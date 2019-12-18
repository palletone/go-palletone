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
package btcadaptor

import (
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcutil"

	"github.com/palletone/btc-adaptor/txscript"

	"github.com/palletone/adaptor"
)

func NewPrivateKey(netID int) ([]byte, error) {
	key, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		return nil, err
	}
	//fmt.Printf("%x\n", key.PubKey().SerializeCompressed())

	return key.Serialize(), nil
}

func GetPublicKey(priKey []byte, netID int) ([]byte, error) {
	_, pubKey := btcec.PrivKeyFromBytes(btcec.S256(), priKey)
	return pubKey.SerializeCompressed(), nil
}

func PubKeyToAddress(pubKey []byte, netID int) (string, error) {
	//chainnet
	realNet := GetNet(netID)
	addressPubKey, err := btcutil.NewAddressPubKey(pubKey, realNet)
	if err != nil {
		return "", err
	}
	return addressPubKey.EncodeAddress(), nil
}

func CreateMultiSigAddress(input *adaptor.CreateMultiSigAddressInput, netID int) (*adaptor.CreateMultiSigAddressOutput, error) {
	//0 < m < n and publicKeys == n
	if 0 >= input.SignCount {
		return nil, fmt.Errorf("Params error :SignCount need be bigger than 0.")
	}

	//chainnet
	realNet := GetNet(netID)

	//convert PublicKeyString to AddressPubKey
	pubkeys := make([]*btcutil.AddressPubKey, 0, len(input.Keys))
	for _, pubKeyBytes := range input.Keys {
		if len(pubKeyBytes) == 0 {
			continue
		}

		addressPubKey, err := btcutil.NewAddressPubKey(pubKeyBytes, realNet)
		if err != nil {
			return nil, err
		}
		pubkeys = append(pubkeys, addressPubKey)
	}
	if len(pubkeys) < input.SignCount {
		return nil, fmt.Errorf("Params error : PublicKeys small than SignCount.")
	}
	//create multisig address
	pkScript, err := txscript.MultiSigScript(pubkeys, input.SignCount)
	if err != nil {
		return nil, err
	}

	//multisig address scriptHash
	scriptAddr, err := btcutil.NewAddressScriptHash(pkScript, realNet)
	if err != nil {
		return nil, err
	}
	//result for return
	var output adaptor.CreateMultiSigAddressOutput
	output.Address = scriptAddr.String()
	output.Extra = pkScript

	return &output, nil
}
