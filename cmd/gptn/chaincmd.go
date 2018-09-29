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
	"github.com/palletone/go-palletone/cmd/utils"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	mp "github.com/palletone/go-palletone/consensus/mediatorplugin"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/core/accounts"
	"github.com/palletone/go-palletone/core/gen"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/storage"
	"gopkg.in/urfave/cli.v1"
	"os"
)

var (
	initCommand = cli.Command{
		Action:    utils.MigrateFlags(initGenesis),
		Name:      "init",
		Usage:     "Bootstrap and initialize a new genesis block",
		ArgsUsage: "<genesisPath>",
		Flags: []cli.Flag{
			//			utils.DataDirFlag,
			GenesisJsonPathFlag,
			GenesisTimestampFlag,
			//			utils.LightModeFlag,
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

func getAccountFromConf(configPath string) (account accounts.Account, passphrase string) {
	cfg := new(FullConfig)

	// 加载配置文件中的配置信息到 cfg中
	err := loadConfig(configPath, cfg)
	if err != nil {
		utils.Fatalf("%v", err)
	}

	med := cfg.MediatorPlugin.Mediators[0]
	addr, err := common.StringToAddress(med.Address)
	if err != nil {
		utils.Fatalf("%v", err)
	}

	account = accounts.Account{Address: addr}
	passphrase = med.Password
	return
}

func initGenesis(ctx *cli.Context) error {
	node := makeFullNode(ctx)

	// Make sure we have a valid genesis JSON
	genesisPath := ctx.Args().First()
	//if len(genesisPath) == 0 {
	//	utils.Fatalf("Must supply path to genesis JSON file")
	//}
	// If no path is specified, the default path is used
	if len(genesisPath) == 0 {
		genesisPath, _ = getGenesisPath(defaultGenesisJsonPath, node.DataDir())
	}
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

	Dbconn, err := node.OpenDatabase("leveldb", 0, 0)
	if err != nil {
		fmt.Println("eveldb init failed")
		return errors.New("leveldb init failed")
	}
	filepath := node.ResolvePath("leveldb")
	dagconfig.DbPath = filepath

	ks := node.GetKeyStore()
	// modify by Albert·Gou
	account, password := unlockAccount(nil, ks, genesis.TokenHolder, 0, nil)

	// 从配置文件中获取账户和密码
	//configPath := defaultConfigPath
	//if temp := ctx.GlobalString(ConfigFileFlag.Name); temp != "" {
	//	configPath, _ = getConfigPath(temp, node.DataDir())
	//}
	//account, password := getAccountFromConf(configPath)

	err = ks.Unlock(account, password)
	if err != nil {
		utils.Fatalf("Failed to unlock account: %v, address: %v", err, account.Address.Str())
		return err
	}

	unit, err := gen.SetupGenesisUnit(Dbconn, genesis, ks, account)
	if err != nil {
		utils.Fatalf("Failed to write genesis unit: %v", err)
		return err
	}

	genesisUnitHash := unit.UnitHash
	log.Info(fmt.Sprintf("Successfully Get Genesis Unit, it's hash: %v", genesisUnitHash.Hex()))

	// 2, 重写配置文件，修改当前节点的mediator的地址和密码
	// @author Albert·Gou
	//configPath := defaultConfigPath
	//if temp := ctx.GlobalString(ConfigFileFlag.Name); temp != "" {
	//	configPath, _ = getConfigPath(temp, node.DataDir())
	//}
	//modifyMediatorInConf(configPath, password, account.Address)

	// 3, 全局属性不是交易，不需要放在Unit中
	// @author Albert·Gou
	gp := modules.InitGlobalProp(genesis)
	storage.StoreGlobalProp(Dbconn, gp)
	if err != nil {
		utils.Fatalf("Failed to write global properties: %v", err)
		return err
	}

	// 4, 动态全局属性不是交易，不需要放在Unit中
	// @author Albert·Gou
	dgp := modules.InitDynGlobalProp(genesis, genesisUnitHash)
	storage.StoreDynGlobalProp(Dbconn, dgp)
	if err != nil {
		utils.Fatalf("Failed to write dynamic global properties: %v", err)
		return err
	}

	// 5, 初始化mediator调度器，并存在数据库
	// @author Albert·Gou
	ms := modules.InitMediatorSchl(gp, dgp)
	storage.StoreMediatorSchl(Dbconn, ms)
	if err != nil {
		utils.Fatalf("Failed to write mediator schedule: %v", err)
		return err
	}

	return nil
}

// 重写配置文件，修改配置的的mediator的地址和密码
// @author Albert·Gou
func modifyMediatorInConf(configPath, password string, address common.Address) error {
	cfg := new(FullConfig)

	// 加载配置文件中的配置信息到 cfg中
	err := loadConfig(configPath, cfg)
	if err != nil {
		utils.Fatalf("%v", err)
		return err
	}

	cfg.MediatorPlugin.EnableStaleProduction = true
	cfg.MediatorPlugin.Mediators = []mp.MediatorConf{
		mp.MediatorConf{address.Str(), password,
			mp.DefaultInitPartSec, mp.DefaultInitPartPub},
	}

	err = makeConfigFile(cfg, configPath)
	if err != nil {
		utils.Fatalf("%v", err)
		return err
	}

	log.Debug(fmt.Sprintf("Rewriting config file at: %v", configPath))

	return nil
}
