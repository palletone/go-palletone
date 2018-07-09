package test

import (
	"os"
	"testing"

	"github.com/palletone/go-palletone/core/vmContractPub/util"
	"github.com/palletone/go-palletone/vm/controller"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"github.com/palletone/go-palletone/core/vmContractPub/mocks/config"

)

func TestMain(m *testing.M) {
	config.SetupTestConfig()
	os.Exit(m.Run())
}

func TestJava_BuildImage(t *testing.T) {

	vm, err := container.NewVM()
	if err != nil {
		t.Fail()
		t.Logf("Error getting VM: %s", err)
		return
	}

	chaincodePath := "../../../../../examples/chaincode/java/SimpleSample"
	//TODO find a better way to launch example java chaincode
	spec := &pb.ChaincodeSpec{Type: pb.ChaincodeSpec_JAVA, ChaincodeId: &pb.ChaincodeID{Name: "ssample", Path: chaincodePath}, Input: &pb.ChaincodeInput{Args: util.ToChaincodeArgs("f")}}
	if err := vm.BuildChaincodeContainer(spec); err != nil {
		t.Fail()
		t.Log(err)
	}

}
