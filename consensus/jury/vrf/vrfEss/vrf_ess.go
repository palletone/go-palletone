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
package vrfEss

import (
	"crypto/ecdsa"
	"github.com/palletone/go-palletone/common/crypto"
	"github.com/palletone/go-palletone/common/util"
)

func Evaluate(pri *ecdsa.PrivateKey,  msg []byte) (proof []byte, err error) {
	h := crypto.Keccak256Hash(util.RHashBytes(msg))
	sign, err := crypto.Sign(h.Bytes(), pri)
	if err != nil {
		return nil, err
	}
	return sign, nil
}

func VerifyWithPK(sign []byte, msg interface{}, publicKey []byte) bool {
	hash := crypto.Keccak256Hash(util.RHashBytes(msg))
	// sig := sign[:len(sign)-1] // remove recovery id
	return crypto.VerifySignature(publicKey, hash.Bytes(), sign)
}
