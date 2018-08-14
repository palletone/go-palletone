package common

import (
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
	if err := SaveConfig(confs); err != nil {
		log.Println(err)
	}
}

func TestGetConfig(t *testing.T) {
	data := GetConfig([]byte("Test Struct"))
	if len(data) <= 0 {
		log.Println("Get config data error")
	} else {

		log.Println("Get Data:", data)
	}

	var st modules.Asset
	if err := rlp.DecodeBytes(data, &st); err != nil {
		t.Error(err.Error())
	}
	log.Println(st)
}

func TestSaveStruct(t *testing.T) {
	if storage.Dbconn == nil {
		storage.Dbconn = storage.ReNewDbConn(dagconfig.DefaultConfig.DbPath)
	}
	aid := modules.IDType16{}
	aid.SetBytes([]byte("1111111111111111222222222222222222"))
	st := modules.Asset{
		AssertId: aid,
		UniqueId: aid,
		ChainId:  1,
	}
	if err := storage.Store(storage.Dbconn, "TestStruct", st); err != nil {
		t.Error(err.Error())
	}
}

func TestReadStruct(t *testing.T) {
	data, err := storage.Get([]byte("TestStruct"))
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
