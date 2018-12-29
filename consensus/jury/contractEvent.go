package jury

import (
	"fmt"
	"time"

	"github.com/palletone/go-palletone/dag/errors"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/event"
	"github.com/palletone/go-palletone/core/gen"
)

func (p *Processor) SubscribeContractEvent(ch chan<- ContractExeEvent) event.Subscription {
	return p.contractExecScope.Track(p.contractExecFeed.Subscribe(ch))
}

func (p *Processor) SubscribeContractSigEvent(ch chan<- ContractSigEvent) event.Subscription {
	return p.contractSigScope.Track(p.contractSigFeed.Subscribe(ch))
}

func (p *Processor) ProcessContractEvent(event *ContractExeEvent) error {
	if event == nil {
		return errors.New("ProcessContractEvent param is nil")
	}
	reqId := event.Tx.RequestHash()
	if _, ok := p.mtx[reqId]; ok {
		return nil
	}
	log.Debug("ProcessContractEvent", "enter, tx req id", reqId)

	if false == checkTxValid(event.Tx) {
		return errors.New(fmt.Sprintf("ProcessContractEvent recv event Tx is invalid, txid:%s", reqId.String()))
	}
	execBool := p.nodeContractExecutable(p.local, event.Tx)
	p.locker.Lock()
	p.mtx[reqId] = &contractTx{
		reqTx:      event.Tx,
		rstTx:      nil,
		tm:         time.Now(),
		valid:      true,
		executable: execBool, //todo
	}
	p.locker.Unlock()
	log.Debug("ProcessContractEvent", "add tx req id ", reqId)
	ctx := p.mtx[reqId]
	if ctx.executable {
		go p.runContractReq(ctx)
	}
	//broadcast contract request transaction event
	go p.ptn.ContractBroadcast(*event)
	return nil
}

func (p *Processor) ProcessContractSigEvent(event *ContractSigEvent) error {
	if event == nil || len(event.Tx.TxMessages) < 1 {
		return errors.New("ProcessContractSigEvent param is nil")
	}
	if false == checkTxValid(event.Tx) {
		return errors.New("ProcessContractSigEvent event Tx is invalid")
	}
	reqId := event.Tx.RequestHash()
	if _, ok := p.mtx[reqId]; !ok {
		log.Debug("ProcessContractSigEvent", "local not find reqId,create it", reqId.String())
		exec := p.nodeContractExecutable(p.local, event.Tx)
		p.locker.Lock()
		p.mtx[reqId] = &contractTx{
			reqTx:      event.Tx, //todo 只截取请求部分
			tm:         time.Now(),
			valid:      true,
			executable: exec, //default
		}
		ctx := p.mtx[reqId]
		ctx.rcvTx = append(ctx.rcvTx, event.Tx)
		p.locker.Unlock()

		if ctx.executable == true {
			go p.runContractReq(ctx)
		}
		return nil
	}
	ctx := p.mtx[reqId]

	//如果是mediator，检查接收到的签名个数是否达到3个，如果3个，添加到rstTx，否则函数返回
	//如果是jury，将接收到tx与本地执行后的tx进行对比，相同则添加签名到sigTx，如果满足三个签名且签名值最小则广播tx，否则函数返回
	nodes, err := p.getLocalNodesInfo()
	if err == nil || len(nodes) < 1 {
		return errors.New("ProcessContractSigEvent getLocalNodesInfo fail")
	}
	node := nodes[0] //todo mult node
	if node.ntype == TMediator /*node.ntype& TMediator != 0*/ { //mediator
		if getTxSigNum(event.Tx) >= CONTRACT_SIG_NUM {
			ctx.rstTx = event.Tx
		}
	} else if node.ntype == TJury /*node.ntype&TJury != 0*/ { //jury
		if ok, err := checkAndAddTxData(ctx.sigTx, event.Tx); err == nil && ok {
			//获取签名数量，计算hash值是否最小，如果是则广播交易给Mediator,这里采用相同的签名广播接口，即ContractSigMsg
			if getTxSigNum(ctx.sigTx) >= CONTRACT_SIG_NUM {
				//计算hash值是否最小，如果最小则广播该交易
				if localIsMinSigure(ctx.sigTx) == true {
					go p.ptn.ContractSigBroadcast(ContractSigEvent{Tx: ctx.sigTx})
				}
			}
		}
	} else { //default
		log.Info("ProcessContractSigEvent this node don't care this ContractSigEvent")
		return nil
	}

	return errors.New(fmt.Sprintf("ProcessContractSigEvent err with tx id:%s", reqId.String()))
}

func (p *Processor) ProcessContractSpecialEvent(event *ContractSpecialEvent) error {
	if event == nil || len(event.Tx.TxMessages) < 1 {
		return errors.New("ProcessContractInstallEvent param is nil")
	}
	if false == checkTxValid(event.Tx) {
		return errors.New("ProcessContractInstallEvent event Tx is invalid")
	}
	exec := p.nodeContractExecutable(p.local, event.Tx)
	if !exec {
		log.Debug("ProcessContractSpecialEvent", "nodeContractExecutable is false, reqId", event.Tx.RequestHash())
		return nil
	}
	reqId := event.Tx.RequestHash()
	if _, ok := p.mtx[reqId]; !ok {
		log.Debug("ProcessContractInstallEvent", "local not find reqId, create it", reqId.String())

		account := p.getLocalAccount()
		if account == nil{
			return errors.New("ProcessContractInstallEvent no local account")
		}
		tx , err := gen.GenContractSigTransction(account.Address, account.Password, event.Tx, p.ptn.GetKeyStore())
		if err != nil{
			return err
		}
		p.locker.Lock()
		p.mtx[reqId] = &contractTx{
			reqTx:      tx, //todo 只截取请求部分
			rstTx:      tx,
			tm:         time.Now(),
			valid:      true,
			executable: exec, //default
		}
		p.locker.Unlock()
	}

	return nil
}
