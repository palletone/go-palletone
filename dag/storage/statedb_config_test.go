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

package storage

import (
	"testing"

	"github.com/palletone/go-palletone/common/ptndb"
)

func MockStateMemDb() IStateDb {
	db, _ := ptndb.NewMemDatabase()
	//l := log.NewTestLog()
	//db, _ := ptndb.NewLDBDatabase("E:\\codes\\go\\src\\github.com\\palletone\\go-palletone\\cmd\\gptn\\gptn\\leveldb", 0, 0)
	statedb := NewStateDb(db)
	return statedb
}

func TestSaveAndGetConfig(t *testing.T) {
	//Dbconn := storage.ReNewDbConn("E:\\codes\\go\\src\\github.com\\palletone\\go-palletone\\cmd\\gptn\\gptn\\leveldb")
	//if Dbconn == nil {
	//	fmt.Println("Connect to db error.")
	//	return
	//}
	//db := MockStateMemDb()
	//confs := []modules.ContractWriteSet{}
	//aid := modules.AssetId{}
	//aid.SetBytes([]byte("1111111111111111222222222222222222"))
	//st := modules.Asset{
	//	AssetId:  aid,
	//	UniqueId: aid,
	//	ChainId:  1,
	//}
	//confs = append(confs, modules.ContractWriteSet{Key: "TestStruct", Value: modules.ToPayloadMapValueBytes(st)})
	//confs = append(confs, modules.ContractWriteSet{Key: "TestInt", Value: modules.ToPayloadMapValueBytes(uint32(10))})
	//stateVersion := modules.StateVersion{
	//	Height: modules.ChainIndex{
	//		AssetID: aid,
	//		IsMain:  true,
	//		Index:   0,
	//	},
	//	TxIndex: 0,
	//}
	//logger.Println(stateVersion)
	//if err := db.SaveConfig(confs, &stateVersion); err != nil {
	//	logger.Println(err)
	//}
	//
	//data := db.GetConfig([]byte("MediatorCandidates"))
	//var mList []core.MediatorInfo
	//fmt.Println(data)
	//if err := rlp.DecodeBytes(data, &mList); err != nil {
	//	logger.Println("Check unit signature when get mediators list", "error", err.Error())
	//	return
	//}
	//// todo get ActiveMediators
	//bNum := db.GetConfig([]byte("ActiveMediators"))
	//var mNum uint16
	//if err := rlp.DecodeBytes(bNum, &mNum); err != nil {
	//	logger.Println("Check unit signature", "error", err.Error())
	//	return
	//}
	//fmt.Println("Num=", mNum)
	//if int(mNum) != len(mList) {
	//	logger.Println("Check unit signature", "error", "mediators info error, pls update network")
	//	return
	//}
	// todo get GenesisAsset
	// genesisAsset, _, err := db.GetConfig([]byte(modules.FIELD_GENESIS_ASSET))
	// if !assert.NotNil(t, err) {
	// 	log.Error("getConfig is failed", "error", err)
	// 	return
	// }
	// var asset modules.Asset
	// if err := rlp.DecodeBytes(genesisAsset, &asset); err != nil {
	// 	log.Error("Check unit signature", "error", err.Error())
	// 	return
	// }
	// fmt.Println("asset=", asset)
	// log.Debug(">>>>>>>>> Pass >>>>>>>>>>.")
}

//
//func TestSaveStruct(t *testing.T) {
//	//Dbconn := storage.ReNewDbConn(dagconfig.DbPath)
//	//if Dbconn == nil {
//	//	fmt.Println("Connect to db error.")
//	//	return
//	//}
//	db:=MockStateMemDb()
//	aid := modules.AssetId{}
//	aid.SetBytes([]byte("1111111111111111222222222222222222"))
//	st := modules.Asset{
//		AssetId:  aid,
//		UniqueId: aid,
//		ChainId:  1,
//	}
//
//	if err := storage.Store(Dbconn, "TestStruct", st); err != nil {
//		t.Error(err.Error())
//	}
//}
//
//func TestReadStruct(t *testing.T) {
//	Dbconn := storage.ReNewDbConn(dagconfig.DbPath)
//	if Dbconn == nil {
//		fmt.Println("Connect to db error.")
//		return
//	}
//
//	data, err := storage.Get(Dbconn, []byte("TestStruct"))
//	if err != nil {
//		t.Error(err.Error())
//	}
//
//	var st modules.Asset
//	if err := rlp.DecodeBytes(data, &st); err != nil {
//		t.Error(err.Error())
//	}
//	logger.Println("Data:", data)
//	logger.Println(st)
//}
