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
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/event"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/dag/errors"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/rwset"
	"github.com/palletone/go-palletone/validator"
	"github.com/palletone/go-palletone/dag/dboperation"
)

func (p *Processor) SubscribeContractEvent(ch chan<- ContractEvent) event.Subscription {
	return p.contractExecScope.Track(p.contractExecFeed.Subscribe(ch))
}

func (p *Processor) ProcessContractEvent(event *ContractEvent) (bool, error) {
	if event == nil || event.Tx == nil || len(event.Tx.Messages()) < 1 {
		return false, errors.New("ProcessContractEvent param is nil")
	}
	var err error
	brd := false
	tx := event.Tx
	reqId := tx.RequestHash()
	if p.checkTxIsExist(tx) {
		return false, fmt.Errorf("[%s]ProcessContractEvent, event Tx is exist, txId:%s",
			reqId.ShortStr(), tx.Hash().String())
	}
	if p.checkTxReqIdIsExist(reqId) {
		return false, fmt.Errorf("[%s]ProcessContractEvent, event Tx reqId is exist, txId:%s",
			reqId.ShortStr(), tx.Hash().String())
	}
	if !tx.IsOnlyContractRequest() && modules.APP_CONTRACT_INVOKE_REQUEST != tx.GetContractTxType() { //!tx.IsSystemContract()
		if _, v, err := p.validator.ValidateTx(tx, false); v != validator.TxValidationCode_VALID {
			return false, fmt.Errorf("[%s]ProcessContractEvent, event Tx is invalid, txId:%s, err:%s",
				reqId.ShortStr(), tx.Hash().String(), err.Error())
		}
		if !p.checkTxAddrValid(tx) {
			return true, fmt.Errorf("[%s]ProcessContractEvent, event Tx addr is invalid, txId:%s",
				reqId.ShortStr(), tx.Hash().String())
		}
	}
	if !p.contractEventExecutable(event.CType, tx, event.Ele) {
		log.Debugf("[%s]ProcessContractEvent, contractEventExecutable is false", reqId.ShortStr())
		return true, nil
	}

	log.Debugf("[%s]ProcessContractEvent, event type:%v ", reqId.ShortStr(), event.CType)
	switch event.CType {
	case CONTRACT_EVENT_EXEC:
		brd, err = p.contractExecEvent(tx, event.Ele)
	case CONTRACT_EVENT_SIG:
		brd, err = p.contractSigEvent(tx, event.Ele)
	case CONTRACT_EVENT_COMMIT:
		brd, err = p.contractCommitEvent(tx)
	case CONTRACT_EVENT_ELE:
		return true, p.contractEleEvent(tx)
	}
	return brd, err
}

func (p *Processor) ProcessUserContractInvokeReqTx(tx *modules.Transaction) {
	if tx == nil {
		log.Error("ProcessUserContractInvokeReqTx, tx is nil")
		return
	}
	reqId := tx.RequestHash()
	log.Debugf("[%s]ProcessUserContractInvokeReqTx enter0", reqId.ShortStr())

	//检查是否为用户合约的请求交易
	if !tx.IsContractTx() || tx.IsSystemContract() || tx.GetContractTxType() != modules.APP_CONTRACT_INVOKE_REQUEST ||
		!tx.IsOnlyContractRequest() {
		return
	}
	log.Debugf("[%s]ProcessUserContractInvokeReqTx enter", reqId.ShortStr())
	var ele *modules.ElectionNode
	ele, err := p.dag.GetContractJury(tx.GetContractId())
	if err != nil {
		log.Errorf("[%s]ProcessUserContractInvokeReqTx GetContractJury err:%s", reqId.ShortStr(), err.Error())
	}
	event := &ContractEvent{
		Ele:   ele,
		CType: CONTRACT_EVENT_EXEC,
		Tx:    tx,
	}

	//todo del
	//sAddr := []string{"P1RS8EfWPxzQMcmjFJ1H7WBGy58FsdAdDF", "P184RUiG5VdY3Y8YUxTmrdsV92MbYQgaPpP", "P1PLs3Cr9Sk8KCV6YfoTTBXRmgMY628SFja"}
	//ele = &modules.ElectionNode{
	//	JuryCount: 3,
	//	EleList:   make([]modules.ElectionInf, 0),
	//}
	//for _, addr := range sAddr {
	//	h := util.RlpHash(addr)
	//	elf := modules.ElectionInf{
	//		EType:    1,
	//		AddrHash: h,
	//	}
	//	ele.EleList = append(ele.EleList, elf)
	//}

	_, err = p.ProcessContractEvent(event)
	if err != nil {
		log.Errorf("[%s]ProcessUserContractInvokeReqTx, ProcessContractEvent err:%s", reqId.ShortStr(), err.Error())
	}
}

