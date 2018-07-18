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

package mediatorplugin

import "gopkg.in/urfave/cli.v1"

var (
	StaleProductionFlag = cli.BoolFlag{
		Name:  "enable-stale-production",
		Usage: "Enable Verified Unit production, even if the chain is stale.",
	}
)

// mediator 结构体 和具体的账户模型有关
type Mediator struct {
	Address    string
	Passphrase string
}

// config data for mediator plugin
type Config struct {
	EnableStaleProduction bool // Enable Verified Unit production, even if the chain is stale.
	//	RequiredParticipation float32	// Percent of mediators (0-99) that must be participating in order to produce
	Mediators []Mediator
}

// mediator plugin default config
var DefaultConfig = Config{
	EnableStaleProduction: false,
		Mediators: []Mediator{
			{"P1XXX","123"},
		},
}

func SetMediatorPluginConfig(ctx *cli.Context, cfg *Config) {
	switch {
	case ctx.GlobalIsSet(StaleProductionFlag.Name):
		cfg.EnableStaleProduction = ctx.GlobalBool(StaleProductionFlag.Name)
		//case :
		//
	}
}
