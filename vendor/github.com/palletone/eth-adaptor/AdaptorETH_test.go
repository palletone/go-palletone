package ethadaptor

import (
	"testing"

	"github.com/palletone/adaptor"
	"github.com/stretchr/testify/assert"
)

func TestAdaptorETH_NewPrivateKey(t *testing.T) {
	ada := NewAdaptorETHTestnet()
	output, err := ada.NewPrivateKey(nil)
	assert.Nil(t, err)
	t.Logf("New private key:%x,len:%d", output.PrivateKey, len(output.PrivateKey))
	pubKeyOutput, err := ada.GetPublicKey(&adaptor.GetPublicKeyInput{PrivateKey: output.PrivateKey})
	assert.Nil(t, err)
	t.Logf("Pub key:%x,len:%d", pubKeyOutput.PublicKey, len(pubKeyOutput.PublicKey))
	addrOutput, err := ada.GetAddress(&adaptor.GetAddressInput{Key: pubKeyOutput.PublicKey})
	assert.Nil(t, err)
	t.Logf("Address:%s", addrOutput.Address)
}
