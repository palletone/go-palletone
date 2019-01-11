package ucc

import (
	"archive/tar"
	"compress/gzip"
	"fmt"

	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/contracts/comm"
	"github.com/palletone/go-palletone/contracts/core"
	"github.com/palletone/go-palletone/contracts/platforms"
	"github.com/palletone/go-palletone/contracts/shim"
	"github.com/palletone/go-palletone/core/vmContractPub/ccprovider"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"github.com/palletone/go-palletone/dag/rwset"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
	"io"
	"os"
	"path"
	"time"
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

//var log = flogging.MustGetLogger("uccapi")

// buildLocal builds a given chaincode code
func buildUserCC(context context.Context, spec *pb.ChaincodeSpec) (*pb.ChaincodeDeploymentSpec, error) {
	var codePackageBytes []byte
	chaincodeDeploymentSpec := &pb.ChaincodeDeploymentSpec{ExecEnv: pb.ChaincodeDeploymentSpec_DOCKER, ChaincodeSpec: spec, CodePackage: codePackageBytes}
	return chaincodeDeploymentSpec, nil
}

func getDeploymentSpec(_ context.Context, spec *pb.ChaincodeSpec) (*pb.ChaincodeDeploymentSpec, error) {
	log.Debugf("getting deployment spec for chaincode spec: %v\n", spec)
	codePackageBytes, err := platforms.GetDeploymentPayload(spec)
	if err != nil {
		return nil, err
	}
	cdDeploymentSpec := &pb.ChaincodeDeploymentSpec{ChaincodeSpec: spec, CodePackage: codePackageBytes}
	return cdDeploymentSpec, nil
}

func mockerDeployUserCC() error{
	log.Debug("==================mockerDeployUserCC enter")
	time.Sleep(time.Duration(10)*time.Second)
	log.Debug("==================mockerDeployUserCC end")

	return nil
}

func DeployUserCC(chaincodeData []byte,spec *pb.ChaincodeSpec, chainID string, usrcc *UserChaincode, txid string, txsim rwset.TxSimulator, timeout time.Duration) error {
	ccprov := ccprovider.GetChaincodeProvider()
	ctxt := context.Background()
	if txsim != nil {
		ctxt = context.WithValue(ctxt, core.TXSimulatorKey, txsim)
	}
	//chaincodeDeploymentSpec, err := getDeploymentSpec(ctxt, spec)
	cdDeploymentSpec := &pb.ChaincodeDeploymentSpec{ChaincodeSpec: spec, CodePackage: chaincodeData}
	//if err != nil {
	//	log.Error(fmt.Sprintf("Error deploying chaincode spec: %v\n\n error: %s", spec, err))
	//	return err
	//}
	// 部署是应该还没有合约ID，返回的才是合约ID
	cccid := ccprov.GetCCContext(nil, chainID, cdDeploymentSpec.ChaincodeSpec.ChaincodeId.Name, usrcc.Version, txid, false, nil, nil)
	_, _, err := ccprov.ExecuteWithErrorFilter(ctxt, cccid, cdDeploymentSpec, timeout)
	if err != nil {
		log.Errorf("ExecuteWithErrorFilter with usercc.Name[%s] chainId[%s] err !!", usrcc.Name, chainID)
	}
	log.Debugf("user chaincode [%s]-[%s]-[%s]-[%s] deployed", usrcc.Name,usrcc.Path,usrcc.Version,chainID)
	return nil
}

func StopUserCC(contractid []byte, chainID string, usrcc *UserChaincode, txid string, deleteImage bool) error {
	ccprov := ccprovider.GetChaincodeProvider()
	chaincodeID := &pb.ChaincodeID{Path: usrcc.Path, Name: usrcc.Name, Version: usrcc.Version}
	spec := &pb.ChaincodeSpec{
		Type:        pb.ChaincodeSpec_Type(pb.ChaincodeSpec_Type_value["GOLANG"]),
		ChaincodeId: chaincodeID,
		Input: &pb.ChaincodeInput{
			Args: usrcc.InitArgs,
		},
	}
	chaincodeDeploymentSpec := &pb.ChaincodeDeploymentSpec{
		ChaincodeSpec: spec,
		CodePackage:   nil,
	}
	cccid := ccprov.GetCCContext(contractid, chainID, usrcc.Name, usrcc.Version, txid, false, nil, nil)
	if err := ccprov.Stop(context.Background(), cccid, chaincodeDeploymentSpec); err != nil {
		return err
	}

	if deleteImage {
		return ccprov.Destory(context.Background(), cccid, chaincodeDeploymentSpec)
	} else {
		return nil
	}
}

func GetUserCCPayload(chainID string, usrcc *UserChaincode) (payload []byte, err error) {
	chaincodeID := &pb.ChaincodeID{Path: usrcc.Path, Name: usrcc.Name, Version: usrcc.Version}
	spec := &pb.ChaincodeSpec{Type: pb.ChaincodeSpec_Type(pb.ChaincodeSpec_Type_value["GOLANG"]), ChaincodeId: chaincodeID, Input: &pb.ChaincodeInput{Args: usrcc.InitArgs}}
	chaincodeData, err := platforms.GetChainCodePayload(spec)
	if err != nil {
		return nil, errors.New("GetChainCodePayload fail")
	}
	return chaincodeData, nil
}

func RecoverChainCodeFromDb(spec *pb.ChaincodeSpec, chainID string, templateId []byte) (*UserChaincode,[]byte, error) {
	//从数据库读取
	dag, err := comm.GetCcDagHand()
	if err != nil {
		log.Error("getCcDagHand err:", "error", err)
		return nil, nil,err
	}
	v, chaincodeData, name, path, tplVer := dag.GetContractTpl(templateId)
	if chaincodeData == nil || name == "" || path == "" || tplVer ==  ""{
		log.Error("getContractTpl err:", "error", v)
		return nil,nil, errors.New("GetContractTpl contract template err")
	}
	usrCC := &UserChaincode{
		Name:    name,
		Version: tplVer,
		Path:    path,
	}
	return usrCC, chaincodeData,nil
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

	//todo del
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

//+++++++++++++
func UnTarGz(srcFilePath string, destDirPath string) error {
	fmt.Println("UnTarGzing enter, srcPath:" + srcFilePath + ", destPath:" + destDirPath)
	_, err := os.Stat(destDirPath)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.Mkdir(destDirPath, os.ModePerm); err != nil {
				fmt.Printf("os.Mkdir err =%s", err)
				return err
			}
		}
	}
	fr, err := os.Open(srcFilePath)
	if err != nil {
		fmt.Printf("os.Open err =%s", err)
		return err
	}
	defer fr.Close()

	// Gzip reader
	gr, err := gzip.NewReader(fr)
	if err != nil {
		fmt.Printf("gzip.NewReader  err =%s", err)
		return err
	}

	// Tar reader
	tr := tar.NewReader(gr)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			// End of tar archive
			break
		}
		//handleError(err)
		//fmt.Println("UnTarGzing file..." + hdr.Name)
		// Check if it is diretory or file
		if hdr.Typeflag != tar.TypeDir {
			// Get files from archive
			// Create diretory before create file
			os.MkdirAll(destDirPath+"/"+path.Dir(hdr.Name), os.ModePerm)
			// Write data to file
			fw, err := os.Create(destDirPath + "/" + hdr.Name)
			if err != nil {
				fmt.Printf("os.Createdoc  err =%s", err)
				return err
			}
			_, err = io.Copy(fw, tr)
			if err != nil {
				fmt.Printf("os.Createdoc  err =%s", err)
				return err
			}
		}
	}
	fmt.Println("Well done!")
	return nil
}

