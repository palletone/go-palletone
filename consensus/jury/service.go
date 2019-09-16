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
	"bytes"
	"fmt"
	"math/big"
	"sync"
	"time"

	"encoding/json"
	"github.com/coocood/freecache"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/crypto"
	"github.com/palletone/go-palletone/common/event"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/p2p"
	"github.com/palletone/go-palletone/common/util"
	"github.com/palletone/go-palletone/contracts"
	"github.com/palletone/go-palletone/core"
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

type PalletOne interface {
	GetKeyStore() *keystore.KeyStore
	TxPool() txspool.ITxPool

	MockContractLocalSend(event ContractEvent)
	ContractBroadcast(event ContractEvent, local bool)
	ElectionBroadcast(event ElectionEvent, local bool)
	AdapterBroadcast(event AdapterEvent)

	LocalHaveActiveMediator() bool
	GetLocalActiveMediators() []common.Address
	SignGenericTransaction(from common.Address, tx *modules.Transaction) (*modules.Transaction, error)
}

type iDag interface {
	GetTxFee(pay *modules.Transaction) (*modules.AmountAsset, error)
	GetUtxoEntry(outpoint *modules.OutPoint) (*modules.Utxo, error)
	GetAddrByOutPoint(outPoint *modules.OutPoint) (common.Address, error)
	GetActiveMediators() []common.Address
	GetTxHashByReqId(reqid common.Hash) (common.Hash, error)
	IsActiveJury(addr common.Address) bool
	JuryCount() uint
	GetContractDevelopers() ([]common.Address, error)
	IsContractDeveloper(addr common.Address) bool
	GetStxoEntry(outpoint *modules.OutPoint) (*modules.Stxo, error)
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
	GetTxRequesterAddress(tx *modules.Transaction) (common.Address, error)
	//GetConfig(name string) ([]byte, *modules.StateVersion, error)
	IsTransactionExist(hash common.Hash) (bool, error)
	GetContractJury(contractId []byte) (*modules.ElectionNode, error)
	GetContractTpl(tplId []byte) (*modules.ContractTemplate, error)
	GetContract(contractId []byte) (*modules.Contract, error)
	//获得系统配置的最低手续费要求
	GetMinFee() (*modules.AmountAsset, error)
	GetContractState(id []byte, field string) ([]byte, *modules.StateVersion, error)
	GetContractStatesByPrefix(id []byte, prefix string) (map[string]*modules.ContractStateValue, error)
	ChainThreshold() int
	GetChainParameters() *core.ChainParameters
	GetMediators() map[common.Address]bool
	GetMediator(add common.Address) *core.Mediator
	GetBlacklistAddress() ([]common.Address, *modules.StateVersion, error)

}

type electionVrf struct {
	rcvEle   []modules.ElectionInf //receive
	sigs     []modules.SignatureSet
	tm       time.Time
	juryCnt  uint64
	brded    bool //broadcasted
	nType    byte //ele type  ?   node type, 1:election sig requester. 0:receiver,not process election sig result event
	invalid  bool //this ele invalid
	sigReqEd bool //sig req evt received
	vrfReqEd bool //vrf request event received
}

type contractTx struct {
	eleNode  *modules.ElectionNode  //dynamic
	reqTx    *modules.Transaction   //request contract
	rstTx    *modules.Transaction   //contract run result---system
	sigTx    *modules.Transaction   //contract sig result---user, 0:local, 1,2 other,signature is the same as local value
	rcvTx    []*modules.Transaction //the local has not received the request contract, the cache has signed the contract
	tm       time.Time              //create time
	valid    bool                   //contract request valid identification
	reqRcvEd bool                   //contract request received
	adaInf   map[uint32]*AdapterInf //adapter event data information
}

