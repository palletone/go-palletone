package jury

import (
	"github.com/dedis/kyber"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/core"
	"gopkg.in/urfave/cli.v1"
)

const (
	DefaultPassword    = "password"
	DefaultInitPartSec = "47gsj9pK3pwYUS1ZrWQjTgWMHUXWdNuCr7hXPXHySyBk"
)

var (
	AccountInfoFlag = cli.BoolFlag{
		Name:  "accountInfo",
		Usage: "The About information account address,password,public and private key and on ",
	}
)

type AccountConf struct {
	Address,
	Password,
	InitPartSec,
	InitPartPub string
}

type JuryAccount struct {
	Address     common.Address
	Password    string
	InitPartSec kyber.Scalar
	InitPartPub kyber.Point
}

type Config struct {
	Accounts []*AccountConf // the set of the mediator info
}

func (aConf *AccountConf) configToAccount() *JuryAccount {
	addr := core.StrToMedAdd(aConf.Address)
	sec, _ := core.StrToScalar(aConf.InitPartSec)
	pub, _ := core.StrToPoint(aConf.InitPartPub)

	medAcc := &JuryAccount{
		addr,
		aConf.Password,
		sec,
		pub,
	}
	return medAcc
}

var DefaultConfig = Config{
	Accounts: []*AccountConf{
		&AccountConf{core.DefaultJuryAddr, DefaultPassword,
			DefaultInitPartSec, core.DefaultJuryInitPartPub},
	},
}

func SetJuryConfig(ctx *cli.Context, cfg *Config) {
	switch {
	case ctx.GlobalIsSet(AccountInfoFlag.Name):
	}
}
