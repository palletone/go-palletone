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
	"encoding/json"
	//"fmt"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/hexutil"
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
func (api *PublicPalletOneAPI) Etherbase() (common.Address, error) {
	return api.p.Etherbase()
}

// Coinbase is the address that mining rewards will be send to (alias for Etherbase)
func (api *PublicPalletOneAPI) Coinbase() (common.Address, error) {
	return api.Etherbase()
}

// Hashrate returns the POW hashrate
func (api *PublicPalletOneAPI) Hashrate() hexutil.Uint64 {
	//return hexutil.Uint64(api.p.Miner().HashRate())
	return hexutil.Uint64(0)
}

type PublicDagAPI struct {
	p *PalletOne
}

func NewPublicDagAPI(p *PalletOne) *PublicDagAPI {
	return &PublicDagAPI{p}
}
func (api *PublicDagAPI) TokenInfos() (string, error) {
	all, err := api.p.dag.GetAllTokenInfo()
	if err != nil {
		return "get failed.", err
	}
	if bytes, err := json.Marshal(all); err != nil {
		return "error", err
	} else {
		return string(bytes), nil
	}
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
	ptn *PalletOne
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

//func (api *PublicPalletOneAPI) TransferPtn(from, to, amount, text, password string) (*mp.TxExecuteResult, error) {
//	// 参数检查
//	fromAdd, err := common.StringToAddress(from)
//	if err != nil {
//		return nil, fmt.Errorf("invalid account address: %s", from)
//	}
//
//	toAdd, err := common.StringToAddress(to)
//	if err != nil {
//		return nil, fmt.Errorf("invalid account address: %s", to)
//	}
//
//	amountPtn, err := decimal.RandFromString(amount)
//	if err != nil {
//		return nil, fmt.Errorf("invalid PTN amount: %s", amount)
//	}
//
//	// 判断本节点是否同步完成，数据是否最新
//	if !api.p.dag.IsSynced() {
//		return nil, fmt.Errorf("the data of this node is not synced, and can't transfer now")
//	}
//
//	// 1. 创建交易
//	tx, fee, err := api.p.dag.GenTransferPtnTx(fromAdd, toAdd, amountPtn, text)
//	if err != nil {
//		return nil, err
//	}
//
//	// 2. 签名和发送交易
//	err = api.p.SignAndSendTransaction(fromAdd, tx)
//	if err != nil {
//		return nil, err
//	}
//
//	// 5. 返回执行结果
//	res := &mp.TxExecuteResult{}
//	res.TxContent = fmt.Sprintf("Account %s transfer %vPTN to account %s with message: %v",
//		from, amount, to, text)
//	res.TxHash = tx.Hash()
//	res.TxSize = tx.Size().TerminalString()
//	res.TxFee = fmt.Sprintf("%vdao", fee)
//	res.Warning = mp.DefaultResult
//
//	return res, nil
//}
