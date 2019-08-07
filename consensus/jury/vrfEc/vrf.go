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
package vrfEc

import (
	"errors"
	"hash"
	"crypto"
	"crypto/elliptic"
	"crypto/ecdsa"

	"github.com/palletone/go-palletone/common/log"
)

var (
	ErrKeyNotSupported = errors.New("only support ECC key")
	ErrEvalVRF         = errors.New("failed to evaluate vrf")
)

func VrfProof2Value(curve *elliptic.CurveParams, proof []byte) []byte {
	params := curve.Params()
	//nilIndex := [32]byte{}
	byteLen := (params.BitSize + 7) >> 3

	if curve == nil || len(proof) != 4*byteLen+1 {
		return nil
	}
	//vrfProof = proof[0 : 2*byteLen]
	//vrfValue = proof[2*byteLen : 2*byteLen+2*byteLen+1]

	return proof[2*byteLen : 2*byteLen+2*byteLen+1]
}

//Vrf returns the verifiable random function evaluated m and a NIZK proof
//func VrfProve(pri *ecdsa.PrivateKey, msg []byte) (vrfValue, vrfProof []byte, err error) {
func VrfProve(pri *ecdsa.PrivateKey, msg []byte) (proof []byte, err error) {
	sk := pri
	h := getHash(sk.Curve)
	//byteLen := (sk.Params().BitSize + 7) >> 3
	_, proof = Evaluate(sk, h, msg)
	if proof == nil {
		return nil, ErrEvalVRF
	}

	//vrfProof = proof[0 : 2*byteLen]
	//vrfValue = proof[2*byteLen : 2*byteLen+2*byteLen+1]
	err = nil
	return
}

//Verify returns true if vrf and nizk is correct for msg
//func VrfVerify(pub *ecdsa.PublicKey, msg, vrfValue, vrfProof []byte) (bool, error) {
func VrfVerify(pub *ecdsa.PublicKey, msg, proof []byte) (bool, error) {
	pk := pub
	h := getHash(pk.Curve)
	byteLen := (pk.Params().BitSize + 7) >> 3
	if len(proof) != byteLen*4+1  {
		return false, nil
	}
	//proof := append(vrfProof, vrfValue...)
	_, err := ProofToHash(pk, h, msg, proof)
	if err != nil {
		log.Debugf("verifying VRF failed: %v", err)
		return false, nil
	}
	return true, nil
}

func getHash(curve elliptic.Curve) hash.Hash {
	bitSize := curve.Params().BitSize

	switch bitSize {
	case 224:
		return crypto.SHA224.New()
	case 256:
		return crypto.SHA256.New()	 //default
		//if curve.Params().Name == "sm2p256v1" {
		//	log.Debug("sm2p256v1 not support!!")
		//	//return sm3.New()
		//} else if curve.Params().Name == "P-256" {
		//	//return crypto.SHA256.New()
		//	return crypto.SHA3_256.New()
		//} else {
		//	return nil
		//}
	case 384:
		return crypto.SHA384.New()
	}
	return nil
}
