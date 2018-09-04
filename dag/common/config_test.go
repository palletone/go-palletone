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
	Dbconn := storage.ReNewDbConn("E:\\codes\\go\\src\\github.com\\palletone\\go-palletone\\cmd\\gptn\\gptn\\leveldb")
	if Dbconn == nil {
		fmt.Println("Connect to db error.")
		return
	}
	confs := []modules.PayloadMapStruct{}
	aid := modules.IDType16{}
	aid.SetBytes([]byte("1111111111111111222222222222222222"))
	st := modules.Asset{
		AssertId: aid,
		UniqueId: aid,
		ChainId:  1,
	}
	confs = append(confs, modules.PayloadMapStruct{Key: "TestStruct", Value: modules.ToPayloadMapValueBytes(st)})
	confs = append(confs, modules.PayloadMapStruct{Key: "TestInt", Value: modules.ToPayloadMapValueBytes(uint32(10))})
	stateVersion := modules.StateVersion{
		Height: modules.ChainIndex{
			AssetID: aid,
			IsMain:  true,
			Index:   0,
		},
		TxIndex: 0,
	}
	log.Println(stateVersion)
	if err := SaveConfig(Dbconn, confs, &stateVersion); err != nil {
		log.Println(err)
	}
}

func TestGetConfig(t *testing.T) {
	Dbconn := storage.ReNewDbConn("E:\\codes\\go\\src\\github.com\\palletone\\go-palletone\\cmd\\gptn\\gptn\\leveldb")
	if Dbconn == nil {
		fmt.Println("Connect to db error.")
		return
	}
	// todo get struct
	data := GetConfig(Dbconn, []byte("TestStruct"))
	if len(data) <= 0 {
		log.Println("Get config data error")
	} else {
		log.Println("Get Data:", data)
	}

	var st modules.Asset
	if err := rlp.DecodeBytes(data, &st); err != nil {
		log.Println("Get config data error:", err.Error())
		return
	}
	log.Println(st.ChainId, st.UniqueId, st.AssertId)
	// todo get int
	int_data := GetConfig(Dbconn, []byte("TestInt"))
	if len(data) <= 0 {
		log.Println("Get config int data error")
	} else {
		log.Println("Get int Data:", int_data)
	}
	var i uint32
	if err := rlp.DecodeBytes(int_data, &i); err != nil {
		log.Println("Get config data error:", err.Error())
		return
	}
	log.Println("int value=", i)
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
