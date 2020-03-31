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
	"log"
	"strings"
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
	txB := newTestPaymentTx(txA.Hash())
	txC := newTestPaymentTx(txB.Hash())
	txD := newTestPaymentTx(txC.Hash())
	txX := newTestPaymentTx(hash1)
	txY := newTestPaymentTx(txX.Hash())
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

	mdag := mockdag1(t, mockCtrl, common.BytesToHash([]byte("0")))
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
	hash0 := common.HexToHash("0x0df89bb8b81b8b0257b5ab8dd7917301cc4dc41809294e80439d23d425e91390")
	mdag := mockdag1(t, mockCtrl, hash0)
	pool := mockTxPool1(mdag)
	defer pool.Stop()
	//真实的几条交易，从前到后是依赖关系
	txf28c := rlpDecodeTx("f90212f9020df8e480b8e1f8dff892f890b86a47304402205746376a5b71857e3c84ec933a884d1b9cc483a535d992a264c21a8fc977658e022035181eef0838df7ca72bdbe6e03294aeb071a60c9458054f86285e41bcb0a483012102b12b2b4dc41fd3a890a3ba1a5ece3ce963890aa4c7badea72f99482af7d4a35e80a00df89bb8b81b8b0257b5ab8dd7917301cc4dc41809294e80439d23d425e913908080f848f8468801634578573891809976a9149deddffbeb485b43dca97af6a6bf46477876396288ace2904000a6a0080000000000000000000000900000000000000000000000000000000080f9012480b90120f9011df893f891b86b483045022100abb33a3a51535aa7ec8fb72c8da02814926d5f9f44c771c098e13fbbd6960f9402207a9e6b8b7cc35d0772d1af44d9f3c387e04e5bd85c245effd2271fc2408cdcd1012102b12b2b4dc41fd3a890a3ba1a5ece3ce963890aa4c7badea72f99482af7d4a35e80a00df89bb8b81b8b0257b5ab8dd7917301cc4dc41809294e80439d23d425e913900101f885f83e079976a9146820d6eca8ec493be799ab9ab455a261887654ca88ace2904000aedb000cd890a37dd4633ea7f00f9000000000000000000000000000000000f84385174876e7889976a9149deddffbeb485b43dca97af6a6bf46477876396288ace2904000aedb000cd890a37dd4633ea7f00f9000000000000000000000000000000000808080")
	tx14f4 := rlpDecodeTx("f90212f9020df8e580b8e2f8e0f893f891b86b483045022100f83b8c726ad0a7d473cd4d0d2cb0208a821c55b0fc0075e38ed360b55676edb10220026aa72a8a07f76e417fa05661bcb54e80b63932409a642caa241541a794d4ea012102b12b2b4dc41fd3a890a3ba1a5ece3ce963890aa4c7badea72f99482af7d4a35e80a0f28c0c23b0db24c043be22dc84cca17f66a15b4b641d351651c9130f1446e30c8080f848f846880163457857294f409976a9149deddffbeb485b43dca97af6a6bf46477876396288ace2904000a6a0080000000000000000000000900000000000000000000000000000000080f9012380b9011ff9011cf892f890b86a473044022062adf148e1e8e3dec35479d1279fdb8a98fba5e79d546978b09544d48e932e9602207942e83c2de30b54bd31cb66153d3bc5f925968003cdb0a51d291a7e045ceccf012102b12b2b4dc41fd3a890a3ba1a5ece3ce963890aa4c7badea72f99482af7d4a35e80a0f28c0c23b0db24c043be22dc84cca17f66a15b4b641d351651c9130f1446e30c0101f885f83e059976a914516831dd03cc8929d933a494b363b63f5088823288ace2904000aedb000cd890a37dd4633ea7f00f9000000000000000000000000000000000f84385174876e7839976a9149deddffbeb485b43dca97af6a6bf46477876396288ace2904000aedb000cd890a37dd4633ea7f00f9000000000000000000000000000000000808080")
	txd380 := rlpDecodeTx("f90212f9020df8e580b8e2f8e0f893f891b86b483045022100ddad5e419b814c66ce56a3d2c9dbfdc3393ed9a36c3a04043c835b9953c83636022036d046c3cd43275204eaf273d67f05acc2d64a318aa9384080d70b53f9f7b5b3012102b12b2b4dc41fd3a890a3ba1a5ece3ce963890aa4c7badea72f99482af7d4a35e80a014f4afbbb5a540bf8171f7ca7c9a92bdf3d0f5c28ef9f6222610c8f5cc1338738080f848f8468801634578571a0d009976a9149deddffbeb485b43dca97af6a6bf46477876396288ace2904000a6a0080000000000000000000000900000000000000000000000000000000080f9012380b9011ff9011cf892f890b86a47304402204f46fc8020a4af500d79f96cc611582afe631dae2c3ac2bc44adc17bd5d0f431022054584e6e43e7312b8246f95300d95004e48ac56555833aebceeb2bc94c6453a2012102b12b2b4dc41fd3a890a3ba1a5ece3ce963890aa4c7badea72f99482af7d4a35e80a014f4afbbb5a540bf8171f7ca7c9a92bdf3d0f5c28ef9f6222610c8f5cc1338730101f885f83e039976a9144471b38799bd66590c5fab640927400e610c4eb188ace2904000aedb000cd890a37dd4633ea7f00f9000000000000000000000000000000000f84385174876e7809976a9149deddffbeb485b43dca97af6a6bf46477876396288ace2904000aedb000cd890a37dd4633ea7f00f9000000000000000000000000000000000808080")
	txb37c := rlpDecodeTx("f90212f9020df8e580b8e2f8e0f893f891b86b483045022100b50a9e60030cac07ed4c0cf27933bb0478231a8efb191ee1652919c02ef3aa3302202a0041514d7108c9712c15b9ee05cac8cdb5716054943f968608e2ae13fa7889012102b12b2b4dc41fd3a890a3ba1a5ece3ce963890aa4c7badea72f99482af7d4a35e80a0d380ea678eab9cdea2330861a7cbaa70d841f0714b2a5f93166521f95d1e8d438080f848f8468801634578570acac09976a9149deddffbeb485b43dca97af6a6bf46477876396288ace2904000a6a0080000000000000000000000900000000000000000000000000000000080f9012380b9011ff9011cf892f890b86a47304402205cd68632bdccbd723c793be47a40e40935b5fcda6d4d346447f02a64c471b21b02206c62f1dcf5d1fe128a34ce434d89711753e6d27de88d0b667d7eb4e9e0258e77012102b12b2b4dc41fd3a890a3ba1a5ece3ce963890aa4c7badea72f99482af7d4a35e80a0d380ea678eab9cdea2330861a7cbaa70d841f0714b2a5f93166521f95d1e8d430101f885f83e089976a9147a5a90d248c9cd2a998752dce32998e44d93900988ace2904000aedb000cd890a37dd4633ea7f00f9000000000000000000000000000000000f84385174876e7789976a9149deddffbeb485b43dca97af6a6bf46477876396288ace2904000aedb000cd890a37dd4633ea7f00f9000000000000000000000000000000000808080")
	tx686b := rlpDecodeTx("f90213f9020ef8e580b8e2f8e0f893f891b86b483045022100e68655aef37a37c0057f5b209b3a5e1cc54e1f416d2cbd78757027fe8739406102204b2c07de6b12a8d64ece8f74fc3b7dc6deb102c78b120322ab0ccce68f7c2c7a012102b12b2b4dc41fd3a890a3ba1a5ece3ce963890aa4c7badea72f99482af7d4a35e80a0b37c5fee9cd08652ccd048cd38acbf886ca7f7d5a86dadcbe714c5f897d196308080f848f846880163457856fb88809976a9149deddffbeb485b43dca97af6a6bf46477876396288ace2904000a6a0080000000000000000000000900000000000000000000000000000000080f9012480b90120f9011df893f891b86b483045022100b63b2587174c7f4ccc7f64e2f092b72bc12e5b8c54d5ba4eb77a37cf9566aa4802206a8a492d8340970346c2bcb06c5ea51bfcb9d8e815a59363ffa3724696b87818012102b12b2b4dc41fd3a890a3ba1a5ece3ce963890aa4c7badea72f99482af7d4a35e80a0b37c5fee9cd08652ccd048cd38acbf886ca7f7d5a86dadcbe714c5f897d196300101f885f83e0a9976a9141e4641bac70f81b0d06ce711c5704e42d9764e3788ace2904000aedb000cd890a37dd4633ea7f00f9000000000000000000000000000000000f84385174876e76e9976a9149deddffbeb485b43dca97af6a6bf46477876396288ace2904000aedb000cd890a37dd4633ea7f00f9000000000000000000000000000000000808080")

	txs := []*modules.Transaction{tx14f4, txd380, txb37c, tx686b, txf28c}
	for _, tx := range txs {
		t.Logf("addLocals tx:%s, error:%v", tx.Hash().String(), pool.AddLocal(tx))
	}
	defer func(p *tp1.TxPool) {
		count := p.Count()
		assert.Equal(t, 5, count)
		sortedTxs, _ := pool.GetSortedTxs()
		for index, tx := range sortedTxs {
			t.Logf("index:%d, hash:%s", index, tx.Tx.Hash().String())
		}
		if len(sortedTxs) == 5 {
			assert.Equal(t, txf28c.Hash().String(), sortedTxs[0].Tx.Hash().String())
			assert.Equal(t, tx14f4.Hash().String(), sortedTxs[1].Tx.Hash().String())
			assert.Equal(t, txd380.Hash().String(), sortedTxs[2].Tx.Hash().String())
			assert.Equal(t, txb37c.Hash().String(), sortedTxs[3].Tx.Hash().String())
			assert.Equal(t, tx686b.Hash().String(), sortedTxs[4].Tx.Hash().String())
		}
	}(pool)
}
func BenchmarkTxPool1_AddLocal(b *testing.B) {
	hash := common.BytesToHash([]byte("0"))
	mockCtrl := gomock.NewController(b)
	defer mockCtrl.Finish()
	mdag := mock.NewMockIDag(mockCtrl)
	mdag.EXPECT().GetUtxoEntry(gomock.Any()).DoAndReturn(func(outpoint *modules.OutPoint) (*modules.Utxo, error) {
		if outpoint.TxHash == hash {
			return &modules.Utxo{Amount: 123}, nil
		}
		return nil, tp2.ErrNotFound
	}).AnyTimes()

	mdag.EXPECT().GetNewestUnit(gomock.Any()).DoAndReturn(func(asset modules.AssetId) (common.Hash, *modules.ChainIndex, error) {
		return Hash("hash"), &modules.ChainIndex{asset, 0}, nil
	}).AnyTimes()
	mdag.EXPECT().IsTransactionExist(gomock.Any()).DoAndReturn(func(hash common.Hash) (bool, error) {
		return false, nil
	}).AnyTimes()
	mdag.EXPECT().GetTxHashByReqId(gomock.Any()).DoAndReturn(func(hash common.Hash) (common.Hash, error) {
		return common.Hash{}, tp2.ErrNotFound
	}).AnyTimes()

	pool := mockTxPool1(mdag)
	txA := mockPaymentTx(hash, 0, 0)
	pool.AddLocal(txA)
	for i := 0; i < b.N; i++ {
		txA = mockPaymentTx(txA.Hash(), 0, 0)
		pool.AddLocal(txA)
	}

	result := printTxPoolSortTxs(pool)
	b.Log("Add Txs", result)
}