type Processor struct {
	name      string //no use
	ptn       PalletOne
	dag       iDag
	validator validator.Validator
	contract  *contracts.Contract

	local   map[common.Address]*JuryAccount          //[]common.Address //local jury account addr
	mtx     map[common.Hash]*contractTx              //all contract buffer
	mel     map[common.Hash]*electionVrf             //election vrf inform
	lockVrf map[common.Address][]modules.ElectionInf //contractId/deployId ----vrfInfo, jury VRF

	quit         chan struct{}
	locker       *sync.Mutex //locker       *sync.Mutex  RWMutex
	errMsgEnable bool        //package contract execution error information into the transaction

	electionNum       int
	contractSigNum    int
	contractExecFeed  event.Feed
	contractExecScope event.SubscriptionScope
}

func NewContractProcessor(ptn PalletOne, dag iDag, contract *contracts.Contract, cfg *Config) (*Processor, error) {
	if ptn == nil || dag == nil {
		return nil, errors.New("NewContractProcessor, param is nil")
	}
	acs := make(map[common.Address]*JuryAccount)
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
	cp := dag.GetChainParameters()
	var contractSigNum int
	if cp.ContractSignatureNum < 1 {
		contractSigNum = core.DefaultContractSignatureNum
	} else {
		contractSigNum = cp.ContractSignatureNum
	}
	var contractEleNum int
	if cp.ContractElectionNum < 1 {
		contractEleNum = core.DefaultContractElectionNum
	} else {
		contractEleNum = cp.ContractElectionNum
	}
	log.Debug("NewContractProcessor", "contractEleNum", contractEleNum, "contractSigNum", contractSigNum)

	cache := freecache.NewCache(20 * 1024 * 1024)
	validator := validator.NewValidate(dag, dag, dag, nil, cache)
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
		electionNum:    contractEleNum, //todo contractSigNum ,cfg.ContractSigNum
		contractSigNum: contractSigNum, //todo contractEleNum ,cfg.ElectionNum
		validator:      validator,
		errMsgEnable:   true,
	}
	log.Info("NewContractProcessor ok", "local address:", p.local, "electionNum", p.electionNum)

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

func (p *Processor) localHaveActiveJury() bool {
	for addr := range p.local {
		if p.isLocalActiveJury(addr) {
			return true
		}
	}
	return false
}

func (p *Processor) GetLocalJuryAddrs() []common.Address {
	num := len(p.local)
	if num <= 0 {
		return nil
	}
	addrs := make([]common.Address, num)
	for addr := range p.local {
		addrs = append(addrs, addr)
	}
	return addrs
}

func (p *Processor) getLocalJuryAccount() *JuryAccount {
	num := len(p.local)
	if num <= 0 {
		return nil
	}
	for _, a := range p.local { //first one
		return a
	}
	return nil
}

