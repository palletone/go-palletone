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
	"github.com/palletone/go-palletone/contracts"
	"github.com/palletone/go-palletone/core/accounts/keystore"
	"github.com/palletone/go-palletone/core/gen"
	cm "github.com/palletone/go-palletone/dag/common"
	mp "github.com/palletone/go-palletone/consensus/mediatorplugin"
	"github.com/palletone/go-palletone/dag/txspool"
	"github.com/palletone/go-palletone/core/accounts"
)

type PeerType int

const (
	_         PeerType = iota
	TUnknow
	TJury
	TMediator
)

type Juror struct {
	name    string
	address common.Address

	//InitPartSec kyber.Scalar
	InitPartPub kyber.Point
}

type PalletOne interface {
	GetKeyStore() *keystore.KeyStore
	TxPool() txspool.ITxPool

	MockContractLocalSend(event ContractExeEvent)
	MockContractSigLocalSend(event ContractSigEvent)

	ContractBroadcast(event ContractExeEvent)
	ContractSigBroadcast(event ContractSigEvent)

	GetLocalMediators() *mp.MediatorAccount
}

type iDag interface {
	GetActiveMediators() []common.Address
}

type contractTx struct {
	list []common.Address //dynamic
	tx   *modules.Transaction
}

type Processor struct {
	name     string
	ptype    PeerType
	ptn      PalletOne
	dag      iDag
	local    *mp.MediatorAccount //local
	contract *contracts.Contract
	txPool   txspool.ITxPool
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
	var address common.Address
	localmediators := ptn.GetLocalMediators()

	p := &Processor{
		name:     "conract processor",
		ptn:      ptn,
		dag:      dag,
		contract: contract,
		quit:     make(chan struct{}),
		mtx:      make(map[common.Hash]*contractTx),
		local:    localmediators,
	}

	log.Info("NewContractProcessor ok", "mediator_address", address.String())
	log.Info("NewContractProcessor", "info:%v", p.local)
	return p, nil
}

func (p *Processor) Start(server *p2p.Server) error {
	//启动消息接收处理线程

	//合约执行节点更新线程

	//合约执行线程
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
	log.Info("ProcessContractEvent", "enter", event.Tx.TxHash)

	for i := 0; i < len(event.Tx.TxMessages); i++ {
		log.Info("ProcessContractEvent:", ":", event.Tx.TxMessages[i].App)
	}

	if event == nil {
		return errors.New("param is nil")
	}

	if false == checkTxValid(event.Tx, p.ptn.GetKeyStore()) {
		return errors.New("ProcessContractEvent recv event Tx is invalid")
	}

	cmsgType, payload, err := runContractCmd(p.contract, event.Tx)
	if err != nil {
		return err
	}
	ks := p.ptn.GetKeyStore()
	ks.Unlock(accounts.Account{Address: p.local.Address}, p.local.Password)

	tx, _, err := gen.GenContractSigTransctions(p.local.Address, event.Tx, cmsgType, payload, p.ptn.GetKeyStore())
	if err != nil {
		log.Error("GenContractSigTransctions", "err:%s", err)
		return err
	}
	//p.contractTx[event.Tx.TxHash] = tx
	p.mtx[event.Tx.TxHash] = &contractTx{
		list: p.dag.GetActiveMediators(),
		tx:   tx,
	}

	log.Info("ProcessContractEvent", "trs:", tx)
	log.Info("ProcessContractEvent", "add tx", event.Tx.TxHash)

	//broadcast
	go p.ptn.ContractSigBroadcast(ContractSigEvent{tx.TxHash, tx})
	//local
	go p.contractSigFeed.Send(ContractSigEvent{tx.TxHash, tx})

	return nil
}

