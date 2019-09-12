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
	"testing"
	"crypto/rand"
	"crypto/ecdsa"
	"crypto/elliptic"
	"math/big"

	"github.com/btcsuite/btcd/btcec"
	"github.com/palletone/go-palletone/common/crypto"
)

var vType = new(Ec)
//func testVrf(t *testing.T, kt keypair.KeyType, curve byte) {
func testVrf1(t *testing.T) {
	//pubs := (*btcec.PublicKey)(pub.(*ecdsa.PublicKey)).SerializeCompressed()
	//t.Logf("pubs:%v", pubs)

	//key, err := crypto.GenerateKey()

	c := elliptic.P256() //ok
	//c := crypto.S256() //fail
	d, x, y, err := elliptic.GenerateKey(c, rand.Reader)

	//key, err := ecdsa.GenerateKey(c, rand.Reader) //fail
	//if err != nil {
	//	log.Error("couldn't generate key: ", "testVrf", err)
	//	return
	//}

	key := &ecdsa.PrivateKey{
		D: new(big.Int).SetBytes(d),
		PublicKey: ecdsa.PublicKey{
			X:     x,
			Y:     y,
			Curve: c,
		},
	}
	key.Curve.Params().Name = "P-256"

	msg := []byte("test")
	proof, _, err := vType.VrfProve(key, msg)
	if err != nil {
		t.Fatalf("compute vrf: %v", err)
	}
	pk:= (*btcec.PublicKey)(&key.PublicKey).SerializeCompressed()
	ret, _, err := vType.VrfVerify(pk, msg, proof)
	if err != nil {
		t.Fatalf("verify vrf: %v", err)
	}
	if !ret {
		t.Fatal("failed")
	}
}

func testVrf2(t *testing.T) { //todo  后面再继续修改为s256进行验证吧
	//t.Logf("pubs:%v", pubs)
	//key, err := crypto.GenerateKey()

	c := elliptic.P256() //ok
	//c := crypto.S256() //fail
	//d, x, y, err := elliptic.GenerateKey(c, rand.Reader)

	key, err := ecdsa.GenerateKey(c, rand.Reader) //fail
	if err != nil {
		return
	}

	//key := &ecdsa.PrivateKey{
	//	D: new(big.Int).SetBytes(d),
	//	PublicKey: ecdsa.PublicKey{
	//		X:     x,
	//		Y:     y,
	//		Curve: c,
	//	},
	//}
	key.Curve.Params().Name = "P-256" //"sm2p256v1"//"P-256"

	msg := []byte("test")
	proof, _, err := vType.VrfProve(key, msg)
	if err != nil {
		t.Fatalf("compute vrf: %v", err)
	}
	pk := crypto.P256FromECDSAPub(&key.PublicKey)
	ret, _, err := vType.VrfVerify(pk, msg, proof)
	if err != nil {
		t.Fatalf("verify vrf: %v", err)
	}
	if !ret {
		t.Fatal("failed")
	}
}

func TestVrf(t *testing.T) {
	//testVrf(t, keypair.PK_ECDSA, keypair.P224)
	//testVrf(t, keypair.PK_ECDSA, keypair.P256)
	//testVrf(t, keypair.PK_ECDSA, keypair.P384)
	//testVrf(t, keypair.PK_SM2, keypair.SM2P256V1)
	//	testVrf1(t) //ok
	testVrf2(t)
}

//func BenchmarkVrf(b *testing.B) {
//	pri, _, err := keypair.GenerateKeyPair(keypair.PK_ECDSA, keypair.P256)
//	if err != nil {
//		b.Fatal(err)
//	}
//	b.ResetTimer()
//
//	for i := 0; i < b.N; i++ {
//		msg := []byte("test")
//		Vrf(pri, msg)
//	}
//}
