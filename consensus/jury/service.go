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
	"sync"
	"time"
	"reflect"
	"bytes"

	"github.com/dedis/kyber"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/event"
	"github.com/palletone/go-palletone/dag/errors"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/p2p"
	"github.com/palletone/go-palletone/common/rlp"
	"github.com/palletone/go-palletone/contracts"
	"github.com/palletone/go-palletone/core/accounts/keystore"
	"github.com/palletone/go-palletone/core/gen"
	"github.com/palletone/go-palletone/dag"
	"github.com/palletone/go-palletone/dag/txspool"
	cm "github.com/palletone/go-palletone/dag/common"
)

type PeerType int

const (
	_         PeerType = iota
	TUnknow
	TJury
	TMediator
)

type Juror struct {
	name        string
	address     common.Address
	InitPartPub kyber.Point
}

type PalletOne interface {
	GetKeyStore() *keystore.KeyStore
	TxPool() txspool.ITxPool

	MockContractLocalSend(event ContractExeEvent)
	MockContractSigLocalSend(event ContractSigEvent)

	ContractBroadcast(event ContractExeEvent)
	ContractSigBroadcast(event ContractSigEvent)

	GetLocalMediators() []common.Address
	SignGenericTransaction(from common.Address, tx *modules.Transaction) (*modules.Transaction, error)
}

type iDag interface {
	GetActiveMediators() []common.Address
	IsActiveMediator(add common.Address) bool
	GetAddr1TokenUtxos(addr common.Address, asset *modules.Asset) (map[modules.OutPoint]*modules.Utxo, error)
	CreateBaseTransaction(from, to common.Address, daoAmount, daoFee uint64) (*modules.Transaction, error)
}

type contractTx struct {
	list       []common.Address     //dynamic
	reqTx      *modules.Transaction //request contract
	rstTx      *modules.Transaction //contract run result
	tm         time.Time            //create time
	valid      bool                 //contract request valid identification
	executable bool                 //contract executable,sys on mediator, user on jury
}

type Processor struct {
	name  string
	ptype PeerType
	ptn   PalletOne
	dag   iDag

	local    []common.Address //local account addr(tmp mediator)
	contract *contracts.Contract
	locker   *sync.Mutex
	quit     chan struct{}
	mtx      map[common.Hash]*contractTx

	contractExecFeed  event.Feed
	contractExecScope event.SubscriptionScope
	contractSigFeed   event.Feed
	contractSigScope  event.SubscriptionScope
	idag              dag.IDag
}

func NewContractProcessor(ptn PalletOne, dag iDag, contract *contracts.Contract) (*Processor, error) {
	if ptn == nil || dag == nil {
		return nil, errors.New("NewContractProcessor, param is nil")
	}

	addrs := ptn.GetLocalMediators()
	p := &Processor{
		name:     "conractProcessor",
		ptn:      ptn,
		dag:      dag,
		contract: contract,
		local:    addrs,
		locker:   new(sync.Mutex),
		quit:     make(chan struct{}),
		mtx:      make(map[common.Hash]*contractTx),
	}

	log.Info("NewContractProcessor ok", "local address:", p.local)
	//log.Info("NewContractProcessor", "info:", p.local)
	return p, nil
}

func (p *Processor) Start(server *p2p.Server) error {
	//启动消息接收处理线程

	//合约执行节点更新线程

	//合约定时清理线程
	go p.ContractTxDeleteLoop()

	return nil
}

func (p *Processor) Stop() error {
	close(p.quit)
	log.Debug("contract processor stop")
	return nil
}

func (p *Processor) SubscribeContractEvent(ch chan<- ContractExeEvent) event.Subscription {
	return p.contractExecScope.Track(p.contractExecFeed.Subscribe(ch))
}

func (p *Processor) ProcessContractEvent(event *ContractExeEvent) error {
	if event == nil {
		return errors.New("ProcessContractEvent param is nil")
	}
	reqId := event.Tx.RequestHash()
	if _, ok := p.mtx[reqId]; ok {
		return nil
	}
	log.Debug("ProcessContractEvent", "enter, tx req id ", reqId)

	if false == checkTxValid(event.Tx) {
		return errors.New(fmt.Sprintf("ProcessContractEvent recv event Tx is invalid, txid:%s", reqId.String()))
	}

	execBool := nodeContractExecutable(p.dag, p.local, event.Tx)
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

	if p.mtx[reqId].executable {
		go runContractReq(p.dag, p.contract, p.mtx[reqId])
	}
	//broadcast contract request transaction event
	go p.ptn.ContractBroadcast(*event)
	return nil
}

