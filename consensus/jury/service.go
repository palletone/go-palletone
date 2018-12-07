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
	"github.com/dedis/kyber"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/event"
	"github.com/palletone/go-palletone/dag/errors"
	"github.com/palletone/go-palletone/dag/modules"
	"sync"
	"time"

	"bytes"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/p2p"
	"github.com/palletone/go-palletone/common/rlp"
	//mp "github.com/palletone/go-palletone/consensus/mediatorplugin"
	"github.com/palletone/go-palletone/contracts"
	"github.com/palletone/go-palletone/core/accounts/keystore"
	"github.com/palletone/go-palletone/core/gen"
	cm "github.com/palletone/go-palletone/dag/common"
	"github.com/palletone/go-palletone/dag/txspool"
	"reflect"
)

type PeerType int

const (
	_ PeerType = iota
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
}

type iDag interface {
	GetActiveMediators() []common.Address
}

type contractTx struct {
	list  []common.Address //dynamic
	tx    *modules.Transaction
	valid bool
	tm    time.Time //creat time
}

type Processor struct {
	name  string
	ptype PeerType
	ptn   PalletOne
	dag   iDag
	//local    *mp.MediatorAccount //local
	contract *contracts.Contract
	locker   *sync.Mutex
	quit     chan struct{}
	mtx      map[common.Hash]*contractTx

	contractExecFeed  event.Feed
	contractExecScope event.SubscriptionScope
	contractSigFeed   event.Feed
	contractSigScope  event.SubscriptionScope
}

func NewContractProcessor(ptn PalletOne, dag iDag, contract *contracts.Contract) (*Processor, error) {
	if ptn == nil || dag == nil {
		return nil, errors.New("NewContractProcessor, param is nil")
	}

	p := &Processor{
		name:     "conractProcessor",
		ptn:      ptn,
		dag:      dag,
		contract: contract,
		locker:   new(sync.Mutex),
		quit:     make(chan struct{}),
		mtx:      make(map[common.Hash]*contractTx),
	}

	//log.Info("NewContractProcessor ok", "local address", localmediator.Address.String())
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
		return errors.New("param is nil")
	}
	if _, ok := p.mtx[event.Tx.TxId]; ok {
		return nil
	}
	log.Debug("ProcessContractEvent", "enter, tx req id ", event.Tx.TxId)

	if false == checkTxValid(event.Tx) {
		return errors.New(fmt.Sprintf("ProcessContractEvent recv event Tx is invalid, txid:%s", event.Tx.TxId.String()))
	}
	p.locker.Lock()
	p.mtx[event.Tx.TxId] = &contractTx{
		tx:    event.Tx,
		tm:    time.Now(),
		valid: true,
	}
	p.locker.Unlock()

	//broadcast contract request transaction event
	//go p.ptn.ContractBroadcast(*event)
	return nil
}

func (p *Processor) RunContractLoop(txpool txspool.ITxPool, addr common.Address, ks *keystore.KeyStore) error {
	//log.Debug("ProcessContractEvent", "enter", addr.String())
	for _, ctx := range p.mtx {
		if false == ctx.valid {
			continue
		}
		ctx.valid = false

		if false == checkTxValid(ctx.tx) {
			log.Error("RunContractLoop recv event Tx is invalid,", "txid", ctx.tx.TxId.String())
			continue
		}
		cmsgType, payload, err := runContractCmd(p.contract, ctx.tx)

		if err != nil {
			log.Error("RunContractLoop runContractCmd", "error", err.Error())
			continue
		}
		tx, err := gen.GenContractSigTransctions(addr, ctx.tx, cmsgType, payload, ks)
		if err != nil {
			log.Error("RunContractLoop GenContractSigTransctions", "error", err.Error())
			continue
		}
		log.Debug("RunContractLoop", "tx", tx)
		if err = txpool.AddLocal(txspool.TxtoTxpoolTx(txpool, tx)); err != nil {
			log.Error("RunContractLoop", "error", err.Error())
			continue
		}

	}

	return nil
}

