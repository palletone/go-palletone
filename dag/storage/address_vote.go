package storage

import (
	"fmt"
	"github.com/palletone/go-palletone/common"
	"sort"
)

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
		if AddressEqual(voted, voter) {
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
func NewAddressVoteBox() *AddressVoteBox {
	return &AddressVoteBox{
		Candidates: make(map[common.Address]uint64, 0),
		voters:     make([]common.Address, 0),
	}
}

//YiRan
type AddressVoteBox struct {
	Candidates map[common.Address]uint64
	voters     []common.Address
}

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