func runContractReq(dag iDag, contract *contracts.Contract, req *contractTx) error {
	if req == nil {
		return errors.New("runContractReq param is nil")
	}
	_, msgs, err := runContractCmd(dag, contract, req.reqTx)
	if err != nil {
		log.Error("runContractReq runContractCmd", "reqTx", req.reqTx.RequestHash().String(), "error", err.Error())
		return err
	}
	tx, err := gen.GenContractTransction(req.reqTx, msgs)
	if err != nil {
		log.Error("runContractReq GenContractSigTransctions", "error", err.Error())
		return err
	}

	req.rstTx = tx
	return nil
}

func (p *Processor) AddContractLoop(txpool txspool.ITxPool, addr common.Address, ks *keystore.KeyStore) error {
	//log.Debug("ProcessContractEvent", "enter", addr.String())
	for _, ctx := range p.mtx {
		if false == ctx.valid || ctx.rstTx == nil {
			continue
		}
		ctx.valid = false
		if false == checkTxValid(ctx.rstTx) {
			log.Error("AddContractLoop recv event Tx is invalid,", "txid", ctx.rstTx.RequestHash().String())
			continue
		}

		tx, err := gen.GenContractSigTransction(addr, ctx.rstTx, ks)
		if err != nil {
			log.Error("AddContractLoop GenContractSigTransctions", "error", err.Error())
			continue
		}
		log.Debug("AddContractLoop", "tx", tx)
		if err = txpool.AddLocal(txspool.TxtoTxpoolTx(txpool, tx)); err != nil {
			log.Error("AddContractLoop", "error", err.Error())
			continue
		}
	}
	return nil
}

func (p *Processor) CheckContractTxValid(tx *modules.Transaction) bool {
	//检查本地是否存
	if tx == nil {
		log.Error("CheckContractTxValid", "param is nil")
		return false
	}
	reqId := tx.RequestHash()
	log.Debug("CheckContractTxValid", "tx req id ", reqId)

	if false == checkTxValid(tx) {
		log.Error("CheckContractTxValid", "checkTxValid fail")
		return false
	}
	//检查本阶段时候有合约执行权限
	if nodeContractExecutable(p.dag, p.local, tx) != true {
		log.Error("CheckContractTxValid", "nodeContractExecutable false")
		return false
	}

	ctx, ok := p.mtx[reqId]
	if ctx != nil && (ctx.valid == false || ctx.executable == false) {
		return false
	}

	if ok && ctx.rstTx != nil {
		//比较msg
		log.Debug("CheckContractTxValid", "compare txid", reqId)
		return msgsCompare(ctx.rstTx.TxMessages, tx.TxMessages, modules.APP_CONTRACT_INVOKE)
	} else {
		//runContractCmd
		//比较msg
		_, msgs, err := runContractCmd(p.dag, p.contract, tx)
		if err != nil {
			log.Error("CheckContractTxValid runContractCmd", "error", err.Error())
			return false
		}
		p.mtx[reqId].valid = false
		return msgsCompare(msgs, tx.TxMessages, modules.APP_CONTRACT_INVOKE)
	}
}

func (p *Processor) SubscribeContractSigEvent(ch chan<- ContractSigEvent) event.Subscription {
	return p.contractSigScope.Track(p.contractSigFeed.Subscribe(ch))
}

func (p *Processor) ContractTxDeleteLoop() {
	for {
		time.Sleep(time.Second * time.Duration(20))
		p.locker.Lock()
		for k, v := range p.mtx {
			if time.Since(v.tm) > time.Second*100 { //todo
				if v.valid == false {
					log.Info("ContractTxDeleteLoop", "delete tx id", k.String())
					delete(p.mtx, k)
				}
			}
		}
		p.locker.Unlock()
	}
}

