// Copyright 2017 The go-ethereum Authors
// This file is part of go-ethereum.
//
// go-ethereum is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-ethereum is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with go-ethereum. If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"unicode"

	"github.com/naoina/toml"
	"github.com/palletone/go-palletone/cmd/utils"
	cli "gopkg.in/urfave/cli.v1"

	"github.com/palletone/go-palletone/adaptor"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/p2p"
	"github.com/palletone/go-palletone/common/p2p/nat"
	"github.com/palletone/go-palletone/configure"
	"github.com/palletone/go-palletone/consensus/consensusconfig"
	"github.com/palletone/go-palletone/core/node"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/ptn"
	"github.com/palletone/go-palletone/statistics/dashboard"
)

var (
	dumpConfigCommand = cli.Command{
		Action:      utils.MigrateFlags(dumpConfig),
		Name:        "dumpconfig",
		Usage:       "Show configuration values",
		ArgsUsage:   "",
		Flags:       append(append(nodeFlags, rpcFlags...)),
		Category:    "MISCELLANEOUS COMMANDS",
		Description: `The dumpconfig command shows configuration values.`,
	}

	configFileFlag = cli.StringFlag{
		Name:  "config",
		Usage: "TOML configuration file",
	}
)

// These settings ensure that TOML keys use the same names as Go struct fields.
var tomlSettings = toml.Config{
	NormFieldName: func(rt reflect.Type, key string) string {
		return key
	},
	FieldToKey: func(rt reflect.Type, field string) string {
		return field
	},
	MissingField: func(rt reflect.Type, field string) error {
		link := ""
		if unicode.IsUpper(rune(rt.Name()[0])) && rt.PkgPath() != "main" {
			link = fmt.Sprintf(", see https://godoc.org/%s#%s for available fields", rt.PkgPath(), rt.Name())
		}
		return fmt.Errorf("field '%s' is not defined in %s%s", field, rt.String(), link)
	},
}

type ethstatsConfig struct {
	URL string `toml:",omitempty"`
}

type FullConfig struct {
	Ptn       ptn.Config
	Node      node.Config
	Ethstats  ethstatsConfig
	Dashboard dashboard.Config
	Consensus consensusconfig.Config
	Log       *log.Config
	Dag       dagconfig.Config
	P2P       p2p.Config
	Ada       adaptor.Config
}

func loadConfig(file string, cfg *FullConfig) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	err = tomlSettings.NewDecoder(bufio.NewReader(f)).Decode(cfg)
	// Add file name to errors that have a line number.
	if _, ok := err.(*toml.LineError); ok {
		err = errors.New(file + ", " + err.Error())
	}
	return err
}

func defaultNodeConfig() node.Config {
	cfg := node.DefaultConfig
	cfg.Name = clientIdentifier
	cfg.Version = configure.VersionWithCommit(gitCommit)
	cfg.HTTPModules = append(cfg.HTTPModules, "eth" /*, "shh"*/)
	cfg.WSModules = append(cfg.WSModules, "eth" /*, "shh"*/)
	cfg.IPCPath = "gptn.ipc"
	return cfg
}

func adaptorConfig(config FullConfig) FullConfig {
	config.Node.P2P = config.P2P
	config.Ptn.Dag = config.Dag
	config.Ptn.Log = *config.Log
	config.Ptn.Consensus = config.Consensus
	return config
}

func makeConfigNode(ctx *cli.Context) (*node.Node, FullConfig) {
	// Load defaults.
	cfg := FullConfig{
		Ptn:       ptn.DefaultConfig,
		Node:      defaultNodeConfig(),
		Dashboard: dashboard.DefaultConfig,
		P2P:       p2p.Config{ListenAddr: ":30303", MaxPeers: 25, NAT: nat.Any()},
		Consensus: consensusconfig.DefaultConfig,
		Dag:       dagconfig.DefaultConfig,
		Log:       &log.DefaultConfig,
		Ada:       adaptor.DefaultConfig,
	}

	// Load config file.
	file := "./palletone.toml"
	if temp := ctx.GlobalString(configFileFlag.Name); temp != "" {
		file = temp
	}
	if err := loadConfig(file, &cfg); err != nil {
		utils.Fatalf("%v", err)
	}

	cfg = adaptorConfig(cfg)
	// Apply flags.
	//utils.SetNodeConfig(ctx, &cfg.Node)
	stack, err := node.New(&cfg.Node)
	if err != nil {
		utils.Fatalf("Failed to create the protocol stack: %v", err)
	}

	utils.SetEthConfig(ctx, stack, &cfg.Ptn)
	if ctx.GlobalIsSet(utils.EthStatsURLFlag.Name) {
		cfg.Ethstats.URL = ctx.GlobalString(utils.EthStatsURLFlag.Name)
	}
	utils.SetDashboardConfig(ctx, &cfg.Dashboard)
	return stack, cfg
}

func makeFullNode(ctx *cli.Context) *node.Node {
	stack, cfg := makeConfigNode(ctx)
	log.InitLogger()
	utils.RegisterEthService(stack, &cfg.Ptn)
	if ctx.GlobalBool(utils.DashboardEnabledFlag.Name) {
		utils.RegisterDashboardService(stack, &cfg.Dashboard, gitCommit)
	}

	// Add the PalletOne Stats daemon if requested.
	if cfg.Ethstats.URL != "" {
		utils.RegisterEthStatsService(stack, cfg.Ethstats.URL)
	}
	return stack
}

// dumpConfig is the dumpconfig command.
func dumpConfig(ctx *cli.Context) error {
	_, cfg := makeConfigNode(ctx)
	comment := ""

	if cfg.Ptn.Genesis != nil {
		cfg.Ptn.Genesis = nil
		comment += "# Note: this config doesn't contain the genesis block.\n\n"
	}

	out, err := tomlSettings.Marshal(&cfg)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	io.WriteString(os.Stdout, comment)
	os.Stdout.Write(out)
	return nil
}
