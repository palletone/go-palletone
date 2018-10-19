/*
   This file is part of go-palletone.
   go-palletone is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.
   go-palletone is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.
   You should have received a copy of the GNU General Public License
   along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
*/
/*
 * @author PalletOne core developers <dev@pallet.one>
 * @date 2018
 */

package palletcache

import (
	"log"
	"testing"

	"github.com/palletone/go-palletone/dag/palletcache/cache"
	"github.com/palletone/go-palletone/dag/palletcache/redis"
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
