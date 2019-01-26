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

package common

import (
	"github.com/coocood/freecache"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/storage"
	"io/ioutil"
	"os"
	"testing"
	"github.com/palletone/go-palletone/common"
)

func BenchmarkPropCacheRepository_RetrieveDynGlobalProp(b *testing.B) {
	dirname, err := ioutil.TempDir(os.TempDir(), "ptndb_test_")
	if err != nil {
		panic("failed to create test file: " + err.Error())
	}
	db, _ := ptndb.NewLDBDatabase(dirname, 0, 0)
	propdb := storage.NewPropertyDb(db)
	cache := freecache.NewCache(10 * 1024 * 1024)
	rep := NewPropCacheRepository(propdb, cache)
	data := &modules.DynamicGlobalProperty{common.Address{},true, 2, 3,4}
	rep.StoreDynGlobalProp(data)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		rep.RetrieveDynGlobalProp()
	}
}
func TestPropCacheRepository_RetrieveGlobalProp(t *testing.T) {
	db, _ := ptndb.NewMemDatabase()
	propdb := storage.NewPropertyDb(db)
	cache := freecache.NewCache(10 * 1024 * 1024)
	rep := NewPropCacheRepository(propdb, cache)
	rep.RetrieveGlobalProp()
}
