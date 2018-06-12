package storage

import (
	//"github.com/facebookgo/ensure"
	"github.com/tecbot/gorocksdb"
	"log"
	"testing"
)

func TestRocks(t *testing.T) {
	log.Println("ok")
	opts := gorocksdb.NewDefaultOptions()
	// test the ratelimiter
	rateLimiter := gorocksdb.NewRateLimiter(1024, 100*1000, 10)
	opts.SetRateLimiter(rateLimiter)
	opts.SetCreateIfMissing(true)

	db, err := gorocksdb.OpenDb(opts, "/Users/jay/code/gocode/src/palletone/bin/rocksdb")
	//opts_w := gorocksdb.NewDefaultWriteOptions()
	//rateLimiter := gorocksdb.NewRateLimiter(1024, 100*1000, 10)
	log.Println("open_db error:", err)
	//ensure.Nil(t, db.Put(opts_w, []byte("123"), []byte("jay")))
	opts_r := gorocksdb.NewDefaultReadOptions()
	v, e1 := db.Get(opts_r, []byte("123"))
	log.Println("value:", string(v.Data()), e1)
}
