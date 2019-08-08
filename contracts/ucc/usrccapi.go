package ucc

import (
	"github.com/pkg/errors"
	"golang.org/x/net/context"
	"time"

	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/contracts/comm"
	cfg "github.com/palletone/go-palletone/contracts/contractcfg"
	"github.com/palletone/go-palletone/contracts/core"
	"github.com/palletone/go-palletone/contracts/platforms"
	"github.com/palletone/go-palletone/core/vmContractPub/ccprovider"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"github.com/palletone/go-palletone/dag/rwset"
)

type UserChaincode struct {
	Name    string //Unique name of the chaincode
	Path    string //Path to the chaincode; currently not used
	Version string //chainCode Version
	//Desciption     string
	//Abi            string
	Language string
	//InitArgs       [][]byte       //InitArgs initialization arguments to startup the chaincode
	//Chaincode      shim.Chaincode // Chaincode is the actual chaincode object
	//InvokableCC2CC bool           //InvokableCC2CC keeps track of whether this chaincode
	Enabled bool //Enabled a convenient switch to enable/disable chaincode
}

//func buildUserCC(context context.Context, spec *pb.ChaincodeSpec) (*pb.ChaincodeDeploymentSpec, error) {
//	var codePackageBytes []byte
//	chaincodeDeploymentSpec := &pb.ChaincodeDeploymentSpec{ExecEnv: pb.ChaincodeDeploymentSpec_DOCKER, ChaincodeSpec: spec, CodePackage: codePackageBytes}
//	return chaincodeDeploymentSpec, nil
//}

func getDeploymentSpec(spec *pb.ChaincodeSpec) (*pb.ChaincodeDeploymentSpec, error) {
	log.Debugf("getting deployment spec for chaincode spec: %v\n", spec)
	codePackageBytes, err := platforms.GetDeploymentPayload(spec)
	if err != nil {
		return nil, err
	}
	cdDeploymentSpec := &pb.ChaincodeDeploymentSpec{ChaincodeSpec: spec, CodePackage: codePackageBytes}
	return cdDeploymentSpec, nil
}

//func mockerDeployUserCC() error {
//	log.Debug("==================mockerDeployUserCC enter")
//	time.Sleep(time.Duration(1) * time.Second)
//	log.Debug("==================mockerDeployUserCC end")
//
//	return nil
//}

func DeployUserCC(contractId []byte, chaincodeData []byte, spec *pb.ChaincodeSpec, chainID string, txid string, txsim rwset.TxSimulator, timeout time.Duration) error {
	//return mockerDeployUserCC()

	cdDeploymentSpec := &pb.ChaincodeDeploymentSpec{}
	var err error
	if cfg.DebugTest {
		cdDeploymentSpec, err = getDeploymentSpec(spec)
		if err != nil {
			return err
		}
	} else {
		cdDeploymentSpec.ChaincodeSpec = spec
		cdDeploymentSpec.CodePackage = chaincodeData
	}
	ccprov := ccprovider.GetChaincodeProvider()
	ctxt := context.Background()
	if txsim != nil {
		ctxt = context.WithValue(ctxt, core.TXSimulatorKey, txsim)
	}
	cccid := ccprov.GetCCContext(contractId, chainID, cdDeploymentSpec.ChaincodeSpec.ChaincodeId.Name, cdDeploymentSpec.ChaincodeSpec.ChaincodeId.Version, txid, false, nil, nil)
	_, _, err = ccprov.ExecuteWithErrorFilter(ctxt, cccid, cdDeploymentSpec, timeout)
	if err != nil {
		log.Errorf("ExecuteWithErrorFilter with usercc.Name[%s] chainId[%s] err !!", cdDeploymentSpec.ChaincodeSpec.ChaincodeId.Name, chainID)
		ccprov.Stop(ctxt, cccid, cdDeploymentSpec, false)
		return err
	}
	log.Debugf("user chaincode chainID[%s]-name[%s]-path[%s]-version[%s] deployed", chainID, cdDeploymentSpec.ChaincodeSpec.ChaincodeId.Name, cdDeploymentSpec.ChaincodeSpec.ChaincodeId.Path, cdDeploymentSpec.ChaincodeSpec.ChaincodeId.Version)
	return nil
}

func StopUserCC(contractid []byte, chainID string, usrcc *UserChaincode, txid string, deleteImage bool, dontRmCon bool) error {
	ccprov := ccprovider.GetChaincodeProvider()
	chaincodeID := &pb.ChaincodeID{Path: usrcc.Path, Name: usrcc.Name, Version: usrcc.Version}
	spec := &pb.ChaincodeSpec{
		Type:        pb.ChaincodeSpec_Type(pb.ChaincodeSpec_Type_value[usrcc.Language]),
		ChaincodeId: chaincodeID,
		//Input: &pb.ChaincodeInput{
		//	Args: usrcc.InitArgs,
		//},
	}
	chaincodeDeploymentSpec := &pb.ChaincodeDeploymentSpec{
		ChaincodeSpec: spec,
		CodePackage:   nil,
	}
	cccid := ccprov.GetCCContext(contractid, chainID, usrcc.Name, usrcc.Version, txid, false, nil, nil)
	if err := ccprov.Stop(context.Background(), cccid, chaincodeDeploymentSpec, dontRmCon); err != nil {
		return err
	}

	if deleteImage {
		return ccprov.Destroy(context.Background(), cccid, chaincodeDeploymentSpec)
	} else {
		return nil
	}
}