func (p *Processor) contractEleEvent(tx *modules.Transaction) error {
	p.locker.Lock()
	defer p.locker.Unlock()

	reqId := tx.RequestHash()
	juryCount := uint64(p.dag.JuryCount())
	if _, ok := p.mtx[reqId]; !ok {
		p.mtx[reqId] = &contractTx{
			reqTx:  tx.GetRequestTx(),
			rstTx:  nil,
			valid:  true,
			tm:     time.Now(),
			adaInf: make(map[uint32]*AdapterInf),
		}
	}
	mtx := p.mtx[reqId]
	eels, err := p.getContractAssignElectionList(tx)
	if err != nil {
		return err
	}
	eelsLen := len(eels)
	if eelsLen > 0 {
		eleNode := &modules.ElectionNode{
			EleList: make([]modules.ElectionInf, 0),
		}
		cfgEleNum := getSysCfgContractElectionNum(p.dag)
		if eelsLen >= cfgEleNum {
			eelsLen = cfgEleNum
			log.Debugf("[%s]contractEleEvent election Num ok", reqId.ShortStr())
		}
		eleNode.EleList = eels[:eelsLen]
		eleNode.JuryCount = juryCount
		mtx.eleNode = eleNode
	}
	if _, ok := p.mel[reqId]; !ok {
		p.mel[reqId] = &electionVrf{
			rcvEle: make([]modules.ElectionInf, 0),
			sigs:   make([]modules.SignatureSet, 0),
			tm:     time.Now(),
		}
	}
	if eelsLen < getSysCfgContractElectionNum(p.dag) {
		reqEvent := &ElectionRequestEvent{
			ReqId:     reqId,
			JuryCount: juryCount,
		}
		go p.ptn.ElectionBroadcast(ElectionEvent{EType: ELECTION_EVENT_VRF_REQUEST, Event: reqEvent}, true) //todo true
	}
	return nil
}

