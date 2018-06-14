package palletcache

import (
	"encoding/json"
	"github.com/palletone/go-palletone/configure"
	"github.com/palletone/go-palletone/dag/palletcache/cache"
	th_redis "github.com/palletone/go-palletone/dag/palletcache/redis"
	"log"
)

var (
	Isredispool bool
)

func init() {
	switch config.TConfig.CacheSource {
	case "redis":
		th_redis.Init()
		Isredispool = true
		log.Println("init redis ,", Isredispool)
	// case "cache":

	default:
		cache.Init()
	}
}
func Store(tag, key string, value interface{}, expire int) error {
	if Isredispool {
		err := th_redis.Store(tag, key, value)
		if err != nil {
			return err
		}
		return th_redis.Expire(key, expire)
	}
	val_byte, _ := json.Marshal(value)
	return cache.Store([]byte(key), val_byte, expire)
}

func Get(tag, key string) (interface{}, bool) {
	log.Println("Isredis?", Isredispool)
	if Isredispool {
		return th_redis.Get(tag, key)
	}
	return cache.Get([]byte(key))
}

func Del(tag, key string) bool {
	if Isredispool {
		return th_redis.RemoveItem(tag, key)
	}
	return cache.Del([]byte(key))
}
