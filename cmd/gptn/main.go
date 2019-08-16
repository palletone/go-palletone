// Copyright 2018 PalletOne
// Copyright 2014 The go-ethereum Authors
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

// gptn is the official command-line client for PalletOne.
package main

import (
	"fmt"
	"os"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/palletone/go-palletone/cmd/console"
	"github.com/palletone/go-palletone/cmd/utils"
	"github.com/palletone/go-palletone/common/log"
	mp "github.com/palletone/go-palletone/consensus/mediatorplugin"
	"github.com/palletone/go-palletone/core/accounts"
	"github.com/palletone/go-palletone/core/node"
	"github.com/palletone/go-palletone/internal/debug"
	"github.com/palletone/go-palletone/ptnclient"
	"github.com/palletone/go-palletone/statistics/metrics"
	"gopkg.in/urfave/cli.v1"
)

const (
	clientIdentifier = "gptn" // Client identifier to advertise over the network
)

var (
	// Git SHA1 commit hash of the release (set via linker flags)
	gitCommit = ""
	// The app that holds all commands and flags.
	// 新建一个全局的app结构，用来管理程序启动，命令行配置等
	app = utils.NewApp(gitCommit, "the go-palletone command line interface")
	// flags that configure the node
	nodeFlags = []cli.Flag{
		utils.IdentityFlag,
		utils.UnlockedAccountFlag,
		utils.PasswordFileFlag,
		utils.BootnodesFlag,
		utils.BootnodesV4Flag,
		utils.BootnodesV5Flag,
		utils.DataDirFlag,
		utils.KeyStoreDirFlag,
		utils.NoUSBFlag,
		utils.DashboardEnabledFlag,
		utils.DashboardAddrFlag,
		utils.DashboardPortFlag,
		utils.DashboardRefreshFlag,
		utils.TxPoolNoLocalsFlag,
		utils.TxPoolJournalFlag,
		utils.TxPoolRejournalFlag,
		utils.TxPoolPriceLimitFlag,
		utils.TxPoolPriceBumpFlag,
		utils.TxPoolGlobalSlotsFlag,
		//utils.TxPoolAccountQueueFlag,
		utils.TxPoolGlobalQueueFlag,
		utils.TxPoolLifetimeFlag,
		utils.FastSyncFlag,
		utils.LightModeFlag,
		utils.SyncModeFlag,
		utils.GCModeFlag,
		utils.LightServFlag,
		utils.LightPeersFlag,
		utils.LightKDFFlag,
		utils.CacheFlag,
		utils.CacheDatabaseFlag,
		utils.CacheGCFlag,
		utils.TrieCacheGenFlag,
		utils.ListenPortFlag,
		utils.MaxPeersFlag,
		utils.MaxPendingPeersFlag,
		utils.EtherbaseFlag,
		utils.CryptoLibFlag,
		utils.MinerThreadsFlag,
		utils.MiningEnabledFlag,
		//utils.TargetGasLimitFlag,
		utils.NATFlag,
		utils.NoDiscoverFlag,
		utils.DiscoveryV5Flag,
		utils.NetrestrictFlag,
		utils.NodeKeyFileFlag,
		utils.NodeKeyHexFlag,
		utils.DeveloperFlag,
		utils.DeveloperPeriodFlag,
		utils.TestnetFlag,
		utils.VMEnableDebugFlag,
		utils.NetworkIdFlag,
		utils.RPCCORSDomainFlag,
		utils.RPCVirtualHostsFlag,
		utils.EthStatsURLFlag,
		utils.MetricsEnabledFlag,
		utils.FakePoWFlag,
		utils.NoCompactionFlag,
		// utils.GpoBlocksFlag,
		// utils.GpoPercentileFlag,
		utils.ExtraDataFlag,
		//utils.DagValue1Flag,
		//utils.DagValue2Flag,
		utils.LogOutputPathFlag,
		utils.LogLevelFlag,
		utils.LogIsDebugFlag,
		utils.LogErrPathFlag,
		utils.LogEncodingFlag,
		utils.LogOpenModuleFlag,
		ConfigFilePathFlag,
		GenesisJsonPathFlag,
		GenesisTimestampFlag,
	}

	rpcFlags = []cli.Flag{
		utils.RPCEnabledFlag,
		utils.RPCListenAddrFlag,
		utils.RPCPortFlag,
		utils.RPCApiFlag,
		utils.WSEnabledFlag,
		utils.WSListenAddrFlag,
		utils.WSPortFlag,
		utils.WSApiFlag,
		utils.WSAllowedOriginsFlag,
		utils.IPCDisabledFlag,
		utils.IPCPathFlag,
	}
)

