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

package ptn

import (
	"context"
	//"fmt"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/hexutil"
	"github.com/palletone/go-palletone/configure"
)

// PublicEthereumAPI provides an API to access PalletOne full node-related
// information.
type PublicEthereumAPI struct {
	e *PalletOne
}

// NewPublicEthereumAPI creates a new PalletOne protocol API for full nodes.
func NewPublicEthereumAPI(e *PalletOne) *PublicEthereumAPI {
	return &PublicEthereumAPI{e}
}

// Etherbase is the address that mining rewards will be send to
func (api *PublicEthereumAPI) Etherbase() (common.Address, error) {
	return api.e.Etherbase()
}

// Coinbase is the address that mining rewards will be send to (alias for Etherbase)
func (api *PublicEthereumAPI) Coinbase() (common.Address, error) {
	return api.Etherbase()
}

// Hashrate returns the POW hashrate
func (api *PublicEthereumAPI) Hashrate() hexutil.Uint64 {
	//return hexutil.Uint64(api.e.Miner().HashRate())
	return hexutil.Uint64(0)
}

/*
// PublicDebugAPI is the collection of PalletOne full node APIs exposed
// over the public debugging endpoint.
type PublicDebugAPI struct {
	eth *PalletOne
}

// NewPublicDebugAPI creates a new API definition for the full node-
// related public debug methods of the PalletOne service.
func NewPublicDebugAPI(eth *PalletOne) *PublicDebugAPI {
	return &PublicDebugAPI{eth: eth}
}
*/
// PrivateDebugAPI is the collection of PalletOne full node APIs exposed over
// the private debugging endpoint.
type PrivateDebugAPI struct {
	config *configure.ChainConfig
	eth    *PalletOne
}

// NewPrivateDebugAPI creates a new API definition for the full node-related
// private debug methods of the PalletOne service.
func NewPrivateDebugAPI(config *configure.ChainConfig, eth *PalletOne) *PrivateDebugAPI {
	return &PrivateDebugAPI{config: config, eth: eth}
}

// Preimage is a debug API function that returns the preimage for a sha3 hash, if known.
func (api *PrivateDebugAPI) Preimage(ctx context.Context, hash common.Hash) (hexutil.Bytes, error) {
	return hexutil.Bytes{}, nil
}

// StorageRangeResult is the result of a debug_storageRangeAt API call.
type StorageRangeResult struct {
	Storage storageMap   `json:"storage"`
	NextKey *common.Hash `json:"nextKey"` // nil if Storage includes the last key in the trie.
}

type storageMap map[common.Hash]storageEntry

type storageEntry struct {
	Key   *common.Hash `json:"key"`
	Value common.Hash  `json:"value"`
}