func (p *Processor) ProcessContractSigEvent(event *ContractSigEvent) error {
	log.Info("ProcessContractSigEvent", "event:", event.Tx.TxHash)
	if event == nil {
		return errors.New("ProcessContractSigEvent param is nil")
	}

	if false == checkTxValid(event.Tx, p.ptn.GetKeyStore()) {
		return errors.New("ProcessContractSigEvent event Tx is invalid")
	}
	tx := p.mtx[event.Tx.TxHash].tx
	if tx == nil {
		log.Info("ProcessContractSigEvent", "tx(%s) is nil", event.Tx.TxHash)
		go func() {
			for i := 0; i < 10; i += 1 {
				time.Sleep(time.Millisecond * 500)
				if p.mtx[event.Tx.TxHash].tx != nil {
					tx = p.mtx[event.Tx.TxHash].tx
					if judge, _ := checkAndAddTxData(tx, event.Tx); judge == true {
						//收集签名数量，达到要求后将tx添加到交易池
						num := getTxSigNum(tx)
						log.Info("ProcessContractSigEvent", "tx sig num=%d", num)
						if num >= 1 { //todo
							if p.addTx2LocalTxTool(tx) != nil {
								log.Error("ProcessContractSigEvent", "tx(%s)addTx2LocalTxTool fail", tx.TxHash)
								return
							}
						}
					}
					return
				}
			}
		}()
	} else {
		log.Info("ProcessContractSigEvent", "tx(%s) is ok", event.Tx.TxHash)
		if judge, _ := checkAndAddTxData(tx, event.Tx); judge == true {
			num := getTxSigNum(tx)
			log.Info("ProcessContractSigEvent", "tx sig num=%d", num)
			if num >= 1 { //todo
				if p.addTx2LocalTxTool(tx) != nil {
					log.Error("ProcessContractSigEvent", "tx(%s)addTx2LocalTxTool fail", tx.TxHash)
				}
			}
		}
	}
	return nil
}

func (p *Processor) SubscribeContractSigEvent(ch chan<- ContractSigEvent) event.Subscription {
	return p.contractSigScope.Track(p.contractSigFeed.Subscribe(ch))
}

