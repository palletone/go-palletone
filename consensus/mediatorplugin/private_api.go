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
	"fmt"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/p2p/discover"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/txspool"
	"github.com/palletone/go-palletone/dag/vote"
)

const defaultResult = "Transaction executed locally, but may not be confirmed by the network yet!"

type PrivateMediatorAPI struct {
	*MediatorPlugin
}

func NewPrivateMediatorAPI(mp *MediatorPlugin) *PrivateMediatorAPI {
	return &PrivateMediatorAPI{mp}
}

// 交易执行结果
type TxExecuteResult struct {
	TxHash  common.Hash        `json:"txHash"`
	TxSize  common.StorageSize `json:"txSize"`
	Warning string             `json:"warning"`
}

// 创建 mediator 所需的参数, 至少包含普通账户地址
type MediatorCreateArgs struct {
	modules.MediatorCreateOperation
}

// 相关参数检查
func (args *MediatorCreateArgs) validate() (common.Address, error) {
	res := common.Address{}
	addr, err := common.StringToAddress(args.AddStr)
	if err != nil {
		return res, err
	}

	res = addr

	_, err = core.StrToPoint(args.InitPartPub)
	if err != nil {
		return res, err
	}

	_, err = discover.ParseNode(args.Node)
	if err != nil {
		return res, err
	}

	return res, nil
}

func (a *PrivateMediatorAPI) Create(args MediatorCreateArgs) (TxExecuteResult, error) {
	res := TxExecuteResult{}
	// 参数验证
	addr, err := args.validate()
	if err != nil {
		return res, err
	}

	// 判断本节点是否同步完成，数据是否最新
	if !a.dag.IsSynced() {
		return res, fmt.Errorf("the data of this node is not up to date, " +
			"and mediator cannot be created at present")
	}

	// 判断是否已经是mediator
	if a.dag.IsMediator(addr) {
		return res, fmt.Errorf("account %v is already a mediator", args.AddStr)
	}

	// 1. 创建交易
	tx, err := a.dag.GenMediatorCreateTx(addr, &args.MediatorCreateOperation)
	if err != nil {
		return res, err
	}

	// 2. 签名和发送交易
	err = a.ptn.SignAndSendTransaction(addr, tx)
	if err != nil {
		return res, err
	}

	// 5. 返回执行结果
	res.TxHash = tx.Hash()
	res.TxSize = tx.Size()
	res.Warning = defaultResult

	return res, nil
}

// 投票 mediator 所需的参数
type VoteMediatorArgs struct {
	Voter    string `json:"voter"`
	Mediator string `json:"mediator"`
}

func (a *PrivateMediatorAPI) Vote(args VoteMediatorArgs) (TxExecuteResult, error) {
	// 参数检查
	res := TxExecuteResult{}
	voter, err := common.StringToAddress(args.Voter)
	if err != nil {
		return res, err
	}

	mediator, err := common.StringToAddress(args.Mediator)
	if err != nil {
		return res, err
	}
	if !a.dag.IsMediator(mediator) {
		return res, fmt.Errorf("%v is not mediator!", mediator.Str())
	}

	// 1. 组装 message
	voting := &vote.VoteInfo{
		VoteType: vote.TYPE_MEDIATOR,
		Contents: mediator.Bytes(),
	}

	msg := &modules.Message{
		App:     modules.APP_VOTE,
		Payload: voting,
	}

	// 2. 组装 tx
	fee := a.dag.CurrentFeeSchedule().VoteMediatorFee
	tx, err := a.dag.CreateBaseTransaction(voter, voter, 0, fee)
	if err != nil {
		return res, err
	}
	tx.TxMessages = append(tx.TxMessages, msg)

	// 3. 签名 tx
	tx, err = a.ptn.SignGenericTransaction(voter, tx)
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
	res.TxHash = tx.Hash()
	res.TxSize = tx.Size()
	res.Warning = defaultResult

	return res, nil
}
