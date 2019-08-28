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
package adaptorbtc

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"

	"github.com/btcsuite/btcd/btcec"
	//"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcutil"

	"github.com/palletone/btc-adaptor/txscript"

	"github.com/palletone/adaptor"
)

func NewPrivateKey(netID int) (wifPriKey string) {
	//rand bytes
	randBytes := make([]byte, 32)
	_, err := rand.Read(randBytes)
	if err != nil {
		return err.Error()
	}

	//chainnet
	realNet := GetNet(netID)

	//wif wallet import format
	key, _ := btcec.PrivKeyFromBytes(btcec.S256(), randBytes)
	wif, err := btcutil.NewWIF(key, realNet, true)
	if err != nil {
		return err.Error()
	}
	return wif.String()
}

func GetPublicKey(wifPriKey string, netID int) (pubKey string) {
	//decode to wif
	wif, err := btcutil.DecodeWIF(wifPriKey)
	if err != nil {
		return err.Error()
	}

	//chainnet
	realNet := GetNet(netID)

	addressPubKey, err := btcutil.NewAddressPubKey(wif.SerializePubKey(),
		realNet)
	return addressPubKey.String()
}

func GetAddress(wifPriKey string, netID int) (address string) {
	//decode to wif
	wif, err := btcutil.DecodeWIF(wifPriKey)
	if err != nil {
		return err.Error()
	}

	//chainnet
	realNet := GetNet(netID)

	addressPubKey, err := btcutil.NewAddressPubKey(wif.SerializePubKey(),
		realNet)

	return addressPubKey.EncodeAddress()
}

func GetAddressByPubkey(pubKeyHex string, netID int) (string, error) {
	//
	pubKeyBytes, err := hex.DecodeString(pubKeyHex)
	if err != nil {
		return "", err
	}

	//chainnet
	realNet := GetNet(netID)

	addressPubKey, err := btcutil.NewAddressPubKey(pubKeyBytes,
		realNet)
	if err != nil {
		return "", err
	}
	return addressPubKey.EncodeAddress(), nil
}

func CreateMultiSigAddress(createMultiSigParams *adaptor.CreateMultiSigParams, netID int) (string, error) {
	//	var createMultiSigParams CreateMultiSigParams
	//	err := json.Unmarshal([]byte(params), &createMultiSigParams)
	//	if err != nil {
	//		log.Fatal(err)
	//		return err.Error()
	//	}

	//0 < m < n and publicKeys == n
	if 0 == createMultiSigParams.M ||
		createMultiSigParams.M > createMultiSigParams.N ||
		0 == createMultiSigParams.N {
		return "", errors.New("Params error : 0 < m < n.")
	}

	//chainnet
	realNet := GetNet(netID)

	//convert PublicKeyString to AddressPubKey
	pubkeys := make([]*btcutil.AddressPubKey, len(createMultiSigParams.PublicKeys))
	for i, publicKeyString := range createMultiSigParams.PublicKeys {
		publicKeyString = strings.TrimSpace(publicKeyString) //Trim whitespace
		if len(publicKeyString) == 0 {
			continue
		}
		pubKeyBytes, err := hex.DecodeString(publicKeyString)
		if err != nil {
			return "", err
		}

		addressPubKey, err := btcutil.NewAddressPubKey(pubKeyBytes, realNet)
		if err != nil {
			return "", err
		}
		pubkeys[i] = addressPubKey
	}
	if len(createMultiSigParams.PublicKeys) != createMultiSigParams.N {
		return "", errors.New("Params error : PublicKeys small than n.")
	}
	//create multisig address
	pkScript, err := txscript.MultiSigScript(pubkeys, createMultiSigParams.M)
	if err != nil {
		return "", err
	}

	//multisig address scriptHash
	scriptAddr, err := btcutil.NewAddressScriptHash(pkScript, realNet)
	if err != nil {
		return "", err
	}
	//result for return
	var createMultiSigResult adaptor.CreateMultiSigResult
	createMultiSigResult.P2ShAddress = scriptAddr.String()
	createMultiSigResult.RedeemScript = hex.EncodeToString(pkScript)

	for _, pubkey := range pubkeys {
		createMultiSigResult.Addresses = append(createMultiSigResult.Addresses, pubkey.EncodeAddress())
	}

	jsonResult, err := json.Marshal(createMultiSigResult)
	if err != nil {
		return "", err
	}

	return string(jsonResult), nil
}
