package mediatorplugin

type MediatorPlugin struct {
	// Enable VerifiedUnit production, even if the chain is stale.
	// 新开启一个区块链时，必须设为true
	ProductionEnabled bool
	// Mediator`s address and passphrase controlled by this node
	Mediators map[string]string
}
