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

import "github.com/palletone/go-palletone/common"

type AddressMultipleVote struct {
	MultipleVote
}


func (amv *AddressMultipleVote) Result(number uint8) []common.Address {
	sorted := amv.GetResult(number)
	return LInterface2LAddress(sorted)
}
func (amv *AddressMultipleVote) Register(addresses []common.Address) {
	listInterface := LAddress2LInterface(addresses)
	amv.RegisterCandidates(listInterface)
}

func (amv *AddressMultipleVote) Add(addresses []common.Address, score uint64) {
	lInterface := LAddress2LInterface(addresses)
	amv.AddNToBox(score, lInterface)
}
