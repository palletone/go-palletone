// Copyright 2016 The go-ethereum Authors
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
	"fmt"
	"github.com/palletone/go-palletone/contracts/comm"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/palletone/go-palletone/cmd/console"
	"github.com/palletone/go-palletone/cmd/utils"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/rpc"
	"github.com/palletone/go-palletone/core/node"
	"gopkg.in/urfave/cli.v1"
	"runtime"
)

var (
	consoleFlags = []cli.Flag{utils.JSpathFlag, utils.ExecFlag, utils.PreloadJSFlag}

	consoleCommand = cli.Command{
		Action: utils.MigrateFlags(localConsole),
		Name:   "console",
		Usage:  "Start an interactive JavaScript environment",
		//Flags:    append(append(append(nodeFlags, rpcFlags...), consoleFlags...), whisperFlags...),
		Flags:    append(append(nodeFlags, rpcFlags...), consoleFlags...),
		Category: "CONSOLE COMMANDS",
		Description: `
The Geth console is an interactive shell for the JavaScript runtime environment
which exposes a node admin interface as well as the Ðapp JavaScript API.
See https://github.com/palletone/go-palletone/wiki/JavaScript-Console.`,
	}

	attachCommand = cli.Command{
		Action:    utils.MigrateFlags(remoteConsole),
		Name:      "attach",
		Usage:     "Start an interactive JavaScript environment (connect to node)",
		ArgsUsage: "[endpoint]",
		Flags:     append(consoleFlags, utils.DataDirFlag, ConfigFilePathFlag),
		Category:  "CONSOLE COMMANDS",
		Description: `
The Geth console is an interactive shell for the JavaScript runtime environment
which exposes a node admin interface as well as the Ðapp JavaScript API.
See https://github.com/palletone/go-palletone/wiki/JavaScript-Console.
This command allows to open a console on a running gptn node.`,
	}

	javascriptCommand = cli.Command{
		Action:    utils.MigrateFlags(ephemeralConsole),
		Name:      "js",
		Usage:     "Execute the specified JavaScript files",
		ArgsUsage: "<jsfile> [jsfile...]",
		Flags:     append(nodeFlags, consoleFlags...),
		Category:  "CONSOLE COMMANDS",
		Description: `
The JavaScript VM exposes a node admin interface as well as the Ðapp
JavaScript API. See https://github.com/palletone/go-palletone/wiki/JavaScript-Console`,
	}
	juryListenerIpCommand = cli.Command{
		Action:             utils.MigrateFlags(getInternalIp),
		Name:               "getJuryIp",
		Usage:              "Get local ip for jury to listener for user contracts",
		Description:        `
gptn getJuryIp`,
	}
)

// localConsole starts a new gptn node, attaching a JavaScript console to it at the
// same time.
func localConsole(ctx *cli.Context) error {
	// Create and start the node based on the CLI flags
	node := makeFullNode(ctx, true)
	startNode(ctx, node)
	defer node.Stop()

	// Attach to the newly started node and start the JavaScript console
	client, err := node.Attach()
	if err != nil {
		utils.Fatalf("Failed to attach to the inproc gptn: %v", err)
	}
	config := console.Config{
		DataDir: node.DataDir(), //utils.MakeDataDir(ctx),
		DocRoot: ctx.GlobalString(utils.JSpathFlag.Name),
		Client:  client,
		Preload: utils.MakeConsolePreloads(ctx),
	}

	console, err := console.New(config)
	if err != nil {
		utils.Fatalf("Failed to start the JavaScript console1: %v", err)
	}
	defer console.Stop(false)

	// If only a short execution was requested, evaluate and return
	if script := ctx.GlobalString(utils.ExecFlag.Name); script != "" {
		console.Evaluate(script)
		return nil
	}
	// Otherwise print the welcome screen and enter interactive mode
	console.Welcome()
	console.Interactive()

	return nil
}

