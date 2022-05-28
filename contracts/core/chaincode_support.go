/*
	This file is part of go-palletone.
	go-palletone is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.
	go-palletone is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.
	You should have received a copy of the GNU General Public License
	along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
*/

/*
 * Copyright IBM Corp. All Rights Reserved.
 * @author PalletOne core developers <dev@pallet.one>
 * @date 2018
 */

package core

import (
	"bytes"
	"fmt"
	docker "github.com/fsouza/go-dockerclient"
	util "github.com/palletone/go-palletone/vm/common"
	"io"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"

	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/contracts/accesscontrol"
	cfg "github.com/palletone/go-palletone/contracts/contractcfg"
	"github.com/palletone/go-palletone/contracts/platforms"
	"github.com/palletone/go-palletone/contracts/shim"
	"github.com/palletone/go-palletone/core/vmContractPub/ccprovider"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"github.com/palletone/go-palletone/dag/rwset"
	"github.com/palletone/go-palletone/vm/api"
	"github.com/palletone/go-palletone/vm/ccintf"
	"github.com/palletone/go-palletone/vm/controller"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
)

type key string

const (
	// DevModeUserRunsChaincode property allows user to run chaincode in development environment
	DevModeUserRunsChaincode string = "dev"
	//chaincodeStartupTimeoutDefault int    = 5000
	//peerAddressDefault             string = "0.0.0.0:7052"

	//TXSimulatorKey is used to attach ledger simulation context
	TXSimulatorKey key = "txsimulatorkey"

	//HistoryQueryExecutorKey is used to attach ledger history query executor context
	//HistoryQueryExecutorKey key = "historyqueryexecutorkey"

	//glh
	// Mutual TLS auth client key and cert paths in the chaincode container
	TLSClientKeyPath      string = "/etc/palletone/client.key"
	TLSClientCertPath     string = "/etc/palletone/client.crt"
	TLSClientRootCertPath string = "/etc/palletone/peer.crt"
)

//this is basically the singleton that supports the
//entire chaincode framework. It does NOT know about
//chains. Chains are per-proposal entities that are
//setup as part of "join" and go through this object
//via calls to Execute and Deploy chaincodes.
var theChaincodeSupport *ChaincodeSupport

//glh
//use this for ledger access and make sure TXSimulator is being used

func getTxSimulator(context context.Context) rwset.TxSimulator {
	if txsim, ok := context.Value(TXSimulatorKey).(rwset.TxSimulator); ok {
		return txsim
	}
	//chaincode will not allow state operations
	return nil
}

/*
//use this for ledger access and make sure HistoryQueryExecutor is being used
func getHistoryQueryExecutor(context context.Context) ledger.HistoryQueryExecutor {
	if historyQueryExecutor, ok := context.Value(HistoryQueryExecutorKey).(ledger.HistoryQueryExecutor); ok {
		return historyQueryExecutor
	}
	//chaincode will not allow state operations
	return nil
}
*/
//
//chaincode runtime environment encapsulates handler and container environment
//This is where the VM that's running the chaincode would hook in
type chaincodeRTEnv struct {
	handler *Handler
}

// runningChaincodes contains maps of chaincodeIDs to their chaincodeRTEs
type runningChaincodes struct {
	sync.RWMutex
	// chaincode environment for each chaincode
	chaincodeMap map[string]*chaincodeRTEnv

	//mark the starting of launch of a chaincode so multiple requests
	//do not attempt to start the chaincode at the same time
	launchStarted map[string]bool
}

//GetChain returns the chaincode framework support object
func GetChain() *ChaincodeSupport {
	return theChaincodeSupport
}

func (chaincodeSupport *ChaincodeSupport) preLaunchSetup(chaincode string, notfy chan bool) {
	chaincodeSupport.runningChaincodes.Lock()
	defer chaincodeSupport.runningChaincodes.Unlock()
	//register placeholder Handler. This will be transferred in registerHandler
	//NOTE: from this point, existence of handler for this chaincode means the chaincode
	//is in the process of getting started (or has been started)
	chaincodeSupport.runningChaincodes.chaincodeMap[chaincode] = &chaincodeRTEnv{handler: &Handler{readyNotify: notfy}}
}

//call this under lock
func (chaincodeSupport *ChaincodeSupport) chaincodeHasBeenLaunched(chaincode string) (*chaincodeRTEnv, bool) {
	chrte, hasbeenlaunched := chaincodeSupport.runningChaincodes.chaincodeMap[chaincode]
	return chrte, hasbeenlaunched
}

//call this under lock
func (chaincodeSupport *ChaincodeSupport) launchStarted(chaincode string) bool {
	if _, launchStarted := chaincodeSupport.runningChaincodes.launchStarted[chaincode]; launchStarted {
		return true
	}
	return false
}