//执行合约命令:install、deploy、invoke、stop，注意同时只支持一种类型
func runContractCmd(contract *contracts.Contract, trs *modules.Transaction) (modules.MessageType, interface{}, error) {
	if trs == nil {
		return 0, nil, errors.New("Transaction is nil")
	}
	if len(trs.TxMessages) <= 0 {
		return 0, nil, errors.New("TxMessages is not exit in Transaction")
	}
	for _, msg := range trs.TxMessages {
		switch msg.App {
		case modules.APP_CONTRACT_TPL_REQUEST:
			{
				return modules.APP_CONTRACT_TPL, nil, errors.New("not support APP_CONTRACT_TPL")
			}
		case modules.APP_CONTRACT_DEPLOY_REQUEST:
			{
				return msg.App, nil, errors.New("not support APP_CONTRACT_DEPLOY")
			}
		case modules.APP_CONTRACT_INVOKE_REQUEST:
			{
				req := ContractInvokeReq{
					chainID:  "palletone",
					deployId: msg.Payload.(modules.ContractInvokeRequestPayload).ContractId,
					args:     msg.Payload.(modules.ContractInvokeRequestPayload).Args,
					txid:     trs.TxHash.String(),
					tx:       trs,
				}
				payload, err := ContractProcess(contract, req)
				if err != nil {
					log.Error("runContractCmd", "ContractProcess fail:%s", err)
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

	//检查除签名msg外的其他msg内容是否相同
	if len(local.TxMessages) != len(recv.TxMessages) {
		return false, errors.New("tx msg is invalid")
	}
	for i := 0; i < len(local.TxMessages); i++ {
		if local.TxMessages[i] != recv.TxMessages[i] {
			return false, errors.New("tx msg not equal")
		}
	}
	for _, msg := range recv.TxMessages {
		if msg.App == modules.APP_SIGNATURE {
			recvSigMsg = msg
		}
	}

	if recvSigMsg == nil {
		return false, errors.New("not find recv sig msg")
	}
	for i, msg := range local.TxMessages {
		if msg.App == modules.APP_SIGNATURE {
			sigPayload := msg.Payload.(modules.SignaturePayload)
			sigs := sigPayload.Signatures
			for _, sig := range sigs {
				if true == bytes.Equal(sig.PubKey, recvSigMsg.Payload.(modules.SignaturePayload).Signatures[0].PubKey) &&
					true == bytes.Equal(sig.Signature, recvSigMsg.Payload.(modules.SignaturePayload).Signatures[0].Signature) {
					log.Info("tx  already recv:", recv.TxHash.String())
					return false, nil
				}
			}
			//直接将签名添加到msg中
			if len(recvSigMsg.Payload.(modules.SignaturePayload).Signatures) > 0 {
				sigPayload.Signatures = append(sigs, recvSigMsg.Payload.(modules.SignaturePayload).Signatures[0])
			}
			local.TxMessages[i].Payload = sigPayload
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

func checkTxValid(tx *modules.Transaction, ks *keystore.KeyStore) bool {
	if tx == nil {
		return false
	}
	printTxInfo(tx)

	return cm.ValidateTxSig(tx, ks)
}

func (p *Processor) addTx2LocalTxTool(tx *modules.Transaction) error {
	if tx == nil {
		return errors.New("addTx2LocalTxTool param is nil")
	}
	txPool := p.ptn.TxPool()
	log.Debug("addTx2LocalTxTool", "tx(%s)", tx.TxHash)

	return txPool.AddLocal(txspool.TxtoTxpoolTx(txPool, tx))
}

//todo txid ?
func (p *Processor) ContractTxReqBroadcast(deployId []byte, txid string, args [][]byte, timeout time.Duration) error {
	log.Info("ContractTxReqBroadcast", "enter", "ok")
	if deployId == nil || args == nil {
		log.Error("ContractTxReqBroadcast", "param is nil")
		return errors.New("transaction request param is nil")
	}
	pay := &modules.PaymentPayload{
		Inputs:   []*modules.Input{},
		Outputs:  []*modules.Output{},
		LockTime: 11111, //todo
	}
	msgPay := &modules.Message{
		App:     modules.APP_PAYMENT,
		Payload: pay,
	}
	msgReq := &modules.Message{
		App: modules.APP_CONTRACT_INVOKE_REQUEST,
		Payload: modules.ContractInvokeRequestPayload{
			ContractId: deployId,
			Args:       args,
			Timeout:    timeout,
		},
	}

	//broadcast
	go p.ptn.ContractBroadcast(ContractExeEvent{modules.NewTransaction([]*modules.Message{msgPay, msgReq})})
	//local
	go p.contractExecFeed.Send(ContractExeEvent{modules.NewTransaction([]*modules.Message{msgPay, msgReq})})

	return nil
}

func printTxInfo(tx *modules.Transaction) {
	if tx == nil {
		return
	}

	log.Info("=========tx info============hash:", tx.TxHash.String())

	for i := 0; i < len(tx.TxMessages); i++ {
		log.Info("---------")
		app := tx.TxMessages[i].App
		log.Info("", "app:", app)
		if app == modules.APP_PAYMENT {
			fmt.Println(tx.TxMessages[i].Payload.(*modules.PaymentPayload).LockTime)
		} else if app == modules.APP_CONTRACT_INVOKE_REQUEST {
			fmt.Println(tx.TxMessages[i].Payload.(modules.ContractInvokeRequestPayload).ContractId, tx.TxMessages[i].Payload.(modules.ContractInvokeRequestPayload).Args)
		} else if app == modules.APP_CONTRACT_INVOKE {
			fmt.Println(tx.TxMessages[i].Payload.(*modules.ContractInvokePayload).ContractId, tx.TxMessages[i].Payload.(*modules.ContractInvokePayload).Args)
		} else if app == modules.APP_SIGNATURE {
			fmt.Println(tx.TxMessages[i].Payload.(modules.SignaturePayload).Signatures)
		}
	}

}
