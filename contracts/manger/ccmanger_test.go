package manger

import (
	"testing"
	"time"
	"os"
	"net"
	"fmt"
	"google.golang.org/grpc"
	"golang.org/x/net/context"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"github.com/palletone/go-palletone/contracts/rwset"
	"github.com/palletone/go-palletone/core/vmContractPub/protos/common"
	"github.com/palletone/go-palletone/core/vmContractPub/protos/utils"
	"github.com/palletone/go-palletone/contracts/core"
	"github.com/palletone/go-palletone/contracts/scc"
	"github.com/palletone/go-palletone/core/vmContractPub/util"
	"github.com/palletone/go-palletone/contracts/example/go/samplesyscc"
	"github.com/palletone/go-palletone/contracts/accesscontrol"
	"github.com/palletone/go-palletone/core/vmContractPub/crypto"
	"github.com/palletone/go-palletone/contracts/ucc"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"sync"
)

type mocksupt struct {}

func (*mocksupt) GetTxSimulator(chainid string, txid string) (*rwset.TxSimulator, error) {
	return nil, nil
}
func (*mocksupt) IsSysCC(name string) bool {
	return true
}

func (*mocksupt) Execute(ctxt context.Context, cid, name, version, txid string, syscc bool, signedProp *pb.SignedProposal, prop *pb.Proposal, spec interface{}, timeout time.Duration) (*pb.Response, *pb.ChaincodeEvent, error) {
	return nil, nil, nil
}

//
//func singedPro(chid, ccid, ccver string, ccargs [][]byte) *pb.SignedProposal {
//	spec := &pb.ChaincodeSpec{Type: 1, ChaincodeId: &pb.ChaincodeID{Name: ccid, Version: ccver}, Input: &pb.ChaincodeInput{Args: ccargs}}
//
//	cis := &pb.ChaincodeInvocationSpec{ChaincodeSpec: spec}
//
//	creator, err := signer.Serialize()
//	prop, _, err := utils.CreateChaincodeProposal(common.HeaderType_ENDORSER_TRANSACTION, chid, cis, creator)
//	propBytes, err := utils.GetBytesProposal(prop)
//	signature, err := signer.Sign(propBytes)
//
//	return &pb.SignedProposal{ProposalBytes: propBytes, Signature: signature}
//
//
//	sprop, prop := putils.MockSignedEndorserProposalOrPanic(chainID, spec, creator, []byte("msg1"))
//	cccid := ccprovider.NewCCContext(chainID, cdInvocationSpec.ChaincodeSpec.ChaincodeId.Name, version, uuid, false, sprop, prop)
//	retval, ccevt, err = ExecuteWithErrorFilter(ctx, cccid, cdInvocationSpec)
//	if err != nil {
//		return nil, uuid, nil, fmt.Errorf("Error invoking chaincode: %s", err)
//	}
//}
//

func getSignedPropWithCHIdAndArgs(chid, ccid, ccver string, ccargs [][]byte, t *testing.T) *pb.SignedProposal {
	spec := &pb.ChaincodeSpec{Type: 1, ChaincodeId: &pb.ChaincodeID{Name: ccid, Version: ccver}, Input: &pb.ChaincodeInput{Args: ccargs}}
	cis := &pb.ChaincodeInvocationSpec{ChaincodeSpec: spec}

	//creator, err := signer.Serialize()
	creator := []byte("glh")
	prop, _, err := utils.CreateChaincodeProposal(common.HeaderType_ENDORSER_TRANSACTION, chid, cis, creator)
	assert.NoError(t, err)
	propBytes, err := utils.GetBytesProposal(prop)
	assert.NoError(t, err)

	//todo ,tmp!!!!!!
	signature := propBytes
	//signature, err := signer.Sign(propBytes)
	assert.NoError(t, err)
	return &pb.SignedProposal{ProposalBytes: propBytes, Signature: signature}
}

