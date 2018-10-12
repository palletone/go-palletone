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
	AddRemote(tx *modules.Transaction) error
	Stop()

	AddLocal(tx *modules.TxPoolTransaction) error
	AddLocals(txs []*modules.TxPoolTransaction) []error
	AllHashs() []*common.Hash

	// AddRemotes should add the given transactions to the pool.
	AddRemotes([]*modules.Transaction) []error

	// Pending should return pending transactions.
	// The slice should be modifiable by the caller.
	Pending() (map[common.Hash]*modules.TxPoolTransaction, error)

	// SubscribeTxPreEvent should return an event subscription of
	// TxPreEvent and send events to the given channel.
	SubscribeTxPreEvent(chan<- modules.TxPreEvent) event.Subscription
	GetSortedTxs() ([]*modules.TxPoolTransaction, common.StorageSize)
	GetNonce(hash common.Hash) uint64
	Get(hash common.Hash) *modules.TxPoolTransaction
	Stats() (int, int)
	Content() (map[common.Hash]*modules.Transaction, map[common.Hash]*modules.Transaction)
}
