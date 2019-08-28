// Copyright 2017 The go-ethereum Authors
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

// Package keystore implements encrypted storage of secp256k1 private keys.
//
// Keys are stored as encrypted JSON files according to the Web3 Secret Storage specification.
// See https://github.com/ethereum/wiki/wiki/Web3-Secret-Storage-Definition for more information.
package keystore

import (
	crand "crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/crypto"
	"github.com/palletone/go-palletone/common/event"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/core/accounts"
	"github.com/palletone/go-palletone/dag/modules"
)

var (
	ErrLocked  = accounts.NewAuthNeededError("password or unlock")
	ErrNoMatch = errors.New("no key for given address or file")
	ErrDecrypt = errors.New("could not decrypt key with given passphrase")
)

// KeyStoreType is the reflect type of a keystore backend.
var KeyStoreType = reflect.TypeOf(&KeyStore{})

// KeyStoreScheme is the protocol scheme prefixing account and wallet URLs.
var KeyStoreScheme = "keystore"

// Maximum time between wallet refreshes (if filesystem notifications don't work).
const walletRefreshCycle = 3 * time.Second

// KeyStore manages a key storage directory on disk.
type KeyStore struct {
	storage  keyStore                     // Storage backend, might be cleartext or encrypted
	cache    *accountCache                // In-memory account cache over the filesystem storage
	changes  chan struct{}                // Channel receiving change notifications from the cache
	unlocked map[common.Address]*unlocked // Currently unlocked account (decrypted private keys)

	wallets     []accounts.Wallet       // Wallet wrappers around the individual key files
	updateFeed  event.Feed              // Event feed to notify wallet additions/removals
	updateScope event.SubscriptionScope // Subscription scope tracking current live listeners
	updating    bool                    // Whether the event notification loop is running

	mu sync.RWMutex
}

type unlocked struct {
	*Key
	abort chan struct{}
}

// NewKeyStore creates a keystore for the given directory.
func NewKeyStore(keydir string, scryptN, scryptP int) *KeyStore {
	keydir, _ = filepath.Abs(keydir)
	ks := &KeyStore{storage: &keyStorePassphrase{keydir, scryptN, scryptP}}
	ks.init(keydir)
	return ks
}

// NewPlaintextKeyStore creates a keystore for the given directory.
// Deprecated: Use NewKeyStore.
func NewPlaintextKeyStore(keydir string) *KeyStore {
	keydir, _ = filepath.Abs(keydir)
	ks := &KeyStore{storage: &keyStorePlain{keydir}}
	ks.init(keydir)
	return ks
}

func (ks *KeyStore) init(keydir string) {
	// Lock the mutex since the account cache might call back with events
	ks.mu.Lock()
	defer ks.mu.Unlock()

	// Initialize the set of unlocked keys and the account cache
	ks.unlocked = make(map[common.Address]*unlocked)
	ks.cache, ks.changes = newAccountCache(keydir)

	// TODO: In order for this finalizer to work, there must be no references
	// to ks. addressCache doesn't keep a reference but unlocked keys do,
	// so the finalizer will not trigger until all timed unlocks have expired.
	runtime.SetFinalizer(ks, func(m *KeyStore) {
		m.cache.close()
	})
	// Create the initial list of wallets from the cache
	accs := ks.cache.accounts()
	ks.wallets = make([]accounts.Wallet, len(accs))
	for i := 0; i < len(accs); i++ {
		ks.wallets[i] = &keystoreWallet{account: accs[i], keystore: ks}
	}
}

// Wallets implements accounts.Backend, returning all single-key wallets from the
// keystore directory.
func (ks *KeyStore) Wallets() []accounts.Wallet {
	// Make sure the list of wallets is in sync with the account cache
	ks.refreshWallets()

	ks.mu.RLock()
	defer ks.mu.RUnlock()

	cpy := make([]accounts.Wallet, len(ks.wallets))
	copy(cpy, ks.wallets)
	return cpy
}