func TestTxpool1ByRealData(t *testing.T) {
	//真实的几条交易，从前到后是依赖关系
	txf28c := rlpDecodeTx("f90212f9020df8e480b8e1f8dff892f890b86a47304402205746376a5b71857e3c84ec933a884d1b9cc483a535d992a264c21a8fc977658e022035181eef0838df7ca72bdbe6e03294aeb071a60c9458054f86285e41bcb0a483012102b12b2b4dc41fd3a890a3ba1a5ece3ce963890aa4c7badea72f99482af7d4a35e80a00df89bb8b81b8b0257b5ab8dd7917301cc4dc41809294e80439d23d425e913908080f848f8468801634578573891809976a9149deddffbeb485b43dca97af6a6bf46477876396288ace2904000a6a0080000000000000000000000900000000000000000000000000000000080f9012480b90120f9011df893f891b86b483045022100abb33a3a51535aa7ec8fb72c8da02814926d5f9f44c771c098e13fbbd6960f9402207a9e6b8b7cc35d0772d1af44d9f3c387e04e5bd85c245effd2271fc2408cdcd1012102b12b2b4dc41fd3a890a3ba1a5ece3ce963890aa4c7badea72f99482af7d4a35e80a00df89bb8b81b8b0257b5ab8dd7917301cc4dc41809294e80439d23d425e913900101f885f83e079976a9146820d6eca8ec493be799ab9ab455a261887654ca88ace2904000aedb000cd890a37dd4633ea7f00f9000000000000000000000000000000000f84385174876e7889976a9149deddffbeb485b43dca97af6a6bf46477876396288ace2904000aedb000cd890a37dd4633ea7f00f9000000000000000000000000000000000808080")
	tx14f4 := rlpDecodeTx("f90212f9020df8e580b8e2f8e0f893f891b86b483045022100f83b8c726ad0a7d473cd4d0d2cb0208a821c55b0fc0075e38ed360b55676edb10220026aa72a8a07f76e417fa05661bcb54e80b63932409a642caa241541a794d4ea012102b12b2b4dc41fd3a890a3ba1a5ece3ce963890aa4c7badea72f99482af7d4a35e80a0f28c0c23b0db24c043be22dc84cca17f66a15b4b641d351651c9130f1446e30c8080f848f846880163457857294f409976a9149deddffbeb485b43dca97af6a6bf46477876396288ace2904000a6a0080000000000000000000000900000000000000000000000000000000080f9012380b9011ff9011cf892f890b86a473044022062adf148e1e8e3dec35479d1279fdb8a98fba5e79d546978b09544d48e932e9602207942e83c2de30b54bd31cb66153d3bc5f925968003cdb0a51d291a7e045ceccf012102b12b2b4dc41fd3a890a3ba1a5ece3ce963890aa4c7badea72f99482af7d4a35e80a0f28c0c23b0db24c043be22dc84cca17f66a15b4b641d351651c9130f1446e30c0101f885f83e059976a914516831dd03cc8929d933a494b363b63f5088823288ace2904000aedb000cd890a37dd4633ea7f00f9000000000000000000000000000000000f84385174876e7839976a9149deddffbeb485b43dca97af6a6bf46477876396288ace2904000aedb000cd890a37dd4633ea7f00f9000000000000000000000000000000000808080")
	txd380 := rlpDecodeTx("f90212f9020df8e580b8e2f8e0f893f891b86b483045022100ddad5e419b814c66ce56a3d2c9dbfdc3393ed9a36c3a04043c835b9953c83636022036d046c3cd43275204eaf273d67f05acc2d64a318aa9384080d70b53f9f7b5b3012102b12b2b4dc41fd3a890a3ba1a5ece3ce963890aa4c7badea72f99482af7d4a35e80a014f4afbbb5a540bf8171f7ca7c9a92bdf3d0f5c28ef9f6222610c8f5cc1338738080f848f8468801634578571a0d009976a9149deddffbeb485b43dca97af6a6bf46477876396288ace2904000a6a0080000000000000000000000900000000000000000000000000000000080f9012380b9011ff9011cf892f890b86a47304402204f46fc8020a4af500d79f96cc611582afe631dae2c3ac2bc44adc17bd5d0f431022054584e6e43e7312b8246f95300d95004e48ac56555833aebceeb2bc94c6453a2012102b12b2b4dc41fd3a890a3ba1a5ece3ce963890aa4c7badea72f99482af7d4a35e80a014f4afbbb5a540bf8171f7ca7c9a92bdf3d0f5c28ef9f6222610c8f5cc1338730101f885f83e039976a9144471b38799bd66590c5fab640927400e610c4eb188ace2904000aedb000cd890a37dd4633ea7f00f9000000000000000000000000000000000f84385174876e7809976a9149deddffbeb485b43dca97af6a6bf46477876396288ace2904000aedb000cd890a37dd4633ea7f00f9000000000000000000000000000000000808080")
	txb37c := rlpDecodeTx("f90212f9020df8e580b8e2f8e0f893f891b86b483045022100b50a9e60030cac07ed4c0cf27933bb0478231a8efb191ee1652919c02ef3aa3302202a0041514d7108c9712c15b9ee05cac8cdb5716054943f968608e2ae13fa7889012102b12b2b4dc41fd3a890a3ba1a5ece3ce963890aa4c7badea72f99482af7d4a35e80a0d380ea678eab9cdea2330861a7cbaa70d841f0714b2a5f93166521f95d1e8d438080f848f8468801634578570acac09976a9149deddffbeb485b43dca97af6a6bf46477876396288ace2904000a6a0080000000000000000000000900000000000000000000000000000000080f9012380b9011ff9011cf892f890b86a47304402205cd68632bdccbd723c793be47a40e40935b5fcda6d4d346447f02a64c471b21b02206c62f1dcf5d1fe128a34ce434d89711753e6d27de88d0b667d7eb4e9e0258e77012102b12b2b4dc41fd3a890a3ba1a5ece3ce963890aa4c7badea72f99482af7d4a35e80a0d380ea678eab9cdea2330861a7cbaa70d841f0714b2a5f93166521f95d1e8d430101f885f83e089976a9147a5a90d248c9cd2a998752dce32998e44d93900988ace2904000aedb000cd890a37dd4633ea7f00f9000000000000000000000000000000000f84385174876e7789976a9149deddffbeb485b43dca97af6a6bf46477876396288ace2904000aedb000cd890a37dd4633ea7f00f9000000000000000000000000000000000808080")
	tx686b := rlpDecodeTx("f90213f9020ef8e580b8e2f8e0f893f891b86b483045022100e68655aef37a37c0057f5b209b3a5e1cc54e1f416d2cbd78757027fe8739406102204b2c07de6b12a8d64ece8f74fc3b7dc6deb102c78b120322ab0ccce68f7c2c7a012102b12b2b4dc41fd3a890a3ba1a5ece3ce963890aa4c7badea72f99482af7d4a35e80a0b37c5fee9cd08652ccd048cd38acbf886ca7f7d5a86dadcbe714c5f897d196308080f848f846880163457856fb88809976a9149deddffbeb485b43dca97af6a6bf46477876396288ace2904000a6a0080000000000000000000000900000000000000000000000000000000080f9012480b90120f9011df893f891b86b483045022100b63b2587174c7f4ccc7f64e2f092b72bc12e5b8c54d5ba4eb77a37cf9566aa4802206a8a492d8340970346c2bcb06c5ea51bfcb9d8e815a59363ffa3724696b87818012102b12b2b4dc41fd3a890a3ba1a5ece3ce963890aa4c7badea72f99482af7d4a35e80a0b37c5fee9cd08652ccd048cd38acbf886ca7f7d5a86dadcbe714c5f897d196300101f885f83e0a9976a9141e4641bac70f81b0d06ce711c5704e42d9764e3788ace2904000aedb000cd890a37dd4633ea7f00f9000000000000000000000000000000000f84385174876e76e9976a9149deddffbeb485b43dca97af6a6bf46477876396288ace2904000aedb000cd890a37dd4633ea7f00f9000000000000000000000000000000000808080")

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mdag := mockdag1(t, mockCtrl, common.HexToHash("0x0df89bb8b81b8b0257b5ab8dd7917301cc4dc41809294e80439d23d425e91390"))
	pool := mockTxPool1(mdag)

	pool.AddLocal(tx14f4)
	pool.AddLocal(txd380)
	pool.AddLocal(txb37c)
	pool.AddLocal(tx686b)
	//先添加后面的，最后添加开头的。
	pool.AddLocal(txf28c)
	result := printTxPoolSortTxs(pool)
	t.Log("Real sort result:", result)
	expect := txf28c.Hash().String() + ";" + tx14f4.Hash().String() + ";" + txd380.Hash().String() + ";" + txb37c.Hash().String() + ";" + tx686b.Hash().String()
	assert.True(t, strings.Contains(result, expect))
}

