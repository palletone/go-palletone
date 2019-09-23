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
	"sync"
	"time"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/event"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/p2p"
	"github.com/palletone/go-palletone/common/rpc"
	"github.com/palletone/go-palletone/consensus/jury"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/core/accounts"
	"github.com/palletone/go-palletone/core/accounts/keystore"
	"github.com/palletone/go-palletone/core/node"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/txspool"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/pairing/bn256"
	"go.dedis.ch/kyber/v3/share/dkg/pedersen"
)

// PalletOne wraps all methods required for producing unit.
type PalletOne interface {
	GetKeyStore() *keystore.KeyStore
	TxPool() txspool.ITxPool
	ContractProcessor() *jury.Processor
}

type iDag interface {
	ChainThreshold() int
	GetSlotAtTime(when time.Time) uint32
	GetSlotTime(slotNum uint32) time.Time
	HeadUnitTime() int64

	GetScheduledMediator(slotNum uint32) common.Address
	GetActiveMediatorInitPubs() []kyber.Point
	ActiveMediatorsCount() int
	GetActiveMediatorAddr(index int) common.Address
	HeadUnitNum() uint64
	GetHeaderByHash(common.Hash) (*modules.Header, error)
	//GetUnitByHash(common.Hash) (*modules.Unit, error)

	GetGlobalProp() *modules.GlobalProperty
	GetActiveMediators() []common.Address
	IsActiveMediator(add common.Address) bool
	//IsSynced() bool
	//ValidateUnitExceptGroupSig(unit *modules.Unit) error
	SetUnitGroupSign(unitHash common.Hash, groupSign []byte, txpool txspool.ITxPool) error

	GenerateUnit(when time.Time, producer common.Address, groupPubKey []byte, ks *keystore.KeyStore,
		txpool txspool.ITxPool) (*modules.Unit, error)

	IsPrecedingMediator(add common.Address) bool
	IsIrreversibleUnit(hash common.Hash) bool
	IsMediator(address common.Address) bool

	PrecedingThreshold() int
	PrecedingMediatorsCount() int
	UnitIrreversibleTime() time.Duration
	LastMaintenanceTime() int64

	IsConsecutiveMediator(nextMediator common.Address) bool
	MediatorParticipationRate() uint32
}

type MediatorPlugin struct {
	ptn  PalletOne     // Full PalletOne service to retrieve other function
	quit chan struct{} // Channel used for graceful exit
	dag  iDag
	//srvr *p2p.Server

	// 标记是否主程序启动时，就开启unit生产功能
	producingEnabled bool
	stopProduce      chan struct{}
	// wait group is used for graceful shutdowns during producing unit
	wg sync.WaitGroup

	// Enable Unit production, even if the chain is stale.
	// 新开启一条链时，第一个运行的节点必须设为true，否则整个链无法启动
	// 其他节点必须设为false，否则容易导致分叉
	productionEnabled bool

	// 允许本节点的mediator可以连续生产unit, 只能使用一次
	consecutiveProduceEnabled bool
	// 本节点要求的mediator参与率，低于该参与率不生产unit
	requiredParticipation uint32
	// 群签名功能开启标记
	groupSigningEnabled bool

	// Mediator`s info controlled by this node, 本节点配置的mediator信息
	mediators map[common.Address]*MediatorAccount

	// 新生产unit的事件订阅
	newProducedUnitFeed  event.Feed              // 订阅的时候自动初始化一次
	newProducedUnitScope event.SubscriptionScope // 零值已准备就绪待用

	// dkg 处理和使用 相关
	suite               *bn256.Suite
	activeDKGs          map[common.Address]*dkg.DistKeyGenerator
	precedingDKGs       map[common.Address]*dkg.DistKeyGenerator
	lastMaintenanceTime int64
	dkgLock             *sync.RWMutex

	// dkg 完成 vss 协议相关
	dealBuf    map[common.Address]chan *dkg.Deal
	respBuf    map[common.Address]map[common.Address]chan *dkg.Response
	vssBufLock *sync.RWMutex
	stopVSS    chan struct{}

	// 广播和处理 vss 协议 deal
	vssDealFeed  event.Feed
	vssDealScope event.SubscriptionScope

	// 广播和处理 vss 协议 response
	vssResponseFeed  event.Feed
	vssResponseScope event.SubscriptionScope

	// unit阈值签名相关
	toTBLSSignBuf    map[common.Address]*sync.Map
	toTBLSRecoverBuf map[common.Address]map[common.Hash]*sigShareSet
	recoverBufLock   *sync.RWMutex

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

func (mp *MediatorPlugin) Start(server *p2p.Server) error {
	log.Debugf("mediator plugin startup begin")
	//mp.srvr = server

	// 解锁本地控制的mediator账户
	mp.unlockLocalMediators()

	// 开启循环生产计划
	if mp.producingEnabled {
		go mp.launchProduction()
	}

	// 开始完成 vss 协议
	// todo albert 待优化
	//if mp.groupSigningEnabled {
	//	go mp.startVSSProtocol()
	//}

	log.Debugf("mediator plugin startup end")
	return nil
}

func (mp *MediatorPlugin) unlockLocalMediators() {
	ks := mp.ptn.GetKeyStore()

	for add, medAcc := range mp.mediators {
		log.Debugf("try to unlock mediator account: %v", add.Str())
		err := ks.Unlock(accounts.Account{Address: add}, medAcc.Password)
		if err != nil {
			log.Debugf("fail to unlock the mediator(%v), error: %v", add.Str(), err.Error())
			//delete(mp.mediators, add)
		}
	}
}

func (mp *MediatorPlugin) launchProduction() {
	// 1. 判断是否满足生产unit的条件，主要判断本节点是否控制至少一个mediator账户
	if len(mp.mediators) == 0 {
		log.Debugf("No mediators configured! Please add mediator(s) to config file, and reboot process")
	} else {
		// 2. 开启循环生产计划
		log.Infof("Launching unit production for %v mediators.", len(mp.mediators))

		if mp.productionEnabled {
			if mp.dag.HeadUnitNum() == 0 {
				mp.newChainBanner()
			}
		}

		// 调度生产unit
		go mp.scheduleProductionLoop()
	}
}

func (mp *MediatorPlugin) Stop() error {
	close(mp.quit)

	mp.newProducedUnitScope.Close()
	mp.vssDealScope.Close()
	mp.vssResponseScope.Close()
	mp.sigShareScope.Close()
	mp.groupSigScope.Close()

	mp.wg.Wait()

	log.Infof("mediator plugin stopped")
	return nil
}

// 匿名函数的好处之一：能在匿名函数内部直接使用本函数之外的变量;
// 函数使用外部变量的特性称之为闭包； 例如，以下匿名方法就直接使用cfg变量
func RegisterMediatorPluginService(stack *node.Node, cfg *Config) {
	log.Debugf("Register Mediator Plugin Service...")

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
		log.Debugf("failed to register the Mediator Plugin service: %v", err)
	}
}

