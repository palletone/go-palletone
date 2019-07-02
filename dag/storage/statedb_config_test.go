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
 *  * @date 2018
 *
 */

package storage

import (
	"testing"

	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/stretchr/testify/assert"
)

func MockStateMemDb() *StateDb {
	db, _ := ptndb.NewMemDatabase()
	statedb := NewStateDb(db)
	return statedb
}

func TestStateDb_GetPartitionChains(t *testing.T) {
	db := MockStateMemDb()
	partitions, err := db.GetPartitionChains()
	assert.Nil(t, err)
	assert.Equal(t, 0, len(partitions))
}

func TestSaveAndGetConfig(t *testing.T) {
	db := MockStateMemDb()
	version := &modules.StateVersion{Height: &modules.ChainIndex{Index: 123}, TxIndex: 1}
	err := db.SaveSysConfigContract("key1", nil, version)
	assert.Nil(t, err)
	data, version, err := db.getSysConfigContract("key1")
	assert.Nil(t, err)
	t.Log(data)
	assert.NotNil(t, version)

}
