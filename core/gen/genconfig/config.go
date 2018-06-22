package genconfig

type Config struct {
	// token total, by default 1 billion
	TotalSupply uint64
	// allocate all token to this address
	HolderAddress string
}

var DefaultConfig Config = Config{
	TotalSupply:   10000000000,
	HolderAddress: "",
}
