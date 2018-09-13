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
	"os"
	"reflect"
	"unicode"
	"path/filepath"
	"gopkg.in/urfave/cli.v1"
	"github.com/naoina/toml"
	"github.com/palletone/go-palletone/cmd/utils"
	"github.com/palletone/go-palletone/adaptor"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/p2p"
	"github.com/palletone/go-palletone/configure"
	"github.com/palletone/go-palletone/core/node"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/ptn"
	"github.com/palletone/go-palletone/ptnjson"
	"github.com/palletone/go-palletone/statistics/dashboard"
	"github.com/palletone/go-palletone/contracts"
	mp "github.com/palletone/go-palletone/consensus/mediatorplugin"
)

const defaultConfigPath = "palletone.toml"

var (
	dumpConfigCommand = cli.Command{
		Action:    utils.MigrateFlags(dumpConfig),
		Name:      "dumpconfig",
		Usage:     "Dumps configuration to a specified file",
		ArgsUsage: "<configFilePath>",
		//		Flags:       append(append(nodeFlags, rpcFlags...)),
		Flags: []cli.Flag{
			ConfigFileFlag,
		},
		Category:    "MISCELLANEOUS COMMANDS",
		Description: `The dumpconfig command dumps configuration to a specified file.`,
	}

	ConfigFileFlag = cli.StringFlag{
		Name:  "configfile",
		Usage: "TOML configuration file",
		Value: defaultConfigPath,
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

// SignRawTransactionCmd defines the signrawtransaction JSON-RPC command.
type SignRawTransactionCmd struct {
	RawTx    string
	Inputs   *[]ptnjson.RawTxInput
	PrivKeys *[]string
	Flags    *string `jsonrpcdefault:"\"ALL\""`
}

const (
	NETID_MAIN = iota
	NETID_TEST
)

type ptnstatsConfig struct {
	URL string `toml:",omitempty"`
}

type FullConfig struct {
	Ptn            ptn.Config
	Node           node.Config
	Ptnstats       ptnstatsConfig
	Dashboard      dashboard.Config
	MediatorPlugin mp.Config
	Log            *log.Config
	Dag            *dagconfig.Config
	P2P            p2p.Config
	Ada            adaptor.Config
	Contract       contracts.Config
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
	cfg.HTTPModules = append(cfg.HTTPModules, "ptn" /*, "shh"*/)
	cfg.WSModules = append(cfg.WSModules, "ptn" /*, "shh"*/)
	cfg.IPCPath = "gptn.ipc"
	return cfg
}

func adaptorConfig(config FullConfig) FullConfig {
	//config.Node.P2P = config.P2P
	config.Ptn.Dag = *config.Dag
	config.Ptn.Log = *config.Log
	config.Ptn.MediatorPlugin = config.MediatorPlugin

	return config
}

// 根据指定路径和配置参数获取配置文件的路径
// @author Albert·Gou
func getConfigPath(configPath, dataDir string) (string, error) {
	if filepath.IsAbs(configPath) {
		return configPath, nil
	}

	if dataDir != "" && configPath == "" {
		return filepath.Join(dataDir, defaultConfigPath), nil
	}

	if configPath != "" {
		return filepath.Abs(configPath)
	}

	return "", nil
}

// 加载指定的或者默认的配置文件，如果不存在则根据默认的配置生成文件
// @author Albert·Gou
func maybeLoadConfig(ctx *cli.Context, cfg *FullConfig) error {
	// 获取配置文件路径: 命令行指定的路径 或者默认的路径
	configPath := defaultConfigPath
	if temp := ctx.GlobalString(ConfigFileFlag.Name); temp != "" {
		configPath, _ = getConfigPath(temp, cfg.Node.DataDir)
	}

	// 如果配置文件不存在，则使用默认的配置生成一个配置文件
	if _, err := os.Stat(configPath); err != nil && os.IsNotExist(err) {
		defaultConfig := makeDefaultConfig()
		err = makeConfigFile(&defaultConfig, configPath)
		if err != nil {
			utils.Fatalf("%v", err)
			return err
		}

		log.Debug(fmt.Sprintf("Writing new config file at: %v", configPath))
	}

	// 加载配置文件中的配置信息到 cfg中
	if err := loadConfig(configPath, cfg); err != nil {
		utils.Fatalf("%v", err)
		return err
	}

	return nil
}

func makeConfigNode(ctx *cli.Context) (*node.Node, FullConfig) {
	// Load defaults.
	// 1. cfg加载系统默认的配置信息，cfg是一个字典结构
	cfg := makeDefaultConfig()

	// Load config file.
	// 2. 获取配置文件中的配置信息，并覆盖cfg中对应的配置
	if err := maybeLoadConfig(ctx, &cfg); err != nil {
		utils.Fatalf("%v", err)
	}

	// Apply flags.
	// 3. 将命令行中的配置参数覆盖cfg中对应的配置,
	// 先处理node的配置信息，再创建node，然后再处理其他service的配置信息，因为其他service的配置依赖node中的协议
	// 注意：不是将命令行中所有的配置都覆盖cfg中对应的配置，例如 Ptnstats 配置目前没有覆盖 (可能通过命令行设置)
	utils.SetNodeConfig(ctx, &cfg.Node)
	//cfg = adaptorConfig(cfg)
	cfg.Node.P2P = cfg.P2P
	// 4. 通过Node的配置来创建一个Node, 变量名叫stack，代表协议栈的含义。
	stack, err := node.New(&cfg.Node)
	if err != nil {
		utils.Fatalf("Failed to create the protocol stack: %v", err)
	}
	cfg = adaptorConfig(cfg)
	utils.SetPtnConfig(ctx, stack, &cfg.Ptn)
	//fmt.Println("cfg.Ptn.Log.OpenModule", cfg.Ptn.Log.OpenModule)
	cfg.Log.OpenModule = cfg.Ptn.Log.OpenModule

	if ctx.GlobalIsSet(utils.EthStatsURLFlag.Name) {
		cfg.Ptnstats.URL = ctx.GlobalString(utils.EthStatsURLFlag.Name)
	}
	utils.SetDashboardConfig(ctx, &cfg.Dashboard)
	mp.SetMediatorPluginConfig(ctx, &cfg.MediatorPlugin)

	return stack, cfg
}

// makeFullNode 函数用创建一个 PalletOne 节点，节点类型根据ctx参数传递的命令行指令来控制。
// 生成node.Node一个结构，里面会有任务函数栈, 然后设置各个服务到serviceFuncs 里面，
// 包括：全节点，dashboard，以及状态stats服务等
func makeFullNode(ctx *cli.Context) *node.Node {
	//	if ctx.String("log.path") != "stdout" {
	//		log.FileInitLogger(ctx.String("log.path"))
	//	}else{
	//		log.InitLogger()
	//	}
	// 1. 根据默认配置、命令行参数和配置文件的配置来创建一个node, 并获取相关配置
	stack, cfg := makeConfigNode(ctx)
	log.InitLogger()
	//	if ctx.String("log.path") != "stdout" {
	//		log.FileInitLogger(ctx.String("log.path"))
	//	}
	//stack.SetDbPath(cfg.Dag.DbPath)

	// 2. 创建 Ptn service、DashBoard service以及 PtnStats service 等 service ,
	// 并注册到 Node 的 serviceFuncs 中，然后在 stack.Start() 执行的时候会调用这些 service
	// 所有的 service 必须实现 node.Service 接口中定义的所有方法
	utils.RegisterPtnService(stack, &cfg.Ptn)
	if ctx.GlobalBool(utils.DashboardEnabledFlag.Name) {
		//注册dashboard仪表盘服务，Dashboard会开启端口监听
		utils.RegisterDashboardService(stack, &cfg.Dashboard, gitCommit)
	}

	// Add the PalletOne Stats daemon if requested.
	if cfg.Ptnstats.URL != "" {
		// 注册状态服务。 默认情况下是没有启动的。
		utils.RegisterPtnStatsService(stack, cfg.Ptnstats.URL)
	}

	// rebuild leveldb
	// if cfg.Dag.DbPath != "" {
	// 	dagconfig.De
	// }

	// comment by Albert·Gou
	//	mp.RegisterMediatorPluginService(stack, &cfg.MediatorPlugin)

	return stack
}

// dumpConfig is the dumpconfig command.
// modify by Albert·Gou
func dumpConfig(ctx *cli.Context) error {
	cfg := makeDefaultConfig()
	//	comment := ""

	//if cfg.Ptn.Genesis != nil {
	//	cfg.Ptn.Genesis = nil
	//	comment += "# Note: this config doesn't contain the genesis block.\n\n"
	//}

	configPath := ctx.Args().First()
	// If no path is specified, the default path is used
	if len(configPath) == 0 {
		configPath, _ = getConfigPath(defaultConfigPath, cfg.Node.DataDir)
	}

	//	io.WriteString(os.Stdout, comment)

	err := makeConfigFile(&cfg, configPath)
	if err != nil {
		utils.Fatalf("%v", err)
		return err
	}

	fmt.Println("Dumping new config file at " + configPath)

	return nil
}

// makeDefaultConfig, create a default config
// @author Albert·Gou
func makeDefaultConfig() FullConfig {
	// 不是所有的配置都有默认值，例如 Ptnstats 目前没有设置默认值
	return FullConfig{
		Ptn:            ptn.DefaultConfig,
		Node:           defaultNodeConfig(),
		Dashboard:      dashboard.DefaultConfig,
		P2P:            p2p.DefaultConfig,
		MediatorPlugin: mp.DefaultConfig,
		Dag:            &dagconfig.DefaultConfig,
		Log:            &log.DefaultConfig,
		Ada:            adaptor.DefaultConfig,
		Contract:       contracts.DefaultConfig,
	}
}

// Create a config file with the specified path and config info
// @author Albert·Gou
func makeConfigFile(cfg *FullConfig, configPath string) error {
	var (
		configFile *os.File = nil
		err        error    = nil
	)

	err = os.MkdirAll(filepath.Dir(configPath), os.ModePerm)
	if err != nil {
		utils.Fatalf("%v", err)
		return err
	}

	configFile, err = os.Create(configPath)
	defer configFile.Close()
	if err != nil {
		utils.Fatalf("%v", err)
		return err
	}

	configToml, err := tomlSettings.Marshal(cfg)
	if err != nil {
		log.Error(err.Error())
		return err
	}

	_, err = configFile.Write(configToml)
	if err != nil {
		utils.Fatalf("%v", err)
		return err
	}

	return nil
}
