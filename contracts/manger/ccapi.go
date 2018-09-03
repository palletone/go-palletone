package manger

import (
	"time"
	"errors"
	"fmt"
	"golang.org/x/net/context"

	cclist "github.com/palletone/go-palletone/contracts/list"
	cp "github.com/palletone/go-palletone/common/crypto"
	"github.com/palletone/go-palletone/contracts/scc"
	"github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"github.com/palletone/go-palletone/core/vmContractPub/crypto"
	"github.com/palletone/go-palletone/contracts/ucc"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	unit "github.com/palletone/go-palletone/dag/modules"
	"bytes"
	"container/list"
	"github.com/spf13/viper"
	"encoding/hex"
)

var debugX bool = true

type TempCC struct {
	templateId []byte
	name       string
	path       string
	vers       string
}

var listCC list.List

func listAdd(cc *TempCC) error {
	if cc != nil {
		//fmt.Printf("==name[%s]", cc.name)
		listCC.PushBack(*cc)
	}
	return nil
}

func listDel(templateId []byte) {
	for e := listCC.Front(); e != nil; e = e.Next() {
		if bytes.Equal(e.Value.(TempCC).templateId, templateId) == true {
			listCC.Remove(e)
		}
	}
}

func listGet(templateId []byte) (*TempCC, error) {
	logger.Infof("listGet [%v]", templateId)
	for e := listCC.Front(); e != nil; e = e.Next() {
		if bytes.Equal(e.Value.(TempCC).templateId, templateId) == true {
			cc := &TempCC{
				templateId: templateId,
				name:       e.Value.(TempCC).name,
				path:       e.Value.(TempCC).path,
				vers:       e.Value.(TempCC).vers,
			}
			//fmt.Printf("==name[%s]", cc.name)
			return cc, nil
		}
	}
	return nil, errors.New("not find")
}

// contract manger module init
func Init() error {
	err := peerServerInit()
	if err != nil {
		logger.Errorf("peerServerInit error:%s", err)
		return err
	}
	err = systemContractInit()
	if err != nil {
		logger.Errorf("systemContractInit error:%s", err)
		return err
	}

	return nil
}

func InitNoSysCCC() error {
	err := peerServerInit()
	if err != nil {
		logger.Errorf("peerServerInit error:%s", err)
		return err
	}
	//err = systemContractInit()
	//if err != nil {
	//	logger.Errorf("systemContractInit error:%s", err)
	//	return err
	//}
	return nil
}

func Deinit() error {
	err := peerServerDeInit()
	if err != nil {
		logger.Errorf("peerServerDeInit error:%s", err)
		return err
	}
	err = systemContractDeInit()
	if err != nil {
		logger.Errorf("systemContractDeInit error:%s", err)
		return err
	}
	return nil
}

func GetSysCCList() (ccInf []cclist.CCInfo, ccCount int, errs error) {
	scclist := make([]cclist.CCInfo, 0)
	ci := cclist.CCInfo{}

	cclist, count, err := scc.SysCCsList()
	for _, ccinf := range cclist {
		ci.Name = ccinf.Name
		ci.Path = ccinf.Path
		ci.Enable = ccinf.Enabled
		ci.SysCC = true

		scclist = append(scclist, ci)
	}
	return scclist, count, err
}

func GetUsrCCList() {
}

//install but not into db
func Install(chainID string, ccName string, ccPath string, ccVersion string) (payload *unit.ContractTplPayload, err error) {
	logger.Infof("==========install enter=======")
	logger.Infof("name[%s]path[%s]version[%s]", ccName, ccPath, ccVersion)
	defer logger.Infof("-----------install exit--------")
	usrcc := &ucc.UserChaincode{
		Name:    ccName,
		Path:    ccPath,
		Version: ccVersion,
		Enabled: true,
	}

	paylod, err := ucc.GetUserCCPayload(chainID, usrcc)
	if err != nil {
		return nil, err
	}

	var buffer bytes.Buffer
	buffer.Write([]byte(ccName))
	buffer.Write([]byte(ccPath))
	buffer.Write([]byte(ccVersion))
	tpid := cp.Keccak256Hash(buffer.Bytes())

	payloadUnit := &unit.ContractTplPayload{
		TemplateId: []byte(tpid[:]),
		Name:       ccName,
		Path:       ccPath,
		Version:    ccVersion,
		Bytecode:   paylod,
	}

	//test
	tcc := &TempCC{templateId: []byte(tpid[:]), name: ccName, path: ccPath, vers: ccVersion}
	listAdd(tcc)

	return payloadUnit, nil
}