func GetUserCCPayload(usrcc *UserChaincode) (payload []byte, err error) {
	chaincodeID := &pb.ChaincodeID{Path: usrcc.Path, Name: usrcc.Name, Version: usrcc.Version}
	spec := &pb.ChaincodeSpec{Type: pb.ChaincodeSpec_Type(pb.ChaincodeSpec_Type_value[usrcc.Language]), ChaincodeId: chaincodeID}
	chaincodeData, err := platforms.GetChainCodePayload(spec)
	if err != nil {
		log.Error("getChainCodePayload err:", "error", err)
		return nil, errors.New("GetChainCodePayload fail")
	}
	return chaincodeData, nil
}

func RecoverChainCodeFromDb(chainID string, templateId []byte) (*UserChaincode, []byte, error) {
	//todo, for test
	if cfg.DebugTest {
		usrCC1 := &UserChaincode{}
		return usrCC1, nil, nil
	}

	dag, err := comm.GetCcDagHand()
	if err != nil {
		log.Error("getCcDagHand err:", "error", err)
		return nil, nil, err
	}
	tpl, err := dag.GetContractTpl(templateId)
	if err != nil {
		return nil, nil, errors.New("GetContractTpl contract template err")
	}
	chaincodeData, err := dag.GetContractTplCode(templateId)
	if err != nil {
		return nil, nil, errors.New("GetContractTpl contract code err")
	}
	usrCC := &UserChaincode{
		Name:     tpl.TplName,
		Version:  tpl.Version,
		Path:     tpl.Path,
		Language: tpl.Language,
	}
	return usrCC, chaincodeData, nil

	//todo, For future testing, please don't delete this code.
	//envpath, err := platforms.GetPlatformEnvPath(spec)
	//if err != nil {
	//	log.Error("getPlatformEnvPath err:", "error", err)
	//	return nil, err
	//}
	//targzFile := envpath + "/" + name + ".tar.gz"
	//decompressFile := envpath
	//log.Debugf("name[%s]path[%s]ver[%v]-tar[%s]untar path[%s]", name, path, v, targzFile, decompressFile)
	//
	//_, err = os.Stat(targzFile)
	//if err != nil {
	//	if os.IsExist(err) {
	//		os.Remove(targzFile)
	//	}
	//}
	//
	//err = ioutil.WriteFile(targzFile, data, 0644)
	//if err != nil {
	//	log.Errorf("write file[%s] fail:%s", targzFile, err)
	//	return nil, errors.New("write file fail")
	//}
	//if err = UnTarGz(targzFile, decompressFile); err != nil {
	//	return nil, err
	//}
	//err = os.Remove(targzFile)
	//if err != nil {
	//	return nil, err
	//}

	//todo del...
	//usrCC := &UserChaincode{}
	//return usrCC, nil

	//if 1 == 1 {
	//	envpath, err := platforms.GetPlatformEnvPath(spec)
	//	if err != nil {
	//		log.Error("getPlatformEnvPath err:", "error", err)
	//		return nil, err
	//	}
	//	dag, err := comdb.GetCcDagHand()
	//	if err != nil {
	//		log.Error("getCcDagHand err:", "error", err)
	//		return nil, err
	//	}
	//	v, data, name, path := dag.GetContractTpl(templateId)
	//	if data == nil || name == "" || path == "" {
	//		log.Error("getContractTpl err:", "error")
	//		return nil, errors.New("GetContractTpl contract template err")
	//	}
	//	targzFile := envpath + "/tmp/" + name + ".tar.gz"
	//	decompressFile := envpath
	//	log.Infof("name[%s]path[%s]ver[%v]-tar[%s]untar path[%s]", name, path, v, targzFile, decompressFile)
	//
	//	_, err = os.Stat(targzFile)
	//	if err != nil {
	//		if os.IsExist(err) {
	//			os.Remove(targzFile)
	//		}
	//	}
	//
	//	err = ioutil.WriteFile(targzFile, data, 0644)
	//	if err != nil {
	//		log.Errorf("write file[%s] fail:%s", targzFile, err)
	//		return nil, errors.New("write file fail")
	//	}
	//	if err = UnTarGz(targzFile, decompressFile); err != nil {
	//		return nil, err
	//	}
	//
	//	usrCC := &UserChaincode{
	//		Name: name,
	//		//Version:ver,
	//		Path: path,
	//	}
	//	//todo
	//	return usrCC, nil
	//} else {
	//	testFile := "/home/glh/go/src/chaincode/abc.tar.gz"
	//	zipName := "test.tar.gz"
	//	dir := "/home/glh/go/src/chaincode/"
	//	//version, zipdata, name, path := storage.GetContractTpl(templateId)
	//	//read
	//	fi, err := os.Open(testFile)
	//	if err != nil {
	//		log.Errorf("open file[%s] fail:%s", testFile, err)
	//		return nil, errors.New("open file fail")
	//	}
	//	defer fi.Close()
	//	filedata, err := ioutil.ReadAll(fi)
	//	if err != nil {
	//		log.Errorf("read file[%s] fail:%s", testFile, err)
	//		return nil, errors.New("read file fail")
	//	}
	//
	//	//write
	//	err = ioutil.WriteFile(dir+zipName, filedata, 0644)
	//	if err != nil {
	//		log.Errorf("write file[%s] fail:%s", testFile, err)
	//		return nil, errors.New("write file fail")
	//	}
	//
	//	if err = UnTarGz(dir+zipName, "./"); err != nil {
	//		log.Errorf("DeCompress[%s] fail:%s", testFile, err)
	//		return nil, err
	//	}
	//
	//	usrCC := &UserChaincode{}
	//	return usrCC, nil
	//}
}
