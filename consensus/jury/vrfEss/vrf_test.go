package vrfEs

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