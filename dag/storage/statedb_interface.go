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
	SaveContractState(id common.Address, name string, value interface{}, version *modules.StateVersion) error
	SaveContractTemplate(templateId []byte, bytecode []byte, version []byte) error
	SaveContractTemplateState(id []byte, name string, value interface{}, version *modules.StateVersion) error
	DeleteState(key []byte) error
	GetContractTpl(templateID []byte) (version *modules.StateVersion, bytecode []byte, name string, path string)
	GetContractState(id common.Address, field string) (*modules.StateVersion, []byte)
	GetTplAllState(id []byte) []*modules.ContractReadSet
	GetContractAllState(id common.Address) []*modules.ContractReadSet
	GetTplState(id []byte, field string) (*modules.StateVersion, []byte)
	GetContract(id common.Address) (*modules.Contract, error)
	GetAccountInfo(address common.Address) (*modules.AccountInfo, error)
	SaveAccountInfo(address common.Address, info *modules.AccountInfo) error
	GetCandidateMediatorAddrList() ([]common.Address, error)
	//GetActiveMediatorAddrList() ([]common.Address, error)

	AddVote(voter common.Address, candidate common.Address) error
	GetSortedVote(CandidateNumber uint) ([]Candidate, error)
	SaveCandidateMediatorAddrList(addrs []common.Address, v *modules.StateVersion) error
	GetAccountMediatorInfo(address common.Address) (*core.MediatorInfo, error)
	SaveAccountMediatorInfo(address common.Address, info *core.MediatorInfo, version *modules.StateVersion) error
}
