/*
 *
 *    This file is part of go-palletone.
 *    go-palletone is free software: you can redistribute it and/or modify
 *    it under the terms of the GNU General Public License as published by
 *    the Free Software Foundation, either version 3 of the License, or
 *    (at your option) any later version.
 *    go-palletone is distributed in the hope that it will be useful,
 *    but WITHOUT ANY WARRANTY; without even the implied warranty of
 *    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *    GNU General Public License for more details.
 *    You should have received a copy of the GNU General Public License
 *    along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
 * /
 *
 *  * @author PalletOne core developer <dev@pallet.one>
 *  * @date 2018
 *
 */

package dag

import (
	"github.com/dedis/kyber"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/event"
	"github.com/palletone/go-palletone/common/p2p/discover"
	"github.com/palletone/go-palletone/core/accounts/keystore"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/txspool"
	"time"
)

type IDag interface {
	CurrentUnit() *modules.Unit
	SaveDag(unit modules.Unit, isGenesis bool) (int, error)
	GetActiveMediatorNodes() []*discover.Node
	VerifyHeader(header *modules.Header, seal bool) error
	GetCurrentUnit(assetId modules.IDType16) *modules.Unit
	InsertDag(units modules.Units) (int, error)
	GetUnitByHash(hash common.Hash) *modules.Unit
	HasHeader(common.Hash, uint64) bool
	GetHeaderByNumber(number modules.ChainIndex) *modules.Header
	// GetHeaderByHash retrieves a header from the local chain.
	GetHeaderByHash(common.Hash) *modules.Header
	GetHeader(hash common.Hash, number uint64) (*modules.Header, error)
	// CurrentHeader retrieves the head header from the local chain.
	CurrentHeader() *modules.Header
	GetTransactionByHash(hash common.Hash) (*modules.Transaction, error)
	// InsertHeaderDag inserts a batch of headers into the local chain.
	InsertHeaderDag([]*modules.Header, int) (int, error)
	HasUnit(hash common.Hash) bool
	SaveUnit(unit modules.Unit, isGenesis bool) error
	//All leaf nodes for dag downloader
	GetAllLeafNodes() ([]*modules.Header, error)
	GetUnit(common.Hash) *modules.Unit
	CreateUnit(mAddr *common.Address, txpool *txspool.TxPool, ks *keystore.KeyStore, t time.Time) ([]modules.Unit, error)
	GetActiveMediatorNode(index int) *discover.Node
	FastSyncCommitHead(common.Hash) error
	GetGenesisUnit(index uint64) (*modules.Unit, error)
	GetContractState(id string, field string) (*modules.StateVersion, []byte)
	GetUnitNumber(hash common.Hash) (modules.ChainIndex, error)
	GetCanonicalHash(number uint64) (common.Hash, error)
	GetHeadHeaderHash() (common.Hash, error)
	GetHeadUnitHash() (common.Hash, error)
	GetHeadFastUnitHash() (common.Hash, error)
	GetUtxoView(tx *modules.Transaction) (*txspool.UtxoViewpoint, error)
	SubscribeChainHeadEvent(ch chan<- modules.ChainHeadEvent) event.Subscription
	GetTrieSyncProgress() (uint64, error)
	GetUtxoEntry(key []byte) (*modules.Utxo, error)
	GetAddrOutput(addr string) ([]modules.Output, error)
	GetAddrTransactions(addr string) (modules.Transactions, error)
	GetContractTpl(templateID []byte) (version *modules.StateVersion, bytecode []byte, name string, path string)
	WalletTokens(addr common.Address) (map[string]*modules.AccountToken, error)
	WalletBalance(address common.Address, assetid []byte, uniqueid []byte, chainid uint64) (uint64, error)
	GetContract(id common.Hash) (*modules.Contract, error)
	GetActiveMediatorAddr(index int) common.Address
	GetActiveMediatorInitPubs() []kyber.Point
	GetCurThreshold() int
	IsActiveMediator(add common.Address) bool
	GetGlobalProp() *modules.GlobalProperty
	GetDynGlobalProp() *modules.DynamicGlobalProperty
	GetMediatorSchl() *modules.MediatorSchedule
	GetActiveMediatorCount() int
}
