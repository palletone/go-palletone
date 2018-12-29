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

	"github.com/dedis/kyber"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/event"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/p2p"
	"github.com/palletone/go-palletone/common/rlp"
	"github.com/palletone/go-palletone/contracts"
	"github.com/palletone/go-palletone/core/gen"
	"github.com/palletone/go-palletone/dag/errors"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/txspool"
	"github.com/palletone/go-palletone/core/accounts/keystore"
)

type PeerType = int

const (
	CONTRACT_SIG_NUM = 3

	TJury     = 2
	TMediator = 4
)

type PalletOne interface {
	GetKeyStore() *keystore.KeyStore
	TxPool() txspool.ITxPool

	MockContractLocalSend(event ContractExeEvent)
	MockContractSigLocalSend(event ContractSigEvent)

	ContractBroadcast(event ContractExeEvent)
	ContractSigBroadcast(event ContractSigEvent)
	ContractSpecialBroadcast(event ContractSpecialEvent)

	GetLocalMediators() []common.Address
	IsLocalActiveMediator(add common.Address) bool

	SignGenericTransaction(from common.Address, tx *modules.Transaction) (*modules.Transaction, error)
}

type iDag interface {
	GetTxFee(pay *modules.Transaction) (*modules.InvokeFees, error)
	GetAddrByOutPoint(outPoint *modules.OutPoint) (common.Address, error)
	GetActiveMediators() []common.Address
	GetTxHashByReqId(reqid common.Hash) (common.Hash, error)
	IsActiveJury(add common.Address) bool
	IsActiveMediator(add common.Address) bool
	GetAddr1TokenUtxos(addr common.Address, asset *modules.Asset) (map[modules.OutPoint]*modules.Utxo, error)
	CreateGenericTransaction(from, to common.Address, daoAmount, daoFee uint64,
		msg *modules.Message, txPool txspool.ITxPool) (*modules.Transaction, uint64, error)
}

type Juror struct {
	name        string
	address     common.Address
	InitPartPub kyber.Point
}

//合约节点类型、地址信息
type nodeInfo struct {
	addr  common.Address
	ntype int //1:default, 2:jury, 4:mediator
}

type contractTx struct {
	state      int                    //contract run state, 0:default, 1:running
	list       []common.Address       //dynamic
	reqTx      *modules.Transaction   //request contract
	rstTx      *modules.Transaction   //contract run result---system
	sigTx      *modules.Transaction   //contract sig result---user, 0:local, 1,2 other
	rcvTx      []*modules.Transaction //todo 本地没有没有接收过请求合约，缓存已经签名合约
	tm         time.Time              //create time
	valid      bool                   //contract request valid identification
	executable bool                   //contract executable,sys on mediator, user on jury
}

type Processor struct {
	name     string //no user
	ptn      PalletOne
	dag      iDag
	local    map[common.Address]*JuryAccount //[]common.Address //local account addr
	contract *contracts.Contract
	locker   *sync.Mutex
	quit     chan struct{}
	mtx      map[common.Hash]*contractTx //all contract buffer

	contractExecFeed  event.Feed
	contractExecScope event.SubscriptionScope
	contractSigFeed   event.Feed
	contractSigScope  event.SubscriptionScope
}

