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
	"github.com/palletone/go-palletone/core"
	"gopkg.in/urfave/cli.v1"
)

const (
	defaultPassword    = "password"
	DefaultInitPartSec = "Vh52_xy-bE5U2mUDFYxLJwRke2IQ7u0Nb9L3_cPyXKY"
	DefaultInitPartPub = "AV0a95Ex-pTGAAYXg277329ewkWDOOdkuo-Va1ogVnEQiCc-efOZlFMFWCYDhld8uxoRggzxRJJzPJ" +
		"0r4hKKVKRZOW-aLJYuGilc5ONNi3riQBCoOry7cX65yLx9_yMENAnWI_fN6USJpDG2dJBTCyuY-N_GOZf9wD-2qhP5-lDj"
)

var (
	StaleProductionFlag = cli.BoolFlag{
		Name:  "enable-stale-production",
		Usage: "Enable Verified Unit production, even if the chain is stale.",
	}
)

// config data for mediator plugin
type Config struct {
	EnableStaleProduction bool // Enable Verified Unit production, even if the chain is stale.
	//	RequiredParticipation float32	// Percent of mediators (0-99) that must be participating in order to produce
	Mediators []MediatorInfo // the set of the mediator info
}

// mediator plugin default config
var DefaultConfig = Config{
	EnableStaleProduction: false,
	Mediators: []MediatorInfo{
		MediatorInfo{core.DefaultTokenHolder, defaultPassword,
			DefaultInitPartSec, DefaultInitPartPub},
	},
}

func SetMediatorPluginConfig(ctx *cli.Context, cfg *Config) {
	switch {
	case ctx.GlobalIsSet(StaleProductionFlag.Name):
		cfg.EnableStaleProduction = ctx.GlobalBool(StaleProductionFlag.Name)
	}
}

type MediatorInfo struct {
	Address,
	Password,
	InitPartSec,
	InitPartPub string
}
