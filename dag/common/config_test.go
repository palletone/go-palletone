package common

import (
	"fmt"
	"github.com/palletone/go-palletone/common/rlp"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/storage"
	"log"
	"testing"
)

func TestSaveConfig(t *testing.T) {
	confs := []modules.PayloadMapStruct{}
	aid := modules.IDType16{}
	aid.SetBytes([]byte("1111111111111111222222222222222222"))
	st := modules.Asset{
		AssertId: aid,
		UniqueId: aid,
		ChainId:  1,
	}
	confs = append(confs, modules.PayloadMapStruct{Key: "TestStruct", Value: st})
	stateVersion := modules.StateVersion{
		Height: modules.ChainIndex{
			AssetID: aid,
			IsMain:  true,
			Index:   0,
		},
		TxIndex: 0,
	}
	log.Println(stateVersion)
	// if err := SaveConfig(confs, &stateVersion); err != nil {
	// 	log.Println(err)
	// }
}

func TestGetConfig(t *testing.T) {
	Dbconn := storage.ReNewDbConn("E:\\codes\\go\\src\\github.com\\palletone\\go-palletone\\cmd\\gptn\\gptn\\leveldb")
	if Dbconn == nil {
		fmt.Println("Connect to db error.")
		return
	}
	data := GetConfig(Dbconn, []byte("tokenAmount"))
	if len(data) <= 0 {
		log.Println("Get config data error")
	} else {

		log.Println("Get Data:", data)
	}
	//var st modules.Asset
	//if err := rlp.DecodeBytes(data, &st); err != nil {
	//	log.Println(err.Error())
	//}
	//log.Println(st)
}

func TestSaveStruct(t *testing.T) {
	Dbconn := storage.ReNewDbConn(dagconfig.DbPath)
	if Dbconn == nil {
		fmt.Println("Connect to db error.")
		return
	}
	aid := modules.IDType16{}
	aid.SetBytes([]byte("1111111111111111222222222222222222"))
	st := modules.Asset{
		AssertId: aid,
		UniqueId: aid,
		ChainId:  1,
	}
	if err := storage.Store(Dbconn, "TestStruct", st); err != nil {
		t.Error(err.Error())
	}
}

func TestReadStruct(t *testing.T) {
	Dbconn := storage.ReNewDbConn(dagconfig.DbPath)
	if Dbconn == nil {
		fmt.Println("Connect to db error.")
		return
	}

	data, err := storage.Get(Dbconn, []byte("TestStruct"))
	if err != nil {
		t.Error(err.Error())
	}

	var st modules.Asset
	if err := rlp.DecodeBytes(data, &st); err != nil {
		t.Error(err.Error())
	}
	log.Println("Data:", data)
	log.Println(st)
}
