package core

// Genesis specifies the header fields, state of a genesis block. It also defines hard
// fork switch-over blocks through the chain configuration.
type SystemConfig struct {
	DepositRate float64 `json:"depositRate"`
}

type Genesis struct {
	Version                   string                   `json:"version"`
	Alias                     string                   `json:"alias"`
	TokenAmount               uint64                   `json:"tokenAmount"`
	TokenDecimal              uint32                   `json:"tokenDecimal"`
	DecimalUnit               string                   `json:"decimal_unit"`
	ChainID                   uint64                   `json:"chainId"`
	TokenHolder               string                   `json:"tokenHolder"`
	InitialParameters         ChainParameters          `json:"initialParameters"`
	ImmutableParameters       ImmutableChainParameters `json:"immutableChainParameters"`
	InitialTimestamp          int64                    `json:"initialTimestamp"`
	InitialActiveMediators    uint16                   `json:"initialActiveMediators"`
	InitialMediatorCandidates []MediatorInfo           `json:"initialMediatorCandidates"`
	SystemConfig              SystemConfig             `json:"systemConfig"`
}

func (g *Genesis) GetTokenAmount() uint64 {
	return g.TokenAmount
}

type MediatorInfo struct {
	Address,
	InitPartPub,
	Node string
}
