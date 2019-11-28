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

package memunit

import (
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTempdb_Get(t *testing.T) {
	db, _ := ptndb.NewMemDatabase()
	tmpdb, _ := NewTempdb(db)
	db.Put([]byte("A"), []byte("1"))
	db.Put([]byte("AB"), []byte("1"))
	a, _ := tmpdb.Get([]byte("A"))
	assert.Equal(t, a, []byte("1"))
	tmpdb.Put([]byte("A"), []byte("2"))
	a, _ = tmpdb.Get([]byte("A"))
	assert.Equal(t, a, []byte("2"))
	a, _ = db.Get([]byte("A"))
	assert.Equal(t, a, []byte("1"))

	hasB, _ := tmpdb.Has([]byte("AB"))
	assert.True(t, hasB)
	tmpdb.Delete([]byte("AB"))
	hasB, _ = tmpdb.Has([]byte("AB"))
	assert.False(t, hasB)

	tmpdb.Put([]byte("AC"), []byte("11"))
	it := tmpdb.NewIteratorWithPrefix([]byte("A"))
	for it.Next() {
		t.Logf("Key:%s,Value:%s", it.Key(), it.Value())
	}
	tmpdb.Put([]byte("AB"), []byte("3"))
	it = tmpdb.NewIteratorWithPrefix([]byte("A"))
	for it.Next() {
		t.Logf("Key:%s,Value:%s", it.Key(), it.Value())
	}
}