func (p *Processor) runContractReq(reqId common.Hash, ele *modules.ElectionNode) error {
	log.Debugf("[%s]runContractReq enter", shortId(reqId.String()))
	defer log.Debugf("[%s]runContractReq exit", shortId(reqId.String()))
	p.locker.Lock()
	ctx := p.mtx[reqId]
	if ctx == nil {
		p.locker.Unlock()
		return fmt.Errorf("runContractReq param is nil, reqId[%s]", reqId)
	}
	reqTx := ctx.reqTx.Clone()
	p.locker.Unlock()

	msgs, err := runContractCmd(rwset.RwM, p.dag, p.contract, &reqTx, ele, p.errMsgEnable) //contract exec long time...
	if err != nil {
		log.Errorf("[%s]runContractReq, runContractCmd reqTx, err：%s", shortId(reqId.String()), err.Error())
		return err
	}
	p.locker.Lock()
	defer p.locker.Unlock()

	tx, err := gen.GenContractTransction(&reqTx, msgs)
	if err != nil {
		log.Error("[%s]runContractReq, GenContractSigTransactions error:%s", shortId(reqId.String()), err.Error())
		return err
	}
	//计算交易费用，将deploy持续时间写入交易中
	addContractDeployDuringTime(p.dag, tx)

	//如果系统合约，直接添加到缓存池
	//如果用户合约，需要签名，添加到缓存池并广播
	if tx.IsSystemContract() {
		ctx.rstTx = tx
	} else {
		account := p.getLocalJuryAccount()
		if account == nil {
			log.Errorf("[%s]runContractReq, not find local account", shortId(reqId.String()))
			return fmt.Errorf("runContractReq no local account, reqId[%s]", reqId.String())
		}
		sigTx, err := p.GenContractSigTransaction(account.Address, account.Password, tx, p.ptn.GetKeyStore())
		if err != nil {
			log.Errorf("[%s]runContractReq, GenContractSigTransctions error:%s", shortId(reqId.String()), err.Error())
			return fmt.Errorf("runContractReq, GenContractSigTransctions error, reqId[%s], err:%s", reqId, err.Error())
		}
		ctx.sigTx = sigTx
		log.Debugf("[%s]runContractReq, gen local signature tx[%s]", shortId(reqId.String()), sigTx.Hash().String())
		//如果rcvTx存在，则比较执行结果，并将结果附加到sigTx上,并删除rcvTx
		if len(ctx.rcvTx) > 0 {
			for _, rtx := range ctx.rcvTx {
				ok, err := checkAndAddTxSigMsgData(ctx.sigTx, rtx)
				if err != nil {
					log.Debugf("[%s]runContractReq, checkAndAddTxSigMsgData error:%s", shortId(reqId.String()), err.Error())
				} else if ok {
					log.Debugf("[%s]runContractReq, checkAndAddTxSigMsgData ok, tx[%s]", shortId(reqId.String()), rtx.Hash().String())
				} else {
					log.Debugf("[%s]runContractReq, checkAndAddTxSigMsgData fail", shortId(reqId.String()))
				}
			}
		}

		sigNum := getTxSigNum(ctx.sigTx)
		log.Debugf("[%s]runContractReq sigNum %d, p.contractSigNum %d", shortId(reqId.String()), sigNum, p.contractSigNum)
		if sigNum >= p.contractSigNum {
			if localIsMinSignature(ctx.sigTx) {
				//签名数量足够，而且当前节点是签名最新的节点，那么合并签名并广播完整交易
				log.Infof("[%s]runContractReq, localIsMinSignature Ok!", shortId(reqId.String()))
				processContractPayout(ctx.sigTx, ele)
				go p.ptn.ContractBroadcast(ContractEvent{CType: CONTRACT_EVENT_COMMIT, Ele: ele, Tx: ctx.sigTx}, true)
				return nil
			}
		}
		//广播
		go p.ptn.ContractBroadcast(ContractEvent{CType: CONTRACT_EVENT_SIG, Ele: ele, Tx: sigTx}, false)
	}
	return nil
}

