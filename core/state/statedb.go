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

// Package state provides a caching layer atop the Ethereum state trie.
package state

import (
	"fmt"
	"math/big"
	"sort"
	"sync"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/crypto"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/trie"
	"github.com/palletone/go-palletone/core/types"
)

type revision struct {
	id           int
	journalIndex int
}

var (
	// emptyState is the known hash of an empty state trie entry.
	emptyState = crypto.Keccak256Hash(nil)

	// emptyCode is the known hash of the empty EVM bytecode.
	emptyCode = crypto.Keccak256Hash(nil)
)

// StateDBs within the ethereum protocol are used to store anything
// within the merkle trie. StateDBs take care of caching and storing
// nested states. It's the general query interface to retrieve:
// * Contracts
// * Accounts
type StateDB struct {
	db   Database
	trie Trie

	// This map holds 'live' objects, which will get modified while processing a state transition.
	stateObjects      map[common.Address]*stateObject
	stateObjectsDirty map[common.Address]struct{}

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error

	// The refund counter, also used by state transitioning.
	refund uint64

	thash, bhash common.Hash
	txIndex      int
	logs         map[common.Hash][]*types.Log
	logSize      uint

	preimages map[common.Hash][]byte

	// Journal of state modifications. This is the backbone of
	// Snapshot and RevertToSnapshot.
	journal        *journal
	validRevisions []revision
	nextRevisionId int

	lock sync.Mutex
}

// Create a new state from a given trie.
func New(root common.Hash, db Database) (*StateDB, error) {
	tr, err := db.OpenTrie(root)
	if err != nil {
		return nil, err
	}
	return &StateDB{
		db:                db,
		trie:              tr,
		stateObjects:      make(map[common.Address]*stateObject),
		stateObjectsDirty: make(map[common.Address]struct{}),
		logs:              make(map[common.Hash][]*types.Log),
		preimages:         make(map[common.Hash][]byte),
		journal:           newJournal(),
	}, nil
}

// setError remembers the first non-nil error it is called with.
func (statedb *StateDB) setError(err error) {
	if statedb.dbErr == nil {
		statedb.dbErr = err
	}
}

func (statedb *StateDB) Error() error {
	return statedb.dbErr
}

// Reset clears out all ephemeral state objects from the state db, but keeps
// the underlying state trie to avoid reloading data for the next operations.
func (statedb *StateDB) Reset(root common.Hash) error {
	tr, err := statedb.db.OpenTrie(root)
	if err != nil {
		return err
	}
	statedb.trie = tr
	statedb.stateObjects = make(map[common.Address]*stateObject)
	statedb.stateObjectsDirty = make(map[common.Address]struct{})
	statedb.thash = common.Hash{}
	statedb.bhash = common.Hash{}
	statedb.txIndex = 0
	statedb.logs = make(map[common.Hash][]*types.Log)
	statedb.logSize = 0
	statedb.preimages = make(map[common.Hash][]byte)
	statedb.clearJournalAndRefund()
	return nil
}

func (statedb *StateDB) AddLog(log *types.Log) {
	statedb.journal.append(addLogChange{txhash: statedb.thash})

	log.TxHash = statedb.thash
	log.BlockHash = statedb.bhash
	log.TxIndex = uint(statedb.txIndex)
	log.Index = statedb.logSize
	statedb.logs[statedb.thash] = append(statedb.logs[statedb.thash], log)
	statedb.logSize++
}

func (statedb *StateDB) GetLogs(hash common.Hash) []*types.Log {
	return statedb.logs[hash]
}

func (statedb *StateDB) Logs() []*types.Log {
	logs := make([]*types.Log, 0, len(statedb.logs))
	for _, lgs := range statedb.logs {
		logs = append(logs, lgs...)
	}
	return logs
}

// AddPreimage records a SHA3 preimage seen by the VM.
func (statedb *StateDB) AddPreimage(hash common.Hash, preimage []byte) {
	if _, ok := statedb.preimages[hash]; !ok {
		statedb.journal.append(addPreimageChange{hash: hash})
		pi := make([]byte, len(preimage))
		copy(pi, preimage)
		statedb.preimages[hash] = pi
	}
}

