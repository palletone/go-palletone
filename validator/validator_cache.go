/*
 *
 *    This file is part of go-palletone.
 *    go-palletone is free software: you can redistribute it and/or modify
 *    it under the terms of the GNU General Public License as published by
 *    the Free Software Foundation, either version 3 of the License, or
 *    (at your option) any later version.
 *    go-palletone is distributed in the hope that it will be useful,
 *    but WITHOUT ANY WARRANTY; without even the implied warranty of
 *    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *    GNU General Public License for more details.
 *    You should have received a copy of the GNU General Public License
 *    along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
 * /
 *
 *  * @author PalletOne core developer <dev@pallet.one>
 *  * @date 2018-2019
 *
 */

package validator

import (
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/dag/palletcache"
	"sync"
)

//如果已经验证通过的对象，那么不需要重复做全部验证
type ValidatorCache struct {
	sync.RWMutex
	cache      palletcache.ICache
	maxEntries uint
}

var prefix = []byte("VA")

func NewValidatorCache(maxEntries uint, cache palletcache.ICache) *ValidatorCache {
	return &ValidatorCache{cache: cache, maxEntries: maxEntries}
}
func (s *ValidatorCache) Exists(sigHash common.Hash) bool {
	if s.cache == nil {
		return false
	}
	s.RLock()
	_, err := s.cache.Get(append(prefix, sigHash.Bytes()...))
	if err != nil {
		return false
	}
	s.RUnlock()
	return true
}
func (s *ValidatorCache) Add(sigHash common.Hash) {
	if s.cache == nil {
		return
	}
	s.Lock()
	defer s.Unlock()

	if s.maxEntries <= 0 {
		return
	}
	s.cache.Set(append(prefix, sigHash.Bytes()...), []byte{0x1}, 60)
}
