package mediatorplugin

import (
	"github.com/palletone/go-palletone/common/p2p"
	"github.com/palletone/go-palletone/common/rpc"
	"github.com/palletone/go-palletone/core/node"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/core/accounts/keystore"
	"github.com/palletone/go-palletone/core/accounts"
	"github.com/palletone/go-palletone/cmd/utils"
	"strings"
	"fmt"
)

func (mp *MediatorPlugin) Protocols() []p2p.Protocol {
	return nil
}

func (mp *MediatorPlugin) APIs() []rpc.API {
	return nil
}

func (mp *MediatorPlugin) Start(server *p2p.Server) error {
	log.Info("mediator plugin startup begin")

	// 1. 判断是否满足生产验证单元的条件，主要判断本节点是否控制至少一个mediator账户
	if len(mp.Mediators) == 0 {
		println("No mediators configured! Please add mediator and private keys to configuration.")
	} else {
		// 2. 开启循环生产计划
		log.Info(fmt.Sprintf("Launching unit verify for %d mediators.", len(mp.Mediators)))

		//if mp.ProductionEnabled {
		//	if mp.DB.DynGlobalProp.LastVerifiedUnitNum == 0 {
		//		println()
		//		println("*   ------- NEW CHAIN -------   *")
		//		println("*   - Welcome to PalletOne! -   *")
		//		println("*   -------------------------   *")
		//		println()
		//	}
		//}
		//
		//mp.ScheduleProductionLoop()
	}

	log.Info("mediator plugin startup end!")

	return nil
}

func (mp *MediatorPlugin) Stop() error {
	return nil
}

// 匿名函数的好处之一：能在匿名函数内部直接使用本函数之外的变量;
// 函数使用外部变量的特性称之为闭包； 例如，以下匿名方法就直接使用cfg变量
func RegisterMediatorPluginService(stack *node.Node, cfg *Config) {
	log.Info("Register Mediator Plugin Service...")

	stack.Register(func(ctx *node.ServiceContext) (node.Service, error) {
		return Initialize(ctx, cfg)
	})
}

func Initialize(ctx *node.ServiceContext, cfg *Config) (*MediatorPlugin, error) {
	log.Info("mediator plugin initialize begin")

	ks := ctx.AccountManager.Backends(keystore.KeyStoreType)[0].(*keystore.KeyStore)

	mss := 	cfg.Mediators
	msm := map[accounts.Account]string{}
	for i := 0; i < len(mss); i++ {
		m := mss[i]

		address := strings.TrimSpace(m.Address)

		account, err := utils.MakeAddress(ks, address)
		if err != nil {
			utils.Fatalf("Could not list accounts: %v", err)
		}

		log.Info("Mediator account address ", address)

		msm[account] = m.Passphrase
	}

	mp := MediatorPlugin{
		ProductionEnabled: cfg.EnableStaleProduction,
		Mediators: msm,
	}

	log.Info("mediator plugin initialize end")

	return &mp, nil
}
