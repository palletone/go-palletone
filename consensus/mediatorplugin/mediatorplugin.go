package mediatorplugin

import (
	"github.com/palletone/go-palletone/core/accounts"
)

type MediatorPlugin struct {
	// Enable VerifiedUnit production, even if the chain is stale.
	// 新开启一个区块链时，必须设为true
	ProductionEnabled bool
	// Mediator`s account and passphrase controlled by this node
	Mediators map[accounts.Account]string
}
