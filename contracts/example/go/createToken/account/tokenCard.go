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

import "github.com/palletone/go-palletone/common"

type pFunc func(address common.Address, params[]string) (string,bool)

//TokenCard : TokenCard class interface
type TokenCard interface {
	GetName() string
	GetSymbol() string
	GetFunc(string) pFunc
}

//TokenCardClass : TokenCard class implement
type TokenCardClass struct {
	name   string
	symbol string
	method map[string]pFunc
}

//GetName :
func (u *TokenCardClass) GetName() string {
	return u.name
}

//GetSymbol :
func (u *TokenCardClass) GetSymbol() string {
	return u.symbol
}
//Call :
func (u *TokenCardClass) GetFunc(m string) pFunc {
	f := u.method[m]
	return f
}

//GenerateTokenCard : generate token-card from symbol
func GenerateTokenCard(symbol string) TokenCard {
	tc := &TokenCardClass{symbol: symbol}
	//设置方法集
	return tc
}


// type IAccount interface {
// 	IsOwner() bool
// 	ChangeTotalSupply(amount uint64) bool
// 	ChangeOwner(des common.Address) bool
// 	OwnerOf(address common.Address, TokenID uint64) common.Address
// 	TotalSupply() uint64
// 	Name() string
// 	Symbol() string
// 	balanceOf(account common.Address) []uint64
// 	GlobalIDByInnerID(id uint64) modules.IDType16
// 	CreateNewToken(additional []byte) bool
// 	transfer(to common.Address, ids []uint64) bool
// }