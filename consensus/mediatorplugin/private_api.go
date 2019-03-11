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
	dagcom "github.com/palletone/go-palletone/dag/common"
	"github.com/palletone/go-palletone/dag/modules"
)

const DefaultResult = "Transaction executed locally, but may not be confirmed by the network yet!"

type PrivateMediatorAPI struct {
	*MediatorPlugin
}

func NewPrivateMediatorAPI(mp *MediatorPlugin) *PrivateMediatorAPI {
	return &PrivateMediatorAPI{mp}
}

func (a *PrivateMediatorAPI) StartProduce() {
	if !a.producingEnabled {
		a.producingEnabled = true
		go a.ScheduleProductionLoop()
	}
}

func (a *PrivateMediatorAPI) StopProduce() {
	if a.producingEnabled {
		a.producingEnabled = false
		go func() {
			a.stopProduce <- struct{}{}
		}()
	}
}

// 交易执行结果
type TxExecuteResult struct {
	TxContent string      `json:"txContent"`
	TxHash    common.Hash `json:"txHash"`
	TxSize    string      `json:"txSize"`
	TxFee     string      `json:"txFee"`
	Tip       string      `json:"tip"`
	Warning   string      `json:"warning"`
}

// 创建 mediator 所需的参数, 至少包含普通账户地址
type MediatorCreateArgs struct {
	*modules.MediatorCreateOperation
}

// 相关参数检查
func (args *MediatorCreateArgs) setDefaults(node *discover.Node) (initPrivKey string) {
	if args.InitPubKey == "" {
		initPrivKey, args.InitPubKey = core.CreateInitDKS()
	}

	if args.Node == "" {
		args.Node = node.String()
	}

	return
}

func (a *PrivateMediatorAPI) Create(args MediatorCreateArgs) (*TxExecuteResult, error) {
	// 参数补全
	initPrivKey := args.setDefaults(a.srvr.Self())

	// 参数验证
	err := args.Validate()
	if err != nil {
		return nil, err
	}

	// 判断本节点是否同步完成，数据是否最新
	if !a.dag.IsSynced() {
		return nil, fmt.Errorf("the data of this node is not synced, " +
			"and mediator cannot be created at present")
	}

	addr := args.FeePayer()
	// 判断是否已经是mediator
	if a.dag.IsMediator(addr) {
		return nil, fmt.Errorf("account %v is already a mediator", args.AddStr)
	}

	// 判断是否申请通过
	if !dagcom.MediatorCreateEvaluate(args.MediatorCreateOperation) {
		return nil, fmt.Errorf("has not successfully paid the deposit")
	}

	// 1. 创建交易
	tx, fee, err := a.dag.GenMediatorCreateTx(addr, args.MediatorCreateOperation, a.ptn.TxPool())
	if err != nil {
		return nil, err
	}

	// 2. 签名和发送交易
	err = a.ptn.SignAndSendTransaction(addr, tx)
	if err != nil {
		return nil, err
	}

	// 5. 返回执行结果
	res := &TxExecuteResult{}
	res.TxContent = fmt.Sprintf("Create mediator %s with initPubKey : %s , node: %s , url: %s",
		args.AddStr, args.InitPubKey, args.Node, args.Url)
	res.TxHash = tx.Hash()
	res.TxSize = tx.Size().TerminalString()
	res.TxFee = fmt.Sprintf("%vdao", fee)
	res.Warning = DefaultResult

	if initPrivKey != "" {
		res.Tip = "Your initial private key is: " + initPrivKey + " , initial public key is: " +
			args.InitPubKey + " , please keep in mind!"
	}

	return res, nil
}

func (a *PrivateMediatorAPI) Vote(voterStr, mediatorStr string) (*TxExecuteResult, error) {
	// 参数检查
	voter, err := common.StringToAddress(voterStr)
	if err != nil {
		return nil, fmt.Errorf("invalid account address: %s", voterStr)
	}

	mediator, err := common.StringToAddress(mediatorStr)
	if err != nil {
		return nil, fmt.Errorf("invalid account address: %s", mediatorStr)
	}

	// 判断本节点是否同步完成，数据是否最新
	if !a.dag.IsSynced() {
		return nil, fmt.Errorf("the data of this node is not synced, and can't vote now")
	}

	// 判断是否是mediator
	if !a.dag.IsMediator(mediator) {
		return nil, fmt.Errorf("%v is not mediator", mediatorStr)
	}

	// 判断是否已经投过该mediator
	voted := a.dag.GetAccountInfo(voter).VotedMediators
	if voted[mediator] {
		return nil, fmt.Errorf("account %v was already voting for mediator %v", voterStr, mediatorStr)
	}

	// 1. 创建交易
	tx, fee, err := a.dag.GenVoteMediatorTx(voter, mediator, a.ptn.TxPool())
	if err != nil {
		return nil, err
	}

	// 2. 签名和发送交易
	err = a.ptn.SignAndSendTransaction(voter, tx)
	if err != nil {
		return nil, err
	}

	// 5. 返回执行结果
	res := &TxExecuteResult{}
	res.TxContent = fmt.Sprintf("Account %s vote mediator %s", voterStr, mediatorStr)
	res.TxHash = tx.Hash()
	res.TxSize = tx.Size().TerminalString()
	res.TxFee = fmt.Sprintf("%vdao", fee)
	res.Warning = DefaultResult

	return res, nil
}

func (a *PrivateMediatorAPI) SetDesiredMediatorCount(accountStr string,
	desiredMediatorCount uint8) (*TxExecuteResult, error) {
	// 参数检查
	account, err := common.StringToAddress(accountStr)
	if err != nil {
		return nil, fmt.Errorf("invalid account address: %s", accountStr)
	}

	maxMediatorCount := a.dag.GetChainParameters().MaximumMediatorCount
	if desiredMediatorCount > maxMediatorCount {
		return nil, fmt.Errorf("the max number of allowed active mediators is: %s", maxMediatorCount)
	}

	// 判断本节点是否同步完成，数据是否最新
	if !a.dag.IsSynced() {
		return nil, fmt.Errorf("the data of this node is not synced, and can't vote now")
	}

	// 判断账户是否已经设置此数量
	ai := a.dag.GetAccountInfo(account)
	if ai.DesiredMediatorCount == desiredMediatorCount {
		return nil, fmt.Errorf("account %v was already setting desired mediator count %v",
			accountStr, desiredMediatorCount)
	}

	// 1. 创建交易
	tx, fee, err := a.dag.GenSetDesiredMediatorCountTx(account, desiredMediatorCount, a.ptn.TxPool())
	if err != nil {
		return nil, err
	}

	// 2. 签名和发送交易
	err = a.ptn.SignAndSendTransaction(account, tx)
	if err != nil {
		return nil, err
	}

	// 5. 返回执行结果
	res := &TxExecuteResult{}
	res.TxContent = fmt.Sprintf("Account %s set desired mediator count %v", accountStr, desiredMediatorCount)
	res.TxHash = tx.Hash()
	res.TxSize = tx.Size().TerminalString()
	res.TxFee = fmt.Sprintf("%vdao", fee)
	res.Warning = DefaultResult

	return res, nil
}