// Preimages returns a list of SHA3 preimages that have been submitted.
func (statedb *StateDB) Preimages() map[common.Hash][]byte {
	return statedb.preimages
}

func (statedb *StateDB) AddRefund(gas uint64) {
	statedb.journal.append(refundChange{prev: statedb.refund})
	statedb.refund += gas
}

// Exist reports whether the given account address exists in the state.
// Notably this also returns true for suicided accounts.
func (statedb *StateDB) Exist(addr common.Address) bool {
	return statedb.getStateObject(addr) != nil
}

// Empty returns whether the state object is either non-existent
// or empty according to the EIP161 specification (balance = nonce = code = 0)
func (statedb *StateDB) Empty(addr common.Address) bool {
	so := statedb.getStateObject(addr)
	return so == nil || so.empty()
}

// Retrieve the balance from the given address or 0 if object not found
func (statedb *StateDB) GetBalance(addr common.Address) *big.Int {
	stateObject := statedb.getStateObject(addr)
	if stateObject != nil {
		return stateObject.Balance()
	}
	return common.Big0
}

func (statedb *StateDB) GetNonce(addr common.Address) uint64 {
	stateObject := statedb.getStateObject(addr)
	if stateObject != nil {
		return stateObject.Nonce()
	}

	return 0
}

func (statedb *StateDB) GetCode(addr common.Address) []byte {
	stateObject := statedb.getStateObject(addr)
	if stateObject != nil {
		return stateObject.Code(statedb.db)
	}
	return nil
}

func (statedb *StateDB) GetCodeSize(addr common.Address) int {
	stateObject := statedb.getStateObject(addr)
	if stateObject == nil {
		return 0
	}
	if stateObject.code != nil {
		return len(stateObject.code)
	}
	size, err := statedb.db.ContractCodeSize(stateObject.addrHash, common.BytesToHash(stateObject.CodeHash()))
	if err != nil {
		statedb.setError(err)
	}
	return size
}

func (statedb *StateDB) GetCodeHash(addr common.Address) common.Hash {
	stateObject := statedb.getStateObject(addr)
	if stateObject == nil {
		return common.Hash{}
	}
	return common.BytesToHash(stateObject.CodeHash())
}

func (statedb *StateDB) GetState(addr common.Address, bhash common.Hash) common.Hash {
	stateObject := statedb.getStateObject(addr)
	if stateObject != nil {
		return stateObject.GetState(statedb.db, bhash)
	}
	return common.Hash{}
}

// Database retrieves the low level database supporting the lower level trie ops.
func (statedb *StateDB) Database() Database {
	return statedb.db
}

// StorageTrie returns the storage trie of an account.
// The return value is a copy and is nil for non-existent accounts.
func (statedb *StateDB) StorageTrie(addr common.Address) Trie {
	stateObject := statedb.getStateObject(addr)
	if stateObject == nil {
		return nil
	}
	cpy := stateObject.deepCopy(statedb)
	return cpy.updateTrie(statedb.db)
}

func (statedb *StateDB) HasSuicided(addr common.Address) bool {
	stateObject := statedb.getStateObject(addr)
	if stateObject != nil {
		return stateObject.suicided
	}
	return false
}

/*
 * SETTERS
 */

// AddBalance adds amount to the account associated with addr.
func (statedb *StateDB) AddBalance(addr common.Address, amount *big.Int) {
	stateObject := statedb.GetOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.AddBalance(amount)
	}
}

// SubBalance subtracts amount from the account associated with addr.
func (statedb *StateDB) SubBalance(addr common.Address, amount *big.Int) {
	stateObject := statedb.GetOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.SubBalance(amount)
	}
}

func (statedb *StateDB) SetBalance(addr common.Address, amount *big.Int) {
	stateObject := statedb.GetOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.SetBalance(amount)
	}
}

