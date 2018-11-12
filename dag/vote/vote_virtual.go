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

type vote interface {
	//Add the candidates into the vote box. using a slice container or just single element as parameter.
	//for every single entity vote struct, the type of candidate should be same.
	RegisterCandidates(candidates interface{})
	//Query whether a candidate exists
	Exist(candidate interface{}) bool
	//Gets a list of all registered candidates and returns []interface.
	GetCandidates() []interface{}
	//Depending on the implementation, the actual operation will be different.
	//base vote: score number will incrsing the all given candidates score directly.
	//tos accept a slice container or just single element as parameter.
	//open vote: score will be ignored, because power of voter is set by another method.
	//this function only implement the operation that set the vote apiration for current voter.
	AddToBox(score uint64, tos interface{})
	//return given candidate's score, return error if candidate not valid .
	GetScore(candidate interface{}) (uint64, error)
	//return current vote status for all candidates.
	GetVoteDetail() map[interface{}]uint64
	//count a slice of elected candiates and replace the val, so val should be a pointer by passing variable's mem address.
	//number describe how many candidate will return .
	// boolen value is false when candidates which passing by method called for RegisterCandidates() or  AddToBox() before is invalid
	GetResult(number uint8, val interface{}) bool
}

type privileger interface {
	GetWeight(voter interface{}) uint64
	SetWeight(voter interface{}, weight uint64)
	SetWeightBatch(voters interface{}, weight uint64)
}

type deligater interface {
	SetAgent(voter interface{}, agent interface{}) bool
}

type processor interface {
	SetCurrentVoter(voter interface{})
	SetProcess(tos []interface{})
}

type openVote interface {
	//Add the candidates into the vote box. using a slice container or just single element as parameter.
	//for every single entity vote struct, the type of candidate should be same.
	RegisterCandidates(candidates interface{})
	//change current voter.
	//Once you have a "voter" set, by default all operations are done with that "voter"
	SetCurrentVoter(voter interface{})
	//tos accept a slice container or just single element as parameter.
	//score is weight set to current voter.
	AddToBox(score uint64, tos interface{})

	// The current voter's vote will be represented by the  passing agent .
	SetAgent(voter interface{})
	// cancel agent
	DeleteAgent()
	// count the result
	GetResult(number uint8, val interface{}) bool
}

func NewBaseVote() vote {
	return NewBaseVoteModel()
}

func NewOpenVote() openVote {
	return NewOpenVoteModel()
}
