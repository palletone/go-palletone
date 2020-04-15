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

	"github.com/coocood/freecache"
	"github.com/ethereum/go-ethereum/event"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/crypto"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/p2p"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/common/util"
	"github.com/palletone/go-palletone/contracts"
	"github.com/palletone/go-palletone/contracts/utils"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/core/accounts"
	"github.com/palletone/go-palletone/core/accounts/keystore"
	"github.com/palletone/go-palletone/core/gen"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/dboperation"
	"github.com/palletone/go-palletone/dag/errors"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/rwset"
	"github.com/palletone/go-palletone/tokenengine"
	"github.com/palletone/go-palletone/txspool"
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
	EnableGasFee() bool
}

type iDag interface {
	GetTxFee(pay *modules.Transaction) (*modules.AmountAsset, error)
	//CurrentHeader(token modules.AssetId) *modules.Header
	GetNewestUnit(token modules.AssetId) (common.Hash, *modules.ChainIndex, error)
	GetAddrUtxos(addr common.Address) (map[modules.OutPoint]*modules.Utxo, error)
	GetContractStatesById(id []byte) (map[string]*modules.ContractStateValue, error)
	GetContractTplCode(tplId []byte) ([]byte, error)
	GetGlobalProp() *modules.GlobalProperty
	GetHeaderByNumber(number *modules.ChainIndex) (*modules.Header, error)
	GetStableTransactionOnly(hash common.Hash) (*modules.Transaction, error)
	GetStableUnit(hash common.Hash) (*modules.Unit, error)
	GetStableUnitByNumber(number *modules.ChainIndex) (*modules.Unit, error)
	UnstableHeadUnitProperty(asset modules.AssetId) (*modules.UnitProperty, error)
	GetDb() ptndb.Database
	//GetTxFee(pay *modules.Transaction) (*modules.AmountAsset, error)
	GetUtxoEntry(outpoint *modules.OutPoint) (*modules.Utxo, error)
	GetAddrByOutPoint(outPoint *modules.OutPoint) (common.Address, error)
	GetActiveMediators() []common.Address
	GetTxHashByReqId(reqid common.Hash) (common.Hash, error)
	IsActiveJury(addr common.Address) bool
	JuryCount() uint
	GetContractDevelopers() ([]common.Address, error)
	IsContractDeveloper(addr common.Address) bool
	GetStxoEntry(outpoint *modules.OutPoint) (*modules.Stxo, error)
	//GetActiveJuries() []common.Address
	IsActiveMediator(addr common.Address) bool
	GetAddr1TokenUtxos(addr common.Address, asset *modules.Asset) (map[modules.OutPoint]*modules.Utxo, error)
	CreateGenericTransaction(from, to common.Address, daoAmount, daoFee uint64, certID *big.Int,
		msg *modules.Message, enableGasFee bool) (*modules.Transaction, uint64, error)
	CreateTokenTransaction(from, to common.Address, token *modules.Asset, daoAmountToken, daoFee uint64,
		msg *modules.Message) (*modules.Transaction, uint64, error)
	GetTransaction(hash common.Hash) (*modules.TransactionWithUnitInfo, error)
	GetTransactionOnly(hash common.Hash) (*modules.Transaction, error)
	GetHeaderByHash(common.Hash) (*modules.Header, error)
	//GetTxRequesterAddress(tx *modules.Transaction) (common.Address, error)
	//GetConfig(name string) ([]byte, *modules.StateVersion, error)
	IsTransactionExist(hash common.Hash) (bool, error)
	GetContractJury(contractId []byte) (*modules.ElectionNode, error)
	GetContractTpl(tplId []byte) (*modules.ContractTemplate, error)
	GetContract(contractId []byte) (*modules.Contract, error)
	GetContractState(id []byte, field string) ([]byte, *modules.StateVersion, error)
	GetContractStatesByPrefix(id []byte, prefix string) (map[string]*modules.ContractStateValue, error)
	ChainThreshold() int
	GetChainParameters() *core.ChainParameters
	GetMediators() map[common.Address]bool
	GetMediator(add common.Address) *core.Mediator
	GetBlacklistAddress() ([]common.Address, *modules.StateVersion, error)
	GetJurorByAddrHash(addrHash common.Hash) (*modules.JurorDeposit, error)
	SaveTransaction(tx *modules.Transaction, txIndex int) error
	//nouse
	GetPtnBalance(addr common.Address) uint64
	GetPayload(from, to common.Address, daoAmount, daoFee uint64,
		utxos map[modules.OutPoint]*modules.Utxo) (*modules.PaymentPayload, error)
	GetNewestUnitTimestamp(token modules.AssetId) (int64, error)
	GetScheduledMediator(slotNum uint32) common.Address
	GetSlotAtTime(when time.Time) uint32
	GetJurorReward(jurorAdd common.Address) common.Address

	CheckReadSetValid(contractId []byte, readSet []modules.ContractReadSet) bool
	GetContractsWithJuryAddr(addr common.Hash) []*modules.Contract
	SaveContract(contract *modules.Contract) error
	GetImmutableChainParameters() *core.ImmutableChainParameters
	NewTemp() (dboperation.IContractDag, error)
	HeadUnitNum() uint64

	SubscribeSaveUnitEvent(ch chan<- modules.SaveUnitEvent) event.Subscription
	//SubscribeUnstableRepositoryUpdatedEvent(ch chan<- modules.UnstableRepositoryUpdatedEvent) event.Subscription
	SubscribeSaveStableUnitEvent(ch chan<- modules.SaveUnitEvent) event.Subscription
	//localdb
	SaveLocalTx(tx *modules.Transaction) error
	GetLocalTx(txId common.Hash) (*modules.Transaction, modules.TxStatus, error)
	SaveLocalTxStatus(txId common.Hash, status modules.TxStatus) error
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
	//stxo         map[modules.OutPoint]bool
	//utxo         map[modules.OutPoint]*modules.Utxo
	quit         chan struct{}
	locker       *sync.Mutex //locker       *sync.Mutex  RWMutex
	errMsgEnable bool        //package contract execution error information into the transaction

	contractExecFeed  event.Feed
	contractExecScope event.SubscriptionScope
	pDocker           *utils.PalletOneDocker
}

