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
	GetConfig(name string) ([]byte, *modules.StateVersion, error)
	GetContractStatesById(id []byte) (map[string]*modules.ContractStateValue, error)
	GetContractStatesByPrefix(id []byte, prefix string) (map[string]*modules.ContractStateValue, error)

	GetContract(id []byte) (*modules.Contract, error)
	GetContractTpl(templateID []byte) (version *modules.StateVersion, bytecode []byte, name string, path string, tplVersion string)
	GetContractDeploy(tempId, contractId []byte, name string) (*modules.ContractDeployPayload, error)
	GetAllAccountStates(address common.Address) (map[string]*modules.ContractStateValue, error)
	GetAccountState(address common.Address, statekey string) (*modules.ContractStateValue, error)

	GetAccountBalance(address common.Address) uint64
	RetrieveMediator(address common.Address) (*core.Mediator, error)
	StoreMediator(med *core.Mediator) error
	GetMediators() map[common.Address]bool
	GetApprovedMediatorList() ([]*modules.MediatorRegisterInfo, error)
	IsApprovedMediator(address common.Address) bool
	IsMediator(address common.Address) bool
	LookupAccount() map[common.Address]*modules.AccountInfo
	RetrieveMediatorInfo(address common.Address) (*modules.MediatorInfo, error)
	GetMinFee() (*modules.AmountAsset, error)
	//GetCurrentChainIndex(assetId modules.AssetId) (*modules.ChainIndex, error)

	GetJuryCandidateList() ([]common.Address, error)
	IsJury(address common.Address) bool
	UpdateSysParams(ver *modules.StateVersion) error
	GetPartitionChains() ([]*modules.PartitionChain, error)
	GetMainChain() (*modules.MainChain, error)
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
	return &StateRepository{statedb: statedb}
}

func (rep *StateRepository) GetContractState(id []byte, field string) ([]byte, *modules.StateVersion, error) {
	return rep.statedb.GetContractState(id, field)
}

func (rep *StateRepository) GetConfig(name string) ([]byte, *modules.StateVersion, error) {
	return rep.statedb.GetSysConfig(name)
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

func (rep *StateRepository) GetContractTpl(templateID []byte) (version *modules.StateVersion, bytecode []byte, name string, path string, tplVersion string) {
	return rep.statedb.GetContractTpl(templateID)
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

func (rep *StateRepository) GetApprovedMediatorList() ([]*modules.MediatorRegisterInfo, error) {
	return rep.statedb.GetApprovedMediatorList()
}

func (rep *StateRepository) IsApprovedMediator(address common.Address) bool {
	return rep.statedb.IsApprovedMediator(address)
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

func (rep *StateRepository) GetContractDeploy(tempId, contractId []byte, name string) (*modules.ContractDeployPayload, error) {
	return rep.statedb.GetContractDeploy(tempId[:])
}

func (rep *StateRepository) GetMinFee() (*modules.AmountAsset, error) {
	return rep.statedb.GetMinFee()
}

func (rep *StateRepository) GetJuryCandidateList() ([]common.Address, error) {
	return rep.statedb.GetJuryCandidateList()
}

func (rep *StateRepository) IsJury(address common.Address) bool {
	return rep.statedb.IsInJuryCandidateList(address)
}

func (rep *StateRepository) UpdateSysParams(ver *modules.StateVersion) error {
	return rep.statedb.UpdateSysParams(ver)
}

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
