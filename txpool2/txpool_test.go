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
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/coocood/freecache"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/golang/mock/gomock"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
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
	mdag.EXPECT().IsTransactionExist(gomock.Any()).Return(false, nil).AnyTimes()
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
	list, _ := pool.GetSortedTxs()
	for _, tx := range list {
		sortedTx += string(tx.TxHash.String()) + ";"
	}
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
	mdag.EXPECT().IsTransactionExist(gomock.Any()).Return(false, nil).AnyTimes()
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
	mdag.EXPECT().IsTransactionExist(gomock.Any()).Return(false, nil).AnyTimes()
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
	mdag.EXPECT().IsTransactionExist(gomock.Any()).Return(false, nil).AnyTimes()
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
	mdag.EXPECT().IsTransactionExist(gomock.Any()).Return(false, nil).AnyTimes()
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
func TestAddContractInstallTx(t *testing.T) {
	data, _ := hex.DecodeString("f909a2f9099df8e580b8e2f8e0f893f891b86b483045022100a79955b54c5af36096d364c931ab4cd01c1a64dfb54fb2b15da0c0d3eb17dd3e0220062c04f84b0d1fe15f8631bbc69acfb1a4be6d9f701295af4b9a55e184bc62f90121020eb815124c063e6ed4750e94306ea53104de7afaefe440fe1a0cd6b98a81208f80a00f8c28444ed0a9f4fe86a8e12032d2037f58cbf34099a35c5f4f9d44b83858688001f848f8468801634575fd925a009976a914495bf804f5d46659cda9d9c986c64f529850fbf188ace290400082bb080000000000000000000000900000000000000000000000000000000080f88664b883f881866a757279303680b8426769746875622e636f6d2f70616c6c65746f6e652f676f2d70616c6c65746f6e652f636f6e7472616374732f6578616d706c652f676f2f636f6e7472616374706179876578616d706c658086474f4c414e47c0a3503137677457374d74373378646d5969547a7433616733464564705637355370553545f9074901b90745f90742a067bb24b761ff7aa10053c3ccbeedfbe97b919f604a291fda0e62a7c64fc45de280b9071a1f8b08000000000000ffec587d6fdbbad5efbfd6a7388f1e74905a5792ddaedbb29be23a6dd26bac373162777745535cd0d291cc452205f2c8a951e4bb0fa424db895f92be0cdb8010416c91e7fc78ceefbc90b2567198719a55d320964558b23c479202c34c3e5b3dc452906231e9103fb3a2cccdf272b2648b763ac8e4a3cd11f5a2e8e58b178fa2288afef4f2e66714f59e477fec35dfed78143defbde8bf7804d116ac1f3e2a4d4c3d8aa2dad9dd7277ad2f8d6f3eff4746f8c401f307003099710d29cf11b88692290299c27a1e048de0fa9c914d15226899d2155378000b5941cc04284cb826c5a7152170022692502a2864c2d34503c5092a91a0029a2110aa429b4dcdc3dbd3f7f016052a96c3a89ae63c86773c46a1119886d2cce81926306da18cd289b164dc580227b21209232e451790d30c15cc51692e053c6fb76930bb205583e331322e2890a551f5818905e48c56dabb7858b99b0017167e264b049a31329e5ef13c872942a531adf26e8332ad087e1b4e7e397b3f81c1e907f86d707e3e389d7cf82b5c719ac98a00e758a3f1a2cc392670c59462821620d306e4d7e3f3d7bf0c4e2783a3e1bbe1e403480527c3c9e9f1780c2767e73080d1e07c327cfdfedde01c46efcf4767e3e300608cc6386c30f6309edab829840489f15cb70c7c9015e899acf204666c8ea030463ec70418c4b25cdc1dcb0687e55264d6df9b0907304c4148ea8246849f6644e541185e5d5d0599a802a9b230af7174f8cada14b6e9fc047e6615cda48291453b130871edc11c7359a2829f129cff5cef154881af5abd8411423feafdf9593feafda5060c1da764f125cb100ac685e3f0a2948ac0733a2e8a58265c64e13fb514aed371d382cc8726154b31771da7e3dea7c7168514612e33f79ef26d4fd6335eb84ea79cc2bdd41486f3e275a33daaa661a924491d9688ea7e5b272c0b0b9954396aa340bc40d7f11d270c616c52145fcf1837a420340703683b0fab05fb5ca0205b9f0e2d4adc50d6a4aa98e08b736db1878213c42ccfa72cbe0485a5428d82b8c8eae21073195b3493760ce216c7e8dadeb69ca92bb160c2c493ae24b0389695200d03d3a5e0c8feb742a498d0292af8075482936976b28001908423a84a29ecbe97e8a49588c1237872cb0bdfdaed69aaa66022152c57c6544d878250a52c461fca69708eba94a6debe381da3108c2a1a1323f4dc922db4f916b95df8f869bab839e7fb4e4721554ad47b8cab3846ad3dc173bf668fe4ef2c49146add655a237559613cde6376c916b2badbf02e309569f8f8c9f43e916df8b1da180e0eadecc7e893d3a98d80dfcd64934bc1d8224ce4c0ac7956b4f7c9773aac68259b920a0624792dd06f0426f2128511f9438b36b00e5aac2ff5f703a8b8a0972f3c5690df05bb7400d692eb9670b638ab6a306f8d326877e842e43b9db4a060a4b8a05c78eec8f20464965d78ba146ddcf17c780aaec997960623b4c2be23727be363a3dffb61b9d5db925bbd3b73eb6e0bfb3fccc2fe160bfbdf67a1acc8c27c77a22f193b5a102ef3b59a066f719368dfe9f0147214de0d351f0e0f2132681d7398187803e37eb9708f9592eac23db870cda12b1013db0b21437a83a5d49c0649a2ec499d61738349ea8536f32edc6bd7e9dca0caa27aed5ebed3b9763ab9cc823738adb274cd6263dd013cd66e176aff6f19ee3b2b06fa7733d0dfc240ffbf9a81fe6e06fa4b06cc51188c73c4d2ebc113a81f31962231ab3b927a4b4ebb35da43ebfcf7b4ce29cb9988bfbfe4b74547a1ae72ea022ab59effd6afa366df2537d654530746faff0ecdb34dfdcd0445a59a6f7e9da30923b6dcc5a46ff02b537ac672afb6e05b71b7316af63294ee66d464fad77779a3b5c9d4e7092f50132b4aaf177dab1b3426b59ef64392cce382bcb59ad446f65651dacbecb2200d8a2dc56dacdc94d9c7ced05e11bf9e1f8358bfbeda3c5c23e9a459198864c4142b905069af26abd5329dd4ad4f38779d360ad6ae7635b4e56cab72735e6de8af5d3d6ac277e9f677e8f6f7ebb6c7f256cb5707f69af91b28a6ca6ea99ba93dfb367de196d27ab7d8b79f4d9d9baacbb2d8e1e766649ab06c97cff0b67c865bc2b8591eee50cc59ce93e66565852a5881011c7f2e31b62f52176e2d71e1bafb53babcc7abc1ed7cde28b6e9dad13775b754235b1360db04e23581f84e0476d75571b7bf2dd1df735c34f71d1b2978d5dc72daf362cb6da93953fc5b1dafb579d9e7eb986f3d76daf7fc419ed798dfda4f75dd4e5dd7e9980bd6252eba3037338a890ca1dede8219c9a787600efa71694efad473ff868b83c7bafb77965778f0f8ffe76eb741d87de63441d34d73dd139a65517f556f6d7f3f39aad22eb40ff602b985bee11bcf6f3936eb379c7bfcf999bd15ee44f46fdc289761dbe37523b372bc605c78beb1bb0db0d522a6c81378e5dd26656b989777afd4ab2fd3a00d80a9fb5a7ff50b4973d145a56c84ae9dfff40fd50fe3613c8c87f1307ee8f857000000ffff0623672b001e0000c28080f8e005b8ddf8dbf8d9f86aa1020eb815124c063e6ed4750e94306ea53104de7afaefe440fe1a0cd6b98a81208fb8463044022013f14212e75aec9f38aadaa607404a244245a140cacd1ce87cbe70a5871ecf6e0220356f1d5780c4d78fac288f6b4dcd65e12a66d3ad31b657e435c48b95b124cdd0f86ba103bdc5747794aae58742cc837027b4d8fd757631e46a9da3b6bac081b19897eb38b8473045022100f883a2898f92bc4cc909077f9e1376322adf63050e686dff3190b0b27e2d3af5022038e9dd13cf803b5ac65266b1ce71268a2b7c07acc22650daf29f9dc47723486b8080")
	installTx := &modules.Transaction{}
	rlp.DecodeBytes(data, installTx)
	t.Log(installTx.IsSystemContract())

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mdag := mock.NewMockIDag(mockCtrl)
	mdag.EXPECT().GetUtxoEntry(gomock.Any()).DoAndReturn(
		func(outpoint *modules.OutPoint) (*modules.Utxo, error) {
			return &modules.Utxo{Amount: 123}, nil
		}).AnyTimes()
	mdag.EXPECT().IsTransactionExist(gomock.Any()).Return(false, nil).AnyTimes()
	pool := mockTxPool(mdag)
	err := pool.AddLocal(installTx)
	t.Log(err)
	assert.Nil(t, err)
}
func TestTxPool_GetUnpackedTxsByAddr(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	addr, _ := common.StringToAddress("P1HXNZReTByQHgWQNGMXotMyTkMG9XeEQfX")
	lockScript := tokenengine.Instance.GenerateLockScript(addr)

	mdag := mock.NewMockIDag(mockCtrl)
	mdag.EXPECT().GetUtxoEntry(gomock.Any()).DoAndReturn(
		func(outpoint *modules.OutPoint) (*modules.Utxo, error) {
			if outpoint.TxHash == Hash("dag") {
				return &modules.Utxo{Amount: 123, PkScript: lockScript}, nil
			}
			return nil, ErrNotFound
		}).AnyTimes()
	mdag.EXPECT().IsTransactionExist(gomock.Any()).Return(false, nil).AnyTimes()
	pool := mockTxPool(mdag)
	pay1 := mockPaymentTx(Hash("dag"), 0, 0)
	pool.AddLocal(pay1)
	t.Log("TxA:", pay1.Hash().String())
	req := mockContractInvokeRequest(pay1.Hash(), 0, 0, []byte("user contract"))
	err := pool.AddLocal(req)
	t.Log("ReqB:", req.Hash().String())
	assert.Nil(t, err)
	fullTx := mockContractInvokeFullTx(pay1.Hash(), 0, 0, []byte("user contract"))
	err = pool.AddLocal(fullTx)
	assert.Nil(t, err)
	t.Log("FullTxB:", fullTx.Hash().String())
	req1 := mockContractInvokeRequest(Hash("new one"), 0, 0, []byte("user contract"))
	err = pool.AddLocal(req1)
	t.Log("ReqX:", req1.Hash().String())
	assert.Nil(t, err)
	fullTx1 := mockContractInvokeFullTx(Hash("new one"), 0, 0, []byte("user contract"))
	err = pool.AddLocal(fullTx1)
	assert.Nil(t, err)
	txs, err := pool.GetUnpackedTxsByAddr(addr)
	assert.Nil(t, err)
	for _, tx := range txs {
		t.Log(tx.TxHash.String())
	}
	assert.Equal(t, 2, len(txs))
}
func TestTxPool_SubscribeTxPreEvent(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mdag := mock.NewMockIDag(mockCtrl)
	mdag.EXPECT().GetUtxoEntry(gomock.Any()).DoAndReturn(func(outpoint *modules.OutPoint) (*modules.Utxo, error) {
		if outpoint.TxHash == Hash("Dag") {
			return &modules.Utxo{Amount: 123}, nil
		}
		return nil, ErrNotFound
	}).AnyTimes()
	mdag.EXPECT().IsTransactionExist(gomock.Any()).Return(false, nil).AnyTimes()
	pool := mockTxPool(mdag)
	txpoolAddTxCh := make(chan modules.TxPreEvent, 50)
	txpoolAddTxSub := pool.SubscribeTxPreEvent(txpoolAddTxCh)
	eventResult := ""
	go func() {
		for {
			select {
			case tx := <-txpoolAddTxCh:
				log.Debugf("Subscribe TxPool add tx event received Tx:%s", tx.Tx.Hash().String())
				if !tx.IsOrphan {
					eventResult += tx.Tx.Hash().String() + ","
				}
			case err := <-txpoolAddTxSub.Err():
				if err != nil {
					log.Error(err.Error())
				}
				return
			}

		}

	}()
	txA := mockPaymentTx(Hash("Dag"), 0, 0)
	t.Logf("Tx A:%s", txA.Hash().String())
	txB := mockPaymentTx(txA.Hash(), 0, 0)
	t.Logf("Tx B:%s", txB.Hash().String())
	txC := mockPaymentTx(txB.Hash(), 0, 0)
	t.Logf("Tx C:%s", txC.Hash().String())
	pool.AddLocal(txB)
	pool.AddLocal(txC)
	pool.AddLocal(txA)
	time.Sleep(time.Second)
	t.Log("Event result:", eventResult)
	pool.Stop()
	expectHashes := txA.Hash().String() + "," + txB.Hash().String() + "," + txC.Hash().String() + ","
	assert.Equal(t, expectHashes, eventResult)
}
func TestReal(t *testing.T) {
	hexTx := "f9017201f90165f9012880b90124f90121f893f891b86b483045022100c64f8a4b24a8ee902fe6a8ba23bba13274fe8cbe60f2d53c8c3dd00625b860c602201a809a697aefaef5d1545c46d006ed49b2c818c7598849f2ce4d4fa3317b06af012103f114bb36b24a4e12d716697d64f6b3992f2a566c35910eb5307efe8ca197445e80a04dafdd75e302664cc254a8923a94d4960cc325d81858d9e3898f9c2aa2ef15ee8080f889f83f8405f5e10096140000000000000000000000000000000000000001c8e290400082bb0800000000000000000000009000000000000000000000000000000000f846880163457857941f009976a914619f00a47ef419402e69f5889d6955396337956d88ace290400082bb080000000000000000000000900000000000000000000000000000000080f83866b6f5940000000000000000000000000000000000000001de9d446576656c6f706572506179546f4465706f736974436f6e74726163748088c7845e82146b8080"
	data, _ := hex.DecodeString(hexTx)
	tx := &modules.Transaction{}
	rlp.DecodeBytes(data, tx)
	t.Log(tx.IsSystemContract())
	t.Logf("Msg count:%d", len(tx.TxMessages()))
	t.Logf("%#v", tx)
	t.Log("ReqHash:", tx.RequestHash().String())
	t.Log("TxHash:", tx.Hash().String())
	tx.SetVersion(0)
	tx.SetNonce(0)
	t.Log("NewTxHash", tx.Hash().String())

}

