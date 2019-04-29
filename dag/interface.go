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
	"time"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/event"
	"github.com/palletone/go-palletone/common/p2p/discover"
	"github.com/palletone/go-palletone/contracts/list"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/txspool"
)

type IDag interface {
	Close()

	//common geter
	GetCommon(key []byte) ([]byte, error)
	GetCommonByPrefix(prefix []byte) map[string][]byte

	IsEmpty() bool
	CurrentUnit(token modules.AssetId) *modules.Unit
	//SaveDag(unit *modules.Unit, isGenesis bool) (int, error)
	VerifyHeader(header *modules.Header) error
	GetCurrentUnit(assetId modules.AssetId) *modules.Unit
	GetMainCurrentUnit() *modules.Unit
	GetCurrentMemUnit(assetId modules.AssetId, index uint64) *modules.Unit
	InsertDag(units modules.Units, txpool txspool.ITxPool) (int, error)
	GetUnitByHash(hash common.Hash) (*modules.Unit, error)
	HasHeader(common.Hash, uint64) bool
	GetHeaderByNumber(number *modules.ChainIndex) (*modules.Header, error)
	// GetHeaderByHash retrieves a header from the local chain.
	GetHeaderByHash(common.Hash) (*modules.Header, error)
	GetUnstableUnits() []*modules.Unit
	//GetPrefix(prefix string) map[string][]byte

	// CurrentHeader retrieves the head header from the local chain.
	CurrentHeader(token modules.AssetId) *modules.Header
	GetUnitTransactions(hash common.Hash) (modules.Transactions, error)
	GetUnitTxsHash(hash common.Hash) ([]common.Hash, error)
	GetTransaction(hash common.Hash) (*modules.TransactionWithUnitInfo, error)
	GetTransactionOnly(hash common.Hash) (*modules.Transaction, error)
	IsTransactionExist(hash common.Hash) (bool, error)
	GetTxSearchEntry(hash common.Hash) (*modules.TxLookupEntry, error)

	// InsertHeaderDag inserts a batch of headers into the local chain.
	InsertHeaderDag([]*modules.Header) (int, error)
	HasUnit(hash common.Hash) bool
	UnitIsConfirmedByHash(hash common.Hash) bool
	ParentsIsConfirmByHash(hash common.Hash) bool
	IsHeaderExist(hash common.Hash) bool
	SaveUnit(unit *modules.Unit, txpool txspool.ITxPool, isGenesis bool) error
	CreateUnit(mAddr *common.Address, txpool txspool.ITxPool, t time.Time) (*modules.Unit, error)

	// validate group signature by hash
	//ValidateUnitGroupSig(hash common.Hash) (bool, error)

	FastSyncCommitHead(common.Hash) error
	GetGenesisUnit() (*modules.Unit, error)

	GetConfig(name string) ([]byte, *modules.StateVersion, error)
	GetContractState(contractid []byte, field string) ([]byte, *modules.StateVersion, error)
	GetContractStatesById(id []byte) (map[string]*modules.ContractStateValue, error)
	GetContractStatesByPrefix(id []byte, prefix string) (map[string]*modules.ContractStateValue, error)

	GetUnitNumber(hash common.Hash) (*modules.ChainIndex, error)
	//GetCanonicalHash(number uint64) (common.Hash, error)
	//GetHeadHeaderHash() (common.Hash, error)
	//GetHeadUnitHash() (common.Hash, error)
	//GetHeadFastUnitHash() (common.Hash, error)
	GetUtxoView(tx *modules.Transaction) (*txspool.UtxoViewpoint, error)
	SubscribeChainHeadEvent(ch chan<- modules.ChainHeadEvent) event.Subscription
	GetTrieSyncProgress() (uint64, error)
	GetUtxoEntry(outpoint *modules.OutPoint) (*modules.Utxo, error)
	//GetUtxoPkScripHexByTxhash(txhash common.Hash, mindex, outindex uint32) (string, error)
	//GetAddrOutput(addr string) ([]modules.Output, error)
	GetAddrOutpoints(addr common.Address) ([]modules.OutPoint, error)
	GetAddrUtxos(addr common.Address) (map[modules.OutPoint]*modules.Utxo, error)
	GetAddr1TokenUtxos(addr common.Address, asset *modules.Asset) (map[modules.OutPoint]*modules.Utxo, error)
	GetAllUtxos() (map[modules.OutPoint]*modules.Utxo, error)
	GetAddrTransactions(addr common.Address) ([]*modules.TransactionWithUnitInfo, error)
	GetAssetTxHistory(asset *modules.Asset) ([]*modules.TransactionWithUnitInfo, error)

	GetContractTpl(templateID []byte) (version *modules.StateVersion, bytecode []byte, name string, path string, tplVersion string)
	//WalletTokens(addr common.Address) (map[string]*modules.AccountToken, error)
	//WalletBalance(address common.Address, assetid []byte, uniqueid []byte, chainid uint64) (uint64, error)
	GetContract(id []byte) (*modules.Contract, error)
	GetUnitByNumber(number *modules.ChainIndex) (*modules.Unit, error)
	GetUnitHashesFromHash(hash common.Hash, max uint64) []common.Hash

	//Mediator
	GetActiveMediator(add common.Address) *core.Mediator
	GetActiveMediatorNode(index int) *discover.Node
	GetActiveMediatorNodes() map[string]*discover.Node

	GetAddrByOutPoint(outPoint *modules.OutPoint) (common.Address, error)
	GetTxFee(pay *modules.Transaction) (*modules.AmountAsset, error)
	SetUnitGroupSign(unitHash common.Hash, groupSign []byte, txpool txspool.ITxPool) error

	IsSynced() bool
	SubscribeActiveMediatorsUpdatedEvent(ch chan<- ActiveMediatorsUpdatedEvent) event.Subscription
	GetPrecedingMediatorNodes() map[string]*discover.Node
	UnitIrreversibleTime() time.Duration
	GenTransferPtnTx(from, to common.Address, daoAmount uint64, text *string, txPool txspool.ITxPool) (*modules.Transaction, uint64, error)

	QueryDbByKey(key []byte) ([]byte, error)
	QueryDbByPrefix(prefix []byte) ([]*modules.DbRow, error)

	// SaveReqIdByTx
	//GetReqIdByTxHash(hash common.Hash) (common.Hash, error)
	GetTxHashByReqId(reqid common.Hash) (common.Hash, error)
	GetTxByReqId(reqid common.Hash) (*modules.TransactionWithUnitInfo, error)

	//SaveReqIdByTx(tx *modules.Transaction) error

	GetTxFromAddress(tx *modules.Transaction) ([]common.Address, error)

	GetFileInfo(filehash []byte) ([]*modules.FileInfo, error)

	//Light Palletone Subprotocal
	GetLightHeaderByHash(headerHash common.Hash) (*modules.Header, error)
	GetLightChainHeight(assetId modules.AssetId) uint64
	InsertLightHeader(headers []*modules.Header) (int, error)
	GetAllLeafNodes() ([]*modules.Header, error)

	HeadUnitTime() int64
	HeadUnitNum() uint64
	HeadUnitHash() common.Hash

	UpdateSysParams() error

	ValidateUnitExceptGroupSig(unit *modules.Unit) error
	RefreshSysParameters()

	SaveChaincode(contractId common.Address, cc *list.CCInfo) error
	GetChaincodes(contractId common.Address) (*list.CCInfo, error)
	GetPartitionChains() ([]*modules.PartitionChain, error)
	GetMainChain() (*modules.MainChain, error)
	GetCoinYearRate() float64
}
