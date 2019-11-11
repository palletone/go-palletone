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

package memunit

import (
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/event"
	common2 "github.com/palletone/go-palletone/dag/common"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/txspool"
)

type IMemDag interface {
	AddStableUnit(unit *modules.Unit)
	SetStableUnit(unit *modules.Unit, isGenesis bool)
	AddUnit(unit *modules.Unit, txpool txspool.ITxPool, isProd bool) (common2.IUnitRepository, common2.IUtxoRepository,
		common2.IStateRepository, common2.IPropRepository, common2.IUnitProduceRepository, error)
	SaveHeader(header *modules.Header) error
	GetLastStableUnitInfo() (common.Hash, uint64)
	GetLastMainChainUnit() *modules.Unit
	GetChainUnits() map[common.Hash]*modules.Unit
	SetStableThreshold(threshold int)
	GetUnstableRepositories() (common2.IUnitRepository, common2.IUtxoRepository, common2.IStateRepository,
		common2.IPropRepository, common2.IUnitProduceRepository)
	SetUnitGroupSign(uHash common.Hash, groupSign []byte, txpool txspool.ITxPool) error
	GetHeaderByHash(hash common.Hash) (*modules.Header, error)
	GetHeaderByNumber(number *modules.ChainIndex) (*modules.Header, error)
	Info() (*modules.MemdagStatus, error)

	SubscribeToGroupSignEvent(ch chan<- modules.ToGroupSignEvent) event.Subscription
	Close()
}
