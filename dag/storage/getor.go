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

package storage

import (
	"encoding/json"
	"log"
	"unsafe"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/dag/modules"

	config "github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/common/rlp"
)

// @author Albert·Gou
func Retrieve(key string, v interface{}) error {
	//rv := reflect.ValueOf(v)
	//if rv.Kind() != reflect.Ptr || rv.IsNil() {
	//	return errors.New("an invalid argument, the argument must be a non-nil pointer")
	//}

	data, err := Get([]byte(key))
	if err != nil {
		return err
	}

	err = rlp.DecodeBytes(data, v)
	if err != nil {
		return err
	}

	return nil
}

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

func GetHeader(hash common.Hash, index uint64) *modules.Header {

	encNum := encodeBlockNumber(index)
	key := append(HEADER_PREFIX, encNum...)
	header_bytes, err := Get(append(key, hash.Bytes()...))
	// rlp  to  Header struct
	log.Println(err)
	var header modules.Header
	json.Unmarshal(header_bytes, &header)

	return &header
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
