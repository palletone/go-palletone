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

package createToken

import (
	"bytes"
	"encoding/binary"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/dag/errors"
	"github.com/palletone/go-palletone/dag/modules"
)

type ICustomToken interface {
}

type Token struct {
	InnerID     uint64
	GlobalID    modules.AssetId
	Holder      common.Address
	Creator     common.Address
	CreatedTime []byte
	Extra       []byte
}

//注意大小写
type CustomToken struct {
	CustomTokenName   string
	CustomTokenSymbol string
	TotalSupply       uint64
	Owner             common.Address
	CurrentIdx        uint64
	Members           map[common.Address][]uint64
	Inventory         map[uint64]Token

	//optional field
	TransFee uint
	//expiryTimeStamp uint64
	Decimals uint
}

//IsOwner(const):Anyone
func (ct *CustomToken) IsOwner(caller common.Address) bool {
	return ct.Owner == caller
}

//ChangeTotalSupply(danger):Owner
func (ct *CustomToken) ChangeTotalSupply(caller common.Address, amount uint64) bool {
	if !ct.IsOwner(caller) {
		return false
	}
	if amount < ct.CurrentIdx {
		return false
	}
	ct.TotalSupply = amount
	return true
}

//GetDecimals(const):AnyOne
func (ct *CustomToken) GetDecimals(caller common.Address) uint {
	return ct.Decimals
}

//ChangeOwner(danger):Owner
func (ct *CustomToken) ChangeOwner(caller common.Address, des common.Address) bool {
	if !ct.IsOwner(caller) && des != caller {
		return false
	}
	ct.Owner = des
	return true
}

//OwnerOf(const):AnyOne
func (ct *CustomToken) OwnerOf(caller common.Address, TokenID uint64) common.Address {
	return ct.Inventory[TokenID].Holder
}

//TotalSupply(const):AnyOne
func (ct *CustomToken) GetTotalSupply(caller common.Address) uint64 {
	return ct.TotalSupply
}

//Name(const):AnyOne
func (ct *CustomToken) GetName(caller common.Address) string {
	return ct.CustomTokenName
}

//Symbol(const):AnyOne
func (ct *CustomToken) GetSymbol(caller common.Address) string {
	return ct.CustomTokenSymbol
}

//balanceOf(const):AnyOne
func (ct *CustomToken) BalanceOf(caller common.Address) []uint64 {
	return ct.Members[caller]
}

//GetUniversalToken(const):AnyOne
func (ct *CustomToken) GlobalIDByInnerID(caller common.Address, id uint64) modules.AssetId {
	return ct.Inventory[id].GlobalID
}

//GetUniversalToken(const):AnyOne
func (ct *CustomToken) GetGlobalID(caller common.Address, innerID uint64) modules.AssetId {
	var buffer bytes.Buffer
	buffer.WriteString(ct.CustomTokenSymbol)
	buffer.WriteString(ct.CustomTokenName)
	buffer.Write(ct.Owner.Bytes())
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, innerID)
	buffer.Write(b)
	id, _ := modules.SetIdTypeByHex(buffer.String())
	return id
}

//CreateNewToken(danger):Owner
func (ct *CustomToken) CreateNewToken(caller common.Address, additional []byte) bool {
	if !ct.IsOwner(caller) {
		return false
	}
	if ct.CurrentIdx == ct.TotalSupply {
		return false
	}
	ct.Members[ct.Owner] = append(ct.Members[ct.Owner], ct.CurrentIdx)
	ct.Inventory[ct.CurrentIdx] = Token{
		InnerID:  ct.CurrentIdx,
		GlobalID: ct.GetGlobalID(common.Address{}, ct.CurrentIdx),
		Holder:   ct.Owner,
		Creator:  ct.Owner,
	}
	ct.CurrentIdx++
	return true
}

//transfer(danger):TokenHolder
func (ct *CustomToken) Transfer(caller common.Address, to common.Address, ids []uint64) bool {
	//quit if any holder of token from ids passing in is invalid.
	for _, id := range ids {
		if ct.Inventory[id].Holder != caller {
			return false
		}
	}
	// change holder
	for _, id := range ids {
		token := ct.Inventory[id]
		token.Holder = to
		ct.Inventory[id] = token
	}
	return true
}

//move to storage modules
type CustomTokenConstructor struct {
}

func (ctor *CustomTokenConstructor) ValidCustomTokenSymbolUniqueness(tokenSymbol string) bool {
	return true
}

func (ctor *CustomTokenConstructor) CreateCustomToken(Name string, Symbol string, ownerAddress common.Address) (*CustomToken, error) {
	if !ctor.ValidCustomTokenSymbolUniqueness(Symbol) {
		return nil, errors.New("InvalidCustomTokenSymbol")
	}
	return &CustomToken{
		Owner:             ownerAddress,
		CustomTokenName:   Name,
		CustomTokenSymbol: Symbol,
	}, nil
}
