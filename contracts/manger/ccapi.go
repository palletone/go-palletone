package manger

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
	"time"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/crypto"
	"github.com/palletone/go-palletone/common/log"
	db "github.com/palletone/go-palletone/contracts/comm"
	"github.com/palletone/go-palletone/contracts/core"
	cclist "github.com/palletone/go-palletone/contracts/list"
	"github.com/palletone/go-palletone/contracts/scc"
	"github.com/palletone/go-palletone/contracts/ucc"

	"github.com/fsouza/go-dockerclient"
	"github.com/palletone/go-palletone/common/util"
	"github.com/palletone/go-palletone/contracts/contractcfg"
	"github.com/palletone/go-palletone/contracts/utils"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"github.com/palletone/go-palletone/dag"
	md "github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/rwset"
)

//type TempCC struct {
//	templateId  []byte
//	name        string
//	path        string
//	version     string
//	description string
//	abi         string
//	language    string
//}

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
		//ci.SysCC = true
		scclist = append(scclist, ci)
	}
	return scclist, count, err
}

//install but not into db
func Install(dag dag.IDag, chainID, ccName, ccPath, ccVersion, ccDescription, ccAbi, ccLanguage string) (payload *md.ContractTplPayload, err error) {
	log.Info("Install enter", "chainID", chainID, "name", ccName, "path", ccPath, "version", ccVersion, "ccdescription", ccDescription, "ccabi", ccAbi, "cclanguage", ccLanguage)
	defer log.Info("Install exit", "chainID", chainID, "name", ccName, "path", ccPath, "version", ccVersion, "ccdescription", ccDescription, "ccabi", ccAbi, "cclanguage", ccLanguage)
	//  用于合约实列
	usrcc := &ucc.UserChaincode{
		Name:    ccName,
		Path:    ccPath,
		Version: ccVersion,
		//Desciption: ccDescription,
		//Abi:        ccAbi,
		Language: ccLanguage,
		Enabled:  true,
	}
	//  产生唯一模板id
	var buffer bytes.Buffer
	buffer.Write([]byte(ccName))
	buffer.Write([]byte(ccPath))
	buffer.Write([]byte(ccVersion))
	tpid := crypto.Keccak256Hash(buffer.Bytes())
	payloadUnit := &md.ContractTplPayload{
		TemplateId: tpid[:],
		//Name:       ccName,
		//Path:       ccPath,
		//Version:    ccVersion,
	}

	//查询一下是否已经安装过
	if tpl, _ := dag.GetContractTpl(tpid[:]); tpl != nil {
		errMsg := fmt.Sprintf("install ,the contractTlp is exist.tplId:%x", tpid)
		log.Debug("Install", "err", errMsg)
		return nil, errors.New(errMsg)
	}
	//将合约代码文件打包成 tar 文件
	paylod, err := ucc.GetUserCCPayload(usrcc)
	if err != nil {
		log.Error("getUserCCPayload err:", "error", err)
		return nil, err
	}
	payloadUnit.ByteCode = paylod

	return payloadUnit, nil
}

