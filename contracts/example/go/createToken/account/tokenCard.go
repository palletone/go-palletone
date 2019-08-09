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
	"github.com/palletone/go-palletone/contracts/example/go/createToken"
)

//type pFunc func(address common.Address, params []string) (string, bool)

//TokenCard : TokenCard class interface
type TokenCard interface {
	GetName() string
	GetSymbol() string
}

//TokenCardClass : TokenCard class implement
type TokenCardClass struct {
	name   string
	symbol string
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

//GenerateTokenCard : generate token-card from symbol
func GenerateTokenCard(symbol string) TokenCard {
	tc := &TokenCardClass{symbol: symbol}
	ct := new(createToken.CustomToken)
	//TODO:从数据库读取 CustomToken
	tc.ctData = ct
	return tc
}
