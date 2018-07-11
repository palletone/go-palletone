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
	"reflect"
	// "strings"

	"math/big"

	"github.com/btcsuite/btcutil/base58"
	"github.com/palletone/go-palletone/common/hexutil"
)

const (
	AddressLength = 35 // ETH length is 20
)

var (
	addressT = reflect.TypeOf(Address{})
)

/////////// Address

// Address represents the 35 byte address of an PalletOne account.
// for personal address, start with P1, script address start with P3, contract address start with Pc
type Address [AddressLength]byte
type AddressType byte

const (
	ErrorAddress  AddressType = iota
	PublicKeyHash AddressType = 1
	ScriptHash    AddressType = 2
	ContractHash  AddressType = 3
)

func (a Address) Validate() (AddressType, error) {
	if a[0] != byte('P') {
		return ErrorAddress, errors.New("PalletOne address must start with 'P'")
	}
	_, version, err := base58.CheckDecode(string(a[1:]))
	if err != nil {
		return ErrorAddress, err
	}
	switch version {
	case 0:
		return PublicKeyHash, nil
	case 5:
		return ScriptHash, nil
	case 28:
		return ContractHash, nil
	default:
		return ErrorAddress, errors.New("Invalid address type")
	}

}
func IsValidAddress(s string) bool {
	addr := StringToAddress(s)
	_, err := addr.Validate()
	return err == nil
}
func BytesToAddress(b []byte) Address {
	var a Address
	a.SetBytes(b)
	return a
}
func StringToAddress(s string) Address { return BytesToAddress([]byte(s)) }
func HexToAddress(s string) Address    { return BytesToAddress(FromHex(s)) }
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
func (a Address) Str() string { return string(a[:]) }
func (a Address) Bytes() []byte {

	result, _, _ := base58.CheckDecode(a.String()[1:])
	return result
}

func (a Address) Big() *big.Int { return new(big.Int).SetBytes(a.Bytes()) }
func (a Address) Hash() Hash    { return BytesToHash(a.Bytes()) }
func (a Address) Hex() string   { return fmt.Sprintf("0x%x", a.Bytes()) }

// Hex returns an EIP55-compliant hex string representation of the address.
// func (a Address) Hex() string {
// 	unchecksummed := hex.EncodeToString(a[:])
// 	sha := sha3.NewKeccak256()
// 	sha.Write([]byte(unchecksummed))
// 	hash := sha.Sum(nil)

// 	result := []byte(unchecksummed)
// 	for i := 0; i < len(result); i++ {
// 		hashByte := hash[i/2]
// 		if i%2 == 0 {
// 			hashByte = hashByte >> 4
// 		} else {
// 			hashByte &= 0xf
// 		}
// 		if result[i] > '9' && hashByte > 7 {
// 			result[i] -= 32
// 		}
// 	}
// 	return "0x" + string(result)
// }

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
func (a *Address) SetString(s string) { a.SetBytes([]byte(s)) }

// Sets a to other
func (a *Address) Set(other Address) {
	for i, v := range other {
		a[i] = v
	}
}

// MarshalText returns the hex representation of a.
func (a Address) MarshalText() ([]byte, error) {
	return hexutil.Bytes(a[:]).MarshalText()
}

// UnmarshalText parses a hash in hex syntax.
func (a *Address) UnmarshalText(input []byte) error {
	return hexutil.UnmarshalFixedText("Address", input, a[:])
}

// UnmarshalJSON parses a hash in hex syntax.
func (a *Address) UnmarshalJSON(input []byte) error {
	return hexutil.UnmarshalFixedJSON(addressT, input, a[:])
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
