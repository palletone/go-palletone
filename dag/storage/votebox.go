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
 * @author PalletOne core developer Yiran <dev@pallet.one>
 * @date 2018
 */

package storage

import (
	"fmt"
	"github.com/palletone/go-palletone/common"
)

//Yiran
// this function connect multiple []byte keys to single []byte.
func KeyConnector(keys ...[]byte) []byte {
	var res []byte
	for _, key := range keys {
		res = append(res, key...)
	}
	return res
}

//Yiran
// print error if exist.
func ErrorLogHandler(err error, errType string) error {
	if err != nil {
		println(errType, "error", err.Error())
		return err
	}
	return nil
}

//Yiran
type VoteBox struct {
	Candidates map[common.Address]uint64
	voters     []common.Address
}

//Yiran
//Sort
func (box *VoteBox) HeadN(num uint) []common.Address {
	ResCandidates := make([]Candidate, num)
	for CurrCandidate, CurrScore := range box.Candidates {
		//insert if result set has space free.
		if uint(len(ResCandidates)) < num {
			ResCandidates = append(ResCandidates, Candidate{Address: CurrCandidate, VoteNumber: CurrScore})
		}
		// insert if current score greater than smallest elem score
		if CurrScore > ResCandidates[len(ResCandidates)-1].VoteNumber {
			for i, c := range ResCandidates {
				if CurrScore > c.VoteNumber {
					backcs := ResCandidates[i+1 : num-1]
					ResCandidates = append(ResCandidates[:i], Candidate{Address: CurrCandidate, VoteNumber: CurrScore})
					ResCandidates = append(ResCandidates, backcs...)
				}
			}
		}
	}
	// Trans []Candidate --> []Address
	ResAddress := make([]common.Address, num)
	for _, SortedCandidate := range ResCandidates {
		ResAddress = append(ResAddress, SortedCandidate.Address)
	}
	return ResAddress
}

//Yiran
func (box *VoteBox) Register(addresses []common.Address) {
	for _, address := range addresses {
		box.Candidates[address] = 1
	}
}

//Yiran
func (box *VoteBox) AddToBoxIfNotVoted(score uint64, voter common.Address, voteAddress common.Address) {
	for i, voted := range box.voters {
		// match, so already voted, do noting.
		if BytesEqual(voted.Bytes(), voter.Bytes()) {
			return
		}
		// no match until the end of the list, so add to VoteBox
		if i == len(box.voters)-1 {
			// 1. mark voter already voted
			box.voters = append(box.voters, voter)
			// 2. increase candidate vote number
			if _, ok := box.Candidates[voter]; ok { // voteAddress must belong to a candidate,
				box.Candidates[voteAddress] += score
			} else {
				fmt.Println("candidate address invalid")
			}
		}
	}
	return

}

//Yiran
func NewVoteBox() *VoteBox {
	return &VoteBox{
		Candidates: make(map[common.Address]uint64, 0),
		voters:     make([]common.Address, 0),
	}
}

//Yiran
type Candidate struct {
	Address    common.Address
	VoteNumber uint64
}

//Yiran
func BytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}

	if (a == nil) != (b == nil) { //[]int{} != []int(nil)
		return false
	}

	for i, v := range a {
		if v != b[i] {
			return false
		}
	}

	return true
}
