package jury

import (
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/core"
	"gopkg.in/urfave/cli.v1"
)

const (
	DefaultContractSigNum = 3
	DefaultPassword       = "password"
	DefaultInitPartSec    = "47gsj9pK3pwYUS1ZrWQjTgWMHUXWdNuCr7hXPXHySyBk"
)

var (
	AccountInfoFlag = cli.BoolFlag{
		Name:  "accountInfo",
		Usage: "The About information account address,password,public and private key and on ",
	}
)

type AccountConf struct {
	Address,
	Password string
}
type JuryAccount struct {
	Address  common.Address
	Password string
}
type Config struct {
	ContractSigNum int
	Accounts       []*AccountConf // the set of the mediator info
}

func (aConf *AccountConf) configToAccount() *JuryAccount {
	addr, _ := common.StringToAddress(aConf.Address)

	medAcc := &JuryAccount{
		addr,
		aConf.Password,
	}
	return medAcc
}

var DefaultConfig = Config{
	ContractSigNum: DefaultContractSigNum,
	Accounts: []*AccountConf{
		&AccountConf{core.DefaultJuryAddr, DefaultPassword},
	},
}

func SetJuryConfig(ctx *cli.Context, cfg *Config) {
	switch {
	case ctx.GlobalIsSet(AccountInfoFlag.Name):
	}
}
