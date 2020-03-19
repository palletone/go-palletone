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
 *  * @date 2018
 *
 */

package txspool

import (
	"github.com/ethereum/go-ethereum/event"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/dag/modules"
)

type ITxPool interface {
	Stop()

	AddLocal(tx *modules.Transaction) error
	//AddLocals(txs []*modules.Transaction) []error
	//AddSequenTx(tx *modules.Transaction) error
	//AddSequenTxs(txs []*modules.Transaction) error
	//AllHashs() []*common.Hash
	//AllTxpoolTxs() map[common.Hash]*TxPoolTransaction

	// AddRemotes should add the given transactions to the pool.
	AddRemote(tx *modules.Transaction) error
	//AddRemotes([]*modules.Transaction) []error
	//ProcessTransaction(tx *modules.Transaction, allowOrphan bool, rateLimit bool, tag Tag) ([]*TxDesc, error)
	//查询已打包的交易，以UnitHash为Key
	Pending() (map[common.Hash][]*TxPoolTransaction, error)
	//查询孤儿交易
	Queued() ([]*TxPoolTransaction, error)
	//将一堆交易修改状态为已打包
	SetPendingTxs(unit_hash common.Hash, num uint64, txs []*modules.Transaction) error
	//将一堆交易设置为未打包
	ResetPendingTxs(txs []*modules.Transaction) error
	//SendStoredTxs(hashs []common.Hash) error
	//将一堆交易标记为删除
	DiscardTxs(txs []*modules.Transaction) error
	//查询UTXO
	GetUtxo(outpoint *modules.OutPoint) (*modules.Utxo, error)
	//订阅事件
	SubscribeTxPreEvent(chan<- modules.TxPreEvent) event.Subscription
	//GetSortedTxs(hash common.Hash, index uint64) ([]*TxPoolTransaction, common.StorageSize)
	//迭代获取未打包的排序好的Tx，迭代执行函数时，如果返回true就继续迭代，如果false停止迭代
	GetSortedTxs(processor func(tx *TxPoolTransaction) (getNext bool, err error)) error
	//从交易池获取某个交易
	GetTx(hash common.Hash) (*TxPoolTransaction, error)
	//获取交易池中某个地址的所有交易
	//GetPoolTxsByAddr(addr string) ([]*TxPoolTransaction, error)
	//获得一个地址的未打包的交易
	GetUnpackedTxsByAddr(addr common.Address) ([]*TxPoolTransaction, error)
	//返回交易池中几种状态的交易数量
	Status() (int, int, int)
	//返回交易池中交易的内容
	Content() (map[common.Hash]*TxPoolTransaction, map[common.Hash]*TxPoolTransaction)
}
