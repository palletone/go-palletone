package test

import (
	"bytes"
	"container/list"
	"context"
	"fmt"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/crypto"
	"github.com/palletone/go-palletone/common/log"
	list2 "github.com/palletone/go-palletone/contracts/list"
	"github.com/palletone/go-palletone/contracts/manger"
	"github.com/palletone/go-palletone/contracts/ucc"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"github.com/palletone/go-palletone/dag"
	"github.com/palletone/go-palletone/dag/errors"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/rwset"
	errors2 "github.com/pkg/errors"
	"time"
)

type TempCC struct {
	templateId  []byte
	name        string
	path        string
	version     string
	description string
	abi         string
	language    string
}

var listCC list.List

func listAdd(cc *TempCC) {
	if cc != nil {
		log.Debug("listAdd", "TempCC", cc)
		listCC.PushBack(*cc)
	}
}

//func listDel(templateId []byte) {
//	for e := listCC.Front(); e != nil; e = e.Next() {
//		if bytes.Equal(e.Value.(TempCC).templateId, templateId) {
//			listCC.Remove(e)
//		}
//	}
//}

func listGet(templateId []byte) (*TempCC, error) {
	for e := listCC.Front(); e != nil; e = e.Next() {
		if bytes.Equal(e.Value.(TempCC).templateId, templateId) {
			cc := &TempCC{
				templateId:  templateId,
				name:        e.Value.(TempCC).name,
				path:        e.Value.(TempCC).path,
				version:     e.Value.(TempCC).version,
				description: e.Value.(TempCC).description,
				abi:         e.Value.(TempCC).abi,
				language:    e.Value.(TempCC).language,
			}
			//fmt.Printf("==name[%s]", cc.name)
			return cc, nil
		}
	}
	return nil, errors.New("not find")
}

