/*
 *
 * 	This file is part of go-palletone.
 * 	go-palletone is free software: you can redistribute it and/or modify
 * 	it under the terms of the GNU General Public License as published by
 * 	the Free Software Foundation, either version 3 of the License, or
 * 	(at your option) any later version.
 * 	go-palletone is distributed in the hope that it will be useful,
 * 	but WITHOUT ANY WARRANTY; without even the implied warranty of
 * 	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * 	GNU General Public License for more details.
 * 	You should have received a copy of the GNU General Public License
 * 	along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
 * /
 *
 *
 *  * @author PalletOne core developer  <dev@pallet.one>
 *  * @date 2018-2020
 *
 */

package dag

import (
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/dag/modules"
)

//通过本地RPC创建或广播的交易
func (dag *Dag) SaveLocalTx(tx *modules.Transaction) error {
	return dag.localRep.SaveLocalTx(tx)
}

//查询某交易的内容和状态
func (dag *Dag) GetLocalTx(txId common.Hash) (*modules.Transaction, modules.TxStatus, error) {
	return dag.localRep.GetLocalTx(txId)
}

//保存某交易的状态
func (dag *Dag) SaveLocalTxStatus(txId common.Hash, status modules.TxStatus) error {
	return dag.localRep.SaveLocalTxStatus(txId, status)
}