// NewChaincodeSupport creates a new ChaincodeSupport instance
func NewChaincodeSupport(ccEndpoint string, userrunsCC bool, ccstartuptimeout time.Duration, ca accesscontrol.CA, jury IAdapterJury) pb.ChaincodeSupportServer {
	//path := config.GetPath("peer.fileSystemPath") + string(filepath.Separator) + "chaincodes"
	//path := cfg.GetConfig().ContractFileSystemPath + string(filepath.Separator) + "chaincodes"
	//log.Infof("NewChaincodeSupport chaincodes path: %s, cfgpath[%s]\n", path, cfg.GetConfig().ContractFileSystemPath)

	//ccprovider.SetChaincodesPath(path)
	pnid := viper.GetString("peer.networkId")
	pid := viper.GetString("peer.id")

	theChaincodeSupport = &ChaincodeSupport{
		ca: ca,
		runningChaincodes: &runningChaincodes{
			chaincodeMap:  make(map[string]*chaincodeRTEnv),
			launchStarted: make(map[string]bool),
		}, peerNetworkID: pnid, peerID: pid, jury: jury,
	}

	theChaincodeSupport.auth = accesscontrol.NewAuthenticator(theChaincodeSupport, ca)
	theChaincodeSupport.peerAddress = ccEndpoint
	log.Infof("Chaincode support using peerAddress: %s\n", theChaincodeSupport.peerAddress)

	theChaincodeSupport.userRunsCC = userrunsCC
	theChaincodeSupport.ccStartupTimeout = ccstartuptimeout

	theChaincodeSupport.peerTLS = viper.GetBool("peer.tls.enabled")
	if !theChaincodeSupport.peerTLS {
		theChaincodeSupport.auth.DisableAccessCheck()
	}

	kadef := 120 //心跳定时维护时间,秒
	if ka := viper.GetString("chaincode.keepalive"); ka == "" {
		theChaincodeSupport.keepalive = time.Duration(kadef) * time.Second
	} else {
		t, terr := strconv.Atoi(ka)
		if terr != nil {
			log.Errorf("Invalid keepalive value %s (%s) defaulting to %d", ka, terr, kadef)
			t = kadef
		} else if t <= 0 {
			log.Debugf("Turn off keepalive(value %s)", ka)
			t = kadef
		}
		theChaincodeSupport.keepalive = time.Duration(t) * time.Second
	}

	//default chaincode execute timeout is 30 secs
	execto := time.Duration(30) * time.Second
	//if eto := viper.GetDuration("chaincode.executetimeout"); eto <= time.Duration(1)*time.Second {
	if eto := cfg.GetConfig().ContractExecutetimeout; eto <= time.Duration(1)*time.Second {
		log.Errorf("Invalid execute timeout value %s (should be at least 1s); defaulting to %s", eto, execto)
	} else {
		log.Debugf("Setting execute timeout value to %s", eto)
		execto = eto
	}

	theChaincodeSupport.executetimeout = execto

	viper.SetEnvPrefix("CORE")
	viper.AutomaticEnv()
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)

	//theChaincodeSupport.chaincodeLogLevel = getLogLevelFromViper("level")
	//theChaincodeSupport.shimLogLevel = getLogLevelFromViper("shim")
	//theChaincodeSupport.logFormat = viper.GetString("chaincode.logging.format")

	return theChaincodeSupport.auth
}

/*
// getLogLevelFromViper gets the chaincode container log levels from viper
func getLogLevelFromViper(module string) string {
	//levelString := viper.GetString("chaincode.logging." + module)
	//_, err := logging.LogLevel(levelString)
	//
	//if err == nil {
	//	log.Debugf("CORE_CHAINCODE_%s set to level %s", strings.ToUpper(module), levelString)
	//} else {
	//	log.Warnf("CORE_CHAINCODE_%s has invalid log level %s. defaulting to %s", strings.ToUpper(module), levelString, flogging.DefaultLevel())
	//	levelString = flogging.DefaultLevel()
	//}
	return flogging.DefaultLevel()
}
*/
// ChaincodeSupport responsible for providing interfacing with chaincodes from the Peer.
type ChaincodeSupport struct {
	ca                accesscontrol.CA
	auth              accesscontrol.Authenticator
	runningChaincodes *runningChaincodes
	peerAddress       string
	ccStartupTimeout  time.Duration
	peerNetworkID     string
	peerID            string
	keepalive         time.Duration
	chaincodeLogLevel string
	shimLogLevel      string
	logFormat         string
	executetimeout    time.Duration
	userRunsCC        bool
	peerTLS           bool
	jury              IAdapterJury
}

// DuplicateChaincodeHandlerError returned if attempt to register same chaincodeID while a stream already exists.
type DuplicateChaincodeHandlerError struct {
	ChaincodeID *pb.PtnChaincodeID
}

func (d *DuplicateChaincodeHandlerError) Error() string {
	return fmt.Sprintf("duplicate chaincodeID error: %s", d.ChaincodeID)
}

func newDuplicateChaincodeHandlerError(chaincodeHandler *Handler) error {
	return &DuplicateChaincodeHandlerError{ChaincodeID: chaincodeHandler.ChaincodeID}
}

func (chaincodeSupport *ChaincodeSupport) registerHandler(chaincodehandler *Handler) error {
	key := chaincodehandler.ChaincodeID.Name + chaincodehandler.ChaincodeID.Version

	chaincodeSupport.runningChaincodes.Lock()
	defer chaincodeSupport.runningChaincodes.Unlock()

	chrte2, ok := chaincodeSupport.chaincodeHasBeenLaunched(key)
	if ok && chrte2.handler.registered {
		log.Debugf("duplicate registered handler(key:%s) return error", key)
		// Duplicate, return error
		return newDuplicateChaincodeHandlerError(chaincodehandler)
	}
	//a placeholder, unregistered handler will be setup by transaction processing that comes
	//through via consensus. In this case we swap the handler and give it the notify channel
	if chrte2 != nil {
		chaincodehandler.readyNotify = chrte2.handler.readyNotify
		chrte2.handler = chaincodehandler
	} else {
		if !chaincodeSupport.userRunsCC {
			//this chaincode was not launched by the peer and is attempting
			//to register. Don't allow this.
			return errors.Errorf("peer will not accept external chaincode connection %v (except in dev mode)", chaincodehandler.ChaincodeID)
		}
		chaincodeSupport.runningChaincodes.chaincodeMap[key] = &chaincodeRTEnv{handler: chaincodehandler}
	}

	chaincodehandler.registered = true

	//now we are ready to receive messages and send back responses
	chaincodehandler.txCtxs = make(map[string]*transactionContext)
	chaincodehandler.txidMap = make(map[string]bool)

	log.Debugf("registered handler complete for chaincode %s", key)

	return nil
}

