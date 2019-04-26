package light

import (
	"github.com/coocood/freecache"
	"github.com/palletone/go-palletone/common"
)

var cache *freecache.Cache

func InitSPVCache(size int) {
	cache = freecache.NewCache(200 * 1024 * 1024)
}

func SetProofPath(txhash, unithash common.Hash, data proofsData) {
	//cache.Set(txhash, unithash)
	//cache.Set(unithash, data)
}
