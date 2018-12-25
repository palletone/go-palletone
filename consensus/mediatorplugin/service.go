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
 * @author PalletOne core developer Albert·Gou <dev@pallet.one>
 * @date 2018
 */

package mediatorplugin

import (
	"fmt"
	"time"

	"github.com/dedis/kyber"
	"github.com/dedis/kyber/pairing/bn256"
	"github.com/dedis/kyber/share/dkg/pedersen"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/event"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/p2p"
	"github.com/palletone/go-palletone/common/rpc"
	"github.com/palletone/go-palletone/consensus/jury"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/core/accounts/keystore"
	"github.com/palletone/go-palletone/core/node"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/storage"
	"github.com/palletone/go-palletone/dag/txspool"
)

// PalletOne wraps all methods required for producing unit.
type PalletOne interface {
	GetKeyStore() *keystore.KeyStore
	TxPool() txspool.ITxPool

	SignGenericTransaction(from common.Address, tx *modules.Transaction) (*modules.Transaction, error)
	SignAndSendTransaction(addr common.Address, tx *modules.Transaction) error

	ContractProcessor() *jury.Processor
}

type iDag interface {
	ChainThreshold() int
	GetSlotAtTime(when time.Time) uint32
	GetSlotTime(slotNum uint32) time.Time
	HeadUnitTime() int64
	GetScheduledMediator(slotNum uint32) common.Address
	GetActiveMediatorInitPubs() []kyber.Point
	GetActiveMediatorCount() int
	GetActiveMediatorAddr(index int) common.Address
	HeadUnitNum() uint64
	GetUnitByHash(common.Hash) (*modules.Unit, error)

	IsActiveMediator(add common.Address) bool
	IsSynced() bool

	ValidateUnitExceptGroupSig(unit *modules.Unit, isGenesis bool) bool

	GenerateUnit(when time.Time, producer common.Address, groupPubKey []byte,
		ks *keystore.KeyStore, txspool txspool.ITxPool) *modules.Unit

	ActiveMediators() map[common.Address]bool

	IsPrecedingMediator(add common.Address) bool
	IsIrreversibleUnit(hash common.Hash) bool

	CurrentFeeSchedule() core.FeeSchedule
	GenMediatorCreateTx(account common.Address, op *modules.MediatorCreateOperation) (*modules.Transaction, uint64, error)
	GenVoteMediatorTx(voter, mediator common.Address) (*modules.Transaction, uint64, error)

	GetMediators() map[common.Address]bool
	IsMediator(address common.Address) bool

	GetAllMediatorInCandidateList() ([]*modules.MediatorInfo, error)
	IsInMediatorCandidateList(address common.Address) bool

	GetVotedMediator(addr common.Address) map[common.Address]bool
	GetDynGlobalProp() *modules.DynamicGlobalProperty
	GetMediatorInfo(address common.Address) *storage.MediatorInfo
}

type MediatorPlugin struct {
	ptn  PalletOne     // Full PalletOne service to retrieve other function
	quit chan struct{} // Channel used for graceful exit
	dag  iDag

	// Enable Unit production, even if the chain is stale.
	// 新开启一条链时，第一个节点必须设为true，其他节点必须设为false
	productionEnabled bool
	// Mediator`s account and passphrase controlled by this node
	mediators map[common.Address]*MediatorAccount

	// 新生产unit的事件订阅
	newProducedUnitFeed  event.Feed              // 订阅的时候自动初始化一次
	newProducedUnitScope event.SubscriptionScope // 零值已准备就绪待用

	// dkg 初始化 相关
	suite         *bn256.Suite
	activeDKGs    map[common.Address]*dkg.DistKeyGenerator
	precedingDKGs map[common.Address]*dkg.DistKeyGenerator

	// dkg 完成 vss 协议相关
	respBuf map[common.Address]map[common.Address]chan *dkg.Response

	// 广播和处理 vss 协议 deal
	vssDealFeed  event.Feed
	vssDealScope event.SubscriptionScope

	// 广播和处理 vss 协议 response
	vssResponseFeed  event.Feed
	vssResponseScope event.SubscriptionScope

	// unit阈值签名相关
	toTBLSSignBuf    map[common.Address]chan *modules.Unit
	toTBLSRecoverBuf map[common.Address]map[common.Hash]*sigShareSet

	// unit 签名分片的事件订阅
	sigShareFeed  event.Feed
	sigShareScope event.SubscriptionScope

	// unit 群签名的事件订阅
	groupSigFeed  event.Feed
	groupSigScope event.SubscriptionScope
}

func (mp *MediatorPlugin) Protocols() []p2p.Protocol {
	return nil
}

func (mp *MediatorPlugin) APIs() []rpc.API {
	return []rpc.API{
		{
			Namespace: "mediator",
			Version:   "1.0",
			Service:   NewPublicMediatorAPI(mp),
			Public:    true,
		},
		{
			Namespace: "mediator",
			Version:   "1.0",
			Service:   NewPrivateMediatorAPI(mp),
			Public:    false,
		},
	}
}

func (mp *MediatorPlugin) isLocalMediator(add common.Address) bool {
	_, ok := mp.mediators[add]

	return ok
}

func (mp *MediatorPlugin) ScheduleProductionLoop() {
	// 1. 判断是否满足生产验证单元的条件，主要判断本节点是否控制至少一个mediator账户
	if len(mp.mediators) == 0 {
		println("No mediators configured! Please add mediator and private keys to configuration.")
	} else {
		// 2. 开启循环生产计划
		go log.Info(fmt.Sprintf("Launching unit production for %d mediators.", len(mp.mediators)))

		if mp.productionEnabled {
			dag := mp.dag
			if dag.HeadUnitNum() == 0 {
				newChainBanner(dag)
			}
		}

		// 调度生产unit
		go mp.scheduleProductionLoop()
	}
}

