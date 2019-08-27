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
 * @author PalletOne core developers <dev@pallet.one>
 * @date 2018
 */
package ethadaptor

import (
	base58 "github.com/btcsuite/btcutil/base58"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/palletone/adaptor"
)

func NewPrivateKey(netID int) ([]byte, error) {
	privateKeyECDSA, err := crypto.GenerateKey()
	if err != nil {
		return nil, err
	}

	return crypto.FromECDSA(privateKeyECDSA), nil
}

func GetPublicKey(priKey []byte) ([]byte, error) {
	privateKeyECDSA, err := crypto.ToECDSA(priKey)
	if err != nil {
		return nil, err
	}

	return crypto.CompressPubkey(&privateKeyECDSA.PublicKey), nil
}

func GetAddress(priKey []byte, netID int) (string, error) {
	privateKeyECDSA, err := crypto.ToECDSA(priKey)
	if err != nil {
		return "", err
	}
	addr := crypto.PubkeyToAddress(privateKeyECDSA.PublicKey)
	//	fmt.Println(addr.String())

	return addr.String(), nil
}
func PubKeyToAddress(pubKey []byte) (string, error) {
	pk, err := crypto.DecompressPubkey(pubKey)
	if err != nil {
		return "", err
	}
	addr := crypto.PubkeyToAddress(*pk)
	return addr.String(), nil
}
func GetPalletOneMappingAddress(addr *adaptor.GetPalletOneMappingAddressInput) (
	*adaptor.GetPalletOneMappingAddressOutput, error) {
	var addrBytes []byte
	if "0x" == addr.ChainAddress[:2] || "0X" == addr.ChainAddress[:2] {
		addrBytes = Hex2Bytes(addr.ChainAddress[2:])
	} else {
		addrBytes = Hex2Bytes(addr.ChainAddress)
	}
	var result adaptor.GetPalletOneMappingAddressOutput
	result.PalletOneAddress = "P" + base58.CheckEncode(addrBytes, 0)
	return &result, nil
}
