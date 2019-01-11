package jury

import (
	"time"

	"github.com/palletone/go-palletone/dag/errors"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/event"
	"github.com/palletone/go-palletone/dag/modules"
)

func (p *Processor) SubscribeContractEvent(ch chan<- ContractEvent) event.Subscription {
	return p.contractExecScope.Track(p.contractExecFeed.Subscribe(ch))
}

func (p *Processor) ProcessContractEvent(event *ContractEvent) error {
	if event == nil || len(event.Tx.TxMessages) < 1 {
		return errors.New("ProcessContractEvent param is nil")
	}
	if false == checkTxValid(event.Tx) {
		return errors.New("ProcessContractEvent event Tx is invalid")
	}
	if !p.contractEventExecutable(event.CType, p.local, event.Tx) {
		log.Debug("ProcessContractEvent", "contractEventExecutable is false, reqId", event.Tx.RequestHash())
		return nil
	}
	log.Debug("ProcessContractEvent", " reqId", event.Tx.RequestHash(), "event", event.CType)

	switch event.CType {
	case CONTRACT_EVENT_EXEC:
		return p.contractExecEvent(event.Tx)
	case CONTRACT_EVENT_SIG:
		return p.contractSigEvent(event.Tx)
	case CONTRACT_EVENT_COMMIT:
		return p.contractCommitEvent(event.Tx)
	}
	return nil
}

func (p *Processor) contractExecEvent(tx *modules.Transaction) error {
	reqId := tx.RequestHash()
	if _, ok := p.mtx[reqId]; ok {
		return nil
	}
	log.Debug("contractExecEvent", "enter, tx req id", reqId)

	p.locker.Lock()
	p.mtx[reqId] = &contractTx{
		reqTx: tx,
		rstTx: nil,
		tm:    time.Now(),
		valid: true,
	}
	p.locker.Unlock()
	log.Debug("contractExecEvent", "add tx req id", reqId)

	if !isSystemContract(tx) { //系统合约在UNIT构建前执行
		go p.runContractReq(reqId)
	}

	//broadcast contract request transaction event
	event := &ContractEvent{
		CType: CONTRACT_EVENT_EXEC,
		Tx:    tx,
	}
	go p.ptn.ContractBroadcast(*event)

	return nil
}

func (p *Processor) contractSigEvent(tx *modules.Transaction) error {
	reqId := tx.RequestHash()
	if _, ok := p.mtx[reqId]; !ok {
		log.Debug("contractSigEvent", "local not find reqId,create it", reqId.String())
		p.locker.Lock()
		p.mtx[reqId] = &contractTx{
			reqTx: tx, //todo 只截取请求部分
			tm:    time.Now(),
			valid: true,
		}
		p.mtx[reqId].rcvTx = append(p.mtx[reqId].rcvTx, tx)
		p.locker.Unlock()

		go p.runContractReq(reqId)
		return nil
	}
	ctx := p.mtx[reqId]
	ctx.rcvTx = append(ctx.rcvTx, tx)
	//如果是jury，将接收到tx与本地执行后的tx进行对比，相同则添加签名到sigTx，如果满足三个签名且签名值最小则广播tx，否则函数返回
	if ok, err := checkAndAddTxData(ctx.sigTx, tx); err == nil && ok {
		if getTxSigNum(ctx.sigTx) >= CONTRACT_SIG_NUM {
			if localIsMinSignature(ctx.sigTx) {
				go p.ptn.ContractBroadcast(ContractEvent{CType: CONTRACT_EVENT_COMMIT, Tx: ctx.sigTx})
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
	if _, ok := p.mtx[reqId]; !ok {
		log.Debug("contractCommitEvent", "local not find reqId,create it", reqId.String())
		p.mtx[reqId] = &contractTx{
			reqTx: tx,
			tm:    time.Now(),
			valid: true,
		}
	}
	p.mtx[reqId].valid = true
	p.mtx[reqId].rstTx = tx
	p.locker.Unlock()

	return nil
}
