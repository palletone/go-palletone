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

package common

import (
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/storage"
)

type IStateRepository interface {
	GetContractState(id []byte, field string) ([]byte, *modules.StateVersion, error)
	SaveContractState(id []byte, w *modules.ContractWriteSet, version *modules.StateVersion) error
	GetContractStatesById(id []byte) (map[string]*modules.ContractStateValue, error)
	GetContractStatesByPrefix(id []byte, prefix string) (map[string]*modules.ContractStateValue, error)

	GetContract(id []byte) (*modules.Contract, error)
	GetAllContracts() ([]*modules.Contract, error)
	GetContractsByTpl(tplId []byte) ([]*modules.Contract, error)
	GetContractTpl(tplId []byte) (*modules.ContractTemplate, error)
	GetContractTplCode(tplId []byte) ([]byte, error)
	GetContractDeploy(tempId, contractId []byte, name string) (*modules.ContractDeployPayload, error)

	GetAllAccountStates(address common.Address) (map[string]*modules.ContractStateValue, error)
	GetAccountState(address common.Address, statekey string) (*modules.ContractStateValue, error)
	GetAccountBalance(address common.Address) uint64
	LookupAccount() map[common.Address]*modules.AccountInfo
	GetAccountVotedMediators(addr common.Address) map[string]bool

	RetrieveMediator(address common.Address) (*core.Mediator, error)
	StoreMediator(med *core.Mediator) error
	GetMediators() map[common.Address]bool
	LookupMediatorInfo() []*modules.MediatorInfo
	IsMediator(address common.Address) bool
	RetrieveMediatorInfo(address common.Address) (*modules.MediatorInfo, error)
	StoreMediatorInfo(add common.Address, mi *modules.MediatorInfo) error

	GetMinFee() (*modules.AmountAsset, error)
	//GetCurrentChainIndex(assetId modules.AssetId) (*modules.ChainIndex, error)

	GetJuryCandidateList() (map[string]bool, error)
	IsJury(address common.Address) bool
	//UpdateSysParams(ver *modules.StateVersion) error
	GetPartitionChains() ([]*modules.PartitionChain, error)
	GetMainChain() (*modules.MainChain, error)
	//获得一个合约的陪审团列表
	GetContractJury(contractId []byte) ([]modules.ElectionInf, error)
	GetAllContractTpl() ([]*modules.ContractTemplate, error)
	GetDataVersion() (*modules.DataVersion, error)
	StoreDataVersion(dv *modules.DataVersion) error

	//RefreshSysParameters()
	GetSysParamWithoutVote() (map[string]string, error)
	GetSysParamsWithVotes() (*modules.SysTokenIDInfo, error)
	SaveSysConfigContract(key string, val []byte, ver *modules.StateVersion) error
	//GetSysConfig(name string) ([]byte, *modules.StateVersion, error)
	//GetAllConfig() (map[string]*modules.ContractStateValue, error)
}

type StateRepository struct {
	statedb storage.IStateDb
	//logger  log.ILogger
}

func NewStateRepository(statedb storage.IStateDb) *StateRepository {
	return &StateRepository{statedb: statedb}
}

func NewStateRepository4Db(db ptndb.Database) *StateRepository {
	statedb := storage.NewStateDb(db)
	return NewStateRepository(statedb)
}

func (rep *StateRepository) GetContractState(id []byte, field string) ([]byte, *modules.StateVersion, error) {
	return rep.statedb.GetContractState(id, field)
}

func (rep *StateRepository) SaveSysConfigContract(key string, val []byte, ver *modules.StateVersion) error {
	return rep.statedb.SaveSysConfigContract(key, val, ver)
}

//func (rep *StateRepository) GetSysConfig(name string) ([]byte, *modules.StateVersion, error) {
//	return rep.statedb.GetSysConfig(name)
//}

//func (rep *StateRepository) GetAllConfig() (map[string]*modules.ContractStateValue, error) {
//	return rep.statedb.GetAllSysConfig()
//}

func (rep *StateRepository) GetSysParamWithoutVote() (map[string]string, error) {
	return rep.statedb.GetSysParamWithoutVote()
}

func (rep *StateRepository) GetSysParamsWithVotes() (*modules.SysTokenIDInfo, error) {
	return rep.statedb.GetSysParamsWithVotes()
}

func (rep *StateRepository) GetContractStatesById(id []byte) (map[string]*modules.ContractStateValue, error) {
	return rep.statedb.GetContractStatesById(id)
}

func (rep *StateRepository) GetContractStatesByPrefix(id []byte, prefix string) (map[string]*modules.ContractStateValue, error) {
	return rep.statedb.GetContractStatesByPrefix(id, prefix)
}

func (rep *StateRepository) GetContract(id []byte) (*modules.Contract, error) {
	return rep.statedb.GetContract(id)
}
func (rep *StateRepository) GetAllContracts() ([]*modules.Contract, error) {
	return rep.statedb.GetAllContracts()
}
func (rep *StateRepository) GetContractsByTpl(tplId []byte) ([]*modules.Contract, error) {
	cids, err := rep.statedb.GetContractIdsByTpl(tplId)
	if err != nil {
		return nil, err
	}
	result := make([]*modules.Contract, 0, len(cids))
	for _, cid := range cids {
		contract, err := rep.statedb.GetContract(cid)
		if err != nil {
			return nil, err
		}
		result = append(result, contract)
	}
	return result, nil
}