func TestEndorserDeployExecSysCC(t *testing.T) {
	SysCCMap := make(map[string]struct{})
	deployedCCName := "sample_syscc"
	SysCCMap[deployedCCName] = struct{}{}
	creator := []byte("glh")
	txid := "c089md9jdopdf32"
	var mksupt Support = &SupportImpl{}

	peerInit()
	t.Logf("TestEndorserDeployExecSysCC run, cc name[%s]", deployedCCName)

	chainID := util.GetTestChainID()
	es := NewEndorserServer(mksupt)

	f := "putval"
	args := util.ToChaincodeArgs(f, "greeting", "hey there")

	//signedProp := getSignedPropWithCHIdAndArgs(util.GetTestChainID(), "lscc", "0", [][]byte{[]byte("deploy"), []byte("a"), cds}, t)
	spec := &pb.ChaincodeSpec{
		ChaincodeId: &pb.ChaincodeID{Name: deployedCCName},
		Type:        pb.ChaincodeSpec_GOLANG,
		Input:       &pb.ChaincodeInput{Args: args},
	}
	cid := &pb.ChaincodeID{
		Path: "/home/glh/project/pallet/src/common/mocks/samplesyscc/samplesyscc", ///home/glh/project/pallet/src/common/mocks/samplesyscc
		Name: "sample_syscc",
		Version:"ptn001",
	}

	sprop, prop, err := signedEndorserProposa(chainID, txid, spec, creator, []byte("msg1"))
	rsp, unit, err := es.ProcessProposal(context.Background(), sprop, prop, chainID, cid, 5*time.Second)
	if err != nil {
		logger.Errorf("ProcessProposal error[%v]", err)
	}
	logger.Infof("ProcessProposal rsp=%v, unit=%v", rsp, unit)
}

func peerMockInitialize() {
//ledgermgmt.InitializeTestEnvWithCustomProcessors(ConfigTxProcessors)
chains.clist = nil
chains.clist = make(map[string]*chain)
//chainInitializer = func(string) { return }
}
func peerMockCreateChain(cid string) error {
	chains.Lock()
	defer chains.Unlock()

	chains.clist[cid] = &chain{
	}
	return nil
}

func peerInitSysCCTests() (*oldSysCCInfo, net.Listener, error) {
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	viper.Set("peer.fileSystemPath", "/home/glh/tmp/chaincodes")
	viper.Set("peer.address", "127.0.0.1:12345")
	viper.Set("chaincode.executetimeout", 20*time.Second)

	defer os.RemoveAll("/home/glh/tmp/chaincodes")

	peerMockInitialize()

	peerAddress := "0.0.0.0:21726"
	lis, err := net.Listen("tcp", peerAddress)
	if err != nil {
		return nil, nil, err
	}

	ccStartupTimeout := time.Duration(5000) * time.Millisecond
	ca, _ := accesscontrol.NewCA()
	pb.RegisterChaincodeSupportServer(grpcServer, core.NewChaincodeSupport(peerAddress, false, ccStartupTimeout, ca))

	go grpcServer.Serve(lis)

	//set systemChaincodes to sample
	sysccs := []*scc.SystemChaincode{
		{
			Enabled:   true,
			Name:      "sample_syscc",
			Path:      "/home/glh/project/pallet/src/common/mocks/samplesyscc/samplesyscc",
			InitArgs:  [][]byte{},
			Chaincode: &samplesyscc.SampleSysCC{},
		},
	}

	sysccinfo := &oldSysCCInfo{origSysCCWhitelist: viper.GetStringMapString("chaincode.system")}

	// System chaincode has to be enabled
	viper.Set("chaincode.system", map[string]string{"sample_syscc": "true"})

	sysccinfo.origSystemCC = scc.MockRegisterSysCCs(sysccs)

	/////^^^ system initialization completed ^^^
	return sysccinfo, lis, nil
}

func peerInit() {
	_, _, err := peerInitSysCCTests() //lis
	if err != nil {
		return
	}

	chainID := util.GetTestChainID()
	peerMockCreateChain(chainID)

	scc.DeploySysCCs(chainID)
	//defer scc.DeDeploySysCCs(chainID)
}

