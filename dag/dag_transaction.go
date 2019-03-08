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
	"github.com/palletone/go-palletone/dag/txspool"
	"github.com/palletone/go-palletone/dag/vote"
	"github.com/palletone/go-palletone/tokenengine"
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

func (dag *Dag) createBaseTransaction(from, to common.Address, daoAmount, daoFee uint64, txPool txspool.ITxPool) (*modules.Transaction, error) {
	// 条件判断
	if daoFee == 0 {
		return nil, fmt.Errorf("transaction fee cannot be 0")
	}

	daoTotal := daoAmount + daoFee
	dbBalance := dag.GetPtnBalance(from)
	if daoTotal > dbBalance {
		return nil, fmt.Errorf("the ptn balance of the account is %v not enough %v", dbBalance, daoTotal)
	}

	// 1. 获取转出账户所有的PTN utxo
	//allUtxos, err := dag.GetAddrUtxos(from)
	coreUtxos, err := dag.getAddrCoreUtxos(from, txPool)
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

	selUtxos, change, err := core.Select_utxo_Greedy(greedyUtxos, daoTotal)
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

func (dag *Dag) createTokenTransaction(from, to, toToken common.Address, daoAmount, daoFee, daoAmountToken uint64,
	assetToken string, txPool txspool.ITxPool) (*modules.Transaction, error) {
	// 条件判断
	if daoFee == 0 {
		return nil, fmt.Errorf("transaction fee cannot be 0")
	}

	daoTotal := daoAmount + daoFee
	if daoTotal > dag.GetPtnBalance(from) {
		return nil, fmt.Errorf("the ptn balance of the account is not enough %v", daoTotal)
	}

	// 1. 获取转出账户所有的PTN utxo
	//allUtxos, err := dag.GetAddrUtxos(from)
	coreUtxos, tokenUtxos, err := dag.getAddrCoreUtxosToken(from, assetToken, txPool)
	if err != nil {
		return nil, err
	}

	if len(coreUtxos) == 0 {
		return nil, fmt.Errorf("%v 's uxto is null", from.Str())
	}
	if len(tokenUtxos) == 0 {
		return nil, fmt.Errorf("%v 's  uxto of this Token is null", from.Str())
	}
	//2. 获取 PaymentPayload
	ploadPTN, err := getPayload(from, to, daoAmount, daoFee, coreUtxos)
	if err != nil {
		return nil, err
	}
	ploadToken, err := getPayload(from, toToken, daoAmountToken, 0, tokenUtxos)
	if err != nil {
		return nil, err
	}
	// 3. 构建Transaction
	tx := &modules.Transaction{
		TxMessages: make([]*modules.Message, 0),
	}
	tx.TxMessages = append(tx.TxMessages, modules.NewMessage(modules.APP_PAYMENT, ploadPTN))
	tx.TxMessages = append(tx.TxMessages, modules.NewMessage(modules.APP_PAYMENT, ploadToken))

	return tx, nil
}

func getPayload(from, to common.Address, daoAmount, daoFee uint64, utxos map[modules.OutPoint]*modules.Utxo) (*modules.PaymentPayload, error) {
	// 1. 利用贪心算法得到指定额度的utxo集合
	greedyUtxos := core.Utxos{}
	for outPoint, utxo := range utxos {
		tg := newTxo4Greedy(outPoint, utxo.Amount)
		greedyUtxos = append(greedyUtxos, tg)
	}

	daoTotal := daoAmount + daoFee
	selUtxos, change, err := core.Select_utxo_Greedy(greedyUtxos, daoTotal)
	if err != nil {
		return nil, fmt.Errorf("select utxo err")
	}

	// 2. 构建PaymentPayload的Inputs
	pload := new(modules.PaymentPayload)
	pload.LockTime = 0

	for _, selTxo := range selUtxos {
		tg := selTxo.(*Txo4Greedy)
		txInput := modules.NewTxIn(&tg.OutPoint, []byte{})
		pload.AddTxIn(txInput)
	}

	// 3. 构建PaymentPayload的Outputs
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

	var asset modules.Asset
	for _, utxo := range utxos {
		if utxo != nil {
			asset.AssetId = utxo.Asset.AssetId
			asset.UniqueId = utxo.Asset.UniqueId
			break
		}
	}

	for _, outAmount := range outAmounts {
		pkScript := tokenengine.GenerateLockScript(outAmount.addr)
		txOut := modules.NewTxOut(outAmount.amount, pkScript, &asset)
		pload.AddTxOut(txOut)
	}
	return pload, nil
}

func (dag *Dag) getAddrCoreUtxos(addr common.Address, txPool txspool.ITxPool) (map[modules.OutPoint]*modules.Utxo, error) {
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

		if ok, _ := txPool.OutPointIsSpend(&outPoint); ok {
			continue
		}

		coreUtxos[outPoint] = utxo
	}

	return coreUtxos, nil
}

