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

import "fmt"

type privilegedVotePlugin struct {
	weightMap map[interface{}]uint64
}

func (pp *privilegedVotePlugin) GetWeight(voter interface{}) uint64 {
	if w, ok := pp.weightMap[voter]; ok {
		return w
	}
	fmt.Println("voter's weight is not initialized")
	return 0
}

func (pp *privilegedVotePlugin) SetWeight(voter interface{}, weight uint64) {
	//check voter validity first
	pp.weightMap[voter] = weight
}

func (pp *privilegedVotePlugin) SetWeightBatch(voters interface{}, weight uint64) {
	//check voter validity first
	is := ToInterfaceSlice(voters)
	for _, voter := range is {
		pp.weightMap[voter] = weight
	}
}