func NewContractProcessor(ptn PalletOne, dag iDag, contract *contracts.Contract, cfg *Config) (*Processor, error) {
	if ptn == nil || dag == nil {
		return nil, errors.New("NewContractProcessor, param is nil")
	}

	accounts := make(map[common.Address]*JuryAccount, 0)
	for _, cfg := range cfg.Accounts {
		account := cfg.configToAccount()
		addr := account.Address
		accounts[addr] = account
	}

	p := &Processor{
		name:     "conractProcessor",
		ptn:      ptn,
		dag:      dag,
		contract: contract,
		local:    accounts,
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

func (p *Processor) isLocalActiveJury(add common.Address) bool {
	if _, ok := p.local[add]; ok {
		return p.dag.IsActiveJury(add)
	}
	return false
}

func (p *Processor) getLocalNodesInfo() ([]*nodeInfo, error) {
	if len(p.local) < 1 {
		return nil, errors.New("getLocalNodeInfo, no local account")
	}

	nodes := make([]*nodeInfo, 0)
	for addr, _ := range p.local {
		nodeType := 0
		if p.ptn.IsLocalActiveMediator(addr) {
			nodeType = TMediator
		} else if p.isLocalActiveJury(addr) {
			nodeType = TJury
		}
		node := &nodeInfo{
			addr:  addr,
			ntype: nodeType,
		}
		nodes = append(nodes, node)
	}
	return nodes, nil
}

func (p *Processor) runContractReq(req *contractTx) error {
	if req == nil {
		return errors.New("runContractReq param is nil")
	}
	_, msgs, err := runContractCmd(p.dag, p.contract, req.reqTx)
	if err != nil {
		log.Error("runContractReq runContractCmd", "reqTx", req.reqTx.RequestHash().String(), "error", err.Error())
		return err
	}
	tx, err := gen.GenContractTransction(req.reqTx, msgs)
	if err != nil {
		log.Error("runContractReq GenContractSigTransctions", "error", err.Error())
		return err
	}

	//如果系统合约，直接添加到缓存池
	//如果用户合约，需要签名，添加到缓存池并广播
	if isSystemContract(tx) {
		req.rstTx = tx
	} else {
		account := p.getLocalAccount()
		if account == nil {
			return errors.New("runContractReq no local account")
		}
		sigTx, err := gen.GenContractSigTransction(account.Address, account.Password, tx, p.ptn.GetKeyStore())
		if err != nil {
			log.Error("runContractReq GenContractSigTransctions", "error", err.Error())
			return errors.New("")
		}
		req.sigTx = sigTx
		//如果rcvTx存在，则比较执行结果，并将结果附加到sigTx上,并删除rcvTx
		if len(req.rcvTx) > 0 {
			for _, rtx := range req.rcvTx {
				if err := checkAndAddSigSet(req.sigTx, rtx); err != nil {
					log.Error("runContractReq", "checkAndAddSigSet error", err.Error())
				} else {
					log.Debug("runContractReq", "checkAndAddSigSet ok")
				}
			}
			req.rcvTx = nil
		}
		//广播
		go p.ptn.ContractSigBroadcast(ContractSigEvent{Tx: sigTx})
	}
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
		txHash, err := p.dag.GetTxHashByReqId(ctx.rstTx.RequestHash())
		if err == nil && txHash != (common.Hash{}) {
			log.Info("AddContractLoop", "transaction request Id already in dag", ctx.rstTx.RequestHash())
			continue
		}
		tx, err := gen.GenContractSigTransction(addr, "", ctx.rstTx, ks)
		if err != nil {
			log.Error("AddContractLoop GenContractSigTransctions", "error", err.Error())
			continue
		}
		if false == checkTxValid(ctx.rstTx) {
			log.Error("AddContractLoop recv event Tx is invalid,", "txid", ctx.rstTx.RequestHash().String())
			continue
		}

		if err = txpool.AddLocal(txspool.TxtoTxpoolTx(txpool, tx)); err != nil {
			log.Error("AddContractLoop", "error", err.Error())
			continue
		}
		log.Debug("AddContractLoop", "Tx reqId", tx.RequestHash().String(), "Tx hash", tx.Hash().String())
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
	if p.nodeContractExecutable(p.local, tx) != true {
		log.Error("CheckContractTxValid", "nodeContractExecutable false")
		return false
	}

	ctx, ok := p.mtx[reqId]
	if ctx != nil && (ctx.valid == false || ctx.executable == false) {
		log.Debug("CheckContractTxValid ctx != nil && (ctx.valid == false || ctx.executable == false)")
		return false
	}
	if ok && ctx.rstTx != nil {
		log.Debug("CheckContractTxValid ok && ctx.rstTx != nil") // todo del
		return msgsCompare(ctx.rstTx.TxMessages, tx.TxMessages, modules.APP_CONTRACT_INVOKE)
	} else {
		log.Debug("CheckContractTxValid  ctx.rstTx == nil") //todo del
		_, msgs, err := runContractCmd(p.dag, p.contract, tx)
		if err != nil {
			log.Error("CheckContractTxValid runContractCmd", "error", err.Error())
			return false
		}
		p.mtx[reqId].valid = false
		return msgsCompare(msgs, tx.TxMessages, modules.APP_CONTRACT_INVOKE)
	}
}

func (p *Processor) nodeContractExecutable(accounts map[common.Address]*JuryAccount /*addrs []common.Address*/ , tx *modules.Transaction) bool {
	if tx == nil {
		return false
	}
	sysContract := isSystemContract(tx)
	if sysContract { //system contract
		for addr, _ := range accounts {
			if p.ptn.IsLocalActiveMediator(addr) {
				log.Debug("nodeContractExecutable", "Mediator, true:tx requestId", tx.RequestHash())
				return true
			}
		}
	} else { //usr contract
		log.Debug("User contract, call docker to run contract.")
		for addr, _ := range accounts {
			if true == p.isLocalActiveJury(addr) {
				log.Debug("nodeContractExecutable", "Jury, true:tx requestId", tx.RequestHash())
				return true
			}
		}
	}
	log.Debug("nodeContractExecutable", "false:tx requestId", tx.RequestHash())

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

func (p *Processor) ContractTxBroadcast(txBytes []byte) ([]byte, error) {
	if txBytes == nil {
		log.Error("ContractTxBroadcast", "param is nil")
		return nil, errors.New("transaction request param is nil")
	}
	log.Info("ContractTxBroadcast enter")

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
	go p.ptn.ContractBroadcast(ContractExeEvent{Tx: tx})

	return req[:], nil
}

func (p *Processor) createContractTxReq(from, to common.Address, daoAmount, daoFee uint64, msg *modules.Message, isLocalInstall bool) ([]byte, *modules.Transaction, error) {
	tx, _, err := p.dag.CreateGenericTransaction(from, to, daoAmount, daoFee, msg, p.ptn.TxPool())
	if err != nil {
		return nil, nil, err
	}
	log.Debug("createContractTxReq", "tx:", tx)
	if tx, err = p.ptn.SignGenericTransaction(from, tx); err != nil {
		return nil, nil, err
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
	ctx := p.mtx[reqId]
	if isLocalInstall {
		if err = p.runContractReq(ctx); err != nil {
			return nil, nil, err
		}
		account := p.getLocalAccount()
		if account == nil {
			return nil, nil, errors.New("createContractTxReq no local account")
		}
		ctx.rstTx, err = gen.GenContractSigTransction(account.Address, account.Password, ctx.rstTx, p.ptn.GetKeyStore())
		if err != nil {
			return nil, nil, err
		}
		tx = ctx.rstTx
	} else if p.nodeContractExecutable(p.local, tx) == true {
		go p.runContractReq(ctx)
	}
	//broadcast
	//go p.ptn.ContractBroadcast(ContractExeEvent{Tx: tx})

	//local
	//go p.contractExecFeed.Send(ContractExeEvent{modules.NewTransaction([]*modules.Message{msgPay, msgReq})})
	//go p.ProcessContractEvent(&ContractExeEvent{Tx: tx})
	return reqId[:], tx, nil
}

func (p *Processor) ContractTxDeleteLoop() {
	for {
		time.Sleep(time.Second * time.Duration(5))
		p.locker.Lock()
		for k, v := range p.mtx {
			if v.valid == false {
				if time.Since(v.tm) > time.Second*120 {
					log.Info("ContractTxDeleteLoop, contract is invalid", "delete tx id", k.String())
					delete(p.mtx, k)
				}
			} else {
				if time.Since(v.tm) > time.Second*600 {
					log.Info("ContractTxDeleteLoop, contract is valid", "delete tx id", k.String())
					delete(p.mtx, k)
				}
			}
		}
		p.locker.Unlock()
	}
}

func (p *Processor) getLocalAccount() *JuryAccount {
	//todo 这里默认取其中一个，实际配置只有一个
	var account *JuryAccount
	for _, account = range p.local {
		break
	}
	return account
}
