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
 *  * @date 2018-2019
 *
 */

package validator

import (
	"time"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/dag/modules"
)

type IUtxoQuery interface {
	GetUtxoEntry(outpoint *modules.OutPoint) (*modules.Utxo, error)
	GetStxoEntry(outpoint *modules.OutPoint) (*modules.Stxo, error)
}

type IStateQuery interface {
	GetContractTpl(tplId []byte) (*modules.ContractTemplate, error)
	GetContractJury(contractId []byte) (*modules.ElectionNode, error)
	GetContractState(id []byte, field string) ([]byte, *modules.StateVersion, error)
	GetContractStatesByPrefix(id []byte, prefix string) (map[string]*modules.ContractStateValue, error)
	GetMediators() map[common.Address]bool
	GetMediator(add common.Address) *core.Mediator
	GetBlacklistAddress() ([]common.Address, *modules.StateVersion, error)
	GetJurorByAddrHash(addrHash common.Hash) (*modules.JurorDeposit, error)
	GetJurorReward(jurorAdd common.Address) common.Address
	IsContractDeveloper(addr common.Address) bool
}

type IDagQuery interface {
	GetTransactionOnly(hash common.Hash) (*modules.Transaction, error)
	IsTransactionExist(hash common.Hash) (bool, error)
	GetHeaderByHash(common.Hash) (*modules.Header, error)
	CheckReadSetValid(contractId []byte, readSet []modules.ContractReadSet) bool
	GetTxRequesterAddress(tx *modules.Transaction) (common.Address, error)
}

type IPropQuery interface {
	GetSlotAtTime(when time.Time) uint32
	GetScheduledMediator(slotNum uint32) common.Address
	GetNewestUnitTimestamp(token modules.AssetId) (int64, error)
	GetChainParameters() *core.ChainParameters
}

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
	SaveTransaction(tx *modules.Transaction) error
	GetUtxoEntry(outpoint *modules.OutPoint) (*modules.Utxo, error)
}