//先添加用户合约Request，然后是连续交易的转账，然后又是用户合约Request
func TestTxPool_AddUserContractAndTransferTx(t *testing.T) {
	addr, _ := common.StringToAddress("P1HXNZReTByQHgWQNGMXotMyTkMG9XeEQfX")
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
	mdag.EXPECT().IsTransactionExist(gomock.Any()).Return(false, nil).AnyTimes()
	pool := mockTxPool(mdag)

	reqA := mockContractInvokeRequest(Hash("dag"), 0, 0, []byte("user contract"))
	err := pool.AddLocal(reqA)
	assert.Nil(t, err)
	txB := mockPaymentTx(reqA.Hash(), 0, 0)
	err = pool.AddLocal(txB)
	assert.Nil(t, err)
	reqC := mockContractInvokeRequest(txB.Hash(), 0, 0, []byte("user contract"))
	err = pool.AddLocal(reqC)
	assert.Nil(t, err)
	sortedTx, err := pool.GetSortedTxs()
	assert.Equal(t, 0, len(sortedTx))
	txs, _ := pool.GetUnpackedTxsByAddr(addr)
	assert.Equal(t, 3, len(txs))
	fullTxA := mockContractInvokeFullTx(Hash("dag"), 0, 0, []byte("user contract"))
	log.Debug("接下来添加FullTxA，那么TxA和TxB都会变成Normal")
	err = pool.AddLocal(fullTxA)
	assert.Nil(t, err)
	sortedTx, err = pool.GetSortedTxs()
	assert.Equal(t, 2, len(sortedTx))
	txs, _ = pool.GetUnpackedTxsByAddr(addr)
	assert.Equal(t, 3, len(txs))
	//第二种情形，ReqA，B，B先完成FullTx
	log.Debug("-------------------")
	pool = mockTxPool(mdag)
	pool.AddLocal(reqA)
	reqB := mockContractInvokeRequest(reqA.Hash(), 0, 0, []byte("user contract"))
	pool.AddLocal(reqB)
	fullTxB := mockContractInvokeFullTx(reqA.Hash(), 0, 0, []byte("user contract"))
	err = pool.AddLocal(fullTxB)
	assert.Nil(t, err)
	sortedTx, _ = pool.GetSortedTxs()
	assert.Equal(t, 0, len(sortedTx))
	txs, _ = pool.GetUnpackedTxsByAddr(addr)
	assert.Equal(t, 2, len(txs))
	pool.AddLocal(fullTxA)
	sortedTx, err = pool.GetSortedTxs()
	assert.Equal(t, 2, len(sortedTx))
	txs, _ = pool.GetUnpackedTxsByAddr(addr)
	assert.Equal(t, 2, len(txs))
}

