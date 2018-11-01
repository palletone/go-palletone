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

package vote

import (
	"fmt"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/dag/errors"
	"sort"
)

type vote interface {
	GetScore(candidate interface{}) (uint64, error)
	GetCandidates() []interface{}
	GetResult(number uint8) []interface{}
	GetVoteDetail() map[interface{}]uint64
	RegisterCandidates(candidates interface{})
	AddToBox(score uint64, i interface{})
	AddNToBox(score uint64, to interface{})
}

type BaseVote struct {
	voteStatus map[interface{}]uint64
}

func (bv *BaseVote) GetScore(c interface{}) (uint64, error) {
	if score, ok := bv.voteStatus[c]; ok {
		return score, nil
	} else {
		return 0, errors.New("getScore:error invalid Candidate!")
	}
}

//getCandidates :
func (bv *BaseVote) GetCandidates() []interface{} {
	res := make([]interface{}, 0)
	for c, _ := range bv.voteStatus {
		res = append(res, c)
	}
	return res
}

func (bv *BaseVote) GetResult(resNumber uint8) []interface{} {
	res := make([]interface{}, 0)
	VoteSorter := NewMapSorter(bv.voteStatus)
	sort.Sort(VoteSorter)
	voteNumber := uint8(len(VoteSorter))
	if resNumber == 0 || resNumber > voteNumber {
		resNumber = voteNumber
	}
	for i := uint8(0); i < resNumber; i++ {
		res = append(res, VoteSorter[i].object)
	}
	return res
}

func (bv *BaseVote) GetVoteDetail() map[interface{}]uint64 {
	return bv.voteStatus
}

func (bv *BaseVote) RegisterCandidates(ShouldBeIList interface{}) {
	bv.voteStatus = make(map[interface{}]uint64, 0)
	IList := ShouldBeIList.([]interface{})
	for _, c := range IList {
		bv.voteStatus[c] = 0
	}
}
func (bv *BaseVote) AddNToBox(score uint64, ShouldBeIList interface{}) {
	IList := ShouldBeIList.([]interface{})
	for _, c := range IList {
		if _, ok := bv.voteStatus[c]; ok {
			bv.voteStatus[c] += score
		} else {
			fmt.Println("addToBox:warning candidate invalid")
		}
	}

}

func (bv *BaseVote) AddToBox(score uint64, i interface{}) {
	if _, ok := bv.voteStatus[i]; ok {
		bv.voteStatus[i] += score
	} else {
		fmt.Println("addToBox:warning candidate invalid")
	}

}

type SingleVote struct {
	BaseVote
	voter map[interface{}]bool
}

type MultipleVote struct {
	BaseVote
}

type IAddressMultipleVote interface {
	vote
	Register(addresses []common.Address)
	Result(number uint8) []common.Address
	Add(addresses []common.Address, score uint64)
}

type ITxhashMultipleVote interface {
	vote
}
