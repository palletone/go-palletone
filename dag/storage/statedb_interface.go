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
	GetConfig(name string) ([]byte, *modules.StateVersion, error)
	GetPrefix(prefix []byte) map[string][]byte
	SaveConfig(confs []modules.ContractWriteSet, stateVersion *modules.StateVersion) error
	//SaveAssetInfo(assetInfo *modules.AssetInfo) error
	//GetAssetInfo(assetId *modules.Asset) (*modules.AssetInfo, error)
	SaveContract(contract *modules.Contract) error
	SaveContractState(id []byte, name string, value interface{}, version *modules.StateVersion) error
	SaveContractTemplate(templateId []byte, bytecode []byte, version []byte) error
	SaveContractTemplateState(id []byte, name string, value interface{}, version *modules.StateVersion) error
	SaveContractDeployReq(deploy *modules.ContractDeployRequestPayload) error
	SaveContractInvokeReq(invoke *modules.ContractInvokeRequestPayload) error

	DeleteState(key []byte) error
	GetContractTpl(templateID []byte) (version *modules.StateVersion, bytecode []byte, name string, path string, tplVersion string)
	GetContractState(id []byte, field string) (*modules.StateVersion, []byte)
	GetTplAllState(id []byte) []*modules.ContractReadSet
	GetContractAllState() []*modules.ContractReadSet
	GetContractStatesById(id []byte) (map[string]*modules.ContractStateValue, error)
	GetTplState(id []byte, field string) (*modules.StateVersion, []byte)
	GetContract(id []byte) (*modules.Contract, error)
	GetContractDeployReq(tempId, txId []byte) (*modules.ContractDeployRequestPayload, error)
	GetContractInvokeReq(contractId []byte, funcName string) (*modules.ContractInvokeRequestPayload, error)
	/* Account_Info */
	RetrieveAccountInfo(address common.Address) (*modules.AccountInfo, error)
	StoreAccountInfo(address common.Address, info *modules.AccountInfo) error
	UpdateAccountInfoBalance(addr common.Address, addAmount int64) error
	//AddVote2Account(address common.Address, voteInfo vote.VoteInfo) error
	//GetAccountVoteInfo(address common.Address, voteType uint8) [][]byte

	//GetSortedMediatorVote(returnNumber int) (map[string]uint64, error)
	//GetVoterList(voteType uint8, MinTermLimit uint16) []common.Address
	//UpdateVoterList(voter common.Address, voteType uint8, term uint16) error
	//GetAccountMediatorVote(voterAddress common.Address) ([]common.Address, uint64, error)
	AppendVotedMediator(voter, mediator common.Address) error

	// world state chainIndex
	//GetCurrentChainIndex(assetId modules.IDType16) (*modules.ChainIndex, error)
	//保存当前最新单元的高度，即使是未稳定的单元，也会更新
	//SaveChainIndex(index *modules.ChainIndex) error
	//GetCurrentUnit(assetId modules.IDType16) *modules.Unit

	CreateUserVote(voter common.Address, detail [][]byte, bHash []byte) error

	StoreMediator(med *core.Mediator) error
	StoreMediatorInfo(add common.Address, mi *modules.MediatorInfo) error
	RetrieveMediator(address common.Address) (*core.Mediator, error)
	GetMediatorCount() int

	GetMediators() map[common.Address]bool
	LookupMediator() map[common.Address]*core.Mediator

	GetApprovedMediatorList() ([]*modules.MediatorRegisterInfo, error)
	IsApprovedMediator(address common.Address) bool
	IsMediator(address common.Address) bool
	LookupAccount() map[common.Address]*modules.AccountInfo
	RetrieveMediatorInfo(address common.Address) (*modules.MediatorInfo, error)
}
