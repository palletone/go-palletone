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
 * @author PalletOne core developer <dev@pallet.one>
 * @date 2018
 */
package account

import (
	"fmt"
	"github.com/palletone/go-palletone/common"
	"testing"
)

func TestAccountModel(t *testing.T) {
	addr := common.StringToAddressGodBlessMe("P1GqZ72gaeq7LiS34KLJoMmCnMnaopkcEPn")
	token := uint64(1)
	//1. generate user model
	u := GenerateUser(addr)

	//2. add different type of tokens to wallet
	u.AddNewTokenCard("wil")
	u.AddNewTokenCard("tom")

	//3. set current token for call
	u.SetCurrentTokenCard("wil")

	//4. system call
	result, ok := u.Call("Tranfer", addr, []uint64{token}) //strict match
	if !ok {
		//check
		fmt.Println(ok)
	}
	fmt.Println(result)

}
