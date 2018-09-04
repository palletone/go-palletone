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
	"github.com/dedis/kyber/share/vss/pedersen"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/event"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/p2p"
	"github.com/palletone/go-palletone/common/p2p/discover"
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
	Dag() *dag.Dag
	GetKeyStore() *keystore.KeyStore
	TxPool() *txspool.TxPool
	GetActiveMediatorNodes() []*discover.Node
}

// toBLSed represents a BLS sign operation.
type toBLSSigned struct {
	origin string
	unit   *modules.Unit
}

// toTBLSSigned, TBLS signed auxiliary structure
type toTBLSSigned struct {
	unit      *modules.Unit
	sigShares [][]byte
}

type MediatorPlugin struct {
	server *p2p.Server   // Peer-to-peer server to maintain the connection with other active mediator peer
	ptn    PalletOne     // Full PalletOne service to retrieve other function
	quit   chan struct{} // Channel used for graceful exit
	// Enable VerifiedUnit production, even if the chain is stale.
	// 新开启一个区块链时，必须设为true
	productionEnabled bool
	// Mediator`s account and passphrase controlled by this node
	mediators map[common.Address]MediatorAccount

	// 新生产unit的事件订阅和数据发送和接收
	newProducedUnitFeed  event.Feed              // 订阅的时候自动初始化一次
	newProducedUnitScope event.SubscriptionScope // 零值已准备就绪待用
	toBLSSigned          chan *toBLSSigned       // 接收新生产的unit

	// dkg生成vss相关
	suite vss.Suite
	dkgs  map[common.Address]*dkg.DistKeyGenerator

	// unit阈值签名相关
	pendingTBLSSign map[common.Hash]*toTBLSSigned // 等待TBLS阈值签名的unit
}

func (mp *MediatorPlugin) Protocols() []p2p.Protocol {
	return nil
}

func (mp *MediatorPlugin) APIs() []rpc.API {
	return nil
}

func (mp *MediatorPlugin) GetLocalActiveMediators() []common.Address {
	ams := make([]common.Address, 0)

	gp := mp.getDag().GlobalProp
	for add := range mp.mediators {
		if _, ok := gp.ActiveMediators[add]; ok {
			ams = append(ams, add)
		}
	}

	return ams
}

func (mp *MediatorPlugin) HaveActiveMediator() bool {
	ams := mp.GetLocalActiveMediators()

	return len(ams) != 0
}

func (mp *MediatorPlugin) AddActiveMediatorPeers() {
	if !mp.HaveActiveMediator() {
		return
	}

	for _, n := range mp.ptn.GetActiveMediatorNodes() {
		mp.server.AddPeer(n)
	}
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
			if dag.DynGlobalProp.LastVerifiedUnitNum == 0 {
				newChainBanner(dag)
			}
		}

		// 调度生产unit
		go mp.scheduleProductionLoop()
	}
}

func (mp *MediatorPlugin) NewActiveMediatorsDKG() {
	ams := mp.GetLocalActiveMediators()
	initPubs := mp.getDag().GetActiveMediatorInitPubs()
	curThreshold := mp.getDag().GetCurThreshold()

	for _, med := range ams {
		initSec := mp.mediators[med].InitPartSec

		dkg, err := dkg.NewDistKeyGenerator(mp.suite, initSec, initPubs, curThreshold)
		if err != nil {
			//panic(err)
		}

		mp.dkgs[med] = dkg
	}
}

func (mp *MediatorPlugin) Start(server *p2p.Server) error {
	log.Debug("mediator plugin startup begin")
	mp.server = server

	// 1. 如果当前节点有活跃mediator，则静态连接其他活跃mediator
	go mp.AddActiveMediatorPeers()

	// 2. 开启循环生产计划
	go mp.ScheduleProductionLoop()

	// 3. 给当前节点控制的活跃mediator，初始化对应的DKG
	go mp.NewActiveMediatorsDKG()

	// 4. BLS签名循环
	go mp.unitBLSSignLoop()

	log.Debug("mediator plugin startup end")

	return nil
}

func (mp *MediatorPlugin) Stop() error {
	close(mp.quit)
	mp.newProducedUnitScope.Close()
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

		return Initialize(ptn, cfg)
	})

	if err != nil {
		log.Error(fmt.Sprintf("failed to register the Mediator Plugin service: %v", err))
	}
}

func Initialize(ptn PalletOne, cfg *Config) (*MediatorPlugin, error) {
	log.Debug("mediator plugin initialize begin")

	mss := cfg.Mediators
	msm := map[common.Address]MediatorAccount{}

	for _, medConf := range mss {
		medAcc := ConfigToAccount(medConf)
		addr := medAcc.Address
		log.Info(fmt.Sprintf("this node controll mediator account address: %v", addr.Str()))

		msm[addr] = medAcc
	}

	mp := MediatorPlugin{
		ptn:               ptn,
		productionEnabled: cfg.EnableStaleProduction,
		mediators:         msm,
		quit:              make(chan struct{}),

		toBLSSigned:     make(chan *toBLSSigned),
		pendingTBLSSign: make(map[common.Hash]*toTBLSSigned),

		suite: bn256.NewSuiteG2(),
		dkgs:  make(map[common.Address]*dkg.DistKeyGenerator),
	}

	log.Debug("mediator plugin initialize end")

	return &mp, nil
}

func (mp *MediatorPlugin) getDag() *dag.Dag {
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