func NewMediatorPlugin( /*ctx *node.ServiceContext, */ cfg *Config, ptn PalletOne, dag iDag) (*MediatorPlugin, error) {
	log.Infof("mediator plugin initialize begin")

	if ptn == nil || dag == nil || cfg == nil {
		err := "pointer parameters of NewMediatorPlugin are nil"
		log.Errorf(err)
		panic(err)
	}

	mp := MediatorPlugin{
		ptn:  ptn,
		quit: make(chan struct{}),
		dag:  dag,

		producingEnabled: cfg.EnableProducing,
		stopProduce:      make(chan struct{}),

		productionEnabled:         cfg.EnableStaleProduction,
		consecutiveProduceEnabled: cfg.EnableConsecutiveProduction,
		requiredParticipation:     cfg.RequiredParticipation * core.PalletOne1Percent,
		groupSigningEnabled:       cfg.EnableGroupSigning,

		suite:               core.Suite,
		activeDKGs:          make(map[common.Address]*dkg.DistKeyGenerator),
		precedingDKGs:       make(map[common.Address]*dkg.DistKeyGenerator),
		lastMaintenanceTime: dag.LastMaintenanceTime(),
		dkgLock:             new(sync.RWMutex),
	}

	mp.initLocalConfigMediator(cfg.Mediators /*, ctx.AccountManager*/)

	if mp.groupSigningEnabled {
		mp.initGroupSignBuf()
	}

	log.Debugf("mediator plugin initialize end")
	return &mp, nil
}

func (mp *MediatorPlugin) initLocalConfigMediator(mcs []*MediatorConf /*, am *accounts.Manager*/) {
	mas := make(map[common.Address]*MediatorAccount, len(mcs))
	//ks := am.Backends(keystore.KeyStoreType)[0].(*keystore.KeyStore)

	for _, medConf := range mcs {
		medAcc := medConf.configToAccount()
		if medAcc == nil {
			continue
		}

		addr := medAcc.Address
		// 解锁本地配置的mediator账户
		//log.Debugf("try to unlock mediator account: %v", medConf.Address)
		//err := ks.Unlock(accounts.Account{Address: addr}, medAcc.Password)
		//if err != nil {
		//	log.Debugf("fail to unlock the mediator(%v), error: %v", medConf.Address, err.Error())
		//	continue
		//}

		log.Infof("this node control mediator account: %v", medConf.Address)
		mas[addr] = medAcc
	}

	log.Infof("This node controls %v mediators.", len(mas))
	mp.mediators = mas
}

// initGroupSignBuf, 初始化与群签名相关(vss和tbls)的buf
func (mp *MediatorPlugin) initGroupSignBuf() {
	lamc := len(mp.mediators)

	mp.dealBuf = make(map[common.Address]chan *dkg.Deal, lamc)
	mp.respBuf = make(map[common.Address]map[common.Address]chan *dkg.Response, lamc)
	mp.stopVSS = make(chan struct{})
	mp.vssBufLock = new(sync.RWMutex)

	mp.toTBLSSignBuf = make(map[common.Address]*sync.Map, lamc)
	mp.toTBLSRecoverBuf = make(map[common.Address]map[common.Hash]*sigShareSet, lamc)
	mp.recoverBufLock = new(sync.RWMutex)
}

func (mp *MediatorPlugin) UpdateMediatorsDKG(isRenew bool) {
	log.Debugf("UpdateMediatorsDKG when after performChainMaintenance")

	if !mp.groupSigningEnabled {
		return
	}

	// 保存旧的 dkg ， 用于之前的unit群签名确认
	mp.dkgLock.RLock()
	mp.precedingDKGs = mp.activeDKGs
	mp.lastMaintenanceTime = mp.dag.LastMaintenanceTime()
	mp.dkgLock.RUnlock()

	// 判断是否重新 初始化DKG 和 VSS 协议
	// todo albert 待优化
	//if !isRenew {
	//	return
	//}

	// 初始化当前节点控制的活跃mediator对应的DKG, 以及初始化完成vss相关的buf
	mp.newDKGAndInitVSSBuf()

	// 开始完成 vss 协议
	mp.startVSSProtocol()
}