/*
func TarGz(srcDirPath string, destFilePath string) {
	fw, err := os.Create(destFilePath)
	handleError(err)
	defer fw.Close()

	// Gzip writer
	gw := gzip.NewWriter(fw)
	defer gw.Close()

	// Tar writer
	tw := tar.NewWriter(gw)
	defer tw.Close()

	// Check if it's a file or a directory
	f, err := os.Open(srcDirPath)
	handleError(err)
	fi, err := f.Stat()
	handleError(err)
	if fi.IsDir() {
		// handle source directory
		fmt.Println("Cerating tar.gz from directory...")
		tarGzDir(srcDirPath, path.Base(srcDirPath), tw)
	} else {
		// handle file directly
		fmt.Println("Cerating tar.gz from " + fi.Name() + "...")
		tarGzFile(srcDirPath, fi.Name(), tw, fi)
	}
	fmt.Println("Well done!")
}

// Deal with directories
// if find files, handle them with tarGzFile
// Every recurrence append the base path to the recPath
// recPath is the path inside of tar.gz
func tarGzDir(srcDirPath string, recPath string, tw *tar.Writer) {
	// Open source diretory
	dir, err := os.Open(srcDirPath)
	handleError(err)
	defer dir.Close()

	// Get file info slice
	fis, err := dir.Readdir(0)
	handleError(err)
	for _, fi := range fis {
		// Append path
		curPath := srcDirPath + "/" + fi.Name()
		// Check it is directory or file
		if fi.IsDir() {
			// Directory
			// (Directory won't add unitl all subfiles are added)
			fmt.Printf("Adding path...%s\\n", curPath)
			tarGzDir(curPath, recPath+"/"+fi.Name(), tw)
		} else {
			// File
			fmt.Printf("Adding file...%s\\n", curPath)
		}

		tarGzFile(curPath, recPath+"/"+fi.Name(), tw, fi)
	}
}

// Deal with files
func tarGzFile(srcFile string, recPath string, tw *tar.Writer, fi os.FileInfo) {
	if fi.IsDir() {
		// Create tar header
		hdr := new(tar.Header)
		// if last character of header name is '/' it also can be directory
		// but if you don't set Typeflag, error will occur when you untargz
		hdr.Name = recPath + "/"
		hdr.Typeflag = tar.TypeDir
		hdr.Size = 0
		//hdr.Mode = 0755 | c_ISDIR
		hdr.Mode = int64(fi.Mode())
		hdr.ModTime = fi.ModTime()

		// Write hander
		err := tw.WriteHeader(hdr)
		handleError(err)
	} else {
		// File reader
		fr, err := os.Open(srcFile)
		handleError(err)
		defer fr.Close()

		// Create tar header
		hdr := new(tar.Header)
		hdr.Name = recPath
		hdr.Size = fi.Size()
		hdr.Mode = int64(fi.Mode())
		hdr.ModTime = fi.ModTime()

		// Write hander
		err = tw.WriteHeader(hdr)
		if err != nil {
			fmt.Printf("gzip.NewReader  err =%s", err)
			return
		}

		// Write file data
		_, err = io.Copy(tw, fr)
		if err != nil {
			fmt.Printf("gzip.NewReader  err =%s", err)
			return
		}
	}
}
*/
