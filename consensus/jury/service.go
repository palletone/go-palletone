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
	"github.com/palletone/go-palletone/core/accounts/keystore"
	"github.com/palletone/go-palletone/core/gen"
	"github.com/palletone/go-palletone/dag/errors"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/txspool"
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

	GetLocalMediators() []common.Address
	IsLocalActiveMediator(add common.Address) bool

	SignGenericTransaction(from common.Address, tx *modules.Transaction) (*modules.Transaction, error)
}

type iDag interface {
	GetTxFee(pay *modules.Transaction) (*modules.InvokeFees, error)
	GetAddrByOutPoint(outPoint *modules.OutPoint) (common.Address, error)
	GetActiveMediators() []common.Address
	IsActiveJury(add common.Address) bool
	IsActiveMediator(add common.Address) bool
	GetAddr1TokenUtxos(addr common.Address, asset *modules.Asset) (map[modules.OutPoint]*modules.Utxo, error)
	CreateGenericTransaction(from, to common.Address, daoAmount, daoFee uint64,
		msg *modules.Message) (*modules.Transaction, uint64, error)
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

func (p *Processor) SubscribeContractEvent(ch chan<- ContractExeEvent) event.Subscription {
	return p.contractExecScope.Track(p.contractExecFeed.Subscribe(ch))
}

func (p *Processor) isLocalActiveJury(add common.Address) bool {
	if _, ok := p.local[add]; ok {
		return p.dag.IsActiveJury(add)
	}
	return false
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

//todo  对于接收到签名交易，而本地合约还未执行完成的情况后面完善
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
		//todo 这里默认取其中一个，实际配置只有一个
		var account *JuryAccount
		for _, account = range p.local {
			break
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
		log.Debug("AddContractLoop", "AddLocal ok, transaction reqId", tx.RequestHash().String())
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
		return false
	}

	if ok && ctx.rstTx != nil {
		//比较msg
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

//tmp
func (p *Processor) creatContractTxReqBroadcast(from, to common.Address, daoAmount, daoFee uint64, msg *modules.Message) ([]byte, error) {
	tx, _, err := p.dag.CreateGenericTransaction(from, to, daoAmount, daoFee, msg)
	if err != nil {
		return nil, err
	}
	log.Debug("creatContractTxReq", "tx:", tx)

	//tx.AddMessage(msg)
	tx, err = p.ptn.SignGenericTransaction(from, tx)
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
	txHex, _ := rlp.EncodeToBytes(tx)
	log.Debugf("Signed ContractRequest hex:%x", txHex)
	if p.mtx[reqId].executable {
		if p.nodeContractExecutable(p.local, tx) == true {
			go p.runContractReq(p.mtx[reqId])
		}
	}
	//broadcast
	go p.ptn.ContractBroadcast(ContractExeEvent{Tx: tx})
	//local
	//go p.contractExecFeed.Send(ContractExeEvent{modules.NewTransaction([]*modules.Message{msgPay, msgReq})})
	//go p.ProcessContractEvent(&ContractExeEvent{Tx: tx})

	return reqId[:], nil
}

func (p *Processor) ContractInstallReq(from, to common.Address, daoAmount, daoFee uint64, tplName, path, version string) ([]byte, error) {
	if from == (common.Address{}) || to == (common.Address{}) || tplName == "" || path == "" || version == "" {
		log.Error("ContractInstallReq", "param is error")
		return nil, errors.New("ContractInstallReq request param is error")
	}

	log.Debug("ContractInstallReq", "enter, tplName ", tplName, "path", path, "version", version)
	msgReq := &modules.Message{
		App: modules.APP_CONTRACT_TPL_REQUEST,
		Payload: &modules.ContractInstallRequestPayload{
			TplName: tplName,
			Path:    path,
			Version: version,
		},
	}
	return p.creatContractTxReqBroadcast(from, to, daoAmount, daoFee, msgReq)
}

func (p *Processor) ContractDeployReq(from, to common.Address, daoAmount, daoFee uint64, templateId []byte, txid string, args [][]byte, timeout time.Duration) ([]byte, error) {
	if from == (common.Address{}) || to == (common.Address{}) || templateId == nil {
		log.Error("ContractDeployReq", "param is error")
		return nil, errors.New("ContractDeployReq request param is error")
	}
	log.Debug("ContractDeployReq", "enter, templateId ", templateId)

	msgReq := &modules.Message{
		App: modules.APP_CONTRACT_DEPLOY_REQUEST,
		Payload: &modules.ContractDeployRequestPayload{
			TplId:   templateId,
			TxId:    txid,
			Args:    args,
			Timeout: timeout,
		},
	}
	return p.creatContractTxReqBroadcast(from, to, daoAmount, daoFee, msgReq)
}

func (p *Processor) ContractInvokeReq(from, to common.Address, daoAmount, daoFee uint64, contractId common.Address, args [][]byte, timeout time.Duration) ([]byte, error) {
	if from == (common.Address{}) || to == (common.Address{}) || contractId == (common.Address{}) || args == nil {
		log.Error("ContractInvokeReq", "param is error")
		return nil, errors.New("ContractInvokeReq request param is error")
	}

	log.Debug("ContractInvokeReq", "enter, contractId ", contractId)
	msgReq := &modules.Message{
		App: modules.APP_CONTRACT_INVOKE_REQUEST,
		Payload: &modules.ContractInvokeRequestPayload{
			ContractId:   contractId.Bytes(),
			FunctionName: "",
			Args:         args,
			Timeout:      timeout,
		},
	}
	return p.creatContractTxReqBroadcast(from, to, daoAmount, daoFee, msgReq)
}

func (p *Processor) ContractStopReq(from, to common.Address, daoAmount, daoFee uint64, contractId common.Address, txid string, deleteImage bool) ([]byte, error) {
	if from == (common.Address{}) || to == (common.Address{}) || contractId == (common.Address{}) {
		log.Error("ContractStopReq", "param is error")
		return nil, errors.New("ContractStopReq request param is error")
	}

	log.Debug("ContractStopReq", "enter, contractId ", contractId)
	msgReq := &modules.Message{
		App: modules.APP_CONTRACT_STOP_REQUEST,
		Payload: &modules.ContractStopRequestPayload{
			ContractId:  contractId[:],
			Txid:        txid,
			DeleteImage: deleteImage,
		},
	}
	return p.creatContractTxReqBroadcast(from, to, daoAmount, daoFee, msgReq)
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
