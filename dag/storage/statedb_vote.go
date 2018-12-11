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
	"encoding/json"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/rlp"
	"github.com/palletone/go-palletone/common/util"
	"github.com/palletone/go-palletone/dag/constants"
)

//UpdateMediatorVote YiRan@
func (statedb *StateDb) UpdateMediatorVote(voter common.Address, candidates []byte) error {
	//1. get current account info
	accountInfo, err := statedb.GetAccountInfo(voter)
	if err != nil {
		return err
	}
	accountInfo.MediatorVote = common.BytesToAddress(candidates)
	statedb.logger.Debugf("Try to save mediator vote result{%s} for address:%s", accountInfo.MediatorVote.Str(), voter.Str())
	//
	//newVotes := []vote.VoteInfo{}
	//mediatorVotes := []vote.VoteInfo{}
	////2. split vote by type
	//for _, voteInfo := range accountInfo.Votes {
	//	if voteInfo.VoteType != 0 {
	//		newVotes = append(newVotes, voteInfo)
	//	} else {
	//		mediatorVotes = append(mediatorVotes, voteInfo)
	//	}
	//}
	//switch {
	//
	//case mode == 0: //[Replace all]
	//	//3. append new data
	//	for _, candidate := range candidates {
	//		newVotes = append(newVotes, vote.VoteInfo{VoteType: vote.TYPE_MEDIATOR, VoteContent: candidate.Bytes()})
	//	}
	//
	//case mode == 1: //[Replace]
	//	//3. format examination
	//	if len(candidates)%2 != 0 {
	//		return errors.New("invalid candidates number, must be even")
	//	}
	//	// 4. replace appointed data
	//	stride := len(candidates)
	//	for i := 0; i < stride; i++ {
	//		for j, voteInfo := range mediatorVotes {
	//			if util.BytesEqual(voteInfo.VoteContent, candidates[i].Bytes()) {
	//				mediatorVotes[j].VoteContent = candidates[i+stride].Bytes()
	//			}
	//		}
	//	}
	//	//5. merge to newVotes
	//	newVotes = append(newVotes, mediatorVotes...)
	//
	//case mode == 2: //[Delete]
	//	// 3. copy votes which not in address list
	//	resMediatorVotes := []vote.VoteInfo{}
	//	pairFlag := false
	//	for _, mediatorVoteInfo := range mediatorVotes {
	//		for _, candidate := range candidates {
	//			if util.BytesEqual(candidate.Bytes(), mediatorVoteInfo.VoteContent) {
	//				pairFlag = true
	//			}
	//		}
	//		if pairFlag == false {
	//			resMediatorVotes = append(resMediatorVotes, mediatorVoteInfo)
	//		}
	//		pairFlag = false
	//	}
	//	// 4. merge vote
	//	newVotes = append(newVotes, resMediatorVotes...)
	//case mode == 3: //[Delete all]
	//default:
	//	return errors.New("Invalid mode")
	//
	//}
	////$. save new account info
	//accountInfo.Votes = newVotes
	err = statedb.SaveAccountInfo(voter, accountInfo)
	if err != nil {
		return err
	}
	return nil
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

//func (statedb *StateDb) GetAccountMediatorVote(voterAddress common.Address) ([]common.Address, uint64, error) {
//	// todo
//	// 1. get account info
//	accountInfo, err := statedb.GetAccountInfo(voterAddress)
//	if err != nil {
//		return nil, 0, err
//	}
//	// 2. get mediator vote
//	mediatorVotes := []common.Address{}
//	for _, voteInfo := range accountInfo.Votes {
//		if voteInfo.VoteType == vote.TYPE_MEDIATOR {
//			mediatorVotes = append(mediatorVotes, common.BytesToAddress(voteInfo.Contents))
//		}
//	}
//	// 3. get weight
//	weight := accountInfo.PtnBalance
//
//	return mediatorVotes, weight, nil
//}

func (statedb *StateDb) GetSortedMediatorVote(returnNumber int) (map[string]uint64, error) {
	// todo
	result := make(map[string]uint64)
	for _, info := range statedb.getAllAccountInfo() {
		if !info.MediatorVote.Equal(common.Address{}) && len(info.MediatorVote) > 0 {
			addrs := []string{}
			json.Unmarshal(info.MediatorVote[:], &addrs)
			for _, addr := range addrs {
				if val, ok := result[addr]; ok {
					result[addr] = val + info.PtnBalance
				} else {
					result[addr] = info.PtnBalance
				}
			}
		}
	}
	//TODO sort
	return result, nil

	//voteBox := vote.NewBaseVote()
	//// 1. get voter list
	//voterList := statedb.GetVoterList(voteType, minTermLimit)
	//
	//// 2. register candidate
	//addresses := statedb.GetMediators()
	//voteBox.RegisterCandidates(addresses)
	//
	//// 3. collect ballot
	//for _, voterAddress := range voterList {
	//	to, weight, err := statedb.GetAccountMediatorVote(voterAddress)
	//	if err != nil {
	//		return nil, err
	//	}
	//	voteBox.AddToBox(weight, to)
	//}
	//
	//// $. return elected mediator
	//res := make([]common.Address, 0)
	//voteBox.GetResult(ReturnNumber, &res)
	//return res, nil

}

//CreateUserVote YiRan@
func (statedb *StateDb) CreateUserVote(voter common.Address, detail [][]byte, bHash []byte) error {
	key := util.KeyConnector(constants.CREATE_VOTE_PREFIX, bHash)
	value := detail
	return StoreBytes(statedb.db, key, value)
}