//install but not into db
func Install(chainID, ccName, ccPath, ccVersion, ccDescription, ccAbi, ccLanguage string) (payload *modules.ContractTplPayload, err error) {
	log.Info("Install enter", "chainID", chainID, "name", ccName, "path", ccPath, "version", ccVersion, "ccdescription", ccDescription, "ccabi", ccAbi, "cclanguage", ccLanguage)
	defer log.Info("Install exit", "chainID", chainID, "name", ccName, "path", ccPath, "version", ccVersion, "ccdescription", ccDescription, "ccabi", ccAbi, "cclanguage", ccLanguage)
	//  产生唯一模板id
	var buffer bytes.Buffer
	buffer.Write([]byte(ccName))
	buffer.Write([]byte(ccPath))
	buffer.Write([]byte(ccVersion))
	tpid := crypto.Keccak256Hash(buffer.Bytes())
	payloadUnit := &modules.ContractTplPayload{
		TemplateId: tpid[:],
		//Name:       ccName,
		//Path:       ccPath,
		//Version:    ccVersion,
	}
	log.Info("enter contract debug test", "templateId", tpid)
	tcc := &TempCC{templateId: tpid[:], name: ccName, path: ccPath, version: ccVersion, description: ccDescription, abi: ccAbi, language: ccLanguage}
	//  保存
	listAdd(tcc)
	return payloadUnit, nil
}
func Deploy(rwM rwset.TxManager, idag dag.IDag, chainID string, templateId []byte, txId string, args [][]byte) (deployId []byte, deployPayload *modules.ContractDeployPayload, e error) {
	log.Info("Deploy enter", "chainID", chainID, "templateId", templateId, "txId", txId)
	defer log.Info("Deploy exit", "chainID", chainID, "templateId", templateId, "txId", txId)
	var mksupt manger.Support = &manger.SupportImpl{}
	setTimeOut := time.Duration(30) * time.Second
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

	log.Info("enter contract debug test")
	tmpcc, err := listGet(templateId)
	if err == nil {
		templateCC.Name = tmpcc.name
		templateCC.Path = tmpcc.path
		templateCC.Version = tmpcc.version
		templateCC.Language = tmpcc.language
		//templateCC.Desciption = tmpcc.description
		templateCC.Language = tmpcc.language
		//templateCC.Abi = tmpcc.abi
	} else {
		errMsg := fmt.Sprintf("Deploy not find tplId[%x] in list", templateId)
		log.Error(errMsg)
		return nil, nil, errors.New(errMsg)
	}
	txsim, err := mksupt.GetTxSimulator(rwM, idag, chainID, txId)
	if err != nil {
		log.Error("getTxSimulator err:", "error", err)
		return nil, nil, errors2.WithMessage(err, "GetTxSimulator error")
	}
	txHash := common.HexToHash(txId)
	depId := crypto.RequestIdToContractAddress(txHash) //common.NewAddress(btxId[:20], common.ContractHash)
	//usrccName := depId.String()
	usrcc := &ucc.UserChaincode{
		Name:    templateCC.Name,
		Path:    templateCC.Path,
		Version: templateCC.Version,
		//Desciption: templateCC.Desciption,
		Language: templateCC.Language,
		//Abi:        templateCC.Abi,
		//InitArgs:   args,
		Enabled: true,
	}
	chaincodeID := &pb.ChaincodeID{
		Name:    usrcc.Name,
		Path:    usrcc.Path,
		Version: usrcc.Version,
	}
	spec.ChaincodeId = chaincodeID
	cp := idag.GetChainParameters()
	spec.CpuQuota = cp.UccCpuQuota  //微妙单位（100ms=100000us=上限为1个CPU）
	spec.CpuShare = cp.UccCpuShares //占用率，默认1024，即可占用一个CPU，相对值
	spec.Memory = cp.UccMemory      //字节单位 物理内存  1073741824  1G 2147483648 2G 209715200 200m 104857600 100m
	err = ucc.DeployUserCC(depId.Bytes(), chaincodeData, spec, chainID, txId, txsim, setTimeOut)
	if err != nil {
		log.Error("deployUserCC err:", "error", err)
		return nil, nil, errors2.WithMessage(err, "Deploy fail")
	}
	cc := &list2.CCInfo{
		Id:      depId.Bytes(),
		Name:    usrcc.Name,
		Path:    usrcc.Path,
		Version: usrcc.Version,
		//Description: usrcc.Desciption,
		//Abi:         usrcc.Abi,
		Language: usrcc.Language,
		TempleId: templateId,
		SysCC:    false,
	}
	//  测试
	err = list2.SetChaincode(chainID, 0, cc)
	if err != nil {
		log.Error("Deploy", "SetChaincode fail, chainId", chainID, "name", cc.Name)
	}

	unit, err := manger.RwTxResult2DagDeployUnit(txsim, templateId, cc.Name, cc.Id, args, setTimeOut)
	if err != nil {
		log.Errorf("chainID[%s] converRwTxResult2DagUnit failed", chainID)
		return nil, nil, errors2.WithMessage(err, "Conver RwSet to dag unit fail")
	}
	return cc.Id, unit, err
}
func Invoke(rwM rwset.TxManager, idag dag.IDag, chainID string, deployId []byte, txid string, args [][]byte) (*modules.ContractInvokeResult, error) {
	log.Info("Invoke enter", "chainID", chainID, "deployId", deployId, "txid", txid)
	defer log.Info("Invoke exit", "chainID", chainID, "deployId", deployId, "txid", txid)
	setTimeOut := time.Duration(30) * time.Second

	var mksupt manger.Support = &manger.SupportImpl{}
	creator := []byte(chainID)
	chain := list2.GetAllChaincode(chainID)
	if chain != nil {
		for k, v := range chain.CClist {
			log.Infof("\n\nchaincode name =======%v", k)
			log.Infof("\n\nchaincode info =======%v", v)
		}
	}
	cc, err := list2.GetChaincode(chainID, deployId, "")
	if err != nil {
		return nil, err
	}
	startTm := time.Now()
	es := manger.NewEndorserServer(mksupt)
	spec := &pb.ChaincodeSpec{
		ChaincodeId: &pb.ChaincodeID{Name: cc.Name},
		Type:        pb.ChaincodeSpec_GOLANG,
		Input:       &pb.ChaincodeInput{Args: args},
	}
	cid := &pb.ChaincodeID{
		Path:    cc.Path, //no use
		Name:    cc.Name,
		Version: cc.Version,
	}
	sprop, prop, err := manger.SignedEndorserProposa(chainID, txid, spec, creator, []byte("msg1"))
	if err != nil {
		log.Errorf("signedEndorserProposa error[%v]", err)
		return nil, err
	}
	rsp, unit, err := es.ProcessProposal(rwM, idag, deployId, context.Background(), sprop, prop, chainID, cid, setTimeOut)
	if err != nil {
		log.Infof("ProcessProposal error[%v]", err)
		return nil, err
	}
	stopTm := time.Now()
	duration := stopTm.Sub(startTm)
	//unit.ExecutionTime = duration
	requstId := common.HexToHash(txid)
	unit.RequestId = requstId
	log.Infof("Invoke Ok, ProcessProposal duration=%v,rsp=%v,%s", duration, rsp, unit.Payload)
	return unit, nil
}

func Stop(contractid []byte, chainID string, deployId []byte, txid string, deleteImage bool) (*modules.ContractStopPayload, error) {
	log.Info("Stop enter", "contractid", contractid, "chainID", chainID, "deployId", deployId, "txid", txid)
	defer log.Info("Stop enter", "contractid", contractid, "chainID", chainID, "deployId", deployId, "txid", txid)

	setChainId := "palletone"
	if chainID != "" {
		setChainId = chainID
	}
	if txid == "" {
		return nil, errors.New("input param txid is nil")
	}
	cc, err := list2.GetChaincode(chainID, deployId, "") //todo
	if err != nil {
		return nil, err
	}
	result, err := StopByName(contractid, setChainId, txid, cc.Name, cc.Path, cc.Version, deleteImage)
	if err == nil {
		list2.DelChaincode(chainID, cc.Name, cc.Version)
	}
	return result, err
}

func StopByName(contractid []byte, chainID string, txid string, ccName string, ccPath string, ccVersion string, deleteImage bool) (*modules.ContractStopPayload, error) {
	usrcc := &ucc.UserChaincode{
		Name:    ccName,
		Path:    ccPath,
		Version: ccVersion,
		Enabled: true,
	}
	err := ucc.StopUserCC(contractid, chainID, usrcc, txid, deleteImage, false)
	if err != nil {
		errMsg := fmt.Sprintf("StopUserCC err[%s]-[%s]-err[%s]", chainID, ccName, err)
		return nil, errors.New(errMsg)
	}
	stopResult := &modules.ContractStopPayload{
		ContractId: contractid,
	}
	return stopResult, nil
}
