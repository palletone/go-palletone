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
package adaptoreth

import (
	"encoding/hex"

	"github.com/ethereum/go-ethereum/crypto"
	//"github.com/palletone/adaptor"
)

func NewPrivateKey(netID int) (prikeyHex string) {
	privateKeyECDSA, err := crypto.GenerateKey()
	if err != nil {
		return err.Error()
	}
	priHex := hex.EncodeToString(crypto.FromECDSA(privateKeyECDSA))
	//	fmt.Println(priHex)

	return priHex
}

func GetPublicKey(priKeyHex string, netID int) (pubKey string) {
	privateKeyECDSA, err := crypto.HexToECDSA(priKeyHex)
	if err != nil {
		return err.Error()
	}
	pubHex := hex.EncodeToString(crypto.CompressPubkey(&privateKeyECDSA.PublicKey))
	//	fmt.Println(pubHex)

	return pubHex
}

func GetAddress(priKeyHex string, netID int) (address string) {
	privateKeyECDSA, err := crypto.HexToECDSA(priKeyHex)
	if err != nil {
		return err.Error()
	}
	addr := crypto.PubkeyToAddress(privateKeyECDSA.PublicKey)
	//	fmt.Println(addr.String())

	return addr.String()
}
