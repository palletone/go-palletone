package crypto

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCryptoP256_Key(t *testing.T) {
	crypto:=&CryptoP256{}
	msg := []byte("ABC")

	privKey,err:= crypto.KeyGen()
	assert.Nil(t,err)
	t.Logf("PrivateKey:%x",privKey)

	assert.Nil(t,err)
	pubKey,err:=crypto.PrivateKeyToPubKey(privKey)
	t.Logf("Pubkey:%x,len:%d",pubKey,len(pubKey))

	sign,err:= crypto.Sign(privKey,msg)
	assert.Nil(t,err)
	t.Logf("Signature:%x,len:%d",sign,len(sign))
	pass1,err:= crypto.Verify(pubKey,sign,msg)
	assert.Nil(t,err)
	assert.True(t,pass1)

	addr1 := PubkeyBytesToAddress(pubKey)
	address1 := addr1.String()
	t.Logf("Address:%s",address1)
	}

