/*
 *
 *     This file is part of go-palletone.
 *     go-palletone is free software: you can redistribute it and/or modify
 *     it under the terms of the GNU General Public License as published by
 *     the Free Software Foundation, either version 3 of the License, or
 *     (at your option) any later version.
 *     go-palletone is distributed in the hope that it will be useful,
 *     but WITHOUT ANY WARRANTY; without even the implied warranty of
 *     MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *     GNU General Public License for more details.
 *     You should have received a copy of the GNU General Public License
 *     along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
 * /
 *
 *  * @author PalletOne core developers <dev@pallet.one>
 *  * @date 2018
 *
 */

package storage

import (
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/dag/modules"
)

type IStateDb interface {
	GetPrefix(prefix []byte) map[string][]byte
	//Contract statedb
	SaveContract(contract *modules.Contract) error
	GetContract(id []byte) (*modules.Contract, error)
	GetAllContracts() ([]*modules.Contract, error)
	SaveContractState(id []byte, w *modules.ContractWriteSet, version *modules.StateVersion) error
	SaveContractStates(id []byte, wset []modules.ContractWriteSet, version *modules.StateVersion) error
	GetContractState(id []byte, field string) ([]byte, *modules.StateVersion, error)
	GetContractStatesByPrefix(id []byte, prefix string) (map[string]*modules.ContractStateValue, error)
	GetContractStatesById(id []byte) (map[string]*modules.ContractStateValue, error)

	SaveContractTpl(tpl *modules.ContractTemplate) error
	SaveContractTplCode(tplId []byte, byteCode []byte) error
	GetContractTpl(tplId []byte) (*modules.ContractTemplate, error)
	GetContractTplCode(tplId []byte) ([]byte, error)
	GetAllContractTpl() ([]*modules.ContractTemplate, error)
	GetContractIdsByTpl(tplId []byte) ([][]byte, error)
	SaveContractDeploy(reqid []byte, deploy *modules.ContractDeployPayload) error
	SaveContractDeployReq(reqid []byte, deploy *modules.ContractDeployRequestPayload) error
	SaveContractInvokeReq(reqid []byte, invoke *modules.ContractInvokeRequestPayload) error
	SaveContractStop(reqid []byte, stop *modules.ContractStopPayload) error
	SaveContractStopReq(reqid []byte, stopr *modules.ContractStopRequestPayload) error
	SaveContractSignature(reqid []byte, sig *modules.SignaturePayload) error

	GetContractDeploy(reqId []byte) (*modules.ContractDeployPayload, error)
	GetContractDeployReq(reqid []byte) (*modules.ContractDeployRequestPayload, error)
	GetContractInvoke(reqId []byte) (*modules.ContractInvokePayload, error)
	GetContractInvokeReq(reqid []byte) (*modules.ContractInvokeRequestPayload, error)
	GetContractStop(reqId []byte) (*modules.ContractStopPayload, error)
	GetContractStopReq(reqId []byte) (*modules.ContractStopRequestPayload, error)
	GetContractSignature(reqId []byte) (*modules.SignaturePayload, error)

	SaveAccountState(address common.Address, write *modules.AccountStateWriteSet, version *modules.StateVersion) error
	SaveAccountStates(address common.Address, writeset []modules.AccountStateWriteSet, version *modules.StateVersion) error
	GetAllAccountStates(address common.Address) (map[string]*modules.ContractStateValue, error)
	GetAccountState(address common.Address, statekey string) (*modules.ContractStateValue, error)
	UpdateAccountBalance(addr common.Address, addAmount int64) error
	GetAccountBalance(address common.Address) uint64
	GetAccountVotedMediators(addr common.Address) map[string]bool
	LookupAccount() map[common.Address]*modules.AccountInfo

	GetMinFee() (*modules.AmountAsset, error)
	//获得一个合约的陪审团列表
	GetContractJury(contractId []byte) ([]modules.ElectionInf, error)
	SaveContractJury(contractId []byte, jury []modules.ElectionInf, version *modules.StateVersion) error
	// world state chainIndex
	//GetCurrentChainIndex(assetId modules.AssetId) (*modules.ChainIndex, error)
	//保存当前最新单元的高度，即使是未稳定的单元，也会更新
	//SaveChainIndex(index *modules.ChainIndex) error
	//GetCurrentUnit(assetId modules.AssetId) *modules.Unit

	StoreMediator(med *core.Mediator) error
	StoreMediatorInfo(add common.Address, mi *modules.MediatorInfo) error
	RetrieveMediator(address common.Address) (*core.Mediator, error)
	GetMediators() map[common.Address]bool
	LookupMediatorInfo() []*modules.MediatorInfo
	IsMediator(address common.Address) bool
	RetrieveMediatorInfo(address common.Address) (*modules.MediatorInfo, error)

	GetJuryCandidateList() (map[string]bool, error)
	IsInJuryCandidateList(address common.Address) bool
	GetDataVersion() (*modules.DataVersion, error)
	SaveDataVersion(dv *modules.DataVersion) error

	GetPartitionChains() ([]*modules.PartitionChain, error)
	GetMainChain() (*modules.MainChain, error)

	//GetSysConfig(name string) ([]byte, *modules.StateVersion, error)
	//GetAllSysConfig() (map[string]*modules.ContractStateValue, error)
	GetSysParamWithoutVote() (map[string]string, error)
	GetSysParamsWithVotes() (*modules.SysTokenIDInfo, error)
	SaveSysConfigContract(key string, val []byte, ver *modules.StateVersion) error
}