// refreshWallets retrieves the current account list and based on that does any
// necessary wallet refreshes.
func (ks *KeyStore) refreshWallets() {
	// Retrieve the current list of accounts
	ks.mu.Lock()
	accs := ks.cache.accounts()

	// Transform the current list of wallets into the new one
	wallets := make([]accounts.Wallet, 0, len(accs))
	events := []accounts.WalletEvent{}

	for _, account := range accs {
		// Drop wallets while they were in front of the next account
		for len(ks.wallets) > 0 && ks.wallets[0].URL().Cmp(account.URL) < 0 {
			events = append(events, accounts.WalletEvent{Wallet: ks.wallets[0], Kind: accounts.WalletDropped})
			ks.wallets = ks.wallets[1:]
		}
		// If there are no more wallets or the account is before the next, wrap new wallet
		if len(ks.wallets) == 0 || ks.wallets[0].URL().Cmp(account.URL) > 0 {
			wallet := &keystoreWallet{account: account, keystore: ks}

			events = append(events, accounts.WalletEvent{Wallet: wallet, Kind: accounts.WalletArrived})
			wallets = append(wallets, wallet)
			continue
		}
		// If the account is the same as the first wallet, keep it
		if ks.wallets[0].Accounts()[0] == account {
			wallets = append(wallets, ks.wallets[0])
			ks.wallets = ks.wallets[1:]
			continue
		}
	}
	// Drop any leftover wallets and set the new batch
	for _, wallet := range ks.wallets {
		events = append(events, accounts.WalletEvent{Wallet: wallet, Kind: accounts.WalletDropped})
	}
	ks.wallets = wallets
	ks.mu.Unlock()

	// Fire all wallet events and return
	for _, event := range events {
		ks.updateFeed.Send(event)
	}
}

// Subscribe implements accounts.Backend, creating an async subscription to
// receive notifications on the addition or removal of keystore wallets.
func (ks *KeyStore) Subscribe(sink chan<- accounts.WalletEvent) event.Subscription {
	// We need the mutex to reliably start/stop the update loop
	ks.mu.Lock()
	defer ks.mu.Unlock()

	// Subscribe the caller and track the subscriber count
	sub := ks.updateScope.Track(ks.updateFeed.Subscribe(sink))

	// Subscribers require an active notification loop, start it
	if !ks.updating {
		ks.updating = true
		go ks.updater()
	}
	return sub
}

// updater is responsible for maintaining an up-to-date list of wallets stored in
// the keystore, and for firing wallet addition/removal events. It listens for
// account change events from the underlying account cache, and also periodically
// forces a manual refresh (only triggers for systems where the filesystem notifier
// is not running).
func (ks *KeyStore) updater() {
	for {
		// Wait for an account update or a refresh timeout
		select {
		case <-ks.changes:
		case <-time.After(walletRefreshCycle):
		}
		// Run the wallet refresher
		ks.refreshWallets()

		// If all our subscribers left, stop the updater
		ks.mu.Lock()
		if ks.updateScope.Count() == 0 {
			ks.updating = false
			ks.mu.Unlock()
			return
		}
		ks.mu.Unlock()
	}
}

// HasAddress reports whether a key with the given address is present.
func (ks *KeyStore) HasAddress(addr common.Address) bool {
	return ks.cache.hasAddress(addr)
}

// Accounts returns all key files present in the directory.
func (ks *KeyStore) Accounts() []accounts.Account {
	return ks.cache.accounts()
}

// Delete deletes the key matched by account if the passphrase is correct.
// If the account contains no filename, the address must match a unique key.
func (ks *KeyStore) Delete(a accounts.Account, passphrase string) error {
	// Decrypting the key isn't really necessary, but we do
	// it anyway to check the password and zero out the key
	// immediately afterwards.
	a, key, err := ks.getDecryptedKey(a, passphrase)
	if key != nil {
		ZeroKey(key.PrivateKey)
	}
	if err != nil {
		return err
	}
	// The order is crucial here. The key is dropped from the
	// cache after the file is gone so that a reload happening in
	// between won't insert it into the cache again.
	err = os.Remove(a.URL.Path)
	if err == nil {
		ks.cache.delete(a)
		ks.refreshWallets()
	}
	return err
}

// SignHash calculates a ECDSA signature for the given hash. The produced
// signature is in the [R || S ] format .
func (ks *KeyStore) SignMessage(addr common.Address, msg []byte) ([]byte, error) {
	// Look up the key to sign with and abort if it cannot be found
	ks.mu.RLock()
	defer ks.mu.RUnlock()

	unlockedKey, found := ks.unlocked[addr]
	if !found {
		return nil, ErrLocked
	}
	return crypto.MyCryptoLib.Sign(unlockedKey.PrivateKey, msg)
	// Sign the hash using plain ECDSA operations
	//return crypto.Sign(hash, unlockedKey.PrivateKey)
}

// SignTx signs the given transaction with the requested account.
func (ks *KeyStore) SignTx(a accounts.Account, tx *modules.Transaction,
	chainID *big.Int) (*modules.Transaction, error) {
	//R, S, V, err := ks.SigTX(tx, a.Address)
	//if err != nil {
	//	return nil, err
	//}
	//// publicKey, err1 := ks.GetPublicKey(a.Address)
	//// if err1 != nil {
	//// 	return nil, err1
	//// }
	//
	//if tx.From == nil {
	//	tx.From = new(modules.Authentifier)
	//}
	//tx.From.Address = a.Address.String()
	//tx.From.R = R
	//tx.From.S = S
	//tx.From.V = V
	return tx, nil
}

