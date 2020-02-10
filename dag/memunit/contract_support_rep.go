package memunit

import (
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/core"
	dagcommon "github.com/palletone/go-palletone/dag/common"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/tokenengine"
)

type ContractSupportRepository struct {
	db          ptndb.Database
	unitRep     dagcommon.IUnitRepository
	propRep     dagcommon.IPropRepository
	stateRep    dagcommon.IStateRepository
	utxoRep     dagcommon.IUtxoRepository
	tokenEngine tokenengine.ITokenEngine
}

func NewContractSupportRepository(db ptndb.Database) *ContractSupportRepository {
	tokenEngine := tokenengine.Instance
	unitRep := dagcommon.NewUnitRepository4Db(db, tokenEngine)
	utxoRep := dagcommon.NewUtxoRepository4Db(db, tokenEngine)
	stateRep := dagcommon.NewStateRepository4Db(db)
	propRep := dagcommon.NewPropRepository4Db(db)
	return &ContractSupportRepository{
		db:          db,
		unitRep:     unitRep,
		propRep:     propRep,
		stateRep:    stateRep,
		utxoRep:     utxoRep,
		tokenEngine: tokenEngine,
	}
}
func (c *ContractSupportRepository) GetDb() ptndb.Database {
	return c.db
}

func (c *ContractSupportRepository) GetContractStatesById(contractid []byte) (map[string]*modules.ContractStateValue, error) {
	return c.stateRep.GetContractStatesById(contractid)
}

func (c *ContractSupportRepository) GetContractState(contractid []byte, field string) ([]byte, *modules.StateVersion, error) {
	return c.stateRep.GetContractState(contractid, field)
}

func (c *ContractSupportRepository) GetContractStatesByPrefix(contractid []byte, prefix string) (map[string]*modules.ContractStateValue, error) {
	return c.stateRep.GetContractStatesByPrefix(contractid, prefix)
}

func (c *ContractSupportRepository) UnstableHeadUnitProperty(asset modules.AssetId) (*modules.UnitProperty, error) {
	return c.propRep.GetHeadUnitProperty(asset)
}

func (c *ContractSupportRepository) GetGlobalProp() *modules.GlobalProperty {
	gp, _ := c.propRep.RetrieveGlobalProp()
	return gp
}

func (c *ContractSupportRepository) GetNewestUnit(token modules.AssetId) (common.Hash, *modules.ChainIndex, error) {
	return c.propRep.GetNewestUnit(token)
}

func (c *ContractSupportRepository) GetHeaderByNumber(number *modules.ChainIndex) (*modules.Header, error) {
	return c.unitRep.GetHeaderByNumber(number)
}

func (c *ContractSupportRepository) GetAddrUtxos(addr common.Address) (map[modules.OutPoint]*modules.Utxo, error) {
	return c.utxoRep.GetAddrUtxos(addr, nil)
}

func (c *ContractSupportRepository) GetAddr1TokenUtxos(addr common.Address, asset *modules.Asset) (map[modules.OutPoint]*modules.Utxo, error) {
	return c.utxoRep.GetAddrUtxos(addr, asset)
}

func (c *ContractSupportRepository) GetStableTransactionOnly(hash common.Hash) (*modules.Transaction, error) {
	return c.unitRep.GetTransactionOnly(hash)
}

func (c *ContractSupportRepository) GetStableUnit(hash common.Hash) (*modules.Unit, error) {
	return c.unitRep.GetUnit(hash)
}

func (c *ContractSupportRepository) GetStableUnitByNumber(number *modules.ChainIndex) (*modules.Unit, error) {
	hash, err := c.unitRep.GetHashByNumber(number)
	if err != nil {
		log.Debug("GetStableUnitByNumber dagdb.GetHashByNumber err:", "error", err)
		return nil, err
	}
	return c.unitRep.GetUnit(hash)
}

func (c *ContractSupportRepository) GetAddrByOutPoint(outPoint *modules.OutPoint) (common.Address, error) {
	utxo, err := c.utxoRep.GetUtxoEntry(outPoint)
	if err != nil {
		return common.Address{}, err
	}
	return c.tokenEngine.GetAddressFromScript(utxo.PkScript)
}

func (c *ContractSupportRepository) GetContract(id []byte) (*modules.Contract, error) {
	return c.stateRep.GetContract(id)
}

func (c *ContractSupportRepository) GetChainParameters() *core.ChainParameters {
	return c.propRep.GetChainParameters()
}

func (c *ContractSupportRepository) GetContractTpl(tplId []byte) (*modules.ContractTemplate, error) {
	return c.stateRep.GetContractTpl(tplId)
}

func (c *ContractSupportRepository) GetContractTplCode(tplId []byte) ([]byte, error) {
	return c.stateRep.GetContractTplCode(tplId)
}

func (c *ContractSupportRepository) SaveTransaction(tx *modules.Transaction) error {
	return c.unitRep.SaveTransaction(tx)
}

func (c *ContractSupportRepository) GetUtxoEntry(outpoint *modules.OutPoint) (*modules.Utxo, error) {
	return c.utxoRep.GetUtxoEntry(outpoint)
}
