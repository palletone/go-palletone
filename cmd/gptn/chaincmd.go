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
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	mp "github.com/palletone/go-palletone/consensus/mediatorplugin"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/core/accounts"
	"github.com/palletone/go-palletone/core/gen"
	"github.com/palletone/go-palletone/dag"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/modules"
	// "github.com/palletone/go-palletone/dag/txspool"
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

	Dbconn, err := node.OpenDatabase("leveldb", 0, 0)
	if err != nil {
		return errors.New("leveldb init failed")
	}
	filepath := node.ResolvePath("leveldb")
	dagconfig.DbPath = filepath
	dag, _ := dag.NewDag4GenesisInit(Dbconn)
	ks := node.GetKeyStore()
	// modify by Albert·Gou
	account, password := unlockAccount(nil, ks, genesis.TokenHolder, 0, nil)
	// 从配置文件中获取账户和密码
	//configPath := getConfigPath(ctx)
	//account, password := getAccountFromConf(configPath)

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
	//将Unit存入数据库中
	err = dag.SaveUnit4GenesisInit(unit, nil)
	if err != nil {
		fmt.Println("Save Genesis unit to db error:", err)
		return err
	}
	// @jay
	// asset 存入数据库中
	// dag.SaveCommon(key,asset)   key=[]byte(modules.FIELD_GENESIS_ASSET)
	//chainIndex := unit.UnitHeader.ChainIndex()
	//if err := dag.SaveChainIndex(chainIndex); err != nil {
	//	log.Info("save chain index is failed.", "error", err)
	//} else {
	token_info := modules.NewTokenInfo("ptncoin", "ptn", "creator_jay")
	idhex, _ := dag.SaveTokenInfo(token_info)
	log.Info("save chain index is success.", "idhex", idhex)
	//}

	genesisUnitHash := unit.UnitHash
	log.Info(fmt.Sprintf("Successfully Get Genesis Unit, it's hash: %v", genesisUnitHash.Hex()))

	// 2, 重写配置文件，修改当前节点的mediator的地址和密码
	// @author Albert·Gou
	//configPath := getConfigPath(ctx)
	//modifyMediatorInConf(configPath, password, account.Address)

	//3. initial globalproperty
	//modified by Yiran
	err = dag.InitPropertyDB(genesis, genesisUnitHash)
	if err != nil {
		utils.Fatalf("Failed toInitPropertyDB: %v", err)
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
	cfg.MediatorPlugin.Mediators = []*mp.MediatorConf{
		&mp.MediatorConf{address.Str(), password,
			mp.DefaultInitPrivKey, core.DefaultInitPubKey},
	}

	err = makeConfigFile(cfg, configPath)
	if err != nil {
		utils.Fatalf("%v", err)
		return err
	}

	log.Debug(fmt.Sprintf("Rewriting config file at: %v", configPath))

	return nil
}