func (p *Processor) GenContractSigTransaction(signer common.Address, password string, orgTx *modules.Transaction,
	ks *keystore.KeyStore) (*modules.Transaction, error) {
	if orgTx == nil || len(orgTx.TxMessages) < 3 {
		return nil, fmt.Errorf("GenContractSigTransctions param is error")
	}
	if password != "" {
		err := ks.Unlock(accounts.Account{Address: signer}, password)
		if err != nil {
			return nil, err
		}
	}
	tx := orgTx
	needSignMsg := true
	//Find contract pay out payment messages
	resultMsg := false
	isSysContract := false
	reqId := tx.RequestHash()
	for msgidx, msg := range tx.TxMessages {
		if msg.App == modules.APP_CONTRACT_INVOKE_REQUEST {
			resultMsg = true
			requestMsg := msg.Payload.(*modules.ContractInvokeRequestPayload)
			isSysContract = common.IsSystemContractAddress(requestMsg.ContractId)
			continue
		}
		if resultMsg {
			if msg.App == modules.APP_PAYMENT {
				//Contract result里面的Payment只有2种，创币或者从合约付出，
				payment := msg.Payload.(*modules.PaymentPayload)
				if !payment.IsCoinbase() && isSysContract {
					//如果是系统合约付出，那么Mediator一个签名就够了
					//Contract Payout, need sign
					needSignMsg = false
					pubKey, _ := ks.GetPublicKey(signer)
					redeemScript := tokenengine.Instance.GenerateRedeemScript(1, [][]byte{pubKey})
					log.Debugf("[%s]GenContractSigTransaction, RedeemScript:%x", shortId(reqId.String()), redeemScript)
					for inputIdx, input := range payment.Inputs {
						var utxo *modules.Utxo
						var err error
						if input.PreviousOutPoint.TxHash.IsSelfHash() { //引用Tx本身
							output := tx.TxMessages[input.PreviousOutPoint.MessageIndex].Payload.(*modules.PaymentPayload).
								Outputs[input.PreviousOutPoint.OutIndex]
							utxo = &modules.Utxo{
								Amount:    output.Value,
								Asset:     output.Asset,
								PkScript:  output.PkScript,
								LockTime:  0,
								Timestamp: 0,
							}
						} else {
							utxo, err = p.dag.GetUtxoEntry(input.PreviousOutPoint)
							if err != nil {
								return nil, err
							}
						}
						log.Debugf("[%s]GenContractSigTransaction, Lock script:%x", shortId(reqId.String()), utxo.PkScript)
						sign, err := tokenengine.Instance.MultiSignOnePaymentInput(tx, tokenengine.SigHashAll, msgidx, inputIdx,
							utxo.PkScript, redeemScript, ks.GetPublicKey, ks.SignMessage, nil)
						if err != nil {
							log.Errorf("[%s]GenContractSigTransaction, Sign error:%s", shortId(reqId.String()), err)
						}
						log.Debugf("[%s]Sign a contract payout payment,tx[%s],sign:%x", shortId(reqId.String()), tx.Hash().String(), sign)
						input.SignatureScript = sign
					}
				}
			}
		}
	}
	if needSignMsg {
		//没有Contract Payout的情况下，那么需要单独附加Signature Message
		pubKey, err := ks.GetPublicKey(signer)
		if err != nil {
			return nil, fmt.Errorf("GenContractSigTransctions GetPublicKey fail, address[%s], reqId[%s]",
				signer.String(), reqId.String())
		}
		//只对合约执行后不包含Jury签名的Tx进行签名
		sig, err := GetTxSig(tx.GetResultRawTx(), ks, signer)
		if err != nil {
			return nil, fmt.Errorf("GenContractSigTransctions GetTxSig fail, address[%s], reqId[%s]",
				signer.String(), reqId.String())
		}
		sigSet := modules.SignatureSet{
			PubKey:    pubKey,
			Signature: sig,
		}
		SigPayload, err := getContractTxContractInfo(tx, modules.APP_SIGNATURE)
		if err != nil {
			return nil, fmt.Errorf("GenContractSigTransctions getContractTxContractInfo err, address[%s], reqId[%s], err:%s",
				signer.String(), reqId.String(), err.Error())
		}
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
			log.Debugf("[%s]GenContractSigTransactions, Add sign message[%s] to tx requestId[%s]",
				shortId(reqId.String()), sigSet.String(), reqId.String())
			tx.TxMessages = append(tx.TxMessages, msgSig)
		}
		log.Debugf("[%s]GenContractSigTransactions, ok, tx[%s]", shortId(reqId.String()), tx.Hash().String())
	}
	return tx, nil
}
func GetTxSig(tx *modules.Transaction, ks *keystore.KeyStore, signer common.Address) ([]byte, error) {
	reqId := tx.RequestHash()

	sign, err := ks.SigData(tx, signer)
	if err != nil {
		return nil, fmt.Errorf("GetTxSig, Failed to singure transaction, reqId%s, err:%s", reqId.String(), err.Error())
	}
	log.DebugDynamic(func() string {
		data, err := rlp.EncodeToBytes(tx)
		if err != nil {
			return err.Error()
		}
		js, err := json.Marshal(tx)
		if err != nil {
			return err.Error()
		}
		return fmt.Sprintf("Jurior[%s] try to sign tx reqid:%s,signature:%x, tx json: %s\n rlpcode for debug: %x",
			signer.String(), reqId.String(), sign, string(js), data)
	})
	return sign, nil
}
func (p *Processor) AddContractLoop(rwM rwset.TxManager, txpool txspool.ITxPool, addr common.Address,
	ks *keystore.KeyStore) error {
	setChainId := "palletone"
	index := 0
	for _, ctx := range p.mtx {
		if !ctx.valid || ctx.reqTx == nil {
			continue
		}
		reqId := ctx.reqTx.RequestHash()
		if !ctx.reqTx.IsSystemContract() {
			defer rwM.CloseTxSimulator(setChainId, reqId.String())
		}
		if ctx.reqTx.IsSystemContract() && p.contractEventExecutable(CONTRACT_EVENT_EXEC, ctx.reqTx, nil) {
			if cType, err := getContractTxType(ctx.reqTx); err == nil && cType != modules.APP_CONTRACT_TPL_REQUEST {
				ctx.valid = false
				log.Debugf("[%s]AddContractLoop, A enter mtx, addr[%s]", shortId(reqId.String()), addr.String())
				if p.checkTxReqIdIsExist(reqId) {
					log.Debugf("[%s]AddContractLoop ,ReqId is exist ", shortId(reqId.String()))
					continue
				}
				if p.runContractReq(reqId, nil) != nil {
					continue
				}
			}
		}
		if ctx.rstTx == nil {
			continue
		}
		ctx.valid = false

		tx := ctx.rstTx
		reqId = tx.RequestHash()
		if p.checkTxReqIdIsExist(reqId) {
			log.Debugf("[%s]AddContractLoop ,ReqId is exist, rst reqId[%s]", shortId(reqId.String()), reqId.String())
			continue
		}
		if p.checkTxIsExist(tx) {
			log.Debugf("[%s]AddContractLoop ,tx is exist, rst reqId[%s]", shortId(reqId.String()), reqId.String())
			continue
		}
		log.Debugf("[%s]AddContractLoop, B enter mtx, addr[%s]", shortId(reqId.String()), addr.String())
		if tx.IsSystemContract() {
			sigTx, err := p.GenContractSigTransaction(addr, "", tx, ks)
			if err != nil {
				log.Error("AddContractLoop GenContractSigTransctions", "error", err.Error())
				continue
			}
			tx = sigTx
		}
		if err := txpool.AddSequenTx(tx); err != nil {
			log.Errorf("[%s]AddContractLoop, error:%s", shortId(reqId.String()), err.Error())
			continue
		}
		log.Debugf("[%s]AddContractLoop, OK, index[%d], Tx hash[%s], txSize[%f]", shortId(reqId.String()), index, tx.Hash().String(), tx.Size().Float64())
		index++
	}
	return nil
}

