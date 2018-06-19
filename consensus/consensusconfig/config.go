package consensusconfig

//config data for consensus
type Config struct {
	Engine string //solo or dpos
}

// Consensus default
var DefaultConfig = Config{
	Engine: "solo",
}