// SignHashWithPassphrase signs hash if the private key matching the given address
// can be decrypted with the given passphrase. The produced signature is in the
// [R || S ] format where V is 0 or 1.
func (ks *KeyStore) SignMessageWithPassphrase(a accounts.Account, passphrase string,
	msg []byte) (signature []byte, err error) {
	_, key, err := ks.getDecryptedKey(a, passphrase)
	if err != nil {
		return nil, err
	}
	defer ZeroKey(key.PrivateKey)
	return crypto.MyCryptoLib.Sign(key.PrivateKey, msg)
	//return crypto.Sign(hash, key.PrivateKey)
}
func (ks *KeyStore) VerifySignatureWithPassphrase(a accounts.Account, passphrase string, hash []byte,
	signature []byte) (pass bool, err error) {
	_, key, err := ks.getDecryptedKey(a, passphrase)
	if err != nil {
		return false, err
	}
	defer ZeroKey(key.PrivateKey)

	pk, _ := crypto.MyCryptoLib.PrivateKeyToPubKey(key.PrivateKey)
	return crypto.MyCryptoLib.Verify(pk, signature, hash)
	//sig := signature[:len(signature)-1] // remove recovery id
	//return crypto.VerifySignature(crypto.FromECDSAPub(&pk), hash, sig), nil
}

// SignTxWithPassphrase signs the transaction if the private key matching the
// given address can be decrypted with the given passphrase.
func (ks *KeyStore) SignTxWithPassphrase(a accounts.Account, passphrase string, tx *modules.Transaction,
	chainID *big.Int) (*modules.Transaction, error) {
	_, key, err := ks.getDecryptedKey(a, passphrase)
	if err != nil {
		return nil, err
	}
	defer ZeroKey(key.PrivateKey)

	//authen, err := ks.SigTXWithPwd(tx, key.PrivateKey)
	//if err != nil {
	//	return nil, err
	//}
	// publicKey := crypto.FromECDSAPub(&key.PrivateKey.PublicKey)

	//if tx.From == nil {
	//	tx.From = new(modules.Authentifier)
	//}
	//
	//tx.From.Address = a.Address.String()
	//tx.From.R = authen[:32]
	//tx.From.S = authen[32:64]
	//tx.From.V = authen[64:]
	return tx, nil
}

// Unlock unlocks the given account indefinitely.
func (ks *KeyStore) Unlock(a accounts.Account, passphrase string) error {
	return ks.TimedUnlock(a, passphrase, 0)
}

// Lock removes the private key with the given address from memory.
func (ks *KeyStore) Lock(addr common.Address) error {
	ks.mu.Lock()
	if unl, found := ks.unlocked[addr]; found {
		ks.mu.Unlock()
		ks.expire(addr, unl, time.Duration(0)*time.Nanosecond)
	} else {
		ks.mu.Unlock()
	}
	return nil
}

// TimedUnlock unlocks the given account with the passphrase. The account
// stays unlocked for the duration of timeout. A timeout of 0 unlocks the account
// until the program exits. The account must match a unique key file.
//
// If the account address is already unlocked for a duration, TimedUnlock extends or
// shortens the active unlock timeout. If the address was previously unlocked
// indefinitely the timeout is not altered.
func (ks *KeyStore) TimedUnlock(a accounts.Account, passphrase string, timeout time.Duration) error {
	a, key, err := ks.getDecryptedKey(a, passphrase)
	if err != nil {
		return err
	}

	ks.mu.Lock()
	defer ks.mu.Unlock()
	u, found := ks.unlocked[a.Address]
	if found {
		if u.abort == nil {
			// The address was unlocked indefinitely, so unlocking
			// it with a timeout would be confusing.
			ZeroKey(key.PrivateKey)
			return nil
		}
		// Terminate the expire goroutine and replace it below.
		close(u.abort)
	}
	if timeout > 0 {
		u = &unlocked{Key: key, abort: make(chan struct{})}
		go ks.expire(a.Address, u, timeout)
	} else {
		u = &unlocked{Key: key}
	}
	ks.unlocked[a.Address] = u
	return nil
}

func (ks *KeyStore) IsUnlock(addr common.Address) bool {
	ks.mu.Lock()
	_, found := ks.unlocked[addr]
	ks.mu.Unlock()

	return found
}

