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
	"errors"
	"unsafe"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/common/rlp"
	"github.com/palletone/go-palletone/dag/modules"
)

const missingNumber = uint64(0xffffffffffffffff)

// DatabaseReader wraps the Get method of a backing data store.
type DatabaseReader interface {
	Get(key []byte) (value []byte, err error)
	NewIterator() ptndb.Iterator
	NewIteratorWithPrefix(prefix []byte) ptndb.Iterator
}

// @author Albert·Gou
func retrieve(db ptndb.Database, key []byte, v interface{}) error {
	data, err := db.Get(key)
	if err != nil {
		return err
	}

	err = rlp.DecodeBytes(data, v)
	if err != nil {
		return err
	}

	return nil
}
func retrieveWithVersion(db ptndb.Database, key []byte) (*modules.StateVersion, []byte, error) {
	data, err := db.Get(key)
	if err != nil {
		return nil, nil, err
	}
	return splitValueAndVersion(data)
}

//将Statedb里的Value分割为Version和用户数据
func splitValueAndVersion(data []byte) (*modules.StateVersion, []byte, error) {
	verBytes := data[:29]
	objData := data[30:]

	version := &modules.StateVersion{}
	version.SetBytes(verBytes)
	return version, objData, nil
}

// get string
func getString(db ptndb.Database, key []byte) (string, error) {
	if re, err := db.Get(key); err != nil {
		return "", err
	} else {
		return *(*string)(unsafe.Pointer(&re)), nil
	}
}

// get prefix
func getprefix(db DatabaseReader, prefix []byte) map[string][]byte {
	iter := db.NewIteratorWithPrefix(prefix)
	result := make(map[string][]byte)
	for iter.Next() {
		key := iter.Key()
		value := make([]byte, 0)
		// 请注意： 直接赋值取得iter.Value()的最后一个指针
		result[*(*string)(unsafe.Pointer(&key))] = append(value, iter.Value()...)
	}
	return result
}

// get row count by prefix
func getCountByPrefix(db DatabaseReader, prefix []byte) int {
	iter := db.NewIteratorWithPrefix(prefix)
	count := 0
	for iter.Next() {
		count++
	}
	return count
}
func GetContractRlp(db DatabaseReader, id common.Hash) (rlp.RawValue, error) {
	if common.EmptyHash(id) {
		return nil, errors.New("the filed not defined")
	}
	con_bytes, err := db.Get(append(CONTRACT_PREFIX, id[:]...))
	if err != nil {
		return nil, err
	}
	return con_bytes, nil
}

// GetAdddrTransactionsHash
func GetAddrTransactionsHash(db DatabaseReader, addr string) ([]common.Hash, error) {
	data, err := db.Get(append(AddrTransactionsHash_Prefix, []byte(addr)...))
	if err != nil {
		return []common.Hash{}, err
	}
	hashs := make([]common.Hash, 0)
	if err := rlp.DecodeBytes(data, hashs); err != nil {
		return []common.Hash{}, err
	}
	return hashs, nil
}
