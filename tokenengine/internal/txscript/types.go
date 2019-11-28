/*
 *
 *    This file is part of go-palletone.
 *    go-palletone is free software: you can redistribute it and/or modify
 *    it under the terms of the GNU General Public License as published by
 *    the Free Software Foundation, either version 3 of the License, or
 *    (at your option) any later version.
 *    go-palletone is distributed in the hope that it will be useful,
 *    but WITHOUT ANY WARRANTY; without even the implied warranty of
 *    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *    GNU General Public License for more details.
 *    You should have received a copy of the GNU General Public License
 *    along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
 * /
 *
 *  * @author PalletOne core developer <dev@pallet.one>
 *  * @date 2018-2019
 *
 */

package txscript

import (
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/crypto"
)

type AddressOriginalData struct {
	Address  common.Address
	Original []byte
}

func NewAddressOriginalData(data []byte, at ScriptClass) AddressOriginalData {
	hash160 := data
	if len(data) != 20 {
		hash160 = crypto.Hash160(data)
	}

	switch at {
	case PubKeyHashTy:
		return AddressOriginalData{Address: common.NewAddress(hash160, common.PublicKeyHash), Original: data}

	case ScriptHashTy:
		return AddressOriginalData{Address: common.NewAddress(hash160, common.ScriptHash), Original: data}

	case ContractHashTy:
		return AddressOriginalData{Address: common.NewAddress(hash160, common.ContractHash), Original: data}
	case PubKeyTy:

		return AddressOriginalData{Address: common.NewAddress(hash160, common.PublicKeyHash), Original: data}
	}
	return AddressOriginalData{Original: data}
}

//根据合约地址，获得该合约对应的陪审团赎回脚本
type PickupJuryRedeemScript func(common.Address) ([]byte, error)

type ICrypto interface {
	Hash(msg []byte) ([]byte, error)
	Sign(address common.Address, msg []byte) ([]byte, error)
	Verify(pubKey, signature, msg []byte) (bool, error)
	GetPubKey(address common.Address) ([]byte, error)
}