func (statedb *StateDB) SetNonce(addr common.Address, nonce uint64) {
	stateObject := statedb.GetOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.SetNonce(nonce)
	}
}

func (statedb *StateDB) SetCode(addr common.Address, code []byte) {
	stateObject := statedb.GetOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.SetCode(crypto.Keccak256Hash(code), code)
	}
}

func (statedb *StateDB) SetState(addr common.Address, key, value common.Hash) {
	stateObject := statedb.GetOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.SetState(statedb.db, key, value)
	}
}

// Suicide marks the given account as suicided.
// This clears the account balance.
//
// The account's state object is still available until the state is committed,
// getStateObject will return a non-nil account after Suicide.
func (statedb *StateDB) Suicide(addr common.Address) bool {
	stateObject := statedb.getStateObject(addr)
	if stateObject == nil {
		return false
	}
	statedb.journal.append(suicideChange{
		account:     &addr,
		prev:        stateObject.suicided,
		prevbalance: new(big.Int).Set(stateObject.Balance()),
	})
	stateObject.markSuicided()
	stateObject.data.Balance = new(big.Int)

	return true
}

//
// Setting, updating & deleting state object methods.
//

// updateStateObject writes the given object to the trie.
func (statedb *StateDB) updateStateObject(stateObject *stateObject) {
	addr := stateObject.Address()
	data, err := rlp.EncodeToBytes(stateObject)
	if err != nil {
		panic(fmt.Errorf("can't encode object at %x: %v", addr[:], err))
	}
	statedb.setError(statedb.trie.TryUpdate(addr[:], data))
}

// deleteStateObject removes the given object from the state trie.
func (statedb *StateDB) deleteStateObject(stateObject *stateObject) {
	stateObject.deleted = true
	addr := stateObject.Address()
	statedb.setError(statedb.trie.TryDelete(addr[:]))
}

// Retrieve a state object given my the address. Returns nil if not found.
func (statedb *StateDB) getStateObject(addr common.Address) (stateObject *stateObject) {
	// Prefer 'live' objects.
	if obj := statedb.stateObjects[addr]; obj != nil {
		if obj.deleted {
			return nil
		}
		return obj
	}

	// Load the object from the database.
	enc, err := statedb.trie.TryGet(addr[:])
	if len(enc) == 0 {
		statedb.setError(err)
		return nil
	}
	var data Account
	if err := rlp.DecodeBytes(enc, &data); err != nil {
		log.Error("Failed to decode state object", "addr", addr, "err", err)
		return nil
	}
	// Insert into the live set.
	obj := newObject(statedb, addr, data)
	statedb.setStateObject(obj)
	return obj
}

func (statedb *StateDB) setStateObject(object *stateObject) {
	statedb.stateObjects[object.Address()] = object
}

// Retrieve a state object or create a new state object if nil.
func (statedb *StateDB) GetOrNewStateObject(addr common.Address) *stateObject {
	stateObject := statedb.getStateObject(addr)
	if stateObject == nil || stateObject.deleted {
		stateObject, _ = statedb.createObject(addr)
	}
	return stateObject
}

// createObject creates a new state object. If there is an existing account with
// the given address, it is overwritten and returned as the second return value.
func (statedb *StateDB) createObject(addr common.Address) (newobj, prev *stateObject) {
	prev = statedb.getStateObject(addr)
	newobj = newObject(statedb, addr, Account{})
	newobj.setNonce(0) // sets the object to dirty
	if prev == nil {
		statedb.journal.append(createObjectChange{account: &addr})
	} else {
		statedb.journal.append(resetObjectChange{prev: prev})
	}
	statedb.setStateObject(newobj)
	return newobj, prev
}

// CreateAccount explicitly creates a state object. If a state object with the address
// already exists the balance is carried over to the new account.
//
// CreateAccount is called during the EVM CREATE operation. The situation might arise that
// a contract does the following:
//
//   1. sends funds to sha(account ++ (nonce + 1))
//   2. tx_create(sha(account ++ nonce)) (note that this gets the address of 1)
//
// Carrying over the balance ensures that Ether doesn't disappear.
func (statedb *StateDB) CreateAccount(addr common.Address) {
	new, prev := statedb.createObject(addr)
	if prev != nil {
		new.setBalance(prev.data.Balance)
	}
}

