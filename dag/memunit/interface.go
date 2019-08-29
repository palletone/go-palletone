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
	"github.com/palletone/go-palletone/dag/txspool"
)

// type IMemDag interface {
// 	Save(unit *modules.Unit, txpool txspool.ITxPool) error
// 	GetUnit(hash common.Hash) (*modules.Unit, error)
// 	GetHashByNumber(chainIndex *modules.ChainIndex) (common.Hash, error)
// 	UpdateMemDag(hash common.Hash, sign []byte, txpool txspool.ITxPool) error
// 	Exists(uHash common.Hash) bool
// 	Prune(assetId string, hashs []common.Hash) error
// 	SwitchMainChain() error
// 	QueryIndex(assetId string, maturedUnitHash common.Hash) (uint64, int)
// 	GetCurrentUnit(assetid modules.AssetId, index uint64) (*modules.Unit, error)
// 	GetNewestUnit(assetid modules.AssetId) (*modules.Unit, error)
// 	GetDelhashs() chan common.Hash
// 	PushDelHashs(hashs []common.Hash)
// }

type IMemDag interface {
	AddStableUnit(unit *modules.Unit)
	AddUnit(unit *modules.Unit, txpool txspool.ITxPool, isProd bool) (common2.IUnitRepository, common2.IUtxoRepository,
		common2.IStateRepository, common2.IPropRepository, common2.IUnitProduceRepository, error)
	GetLastStableUnitInfo() (common.Hash, uint64)
	GetLastMainChainUnit() *modules.Unit
	GetChainUnits() map[common.Hash]*modules.Unit
	SetStableThreshold(threshold int)
	GetUnstableRepositories() (common2.IUnitRepository, common2.IUtxoRepository, common2.IStateRepository,
		common2.IPropRepository, common2.IUnitProduceRepository)
	SetUnitGroupSign(uHash common.Hash /*, groupPubKey []byte*/, groupSign []byte, txpool txspool.ITxPool) error
	GetHeaderByHash(hash common.Hash) (*modules.Header, error)
	GetHeaderByNumber(number *modules.ChainIndex) (*modules.Header, error)

	SubscribeToGroupSignEvent(ch chan<- modules.ToGroupSignEvent) event.Subscription
	Close()
}
