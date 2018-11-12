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

type Token struct {
	InnerID     uint64
	GlobalID    modules.IDType16
	Holder      common.Address
	Creator     common.Address
	CreatedTime []byte
	Extra       []byte
}

type CustomToken struct {
	customTokenName   string
	customTokenSymbol string
	totalSupply       uint64
	owner             common.Address
	currentIdx        uint64
	Members           map[common.Address][]uint64
	inventory         map[uint64]Token

	//optional field
	transFee uint
	//expiryTimeStamp uint64
	Decimals uint
}

//IsOwner(const):Anyone
func (ct *CustomToken) IsOwner(address common.Address) bool {
	return ct.owner == address
}

//ChangeTotalSupply(danger):Owner
func (ct *CustomToken) ChangeTotalSupply(address common.Address, amount uint64) bool {
	if !ct.IsOwner(address) {
		return false
	}
	if amount < ct.currentIdx {
		return false
	}
	ct.totalSupply = amount
	return true
}

//ChangeOwner(danger):Owner
func (ct *CustomToken) ChangeOwner(address common.Address, des common.Address) bool {
	if !ct.IsOwner(address) && des != address {
		return false
	}
	ct.owner = des
	return true
}

//OwnerOf(const):AnyOne
func (ct *CustomToken) OwnerOf(address common.Address, TokenID uint64) common.Address {
	return ct.inventory[TokenID].Holder
}

//TotalSupply(const):AnyOne
func (ct *CustomToken) TotalSupply(address common.Address) uint64 {
	return ct.totalSupply
}

//Name(const):AnyOne
func (ct *CustomToken) Name(address common.Address) string {
	return ct.customTokenName
}

//Symbol(const):AnyOne
func (ct *CustomToken) Symbol(address common.Address) string {
	return ct.customTokenSymbol
}

//balanceOf(const):AnyOne
func (ct *CustomToken) balanceOf(account common.Address) []uint64 {
	return ct.Members[account]
}

//GetUniversalToken(const):AnyOne
func (ct *CustomToken) GlobalIDByInnerID(account common.Address, id uint64) modules.IDType16 {
	return ct.inventory[id].GlobalID
}

//GetUniversalToken(const):AnyOne
func (ct *CustomToken) GetGlobalID(account common.Address, innerID uint64) modules.IDType16 {
	var buffer bytes.Buffer
	buffer.WriteString(ct.customTokenSymbol)
	buffer.WriteString(ct.customTokenName)
	buffer.Write(ct.owner.Bytes())
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, innerID)
	buffer.Write(b)
	id, _ := modules.SetIdTypeByHex(buffer.String())
	return id
}

//CreateNewToken(danger):Owner
func (ct *CustomToken) CreateNewToken(account common.Address, additional []byte) bool {
	if !ct.IsOwner(account) {
		return false
	}
	if ct.currentIdx == ct.totalSupply {
		return false
	}
	ct.Members[ct.owner] = append(ct.Members[ct.owner], ct.currentIdx)
	ct.inventory[ct.currentIdx] = Token{
		InnerID:  ct.currentIdx,
		GlobalID: ct.GetGlobalID(common.Address{}, ct.currentIdx),
		Holder:   ct.owner,
		Creator:  ct.owner,
	}
	ct.currentIdx++
	return true
}

//transferFrom(danger):TokenHolder
func (ct *CustomToken) transferFrom(from common.Address, to common.Address, ids []uint64) bool {
	//quit if any holder of token from ids passing in is invalid.
	for _, id := range ids {
		if ct.inventory[id].Holder != from {
			return false
		}
	}
	// change holder
	for _, id := range ids {
		token := ct.inventory[id]
		token.Holder = to
		ct.inventory[id] = token
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
		owner:             ownerAddress,
		customTokenName:   Name,
		customTokenSymbol: Symbol,
	}, nil
}
