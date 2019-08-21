// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package common

import (
	// "encoding/hex"
	//"encoding/json"
	"errors"
	"fmt"
	// "math/big"
	// "math/rand"
	//"reflect"
	// "strings"

	"math/big"

	"bytes"
	"encoding/json"
	"github.com/btcsuite/btcutil/base58"
)

const (
	AddressLength = 21 //byte[0:20] is hash160, byte[20] is AddressType
)

//var (
//	addressT = reflect.TypeOf(Address{})
//)

/////////// Address

// Address represents the 35 byte address of an PalletOne account.
// for personal address, start with P1 (version 0), script address start with P3(version 5),
// contract address start with Pc(version 28)
type Addresses []Address
type Address [AddressLength]byte
type AddressType byte

const (
	ErrorAddress  AddressType = 0xff
	PublicKeyHash AddressType = 0
	ScriptHash    AddressType = 5
	ContractHash  AddressType = 28
)

func (a *Address) GetType() AddressType {
	return AddressType(a[20])
}

//如果是合约地址，那么是不是一个系统合约地址？
func (a *Address) IsSystemContractAddress() bool {
	return IsSystemContractAddress(a.Bytes())
}
func IsSystemContractAddress(addr []byte) bool {
	bb := make([]byte, 20)
	bb[18] = 0xff
	bb[19] = 0xff
	return bytes.Compare(addr, bb) < 0
}
func NewAddress(hash160 []byte, ty AddressType) Address {
	newBytes := make([]byte, 21)
	copy(newBytes, hash160)
	newBytes[20] = byte(ty)
	return BytesToAddress(newBytes)
}

//将一个字符串格式的Address转换为Address对象
func StringToAddress(a string) (Address, error) {
	if len(a) <= 0 {
		return Address{}, errors.New("Address cannot be null")
	}
	if a[0] != byte('P') {
		return Address{}, errors.New("PalletOne address must start with 'P'")
	}
	addrb, version, err := base58.CheckDecode(a[1:])
	if err != nil {
		return Address{}, err
	}
	switch version {
	case 0:
		return BytesToAddress(append(addrb, byte(PublicKeyHash))), nil
	case 5:
		return BytesToAddress(append(addrb, byte(ScriptHash))), nil
	case 28:
		return BytesToAddress(append(addrb, byte(ContractHash))), nil
	default:
		return Address{}, errors.New("Invalid address type")
	}
}

func StringToAddressGodBlessMe(a string) Address {
	addr, _ := StringToAddress(a)
	return addr
}

func (a Address) Validate() (AddressType, error) {
	var ty AddressType = AddressType(a[20])
	return ty, nil
}

func IsValidAddress(s string) bool {
	_, err := StringToAddress(s)
	// if err!=nil{
	// 	return ErrorAddress,err
	// }
	// return a.GetType(),nil
	return err == nil
}
func BytesToAddress(b []byte) Address {
	var a Address
	a.SetBytes(b)
	return a
}

func BytesListToAddressList(b []byte) []Address {

	var stringArray []string
	json.Unmarshal(b, &stringArray)
	Addresses := make([]Address, 0, len(stringArray))

	for _, str := range stringArray {
		addr, _ := StringToAddress(str)
		Addresses = append(Addresses, addr)
	}
	return Addresses
}

func (a Addresses) Len() int {
	return len(a)
}

func (a Addresses) Less(i, j int) bool {
	return a[i].Less(a[j])
}

func (a Addresses) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func HexToAddress(s string) Address {
	return BytesToAddress(FromHex(s))
}

func PubKeyHashHexToAddress(s string) Address {
	pubKeyHash := FromHex(s)
	addrStr := "P" + base58.CheckEncode(pubKeyHash, byte(0))
	return BytesToAddress([]byte(addrStr))
}

// IsHexAddress verifies whether a string can represent a valid hex-encoded
// PalletOne address or not.
func IsHexAddress(s string) bool {
	if hasHexPrefix(s) {
		s = s[2:]
	}
	return len(s) == 2*AddressLength && isHex(s)
}

// Get the string representation of the underlying address
func (a Address) Str() string {
	return "P" + base58.CheckEncode(a[0:20], a[20])
}

//Return account 20 bytes without address type
func (a Address) Bytes() []byte {
	return a[0:20]
}

//Return address all 21 bytes, you can use SetBytes function to get Address object
func (a Address) Bytes21() []byte {
	return a[:]
}

func (a Address) Big() *big.Int { return new(big.Int).SetBytes(a.Bytes()) }
func (a Address) Hash() Hash    { return BytesToHash(a.Bytes()) }
func (a Address) Hex() string   { return fmt.Sprintf("%#x", a.Bytes()) }

// String implements the stringer interface and is used also by the logger.
func (a Address) String() string {
	return a.Str()
}

