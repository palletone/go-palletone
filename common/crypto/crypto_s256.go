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

package crypto

import (
	"github.com/palletone/go-palletone/dag/errors"
	"golang.org/x/crypto/sha3"
	"hash"
)

type CryptoS256 struct {
}

func (c *CryptoS256) KeyGen() (privKey []byte, err error) {
	key, err := GenerateKey()
	if err != nil {
		return nil, err
	}
	return FromECDSA(key), nil
}
func (c *CryptoS256) PrivateKeyToPubKey(privKey []byte) ([]byte, error) {
	prvKey, err := ToECDSA(privKey)
	if err != nil {
		return nil, err
	}
	pubKey := prvKey.PublicKey
	return CompressPubkey(&pubKey), nil
}
func (c *CryptoS256) Hash(msg []byte) (hash []byte, err error) {
	return Keccak256(msg), nil
}
func (c *CryptoS256) GetHash() (h hash.Hash, err error) {
	return sha3.New256(), nil
}
func (c *CryptoS256) Sign(privKey, digest []byte) (signature []byte, err error) {
	prvKey, err := ToECDSA(privKey)
	if err != nil {
		return nil, err
	}
	return Sign(digest, prvKey)
}
func (c *CryptoS256) Verify(pubKey, signature, digest []byte) (valid bool, err error) {
	return VerifySignature(pubKey, digest, signature), nil
}
func (c *CryptoS256) Encrypt(key []byte, plaintext []byte) (ciphertext []byte, err error) {
	return nil, errors.New("Not implement")
}
func (c *CryptoS256) Decrypt(key, ciphertext []byte) (plaintext []byte, err error) {
	return nil, errors.New("Not implement")
}