func (dag *Dag) getAddrCoreUtxosToken(addr common.Address, assetToken string, txPool txspool.ITxPool) (map[modules.OutPoint]*modules.Utxo, map[modules.OutPoint]*modules.Utxo, error) {
	// todo 待优化
	allTxos, err := dag.GetAddrUtxos(addr)
	if err != nil {
		return nil, nil, err
	}

	coreUtxos := make(map[modules.OutPoint]*modules.Utxo, len(allTxos))
	tokenUtxos := make(map[modules.OutPoint]*modules.Utxo, len(allTxos))
	for outPoint, utxo := range allTxos {
		// 剔除非PTN资产
		isPTN := true
		if !utxo.Asset.IsSimilar(modules.CoreAsset) {
			if utxo.Asset.String() != assetToken {
				continue
			}
			isPTN = false
		}

		// 剔除已花费的TXO
		if utxo.IsSpent() {
			continue
		}

		if ok, _ := txPool.OutPointIsSpend(&outPoint); ok {
			continue
		}

		if isPTN {
			coreUtxos[outPoint] = utxo
		} else {
			tokenUtxos[outPoint] = utxo
		}
	}

	return coreUtxos, tokenUtxos, nil
}

func (dag *Dag) calculateDataFee(data interface{}) uint64 {
	size := float64(modules.CalcDateSize(data))
	pricePerKByte := dag.CurrentFeeSchedule().TransferFee.PricePerKByte

	return uint64(size * float64(pricePerKByte) / 1024)
}

func (dag *Dag) CreateGenericTransaction(from, to common.Address, daoAmount, daoFee uint64,
	msg *modules.Message, txPool txspool.ITxPool) (*modules.Transaction, uint64, error) {
	// 如果是 text，则增加费用，以防止用户任意增加文本，导致网络负担加重
	if msg.App == modules.APP_DATA {
		daoFee += dag.calculateDataFee(msg.Payload)
	}
	tx, err := dag.createBaseTransaction(from, to, daoAmount, daoFee, txPool)
	if err != nil {
		return nil, 0, err
	}
	tx.AddMessage(msg)

	return tx, daoFee, nil
}

func (dag *Dag) CreateTokenTransaction(from, to, toToken common.Address, daoAmount, daoFee uint64, daoAmountToken uint64,
	assetToken string, msg *modules.Message, txPool txspool.ITxPool) (*modules.Transaction, uint64, error) {
	// 如果是 text，则增加费用，以防止用户任意增加文本，导致网络负担加重
	if msg.App == modules.APP_DATA {
		daoFee += dag.calculateDataFee(msg.Payload)
	}
	tx, err := dag.createTokenTransaction(from, to, toToken, daoAmount, daoFee, daoAmountToken, assetToken, txPool)
	if err != nil {
		return nil, 0, err
	}
	tx.AddMessage(msg)

	return tx, daoFee, nil
}

func (dag *Dag) GenMediatorCreateTx(account common.Address,
	op *modules.MediatorCreateOperation, txPool txspool.ITxPool) (*modules.Transaction, uint64, error) {
	// 1. 组装 message
	msg := &modules.Message{
		App:     modules.OP_MEDIATOR_CREATE,
		Payload: op,
	}

	// 2. 组装 tx
	fee := dag.CurrentFeeSchedule().MediatorCreateFee
	tx, fee, err := dag.CreateGenericTransaction(account, account, 0, fee, msg, txPool)
	if err != nil {
		return nil, 0, err
	}

	return tx, fee, nil
}

func (dag *Dag) GenVoteMediatorTx(voter, mediator common.Address, txPool txspool.ITxPool) (*modules.Transaction, uint64, error) {
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
	tx, fee, err := dag.CreateGenericTransaction(voter, voter, 0, fee, msg, txPool)
	if err != nil {
		return nil, 0, err
	}

	return tx, fee, nil
}

func (dag *Dag) GenSetDesiredMediatorCountTx(account common.Address, desiredMediatorCount uint8,
	txPool txspool.ITxPool) (*modules.Transaction, uint64, error) {
	return nil, 0, nil
}

func (dag *Dag) GenTransferPtnTx(from, to common.Address, daoAmount uint64, text *string,
	txPool txspool.ITxPool) (*modules.Transaction, uint64, error) {
	fee := dag.CurrentFeeSchedule().TransferFee.BaseFee
	var tx *modules.Transaction
	var err error

	// 如果没有文本，或者文本为空
	if text == nil || *text == "" {
		tx, err = dag.createBaseTransaction(from, to, daoAmount, fee, txPool)
	} else {
		// 1. 组装 message
		msg := &modules.Message{
			App:     modules.APP_DATA,
			Payload: &modules.DataPayload{MainData: []byte(*text)},
		}

		// 2. 创建 tx
		tx, fee, err = dag.CreateGenericTransaction(from, to, daoAmount, fee, msg, txPool)
	}

	if err != nil {
		return nil, 0, err
	}

	return tx, fee, nil
}
