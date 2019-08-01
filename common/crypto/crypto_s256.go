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
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
	"github.com/btcsuite/btcd/btcec"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/dag/errors"
	"golang.org/x/crypto/sha3"
	"hash"
	"math/big"
)

type CryptoS256 struct {
}

func (c *CryptoS256) KeyGen() (privKey []byte, err error) {
	key, err := ecdsa.GenerateKey(btcec.S256(), rand.Reader)
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
	return compressPubkey(&pubKey), nil
}
func (c *CryptoS256) PrivateKeyToInstance(privKey []byte) (interface{}, error){
	return ToECDSA(privKey)
}
// DecompressPubkey parses a public key in the 33-byte compressed format.
//func decompressPubkey(pubkey []byte) (*ecdsa.PublicKey, error) {
//	if len(pubkey) != 33 {
//		return nil, errors.New("invalid compressed public key length")
//	}
//	key, err := btcec.ParsePubKey(pubkey, btcec.S256())
//	if err != nil {
//		return nil, err
//	}
//	return key.ToECDSA(), nil
//}

// CompressPubkey encodes a public key to the 33-byte compressed format.
func compressPubkey(pubkey *ecdsa.PublicKey) []byte {
	return (*btcec.PublicKey)(pubkey).SerializeCompressed()
}
func (c *CryptoS256) Hash(msg []byte) (hash []byte, err error) {
	d := sha3.New256()
	d.Write(msg)
	return d.Sum(nil), nil
}
func (c *CryptoS256) GetHash() (h hash.Hash, err error) {
	return sha3.New256(), nil
}
func (c *CryptoS256) Sign(privKey, message []byte) (signature []byte, err error) {
	prvKey, err := ToECDSA(privKey)
	if err != nil {
		return nil, err
	}
	digest, err := c.Hash(message)
	if err != nil {
		return nil, err
	}
	return sign(digest, prvKey)
}
func sign(hash []byte, prv *ecdsa.PrivateKey) ([]byte, error) {
	if len(hash) != 32 {
		return nil, fmt.Errorf("hash is required to be exactly 32 bytes (%d)", len(hash))
	}
	if prv.Curve != btcec.S256() {
		return nil, fmt.Errorf("private key curve is not secp256k1")
	}
	key := (*btcec.PrivateKey)(prv)
	sign, _ := key.Sign(hash)
	return sign.Serialize(), nil
}
func (c *CryptoS256) Verify(pubKey, signature, message []byte) (valid bool, err error) {
	key, err := btcec.ParsePubKey(pubKey, btcec.S256())
	if err != nil {
		log.Info("parsePubKey error:" + err.Error())
		return false, err
	}
	var sig *btcec.Signature
	if len(signature) == 64 { // R||S
		sig = &btcec.Signature{R: new(big.Int).SetBytes(signature[:32]), S: new(big.Int).SetBytes(signature[32:])}
	} else {
		sig, err = btcec.ParseSignature(signature, btcec.S256())
		if err != nil {
			log.Info("ParseSignature error:" + err.Error())
			return false, err
		}
	}
	// Reject malleable signatures. libsecp256k1 does this check but btcec doesn't.
	if sig.S.Cmp(secp256k1_halfN) > 0 {

		return false, errors.New("sig.S.Cmp > 0")
	}
	digest, err := c.Hash(message)
	if err != nil {
		return false, err
	}
	return sig.Verify(digest, key), nil
}
func (c *CryptoS256) Encrypt(key []byte, plaintext []byte) (ciphertext []byte, err error) {
	return nil, errors.New("Not implement")
}
func (c *CryptoS256) Decrypt(key, ciphertext []byte) (plaintext []byte, err error) {
	return nil, errors.New("Not implement")
}