func (p *Processor) CheckContractTxValid(rwM rwset.TxManager, tx *modules.Transaction, execute bool) bool {
	if tx == nil {
		log.Error("CheckContractTxValid, param is nil")
		return false
	}
	reqId := tx.RequestHash()
	log.Debugf("[%s]CheckContractTxValid, exec:%v", shortId(reqId.String()), execute)
	if !execute || !tx.IsSystemContract() {
		//不执行合约或者用户合约
		return true
	}
	if p.checkTxReqIdIsExist(tx.RequestHash()) {
		return false
	}
	if p.validator.CheckTxIsExist(tx) {
		return false
	}
	if !p.checkTxValid(tx) {
		log.Errorf("[%s]CheckContractTxValid checkTxValid fail", shortId(reqId.String()))
		return false
	}
	//只检查invoke类型
	if txType, err := getContractTxType(tx); err == nil {
		if txType != modules.APP_CONTRACT_INVOKE_REQUEST {
			//other type is ok
			return true
		}
	}
	//检查本阶段是否有合约执行权限
	if !p.contractEventExecutable(CONTRACT_EVENT_EXEC, tx, nil) {
		log.Debugf("[%s]CheckContractTxValid, nodeContractExecutable false", shortId(reqId.String()))
		return false
	}
	msgs, err := runContractCmd(rwM, p.dag, p.contract, tx, nil, p.errMsgEnable) // long time ...
	if err != nil {
		log.Errorf("[%s]CheckContractTxValid runContractCmd,error:%s", shortId(reqId.String()), err.Error())
		return false
	}
	return msgsCompare(msgs, tx.TxMessages, modules.APP_CONTRACT_INVOKE)
}

func (p *Processor) IsSystemContractTx(tx *modules.Transaction) bool {
	return tx.IsSystemContract()
}

