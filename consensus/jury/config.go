/*
	This file is part of go-palletone.
	go-palletone is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.
	go-palletone is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.
	You should have received a copy of the GNU General Public License
	along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
*/

/*
 * @author PalletOne core developers <dev@pallet.one>
 * @date 2018
 */
package jury

import (
	"gopkg.in/urfave/cli.v1"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/core"
)

const (
	DefaultContractSigNum = 1
	DefaultElectionNum    = 4
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
	ContractSigNum int            //user contract jury sig number
	ElectionNum    int            //vrf election jury number
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
	ElectionNum:    DefaultElectionNum,
	Accounts: []*AccountConf{
		&AccountConf{core.DefaultJuryAddr, DefaultPassword},
	},
}

func SetJuryConfig(ctx *cli.Context, cfg *Config) {
	switch {
	case ctx.GlobalIsSet(AccountInfoFlag.Name):
	}
}
