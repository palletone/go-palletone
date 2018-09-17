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

	genesisOut, err := getGenesisPathFromConfig(ctx)
	if err != nil {
		return err
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

func getGenesisPathFromConfig(ctx *cli.Context) (string, error) {
	cfg := FullConfig{Node: defaultNodeConfig()}
	// Load config file.
	if err := maybeLoadConfig(ctx, &cfg); err != nil {
		utils.Fatalf("%v", err)
		return "", err
	}

	// Make sure we have a valid genesis JSON
	genesisOut := ctx.Args().First()
	// If no path is specified, the default path is used
	if len(genesisOut) == 0 {
		//		utils.Fatalf("Must supply path to genesis JSON file")
		genesisOut, _ = getGenesisPath(defaultGenesisJsonPath, cfg.Node.DataDir)
	}

	return genesisOut, nil
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

func createExampleAccount(ctx *cli.Context) (addrStr, password string) {
	password = core.DefaultPassword
	address, _ := createAccount(ctx, password)
	addrStr = address.Str()
	return
}

// createExampleGenesis, create the genesis state of new chain with the specified account
func createExampleGenesis(account string) *core.Genesis {
	SystemConfig := core.SystemConfig{
		DepositRate: core.DefaultDepositRate,
	}

	initParams := core.NewChainParams()

	return &core.Genesis{
		Alias:                  core.DefaultAlias,
		Version:                configure.Version,
		TokenAmount:            core.DefaultTokenAmount,
		TokenDecimal:           core.DefaultTokenDecimal,
		ChainID:                core.DefaultChainID,
		TokenHolder:            account,
		Text:                   core.DefaultText,
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