func (p *Processor) isValidateElection(tx *modules.Transaction, ele *modules.ElectionNode, checkExit bool) bool {
	if tx == nil || ele == nil {
		log.Error("isValidateElection, param is nil")
		return false
	}
	reqId := tx.RequestHash()
	if len(ele.EleList) < p.electionNum {
		log.Infof("[%s]isValidateElection, ElectionInf number not enough ,len(ele)[%d], set electionNum[%d]",
			shortId(reqId.String()), len(ele.EleList), p.electionNum)
		return false
	}
	contractId := tx.ContractIdBytes()
	reqAddr, err := p.dag.GetTxRequesterAddress(tx)
	if err != nil {
		log.Errorf("[%s]isValidateElection, GetTxRequesterAddress fail, err:%s", shortId(reqId.String()), err)
		return false
	}
	//从dag中根据contractId获取eleNode信息
	//eleNode, err := p.getContractElectionList(common.NewAddress(contractId, common.ContractHash))
	//if err != nil {
	//	log.Errorf("[%s]isValidateElection, getContractElectionList fail, contractId[%s], err:%s",
	//		shortId(reqId.String()), string(contractId), err)
	//	return false
	//}
	isExit := false
	elr := newElector(uint(p.electionNum), ele.JuryCount, common.Address{},"", p.ptn.GetKeyStore())
	for i, e := range ele.EleList {
		isVerify := false
		//检查地址hash是否在本地
		if checkExit && !isExit {
			jury := p.getLocalJuryAccount()
			log.Debugf("[%s]isValidateElection, addrHash[%s]", shortId(reqId.String()), e.AddrHash.String())
			if bytes.Equal(e.AddrHash.Bytes(), util.RlpHash(jury.Address).Bytes()) {
				isExit = true
			}
		}
		//检查指定节点模式下，是否为jjh请求地址
		if e.EType == 1 {
			jjhAd := p.dag.GetChainParameters().FoundationAddress
			if jjhAd == reqAddr.Str() { //true
				log.Debugf("[%s]isValidateElection, e.EType == 1, ok", shortId(reqId.String()))
				continue
			} else {
				log.Debugf("[%s]isValidateElection, e.EType == 1, but not jjh request addr", shortId(reqId.String()))
				log.Debugf("[%s]isValidateElection, reqAddr[%s], jjh[%s]", shortId(reqId.String()), reqAddr.Str(), jjhAd)
				return false
			}
		}

		//检查地址与pubKey是否匹配:获取当前pubKey下的Addr，将地址hash后与输入比较
		addr := crypto.PubkeyBytesToAddress(e.PublicKey)
		if e.AddrHash != util.RlpHash(addr) {
			log.Errorf("[%s]isValidateElection, publicKey not match address, addrHash[%v]", shortId(reqId.String()), e.AddrHash)
			return false
		}
		//从数据库中查询该地址是否为Jury
		if !p.dag.IsActiveJury(addr) {
			log.Errorf("[%s]isValidateElection, not active Jury, addrHash[%v]", shortId(reqId.String()), e.AddrHash)
			return false
		}
		isVerify, err := elr.verifyVrf(e.Proof, conversionElectionSeedData(contractId), e.PublicKey)
		if err != nil || !isVerify {
			log.Infof("[%s]isValidateElection, index[%d],verifyVrf fail, contractId[%s]",
				shortId(reqId.String()), i, string(contractId))
			return false
		}
	}
	if checkExit {
		if !isExit {
			log.Debugf("[%s]isValidateElection, election addr not in local", shortId(reqId.String()))
			return false
		}
	}
	return true
}