func Deploy(rwM rwset.TxManager, idag dag.IDag, chainID string, templateId []byte, txId string, args [][]byte, timeout time.Duration) (deployId []byte, deployPayload *md.ContractDeployPayload, e error) {
	log.Info("Deploy enter", "chainID", chainID, "templateId", templateId, "txId", txId)
	defer log.Info("Deploy exit", "chainID", chainID, "templateId", templateId, "txId", txId)
	setTimeOut := time.Duration(30) * time.Second
	if timeout > 0 {
		setTimeOut = timeout
	}
	templateCC, chaincodeData, err := ucc.RecoverChainCodeFromDb(chainID, templateId)
	if err != nil {
		log.Error("Deploy", "chainid:", chainID, "templateId:", templateId, "RecoverChainCodeFromDb err", err)
		return nil, nil, err
	}
	mksupt := &SupportImpl{}
	txsim, err := mksupt.GetTxSimulator(rwM, idag, chainID, txId)
	if err != nil {
		log.Error("getTxSimulator err:", "error", err)
		return nil, nil, errors.WithMessage(err, "GetTxSimulator error")
	}
	txHash := common.HexToHash(txId)
	depId := crypto.RequestIdToContractAddress(txHash) //common.NewAddress(btxId[:20], common.ContractHash)
	usrccName := depId.String()
	usrcc := &ucc.UserChaincode{
		Name:    usrccName,
		Path:    templateCC.Path,
		Version: templateCC.Version,
		//Desciption: templateCC.Desciption,
		Language: templateCC.Language,
		//Abi:        templateCC.Abi,
		//InitArgs:   args,
		Enabled: true,
	}
	//  TODO 可以开启单机多容器,防止容器名冲突
	usrcc.Version += ":"
	usrcc.Version += contractcfg.GetConfig().ContractAddress
	spec := &pb.ChaincodeSpec{
		Type: pb.ChaincodeSpec_Type(pb.ChaincodeSpec_Type_value[templateCC.Language]),
		Input: &pb.ChaincodeInput{
			Args: args,
		},
		ChaincodeId: &pb.ChaincodeID{
			Name:    usrcc.Name,
			Path:    usrcc.Path,
			Version: usrcc.Version,
		},
	}
	//TODO 这里获取运行用户合约容器的相关资源  CpuQuota  CpuShare  MEMORY
	cp := idag.GetChainParameters()
	spec.CpuQuota = cp.UccCpuQuota  //微妙单位（100ms=100000us=上限为1个CPU）
	spec.CpuShare = cp.UccCpuShares //占用率，默认1024，即可占用一个CPU，相对值
	spec.Memory = cp.UccMemory      //字节单位 物理内存  1073741824  1G 2147483648 2G 209715200 200m 104857600 100m
	err = ucc.DeployUserCC(depId.Bytes(), chaincodeData, spec, chainID, txId, txsim, setTimeOut)
	if err != nil {
		log.Error("deployUserCC err:", "error", err)
		return nil, nil, errors.WithMessage(err, "Deploy fail")
	}
	cc := &cclist.CCInfo{
		Id:       depId.Bytes(),
		Name:     usrcc.Name,
		Path:     usrcc.Path,
		TempleId: templateId,
		Version:  usrcc.Version,
		Language: usrcc.Language,
		SysCC:    false,
	}
	//if depId.IsSystemContractAddress() {
	//	cc.SysCC = true
	//	err = cclist.SetChaincode(chainID, 0, cc)
	//	if err != nil {
	//		log.Error("Deploy", "SetChaincode fail, chainId", chainID, "name", cc.Name)
	//	}
	//} else {
	err = SaveChaincode(idag, depId, cc)
	if err != nil {
		log.Error("Deploy saveChaincodeSet", "SetChaincode fail, channel", chainID, "name", cc.Name, "error", err.Error())
	}
	//}
	unit, err := RwTxResult2DagDeployUnit(txsim, templateId, cc.Name, cc.Id, args, timeout)
	if err != nil {
		log.Errorf("chainID[%s] converRwTxResult2DagUnit failed", chainID)
		return nil, nil, errors.WithMessage(err, "Conver RwSet to dag unit fail")
	}
	return cc.Id, unit, err
}

func GetChaincode(dag dag.IDag, contractId common.Address) (*cclist.CCInfo, error) {
	return dag.GetChaincode(contractId)
}

func SaveChaincode(dag dag.IDag, contractId common.Address, chaincode *cclist.CCInfo) error {
	return dag.SaveChaincode(contractId, chaincode)
}

func GetChaincodes(dag dag.IDag) ([]*cclist.CCInfo, error) {
	return dag.RetrieveChaincodes()
}