func init() {
	// 先调用初始化函数，设置app的各个参数
	// Initialize the CLI app and start Gptn
	// gptn处理函数会在 app.HandleAction 里面调用
	app.Action = gptn      //默认的操作，就是启动一个gptn节点， 如果有其他子命令行参数，会调用到下面的Commands里面去
	app.HideVersion = true // we have a command to print the version
	app.Copyright = "Copyright 2017-2018 The go-palletone Authors"
	// 设置各个子命令的处理类/函数，比如consoleCommand 最后调用到 localConsole
	// 如果命令行参数里面有下面的指令，就会直接调用下面的Command.Run方法，而不调用默认的app.Action方法
	// Commands 是程序支持的所有子命令
	app.Commands = []cli.Command{
		// See chaincmd.go:
		initCommand, //初始化创世单元命令
		//importCommand,
		//exportCommand,
		//importPreimagesCommand,
		//exportPreimagesCommand,
		copydbCommand,
		removedbCommand,
		//dumpCommand,	//转储命令
		// See monitorcmd.go:
		monitorCommand,
		// See accountcmd.go:
		accountCommand,
		// walletCommand,
		// See consolecmd.go:
		consoleCommand, //js控制台命令
		attachCommand,
		javascriptCommand,
		// See misccmd.go:
		makecacheCommand,
		makedagCommand,
		versionCommand,
		bugCommand,
		licenseCommand,
		dumpConfigCommand,        //转储配置文件命令
		dumpUserCfgCommand,       //no mediator,no jury
		dumpJsonCommand,          //create genesis.json
		createGenesisJsonCommand, // 创建创世json文件命令
		nodeInfoCommand,          // 获取本节点信息
		timestampCommand,         // 获取指定时间的时间戳
		mediatorCommand,          // mediator 管理
		//certCommand,              //证书管理

	}
	sort.Sort(cli.CommandsByName(app.Commands))

	// 所有从命令行能够解析的Options
	app.Flags = append(app.Flags, nodeFlags...)
	app.Flags = append(app.Flags, rpcFlags...)
	app.Flags = append(app.Flags, consoleFlags...)
	app.Flags = append(app.Flags, debug.Flags...)
	app.Flags = append(app.Flags, mp.MediatorFlags...)

	// before函数在app.Run的开始会先调用，也就是gopkg.in/urfave/cli.v1/app.go Run函数的前面
	app.Before = func(ctx *cli.Context) error {
		// 设置最大可用处理器数
		runtime.GOMAXPROCS(runtime.NumCPU())
		if err := debug.Setup(ctx); err != nil {
			return err
		}
		// Start system runtime metrics collection
		// 创建一个goroutine，每3秒监测一次系统的ram和disk状态
		go metrics.CollectProcessMetrics(3 * time.Second)

		utils.SetupNetwork(ctx)
		return nil
	}

	//after函数在最后调用，app.Run 里面会设置defber function
	app.After = func(ctx *cli.Context) error {
		debug.Exit()
		console.Stdin.Close() // Resets terminal mode.
		return nil
	}
}

