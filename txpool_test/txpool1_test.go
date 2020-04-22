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
	"encoding/hex"
	"log"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/rlp"
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
		count := p.AllLength()
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

func TestPool1AddContractInstallTx(t *testing.T) {
	data, _ := hex.DecodeString("f909a2f9099df8e580b8e2f8e0f893f891b86b483045022100a79955b54c5af36096d364c931ab4cd01c1a64dfb54fb2b15da0c0d3eb17dd3e0220062c04f84b0d1fe15f8631bbc69acfb1a4be6d9f701295af4b9a55e184bc62f90121020eb815124c063e6ed4750e94306ea53104de7afaefe440fe1a0cd6b98a81208f80a00f8c28444ed0a9f4fe86a8e12032d2037f58cbf34099a35c5f4f9d44b83858688001f848f8468801634575fd925a009976a914495bf804f5d46659cda9d9c986c64f529850fbf188ace290400082bb080000000000000000000000900000000000000000000000000000000080f88664b883f881866a757279303680b8426769746875622e636f6d2f70616c6c65746f6e652f676f2d70616c6c65746f6e652f636f6e7472616374732f6578616d706c652f676f2f636f6e7472616374706179876578616d706c658086474f4c414e47c0a3503137677457374d74373378646d5969547a7433616733464564705637355370553545f9074901b90745f90742a067bb24b761ff7aa10053c3ccbeedfbe97b919f604a291fda0e62a7c64fc45de280b9071a1f8b08000000000000ffec587d6fdbbad5efbfd6a7388f1e74905a5792ddaedbb29be23a6dd26bac373162777745535cd0d291cc452205f2c8a951e4bb0fa424db895f92be0cdb8010416c91e7fc78ceefbc90b2567198719a55d320964558b23c479202c34c3e5b3dc452906231e9103fb3a2cccdf272b2648b763ac8e4a3cd11f5a2e8e58b178fa2288afef4f2e66714f59e477fec35dfed78143defbde8bf7804d116ac1f3e2a4d4c3d8aa2dad9dd7277ad2f8d6f3eff4746f8c401f307003099710d29cf11b88692290299c27a1e048de0fa9c914d15226899d2155378000b5941cc04284cb826c5a7152170022692502a2864c2d34503c5092a91a0029a2110aa429b4dcdc3dbd3f7f016052a96c3a89ae63c86773c46a1119886d2cce81926306da18cd289b164dc580227b21209232e451790d30c15cc51692e053c6fb76930bb205583e331322e2890a551f5818905e48c56dabb7858b99b0017167e264b049a31329e5ef13c872942a531adf26e8332ad087e1b4e7e397b3f81c1e907f86d707e3e389d7cf82b5c719ac98a00e758a3f1a2cc392670c59462821620d306e4d7e3f3d7bf0c4e2783a3e1bbe1e403480527c3c9e9f1780c2767e73080d1e07c327cfdfedde01c46efcf4767e3e300608cc6386c30f6309edab829840489f15cb70c7c9015e899acf204666c8ea030463ec70418c4b25cdc1dcb0687e55264d6df9b0907304c4148ea8246849f6644e541185e5d5d0599a802a9b230af7174f8cada14b6e9fc047e6615cda48291453b130871edc11c7359a2829f129cff5cef154881af5abd8411423feafdf9593feafda5060c1da764f125cb100ac685e3f0a2948ac0733a2e8a58265c64e13fb514aed371d382cc8726154b31771da7e3dea7c7168514612e33f79ef26d4fd6335eb84ea79cc2bdd41486f3e275a33daaa661a924491d9688ea7e5b272c0b0b9954396aa340bc40d7f11d270c616c52145fcf1837a420340703683b0fab05fb5ca0205b9f0e2d4adc50d6a4aa98e08b736db1878213c42ccfa72cbe0485a5428d82b8c8eae21073195b3493760ce216c7e8dadeb69ca92bb160c2c493ae24b0389695200d03d3a5e0c8feb742a498d0292af8075482936976b28001908423a84a29ecbe97e8a49588c1237872cb0bdfdaed69aaa66022152c57c6544d878250a52c461fca69708eba94a6debe381da3108c2a1a1323f4dc922db4f916b95df8f869bab839e7fb4e4721554ad47b8cab3846ad3dc173bf668fe4ef2c49146add655a237559613cde6376c916b2badbf02e309569f8f8c9f43e916df8b1da180e0eadecc7e893d3a98d80dfcd64934bc1d8224ce4c0ac7956b4f7c9773aac68259b920a0624792dd06f0426f2128511f9438b36b00e5aac2ff5f703a8b8a0972f3c5690df05bb7400d692eb9670b638ab6a306f8d326877e842e43b9db4a060a4b8a05c78eec8f20464965d78ba146ddcf17c780aaec997960623b4c2be23727be363a3dffb61b9d5db925bbd3b73eb6e0bfb3fccc2fe160bfbdf67a1acc8c27c77a22f193b5a102ef3b59a066f719368dfe9f0147214de0d351f0e0f2132681d7398187803e37eb9708f9592eac23db870cda12b1013db0b21437a83a5d49c0649a2ec499d61738349ea8536f32edc6bd7e9dca0caa27aed5ebed3b9763ab9cc823738adb274cd6263dd013cd66e176aff6f19ee3b2b06fa7733d0dfc240ffbf9a81fe6e06fa4b06cc51188c73c4d2ebc113a81f31962231ab3b927a4b4ebb35da43ebfcf7b4ce29cb9988bfbfe4b74547a1ae72ea022ab59effd6afa366df2537d654530746faff0ecdb34dfdcd0445a59a6f7e9da30923b6dcc5a46ff02b537ac672afb6e05b71b7316af63294ee66d464fad77779a3b5c9d4e7092f50132b4aaf177dab1b3426b59ef64392cce382bcb59ad446f65651dacbecb2200d8a2dc56dacdc94d9c7ced05e11bf9e1f8358bfbeda3c5c23e9a459198864c4142b905069af26abd5329dd4ad4f38779d360ad6ae7635b4e56cab72735e6de8af5d3d6ac277e9f677e8f6f7ebb6c7f256cb5707f69af91b28a6ca6ea99ba93dfb367de196d27ab7d8b79f4d9d9baacbb2d8e1e766649ab06c97cff0b67c865bc2b8591eee50cc59ce93e66565852a5881011c7f2e31b62f52176e2d71e1bafb53babcc7abc1ed7cde28b6e9dad13775b754235b1360db04e23581f84e0476d75571b7bf2dd1df735c34f71d1b2978d5dc72daf362cb6da93953fc5b1dafb579d9e7eb986f3d76daf7fc419ed798dfda4f75dd4e5dd7e9980bd6252eba3037338a890ca1dede8219c9a787600efa71694efad473ff868b83c7bafb77965778f0f8ffe76eb741d87de63441d34d73dd139a65517f556f6d7f3f39aad22eb40ff602b985bee11bcf6f3936eb379c7bfcf999bd15ee44f46fdc289761dbe37523b372bc605c78beb1bb0db0d522a6c81378e5dd26656b989777afd4ab2fd3a00d80a9fb5a7ff50b4973d145a56c84ae9dfff40fd50fe3613c8c87f1307ee8f857000000ffff0623672b001e0000c28080f8e005b8ddf8dbf8d9f86aa1020eb815124c063e6ed4750e94306ea53104de7afaefe440fe1a0cd6b98a81208fb8463044022013f14212e75aec9f38aadaa607404a244245a140cacd1ce87cbe70a5871ecf6e0220356f1d5780c4d78fac288f6b4dcd65e12a66d3ad31b657e435c48b95b124cdd0f86ba103bdc5747794aae58742cc837027b4d8fd757631e46a9da3b6bac081b19897eb38b8473045022100f883a2898f92bc4cc909077f9e1376322adf63050e686dff3190b0b27e2d3af5022038e9dd13cf803b5ac65266b1ce71268a2b7c07acc22650daf29f9dc47723486b8080")
	installTx := &modules.Transaction{}
	rlp.DecodeBytes(data, installTx)
	t.Log(installTx.IsSystemContract())
	hash := common.HexToHash("0x0f8c28444ed0a9f4fe86a8e12032d2037f58cbf34099a35c5f4f9d44b8385868")
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mdag := mockdag1(t, mockCtrl, hash)
	pool := mockTxPool1(mdag)
	err := pool.AddLocal(installTx)
	t.Log(err)
	assert.Nil(t, err)
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