func TestTxpoolByRealUserContractTx(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mdag := mock.NewMockIDag(mockCtrl)
	mdag.EXPECT().GetUtxoEntry(gomock.Any()).DoAndReturn(func(outpoint *modules.OutPoint) (*modules.Utxo, error) {
		if outpoint.TxHash.String() == "0x8df043145b8bd7c3a295495236c68dbb14ca17a653f0ebea68231470b2b5a43a" {
			return &modules.Utxo{Amount: 123}, nil
		}
		return nil, ErrNotFound
	}).AnyTimes()
	mdag.EXPECT().IsTransactionExist(gomock.Any()).Return(false, nil).AnyTimes()
	pool := mockTxPool(mdag)

	//真实的几条交易，从前到后是依赖关系
	txb12e := rlpDecodeTx("f90111f9010cf8e580b8e2f8e0f893f891b86b483045022100b412470a5730d12dc45d1f6a000784c0f57c4a87e9f4fd32fcfeda549a53b5ca022009192bdd788d2a0ed73937af40c9755daa826709b53195ece233519cd4e3a7900121020266ff2e2a30b4e7bbf32ad9317b1d3d33d9ffd5f77cc3c203476d5d2b16b88380a08df043145b8bd7c3a295495236c68dbb14ca17a653f0ebea68231470b2b5a43a8080f848f846880163456ef582ec009976a9149c6ab12cc402fb5d064807861da793f6d7a32af788ace290400082bb080000000000000000000000900000000000000000000000000000000080e466a2e19483bd50a65bf2762db81e0fb85912c34473188aedca89706179737461746531808080")
	pool.AddLocal(txb12e)
	tx0f38 := rlpDecodeTx("f90111f9010cf8e580b8e2f8e0f893f891b86b483045022100d1624680430c9cc4942ae6f663ee94b27516383aa73cf8aec4f130f9870eb68f02205e7b95c718bb37dd5f06d0ce83daaa923f70507a4b37d3c31ebe3b62eed326340121020266ff2e2a30b4e7bbf32ad9317b1d3d33d9ffd5f77cc3c203476d5d2b16b88380a0b12e689c4dece67d376e38d71db9d0f42f7864c5b1e8015786d47602aa1dc5fb8080f848f846880163456eef8d0b009976a9149c6ab12cc402fb5d064807861da793f6d7a32af788ace290400082bb080000000000000000000000900000000000000000000000000000000080e466a2e19483bd50a65bf2762db81e0fb85912c34473188aedca89706179737461746531808080")
	pool.AddLocal(tx0f38)
	tx0266 := rlpDecodeTx("f90111f9010cf8e580b8e2f8e0f893f891b86b483045022100cbaf83c30b8a6a6fe62c487cd2bf5048efa0a2461fe018595948d2580c2e59ed0220659388425dc9e9f80ee5f96749fa02b342d3a5e264b094bceee12f068441ed540121020266ff2e2a30b4e7bbf32ad9317b1d3d33d9ffd5f77cc3c203476d5d2b16b88380a00f385cb12e524de9f0d4914abd226c3d0ffcb7f9ebaf4e44d852270fba949ec28080f848f846880163456ee9972a009976a9149c6ab12cc402fb5d064807861da793f6d7a32af788ace290400082bb080000000000000000000000900000000000000000000000000000000080e466a2e19483bd50a65bf2762db81e0fb85912c34473188aedca89706179737461746531808080")
	pool.AddLocal(tx0266)
	txc7b8 := rlpDecodeTx("f90110f9010bf8e480b8e1f8dff892f890b86a47304402202d0daac83c620ebb19184384f5e37ea75c5e49cd5e06a283b59a5d9eb0f5e9d102202da5ff1b7bc6ebeb975e3d44fc2bc2de52c43029a2ae2bcd72b687f2f077f8a60121020266ff2e2a30b4e7bbf32ad9317b1d3d33d9ffd5f77cc3c203476d5d2b16b88380a00266c512f86562578a06737ca1f53c84f225edbd63790e0f1f50f07f06df4b028080f848f846880163456ee3a149009976a9149c6ab12cc402fb5d064807861da793f6d7a32af788ace290400082bb080000000000000000000000900000000000000000000000000000000080e466a2e19483bd50a65bf2762db81e0fb85912c34473188aedca89706179737461746531808080")
	pool.AddLocal(txc7b8)
	tx9ee7 := rlpDecodeTx("f90110f9010bf8e480b8e1f8dff892f890b86a473044022007ebc489ac19b91b8ff90749d4658f55ca8a4e190a984dd643a7aeda8ba586cc02200d8b750f89bd85d753095b70df39ab6de94a2561dfabd4fc5331381c331716700121020266ff2e2a30b4e7bbf32ad9317b1d3d33d9ffd5f77cc3c203476d5d2b16b88380a0c7b8997dfc6c212dc0744033301934680cc5a226489d68c6384a1ced0315da6d8080f848f846880163456eddab68009976a9149c6ab12cc402fb5d064807861da793f6d7a32af788ace290400082bb080000000000000000000000900000000000000000000000000000000080e466a2e19483bd50a65bf2762db81e0fb85912c34473188aedca89706179737461746531808080")
	pool.AddLocal(tx9ee7)
	tx953f := rlpDecodeTx("f90111f9010cf8e580b8e2f8e0f893f891b86b4830450221009a87ffc711e81c673d5227218a5aa788d2ba259fb7a60e16f8f95a11c0ae427c022071e63749c2aef72ca7ebd592e856771747c00e1784465fc667bc59f2c43423590121020266ff2e2a30b4e7bbf32ad9317b1d3d33d9ffd5f77cc3c203476d5d2b16b88380a09ee7928cfd2b746030f238737cf60edf343a9bc313f4f14fbcce1ac100cf03438080f848f846880163456ed7b587009976a9149c6ab12cc402fb5d064807861da793f6d7a32af788ace290400082bb080000000000000000000000900000000000000000000000000000000080e466a2e19483bd50a65bf2762db81e0fb85912c34473188aedca89706179737461746531808080")
	pool.AddLocal(tx953f)
	tx34e4 := rlpDecodeTx("f90110f9010bf8e480b8e1f8dff892f890b86a47304402200db4478023c1e35e7095fe26387257f9cff6161c87974ef23863d9ceb171885e02206bc9e81b64afd11be4637bbf0236c8e4edddce154b98c1b1178621ed4d0a0f490121020266ff2e2a30b4e7bbf32ad9317b1d3d33d9ffd5f77cc3c203476d5d2b16b88380a0953f621dd3354ec781f5581b9405b01e9edc1f4cc9979e33ce1af9d295867a678080f848f846880163456ed1bfa6009976a9149c6ab12cc402fb5d064807861da793f6d7a32af788ace290400082bb080000000000000000000000900000000000000000000000000000000080e466a2e19483bd50a65bf2762db81e0fb85912c34473188aedca89706179737461746531808080")
	pool.AddLocal(tx34e4)
	log.Debug("-------Add FullTx---------")
	tx0f38_8696 := rlpDecodeTx("f9029cf90297f8e580b8e2f8e0f893f891b86b483045022100d1624680430c9cc4942ae6f663ee94b27516383aa73cf8aec4f130f9870eb68f02205e7b95c718bb37dd5f06d0ce83daaa923f70507a4b37d3c31ebe3b62eed326340121020266ff2e2a30b4e7bbf32ad9317b1d3d33d9ffd5f77cc3c203476d5d2b16b88380a0b12e689c4dece67d376e38d71db9d0f42f7864c5b1e8015786d47602aa1dc5fb8080f848f846880163456eef8d0b009976a9149c6ab12cc402fb5d064807861da793f6d7a32af788ace290400082bb080000000000000000000000900000000000000000000000000000000080e466a2e19483bd50a65bf2762db81e0fb85912c34473188aedca8970617973746174653180f603b4f39483bd50a65bf2762db81e0fb85912c34473188aedc0c0d7d68089706179737461746531897061797374617465318080c28080f9015105b9014df9014af90147f86ba102d32a2a4e73700a0b6630fa1745de0407a4378edc44c71cf9bca4c67db1af742bb8473045022100b351a2608f624280983eb06d85c371b9b2f06d464792edd374914f31d00e3526022056d4ae199106518aa4a0bb88e9e18592a1c9818358284932d25abc743bed38d8f86ba102934f018117879da80dd67d568eb64d90df1f135201e616d0d35cf6f411aa8cc8b8473045022100f34b407c28cc49ea8ae6409672e02cd1c8eb3c77a4006e944cbe79d5c43dd1330220312baa743d71ee4a8a6b54671796e6b1a18b32229b0fd81ce2d88dd9fd57bc6ef86ba102ae223fee2635d11808e24b491a7f9be0b5398f32339cb34416dd4eb73fd62ca7b8473045022100e781848f7cf3128f0baab279d7762bb280d9aff691a201eae64e2da31454a29a02202b80d300ab2650fa9e46b274caa3d68f95edcb35a6206a0b50f91c738ec653098080")
	pool.AddLocal(tx0f38_8696)
	log.Debug("-------Add FullTx 1---------")
	txc7b8_27bc := rlpDecodeTx("f90298f90293f8e480b8e1f8dff892f890b86a47304402202d0daac83c620ebb19184384f5e37ea75c5e49cd5e06a283b59a5d9eb0f5e9d102202da5ff1b7bc6ebeb975e3d44fc2bc2de52c43029a2ae2bcd72b687f2f077f8a60121020266ff2e2a30b4e7bbf32ad9317b1d3d33d9ffd5f77cc3c203476d5d2b16b88380a00266c512f86562578a06737ca1f53c84f225edbd63790e0f1f50f07f06df4b028080f848f846880163456ee3a149009976a9149c6ab12cc402fb5d064807861da793f6d7a32af788ace290400082bb080000000000000000000000900000000000000000000000000000000080e466a2e19483bd50a65bf2762db81e0fb85912c34473188aedca8970617973746174653180f603b4f39483bd50a65bf2762db81e0fb85912c34473188aedc0c0d7d68089706179737461746531897061797374617465318080c28080f9014e05b9014af90147f90144f86aa102d32a2a4e73700a0b6630fa1745de0407a4378edc44c71cf9bca4c67db1af742bb846304402202cb4bf66b755697ea61c05735c945c0ba8307ebc269ef1363eba01cb52f7063302205612b8572323d66e45ec37930dd3efef385882c94970a5b36e619eaf2ec936d7f86aa102934f018117879da80dd67d568eb64d90df1f135201e616d0d35cf6f411aa8cc8b8463044022030a60503a6b044db7622391148eb0c4e3ec74bf2435f02f7233385f1f465553502207fccacae3b41a12bdb1ec8ac3b39b2adb77bf7918bb73fcf6d79f78054a6ad92f86aa102ae223fee2635d11808e24b491a7f9be0b5398f32339cb34416dd4eb73fd62ca7b846304402207b7f2026b249da60c5269b75403d78c46d3678eb076f07bc5a81f1630e016af802207b0b341f41660f7ec12485797c987eb60a7bb7e04afa48519981dd1313196d3e8080")
	pool.AddLocal(txc7b8_27bc)
	log.Debug("-------Add FullTx 2---------")
	txb12e_f25b := rlpDecodeTx("f90299f90294f8e580b8e2f8e0f893f891b86b483045022100b412470a5730d12dc45d1f6a000784c0f57c4a87e9f4fd32fcfeda549a53b5ca022009192bdd788d2a0ed73937af40c9755daa826709b53195ece233519cd4e3a7900121020266ff2e2a30b4e7bbf32ad9317b1d3d33d9ffd5f77cc3c203476d5d2b16b88380a08df043145b8bd7c3a295495236c68dbb14ca17a653f0ebea68231470b2b5a43a8080f848f846880163456ef582ec009976a9149c6ab12cc402fb5d064807861da793f6d7a32af788ace290400082bb080000000000000000000000900000000000000000000000000000000080e466a2e19483bd50a65bf2762db81e0fb85912c34473188aedca8970617973746174653180f603b4f39483bd50a65bf2762db81e0fb85912c34473188aedc0c0d7d68089706179737461746531897061797374617465318080c28080f9014e05b9014af90147f90144f86aa102934f018117879da80dd67d568eb64d90df1f135201e616d0d35cf6f411aa8cc8b84630440220489404f9aeb44d420deefca8aef32d7571c8ab0c1b9dd573679af7b40719ed0902203185f1a9cc43e075418243ec87f68f4ae5ca1e01a63b11e7988956c6dc474d01f86aa102ae223fee2635d11808e24b491a7f9be0b5398f32339cb34416dd4eb73fd62ca7b84630440220703026356d545755489d14fa5c39786140c4cd571eafff9869efe7d46fc188560220745926802e987a6f594cc4879340fb28e59de9aaf253c83737eae3262588ec27f86aa102d32a2a4e73700a0b6630fa1745de0407a4378edc44c71cf9bca4c67db1af742bb84630440220528937cb84c38287400289145a80a43cbe54e5124a95e20147eb45e19540d6b10220060029e8d32c068df55ca2b612a57f4ad61c99822e5db24c423157754f8fef0f8080")
	pool.AddLocal(txb12e_f25b)
	log.Debug("-------Add FullTx 3---------")
	for hash := range pool.userContractRequests {
		log.Debug("current req[%s]", hash.String())
	}

	tx9ee7_9d51 := rlpDecodeTx("f90298f90293f8e480b8e1f8dff892f890b86a473044022007ebc489ac19b91b8ff90749d4658f55ca8a4e190a984dd643a7aeda8ba586cc02200d8b750f89bd85d753095b70df39ab6de94a2561dfabd4fc5331381c331716700121020266ff2e2a30b4e7bbf32ad9317b1d3d33d9ffd5f77cc3c203476d5d2b16b88380a0c7b8997dfc6c212dc0744033301934680cc5a226489d68c6384a1ced0315da6d8080f848f846880163456eddab68009976a9149c6ab12cc402fb5d064807861da793f6d7a32af788ace290400082bb080000000000000000000000900000000000000000000000000000000080e466a2e19483bd50a65bf2762db81e0fb85912c34473188aedca8970617973746174653180f603b4f39483bd50a65bf2762db81e0fb85912c34473188aedc0c0d7d68089706179737461746531897061797374617465318080c28080f9014e05b9014af90147f90144f86aa102934f018117879da80dd67d568eb64d90df1f135201e616d0d35cf6f411aa8cc8b846304402200f5771ef18747303de423562d1824a4863be6cdb5b374e042dc33ff92e6dca7d02206d6f4976509e0388c3bc5052c714b7b5d05d7b5f6b2ec382e52bd09854ff56bef86aa102ae223fee2635d11808e24b491a7f9be0b5398f32339cb34416dd4eb73fd62ca7b846304402205188ebcbef0de5ebb8e0a40e3aae5777257716fd9370fdce438e68789871978002202a32b8762ad4fcf49da8fdb4618966c457df5153df9a9854c85b97773a8932d9f86aa102d32a2a4e73700a0b6630fa1745de0407a4378edc44c71cf9bca4c67db1af742bb846304402201e177514e64dfdc93c450abd412063ebc3a02e73b3ce99f0c0d0845a4e14a82d022037886160aaf89c5b4179450edb57305d89738afc983f4bff31cdf82995c90b438080")
	pool.AddLocal(tx9ee7_9d51)
	log.Debug("-------Add FullTx 4---------")
	tx34e4_cb85 := rlpDecodeTx("f90298f90293f8e480b8e1f8dff892f890b86a47304402200db4478023c1e35e7095fe26387257f9cff6161c87974ef23863d9ceb171885e02206bc9e81b64afd11be4637bbf0236c8e4edddce154b98c1b1178621ed4d0a0f490121020266ff2e2a30b4e7bbf32ad9317b1d3d33d9ffd5f77cc3c203476d5d2b16b88380a0953f621dd3354ec781f5581b9405b01e9edc1f4cc9979e33ce1af9d295867a678080f848f846880163456ed1bfa6009976a9149c6ab12cc402fb5d064807861da793f6d7a32af788ace290400082bb080000000000000000000000900000000000000000000000000000000080e466a2e19483bd50a65bf2762db81e0fb85912c34473188aedca8970617973746174653180f603b4f39483bd50a65bf2762db81e0fb85912c34473188aedc0c0d7d68089706179737461746531897061797374617465318080c28080f9014e05b9014af90147f90144f86aa102934f018117879da80dd67d568eb64d90df1f135201e616d0d35cf6f411aa8cc8b8463044022045d9639258d1890833ffff100cad47a4b22a5a92505e060fcd0b25d2f98ca5f702203562b4f404f323f0166d16be8eecc27e4083bcc4454677ca045b838c8fe2b3b6f86aa102ae223fee2635d11808e24b491a7f9be0b5398f32339cb34416dd4eb73fd62ca7b846304402204c44c1440c10d4a87d442599c0cf38f25c56d8473e62c1fe9031b802e17f9c8302204814dcc44680ee41a69c493924c94433c141c1dcb93a9c1ed3aa6c0903bd5398f86aa102d32a2a4e73700a0b6630fa1745de0407a4378edc44c71cf9bca4c67db1af742bb8463044022053492abd61767567fd030e76b344ea01ca963f2d189acb37bcdcc37460e51e5302205e55518a01b50e5311d21c714f5c9b7e69a299f89d1a6d349e73dba7a4c89bec8080")
	pool.AddLocal(tx34e4_cb85)
	log.Debug("-------Add FullTx 5---------")
	tx953f_af82 := rlpDecodeTx("f9029bf90296f8e580b8e2f8e0f893f891b86b4830450221009a87ffc711e81c673d5227218a5aa788d2ba259fb7a60e16f8f95a11c0ae427c022071e63749c2aef72ca7ebd592e856771747c00e1784465fc667bc59f2c43423590121020266ff2e2a30b4e7bbf32ad9317b1d3d33d9ffd5f77cc3c203476d5d2b16b88380a09ee7928cfd2b746030f238737cf60edf343a9bc313f4f14fbcce1ac100cf03438080f848f846880163456ed7b587009976a9149c6ab12cc402fb5d064807861da793f6d7a32af788ace290400082bb080000000000000000000000900000000000000000000000000000000080e466a2e19483bd50a65bf2762db81e0fb85912c34473188aedca8970617973746174653180f603b4f39483bd50a65bf2762db81e0fb85912c34473188aedc0c0d7d68089706179737461746531897061797374617465318080c28080f9015005b9014cf90149f90146f86aa102934f018117879da80dd67d568eb64d90df1f135201e616d0d35cf6f411aa8cc8b8463044022013704dac9ab8a24485af3b3fe028f19d6444b08046b5f7e3ae3124011d6eda070220311d0097186ee8530f24e8625c66d402b40368851261e520509982654d40158ff86ba102d32a2a4e73700a0b6630fa1745de0407a4378edc44c71cf9bca4c67db1af742bb8473045022100c22f641401d36b19a44af69b0e12803a85859fcbaaab6b1cee2fc1143a6b63b60220741d3f43a61fabe1f14908680cc6a0b02663e5aa8af3ed590f31baa2d9bd2169f86ba102ae223fee2635d11808e24b491a7f9be0b5398f32339cb34416dd4eb73fd62ca7b8473045022100a615145f897e7efc5c6d7bac28a01eca972ece7a4a1a74a5da31ad85c73ccc3c022058dcf04cf9b9d4af3cdd0bef2e46463879e8a3642a4619b0d70623fd95b68eb28080")
	pool.AddLocal(tx953f_af82)
	tx0266_2dd2 := rlpDecodeTx("f9029af90295f8e580b8e2f8e0f893f891b86b483045022100cbaf83c30b8a6a6fe62c487cd2bf5048efa0a2461fe018595948d2580c2e59ed0220659388425dc9e9f80ee5f96749fa02b342d3a5e264b094bceee12f068441ed540121020266ff2e2a30b4e7bbf32ad9317b1d3d33d9ffd5f77cc3c203476d5d2b16b88380a00f385cb12e524de9f0d4914abd226c3d0ffcb7f9ebaf4e44d852270fba949ec28080f848f846880163456ee9972a009976a9149c6ab12cc402fb5d064807861da793f6d7a32af788ace290400082bb080000000000000000000000900000000000000000000000000000000080e466a2e19483bd50a65bf2762db81e0fb85912c34473188aedca8970617973746174653180f603b4f39483bd50a65bf2762db81e0fb85912c34473188aedc0c0d7d68089706179737461746531897061797374617465318080c28080f9014f05b9014bf90148f90145f86aa102ae223fee2635d11808e24b491a7f9be0b5398f32339cb34416dd4eb73fd62ca7b8463044022029457f6be42f800d95122bc3f61c5e18bd101ed9642d766f9e548e936919901202201e1919e821cd63db084cd9e2f98466f34c79fec7783df472dfdafd7a05b5090ef86ba102d32a2a4e73700a0b6630fa1745de0407a4378edc44c71cf9bca4c67db1af742bb8473045022100afbca75944d1625c8b322b03da111a1e027e3da9b133d776bc4a54a9915c3a7a02203244641985f6d4ad9ec1a76bfb4df62e5f081489da6c6ac6cf778142a2fa35ccf86aa102934f018117879da80dd67d568eb64d90df1f135201e616d0d35cf6f411aa8cc8b846304402206d880278401648b32951e3f71fc8054fedc028b41f9f335905a5cc17796b4211022075e157cd183e04b72fcac1ee2f5d8fefc191b40c89126e403fc3256c186f38a58080")
	pool.AddLocal(tx0266_2dd2)

	sortedTx := ""
	list, _ := pool.GetSortedTxs()
	for _, tx := range list {
		sortedTx += string(tx.ReqHash.String()) + ";"
	}

	t.Log("Real sort result:", sortedTx)
	match, _ := regexp.MatchString(".*b12e.*0f38.*0266.*c7b8.*9ee7.*953f.*34e4.*", sortedTx)
	assert.True(t, match)
}