func main() {
	// 如果是gptn命令行启动，不带子命令，那么直接调用app.Action = gptn()函数；
	// 如果带有子命令比如gptn console，那么会调用Command.Run, 最终会执行该子命令对应的Command.Action
	// 对于console子命令来说就是调用的 localConsole()函数；调用路径为：
	/*
		1. app.Run(os.Args);
		2. c.Run(context)
		3. HandleAction(c.Action, context)
	*/
	//welcomePalletOne()
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// gptn is the main entry point into the system if no special subcommand is ran.
// It creates a default node based on the command line arguments and runs it in
// blocking mode, waiting for it to be shut down.
func gptn(ctx *cli.Context) error {
	welcomePalletOne()
	node := makeFullNode(ctx, false)
	startNode(ctx, node)
	node.Wait()
	return nil
}

// startNode boots up the system node and all registered protocols, after which
// it unlocks any requested accounts, and starts the RPC/IPC interfaces and the
// miner.
func startNode(ctx *cli.Context, stack *node.Node) {
	debug.Memsize.Add("node", stack)

	// Start up the node itself
	// 启动节点本身, 将之前注册的所有 service 交给 p2p.Server, 然后启动
	utils.StartNode(stack)

	// Unlock any account specifically requested
	ks := stack.GetKeyStore()

	//自动解锁指定的账号，配置的, 这样非交互状态下方便使用
	passwords := utils.MakePasswordList(ctx)
	unlocks := strings.Split(ctx.GlobalString(utils.UnlockedAccountFlag.Name), ",")
	for i, account := range unlocks {
		if trimmed := strings.TrimSpace(account); trimmed != "" {
			unlockAccount(ks, trimmed, i, passwords)
		}
	}

	// Register wallet event handlers to open and auto-derive wallets
	// 注册钱包事件, 来处理打开钱包和自动派生钱包
	events := make(chan accounts.WalletEvent, 16)
	stack.AccountManager().Subscribe(events)

	// 创建协程，创建链状态读取器、用 RPC 监听钱包供远程调用
	go func() {
		// Create a chain state reader for self-derivation
		rpcClient, err := stack.Attach()
		if err != nil {
			utils.Fatalf("Failed to attach to self: %v", err)
		}
		stateReader := ptnclient.NewClient(rpcClient)

		// Open any wallets already attached
		for _, wallet := range stack.AccountManager().Wallets() {
			if err := wallet.Open(""); err != nil {
				log.Warn("Failed to open wallet", "url", wallet.URL(), "err", err)
			}
		}
		// Listen for wallet event till termination
		for event := range events {
			switch event.Kind {
			case accounts.WalletArrived:
				if err := event.Wallet.Open(""); err != nil {
					log.Warn("New wallet appeared, failed to open", "url", event.Wallet.URL(), "err", err)
				}
			case accounts.WalletOpened:
				status, _ := event.Wallet.Status()
				log.Info("New wallet appeared", "url", event.Wallet.URL(), "status", status)

				if event.Wallet.URL().Scheme == "ledger" {
					event.Wallet.SelfDerive(accounts.DefaultLedgerBaseDerivationPath, stateReader)
				} else {
					event.Wallet.SelfDerive(accounts.DefaultBaseDerivationPath, stateReader)
				}

			case accounts.WalletDropped:
				log.Info("Old wallet dropped", "url", event.Wallet.URL())
				event.Wallet.Close()
			}
		}
	}()
	// Start auxiliary services if enabled
	//if ctx.GlobalBool(utils.MiningEnabledFlag.Name) || ctx.GlobalBool(utils.DeveloperFlag.Name) {
	//	// Mining only makes sense if a full PalletOne node is running
	//	if ctx.GlobalBool(utils.LightModeFlag.Name) || ctx.GlobalString(utils.SyncModeFlag.Name) == "light" {
	//		utils.Fatalf("Light clients do not support mining")
	//	}
	//	var palletone *ptn.PalletOne
	//	if err := stack.Service(&palletone); err != nil {
	//		utils.Fatalf("PalletOne service not running: %v", err)
	//	}
	//}
}

func welcomePalletOne() {
	///*
	//"*    _____      _ _      _    ____                    *\n"
	//"*   |  __ \    | | |    | |  / __ \                   *\n"
	//"*   | |__) |_ _| | | ___| |_| |  | |_ __   ___        *\n"
	//"*   |  ___/ _` | | |/ _ \ __| |  | | '_ \ / _ \       *\n"
	//"*   | |  | (_| | | |  __/ |_| |__| | | | |  __/       *\n"
	//"*   |_|   \__,_|_|_|\___|\__|\____/|_| |_|\___|       *\n"
	//*/
	//
	//
	fmt.Print("\n" +
		"    * * * * * Welcome to PalletOne! * * * * *        \n" +
		"    _____      _ _      _    ____                    \n" +
		"   |  __ \\    | | |    | |  / __ \\                 \n" +
		"   | |__) |_ _| | | ___| |_| |  | |_ __   ___        \n" +
		"   |  ___/ _` | | |/ _ \\ __| |  | | '_ \\ / _ \\    \n" +
		"   | |  | (_| | | |  __/ |_| |__| | | | |  __/       \n" +
		"   |_|   \\__,_|_|_|\\___|\\__|\\____/|_| |_|\\___|  \n" +
		"                                                     \n" +
		"    * * * * * * * * * * * * * * * * * * * * *        \n" +
		"\n")
}
