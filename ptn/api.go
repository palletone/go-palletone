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

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/hexutil"
	// mp "github.com/palletone/go-palletone/consensus/mediatorplugin"
	"github.com/palletone/go-palletone/dag/errors"
	"github.com/shopspring/decimal"
)

// PublicPalletOneAPI provides an API to access PalletOne full node-related
// information.
type PublicPalletOneAPI struct {
	p *PalletOne
}

// NewPublicPalletOneAPI creates a new PalletOne protocol API for full nodes.
func NewPublicPalletOneAPI(p *PalletOne) *PublicPalletOneAPI {
	return &PublicPalletOneAPI{p}
}

// Etherbase is the address that mining rewards will be send to
//func (api *PublicPalletOneAPI) Etherbase() (common.Address, error) {
//	return api.p.Etherbase()
//}

// Coinbase is the address that mining rewards will be send to (alias for Etherbase)
func (api *PublicPalletOneAPI) Coinbase() (common.Address, error) {
	//return api.Etherbase()
	err := errors.New("TODO read mediator address from toml config")
	return common.Address{}, err
}

// Hashrate returns the POW hashrate
func (api *PublicPalletOneAPI) Hashrate() hexutil.Uint64 {
	//return hexutil.Uint64(api.p.Miner().HashRate())
	return hexutil.Uint64(0)
}

/*
// PublicDebugAPI is the collection of PalletOne full node APIs exposed
// over the public debugging endpoint.
type PublicDebugAPI struct {
	ptn *PalletOne
}

// NewPublicDebugAPI creates a new API definition for the full node-
// related public debug methods of the PalletOne service.
func NewPublicDebugAPI(ptn *PalletOne) *PublicDebugAPI {
	return &PublicDebugAPI{ptn: ptn}
}
*/
// PrivateDebugAPI is the collection of PalletOne full node APIs exposed over
// the private debugging endpoint.
type PrivateDebugAPI struct {
	//config *configure.ChainConfig
	//ptn *PalletOne
}

// NewPrivateDebugAPI creates a new API definition for the full node-related
// private debug methods of the PalletOne service.
//func NewPrivateDebugAPI(config *configure.ChainConfig, ptn *PalletOne) *PrivateDebugAPI {
//	return &PrivateDebugAPI{config: config, ptn: ptn}
//}

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

type TransferPtnArgs struct {
	From   string          `json:"from"`
	To     string          `json:"to"`
	Amount decimal.Decimal `json:"amount"`
	Text   *string         `json:"text"`
}

// func (api *PublicPalletOneAPI) TransferPtn(args TransferPtnArgs) (*ptnapi.TxExecuteResult, error) {
// 	return api.p.TransferPtn(args.From, args.To, args.Amount, args.Text)
// }
