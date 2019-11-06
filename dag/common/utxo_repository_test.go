/*
 *
 *    This file is part of go-palletone.
 *    go-palletone is free software: you can redistribute it and/or modify
 *    it under the terms of the GNU General Public License as published by
 *    the Free Software Foundation, either version 3 of the License, or
 *    (at your option) any later version.
 *    go-palletone is distributed in the hope that it will be useful,
 *    but WITHOUT ANY WARRANTY; without even the implied warranty of
 *    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *    GNU General Public License for more details.
 *    You should have received a copy of the GNU General Public License
 *    along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
 * /
 *
 *  * @author PalletOne core developer <dev@pallet.one>
 *  * @date 2018
 *
 */

package common

import (
	"io/ioutil"
	"log"
	"testing"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/ptndb"

	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/storage"
	"github.com/palletone/go-palletone/tokenengine"
	"github.com/stretchr/testify/assert"
	"time"
)

func mockUtxoRepository() *UtxoRepository {
	db, _ := ptndb.NewMemDatabase()
	utxodb := storage.NewUtxoDb(db, tokenengine.Instance)
	idxdb := storage.NewIndexDb(db)
	statedb := storage.NewStateDb(db)
	propDb := storage.NewPropertyDb(db)
	return NewUtxoRepository(utxodb, idxdb, statedb, propDb, tokenengine.Instance)
}

func TestUpdateUtxo(t *testing.T) {
	rep := mockUtxoRepository()
	rep.UpdateUtxo(time.Now().Unix(), common.Hash{}, &modules.PaymentPayload{}, uint32(0))
}

//func TestReadUtxos(t *testing.T) {
//	rep := mockUtxoRepository()
//	utxos, totalAmount := rep.ReadUtxos(common.Address{}, modules.Asset{})
//	log.Println(utxos, totalAmount)
//}

func TestGetUxto(t *testing.T) {
	dagconfig.DagConfig.DbPath = getTempDir(t)
	log.Println(modules.Input{})
}

func getTempDir(t *testing.T) string {
	d, err := ioutil.TempDir("", "leveldb-test")
	if err != nil {
		t.Fatal(err)
	}
	return d
}

//func TestSaveAssetInfo(t *testing.T) {
//	assetid := modules.PTNCOIN
//	asset := modules.Asset{
//		AssetId:  assetid,
//		UniqueId: assetid,
//	}
//	assetInfo := modules.AssetInfo{
//		GasToken:        "Test",
//		AssetID:      &asset,
//		InitialTotal: 1000000000,
//		Decimal:      100000000,
//	}
//	assetInfo.OriginalHolder.SetString("Mytest")
//}

//func TestWalletBalance(t *testing.T) {
//	rep := mockUtxoRepository()
//	addr := common.Address{}
//	addr.SetString("P1CXn936dYuPKGyweKPZRycGNcwmTnqeDaA")
//	balance := rep.WalletBalance(addr, modules.Asset{})
//	log.Println("Address total =", balance)
//}

//
//func TestGetAccountTokens(t *testing.T) {
//	rep := mockUtxoRepository()
//	addr := common.Address{}
//	addr.SetString("P12EA8oRMJbAtKHbaXGy8MGgzM8AMPYxkNr")
//	tokens, err := rep.GetAccountTokens(addr)
//	if err != nil {
//		log.Println("Get account error:", err.Error())
//	} else if len(tokens) == 0 {
//		log.Println("Get none account")
//	} else {
//		for _, token := range tokens {
//			log.Printf("Token (%s, %v) = %v\n",
//				token.GasToken, token.AssetID.AssetId, token.Balance)
//			// test WalletBalance method
//			log.Println(rep.WalletBalance(addr, *token.AssetID))
//			// test ReadUtxos method
//			utxos, amount := rep.ReadUtxos(addr, *token.AssetID)
//			log.Printf("Addr(%s) balance=%v\n", addr.String(), amount)
//			for outpoint, utxo := range utxos {
//				log.Println(">>> UTXO txhash =", outpoint.TxHash.String())
//				log.Println("    UTXO msg index =", outpoint.MessageIndex)
//				log.Println("    UTXO out index =", outpoint.OutIndex)
//				log.Println("    UTXO amount =", utxo.Amount)
//			}
//		}
//	}
//
//}

func Test_SaveUtxoView(t *testing.T) {
	db, _ := ptndb.NewMemDatabase()
	rep := NewUtxoRepository4Db(db, tokenengine.Instance)
	utxoViews := mockUtxoViews()
	err := rep.SaveUtxoView(utxoViews)
	assert.Nil(t, err)
	addr, _ := common.StringToAddress("P124gB1bXHDTXmox58g4hd4u13HV3e5vKie")
	utxos, err := rep.GetAddrUtxos(addr, modules.NewPTNAsset())
	assert.Nil(t, err)
	assert.Equal(t, len(utxos), 2)
	for u := range utxos {
		t.Logf("Utxo:%s", u.String())
	}
	addr2, _ := common.StringToAddress("P1LWaK3KBCuPVsXUPHXkMZr2Cm5tZquRDK8")
	utxos2, err := rep.GetAddrUtxos(addr2, modules.NewPTNAsset())
	assert.Equal(t, len(utxos2), 0)
	//Clear db and query again!
	err = rep.ClearUtxo()
	assert.Nil(t, err)
	utxos, err = rep.GetAddrUtxos(addr, modules.NewPTNAsset())
	assert.Nil(t, err)
	assert.Equal(t, len(utxos), 0)
}
func mockUtxoViews() map[modules.OutPoint]*modules.Utxo {

	result := make(map[modules.OutPoint]*modules.Utxo)
	addr, _ := common.StringToAddress("P124gB1bXHDTXmox58g4hd4u13HV3e5vKie")
	lockScript := tokenengine.Instance.GenerateLockScript(addr)
	o1 := modules.NewOutPoint(common.HexToHash("1111111"), 0, 0)
	utxo1 := modules.NewUtxo(&modules.Output{Value: 123, Asset: modules.NewPTNAsset(), PkScript: lockScript}, 0, 0)
	result[*o1] = utxo1

	o2 := modules.NewOutPoint(common.HexToHash("666666666"), 0, 0)
	utxo2 := modules.NewUtxo(&modules.Output{Value: 6666, Asset: modules.NewPTNAsset(), PkScript: lockScript}, 0, 0)
	result[*o2] = utxo2
	return result
}
