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
	//SaveConfig(confs []modules.ContractWriteSet, stateVersion *modules.StateVersion) error
	//SaveAssetInfo(assetInfo *modules.AssetInfo) error
	//GetAssetInfo(assetId *modules.Asset) (*modules.AssetInfo, error)
	SaveContract(contract *modules.Contract) error
	SaveContractState(id []byte, name string, value interface{}, version *modules.StateVersion) error
	SaveContractStates(id []byte, wset []modules.ContractWriteSet, version *modules.StateVersion) error
	SaveContractTemplate(templateId []byte, bytecode []byte, version []byte) error
	SaveContractTemplateState(id []byte, name string, value interface{}, version *modules.StateVersion) error
	SaveContractDeploy(reqid []byte, deploy *modules.ContractDeployPayload) error
	SaveContractDeployReq(reqid []byte, deploy *modules.ContractDeployRequestPayload) error
	SaveContractInvoke(reqid []byte, invoke *modules.ContractInvokePayload) error
	SaveContractInvokeReq(reqid []byte, invoke *modules.ContractInvokeRequestPayload) error
	SaveContractStop(reqid []byte, stop *modules.ContractStopPayload) error
	SaveContractStopReq(reqid []byte, stopr *modules.ContractStopRequestPayload) error
	SaveContractSignature(reqid []byte, sig *modules.SignaturePayload) error

	DeleteState(key []byte) error
	GetContractTpl(templateID []byte) (version *modules.StateVersion, bytecode []byte, name string, path string, tplVersion string)
	GetContractState(id []byte, field string) ([]byte, *modules.StateVersion, error)
	GetTplAllState(id []byte) []*modules.ContractReadSet
	//GetContractAllState() []*modules.ContractReadSet
	GetContractStatesByPrefix(id []byte, prefix string) (map[string]*modules.ContractStateValue, error)
	GetContractStatesById(id []byte) (map[string]*modules.ContractStateValue, error)
	GetTplState(id []byte, field string) (*modules.StateVersion, []byte)
	GetContract(id []byte) (*modules.Contract, error)
	GetContractDeploy(reqId []byte) (*modules.ContractDeployPayload, error)
	GetContractDeployReq(reqid []byte) (*modules.ContractDeployRequestPayload, error)
	GetContractInvoke(reqId []byte) (*modules.ContractInvokePayload, error)
	GetContractInvokeReq(reqid []byte) (*modules.ContractInvokeRequestPayload, error)
	GetContractStop(reqId []byte) (*modules.ContractStopPayload, error)
	GetContractStopReq(reqId []byte) (*modules.ContractStopRequestPayload, error)
	GetContractSignature(reqId []byte) (*modules.SignaturePayload, error)
	/* Account_Info */
	RetrieveAccountInfo(address common.Address) (*modules.AccountInfo, error)
	StoreAccountInfo(address common.Address, info *modules.AccountInfo) error
	UpdateAccountBalance(addr common.Address, addAmount int64) error
	GetAccountBalance(address common.Address) uint64
	//AddVote2Account(address common.Address, voteInfo vote.VoteInfo) error
	//GetAccountVoteInfo(address common.Address, voteType uint8) [][]byte
	GetMinFee() (*modules.AmountAsset, error)
	//GetVoterList(voteType uint8, MinTermLimit uint16) []common.Address
	//UpdateVoterList(voter common.Address, voteType uint8, term uint16) error
	AppendVotedMediator(voter, mediator common.Address) error

	// world state chainIndex
	//GetCurrentChainIndex(assetId modules.AssetId) (*modules.ChainIndex, error)
	//保存当前最新单元的高度，即使是未稳定的单元，也会更新
	//SaveChainIndex(index *modules.ChainIndex) error
	//GetCurrentUnit(assetId modules.AssetId) *modules.Unit

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
	UpdateDesiredMediatorCount(account common.Address, mediatorCountSet uint8) error

	GetJuryCandidateList() ([]string, error)
	IsInJuryCandidateList(address common.Address) bool
}
