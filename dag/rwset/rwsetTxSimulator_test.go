package rwset

import (
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/dag"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

func TestRwSetTxSimulator_GetTokenBalance(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	dag := dag.NewMockIDag(mockCtrl)
	simulator := &RwSetTxSimulator{}
	simulator.dag = dag
	mockUtxos := mockUtxos()
	dag.EXPECT().GetAddrUtxos(gomock.Any()).Return(mockUtxos, nil).AnyTimes()
	balance, err := simulator.GetTokenBalance([]byte{}, "PalletOne")
	assert.Nil(t, err)
	assert.True(t, len(balance) == 2, "mock has 2 asset,but current is "+strconv.Itoa(len(balance)))
	for k, v := range balance {
		t.Logf("Key:{%s},Value:%d", k.String(), v)
	}
}
func mockUtxos() map[modules.OutPoint]*modules.Utxo {
	result := map[modules.OutPoint]*modules.Utxo{}
	p1 := modules.NewOutPoint(&common.Hash{}, 0, 0)
	asset1 := &modules.Asset{AssetId: modules.PTNCOIN, ChainId: 1}
	fmt.Printf("Mock asset1:%s\n", asset1.String())
	utxo1 := &modules.Utxo{Asset: asset1, Amount: 100, LockTime: 0}
	utxo2 := &modules.Utxo{Asset: asset1, Amount: 200, LockTime: 0}
	asset2 := &modules.Asset{AssetId: modules.BTCCOIN, ChainId: 1}
	fmt.Printf("Mock asset2:%s\n", asset2.String())
	utxo3 := &modules.Utxo{Asset: asset2, Amount: 500, LockTime: 0}
	result[*p1] = utxo1
	p2 := modules.NewOutPoint(&common.Hash{}, 1, 0)
	result[*p2] = utxo2
	p3 := modules.NewOutPoint(&common.Hash{}, 2, 1)
	result[*p3] = utxo3
	return result
}
