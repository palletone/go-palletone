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

package main

import (
	"os"
	"fmt"
	"encoding/json"
	"gopkg.in/urfave/cli.v1"

	"github.com/palletone/go-palletone/cmd/utils"
	"github.com/palletone/go-palletone/core/gen"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/configure"
	"github.com/palletone/go-palletone/cmd/console"
	"path/filepath"
)

const defaultGenesisJsonPath = "./genesis.json"

var (
	//GenesisJsonPathFlag = utils.DirectoryFlag{
	//	Name:  "genesisjsonpath",
	//	Usage: "Path to create a Genesis State at.",
	//	//	Value: utils.DirectoryString{node.DefaultDataDir()},
	//}

	GenesisJsonPathFlag = cli.StringFlag{
		Name:  "genesis-json-path",
		Usage: "Path to create a Genesis State at.",
		Value: defaultGenesisJsonPath,
	}

	createGenesisJsonCommand = cli.Command{
		Action:utils.MigrateFlags(createGenesisJson),
		Name:"create-genesis-json",
		Usage:"Create a genesis json file template",
		ArgsUsage: "<genesisJsonPath>",
		Flags:[]cli.Flag{
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

// createGenesisJson, Create a json file for the genesis state of a new chain.
func createGenesisJson(ctx *cli.Context) error {
	// Make sure we have a valid genesis JSON
	genesisOut := ctx.Args().First()
	// If no path is specified, the default path is used
	if len(genesisOut) == 0 {
//		utils.Fatalf("Must supply path to genesis JSON file")
		genesisOut = defaultGenesisJsonPath
	}

	var (
		genesisFile *os.File
		err error
	)

	err = os.MkdirAll(filepath.Dir(genesisOut), os.ModePerm)
	if err != nil {
		utils.Fatalf("%v", err)
		return err
	}

	genesisFile,err = os.Create(genesisOut)
	defer genesisFile.Close()
	if err != nil {
		utils.Fatalf("%v", err)
		return err
	}

	var confirm bool
	confirm, err = console.Stdin.PromptConfirm("Do you use an existing account?")
	if err != nil {
		utils.Fatalf("%v", err)
	}

	var account string
	if confirm {
		account, err = console.Stdin.PromptInput("Please enter an existing account address: ")
		if err != nil {
			utils.Fatalf("%v", err)
			return err
		}
	} else {
		account, err = initialAccount(ctx)
		if err != nil {
			utils.Fatalf("%v", err)
			return err
		}
	}

	genesisState := createExampleGenesis(account)

	var genesisJson []byte
	genesisJson, err = json.MarshalIndent(*genesisState, "", "  ")
	if err != nil {
		utils.Fatalf("%v", err)
		return err
	}

	_ ,err = genesisFile.Write(genesisJson)
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

	fmt.Printf("Initial account address: %s\n", address)

	return address.Str(), nil
}

// createExampleGenesis, create the genesis state of new chain with the specified account
func createExampleGenesis(account string)  *core.Genesis  {
	SystemConfig := core.SystemConfig{
		MediatorInterval: gen.DefaultMediatorInterval,
		DepositRate:   gen.DefaultDepositRate,
	}

	return &core.Genesis{
		Version:                   configure.Version,
		TokenAmount:               gen.DefaultTokenAmount,
		TokenDecimal:              gen.DefaultTokenDecimal,
		ChainID:                   1,
		TokenHolder:               account,
		SystemConfig:              SystemConfig,
		InitialActiveMediators:    gen.DefaultMediatorCount,
		InitialMediatorCandidates: gen.InitialMediatorCandidates(
			gen.DefaultMediatorCount, account),
	}
}
