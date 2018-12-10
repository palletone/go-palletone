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
 * @author PalletOne core developer Albert·Gou <dev@pallet.one>
 * @date 2018/11/05
 */

package mediatorplugin

import (
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/txspool"
)

type PrivateMediatorAPI struct {
	*MediatorPlugin
}

func NewPrivateMediatorAPI(mp *MediatorPlugin) *PrivateMediatorAPI {
	return &PrivateMediatorAPI{mp}
}

// 创建 mediator 所需的参数, 至少包含普通账户地址
type MediatorCreateArgs struct {
	modules.MediatorCreateOperation
}

// 创建 mediator 的执行结果，包含交易哈希，初始dks
type MediatorCreateResult struct {
	TxHash  common.Hash        `json:"txHash"`
	TxSize  common.StorageSize `json:"txSize"`
	Warning string             `json:"warning"`
}

func (a *PrivateMediatorAPI) Register(args MediatorCreateArgs) (MediatorCreateResult, error) {
	res := MediatorCreateResult{}
	addr, err := common.StringToAddress(args.AddStr)
	if err != nil {
		return res, err
	}

	// 1. 组装 message
	msg := &modules.Message{
		App:     modules.OP_MEDIATOR_CREATE,
		Payload: &args.MediatorCreateOperation,
	}

	// 2. 组装 tx
	fee := a.dag.CurrentFeeSchedule().MediatorCreateFee
	tx, err := a.dag.CreateBaseTransaction(addr, addr, 0, fee)
	if err != nil {
		return res, err
	}

	tx.TxMessages = append(tx.TxMessages, msg)
	//tx.TxHash = tx.Hash()

	// 3. 签名 tx
	tx, err = a.ptn.SignGenericTransaction(addr, tx)
	if err != nil {
		return res, err
	}

	// 4. 将 tx 放入 pool
	txPool := a.ptn.TxPool()
	err = txPool.AddLocal(txspool.TxtoTxpoolTx(txPool, tx))
	if err != nil {
		return res, err
	}

	// 5. 返回执行结果
	//res.TxHash = tx.TxHash
	res.TxSize = tx.Size()
	res.Warning = "transaction executed locally, but may not be confirmed by the network yet!"

	return res, nil
}
