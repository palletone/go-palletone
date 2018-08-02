package ucc

import (
	"fmt"
	"golang.org/x/net/context"

	"github.com/palletone/go-palletone/core/vmContractPub/flogging"
	"github.com/palletone/go-palletone/core/vmContractPub/util"
	"github.com/palletone/go-palletone/contracts/shim"
	"github.com/palletone/go-palletone/core/vmContractPub/ccprovider"

	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"github.com/palletone/go-palletone/vm/controller"
	"time"
)


type UserChaincode struct {
	//Unique name of the chaincode
	Name string

	//Path to the chaincode; currently not used
	Path string

	Version string

	//InitArgs initialization arguments to startup the chaincode
	InitArgs [][]byte

	// Chaincode is the actual chaincode object
	Chaincode shim.Chaincode

	//InvokableExternal bool

	// InvokableCC2CC keeps track of whether this chaincode can be invoked
	// by way of a chaincode-to-chaincode invocation
	InvokableCC2CC bool

	// Enabled a convenient switch to enable/disable chaincode without
	// having to remove entry from importsysccs.go
	Enabled bool
}
var logger = flogging.MustGetLogger("uccapi")

// buildLocal builds a given chaincode code
func buildUserCC(context context.Context, spec *pb.ChaincodeSpec) (*pb.ChaincodeDeploymentSpec, error) {
	var codePackageBytes []byte
	chaincodeDeploymentSpec := &pb.ChaincodeDeploymentSpec{ExecEnv: pb.ChaincodeDeploymentSpec_DOCKER, ChaincodeSpec: spec, CodePackage: codePackageBytes}
	return chaincodeDeploymentSpec, nil
}

func getDeploymentSpec(_ context.Context, spec *pb.ChaincodeSpec) (*pb.ChaincodeDeploymentSpec, error) {
	fmt.Printf("getting deployment spec for chaincode spec: %v\n", spec)
	codePackageBytes, err := controller.GetChaincodePackageBytes(spec)
	if err != nil {
		return nil, err
	}
	cdDeploymentSpec := &pb.ChaincodeDeploymentSpec{ChaincodeSpec: spec, CodePackage: codePackageBytes}
	return cdDeploymentSpec, nil
}


func DeployUserCC(chainID string, usrcc *UserChaincode, timeout time.Duration) error {
	//if !usrcc.Enabled || !isWhitelisted(usrcc) {
	//	logger.Info(fmt.Sprintf("system chaincode (%s,%s) disabled", usrcc.Name, usrcc.Path))
	//	return nil
	//}
	var err error

	ccprov := ccprovider.GetChaincodeProvider()
	txid := util.GenerateUUID()
	ctxt := context.Background()
	//glh
	/*
	if chainID != "" {
		lgr := peer.GetLedger(chainID)
		if lgr == nil {
			panic(fmt.Sprintf("syschain %s start up failure - unexpected nil ledger for channel %s", syscc.Name, chainID))
		}

		//init can do GetState (and other Get's) even if Puts cannot be
		//be handled. Need ledger for this
		ctxt2, txsim, err := ccprov.GetContext(lgr, txid)
		if err != nil {
			return err
		}
		ctxt = ctxt2
		defer txsim.Done()
	}
*/
	chaincodeID := &pb.ChaincodeID{Path: usrcc.Path, Name: usrcc.Name, Version:usrcc.Version}
	spec := &pb.ChaincodeSpec{Type: pb.ChaincodeSpec_Type(pb.ChaincodeSpec_Type_value["GOLANG"]), ChaincodeId: chaincodeID, Input: &pb.ChaincodeInput{Args: usrcc.InitArgs}}

	// First build and get the deployment spec
	chaincodeDeploymentSpec, err := getDeploymentSpec(ctxt, spec)

	if err != nil {
		logger.Error(fmt.Sprintf("Error deploying chaincode spec: %v\n\n error: %s", spec, err))
		return err
	}
	version := "aaaaa"

	cccid := ccprov.GetCCContext(chainID, chaincodeDeploymentSpec.ChaincodeSpec.ChaincodeId.Name, version, txid, false, nil, nil)
	_, _, err = ccprov.ExecuteWithErrorFilter(ctxt, cccid, chaincodeDeploymentSpec,timeout)

	if err != nil {
		logger.Errorf("ExecuteWithErrorFilter with usercc.Name[%s] chainId[%s] err !!", usrcc.Name, chainID)
	}
	logger.Infof("user chaincode %s/%s(%s) deployed", usrcc.Name, chainID, usrcc.Path)

	return err
}

func StopUserCC(chainID string, usrcc *UserChaincode, timeout time.Duration) error {
	//theChaincodeSupport.Stop(ctxt, cccid, &pb.ChaincodeDeploymentSpec{ChaincodeSpec: spec})
	return nil
}