func (chaincodeSupport *ChaincodeSupport) deregisterHandler(chaincodehandler *Handler) error {

	// clean up queryIteratorMap
	//for _, txcontext := range chaincodehandler.txCtxs {
	//	for _, v := range txcontext.queryIteratorMap {
	//		v.Close()
	//	}
	//}

	key := chaincodehandler.ChaincodeID.Name + chaincodehandler.ChaincodeID.Version
	log.Debugf("Deregister handler: %s", key)
	chaincodeSupport.runningChaincodes.Lock()
	defer chaincodeSupport.runningChaincodes.Unlock()
	if _, ok := chaincodeSupport.chaincodeHasBeenLaunched(key); !ok {
		// Handler NOT found
		return errors.Errorf("error deregistering handler, could not find handler with key: %s", key)
	}
	delete(chaincodeSupport.runningChaincodes.chaincodeMap, key)
	log.Debugf("Deregistered handler with key: %s", key)
	return nil
}

// send ready to move to ready state
func (chaincodeSupport *ChaincodeSupport) sendReady(context context.Context, cccid *ccprovider.CCContext, timeout time.Duration) error {
	canName := cccid.GetCanonicalName()
	chaincodeSupport.runningChaincodes.Lock()
	//if its in the map, there must be a connected stream...nothing to do
	var chrte *chaincodeRTEnv
	var ok bool
	if chrte, ok = chaincodeSupport.chaincodeHasBeenLaunched(canName); !ok {
		chaincodeSupport.runningChaincodes.Unlock()
		err := errors.Errorf("handler not found for chaincode %s", canName)
		log.Debugf("%+v", err)
		return err
	}
	chaincodeSupport.runningChaincodes.Unlock()

	var notfy chan *pb.PtnChaincodeMessage
	var err error
	if notfy, err = chrte.handler.ready(context, cccid.ChainID, cccid.TxID, cccid.SignedProposal, cccid.Proposal); err != nil {
		return errors.WithMessage(err, fmt.Sprintf("error sending %s", pb.PtnChaincodeMessage_READY))
	}
	if notfy != nil {
		select {
		case ccMsg := <-notfy:
			if ccMsg.Type == pb.PtnChaincodeMessage_ERROR {
				err = errors.Errorf("error initializing container %s: %s", canName, string(ccMsg.Payload))
			}
			if ccMsg.Type == pb.PtnChaincodeMessage_COMPLETED {
				res := &pb.Response{}
				_ = proto.Unmarshal(ccMsg.Payload, res)
				if res.Status != shim.OK {
					err = errors.Errorf("error initializing container %s: %s", canName, res.Message)
				}
				// TODO
				// return res so that endorser can anylyze it.
			}
		case <-time.After(timeout):
			err = errors.New("timeout expired while executing send init message")
		}
	}

	//if initOrReady succeeded, our responsibility to delete the context
	chrte.handler.deleteTxContext(cccid.ChainID, cccid.TxID)

	return err
}

// returns a map of file path <-> []byte for all files related to TLS
func (chaincodeSupport *ChaincodeSupport) getTLSFiles(keyPair *accesscontrol.CertAndPrivKeyPair) map[string][]byte {
	if keyPair == nil {
		return nil
	}

	return map[string][]byte{
		TLSClientKeyPath:      []byte(keyPair.Key),
		TLSClientCertPath:     []byte(keyPair.Cert),
		TLSClientRootCertPath: chaincodeSupport.ca.CertBytes(),
	}
}