//处理TxMsg中的Tx
//install:接收请求交易，由commit事件进行处理
//deploy:接收请求交易，需要进行election
//invoke:执行请求交易，异步执行，将结果记录到rst
//stop:执行请求交易，异步执行，将结果记录rst ???
func (p *Processor) contractTxExec(tx *modules.Transaction, rw rwset.TxManager, dag dboperation.IContractDag, ele *modules.ElectionNode) (*modules.Transaction, error) {
	if tx == nil {
		return nil, errors.New("contractTxExec, tx is nil")
	}
	reqId := tx.RequestHash()
	p.locker.Lock()
	if p.mtx[reqId] == nil {
		p.mtx[reqId] = &contractTx{
			rstTx:  nil,
			tm:     time.Now(),
			valid:  true,
			adaInf: make(map[uint32]*AdapterInf),
		}
	} else {
		if p.mtx[reqId].reqRcvEd {
			p.locker.Unlock()
			return nil, nil
		}
	}
	p.mtx[reqId].reqTx = tx.GetRequestTx()
	p.mtx[reqId].eleNode = ele
	p.mtx[reqId].reqRcvEd = true

	mtx := p.mtx[reqId]
	//关闭mel
	if e, ok := p.mel[reqId]; ok {
		e.invalid = true
	}
	p.locker.Unlock()

	log.Debugf("[%s]contractTxExec, add tx reqId:%s", reqId.ShortStr(), reqId.String())
	if tx.GetContractTxType() != modules.APP_CONTRACT_INVOKE_REQUEST {
		return tx, nil
	}

	//if !tx.IsSystemContract() { //系统合约在UNIT构建前执行
	//	go p.runContractReq(reqId, ele, rwset.RwM, p.dag)
	//}

	account := p.getLocalJuryAccount()
	if account == nil {
		return nil, fmt.Errorf("[%s]contractTxExec, getLocalJuryAccount is nil", reqId.ShortStr())
	}
	sigTx, err := p.RunAndSignTx(tx, rw, dag, account.Address) //long time ...
	if err != nil {
		log.Errorf("[%s]contractTxExec, RunAndSignTx err:%s", reqId.ShortStr(), err.Error())
		return nil, err
	}

	mtx.sigTx = sigTx
	if len(mtx.rcvTx) > 0 {
		for _, rtx := range mtx.rcvTx {
			ok, err := checkAndAddTxSigMsgData(mtx.sigTx, rtx)
			if err != nil {
				log.Debugf("[%s]contractTxExec, checkAndAddTxSigMsgData error:%s", reqId.ShortStr(), err.Error())
			} else if ok {
				log.Debugf("[%s]contractTxExec, checkAndAddTxSigMsgData ok, tx[%s]", reqId.ShortStr(), rtx.Hash().String())
			} else {
				log.Debugf("[%s]contractTxExec, checkAndAddTxSigMsgData fail", reqId.ShortStr())
			}
		}
	}

	sigNum := getTxSigNum(mtx.sigTx)
	cfgSigNum := getSysCfgContractSignatureNum(p.dag)
	log.Debugf("[%s]contractTxExec sigNum %d, p.contractSigNum %d", reqId.ShortStr(), sigNum, cfgSigNum)
	if sigNum >= cfgSigNum {
		if localIsMinSignature(mtx.sigTx) {
			//签名数量足够，而且当前节点是签名最新的节点，那么合并签名并广播完整交易
			log.Infof("[%s]contractTxExec, localIsMinSignature Ok!", reqId.ShortStr())
			p.processContractPayout(mtx.sigTx, ele)
			go p.ptn.ContractBroadcast(ContractEvent{CType: CONTRACT_EVENT_COMMIT, Ele: ele, Tx: mtx.sigTx}, true)
			return sigTx, nil
		}
	}
	//广播
	go p.ptn.ContractBroadcast(ContractEvent{CType: CONTRACT_EVENT_SIG, Ele: ele, Tx: sigTx}, false)

	return sigTx, nil
}

func (p *Processor) contractExecEvent(tx *modules.Transaction, ele *modules.ElectionNode) (broadcast bool, err error) {
	if tx == nil {
		return false, errors.New("contractExecEvent, tx is nil")
	}
	reqId := tx.RequestHash()
	p.locker.Lock()
	if p.mtx[reqId] == nil {
		p.mtx[reqId] = &contractTx{
			rstTx:  nil,
			tm:     time.Now(),
			valid:  true,
			adaInf: make(map[uint32]*AdapterInf),
		}
	} else {
		if p.mtx[reqId].reqRcvEd {
			p.locker.Unlock()
			return false, nil
		}
	}
	p.mtx[reqId].reqTx = tx.GetRequestTx()
	p.mtx[reqId].eleNode = ele
	p.mtx[reqId].reqRcvEd = true
	//关闭mel
	if e, ok := p.mel[reqId]; ok {
		e.invalid = true
	}
	p.locker.Unlock()
	log.Debugf("[%s]contractExecEvent, add tx reqId:%s", reqId.ShortStr(), reqId.String())

	if !tx.IsSystemContract() { //系统合约在UNIT构建前执行
		go p.runContractReq(reqId, ele, rwset.RwM, p.dag)
	}
	return true, nil
}

