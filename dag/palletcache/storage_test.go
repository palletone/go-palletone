package palletcache

import (
	"github.com/palletone/go-palletone/dag/palletcache/cache"
	"github.com/palletone/go-palletone/dag/palletcache/redis"
	"log"
	"testing"
)

var configs string = "redis"

func Init() {
	switch configs {
	case "redis":
		redis.Init()
		Isredispool = true
		log.Println("init redis ,", Isredispool)
	// case "cache":

	default:
		cache.Init()
	}
}
func TestCacheStore(t *testing.T) {

	err := Store("unit", "unit2", 87654321, 0)
	log.Println("re", err)

	re, has := cache.Get([]byte("unit2"))
	log.Println("result:", string(re), has) //  result: 87654321 true

}

func TestRedisStore(t *testing.T) {
	Init()
	err := Store("unit", "unit2", "hello2", 0)
	log.Println("re", err)

	re1, ok1 := Get("unit", "unit2")
	log.Println("re1:", ok1, re1) // true  hello2

	Del("unit", "unit2")
	re2, ok2 := Get("unit", "unit2")
	log.Println("re2:", ok2, re2) // false <nil>
}