//get args and env given chaincodeID
func (chaincodeSupport *ChaincodeSupport) getLaunchConfigs(cccid *ccprovider.CCContext, cLang pb.PtnChaincodeSpec_Type) (args []string, envs []string, filesToUpload map[string][]byte, err error) {
	canName := cccid.GetCanonicalName()
	envs = []string{"CORE_CHAINCODE_ID_NAME=" + canName}

	// ----------------------------------------------------------------------------
	// Pass TLS options to chaincode
	// ----------------------------------------------------------------------------
	// Note: The peer certificate is only baked into the image during the build
	// phase (see core/chaincode/platforms).  This logic below merely assumes the
	// image is already configured appropriately and is simply toggling the feature
	// on or off.  If the peer's x509 has changed since the chaincode was deployed,
	// the image may be stale and the admin will need to remove the current containers
	// before restarting the peer.
	// ----------------------------------------------------------------------------
	var certKeyPair *accesscontrol.CertAndPrivKeyPair
	if chaincodeSupport.peerTLS {
		certKeyPair, err = chaincodeSupport.auth.Generate(cccid.GetCanonicalName())
		if err != nil {
			return nil, nil, nil, errors.WithMessage(err, fmt.Sprintf("failed generating TLS cert for %s", cccid.GetCanonicalName()))
		}
		envs = append(envs, "CORE_PEER_TLS_ENABLED=true")
		envs = append(envs, fmt.Sprintf("CORE_TLS_CLIENT_KEY_PATH=%s", TLSClientKeyPath))
		envs = append(envs, fmt.Sprintf("CORE_TLS_CLIENT_CERT_PATH=%s", TLSClientCertPath))
		envs = append(envs, fmt.Sprintf("CORE_PEER_TLS_ROOTCERT_FILE=%s", TLSClientRootCertPath))
	} else {
		envs = append(envs, "CORE_PEER_TLS_ENABLED=false")
	}
	chaincodeSupport.chaincodeLogLevel = "info"
	if chaincodeSupport.chaincodeLogLevel != "" {
		envs = append(envs, "CORE_CHAINCODE_LOGGING_LEVEL="+chaincodeSupport.chaincodeLogLevel)
	}

	if chaincodeSupport.shimLogLevel != "" {
		envs = append(envs, "CORE_CHAINCODE_LOGGING_SHIM="+chaincodeSupport.shimLogLevel)
	}
	if chaincodeSupport.peerAddress != "" {
		log.Debugf("-------------------------------------------%s\n\n", chaincodeSupport.peerAddress)
		envs = append(envs, "CORE_CHAINCODE_PEER_ADDRESS="+chaincodeSupport.peerAddress)
	}
	if chaincodeSupport.logFormat != "" {
		envs = append(envs, "CORE_CHAINCODE_LOGGING_FORMAT="+chaincodeSupport.logFormat)
	}
	switch cLang {
	case pb.PtnChaincodeSpec_GOLANG, pb.PtnChaincodeSpec_CAR:
		//args = []string{"chaincode", fmt.Sprintf("-peer.address=%s", chaincodeSupport.peerAddress)}
		//args = []string{"/bin/sh", "-c", "cd / && tar -xvf binpackage.tar -C $GOPATH/bin && rm binpackage.tar && rm Dockerfile && cd $GOPATH/bin && ./chaincode"}
		args = []string{"/bin/sh", "-c", "cd / && tar -xvf binpackage.tar -C $GOPATH/bin && cd $GOPATH/bin && chmod 777 -R ./chaincode && ./chaincode"}
	case pb.PtnChaincodeSpec_JAVA:
		args = []string{"java", "-jar", "chaincode.jar", "--peerAddress", chaincodeSupport.peerAddress}
	case pb.PtnChaincodeSpec_NODE:
		args = []string{"/bin/sh", "-c", fmt.Sprintf("cd /usr/local/src; npm start -- --peer.address %s", chaincodeSupport.peerAddress)}

	default:
		return nil, nil, nil, errors.Errorf("unknown chaincodeType: %s", cLang)
	}

	filesToUpload = theChaincodeSupport.getTLSFiles(certKeyPair)

	log.Debugf("Executable is %s", args[0])
	log.Debugf("Args %v", args)
	log.Debugf("Envs %v", envs)
	log.Debugf("FilesToUpload %v", reflect.ValueOf(filesToUpload).MapKeys())

	return args, envs, filesToUpload, nil
}

//---------- Begin - launchAndWaitForRegister related functionality --------

//a launcher interface to encapsulate chaincode execution. This
//helps with UT of launchAndWaitForRegister
type launcherIntf interface {
	launch(ctxt context.Context, notfy chan bool) (interface{}, error)
}

//ccLaucherImpl will use the container launcher mechanism to launch the actual chaincode
type ccLauncherImpl struct {
	ctxt      context.Context
	ccSupport *ChaincodeSupport
	cccid     *ccprovider.CCContext
	cds       *pb.PtnChaincodeDeploymentSpec
	builder   api.BuildSpecFactory
}

//launches the chaincode using the supplied context and notifier
func (ccl *ccLauncherImpl) launch(ctxt context.Context, notfy chan bool) (interface{}, error) {
	//launch the chaincode，cmd命令参数，环境变量，TLS文件
	args, env, filesToUpload, err := ccl.ccSupport.getLaunchConfigs(ccl.cccid, ccl.cds.ChaincodeSpec.Type)
	if err != nil {
		return nil, err
	}
	canName := ccl.cccid.GetCanonicalName()
	log.Debugf("start container: %s(networkid:%s,peerid:%s)", canName, ccl.ccSupport.peerNetworkID, ccl.ccSupport.peerID)
	log.Debugf("start container with args: %s", strings.Join(args, " "))
	log.Debugf("start container with env:\n\t%s", strings.Join(env, "\n\t"))

	//set up the shadow handler JIT before container launch to
	//reduce window of when an external chaincode can sneak in
	//and use the launching context and make it its own
	preLaunchFunc := func() error {
		ccl.ccSupport.preLaunchSetup(canName, notfy)
		return nil
	}

	ccid := ccintf.CCID{ChaincodeSpec: ccl.cds.ChaincodeSpec, NetworkID: ccl.ccSupport.peerNetworkID, PeerID: ccl.ccSupport.peerID, Version: ccl.cccid.Version}
	sir := controller.StartImageReq{CCID: ccid, Builder: ccl.builder, Args: args, Env: env, FilesToUpload: filesToUpload, PrelaunchFunc: preLaunchFunc}
	ipcCtxt := context.WithValue(ctxt, ccintf.GetCCHandlerKey(), ccl.ccSupport)

	vmtype := ccl.ccSupport.getVMType(ccl.cds)
	resp, err := controller.VMCProcess(ipcCtxt, vmtype, sir)

	return resp, err
}

