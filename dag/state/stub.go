// Copyright 2015 The go-ethereum Authors
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

package state

import (
	"math/big"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/trie"
)

type StateDB struct{}

type ManagedState struct{}

func (self *StateDB) GetBalance(addr common.Address) *big.Int       { return &big.Int{} }
func (self *StateDB) GetNonce(addr common.Address) uint64           { return uint64(0) }
func (ms *ManagedState) SetNonce(addr common.Address, nonce uint64) {}
func (ms *ManagedState) GetNonce(addr common.Address) uint64        { return uint64(0) }
func NewStateSync(root common.Hash) *trie.TrieSync {
	return &trie.TrieSync{}
}

type Trie interface {
	NodeIterator(startKey []byte) trie.NodeIterator
	GetKey([]byte) []byte
}
type Dump struct{}

func NodeIterator(startKey []byte) trie.NodeIterator { return nil }
func GetKey([]byte) []byte                           { return []byte{} }

var MaxTrieCacheGen = uint16(120)
