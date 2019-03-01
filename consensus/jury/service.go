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
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"sync"
	"time"

	"github.com/dedis/kyber"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/event"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/p2p"
	"github.com/palletone/go-palletone/common/util"
	"github.com/palletone/go-palletone/contracts"
	"github.com/palletone/go-palletone/core/accounts/keystore"
	"github.com/palletone/go-palletone/core/gen"
	"github.com/palletone/go-palletone/dag/errors"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/txspool"
	"github.com/palletone/go-palletone/validator"
)

type PeerType = uint8

const (
	TJury     = 2
	TMediator = 4
)

type PalletOne interface {
	GetKeyStore() *keystore.KeyStore
	TxPool() txspool.ITxPool

	MockContractLocalSend(event ContractEvent)
	ContractBroadcast(event ContractEvent, local bool)
	ElectionBroadcast(event ElectionEvent)

	GetLocalMediators() []common.Address
	IsLocalActiveMediator(add common.Address) bool

	SignGenericTransaction(from common.Address, tx *modules.Transaction) (*modules.Transaction, error)
}

type iDag interface {
	GetTxFee(pay *modules.Transaction) (*modules.AmountAsset, error)
	GetUtxoEntry(outpoint *modules.OutPoint) (*modules.Utxo, error)
	GetAddrByOutPoint(outPoint *modules.OutPoint) (common.Address, error)
	GetActiveMediators() []common.Address
	GetTxHashByReqId(reqid common.Hash) (common.Hash, error)
	IsActiveJury(add common.Address) bool
	IsActiveMediator(add common.Address) bool
	GetAddr1TokenUtxos(addr common.Address, asset *modules.Asset) (map[modules.OutPoint]*modules.Utxo, error)
	CreateGenericTransaction(from, to common.Address, daoAmount, daoFee uint64,
		msg *modules.Message, txPool txspool.ITxPool) (*modules.Transaction, uint64, error)
	CreateTokenTransaction(from, to, toToken common.Address, daoAmount, daoFee, daoAmountToken uint64, assetToken string,
		msg *modules.Message, txPool txspool.ITxPool) (*modules.Transaction, uint64, error)
	GetTransaction(hash common.Hash) (*modules.Transaction, common.Hash, uint64, uint64, error)
	GetTransactionOnly(hash common.Hash) (*modules.Transaction, error)
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

type electionInfo struct {
	contractId common.Address

	eleChan  chan bool
	eleNum   uint //选举jury的数量
	seedData []byte
}

type contractTx struct {
	state    int                    //contract run state, 0:default, 1:running
	addrHash []common.Hash          //dynamic
	reqTx    *modules.Transaction   //request contract
	rstTx    *modules.Transaction   //contract run result---system
	sigTx    *modules.Transaction   //contract sig result---user, 0:local, 1,2 other
	rcvTx    []*modules.Transaction //the local has not received the request contract, the cache has signed the contract
	tm       time.Time              //create time
	valid    bool                   //contract request valid identification
	eleInfo  electionInfo           //vrf election jury list
}

type Processor struct {
	name      string //no user
	ptn       PalletOne
	dag       iDag
	validator validator.Validator
	contract  *contracts.Contract
	vrfAct    vrfAccount
	local     map[common.Address]*JuryAccount  //[]common.Address //local jury account addr
	mtx       map[common.Hash]*contractTx      //all contract buffer
	lockAddr  map[common.Address][]common.Hash //contractId/deployId ----addrHash, jury VRF
	quit      chan struct{}
	locker    *sync.Mutex

	electionNum       int
	contractSigNum    int
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

	c := elliptic.P256()
	key, err := ecdsa.GenerateKey(c, rand.Reader)
	if err != nil {
		return nil, errors.New("NewContractProcessor, GenerateKey fail")
	}
	va := vrfAccount{
		priKey: key,
		pubKey: &key.PublicKey,
	}

	validator := validator.NewValidate(dag, dag, nil)
	p := &Processor{
		name:           "conractProcessor",
		ptn:            ptn,
		dag:            dag,
		contract:       contract,
		vrfAct:         va,
		local:          accounts,
		locker:         new(sync.Mutex),
		quit:           make(chan struct{}),
		mtx:            make(map[common.Hash]*contractTx),
		lockAddr:       make(map[common.Address][]common.Hash),
		electionNum:    cfg.ElectionNum,
		contractSigNum: cfg.ContractSigNum,
		validator:      validator,
	}

	log.Info("NewContractProcessor ok", "local address:", p.local)
	log.Info("NewContractProcessor", "vrf Account publicKey", p.vrfAct.pubKey, "privateKey", p.vrfAct.priKey)
	log.Info("NewContractProcessor", "electionNum:", p.electionNum)
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

func (p *Processor) runContractReq(reqId common.Hash) error {
	req := p.mtx[reqId]
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

		if getTxSigNum(req.sigTx) >= p.contractSigNum {
			if localIsMinSignature(req.sigTx) {
				go p.ptn.ContractBroadcast(ContractEvent{AddrHash: req.addrHash, CType: CONTRACT_EVENT_COMMIT, Tx: req.sigTx}, true)
				return nil
			}
		}
		//广播
		go p.ptn.ContractBroadcast(ContractEvent{AddrHash: req.addrHash, CType: CONTRACT_EVENT_SIG, Tx: sigTx}, false)
	}
	return nil
}

func (p *Processor) AddContractLoop(txpool txspool.ITxPool, addr common.Address, ks *keystore.KeyStore) error {
	//log.Debug("ProcessContractEvent", "enter", addr.String())
	for _, ctx := range p.mtx {
		if false == ctx.valid {
			continue
		}
		ctx.valid = false
		if isSystemContract(ctx.reqTx) && p.contractEventExecutable(CONTRACT_EVENT_EXEC, ctx.reqTx, nil) {
			if cType, err := getContractTxType(ctx.reqTx); err == nil && cType != modules.APP_CONTRACT_TPL_REQUEST {
				if p.runContractReq(ctx.reqTx.RequestHash()) != nil {
					continue
				}
			}
		}
		if ctx.rstTx == nil {
			continue
		}

		if !p.checkTxValid(ctx.rstTx) {
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
		//if false == checkTxValid(ctx.rstTx) {
		//	log.Error("AddContractLoop recv event Tx is invalid,", "txid", ctx.rstTx.RequestHash().String())
		//	continue
		//}

		if err = txpool.AddLocal(txspool.TxtoTxpoolTx(txpool, tx)); err != nil {
			log.Error("AddContractLoop", "error", err.Error())
			continue
		}
		log.Debug("AddContractLoop", "Tx reqId", tx.RequestHash().String(), "Tx hash", tx.Hash().String())
	}
	return nil
}

func (p *Processor) CheckContractTxValid(tx *modules.Transaction, execute bool) bool {
	if tx == nil {
		log.Error("CheckContractTxValid", "param is nil")
		return false
	}
	log.Debug("CheckContractTxValid", "reqId:", tx.RequestHash().String(), "exec:", execute)
	if !p.checkTxValid(tx) {
		log.Error("CheckContractTxValid checkTxValid fail")
		return false
	}
	if !execute || !isSystemContract(tx) { //不执行合约或者用户合约
		return true
	}
	//只检查invoke类型
	if txType, err := getContractTxType(tx); err == nil {
		if txType != modules.APP_CONTRACT_INVOKE_REQUEST {
			return true
		}
	}
	//检查本阶段时候有合约执行权限
	if !p.contractEventExecutable(CONTRACT_EVENT_EXEC, tx, nil) {
		log.Error("CheckContractTxValid", "nodeContractExecutable false")
		return false
	}
	_, msgs, err := runContractCmd(p.dag, p.contract, tx) // long time ...
	if err != nil {
		log.Error("CheckContractTxValid runContractCmd", "error", err.Error())
		return false
	}
	return msgsCompare(msgs, tx.TxMessages, modules.APP_CONTRACT_INVOKE)
}

func (p *Processor) IsSystemContractTx(tx *modules.Transaction) bool {
	return isSystemContract(tx)
}

func (p *Processor) isInLocalAddr(addrHash []common.Hash) bool {
	if len(addrHash) <= 0 {
		return true
	}
	for _, hs := range addrHash {
		for addr, _ := range p.local {
			if hs == util.RlpHash(addr) {
				return true
			}
		}
	}
	return false
}

func (p *Processor) contractEventExecutable(event ContractEventType, tx *modules.Transaction, addrHash []common.Hash) bool {
	if tx == nil {
		return false
	}
	isSysContract := isSystemContract(tx)
	isMediator, isJury := func(acs map[common.Address]*JuryAccount) (isM bool, isJ bool) {
		isM = false
		isJ = false
		for addr, _ := range p.local {
			if p.ptn.IsLocalActiveMediator(addr) {
				log.Debug("contractEventExecutable", "is Mediator, addr:", addr.String())
				isM = true
			}
			if true == p.isLocalActiveJury(addr) { //todo
				log.Debug("contractEventExecutable", "is Jury, addr:", addr.String())
				isJ = true
			}
		}
		return isM, isJ
	}(p.local)

	switch event {
	case CONTRACT_EVENT_EXEC:
		if isSysContract && isMediator {
			log.Debug("contractEventExecutable", "CONTRACT_EVENT_EXEC, Mediator, true:tx requestId", tx.RequestHash())
			return true
		} else if !isSysContract && isJury {
			if p.isInLocalAddr(addrHash) {
				log.Debug("contractEventExecutable", "CONTRACT_EVENT_EXEC, Jury, true:tx requestId", tx.RequestHash())
				return true
			} else {
				log.Debug("contractEventExecutable", "CONTRACT_EVENT_EXEC, Jury,not in local addr, false:tx requestId", tx.RequestHash())
			}
		}
	case CONTRACT_EVENT_SIG:
		if !isSysContract && isJury {
			if p.isInLocalAddr(addrHash) {
				log.Debug("contractEventExecutable", "CONTRACT_EVENT_SIG, Jury, true:tx requestId", tx.RequestHash())
				return true
			} else {
				log.Debug("contractEventExecutable", "CONTRACT_EVENT_SIG, Jury, not in local addr,false:tx requestId", tx.RequestHash())
			}
		}
	case CONTRACT_EVENT_COMMIT:
		if isMediator {
			log.Debug("contractEventExecutable", "CONTRACT_EVENT_COMMIT, Mediator, true:tx requestId", tx.RequestHash())
			return true
		}
	}
	return false
}

func (p *Processor) createContractTxReqToken(from, to, toToken common.Address, daoAmount, daoFee, daoAmountToken uint64, assetToken string, msg *modules.Message, isLocalInstall bool) ([]byte, *modules.Transaction, error) {
	tx, _, err := p.dag.CreateTokenTransaction(from, to, toToken, daoAmount, daoFee, daoAmountToken, assetToken, msg, p.ptn.TxPool())
	if err != nil {
		return nil, nil, err
	}
	log.Debug("createContractTxReq", "tx:", tx)

	return p.signAndExecute(getContractIdFromMsg(msg), from, tx, isLocalInstall)
}

func (p *Processor) createContractTxReq(from, to common.Address, daoAmount, daoFee uint64, msg *modules.Message, isLocalInstall bool) ([]byte, *modules.Transaction, error) {
	tx, _, err := p.dag.CreateGenericTransaction(from, to, daoAmount, daoFee, msg, p.ptn.TxPool())
	if err != nil {
		return nil, nil, err
	}
	log.Debug("createContractTxReq", "tx:", tx)

	return p.signAndExecute(getContractIdFromMsg(msg), from, tx, isLocalInstall)
}

func (p *Processor) signAndExecute(contractId common.Address, from common.Address, tx *modules.Transaction, isLocalInstall bool) ([]byte, *modules.Transaction, error) {
	tx, err := p.ptn.SignGenericTransaction(from, tx)
	if err != nil {
		return nil, nil, err
	}
	reqId := tx.RequestHash()
	p.mtx[reqId] = &contractTx{
		reqTx: tx,
		tm:    time.Now(),
		valid: true,
	}
	ctx := p.mtx[reqId]
	if !isSystemContract(tx) && contractId != (common.Address{}) {
		//获取合约Id
		//检查合约Id下是否存在addrHash,并检查数量是否满足要求
		if addrs, ok := p.lockAddr[contractId]; !ok || len(addrs) < p.electionNum {
			p.lockAddr[contractId] = []common.Hash{}                       //清空
			if err = p.ElectionRequest(reqId, time.Second*5); err != nil { //todo ,Single-threaded timeout wait mode
				return nil, nil, err
			}
		}
	}
	if isLocalInstall {
		if err = p.runContractReq(reqId); err != nil {
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
	} else if p.contractEventExecutable(CONTRACT_EVENT_EXEC, tx, ctx.addrHash) && !isSystemContract(tx) {
		go p.runContractReq(reqId)
	}
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
