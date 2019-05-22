package partitioncc

import (
	"testing"

	"github.com/palletone/go-palletone/dag/modules"
	"github.com/stretchr/testify/assert"

	"encoding/hex"
	"encoding/json"
	"github.com/golang/mock/gomock"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/contracts/shim"
	"github.com/palletone/go-palletone/dag/errors"
	"strings"
)

var headerHexString = "f8f3e1a00000000000000000000000000000000000000000000000000000000000000000f86aa1026a16514269bc7324ff9a3b0a39a13e8928cc3d2b2b7d650b5d2c8580fd3e9a23b84630440220545062adc8304f45f429a7f4571f04f39d76125d938a27a7599ddc2619fb5f7a02201af270ab69a957e8626eba0fa181eb8ff9de817bfc026bcb8476aa4d825b16ed8a67726f75705f7369676e8c67726f75705f7075624b6579a0c35639062e40f8891cef2526b387f42e353b8f403b930106bb5aa3519e59e35fd2908157c40f080000000000000000000000809400000000000000000000000000000000000000007b820102"
var headerRlp, _ = hex.DecodeString(headerHexString)

func TestAddPartition(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	stub := shim.NewMockChaincodeStubInterface(mockCtrl)
	//db := make(map[string][]byte)
	//type put func(a string, b []byte) {db[a]=b}
	stub.EXPECT().PutState(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	partition := newTestPartition()
	err := addPartitionChain(stub, partition)
	assert.Nil(t, err)
}
func TestRegisterPartition(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	stub := shim.NewMockChaincodeStubInterface(mockCtrl)
	stub.EXPECT().PutState(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	fundationAddr, _ := common.StringToAddress("P1EZE1HAqMZATrdTkpgmizoRV21rj4pm3db")
	stub.EXPECT().GetInvokeParameters().Return(fundationAddr, nil, nil, "registerPartition", nil, nil).AnyTimes()
	stub.EXPECT().GetSystemConfig("FoundationAddress").Return("P1EZE1HAqMZATrdTkpgmizoRV21rj4pm3db", nil).AnyTimes()
	args := []string{headerHexString, "222222222", "2", "PTN", "1", "1", "1", "1", "2", "[\"127.0.0.1:1234\",\"192.168.100.2:9090\"]"}
	result := registerPartition(args, stub)
	t.Log(result.Status)
}
func TestListPartition(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	stub := shim.NewMockChaincodeStubInterface(mockCtrl)
	db := make(map[string][]byte)
	stub.EXPECT().PutState(gomock.Any(), gomock.Any()).Do(func(key string, value []byte) {
		db[key] = value
	}).AnyTimes()
	stub.EXPECT().GetState(gomock.Any()).DoAndReturn(func(key string) ([]byte, error) {
		value, ok := db[key]
		if ok {
			return value, nil
		} else {
			return nil, errors.New("not found")
		}
	}).AnyTimes()
	stub.EXPECT().GetStateByPrefix(gomock.Any()).DoAndReturn(func(prefix string) ([]*modules.KeyValue, error) {
		rows := []*modules.KeyValue{}
		for k, v := range db {
			if strings.HasPrefix(k, prefix) {
				rows = append(rows, &modules.KeyValue{Key: k, Value: v})
			}
		}
		return rows, nil
	}).AnyTimes()

	peers := []string{"127.0.0.1:1234", "192.168.100.2:9090"}
	peersJson, _ := json.Marshal(peers)
	args := []string{headerHexString, "222222222", "2", "PTN", "1", "1", "1", "1", "2", string(peersJson)}
	partitionChain, _ := buildPartitionChain(args)

	err := addPartitionChain(stub, partitionChain)
	assert.Nil(t, err)
	result := listPartition(stub)
	t.Log(string(result.Payload))
}

func TestUpdatePartition(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	stub := shim.NewMockChaincodeStubInterface(mockCtrl)
	db := make(map[string][]byte)
	stub.EXPECT().PutState(gomock.Any(), gomock.Any()).Do(func(key string, value []byte) {
		db[key] = value
	}).AnyTimes()
	stub.EXPECT().GetState(gomock.Any()).DoAndReturn(func(key string) ([]byte, error) {
		value, ok := db[key]
		if ok {
			return value, nil
		} else {
			return nil, errors.New("not found")
		}
	}).AnyTimes()
	stub.EXPECT().GetStateByPrefix(gomock.Any()).DoAndReturn(func(prefix string) ([]*modules.KeyValue, error) {
		rows := []*modules.KeyValue{}
		for k, v := range db {
			if strings.HasPrefix(k, prefix) {
				rows = append(rows, &modules.KeyValue{Key: k, Value: v})
			}
		}
		return rows, nil
	}).AnyTimes()
	fundationAddr, _ := common.StringToAddress("P1EZE1HAqMZATrdTkpgmizoRV21rj4pm3db")
	stub.EXPECT().GetInvokeParameters().Return(fundationAddr, nil, nil, "registerPartition", nil, nil).AnyTimes()
	stub.EXPECT().GetSystemConfig("FoundationAddress").Return("P1EZE1HAqMZATrdTkpgmizoRV21rj4pm3db", nil).AnyTimes()

	peers := []string{"127.0.0.1:1234", "192.168.100.2:9090"}
	peersJson, _ := json.Marshal(peers)
	args := []string{headerHexString, "222222222", "2", "PTN", "1", "1", "1", "1", "2", string(peersJson)}
	partitionChain, _ := buildPartitionChain(args)

	err := addPartitionChain(stub, partitionChain)
	assert.Nil(t, err)
	result := listPartition(stub)
	t.Log(string(result.Payload))
	args[2] = "FFFFFFFF"
	args[3] = "666"
	response := updatePartition(args, stub)
	assert.Equal(t, response.Status, int32(200))
	result2 := listPartition(stub)
	t.Log(string(result2.Payload))
}

func newTestPartition() *modules.PartitionChain {
	return &modules.PartitionChain{
		GenesisHeaderRlp: headerRlp,
		GasToken:         modules.PTNCOIN,
		Status:           1,
	}
}

func TestSetMainChainAndQuery(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	stub := shim.NewMockChaincodeStubInterface(mockCtrl)
	db := make(map[string][]byte)
	stub.EXPECT().PutState(gomock.Any(), gomock.Any()).Do(func(key string, value []byte) {
		db[key] = value
	}).AnyTimes()
	stub.EXPECT().GetState(gomock.Any()).DoAndReturn(func(key string) ([]byte, error) {
		value, ok := db[key]
		if ok {
			return value, nil
		} else {
			return nil, errors.New("not found")
		}
	}).AnyTimes()
	fundationAddr, _ := common.StringToAddress("P1EZE1HAqMZATrdTkpgmizoRV21rj4pm3db")
	stub.EXPECT().GetInvokeParameters().Return(fundationAddr, nil, nil, "registerPartition", nil, nil).AnyTimes()
	stub.EXPECT().GetSystemConfig("FoundationAddress").Return("P1EZE1HAqMZATrdTkpgmizoRV21rj4pm3db", nil).AnyTimes()

	args := []string{"111111", "PTN", "1", "1", "1", "1", "[\"127.0.0.1:1234\",\"192.168.100.2:9090\"]"}
	response := setMainChain(args, stub)
	assert.Equal(t, response.Status, int32(200))
	mainCh := getMainChain(stub)
	t.Log(string(mainCh.Payload))
}
