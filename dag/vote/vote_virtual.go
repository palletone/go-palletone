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
	RegisterCandidates(candidates interface{})
	Exist(candidate interface{}) bool
	GetCandidates() []interface{}
	AddToBox(score uint64, tos interface{})
	GetScore(candidate interface{}) (uint64, error)
	GetVoteDetail() map[interface{}]uint64
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
	RegisterCandidates(candidates interface{})
	SetCurrentVoter(voter interface{})
	AddToBox(score uint64, tos interface{})
	DeleteAgent()
	SetAgent(voter interface{})
}

func NewBaseVote () vote {
	return NewBaseVoteModel()
}

func NewOpenVote () openVote {
	return NewOpenVoteModel()
}
