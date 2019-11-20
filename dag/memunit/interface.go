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

//不稳定单元和新单元的操作
type IMemDag interface {
	//增加一个稳定单元
	AddStableUnit(unit *modules.Unit) error
	//设置MemDag的稳定单元
	SetStableUnit(unit *modules.Unit, isGenesis bool)
	//增加一个单元到MemDag
	AddUnit(unit *modules.Unit, txpool txspool.ITxPool, isProd bool) (common2.IUnitRepository, common2.IUtxoRepository,
		common2.IStateRepository, common2.IPropRepository, common2.IUnitProduceRepository, error)
	//保存Header
	SaveHeader(header *modules.Header) error
	//获取最新稳定单元的信息
	GetLastStableUnitInfo() (common.Hash, uint64)
	//获取主链的最新单元
	GetLastMainChainUnit() *modules.Unit
	//获取所有的不稳定单元
	GetChainUnits() map[common.Hash]*modules.Unit
	//设置要形成稳定单元的阈值，一般是2/3*Count(Mediator)
	SetStableThreshold(threshold int)
	//获得不稳定的Repository
	GetUnstableRepositories() (common2.IUnitRepository, common2.IUtxoRepository, common2.IStateRepository,
		common2.IPropRepository, common2.IUnitProduceRepository)
	//设置一个单元的群签名，使得该单元稳定
	SetUnitGroupSign(uHash common.Hash, groupSign []byte, txpool txspool.ITxPool) error
	//通过Hash获得Header
	GetHeaderByHash(hash common.Hash) (*modules.Header, error)
	//通过高度获得Header
	GetHeaderByNumber(number *modules.ChainIndex) (*modules.Header, error)
	//获得MemDag的信息，包括分叉情况，孤儿块等
	Info() (*modules.MemdagStatus, error)
	//订阅切换主链事件
	SubscribeSwitchMainChainEvent(ob SwitchMainChainEventFunc)
	SubscribeToGroupSignEvent(ch chan<- modules.ToGroupSignEvent) event.Subscription
	//关闭
	Close()
}
