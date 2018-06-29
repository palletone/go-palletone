package storage

import (
	"encoding/json"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/dag/modules"
	"log"
	"unsafe"

	config "github.com/palletone/go-palletone/dag/dagconfig"
)

// get bytes
func Get(key []byte) ([]byte, error) {
	if Dbconn == nil {
		Dbconn = ReNewDbConn(config.DefaultConfig.DbPath)
	}
	return Dbconn.Get(key)
}

// get string
func GetString(key []byte) (string, error) {
	if Dbconn == nil {
		Dbconn = ReNewDbConn(config.DefaultConfig.DbPath)
	}
	if re, err := Dbconn.Get(key); err != nil {
		return "", err
	} else {
		return *(*string)(unsafe.Pointer(&re)), nil
	}
}

// get prefix: return maps
func GetPrefix(prefix []byte) map[string][]byte {
	if Dbconn == nil {
		Dbconn = ReNewDbConn(config.DefaultConfig.DbPath)
	}
	return getprefix(prefix)
}

// get prefix
func getprefix(prefix []byte) map[string][]byte {
	iter := Dbconn.NewIteratorWithPrefix(prefix)
	result := make(map[string][]byte)
	for iter.Next() {
		key := iter.Key()
		value := make([]byte, 0)
		// 请注意： 直接赋值取得iter.Value()的最后一个指针
		result[string(key)] = append(value, iter.Value()...)
	}
	return result
}

func GetUnit(hash common.Hash, index uint64) *modules.Unit {
	unit_bytes, err := Get(append(UNIT_PREFIX, hash.Bytes()...))
	log.Println(err)
	var unit modules.Unit
	json.Unmarshal(unit_bytes, &unit)

	return &unit
}

// func GetFreeUnits() []string {
// 	if list_bytes, err := Get([]byte(modules.FREEUNITS)); err != nil {
// 		log.Error("get free units error:" + err.Error())
// 	} else {
// 		var list []string
// 		json.Unmarshal(list_bytes, &list)
// 		if len(modules.FreeUnitslist) == 0 {
// 			return list
// 		}
// 		lfu := len(modules.FreeUnitslist)
// 		var free []string
// 		for _, v := range list {
// 			for j, f := range modules.FreeUnitslist {
// 				if v == f {
// 					break
// 				} else if v != f && j == (lfu-1) {
// 					free = append(free, v)
// 				}
// 			}
// 		}
// 		if len(free) > 0 {
// 			modules.FreeUnitslist = append(modules.FreeUnitslist, free...)
// 		}
// 	}
// 	return modules.FreeUnitslist
// }