func DeployByName(chainID string, txid string, ccName string, ccPath string, ccVersion string, args [][]byte, timeout time.Duration) (depllyId []byte, respPayload *peer.ContractDeployPayload, e error) {
	var mksupt Support = &SupportImpl{}
	setChainId := "palletone"
	setTimeOut := time.Duration(30) * time.Second

	if chainID != "" {
		setChainId = chainID
	}
	if timeout > 0 {
		setTimeOut = timeout
	}
	if txid == "" || ccName == "" || ccPath == "" {
		return nil, nil, errors.New("input param is nil")
	}
	randNum, err := crypto.GetRandomNonce()
	if err != nil {
		return nil, nil, errors.New("crypto.GetRandomNonce error")
	}

	txsim, err := mksupt.GetTxSimulator(chainID, txid)
	if err != nil {
		return nil, nil, errors.New("GetTxSimulator error")
	}
	//randNum, err := crypto.GetRandomNonce()
	//if err != nil {
	//	return nil, nil, errors.New("crypto.GetRandomNonce error")
	//}

	usrcc := &ucc.UserChaincode{
		Name:     ccName,
		Path:     ccPath,
		Version:  ccVersion,
		InitArgs: args,
		Enabled:  true,
	}

	err = ucc.DeployUserCC(setChainId, usrcc, txid, txsim, setTimeOut)
	if err != nil {
		return nil, nil, errors.New("Deploy fail")
	}

	cc := &cclist.CCInfo{
		Id:      randNum,
		Name:    ccName,
		Path:    ccPath,
		Version: ccVersion,
		SysCC:   false,
		Enable:  true,
	}
	err = cclist.SetChaincode(setChainId, 0, cc)
	if err != nil {
		logger.Errorf("setchaincode[%s]-[%s] fail", setChainId, cc.Name)
	}

	return cc.Id, nil, err
}

func Deploy(chainID string, templateId []byte, txid string, args [][]byte, timeout time.Duration) (deployId []byte, deployPayload *unit.ContractDeployPayload, e error) {
	logger.Infof("==========Deploy enter=======")
	logger.Infof("templateId[%s]txid[%s]", hex.EncodeToString(templateId), txid)
	defer logger.Infof("-----------Deploy exit--------")

	var mksupt Support = &SupportImpl{}
	setChainId := "palletone"
	setTimeOut := time.Duration(30) * time.Second
	logger.Infof("enter, chainId[%s], txid[%s],templateId[%s]", chainID, txid, string(templateId))
	defer logger.Info("exit")

	if chainID != "" {
		setChainId = chainID
	}
	if timeout > 0 {
		setTimeOut = timeout
	}

	templateCC, err := ucc.RecoverChainCodeFromDb(chainID, templateId)
	if err != nil {
		logger.Errorf("chainid[%s]-templateId[%s], RecoverChainCodeFromDb fail", chainID, templateId)
		return nil, nil, err
	}

	//test!!!!!!
	//todo del
	if txid == "" || templateCC.Name == "" || templateCC.Path == "" {
		//logger.Errorf("cc param is null")
		//return "", nil, errors.New("cc param is nil")

		//test
		tmpcc, err := listGet(templateId)
		if err == nil {
			templateCC.Name = tmpcc.name
			templateCC.Path = tmpcc.path
			templateCC.Version = tmpcc.vers
		}
	}

	txsim, err := mksupt.GetTxSimulator(chainID, txid)
	if err != nil {
		return nil, nil, errors.New("GetTxSimulator error")
	}
	randNum, err := crypto.GetRandomNonce()
	if err != nil {
		return nil, nil, errors.New("crypto.GetRandomNonce error")
	}

	usrccName := templateCC.Name + "_" + hex.EncodeToString(randNum)[0:8]//createDeployId(templateCC.Name)

	usrcc := &ucc.UserChaincode{
		Name:     usrccName,
		Path:     templateCC.Path,
		Version:  templateCC.Version,
		InitArgs: args,
		Enabled:  true,
	}

	err = ucc.DeployUserCC(setChainId, usrcc, txid, txsim, setTimeOut)
	if err != nil {
		return nil, nil, errors.New("Deploy fail")
	}
	cc := &cclist.CCInfo{
		Id:      randNum,
		Name:    usrccName,
		Path:    templateCC.Path,
		Version: templateCC.Version,
		SysCC:   false,
		Enable:  true,
	}
	err = cclist.SetChaincode(setChainId, 0, cc)
	if err != nil {
		logger.Errorf("setchaincode[%s]-[%s] fail", setChainId, cc.Name)
	}

	unit, err := RwTxResult2DagDeployUnit(txsim, templateId, txid, cc.Name, cc.Id, args, timeout)
	if err != nil {
		logger.Errorf("chainID[%s] converRwTxResult2DagUnit failed", chainID)
		return nil, nil, errors.New("Conver RwSet to dag unit fail")
	}

	return cc.Id, unit, err
}

