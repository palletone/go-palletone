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
	GetConfig(name []byte) ([]byte, *modules.StateVersion, error)
	GetPrefix(prefix []byte) map[string][]byte
	SaveConfig(confs []modules.PayloadMapStruct, stateVersion *modules.StateVersion) error
	SaveAssetInfo(assetInfo *modules.AssetInfo) error
	GetAssetInfo(assetId *modules.Asset) (*modules.AssetInfo, error)
	SaveContract(contract *modules.Contract) error
	SaveContractState(id []byte, name string, value interface{}, version *modules.StateVersion) error
	SaveContractTemplate(templateId []byte, bytecode []byte, version []byte) error
	SaveContractTemplateState(id []byte, name string, value interface{}, version *modules.StateVersion) error
	DeleteState(key []byte) error
	GetContractTpl(templateID []byte) (version *modules.StateVersion, bytecode []byte, name string, path string)
	GetContractState(id []byte, field string) (*modules.StateVersion, []byte)
	GetTplAllState(id []byte) []*modules.ContractReadSet
	GetContractAllState() []*modules.ContractReadSet
	GetContractStatesById(id []byte) (map[modules.StateVersion][]byte, error)
	GetTplState(id []byte, field string) (*modules.StateVersion, []byte)
	GetContract(id []byte) (*modules.Contract, error)

	/* Account_Info */
	GetAccountInfo(address common.Address) (*modules.AccountInfo, error)
	SaveAccountInfo(address common.Address, info *modules.AccountInfo) error
	GetAccountVoteInfo(address common.Address, voteType uint8) [][]byte
	AddVote2Account(address common.Address, voteInfo modules.VoteInfo) error

	GetSortedVote(ReturnNumber uint, voteType uint8, minTermLimit uint16) ([]common.Address, error)
	GetVoterList(voteType uint8, MinTermLimit uint16) []common.Address
	UpdateVoterList(voter common.Address, voteType uint8, term uint16) error
	UpdateMediatorVote(voter common.Address, candidates []common.Address, mode uint8, term uint16) error
	GetAccountMediatorVote(voterAddress common.Address) ([]common.Address, uint64, error)
	CreateUserVote(voter common.Address, detail [][]byte, bHash []byte) error

	StoreMediatorInfo(mi *core.MediatorInfo) error
	RetrieveMediator(address common.Address) (*core.Mediator, error)
	GetMediatorCount() int
	IsMediator(address common.Address) bool
	GetMediators() map[common.Address]bool
	LookupMediator() map[common.Address]core.Mediator
}