func (db *StateDB) ForEachStorage(addr common.Address, cb func(key, value common.Hash) bool) {
	so := db.getStateObject(addr)
	if so == nil {
		return
	}

	// When iterating over the storage check the cache first
	for h, value := range so.cachedStorage {
		cb(h, value)
	}

	it := trie.NewIterator(so.getTrie(db.db).NodeIterator(nil))
	for it.Next() {
		// ignore cached values
		key := common.BytesToHash(db.trie.GetKey(it.Key))
		if _, ok := so.cachedStorage[key]; !ok {
			cb(key, common.BytesToHash(it.Value))
		}
	}
}

// Copy creates a deep, independent copy of the state.
// Snapshots of the copied state cannot be applied to the copy.
func (statedb *StateDB) Copy() *StateDB {
	statedb.lock.Lock()
	defer statedb.lock.Unlock()

	// Copy all the basic fields, initialize the memory ones
	state := &StateDB{
		db:                statedb.db,
		trie:              statedb.db.CopyTrie(statedb.trie),
		stateObjects:      make(map[common.Address]*stateObject, len(statedb.journal.dirties)),
		stateObjectsDirty: make(map[common.Address]struct{}, len(statedb.journal.dirties)),
		refund:            statedb.refund,
		logs:              make(map[common.Hash][]*types.Log, len(statedb.logs)),
		logSize:           statedb.logSize,
		preimages:         make(map[common.Hash][]byte),
		journal:           newJournal(),
	}
	// Copy the dirty states, logs, and preimages
	for addr := range statedb.journal.dirties {
		// As documented [here](https://github.com/ethereum/go-ethereum/pull/16485#issuecomment-380438527),
		// and in the Finalize-method, there is a case where an object is in the journal but not
		// in the stateObjects: OOG after touch on ripeMD prior to Byzantium. Thus, we need to check for
		// nil
		if object, exist := statedb.stateObjects[addr]; exist {
			state.stateObjects[addr] = object.deepCopy(state)
			state.stateObjectsDirty[addr] = struct{}{}
		}
	}
	// Above, we don't copy the actual journal. This means that if the copy is copied, the
	// loop above will be a no-op, since the copy's journal is empty.
	// Thus, here we iterate over stateObjects, to enable copies of copies
	for addr := range statedb.stateObjectsDirty {
		if _, exist := state.stateObjects[addr]; !exist {
			state.stateObjects[addr] = statedb.stateObjects[addr].deepCopy(state)
			state.stateObjectsDirty[addr] = struct{}{}
		}
	}

	for hash, logs := range statedb.logs {
		state.logs[hash] = make([]*types.Log, len(logs))
		copy(state.logs[hash], logs)
	}
	for hash, preimage := range statedb.preimages {
		state.preimages[hash] = preimage
	}
	return state
}

// Snapshot returns an identifier for the current revision of the state.
func (statedb *StateDB) Snapshot() int {
	id := statedb.nextRevisionId
	statedb.nextRevisionId++
	statedb.validRevisions = append(statedb.validRevisions, revision{id, statedb.journal.length()})
	return id
}

// RevertToSnapshot reverts all state changes made since the given revision.
func (statedb *StateDB) RevertToSnapshot(revid int) {
	// Find the snapshot in the stack of valid snapshots.
	idx := sort.Search(len(statedb.validRevisions), func(i int) bool {
		return statedb.validRevisions[i].id >= revid
	})
	if idx == len(statedb.validRevisions) || statedb.validRevisions[idx].id != revid {
		panic(fmt.Errorf("revision id %v cannot be reverted", revid))
	}
	snapshot := statedb.validRevisions[idx].journalIndex

	// Replay the journal to undo changes and remove invalidated snapshots
	statedb.journal.revert(statedb, snapshot)
	statedb.validRevisions = statedb.validRevisions[:idx]
}

