// Copyright 2014 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package ptndb

import (
	"encoding/binary"
	"errors"
	"reflect"
	"strings"
	"sync"

	"github.com/palletone/go-palletone/common"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/util"
)

var (
	ErrReleased    = errors.New("leveldb: resource already relesed")
	ErrHasReleaser = errors.New("leveldb: releaser already defined")
)

/*
 * This is a test memory database. Do not use for any production it does not get persisted
 */
type MemDatabase struct {
	db   map[string][]byte
	lock sync.RWMutex
}
type KeyValue struct {
	Key   []byte
	Value []byte
}
type MemIterator struct {
	result []KeyValue
	idx    int
}

func (i *MemIterator) Next() bool {
	i.idx++
	return i.idx < len(i.result)
}
func (i *MemIterator) Key() []byte {
	if i.idx == -1 {
		return nil
	}
	return i.result[i.idx].Key
}
func (i *MemIterator) Value() []byte {
	if i.idx == -1 {
		return nil
	}
	return i.result[i.idx].Value
}

//implement iterator interface
func (i *MemIterator) First() bool {
	return i.idx != -1
	// if i.idx == -1 {
	// 	return false
	// }
	// return true
}
func (i *MemIterator) Last() bool {
	return i.idx != -1
	// if i.idx == -1 {
	// 	return false
	// }
	// return true
}
func (i *MemIterator) Seek(key []byte) bool {
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
func (i *MemIterator) Prev() bool {
	return i.idx != -1
	// if i.idx == -1 {
	// 	return false
	// }
	// return true

}
func (i *MemIterator) Valid() bool {
	return i.idx != -1
	// if i.idx == -1 {
	// 	return false
	// }
	// return true
}
func (i *MemIterator) Error() error {
	return nil
}
func (i *MemIterator) Release() {}
func (i *MemIterator) SetReleaser(releaser util.Releaser) {

}

func (db *MemDatabase) NewIterator() iterator.Iterator {
	result := []KeyValue{}
	for key := range db.db {
		kv := KeyValue{[]byte(key), db.db[key]}
		result = append(result, kv)
	}
	return &MemIterator{result: result, idx: -1}
}

// NewIteratorWithPrefix returns a iterator to iterate over subset of database content with a particular prefix.
func (db *MemDatabase) NewIteratorWithPrefix(prefix []byte) iterator.Iterator {
	result := []KeyValue{}
	for key := range db.db {
		if strings.HasPrefix(key, string(prefix)) {
			kv := KeyValue{[]byte(key), db.db[key]}
			result = append(result, kv)
		}
	}
	return &MemIterator{result: result, idx: -1}
}
func NewMemDatabase() (*MemDatabase, error) {
	return &MemDatabase{
		db: make(map[string][]byte),
	}, nil
}

func NewMemDatabaseWithCap(size int) (*MemDatabase, error) {
	return &MemDatabase{
		db: make(map[string][]byte, size),
	}, nil
}

func (db *MemDatabase) Put(key []byte, value []byte) error {
	db.lock.Lock()
	defer db.lock.Unlock()
	db.db[string(key)] = common.CopyBytes(value)
	return nil
}

func (db *MemDatabase) Has(key []byte) (bool, error) {
	db.lock.RLock()
	defer db.lock.RUnlock()

	_, ok := db.db[string(key)]
	return ok, nil
}

func (db *MemDatabase) Get(key []byte) ([]byte, error) {
	db.lock.RLock()
	defer db.lock.RUnlock()
	if entry, ok := db.db[string(key)]; ok {
		return common.CopyBytes(entry), nil
	}
	return nil, errors.New("leveldb: not found")
}

func (db *MemDatabase) Keys() [][]byte {
	db.lock.RLock()
	defer db.lock.RUnlock()

	keys := [][]byte{}
	for key := range db.db {
		keys = append(keys, []byte(key))
	}
	return keys
}

func (db *MemDatabase) Delete(key []byte) error {
	db.lock.Lock()
	defer db.lock.Unlock()

	delete(db.db, string(key))
	return nil
}

func (db *MemDatabase) Close() {}

func (db *MemDatabase) NewBatch() Batch {
	return &memBatch{db: db}
}

func (db *MemDatabase) Len() int { return len(db.db) }

type kv struct {
	k, v []byte
	del  bool
}

type memBatch struct {
	db     *MemDatabase
	writes []kv
	size   int
}

func (b *memBatch) Put(key, value []byte) error {
	b.writes = append(b.writes, kv{common.CopyBytes(key), common.CopyBytes(value), false})
	b.size += len(value)
	return nil
}

func (b *memBatch) Delete(key []byte) error {
	b.writes = append(b.writes, kv{common.CopyBytes(key), nil, true})
	b.size += 1
	return nil
}

func (b *memBatch) Write() error {
	b.db.lock.Lock()
	defer b.db.lock.Unlock()

	for _, kv := range b.writes {
		if kv.del {
			delete(b.db.db, string(kv.k))
		} else {
			b.db.db[string(kv.k)] = kv.v
		}
	}
	return nil
}

func (b *memBatch) ValueSize() int {
	return b.size
}

func (b *memBatch) Reset() {
	b.writes = b.writes[:0]
	b.size = 0
}

// encodeBlockNumber encodes a block number as big endian uint64
func EncodeBlockNumber(number uint64) []byte {
	enc := make([]byte, 8)
	binary.BigEndian.PutUint64(enc, number)
	return enc
}
