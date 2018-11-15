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
	"sync"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/event"
	"github.com/dedis/kyber"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/errors"

	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/core/gen"
	"github.com/palletone/go-palletone/dag/txspool"
	"github.com/palletone/go-palletone/core/accounts/keystore"
	"github.com/palletone/go-palletone/common/p2p"
	"bytes"
	"github.com/palletone/go-palletone/contracts"
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
}

type iDag interface {
}

type Processor struct {
	name     string
	ptn      PalletOne
	dag      iDag
	ptype    PeerType
	address  common.Address
	contract *contracts.Contract

	txPool txspool.ITxPool
	locker *sync.Mutex
	quit   chan struct{}
	jurors map[common.Address]Juror //记录所有执行合约的节点信息

	//contracts  map[modules.MessageType]map[common.Hash]interface{} //本地记录合约执行结果，其中interface为对应的payload
	contractTx map[common.Hash]*modules.Transaction

	contractExecFeed  event.Feed
	contractExecScope event.SubscriptionScope

	contractSigFeed  event.Feed
	contractSigScope event.SubscriptionScope
}

func NewContractProcessor(ptn PalletOne, dag iDag, contract *contracts.Contract) (*Processor, error) {
	if ptn == nil || dag == nil {
		return nil, errors.New("NewContractProcessor, param is nil")
	}
	p := &Processor{
		name:       "conract processor",
		ptn:        ptn,
		dag:        dag,
		contract:   contract,
		quit:       make(chan struct{}),
		contractTx: make(map[common.Hash]*modules.Transaction),
	}

	log.Info("NewContractProcessor ok")
	return p, nil
}

func (p *Processor) Start(server *p2p.Server) error {
	//启动消息接收处理线程

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
	if event == nil {
		return errors.New("param is nil")
	}

	if false == checkTxValid(event.Tx) {
		return errors.New("ProcessContractEvent recv event Tx is invalid")
	}

	cmsgType, payload, err := runContractCmd(p.contract, event.Tx)
	if err != nil {
		return err
	}

	tx, _, err := gen.GenContractSigTransctions(p.address, event.Tx, cmsgType, payload, nil)
	if err != nil {
		log.Error("GenContractSigTransctions", "err:%s", err)
		return err
	}
	p.contractTx[event.Tx.TxHash] = tx
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

	if false == checkTxValid(event.Tx) {
		return errors.New("ProcessContractSigEvent event Tx is invalid")
	}
	tx := p.contractTx[event.Tx.TxHash]

	if tx == nil {
		log.Info("ProcessContractSigEvent", "tx(%s) is nil", event.Tx.TxHash)
		go func() {
			for i := 0; i < 10; i += 1 {
				time.Sleep(time.Millisecond * 500)
				if p.contractTx[event.Tx.TxHash] != nil {
					tx = p.contractTx[event.Tx.TxHash]
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
	//检查接收到的tx有效性
	if checkTxValid(local) != true || checkTxValid(recv) != true {
		return false, errors.New("checkAndAddTxData local or recv Tx is invalid")
	}

	//检查除签名msg外的其他msg内容是否相同
	if len(local.TxMessages) != len(recv.TxMessages) {
		return false, errors.New("tx msg is invalid")
	}
	for i := 0; i < len(local.TxMessages); i++ {
		if local.TxMessages[i] != recv.TxMessages[i] {
			return false, errors.New("tx msg not equal")
		}
	}

	//检查签名是否已存在
	for _, msg := range recv.TxMessages {
		if msg.App == modules.APP_SIGNATURE {
			recvSigMsg = msg
		}
	}

	if recvSigMsg == nil {
		return false, errors.New("not find recv sig msg")
	}
	//todo 验证签名的有效性
	for i, msg := range local.TxMessages {
		if msg.App == modules.APP_SIGNATURE {
			sigPayload := msg.Payload.(*modules.SignaturePayload)
			sigs := sigPayload.Signatures
			for _, sig := range sigs {
				if true == bytes.Equal(sig.PubKey, recvSigMsg.Payload.(*modules.SignaturePayload).Signatures[0].PubKey) &&
					true == bytes.Equal(sig.Signature, recvSigMsg.Payload.(*modules.SignaturePayload).Signatures[0].Signature) {
					log.Info("tx %s already recv", recv.TxHash)
					return false, nil
				}
			}
			//直接将签名添加到msg中
			if len(recvSigMsg.Payload.(*modules.SignaturePayload).Signatures) > 0 {
				sigPayload.Signatures = append(sigs, recvSigMsg.Payload.(*modules.SignaturePayload).Signatures[0])
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

func checkTxValid(tx *modules.Transaction) bool {
	if tx == nil {
		return false
	}
	//签名检查

	return true
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
	msgPay := &modules.Message{
		App:     modules.APP_PAYMENT,
		Payload: modules.PaymentPayload{},
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