func (p *Processor) CheckContractTxValid(tx *modules.Transaction) bool {
	//检查本地是否存在合约执行交易，如果不存在则执行并记录到本地，并与接收到的tx进行合约比较
	if tx == nil {
		log.Error("CheckContractTxValid", "param is nil")
		return false
	}
	log.Debug("CheckContractTxValid", "tx req id ", tx.TxId)

	if false == checkTxValid(tx) {
		log.Error("CheckContractTxValid", "checkTxValid fail")
		return false
	}
	_, payload, err := runContractCmd(p.contract, tx)
	if err != nil {
		log.Error("CheckContractTxValid runContractCmd", "error", err.Error())
		return false
	}

	for i := 0; i < len(tx.TxMessages); i++ {
		if tx.TxMessages[i].App == modules.APP_CONTRACT_INVOKE {
			if reflect.DeepEqual(tx.TxMessages[i], payload) != true {
				log.Error("CheckContractTxValid", "invoke msg not equal")
				return false
			}
		}
	}
	p.locker.Lock()
	if _, ok := p.mtx[tx.TxId]; ok {
		p.mtx[tx.TxId].valid = false
	}
	p.locker.Unlock()
	log.Debug("CheckContractTxValid", "local txid", tx.TxId, "contract transaction:", p.mtx[tx.TxId].list)
	return true
}

func (p *Processor) ProcessContractSigEvent(event *ContractSigEvent) error {
	if event == nil || len(event.Tx.TxMessages) < 1 {
		return errors.New("ProcessContractSigEvent param is nil")
	}

	log.Info("ProcessContractSigEvent", "enter,event tx req id:", event.Tx.TxId.String())
	if false == checkTxValid(event.Tx) {
		return errors.New("ProcessContractSigEvent event Tx is invalid")
	}
	if _, ok := p.mtx[event.Tx.TxId]; ok != true {
		errMsg := fmt.Sprintf("local not find txid: %s", event.Tx.TxId.String())
		log.Error("ProcessContractSigEvent", errMsg)
		return errors.New(errMsg)
	}

	cx := p.mtx[event.Tx.TxId]
	if cx.tx == nil {
		log.Info("ProcessContractSigEvent", "local no tx id, wait for moment:", event.Tx.TxId.String())
		go func() error {
			for i := 0; i < 10; i += 1 {
				time.Sleep(time.Millisecond * 500)
				if cx.tx != nil {
					if judge, err := checkAndAddTxData(cx.tx, event.Tx); err == nil && judge == true {
						if err = p.addTx2LocalTxTool(cx.tx, len(cx.list)); err == nil {
							p.locker.Lock()
							delete(p.mtx, event.Tx.TxId)
							p.locker.Unlock()
						} else {
							return err
						}
					}
					return errors.New("checkAndAddTxData fail")
				}
			}
			return errors.New(fmt.Sprintf("ProcessContractSigEvent checkAndAddTxData wait local transaction timeout, tx id:%s", cx.tx.TxId))
		}()
	} else {
		log.Info("ProcessContractSigEvent", "tx is ok", event.Tx.TxId)
		if judge, err := checkAndAddTxData(cx.tx, event.Tx); err != nil {
			log.Error("ProcessContractSigEvent", "checkAndAddTxData err:", err.Error())
			return err
		} else if judge == true {
			if err = p.addTx2LocalTxTool(cx.tx, len(cx.list)); err == nil {
				p.locker.Lock()
				delete(p.mtx, event.Tx.TxId)
				p.locker.Unlock()
			} else {
				return err
			}
		}
	}
	return errors.New(fmt.Sprintf("ProcessContractSigEvent err with tx id:%s", cx.tx.TxId.String()))
}

func (p *Processor) SubscribeContractSigEvent(ch chan<- ContractSigEvent) event.Subscription {
	return p.contractSigScope.Track(p.contractSigFeed.Subscribe(ch))
}

func (p *Processor) ContractTxDeleteLoop() {
	for {
		time.Sleep(time.Second * time.Duration(50))

		p.locker.Lock()
		for k, v := range p.mtx {
			if time.Since(v.tm) > time.Second*100 { //todo
				log.Info("ContractTxDeleteLoop", "delete id", k.String())
				if v.valid == false {
					//delete(p.mtx, k)
				}
			}
		}
		p.locker.Unlock()
	}
}

