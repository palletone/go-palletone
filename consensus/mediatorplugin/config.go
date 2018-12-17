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
 * @author PalletOne core developer Albert·Gou <dev@pallet.one>
 * @date 2018
 */

package mediatorplugin

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
	StaleProductionFlag = cli.BoolFlag{
		Name:  "enable-stale-production",
		Usage: "Enable Unit production, even if the chain is stale.",
	}
)

// config data for mediator plugin
type Config struct {
	EnableStaleProduction bool // Enable Unit production, even if the chain is stale.
	//	RequiredParticipation float32	// Percent of mediators (0-99) that must be participating in order to produce
	Mediators []*MediatorConf // the set of the mediator info
}

// mediator plugin default config
var DefaultConfig = Config{
	EnableStaleProduction: false,
	Mediators: []*MediatorConf{
		&MediatorConf{core.DefaultMediator, DefaultPassword,
			DefaultInitPartSec, core.DefaultInitPartPub},
	},
}

func SetMediatorPluginConfig(ctx *cli.Context, cfg *Config) {
	switch {
	case ctx.GlobalIsSet(StaleProductionFlag.Name):
		cfg.EnableStaleProduction = ctx.GlobalBool(StaleProductionFlag.Name)
	}
}

type MediatorConf struct {
	Address,
	Password,
	InitPartSec,
	InitPartPub string
}

func (medConf *MediatorConf) configToAccount() *MediatorAccount {
	// 1. 解析 mediator 账户地址
	addr := core.StrToMedAdd(medConf.Address)

	// 2. 解析 mediator 的 DKS 初始公私钥
	sec := core.StrToScalar(medConf.InitPartSec)
	pub := core.StrToPoint(medConf.InitPartPub)

	medAcc := &MediatorAccount{
		addr,
		medConf.Password,
		sec,
		pub,
	}

	return medAcc
}

type MediatorAccount struct {
	Address     common.Address
	Password    string
	InitPartSec kyber.Scalar
	InitPartPub kyber.Point
}
