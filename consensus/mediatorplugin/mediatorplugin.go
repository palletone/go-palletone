package mediatorplugin

import (
	"github.com/palletone/go-palletone/core/node"
	"github.com/palletone/go-palletone/common"
)

type MediatorPlugin struct {
	node *node.Node
	// Enable VerifiedUnit production, even if the chain is stale.
	// 新开启一个区块链时，必须设为true
	productionEnabled bool
	// Mediator`s account and passphrase controlled by this node
	mediators map[common.Address]string
}
