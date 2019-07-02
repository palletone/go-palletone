/*
 *
 *     This file is part of go-palletone.
 *     go-palletone is free software: you can redistribute it and/or modify
 *     it under the terms of the GNU General Public License as published by
 *     the Free Software Foundation, either version 3 of the License, or
 *     (at your option) any later version.
 *     go-palletone is distributed in the hope that it will be useful,
 *     but WITHOUT ANY WARRANTY; without even the implied warranty of
 *     MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *     GNU General Public License for more details.
 *     You should have received a copy of the GNU General Public License
 *     along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
 * /
 *
 *  * @author PalletOne core developers <dev@pallet.one>
 *  * @date 2018
 *
 */

package storage

import (
	"testing"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/stretchr/testify/assert"
)

func TestStateDb_AccountInfo(t *testing.T) {
	db, _ := ptndb.NewMemDatabase()
	statedb := NewStateDb(db)
	addr, _ := common.StringToAddress("P173mPBwP1kXmfpg4p7rzZ5XRsGN1G1WQC8")
	// store
	key := "Name"
	devin := []byte("Devin Zeng")
	version := &modules.StateVersion{TxIndex: 2, Height: modules.NewChainIndex(modules.PTNCOIN, 123)}
	writeSet := &modules.AccountStateWriteSet{IsDelete: false, Key: key, Value: devin}
	statedb.SaveAccountState(addr, writeSet, version)

	state, err := statedb.GetAccountState(addr, key)
	assert.Nil(t, err)
	assert.Equal(t, state.Value, devin)
	allState, err := statedb.GetAllAccountStates(addr)
	assert.Nil(t, err)
	assert.Equal(t, len(allState), 1)
	assert.Equal(t, allState[key].Value, devin)
}

func TestStateDb_GetAccountBalance(t *testing.T) {
	db, _ := ptndb.NewMemDatabase()
	statedb := NewStateDb(db)
	addr, _ := common.StringToAddress("P173mPBwP1kXmfpg4p7rzZ5XRsGN1G1WQC8")
	err := statedb.UpdateAccountBalance(addr, 123)
	assert.Nil(t, err)
	balance := statedb.GetAccountBalance(addr)
	assert.Equal(t, balance, uint64(123))
}
