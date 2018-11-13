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
 * @author PalletOne core developer  <dev@pallet.one>
 * @date 2018
 */

package account

import (
	"github.com/palletone/go-palletone/common"
	"reflect"
)

//User : user calss interface
type User interface {
	AddNewTokenCard(symbol string) bool
	SetCurrentTokenCard(symbol string) bool
	Call(m string, params ...interface{}) (string,bool)
}

//UserClass : user class implement
type UserClass struct {
	Address     common.Address
	Wallet      []TokenCard
	CurrentCard uint
}

//AddNewTokenCard : generate a token-card if not exist in wallet for given symbol then put into user's wallet
func (u *UserClass) AddNewTokenCard(symbol string) bool {
	for _, tokenCard := range u.Wallet {
		if tokenCard.GetSymbol() == symbol {
			return false
		}
	}
	tk := GenerateTokenCard(symbol)
	u.CurrentCard = uint(len(u.Wallet) + 1)
	u.Wallet = append(u.Wallet, tk)
	return true
}

//SetCurrentTokenCard :
func (u *UserClass) SetCurrentTokenCard(symbol string) bool {
	for i, tokenCard := range u.Wallet {
		if tokenCard.GetSymbol() == symbol {
			u.CurrentCard = uint(i)
			return true
		}
	}
	return false
}
//Call :
func (u *UserClass) Call(m string, params ...interface{}) (string,bool) {
	currTokenCard := u.Wallet[u.CurrentCard]
	vparams := make([]reflect.Value,0)
	for _,p := range params {
		vparams = append(vparams,reflect.ValueOf(p))
	}
	res := reflect.ValueOf(currTokenCard).MethodByName(m).Call(vparams)
	if len(res)!= 2 {
		return "return type of func invalid",false
	}
	return res[0].String(),res[1].Bool()
}

//GenerateUser : generate a user model for contract call.
func GenerateUser(address common.Address) User {
	u := &UserClass{Address:address}
	return u
}