//执行合约命令:install、deploy、invoke、stop，同时只支持一种类型
func runContractCmd(dag iDag, contract *contracts.Contract, trs *modules.Transaction) (modules.MessageType, []*modules.Message, error) {
	if trs == nil || len(trs.TxMessages) <= 0 {
		return 0, nil, errors.New("runContractCmd transaction or msg is nil")
	}

	for _, msg := range trs.TxMessages {
		switch msg.App {
		case modules.APP_CONTRACT_TPL_REQUEST:
			{
				return modules.APP_CONTRACT_TPL, nil, errors.New("runContractCmd not support APP_CONTRACT_TPL")
			}
		case modules.APP_CONTRACT_DEPLOY_REQUEST:
			{
				return modules.APP_CONTRACT_DEPLOY, nil, errors.New("runContractCmd not support APP_CONTRACT_DEPLOY")
			}
		case modules.APP_CONTRACT_INVOKE_REQUEST:
			{
				msgs := []*modules.Message{}
				req := ContractInvokeReq{
					chainID:  "palletone",
					deployId: msg.Payload.(*modules.ContractInvokeRequestPayload).ContractId,
					args:     msg.Payload.(*modules.ContractInvokeRequestPayload).Args,
					txid:     trs.RequestHash().String(),
					tx:       trs,
				}
				invokeResult, err := ContractProcess(contract, req)
				if err != nil {
					log.Error("runContractCmd ContractProcess ", "error", err)
					return msg.App, nil, errors.New(fmt.Sprintf("runContractCmd APP_CONTRACT_INVOKE txid(%s) rans err:%s", req.txid, err))
				}
				result := invokeResult.(*modules.ContractInvokeResult)
				payload := modules.NewContractInvokePayload(result.ContractId, result.FunctionName, result.Args, result.ExecutionTime, result.ReadSet, result.WriteSet, result.Payload)

				if payload != nil {
					msgs = append(msgs, modules.NewMessage(modules.APP_CONTRACT_INVOKE, payload))
				}
				toContractPayments, err := resultToContractPayments(dag, result)
				if err != nil {
					return modules.APP_CONTRACT_INVOKE, nil, err
				}
				if toContractPayments != nil && len(toContractPayments) > 0 {
					for _, contractPayment := range toContractPayments {
						msgs = append(msgs, modules.NewMessage(modules.APP_PAYMENT, contractPayment))
					}
				}
				cs, err := resultToCoinbase(result)
				if err != nil {
					return modules.APP_CONTRACT_INVOKE, nil, err
				}
				if cs != nil && len(cs) > 0 {
					for _, coinbase := range cs {
						msgs = append(msgs, modules.NewMessage(modules.APP_PAYMENT, coinbase))
					}
				}

				return modules.APP_CONTRACT_INVOKE, msgs, nil
			}
		case modules.APP_CONTRACT_STOP_REQUEST:
			{
				return modules.APP_CONTRACT_STOP, nil, errors.New("not support APP_CONTRACT_STOP")
			}
		}
	}

	return 0, nil, errors.New(fmt.Sprintf("runContractCmd err, txid=%s", trs.RequestHash().String()))
}

func checkAndAddTxData(local *modules.Transaction, recv *modules.Transaction) (bool, error) {
	var recvSigMsg *modules.Message

	if len(local.TxMessages) != len(recv.TxMessages) {
		return false, errors.New("tx msg is invalid")
	}
	for i := 0; i < len(local.TxMessages); i++ {
		if recv.TxMessages[i].App == modules.APP_SIGNATURE {
			recvSigMsg = recv.TxMessages[i]
		} else if reflect.DeepEqual(*local.TxMessages[i], *recv.TxMessages[i]) != true {
			return false, errors.New("tx msg is not equal")
		}
	}

	if recvSigMsg == nil {
		return false, errors.New("not find recv sig msg")
	}
	for i, msg := range local.TxMessages {
		if msg.App == modules.APP_SIGNATURE {
			sigPayload := msg.Payload.(*modules.SignaturePayload)
			sigs := sigPayload.Signatures
			for _, sig := range sigs {
				if true == bytes.Equal(sig.PubKey, recvSigMsg.Payload.(*modules.SignaturePayload).Signatures[0].PubKey) &&
					true == bytes.Equal(sig.Signature, recvSigMsg.Payload.(*modules.SignaturePayload).Signatures[0].Signature) {
					log.Info("tx  already recv:", recv.RequestHash().String())
					return false, nil
				}
			}
			//直接将签名添加到msg中
			if len(recvSigMsg.Payload.(*modules.SignaturePayload).Signatures) > 0 {
				sigPayload.Signatures = append(sigs, recvSigMsg.Payload.(*modules.SignaturePayload).Signatures[0])
			}

			local.TxMessages[i].Payload = sigPayload
			//local.TxHash = common.Hash{}
			//local.TxHash = local.Hash()

			log.Info("checkAndAddTxData", "add sig payload:", sigPayload.Signatures)
			return true, nil
		}
	}

	return false, errors.New("")
}