func TestTxpoolByRealUserContractTx2(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mdag := mock.NewMockIDag(mockCtrl)
	mdag.EXPECT().GetUtxoEntry(gomock.Any()).DoAndReturn(func(outpoint *modules.OutPoint) (*modules.Utxo, error) {
		if outpoint.TxHash.String() == "0xfa2429e478938a9d4db6774d2e5450b898801665f467f18f19962a6ae189ef99" {
			return &modules.Utxo{Amount: 123}, nil
		}
		return nil, ErrNotFound
	}).AnyTimes()
	mdag.EXPECT().IsTransactionExist(gomock.Any()).Return(false, nil).AnyTimes()
	pool := mockTxPool(mdag)

	//真实的几条交易，从前到后是依赖关系
	txc06d := rlpDecodeTx("f90299f90294f8e580b8e2f8e0f893f891b86b483045022100a6a59963166d7395c61ec8e4b93889eb2ca73b3d36cf9354aeecf9082b271d5a02205e38f14f301320eb2eda45ff49d0b26e8dc08ec3d0ec87ec63993de951a2fbf301210303136e70f20ead1393f1836707e7cf216dfc8bc4b5170965ce5a83478867ec3480a0fa2429e478938a9d4db6774d2e5450b898801665f467f18f19962a6ae189ef998080f848f846880163433385d4ed009976a91401b6c3f48ed75b4356af94d92e6924dcca7415ad88ace290400082bb080000000000000000000000900000000000000000000000000000000080e466a2e19457b99654a11da2d35bcd74c2fb6819f366e44b2aca8970617973746174653180f603b4f39457b99654a11da2d35bcd74c2fb6819f366e44b2ac0c0d7d68089706179737461746531897061797374617465318080c28080f9014e05b9014af90147f90144f86aa1038d82d3ff638ab2d7c06432201794e224650be3605951d43af294d886422aeb9bb84630440220468e2b37772973fc1172cb022d2c7b67b8c2a74f5ca18ed079aee705dfd45819022003df4d6e1bb99e0a1a48e7a6e9a748f914eefd2600e97bf3f46d1cc833820158f86aa103737c5b180f504ff5bacbe39def0189b95a58ff134ceaa2bc61f2bf45e06490cbb84630440220733f7df3e291229e50d092c37cd569fe43c7549f4621f00d72c2860b382e5dee022001dc2138f39e19ff298cf805cc02dfaa81734910206d9629b70fc41a14fc0323f86aa102dd34f69ab9b78973ccbe24be17223006022a7e283cf1cf83d613c77ed08bd49cb84630440220730c54b939bbedd22ee35006eddf5f47e31a38989a5302b14315254fd4f1c5f002202251fdf3f9c496325e6c931c4065647596c80b05bb4e10e1a166fb35f6dba9bd8080")
	tx428f := rlpDecodeTx("f90299f90294f8e480b8e1f8dff892f890b86a4730440220150a3e8d2335494073140ee8c5a4d5e14cbceeebf174690f3332ee70470cc63d022058f54fff0e685a4a008167b5be7e30a0ceafc78942707b840427a2494c3df31e01210303136e70f20ead1393f1836707e7cf216dfc8bc4b5170965ce5a83478867ec3480a0aeb01a81907882be35343daac3017ead26ef2fc0a0d85c9067f0460f8ad8915a8080f848f84688016343337fdf0c009976a91401b6c3f48ed75b4356af94d92e6924dcca7415ad88ace290400082bb080000000000000000000000900000000000000000000000000000000080e466a2e19457b99654a11da2d35bcd74c2fb6819f366e44b2aca8970617973746174653180f603b4f39457b99654a11da2d35bcd74c2fb6819f366e44b2ac0c0d7d68089706179737461746531897061797374617465318080c28080f9014f05b9014bf90148f90145f86aa102dd34f69ab9b78973ccbe24be17223006022a7e283cf1cf83d613c77ed08bd49cb84630440220337a1c444049c3eea627cc0fe11abd0013a5bf8bebb190837dc3da931868fe2302202149c092679cd0fbc737c33766b4da02b4277bd541d7c728a6e7045856fcc8fff86aa1038d82d3ff638ab2d7c06432201794e224650be3605951d43af294d886422aeb9bb846304402204b8e7180611f4f7d772787ed9cc65ffc4732baa1400dbe064272a662e261cf1102206866c471b51ef596b039dee7f01e8976ed4e7e4e4467a7d0f67e45a98816f589f86ba103737c5b180f504ff5bacbe39def0189b95a58ff134ceaa2bc61f2bf45e06490cbb8473045022100e5ee3b05a4faa95f6eb650c6c629c8a98ec97defa31bbcc8b50585e695c49821022051b3c1f5e48ceab3c9274931d451949fc230130fea30f3b0c41f5ce19022da738080")
	pool.AddLocal(txc06d)
	pool.AddLocal(tx428f)
}
