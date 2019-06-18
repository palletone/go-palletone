package partitioncc

import (
	"testing"

	"github.com/palletone/go-palletone/dag/modules"
	"github.com/stretchr/testify/assert"

	"encoding/hex"
	"encoding/json"
	"strings"

	"github.com/golang/mock/gomock"
	"github.com/palletone/go-palletone/contracts/shim"
	"github.com/palletone/go-palletone/dag/errors"
)

var headerHexString = "f8acc0f86ba1023a781c75b6af6786ab68cf82877dff850d2cfb80cb308e286b3eda0ae15a7679b8473045022100e794e6e513f0817403fc912d4fd64c0bd9ccdfc763b9e3bf81f882813605c81202205ab7bc08bea6009c93ac06f93b1dd8b70f7babf75191c0a56679a084210e6c298080a014f1fcbe671fd3b8fd357935e046f6a1fb0fa311f88835ce7a0340695c453ee6c0d290400082bb0800000000000000000000008080845ce5535580"
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

//func TestRegisterPartition(t *testing.T) {
//	mockCtrl := gomock.NewController(t)
//	defer mockCtrl.Finish()
//	stub := shim.NewMockChaincodeStubInterface(mockCtrl)
//	stub.EXPECT().PutState(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
//	fundationAddr, _ := common.StringToAddress("P1EZE1HAqMZATrdTkpgmizoRV21rj4pm3db")
//	stub.EXPECT().GetInvokeParameters().Return(fundationAddr, nil, nil, "registerPartition", nil, nil).AnyTimes()
//	//stub.EXPECT().GetSystemConfig("FoundationAddress").Return("P1EZE1HAqMZATrdTkpgmizoRV21rj4pm3db", nil).AnyTimes()
//	args := []string{headerHexString, "222222222", "2", "PTN", "1", "1", "1", "1", "2", "[\"127.0.0.1:1234\",\"192.168.100.2:9090\"]", "[]"}
//	result := registerPartition(args, stub)
//	t.Log(result.Message)
//	assert.EqualValues(t, 200, result.Status)
//}

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
	args := []string{headerHexString, "222222222", "2", "PTN", "1", "1", "1", "1", "2", string(peersJson), "[\"PTN\"]"}
	partitionChain, _ := buildPartitionChain(args)

	err := addPartitionChain(stub, partitionChain)
	assert.Nil(t, err)
	result := listPartition(stub)
	t.Log(string(result.Payload))
}

//func TestUpdatePartition(t *testing.T) {
//	mockCtrl := gomock.NewController(t)
//	defer mockCtrl.Finish()
//	stub := shim.NewMockChaincodeStubInterface(mockCtrl)
//	db := make(map[string][]byte)
//	stub.EXPECT().PutState(gomock.Any(), gomock.Any()).Do(func(key string, value []byte) {
//		db[key] = value
//	}).AnyTimes()
//	stub.EXPECT().GetState(gomock.Any()).DoAndReturn(func(key string) ([]byte, error) {
//		value, ok := db[key]
//		if ok {
//			return value, nil
//		} else {
//			return nil, errors.New("not found")
//		}
//	}).AnyTimes()
//	stub.EXPECT().GetStateByPrefix(gomock.Any()).DoAndReturn(func(prefix string) ([]*modules.KeyValue, error) {
//		rows := []*modules.KeyValue{}
//		for k, v := range db {
//			if strings.HasPrefix(k, prefix) {
//				rows = append(rows, &modules.KeyValue{Key: k, Value: v})
//			}
//		}
//		return rows, nil
//	}).AnyTimes()
//	fundationAddr, _ := common.StringToAddress("P1EZE1HAqMZATrdTkpgmizoRV21rj4pm3db")
//	stub.EXPECT().GetInvokeParameters().Return(fundationAddr, nil, nil, "registerPartition", nil, nil).AnyTimes()
//	//stub.EXPECT().GetSystemConfig("FoundationAddress").Return("P1EZE1HAqMZATrdTkpgmizoRV21rj4pm3db", nil).AnyTimes()
//
//	peers := []string{"127.0.0.1:1234", "192.168.100.2:9090"}
//	peersJson, _ := json.Marshal(peers)
//	args := []string{headerHexString, "222222222", "2", "PTN", "1", "1", "1", "1", "2", string(peersJson), "[]"}
//	partitionChain, _ := buildPartitionChain(args)
//
//	err := addPartitionChain(stub, partitionChain)
//	assert.Nil(t, err)
//	result := listPartition(stub)
//	t.Log(string(result.Payload))
//	args[2] = "FFFFFFFF"
//	args[3] = "666"
//	response := updatePartition(args, stub)
//	assert.Equal(t, response.Status, int32(200))
//	result2 := listPartition(stub)
//	t.Log(string(result2.Payload))
//}

func newTestPartition() *modules.PartitionChain {
	return &modules.PartitionChain{
		GenesisHeaderRlp: headerRlp,
		GasToken:         modules.PTNCOIN,
		Status:           1,
	}
}

//func TestSetMainChainAndQuery(t *testing.T) {
//	mockCtrl := gomock.NewController(t)
//	defer mockCtrl.Finish()
//	stub := shim.NewMockChaincodeStubInterface(mockCtrl)
//	db := make(map[string][]byte)
//	stub.EXPECT().PutState(gomock.Any(), gomock.Any()).Do(func(key string, value []byte) {
//		db[key] = value
//	}).AnyTimes()
//	stub.EXPECT().GetState(gomock.Any()).DoAndReturn(func(key string) ([]byte, error) {
//		value, ok := db[key]
//		if ok {
//			return value, nil
//		} else {
//			return nil, errors.New("not found")
//		}
//	}).AnyTimes()
//	fundationAddr, _ := common.StringToAddress("P1EZE1HAqMZATrdTkpgmizoRV21rj4pm3db")
//	stub.EXPECT().GetInvokeParameters().Return(fundationAddr, nil, nil, "registerPartition", nil, nil).AnyTimes()
//	//stub.EXPECT().GetSystemConfig("FoundationAddress").Return("P1EZE1HAqMZATrdTkpgmizoRV21rj4pm3db", nil).AnyTimes()
//
//	args := []string{headerHexString, "PTN", "1", "1", "1", "1", "3", "[\"127.0.0.1:1234\",\"192.168.100.2:9090\"]", "[]"}
//	response := setMainChain(args, stub)
//	assert.Equal(t, response.Status, int32(200))
//	mainCh := getMainChain(stub)
//	t.Log(string(mainCh.Payload))
//}
