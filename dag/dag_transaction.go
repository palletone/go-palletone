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
 * @date 2018
 *
 */

package dag

import (
	"fmt"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/vote"
	"github.com/palletone/go-palletone/ptnjson"
	"github.com/palletone/go-palletone/tokenengine"
	"github.com/shopspring/decimal"
)

type Txo4Greedy struct {
	modules.OutPoint
	Amount uint64
}

func (txo *Txo4Greedy) GetAmount() uint64 {
	return txo.Amount
}

func newTxo4Greedy(outPoint modules.OutPoint, amount uint64) *Txo4Greedy {
	return &Txo4Greedy{
		OutPoint: outPoint,
		Amount:   amount,
	}
}

func (dag *Dag) createBaseTransaction(from, to common.Address, daoAmount, daoFee uint64) (*modules.Transaction, error) {
	if daoFee == 0 {
		return nil, fmt.Errorf("transaction's fee id zero")
	}

	// 1. 获取转出账户所有的PTN utxo
	//allUtxos, err := dag.GetAddrUtxos(from)
	coreUtxos, err := dag.getAddrCoreUtxos(from)
	if err != nil {
		return nil, err
	}

	if len(coreUtxos) == 0 {
		return nil, fmt.Errorf("%v 's uxto is null", from.Str())
	}

	// 2. 利用贪心算法得到指定额度的utxo集合
	greedyUtxos := core.Utxos{}
	for outPoint, utxo := range coreUtxos {
		tg := newTxo4Greedy(outPoint, utxo.Amount)
		greedyUtxos = append(greedyUtxos, tg)
	}

	selUtxos, change, err := core.Select_utxo_Greedy(greedyUtxos, daoAmount+daoFee)
	if err != nil {
		return nil, fmt.Errorf("select utxo err")
	}

	// 3. 构建PaymentPayload的Inputs
	pload := new(modules.PaymentPayload)
	pload.LockTime = 0

	for _, selTxo := range selUtxos {
		tg := selTxo.(*Txo4Greedy)
		txInput := modules.NewTxIn(&tg.OutPoint, []byte{})
		pload.AddTxIn(txInput)
	}

	// 4. 构建PaymentPayload的Outputs
	// 为了保证顺序， 将map改为结构体数组
	type OutAmount struct {
		addr   common.Address
		amount uint64
	}

	outAmounts := make([]*OutAmount, 1, 2)
	outAmount := &OutAmount{to, daoAmount}
	outAmounts[0] = outAmount

	if change > 0 {
		// 处理from和to是同一个地址的特殊情况
		if from.Equal(to) {
			outAmount.amount = outAmount.amount + change
			outAmounts[0] = outAmount
		} else {
			outAmounts = append(outAmounts, &OutAmount{from, change})
		}
	}

	for _, outAmount := range outAmounts {
		pkScript := tokenengine.GenerateLockScript(outAmount.addr)
		txOut := modules.NewTxOut(outAmount.amount, pkScript, modules.CoreAsset)
		pload.AddTxOut(txOut)
	}

	// 5. 构建Transaction
	tx := &modules.Transaction{
		TxMessages: make([]*modules.Message, 0),
	}
	tx.TxMessages = append(tx.TxMessages, modules.NewMessage(modules.APP_PAYMENT, pload))

	return tx, nil
}

func (dag *Dag) getAddrCoreUtxos(addr common.Address) (map[modules.OutPoint]*modules.Utxo, error) {
	// todo 待优化
	allTxos, err := dag.GetAddrUtxos(addr)
	if err != nil {
		return nil, err
	}

	coreUtxos := make(map[modules.OutPoint]*modules.Utxo, len(allTxos))
	for outPoint, utxo := range allTxos {
		// 剔除非PTN资产
		if !utxo.Asset.IsSimilar(modules.CoreAsset) {
			continue
		}

		// 剔除已花费的TXO
		if utxo.IsSpent() {
			continue
		}

		coreUtxos[outPoint] = utxo
	}

	return coreUtxos, nil
}

func (dag *Dag) calculateDataFee(data interface{}) uint64 {
	size := float64(modules.CalcDateSize(data))
	pricePerKByte := dag.CurrentFeeSchedule().TransferFee.PricePerKByte

	return uint64(size * float64(pricePerKByte) / 1024)
}

func (dag *Dag) CreateGenericTransaction(from, to common.Address, daoAmount, daoFee uint64,
	msg *modules.Message) (*modules.Transaction, uint64, error) {
	// 如果是 text，则增加费用，以防止用户任意增加文本，导致网络负担加重
	if msg.App == modules.APP_TEXT {
		daoFee += dag.calculateDataFee(msg.Payload)
	}

	tx, err := dag.createBaseTransaction(from, to, daoAmount, daoFee)
	if err != nil {
		return nil, 0, err
	}

	tx.AddMessage(msg)
	//tx.TxMessages = append(tx.TxMessages, msgs...)
	return tx, 0, nil
}

func (dag *Dag) GenMediatorCreateTx(account common.Address,
	op *modules.MediatorCreateOperation) (*modules.Transaction, uint64, error) {
	// 1. 组装 message
	msg := &modules.Message{
		App:     modules.OP_MEDIATOR_CREATE,
		Payload: op,
	}

	// 2. 组装 tx
	fee := dag.CurrentFeeSchedule().MediatorCreateFee
	tx, fee, err := dag.CreateGenericTransaction(account, account, 0, fee, msg)
	if err != nil {
		return nil, 0, err
	}

	return tx, fee, nil
}

func (dag *Dag) GenVoteMediatorTx(voter, mediator common.Address) (*modules.Transaction, uint64, error) {
	// 1. 组装 message
	voting := &vote.VoteInfo{
		VoteType: vote.TypeMediator,
		Contents: mediator.Bytes21(),
	}

	msg := &modules.Message{
		App:     modules.APP_VOTE,
		Payload: voting,
	}

	// 2. 组装 tx
	fee := dag.CurrentFeeSchedule().VoteMediatorFee
	tx, fee, err := dag.CreateGenericTransaction(voter, voter, 0, fee, msg)
	if err != nil {
		return nil, 0, err
	}

	return tx, fee, nil
}

func (dag *Dag) GenTransferPtnTx(from, to common.Address, amount decimal.Decimal, text string) (*modules.Transaction, uint64, error) {
	// 1. 组装 message
	msg := &modules.Message{
		App:     modules.APP_TEXT,
		Payload: &modules.TextPayload{Text: []byte(text)},
	}

	// 2. 组装 tx
	fee := dag.CurrentFeeSchedule().TransferFee.BaseFee
	tx, fee, err := dag.CreateGenericTransaction(from, to, ptnjson.Ptn2Dao(amount), fee, msg)
	if err != nil {
		return nil, 0, err
	}

	return tx, fee, nil
}
