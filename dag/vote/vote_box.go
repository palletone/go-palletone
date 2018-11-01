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
	GetResult(number uint8) interface{}
	GetVoteDetail() map[interface{}]uint64
	RegisterCandidates(candidates interface{})
	AddToBox(score uint64, to interface{})
}

type candidateVote struct {
	candidate interface{}
	score     uint64
}

type voteSorter []candidateVote

func (vs voteSorter) Len() int {
	return len(vs)
}
func (vs voteSorter) Less(i, j int) bool {
	return vs[i].score > vs[j].score //Descending order
}
func (vs voteSorter) Swap(i, j int) {
	vs[i], vs[j] = vs[j], vs[i]
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

func (bv *BaseVote) GetResult(number uint8) interface{} {
	res := make([]interface{}, 0)
	voteSorter := make(voteSorter, 0)
	for c, s := range bv.voteStatus {
		voteSorter = append(voteSorter, candidateVote{candidate: c, score: s})
	}
	sort.Sort(voteSorter)
	if number == 0 || number > uint8(len(voteSorter)) {
		number = uint8(len(voteSorter))
	}
	for i := uint8(0); i < number; i++ {
		res = append(res, voteSorter[i].candidate)
	}
	return res
}

func (bv *BaseVote) GetVoteDetail() map[interface{}]uint64 {
	return bv.voteStatus
}

func (bv *BaseVote) RegisterCandidates(candidates interface{}) {
	bv.voteStatus = make(map[interface{}]uint64, 0)
	ic := candidates.([]interface{})
	for _, c := range ic {
		bv.voteStatus[c] = 0
	}
}
func (bv *BaseVote) AddToBox(score uint64, ShouldBeCList interface{}) {
	CList := ShouldBeCList.([]interface{})
	for _, c := range CList {
		if _, ok := bv.voteStatus[c]; ok {
			bv.voteStatus[c] += score
		} else {
			fmt.Println("addToBox:warning Candidate invalid")
		}
	}

}

type SingleVote struct {
	BaseVote
	voter []common.Address
}

type MultipleVote struct {
	BaseVote
}

type AddressMultipleVote struct {
	MultipleVote
}
type IAddressMultipleVote interface {
	vote
	Register(addresses map[common.Address]bool)
	Result(number uint8) []interface{}
}

func MapAddresses2Candidates(addresses map[common.Address]bool) map[interface{}]bool {
	res := make(map[interface{}]bool, 0)
	for addr, _ := range addresses {
		res[addr] = true
	}
	return res
}

func ListAddresses2Candidates(addresses []common.Address) *[]interface{} {
	res := make([]interface{}, 0)
	for _, addr := range addresses {
		res = append(res, addr)
	}
	return &res
}
func (amv *AddressMultipleVote) Result(number uint8) []interface{} {
	res := amv.GetResult(number)
	res2 := res.([]interface{})
	return res2
}
func (amv *AddressMultipleVote) Register(addresses []common.Address) {
	listInterface := ListAddresses2Candidates(addresses)
	amv.RegisterCandidates(*listInterface)
}

type TxHashMultipleVote struct {
	MultipleVote
}
