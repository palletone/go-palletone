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
 *  * @date 2018-2019
 *
 */
package storage

import (
	"testing"

	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/stretchr/testify/assert"
)

func TestStateDb_Version(t *testing.T) {
	db, _ := ptndb.NewMemDatabase()
	version := &modules.StateVersion{Height: &modules.ChainIndex{Index: 123, AssetID: modules.PTNCOIN}, TxIndex: 1}
	key := "Name"
	value := []byte("Devin")
	err := storeBytesWithVersion(db, []byte(key), version, value)
	assert.Nil(t, err)
	value1, version1, err := retrieveWithVersion(db, []byte(key))
	assert.Nil(t, err)
	assert.Equal(t, value, value1)
	assert.Equal(t, version.String(), version1.String())
}
