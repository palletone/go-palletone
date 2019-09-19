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
	"math/big"
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

	GetCommon(key []byte) ([]byte, error)
	GetCommonByPrefix(prefix []byte) map[string][]byte
	SaveCommon(key, val []byte) error

	IsEmpty() bool
	GetStableChainIndex(token modules.AssetId) *modules.ChainIndex
	CurrentUnit(token modules.AssetId) *modules.Unit
	GetCurrentUnit(assetId modules.AssetId) *modules.Unit
	GetMainCurrentUnit() *modules.Unit
	GetCurrentMemUnit(assetId modules.AssetId, index uint64) *modules.Unit
	InsertDag(units modules.Units, txpool txspool.ITxPool, is_stable bool) (int, error)
	GetUnitByHash(hash common.Hash) (*modules.Unit, error)
	HasHeader(common.Hash, uint64) bool
	GetHeaderByNumber(number *modules.ChainIndex) (*modules.Header, error)
	GetHeaderByHash(common.Hash) (*modules.Header, error)
	GetUnstableUnits() []*modules.Unit

	CurrentHeader(token modules.AssetId) *modules.Header
	GetUnitTransactions(hash common.Hash) (modules.Transactions, error)
	GetUnitTxsHash(hash common.Hash) ([]common.Hash, error)
	GetTransaction(hash common.Hash) (*modules.TransactionWithUnitInfo, error)
	GetTransactionOnly(hash common.Hash) (*modules.Transaction, error)
	IsTransactionExist(hash common.Hash) (bool, error)
	GetTxSearchEntry(hash common.Hash) (*modules.TxLookupEntry, error)
	GetTxRequesterAddress(tx *modules.Transaction) (common.Address, error)

	// InsertHeaderDag inserts a batch of headers into the local chain.
	InsertHeaderDag([]*modules.Header) (int, error)
	HasUnit(hash common.Hash) bool
	//UnitIsConfirmedByHash(hash common.Hash) bool
	ParentsIsConfirmByHash(hash common.Hash) bool
	IsHeaderExist(hash common.Hash) bool
	SaveUnit(unit *modules.Unit, txpool txspool.ITxPool, isGenesis bool) error
	CreateUnit(mAddr common.Address, txpool txspool.ITxPool, t time.Time) (*modules.Unit, error)

	FastSyncCommitHead(common.Hash) error
	GetGenesisUnit() (*modules.Unit, error)

	GetContractState(contractid []byte, field string) ([]byte, *modules.StateVersion, error)
	GetContractStatesById(id []byte) (map[string]*modules.ContractStateValue, error)
	GetContractStatesByPrefix(id []byte, prefix string) (map[string]*modules.ContractStateValue, error)
	GetContractJury(contractId []byte) (*modules.ElectionNode, error)
	GetUnitNumber(hash common.Hash) (*modules.ChainIndex, error)

	GetUtxoView(tx *modules.Transaction) (*txspool.UtxoViewpoint, error)
	IsUtxoSpent(outpoint *modules.OutPoint) (bool, error)
	SubscribeChainHeadEvent(ch chan<- modules.ChainHeadEvent) event.Subscription
	SubscribeChainEvent(ch chan<- modules.ChainEvent) event.Subscription
	PostChainEvents(events []interface{})

	GetTrieSyncProgress() (uint64, error)
	GetUtxoEntry(outpoint *modules.OutPoint) (*modules.Utxo, error)
	GetStxoEntry(outpoint *modules.OutPoint) (*modules.Stxo, error)
	//Include Utxo and Stxo
	GetTxOutput(outpoint *modules.OutPoint) (*modules.Utxo, error)
	GetAddrOutpoints(addr common.Address) ([]modules.OutPoint, error)
	GetAddrUtxos(addr common.Address) (map[modules.OutPoint]*modules.Utxo, error)
	GetAddrStableUtxos(addr common.Address) (map[modules.OutPoint]*modules.Utxo, error)
	GetAddr1TokenUtxos(addr common.Address, asset *modules.Asset) (map[modules.OutPoint]*modules.Utxo, error)
	GetAllUtxos() (map[modules.OutPoint]*modules.Utxo, error)
	GetAddrTransactions(addr common.Address) ([]*modules.TransactionWithUnitInfo, error)
	GetAssetTxHistory(asset *modules.Asset) ([]*modules.TransactionWithUnitInfo, error)

	GetContractTpl(tplId []byte) (*modules.ContractTemplate, error)
	GetContractTplCode(tplId []byte) ([]byte, error)
	GetAllContractTpl() ([]*modules.ContractTemplate, error)

	GetContract(id []byte) (*modules.Contract, error)
	GetAllContracts() ([]*modules.Contract, error)
	GetContractsByTpl(tplId []byte) ([]*modules.Contract, error)
	GetUnitByNumber(number *modules.ChainIndex) (*modules.Unit, error)
	GetUnitHashesFromHash(hash common.Hash, max uint64) []common.Hash
	GetUnitHash(number *modules.ChainIndex) (common.Hash, error)

	//Mediator
	GetActiveMediator(add common.Address) *core.Mediator
	GetActiveMediatorNode(index int) *discover.Node
	GetActiveMediatorNodes() map[string]*discover.Node

	GetAddrByOutPoint(outPoint *modules.OutPoint) (common.Address, error)
	GetTxFee(pay *modules.Transaction) (*modules.AmountAsset, error)
	SetUnitGroupSign(unitHash common.Hash, groupSign []byte, txpool txspool.ITxPool) error
	SubscribeToGroupSignEvent(ch chan<- modules.ToGroupSignEvent) event.Subscription

	IsSynced() bool
	SubscribeActiveMediatorsUpdatedEvent(ch chan<- modules.ActiveMediatorsUpdatedEvent) event.Subscription
	GetPrecedingMediatorNodes() map[string]*discover.Node
	UnitIrreversibleTime() time.Duration
	GenTransferPtnTx(from, to common.Address, daoAmount uint64, text *string,
		txPool txspool.ITxPool) (*modules.Transaction, uint64, error)

	QueryDbByKey(key []byte) ([]byte, error)
	QueryDbByPrefix(prefix []byte) ([]*modules.DbRow, error)

	// SaveReqIdByTx
	GetTxHashByReqId(reqid common.Hash) (common.Hash, error)
	GetTxByReqId(reqid common.Hash) (*modules.TransactionWithUnitInfo, error)

	GetTxFromAddress(tx *modules.Transaction) ([]common.Address, error)

	GetFileInfo(filehash []byte) ([]*modules.FileInfo, error)

	GetLightHeaderByHash(headerHash common.Hash) (*modules.Header, error)
	GetLightChainHeight(assetId modules.AssetId) uint64
	InsertLightHeader(headers []*modules.Header) (int, error)
	GetAllLeafNodes() ([]*modules.Header, error)
	ClearUtxo(addr common.Address) error
	SaveUtxoView(view map[modules.OutPoint]*modules.Utxo) error

	HeadUnitTime() int64
	HeadUnitNum() uint64
	HeadUnitHash() common.Hash
	GetIrreversibleUnitNum(id modules.AssetId) uint64

	SaveChaincode(contractId common.Address, cc *list.CCInfo) error
	GetChaincode(contractId common.Address) (*list.CCInfo, error)
	RetrieveChaincodes() ([]*list.CCInfo, error)
	GetPartitionChains() ([]*modules.PartitionChain, error)
	GetMainChain() (*modules.MainChain, error)

	RefreshAddrTxIndex() error
	GetMinFee() (*modules.AmountAsset, error)

	GenVoteMediatorTx(voter common.Address, mediators map[string]bool,
		txPool txspool.ITxPool) (*modules.Transaction, uint64, error)
	GetDynGlobalProp() *modules.DynamicGlobalProperty
	GetGlobalProp() *modules.GlobalProperty
	GetMediatorCount() int

	IsMediator(address common.Address) bool
	GetMediators() map[common.Address]bool
	GetActiveMediators() []common.Address
	GetAccountVotedMediators(addr common.Address) map[string]bool
	GetMediatorInfo(address common.Address) *modules.MediatorInfo

	GetVotingForMediator(addStr string) (map[string]uint64, error)
	MediatorVotedResults() (map[string]uint64, error)
	LookupMediatorInfo() []*modules.MediatorInfo
	IsActiveMediator(add common.Address) bool
	GetMediator(add common.Address) *core.Mediator

	GetNewestUnitTimestamp(token modules.AssetId) (int64, error)
	GetScheduledMediator(slotNum uint32) common.Address
	GetSlotAtTime(when time.Time) uint32
	GetChainParameters() *core.ChainParameters
	GetImmutableChainParameters() *core.ImmutableChainParameters

	GetDataVersion() (*modules.DataVersion, error)
	StoreDataVersion(dv *modules.DataVersion) error
	QueryProofOfExistenceByReference(ref []byte) ([]*modules.ProofOfExistence, error)
	GetAssetReference(asset []byte) ([]*modules.ProofOfExistence, error)

	IsActiveJury(addr common.Address) bool
	JuryCount() uint
	GetContractDevelopers() ([]common.Address, error)
	IsContractDeveloper(addr common.Address) bool
	GetActiveJuries() []common.Address
	CreateGenericTransaction(from, to common.Address, daoAmount, daoFee uint64, certID *big.Int,
		msg *modules.Message, txPool txspool.ITxPool) (*modules.Transaction, uint64, error)
	CreateTokenTransaction(from, to, toToken common.Address, daoAmount, daoFee, daoAmountToken uint64, assetToken string,
		msg *modules.Message, txPool txspool.ITxPool) (*modules.Transaction, uint64, error)
	ChainThreshold() int
	CheckHeaderCorrect(number int) error
	GetBlacklistAddress() ([]common.Address, *modules.StateVersion, error)
	RebuildAddrTxIndex() error
}