// remoteConsole will connect to a remote gptn instance, attaching a JavaScript
// console to it.
func remoteConsole(ctx *cli.Context) error {
	// 1. 获取和计算 endpoint 和 dataDir 的实际路径
	endpoint := ctx.Args().First()
	dataDir := ctx.GlobalString(utils.DataDirFlag.Name)
	cfg := DefaultConfig()

	if endpoint == "" || dataDir == "" {
		if runtime.GOOS != "windows" {
			// 在非windows系统的中，ipc文件放在 dataDir 下

			if endpoint == "" && dataDir == "" {
				dataDir = cfg.Node.DataDir
				endpoint = filepath.Join(dataDir, cfg.Node.IPCPath)

				// 如果当前目录没有 ipc文件，则从配置文件中读取
				if !common.IsExisted(endpoint) {
					configPath := getConfigPath(ctx)

					err := loadConfig(configPath, &cfg)
					if err != nil {
						utils.Fatalf("%v", err)
						return err
					}

					utils.SetNodeConfig(ctx, &cfg.Node, filepath.Dir(configPath))

					dataDir = cfg.Node.DataDir
					endpoint = cfg.Node.IPCPath

					if !filepath.IsAbs(endpoint) {
						endpoint = filepath.Join(dataDir, endpoint)
					}
				}
			} else if endpoint == "" {
				endpoint = fmt.Sprintf("%s/gptn.ipc", dataDir)
			} else {
				dataDir = filepath.Dir(endpoint)
			}
		} else {
			// 在windows系统下，ipc文件在\\.\pipe\ 目录下, 没在 dataDir 下

			if endpoint == "" && dataDir == "" {
				configPath := getConfigPath(ctx)
				if common.IsExisted(configPath) {
					err := loadConfig(configPath, &cfg)
					if err != nil {
						utils.Fatalf("%v", err)
						return err
					}

					utils.SetNodeConfig(ctx, &cfg.Node, filepath.Dir(configPath))
				}

				endpoint = cfg.Node.IPCPath
				if !strings.HasPrefix(endpoint, `\\.\pipe\`) {
					endpoint = `\\.\pipe\` + endpoint
				}

				dataDir = cfg.Node.DataDir
			} else if endpoint == "" {
				endpoint = cfg.Node.IPCPath
				if !strings.HasPrefix(endpoint, `\\.\pipe\`) {
					endpoint = `\\.\pipe\` + endpoint
				}
			} else {
				dataDir = cfg.Node.DataDir
			}
		}
	}

	// 设置 log 的配置
	utils.SetLogConfig(ctx, &cfg.Log, filepath.Dir(dataDir), true)

	// 2. 连接 gptn
	client, err := dialRPC(endpoint)
	if err != nil {
		utils.Fatalf("Unable to attach to remote gptn: %v", err)
	}
	config := console.Config{
		DataDir: dataDir, //utils.MakeDataDir(ctx),
		DocRoot: ctx.GlobalString(utils.JSpathFlag.Name),
		Client:  client,
		Preload: utils.MakeConsolePreloads(ctx),
	}

	// 3. 创建 console
	console, err := console.New(config)
	if err != nil {
		utils.Fatalf("Failed to start the JavaScript console2: %v", err)
	}
	defer console.Stop(false)

	if script := ctx.GlobalString(utils.ExecFlag.Name); script != "" {
		console.Evaluate(script)
		return nil
	}

	// 4. Otherwise print the welcome screen and enter interactive mode
	console.Welcome()
	console.Interactive()

	return nil
}

// dialRPC returns a RPC client which connects to the given endpoint.
// The check for empty endpoint implements the defaulting logic
// for "gptn attach" and "gptn monitor" with no argument.
func dialRPC(endpoint string) (*rpc.Client, error) {
	if endpoint == "" {
		endpoint = node.DefaultIPCEndpoint(clientIdentifier)
	} else if strings.HasPrefix(endpoint, "rpc:") || strings.HasPrefix(endpoint, "ipc:") {
		// Backwards compatibility with gptn < 1.5 which required
		// these prefixes.
		endpoint = endpoint[4:]
	}
	return rpc.Dial(endpoint)
}

// ephemeralConsole starts a new gptn node, attaches an ephemeral JavaScript
// console to it, executes each of the files specified as arguments and tears
// everything down.
func ephemeralConsole(ctx *cli.Context) error {
	// Create and start the node based on the CLI flags
	node := makeFullNode(ctx, true)
	startNode(ctx, node)
	defer node.Stop()
	// Attach to the newly started node and start the JavaScript console
	client, err := node.Attach()
	if err != nil {
		utils.Fatalf("Failed to attach to the inproc gptn: %v", err)
	}
	config := console.Config{
		DataDir: node.DataDir(), //utils.MakeDataDir(ctx),
		DocRoot: ctx.GlobalString(utils.JSpathFlag.Name),
		Client:  client,
		Preload: utils.MakeConsolePreloads(ctx),
	}

	console, err := console.New(config)
	if err != nil {
		utils.Fatalf("Failed to start the JavaScript console3: %v", err)
	}
	defer console.Stop(false)

	// Evaluate each of the specified JavaScript files
	for _, file := range ctx.Args() {
		if err = console.Execute(file); err != nil {
			utils.Fatalf("Failed to execute %s: %v", file, err)
		}
	}
	// Wait for pending callbacks, but stop for Ctrl-C.
	abort := make(chan os.Signal, 1)
	signal.Notify(abort, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-abort
		os.Exit(0)
	}()
	console.Stop(true)

	return nil
}

func getInternalIp(ctx *cli.Context) error {
	localIp := comm.GetInternalIp()
	if localIp != "" {
		fmt.Println(localIp)
		return nil
	}
	fmt.Println("Not find your local ip")
	return nil
}