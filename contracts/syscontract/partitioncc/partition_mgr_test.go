package partitioncc

import (
	"testing"

	"github.com/palletone/go-palletone/dag/modules"
	"github.com/stretchr/testify/assert"

	"github.com/golang/mock/gomock"
	"github.com/palletone/go-palletone/contracts/shim"
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

func newTestPartition() *PartitionChain {
	return &PartitionChain{
		GenesisHash:   "1111111",
		GenesisHeight: 0,
		GasToken:      modules.PTNCOIN,
		Status:        1,
	}
}
