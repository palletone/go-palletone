package dboperation

import (
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/dag/modules"
)

type IContractDag interface {
	GetDb() ptndb.Database
	GetContractStatesById(contractid []byte) (map[string]*modules.ContractStateValue, error)
	GetContractState(contractid []byte, field string) ([]byte, *modules.StateVersion, error)
	GetContractStatesByPrefix(contractid []byte, prefix string) (map[string]*modules.ContractStateValue, error)

	UnstableHeadUnitProperty(asset modules.AssetId) (*modules.UnitProperty, error)
	GetGlobalProp() *modules.GlobalProperty

	GetNewestUnit(token modules.AssetId) (common.Hash, *modules.ChainIndex, error)
	GetHeaderByNumber(number *modules.ChainIndex) (*modules.Header, error)

	GetAddrUtxos(addr common.Address) (map[modules.OutPoint]*modules.Utxo, error)
	GetAddr1TokenUtxos(addr common.Address, asset *modules.Asset) (map[modules.OutPoint]*modules.Utxo, error)
	GetStableTransactionOnly(hash common.Hash) (*modules.Transaction, error)
	GetStableUnit(hash common.Hash) (*modules.Unit, error)
	GetStableUnitByNumber(number *modules.ChainIndex) (*modules.Unit, error)
	GetAddrByOutPoint(outPoint *modules.OutPoint) (common.Address, error)
	//GetTxFee(pay *modules.Transaction) (*modules.AmountAsset, error)
	GetContract(id []byte) (*modules.Contract, error)
	GetChainParameters() *core.ChainParameters
	GetContractTpl(tplId []byte) (*modules.ContractTemplate, error)
	GetContractTplCode(tplId []byte) ([]byte, error)
	SaveTransaction(tx *modules.Transaction, txIndex int) error
	GetUtxoEntry(outpoint *modules.OutPoint) (*modules.Utxo, error)
	GetContractsWithJuryAddr(addr common.Hash) []*modules.Contract
	SaveContract(contract *modules.Contract) error
	GetImmutableChainParameters() *core.ImmutableChainParameters
	NewTemp() (IContractDag, error)
}