//先添加用户合约Request，然后是连续交易的转账，然后又是用户合约Request
func TestTxPool1_AddUserContractAndTransferTx(t *testing.T) {
	addr, _ := common.StringToAddress("P1HXNZReTByQHgWQNGMXotMyTkMG9XeEQfX")
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mdag := mockdag1(t, mockCtrl, Hash("dag"))
	pool := mockTxPool1(mdag)

	reqA := mockContractInvokeRequest(Hash("dag"), 0, 0, []byte("user contract"))
	log.Println("reqA:", reqA.Hash().String())
	err := pool.AddLocal(reqA)
	assert.Nil(t, err)
	txB := mockPaymentTx(reqA.Hash(), 0, 0)
	log.Println("txB:", txB.Hash().String())
	err = pool.AddLocal(txB)
	assert.Nil(t, err)
	reqC := mockContractInvokeRequest(txB.Hash(), 0, 0, []byte("user contract"))
	log.Println("reqC:", reqC.Hash().String())
	err = pool.AddLocal(reqC)
	assert.Nil(t, err)
	txs, _ := pool.GetUnpackedTxsByAddr(addr)
	assert.Equal(t, 3, len(txs))
	sortedTx, err := pool.GetSortedTxs()
	assert.Equal(t, 0, len(sortedTx))

	//pool.AddLocals([]*modules.Transaction{reqA, txB, reqC})
	fullTxA := mockContractInvokeFullTx(Hash("dag"), 0, 0, []byte("user contract"))
	log.Println("fullA:", fullTxA.Hash().String())
	err = pool.AddLocal(fullTxA)
	assert.Nil(t, err)
	sortedTx, err = pool.GetSortedTxs()
	assert.Equal(t, 2, len(sortedTx))
	txs, _ = pool.GetUnpackedTxsByAddr(addr)
	assert.Equal(t, 3, len(txs))

	//第二种情形，ReqA，B，B先完成FullTx
	log.Println("-------------------")
	pl := mockTxPool1(mdag)
	pl.AddLocal(reqA)
	reqB := mockContractInvokeRequest(reqA.Hash(), 0, 0, []byte("user contract"))
	log.Println("reqB:", reqB.Hash().String())
	pl.AddLocal(reqB)
	fullTxB := mockContractInvokeFullTx(reqA.Hash(), 0, 0, []byte("user contract"))
	log.Println("fullB:", fullTxB.Hash().String())
	err = pl.AddLocal(fullTxB)
	assert.Nil(t, err)
	sortedTx, _ = pl.GetSortedTxs()
	assert.Equal(t, 0, len(sortedTx))
	txs, _ = pl.GetUnpackedTxsByAddr(addr)
	assert.Equal(t, 2, len(txs))
	pl.AddLocal(fullTxA)
	sortedTx, err = pl.GetSortedTxs()
	assert.Equal(t, 2, len(sortedTx))
	txs, _ = pl.GetUnpackedTxsByAddr(addr)
	assert.Equal(t, 2, len(txs))
}
