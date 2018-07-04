package core

// Genesis specifies the header fields, state of a genesis block. It also defines hard
// fork switch-over blocks through the chain configuration.
type SystemConfig struct {
	MediatorSlot  int      `json:"mediatorSlot"`
	MediatorCount int      `json:"mediatorCount"`
	MediatorList  []string `json:"mediatorList"`
	MediatorCycle int      `json:"mediatorCycle"`
	DepositRate   float64  `json:"depositRate"`
}

type Genesis struct {
	Height       string       `json:"height"`
	Version      string       `json:"version"`
	TokenAmount  uint64       `json:"tokenAmount"`
	TokenDecimal int          `json:"tokenDecimal"`
	ChainID      int          `json:"chainId"`
	TokenHolder  string       `json:"tokenHolder"`
	SystemConfig SystemConfig `json:"systemConfig"`
}


func (g *Genesis) GetTokenAmount() uint64 {
	return g.TokenAmount
}
