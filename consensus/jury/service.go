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
	"math/big"
	"sync"
	"time"
	"fmt"
	"bytes"
	
	"github.com/dedis/kyber"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/crypto"
	"github.com/palletone/go-palletone/common/event"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/p2p"
	"github.com/palletone/go-palletone/common/util"
	"github.com/palletone/go-palletone/contracts"
	"github.com/palletone/go-palletone/core/accounts"
	"github.com/palletone/go-palletone/core/accounts/keystore"
	"github.com/palletone/go-palletone/core/gen"
	"github.com/palletone/go-palletone/dag/errors"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/rwset"
	"github.com/palletone/go-palletone/dag/txspool"
	"github.com/palletone/go-palletone/tokenengine"
	"github.com/palletone/go-palletone/validator"
)

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
	AdapterBroadcast(event AdapterEvent)

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
	IsActiveJury(addr common.Address) bool
	JuryCount() int
	GetActiveJuries() []common.Address
	IsActiveMediator(addr common.Address) bool
	GetAddr1TokenUtxos(addr common.Address, asset *modules.Asset) (map[modules.OutPoint]*modules.Utxo, error)
	CreateGenericTransaction(from, to common.Address, daoAmount, daoFee uint64, certID *big.Int,
		msg *modules.Message, txPool txspool.ITxPool) (*modules.Transaction, uint64, error)
	CreateTokenTransaction(from, to, toToken common.Address, daoAmount, daoFee, daoAmountToken uint64, assetToken string,
		msg *modules.Message, txPool txspool.ITxPool) (*modules.Transaction, uint64, error)
	GetTransaction(hash common.Hash) (*modules.TransactionWithUnitInfo, error)
	GetTransactionOnly(hash common.Hash) (*modules.Transaction, error)
	GetHeaderByHash(common.Hash) (*modules.Header, error)
	GetContractState(id []byte, field string) ([]byte, *modules.StateVersion, error)
	GetTxRequesterAddress(tx *modules.Transaction) (common.Address, error)
	GetConfig(name string) ([]byte, *modules.StateVersion, error)
	IsTransactionExist(hash common.Hash) (bool, error)
}

type Juror struct {
	name        string
	address     common.Address
	initPartPub kyber.Point
}

//合约节点类型、地址信息
type nodeInfo struct {
	addr  common.Address
	ntype int //1:default, 2:jury, 4:mediator
}

type electionVrf struct {
	eChan chan bool             //election event chan
	eInf  []modules.ElectionInf //receive ele info already

	rst []common.Hash //receive result event address hash
	req bool          //receive request event
	tm  time.Time
}

type contractTx struct {
	state    int                    //contract run state, 0:default, 1:running
	addrHash []common.Hash          //dynamic
	eleInf   []modules.ElectionInf  //dynamic
	reqTx    *modules.Transaction   //request contract
	rstTx    *modules.Transaction   //contract run result---system
	sigTx    *modules.Transaction   //contract sig result---user, 0:local, 1,2 other,signature is the same as local value
	rcvTx    []*modules.Transaction //the local has not received the request contract, the cache has signed the contract
	tm       time.Time              //create time
	valid    bool                   //contract request valid identification
	adaChan  chan bool              //adapter event chan
	adaInf   map[uint32]*AdapterInf //adapter event data information
}

