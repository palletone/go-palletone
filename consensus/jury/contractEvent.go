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
 * @author PalletOne core developers <dev@pallet.one>
 * @date 2018
 */
package jury

import (
	"time"

	"github.com/palletone/go-palletone/common/event"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/dag/errors"
	"github.com/palletone/go-palletone/dag/modules"
)

func (p *Processor) SubscribeContractEvent(ch chan<- ContractEvent) event.Subscription {
	return p.contractExecScope.Track(p.contractExecFeed.Subscribe(ch))
}

func (p *Processor) ProcessContractEvent(event *ContractEvent) error {
	if event == nil || len(event.Tx.TxMessages) < 1 {
		return errors.New("ProcessContractEvent param is nil")
	}
	//if !p.checkTxIsExist(event.Tx) {
	//	return errors.New("ProcessContractEvent event Tx is exist")
	//}
	if !p.checkTxValid(event.Tx) {
		return errors.New("ProcessContractEvent event Tx is invalid")
	}
	if !p.contractEventExecutable(event.CType, event.Tx, event.Ele) {
		log.Debug("ProcessContractEvent", "contractEventExecutable is false, reqId", event.Tx.RequestHash())
		return nil
	}
	log.Debug("ProcessContractEvent", " reqId", event.Tx.RequestHash(), "event", event.CType)

	switch event.CType {
	case CONTRACT_EVENT_EXEC:
		return p.contractExecEvent(event.Tx, event.Ele)
	case CONTRACT_EVENT_SIG:
		return p.contractSigEvent(event.Tx, event.Ele)
	case CONTRACT_EVENT_COMMIT:
		return p.contractCommitEvent(event.Tx)
	}
	return nil
}

func (p *Processor) contractExecEvent(tx *modules.Transaction, ele []ElectionInf) error {
	reqId := tx.RequestHash()
	if _, ok := p.mtx[reqId]; ok {
		return nil
	}
	log.Debug("contractExecEvent", "enter, tx req id", reqId)

	p.locker.Lock()
	p.mtx[reqId] = &contractTx{
		reqTx:  tx.GetRequestTx(),
		rstTx:  nil,
		eleInf: ele,
		tm:     time.Now(),
		valid:  true,
		adaInf: make(map[uint32]*AdapterInf),
	}
	p.locker.Unlock()
	log.Debug("contractExecEvent", "add tx req id", reqId)

	if !tx.IsSystemContract() { //系统合约在UNIT构建前执行
		go p.runContractReq(reqId)
	}
	//broadcast contract request transaction event
	event := &ContractEvent{
		Ele:   ele,
		CType: CONTRACT_EVENT_EXEC,
		Tx:    tx,
	}
	go p.ptn.ContractBroadcast(*event, false)
	return nil
}

func (p *Processor) contractSigEvent(tx *modules.Transaction, ele []ElectionInf) error {
	reqId := tx.RequestHash()
	if _, ok := p.mtx[reqId]; !ok {
		log.Debug("contractSigEvent", "local not find reqId,create it", reqId.String())
		p.locker.Lock()
		p.mtx[reqId] = &contractTx{
			reqTx:  tx.GetRequestTx(), 
			eleInf: ele,
			tm:     time.Now(),
			valid:  true,
			adaInf: make(map[uint32]*AdapterInf),
		}
		p.mtx[reqId].rcvTx = append(p.mtx[reqId].rcvTx, tx)
		p.locker.Unlock()

		go p.runContractReq(reqId)
		return nil
	}
	ctx := p.mtx[reqId]
	ctx.rcvTx = append(ctx.rcvTx, tx)
	//如果是jury，将接收到tx与本地执行后的tx进行对比，相同则添加签名到sigTx，如果满足三个签名且签名值最小则广播tx，否则函数返回

	if ok, err := checkAndAddTxSigMsgData(ctx.sigTx, tx); err == nil && ok {
		if getTxSigNum(ctx.sigTx) >= p.contractSigNum {
			if localIsMinSignature(ctx.sigTx) {
				go p.ptn.ContractBroadcast(ContractEvent{Ele: ele, CType: CONTRACT_EVENT_COMMIT, Tx: ctx.sigTx}, true)
			}
		}
	} else if err != nil {
		return err
	}
	return nil
}

func (p *Processor) contractCommitEvent(tx *modules.Transaction) error {
	reqId := tx.RequestHash()
	//todo
	//合约安装，检查合约签名数据
	//用户合约，检查签名数量及有效性
	p.locker.Lock()
	defer p.locker.Unlock()
	if _, ok := p.mtx[reqId]; !ok {
		log.Debug("contractCommitEvent", "local not find reqId,create it", reqId.String())
		p.mtx[reqId] = &contractTx{
			reqTx:  tx.GetRequestTx(),
			tm:     time.Now(),
			valid:  true,
			adaInf: make(map[uint32]*AdapterInf),
		}
	} else if p.mtx[reqId].rstTx != nil {
		log.Info("contractCommitEvent", "rstTx already receive,reqId", reqId)
		return nil //rstTx already receive
	}
	p.mtx[reqId].valid = true
	p.mtx[reqId].rstTx = tx

	return nil
}
