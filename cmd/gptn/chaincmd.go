// Copyright 2015 The go-ethereum Authors
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
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/palletone/go-palletone/cmd/utils"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/core/gen"
	"github.com/palletone/go-palletone/dag"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/modules"

	"gopkg.in/urfave/cli.v1"
)

var (
	initCommand = cli.Command{
		Action:    utils.MigrateFlags(initGenesis),
		Name:      "init",
		Usage:     "Bootstrap and initialize a new genesis block",
		ArgsUsage: "<genesisPath>",
		Flags: []cli.Flag{
			utils.DataDirFlag,
			GenesisJsonPathFlag,
			GenesisTimestampFlag,
			// utils.LightModeFlag,
		},
		Category: "BLOCKCHAIN COMMANDS",
		Description: `
The init command initializes a new genesis block and definition for the network.
This is a destructive action and changes the network in which you will be
participating.

It expects the genesis file as argument.`,
	}

	copydbCommand = cli.Command{
		Action:    utils.MigrateFlags(copyDb),
		Name:      "copydb",
		Usage:     "Create a local chain from a target chaindata folder",
		ArgsUsage: "<sourceChaindataDir>",
		Flags: []cli.Flag{
			utils.DataDirFlag,
			utils.CacheFlag,
			utils.SyncModeFlag,
			utils.FakePoWFlag,
			utils.TestnetFlag,
		},
		Category: "BLOCKCHAIN COMMANDS",
		Description: `
The first argument must be the directory containing the blockchain to download from`,
	}
	removedbCommand = cli.Command{
		Action:    utils.MigrateFlags(removeDB),
		Name:      "removedb",
		Usage:     "Remove blockchain and state databases",
		ArgsUsage: " ",
		Flags: []cli.Flag{
			utils.DataDirFlag,
			utils.LightModeFlag,
		},
		Category: "BLOCKCHAIN COMMANDS",
		Description: `
Remove blockchain and state databases`,
	}
)

func copyDb(ctx *cli.Context) error {
	return nil
}

func removeDB(ctx *cli.Context) error {
	return nil
}

func initGenesis(ctx *cli.Context) error {
	node := makeFullNode(ctx, false)

	// Make sure we have a valid genesis JSON
	genesisPath := getGenesisPath(ctx)

	file, err := os.Open(genesisPath)
	if err != nil {
		utils.Fatalf("Failed to read genesis file: %v", err)
	}
	defer file.Close()

	genesis := new(core.Genesis)
	if err := json.NewDecoder(file).Decode(genesis); err != nil {
		utils.Fatalf("invalid genesis file: %v", err)
	}

	regulateGenesisTimestamp(ctx, genesis)

	validateGenesis(genesis)

	dbPath := dagconfig.DagConfig.DbPath
	Dbconn, err := node.OpenDatabase(dbPath, 0, 0)
	if err != nil {
		return errors.New("leveldb init failed")
	}
	dag, _ := dag.NewDag4GenesisInit(Dbconn)
	ks := node.GetKeyStore()
	account, password := unlockAccount(ks, genesis.TokenHolder, 0, nil)

	err = ks.Unlock(account, password)
	if err != nil {
		utils.Fatalf("Failed to unlock account: %v, address: %v", err, account.Address.Str())
		return err
	}
	//判断当前DB是否为空，不为空则报错。
	if !dag.IsEmpty() {
		return errors.New("leveldb is not empty")
	}
	//将Genesis对象转换为一个Unit
	unit, err := gen.SetupGenesisUnit(genesis, ks, account)
	if err != nil {
		utils.Fatalf("Failed to generate genesis unit: %v", err)
		return err
	}
	log.Debugf("Save unit:%#v", unit)
	//将Unit存入数据库中
	err = dag.SaveUnit(unit, nil, true)
	if err != nil {
		fmt.Println("Save Genesis unit to db error:", err)
		return err
	}

	// 初始化 stateDB
	// append by albert·gou
	err = dag.InitStateDB(genesis, unit)
	if err != nil {
		utils.Fatalf("Failed to InitStateDB: %v", err)
		return err
	}

	// 初始化属性数据库
	// modified by Yiran
	err = dag.InitPropertyDB(genesis, unit)
	if err != nil {
		utils.Fatalf("Failed to InitPropertyDB: %v", err)
		return err
	}
	dv := new(modules.DataVersion)
	dv.Name = "Gptn"
	dv.Version = genesis.Version
	dag.StoreDataVersion(dv)

	//MUST DO NOT MODIFY THIS LOG. For deploy.sh
	log.Infof("gptn (version[%s] hash[%s]) init success", dv.Version, unit.UnitHash.Hex())
	return nil
}
