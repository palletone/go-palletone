package manger

import (
	"golang.org/x/net/context"
	"github.com/palletone/go-palletone/contracts/scc"
	"github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"time"
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

func Deinit() error{
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
func Invoke(chainID string, ccName string,  args [][]byte, timeout time.Duration) (*peer.ContractInvokePayload, error){
	var mksupt Support = &SupportImpl{}
	creator := []byte("palletone")  //default
	ccVersion := "ptn001"  //default

	logger.Infof("===== Invoke [%s][%s]======", chainID, ccName)
	es := NewEndorserServer(mksupt)
	spec := &pb.ChaincodeSpec{
		ChaincodeId: &pb.ChaincodeID{Name: ccName},
		Type:        pb.ChaincodeSpec_GOLANG,
		Input:       &pb.ChaincodeInput{Args: args},
	}

	cid := &pb.ChaincodeID {
		Path:     "", //no use
		Name:     ccName,
		Version:  ccVersion,
	}

	sprop, prop, err := signedEndorserProposa(chainID, spec, creator, []byte("msg1"))
	if err != nil {
		logger.Errorf("signedEndorserProposa error[%v]", err)
		return nil, err
	}
	if timeout != 0 {
		timeoutProcess := func () {
			logger.Infof("timeoutProcess")
		}
		time.AfterFunc(timeout, timeoutProcess)
	}
	rsp, unit, err := es.ProcessProposal(context.Background(), sprop, prop, chainID, cid, timeout)
	if err != nil {
		logger.Errorf("ProcessProposal error[%v]", err)
		return nil, err
	}
	logger.Infof("ProcessProposal rsp=%v", rsp)

	return unit, nil
}

func Deploy() (contractId string, err error){

	return "", nil
}

func Stop(contractId string) error {

	return nil
}
