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
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/core/accounts/keystore"
	"github.com/palletone/go-palletone/core"
)

const defaultGenesisJsonPath = "./exampleGenesis.json"

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
If a well-formed JSON file exists at the path, it will be parsed and any 
missing fields in a Genesis State will be added, and any unknown fields will be removed. 

If no file or an invalid file is found, it will be replaced with an example Genesis State.`,
	}
)

// createGenesisJson
func createGenesisJson(ctx *cli.Context) error {
	// Make sure we have a valid genesis JSON
	genesisOut := ctx.Args().First()
	// 如果没有指定路径，则使用默认的路径
	if len(genesisOut) == 0 {
//		utils.Fatalf("Must supply path to genesis JSON file")
		genesisOut = defaultGenesisJsonPath
	}

	var (
		genesisFile *os.File
		err error
	)
	if _, err = os.Stat(genesisOut); err != nil && os.IsNotExist(err) {
		genesisFile,err = os.Create(genesisOut)
	}else {
		genesisFile,err = os.OpenFile(genesisOut, os.O_RDWR,0600)
	}
	defer genesisFile.Close()
	if err != nil {
//		return err
		utils.Fatalf("%v", err)
	}

//	genesisState := gen.DefaultGenesisBlock()

	account, err := initialAccount(ctx)
	if err != nil {
		utils.Fatalf("%v", err)
	}
	genesisState := createExampleGenesis(account.Str())

	var genesisJson []byte
	genesisJson, err = json.MarshalIndent(*genesisState, "", "  ")
	if err != nil {
//		return err
		utils.Fatalf("%v", err)
	}
//	log.Debug(string(genesisJson))

	_ ,err = genesisFile.Write(genesisJson)
	if err != nil {
//		return err
		utils.Fatalf("%v", err)
	}

	fmt.Print("Creating example genesis state in file " + genesisOut)

	return nil
}

func initialAccount(ctx *cli.Context) (common.Address, error) {
	cfg := FullConfig{Node: defaultNodeConfig()}
	// Load config file.
	file := "./palletone.toml"
	if temp := ctx.GlobalString(configFileFlag.Name); temp != "" {
		file = temp
	}
	if err := loadConfig(file, &cfg); err != nil {
		utils.Fatalf("%v", err)
	}

	//utils.SetNodeConfig(ctx, cfg.Node)
	scryptN, scryptP, keydir, err := cfg.Node.AccountConfig()

	password := getPassPhrase("Your new account is locked with a password. Please give a password. Do not forget this password.", true, 0, utils.MakePasswordList(ctx))

	address, err := keystore.StoreKey(keydir, password, scryptN, scryptP)

	if err != nil {
		utils.Fatalf("Failed to create account: %v", err)
	}
//	fmt.Printf("Address Hex: {%x}\n", address)
	fmt.Printf("Address: %s\n", address)
	return address, nil
}

func createExampleGenesis(account string)  *core.Genesis  {
	SystemConfig := core.SystemConfig{
		MediatorInterval: gen.DefaultMediatorInterval,
		DepositRate:   0.02,
	}

	return &core.Genesis{
		Height:                    "0",
		Version:                   "0.6.0",
		TokenAmount:               11111111111,
		TokenDecimal:              8,
		ChainID:                   1,
		TokenHolder:               account,
		SystemConfig:              SystemConfig,
		InitialActiveMediators:    gen.DefaultMediatorCount,
		InitialMediatorCandidates: gen.InitialMediatorCandidates(
			gen.DefaultMediatorCount, account),
	}
}
