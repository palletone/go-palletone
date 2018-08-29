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
	"gopkg.in/urfave/cli.v1"
	"github.com/palletone/go-palletone/core"
)

const (
	//defaultMediator = "P113o9hNFyPv7ys5TxpCcaJDYbZ4b7hsPyt"
	defaultPassword = "password"
	DefaultInitPartSec = "Tj1wNXLM6YkAxnvmag5ydh8qbBJ9wEtZDMeTE0qfacI="
	DefaultInitPartPub = "AXkVypTVHRI94Vd7ggAdN62rLW57tcJrTw/0nErnXfFlcQpaSA1qMaYiJI7HxKJBhxWmTzRloEfB" +
		"93mSTDuqs+40LJFNJ/h+gNlZxTHiWEw9MuFhNMC0j5CzfGFvUkDcs0XTIYLGhMomImqw9n5cURXOm0Qhhr92gdO8kYZheIq/"
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
	Mediators map[string]MediatorInfo // the map of  mediator`s address and  the mediator info
}

// mediator plugin default config
var DefaultConfig = Config{
	EnableStaleProduction: false,
	Mediators:             map[string]MediatorInfo{
		core.DefaultTokenHolder:
			MediatorInfo{defaultPassword,DefaultInitPartSec, DefaultInitPartPub},
		//defaultMediator:
		//	MediatorInfo{defaultPassword,DefaultInitPartSec, DefaultInitPartPub},
	},
}

func SetMediatorPluginConfig(ctx *cli.Context, cfg *Config) {
	switch {
	case ctx.GlobalIsSet(StaleProductionFlag.Name):
		cfg.EnableStaleProduction = ctx.GlobalBool(StaleProductionFlag.Name)
	}
}

type MediatorInfo struct {
	Password string
	InitPartSec string
	InitPartPub string
}