//timeout:ms
// ccName can be contract Id
//func Invoke(chainID string, deployId []byte, txid string, args [][]byte, timeout time.Duration) (*peer.ContractInvokePayload, error) {
func Invoke(chainID string, deployId []byte, txid string, args [][]byte, timeout time.Duration) (*unit.ContractInvokePayload, error) {
	logger.Infof("==========Invoke enter=======")
	logger.Infof("deployId[%s] txid[%s]", hex.EncodeToString(deployId), txid)
	defer logger.Infof("-----------Invoke exit--------")

	var mksupt Support = &SupportImpl{}
	creator := []byte("palletone") //default

	cc, err := cclist.GetChaincode(chainID, deployId)
	if err != nil {
		return nil, err
	}
	if cc.Name == "" {
		errstr := fmt.Sprintf("chainCode[%v] not deplay!!", deployId)
		return nil, errors.New(errstr)
	}

	logger.Infof("Invoke [%s][%s]", chainID, cc.Name)
	start := time.Now()
	es := NewEndorserServer(mksupt)
	spec := &pb.ChaincodeSpec{
		ChaincodeId: &pb.ChaincodeID{Name: cc.Name},
		Type:        pb.ChaincodeSpec_GOLANG,
		Input:       &pb.ChaincodeInput{Args: args},
	}

	cid := &pb.ChaincodeID{
		Path:    "", //no use
		Name:    cc.Name,
		Version: cc.Version,
	}

	sprop, prop, err := signedEndorserProposa(chainID, txid, spec, creator, []byte("msg1"))
	if err != nil {
		logger.Errorf("signedEndorserProposa error[%v]", err)
		return nil, err
	}

	rsp, unit, err := es.ProcessProposal(deployId, context.Background(), sprop, prop, chainID, cid, timeout)
	t0 := time.Now()
	duration := t0.Sub(start)

	if err != nil {
		logger.Errorf("ProcessProposal error[%v]", err)
		return nil, err
	}

	logger.Infof("Invoke Ok, ProcessProposal duration=%v,rsp=%v,%s", duration, rsp, unit.Payload)

	return unit, nil
}

func StopByName(chainID string, txid string, ccName string, ccPath string, ccVersion string, deleteImage bool) error {
	usrcc := &ucc.UserChaincode{
		Name:    ccName,
		Path:    ccPath,
		Version: ccVersion,
		Enabled: true,
	}
	err := ucc.StopUserCC(chainID, usrcc, txid, deleteImage)
	if err != nil {
		errMsg := fmt.Sprintf("StopUserCC err[%s]-[%s]-err[%s]", chainID, ccName, err)
		return errors.New(errMsg)
	}

	return nil
}

func Stop(chainID string, deployId []byte, txid string, deleteImage bool) error {
	logger.Infof("==========Stop enter=======")
	logger.Infof("deployId[%s]txid[%s]", hex.EncodeToString(deployId), txid)
	defer logger.Infof("-----------Stop exit--------")

	setChainId := "palletone"
	if chainID != "" {
		setChainId = chainID
	}
	if txid == "" {
		return errors.New("input param txid is nil")
	}

	//clist, err := getChaincodeList(chainID)
	//if err != nil {
	//	logger.Errorf("not find chainlist for chainId[%s]", chainID)
	//	return errors.New("getChaincodeList failed")
	//}
	//
	//for k, v := range clist.cclist {
	//	logger.Infof("chaincode[%s]:%v", k, *v)
	//	if k == chainID {
	//		if bytes.Equal(v.Id, deployId) == true {
	//			return StopByName(setChainId, txid, v.Name, v.Path, v.Version, deleteImage)
	//		}
	//	}
	//}

	cc, err := cclist.GetChaincode(chainID, deployId)
	if err != nil {
		return err
	}
	err = StopByName(setChainId, txid, cc.Name, cc.Path, cc.Version, deleteImage)
	if err == nil {
		cclist.DelChaincode(chainID, cc.Name, cc.Version)
	}
	return err
}

func peerContractMockConfigInit() {
	viper.Set("peer.fileSystemPath", "./chaincodes")
	viper.Set("peer.address", "127.0.0.1:12345")
	viper.Set("chaincode.executetimeout", 20*time.Second)

	viper.Set("vm.endpoint", "unix:///var/run/docker.sock")
	viper.Set("chaincode.builder", "palletimg")

	viper.Set("chaincode.system", map[string]string{"sample_syscc": "true"})
}

func init() {
	peerContractMockConfigInit()

	//InitNoSysCCC()
	Init()
}
