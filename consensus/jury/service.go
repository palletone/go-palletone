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
	mp "github.com/palletone/go-palletone/consensus/mediatorplugin"
	"github.com/palletone/go-palletone/contracts"
	"github.com/palletone/go-palletone/core/accounts"
	"github.com/palletone/go-palletone/core/accounts/keystore"
	"github.com/palletone/go-palletone/core/gen"
	cm "github.com/palletone/go-palletone/dag/common"
	"github.com/palletone/go-palletone/dag/txspool"
)

type PeerType int

const (
	_ PeerType = iota
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
	log.Info("NewContractProcessor", "info:", p.local)
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
	if event == nil {
		return errors.New("param is nil")
	}

	if false == checkTxValid(event.Tx) {
		return errors.New("ProcessContractEvent recv event Tx is invalid")
	}

	payload, err := runContractCmd(p.contract, event.Tx)
	if err != nil {
		log.Error(fmt.Sprintf("ProcessContractEvent runContractCmd err:%s", err))
		return err
	}
	ks := p.ptn.GetKeyStore()
	err = ks.Unlock(accounts.Account{Address: p.local.Address}, p.local.Password)
	if err != nil {
		log.Error(fmt.Sprintf("ProcessContractEvent account add[%s], password[%s], err[%s]", p.local.Address.String(), p.local.Password, err))
		return err
	}
	tx, _, err := gen.GenContractSigTransctions(p.local.Address, event.Tx, payload, p.ptn.GetKeyStore())
	if err != nil {
		log.Error(fmt.Sprintf("ProcessContractEvent GenContractSigTransctions, err:%s", err.Error()))
		return err
	}

	p.mtx[event.Tx.TxId] = &contractTx{
		list: p.dag.GetActiveMediators(),
		tx:   tx,
	}

	log.Info("ProcessContractEvent", "trs:", tx)
	log.Info("ProcessContractEvent", "add tx req id:", event.Tx.TxId)

	//broadcast
	go p.ptn.ContractSigBroadcast(ContractSigEvent{tx})
	//local
	//go p.contractSigFeed.Send(ContractSigEvent{tx.TxHash, tx})
	go p.ProcessContractSigEvent(&ContractSigEvent{tx})

	return nil
}

func (p *Processor) ProcessContractSigEvent(event *ContractSigEvent) error {
	log.Info("ProcessContractSigEvent", "event tx req id:", event.Tx.TxId.String())
	if event == nil {
		return errors.New("ProcessContractSigEvent param is nil")
	}

	if false == checkTxValid(event.Tx) {
		return errors.New("ProcessContractSigEvent event Tx is invalid")
	}
	tx := p.mtx[event.Tx.TxId].tx
	if tx == nil {
		log.Info("ProcessContractSigEvent", "local no tx id:", event.Tx.TxId.String())
		go func() {
			for i := 0; i < 10; i += 1 {
				time.Sleep(time.Millisecond * 500)
				if p.mtx[event.Tx.TxId].tx != nil {
					tx = p.mtx[event.Tx.TxId].tx
					judge, err := checkAndAddTxData(tx, event.Tx)
					if err == nil && judge == true {
						//收集签名数量，达到要求后将tx添加到交易池
						num := getTxSigNum(tx)
						log.Info("ProcessContractSigEvent", "tx sig num:", num)
						if num >= 1 { //todo
							if p.addTx2LocalTxTool(tx) != nil {
								log.Error("ProcessContractSigEvent", "tx id addTx2LocalTxTool fail:", tx.TxId.String())
								return
							}
						}
					}
					return
				}
			}
		}()
	} else {
		log.Info("ProcessContractSigEvent", "tx is ok", event.Tx.TxId)
		judge, err := checkAndAddTxData(tx, event.Tx)
		if err != nil {
			return err
		} else if judge == true {
			num := getTxSigNum(tx)
			log.Info("ProcessContractSigEvent", "tx sig num:", num)
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
func runContractCmd(contract *contracts.Contract, trs *modules.Transaction) ([]*modules.Message, error) {
	if trs == nil {
		return nil, errors.New("Transaction is nil")
	}
	if len(trs.TxMessages) <= 0 {
		return nil, errors.New("TxMessages is not exit in Transaction")
	}
	for _, msg := range trs.TxMessages {
		switch msg.App {
		case modules.APP_CONTRACT_TPL_REQUEST:
			{
				return nil, errors.New("not support APP_CONTRACT_TPL")
			}
		case modules.APP_CONTRACT_DEPLOY_REQUEST:
			{
				return nil, errors.New("not support APP_CONTRACT_DEPLOY")
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
					log.Error("runContractCmd", "ContractProcess fail:", err)
					return nil, errors.New(fmt.Sprintf("txid(%s)APP_CONTRACT_INVOKE rans err:%s", req.txid, err))
				}
				return payload, nil
			}
		case modules.APP_CONTRACT_STOP_REQUEST:
			{
				return nil, errors.New("not support APP_CONTRACT_STOP")
			}
		}
	}

	return nil, errors.New(fmt.Sprintf("Transaction err, txid=%s", trs.TxHash))
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
				return len(msg.Payload.(modules.SignaturePayload).Signatures)
			}
		}
	}
	return 0
}

func checkTxValid(tx *modules.Transaction) bool {
	if tx == nil {
		return false
	}
	return cm.ValidateTxSig(tx)
}

func (p *Processor) addTx2LocalTxTool(tx *modules.Transaction) error {
	if tx == nil {
		return errors.New("addTx2LocalTxTool param is nil")
	}
	txPool := p.ptn.TxPool()
	log.Debug("addTx2LocalTxTool", "tx:", tx.TxHash.String())

	return txPool.AddLocal(txspool.TxtoTxpoolTx(txPool, tx))
}

//todo txid ?
func (p *Processor) ContractTxReqBroadcast(deployId []byte, txid string, txBytes []byte, args [][]byte, timeout time.Duration) error {
	log.Info("ContractTxReqBroadcast", fmt.Sprintf("enter, deployId[%v], txid[%s]", deployId, txid))
	if deployId == nil || args == nil {
		log.Error("ContractTxReqBroadcast", "param is nil")
		return errors.New("transaction request param is nil")
	}

	tx := &modules.Transaction{}
	if txBytes != nil {
		if err := rlp.DecodeBytes(txBytes, tx); err != nil {
			return err
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
	tx.TxHash = tx.Hash()
	tx.TxId = tx.TxHash
	//broadcast
	go p.ptn.ContractBroadcast(ContractExeEvent{Tx: tx})
	//local
	//go p.contractExecFeed.Send(ContractExeEvent{modules.NewTransaction([]*modules.Message{msgPay, msgReq})})
	go p.ProcessContractEvent(&ContractExeEvent{Tx: tx})

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
			p := pay.(modules.SignaturePayload)
			fmt.Printf("Signatures:[%v]", p.Signatures)
		}
	}

}
