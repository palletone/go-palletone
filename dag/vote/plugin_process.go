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

type processPlugin struct {
	currentVoter interface{}
	processMap   map[interface{}][]interface{}
}

func (pp *processPlugin) SetCurrentVoter(voter interface{}) {
	pp.currentVoter = voter
}

func (pp *processPlugin) SetProcess(tosi interface{}) {
	var NilInterfaceSlice interface{}
	// if tosi is nil ,delete process.
	if tosi == NilInterfaceSlice{
		delete(pp.processMap, pp.currentVoter)
		return
	}
	tos := ToInterfaceSlice(tosi)
	pp.processMap[pp.currentVoter] = tos
}