//launchAndWaitForRegister will launch container if not already running. Use
//the targz to create the image if not found. It uses the supplied launcher
//for launching the chaincode. UTs use the launcher freely to test various
//conditions such as timeouts, failed launches and other errors
func (chaincodeSupport *ChaincodeSupport) launchAndWaitForRegister(ctxt context.Context, cccid *ccprovider.CCContext, cds *pb.PtnChaincodeDeploymentSpec, launcher launcherIntf) error {
	canName := cccid.GetCanonicalName()
	if canName == "" {
		return errors.New("chaincode name not set")
	}

	chaincodeSupport.runningChaincodes.Lock()
	//if its in the map, its either up or being launched. Either case break the
	//multiple launch by failing
	if _, hasBeenLaunched := chaincodeSupport.chaincodeHasBeenLaunched(canName); hasBeenLaunched {
		chaincodeSupport.runningChaincodes.Unlock()
		return errors.Errorf("error chaincode has been launched: %s", canName)
	}

	//prohibit multiple simultaneous invokes (for example while flooding the
	//system with invokes as in a stress test scenario) from attempting to launch
	//the chaincode. The first one wins. Others receive an error.
	//NOTE - this transient behavior as the chaincode is being launched is nothing
	//new. All invokes (except the one launching the CC) will fail in any case
	//until the container is up and registered.
	if chaincodeSupport.launchStarted(canName) {
		chaincodeSupport.runningChaincodes.Unlock()
		return errors.Errorf("error chaincode is already launching: %s", canName)
	}

	//Chaincode is not up and is not in the process of being launched. Setup flag
	//for launching so we can proceed to do that undisturbed by other requests on
	//this chaincode
	log.Debugf("chaincode %s is being launched", canName)
	chaincodeSupport.runningChaincodes.launchStarted[canName] = true

	//now that chaincode launch sequence is done (whether successful or not),
	//unset launch flag as we get out of this function. If launch was not
	//successful (handler was not created), next invoke will try again.
	defer func() {
		chaincodeSupport.runningChaincodes.Lock()
		defer chaincodeSupport.runningChaincodes.Unlock()

		delete(chaincodeSupport.runningChaincodes.launchStarted, canName)
		log.Debugf("chaincode %s launch seq completed", canName)
	}()

	chaincodeSupport.runningChaincodes.Unlock()

	//loopback notifier when everything goes ok and chaincode registers
	//correctly
	notfy := make(chan bool, 1)
	errChan := make(chan error)
	go func() {

		var err error
		defer func() {
			//notify ONLY if we encountered an error
			//else either timeout or ready notify should
			//kick in
			if err != nil {
				errChan <- err
			}
		}()

		log.Debugf("launch info: %+v", launcher)
		resp, err := launcher.launch(ctxt, notfy)
		if err != nil || (resp != nil && resp.(controller.VMCResp).Err != nil) {
			if err == nil {
				err = resp.(controller.VMCResp).Err
			} else {
				log.Debugf("launch error: %+v", err)
			}

			//if the launch was successful and leads to proper registration,
			//this error could be ignored in the select below. On the other
			//hand the error might be triggered for select in which case
			//the launch will be cleaned up
			err = errors.WithMessage(err, "error starting container")
		}
		log.Debugf("")
	}()

	var err error

	//wait for REGISTER state
	select {
	case ok := <-notfy:
		if !ok {
			err = errors.Errorf("registration failed for %s(networkid:%s,peerid:%s,tx:%s)", canName, chaincodeSupport.peerNetworkID, chaincodeSupport.peerID, cccid.TxID)
		}
	case err = <-errChan:
		// When the launch completed, errors from the launch if any will be handled below.
		// Just test for invalid nil error notification (we expect only errors to be notified)
		if err == nil {
			// TODO
			return errors.New("nil error notified. the launch contract is to notify errors only")
			//panic("nil error notified. the launch contract is to notify errors only")
		}
	case <-time.After(chaincodeSupport.ccStartupTimeout):
		err = errors.Errorf("timeout expired while starting chaincode %s(networkid:%s,peerid:%s,tx:%s)", canName, chaincodeSupport.peerNetworkID, chaincodeSupport.peerID, cccid.TxID)
	}
	if err != nil {
		log.Debugf("stopping due to error while launching: %+v", err)
		errIgnore := chaincodeSupport.Stop(ctxt, cccid, cds, false)
		if errIgnore != nil {
			log.Debugf("stop failed: %+v", errIgnore)
		}
		//TODO
	}
	return err
}

//---------- End - launchAndWaitForRegister related functionality --------

