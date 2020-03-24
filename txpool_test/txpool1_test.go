/*
   This file is part of go-palletone.
   go-palletone is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.
   go-palletone is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.
   You should have received a copy of the GNU General Public License
   along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
*/

/*
 * @author PalletOne core developers <dev@pallet.one>
 * @date 2018
 */

package txpool_test

import (
	"encoding/json"
	"log"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/dag/mock"
	"github.com/palletone/go-palletone/dag/modules"
	tp2 "github.com/palletone/go-palletone/txpool2"
	tp1 "github.com/palletone/go-palletone/txspool"
	"github.com/stretchr/testify/assert"
)

// build test case.
func newTestPaymentTx(preTxHash common.Hash) *modules.Transaction {
	pay1s := &modules.PaymentPayload{
		LockTime: 0,
	}

	output := modules.NewTxOut(Ptn2Dao(10), []byte{0xee, 0xbb}, modules.NewPTNAsset())
	pay1s.AddTxOut(output)

	input := modules.Input{}
	input.PreviousOutPoint = modules.NewOutPoint(preTxHash, 0, 0)
	input.SignatureScript = []byte{}
	input.Extra = []byte("Test")

	pay1s.AddTxIn(&input)

	msg := &modules.Message{
		App:     modules.APP_PAYMENT,
		Payload: pay1s,
	}
	tx := modules.NewTransaction([]*modules.Message{msg})
	return tx
}
func Ptn2Dao(ptn uint64) uint64 {
	return ptn * 100000000
}

func newCcInvokeRequest(preTxHash common.Hash) *modules.Transaction {
	req := newTestPaymentTx(preTxHash)
	invoke := &modules.ContractInvokeRequestPayload{
		ContractId: []byte("PC1"),
		Args:       [][]byte{[]byte("put"), []byte("a")},
		Timeout:    0,
	}
	req.AddMessage(modules.NewMessage(modules.APP_CONTRACT_INVOKE_REQUEST, invoke))
	return req
}
func newCcInvokeFullTx(preTxHash common.Hash) *modules.Transaction {
	req := newCcInvokeRequest(preTxHash)
	result := &modules.ContractInvokePayload{
		ContractId: []byte("PC1"),
		Args:       [][]byte{[]byte("put"), []byte("a")},
		ReadSet:    nil,
		WriteSet:   nil,
		Payload:    []byte("ok"),
		ErrMsg:     modules.ContractError{},
	}
	req.AddMessage(modules.NewMessage(modules.APP_CONTRACT_INVOKE, result))
	return req
}

// create txs
func createTxs() []*modules.Transaction {
	txs := make([]*modules.Transaction, 0)
	hash0 := common.BytesToHash([]byte("0"))
	hash1 := common.BytesToHash([]byte("1"))
	txA := newTestPaymentTx(hash0)
	log.Printf("txA:%s\n", txA.Hash().String())
	txB := newTestPaymentTx(txA.Hash())
	log.Printf("txB:%s\n", txB.Hash().String())
	txC := newTestPaymentTx(txB.Hash())
	log.Printf("txC:%s\n", txC.Hash().String())
	txD := newTestPaymentTx(txC.Hash())
	log.Printf("txD:%s\n", txD.Hash().String())
	txX := newTestPaymentTx(hash1)
	log.Printf("txX:%s\n", txX.Hash().String())
	txY := newTestPaymentTx(txX.Hash())
	log.Printf("txY:%s\n", txY.Hash().String())
	txs = append(txs, txA, txB, txC, txD, txX, txY)

	return txs
}

