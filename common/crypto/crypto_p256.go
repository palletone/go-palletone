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
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/asn1"
	"fmt"
	"github.com/palletone/go-palletone/dag/errors"
	"hash"
	"math/big"
)

type CryptoP256 struct {
}

type ECDSASignature struct {
	R, S *big.Int
}

func (c *CryptoP256) KeyGen() (privKey []byte, err error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	return FromECDSA(key), nil
}

func P256ToECDSA(d []byte) (*ecdsa.PrivateKey, error) {
	strict := false
	priv := new(ecdsa.PrivateKey)
	priv.PublicKey.Curve = elliptic.P256()
	if strict && 8*len(d) != priv.Params().BitSize {
		return nil, fmt.Errorf("invalid length, need %d bits", priv.Params().BitSize)
	}
	priv.D = new(big.Int).SetBytes(d)

	// // The priv.D must < N
	// if priv.D.Cmp(secp256k1_N) >= 0 {
	// 	return nil, fmt.Errorf("invalid private key, >=N")
	// }
	// // The priv.D must not be zero or negative.
	// if priv.D.Sign() <= 0 {
	// 	return nil, fmt.Errorf("invalid private key, zero or negative")
	// }

	priv.PublicKey.X, priv.PublicKey.Y = priv.PublicKey.Curve.ScalarBaseMult(d)
	if priv.PublicKey.X == nil {
		return nil, errors.New("invalid private key")
	}
	return priv, nil
}

func P256ToECDSAPub(pub []byte) *ecdsa.PublicKey {
	if len(pub) == 0 {
		return nil
	}
	x, y := elliptic.Unmarshal(elliptic.P256(), pub)
	return &ecdsa.PublicKey{Curve: elliptic.P256(), X: x, Y: y}
}

func P256FromECDSAPub(pub *ecdsa.PublicKey) []byte {
	if pub == nil || pub.X == nil || pub.Y == nil {
		return nil
	}
	return elliptic.Marshal(elliptic.P256(), pub.X, pub.Y)
}

func (c *CryptoP256) PrivateKeyToPubKey(privKey []byte) ([]byte, error) {
	prvKey, err := P256ToECDSA(privKey)
	if err != nil {
		return nil, err
	}
	pubKey := P256FromECDSAPub(&prvKey.PublicKey)
	return pubKey, nil
}
func (c *CryptoP256) PrivateKeyToInstance(privKey []byte) (interface{}, error) {
	return P256ToECDSA(privKey)
}
func (c *CryptoP256) Hash(msg []byte) (hash []byte, err error) {
	d := sha256.New()
	d.Write(msg)
	return d.Sum(nil), nil
}

func (c *CryptoP256) GetHash() (h hash.Hash, err error) {
	return sha256.New(), nil
}
func (c *CryptoP256) Sign(privKey, message []byte) (signature []byte, err error) {
	prvKey, err := P256ToECDSA(privKey)
	if err != nil {
		return nil, err
	}
	digest, err := c.Hash(message)
	if err != nil {
		return nil, err
	}
	r, s, err := ecdsa.Sign(rand.Reader, prvKey, digest)
	if err != nil {
		return nil, err
	}
	return asn1.Marshal(ECDSASignature{r, s})
}

func (c *CryptoP256) Verify(pubKey, signature []byte, message []byte) (valid bool, err error) {
	pub := P256ToECDSAPub(pubKey)
	var s ECDSASignature
	_, err = asn1.Unmarshal(signature, &s)
	if err != nil {
		return false, err
	}
	digest, err := c.Hash(message)
	if err != nil {
		return false, err
	}
	result := ecdsa.Verify(pub, digest, s.R, s.S)
	return result, nil
}

func (c *CryptoP256) Encrypt(key []byte, plaintext []byte) (ciphertext []byte, err error) {
	return nil, errors.New("Not implement")
}
func (c *CryptoP256) Decrypt(key, ciphertext []byte) (plaintext []byte, err error) {
	return nil, errors.New("Not implement")
}
