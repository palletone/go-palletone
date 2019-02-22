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
 
package algorithm

import (

	"crypto/ecdsa"
)

type PublicKey struct {
	//pk ed25519.PublicKey
	pk ecdsa.PublicKey
}

//func (pub *PublicKey) Bytes() []byte {
//	return pub.pk
//}

//func (pub *PublicKey) Address() common.Address {
//	return common.BytesToAddress(pub.pk)
//}

//func (pub *PublicKey) VerifySign(m, sign []byte) error {
//	signature := sign[ed25519.PublicKeySize:]
//	if ok := ed25519.Verify(pub.pk, m, signature); !ok {
//		return fmt.Errorf("signature invalid")
//	}
//	return nil
//}

//验证vrf
//func (pub *PublicKey) VerifyVRF(proof, m []byte) (bool, error) {
//	return vrf.ECVRF_verify(pub.pk, proof, m)
//}

type PrivateKey struct {
	//sk ed25519.PrivateKey
	sk ecdsa.PrivateKey
}

func (priv *PrivateKey) PublicKey() *PublicKey {
	//return &PublicKey{priv.sk.Public().(ed25519.PublicKey)}
	return &PublicKey{priv.sk.Public().(ecdsa.PublicKey)}
}

//func (priv *PrivateKey) Sign(m []byte) ([]byte, error) {
//	sign, err := priv.sk.Sign(rand.Reader, m, crypto.Hash(0))
//	if err != nil {
//		return nil, err
//	}
//	//pubkey := priv.sk.Public().(ed25519.PublicKey)
//	pubkey := priv.sk.Public().(ecdsa.PublicKey)
//	return append(pubkey, sign...), nil
//}

//产生vrf证明
//func (priv *PrivateKey) Evaluate(m []byte) (value, proof []byte, err error) {
//	proof, err = vrf.ECVRF_prove(priv.PublicKey().pk, priv.sk, m)
//	if err != nil {
//		return
//	}
//	value = vrf.ECVRF_proof2hash(proof)
//	return
//}

//func NewKeyPair() (*PublicKey, *PrivateKey, error) {
//	pk, sk, err := ed25519.GenerateKey(rand.Reader)
//	if err != nil {
//		return nil, nil, err
//	}
//
//	return &PublicKey{pk}, &PrivateKey{sk}, nil
//}

//func recoverPubkey(sign []byte) *PublicKey {
//	pubkey := sign[:ed25519.PublicKeySize]
//	return &PublicKey{pubkey}
//}
