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

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/palletone/go-palletone/cmd/console"
	"github.com/palletone/go-palletone/cmd/utils"
	"github.com/palletone/go-palletone/configure"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/core/gen"
	"gopkg.in/urfave/cli.v1"
)

const defaultGenesisJsonPath = "ptn-genesis.json"

var (
	GenesisTimestampFlag = cli.Int64Flag{
		Name:  "genesistime",
		Usage: "Replace timestamp from genesis.json with current time plus this many seconds (experts only!)",
		//		Value: 0,
	}

	GenesisJsonPathFlag = cli.StringFlag{
		Name:  "genesispath",
		Usage: "Path to create a Genesis State at.",
		Value: defaultGenesisJsonPath,
	}

	createGenesisJsonCommand = cli.Command{
		Action:    utils.MigrateFlags(createGenesisJson),
		Name:      "newgenesis",
		Usage:     "Create a genesis json file template",
		ArgsUsage: "<genesisJsonPath>",
		Flags: []cli.Flag{
			GenesisJsonPathFlag,
		},
		Category: "BLOCKCHAIN COMMANDS",
		Description: `
Create a json file for the genesis state of a new chain 
with an existing account or a newly created account.

If a well-formed JSON file exists at the path, 
it will be replaced with an example Genesis State.`,
	}
)

func getTokenAccount(ctx *cli.Context) (string, error) {
	confirm, err := console.Stdin.PromptConfirm(
		"Do you want to create a new account as the holder of the token?")
	if err != nil {
		utils.Fatalf("%v", err)
	}

	var account string
	if !confirm {
		account, err = console.Stdin.PromptInput("Please enter an existing account address: ")
		if err != nil {
			utils.Fatalf("%v", err)
			return "", err
		}
	} else {
		account, err = initialAccount(ctx)
		if err != nil {
			utils.Fatalf("%v", err)
			return "", err
		}
	}

	return account, nil
}

// createGenesisJson, Create a json file for the genesis state of a new chain.
func createGenesisJson(ctx *cli.Context) error {
	var (
		genesisFile *os.File
		err         error
	)

	account, err := getTokenAccount(ctx)
	if err != nil {
		return err
	}
	genesisState := createExampleGenesis(account)

	var genesisJson []byte
	genesisJson, err = json.MarshalIndent(*genesisState, "", "  ")
	if err != nil {
		utils.Fatalf("%v", err)
		return err
	}

	cfg := FullConfig{Node: defaultNodeConfig()}
	// Load config file.
	if err := maybeLoadConfig(ctx, &cfg); err != nil {
		utils.Fatalf("%v", err)
	}

	// Make sure we have a valid genesis JSON
	genesisOut := ctx.Args().First()
	// If no path is specified, the default path is used
	if len(genesisOut) == 0 {
		//		utils.Fatalf("Must supply path to genesis JSON file")
		genesisOut, _ = getGenesisPath(defaultGenesisJsonPath, cfg.Node.DataDir)
	}

	err = os.MkdirAll(filepath.Dir(genesisOut), os.ModePerm)
	if err != nil {
		utils.Fatalf("%v", err)
		return err
	}

	genesisFile, err = os.Create(genesisOut)
	defer genesisFile.Close()
	if err != nil {
		utils.Fatalf("%v", err)
		return err
	}

	_, err = genesisFile.Write(genesisJson)
	if err != nil {
		utils.Fatalf("%v", err)
		return err
	}

	fmt.Println("Creating example genesis state in file " + genesisOut)

	return nil
}

// initialAccount, create a initial account for a new account
func initialAccount(ctx *cli.Context) (string, error) {
	address, err := newAccount(ctx)
	if err != nil {
		utils.Fatalf("%v", err)
	}

	fmt.Printf("Initial token holder's account address: %s\n", address.String())

	return address.Str(), nil
}

// createExampleGenesis, create the genesis state of new chain with the specified account
func createExampleGenesis(account string) *core.Genesis {
	SystemConfig := core.SystemConfig{
		DepositRate: core.DefaultDepositRate,
	}

	initParams := core.NewChainParams()

	return &core.Genesis{
		Alias:        core.DefaultAlias,
		Version:      configure.Version,
		TokenAmount:  core.DefaultTokenAmount,
		TokenDecimal: core.DefaultTokenDecimal,
		ChainID:      1,
		TokenHolder:  account,
		//Text:         "Hello PalletOne!",
                Text: "姓名 丨 坐标 丨 简介   \r\n" +
		"孟岩丨北京丨通证派倡导者、CSDN副总裁、柏链道捷CEO.\r\n" +
		"刘百祥丨上海丨 GoC-lab发起人兼技术社群负责人,复旦大学计算机博士.\r\n" +
		"陈澄丨上海丨引力区开发者社区总理事,EOS超级节点负责人.\r\n" +
		"孙红景丨北京丨CTO、13年IT开发和管理经验.\r\n" +
		"kobegpfan丨北京丨世界500强企业技术总监.\r\n" +
		"余奎丨上海丨加密经济学研究员、产品研发经理.\r\n" +
		"Shangsong丨北京丨Fabric、 多链、 分片 、跨链技术.\r\n" +
		"郑广军丨上海丨区块链java应用开发.\r\n" +
		"钮祜禄虫丨北京丨大数据架构、Dapp开发.\r\n" +
		"彭敏丨四川丨计算机网络和系统集成十余年有经验.\r\n",
		SystemConfig:           SystemConfig,
		InitialParameters:      initParams,
		ImmutableParameters:    core.NewImmutChainParams(),
		InitialTimestamp:       gen.InitialTimestamp(initParams.MediatorInterval),
		InitialActiveMediators: core.DefaultMediatorCount,
		InitialMediatorCandidates: gen.InitialMediatorCandidates(
			core.DefaultMediatorCount, account),
	}
}

// 根据指定路径和配置参数获取创世Json文件的路径
// @author Albert·Gou
func getGenesisPath(genesisPath, dataDir string) (string, error) {
	if filepath.IsAbs(genesisPath) {
		return genesisPath, nil
	}

	if dataDir != "" && genesisPath == "" {
		return filepath.Join(dataDir, defaultGenesisJsonPath), nil
	}

	if genesisPath != "" {
		return filepath.Abs(genesisPath)
	}

	return "", nil
}