func (mp *MediatorPlugin) newActiveMediatorsDKG() {
	dag := mp.dag
	if !mp.productionEnabled && !dag.IsSynced() {
		return //we're not synced.
	}

	lams := mp.GetLocalActiveMediators()
	initPubs := dag.GetActiveMediatorInitPubs()
	curThreshold := dag.ChainThreshold()
	lamc := len(lams)

	mp.activeDKGs = make(map[common.Address]*dkg.DistKeyGenerator, lamc)
	mp.respBuf = make(map[common.Address]map[common.Address]chan *dkg.Response, lamc)

	for _, localMed := range lams {
		initSec := mp.mediators[localMed].InitPartSec

		//dkgr, err := dkg.NewDistKeyGeneratorWithoutSecret(mp.suite, initSec, initPubs, curThreshold)
		dkgr, err := dkg.NewDistKeyGenerator(mp.suite, initSec, initPubs, curThreshold)
		if err != nil {
			log.Debug(err.Error())
			continue
		}

		mp.activeDKGs[localMed] = dkgr
		mp.initRespBuf(localMed)
	}
}

func (mp *MediatorPlugin) initRespBuf(localMed common.Address) {
	aSize := mp.dag.GetActiveMediatorCount()
	mp.respBuf[localMed] = make(map[common.Address]chan *dkg.Response, aSize)

	for i := 0; i < aSize; i++ {
		vrfrMed := mp.dag.GetActiveMediatorAddr(i)
		mp.respBuf[localMed][vrfrMed] = make(chan *dkg.Response, aSize-1)
	}
}

func (mp *MediatorPlugin) Start(server *p2p.Server) error {
	log.Debug("mediator plugin startup begin")

	// 1. 开启循环生产计划
	go mp.ScheduleProductionLoop()

	log.Debug("mediator plugin startup end")

	return nil
}

func (mp *MediatorPlugin) UpdateMediatorsDKG() {
	// 1. 保存旧的 dkg ， 用于之前的unit群签名确认
	mp.switchMediatorsDKG()

	// 2. 初始化当前节点控制的活跃mediator对应的DKG.
	mp.newActiveMediatorsDKG()

	// 3. 开始完成 vss 协议
	go mp.startVSSProtocol()
}

func (mp *MediatorPlugin) switchMediatorsDKG() {
	mp.precedingDKGs = mp.activeDKGs
	mp.activeDKGs = make(map[common.Address]*dkg.DistKeyGenerator)
}

func (mp *MediatorPlugin) Stop() error {
	close(mp.quit)

	mp.newProducedUnitScope.Close()
	mp.vssDealScope.Close()
	mp.vssResponseScope.Close()
	mp.sigShareScope.Close()
	mp.groupSigScope.Close()

	log.Debug("mediator plugin stopped")

	return nil
}

// 匿名函数的好处之一：能在匿名函数内部直接使用本函数之外的变量;
// 函数使用外部变量的特性称之为闭包； 例如，以下匿名方法就直接使用cfg变量
func RegisterMediatorPluginService(stack *node.Node, cfg *Config) {
	log.Debug("Register Mediator Plugin Service...")

	err := stack.Register(func(ctx *node.ServiceContext) (node.Service, error) {
		// Retrieve ptn service
		var ptn PalletOne
		err := ctx.Service(&ptn)
		if err != nil {
			return nil, fmt.Errorf("the PalletOne service not found: %v", err)
		}

		//return NewMediatorPlugin(ptn, ptn.Dag(), cfg)
		return nil, nil
	})

	if err != nil {
		log.Debug(fmt.Sprintf("failed to register the Mediator Plugin service: %v", err))
	}
}

func NewMediatorPlugin(ptn PalletOne, dag iDag, cfg *Config) (*MediatorPlugin, error) {
	log.Debug("mediator plugin initialize begin")

	if ptn == nil || dag == nil || cfg == nil {
		err := "pointer parameters of NewMediatorPlugin are nil!"
		//log.Error(err)
		panic(err)
	}

	mss := cfg.Mediators
	msm := make(map[common.Address]*MediatorAccount, 0)

	for _, medConf := range mss {
		medAcc := medConf.configToAccount()
		addr := medAcc.Address
		//log.Debug(fmt.Sprintf("this node control mediator account address: %v", addr.Str()))

		msm[addr] = medAcc
	}

	log.Debug(fmt.Sprintf("This node controls %v mediators.", len(msm)))

	mp := MediatorPlugin{
		ptn:  ptn,
		quit: make(chan struct{}),
		dag:  dag,

		productionEnabled: cfg.EnableStaleProduction,
		mediators:         msm,

		suite:         core.Suite,
		activeDKGs:    make(map[common.Address]*dkg.DistKeyGenerator),
		precedingDKGs: make(map[common.Address]*dkg.DistKeyGenerator),
	}
	mp.initTBLSBuf()

	log.Debug("mediator plugin initialize end")

	return &mp, nil
}

// initTBLSBuf, 初始化与TBLS签名相关的buf
func (mp *MediatorPlugin) initTBLSBuf() {
	lams := mp.GetLocalActiveMediators()
	lamc := len(mp.mediators)

	mp.toTBLSSignBuf = make(map[common.Address]chan *modules.Unit, lamc)
	mp.toTBLSRecoverBuf = make(map[common.Address]map[common.Hash]*sigShareSet, lamc)

	curThrshd := mp.dag.ChainThreshold()
	for _, localMed := range lams {
		mp.toTBLSSignBuf[localMed] = make(chan *modules.Unit, curThrshd)
		mp.toTBLSRecoverBuf[localMed] = make(map[common.Hash]*sigShareSet, curThrshd)
	}
}
