package keystore

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tyler-smith/go-bip39"
)

func TestSeedToMnemonic(t *testing.T) {
	entropy, err := bip39.NewEntropy(256)
	assert.Nil(t, err)
	t.Logf("entropy:%x", entropy)
	mnemonic, err := bip39.NewMnemonic(entropy)
	assert.Nil(t, err)
	t.Log(mnemonic)
	seed, err := bip39.NewSeedWithErrorChecking(mnemonic, "")
	assert.Nil(t, err)
	t.Logf("Seed:%x", seed)

}
