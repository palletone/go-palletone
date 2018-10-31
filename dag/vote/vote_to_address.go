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
	"sort"
)

//YiRan
type AddressVote struct {
	Address common.Address
	Score   uint64
}

//YiRan
type AddressVoteBox struct {
	Candidates map[common.Address]uint64
	voters     []common.Address
}

//YiRan
func NewAddressVoteBox() *AddressVoteBox {
	return &AddressVoteBox{
		Candidates: make(map[common.Address]uint64, 0),
		voters:     make([]common.Address, 0),
	}
}

//YiRan
//return sorted data of given number
func (box *AddressVoteBox) HeadN(num uint) []common.Address {
	Candidates := NewAddressVoteBoxSorter(box.Candidates)
	sort.Sort(Candidates)

	ResAddresses := make([]common.Address, num)
	for i := uint(0); i < num; i++ {
		ResAddresses = append(ResAddresses, Candidates[i].Address)
	}
	return ResAddresses
}

// This function must be used with AddToBoxIfNotVoted
func (box *AddressVoteBox) InitBlackList(addresses []common.Address) {
	for _, address := range addresses {
		box.voters = append(box.voters, address)
	}
}

//YiRan
//Initialize the score for the given accounts
func (box *AddressVoteBox) Register(addresses map[common.Address]bool, initialValue uint64) {
	for address, _ := range addresses {
		box.Candidates[address] = initialValue
	}
}

func (box *AddressVoteBox) IsVoted(voter common.Address) bool {
	for _, voted := range box.voters {
		if common.AddressEqual(voted, voter) {
			return true
		}
	}
	return false
}

//@YiRan
//Vote Rule:
//1.The target of the vote is the candidate
func (box *AddressVoteBox) AddToBox(Weight uint64, to []common.Address) {
	for _, candidate := range to {
		if _, ok := box.Candidates[candidate]; ok {
			box.Candidates[candidate] += Weight
		} else {
			fmt.Println("candidate address invalid")
		}
	}
}

//@YiRan
//Vote Rule:
//1.The voters did not vote
//2.The target of the vote is the candidate
func (box *AddressVoteBox) AddToBoxIfNotVoted(Weight uint64, voter common.Address, to []common.Address) {
	if box.IsVoted(voter) {
		return
	}
	box.AddToBox(Weight, to)
	box.voters = append(box.voters, voter)
}

//YiRan
type AddressVoteBoxSorter []AddressVote

func NewAddressVoteBoxSorter(m map[common.Address]uint64) AddressVoteBoxSorter {
	s := make(AddressVoteBoxSorter, 0, len(m))
	for k, v := range m {
		s = append(s, AddressVote{Address: k, Score: v})
	}
	return s
}

func (ms AddressVoteBoxSorter) Len() int {
	return len(ms)
}
func (ms AddressVoteBoxSorter) Less(i, j int) bool {
	return ms[i].Score > ms[j].Score //Descending order
}
func (ms AddressVoteBoxSorter) Swap(i, j int) {
	ms[i], ms[j] = ms[j], ms[i]
}
