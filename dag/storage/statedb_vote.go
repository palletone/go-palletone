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
 * @author PalletOne core developer YiRan <dev@pallet.one>
 * @date 2018
 */

package storage

import (
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/util"
	"github.com/palletone/go-palletone/dag/constants"
	"github.com/palletone/go-palletone/dag/modules"
)

func (statedb *StateDb) AppendVotedMediator(voter, mediator common.Address) error {
	//1. get current account info
	accountInfo, err := statedb.RetrieveAccountInfo(voter)
	if accountInfo == nil || err != nil {
		accountInfo = modules.NewAccountInfo()
	}

	accountInfo.VotedMediators[mediator] = true
	log.Debugf("Try to save voted mediator(%v) for account(%v)", mediator.Str(), voter.Str())

	return statedb.StoreAccountInfo(voter, accountInfo)
}

//UpdateVoterList YiRan@
func (statedb *StateDb) UpdateVoterList(voter common.Address, voteType uint8, term uint16) error {
	key := util.KeyConnector(constants.STATE_VOTER_LIST, []byte{byte(voteType)}, voter.Bytes())
	return StoreBytes(statedb.db, key, term)
}

//UpdateVoterList YiRan@
func (statedb *StateDb) GetVoterList(voteType uint8, MinTermLimit uint16) []common.Address {
	key := util.KeyConnector(constants.STATE_VOTER_LIST, []byte{byte(voteType)})
	bVoterMap := getprefix(statedb.db, key)
	res := []common.Address{}
	for voter, term := range bVoterMap {
		var pTerm *uint16
		rlp.DecodeBytes(term, pTerm)
		if *pTerm >= MinTermLimit {
			address, _ := common.StringToAddress(voter)
			res = append(res, address)
		}
	}
	return res
}
