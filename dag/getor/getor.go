package getor

import (
	"log"

	"github.com/palletone/go-palletone/dag/storage"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/filter"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
)

var cache, handles int
var DB *leveldb.DB

func Init() {
	var err error
	if cache < 16 {
		cache = 16
	}
	if handles < 16 {
		handles = 16
	}
	DB, err = leveldb.OpenFile(storage.DBPath, &opt.Options{
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
	if DB == nil {
		Init()
	}
	return getprefix(prefix)
}

// get prefix
func getprefix(prefix []byte) map[string][]byte {
	iter := DB.NewIterator(util.BytesPrefix(prefix), nil)
	result := make(map[string][]byte)
	for iter.Next() {
		key := iter.Key()
		value := make([]byte, 0)
		//  直接赋值取得iter.Value()的最后一个指针，值是错的。
		result[string(key)] = append(value, iter.Value()...)
	}
	return result
}