//Stop stops a chaincode if running
func (chaincodeSupport *ChaincodeSupport) Stop(context context.Context, cccid *ccprovider.CCContext, cds *pb.PtnChaincodeDeploymentSpec, dontRmCon bool) error {
	canName := cccid.GetCanonicalName()
	if canName == "" {
		return errors.New("chaincode name not set")
	} else {
		log.Debugf("stopping : %+v", canName)
	}

	//stop the chaincode
	//sir := container.StopImageReq{CCID: ccintf.CCID{ChaincodeSpec: cds.ChaincodeSpec, NetworkID: chaincodeSupport.peerNetworkID, PeerID: chaincodeSupport.peerID, Version: cccid.Version}, Timeout: 0}
	// The line below is left for debugging. It replaces the line above to keep
	// the chaincode container around to give you a chance to get data
	sir := controller.StopImageReq{CCID: ccintf.CCID{ChaincodeSpec: cds.ChaincodeSpec, NetworkID: chaincodeSupport.peerNetworkID, PeerID: chaincodeSupport.peerID, ChainID: "" /*cccid.ChainID*/ , Version: cccid.Version}, Timeout: 0, Dontremove: dontRmCon}
	vmtype := chaincodeSupport.getVMType(cds)

	_, err := controller.VMCProcess(context, vmtype, sir)
	if err != nil {
		err = errors.WithMessage(err, "error stopping container")
		//but proceed to cleanup
	}

	chaincodeSupport.runningChaincodes.Lock()
	if _, ok := chaincodeSupport.chaincodeHasBeenLaunched(canName); !ok {
		//nothing to do
		chaincodeSupport.runningChaincodes.Unlock()
		return nil
	}

	delete(chaincodeSupport.runningChaincodes.chaincodeMap, canName)
	chaincodeSupport.runningChaincodes.Unlock()

	return err
}

func (chaincodeSupport *ChaincodeSupport) Destroy(context context.Context, cccid *ccprovider.CCContext,
	cds *pb.PtnChaincodeDeploymentSpec) error {
	canName := cccid.GetCanonicalName()
	if canName == "" {
		return errors.New("chaincode name not set")
	} else {
		log.Debugf("destroy : %+v", canName)
	}

	sir := controller.DestroyImageReq{
		CCID: ccintf.CCID{
			ChaincodeSpec: cds.ChaincodeSpec,
			NetworkID:     chaincodeSupport.peerNetworkID,
			PeerID:        chaincodeSupport.peerID,
			ChainID:       "",
			Version:       cccid.Version,
		},
		Timeout: 0,
		Force:   true,
		NoPrune: false,
	}
	vmtype := chaincodeSupport.getVMType(cds)

	_, err := controller.VMCProcess(context, vmtype, sir)
	if err != nil {
		err = errors.WithMessage(err, "error destroy container")
	}

	//chaincodeSupport.runningChaincodes.Lock()
	//if _, ok := chaincodeSupport.chaincodeHasBeenLaunched(canName); !ok {
	//	chaincodeSupport.runningChaincodes.Unlock()
	//	return nil
	//}
	//
	//delete(chaincodeSupport.runningChaincodes.chaincodeMap, canName)
	//chaincodeSupport.runningChaincodes.Unlock()

	return err
}

