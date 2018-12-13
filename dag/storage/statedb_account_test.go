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
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/stretchr/testify/assert"
)

func TestStateDb_AccountInfo(t *testing.T) {
	db, _ := ptndb.NewMemDatabase()
	l := log.NewTestLog()
	statedb := NewStateDb(db, l)

	addr, _ := common.StringToAddress("P173mPBwP1kXmfpg4p7rzZ5XRsGN1G1WQC8")
	info, err := statedb.RetrieveAccountInfo(addr)
	assert.NotNil(t, err)
	t.Logf("correct throw error:%s", err)
	info.PtnBalance = 12345
	info.VotedMediator = append(info.VotedMediator, addr)

	//Votes: []vote.VoteInfo{{Contents: , VoteType: vote.TYPE_MEDIATOR}}
	err = statedb.StoreAccountInfo(addr, info)
	assert.Nil(t, err)
	info2, err := statedb.RetrieveAccountInfo(addr)
	assert.NotNil(t, info2)
	assert.Equal(t, info.PtnBalance, info2.PtnBalance)
}