func (rep *StateRepository) GetContractTpl(tplId []byte) (*modules.ContractTemplate, error) {
	return rep.statedb.GetContractTpl(tplId)
}
func (rep *StateRepository) GetContractTplCode(tplId []byte) ([]byte, error) {
	return rep.statedb.GetContractTplCode(tplId)
}

func (rep *StateRepository) RetrieveMediator(address common.Address) (*core.Mediator, error) {
	return rep.statedb.RetrieveMediator(address)
}

func (rep *StateRepository) StoreMediator(med *core.Mediator) error {
	return rep.statedb.StoreMediator(med)
}

func (rep *StateRepository) GetMediators() map[common.Address]bool {
	return rep.statedb.GetMediators()
}

func (rep *StateRepository) SaveContractState(contractId []byte, ws *modules.ContractWriteSet,
	version *modules.StateVersion) error {
	return rep.statedb.SaveContractState(contractId, ws, version)
}

func (rep *StateRepository) IsMediator(address common.Address) bool {
	return rep.statedb.IsMediator(address)
}

func (rep *StateRepository) GetAccountBalance(address common.Address) uint64 {
	return rep.statedb.GetAccountBalance(address)
}

func (rep *StateRepository) LookupAccount() map[common.Address]*modules.AccountInfo {
	return rep.statedb.LookupAccount()
}

func (rep *StateRepository) RetrieveMediatorInfo(address common.Address) (*modules.MediatorInfo, error) {
	return rep.statedb.RetrieveMediatorInfo(address)
}

func (rep *StateRepository) LookupMediatorInfo() []*modules.MediatorInfo {
	return rep.statedb.LookupMediatorInfo()
}

func (rep *StateRepository) StoreMediatorInfo(add common.Address, mi *modules.MediatorInfo) error {
	return rep.statedb.StoreMediatorInfo(add, mi)
}

func (rep *StateRepository) GetContractDeploy(tempId, contractId []byte, name string) (*modules.ContractDeployPayload, error) {
	return rep.statedb.GetContractDeploy(tempId[:])
}

func (rep *StateRepository) GetMinFee() (*modules.AmountAsset, error) {
	return rep.statedb.GetMinFee()
}

func (rep *StateRepository) GetJuryCandidateList() (map[string]bool, error) {
	return rep.statedb.GetJuryCandidateList()
}

func (rep *StateRepository) IsJury(address common.Address) bool {
	return rep.statedb.IsInJuryCandidateList(address)
}

//func (rep *StateRepository) UpdateSysParams(ver *modules.StateVersion) error {
//	return rep.statedb.UpdateSysParams(ver)
//}

func (rep *StateRepository) GetPartitionChains() ([]*modules.PartitionChain, error) {
	return rep.statedb.GetPartitionChains()
}

func (rep *StateRepository) GetMainChain() (*modules.MainChain, error) {
	return rep.statedb.GetMainChain()
}

func (rep *StateRepository) GetAllAccountStates(address common.Address) (map[string]*modules.ContractStateValue, error) {
	return rep.statedb.GetAllAccountStates(address)
}

func (rep *StateRepository) GetAccountState(address common.Address, statekey string) (*modules.ContractStateValue, error) {
	return rep.statedb.GetAccountState(address, statekey)
}

//获得一个合约的陪审团列表
func (rep *StateRepository) GetContractJury(contractId []byte) ([]modules.ElectionInf, error) {
	return rep.statedb.GetContractJury(contractId)
}
func (rep *StateRepository) GetAllContractTpl() ([]*modules.ContractTemplate, error) {
	return rep.statedb.GetAllContractTpl()
}

func (rep *StateRepository) GetAccountVotedMediators(addr common.Address) map[string]bool {
	return rep.statedb.GetAccountVotedMediators(addr)
}

//func (rep *StateRepository) RefreshSysParameters() {
//	deposit, _, _ := rep.statedb.GetSysConfig("DepositRate")
//	depositYearRate, _ := strconv.ParseFloat(string(deposit), 64)
//	parameter.CurrentSysParameters.DepositContractInterest = depositYearRate / 365
//	log.Debugf("Load SysParameter DepositContractInterest value:%f",
//		parameter.CurrentSysParameters.DepositContractInterest)
//
//	txCoinYearRateStr, _, _ := rep.statedb.GetSysConfig("TxCoinYearRate")
//	txCoinYearRate, _ := strconv.ParseFloat(string(txCoinYearRateStr), 64)
//	parameter.CurrentSysParameters.TxCoinDayInterest = txCoinYearRate / 365
//	log.Debugf("Load SysParameter TxCoinDayInterest value:%f", parameter.CurrentSysParameters.TxCoinDayInterest)
//
//	generateUnitRewardStr, _, _ := rep.statedb.GetSysConfig("GenerateUnitReward")
//	generateUnitReward, _ := strconv.ParseUint(string(generateUnitRewardStr), 10, 64)
//	parameter.CurrentSysParameters.GenerateUnitReward = generateUnitReward
//	log.Debugf("Load SysParameter GenerateUnitReward value:%d", parameter.CurrentSysParameters.GenerateUnitReward)
//}

func (rep *StateRepository) GetDataVersion() (*modules.DataVersion, error) {
	return rep.statedb.GetDataVersion()
}
func (rep *StateRepository) StoreDataVersion(dv *modules.DataVersion) error {
	return rep.statedb.SaveDataVersion(dv)
}
