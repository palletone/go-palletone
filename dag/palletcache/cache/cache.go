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
