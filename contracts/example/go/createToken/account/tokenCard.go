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
	methodMap map[string]pFunc
	ctData interface{}
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
	f := u.methodMap[m]
	return f
}

//GenerateTokenCard : generate token-card from symbol
func GenerateTokenCard(symbol string) TokenCard {
	tc := &TokenCardClass{symbol: symbol}
	//1. 从数据库读取 CustomToken
	//2. 将CustomToken 赋值给 tc.ctData
	//3. 将tc.ctData的所有方法 赋值给 tc.methodMap
	return tc
}