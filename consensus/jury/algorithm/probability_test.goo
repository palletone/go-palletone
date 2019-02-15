package algorithm

import (
	"testing"
	"github.com/tinychain/algorand/common"
	"math/rand"
	"time"
	"fmt"
)

func BenchmarkSubUsers(b *testing.B) {
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < b.N; i++ { //
		hash := common.BytesToHash(common.Uint2Bytes(rand.Uint64()))
		num := Selected(26, 1000, hash.Bytes())
		fmt.Printf("hash[%s], num[%d]", hash.String(), num)
	}
}

func TestBenchmarkSubUsers(t *testing.T) {
	fmt.Println("enter TestBenchmarkSubUsers")
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < 100; i++ { //b.N
		hash := common.BytesToHash(common.Uint2Bytes(rand.Uint64()))
		num := Selected(10, 10, hash.Bytes())
		//t.Logf("hash[%s], num[%d]", hash.String(), num)
		fmt.Printf("hash[%s], num[%d]\n", hash.String(), num)
	}
}

func TestSubUsersSingle(t *testing.T) {
	begin := time.Now().UnixNano()
	Selected(26, 1000, common.BytesToHash(common.Uint2Bytes(uint64(0))).Bytes())
	fmt.Printf("subusers cost %v\n", time.Now().UnixNano()-begin)
}