func TestExecSysCC(t *testing.T) {
	viper.Set("peer.fileSystemPath", "/home/glh/tmp/chaincodes")
	viper.Set("peer.address", "127.0.0.1:12345")
	viper.Set("chaincode.executetimeout", 20*time.Second)

	// System chaincode has to be enabled
	viper.Set("chaincode.system", map[string]string{"sample_syscc": "true"})

	chainID := util.GetTestChainID()
	f := "putval"
	args := util.ToChaincodeArgs(f, "greeting", "hey there")

	Init()

	var txid string = "1234567890" //default
	nonce, err := crypto.GetRandomNonce()
	if err == nil {
		txid, err = computeProposalTxID(nonce, []byte("glh"))
	}

	Invoke(chainID, "sample_syscc", txid, args, 0)
}

func multSys(t *testing.T){
	var txid string = "1234567890" //default
	fmt.Println("abc enter..................")
	chainID := util.GetTestChainID()
	f := "putval"
	args1 := util.ToChaincodeArgs(f, "greeting", "my test1")
	args2 := util.ToChaincodeArgs(f, "greeting", "my test2")
	args3 := util.ToChaincodeArgs(f, "greeting", "my test3")

	go func() {
		nonce, err := crypto.GetRandomNonce()
		if err == nil {
			txid, err = computeProposalTxID(nonce, []byte("glh"))
		}

		unit, err := Invoke(chainID, "sample_syscc", txid, args1, 0)
		if err != nil {
			t.Error(err)
		}else {
			//t.Logf("ContractId[%s], Function[%s], ReadSet:%v ,WriteSet:%v", unit.ContractId, unit.Function, unit.ReadSet[unit.ContractId], unit.WriteSet[unit.ContractId])
			for k,v := range unit.WriteSet {
				t.Logf("k[%s], v[%v]", k, v)
			}
		}
	}()
	go func() {
		nonce, err := crypto.GetRandomNonce()
		if err == nil {
			txid, err = computeProposalTxID(nonce, []byte("glh"))
		}
		unit, err := Invoke(chainID, "sample_syscc", txid, args2, 0)
		if err != nil {
			t.Error(err)
		}else {
			//t.Logf("ContractId[%s], Function[%s], ReadSet:%v ,WriteSet:%v", unit.ContractId, unit.Function, unit.ReadSet[unit.ContractId], unit.WriteSet[unit.ContractId])
			for k,v := range unit.WriteSet {
				t.Logf("k[%s], v[%v]", k, v)
			}
		}
	}()
	go func() {
		nonce, err := crypto.GetRandomNonce()
		if err == nil {
			txid, err = computeProposalTxID(nonce, []byte("glh"))
		}
		unit, err := Invoke(chainID, "sample_syscc", txid, args3, 0)
		if err != nil {
			t.Error(err)
		}else {
			//t.Logf("ContractId[%s], Function[%s], ReadSet:%v ,WriteSet:%v", unit.ContractId, unit.Function, unit.ReadSet[unit.ContractId], unit.WriteSet[unit.ContractId])
			for k,v := range unit.WriteSet {
				t.Logf("k[%s], v[%v]", k, v)
			}
		}
	}()
//	go Invoke(chainID, "sample_syscc", args)
}

