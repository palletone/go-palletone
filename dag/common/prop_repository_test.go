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
	"io/ioutil"
	"os"
	"testing"

	"encoding/json"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/crypto"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/storage"
)

func BenchmarkPropRepository_RetrieveDynGlobalProp(b *testing.B) {
	dirname, err := ioutil.TempDir(os.TempDir(), "ptndb_test_")
	if err != nil {
		panic("failed to create test file: " + err.Error())
	}
	db, _ := ptndb.NewLDBDatabase(dirname, 0, 0)
	propdb := storage.NewPropertyDb(db)

	rep := NewPropRepository(propdb)
	data := modules.NewDynGlobalProp()
	rep.StoreDynGlobalProp(data)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		rep.RetrieveDynGlobalProp()
	}
}

func TestShuffle(t *testing.T) {
	addr1 := common.NewAddress(crypto.Hash160([]byte("1")), common.PublicKeyHash)
	addr2 := common.NewAddress(crypto.Hash160([]byte("2")), common.PublicKeyHash)
	addr3 := common.NewAddress(crypto.Hash160([]byte("3")), common.PublicKeyHash)
	addr4 := common.NewAddress(crypto.Hash160([]byte("4")), common.PublicKeyHash)
	addr5 := common.NewAddress(crypto.Hash160([]byte("5")), common.PublicKeyHash)

	for i := 0; i < 10; i++ {
		addrs := []common.Address{addr1, addr2, addr3, addr4, addr5}
		shuffleMediators(addrs, uint64(i))
		addrJs, _ := json.Marshal(addrs)
		t.Logf("i:%d,addr:%s", i, addrJs)
	}
}