func (p *Processor) contractEventExecutable(event ContractEventType, tx *modules.Transaction,
	ele *modules.ElectionNode) bool {
	if tx == nil {
		log.Errorf("tx is nil")
		return false
	}
	reqId := tx.RequestHash()
	isSysContract := tx.IsSystemContract()
	isMediator := p.ptn.LocalHaveActiveMediator()
	isJury := p.localHaveActiveJury()

	switch event {
	case CONTRACT_EVENT_ELE:
		if !isSysContract && isMediator {
			log.Debugf("[%s]contractEventExecutable, CONTRACT_EVENT_ELE, Mediator, true", shortId(reqId.String()))
			return true
		}
	case CONTRACT_EVENT_EXEC:
		if isSysContract && isMediator {
			log.Debugf("[%s]contractEventExecutable, CONTRACT_EVENT_EXEC, Mediator, true", shortId(reqId.String()))
			return true
		} else if !isSysContract && isJury {
			if p.isValidateElection(tx, ele, true) {
				log.Debugf("[%s]contractEventExecutable, CONTRACT_EVENT_EXEC, Jury, true", shortId(reqId.String()))
				return true
			} else {
				log.Debugf("[%s]contractEventExecutable, CONTRACT_EVENT_EXEC, Jury, isValidateElection fail, false",
					shortId(reqId.String()))
			}
		}
	case CONTRACT_EVENT_SIG:
		if !isSysContract && isJury {
			if p.isValidateElection(tx, ele, false) {
				log.Debugf("[%s]contractEventExecutable, CONTRACT_EVENT_SIG, Jury, true", shortId(reqId.String()))
				return true
			} else {
				log.Debugf("[%s]contractEventExecutable, CONTRACT_EVENT_SIG, Jury, isValidateElection fail, false",
					shortId(reqId.String()))
			}
		}
	case CONTRACT_EVENT_COMMIT:
		if isMediator {
			if isSysContract {
				log.Debugf("[%s]contractEventExecutable, CONTRACT_EVENT_COMMIT, Mediator, sysContract, true",
					shortId(reqId.String()))
				return true
			} else if !isSysContract && p.isValidateElection(tx, ele, false) {
				log.Debugf("[%s]contractEventExecutable, CONTRACT_EVENT_COMMIT, Mediator, userContract, true",
					shortId(reqId.String()))
				return true
			} else {
				log.Debugf("[%s]contractEventExecutable, CONTRACT_EVENT_COMMIT, Mediator, isValidateElection fail, false:",
					shortId(reqId.String()))
			}
		}
	}
	return false
}

func (p *Processor) createContractTxReqToken(contractId, from, to, toToken common.Address, daoAmount, daoFee, daoAmountToken uint64,
	assetToken string, msg *modules.Message) (common.Hash, *modules.Transaction, error) {
	tx, _, err := p.dag.CreateTokenTransaction(from, to, toToken, daoAmount, daoFee, daoAmountToken, assetToken,
		msg, p.ptn.TxPool())
	if err != nil {
		return common.Hash{}, nil, err
	}
	log.Debugf("[%s]createContractTxReqToken,contractId[%s],tx[%v]",
		shortId(tx.RequestHash().String()), contractId.String(), tx)
	return p.signAndExecute(contractId, from, tx)
}