func multMoreSys(t *testing.T){
	fmt.Println("mult enter..................")
	chainID := util.GetTestChainID()
	//f := "putval"
	var wg sync.WaitGroup

	var invokeCount int = 1
	var tmout time.Duration = 500 * time.Second
	var txid string = "1234567890" //default

	for num := 0; num <invokeCount; num += 1 {
		nonce, err := crypto.GetRandomNonce()
		if err == nil {
			txid, err = computeProposalTxID(nonce, []byte("glh"))
		}

		////test
		//testStr := fmt.Sprintf("mytest_%d", num)
		//args := util.ToChaincodeArgs(f, "test", testStr)
		//fmt.Println("++++++++++++++++"+ testStr, "  --txid:" + txid)

		////MultiAddr
		//f := "addrBTC"
		//pubkeyAlice := "029d80ff37838e4989a6aa26af41149d4f671976329e9ddb9b78fdea9814ae6ef5"
		//pubkeyBob := "020106ca23b4f28dbc83838ee4745accf90e5621fe70df5b1ee8f7e1b3b41b64cb"
		//args := util.ToChaincodeArgs(f, "btc", pubkeyAlice, pubkeyBob)

		//		//GetBlance
		//		f := "queryBTC"
		//		//addr := "miZqthevf8LWguQmUR6EwynULqjKmYWxyY"
		//		addr := "2N4jXJyMo8eRKLPWqi5iykAyFLXd6szehwA"
		//		minConf := "5"
		//		args := util.ToChaincodeArgs(f, "btc", addr, minConf)

		////SignTransaction
		//f := "transactionBTC"
		//transactionhex := "010000000236045404e65bd741109db92227ca0dc9274ef717a6612c96cd77b24a17d1bcd700000000b400473044022024e6a6ca006f25ccd3ebf5dadf21397a6d7266536cd336061cd17cff189d95e402205af143f6726d75ac77bc8c80edcb6c56579053d2aa31601b23bc8da41385dd86014c69522103940ab29fbf214da2d8ec99c47db63879957311bd90d2f1c635828604d541051421020106ca23b4f28dbc83838ee4745accf90e5621fe70df5b1ee8f7e1b3b41b64cb21029d80ff37838e4989a6aa26af41149d4f671976329e9ddb9b78fdea9814ae6ef553aeffffffff7c1f7d5407b41abf29d41cf6f122ef2d40f76d956900d2c89314970951ef5b9400000000b40047304402206a1d7a2ae07840957bee708b6d3e1fbe7858760ac378b1e21209b348c1e2a5c402204255cd4cd4e5b5805d44bbebe7464aa021377dca5fc6bf4a5632eb2d8bc9f9e4014c69522103940ab29fbf214da2d8ec99c47db63879957311bd90d2f1c635828604d541051421020106ca23b4f28dbc83838ee4745accf90e5621fe70df5b1ee8f7e1b3b41b64cb21029d80ff37838e4989a6aa26af41149d4f671976329e9ddb9b78fdea9814ae6ef553aeffffffff014431d309000000001976a914bddc9a62e9b7c3cfdbe1c817520e24e32c339f3288ac00000000"
		//redeemhex := "522103940ab29fbf214da2d8ec99c47db63879957311bd90d2f1c635828604d541051421020106ca23b4f28dbc83838ee4745accf90e5621fe70df5b1ee8f7e1b3b41b64cb21029d80ff37838e4989a6aa26af41149d4f671976329e9ddb9b78fdea9814ae6ef553ae"
		//args := util.ToChaincodeArgs(f, "btc", transactionhex,redeemhex)

		//MultiAddr
		f := "addrETH"
		addrAlice := "0x7d7116a8706ae08baa7f4909e26728fa7a5f0365"
		addrBob := "0xaAA919a7c465be9b053673C567D73Be860317963"
		args := util.ToChaincodeArgs(f, "ETH", addrAlice, addrBob)

		fmt.Println("++++++++++++++++ txid:" + txid)

		if num > 2 {
			tmout = 1
		}

		wg.Add(1)
		go func(timeout time.Duration, txid string) {
			unit, err := Invoke(chainID, "sample_syscc", txid, args, timeout)
			if err != nil {
				t.Error(err)
			} else {
				//t.Logf("ContractId[%s], Function[%s], ReadSet:%v ,WriteSet:%v", unit.ContractId, unit.Function, unit.ReadSet[unit.ContractId], unit.WriteSet[unit.ContractId])
				if unit != nil {
					fmt.Println("len(unit.WriteSet) ==== ==== ", len(unit.WriteSet))
					for k, v := range unit.WriteSet {
						t.Logf("k[%s], v[%v]", k, v)
					}
				} else {
					fmt.Println("Not nil error. But nil unit !!!")
				}
			}
			wg.Done()
		}(tmout, txid)
	}
	wg.Wait()

	//	go Invoke(chainID, "sample_syscc", args)
}

