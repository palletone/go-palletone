// Copyright 2017 The go-palletone Authors
// This file is part of the go-palletone library.
//
// The go-palletone library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-palletone library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-palletone library. If not, see <http://www.gnu.org/licenses/>.

package downloader

import (
	"math/big"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/dag/coredata"
	//"github.com/palletone/go-palletone/contracts/types"
	"github.com/palletone/go-palletone/common/ptndb"
)

// FakePeer is a mock downloader peer that operates on a local database instance
// instead of being an actual live node. It's useful for testing and to implement
// sync commands from an existing local database.
type FakePeer struct{}

func NewFakePeer(id string, db ptndb.Database, hc *coredata.HeaderChain, dl *Downloader) *FakePeer {
	return &FakePeer{}
}
func (p *FakePeer) Head() (common.Hash, *big.Int) {
	return common.Hash{}, &big.Int{}
}
func (p *FakePeer) RequestHeadersByHash(hash common.Hash, amount int, skip int, reverse bool) error {
	return nil
}

// RequestHeadersByNumber implements downloader.Peer, returning a batch of headers
// defined by the origin number and the associated query parameters.
func (p *FakePeer) RequestHeadersByNumber(number uint64, amount int, skip int, reverse bool) error {
	return nil
}

// RequestBodies implements downloader.Peer, returning a batch of block bodies
// corresponding to the specified block hashes.
func (p *FakePeer) RequestBodies(hashes []common.Hash) error {
	return nil
}

// RequestReceipts implements downloader.Peer, returning a batch of transaction
// receipts corresponding to the specified block hashes.
func (p *FakePeer) RequestReceipts(hashes []common.Hash) error {
	return nil
}

// RequestNodeData implements downloader.Peer, returning a batch of state trie
// nodes corresponding to the specified trie hashes.
func (p *FakePeer) RequestNodeData(hashes []common.Hash) error {
	return nil
}