// Find resolves the given account into a unique entry in the keystore.
func (ks *KeyStore) Find(a accounts.Account) (accounts.Account, error) {
	ks.cache.maybeReload()
	ks.cache.mu.Lock()
	a, err := ks.cache.find(a)
	ks.cache.mu.Unlock()
	return a, err
}

func (ks *KeyStore) getDecryptedKey(a accounts.Account, auth string) (accounts.Account, *Key, error) {
	a, err := ks.Find(a)
	if err != nil {
		return a, nil, err
	}
	key, err := ks.storage.GetKey(a.Address, a.URL.Path, auth)
	return a, key, err
}

func (ks *KeyStore) expire(addr common.Address, u *unlocked, timeout time.Duration) {
	t := time.NewTimer(timeout)
	defer t.Stop()
	select {
	case <-u.abort:
		// just quit
	case <-t.C:
		ks.mu.Lock()
		// only drop if it's still the same key instance that dropLater
		// was launched with. we can check that using pointer equality
		// because the map stores a new pointer every time the key is
		// unlocked.
		if ks.unlocked[addr] == u {
			ZeroKey(u.PrivateKey)
			delete(ks.unlocked, addr)
		}
		ks.mu.Unlock()
	}
}

// NewAccount generates a new key and stores it into the key directory,
// encrypting it with the passphrase.
func (ks *KeyStore) NewAccount(passphrase string) (accounts.Account, error) {
	_, account, err := storeNewKey(ks.storage, crand.Reader, passphrase)
	if err != nil {
		return accounts.Account{}, err
	}
	// Add the account to the cache immediately rather
	// than waiting for file system notifications to pick it up.
	ks.cache.add(account)
	ks.refreshWallets()
	return account, nil
}

// Export exports as a JSON key, encrypted with newPassphrase.
func (ks *KeyStore) Export(a accounts.Account, passphrase, newPassphrase string) (keyJSON []byte, err error) {
	_, key, err := ks.getDecryptedKey(a, passphrase)
	if err != nil {
		return nil, err
	}
	var N, P int
	if store, ok := ks.storage.(*keyStorePassphrase); ok {
		N, P = store.scryptN, store.scryptP
	} else {
		N, P = StandardScryptN, StandardScryptP
	}
	return EncryptKey(key, newPassphrase, N, P)
}

func (ks *KeyStore) DumpKey(a accounts.Account, passphrase string) (privateKey []byte, err error) {
	_, key, err := ks.getDecryptedKey(a, passphrase)
	if err != nil {
		return nil, err
	}
	return key.PrivateKey, nil

}
func (ks *KeyStore) DumpPrivateKey(a accounts.Account, passphrase string) (privateKey interface{}, err error) {
	_, key, err := ks.getDecryptedKey(a, passphrase)
	if err != nil {
		return nil, err
	}

	return crypto.MyCryptoLib.PrivateKeyToInstance(key.PrivateKey)

}

// Import stores the given encrypted JSON key into the key directory.
func (ks *KeyStore) Import(keyJSON []byte, passphrase, newPassphrase string) (accounts.Account, error) {
	key, err := DecryptKey(keyJSON, passphrase)
	if key != nil && key.PrivateKey != nil {
		defer ZeroKey(key.PrivateKey)
	}
	if err != nil {
		return accounts.Account{}, err
	}
	return ks.importKey(key, newPassphrase)
}

// Import stores the given encrypted JSON key into the key directory.
func (ks *KeyStore) ImportFromHex(hexhash string, newPassphrase string) (accounts.Account, error) {
	priv, err := hex.DecodeString(hexhash)
	if err != nil {
		return accounts.Account{}, errors.New("invalid hex string")
	}
	//priv, err := crypto.ToECDSA(b)
	key := newKeyFromECDSA(priv)
	if key != nil && key.PrivateKey != nil {
		defer ZeroKey(key.PrivateKey)
	}
	return ks.importKey(key, newPassphrase)
}

// ImportECDSA stores the given key into the key directory, encrypting it with the passphrase.
func (ks *KeyStore) ImportECDSA(priv []byte, passphrase string) (accounts.Account, error) {

	key := newKeyFromECDSA(priv)
	if ks.cache.hasAddress(key.Address) {
		return accounts.Account{}, fmt.Errorf("account already exists")
	}
	return ks.importKey(key, passphrase)
}

func (ks *KeyStore) importKey(key *Key, passphrase string) (accounts.Account, error) {
	a := accounts.Account{Address: key.Address, URL: accounts.URL{Scheme: KeyStoreScheme,
		Path: ks.storage.JoinPath(keyFileName(key.Address))}}
	if err := ks.storage.StoreKey(a.URL.Path, key, passphrase); err != nil {
		return accounts.Account{}, err
	}
	ks.cache.add(a)
	ks.refreshWallets()
	return a, nil
}