func (p *Processor) contractSigEvent(tx *modules.Transaction, ele *modules.ElectionNode) (broadcast bool, err error) {
	if tx == nil {
		return false, errors.New("contractSigEvent, tx is nil")
	}
	p.locker.Lock()
	defer p.locker.Unlock()
	reqId := tx.RequestHash()
	if _, ok := p.mtx[reqId]; ok {
		if checkTxReceived(p.mtx[reqId].rcvTx, tx) {
			return false, nil
		}
	}
	log.Debugf("[%s]contractSigEvent, receive sig tx[%s]", reqId.ShortStr(), tx.Hash().String())
	if _, ok := p.mtx[reqId]; !ok {
		log.Debugf("[%s]contractSigEvent, local not find reqId,create it", reqId.ShortStr())
		p.mtx[reqId] = &contractTx{
			reqTx:   tx.GetRequestTx(),
			eleNode: ele,
			tm:      time.Now(),
			valid:   true,
			adaInf:  make(map[uint32]*AdapterInf),
		}
		p.mtx[reqId].rcvTx = append(p.mtx[reqId].rcvTx, tx)
		return true, nil
	}
	ctx := p.mtx[reqId]
	ctx.rcvTx = append(ctx.rcvTx, tx)

	//如果是jury，将接收到tx与本地执行后的tx进行对比，相同则添加签名到sigTx，如果满足签名数量且签名值最小则广播tx，否则函数返回
	if ok, err := checkAndAddTxSigMsgData(ctx.sigTx, tx); err == nil && ok {
		if getTxSigNum(ctx.sigTx) >= getSysCfgContractSignatureNum(p.dag) {
			if localIsMinSignature(ctx.sigTx) { //todo
				//签名数量足够，而且当前节点是签名最新的节点，那么合并签名并广播完整交易
				log.Infof("[%s]runContractReq, localIsMinSignature Ok!", reqId.ShortStr())
				p.processContractPayout(ctx.sigTx, ele)
				go p.ptn.ContractBroadcast(ContractEvent{CType: CONTRACT_EVENT_COMMIT, Ele: ele, Tx: ctx.sigTx}, true)
			}
		}
	} else if err != nil {
		return true, err
	}
	return true, nil
}

func (p *Processor) contractCommitEvent(tx *modules.Transaction) (broadcast bool, err error) {
	if tx == nil {
		return false, errors.New("contractCommitEvent, tx is nil")
	}
	reqId := tx.RequestHash()
	p.locker.Lock()
	defer p.locker.Unlock()
	if _, ok := p.mtx[reqId]; !ok {
		//log.Debug("contractCommitEvent", "local not find reqId,create it", reqId)
		p.mtx[reqId] = &contractTx{
			reqTx:  tx.GetRequestTx(),
			tm:     time.Now(),
			valid:  true,
			adaInf: make(map[uint32]*AdapterInf),
		}
	} else if p.mtx[reqId].rstTx != nil {
		log.Debugf("[%s]contractCommitEvent, rstTx already receive", reqId.ShortStr())
		return false, nil //rstTx already receive
	}

	//添加到交易池，等待打包
	log.Debug("contractCommitEvent", "tx:", tx)
	tx1 := tx.Clone()
	p.ptn.TxPool().AddLocal(tx1)

	log.Debugf("[%s]contractCommitEvent, rstTx receive", reqId.ShortStr())
	//err = p.dag.SaveTransaction(tx)
	//if err != nil {
	//	log.Errorf("[%s]contractCommitEvent SaveTransaction err:%s", reqId.ShortStr(), err.Error())
	//	return false, err
	//}
	p.mtx[reqId].valid = true
	p.mtx[reqId].rstTx = tx

	return true, nil
}
