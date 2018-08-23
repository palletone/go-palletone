package ucc

import (
	"fmt"
	"time"
	"golang.org/x/net/context"
	"github.com/pkg/errors"
	"github.com/palletone/go-palletone/core/vmContractPub/flogging"
	"github.com/palletone/go-palletone/core/vmContractPub/util"
	"github.com/palletone/go-palletone/contracts/shim"
	"github.com/palletone/go-palletone/core/vmContractPub/ccprovider"
	"github.com/palletone/go-palletone/contracts/platforms"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"os"
	"io/ioutil"
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
	codePackageBytes, err := platforms.GetDeploymentPayload(spec)
	if err != nil {
		return nil, err
	}

	cdDeploymentSpec := &pb.ChaincodeDeploymentSpec{ChaincodeSpec: spec, CodePackage: codePackageBytes}
	return cdDeploymentSpec, nil
}

func DeployUserCC(chainID string, usrcc *UserChaincode, txid string, timeout time.Duration) error {
	var err error

	ccprov := ccprovider.GetChaincodeProvider()
	ctxt := context.Background()
	//todo
	//从数据库中检查并恢复出保存的context数据

	chaincodeID := &pb.ChaincodeID{Path: usrcc.Path, Name: usrcc.Name, Version: usrcc.Version}
	spec := &pb.ChaincodeSpec{Type: pb.ChaincodeSpec_Type(pb.ChaincodeSpec_Type_value["GOLANG"]), ChaincodeId: chaincodeID, Input: &pb.ChaincodeInput{Args: usrcc.InitArgs}}

	// First build and get the deployment spec
	chaincodeDeploymentSpec, err := getDeploymentSpec(ctxt, spec)
	if err != nil {
		logger.Error(fmt.Sprintf("Error deploying chaincode spec: %v\n\n error: %s", spec, err))
		return err
	}

	cccid := ccprov.GetCCContext(chainID, chaincodeDeploymentSpec.ChaincodeSpec.ChaincodeId.Name, usrcc.Version, txid, false, nil, nil)
	_, _, err = ccprov.ExecuteWithErrorFilter(ctxt, cccid, chaincodeDeploymentSpec, timeout)
	if err != nil {
		logger.Errorf("ExecuteWithErrorFilter with usercc.Name[%s] chainId[%s] err !!", usrcc.Name, chainID)
	}
	//logger.Info("rspPaloyd =%v", rspPaloyd)

	logger.Infof("user chaincode %s/%s(%s) deployed", usrcc.Name, chainID, usrcc.Path)

	return err
}

//delete,  not use
func InvokeUserCC(chainID string, usrcc *UserChaincode, timeout time.Duration) error {
	//if !usrcc.Enabled || !isWhitelisted(usrcc) {
	//	logger.Info(fmt.Sprintf("chaincode (%s,%s) disabled", usrcc.Name, usrcc.Path))
	//	return nil
	//}
	var err error

	ccprov := ccprovider.GetChaincodeProvider()
	txid := util.GenerateUUID()
	ctxt := context.Background()

	chaincodeID := &pb.ChaincodeID{Path: usrcc.Path, Name: usrcc.Name, Version: usrcc.Version}
	spec := &pb.ChaincodeSpec{Type: pb.ChaincodeSpec_Type(pb.ChaincodeSpec_Type_value["GOLANG"]), ChaincodeId: chaincodeID, Input: &pb.ChaincodeInput{Args: usrcc.InitArgs}}

	// First build and get the deployment spec
	chaincodeDeploymentSpec, err := getDeploymentSpec(ctxt, spec)

	if err != nil {
		logger.Error(fmt.Sprintf("Error deploying chaincode spec: %v\n\n error: %s", spec, err))
		return err
	}
	version := "aaaaa"

	cccid := ccprov.GetCCContext(chainID, chaincodeDeploymentSpec.ChaincodeSpec.ChaincodeId.Name, version, txid, false, nil, nil)
	_, _, err = ccprov.ExecuteWithErrorFilter(ctxt, cccid, chaincodeDeploymentSpec, timeout)

	if err != nil {
		logger.Errorf("ExecuteWithErrorFilter with usercc.Name[%s] chainId[%s] err !!", usrcc.Name, chainID)
	}
	logger.Infof("user chaincode %s/%s(%s) deployed", usrcc.Name, chainID, usrcc.Path)

	return err
}

