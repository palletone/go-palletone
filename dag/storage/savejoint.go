package storage

import (
	"encoding/json"
	"errors"
	"log"

	palletdb "github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/dag/constants"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/modules"
)

var (
	Dbconn             *palletdb.LDBDatabase = nil
	AssocUnstableUnits map[string]modules.Joint
	//DBPath             string = "/Users/jay/code/gocode/src/palletone/bin/leveldb"
	DBPath string = dagconfig.DefaultConfig.DbPath
)

func SaveJoint(objJoint *modules.Joint, objValidationState *modules.ValidationState, onDone func()) (err error) {
	log.Println("Start save unit... ")
	if objJoint.Unsigned != "" {
		return errors.New(objJoint.Unsigned)
	}
	if Dbconn == nil {
		Dbconn = ReNewDbConn(dagconfig.DefaultConfig.DbPath)
	}
	obj_unit := objJoint.Unit
	obj_unit_byte, _ := json.Marshal(obj_unit)

	if err = Dbconn.Put(append(UNIT_PREFIX, obj_unit.Hash().Bytes()...), obj_unit_byte); err != nil {
		return
	}
	// add key in  unit_keys
	log.Println("add unit key:", AddUnitKeys(string(UNIT_PREFIX)+obj_unit.Hash().String()))

	if dagconfig.SConfig.Blight {
		// save  update utxo , message , transaction

	}

	if onDone != nil {
		onDone()
	}
	return
}
func GetUnitKeys() []string {
	var keys []string
	if Dbconn == nil {
		Dbconn = ReNewDbConn(dagconfig.DefaultConfig.DbPath)
	}
	if keys_byte, err := Dbconn.Get([]byte("array_units")); err != nil {
		log.Println("get units error:", err)
	} else {
		if err := json.Unmarshal(keys_byte, &keys); err != nil {
			log.Println("error:", err)
		}
	}
	return keys
}
func AddUnitKeys(key string) error {
	keys := GetUnitKeys()
	if len(keys) <= 0 {
		return errors.New("null keys.")
	}
	for _, v := range keys {

		if v == key {
			return errors.New("key is already exist.")
		}
	}
	keys = append(keys, key)
	if Dbconn == nil {
		Dbconn = ReNewDbConn(dagconfig.DefaultConfig.DbPath)
	}

	if err := Dbconn.Put([]byte("array_units"), ConvertBytes(keys)); err != nil {
		return err
	}
	return nil

}
func ConvertBytes(val interface{}) (re []byte) {
	var err error
	if re, err = json.Marshal(val); err != nil {
		log.Println("json.marshal error:", err)
	}
	return re[:]
}
func IsGenesisUnit(unit string) bool {
	return unit == constants.GENESIS_UNIT
}

func GetKeysWithTag(tag string) []string {
	var keys []string
	if Dbconn == nil {
		Dbconn = ReNewDbConn(dagconfig.DefaultConfig.DbPath)
	}
	if keys_byte, err := Dbconn.Get([]byte(tag)); err != nil {
		log.Println("get keys error:", err)
	} else {
		if err := json.Unmarshal(keys_byte, &keys); err != nil {
			log.Println("error:", err)
		}
	}
	return keys
}
func AddKeysWithTag(key, tag string) error {
	keys := GetKeysWithTag(tag)
	if len(keys) <= 0 {
		return errors.New("null keys.")
	}
	log.Println("keys:=", keys)
	for _, v := range keys {
		if v == key {
			return errors.New("key is already exist.")
		}
	}
	keys = append(keys, key)
	if Dbconn == nil {
		Dbconn = ReNewDbConn(dagconfig.DefaultConfig.DbPath)
	}

	if err := Dbconn.Put([]byte(tag), ConvertBytes(keys)); err != nil {
		return err
	}
	return nil

}
