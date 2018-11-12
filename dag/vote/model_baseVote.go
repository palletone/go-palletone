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
	"errors"
	"reflect"
	"sort"
)

//BaseVoteModel : virtual struct
type BaseVoteModel struct {
	candidatesStatus map[interface{}]uint64
	elemType         reflect.Type
}

func NewBaseVoteModel() *BaseVoteModel {
	m := &BaseVoteModel{}
	m.candidatesStatus = make(map[interface{}]uint64, 0)
	return m
}

//GetScore : get given candidate's current score.
func (bv *BaseVoteModel) GetCandidateScore(c interface{}) (uint64, error) {
	if bv.ExistCandidate(c) {
		return bv.candidatesStatus[c], nil
	}
	return 0, errors.New("getScore:error invalid Candidate ")
}

//GetCandidates : get a slice of all candidates.
func (bv *BaseVoteModel) GetCandidates() []interface{} {
	res := make([]interface{}, 0)
	for c := range bv.candidatesStatus {
		res = append(res, c)
	}
	return res
}

//GetResult : get head n of vote result by descending order.
func (bv *BaseVoteModel) GetResult(number uint8, val interface{}) bool {
	VoteSorter := NewMapSorter(bv.candidatesStatus)
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
func (bv *BaseVoteModel) GetVoteDetail() map[interface{}]uint64 {
	return bv.candidatesStatus
}

//RegisterCandidates : init the vote & grant the right to vote for those candiates.
//candidates is slice for now.
func (bv *BaseVoteModel) RegisterCandidates(candidates interface{}) {
	bv.elemType = reflect.TypeOf(candidates).Elem()

	for _, c := range ToInterfaceSlice(candidates) {
		bv.candidatesStatus[c] = 0
	}
}

//Exist : check the existence of given candidate.
func (bv *BaseVoteModel) ExistCandidate(c interface{}) bool {
	_, ok := bv.candidatesStatus[c]
	if !ok {
		//fmt.Printf("candidate %v doesn't exist ", c)
	}
	return ok
}

//AddToBox : give n candidates score.
// candidates shoud be single element or slice
func (bv *BaseVoteModel) AddToBox(score uint64, candidates interface{}) {

	switch reflect.ValueOf(candidates).Kind() {
	case reflect.Slice:

			for _, c := range ToInterfaceSlice(candidates) {
				if bv.ExistCandidate(c) {
					bv.candidatesStatus[c] += score
				}
			}


	default:
		if bv.ExistCandidate(candidates) {
			bv.candidatesStatus[candidates] += score
		}
	}

}
