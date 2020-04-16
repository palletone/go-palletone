// Copyright 2014 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package keystore

import (
	"fmt"
	"io"
	"time"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/crypto"
	"github.com/palletone/go-palletone/core/accounts"
	"github.com/pborman/uuid"
)

func newKeyOutchainFromECDSA(privateKeyECDSA []byte) *Key {
	id := uuid.NewRandom()
	pubKey, _ := crypto.MyCryptoLib.PrivateKeyToPubKey(privateKeyECDSA)
	key := &Key{
		Id:         id,
		Address:    crypto.PubkeyBytesToAddressOutchain(pubKey),
		PrivateKey: privateKeyECDSA,
		KeyType:    KeyType_Outchain_KEY,
	}
	return key
}

func newKeyOutchain(rand io.Reader) (*Key, error) {
	privateKeyECDSA, err := crypto.MyCryptoLib.KeyGen()
	if err != nil {
		return nil, err
	}
	return newKeyOutchainFromECDSA(privateKeyECDSA), nil
}

func storeNewKeyOutchain(ks keyStore, rand io.Reader, auth string) (*Key, accounts.Account, error) {
	key, err := newKeyOutchain(rand)
	if err != nil {
		return nil, accounts.Account{}, err
	}
	a := accounts.Account{Address: key.Address, URL: accounts.URL{Scheme: KeyStoreScheme,
		Path: ks.JoinPath(keyFileNameOutchain(key.Address))}}
	if err := ks.StoreKey(a.URL.Path, key, auth); err != nil {
		ZeroKey(key.PrivateKey)
		return nil, a, err
	}
	return key, a, err
}

// keyFileName implements the naming convention for keyfiles:
// UTC--<created_at UTC ISO8601>-<address hex>
func keyFileNameOutchain(keyAddr common.Address) string {
	ts := time.Now().UTC()
	return fmt.Sprintf("UTC--%s--%s", toISO8601(ts), keyAddr.Str())
}

