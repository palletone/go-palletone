package manger

import (
	"bytes"
	"container/list"
	"encoding/hex"
	"fmt"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
	"time"

	"github.com/palletone/go-palletone/common"
	cp "github.com/palletone/go-palletone/common/crypto"
	"github.com/palletone/go-palletone/common/log"
	db "github.com/palletone/go-palletone/contracts/comm"
	cfg "github.com/palletone/go-palletone/contracts/contractcfg"
	"github.com/palletone/go-palletone/contracts/core"
	cclist "github.com/palletone/go-palletone/contracts/list"
	"github.com/palletone/go-palletone/contracts/scc"
	"github.com/palletone/go-palletone/contracts/ucc"
	"github.com/palletone/go-palletone/core/vmContractPub/crypto"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"github.com/palletone/go-palletone/dag"
	md "github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/rwset"
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
		if bytes.Equal(e.Value.(TempCC).templateId, templateId) {
			listCC.Remove(e)
		}
	}
}

func listGet(templateId []byte) (*TempCC, error) {
	for e := listCC.Front(); e != nil; e = e.Next() {
		if bytes.Equal(e.Value.(TempCC).templateId, templateId) {
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
func Init(dag dag.IDag, jury core.IAdapterJury) error {
	if err := db.SetCcDagHand(dag); err != nil {
		return err
	}
	if err := peerServerInit(jury); err != nil {
		log.Errorf("peerServerInit:%s", err)
		return err
	}
	if err := systemContractInit(); err != nil {
		log.Errorf("systemContractInit error:%s", err)
		return err
	}
	log.Info("contract manger init ok")

	return nil
}

func InitNoSysCCC(jury core.IAdapterJury) error {
	if err := peerServerInit(jury); err != nil {
		log.Errorf("peerServerInit error:%s", err)
		return err
	}
	return nil
}

func Deinit() error {
	if err := peerServerDeInit(); err != nil {
		log.Errorf("peerServerDeInit error:%s", err)
		return err
	}

	if err := systemContractDeInit(); err != nil {
		log.Errorf("systemContractDeInit error:%s", err)
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
		//ci.Enable = ccinf.Enabled
		ci.SysCC = true
		scclist = append(scclist, ci)
	}
	return scclist, count, err
}

func GetUsrCCList() {
}

//install but not into db
func Install(dag dag.IDag, chainID string, ccName string, ccPath string, ccVersion string) (payload *md.ContractTplPayload, err error) {
	log.Infof("enter ccapi.go Install")
	defer log.Infof("exit ccapi.go Install")
	log.Infof("chainID[%s]-name[%s]-path[%s]-version[%s]", chainID, ccName, ccPath, ccVersion)
	usrcc := &ucc.UserChaincode{
		Name:    ccName,
		Path:    ccPath,
		Version: ccVersion,
		Enabled: true,
	}
	var buffer bytes.Buffer
	buffer.Write([]byte(ccName))
	buffer.Write([]byte(ccPath))
	buffer.Write([]byte(ccVersion))
	tpid := cp.Keccak256Hash(buffer.Bytes())
	payloadUnit := &md.ContractTplPayload{
		TemplateId: []byte(tpid[:]),
		Name:       ccName,
		Path:       ccPath,
		Version:    ccVersion,
	}

	if cfg.DebugTest {
		log.Info("enter contract debug test")
		tcc := &TempCC{templateId: []byte(tpid[:]), name: ccName, path: ccPath, vers: ccVersion}
		listAdd(tcc)
	} else {
		//查询一下是否已经安装过
		if v, _, _, _, _ := dag.GetContractTpl(tpid[:]); v != nil {
			log.Error("getContractTpl err:", "error", "the contractTlp is exist")
			return nil, errors.New("the contractTlp is exist.")
		}
		//将合约代码文件打包成 tar 文件
		paylod, err := ucc.GetUserCCPayload(chainID, usrcc)
		if err != nil {
			log.Error("getUserCCPayload err:", "error", err)
			return nil, err
		}
		payloadUnit.ByteCode = paylod
	}
	log.Info("user contract template id", "byte:", payloadUnit.TemplateId, " string:", hex.EncodeToString(payloadUnit.TemplateId))
	//type ContractTplPayload struct {
	//	TemplateId []byte `json:"template_id"` // contract template id
	//	Name       string `json:"name"`        // contract template name
	//	Path       string `json:"path"`        // contract template execute path
	//	Version    string `json:"version"`     // contract template version
	//	Memory     uint16 `json:"memory"`      // contract template bytecode memory size(Byte), use to compute transaction fee
	//	Bytecode   []byte `json:"bytecode"`    // contract bytecode
	//}
	fmt.Println("Install result:==========================================================", payloadUnit)
	return payloadUnit, nil
}

func Deploy(rwM rwset.TxManager, idag dag.IDag, chainID string, templateId []byte, txId string, args [][]byte, timeout time.Duration) (deployId []byte, deployPayload *md.ContractDeployPayload, e error) {
	log.Info("enter Deploy", "chainID", chainID, "templateId", hex.EncodeToString(templateId), "txId", txId)
	defer log.Info("exit Deploy", "txId", txId)

	var mksupt Support = &SupportImpl{}
	setChainId := "palletone"
	setTimeOut := time.Duration(30) * time.Second
	if chainID != "" {
		setChainId = chainID
	}
	if timeout > 0 {
		setTimeOut = timeout
	}
	spec := &pb.ChaincodeSpec{
		Type: pb.ChaincodeSpec_Type(pb.ChaincodeSpec_Type_value["GOLANG"]),
		Input: &pb.ChaincodeInput{
			Args: args,
		},
		ChaincodeId: &pb.ChaincodeID{},
	}
	templateCC := &ucc.UserChaincode{}
	var err error
	var chaincodeData []byte
	//test
	if cfg.DebugTest {
		log.Info("enter contract debug test")
		tmpcc, err := listGet(templateId)
		if err == nil {
			templateCC.Name = tmpcc.name
			templateCC.Path = tmpcc.path
			templateCC.Version = tmpcc.vers
		} else {
			errMsg := fmt.Sprintf("Deploy not find tplId[%s] in list", hex.EncodeToString(templateId))
			log.Error(errMsg)
			return nil, nil, errors.New(errMsg)
		}
	} else {
		templateCC, chaincodeData, err = ucc.RecoverChainCodeFromDb(spec, chainID, templateId)
		if err != nil {
			log.Error("Deploy", "chainid:", chainID, "templateId:", templateId, "RecoverChainCodeFromDb err", err)
			return nil, nil, err
		}
	}
	txsim, err := mksupt.GetTxSimulator(rwM, idag, chainID, txId)
	if err != nil {
		log.Error("getTxSimulator err:", "error", err)
		return nil, nil, errors.WithMessage(err, "GetTxSimulator error")
	}
	btxId, err := hex.DecodeString(txId)
	depId := common.NewAddress(btxId[:20], common.ContractHash)
	usrccName := depId.String() //+ "_" + txId
	//usrccName := templateCC.Name //+ "_" + txId
	usrcc := &ucc.UserChaincode{
		Name:     usrccName,
		Path:     templateCC.Path,
		Version:  templateCC.Version,
		InitArgs: args,
		Enabled:  true,
	}
	chaincodeID := &pb.ChaincodeID{
		Name:    usrccName,
		Path:    usrcc.Path,
		Version: usrcc.Version,
	}
	spec.ChaincodeId = chaincodeID
	err = ucc.DeployUserCC(chaincodeData, spec, setChainId, usrcc, txId, txsim, setTimeOut)
	if err != nil {
		log.Error("deployUserCC err:", "error", err)
		return nil, nil, errors.WithMessage(err, "Deploy fail")
	}
	cc := &cclist.CCInfo{
		Id:      depId.Bytes21(),
		Name:    usrccName,
		Version: usrcc.Version,
		SysCC:   false,
	}
	if depId.IsSystemContractAddress() {
		err = cclist.SetChaincode(setChainId, 0, cc)
		if err != nil {
			log.Error("Deploy", "SetChaincode fail, chainId", setChainId, "name", cc.Name)
		}
	} else {
		err = saveChaincode(idag, setChainId, depId, cc)
		if err != nil {
			log.Error("Deploy saveChaincodeSet", "SetChaincode fail, channel", setChainId, "name", cc.Name, "error", err.Error())
		}
	}
	unit, err := RwTxResult2DagDeployUnit(txsim, templateId, cc.Name, cc.Id, args, timeout)
	if err != nil {
		log.Errorf("chainID[%s] converRwTxResult2DagUnit failed", chainID)
		return nil, nil, errors.WithMessage(err, "Conver RwSet to dag unit fail")
	}
	return cc.Id, unit, err
}

func getChaincode(dag dag.IDag, contractId common.Address) (*cclist.CCInfo, error) {
	log.Infof("===========enter getChaincode")
	return dag.GetChaincodes(contractId)
}

func saveChaincode(dag dag.IDag, channel string, contractId common.Address, chaincode *cclist.CCInfo) error {
	log.Infof("===========enter saveChaincodeSet")
	err := dag.SaveChaincode(contractId, chaincode)
	if err != nil {
		return err
	}
	return nil
}

//timeout:ms
// ccName can be contract Id
//func Invoke(chainID string, deployId []byte, txid string, args [][]byte, timeout time.Duration) (*peer.ContractInvokePayload, error) {
func Invoke(rwM rwset.TxManager, idag dag.IDag, chainID string, deployId []byte, txid string, args [][]byte, timeout time.Duration) (*md.ContractInvokeResult, error) {
	log.Infof("enter ccapi.go Invoke")
	defer log.Infof("exit ccapi.go Invoke")
	log.Infof("chainID[%s]-deployId[%s]-txid[%s]", chainID, hex.EncodeToString(deployId), txid)

	var mksupt Support = &SupportImpl{}
	creator := []byte("palletone")
	address := common.Address{}
	address.SetBytes(deployId)
	cc := &cclist.CCInfo{}
	var err error
	if address.IsSystemContractAddress() {
		cc, err = cclist.GetChaincode(chainID, deployId)
		if err != nil {
			return nil, err
		}
	} else {
		cc, err = getChaincode(idag, address)
		if err != nil {
			return nil, err
		}
	}
	//cc, err := cclist.GetChaincode(chainID, deployId)
	//if err != nil {
	//	return nil, err
	//}
	//if cc.Name == "" {
	//	errstr := fmt.Sprintf("chainCode[%v] not deplay!!", deployId)
	//	return nil, errors.New(errstr)
	//}

	log.Infof("Invoke [%s][%s]", chainID, address.String())
	startTm := time.Now()
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
		log.Errorf("signedEndorserProposa error[%v]", err)
		return nil, err
	}
	rsp, unit, err := es.ProcessProposal(rwM, idag, deployId, context.Background(), sprop, prop, chainID, cid, timeout)
	if err != nil {
		log.Errorf("ProcessProposal error[%v]", err)
		return nil, err
	}
	stopTm := time.Now()
	duration := stopTm.Sub(startTm)
	//unit.ExecutionTime = duration
	requstId := common.HexToHash(txid)
	unit.RequestId = requstId
	log.Infof("Invoke Ok, ProcessProposal duration=%v,rsp=%v,%s", duration, rsp, unit.Payload)
	//type ContractInvokeResult struct {
	//	ContractId   []byte             `json:"contract_id"` // contract id
	//	RequestId    common.Hash        `json:"request_id"`
	//	FunctionName string             `json:"function_name"`
	//	Args         [][]byte           `json:"args"`         // contract arguments list
	//	ReadSet      []ContractReadSet  `json:"read_set"`     // the set data of read, and value could be any type
	//	WriteSet     []ContractWriteSet `json:"write_set"`    // the set data of write, and value could be any type
	//	Payload      []byte             `json:"payload"`      // the contract execution result
	//	TokenPayOut  []*TokenPayOut     `json:"token_payout"` //从合约地址付出Token
	//	TokenSupply  []*TokenSupply     `json:"token_supply"` //增发Token请求产生的结果
	//	TokenDefine  *TokenDefine       `json:"token_define"` //定义新Token
	//}
	//TODO
	//getRT(name)
	fmt.Println("Invoke result:==========================================================", unit)
	return unit, nil
}

func Stop(rwM rwset.TxManager, idag dag.IDag, contractid []byte, chainID string, deployId []byte, txid string, deleteImage bool) (*md.ContractStopPayload, error) {
	log.Infof("enter ccapi.go Stop")
	defer log.Infof("exit ccapi.go Stop")
	log.Infof("deployId[%s]txid[%s]", hex.EncodeToString(deployId), txid)
	setChainId := "palletone"
	if chainID != "" {
		setChainId = chainID
	}
	if txid == "" {
		return nil, errors.New("input param txid is nil")
	}
	//cc, err := cclist.GetChaincode(chainID, deployId)
	//if err != nil {
	//	return nil, err
	//}
	address := common.Address{}
	address.SetBytes(deployId)
	cc, err := getChaincode(idag, address)
	if err != nil {
		return nil, err
	}
	stopResult, err := StopByName(contractid, setChainId, txid, cc.Name, cc.Path, cc.Version, deleteImage)
	if err == nil {
		cclist.DelChaincode(chainID, cc.Name, cc.Version)
	}
	//rwM.CloseTxSimulator(setChainId, txid)
	return stopResult, err
}

func DeployByName(rwM rwset.TxManager, idag dag.IDag, chainID string, txid string, ccName string, ccPath string, ccVersion string, args [][]byte, timeout time.Duration) (depllyId []byte, respPayload *md.ContractDeployPayload, e error) {
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
	txsim, err := mksupt.GetTxSimulator(rwM, idag, chainID, txid)
	if err != nil {
		return nil, nil, errors.New("GetTxSimulator error")
	}
	usrcc := &ucc.UserChaincode{
		Name:     ccName,
		Path:     ccPath,
		Version:  ccVersion,
		InitArgs: args,
		Enabled:  true,
	}
	spec := &pb.ChaincodeSpec{
		Type: pb.ChaincodeSpec_Type(pb.ChaincodeSpec_Type_value["GOLANG"]),
		Input: &pb.ChaincodeInput{
			Args: args,
		},
		ChaincodeId: &pb.ChaincodeID{
			Name:    ccName,
			Path:    ccPath,
			Version: ccVersion,
		},
	}
	err = ucc.DeployUserCC(nil, spec, setChainId, usrcc, txid, txsim, setTimeOut)
	if err != nil {
		return nil, nil, errors.New("Deploy fail")
	}
	cc := &cclist.CCInfo{
		Id:      randNum,
		Name:    ccName,
		Path:    ccPath,
		Version: ccVersion,
		SysCC:   false,
		//Enable:  true,
	}
	err = cclist.SetChaincode(setChainId, 0, cc)
	if err != nil {
		log.Errorf("setchaincode[%s]-[%s] fail", setChainId, cc.Name)
	}
	return cc.Id, nil, err
}

func StopByName(contractid []byte, chainID string, txid string, ccName string, ccPath string, ccVersion string, deleteImage bool) (*md.ContractStopPayload, error) {
	usrcc := &ucc.UserChaincode{
		Name:    ccName,
		Path:    ccPath,
		Version: ccVersion,
		Enabled: true,
	}
	err := ucc.StopUserCC(contractid, chainID, usrcc, txid, deleteImage)
	if err != nil {
		errMsg := fmt.Sprintf("StopUserCC err[%s]-[%s]-err[%s]", chainID, ccName, err)
		return nil, errors.New(errMsg)
	}
	stopResult := &md.ContractStopPayload{
		ContractId: contractid,
	}
	return stopResult, nil
}
