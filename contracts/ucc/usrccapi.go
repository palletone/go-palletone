package ucc

import (
	"fmt"
	"github.com/palletone/go-palletone/contracts/core"
	"github.com/palletone/go-palletone/contracts/platforms"
	"github.com/palletone/go-palletone/contracts/rwset"
	"github.com/palletone/go-palletone/contracts/shim"
	"github.com/palletone/go-palletone/core/vmContractPub/ccprovider"
	"github.com/palletone/go-palletone/core/vmContractPub/flogging"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"github.com/palletone/go-palletone/core/vmContractPub/util"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
	"time"
	"os"
	"io/ioutil"
	"io"
	"strings"
	"compress/gzip"
	"archive/tar"
	"github.com/palletone/go-palletone/dag/storage"
	comdb "github.com/palletone/go-palletone/contracts/comm"
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
	logger.Infof("getting deployment spec for chaincode spec: %v\n", spec)
	codePackageBytes, err := platforms.GetDeploymentPayload(spec)
	if err != nil {
		return nil, err
	}

	cdDeploymentSpec := &pb.ChaincodeDeploymentSpec{ChaincodeSpec: spec, CodePackage: codePackageBytes}
	return cdDeploymentSpec, nil
}

func DeployUserCC(chainID string, usrcc *UserChaincode, txid string, txsim rwset.TxSimulator, timeout time.Duration) error {
	var err error

	ccprov := ccprovider.GetChaincodeProvider()
	ctxt := context.Background()
	if txsim != nil {
		ctxt = context.WithValue(ctxt, core.TXSimulatorKey, txsim)
	}

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

func RecoverChainCodeFromDb(chainID string, templateId []byte) (*UserChaincode, error) {
	//从数据库读取
	//解压到指定路径下

	usrCC := &UserChaincode{
	}
	return usrCC, nil

	if 1 == 0 {
		dag, err := comdb.GetCcDagHand()
		if err != nil {
			return nil, err
		}
		_, data, name, path := storage.GetContractTpl(dag.Db, templateId)
		if data == nil || name == "" || path == "" {
			return nil, errors.New("et contract template err")
		}

		targzFile := path + "/" + name + ".tar.gz"
		decompressFile := path + "/" + name

		logger.Infof("write file contract template info [%s]", targzFile)
		err = ioutil.WriteFile(targzFile, data, 0644)
		if err != nil {
			logger.Errorf("write file[%s] fail:%s", targzFile, err)
			return nil, errors.New("write file fail")
		}

		if err = DeCompress(targzFile, decompressFile); err != nil {
			return nil, err
		}

		usrCC := &UserChaincode{
			Name: name,
			//Version:ver,
			Path: path,
		}
		return usrCC, nil
	} else {
		testFile := "/home/glh/go/src/chaincode/abc.tar.gz"
		zipName := "test.tar.gz"
		dir := "/home/glh/go/src/chaincode/"
		//version, zipdata, name, path := storage.GetContractTpl(templateId)
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
		err = ioutil.WriteFile(dir+zipName, filedata, 0644)
		if err != nil {
			logger.Errorf("write file[%s] fail:%s", testFile, err)
			return nil, errors.New("write file fail")
		}

		if err = DeCompress(dir+zipName, "./"); err != nil {
			return nil, err
		}

		usrCC := &UserChaincode{
		}
		return usrCC, nil
	}
}

//压缩 使用gzip压缩成tar.gz
func Compress(files []*os.File, dest string) error {
	d, _ := os.Create(dest)
	defer d.Close()
	gw := gzip.NewWriter(d)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()
	for _, file := range files {
		err := compress(file, "", tw)
		if err != nil {
			return err
		}
	}
	return nil
}

func compress(file *os.File, prefix string, tw *tar.Writer) error {
	info, err := file.Stat()
	if err != nil {
		return err
	}
	if info.IsDir() {
		prefix = prefix + "/" + info.Name()
		fileInfos, err := file.Readdir(-1)
		if err != nil {
			return err
		}
		for _, fi := range fileInfos {
			f, err := os.Open(file.Name() + "/" + fi.Name())
			if err != nil {
				return err
			}
			err = compress(f, prefix, tw)
			if err != nil {
				return err
			}
		}
	} else {
		header, err := tar.FileInfoHeader(info, "")
		header.Name = prefix + "/" + header.Name
		if err != nil {
			return err
		}
		err = tw.WriteHeader(header)
		if err != nil {
			return err
		}
		_, err = io.Copy(tw, file)
		file.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func DeCompress(tarFile, dest string) error {
	srcFile, err := os.Open(tarFile)
	if err != nil {
		return err
	}
	defer srcFile.Close()
	gr, err := gzip.NewReader(srcFile)
	if err != nil {
		return err
	}
	defer gr.Close()
	tr := tar.NewReader(gr)
	for {
		hdr, err := tr.Next()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return err
			}
		}
		filename := dest + hdr.Name
		file, err := createFile(filename)
		if err != nil {
			return err
		}
		io.Copy(file, tr)
	}
	return nil
}

func createFile(name string) (*os.File, error) {
	err := os.MkdirAll(string([]rune(name)[0:strings.LastIndex(name, "/")]), 0755)
	if err != nil {
		return nil, err
	}
	return os.Create(name)
}
