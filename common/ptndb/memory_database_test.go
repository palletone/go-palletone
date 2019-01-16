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
 *  * @author PalletOne core developer  <dev@pallet.one>
 *  * @date 2018
 *
 */

package ptndb

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMemDatabase_NewIteratorWithPrefix(t *testing.T) {

	db, _ := NewMemDatabase()
	db.Put([]byte("a"), []byte("aaa"))
	db.Put([]byte("ab"), []byte("aaabbb"))
	db.Put([]byte("b"), []byte("bbb"))
	db.Put([]byte("c"), []byte("ccc"))
	db.Put([]byte("ba"), []byte("bbbaaa"))
	db.Put([]byte("abc"), []byte("abcabc"))
	it := db.NewIteratorWithPrefix([]byte("a"))
	itCount := 0
	for it.Next() {
		t.Logf("{%d} Key[%s], Value[%s]", itCount, it.Key(), it.Value())
		itCount++
	}
	assert.True(t, itCount == 3, "Result count not match")

	it2 := db.NewIteratorWithPrefix([]byte("x"))
	assert.False(t, it2.Next())

}
func TestMemDatabase_NewIterator(t *testing.T) {
	db, _ := NewMemDatabase()
	db.Put([]byte("a"), []byte("aaa"))
	db.Put([]byte("ab"), []byte("aaabbb"))
	db.Put([]byte("b"), []byte("bbb"))
	db.Put([]byte("c"), []byte("ccc"))
	db.Put([]byte("ba"), []byte("bbbaaa"))
	db.Put([]byte("abc"), []byte("abcabc"))
	it := db.NewIterator()
	itCount := 0
	for it.Next() {
		t.Logf("{%d} Key[%s], Value[%s]", itCount, it.Key(), it.Value())
		itCount++
	}
	assert.True(t, itCount == 6, "Result count not match")

}