func (p *Processor) createContractTxReq(contractId, from, to common.Address, daoAmount, daoFee uint64, certID *big.Int,
	msg *modules.Message) (common.Hash, *modules.Transaction, error) {
	tx, _, err := p.dag.CreateGenericTransaction(from, to, daoAmount, daoFee, certID, msg, p.ptn.TxPool())
	if err != nil {
		return common.Hash{}, nil, err
	}
	return p.signAndExecute(contractId, from, tx)
}
func (p *Processor) SignAndExecuteAndSendRequest(from common.Address,
	tx *modules.Transaction) (*modules.Transaction, error) {
	requestMsg := tx.TxMessages[tx.GetRequestMsgIndex()]
	if requestMsg.App == modules.APP_CONTRACT_INVOKE_REQUEST {
		request := requestMsg.Payload.(*modules.ContractInvokeRequestPayload)
		contractId := common.NewAddress(request.ContractId, common.ContractHash)
		reqId, tx, err := p.signAndExecute(contractId, from, tx)
		if err != nil {
			return nil, err
		}
		//broadcast
		go p.ptn.ContractBroadcast(ContractEvent{CType: CONTRACT_EVENT_EXEC, Ele: p.mtx[reqId].eleNode, Tx: tx}, true)
		return tx, nil
	}
	return nil, errors.New("Not support request")
}
func (p *Processor) signAndExecute(contractId common.Address, from common.Address,
	tx *modules.Transaction) (common.Hash, *modules.Transaction, error) {
	tx, err := p.ptn.SignGenericTransaction(from, tx)
	if err != nil {
		return common.Hash{}, nil, err
	}
	p.locker.Lock()
	defer p.locker.Unlock()

	reqId := tx.RequestHash()
	if !checkContractTxFeeValid(p.dag, tx) {
		log.Errorf("[%s]signAndExecute, checkContractTxFeeValid fail", shortId(reqId.String()))
		return common.Hash{}, nil, errors.New("checkContractTxFeeValid false")
	}
	log.Debugf("[%s]signAndExecute, contractId[%s]", shortId(reqId.String()), contractId.String())
	if p.mtx[reqId] != nil {
		return reqId, nil, fmt.Errorf("contract request transaction[%s] already created", shortId(reqId.String()))
	}
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
		} else { //invoke,stop
			eleNode, err := p.getContractElectionList(contractId)
			if err != nil {
				log.Errorf("[%s]signAndExecute, getContractElectionList fail,err:%s",
					shortId(tx.RequestHash().String()), err.Error())
				return common.Hash{}, nil, err
			}
			ctx.eleNode = eleNode
		}
	}
	return reqId, tx, nil
}

func (p *Processor) ContractTxDeleteLoop() {
	for {
		time.Sleep(time.Second * time.Duration(5))
		p.locker.Lock()
		for k, v := range p.mtx {
			if !v.valid {
				if time.Since(v.tm) > time.Second*120 {
					log.Infof("[%s]ContractTxDeleteLoop, contract is invalid, delete tx id", shortId(k.String()))
					delete(p.mtx, k)
				}
			} else {
				if time.Since(v.tm) > time.Second*600 {
					log.Infof("[%s]ContractTxDeleteLoop, contract is valid, delete tx id", shortId(k.String()))
					delete(p.mtx, k)
				}
			}
		}
		for k, v := range p.mel {
			if time.Since(v.tm) > time.Second*300 {
				log.Infof("[%s]ContractTxDeleteLoop, delete electionVrf ", shortId(k.String()))
				delete(p.mel, k)
			}
		}
		p.locker.Unlock()
	}
}

func (p *Processor) getContractElectionList(contractId common.Address) (*modules.ElectionNode, error) {
	return p.dag.GetContractJury(contractId.Bytes())
}

func (p *Processor) getContractAssignElectionList(tx *modules.Transaction) ([]modules.ElectionInf, error) {
	if tx == nil {
		return nil, errors.New("getContractAssignElectionList, param is nil")
	}
	reqId := tx.RequestHash()
	payload, err := getContractTxContractInfo(tx, modules.APP_CONTRACT_DEPLOY_REQUEST)
	if err != nil {
		return nil, fmt.Errorf("[%s]getContractAssignElectionList, getContractTxContractInfo fail", shortId(reqId.String()))
	}

	num := 0
	eels := make([]modules.ElectionInf, 0)
	tplId := payload.(*modules.ContractDeployRequestPayload).TemplateId
	//find the address of the contract template binding in the dag
	//addrHash, err := p.getTemplateAddrHash(tplId)
	tpl, err := p.dag.GetContractTpl(tplId)
	if err != nil {
		log.Debugf("[%s]getContractAssignElectionList, getTemplateAddrHash fail,templateId[%x], fail:%s",
			shortId(reqId.String()), tplId, err.Error())
		return nil, fmt.Errorf("[%s]getContractAssignElectionList, GetContractTpl fail", shortId(reqId.String()))
	}
	addrHash := tpl.AddrHash
	if len(addrHash) >= p.electionNum {
		num = p.electionNum
	} else {
		num = len(addrHash)
	}
	//add election node form template install assignation
	for i := 0; i < num; i++ {
		e := modules.ElectionInf{EType: 1, AddrHash: addrHash[i]}
		eels = append(eels, e)
	}
	return eels, nil
}