//执行合约命令:install、deploy、invoke、stop，同时只支持一种类型
func runContractCmd(contract *contracts.Contract, trs *modules.Transaction) (modules.MessageType, []*modules.Message, error) {
	if trs == nil || len(trs.TxMessages) <= 0 {
		return 0, nil, errors.New("Transaction or msg is nil")
	}

	for _, msg := range trs.TxMessages {
		switch msg.App {
		case modules.APP_CONTRACT_TPL_REQUEST:
			{
				return modules.APP_CONTRACT_TPL, nil, errors.New("not support APP_CONTRACT_TPL")
			}
		case modules.APP_CONTRACT_DEPLOY_REQUEST:
			{
				return modules.APP_CONTRACT_DEPLOY, nil, errors.New("not support APP_CONTRACT_DEPLOY")
			}
		case modules.APP_CONTRACT_INVOKE_REQUEST:
			{
				req := ContractInvokeReq{
					chainID:  "palletone",
					deployId: msg.Payload.(*modules.ContractInvokeRequestPayload).ContractId,
					args:     msg.Payload.(*modules.ContractInvokeRequestPayload).Args,
					txid:     trs.TxId.String(),
					tx:       trs,
				}
				payload, err := ContractProcess(contract, req)
				if err != nil {
					log.Error("runContractCmd ContractProcess ", "error", err)
					return msg.App, nil, errors.New(fmt.Sprintf("txid(%s)APP_CONTRACT_INVOKE rans err:%s", req.txid, err))
				}
				return modules.APP_CONTRACT_INVOKE, payload, nil
			}
		case modules.APP_CONTRACT_STOP_REQUEST:
			{
				return modules.APP_CONTRACT_STOP, nil, errors.New("not support APP_CONTRACT_STOP")
			}
		}
	}

	return 0, nil, errors.New(fmt.Sprintf("Transaction err, txid=%s", trs.TxHash))
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
					log.Info("tx  already recv:", recv.TxId.String())
					return false, nil
				}
			}
			//直接将签名添加到msg中
			if len(recvSigMsg.Payload.(*modules.SignaturePayload).Signatures) > 0 {
				sigPayload.Signatures = append(sigs, recvSigMsg.Payload.(*modules.SignaturePayload).Signatures[0])
			}

			local.TxMessages[i].Payload = sigPayload
			local.TxHash = common.Hash{}
			local.TxHash = local.Hash()

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

func (p *Processor) addTx2LocalTxTool(tx *modules.Transaction, cnt int) error {
	if tx == nil || cnt < 4 {
		return errors.New(fmt.Sprintf("addTx2LocalTxTool param error, node count is [%d]", cnt))
	}
	if num := getTxSigNum(tx); num < (cnt*2/3 + 1) {
		log.Error("addTx2LocalTxTool sig num is", num)
		return errors.New(fmt.Sprintf("addTx2LocalTxTool tx sig num is:%d", num))
	}

	txPool := p.ptn.TxPool()
	log.Debug("addTx2LocalTxTool", "tx:", tx.TxHash.String())

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
	tx.TxHash = common.Hash{}
	tx.TxId = tx.Hash()

	tx.TxHash = common.Hash{}
	tx.TxHash = tx.Hash()

	return rlp.EncodeToBytes(tx)
}
func (p *Processor) ContractTxReqBroadcast(deployId []byte, txid string, txBytes []byte, args [][]byte, timeout time.Duration) ([]byte, error) {
	log.Info("ContractTxReqBroadcast", fmt.Sprintf("enter, deployId[%v], txid[%s]", deployId, txid))
	if deployId == nil || args == nil {
		log.Error("ContractTxReqBroadcast", "param is nil")
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
			LockTime: 0, //todo
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
	tx.TxHash = common.Hash{}
	tx.TxId = tx.Hash()

	tx.TxHash = common.Hash{}
	tx.TxHash = tx.Hash()

	p.locker.Lock()
	p.mtx[tx.TxId] = &contractTx{
		tx:    tx,
		tm:    time.Now(),
		valid: true,
	}
	p.locker.Unlock()

	//broadcast
	go p.ptn.ContractBroadcast(ContractExeEvent{Tx: tx})
	//local
	//go p.contractExecFeed.Send(ContractExeEvent{modules.NewTransaction([]*modules.Message{msgPay, msgReq})})
	//go p.ProcessContractEvent(&ContractExeEvent{Tx: tx})

	return tx.TxId[:], nil
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

	tx.TxHash = common.Hash{}
	tx.TxHash = tx.Hash()

	tx.TxId = tx.Hash()

	p.locker.Lock()
	p.mtx[tx.TxId] = &contractTx{
		tx:    tx,
		tm:    time.Now(),
		valid: true,
	}
	p.locker.Unlock()

	//broadcast
	go p.ptn.ContractBroadcast(ContractExeEvent{Tx: tx})

	return tx.TxId[:], nil
}

func printTxInfo(tx *modules.Transaction) {
	if tx == nil {
		return
	}

	log.Info("=========tx info============hash:", tx.TxHash.String())
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
