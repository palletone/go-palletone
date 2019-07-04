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
 *  * @author PalletOne core developers <dev@pallet.one>
 *  * @date 2018-2019
 *  
 *
 */

package crypto

import (
	"hash"
	"github.com/tjfoc/gmsm/sm2"
	"fmt"
	"github.com/palletone/go-palletone/common/math"
	"math/big"
	"errors"
	"github.com/tjfoc/gmsm/sm3"
	"encoding/asn1"
)

type CryptoGm struct {
}

func (c *CryptoGm) KeyGen() ( []byte,  error) {
	privKey, err := sm2.GenerateKey()
	if err != nil {
		return nil, fmt.Errorf("Failed generating GMSM2 key  [%s]", err)
	}

	return sm2FromECDSA(privKey), nil
}
func sm2FromECDSA(priv *sm2.PrivateKey) []byte {
	if priv == nil {
		return nil
	}
	return math.PaddedBigBytes(priv.D, priv.Params().BitSize/8)
}
func sm2ToECDSA(d []byte ) (*sm2.PrivateKey, error) {
	strict:=false
	priv := new(sm2.PrivateKey)
	priv.PublicKey.Curve = sm2.P256Sm2()
	if strict && 8*len(d) != priv.Params().BitSize {
		return nil, fmt.Errorf("invalid length, need %d bits", priv.Params().BitSize)
	}
	priv.D = new(big.Int).SetBytes(d)

	// The priv.D must < N
	if priv.D.Cmp(secp256k1_N) >= 0 {
		return nil, fmt.Errorf("invalid private key, >=N")
	}
	// The priv.D must not be zero or negative.
	if priv.D.Sign() <= 0 {
		return nil, fmt.Errorf("invalid private key, zero or negative")
	}

	priv.PublicKey.X, priv.PublicKey.Y = priv.PublicKey.Curve.ScalarBaseMult(d)
	if priv.PublicKey.X == nil {
		return nil, errors.New("invalid private key")
	}
	return priv, nil
}
func (c *CryptoGm) PrivateKeyToPubKey(privKey []byte) ([]byte, error) {
	prvKey, err := sm2ToECDSA(privKey)
	if err != nil {
		return nil, err
	}
	pubKey := prvKey.PublicKey

	return sm2.Compress(&pubKey), nil
}
func (c *CryptoGm) Hash(msg []byte) (hash []byte, err error) {
	d := sm3.New()
	d.Write(msg)
	return d.Sum(nil),nil
}
func (c *CryptoGm) GetHash() (h hash.Hash, err error) {
	return sm3.New(), nil
}
func (c *CryptoGm) Sign(privKey, digest []byte) (signature []byte, err error) {
	prvKey, err := sm2ToECDSA(privKey)
	if err != nil {
		return nil, err
	}
	r,s,err:= sm2.Sign(prvKey,digest)
	if err!=nil{
		return nil,err
	}
	return MarshalSM2Signature(r,s)
	//return Sign(digest, prvKey)
}
type SM2Signature struct {
	R, S *big.Int
}
func MarshalSM2Signature(r, s *big.Int) ([]byte, error) {
	return asn1.Marshal(SM2Signature{r, s})
}

func UnmarshalSM2Signature(raw []byte) (*big.Int, *big.Int, error) {
	// Unmarshal
	sig := new(SM2Signature)
	_, err := asn1.Unmarshal(raw, sig)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed unmashalling signature [%s]", err)
	}

	// Validate sig
	if sig.R == nil {
		return nil, nil, errors.New("Invalid signature. R must be different from nil.")
	}
	if sig.S == nil {
		return nil, nil, errors.New("Invalid signature. S must be different from nil.")
	}

	if sig.R.Sign() != 1 {
		return nil, nil, errors.New("Invalid signature. R must be larger than zero")
	}
	if sig.S.Sign() != 1 {
		return nil, nil, errors.New("Invalid signature. S must be larger than zero")
	}

	return sig.R, sig.S, nil
}
func (c *CryptoGm) Verify(pubKey, signature, digest []byte) (valid bool, err error) {
	r,s,err:= UnmarshalSM2Signature(signature)
	if err!=nil{
		return false,err
	}
	publicKey:= sm2.Decompress(pubKey)
	return sm2.Verify(publicKey,digest,r,s),nil
	//return VerifySignature(pubKey, digest, signature), nil
}
func (c *CryptoGm) Encrypt(key []byte, plaintext []byte) (ciphertext []byte, err error) {
	return nil, errors.New("Not implement")
}
func (c *CryptoGm) Decrypt(key, ciphertext []byte) (plaintext []byte, err error) {
	return nil, errors.New("Not implement")
}