//var instanceProcessor *Processor

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
	cfgSigNum := getSysCfgContractSignatureNum(dag)
	cfgEleNum := getSysCfgContractElectionNum(dag)
	log.Debug("NewContractProcessor", "contractEleNum", cfgEleNum, "contractSigNum", cfgSigNum)

	cache := freecache.NewCache(20 * 1024 * 1024)
	val := validator.NewValidate(dag, dag, dag, dag, dag, cache, false, ptn.EnableGasFee())
	//val.SetContractTxCheckFun(CheckTxContract)
	//TODO Devin

	p := &Processor{
		name:         "contractProcessor",
		ptn:          ptn,
		dag:          dag,
		contract:     contract,
		local:        acs,
		locker:       new(sync.Mutex),
		quit:         make(chan struct{}),
		mtx:          make(map[common.Hash]*contractTx),
		mel:          make(map[common.Hash]*electionVrf),
		lockVrf:      make(map[common.Address][]modules.ElectionInf),
		validator:    val,
		errMsgEnable: true,
	}
	//instanceProcessor = p
	log.Info("NewContractProcessor ok", "local address:", p.local)

	return p, nil
}
func (p *Processor) SetContract(contract *contracts.Contract) {
	p.contract = contract
}

func (p *Processor) SetDocker(pDocker *utils.PalletOneDocker) {
	p.pDocker = pDocker
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
	addrs := make([]common.Address, 0)
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

func (p *Processor) runContractReq(reqId common.Hash, ele *modules.ElectionNode, txMgr rwset.TxManager, dag dboperation.IContractDag) error {
	log.Debugf("[%s]runContractReq enter", reqId.ShortStr())
	defer log.Debugf("[%s]runContractReq exit", reqId.ShortStr())
	p.locker.Lock()
	ctx := p.mtx[reqId]
	if ctx == nil {
		p.locker.Unlock()
		return fmt.Errorf("runContractReq param is nil, reqId[%s]", reqId)
	}
	if ctx.rstTx != nil && ctx.rstTx.IsSystemContract() {
		log.Debugf("[%s]runContractReq, tstTx already exist", reqId.ShortStr())
		p.locker.Unlock()
		return nil
	}
	reqTx := ctx.reqTx.Clone()
	p.locker.Unlock()
	cctx := &contracts.ContractProcessContext{
		RequestId:    reqTx.RequestHash(),
		Dag:          dag,
		Ele:          ele,
		RwM:          txMgr,
		TxPool:       p.ptn.TxPool(),
		Contract:     p.contract,
		ErrMsgEnable: p.errMsgEnable,
	}
	msgs, err := runContractCmd(cctx, reqTx) //contract exec long time...
	if err != nil {
		log.Errorf("[%s]runContractReq, runContractCmd reqTx, err：%s", reqId.ShortStr(), err.Error())
		return err
	}
	p.locker.Lock()
	defer p.locker.Unlock()
	//TODO update utxo,stxo

	tx, err := p.GenContractTransaction(reqTx, msgs)
	if err != nil {
		log.Error("[%s]runContractReq, GenContractSigTransactions error:%s", reqId.ShortStr(), err.Error())
		return err
	}

	//如果系统合约，直接添加到缓存池
	//如果用户合约，需要签名，添加到缓存池并广播
	if tx.IsSystemContract() {
		log.Debugf("[%s]runContractReq, is system contract, add rstTx", reqId.ShortStr())
		//reqType, _ := tx.GetContractTxType()
		//if reqType != modules.APP_CONTRACT_TPL_REQUEST {//合约模板交易需要签名生成最终交易
		//	err = dag.SaveTransaction(tx)
		//	if err != nil {
		//		log.Errorf("[%s]runContractReq, SaveTransaction err:%s", reqId.ShortStr(), err.Error())
		//		return err
		//	}
		//}
		//Devin:还没有签名，不能SaveTx
		ctx.rstTx = tx
	} else {
		account := p.getLocalJuryAccount()
		if account == nil {
			log.Errorf("[%s]runContractReq, not find local account", reqId.ShortStr())
			return fmt.Errorf("runContractReq no local account, reqId[%s]", reqId.String())
		}
		sigTx, err := p.GenContractSigTransaction(account.Address, account.Password, tx, p.ptn.GetKeyStore(), dag.GetUtxoEntry)
		if err != nil {
			log.Errorf("[%s]runContractReq, GenContractSigTransctions error:%s", reqId.ShortStr(), err.Error())
			return fmt.Errorf("runContractReq, GenContractSigTransctions error, reqId[%s], err:%s", reqId, err.Error())
		}
		ctx.sigTx = sigTx
		log.Debugf("[%s]runContractReq, gen local signature tx[%s]", reqId.ShortStr(), sigTx.Hash().String())
		//如果rcvTx存在，则比较执行结果，并将结果附加到sigTx上,并删除rcvTx
		if len(ctx.rcvTx) > 0 {
			for _, rtx := range ctx.rcvTx {
				ok, err := checkAndAddTxSigMsgData(ctx.sigTx, rtx)
				if err != nil {
					log.Debugf("[%s]runContractReq, checkAndAddTxSigMsgData error:%s", reqId.ShortStr(), err.Error())
				} else if ok {
					log.Debugf("[%s]runContractReq, checkAndAddTxSigMsgData ok, tx[%s]", reqId.ShortStr(), rtx.Hash().String())
				} else {
					log.Debugf("[%s]runContractReq, checkAndAddTxSigMsgData fail", reqId.ShortStr())
				}
			}
		}

		sigNum := getTxSigNum(ctx.sigTx)
		cfgSigNum := getSysCfgContractSignatureNum(p.dag)
		log.Debugf("[%s]runContractReq sigNum %d, p.contractSigNum %d", reqId.ShortStr(), sigNum, cfgSigNum)
		if sigNum >= cfgSigNum {
			if localIsMinSignature(ctx.sigTx) {
				//签名数量足够，而且当前节点是签名最新的节点，那么合并签名并广播完整交易
				log.Infof("[%s]runContractReq, localIsMinSignature Ok!", reqId.ShortStr())
				p.processContractPayout(ctx.sigTx, ele)
				go p.ptn.ContractBroadcast(ContractEvent{CType: CONTRACT_EVENT_COMMIT, Ele: ele, Tx: ctx.sigTx}, true)
				return nil
			}
		}
		//广播
		go p.ptn.ContractBroadcast(ContractEvent{CType: CONTRACT_EVENT_SIG, Ele: ele, Tx: sigTx}, false)
	}
	return nil
}

func (p *Processor) GenContractTransaction(orgTx *modules.Transaction, msgs []*modules.Message) (*modules.Transaction, error) {
	if orgTx == nil || msgs == nil {
		log.Error("GenContractTransaction, param is nil")
		return nil, fmt.Errorf("GenContractTransaction, param is nil")
	}
	reqId := orgTx.RequestHash()
	tx, err := gen.GenContractTransction(orgTx, msgs)
	if err != nil {
		log.Errorf("[%s]GenContractTransaction, gen.GenContractTransction err:%s", reqId.ShortStr(), err.Error())
		return nil, fmt.Errorf("GenContractTransaction err:%s", err.Error())
	}

	payInputNum := getContractInvokeMulPaymentInputNum(tx)
	if payInputNum > 0 {
		log.Debugf("[%s]GenContractTransaction,payInputNum[%d]", reqId.ShortStr(), payInputNum)
	}
	//extSize := ContractDefaultSignatureSize + ContractDefaultPayInputSignatureSize*float64(payInputNum)
	//Devin:没有当前dag，无法正确ValidateTxFeeEnough
	//if p.validator.ValidateTxFeeEnough(tx, extSize, 0) != validator.TxValidationCode_VALID {
	//	msgs, err = genContractErrorMsg(tx, errors.New("tx fee is invalid"), true)
	//	if err != nil {
	//		log.Errorf("[%s]GenContractTransaction, genContractErrorMsg,error:%s", reqId.ShortStr(), err.Error())
	//		return nil, err
	//	}
	//	tx, err = gen.GenContractTransction(orgTx.GetRequestTx(), msgs)
	//	if err != nil {
	//		log.Error("[%s]GenContractTransaction,fee is not enough, GenContractTransaction error:%s",
	//			reqId.ShortStr(), err.Error())
	//		return nil, err
	//	}
	//} else {
	//计算交易费用，将deploy持续时间写入交易中
	err = addContractDeployDuringTime(p.dag, tx)
	if err != nil {
		log.Debugf("[%s]runContractReq, addContractDeployDuringTime error:%s", reqId.ShortStr(), err.Error())
	}
	//}
	return tx, nil
}

func (p *Processor) GenContractSigTransaction(signer common.Address, password string, orgTx *modules.Transaction,
	ks *keystore.KeyStore, utxoFunc modules.QueryUtxoFunc) (*modules.Transaction, error) {
	if orgTx == nil || len(orgTx.Messages()) < 3 {
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
	isSysContract := false
	reqId := tx.RequestHash()
	msgs := tx.TxMessages()
	reqMsgCount := tx.GetRequestMsgCount()
	pubKey, err := ks.GetPublicKey(signer)
	if err != nil {
		return nil, fmt.Errorf("GenContractSigTransaction GetPublicKey fail, address[%s], reqId[%s]",
			signer.String(), reqId.String())
	}
	for msgidx, msg := range msgs {
		if msg.App == modules.APP_CONTRACT_INVOKE_REQUEST {

			requestMsg := msg.Payload.(*modules.ContractInvokeRequestPayload)
			isSysContract = common.IsSystemContractId(requestMsg.ContractId)
			continue
		}
		if msgidx >= reqMsgCount { //invoke result
			if msg.App == modules.APP_PAYMENT {
				//Contract result里面的Payment只有2种，创币或者从合约付出
				payment := msg.Payload.(*modules.PaymentPayload)
				if !payment.IsCoinbase() && isSysContract {
					//如果是系统合约付出，那么Mediator一个签名就够了
					//Contract Payout, need sign
					needSignMsg = false
					redeemScript := tokenengine.Instance.GenerateRedeemScript(1, [][]byte{pubKey})
					log.Debugf("[%s]GenContractSigTransaction, RedeemScript:%x", reqId.ShortStr(), redeemScript)
					for inputIdx, input := range payment.Inputs {
						var utxo *modules.Utxo
						var err error
						if input.PreviousOutPoint.TxHash.IsSelfHash() { //引用Tx本身
							output := msgs[input.PreviousOutPoint.MessageIndex].Payload.(*modules.PaymentPayload).
								Outputs[input.PreviousOutPoint.OutIndex]
							utxo = &modules.Utxo{
								Amount:    output.Value,
								Asset:     output.Asset,
								PkScript:  output.PkScript,
								LockTime:  0,
								Timestamp: 0,
							}
						} else {
							utxo, err = utxoFunc(input.PreviousOutPoint)
							if err != nil {
								return nil, fmt.Errorf("query utxo[%s] get error:%s",
									input.PreviousOutPoint.String(), err.Error())
							}
						}
						log.Debugf("[%s]GenContractSigTransaction, Lock script:%x", reqId.ShortStr(), utxo.PkScript)
						sign, err := tokenengine.Instance.MultiSignOnePaymentInput(tx, tokenengine.SigHashAll, msgidx, inputIdx,
							utxo.PkScript, redeemScript, ks.GetPublicKey, ks.SignMessage, nil)
						if err != nil {
							log.Errorf("[%s]GenContractSigTransaction, Sign error:%s", reqId.ShortStr(), err)
						}
						log.Debugf("[%s]Sign a contract payout payment,sign:%x", reqId.ShortStr(), sign)
						input.SignatureScript = sign

						log.Debugf("[%s]Sign a contract payout payment,sign size:%d", reqId.ShortStr(), len(sign))

						tx.ModifiedMsg(msgidx, msg)
					}
				}
			}
		}
	}
	if needSignMsg {
		//没有Contract Payout的情况下，那么需要单独附加Signature Message
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
		err = addContractSignatureSet(tx, &sigSet)
		if err != nil {
			log.Errorf("[%s]GenContractSigTransactions, addContractSignatureSet,error:%s", reqId.ShortStr(), err.Error())
			return nil, err
		}
		log.Debugf("[%s]GenContractSigTransactions, ok", reqId.ShortStr())
	}

	return tx, nil
}
func (p *Processor) RunAndSignTx(reqTx *modules.Transaction, txMgr rwset.TxManager, dag dboperation.IContractDag,
	addr common.Address) (*modules.Transaction, error) {
	cctx := &contracts.ContractProcessContext{
		RequestId:    reqTx.RequestHash(),
		Dag:          dag,
		Ele:          nil,
		RwM:          txMgr,
		Contract:     p.contract,
		ErrMsgEnable: p.errMsgEnable,
		TxPool:       p.ptn.TxPool(),
	}
	reqId := reqTx.Hash()
	log.Debugf("run contract request[%s]", reqId.String())
	msgs, err := runContractCmd(cctx, reqTx) //contract exec long time...
	if err != nil {
		log.Errorf("[%s]runContractReq, runContractCmd reqTx, err：%s", reqId.ShortStr(), err.Error())
		return nil, err
	}
	p.locker.Lock()
	defer p.locker.Unlock()

	tx, err := p.GenContractTransaction(reqTx, msgs)
	if err != nil {
		log.Error("[%s]runContractReq, GenContractSigTransactions error:%s", reqId.ShortStr(), err.Error())
		return nil, err
	}
	sigTx, err := p.GenContractSigTransaction(addr, "", tx, p.ptn.GetKeyStore(), dag.GetUtxoEntry)
	if err != nil {
		log.Error("GenContractSigTransctions", "error", err.Error())
		return nil, err
	}
	log.Debugf("processed req[%s], result tx[%s]", reqId.String(), sigTx.Hash().String())
	return sigTx, nil
}

func (p *Processor) CheckContractTxValid(rwM rwset.TxManager, tx *modules.Transaction, execute bool) bool {
	if tx == nil {
		log.Error("CheckContractTxValid, param is nil")
		return false
	}
	reqId := tx.RequestHash()
	log.Debugf("[%s]CheckContractTxValid, exec:%v", reqId.ShortStr(), execute)
	if !execute || !tx.IsSystemContract() {
		//不执行合约或者用户合约
		return true
	}
	if !p.ptn.LocalHaveActiveMediator() {
		//只有mediator处理系统合约
		return true
	}
	if p.checkTxReqIdIsExist(tx.RequestHash()) {
		return false
	}
	if p.validator.CheckTxIsExist(tx) {
		return false
	}

	if _, v, err := p.validator.ValidateTx(tx, false); v != validator.TxValidationCode_VALID && err != nil {
		log.Errorf("[%s]CheckContractTxValid checkTxValid fail, err:%s", reqId.ShortStr(), err.Error())
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
		log.Debugf("[%s]CheckContractTxValid, nodeContractExecutable false", reqId.ShortStr())
		return false
	}

	if contractTx, ok := p.mtx[reqId]; ok && contractTx.rstTx != nil {
		log.Debugf("[%s]CheckContractTxValid, already exit rstTx", reqId.ShortStr())
		return msgsCompareInvoke(tx.TxMessages(), contractTx.rstTx.TxMessages())
	}

	ctx := &contracts.ContractProcessContext{
		RequestId:    tx.RequestHash(),
		Dag:          p.dag,
		Ele:          nil,
		RwM:          rwM,
		Contract:     p.contract,
		ErrMsgEnable: p.errMsgEnable,
	}
	msgs, err := runContractCmd(ctx, tx) // long time ...
	if err != nil {
		log.Errorf("[%s]CheckContractTxValid, runContractCmd,error:%s", reqId.ShortStr(), err.Error())
		return false
	}
	reqTx := tx.GetRequestTx()
	txTmp, err := p.GenContractTransaction(reqTx, msgs)
	if err != nil {
		log.Errorf("[%s]CheckContractTxValid, GenContractTransction, error:%s", reqId.ShortStr(), err.Error())
		return false
	}

	if p.mtx[reqId] == nil {
		p.mtx[reqId] = &contractTx{
			rstTx:  nil,
			tm:     time.Now(),
			valid:  true,
			adaInf: make(map[uint32]*AdapterInf),
		}
	}
	//err = p.dag.SaveTransaction(txTmp)
	//if err != nil {
	//	log.Errorf("[%s]CheckContractTxValid, SaveTransaction err:%s", reqId.ShortStr(), err.Error())
	//	return false
	//}
	p.mtx[reqId].reqTx = reqTx
	p.mtx[reqId].rstTx = txTmp

	return msgsCompareInvoke(txTmp.TxMessages(), tx.TxMessages())
}

//验证一个系统合约的执行结果是否正确
func CheckContractTxResult(tx *modules.Transaction, rwM rwset.TxManager, dag dboperation.IContractDag) bool {
	if tx == nil {
		log.Error("CheckContractTxResult, param is nil")
		return false
	}
	reqTx := tx.GetRequestTx()
	reqId := reqTx.Hash()
	log.Debugf("CheckContractTxResult enter reqId: [%s]", reqId.ShortStr())
	//只检查系统合约
	if !tx.IsSystemContract() {
		return true
	}

	if txType, err := getContractTxType(tx); err == nil { //只检查invoke类型
		if txType != modules.APP_CONTRACT_INVOKE_REQUEST {
			return true
		}
	}
	_, m, _ := getContractTxContractInfo(tx, modules.APP_CONTRACT_INVOKE)
	if m == nil {
		log.Debugf("[%s]CheckContractTxResult, msg not include invoke payload", reqId.ShortStr())
		return true
	}

	ctx := &contracts.ContractProcessContext{
		RequestId:    reqId,
		Dag:          dag,
		Ele:          nil,
		RwM:          rwM,
		Contract:     &contracts.Contract{},
		ErrMsgEnable: true,
	}
	msgs, err := runContractCmd(ctx, tx) // long time ...
	if err != nil {
		log.Errorf("[%s]CheckContractTxResult, runContractCmd,error:%s", reqId.ShortStr(), err.Error())
		return false
	}
	resultMsgs := []*modules.Message{}
	isResult := false
	for _, msg := range tx.GetResultRawTx().TxMessages() {
		if msg.App.IsRequest() {
			isResult = true
		}
		if isResult {
			resultMsgs = append(resultMsgs, msg)
		}
	}
	isMsgSame := msgsCompareInvoke(msgs, resultMsgs)
	log.Debugf("CheckContractTxResult, compare request[%s] and execute result:%t", reqId.String(), isMsgSame)
	return isMsgSame
}

//func (p *Processor) IsSystemContractTx(tx *modules.Transaction) bool {
//	return tx.IsSystemContract()
//}
func (p *Processor) getUtxoFromPoolAndDag(outpoint *modules.OutPoint) (*modules.Utxo, error) {
	utxo, err := p.ptn.TxPool().GetUtxoFromAll(outpoint)
	if err == nil {
		return utxo, nil
	}
	log.DebugDynamic(func() string {
		return fmt.Sprintf("GetUtxo(%s) not in txpool,try dag query", outpoint.String())
	})
	return p.dag.GetUtxoEntry(outpoint)
}
func (p *Processor) isValidateElection(tx *modules.Transaction, ele *modules.ElectionNode, checkExit bool) bool {
	if tx == nil {
		log.Error("isValidateElection, param tx is nil")
		return false
	}
	if tx.GetContractTxType() == modules.APP_CONTRACT_TPL_REQUEST {
		return true
	}
	reqId := tx.RequestHash()
	if ele == nil {
		log.Errorf("[%s]isValidateElection, param ele is nil", reqId.ShortStr())
		return false
	}
	cfgEleNum := getSysCfgContractElectionNum(p.dag)
	if len(ele.EleList) < cfgEleNum {
		log.Infof("[%s]isValidateElection, ElectionInf number not enough ,len(ele)[%d], set electionNum[%d]",
			reqId.ShortStr(), len(ele.EleList), cfgEleNum)
		return false
	}
	contractId := tx.GetContractId()
	reqAddrs, err := tx.GetFromAddrs(p.getUtxoFromPoolAndDag, tokenengine.Instance.GetAddressFromScript)
	//reqAddr, err := p.dag.GetTxRequesterAddress(tx)
	if err != nil {
		log.Errorf("[%s]isValidateElection, GetTxRequesterAddress fail, err:%s", reqId.ShortStr(), err)
		return false
	}
	cType, err := getContractTxType(tx)
	if err != nil {
		log.Errorf("[%s]isValidateElection, getContractTxType fail", reqId.ShortStr())
		return false
	}
	jjhAd := p.dag.GetChainParameters().FoundationAddress
	isExit := false
	elr := newElector(uint(cfgEleNum), ele.JuryCount, common.Address{}, "", p.ptn.GetKeyStore())
	for i, e := range ele.EleList {
		isVerify := false
		//检查地址hash是否在本地
		if checkExit && !isExit {
			jury := p.getLocalJuryAccount()
			log.Debugf("[%s]isValidateElection, addrHash[%s]", reqId.ShortStr(), e.AddrHash.String())
			if bytes.Equal(e.AddrHash.Bytes(), util.RlpHash(jury.Address).Bytes()) {
				isExit = true
			}
		}
		//检查指定节点模式下，是否为jjh请求地址
		if e.EType == 1 {
			if cType == modules.APP_CONTRACT_INVOKE_REQUEST {
				continue
			} else {
				isJjh := false
				for _, reqAddr := range reqAddrs {
					if jjhAd == reqAddr.Str() { //true
						log.Debugf("[%s]isValidateElection, e.EType == 1,jjh request addr, ok", reqId.ShortStr())
						isJjh = true
						break
					}
				}
				if isJjh {
					continue
				} else {
					log.Debugf("[%s]isValidateElection, e.EType == 1, but not jjh request addr", reqId.ShortStr())
					//log.Debugf("[%s]isValidateElection, reqAddr[%s], jjh[%s]", reqId.ShortStr(), reqAddr.Str(), jjhAd)
					return false
				}
			}
		}
		//检查地址与pubKey是否匹配:获取当前pubKey下的Addr，将地址hash后与输入比较
		addr := crypto.PubkeyBytesToAddress(e.PublicKey)
		if e.AddrHash != util.RlpHash(addr) {
			log.Errorf("[%s]isValidateElection, publicKey not match address, addrHash[%v]", reqId.ShortStr(), e.AddrHash)
			return false
		}
		//从数据库中查询该地址是否为Jury
		if !p.dag.IsActiveJury(addr) {
			log.Errorf("[%s]isValidateElection, not active Jury, addrHash[%v]", reqId.ShortStr(), e.AddrHash)
			return false
		}
		isVerify, err := elr.verifyVrf(e.Proof, conversionElectionSeedData(contractId), e.PublicKey)
		if err != nil || !isVerify {
			log.Infof("[%s]isValidateElection, index[%d],verifyVrf fail, contractId[%s]",
				reqId.ShortStr(), i, string(contractId))
			return false
		}
	}
	if checkExit {
		if !isExit {
			log.Debugf("[%s]isValidateElection, election addr not in local", reqId.ShortStr())
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
			log.Debugf("[%s]contractEventExecutable, CONTRACT_EVENT_ELE, Mediator, true", reqId.ShortStr())
			return true
		}
	case CONTRACT_EVENT_EXEC:
		if isSysContract && isMediator {
			log.Debugf("[%s]contractEventExecutable, CONTRACT_EVENT_EXEC, Mediator, true", reqId.ShortStr())
			return true
		} else if !isSysContract && isJury {
			if p.isValidateElection(tx, ele, true) {
				log.Debugf("[%s]contractEventExecutable, CONTRACT_EVENT_EXEC, Jury, true", reqId.ShortStr())
				return true
			} else {
				log.Debugf("[%s]contractEventExecutable, CONTRACT_EVENT_EXEC, Jury, isValidateElection fail, false",
					reqId.ShortStr())
			}
		}
	case CONTRACT_EVENT_SIG:
		if !isSysContract && isJury {
			if p.isValidateElection(tx, ele, false) {
				log.Debugf("[%s]contractEventExecutable, CONTRACT_EVENT_SIG, Jury, true", reqId.ShortStr())
				return true
			} else {
				log.Debugf("[%s]contractEventExecutable, CONTRACT_EVENT_SIG, Jury, isValidateElection fail, false",
					reqId.ShortStr())
			}
		}
	case CONTRACT_EVENT_COMMIT:
		if isMediator {
			if isSysContract {
				log.Debugf("[%s]contractEventExecutable, CONTRACT_EVENT_COMMIT, Mediator, sysContract, true",
					reqId.ShortStr())
				return true
			} else if !isSysContract && p.isValidateElection(tx, ele, false) {
				log.Debugf("[%s]contractEventExecutable, CONTRACT_EVENT_COMMIT, Mediator, userContract, true",
					reqId.ShortStr())
				return true
			} else {
				log.Debugf("[%s]contractEventExecutable, CONTRACT_EVENT_COMMIT, Mediator, isValidateElection fail, false:",
					reqId.ShortStr())
			}
		}
	}
	return false
}
func (p *Processor) CreateTokenTransaction(from, to common.Address, token *modules.Asset, daoAmountToken, daoFee uint64,
	msg *modules.Message) (*modules.Transaction, uint64, error) {
	
	if msg.App == modules.APP_DATA {
		size := float64(modules.CalcDateSize(msg.Payload))
		pricePerKByte := p.dag.GetChainParameters().TransferPtnPricePerKByte
		daoFee += uint64(size * float64(pricePerKByte) / 1024)
	}
	tx := &modules.Transaction{}
	// 1. 获取转出账户所有的PTN utxo
	assetId := dagconfig.DagConfig.GetGasToken()
	coreUtxos, err := p.ptn.TxPool().GetAddrUtxos(from, assetId.ToAsset())
	if err != nil {
		return nil, 0, err
	}
	if len(coreUtxos) == 0 {
		return nil, 0, err
	}
	//if p.ptn.EnableGasFee() {
	// must pay gasfee
	if daoFee == 0 {
		return nil, 0, fmt.Errorf("transaction fee cannot be 0")
	}
	daoTotal := daoAmountToken + daoFee
	if daoTotal > p.dag.GetPtnBalance(from) {
		return nil, 0, fmt.Errorf("the ptn balance of the account is not enough %v", daoTotal)
	}
	//2. 获取 PaymentPayload
	ploadPTN, err := p.dag.GetPayload(from, to, daoAmountToken, daoFee, coreUtxos)
	if err != nil {
		return nil, 0, err
	}
	tx.AddMessage(modules.NewMessage(modules.APP_PAYMENT, ploadPTN))
	return tx, daoFee, nil
}

func (p *Processor) createContractTxReqToken(contractId, from, to common.Address, token *modules.Asset,
	daoAmountToken, daoFee uint64, msg *modules.Message) (common.Hash, *modules.Transaction, error) {
	tx, _, err := p.CreateTokenTransaction(from, to, token, daoAmountToken, daoFee, msg)
	if err != nil {
		return common.Hash{}, nil, err
	}
	log.Debugf("[%s]createContractTxReqToken,contractId[%s],tx[%v]",
		tx.RequestHash().ShortStr(), contractId.String(), tx)
	return p.signGenericTx(contractId, from, tx)
}

func (p *Processor) createContractTxReq(contractId, from, to common.Address, daoAmount, daoFee uint64, certID *big.Int,
	msg *modules.Message) (common.Hash, *modules.Transaction, error) {
	tx, _, err := p.dag.CreateGenericTransaction(from, to, daoAmount, daoFee, certID, msg, p.ptn.EnableGasFee())
	if err != nil {
		return common.Hash{}, nil, err
	}
	//此处应该使用dag 和pool 中的所有utxo
	//构造 signature
	if !p.ptn.EnableGasFee() && from.Equal(to) {
		ks := p.ptn.GetKeyStore()
		pubKey, err := ks.GetPublicKey(from)
		if err != nil {
			return common.Hash{}, nil, err
		}
		sign, err := ks.SigData(tx, from)
		if err != nil {
			return common.Hash{}, nil, err
		}
		tx.AddMessage(modules.NewMessage(modules.APP_SIGNATURE, modules.NewSignaturePayload(pubKey, sign)))
	}
	return p.signGenericTx(contractId, from, tx)
}
func (p *Processor) SignAndExecuteAndSendRequest(from common.Address,
	tx *modules.Transaction) (*modules.Transaction, error) {
	requestMsg := tx.Messages()[tx.GetRequestMsgIndex()]
	if requestMsg.App == modules.APP_CONTRACT_INVOKE_REQUEST {
		request := requestMsg.Payload.(*modules.ContractInvokeRequestPayload)
		contractId := common.NewAddress(request.ContractId, common.ContractHash)
		reqId, tx, err := p.signGenericTx(contractId, from, tx)
		if err != nil {
			return nil, err
		}
		//broadcast
		go p.ptn.ContractBroadcast(ContractEvent{CType: CONTRACT_EVENT_EXEC, Ele: p.mtx[reqId].eleNode, Tx: tx}, true)
		return tx, nil
	}
	return nil, errors.New("Not support request")
}
func (p *Processor) signGenericTx(contractId common.Address, from common.Address,
	tx *modules.Transaction) (common.Hash, *modules.Transaction, error) {
	tx, err := p.ptn.SignGenericTransaction(from, tx)
	if err != nil {
		return common.Hash{}, nil, err
	}
	p.locker.Lock()
	defer p.locker.Unlock()

	reqId := tx.RequestHash()
	if p.ptn.EnableGasFee() {
		if p.validator.ValidateTxFeeEnough(tx, ContractDefaultSignatureSize+ContractDefaultRWSize,
			0) != validator.TxValidationCode_VALID {
			return common.Hash{}, nil, fmt.Errorf("signGenericTx, tx fee is invalid")
		}
	}
	log.Debugf("[%s]signGenericTx, contractId[%s]", reqId.ShortStr(), contractId.String())
	if p.mtx[reqId] != nil {
		return reqId, nil, fmt.Errorf("signGenericTx, contract request transaction[%s] already created", reqId.ShortStr())
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
				log.Errorf("[%s]signGenericTx, getContractElectionList fail,err:%s",
					tx.RequestHash().ShortStr(), err.Error())
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
					log.Infof("[%s]ContractTxDeleteLoop, contract is invalid, delete tx id", k.ShortStr())
					delete(p.mtx, k)
				}
			} else {
				if time.Since(v.tm) > time.Second*600 {
					log.Infof("[%s]ContractTxDeleteLoop, contract is valid, delete tx id", k.ShortStr())
					delete(p.mtx, k)
				}
			}
		}
		for k, v := range p.mel {
			if time.Since(v.tm) > time.Second*300 {
				log.Infof("[%s]ContractTxDeleteLoop, delete electionVrf ", k.ShortStr())
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
	_, msg, err := getContractTxContractInfo(tx, modules.APP_CONTRACT_DEPLOY_REQUEST)
	if err != nil {
		return nil, fmt.Errorf("[%s]getContractAssignElectionList, getContractTxContractInfo fail", reqId.ShortStr())
	}

	num := 0
	eels := make([]modules.ElectionInf, 0)
	tplId := msg.Payload.(*modules.ContractDeployRequestPayload).TemplateId
	//find the address of the contract template binding in the dag
	//addrHash, err := p.getTemplateAddrHash(tplId)
	tpl, err := p.dag.GetContractTpl(tplId)
	if err != nil {
		log.Debugf("[%s]getContractAssignElectionList, getTemplateAddrHash fail,templateId[%x], fail:%s",
			reqId.ShortStr(), tplId, err.Error())
		return nil, fmt.Errorf("[%s]getContractAssignElectionList, GetContractTpl fail", reqId.ShortStr())
	}
	addrHash := tpl.AddrHash
	cfgEleNum := getSysCfgContractElectionNum(p.dag)
	if len(addrHash) >= cfgEleNum {
		num = cfgEleNum
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
func (p *Processor) BuildUnitTxs(rwM *rwset.RwSetTxMgr, mDag dboperation.IContractDag, sortedTxs []*modules.Transaction, addr common.Address) ([]*modules.Transaction, error) {
	txs := []*modules.Transaction{}
	for i, tx := range sortedTxs {
		//var saveTx *modules.Transaction
		saveTx := tx
		log.Debugf("buildUnitTxs, idx[%d] txReqId[%s]:IsContractTx[%v]",
			i, tx.RequestHash().String(), tx.IsContractTx())
		if tx.IsContractTx() && tx.IsOnlyContractRequest() { //只处理请求合约
			if tx.IsSystemContract() { //执行系统合约
				signedTx, err := p.RunAndSignTx(tx, rwM, mDag, addr)
				if err != nil {
					log.Errorf("BuildUnitTxs,run contract request[%s] fail:%s", tx.Hash(), err.Error())
					continue
				}
				saveTx = signedTx
			} else { //用户合约,需要从交易池中获取交易请求,根据请求Id再从Processor中获取最终执行后的交易

				//用户合约请求不打包到Unit中
				continue
				//if mtx, ok := p.mtx[tx.RequestHash()]; ok {
				//	log.Debugf("[%s]BuildUnitTxs, get tx from mtx", shortId(tx.RequestHash().String()))
				//
				//	if mtx.rstTx != nil {
				//		saveTx = mtx.rstTx
				//		mtx.valid = false
				//		log.Debugf("[%s]BuildUnitTxs, mtx include tx", shortId(tx.RequestHash().String()))
				//	}
				//}
			}
		} // else { //直接保存交易
		//saveTx = tx
		//}

		if saveTx == nil {
			log.Debugf("buildUnitTxs,  saveTx is nil, idx[%d]tx[%s]", i, tx.RequestHash().String())
			continue
		}

		err := mDag.SaveTransaction(saveTx, i+1) //第0条是Coinbase
		if err != nil {
			log.Errorf("buildUnitTxs, idx[%d] txReqId[%s]-hash[%s]:",
				i, saveTx.RequestHash().String(), saveTx.Hash().String())
		}
		txs = append(txs, saveTx)
		log.Debugf("buildUnitTxs, add idx[%d] txReqId[%s]-hash[%s]:",
			i, saveTx.RequestHash().String(), saveTx.Hash().String())
	}

	return txs, nil
}

func (p *Processor) AddLocalTx(tx *modules.Transaction) error {
	if tx == nil {
		return errors.New("AddLocalTx, tx is nil")
	}

	reqId := tx.RequestHash()
	txHash := tx.Hash()
	poolTx, _ := p.ptn.TxPool().GetTx(txHash)
	if poolTx == nil { //tx not in txpool
		err := p.ptn.TxPool().AddLocal(tx)
		if err != nil {
			log.Errorf("[%s]AddLocalTx, AddLocal err:%s", reqId.ShortStr(), err.Error())
			return err
		}
	}

	isExist, _ := p.dag.IsTransactionExist(txHash)
	if isExist {
		log.Debugf("[%s]AddLocalTx,tx already exist dag", reqId.ShortStr())
		return nil
	}

	err := p.dag.SaveLocalTx(tx)
	if err != nil {
		log.Errorf("[%s]AddLocalTx, SaveLocalTx err:%s", reqId.ShortStr(), err.Error())
		return err
	}

	//先将状态改为交易池中
	err = p.dag.SaveLocalTxStatus(tx.Hash(), modules.TxStatus_InPool)
	if err != nil {
		log.Warnf("[%s]AddLocalTx, SaveLocalTxStatus err:%s", reqId.ShortStr(), err.Error())
		return err
	}

	//更新Tx的状态到LocalDB
	go func(txHash common.Hash) {
		saveUnitCh := make(chan modules.SaveUnitEvent, 10)
		defer close(saveUnitCh)
		saveUnitSub := p.dag.SubscribeSaveUnitEvent(saveUnitCh)
		headCh := make(chan modules.SaveUnitEvent, 10)
		defer close(headCh)
		headSub := p.dag.SubscribeSaveStableUnitEvent(headCh)
		defer saveUnitSub.Unsubscribe()
		defer headSub.Unsubscribe()
		timeout := time.NewTimer(100 * time.Second)
		for {
			select {
			case u := <-saveUnitCh:
				log.Infof("AddLocalTx, SubscribeSaveUnitEvent received unit:%s", u.Unit.DisplayId())
				for _, utx := range u.Unit.Transactions() {
					if utx.Hash() == txHash || utx.RequestHash() == txHash {
						log.Infof("[%s]AddLocalTx, Change local tx[%s] status to unstable",
							reqId.ShortStr(), txHash.String())
						err = p.dag.SaveLocalTxStatus(txHash, modules.TxStatus_Unstable)
						if err != nil {
							log.Warnf("[%s]AddLocalTx, Save tx[%s] status to local err:%s",
								reqId.ShortStr(), txHash.String(), err.Error())
						}
					}
				}
			case u := <-headCh:
				log.Infof("AddLocalTx, SubscribeSaveStableUnitEvent received unit:%s", u.Unit.DisplayId())
				for _, utx := range u.Unit.Transactions() {
					if utx.Hash() == txHash || utx.RequestHash() == txHash {
						log.Debugf("[%s]AddLocalTx, Change local tx[%s] status to stable",
							reqId.ShortStr(), txHash.String())
						err = p.dag.SaveLocalTxStatus(txHash, modules.TxStatus_Stable)
						if err != nil {
							log.Warnf("[%s]AddLocalTx, Save tx[%s] status to local err:%s",
								reqId.ShortStr(), txHash.String(), err.Error())
						}
						return
					}
				}
			case <-timeout.C:
				log.Warnf("[%s]AddLocalTx, SubscribeSaveStableUnitEvent timeout for tx[%s]",
					reqId.ShortStr(), txHash.String())
				return
			// Err() channel will be closed when unsubscribing.
			case <-headSub.Err():
				log.Debugf("SubscribeSaveStableUnitEvent err")
				return
			case <-saveUnitSub.Err():
				log.Debugf("SubscribeSaveUnitEvent err")
				return
			}
		}
	}(tx.Hash())

	return nil
}
