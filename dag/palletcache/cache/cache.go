package cache

import (
	"github.com/coocood/freecache"
)

var cache0 *freecache.Cache

//  init cache.
func Init() {
	cache0 = freecache.NewCache(200 * 1024 * 1024)

}

func Store(key, val []byte, expir int) error {
	if cache0 == nil {
		cache0 = freecache.NewCache(200 * 1024 * 1024)
	}
	return cache0.Set(key, val, expir)

}

func Get(key []byte) ([]byte, bool) {
	if cache0 == nil {
		cache0 = freecache.NewCache(200 * 1024 * 1024)
	}
	if re, err := cache0.Get(key); err != nil {
		return re, false
	} else {
		return re, true
	}
}

func Del(key []byte) bool {
	if cache0 == nil {
		cache0 = freecache.NewCache(200 * 1024 * 1024)
	}
	return cache0.Del(key)
}
