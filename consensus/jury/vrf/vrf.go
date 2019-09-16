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
package vrf

import (
	es "github.com/palletone/go-palletone/consensus/jury/vrf/vrfEs"
	//ec "github.com/palletone/go-palletone/consensus/jury/vrf/vrfEss"
	//ess "github.com/palletone/go-palletone/consensus/jury/vrf/vrfEss"
)

type Vrf interface {
	VrfProve(priKey interface{}, msg []byte) (proof, selData []byte, err error)
	VrfVerify(pubKey, msg, proof []byte) (verify bool, selData []byte, err error)
}

var vTpye = new(es.Es) //VRF type

func VrfProve(priKey interface{}, msg []byte) (proof, selData []byte, err error) {
	return vTpye.VrfProve(priKey, msg)
}

func VrfVerify(pubKey, msg, proof []byte) (verify bool, selData []byte, err error) {
	return vTpye.VrfVerify(pubKey, msg, proof)
}
