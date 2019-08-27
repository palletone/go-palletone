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
 *  * @date 2018-2019
 *
 */

package memunit

import (
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"reflect"
	"strings"
	"sync"

	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/dag/errors"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type Tempdb struct {
	kv      map[string][]byte //Key value
	deleted map[string]bool   //Deleted Keys
	db      ptndb.Database
	lock    sync.RWMutex
}

func NewTempdb(db ptndb.Database) (*Tempdb, error) {
	tempdb := &Tempdb{kv: make(map[string][]byte), deleted: make(map[string]bool), db: db}
	return tempdb, nil
}
func (db *Tempdb) Clear() {
	db.lock.Lock()
	defer db.lock.Unlock()

	db.kv = make(map[string][]byte)
	db.deleted = make(map[string]bool)
}

type KeyValue struct {
	Key, Value []byte
}

type TempdbIterator struct {
	result []KeyValue
	idx    int
}

func (i *TempdbIterator) Next() bool {
	i.idx++
	return i.idx < len(i.result)
}

func (i *TempdbIterator) Key() []byte {
	if i.idx == -1 {
		return nil
	}
	return i.result[i.idx].Key
}

func (i *TempdbIterator) Value() []byte {
	if i.idx == -1 {
		return nil
	}
	return i.result[i.idx].Value
}

//implement iterator interface
func (i *TempdbIterator) First() bool {
	return i.idx != -1
}
func (i *TempdbIterator) Last() bool {
	return i.idx != -1
}
func (i *TempdbIterator) Seek(key []byte) bool {
	if i.idx == -1 {
		return false
	}
	for j := 0; j <= i.idx; j++ {
		val := i.result[j]
		if reflect.DeepEqual(key, val.Key) {
			return true
		}
	}
	return false
}
func (i *TempdbIterator) Prev() bool {
	return i.idx != -1

}
func (i *TempdbIterator) Valid() bool {
	return i.idx != -1
}
func (i *TempdbIterator) Error() error {
	return nil
}
func (i *TempdbIterator) Release() {}
func (i *TempdbIterator) SetReleaser(releaser util.Releaser) {

}

func (db *Tempdb) NewIterator() iterator.Iterator {
	log.Warn("This function may be has a bug, it doesn't include temp data. --Devin")
	return db.db.NewIterator()
}

// NewIteratorWithPrefix returns a iterator to iterate over subset of database content with a particular prefix.
//这个最复杂，需要先去db数据库查询出map，然后把temp的列举出来，同样key的会被替换成新值，如果出现在del里面就删除，然后map转KeyValue数组
func (db *Tempdb) NewIteratorWithPrefix(prefix []byte) iterator.Iterator {
	result := getprefix(db.db, prefix)
	//Replace by tempdb newest value
	db.lock.RLock()
	for key := range db.kv {
		if strings.HasPrefix(key, string(prefix)) {
			result[key] = db.kv[key]
		}
	}
	//Delete some keys
	for key := range db.deleted {
		if strings.HasPrefix(key, string(prefix)) {
			delete(result, key)
		}
	}
	db.lock.RUnlock()
	kv := []KeyValue{}
	for k, v := range result {
		kv = append(kv, KeyValue{[]byte(k), v})
	}
	return &TempdbIterator{result: kv, idx: -1}
}

// get prefix
func getprefix(db ptndb.Database, prefix []byte) map[string][]byte {
	iter := db.NewIteratorWithPrefix(prefix)
	result := make(map[string][]byte)
	for iter.Next() {
		key := iter.Key()
		value := make([]byte, 0)
		// 请注意： 直接赋值取得iter.Value()的最后一个指针
		//result[*(*string)(unsafe.Pointer(&key))] = append(value, iter.Value()...)
		result[string(key)] = append(value, iter.Value()...)
	}
	return result
}
func (db *Tempdb) Put(key []byte, value []byte) error {
	db.lock.Lock()
	defer db.lock.Unlock()
	delete(db.deleted, string(key))
	//db.kv[string(key)] = common.CopyBytes(value)
	db.kv[string(key)] = value[:]
	return nil
}

func (db *Tempdb) Has(key []byte) (bool, error) {
	db.lock.RLock()
	defer db.lock.RUnlock()
	_, del := db.deleted[string(key)]
	if del {
		return false, nil
	}
	_, ok := db.kv[string(key)]
	if ok {
		return true, nil
	}
	return db.db.Has(key)
}

func (db *Tempdb) Get(key []byte) ([]byte, error) {
	db.lock.RLock()
	defer db.lock.RUnlock()
	_, del := db.deleted[string(key)]
	if del {
		return nil, errors.ErrNotFound
	}
	if entry, ok := db.kv[string(key)]; ok {
		return common.CopyBytes(entry), nil
	}
	return db.db.Get(key)
}

func (db *Tempdb) Delete(key []byte) error {
	db.lock.Lock()
	defer db.lock.Unlock()
	db.deleted[string(key)] = true
	delete(db.kv, string(key))
	return nil
}

func (db *Tempdb) Close() {}

func (db *Tempdb) NewBatch() ptndb.Batch {
	return &tempBatch{db: db}
}

func (db *Tempdb) Len() int { return len(db.kv) }

type kv struct {
	k, v []byte
	del  bool
}

type tempBatch struct {
	db     *Tempdb
	writes []kv
	size   int
}

func (b *tempBatch) Put(key, value []byte) error {
	b.writes = append(b.writes, kv{common.CopyBytes(key), common.CopyBytes(value), false})

	b.size += len(value)
	return nil

}

func (b *tempBatch) Delete(key []byte) error {
	b.writes = append(b.writes, kv{common.CopyBytes(key), nil, true})
	b.size += 1
	return nil
}

func (b *tempBatch) Write() error {
	b.db.lock.Lock()
	defer b.db.lock.Unlock()

	for _, kv := range b.writes {
		// b.db.kv[string(kv.k)] = kv.v
		if kv.del {
			b.db.deleted[string(kv.k)] = true
			delete(b.db.kv, string(kv.k))
		} else {
			b.db.kv[string(kv.k)] = kv.v
		}
	}
	return nil
}

func (b *tempBatch) ValueSize() int {
	return b.size
}

func (b *tempBatch) Reset() {
	b.writes = b.writes[:0]
	b.size = 0
}
