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

type openVoteModel struct {
	BaseVoteModel
	processPlugin
	privilegedVotePlugin
	deligateVotePlugin
}

func NewOpenVoteModel() *openVoteModel {
	m := &openVoteModel{}
	m.BaseVoteModel.candidatesStatus = make(map[interface{}]uint64, 0)
	m.processPlugin.processMap = make(map[interface{}][]interface{}, 0)
	m.privilegedVotePlugin.weightMap = make(map[interface{}]uint64, 0)
	m.deligateVotePlugin.agentMap = make(map[interface{}]interface{}, 0)
	return m
}
func (dpv *openVoteModel) RegisterCandidates(candidates interface{}) {
	dpv.BaseVoteModel.RegisterCandidates(candidates)
}

//Exist : wheither exist the given candidate in vote box.
func (dpv *openVoteModel) Exist(candidate interface{}) bool {
	return dpv.BaseVoteModel.Exist(candidate)
}

func (dpv *openVoteModel) GetCandidates() []interface{} {
	return dpv.BaseVoteModel.GetCandidates()
}

func (dpv *openVoteModel) SetCurrentVoter(voter interface{}) {
	dpv.processPlugin.SetCurrentVoter(voter)
}

func (dpv *openVoteModel) AddToBox(score uint64, tos interface{}) {
	dpv.privilegedVotePlugin.SetWeight(dpv.processPlugin.currentVoter, score)
	dpv.processPlugin.SetProcess(tos)
	dpv.DeleteAgent()
}

//GetScore :get data counted last calling of CountVote()
func (dpv *openVoteModel) GetScore(candidate interface{}) (uint64, error) {
	return dpv.BaseVoteModel.GetScore(candidate)
}

func (dpv *openVoteModel) GetVoteDetail() map[interface{}]uint64 {
	return dpv.BaseVoteModel.candidatesStatus
}

func (dpv *openVoteModel) SetAgent(agent interface{}) {
	dpv.deligateVotePlugin.SetAgent(dpv.processPlugin.currentVoter, agent)
	dpv.SetProcess(nil)
}

func (dpv *openVoteModel) DeleteAgent() {
	dpv.deligateVotePlugin.SetAgent(dpv.processPlugin.currentVoter, nil)
}

func (dpv *openVoteModel) CountVote() error {
	for from, to := range dpv.deligateVotePlugin.agentMap {
		dpv.SetWeight(to, dpv.GetWeight(to)+dpv.GetWeight(from))
	}
	for from, tos := range dpv.processPlugin.processMap {
		dpv.BaseVoteModel.AddToBox(dpv.privilegedVotePlugin.GetWeight(from), tos)
	}
	return nil
}

func (dpv *openVoteModel) GetResult(number uint8, val interface{}) bool {
	if err := dpv.CountVote(); err != nil {
		return false
	}
	return dpv.BaseVoteModel.GetResult(number, val)
}
