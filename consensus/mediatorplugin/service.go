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

	"github.com/dedis/kyber"
	"github.com/dedis/kyber/pairing/bn256"
	"github.com/dedis/kyber/share/dkg/pedersen"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/event"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/p2p"
	"github.com/palletone/go-palletone/common/rpc"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/core/accounts/keystore"
	"github.com/palletone/go-palletone/core/node"
	"github.com/palletone/go-palletone/dag"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/txspool"
)

// PalletOne wraps all methods required for producing unit.
type PalletOne interface {
	Dag() dag.IDag
	GetKeyStore() *keystore.KeyStore
	TxPool() txspool.ITxPool
}

type MediatorPlugin struct {
	ptn  PalletOne     // Full PalletOne service to retrieve other function
	quit chan struct{} // Channel used for graceful exit
	// Enable VerifiedUnit production, even if the chain is stale.
	// 新开启一个区块链时，必须设为true
	productionEnabled bool
	// Mediator`s account and passphrase controlled by this node
	mediators map[common.Address]MediatorAccount

	// 新生产unit的事件订阅
	newUnitFeed  event.Feed              // 订阅的时候自动初始化一次
	newUnitScope event.SubscriptionScope // 零值已准备就绪待用

	// dkg 初始化 相关
	suite *bn256.Suite
	dkgs  map[common.Address]*dkg.DistKeyGenerator

	// dkg 完成 vss 协议相关
	respBuf map[common.Address]map[common.Address]chan *dkg.Response
	certifiedFlag map[common.Address]bool

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
	return nil
}

func (mp *MediatorPlugin) GetLocalActiveMediators() []common.Address {
	lams := make([]common.Address, 0)

	dag := mp.getDag()
	for add := range mp.mediators {
		if dag.IsActiveMediator(add) {
			lams = append(lams, add)
		}
	}

	return lams
}

func (mp *MediatorPlugin) LocalHaveActiveMediator() bool {
	lams := mp.GetLocalActiveMediators()

	return len(lams) != 0
}

func (mp *MediatorPlugin) isLocalMediator(add common.Address) bool {
	_, ok := mp.mediators[add]

	return ok
}

func (mp *MediatorPlugin) IsLocalActiveMediator(add common.Address) bool {
	if mp.isLocalMediator(add) {
		return mp.getDag().IsActiveMediator(add)
	}

	return false
}

func (mp *MediatorPlugin) ScheduleProductionLoop() {
	// 1. 判断是否满足生产验证单元的条件，主要判断本节点是否控制至少一个mediator账户
	if len(mp.mediators) == 0 {
		println("No mediators configured! Please add mediator and private keys to configuration.")
	} else {
		// 2. 开启循环生产计划
		log.Info(fmt.Sprintf("Launching verified unit production for %d mediators.", len(mp.mediators)))

		if mp.productionEnabled {
			dag := mp.getDag()
			if dag.GetDynGlobalProp().LastVerifiedUnitNum == 0 {
				newChainBanner(dag)
			}
		}

		// 调度生产unit
		go mp.scheduleProductionLoop()
	}
}

func (mp *MediatorPlugin) NewActiveMediatorsDKG() {
	log.Info("instantiate the DistKeyGenerator (DKG) struct.")

	dag := mp.getDag()
	if !mp.productionEnabled && !dag.IsSynced() {
		return //we're not synced.
	}

	lams := mp.GetLocalActiveMediators()
	initPubs := dag.GetActiveMediatorInitPubs()
	curThreshold := dag.GetCurThreshold()
	lamc := len(lams)

	mp.dkgs = make(map[common.Address]*dkg.DistKeyGenerator, lamc)
	mp.respBuf = make(map[common.Address]map[common.Address]chan *dkg.Response, lamc)
	mp.certifiedFlag = make(map[common.Address]bool, lamc)

	for _, localMed := range lams {
		initSec := mp.mediators[localMed].InitPartSec

		//dkgr, err := dkg.NewDistKeyGeneratorWithoutSecret(mp.suite, initSec, initPubs, curThreshold)
		dkgr, err := dkg.NewDistKeyGenerator(mp.suite, initSec, initPubs, curThreshold)
		if err != nil {
			log.Error(err.Error())
			continue
		}

		mp.dkgs[localMed] = dkgr
		mp.initRespBuf(localMed)
	}

	go mp.StartVSSProtocol()
}

func (mp *MediatorPlugin) initRespBuf(localMed common.Address) {
	aSize := mp.getDag().GetActiveMediatorCount()
	mp.respBuf[localMed] = make(map[common.Address]chan *dkg.Response, aSize)

	for i := 0; i < aSize; i++ {
		vrfrMed := mp.getDag().GetActiveMediatorAddr(i)
		mp.respBuf[localMed][vrfrMed] = make(chan *dkg.Response, aSize-1)
	}
}

func (mp *MediatorPlugin) Start(server *p2p.Server) error {
	log.Debug("mediator plugin startup begin")

	// 1. 开启循环生产计划
	go mp.ScheduleProductionLoop()

	// 2. 给当前节点控制的活跃mediator，初始化对应的DKG.
	// todo 数据同步后再初始化，换届后需重新生成
	go mp.NewActiveMediatorsDKG()

	log.Debug("mediator plugin startup end")

	return nil
}

func (mp *MediatorPlugin) Stop() error {
	close(mp.quit)
	mp.newUnitScope.Close()
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

		return NewMediatorPlugin(ptn, cfg)
	})

	if err != nil {
		log.Error(fmt.Sprintf("failed to register the Mediator Plugin service: %v", err))
	}
}

func NewMediatorPlugin(ptn PalletOne, cfg *Config) (*MediatorPlugin, error) {
	log.Debug("mediator plugin initialize begin")

	mss := cfg.Mediators
	msm := map[common.Address]MediatorAccount{}

	for _, medConf := range mss {
		medAcc := ConfigToAccount(medConf)
		addr := medAcc.Address
		//log.Debug(fmt.Sprintf("this node control mediator account address: %v", addr.Str()))

		msm[addr] = medAcc
	}

	log.Debug(fmt.Sprintf("This node controls %v mediators.", len(msm)))

	mp := MediatorPlugin{
		ptn:               ptn,
		productionEnabled: cfg.EnableStaleProduction,
		mediators:         msm,
		quit:              make(chan struct{}),

		suite: bn256.NewSuiteG2(),
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

	curThrshd := mp.getDag().GetCurThreshold()
	for _, localMed := range lams {
		mp.toTBLSSignBuf[localMed] = make(chan *modules.Unit, curThrshd)
		mp.toTBLSRecoverBuf[localMed] = make(map[common.Hash]*sigShareSet, curThrshd)
	}
}

func (mp *MediatorPlugin) getDag() dag.IDag {
	return mp.ptn.Dag()
}

func ConfigToAccount(medConf MediatorConf) MediatorAccount {
	// 1. 解析 mediator 账户地址
	addr := core.StrToMedAdd(medConf.Address)

	// 2. 解析 mediator 的 DKS 初始公私钥
	sec := core.StrToScalar(medConf.InitPartSec)
	pub := core.StrToPoint(medConf.InitPartPub)

	medAcc := MediatorAccount{
		addr,
		medConf.Password,
		sec,
		pub,
	}

	return medAcc
}

type MediatorAccount struct {
	Address     common.Address
	Password    string
	InitPartSec kyber.Scalar
	InitPartPub kyber.Point
}