type Processor struct {
	name      string //no use
	ptn       PalletOne
	dag       iDag
	validator validator.Validator
	contract  *contracts.Contract
	local     map[common.Address]*JuryAccount          //[]common.Address //local jury account addr
	mtx       map[common.Hash]*contractTx              //all contract buffer
	mel       map[common.Hash]*electionVrf             //election vrf inform
	lockVrf   map[common.Address][]modules.ElectionInf //contractId/deployId ----vrfInfo, jury VRF
	quit      chan struct{}
	locker    *sync.Mutex
	//vrfAct    vrfAccount
	errMsgEnable      bool //package contract execution error information into the transaction
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
	acs := make(map[common.Address]*JuryAccount, 0)
	for _, cfg := range cfg.Accounts {
		account := cfg.configToAccount()
		if account != nil {
			err := ptn.GetKeyStore().Unlock(accounts.Account{Address: account.Address}, account.Password)
			if err == nil {
				addr := account.Address
				acs[addr] = account
			}
		}
	}

	validator := validator.NewValidate(dag, dag, nil)
	p := &Processor{
		name:           "contractProcessor",
		ptn:            ptn,
		dag:            dag,
		contract:       contract,
		local:          acs,
		locker:         new(sync.Mutex),
		quit:           make(chan struct{}),
		mtx:            make(map[common.Hash]*contractTx),
		mel:            make(map[common.Hash]*electionVrf),
		lockVrf:        make(map[common.Address][]modules.ElectionInf),
		electionNum:    cfg.ElectionNum,
		contractSigNum: cfg.ContractSigNum,
		validator:      validator,
		errMsgEnable:   true,
	}

	log.Info("NewContractProcessor ok", "local address:", p.local, "electionNum", p.electionNum)
	//log.Info("NewContractProcessor", "vrf Account publicKey", p.vrfAct.pubKey, "privateKey", p.vrfAct.priKey)
	return p, nil
}
func (p *Processor) SetContract(contract *contracts.Contract) {
	p.contract = contract
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

func (p *Processor) runContractReq(reqId common.Hash, elf []modules.ElectionInf) error {
	req := p.mtx[reqId]
	if req == nil {
		return errors.New("runContractReq param is nil")
	}
	msgs, err := runContractCmd(rwset.RwM, p.dag, p.contract, req.reqTx, elf, p.errMsgEnable)
	if err != nil {
		log.Error("runContractReq", "runContractCmd reqTx", req.reqTx.RequestHash().String(), "error", err.Error())
		return err
	}
	tx, err := gen.GenContractTransction(req.reqTx, msgs)
	if err != nil {
		log.Error("runContractReq", "GenContractSigTransactions error", err.Error())
		return err
	}

	//如果系统合约，直接添加到缓存池
	//如果用户合约，需要签名，添加到缓存池并广播
	if tx.IsSystemContract() {
		req.rstTx = tx
	} else {
		account := p.getLocalAccount()
		if account == nil {
			log.Error("runContractReq", "not find local account, reqId", reqId)
			return errors.New("runContractReq no local account")
		}
		sigTx, err := p.GenContractSigTransaction(account.Address, account.Password, tx, p.ptn.GetKeyStore())
		if err != nil {
			log.Error("runContractReq", "GenContractSigTransctions error", err.Error())
			return errors.New("")
		}
		req.sigTx = sigTx
		//如果rcvTx存在，则比较执行结果，并将结果附加到sigTx上,并删除rcvTx
		if len(req.rcvTx) > 0 {
			for _, rtx := range req.rcvTx {
				ok, err := checkAndAddTxSigMsgData(req.sigTx, rtx)
				if err != nil {
					log.Debug("runContractReq", "checkAndAddTxSigMsgData error", err.Error())
				} else if ok {
					log.Debug("runContractReq", "checkAndAddTxSigMsgData ok, reqId", reqId)
				} else {
					log.Debug("runContractReq", "checkAndAddTxSigMsgData fail, reqId", reqId)
				}
			}
			req.rcvTx = nil
		}

		if getTxSigNum(req.sigTx) >= p.contractSigNum {
			if localIsMinSignature(req.sigTx) {
				log.Info("runContractReq", "localIsMinSignature Ok!, reqId", reqId)
				go p.ptn.ContractBroadcast(ContractEvent{Ele: req.eleInf, CType: CONTRACT_EVENT_COMMIT, Tx: req.sigTx}, true)
				return nil
			}
		}
		//广播
		go p.ptn.ContractBroadcast(ContractEvent{Ele: req.eleInf, CType: CONTRACT_EVENT_SIG, Tx: sigTx}, false)
	}
	return nil
}

func (p *Processor) GenContractSigTransaction(singer common.Address, password string, orgTx *modules.Transaction, ks *keystore.KeyStore) (*modules.Transaction, error) {
	if orgTx == nil || len(orgTx.TxMessages) < 3 {
		return nil, errors.New(fmt.Sprintf("GenContractSigTransctions param is error"))
	}
	if password != "" {
		err := ks.Unlock(accounts.Account{Address: singer}, password)
		if err != nil {
			return nil, err
		}
	}
	tx := orgTx
	needSignMsg := true
	//Find contract pay out payment messages
	resultMsg := false
	for msgidx, msg := range tx.TxMessages {
		if msg.App == modules.APP_CONTRACT_INVOKE_REQUEST {
			resultMsg = true
			continue
		}
		if resultMsg {
			if msg.App == modules.APP_PAYMENT {
				//Contract result里面的Payment只有2种，创币或者从合约付出，
				payment := msg.Payload.(*modules.PaymentPayload)
				if !payment.IsCoinbase() {
					//Contract Payout, need sign
					needSignMsg = false
					pubKey, _ := ks.GetPublicKey(singer)
					redeemScript := tokenengine.GenerateRedeemScript(1, [][]byte{pubKey})
					log.Debugf("RedeemScript:%x", redeemScript)
					for inputIdx, input := range payment.Inputs {
						utxo, err := p.dag.GetUtxoEntry(input.PreviousOutPoint)
						if err != nil {
							return nil, err
						}
						log.Debugf("Lock script:%x", utxo.PkScript)
						sign, err := tokenengine.MultiSignOnePaymentInput(tx, tokenengine.SigHashAll, msgidx, inputIdx, utxo.PkScript, redeemScript, ks.GetPublicKey, ks.SignHash, nil)
						if err != nil {
							log.Errorf("Sign error:%s", err)
						}
						log.Debugf("Sign a contract payout payment,tx[%s],sign:%x", tx.Hash().String(), sign)
						input.SignatureScript = sign
					}
				}
			}
		}
	}
	if needSignMsg {
		//没有Contract Payout的情况下，那么需要单独附加Signature Message
		pubKey, err := ks.GetPublicKey(singer)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("GenContractSigTransctions GetPublicKey fail, address[%s]", singer.String()))
		}
		sig, err := GetTxSig(tx, ks, singer)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("GenContractSigTransctions GetTxSig fail, address[%s], tx[%s]", singer.String(), orgTx.RequestHash().String()))
		}
		sigSet := modules.SignatureSet{
			PubKey:    pubKey,
			Signature: sig,
		}
		SigPayload, err := getContractTxContractInfo(tx, modules.APP_SIGNATURE)
		if SigPayload != nil {
			SigPayload.(*modules.SignaturePayload).Signatures = append(SigPayload.(*modules.SignaturePayload).Signatures, sigSet)
		} else {
			sigs := make([]modules.SignatureSet, 0)
			sigs = append(sigs, sigSet)
			msgSig := &modules.Message{
				App: modules.APP_SIGNATURE,
				Payload: &modules.SignaturePayload{
					Signatures: sigs,
				},
			}
			tx.TxMessages = append(tx.TxMessages, msgSig)
		}
		log.Debug("GenContractSigTransactions", "orgTx.TxId id ok:", tx.Hash())
	}
	return tx, nil
}
func GetTxSig(tx *modules.Transaction, ks *keystore.KeyStore, signer common.Address) ([]byte, error) {
	sign, err := ks.SigData(tx, signer)
	if err != nil {
		msg := fmt.Sprintf("Failed to singure transaction:%v", err)
		log.Error(msg)
		return nil, errors.New(msg)
	}

	return sign, nil
}
func (p *Processor) AddContractLoop(rwM rwset.TxManager, txpool txspool.ITxPool, addr common.Address, ks *keystore.KeyStore) error {
	//log.Debug("AddContractLoop", "loop", addr.String())
	setChainId := "palletone"
	for _, ctx := range p.mtx {
		if false == ctx.valid || ctx.reqTx == nil {
			continue
		}
		reqhash := ctx.reqTx.RequestHash()
		if !ctx.reqTx.IsSystemContract() {
			defer rwM.CloseTxSimulator(setChainId, reqhash.String())
		}
		log.Debug("AddContractLoop", "enter mtx", addr.String())
		ctx.valid = false
		if ctx.reqTx.IsSystemContract() && p.contractEventExecutable(CONTRACT_EVENT_EXEC, ctx.reqTx, nil) {
			if cType, err := getContractTxType(ctx.reqTx); err == nil && cType != modules.APP_CONTRACT_TPL_REQUEST {
				if p.checkTxReqIdIsExist(reqhash) {
					log.Debug("AddContractLoop ,checkTxReqIdIsExist is ok", "reqId", reqhash.String())
					continue
				}
				if p.runContractReq(reqhash, nil) != nil {
					continue
				}
			}
		}
		if ctx.rstTx == nil {
			continue
		}
		tx, err := p.GenContractSigTransaction(addr, "", ctx.rstTx, ks)
		if err != nil {
			log.Error("AddContractLoop GenContractSigTransctions", "error", err.Error())
			continue
		}

		//if !p.checkTxValid(ctx.rstTx) {
		//	log.Error("AddContractLoop recv event Tx is invalid,", "txid", ctx.rstTx.RequestHash().String())
		//	continue
		//}
		txHash, err := p.dag.GetTxHashByReqId(ctx.rstTx.RequestHash())
		if err == nil && txHash != (common.Hash{}) {
			log.Info("AddContractLoop", "transaction request Id already in dag", ctx.rstTx.RequestHash())
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
	rwM.Close()
	return nil
}

func (p *Processor) CheckContractTxValid(rwM rwset.TxManager, tx *modules.Transaction, execute bool) bool {
	if tx == nil {
		log.Error("CheckContractTxValid, param is nil")
		return false
	}
	log.Debug("CheckContractTxValid", "reqId:", tx.RequestHash().String(), "exec:", execute)
	if !execute || !tx.IsSystemContract() {
		//不执行合约或者用户合约
		return true
	}
	if !p.checkTxValid(tx) {
		log.Error("CheckContractTxValid checkTxValid fail")
		return false
	}
	//只检查invoke类型
	if txType, err := getContractTxType(tx); err == nil {
		if txType != modules.APP_CONTRACT_INVOKE_REQUEST {
			return true
		}
	}
	//检查本阶段时候有合约执行权限
	if !p.contractEventExecutable(CONTRACT_EVENT_EXEC, tx, nil) {
		log.Error("CheckContractTxValid, nodeContractExecutable false")
		return false
	}
	msgs, err := runContractCmd(rwM, p.dag, p.contract, tx, nil, p.errMsgEnable) // long time ...
	if err != nil {
		log.Error("CheckContractTxValid runContractCmd", "error", err.Error())
		return false
	}
	return msgsCompare(msgs, tx.TxMessages, modules.APP_CONTRACT_INVOKE)
}

func (p *Processor) IsSystemContractTx(tx *modules.Transaction) bool {
	return tx.IsSystemContract()
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

func (p *Processor) isValidateElection(tx *modules.Transaction, ele []modules.ElectionInf, checkExit bool) bool {
	if len(ele) < p.electionNum {
		log.Info("isValidateElection, ElectionInf number not enough ", "len(ele)=", len(ele), "set electionNum=", p.electionNum)
		return false
	}
	reqId := tx.ContractIdBytes()
	reqAddr, err := p.dag.GetTxRequesterAddress(tx)
	if err != nil {
		log.Error("isValidateElection", "GetTxRequesterAddress fail, reqId", reqId, "err", err)
		return false
	}
	isExit := false
	etor := &elector{
		num:   uint(p.electionNum),
		total: uint64(p.dag.JuryCount()), //todo from dag
	}
	etor.weight = electionWeightValue(etor.total)
	for i, e := range ele {
		isVerify := false
		//检查地址hash是否在本地
		if checkExit && !isExit {
			for addr, _ := range p.local {
				log.Debug("isValidateElection", "local addr", addr, "hash",util.RlpHash(addr).Bytes(), "Hash:",util.RlpHash(addr))
				log.Debug("isValidateElection", "addrHash", e.AddrHash, "ADDRHASH", e.AddrHash.Bytes())
				if bytes.Equal(e.AddrHash.Bytes(), util.RlpHash(addr).Bytes()) {
					isExit = true
					break
				}
			}
		}
		//检查指定节点模式下，是否为jjh请求地址
		if e.Etype == 1 {
			jjhAd, _, err := p.dag.GetConfig("FoundationAddress")
			if err == nil && bytes.Equal(reqAddr[:], jjhAd) {
				log.Debug("isValidateElection", "e.Etype == 1, ok, reqId", reqId)
				continue
			} else {
				log.Debug("isValidateElection", "e.Etype == 1, but not jjh request addr, reqId", reqId)
				log.Debug("isValidateElection", "reqAddr", reqAddr[:], "jjh", jjhAd)

				continue //todo test
				//return false
			}
		}
		//检查地址与pubKey是否匹配:获取当前pubKey下的Addr，将地址hash后与输入比较
		addr := crypto.PubkeyBytesToAddress(e.PublicKey)
		if e.AddrHash != util.RlpHash(addr) {
			log.Error("isValidateElection", "publicKey not match address, addrHash is", e.AddrHash)
			return false
		}
		//从数据库中查询该地址是否为Jury
		if !p.dag.IsActiveJury(addr) {
			log.Error("isValidateElection", "not active Jury, addrHash is", e.AddrHash)
			return false
		}
		//验证proof是否通过
		isVerify, err := etor.verifyVrf(e.Proof, conversionElectionSeedData(reqId), e.PublicKey)
		if err != nil || !isVerify {
			log.Info("isValidateElection", "index", i, "verifyVrf fail, reqId", reqId)
			return false
		}
	}
	if checkExit {
		if !isExit {
			log.Debug("isValidateElection, election addr not in local")
			return false
		}
	}
	return true
}

func (p *Processor) contractEventExecutable(event ContractEventType, tx *modules.Transaction, ele []modules.ElectionInf) bool {
	if tx == nil {
		return false
	}
	isSysContract := tx.IsSystemContract()
	isMediator, isJury := func(acs map[common.Address]*JuryAccount) (isM bool, isJ bool) {
		isM = false
		isJ = false
		for addr, _ := range p.local {
			if p.ptn.IsLocalActiveMediator(addr) {
				log.Debug("contractEventExecutable", "is Mediator, addr:", addr.String())
				isM = true
			}
			if true == p.isLocalActiveJury(addr) {
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
			if p.isValidateElection(tx, ele, true) {
				log.Debug("contractEventExecutable", "CONTRACT_EVENT_EXEC, Jury, true:tx requestId", tx.RequestHash())
				return true
			} else {
				log.Debug("contractEventExecutable", "CONTRACT_EVENT_EXEC, Jury, isValidateElection fail, false:tx requestId", tx.RequestHash())
			}
		}
	case CONTRACT_EVENT_SIG:
		if !isSysContract && isJury {
			if p.isValidateElection(tx, ele, false) {
				log.Debug("contractEventExecutable", "CONTRACT_EVENT_SIG, Jury, true:tx requestId", tx.RequestHash())
				return true
			} else {
				log.Debug("contractEventExecutable", "CONTRACT_EVENT_SIG, Jury, isValidateElection fail, false:tx requestId", tx.RequestHash())
			}
		}
	case CONTRACT_EVENT_COMMIT:
		if isMediator {
			if isSysContract {
				log.Debug("contractEventExecutable", "CONTRACT_EVENT_COMMIT, Mediator, sysContract, true:tx requestId", tx.RequestHash())
				return true
			} else if !isSysContract && p.isValidateElection(tx, ele, false) {
				log.Debug("contractEventExecutable", "CONTRACT_EVENT_COMMIT, Mediator, userContract, true:tx requestId", tx.RequestHash())
				return true
			} else {
				log.Debug("contractEventExecutable", "CONTRACT_EVENT_COMMIT, Mediator, isValidateElection fail, false:tx requestId", tx.RequestHash())
			}
		}
	}
	return false
}

func (p *Processor) createContractTxReqToken(from, to, toToken common.Address, daoAmount, daoFee, daoAmountToken uint64, assetToken string, msg *modules.Message, isLocalInstall bool) (common.Hash, *modules.Transaction, error) {
	tx, _, err := p.dag.CreateTokenTransaction(from, to, toToken, daoAmount, daoFee, daoAmountToken, assetToken, msg, p.ptn.TxPool())
	if err != nil {
		return common.Hash{}, nil, err
	}
	log.Debug("createContractTxReq", "tx:", tx)

	return p.signAndExecute(common.BytesToAddress(tx.RequestHash().Bytes()), from, tx, isLocalInstall)
}

func (p *Processor) createContractTxReq(contractId, from, to common.Address, daoAmount, daoFee uint64, certID *big.Int, msg *modules.Message, isLocalInstall bool) (common.Hash, *modules.Transaction, error) {
	tx, _, err := p.dag.CreateGenericTransaction(from, to, daoAmount, daoFee, certID, msg, p.ptn.TxPool())
	if err != nil {
		return common.Hash{}, nil, err
	}
	log.Debug("createContractTxReq", "contractId", contractId, "tx:", tx)
	return p.signAndExecute(contractId, from, tx, isLocalInstall)
}

func (p *Processor) signAndExecute(contractId common.Address, from common.Address, tx *modules.Transaction, isLocalInstall bool) (common.Hash, *modules.Transaction, error) {
	tx, err := p.ptn.SignGenericTransaction(from, tx)
	if err != nil {
		return common.Hash{}, nil, err
	}
	reqId := tx.RequestHash()
	p.mtx[reqId] = &contractTx{
		reqTx:  tx.GetRequestTx(),
		tm:     time.Now(),
		valid:  true,
		adaInf: make(map[uint32]*AdapterInf),
	}
	ctx := p.mtx[reqId]
	if !tx.IsSystemContract() {
		//获取合约Id
		//检查合约Id下是否存在addrHash,并检查数量是否满足要求
		if contractId == (common.Address{}) { //deploy
			cId := common.NewAddress(common.BytesToAddress(reqId.Bytes()).Bytes(), common.ContractHash)
			//if ele, ok := p.lockVrf[cId]; !ok || len(ele) < p.electionNum {
			//	p.lockVrf[cId] = []modules.ElectionInf{} //清空
			//	if err = p.ElectionRequest(reqId, time.Second*5); err != nil { //todo ,Single-threaded timeout wait mode
			//		return common.Hash{}, nil, err
			//	}
			//}
			//ctx.eleInf = p.lockVrf[cId]
			elist, err := p.genContractElectionList(tx, cId)
			if err != nil {
				return common.Hash{}, nil, err
			}
			ctx.eleInf = elist
		} else { //invoke,stop
			elist, err := p.getContractElectionList(contractId)
			if err != nil {
				return common.Hash{}, nil, err
			}
			ctx.eleInf = elist
			//ctx.eleInf = p.lockArf[contractId]
		}
	}
	if isLocalInstall {
		if err = p.runContractReq(reqId, nil); err != nil {
			return common.Hash{}, nil, err
		}
		account := p.getLocalAccount()
		if account == nil {
			return common.Hash{}, nil, errors.New("createContractTxReq no local account")
		}
		ctx.rstTx, err = p.GenContractSigTransaction(account.Address, account.Password, ctx.rstTx, p.ptn.GetKeyStore())
		if err != nil {
			return common.Hash{}, nil, err
		}
		tx = ctx.rstTx
	} else if p.contractEventExecutable(CONTRACT_EVENT_EXEC, tx, ctx.eleInf) && !tx.IsSystemContract() {
		go p.runContractReq(reqId, ctx.eleInf)
	}

	return reqId, tx, nil
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
		for k, v := range p.mel {
			if time.Since(v.tm) > time.Second*30 {
				log.Info("ContractTxDeleteLoop", "delete electionVrf,  id", k.String())
				delete(p.mel, k)
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

func (p *Processor) getContractElectionList(contractId common.Address) ([]modules.ElectionInf, error) {
	eleByte, _, err := p.dag.GetContractState(contractId[:], "ElectionList")
	if err != nil {
		log.Debug("getContractElectionList", "not find contract election, contractId", contractId)
		return nil, err
	}
	var ele []modules.ElectionInf
	err = rlp.DecodeBytes(eleByte, &ele)
	if err != nil {
		errs := fmt.Sprintf("getContractElectionList, DecodeBytes fail, contractId:%v", contractId)
		log.Debug(errs)
		return nil, errors.New(errs)
	}
	log.Debug("getContractElectionList", "contractId", contractId, "ElectionInf", ele)
	return ele, nil
}

func (p *Processor) getTemplateAddrHash(tplId []byte) ([]common.Hash, error) {
	addrBytes, _, err := p.dag.GetContractState(tplId[:], "TplAddrHash")
	if err != nil {
		log.Debug("getTemplateAddrHash", "not find contract template addrHash, tplId", tplId)
		return nil, err
	}

	var addh []common.Hash
	err = rlp.DecodeBytes(addrBytes, &addh)
	if err != nil {
		log.Debug("getTemplateAddrHash", "err", err)
		errs := fmt.Sprintf("getTemplateAddrHash, DecodeBytes fail, templateId:%v", tplId)
		log.Debug(errs)
		return nil, errors.New(errs)
	}
	log.Debug("getContractElectionList", "templateId", tplId, "addrHash", addh)
	return addh, nil
}

func (p *Processor) genContractElectionList(tx *modules.Transaction, contractId common.Address) ([]modules.ElectionInf, error) {
	if tx == nil {
		return nil, errors.New("genContractElectionList, param is nil")
	}
	reqId := tx.RequestHash()
	payload, err := getContractTxContractInfo(tx, modules.APP_CONTRACT_DEPLOY_REQUEST)
	if err != nil {
		return nil, errors.New("genContractElectionList, getContractTxContractInfo fail")
	}
	num := 0
	eles := make([]modules.ElectionInf, 0)
	tplId := payload.(*modules.ContractDeployRequestPayload).TplId

	//find the address of the contract template binding in the dag
	addrHash, err := p.getTemplateAddrHash(tplId)
	if err != nil {
		log.Debug("genContractElectionList", "getTemplateAddrHash fail,templateId", tplId, "err", err)
	}
	if len(addrHash) >= p.electionNum {
		num = p.electionNum
	} else {
		num = len(addrHash)
	}
	//add election node form template install assignation
	for i := 0; i < num; i++ {
		e := modules.ElectionInf{Etype: 1, AddrHash: addrHash[i]}
		eles = append(eles, e)
	}
	if len(eles) >= p.electionNum {
		log.Debug("genContractElectionList", "all from dag, ele:", eles)
		return eles, nil
	}
	//add election node form vrf request
	if ele, ok := p.lockVrf[contractId]; !ok || len(ele) < p.electionNum {
		p.lockVrf[contractId] = []modules.ElectionInf{}                 //清空
		if err := p.ElectionRequest(reqId, time.Second*5); err != nil { //todo ,Single-threaded timeout wait mode
			return nil, err
		}
	}
	for _, e := range p.lockVrf[contractId] {
		isExist := false
		for _, a := range eles {
			if e.AddrHash == a.AddrHash {
				isExist = true
				break
			}
		}
		if isExist {
			continue
		}
		eles = append(eles, e)
		if len(eles) >= p.electionNum {
			log.Debug("genContractElectionList", "ele:", eles)
			return eles, nil
		}
	}
	return nil, nil
}
