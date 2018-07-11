package core

// Genesis specifies the header fields, state of a genesis block. It also defines hard
// fork switch-over blocks through the chain configuration.
type SystemConfig struct {
	//变量名一定要大些，否则外部无法访问，导致无法进行json编码和解码
	MediatorInterval int     `json:"mediatorInterval"`
	DepositRate      float64 `json:"depositRate"`
}

type Genesis struct {
	Version                   string       `json:"version"`
	TokenAmount               uint64       `json:"tokenAmount"`
	TokenDecimal              uint32          `json:"tokenDecimal"`
	ChainID                   uint64          `json:"chainId"`
	TokenHolder               string       `json:"tokenHolder"`
	InitialActiveMediators    uint16          `json:"initialActiveMediators"`
	InitialMediatorCandidates []string     `json:"initialMediatorCandidates"`
	SystemConfig              SystemConfig `json:"systemConfig"`
}

func (g *Genesis) GetTokenAmount() uint64 {
	return g.TokenAmount
}
