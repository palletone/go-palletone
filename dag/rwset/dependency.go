/*
	This file is part of go-palletone.
	go-palletone is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.
	go-palletone is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.
	You should have received a copy of the GNU General Public License
	along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
*/

/*
 * @author PalletOne core developers <dev@pallet.one>
 * @date 2018-2020
 */
package rwset

import (
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/dag/modules"
)

type IDataQuery interface {
	IStateQuery
	UnstableHeadUnitProperty(asset modules.AssetId) (*modules.UnitProperty, error)
	GetGlobalProp() *modules.GlobalProperty

	GetNewestUnit(token modules.AssetId) (common.Hash, *modules.ChainIndex, error)
	GetHeaderByNumber(number *modules.ChainIndex) (*modules.Header, error)
	GetAddrUtxos(addr common.Address) (map[modules.OutPoint]*modules.Utxo, error)
	GetAddr1TokenUtxos(addr common.Address, asset *modules.Asset) (map[modules.OutPoint]*modules.Utxo, error)
	GetStableTransactionOnly(hash common.Hash) (*modules.Transaction, error)
	GetStableUnit(hash common.Hash) (*modules.Unit, error)
	GetStableUnitByNumber(number *modules.ChainIndex) (*modules.Unit, error)
}
type IStateQuery interface {
	GetContractStatesById(contractid []byte) (map[string]*modules.ContractStateValue, error)
	GetContractState(contractid []byte, field string) ([]byte, *modules.StateVersion, error)
	GetContractStatesByPrefix(contractid []byte, prefix string) (map[string]*modules.ContractStateValue, error)
}
