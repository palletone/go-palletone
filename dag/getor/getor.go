package getor

import (
	"github.com/palletone/go-palletone/dag/storage"
)

func GetPrefix(prefix []byte) map[string][]byte {
	if storage.Dbconn == nil {
		storage.Init()
	}
	return getprefix(prefix)
}

// get prefix
func getprefix(prefix []byte) map[string][]byte {
	iter := storage.Dbconn.NewIteratorWithPrefix(prefix)
	result := make(map[string][]byte)
	for iter.Next() {
		key := iter.Key()
		value := make([]byte, 0)
		//  直接赋值取得iter.Value()的最后一个指针，值是错的。
		result[string(key)] = append(value, iter.Value()...)
	}
	return result
}
