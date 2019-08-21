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
	"encoding/json"

	"github.com/palletone/go-palletone/dag/palletcache/cache"
	th_redis "github.com/palletone/go-palletone/dag/palletcache/redis"
)

var (
	Isredispool bool
)

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
