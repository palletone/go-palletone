package getor

import (
	"github.com/palletone/go-palletone/dag/storage"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/filter"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
	"log"
)

var cache, handles int
var db *leveldb.DB

func Init() {
	var err error
	if cache < 16 {
		cache = 16
	}
	if handles < 16 {
		handles = 16
	}
	db, err = leveldb.OpenFile(storage.DBPath, &opt.Options{
		OpenFilesCacheCapacity: handles,
		BlockCacheCapacity:     cache / 2 * opt.MiB,
		WriteBuffer:            cache / 4 * opt.MiB, // Two of these are used internally
		Filter:                 filter.NewBloomFilter(10),
	})

	if err != nil {
		log.Println("opendb error:", err)

	}
}

func GetPrefix(prefix []byte) map[string][]byte {
	if db == nil {
		Init()
	}
	return getprefix(prefix)
}

// get prefix
func getprefix(prefix []byte) map[string][]byte {
	iter := db.NewIterator(util.BytesPrefix(prefix), nil)
	result := make(map[string][]byte)
	for iter.Next() {
		key := iter.Key()
		value := make([]byte, 0)
		//  直接赋值取得iter.Value()的最后一个指针，值是错的。
		result[string(key)] = append(value, iter.Value()...)
	}
	return result
}
