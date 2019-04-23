package partitioncc

import (
	"testing"

	"github.com/palletone/go-palletone/dag/modules"
	"github.com/stretchr/testify/assert"

	"encoding/json"
	"github.com/golang/mock/gomock"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/contracts/shim"
	"github.com/palletone/go-palletone/dag/errors"
	"strings"
)

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
	args := []string{"111111", "0", "222222222", "2", "PTN", "1", "1", "[\"127.0.0.1:1234\",\"192.168.100.2:9090\"]"}
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
	args := []string{"111111", "0", "222222222", "2", "PTN", "1", "1", string(peersJson)}
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
	args := []string{"111111", "0", "222222222", "2", "PTN", "1", "1", string(peersJson)}
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

func newTestPartition() *PartitionChain {
	return &PartitionChain{
		GenesisHash:   "1111111",
		GenesisHeight: 0,
		GasToken:      modules.PTNCOIN,
		Status:        1,
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

	args := []string{"111111", "PTN", "1", "1", "[\"127.0.0.1:1234\",\"192.168.100.2:9090\"]"}
	response := setMainChain(args, stub)
	assert.Equal(t, response.Status, int32(200))
	mainCh := getMainChain(stub)
	t.Log(string(mainCh.Payload))
}