//timeout:ms
// ccName can be contract Id
//func Invoke(chainID string, deployId []byte, txid string, args [][]byte, timeout time.Duration) (*peer.ContractInvokePayload, error) {
func Invoke(rwM rwset.TxManager, idag dag.IDag, chainID string, deployId []byte, txid string, args [][]byte, timeout time.Duration) (*md.ContractInvokeResult, error) {
	log.Debugf("Invoke enter")
	log.Info("Invoke enter", "chainID", chainID, "deployId", deployId, "txid", txid, "timeout", timeout)
	defer log.Info("Invoke exit", "chainID", chainID, "deployId", deployId, "txid", txid, "timeout", timeout)

	var mksupt Support = &SupportImpl{}
	creator := []byte("palletone")
	address := common.NewAddress(deployId, common.ContractHash)
	var cc *cclist.CCInfo
	var err error

	if address.IsSystemContractAddress() {
		ver := getContractSysVersion(deployId, idag.GetChainParameters().ContractSystemVersion)
		cc, err = cclist.GetChaincode(chainID, deployId, ver)
		if err != nil {
			return nil, err
		}
	} else {
		cc, err = GetChaincode(idag, address)
		log.Debugf("get chain code")
		if err != nil {
			return nil, err
		}
	}
	startTm := time.Now()
	es := NewEndorserServer(mksupt)
	log.Debugf("new endorser server")
	spec := &pb.ChaincodeSpec{
		ChaincodeId: &pb.ChaincodeID{Name: cc.Name, Version: cc.Version},
		Type:        pb.ChaincodeSpec_Type(pb.ChaincodeSpec_Type_value[cc.Language]),
		Input:       &pb.ChaincodeInput{Args: args},
	}
	cid := &pb.ChaincodeID{
		Path:    cc.Path,
		Name:    cc.Name,
		Version: cc.Version,
	}
	sprop, prop, err := SignedEndorserProposa(chainID, txid, spec, creator, []byte("msg1"))
	log.Debugf("signed endorser proposal")
	if err != nil {
		log.Errorf("signedEndorserProposa error[%v]", err)
		return nil, err
	}
	rsp, unit, err := es.ProcessProposal(rwM, idag, deployId, context.Background(), sprop, prop, chainID, cid, timeout)
	log.Debugf("process proposal")
	if err != nil {
		log.Infof("ProcessProposal error[%v]", err)
		return nil, err
	}
	//
	if !cc.SysCC {
		sizeRW, disk, isOver := utils.RemoveConWhenOverDisk(cc, idag)
		if isOver {
			log.Debugf("utils.KillAndRmWhenOver name = %s,sizeRW = %d,disk = %d", cc.Name, sizeRW, disk)
			return nil, fmt.Errorf("utils.KillAndRmWhenOver name = %s,sizeRW = %d bytes,disk = %d bytes", cc.Name, sizeRW, disk)
		}
	}
	stopTm := time.Now()
	duration := stopTm.Sub(startTm)
	//unit.ExecutionTime = duration
	requstId := common.HexToHash(txid)
	unit.RequestId = requstId
	log.Infof("Invoke Ok, ProcessProposal duration=%v,rsp=%v,%s", duration, rsp, unit.Payload)
	return unit, nil
}

func Stop(rwM rwset.TxManager, idag dag.IDag, contractid []byte, chainID string, deployId []byte, txid string, deleteImage bool, dontRmCon bool) (*md.ContractStopPayload, error) {
	log.Info("Stop enter", "contractid", contractid, "chainID", chainID, "deployId", deployId, "txid", txid)
	defer log.Info("Stop enter", "contractid", contractid, "chainID", chainID, "deployId", deployId, "txid", txid)

	setChainId := "palletone"
	if chainID != "" {
		setChainId = chainID
	}
	if txid == "" {
		return nil, errors.New("input param txid is nil")
	}
	address := common.NewAddress(deployId, common.ContractHash)
	cc, err := GetChaincode(idag, address)
	if err != nil {
		return nil, err
	}
	stopResult, err := StopByName(contractid, setChainId, txid, cc, deleteImage, dontRmCon)
	if err != nil {
		return nil, err
	}
	if !dontRmCon {
		err := SaveChaincode(idag, address, nil)
		if err != nil {
			return nil, err
		}
	}
	return stopResult, err
}

