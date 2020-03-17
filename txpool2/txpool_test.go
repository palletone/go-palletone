/*
 *  This file is part of go-palletone.
 *  go-palletone is free software: you can redistribute it and/or modify
 *  it under the terms of the GNU General Public License as published by
 *  the Free Software Foundation, either version 3 of the License, or
 *  (at your option) any later version.
 *  go-palletone is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *  GNU General Public License for more details.
 *  You should have received a copy of the GNU General Public License
 *  along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
 *
 *  @author PalletOne core developer <dev@pallet.one>
 *  @date 2018-2020
 */

package txpool2

import (
	"encoding/hex"
	"strings"
	"testing"

	"github.com/coocood/freecache"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/golang/mock/gomock"
	"github.com/palletone/go-palletone/contracts/syscontract"
	"github.com/palletone/go-palletone/dag/mock"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/tokenengine"
	"github.com/palletone/go-palletone/txspool"
	"github.com/palletone/go-palletone/validator"
	"github.com/stretchr/testify/assert"
)

//func TestTxPool_Instance(t *testing.T){
//	Instance=NewTxPool(nil,nil,nil)
//}

func TestTxPool_GetSortTxs(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mdag := mock.NewMockIDag(mockCtrl)
	mdag.EXPECT().GetUtxoEntry(gomock.Any()).DoAndReturn(func(outpoint *modules.OutPoint) (*modules.Utxo, error) {
		if outpoint.TxHash == Hash("Dag") {
			return &modules.Utxo{Amount: 123}, nil
		}
		return nil, ErrNotFound
	}).AnyTimes()
	pool := mockTxPool(mdag)

	txA := mockPaymentTx(Hash("Dag"), 0, 0)
	t.Logf("Tx A:%s", txA.Hash().String())
	txB := mockPaymentTx(txA.Hash(), 0, 0)
	t.Logf("Tx B:%s", txB.Hash().String())
	txC := mockPaymentTx(txB.Hash(), 0, 0)
	t.Logf("Tx C:%s", txC.Hash().String())
	pool.AddLocal(txA)
	result := printTxPoolSortTxs(pool)
	t.Log("Add TxA", result)
	pool.AddLocal(txC)
	result = printTxPoolSortTxs(pool)
	t.Log("Add Tx A,C", result)

	pool.AddLocal(txB)
	result = printTxPoolSortTxs(pool)
	t.Log("Add Tx A,C,B", result)
}
func mockTxPool(mdag txspool.IDag) *TxPool {
	val := &mockValidator{query: mdag.GetUtxoEntry}
	return NewTxPool4DI(txspool.DefaultTxPoolConfig, freecache.NewCache(10000), mdag, tokenengine.Instance, val)
}
func printTxPoolSortTxs(pool *TxPool) string {
	sortedTx := ""
	pool.GetSortedTxs(func(tx *txspool.TxPoolTransaction) (getNext bool, err error) {
		sortedTx += tx.TxHash.String() + ";"
		return true, nil
	})
	return sortedTx
}

type mockValidator struct {
	query modules.QueryUtxoFunc
}