// Launch will launch the chaincode if not running (if running return nil) and will wait for handler of the chaincode to get into FSM ready state.
func (chaincodeSupport *ChaincodeSupport) Launch(context context.Context,
	cccid *ccprovider.CCContext, spec interface{}) (*pb.PtnChaincodeID, *pb.PtnChaincodeInput, error) {
	log.Debugf("launch enter")
	//build the chaincode
	var cID *pb.PtnChaincodeID
	var cMsg *pb.PtnChaincodeInput

	var cds *pb.PtnChaincodeDeploymentSpec
	var ci *pb.PtnChaincodeInvocationSpec

	log.Infof("chainId=%s, name=%s, version=%s, syscc=%v", cccid.ChainID, cccid.Name, cccid.Version, cccid.Syscc)
	if cds, _ = spec.(*pb.PtnChaincodeDeploymentSpec); cds == nil {
		if ci, _ = spec.(*pb.PtnChaincodeInvocationSpec); ci == nil {
			//  TODO
			return cID, cMsg, errors.New("Launch should be called with deployment or invocation spec")
			//panic("Launch should be called with deployment or invocation spec")
		}
	}
	if cds != nil {
		cID = cds.ChaincodeSpec.ChaincodeId
		cMsg = cds.ChaincodeSpec.Input
		log.Debugf("cds != nil-------这是部署合约---------------， cID=%v", cID)
	} else {
		cID = ci.ChaincodeSpec.ChaincodeId
		cMsg = ci.ChaincodeSpec.Input
		log.Debugf("cds == nil---------这是调用合约-------------, cID=%v", cID)
	}

	canName := cccid.GetCanonicalName()
	log.Debugf("canName= %s", canName)
	chaincodeSupport.runningChaincodes.Lock()
	var chrte *chaincodeRTEnv
	var ok bool
	var err error
	//if its in the map, there must be a connected stream...nothing to do
	if chrte, ok = chaincodeSupport.chaincodeHasBeenLaunched(canName); ok {
		log.Debugf("chaincode has been launched")
		if !chrte.handler.registered {
			chaincodeSupport.runningChaincodes.Unlock()
			err = errors.Errorf("premature execution - chaincode (%s) launched and waiting for registration", canName)
			log.Debugf("%+v", err)
			return cID, cMsg, err
		}
		if chrte.handler.isRunning() {
			//if log.IsEnabledFor(logging.DEBUG) {
			log.Debugf("chaincode is running(no need to launch) : %s", canName)
			//}
			chaincodeSupport.runningChaincodes.Unlock()
			return cID, cMsg, nil
		}
		log.Debugf("Container not in READY state(%s)...send init/ready", chrte.handler.FSM.Current())
	} else {
		log.Debugf("chaincode is not up,but launch started")
		//chaincode is not up... but is the launch process underway? this is
		//strictly not necessary as the actual launch process will catch this
		//(in launchAndWaitForRegister), just a bit of optimization for thundering
		//herds
		if chaincodeSupport.launchStarted(canName) {
			chaincodeSupport.runningChaincodes.Unlock()
			err = errors.Errorf("premature execution - chaincode (%s) is being launched", canName)
			return cID, cMsg, err
		}
	}
	chaincodeSupport.runningChaincodes.Unlock()
	if cds == nil {
		return cID, cMsg, errors.Errorf("contract not running:%s", canName)
		//if cccid.Syscc {
		//	return cID, cMsg, errors.Errorf("a syscc should be running (it cannot be launched) %s", canName)
		//}
		//
		//if chaincodeSupport.userRunsCC {
		//	log.Error("You are attempting to perform an action other than Deploy on Chaincode that is not ready and you are in developer mode. Did you forget to Deploy your chaincode?")
		//}
		//
		//var depPayload []byte
		////hopefully we are restarting from existing image and the deployed transaction exists
		////(this will also validate the ID from the LSCC if we're not using the config-tree approach)
		//depPayload, err = GetCDS(cccid.ContractId, context, cccid.TxID, cccid.SignedProposal, cccid.Proposal, cccid.ChainID, cID.Name)
		//if err != nil {
		//	return cID, cMsg, errors.WithMessage(err, fmt.Sprintf("could not get ChaincodeDeploymentSpec for %s", canName))
		//}
		//if depPayload == nil {
		//	return cID, cMsg, errors.WithMessage(err, fmt.Sprintf("nil ChaincodeDeploymentSpec for %s", canName))
		//}
		//
		//cds = &pb.PtnChaincodeDeploymentSpec{}
		//err = proto.Unmarshal(depPayload, cds)
		//if err != nil {
		//	return cID, cMsg, errors.Wrap(err, fmt.Sprintf("failed to unmarshal deployment transactions for %s", canName))
		//}
	}

	//from here on : if we launch the container and get an error, we need to stop the container
	//launch container if it is a System container or not in dev mode
	if (!chaincodeSupport.userRunsCC || cds.ExecEnv == pb.PtnChaincodeDeploymentSpec_SYSTEM) && (chrte == nil || chrte.handler == nil) {
		//NOTE-We need to streamline code a bit so the data from LSCC gets passed to this thus
		//avoiding the need to go to the FS. In particular, we should use cdsfs completely. It is
		//just a vestige of old protocol that we continue to use ChaincodeDeploymentSpec for
		//anything other than Install. In particular, instantiate, invoke, upgrade should be using
		//just some form of ChaincodeInvocationSpec.
		//
		//But for now, if we are invoking we have gone through the LSCC path above. If  instantiating
		//or upgrading currently we send a CDS with nil CodePackage. In this case the codepath
		//in the endorser has gone through LSCC validation. Just get the code from the FS.
		//if cds.CodePackage == nil {
		//	//no code bytes for these situations
		//	if !(chaincodeSupport.userRunsCC || cds.ExecEnv == pb.PtnChaincodeDeploymentSpec_SYSTEM) {
		//		ccpack, err := ccprovider.GetChaincodeFromFS(cID.Name, cID.Version)
		//		if err != nil {
		//			return cID, cMsg, err
		//		}
		//		cds = ccpack.GetDepSpec()
		//		log.Debugf("launchAndWaitForRegister fetched %d bytes from file system", len(cds.CodePackage))
		//	}
		//}

		builder := func() (io.Reader, error) { return platforms.GenerateDockerBuild(cds) }
		err = chaincodeSupport.launchAndWaitForRegister(context, cccid, cds, &ccLauncherImpl{context, chaincodeSupport, cccid, cds, builder})
		if err != nil {
			log.Debugf("launchAndWaitForRegister failed: %+v", err)
			return cID, cMsg, err
		}
	}

	if err == nil {
		//launch will set the chaincode in Ready state
		err = chaincodeSupport.sendReady(context, cccid, chaincodeSupport.ccStartupTimeout)
		if err != nil {
			err = errors.WithMessage(err, "failed to init chaincode")
			log.Errorf("%+v", err)
			errIgnore := chaincodeSupport.Stop(context, cccid, cds, false)
			if errIgnore != nil {
				log.Errorf("stop failed: %+v", errIgnore)
			}
		}
		log.Debug("sending init completed")
	}
	log.Debug("LaunchChaincode complete")

	return cID, cMsg, err
}

//getVMType - just returns a string for now. Another possibility is to use a factory method to
//return a VM executor
func (chaincodeSupport *ChaincodeSupport) getVMType(cds *pb.PtnChaincodeDeploymentSpec) string {
	if cds.ExecEnv == pb.PtnChaincodeDeploymentSpec_SYSTEM {
		return controller.SYSTEM
	}
	return controller.DOCKER
}

// HandleChaincodeStream implements ccintf.HandleChaincodeStream for all vms to call with appropriate stream
func (chaincodeSupport *ChaincodeSupport) HandleChaincodeStream(ctxt context.Context, stream ccintf.ChaincodeStream) error {
	return HandleChaincodeStream(chaincodeSupport, ctxt, stream, chaincodeSupport.jury)
}