// Format implements fmt.Formatter, forcing the byte slice to be formatted as is,
// without going through the stringer interface used for logging.
func (a Address) Format(s fmt.State, c rune) {
	fmt.Fprintf(s, "%"+string(c), a[:])
}

// Sets the address to the value of b. If b is larger than len(a) it will panic
func (a *Address) SetBytes(b []byte) {
	if len(b) > len(a) {
		b = b[len(b)-AddressLength:]
	}
	copy(a[AddressLength-len(b):], b)
}

// Set string `s` to a. If s is larger than len(a) it will panic
func (a *Address) SetString(s string) error {
	b, err := StringToAddress(s)
	if err != nil {
		return err
	}
	a.Set(b)
	return nil
}

// Sets a to other
func (a *Address) Set(other Address) {
	for i, v := range other {
		a[i] = v
	}
}

// MarshalText returns the hex representation of a.
func (a Address) MarshalText() ([]byte, error) {
	return []byte(a.String()), nil
}

// UnmarshalText parses a hash in hex syntax.
func (a *Address) UnmarshalText(input []byte) error {
	return a.SetString(string(input))
}

// UnmarshalJSON parses a hash in hex syntax.
func (a *Address) UnmarshalJSON(input []byte) error {
	addrStr := string(input[1 : len(input)-1])
	return a.SetString(addrStr)
}

func (a *Address) MarshalJSON() ([]byte, error) {
	str := a.String()
	return json.Marshal(str)
}

//YiRan
//Returns true when the contents of the two Address are exactly the same
func (a *Address) Equal(b Address) bool {
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
func (a *Address) IsZero() bool {
	for _, v := range a {
		if v != byte(0) {
			return false
		}
	}
	return true
}

func (a *Address) Less(b Address) bool {
	for i, v := range a {
		if v < b[i] {
			return true
		} else if v > b[i] {
			return false
		}
	}

	// 两个地址相同
	return false
}

// UnprefixedHash allows marshaling an Address without 0x prefix.
// type UnprefixedAddress Address

// // UnmarshalText decodes the address from hex. The 0x prefix is optional.
// func (a *UnprefixedAddress) UnmarshalText(input []byte) error {
// 	return hexutil.UnmarshalFixedUnprefixedText("UnprefixedAddress", input, a[:])
// }

// // MarshalText encodes the address as hex.
// func (a UnprefixedAddress) MarshalText() ([]byte, error) {
// 	return []byte(hex.EncodeToString(a[:])), nil
// }

//MixedcaseAddress retains the original string, which may or may not be
//correctly checksummed.
/////////////////(for wallet core tyeps SendTxArgs)
type MixedcaseAddress struct {
	addr     Address
	original string
}

// Address returns the address
func (ma *MixedcaseAddress) Address() Address {
	return ma.addr
}

// String implements fmt.Stringer
func (ma *MixedcaseAddress) String() string {
	if ma.ValidChecksum() {
		return fmt.Sprintf("%s [chksum ok]", ma.original)
	}
	return fmt.Sprintf("%s [chksum INVALID]", ma.original)
}

// ValidChecksum returns true if the address has valid checksum
func (ma *MixedcaseAddress) ValidChecksum() bool {
	return ma.original == ma.addr.Hex()
}

// // NewMixedcaseAddress constructor (mainly for testing)
// func NewMixedcaseAddress(addr Address) MixedcaseAddress {
// 	return MixedcaseAddress{addr: addr, original: addr.Str()}
// }

// // NewMixedcaseAddressFromString is mainly meant for unit-testing
// func NewMixedcaseAddressFromString(hexaddr string) (*MixedcaseAddress, error) {
// 	if !IsHexAddress(hexaddr) {
// 		return nil, fmt.Errorf("Invalid address")
// 	}
// 	a := FromHex(hexaddr)
// 	return &MixedcaseAddress{addr: BytesToAddress(a), original: hexaddr}, nil
// }

// // UnmarshalJSON parses MixedcaseAddress
// func (ma *MixedcaseAddress) UnmarshalJSON(input []byte) error {
// 	if err := hexutil.UnmarshalFixedJSON(addressT, input, ma.addr[:]); err != nil {
// 		return err
// 	}
// 	return json.Unmarshal(input, &ma.original)
// }

// // MarshalJSON marshals the original value
// func (ma *MixedcaseAddress) MarshalJSON() ([]byte, error) {
// 	if strings.HasPrefix(ma.original, "0x") || strings.HasPrefix(ma.original, "0X") {
// 		return json.Marshal(fmt.Sprintf("0x%s", ma.original[2:]))
// 	}
// 	return json.Marshal(fmt.Sprintf("0x%s", ma.original))
// }

// Original returns the mixed-case input string
func (ma *MixedcaseAddress) Original() string {
	return ma.original
}

type SignatureError struct {
	InputIndex uint32
	MsgIndex   uint32
	Error      error
}
