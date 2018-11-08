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
func NewOpenVoteModel () *openVoteModel {
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


/*
switch reflect.ValueOf(candidates).Kind() {
	case reflect.Slice:
		if bv.elemType == reflect.TypeOf(candidates).Elem() {
			for _, c := range ToInterfaceSlice(candidates) {
				if bv.Exist(c) {
					bv.candidatesStatus[c] += score
				}
			}
		}

	default:
		if bv.Exist(candidates) {
			bv.candidatesStatus[candidates] += score
		}
	}
*/
func (dpv *openVoteModel) AddToBox(score uint64, tos interface{}) {
	//check validity of candidates
	dpv.BaseVoteModel.AddToBox(0,tos)

	dpv.privilegedVotePlugin.SetWeight(dpv.processPlugin.currentVoter, score)
	dpv.processPlugin.SetProcess(ToInterfaceSlice(tos))
	delete(dpv.deligateVotePlugin.agentMap, dpv.processPlugin.currentVoter)
}

//GetScore : score data may out of date
func (dpv *openVoteModel) GetScore(candidate interface{}) (uint64, error) {
	return dpv.BaseVoteModel.GetScore(candidate)
}

func (dpv *openVoteModel) GetVoteDetail() map[interface{}]uint64 {
	return dpv.BaseVoteModel.candidatesStatus
}

func (dpv *openVoteModel) SetAgent(voter interface{}) {
	dpv.deligateVotePlugin.SetAgent(dpv.processPlugin.currentVoter, voter)
}

func (dpv *openVoteModel) DeleteAgent() {
	delete(dpv.deligateVotePlugin.agentMap, dpv.processPlugin.currentVoter)
}

func (dpv *openVoteModel) CountVote() error {
	return nil
}

func (dpv *openVoteModel) GetResult(number uint8, val interface{}) bool {
	if err := dpv.CountVote() ;err!= nil {
		return false
	}
	return dpv.BaseVoteModel.GetResult(number, val)
}
