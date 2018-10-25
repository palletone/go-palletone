package storage

import (
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/dag/constants"
	"github.com/palletone/go-palletone/dag/errors"
	"github.com/palletone/go-palletone/dag/modules"
)

//UpdateMediatorVote YiRan@
func (statedb *StateDb) UpdateMediatorVote(voter common.Address, candidates []common.Address, mode uint8) error {
	//1. get current account info
	accountInfo, err := statedb.GetAccountInfo(voter)
	if err != nil {
		return err
	}
	newVotes := []modules.VoteInfo{}
	mediatorVotes := []modules.VoteInfo{}
	//2. split vote by type
	for _, voteInfo := range accountInfo.Votes {
		if voteInfo.VoteType != 0 {
			newVotes = append(newVotes, voteInfo)
		} else {
			mediatorVotes = append(mediatorVotes, voteInfo)
		}
	}
	switch {

	case mode == 0: //[Replace all]
		//3. append new data
		for _, candidate := range candidates {
			newVotes = append(newVotes, modules.VoteInfo{VoteType: 0, VoteContent: candidate.Bytes()})
		}

	case mode == 1: //[Edit]
		//3. format examination
		if len(candidates)%2 != 0 {
			return errors.New("invalid candidates number, must be even")
		}
		// 4. replace appointed data
		stride := len(candidates)
		for i := 0; i < stride; i++ {
			for j, voteInfo := range mediatorVotes {
				if BytesEqual(voteInfo.VoteContent, candidates[i].Bytes()) {
					mediatorVotes[j].VoteContent = candidates[i+stride].Bytes()
				}
			}
		}
		//5. merge to newVotes
		newVotes = append(newVotes, mediatorVotes...)

	case mode == 2: //[Delete]
		// 3. copy votes which not in address list
		resMediatorVotes := []modules.VoteInfo{}
		pairFlag := false
		for _, mediatorVoteInfo := range mediatorVotes {
			for _, candidate := range candidates {
				if BytesEqual(candidate.Bytes(), mediatorVoteInfo.VoteContent) {
					pairFlag = true
				}
			}
			if pairFlag == false {
				resMediatorVotes = append(resMediatorVotes, mediatorVoteInfo)
			}
			pairFlag = false
		}
		// 4. merge vote
		newVotes = append(newVotes, resMediatorVotes...)

	default:
		return errors.New("Invalid mode")

	}
	//$. save new account info
	accountInfo.Votes = newVotes
	err = statedb.SaveAccountInfo(voter, accountInfo)
	if err != nil {
		return err
	}
	return nil
}

//UpdateVoterList YiRan@ TODO:SAVE BY TERM
func (statedb *StateDb) UpdateVoterList(voter common.Address) error {
	key := KeyConnector(constants.STATE_VOTER_LIST, voter.Bytes())
	return StoreBytes(statedb.db, key, 0)
}

func (statedb *StateDb) GetAccountMediatorVote(voterAddress common.Address) ([]common.Address, uint64, error) {
	// 1. get account info
	accountInfo, err := statedb.GetAccountInfo(voterAddress)
	if err != nil {
		return nil, 0, err
	}
	// 2. get mediator vote
	mediatorVotes := []common.Address{}
	for _, voteInfo := range accountInfo.Votes {
		if voteInfo.VoteType == 0 {
			mediatorVotes = append(mediatorVotes, common.BytesToAddress(voteInfo.VoteContent))
		}
	}
	// 3. get weight
	weight := accountInfo.PtnBalance

	return mediatorVotes, weight, nil
}

//GetSortedVote YiRan@
func (statedb *StateDb) GetSortedVote(ReturnNumber uint) ([]common.Address, error) {

	// 1. get voter list
	key := constants.STATE_VOTER_LIST
	bVoterMap := getprefix(statedb.db, key)
	voteBox := NewVoteBox()

	// 2. get candidate list
	addresses, err := statedb.GetCandidateMediatorAddrList()
	if err != nil { // get candidates address list error
		return nil, err
	}

	// 3. register candidate
	voteBox.Register(addresses)

	// 4. collect ballot
	for voter, _ := range bVoterMap {
		voterAddress, err := common.StringToAddress(voter)
		if err != nil {
			return nil, err
		}
		to, weight, err := statedb.GetAccountMediatorVote(voterAddress)
		if err != nil {
			return nil, err
		}
		voteBox.AddToBox(weight, to)
	}

	// $. return elected mediator
	return voteBox.HeadN(ReturnNumber), nil

}