func (v *mockValidator) ValidateTx(tx *modules.Transaction, isFullTx bool) ([]*modules.Addition, validator.ValidationCode, error) {
	_, err := v.query(tx.GetSpendOutpoints()[0])
	if err != nil {
		return nil, validator.TxValidationCode_ORPHAN, nil
	}
	return []*modules.Addition{}, validator.TxValidationCode_VALID, nil
}
func (v *mockValidator) SetUtxoQuery(query validator.IUtxoQuery) {
	v.query = query.GetUtxoEntry
}
func BenchmarkTxPool_AddLocal(b *testing.B) {
	mockCtrl := gomock.NewController(b)
	defer mockCtrl.Finish()
	mdag := mock.NewMockIDag(mockCtrl)
	mdag.EXPECT().GetUtxoEntry(gomock.Any()).DoAndReturn(func(outpoint *modules.OutPoint) (*modules.Utxo, error) {
		if outpoint.TxHash == Hash("Dag") {
			return &modules.Utxo{Amount: 123}, nil
		}
		return nil, ErrNotFound
	}).AnyTimes()
	pool := mockTxPool(mdag)

	txA := mockPaymentTx(Hash("Dag"), 0, 0)
	//b.Logf("Tx A:%s", txA.Hash().String())
	pool.AddLocal(txA)
	for i := 0; i < b.N; i++ {
		txA = mockPaymentTx(txA.Hash(), 0, 0)
		//b.Logf("Tx %d:%s",i, txA.Hash().String())
		pool.AddLocal(txA)
	}

	result := printTxPoolSortTxs(pool)
	b.Log("Add Txs", result)
}
func TestTxpoolByRealData(t *testing.T) {
	//真实的几条交易，从前到后是依赖关系
	txf28c := rlpDecodeTx("f90212f9020df8e480b8e1f8dff892f890b86a47304402205746376a5b71857e3c84ec933a884d1b9cc483a535d992a264c21a8fc977658e022035181eef0838df7ca72bdbe6e03294aeb071a60c9458054f86285e41bcb0a483012102b12b2b4dc41fd3a890a3ba1a5ece3ce963890aa4c7badea72f99482af7d4a35e80a00df89bb8b81b8b0257b5ab8dd7917301cc4dc41809294e80439d23d425e913908080f848f8468801634578573891809976a9149deddffbeb485b43dca97af6a6bf46477876396288ace2904000a6a0080000000000000000000000900000000000000000000000000000000080f9012480b90120f9011df893f891b86b483045022100abb33a3a51535aa7ec8fb72c8da02814926d5f9f44c771c098e13fbbd6960f9402207a9e6b8b7cc35d0772d1af44d9f3c387e04e5bd85c245effd2271fc2408cdcd1012102b12b2b4dc41fd3a890a3ba1a5ece3ce963890aa4c7badea72f99482af7d4a35e80a00df89bb8b81b8b0257b5ab8dd7917301cc4dc41809294e80439d23d425e913900101f885f83e079976a9146820d6eca8ec493be799ab9ab455a261887654ca88ace2904000aedb000cd890a37dd4633ea7f00f9000000000000000000000000000000000f84385174876e7889976a9149deddffbeb485b43dca97af6a6bf46477876396288ace2904000aedb000cd890a37dd4633ea7f00f9000000000000000000000000000000000808080")
	tx14f4 := rlpDecodeTx("f90212f9020df8e580b8e2f8e0f893f891b86b483045022100f83b8c726ad0a7d473cd4d0d2cb0208a821c55b0fc0075e38ed360b55676edb10220026aa72a8a07f76e417fa05661bcb54e80b63932409a642caa241541a794d4ea012102b12b2b4dc41fd3a890a3ba1a5ece3ce963890aa4c7badea72f99482af7d4a35e80a0f28c0c23b0db24c043be22dc84cca17f66a15b4b641d351651c9130f1446e30c8080f848f846880163457857294f409976a9149deddffbeb485b43dca97af6a6bf46477876396288ace2904000a6a0080000000000000000000000900000000000000000000000000000000080f9012380b9011ff9011cf892f890b86a473044022062adf148e1e8e3dec35479d1279fdb8a98fba5e79d546978b09544d48e932e9602207942e83c2de30b54bd31cb66153d3bc5f925968003cdb0a51d291a7e045ceccf012102b12b2b4dc41fd3a890a3ba1a5ece3ce963890aa4c7badea72f99482af7d4a35e80a0f28c0c23b0db24c043be22dc84cca17f66a15b4b641d351651c9130f1446e30c0101f885f83e059976a914516831dd03cc8929d933a494b363b63f5088823288ace2904000aedb000cd890a37dd4633ea7f00f9000000000000000000000000000000000f84385174876e7839976a9149deddffbeb485b43dca97af6a6bf46477876396288ace2904000aedb000cd890a37dd4633ea7f00f9000000000000000000000000000000000808080")
	txd380 := rlpDecodeTx("f90212f9020df8e580b8e2f8e0f893f891b86b483045022100ddad5e419b814c66ce56a3d2c9dbfdc3393ed9a36c3a04043c835b9953c83636022036d046c3cd43275204eaf273d67f05acc2d64a318aa9384080d70b53f9f7b5b3012102b12b2b4dc41fd3a890a3ba1a5ece3ce963890aa4c7badea72f99482af7d4a35e80a014f4afbbb5a540bf8171f7ca7c9a92bdf3d0f5c28ef9f6222610c8f5cc1338738080f848f8468801634578571a0d009976a9149deddffbeb485b43dca97af6a6bf46477876396288ace2904000a6a0080000000000000000000000900000000000000000000000000000000080f9012380b9011ff9011cf892f890b86a47304402204f46fc8020a4af500d79f96cc611582afe631dae2c3ac2bc44adc17bd5d0f431022054584e6e43e7312b8246f95300d95004e48ac56555833aebceeb2bc94c6453a2012102b12b2b4dc41fd3a890a3ba1a5ece3ce963890aa4c7badea72f99482af7d4a35e80a014f4afbbb5a540bf8171f7ca7c9a92bdf3d0f5c28ef9f6222610c8f5cc1338730101f885f83e039976a9144471b38799bd66590c5fab640927400e610c4eb188ace2904000aedb000cd890a37dd4633ea7f00f9000000000000000000000000000000000f84385174876e7809976a9149deddffbeb485b43dca97af6a6bf46477876396288ace2904000aedb000cd890a37dd4633ea7f00f9000000000000000000000000000000000808080")
	txb37c := rlpDecodeTx("f90212f9020df8e580b8e2f8e0f893f891b86b483045022100b50a9e60030cac07ed4c0cf27933bb0478231a8efb191ee1652919c02ef3aa3302202a0041514d7108c9712c15b9ee05cac8cdb5716054943f968608e2ae13fa7889012102b12b2b4dc41fd3a890a3ba1a5ece3ce963890aa4c7badea72f99482af7d4a35e80a0d380ea678eab9cdea2330861a7cbaa70d841f0714b2a5f93166521f95d1e8d438080f848f8468801634578570acac09976a9149deddffbeb485b43dca97af6a6bf46477876396288ace2904000a6a0080000000000000000000000900000000000000000000000000000000080f9012380b9011ff9011cf892f890b86a47304402205cd68632bdccbd723c793be47a40e40935b5fcda6d4d346447f02a64c471b21b02206c62f1dcf5d1fe128a34ce434d89711753e6d27de88d0b667d7eb4e9e0258e77012102b12b2b4dc41fd3a890a3ba1a5ece3ce963890aa4c7badea72f99482af7d4a35e80a0d380ea678eab9cdea2330861a7cbaa70d841f0714b2a5f93166521f95d1e8d430101f885f83e089976a9147a5a90d248c9cd2a998752dce32998e44d93900988ace2904000aedb000cd890a37dd4633ea7f00f9000000000000000000000000000000000f84385174876e7789976a9149deddffbeb485b43dca97af6a6bf46477876396288ace2904000aedb000cd890a37dd4633ea7f00f9000000000000000000000000000000000808080")
	tx686b := rlpDecodeTx("f90213f9020ef8e580b8e2f8e0f893f891b86b483045022100e68655aef37a37c0057f5b209b3a5e1cc54e1f416d2cbd78757027fe8739406102204b2c07de6b12a8d64ece8f74fc3b7dc6deb102c78b120322ab0ccce68f7c2c7a012102b12b2b4dc41fd3a890a3ba1a5ece3ce963890aa4c7badea72f99482af7d4a35e80a0b37c5fee9cd08652ccd048cd38acbf886ca7f7d5a86dadcbe714c5f897d196308080f848f846880163457856fb88809976a9149deddffbeb485b43dca97af6a6bf46477876396288ace2904000a6a0080000000000000000000000900000000000000000000000000000000080f9012480b90120f9011df893f891b86b483045022100b63b2587174c7f4ccc7f64e2f092b72bc12e5b8c54d5ba4eb77a37cf9566aa4802206a8a492d8340970346c2bcb06c5ea51bfcb9d8e815a59363ffa3724696b87818012102b12b2b4dc41fd3a890a3ba1a5ece3ce963890aa4c7badea72f99482af7d4a35e80a0b37c5fee9cd08652ccd048cd38acbf886ca7f7d5a86dadcbe714c5f897d196300101f885f83e0a9976a9141e4641bac70f81b0d06ce711c5704e42d9764e3788ace2904000aedb000cd890a37dd4633ea7f00f9000000000000000000000000000000000f84385174876e76e9976a9149deddffbeb485b43dca97af6a6bf46477876396288ace2904000aedb000cd890a37dd4633ea7f00f9000000000000000000000000000000000808080")

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mdag := mock.NewMockIDag(mockCtrl)
	mdag.EXPECT().GetUtxoEntry(gomock.Any()).DoAndReturn(func(outpoint *modules.OutPoint) (*modules.Utxo, error) {
		if outpoint.TxHash.String() == "0x0df89bb8b81b8b0257b5ab8dd7917301cc4dc41809294e80439d23d425e91390" {
			return &modules.Utxo{Amount: 123}, nil
		}
		return nil, ErrNotFound
	}).AnyTimes()
	pool := mockTxPool(mdag)

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
func rlpDecodeTx(str string) *modules.Transaction {
	data, _ := hex.DecodeString(str)
	tx := modules.Transaction{}
	rlp.DecodeBytes(data, &tx)
	return &tx
}

func TestTxPool_AddSysContractTx(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mdag := mock.NewMockIDag(mockCtrl)
	mdag.EXPECT().GetUtxoEntry(gomock.Any()).DoAndReturn(
		func(outpoint *modules.OutPoint) (*modules.Utxo, error) {
			if outpoint.TxHash == Hash("dag") {
				return &modules.Utxo{Amount: 123}, nil
			}
			return nil, ErrNotFound
		}).AnyTimes()
	pool := mockTxPool(mdag)

	req := mockContractInvokeRequest(Hash("dag"), 0, 0, syscontract.TestContractAddress.Bytes())
	err := pool.AddLocal(req)
	assert.Nil(t, err)
	fullTx := mockContractInvokeFullTx(Hash("dag"), 0, 0, syscontract.TestContractAddress.Bytes())
	err = pool.AddLocal(fullTx)
	assert.NotNil(t, err)
	req1 := mockContractInvokeRequest(Hash("new one"), 0, 0, syscontract.TestContractAddress.Bytes())
	err = pool.AddLocal(req1)
	assert.Nil(t, err)

	all, orphan := pool.Content()
	assert.Equal(t, 1, len(all))
	assert.Equal(t, 1, len(orphan))
}

func TestTxPool_AddUserContractTx(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mdag := mock.NewMockIDag(mockCtrl)
	mdag.EXPECT().GetUtxoEntry(gomock.Any()).DoAndReturn(
		func(outpoint *modules.OutPoint) (*modules.Utxo, error) {
			if outpoint.TxHash == Hash("dag") {
				return &modules.Utxo{Amount: 123}, nil
			}
			return nil, ErrNotFound
		}).AnyTimes()
	pool := mockTxPool(mdag)

	req := mockContractInvokeRequest(Hash("dag"), 0, 0, []byte("user contract"))
	err := pool.AddLocal(req)
	assert.Nil(t, err)
	fullTx := mockContractInvokeFullTx(Hash("dag"), 0, 0, []byte("user contract"))
	err = pool.AddLocal(fullTx)
	assert.Nil(t, err)

	req1 := mockContractInvokeRequest(Hash("new one"), 0, 0, []byte("user contract"))
	err = pool.AddLocal(req1)
	assert.Nil(t, err)
	fullTx1 := mockContractInvokeFullTx(Hash("new one"), 0, 0, []byte("user contract"))
	err = pool.AddLocal(fullTx1)
	assert.Nil(t, err)
	all, orphan := pool.Content()
	assert.Equal(t, 1, len(all))
	assert.Equal(t, 2, len(orphan))
}
