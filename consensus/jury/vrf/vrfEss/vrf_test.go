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
	"testing"
	"crypto/ecdsa"
	"crypto/rand"
	"github.com/palletone/go-palletone/common/crypto"
)

var vType =new(Ess)

func TestVrf(t *testing.T) {
	msg := []byte("test")
	//c := elliptic.P256() //ok
	c := crypto.S256() //fail
	//d, x, y, err := elliptic.GenerateKey(c, rand.Reader)

	priKey, err := ecdsa.GenerateKey(c, rand.Reader) //fail
	if err != nil {
		return
	}
	pubKey := &priKey.PublicKey
	proof,_, err := vType.VrfProve(priKey, msg)
	if err != nil {
		t.Fatalf("compute vrf: %v", err)
	}
	pk := crypto.CompressPubkey(pubKey)
	ret, _,err := vType.VrfVerify(pk, msg, proof)
	if err != nil {
		t.Fatalf("verify vrf: %v", err)
	}
	if !ret {
		t.Fatal("failed")
	}
}