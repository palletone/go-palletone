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
	"crypto/ecdsa"

	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/crypto"
)

var (
	ErrKeyNotSupported = errors.New("only support ECC key")
	ErrEvalVRF         = errors.New("failed to evaluate vrf")
)

type Ec struct{
}
//Vrf returns the verifiable random function evaluated m and a NIZK proof
//func VrfProve(pri *ecdsa.PrivateKey, msg []byte) (vrfValue, vrfProof []byte, err error) {
func (e *Ec) VrfProve(priKey interface{}, msg []byte) (proof ,selData []byte, err error) {
	sk := priKey.(*ecdsa.PrivateKey)
	h := getHash(sk.Curve)
	//byteLen := (sk.Params().BitSize + 7) >> 3
	index, proof := Evaluate(sk, h, msg)
	if proof == nil {
		return proof, index[:], ErrEvalVRF
	}

	//vrfProof = proof[0 : 2*byteLen]
	//vrfValue = proof[2*byteLen : 2*byteLen+2*byteLen+1]
	return proof, index[:], nil
}

//Verify returns true if vrf and nizk is correct for msg
//func VrfVerify(pub *ecdsa.PublicKey, msg, vrfValue, vrfProof []byte) (bool, error) {
//func VrfVerify(pub *ecdsa.PublicKey, msg, proof []byte) (bool, error) {
func (e *Ec) VrfVerify(pubKey, msg, proof []byte) (bool, []byte, error) {
	pk:= crypto.P256ToECDSAPub(pubKey)
	if pk == nil {
		log.Error("VrfVerify, P256ToECDSAPub fail")
		return false, nil, errors.New("VrfVerify, P256ToECDSAPub fail")
	}
	h := getHash(pk.Curve)
	byteLen := (pk.Params().BitSize + 7) >> 3
	if len(proof) != byteLen*4+1  {
		return false, nil,nil
	}
	//proof := append(vrfProof, vrfValue...)
	index, err := ProofToHash(pk, h, msg, proof)
	if err != nil {
		log.Debugf("verifying VRF failed: %v", err)
		return false,nil,  nil
	}
	return true,index[:], nil
}