func StopUserCC(chainID string, usrcc *UserChaincode, txid string, deleteImage bool) error {
	var err error
	ccprov := ccprovider.GetChaincodeProvider()
	chaincodeID := &pb.ChaincodeID{Path: usrcc.Path, Name: usrcc.Name, Version: usrcc.Version}
	spec := &pb.ChaincodeSpec{
		Type:        pb.ChaincodeSpec_Type(pb.ChaincodeSpec_Type_value["GOLANG"]),
		ChaincodeId: chaincodeID,
		Input: &pb.ChaincodeInput{
			Args: usrcc.InitArgs,
		},
	}

	chaincodeDeploymentSpec := &pb.ChaincodeDeploymentSpec{ChaincodeSpec: spec, CodePackage: nil}
	//chaincodeDeploymentSpec, err := getDeploymentSpec(context.Background(), spec)
	//if err != nil {
	//}

	cccid := ccprov.GetCCContext(chainID, usrcc.Name, usrcc.Version, txid, false, nil, nil)
	err = ccprov.Stop(context.Background(), cccid, chaincodeDeploymentSpec)
	if err != nil {
	}

	if deleteImage {
		logger.Info("destroyImage not complete")
		//dir := controller.DestroyImageReq{CCID: ccintf.CCID{
		//	ChaincodeSpec: spec,
		//	NetworkID:     theChaincodeSupport.peerNetworkID,
		//	PeerID:        theChaincodeSupport.peerID,
		//	ChainID:       cccid.ChainID},Force: true,
		//	NoPrune: true,
		//}
		//_, err = controller.VMCProcess(context.Background(), controller.DOCKER, dir)
		//if err != nil {
		//	err = fmt.Errorf("Error destroying image: %s", err)
		//	return err
		//}
	}

	return nil
}

func GetUserCCPayload(chainID string, usrcc *UserChaincode) (payload []byte, err error) {
	chaincodeID := &pb.ChaincodeID{Path: usrcc.Path, Name: usrcc.Name, Version: usrcc.Version}
	spec := &pb.ChaincodeSpec{Type: pb.ChaincodeSpec_Type(pb.ChaincodeSpec_Type_value["GOLANG"]), ChaincodeId: chaincodeID, Input: &pb.ChaincodeInput{Args: usrcc.InitArgs}}

	data, err := platforms.GetChainCodePayload(spec)
	if err != nil {

		return nil, errors.New("GetChainCodePayload fail")
	}

	return data, nil
}

func RecoverChainCodeFromDb(chainID string, templateId []byte) ( *UserChaincode, error) {
	//从数据库读取
	//解压到指定路径下

	testFile := "/home/glh/go/src/chaincode/abc.tar.gz"

	zipName := "test.tar.gz"
	dir := "/home/glh/go/src/chaincode/"

	//read
	fi, err := os.Open(testFile)
	if err != nil {
		logger.Errorf("open file[%s] fail:%s", testFile, err)
		return nil, errors.New("open file fail")
	}
	defer fi.Close()
	filedata, err := ioutil.ReadAll(fi)
	if err != nil {
		logger.Errorf("read file[%s] fail:%s", testFile, err)
		return nil, errors.New("read file fail")
	}

	//write
	err = ioutil.WriteFile(dir + zipName, filedata, 0644)
	if err != nil {
		logger.Errorf("write file[%s] fail:%s", testFile, err)
		return nil, errors.New("write file fail")
	}

	usrCC := &UserChaincode{
	}

	return usrCC, nil
}