// Register the bidi stream entry point called by chaincode to register with the Peer.
func (chaincodeSupport *ChaincodeSupport) Register(stream pb.ChaincodeSupport_RegisterServer) error {
	return chaincodeSupport.HandleChaincodeStream(stream.Context(), stream)
}

// createCCMessage creates a transaction message.
func createCCMessage(contractid []byte, typ pb.PtnChaincodeMessage_Type, cid string, txid string, cMsg *pb.PtnChaincodeInput) (*pb.PtnChaincodeMessage, error) {
	payload, err := proto.Marshal(cMsg)
	if err != nil {
		return nil, err
	}
	return &pb.PtnChaincodeMessage{Type: typ, Payload: payload, Txid: txid, ChannelId: cid, ContractId: contractid}, nil
}

// Execute executes a transaction and waits for it to complete until a timeout value.
func (chaincodeSupport *ChaincodeSupport) Execute(ctxt context.Context, cccid *ccprovider.CCContext, msg *pb.PtnChaincodeMessage, timeout time.Duration) (*pb.PtnChaincodeMessage, error) {
	log.Debugf("chain code support execute")
	log.Debugf("Entry, chainId[%s], txid[%s]", msg.ChannelId, msg.Txid)
	defer log.Debugf("Exit")
	//glh
	setTimeout := 5 * time.Second //default chaincode exectute timeout
	if timeout > 0 {
		setTimeout = timeout
	}

	canName := cccid.GetCanonicalName()
	log.Debugf("chaincode canonical name: %s", canName)
	chaincodeSupport.runningChaincodes.Lock()
	//we expect the chaincode to be running... sanity check
	chrte, ok := chaincodeSupport.chaincodeHasBeenLaunched(canName)
	if !ok {
		chaincodeSupport.runningChaincodes.Unlock()
		log.Debugf("cannot execute-chaincode is not running: %s", canName)
		return nil, errors.Errorf("cannot execute transaction for %s", canName)
	}
	chaincodeSupport.runningChaincodes.Unlock()

	var notfy chan *pb.PtnChaincodeMessage
	var err error
	if notfy, err = chrte.handler.sendExecuteMessage(ctxt, cccid.ChainID, msg, cccid.SignedProposal, cccid.Proposal); err != nil {
		return nil, errors.WithMessage(err, fmt.Sprintf("error sending"))
	}
	var ccresp *pb.PtnChaincodeMessage
	select {
	case ccresp = <-notfy:
		log.Debugf("notfy = %v", ccresp)
		//response is sent to user or calling chaincode. ChaincodeMessage_ERROR
		//are typically treated as error
		//log.Errorf("{{{{{ time out [%d]", setTimeout)
	case <-time.After(setTimeout):
		log.Debugf("time out when execute,time = %v", setTimeout)
		//err = errors.New("timeout expired while executing transaction")
		//log.Info("====================================timeout expired while executing transaction")
		//  试图从容器获取错误信息
		containerErrStr := getLogFromContainer(cccid.GetContainerName())
		if containerErrStr != "" {
			log.Error("error from container %s", containerErrStr)
			err = errors.New(containerErrStr)
		} else {
			//log.Info("===================2=================timeout expired while executing transaction")
			log.Errorf("<<<txid[%s] time out [%d]", cccid.TxID, setTimeout)
			err = errors.New("timeout expired while executing transaction")
		}
		//  调用合约超时，停止该容器
		stopContainerWhenInvokeTimeOut(cccid.GetContainerName())

	}
	//our responsibility to delete transaction context if sendExecuteMessage succeeded
	chrte.handler.deleteTxContext(msg.ChannelId, msg.Txid)
	return ccresp, err
}

// IsDevMode returns true if the peer was configured with development-mode enabled
func IsDevMode() bool {
	mode := viper.GetString("chaincode.mode")

	return mode == DevModeUserRunsChaincode
}

//  当调用合约时，发生超时，即停止掉容器
func stopContainerWhenInvokeTimeOut(name string) {
	log.Debugf("enter StopContainerWhenInvokeTimeOut name = %s", name)
	defer log.Debugf("exit StopContainerWhenInvokeTimeOut name = %s", name)
	client, err := util.NewDockerClient()
	if err != nil {
		log.Error("util.NewDockerClient", "error", err)
		return
	}
	err = client.StopContainer(name, 3)
	if err != nil {
		log.Infof("stop container error: %s", err.Error())
		return
	}
}

//  通过容器名称获取容器里面的错误信息，返回最后一条
func getLogFromContainer(name string) string {
	client, err := util.NewDockerClient()
	if err != nil {
		log.Error("util.NewDockerClient", "error", err)
		return ""
	}
	var buf bytes.Buffer
	logsO := docker.LogsOptions{
		Container:         name,
		ErrorStream:       &buf,
		Follow:            true,
		Stderr:            true,
		InactivityTimeout: 3 * time.Second,
	}
	log.Debugf("start docker logs")
	err = client.Logs(logsO)
	log.Debugf("end docker logs")
	if err != nil {
		log.Infof("get log from container %s error: %s", name, err.Error())
		return ""
	}
	errArray := make([]string, 0)
	for {
		line, err := buf.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return ""
		}
		line = strings.TrimSpace(line)
		if strings.Contains(line, "panic: runtime error") || strings.Contains(line, "fatal error: runtime") {
			log.Infof("container %s error %s", name, line)
			errArray = append(errArray, line)
		}
	}
	if len(errArray) != 0 {
		return errArray[len(errArray)-1]
	}
	return ""
}
