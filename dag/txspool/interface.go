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
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/event"
	"github.com/palletone/go-palletone/dag/modules"
)

type ITxPool interface {
	Stop()

	AddLocal(tx *modules.Transaction) error
	AddLocals(txs []*modules.Transaction) []error
	AddSequenTx(tx *modules.Transaction) error
	AddSequenTxs(txs []*modules.Transaction) error
	AllHashs() []*common.Hash
	AllTxpoolTxs() map[common.Hash]*modules.TxPoolTransaction

	// AddRemotes should add the given transactions to the pool.
	AddRemote(tx *modules.Transaction) error
	AddRemotes([]*modules.Transaction) []error
	ProcessTransaction(tx *modules.Transaction, allowOrphan bool, rateLimit bool, tag Tag) ([]*TxDesc, error)

	Pending() (map[common.Hash][]*modules.TxPoolTransaction, error)
	Queued() ([]*modules.TxPoolTransaction, error)
	SetPendingTxs(unit_hash common.Hash, num uint64, txs []*modules.Transaction) error
	ResetPendingTxs(txs []*modules.Transaction) error
	SendStoredTxs(hashs []common.Hash) error
	DiscardTxs(hashs []common.Hash) error
	GetUtxoEntry(outpoint *modules.OutPoint) (*modules.Utxo, error)

	SubscribeTxPreEvent(chan<- modules.TxPreEvent) event.Subscription
	GetSortedTxs(hash common.Hash, index uint64) ([]*modules.TxPoolTransaction, common.StorageSize)
	Get(hash common.Hash) (*modules.TxPoolTransaction, common.Hash)
	GetPoolTxsByAddr(addr string) ([]*modules.TxPoolTransaction, error)
	Stats() (int, int, int)
	Content() (map[common.Hash]*modules.TxPoolTransaction, map[common.Hash]*modules.TxPoolTransaction)
	GetTxFee(tx *modules.Transaction) (*modules.AmountAsset, error)
	OutPointIsSpend(outPoint *modules.OutPoint) (bool, error)
	ValidateOrphanTx(tx *modules.Transaction) (bool, error)
}