func getTxSigNum(tx *modules.Transaction) int {
	if tx != nil {
		for _, msg := range tx.TxMessages {
			if msg.App == modules.APP_SIGNATURE {
				return len(msg.Payload.(*modules.SignaturePayload).Signatures)
			}
		}
	}
	return 0
}

func checkTxValid(tx *modules.Transaction) bool {
	return cm.ValidateTxSig(tx)
}

func msgsCompare(msgsA []*modules.Message, msgsB []*modules.Message, msgType modules.MessageType) bool {
	if msgsA == nil || msgsB == nil {
		log.Error("msgsCompare", "param is nil")
		return false
	}
	var msg1, msg2 *modules.Message
	for _, v := range msgsA {
		if v.App == msgType {
			msg1 = v
		}
	}
	for _, v := range msgsB {
		if v.App == msgType {
			msg2 = v
		}
	}

	if msg1 != nil && msg2 != nil {
		if reflect.DeepEqual(msg1, msg2) == true {
			log.Debug("msgsCompare", "msg is equal, type", msgType)
			return true
		}
	}
	log.Debug("msgsCompare", "msg is not equal")

	return false
}

func nodeContractExecutable(dag iDag, addrs []common.Address, tx *modules.Transaction) bool {
	if tx == nil {
		return false
	}
	var contractId []byte

	for _, msg := range tx.TxMessages {
		if msg.App == modules.APP_CONTRACT_INVOKE_REQUEST {
			contractId = msg.Payload.(*modules.ContractInvokeRequestPayload).ContractId
			log.Debug("nodeContractExecutable", "contract id", contractId, "len", len(contractId))
			break
		}
	}

	if len(contractId) <= 2 && len(contractId) > 0 { //system contract
		for _, addr := range addrs {
			log.Debug("nodeContractExecutable", "contract id", contractId, "addr", addr.String())
			if true == dag.IsActiveMediator(addr) {
				log.Debug("nodeContractExecutable", "true:contract id", contractId, "addr", addr.String())
				return true
			}
		}
	} else { //usr contract
	}
	log.Debug("nodeContractExecutable", "false:contract id", contractId)

	return false
}

func (p *Processor) addTx2LocalTxTool(tx *modules.Transaction, cnt int) error {
	if tx == nil || cnt < 4 {
		return errors.New(fmt.Sprintf("addTx2LocalTxTool param error, node count is [%d]", cnt))
	}
	if num := getTxSigNum(tx); num < (cnt*2/3 + 1) {
		log.Error("addTx2LocalTxTool sig num is", num)
		return errors.New(fmt.Sprintf("addTx2LocalTxTool tx sig num is:%d", num))
	}

	txPool := p.ptn.TxPool()
	log.Debug("addTx2LocalTxTool", "tx:", tx.Hash().String())

	return txPool.AddLocal(txspool.TxtoTxpoolTx(txPool, tx))
}

func (p *Processor) ContractTxCreat(deployId []byte, txBytes []byte, args [][]byte, timeout time.Duration) (rspPayload []byte, err error) {
	log.Info("ContractTxCreat", fmt.Sprintf("enter, deployId[%v],", deployId))

	if deployId == nil || args == nil {
		log.Error("ContractTxCreat", "param is nil")
		return nil, errors.New("transaction request param is nil")
	}

	tx := &modules.Transaction{}
	if txBytes != nil {
		if err := rlp.DecodeBytes(txBytes, tx); err != nil {
			return nil, err
		}
	} else {
		pay := &modules.PaymentPayload{
			Inputs:   []*modules.Input{},
			Outputs:  []*modules.Output{},
			LockTime: 11111, //todo
		}
		msgPay := &modules.Message{
			App:     modules.APP_PAYMENT,
			Payload: pay,
		}
		tx.AddMessage(msgPay)
	}

	msgReq := &modules.Message{
		App: modules.APP_CONTRACT_INVOKE_REQUEST,
		Payload: &modules.ContractInvokeRequestPayload{
			ContractId: deployId,
			Args:       args,
			Timeout:    timeout,
		},
	}

	tx.AddMessage(msgReq)

	return rlp.EncodeToBytes(tx)
}

