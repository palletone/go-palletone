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
	"reflect"
	"sort"

	"github.com/palletone/go-palletone/dag/errors"
)

type vote interface {
	RegisterCandidates(candidates interface{})
	Exist(candidate interface{}) bool
	GetCandidates() []interface{}
	AddToBox(score uint64, tos interface{})
	GetScore(candidate interface{}) (uint64, error)
	GetVoteDetail() map[interface{}]uint64
	GetResult(number uint8) []interface{}
}

//BaseVote : virtual struct
type BaseVote struct {
	voteStatus map[interface{}]uint64
	elemType   reflect.Type
}

//GetScore : get given candidate's current score.
func (bv *BaseVote) GetScore(c interface{}) (uint64, error) {
	if bv.Exist(c) {
		return bv.voteStatus[c], nil
	}
	return 0, errors.New("getScore:error invalid Candidate ")
}

//GetCandidates : get a slice of all candidates.
func (bv *BaseVote) GetCandidates() []interface{} {
	res := make([]interface{}, 0)
	for c := range bv.voteStatus {
		res = append(res, c)
	}
	return res
}

//GetResult : get head n of vote result by descending order.
func (bv *BaseVote) GetResult(number uint8, val interface{}) bool {
	VoteSorter := NewMapSorter(bv.voteStatus)
	sort.Sort(VoteSorter)
	resNumber := resultNumber(number, uint8(len(VoteSorter)))
	rtyp := reflect.TypeOf(val).Elem()

	rs := reflect.MakeSlice(rtyp, 0, 0)
	for i := uint8(0); i < resNumber; i++ {
		rs = reflect.Append(rs, reflect.ValueOf(VoteSorter[i].object))
	}
	reflect.ValueOf(val).Elem().Set(rs)
	return true
}

//GetVoteDetail : get a map of vote status of all candidates/
func (bv *BaseVote) GetVoteDetail() map[interface{}]uint64 {
	return bv.voteStatus
}

//RegisterCandidates : init the vote & grant the right to vote for those candiates.
//candidates is slice for now.
func (bv *BaseVote) RegisterCandidates(candidates interface{}) {
	bv.elemType = reflect.TypeOf(candidates).Elem()
	bv.voteStatus = make(map[interface{}]uint64, 0)
	for _, c := range ToInterfaceSlice(candidates) {
		bv.voteStatus[c] = 0
	}
}

//Exist : check the existence of given candidate.
func (bv *BaseVote) Exist(c interface{}) bool {
	_, ok := bv.voteStatus[c]
	if !ok {
		fmt.Printf("candidate %v doesn't exist ", c)
	}
	return ok
}

//AddToBox : give n candidates score.
// candidates shoud be single element or slice
func (bv *BaseVote) AddToBox(score uint64, candidates interface{}) {

	switch reflect.ValueOf(candidates).Kind() {
	case reflect.Slice:
		for _, c := range ToInterfaceSlice(candidates) {
			if bv.Exist(c) {
				bv.voteStatus[c] += score
			}
		}
	default:
		if bv.Exist(candidates) {
			bv.voteStatus[candidates] += score
		}
	}

}

//SingleVote : one men , one ballot
type SingleVote struct {
	BaseVote
	voted map[interface{}]bool
}

//MultipleVote : one men, N ballots.
type MultipleVote struct {
	BaseVote
	voteLimit uint8
}
