package manger

import (
	"time"
	"errors"
	"fmt"
	"golang.org/x/net/context"

	"github.com/palletone/go-palletone/contracts/scc"
	"github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"github.com/palletone/go-palletone/core/vmContractPub/crypto"
	"github.com/palletone/go-palletone/contracts/ucc"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
)

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

func GetSysCCList() (ccInf []CCInfo, ccCount int, errs error) {
	scclist := make([]CCInfo, 0)
	ci := CCInfo{}

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

//timeout:ms
// ccName can be contract Id
func Invoke(chainID string, ccName string, txid string, args [][]byte, timeout time.Duration) (*peer.ContractInvokePayload, error) {
	//func Invoke(chainID string, ccName string, txid string, args [][]byte, timeout time.Duration) (*peer.ContractInvokePayload, error) {
	var mksupt Support = &SupportImpl{}
	creator := []byte("palletone") //default
	ccVersion := "ptn001"          //default

	logger.Infof("===== Invoke [%s][%s]======", chainID, ccName)
	es := NewEndorserServer(mksupt)
	spec := &pb.ChaincodeSpec{
		ChaincodeId: &pb.ChaincodeID{Name: ccName},
		Type:        pb.ChaincodeSpec_GOLANG,
		Input:       &pb.ChaincodeInput{Args: args},
	}

	cid := &pb.ChaincodeID{
		Path:    "", //no use
		Name:    ccName,
		Version: ccVersion,
	}

	sprop, prop, err := signedEndorserProposa(chainID, txid, spec, creator, []byte("msg1"))
	if err != nil {
		logger.Errorf("signedEndorserProposa error[%v]", err)
		return nil, err
	}

	rsp, unit, err := es.ProcessProposal(context.Background(), sprop, prop, chainID, cid, timeout)
	if err != nil {
		logger.Errorf("ProcessProposal error[%v]", err)
		return nil, err
	}
	logger.Infof("Invoke Ok, ProcessProposal rsp=%v", rsp)

	return unit, nil
}

func Deploy(chainID string, txid string, ccName string, ccPath string, ccVersion string, args [][]byte, timeout time.Duration) (depllyId string, respPayload *peer.ContractDeployPayload, e error) {
	setChainId := "palletone"
	setTimeOut := time.Duration(30) * time.Second

	if chainID != "" {
		setChainId = chainID
	}
	if timeout > 0 {
		setTimeOut = timeout
	}
	if txid == "" || ccName == "" || ccPath == "" {
		return "", nil, errors.New("input param is nil")
	}
	randNum, err := crypto.GetRandomNonce()
	if err != nil {
		return "", nil, errors.New("crypto.GetRandomNonce error")
	}

	usrcc := &ucc.UserChaincode{
		Name:     ccName,
		Path:     ccPath,
		Version:  ccVersion,
		InitArgs: args,
		Enabled:  true,
	}

	err = ucc.DeployUserCC(setChainId, usrcc, txid, setTimeOut)
	if err != nil {
		return "", nil, errors.New("Deploy fail")
	}

	cc := &CCInfo{
		Id:      string(randNum),
		Name:    ccName,
		Path:    ccPath,
		Version: ccVersion,
		SysCC:   false,
		Enable:  true,
	}
	err = setChaincode(setChainId, 0, cc)
	if err != nil {
		logger.Errorf("setchaincode[%s]-[%s] fail", setChainId, cc.Name)
	}

	return cc.Id, nil, err
}

func Stop(chainID string, txid string, ccName string, ccPath string, ccVersion string, deleteImage bool) error {
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

func StopById(chainID string, txid string, deployId string, deleteImage bool) error {
	setChainId := "palletone"

	if chainID != "" {
		setChainId = chainID
	}
	if txid == "" {
		return errors.New("input param txid is nil")
	}

	clist, err := getChaincodeList(chainID)
	if err != nil {
		logger.Errorf("not find chainlist for chainId[%s]", chainID)
		return errors.New("getChaincodeList failed")
	}

	for k, v := range clist.cclist {
		logger.Infof("chaincode[%s]:%v", k, *v)
		if k == chainID {
			if v.Id == deployId {
				return Stop(setChainId, txid, v.Name, v.Path, v.Version, deleteImage)
			}
		}
	}

	return errors.New("not find deployId")
}