func StopByName(contractid []byte, chainID string, txid string, usercc *cclist.CCInfo, deleteImage bool, dontRmCon bool) (*md.ContractStopPayload, error) {
	usrcc := &ucc.UserChaincode{
		Name:     usercc.Name,
		Path:     usercc.Path,
		Version:  usercc.Version,
		Enabled:  true,
		Language: usercc.Language,
	}
	err := ucc.StopUserCC(contractid, chainID, usrcc, txid, deleteImage, dontRmCon)
	if err != nil {
		errMsg := fmt.Sprintf("StopUserCC err[%s]-[%s]-err[%s]", chainID, usrcc.Name, err)
		return nil, errors.New(errMsg)
	}
	stopResult := &md.ContractStopPayload{
		ContractId: contractid,
	}
	return stopResult, nil
}

func RestartContainers(client *docker.Client, dag dag.IDag, cons []docker.APIContainers) {
	//  获取所有退出容器
	addrs, err := utils.GetAllExitedContainer(cons)
	if err != nil {
		log.Infof("client.ListContainers err: %s\n", err.Error())
		return
	}
	if len(addrs) > 0 {
		for _, v := range addrs {
			rd, _ := crypto.GetRandomBytes(32)
			txid := util.RlpHash(rd)
			log.Infof("==============需要重启====容器地址为--->%s", hex.EncodeToString(v.Bytes21()))
			_, err = RestartContainer(dag, "palletone", v.Bytes21(), txid.String())
			if err != nil {
				log.Infof("RestartContainer err: %s", err.Error())
				return
			}
		}
	}
}

//删除所有过期容器
func RemoveExpiredConatiners(client *docker.Client, dag dag.IDag, rmExpConFromSysParam bool, con []docker.APIContainers) {
	//获取容器id，以及对应用户合约的地址，更新状态
	idStrMap := utils.RetrieveExpiredContainers(dag, con, rmExpConFromSysParam)
	if len(idStrMap) > 0 {
		for id, str := range idStrMap {
			err := client.RemoveContainer(docker.RemoveContainerOptions{ID: id, Force: true})
			if err != nil {
				log.Errorf("client.RemoveContainer id=%s error=%s", id, err.Error())
			}
			cc, err := GetChaincode(dag, str)
			if err != nil {
				log.Error("get chaincode error %s", err.Error())
			}
			cc.IsExpired = true
			err = SaveChaincode(dag, str, cc)
			if err != nil {
				log.Error("save chaincode error %s", err.Error())
			}
		}
	}
}

