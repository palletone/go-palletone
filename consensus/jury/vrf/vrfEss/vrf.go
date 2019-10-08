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
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/dag/errors"
)

type Ess struct {
}

func (e *Ess) VrfProve(priKey interface{}, msg []byte) (proof, selData []byte, err error) {
	proof, err = Evaluate(priKey.(*ecdsa.PrivateKey), msg)
	if err != nil {
		log.Error("VrfProve Evaluate fail")
	}
	return proof, proof, nil
}

func (e *Ess) VrfVerify(pk, msg, proof []byte) (bool, []byte, error) {
	if pk == nil || msg == nil || proof == nil {
		log.Error("VrfVerify param is nil")
		return false, nil, errors.New("VrfVerify fail, param is nil")
	}
	return VerifyWithPK(proof, msg, pk), proof, nil
}