// GetRefund returns the current value of the refund counter.
func (statedb *StateDB) GetRefund() uint64 {
	return statedb.refund
}

// Finalize finalizes the state by removing the self destructed objects
// and clears the journal as well as the refunds.
func (s *StateDB) Finalize(deleteEmptyObjects bool) {
	for addr := range s.journal.dirties {
		stateObject, exist := s.stateObjects[addr]
		if !exist {
			// ripeMD is 'touched' at block 1714175, in tx 0x1237f737031e40bcde4a8b7e717b2d15e3ecadfe49bb1bbc71ee9deb09c6fcf2
			// That tx goes out of gas, and although the notion of 'touched' does not exist there, the
			// touch-event will still be recorded in the journal. Since ripeMD is a special snowflake,
			// it will persist in the journal even though the journal is reverted. In this special circumstance,
			// it may exist in `s.journal.dirties` but not in `s.stateObjects`.
			// Thus, we can safely ignore it here
			continue
		}

		if stateObject.suicided || (deleteEmptyObjects && stateObject.empty()) {
			s.deleteStateObject(stateObject)
		} else {
			stateObject.updateRoot(s.db)
			s.updateStateObject(stateObject)
		}
		s.stateObjectsDirty[addr] = struct{}{}
	}
	// Invalidate journal because reverting across transactions is not allowed.
	s.clearJournalAndRefund()
}

// IntermediateRoot computes the current root hash of the state trie.
// It is called in between transactions to get the root hash that
// goes into transaction receipts.
func (s *StateDB) IntermediateRoot(deleteEmptyObjects bool) common.Hash {
	s.Finalize(deleteEmptyObjects)
	return s.trie.Hash()
}

// Prepare sets the current transaction hash and index and block hash which is
// used when the EVM emits new state logs.
func (statedb *StateDB) Prepare(thash, bhash common.Hash, ti int) {
	statedb.thash = thash
	statedb.bhash = bhash
	statedb.txIndex = ti
}

func (s *StateDB) clearJournalAndRefund() {
	s.journal = newJournal()
	s.validRevisions = s.validRevisions[:0]
	s.refund = 0
}

// Commit writes the state to the underlying in-memory trie database.
func (s *StateDB) Commit(deleteEmptyObjects bool) (root common.Hash, err error) {
	defer s.clearJournalAndRefund()

	for addr := range s.journal.dirties {
		s.stateObjectsDirty[addr] = struct{}{}
	}
	// Commit objects to the trie.
	for addr, stateObject := range s.stateObjects {
		_, isDirty := s.stateObjectsDirty[addr]
		switch {
		case stateObject.suicided || (isDirty && deleteEmptyObjects && stateObject.empty()):
			// If the object has been removed, don't bother syncing it
			// and just mark it for deletion in the trie.
			s.deleteStateObject(stateObject)
		case isDirty:
			// Write any contract code associated with the state object
			if stateObject.code != nil && stateObject.dirtyCode {
				s.db.TrieDB().Insert(common.BytesToHash(stateObject.CodeHash()), stateObject.code)
				stateObject.dirtyCode = false
			}
			// Write any storage changes in the state object to its storage trie.
			if err := stateObject.CommitTrie(s.db); err != nil {
				return common.Hash{}, err
			}
			// Update the object in the main account trie.
			s.updateStateObject(stateObject)
		}
		delete(s.stateObjectsDirty, addr)
	}
	// Write trie changes.
	root, err = s.trie.Commit(func(leaf []byte, parent common.Hash) error {
		var account Account
		if err := rlp.DecodeBytes(leaf, &account); err != nil {
			return nil
		}
		if account.Root != emptyState {
			s.db.TrieDB().Reference(account.Root, parent)
		}
		code := common.BytesToHash(account.CodeHash)
		if code != emptyCode {
			s.db.TrieDB().Reference(code, parent)
		}
		return nil
	})
	log.Debug("Trie cache stats after commit", "misses", trie.CacheMisses(), "unloads", trie.CacheUnloads())
	return root, err
}