func TestExecSysCCMult(t *testing.T) {
	viper.Set("peer.fileSystemPath", "d:\\chaincodes")
	viper.Set("peer.address", "127.0.0.1:12345")
	viper.Set("chaincode.executetimeout", 20*time.Second)

	// System chaincode has to be enabled
	viper.Set("chaincode.system", map[string]string{"sample_syscc": "true"})

	//chainID := util.GetTestChainID()
	//f := "putval"
	//args := util.ToChaincodeArgs(f, "greeting", "hey there")

	Init()
	//Invoke(chainID, "sample_syscc", args)
	//multSys(t)
	multMoreSys(t)
	//func () {
	//	go	Invoke(chainID, "sample_syscc", args)
	//	go	Invoke(chainID, "sample_syscc", args)
	//	go	Invoke(chainID, "sample_syscc", args)
	//}()

	//time.Sleep(20*time.Second)

}

func TestGetSysCCList(t *testing.T) {
	cclist, count, err := GetSysCCList()
	if err != nil {
		t.Log(err)
	}

	//t.Logf("cclist:%v", cclist)
	t.Logf("count:%d", count)

	for idx, cc := range  cclist {
		t.Logf("%d, %s---%s---%v", idx, cc.Name, cc.Path, cc.Enable)
	}
}

func TestInstallCC(t *testing.T) {
	viper.Set("peer.fileSystemPath", "/home/glh/tmp/chaincodes")
	viper.Set("peer.address", "127.0.0.1:12345")
	viper.Set("chaincode.executetimeout", 20*time.Second)

	viper.Set("vm.endpoint", "unix:///var/run/docker.sock")
	viper.Set("chaincode.builder", "palletimg")

	viper.Set("chaincode.system", map[string]string{"sample_syscc": "true"})
	chainID := util.GetTestChainID()

	var txid string = "1234567890" //default
	nonce, err := crypto.GetRandomNonce()
	if err == nil {
		txid, err = computeProposalTxID(nonce, []byte("glh"))
	}

	f := "init"
	args := util.ToChaincodeArgs(f, "a", "100", "b", "200")

	usercc := &ucc.UserChaincode{
		Enabled:   true,
		Name:      "example01",
		Path:      "chaincode/example01",
		Version:   "ptn001",
		InitArgs:  args,
		//Chaincode: &samplesyscc.SampleSysCC{},
	}
	InitNoSysCCC()

	//deploy
	fmt.Print("=======================DeployUserCC=============================")
	err = ucc.DeployUserCC(chainID, usercc, txid, 30*time.Second)
	if err != nil {
		t.Errorf("DeployUserCC err:%s", err)
	}
	time.Sleep(1*time.Second)

	//invoke
	fmt.Print("=======================Invoke=============================")
	f = "invoke"
	args = util.ToChaincodeArgs(f, "111")
	_, err = Invoke(chainID, "example01", txid, args, 0)
	if err != nil {
		t.Errorf("Invoke err:%s", err)
	}
	time.Sleep(2*time.Second)

	//stop
	usercc.Name = "example01"
	fmt.Print("=======================StopUserCC=============================")
	err = ucc.StopUserCC(chainID, usercc, txid,true)
	if err != nil {
		t.Errorf("StopUserCC err:%s", err)
	}
	time.Sleep(2*time.Second)


}

func invokeUserCC(t *testing.T, txid string) {

	chainID := util.GetTestChainID()
	f := "invoke"
	args := util.ToChaincodeArgs(f, "111")


	Invoke(chainID, "example01", txid, args, 0)


	usercc := &ucc.UserChaincode{
		Enabled:   true,
		Name:      "example01",
		Path:      "chaincode/example01",
		Version:   "ptn001",
		InitArgs:  args,
		Chaincode: &samplesyscc.SampleSysCC{},
	}

	fmt.Print("=======================#################=============================")

	time.Sleep(5*time.Second)

	ucc.StopUserCC(chainID, usercc, txid, true)
}



