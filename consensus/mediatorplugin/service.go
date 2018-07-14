package mediatorplugin

import (
	"github.com/palletone/go-palletone/common/p2p"
	"github.com/palletone/go-palletone/common/rpc"
	"github.com/palletone/go-palletone/core/node"
)

func (mp *MediatorPlugin) Protocols() []p2p.Protocol {
	return nil
}

func (mp *MediatorPlugin) APIs() []rpc.API {
	return nil
}

func (mp *MediatorPlugin) Start(server *p2p.Server) error {
	return nil
}

func (mp *MediatorPlugin) Stop() error {
	return nil
}

// 匿名函数的好处之一：能在匿名函数内部直接使用本函数之外的变量，
// 例如，以下匿名方法就直接使用cfg变量
func RegisterMediatorPluginService(stack *node.Node, cfg *Config) {
	stack.Register(func(ctx *node.ServiceContext) (node.Service, error) {
		return New(ctx, cfg)
	})
}

func New(ctx *node.ServiceContext, config *Config) (*MediatorPlugin, error) {
	return nil, nil
}