// Update changes the passphrase of an existing account.
func (ks *KeyStore) Update(a accounts.Account, passphrase, newPassphrase string) error {
	a, key, err := ks.getDecryptedKey(a, passphrase)
	if err != nil {
		return err
	}
	return ks.storage.StoreKey(a.URL.Path, key, newPassphrase)
}

// ZeroKey zeroes a private key in memory.
func ZeroKey(k []byte) {
	//b := k.D.Bits()
	for idx := range k {
		k[idx] = 0
	}
}

func (ks *KeyStore) getPrivateKey(address common.Address) ([]byte, error) {
	// Look up the key to sign with and abort if it cannot be found
	ks.mu.RLock()
	defer ks.mu.RUnlock()
	unlockedKey, found := ks.unlocked[address]
	if !found {
		return nil, ErrLocked
	}
	return unlockedKey.PrivateKey, nil
}

func (ks *KeyStore) GetPublicKey(address common.Address) ([]byte, error) {
	// Look up the key to sign with and abort if it cannot be found
	ks.mu.RLock()
	defer ks.mu.RUnlock()
	unlockedKey, found := ks.unlocked[address]
	if !found {
		return nil, ErrLocked
	}
	return crypto.MyCryptoLib.PrivateKeyToPubKey(unlockedKey.PrivateKey)
	//return crypto.CompressPubkey(&unlockedKey.PrivateKey.PublicKey), nil
}

func (ks *KeyStore) SigUnit(unitHeader *modules.Header, address common.Address) ([]byte, error) {
	emptyHeader := modules.CopyHeader(unitHeader)
	emptyHeader.Authors = modules.Authentifier{} //Clear exist sign
	emptyHeader.GroupSign = make([]byte, 0)      //Clear group sign
	return ks.SigData(emptyHeader, address)
}

func (ks *KeyStore) SigData(data interface{}, address common.Address) ([]byte, error) {
	privateKey, err := ks.getPrivateKey(address)
	if err != nil {
		return nil, err
	}
	//defer ZeroKey(privateKey)
	msg, err := rlp.EncodeToBytes(data)
	//hash := util.RlpHash(data) //crypto.Keccak256Hash(util.RHashBytes(data))
	if err != nil {
		return nil, err
	}
	sign, err := crypto.MyCryptoLib.Sign(privateKey, msg)
	//sign, err := crypto.Sign(hash.Bytes(), privateKey)
	if err != nil {
		return nil, err
	}
	//log.Debugf("Try to sign data:%x,sign result:%x", msg, sign)

	return sign, nil
}

func (ks *KeyStore) SigUnitWithPwd(unit interface{}, privateKey []byte) ([]byte, error) {
	//hash := util.RlpHash(unit) //crypto.Keccak256Hash(util.RHashBytes(unit))
	msg, err := rlp.EncodeToBytes(unit)
	//hash := util.RlpHash(data) //crypto.Keccak256Hash(util.RHashBytes(data))
	if err != nil {
		return nil, err
	}
	//unit signature
	sign, err := crypto.MyCryptoLib.Sign(privateKey, msg)
	if err != nil {
		return nil, err
	}
	return sign, nil
}

func VerifyUnitWithPK(sign []byte, unit interface{}, publicKey []byte) bool {
	msg, err := rlp.EncodeToBytes(unit)
	//hash := util.RlpHash(data) //crypto.Keccak256Hash(util.RHashBytes(data))
	if err != nil {
		return false
	}
	if len(sign) <= 0 {
		log.Error("Unit sigature is none.")
		return false
	}
	sig := sign[:len(sign)-1] // remove recovery id
	pass, _ := crypto.MyCryptoLib.Verify(publicKey, sig, msg)
	return pass
}

//tx:TxMessages   []Message
func (ks *KeyStore) SigTX(tx interface{}, address common.Address) (R, S, V []byte, e error) {
	sig, err := ks.SigData(tx, address)
	if err != nil {
		e = err
		return
	}

	R = sig[:32]
	S = sig[32:64]
	V = append(V, sig[64])
	e = nil
	return
}

func VerifyTXWithPK(sign []byte, tx interface{}, publicKey []byte) bool {
	return VerifyUnitWithPK(sign, tx, publicKey)
}

func (ks *KeyStore) SigTXWithPwd(tx interface{}, privateKey []byte) ([]byte, error) {
	return ks.SigUnitWithPwd(tx, privateKey)
}