// Tests that if the transaction count belonging to multiple accounts go above
// some hard threshold, if they are under the minimum guaranteed slot count then
// the transactions are still kept.
func TestTransactionAddingTxs(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mdag := mock.NewMockIDag(mockCtrl)
	mdag.EXPECT().GetNewestUnit(gomock.Any()).DoAndReturn(func(asset modules.AssetId) (common.Hash, *modules.ChainIndex, error) {
		return Hash("hash"), &modules.ChainIndex{asset, 0}, nil
	}).AnyTimes()
	mdag.EXPECT().GetUtxoEntry(gomock.Any()).DoAndReturn(func(outpoint *modules.OutPoint) (*modules.Utxo, error) {
		if outpoint.TxHash == common.BytesToHash([]byte("0")) {
			return &modules.Utxo{Amount: 123}, nil
		}
		return nil, tp2.ErrNotFound
	}).AnyTimes()
	mdag.EXPECT().IsTransactionExist(gomock.Any()).DoAndReturn(func(hash common.Hash) (bool, error) {
		return false, nil
	}).AnyTimes()
	mdag.EXPECT().GetTxHashByReqId(gomock.Any()).DoAndReturn(func(hash common.Hash) (common.Hash, error) {
		return common.Hash{}, tp2.ErrNotFound
	}).AnyTimes()
	pool := mockTxPool1(mdag)
	defer pool.Stop()
	var pending_cache, queue_cache, all, origin int
	txs := createTxs()
	origin = len(txs)
	tx := txs[origin-1]

	for _, tx := range txs {
		t.Logf("addLocals tx:%s, error:%v", tx.Hash().String(), pool.AddLocal(tx))
	}
	//  test GetSortedTxs{}
	defer func(p *tp1.TxPool) {
		sortedtxs, _ := p.GetSortedTxs()
		total := 0
		for _, tx := range sortedtxs {
			total += tx.Tx.SerializeSize()
		}

		all = len(sortedtxs)
		for i := 0; i < all-1; i++ {
			txpl := sortedtxs[i].Priority_lvl
			if txpl < sortedtxs[i+1].Priority_lvl {
				t.Error("sorted failed.", i, txpl)
			}
		}

		poolTxs := p.AllTxpoolTxs()
		for _, tx := range poolTxs {
			if tx.Pending {
				pending_cache++ //4
			} else {
				queue_cache++ // 2
			}
		}
		//  add tx : failed , and discared the tx.
		err := p.AddLocal(tx)
		assert.NotNil(t, err)
		err1 := p.DeleteTxByHash(tx.Hash())
		if err1 != nil {
			t.Error("DeleteTxByHash failed ", "error", err1)
		}
		err2 := p.AddLocal(tx)
		if err2 == nil {
			t.Log("addtx again info success")
		} else {
			t.Error("test added tx failed.", "error", err2)
		}
		t.Logf("data:%d,%d,%d,%d,%d", origin, all, p.AllLength(), pending_cache, queue_cache)
	}(pool)
}

func TestGetProscerTx(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mdag := mock.NewMockIDag(mockCtrl)
	mdag.EXPECT().GetNewestUnit(gomock.Any()).DoAndReturn(func(asset modules.AssetId) (common.Hash, *modules.ChainIndex, error) {
		return Hash("hash"), &modules.ChainIndex{asset, 0}, nil
	}).AnyTimes()
	mdag.EXPECT().GetUtxoEntry(gomock.Any()).DoAndReturn(func(outpoint *modules.OutPoint) (*modules.Utxo, error) {
		if outpoint.TxHash == common.BytesToHash([]byte("0")) {
			return &modules.Utxo{Amount: 10}, nil
		}
		return nil, tp2.ErrNotFound
	}).AnyTimes()
	mdag.EXPECT().IsTransactionExist(gomock.Any()).DoAndReturn(func(hash common.Hash) (bool, error) {
		return false, nil
	}).AnyTimes()
	mdag.EXPECT().GetTxHashByReqId(gomock.Any()).DoAndReturn(func(hash common.Hash) (common.Hash, error) {
		return common.Hash{}, tp2.ErrNotFound
	}).AnyTimes()
	pool := mockTxPool1(mdag)
	defer pool.Stop()

	hash0 := common.BytesToHash([]byte("0"))
	txA := newTestPaymentTx(hash0)
	t.Logf("Tx A:%s", txA.Hash().String())
	txB := newCcInvokeRequest(txA.Hash())
	t.Logf("Tx B:%s", txB.Hash().String())
	data0, _ := json.Marshal(txB)
	t.Logf("Tx hash:%s,  B:%s, ", txB.Hash().String(), string(data0))
	txC := newCcInvokeFullTx(txB.Hash())
	t.Logf("Tx C:%s", txC.Hash().String())
	t.Logf("Tx C req:%s", txC.RequestHash().String())

	data, _ := json.Marshal(txC)
	t.Logf("Tx hash:%s,  C:%s, ", txC.Hash().String(), string(data))

	txD := newTestPaymentTx(txC.RequestHash()) //交易D是基于TxC在Request的时候的UTXO产生的
	t.Logf("Tx D:%s", txD.Hash().String())
	data1, _ := json.Marshal(txD)
	t.Logf("Tx hash:%s,  D:%s, ", txD.Hash().String(), string(data1))

	txs := make([]*modules.Transaction, 0)
	txs = append(txs, txD, txB, txA, txC)
	//errs := pool.AddLocals(txs)
	for _, tx := range txs {
		t.Logf("addLocals tx:%s, error:%v", tx.Hash().String(), pool.AddLocal(tx))
	}
	defer func(p *tp1.TxPool) {
		count := p.Count()
		assert.Equal(t, 4, count)
		sortedTxs, _ := pool.GetSortedTxs()
		for index, tx := range sortedTxs {
			t.Logf("index:%d, hash:%s", index, tx.Tx.Hash().String())
		}
		if len(sortedTxs) == 4 {
			assert.Equal(t, txA.Hash().String(), sortedTxs[0].Tx.Hash().String())
			assert.Equal(t, txB.Hash().String(), sortedTxs[1].Tx.Hash().String())
			assert.Equal(t, txC.Hash().String(), sortedTxs[2].Tx.Hash().String())
			assert.Equal(t, txD.Hash().String(), sortedTxs[3].Tx.Hash().String())
		}
	}(pool)
}
