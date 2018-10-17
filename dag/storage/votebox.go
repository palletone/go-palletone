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

import "github.com/palletone/go-palletone/common"

//@Yiran
//var (
//	MEDIATORVOTE_PREFIX = []byte("01")
//	COMMONVOTE_PREFIX   = []byte("00")
//
//	MEDIATORTERMINTERVAL = 3000
//)
//func (propdb *PropertyDb) UpdateActiveMediators () error{
//	term := unit
//	activeMediators, err := propdb.GetActiveMediators(,MEDIATORTERMINTERVAL)
//	if err != nil {
//		return ErrorLogHandler(err,"GetActiveMediators")
//	}
//
//}
//func (propdb *PropertyDb) GetActiveMediators(term []byte) ([]common.Address, error) {
//	key := KeyConnector(MEDIATOR_CANDIDATE_PREFIX,term)
//	// 1. Load Addresses of MediatorCandidates
//	addresses := make([]common.Address, 0)
//	ErrorLogHandler(Retrieve(propdb.db, string(key), addresses),"RetrieveMediatorCandidatesAddress")
//	// 2. Load VoteNumber of each MediatorCandidates
//	for _, address := range(addresses) {
//		tempKey := KeyConnector(key,address[:])
//		Retrieve
//	}
//
//}

//@Yiran This function checks that a transaction contains a action which creates a vote.
//func IsVoteInitiationTx(transactionIndex []byte) error {
//	//TODO
//	return nil
//}

//@Yiran this function connect multiple []byte keys to single []byte.
func KeyConnector(keys ...[]byte) []byte {
	var res []byte
	for _, key := range keys {
		res = append(res, key...)
	}
	return res
}

//@Yiran print error if exist.
func ErrorLogHandler(err error, errType string) error {
	if err != nil {
		println(errType, "error", err.Error())
		return err
	}
	return nil
}

//@Yiran
type VoteBox struct {
	Candidate common.Address
	Voter      common.Address
}

//func (box *VoteBox) Sort() {
//	//TODO
//}
//func (box *VoteBox) AddToBoxIfNotVoted(voter common.Address, vote common.Address) {
//	//TODO
//	//for addr := range box.voter {
//	//	if addr == voter{
//	//		return
//	//	}
//	//}
//
//}
//
//func NewVoteBox() *VoteBox {
//	return &VoteBox{
//		Candidates: make([]Candidate, 0),
//		Voter:      make([]common.Address, 0),
//	}
//}
//
////@Yiran
//type Candidate struct {
//	Address    common.Address
//	VoteNumber uint64
//}
