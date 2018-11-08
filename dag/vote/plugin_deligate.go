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

type deligateVotePlugin struct {
	agentMap map[interface{}]interface{}
}

func (dp *deligateVotePlugin) SetAgent(voter interface{}, agent interface{}) bool {
	//check loop reference
	//check agent and voter is valid
	nextAgent := dp.agentMap[agent]
	currAgent := agent
	var nilInterface interface{}
	for nextAgent != nilInterface {
		if nextAgent == voter {
			return false
		}
		currAgent = nextAgent
		nextAgent = dp.agentMap[nextAgent]
	}
	dp.agentMap[voter] = currAgent
	return true

}
