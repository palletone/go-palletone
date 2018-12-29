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
func (ovm *openVoteModel) RegisterCandidates(candidates interface{}) {
	ovm.BaseVoteModel.RegisterCandidates(candidates)
}

func (ovm *openVoteModel) GetCandidates() []interface{} {
	return ovm.BaseVoteModel.GetCandidates()
}

func (ovm *openVoteModel) SetCurrentVoter(voter interface{}) {
	ovm.processPlugin.SetCurrentVoter(voter)
}

func (ovm *openVoteModel) AddToBox(tos interface{}) {
	ovm.processPlugin.SetProcess(tos)
	ovm.deligateVotePlugin.SetAgent(ovm.currentVoter,nil)

}

//GetScore :get data counted last calling of CountVote()
func (ovm *openVoteModel) GetCandidateScore(candidate interface{}) (uint64, error) {
	return ovm.BaseVoteModel.GetCandidateScore(candidate)
}

func (ovm *openVoteModel) GetVoteDetail() map[interface{}]uint64 {
	return ovm.BaseVoteModel.candidatesStatus
}

func (ovm *openVoteModel) SetAgent(agent interface{}) {
	ovm.deligateVotePlugin.SetAgent(ovm.processPlugin.currentVoter, agent)
	ovm.processPlugin.SetProcess(nil)
}

func (ovm *openVoteModel) SetWeight(weight uint64) {
	ovm.privilegedVotePlugin.SetWeight(ovm.currentVoter, weight)
}

//func (ovm *openVoteModel) DeleteAgent() {
//	ovm.deligateVotePlugin.SetAgent(ovm.processPlugin.currentVoter, nil)
//}

func (ovm *openVoteModel) CountVote() {
	//backup weight map
	BackUpWeightMap := make(map[interface{}]uint64, 0)
	for k,v := range ovm.privilegedVotePlugin.weightMap{
		BackUpWeightMap[k]=v
	}

	for from, to := range ovm.deligateVotePlugin.agentMap {
		ovm.privilegedVotePlugin.SetWeight(to, ovm.GetWeight(to)+ovm.GetWeight(from))
	}
	for from, tos := range ovm.processPlugin.processMap {
		ovm.BaseVoteModel.AddToBox(ovm.privilegedVotePlugin.GetWeight(from), tos)
	}
	//recover weight map
	ovm.privilegedVotePlugin.weightMap = BackUpWeightMap
}

func (ovm *openVoteModel) GetResult(number uint8, val interface{}) bool {
	ovm.CountVote()
	return ovm.BaseVoteModel.GetResult(number, val)
}


