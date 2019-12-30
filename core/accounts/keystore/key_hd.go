package keystore

import (
	"github.com/palletone/go-palletone/common/crypto"
	"github.com/palletone/go-palletone/core/accounts"
	"github.com/pborman/uuid"
	"github.com/tyler-smith/go-bip32"
	"github.com/tyler-smith/go-bip39"
)

func storeNewHdSeed(ks keyStore, auth string) (*Key, accounts.Account, error) {

	key, err := newSeed()
	if err != nil {
		return nil, accounts.Account{}, err
	}

	a := accounts.Account{Address: key.Address, URL: accounts.URL{Scheme: KeyStoreScheme,
		Path: ks.JoinPath(keyFileName(key.Address))}}
	if err := ks.StoreKey(a.URL.Path, key, auth); err != nil {
		ZeroKey(key.PrivateKey)
		return nil, a, err
	}
	return key, a, err
}
func newSeed() (*Key, error) {
	entropy, err := bip39.NewEntropy(256)
	if err != nil {
		return nil, err
	}
	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return nil, err
	}
	// Generate a Bip32 HD wallet for the mnemonic and a user supplied password
	seed := bip39.NewSeed(mnemonic, "")
	return newKeyFromHdSeed(seed), nil
}
func newKeyFromHdSeed(hdSeed []byte) *Key {
	id := uuid.NewRandom()
	key0, _ := newKey0(hdSeed)
	pubKey := key0.PublicKey().Key

	key := &Key{
		Id:         id,
		Address:    crypto.PubkeyBytesToAddress(pubKey),
		PrivateKey: hdSeed,
	}
	return key
}
func newKey0(seed []byte) (*bip32.Key, error) {
	masterKey, err := bip32.NewMasterKey(seed)
	if err != nil {
		return nil, err
	}
	return NewKeyFromMasterKey(masterKey, PTN_COIN_TYPE, ACCOUNT0, 0, 0)
}

const Purpose uint32 = 0x8000002C
const PTN_COIN_TYPE uint32 = 0x8050544e //PTN
const ACCOUNT0 = 0x80000000

func NewKeyFromMasterKey(masterKey *bip32.Key, coin, account, chain, addressIndex uint32) (*bip32.Key, error) {
	child, err := masterKey.NewChildKey(Purpose)
	if err != nil {
		return nil, err
	}

	child, err = child.NewChildKey(coin)
	if err != nil {
		return nil, err
	}

	child, err = child.NewChildKey(account)
	if err != nil {
		return nil, err
	}

	child, err = child.NewChildKey(chain)
	if err != nil {
		return nil, err
	}

	key, err := child.NewChildKey(addressIndex)
	if err != nil {
		return nil, err
	}

	return key, nil
}