//func (p *Processor) ContractTxReqBroadcast(deployId []byte,txBytes []byte, args [][]byte, timeout time.Duration) ([]byte, error) {
func (p *Processor) ContractTxReqBroadcast(deployId []byte, signer, from, to common.Address, daoAmount, daoFee uint64, args [][]byte, timeout time.Duration) ([]byte, error) {
	if deployId == nil || args == nil {
		log.Error("ContractTxReqBroadcast", "param is nil")
		return nil, errors.New("ContractTxReqBroadcast request param is nil")
	}
	log.Debug("ContractTxReqBroadcast", "enter, deployId", deployId)

	tx, err := p.dag.CreateBaseTransaction(from, to, daoAmount, daoFee)
	if err != nil {
		return nil, err
	}
	log.Debug("ContractTxReqBroadcast","tx:", tx)

	msgReq := &modules.Message{
		App: modules.APP_CONTRACT_INVOKE_REQUEST,
		Payload: &modules.ContractInvokeRequestPayload{
			ContractId: deployId,
			Args:       args,
			Timeout:    timeout,
		},
	}
	tx.AddMessage(msgReq)
	tx, err = p.ptn.SignGenericTransaction(signer, tx)
	if err != nil {
		return nil, err
	}
	reqId := tx.RequestHash()
	p.locker.Lock()
	p.mtx[reqId] = &contractTx{
		reqTx:      tx,
		tm:         time.Now(),
		valid:      true,
		executable: true, //default
	}
	p.locker.Unlock()
	log.Debug("ContractTxReqBroadcast ok", "deployId", deployId, "TxId", reqId.String(), "TxHash", tx.Hash().String())

	if p.mtx[reqId].executable {
		if nodeContractExecutable(p.dag, p.local, tx) == true {
			go runContractReq(p.dag, p.contract, p.mtx[reqId])
		}
	}
	//broadcast
	go p.ptn.ContractBroadcast(ContractExeEvent{Tx: tx})
	//local
	//go p.contractExecFeed.Send(ContractExeEvent{modules.NewTransaction([]*modules.Message{msgPay, msgReq})})
	//go p.ProcessContractEvent(&ContractExeEvent{Tx: tx})

	return reqId[:], nil
}

func (p *Processor) ContractTxBroadcast(txBytes []byte) ([]byte, error) {
	log.Info("ContractTxBroadcast enter")
	if txBytes == nil {
		log.Error("ContractTxBroadcast", "param is nil")
		return nil, errors.New("transaction request param is nil")
	}

	tx := &modules.Transaction{}
	if err := rlp.DecodeBytes(txBytes, tx); err != nil {
		return nil, err
	}

	req := tx.RequestHash()
	p.locker.Lock()
	p.mtx[req] = &contractTx{
		reqTx:      tx,
		tm:         time.Now(),
		valid:      true,
		executable: true, //default
	}
	p.locker.Unlock()

	//broadcast
	go p.ptn.ContractBroadcast(ContractExeEvent{Tx: tx})

	return req[:], nil
}

func printTxInfo(tx *modules.Transaction) {
	if tx == nil {
		return
	}

	log.Info("=========tx info============hash:", tx.Hash().String())
	for i := 0; i < len(tx.TxMessages); i++ {
		log.Info("---------")
		app := tx.TxMessages[i].App
		pay := tx.TxMessages[i].Payload
		log.Info("", "app:", app)
		if app == modules.APP_PAYMENT {
			p := pay.(*modules.PaymentPayload)
			fmt.Println(p.LockTime)
		} else if app == modules.APP_CONTRACT_INVOKE_REQUEST {
			p := pay.(*modules.ContractInvokeRequestPayload)
			fmt.Println(p.ContractId)
		} else if app == modules.APP_CONTRACT_INVOKE {
			p := pay.(*modules.ContractInvokePayload)
			fmt.Println(p.Args)
			for idx, v := range p.WriteSet {
				fmt.Printf("WriteSet:idx[%d], k[%v]-v[%v]", idx, v.Key, v.Value)
			}
			for idx, v := range p.ReadSet {
				fmt.Printf("ReadSet:idx[%d], k[%v]-v[%v]", idx, v.Key, v.Value)
			}
		} else if app == modules.APP_SIGNATURE {
			p := pay.(*modules.SignaturePayload)
			fmt.Printf("Signatures:[%v]", p.Signatures)
		}
	}
}
