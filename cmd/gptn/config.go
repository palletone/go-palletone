// Copyright 2018 PalletOne
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
	"path/filepath"
	"reflect"
	"unicode"

	"bytes"
	"github.com/naoina/toml"
	"github.com/palletone/go-palletone/adaptor"
	"github.com/palletone/go-palletone/cmd/utils"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/crypto"
	"github.com/palletone/go-palletone/common/files"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/p2p"
	"github.com/palletone/go-palletone/configure"
	"github.com/palletone/go-palletone/consensus/jury"
	mp "github.com/palletone/go-palletone/consensus/mediatorplugin"
	"github.com/palletone/go-palletone/contracts/contractcfg"
	"github.com/palletone/go-palletone/core/certficate"
	"github.com/palletone/go-palletone/core/node"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/txspool"
	"github.com/palletone/go-palletone/ptn"
	//"github.com/palletone/go-palletone/ptnjson"
	"github.com/coocood/freecache"
	"github.com/palletone/go-palletone/statistics/dashboard"
	"gopkg.in/urfave/cli.v1"
)

const defaultConfigPath = "./ptn-config.toml"

var (
	dumpConfigCommand = cli.Command{
		Action:    utils.MigrateFlags(dumpConfig),
		Name:      "dumpconfig",
		Usage:     "Dumps configuration to a specified file",
		ArgsUsage: "<configFilePath>",
		Flags: []cli.Flag{
			ConfigFilePathFlag,
		},
		Category:    "MISCELLANEOUS COMMANDS",
		Description: `The dumpconfig command dumps configuration to a specified file.`,
	}

	dumpUserCfgCommand = cli.Command{
		Action:    utils.MigrateFlags(dumpUserCfg),
		Name:      "dumpuserconfig",
		Usage:     "Dumps configuration to a specified file",
		ArgsUsage: "<configFilePath>",
		Flags: []cli.Flag{
			ConfigFilePathFlag,
		},
		Category:    "MISCELLANEOUS COMMANDS",
		Description: `The dumpconfig command dumps configuration to a specified file.`,
	}

	ConfigFilePathFlag = cli.StringFlag{
		Name:  "configfile",
		Usage: "TOML configuration file",
		Value: "", //defaultConfigPath,
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
//type SignRawTransactionCmd struct {
//	RawTx    string
//	Inputs   *[]ptnjson.RawTxInput
//	PrivKeys *[]string
//	Flags    *string `jsonrpcdefault:"\"ALL\""`
//}

//const (
//	NETID_MAIN = iota
//	NETID_TEST
//)

type ptnstatsConfig struct {
	URL string `toml:",omitempty"`
}

type FullConfig struct {
	Ptn            ptn.Config
	TxPool         txspool.TxPoolConfig
	Node           node.Config
	Ptnstats       ptnstatsConfig
	Dashboard      dashboard.Config
	Jury           jury.Config
	MediatorPlugin mp.Config
	Log            log.Config // log的配置比较特殊，不属于任何模块，顶级配置，程序开始运行就使用
	Dag            dagconfig.Config
	P2P            p2p.Config
	Ada            adaptor.Config
	Contract       contractcfg.Config
	Certficate     certficate.CAConfig
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
	cfg.P2P = p2p.DefaultConfig
	cfg.Name = clientIdentifier
	cfg.Version = configure.VersionWithCommit(gitCommit)
	cfg.HTTPModules = append(cfg.HTTPModules, "ptn" /*, "shh"*/)
	cfg.WSModules = append(cfg.WSModules, "ptn" /*, "shh"*/)
	cfg.IPCPath = "gptn.ipc"
	cfg.WSExposeAll = false
	return cfg
}

func adaptorNodeConfig(config *FullConfig) *FullConfig {
	config.Node.P2P = config.P2P
	return config
}

func adaptorPtnConfig(config *FullConfig) *FullConfig {
	config.Ptn.TxPool = config.TxPool
	config.Ptn.Dag = config.Dag
	config.Ptn.Jury = config.Jury
	config.Ptn.MediatorPlugin = config.MediatorPlugin
	config.Ptn.Contract = config.Contract

	return config
}

// 根据指定路径和配置参数获取配置文件的路径
// @author Albert·Gou
func getConfigPath(ctx *cli.Context) string {
	// 获取配置文件路径: 命令行指定的路径 或者默认的路径
	configPath := defaultConfigPath
	if temp := ctx.GlobalString(ConfigFilePathFlag.Name); temp != "" {
		if files.IsDir(temp) {
			temp = filepath.Join(temp, filepath.Base(defaultConfigPath))
		}
		configPath = temp
	}

	return common.GetAbsPath(configPath)
}

// 加载指定的或者默认的配置文件，如果不存在则根据默认的配置生成文件
// @author Albert·Gou
func maybeLoadConfig(ctx *cli.Context) (FullConfig, string, error) {
	// 1. cfg加载系统默认的配置信息，cfg是一个字典结构
	configPath := getConfigPath(ctx)
	cfg := DefaultConfig()

	// 如果配置文件不存在，则使用默认的配置生成一个配置文件
	if !common.IsExisted(configPath) {
		//listenAddr := cfg.P2P.ListenAddr
		//if strings.HasPrefix(listenAddr, ":") {
		//	cfg.P2P.ListenAddr = "127.0.0.1" + listenAddr
		//}

		err := makeConfigFile(&cfg, configPath)
		if err != nil {
			utils.Fatalf("%v", err)
			return FullConfig{}, "", err
		}

		fmt.Println("Writing new config file at: ", configPath)
	}

	// 加载配置文件中的配置信息到 cfg中
	if err := loadConfig(configPath, &cfg); err != nil {
		utils.Fatalf("%v", err)
		return FullConfig{}, "", err
	}

	return cfg, filepath.Dir(configPath), nil
}

func makeConfigNode(ctx *cli.Context, isInConsole bool) (*node.Node, FullConfig) {
	// Load config file.
	cfg, configDir, err := maybeLoadConfig(ctx)
	if err != nil {
		utils.Fatalf("%v", err)
	}

	// Apply flags.
	// 将命令行中的配置参数覆盖cfg中对应的配置,
	// 先处理node的配置信息，再创建node，然后再处理其他service的配置信息，因为其他service的配置依赖node中的协议
	// 注意：不是将命令行中所有的配置都覆盖cfg中对应的配置，例如 Ptnstats 配置目前没有覆盖 (可能通过命令行设置)

	// log的配置比较特殊，不属于任何模块，顶级配置，程序开始运行就使用
	utils.SetLogConfig(ctx, &cfg.Log, configDir, isInConsole)
	utils.SetP2PConfig(ctx, &cfg.P2P)
	adaptorNodeConfig(&cfg)

	dataDir := utils.SetNodeConfig(ctx, &cfg.Node, configDir)
	//通过Node的配置来创建一个Node, 变量名叫stack，代表协议栈的含义。
	stack, err := node.New(&cfg.Node)
	if err != nil {
		utils.Fatalf("Failed to create the protocol stack: %v", err)
	}

	//utils.SetContractConfig(ctx, &cfg.Contract, dataDir)
	utils.SetTxPoolConfig(ctx, &cfg.TxPool)
	utils.SetDagConfig(ctx, &cfg.Dag, dataDir)
	mp.SetMediatorConfig(ctx, &cfg.MediatorPlugin)
	jury.SetJuryConfig(ctx, &cfg.Jury)

	// 为了方便用户配置，所以将各个子模块的配置提升到与ptn同级，
	// 然而在RegisterPtnService()中，只能使用ptn下的配置
	// 所以在此处将各个子模块的配置，复制到ptn下
	adaptorPtnConfig(&cfg)

	utils.SetPtnConfig(ctx, stack, &cfg.Ptn)
	if bytes.Equal(cfg.Ptn.CryptoLib, []byte{1, 1}) {
		fmt.Println("Use GM crypto lib")
		crypto.MyCryptoLib = &crypto.CryptoGm{}
	}
	if ctx.GlobalIsSet(utils.EthStatsURLFlag.Name) {
		cfg.Ptnstats.URL = ctx.GlobalString(utils.EthStatsURLFlag.Name)
	}
	utils.SetDashboardConfig(ctx, &cfg.Dashboard)
	//  init node.cache
	stack.CacheDb = freecache.NewCache(cfg.Dag.DbCache)
	return stack, cfg
}

// makeFullNode 函数用创建一个 PalletOne 节点，节点类型根据ctx参数传递的命令行指令来控制。
// 生成node.Node一个结构，里面会有任务函数栈, 然后设置各个服务到serviceFuncs 里面，
// 包括：全节点，dashboard，以及状态stats服务等
func makeFullNode(ctx *cli.Context, isInConsole bool) *node.Node {
	// 1. 根据默认配置、命令行参数和配置文件的配置来创建一个node, 并获取相关配置
	stack, cfg := makeConfigNode(ctx, isInConsole)

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

	return stack
}

// dumpConfig is the dumpconfig command.
func dumpConfig(ctx *cli.Context) error {
	cfg := DefaultConfig()
	configPath := getDumpConfigPath(ctx)

	err := makeConfigFile(&cfg, configPath)
	if err != nil {
		utils.Fatalf("%v", err)
		return err
	}
	fmt.Println("Dumping new config file at: " + configPath)
	return nil
}

func getDumpConfigPath(ctx *cli.Context) string {
	configPath := ctx.Args().First()

	// If no path is specified, the default path is used
	if len(configPath) == 0 {
		configPath = defaultConfigPath
	}

	if files.IsDir(configPath) {
		configPath = filepath.Join(configPath, filepath.Base(defaultConfigPath))
	}

	return common.GetAbsPath(configPath)
}

func dumpUserCfg(ctx *cli.Context) error {
	cfg := DefaultConfig()
	cfg.MediatorPlugin = mp.MakeConfig()
	cfg.Jury = jury.MakeConfig()

	configPath := getDumpConfigPath(ctx)
	err := makeConfigFile(&cfg, configPath)
	if err != nil {
		utils.Fatalf("%v", err)
		return err
	}

	fmt.Println("Dumping new config file at: " + configPath)
	return nil
}

// DefaultConfig, create a default config
// @author Albert·Gou
func DefaultConfig() FullConfig {
	// 不是所有的配置都有默认值，例如 Ptnstats 目前没有设置默认值
	return FullConfig{
		Ptn:            ptn.DefaultConfig,
		TxPool:         txspool.DefaultTxPoolConfig,
		Node:           defaultNodeConfig(),
		Dashboard:      dashboard.DefaultConfig,
		P2P:            p2p.DefaultConfig,
		Jury:           jury.DefaultConfig,
		MediatorPlugin: mp.DefaultConfig,
		Dag:            dagconfig.DefaultConfig,
		Log:            log.DefaultConfig,
		Ada:            adaptor.DefaultConfig,
		Contract:       contractcfg.DefaultConfig,
		Certficate:     certficate.DefaultCAConfig,
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
	if err != nil {
		utils.Fatalf("%v", err)
		return err
	}
	defer configFile.Close()
	configToml, err := tomlSettings.Marshal(cfg)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	//log.Debugf("%v", string(configToml))

	_, err = configFile.Write(configToml)
	if err != nil {
		utils.Fatalf("%v", err)
		return err
	}

	return nil
}
