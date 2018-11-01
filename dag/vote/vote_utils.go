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

func MAddress2MInterface(addresses map[common.Address]bool) map[interface{}]bool {
	res := make(map[interface{}]bool, 0)
	for addr, _ := range addresses {
		res[addr] = true
	}
	return res
}


func MAddress2LInterface(addresses map[common.Address]bool) []interface{} {
	res := make([]interface{}, 0)
	for addr, _ := range addresses {
		res = append(res, addr)
	}
	return res
}

func LAddress2LInterface(addresses []common.Address) []interface{} {
	res := make([]interface{}, 0)
	for _, addr := range addresses {
		res = append(res, addr)
	}
	return res
}

func LInterface2LAddress(li []interface{}) []common.Address {
	res := make([]common.Address, 0)
	for _, iaddr := range li {
		res = append(res, iaddr.(common.Address))
	}
	return res
}