func RestartContainer(idag dag.IDag, chainID string, deployId []byte, txId string) ([]byte, error) {
	_, err := Stop(nil, idag, deployId, chainID, deployId, txId, false, true)
	if err != nil {
		return nil, err
	}
	log.Info("enter Deploy", "chainID", chainID, "templateId", hex.EncodeToString(deployId), "txId", txId)
	defer log.Info("exit Deploy", "txId", txId)
	//setChainId := "palletone"
	setTimeOut := time.Duration(50) * time.Second
	//if chainID != "" {
	//	setChainId = chainID
	//}
	//test
	address := common.NewAddress(deployId, common.ContractHash)
	cc, err := GetChaincode(idag, address)
	if err != nil {
		return nil, err
	}
	usrcc := &ucc.UserChaincode{
		Name:    cc.Name,
		Path:    cc.Path,
		Version: cc.Version,
		//InitArgs: [][]byte{},
		Enabled: true,
	}
	spec := &pb.ChaincodeSpec{
		Type: pb.ChaincodeSpec_Type(pb.ChaincodeSpec_Type_value[cc.Language]),
		Input: &pb.ChaincodeInput{
			Args: [][]byte{},
		},
		ChaincodeId: &pb.ChaincodeID{
			Name:    usrcc.Name,
			Path:    usrcc.Path,
			Version: usrcc.Version,
		},
	}
	cp := idag.GetChainParameters()
	spec.CpuQuota = cp.UccCpuQuota  //微妙单位（100ms=100000us=上限为1个CPU）
	spec.CpuShare = cp.UccCpuShares //占用率，默认1024，即可占用一个CPU，相对值
	spec.Memory = cp.UccMemory      //字节单位 物理内存  1073741824  1G 2147483648 2G 209715200 200m 104857600 100m
	_, chaincodeData, err := ucc.RecoverChainCodeFromDb(chainID, cc.TempleId)
	if err != nil {
		log.Error("Deploy", "chainid:", chainID, "templateId:", cc.TempleId, "RecoverChainCodeFromDb err", err)
		return nil, err
	}
	err = ucc.DeployUserCC(address.Bytes(), chaincodeData, spec, chainID, txId, nil, setTimeOut)
	if err != nil {
		log.Error("deployUserCC err:", "error", err)
		return nil, errors.WithMessage(err, "Deploy fail")
	}
	return cc.Id, err
}

//func StartChaincodeContainer(idag dag.IDag, chainID string, deployId []byte, txId string) ([]byte, error) {
//	//GoStart()
//	return nil, nil
//}

//func DeployByName(rwM rwset.TxManager, idag dag.IDag, chainID string, txid string, ccName string, ccPath string, ccVersion string, args [][]byte, timeout time.Duration) (depllyId []byte, respPayload *md.ContractDeployPayload, e error) {
//	var mksupt Support = &SupportImpl{}
//	setChainId := "palletone"
//	setTimeOut := time.Duration(30) * time.Second
//	if chainID != "" {
//		setChainId = chainID
//	}
//	if timeout > 0 {
//		setTimeOut = timeout
//	}
//	if txid == "" || ccName == "" || ccPath == "" {
//		return nil, nil, errors.New("input param is nil")
//	}
//	randNum, err := crypto.GetRandomNonce()
//	if err != nil {
//		return nil, nil, errors.New("crypto.GetRandomNonce error")
//	}
//	txsim, err := mksupt.GetTxSimulator(rwM, idag, chainID, txid)
//	if err != nil {
//		return nil, nil, errors.New("GetTxSimulator error")
//	}
//	usrcc := &ucc.UserChaincode{
//		Name:     ccName,
//		Path:     ccPath,
//		Version:  ccVersion,
//		InitArgs: args,
//		Enabled:  true,
//	}
//	spec := &pb.ChaincodeSpec{
//		Type: pb.ChaincodeSpec_Type(pb.ChaincodeSpec_Type_value["GOLANG"]),
//		Input: &pb.ChaincodeInput{
//			Args: args,
//		},
//		ChaincodeId: &pb.ChaincodeID{
//			Name:    ccName,
//			Path:    ccPath,
//			Version: ccVersion,
//		},
//	}
//	err = ucc.DeployUserCC(nil, spec, setChainId, usrcc, txid, txsim, setTimeOut)
//	if err != nil {
//		return nil, nil, errors.New("Deploy fail")
//	}
//	cc := &cclist.CCInfo{
//		Id:      randNum,
//		Name:    ccName,
//		Path:    ccPath,
//		Version: ccVersion,
//		SysCC:   false,
//		//Enable:  true,
//	}
//	err = cclist.SetChaincode(setChainId, 0, cc)
//	if err != nil {
//		log.Errorf("setchaincode[%s]-[%s] fail", setChainId, cc.Name)
//	}
//	return cc.Id, nil, err
//}
