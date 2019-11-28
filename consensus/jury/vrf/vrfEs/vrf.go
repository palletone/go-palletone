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
package vrfEs

import (
	"crypto/ecdsa"
	"github.com/btcsuite/btcd/btcec"
	"github.com/palletone/go-palletone/common/log"
)

type Es struct {
}

func (e *Es) VrfProve(priKey interface{}, msg []byte) (proof ,selData []byte, err error) {
	siger, err := NewVRFSigner(priKey.(*ecdsa.PrivateKey))
	if err != nil {
		log.Errorf("VrfProve, NewVRFSigner err:%s", err.Error())
		return nil, nil,err
	}
	idx, proof := siger.Evaluate(msg)
	log.Debugf("VrfProve, msg[%v], idx[%v], proof[%v]", msg, idx, proof)

	return proof, idx[:],nil
}

func (e *Es) VrfVerify(pubKey, msg, proof []byte) (bool, []byte, error) {
	key, err := btcec.ParsePubKey(pubKey, btcec.S256())
	if err != nil {
		log.Errorf("VrfVerify, parsePubKey error:%s", err.Error())
		return false, nil, err
	}
	pk, err := NewVRFVerifier(key.ToECDSA())
	if err != nil {
		log.Errorf("VrfVerify, NewVRFVerifier error:%s", err.Error())
		return false, nil, err
	}
	idx, err := pk.ProofToHash(msg, proof)
	if err != nil {
		log.Errorf("VrfVerify, ProofToHash error:%s", err.Error())
		return false, nil, err
	}
	return true, idx[:], nil
}
