package mediatorplugin

import (
	"github.com/palletone/go-palletone/common/p2p"
	"github.com/palletone/go-palletone/common/rpc"
	"github.com/palletone/go-palletone/core/node"
	"github.com/palletone/go-palletone/common/log"
)

func (mp *MediatorPlugin) Protocols() []p2p.Protocol {
	return nil
}

func (mp *MediatorPlugin) APIs() []rpc.API {
	return nil
}

func (mp *MediatorPlugin) Start(server *p2p.Server) error {
	log.Info("mediator plugin startup begin")



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

	mss := 	cfg.Mediators
	msm := map[string]string{}
	for i := 0; i < len(mss); i++ {
		m := mss[i]
		msm[m.Address] = m.Passphrase
	}

	mp := MediatorPlugin{
		ProductionEnabled: cfg.EnableStaleProduction,
		Mediators: msm,
	}

	log.Info("mediator plugin initialize end")

	return &mp, nil
}
