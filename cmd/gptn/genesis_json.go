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
	"strings"

	"github.com/palletone/go-palletone/cmd/console"
	"github.com/palletone/go-palletone/cmd/utils"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/files"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/configure"
	mp "github.com/palletone/go-palletone/consensus/mediatorplugin"
	"github.com/palletone/go-palletone/contracts/syscontract"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/core/gen"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/errors"
	"gopkg.in/urfave/cli.v1"
)

const defaultGenesisJsonPath = "./ptn-genesis.json"

var (
	GenesisTimestampFlag = cli.Int64Flag{
		Name:  "genesistime",
		Usage: "Replace timestamp from genesis.json with current time plus this many seconds (experts only!)",
		//		Value: 0,
	}

	GenesisJsonPathFlag = cli.StringFlag{
		Name:  "genesispath",
		Usage: "Path to create a Genesis State at.",
		Value: "", //defaultGenesisJsonPath,
	}

	createGenesisJsonCommand = cli.Command{
		Action:    utils.MigrateFlags(createGenesisJson),
		Name:      "newgenesis",
		Usage:     "Create a genesis json file template",
		ArgsUsage: "<genesisJsonPath> <openStdout>",
		Flags: []cli.Flag{
			GenesisJsonPathFlag,
			utils.CryptoLibFlag,
		},
		Category: "BLOCKCHAIN COMMANDS",
		Description: `
Create a json file for the genesis state of a new chain 
with an existing account or a newly created account.

If a well-formed JSON file exists at the path, 
it will be replaced with an example Genesis State.`,
	}

	dumpJsonCommand = cli.Command{
		Action:    utils.MigrateFlags(dumpJson),
		Name:      "dumpjson",
		Usage:     "Dumps genesis json to a specified file",
		ArgsUsage: "<jsonFilePath>",
		Flags: []cli.Flag{
			GenesisJsonPathFlag,
		},
		Category:    "MISCELLANEOUS COMMANDS",
		Description: `The dumpjson command dumps genesis json to a specified file.`,
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
		if err != nil || len(account) == 0 || !common.IsValidAddress(account) {
			errStr := "Invalid Token Account Address!"
			utils.Fatalf(errStr)
			return "", errors.New(errStr)
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

func createExampleMediators(ctx *cli.Context, mcLen int) []*mp.MediatorConf {
	exampleMediators := make([]*mp.MediatorConf, mcLen)
	for i := 0; i < mcLen; i++ {
		account, password, _ := createExampleAccount(ctx)
		secStr, pubStr := core.CreateInitDKS()

		exampleMediators[i] = &mp.MediatorConf{
			Address:     account,
			Password:    password,
			InitPrivKey: secStr,
			InitPubKey:  pubStr,
		}
	}

	return exampleMediators
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

	mcs := createExampleMediators(ctx, core.DefaultActiveMediatorCount)
	nodeStr /*, err*/ := getNodeInfo(ctx)
	//if err != nil {
	//	return err
	//}

	genesisState := createExampleGenesis()
	genesisState.TokenHolder = account
	genesisState.InitialParameters.FoundationAddress = account
	genesisState.DigitalIdentityConfig.RootCAHolder = account
	genesisState.ImmutableParameters.MinMaintSkipSlots = 2

	genesisState.InitialParameters.MediatorInterval = 3
	genesisState.InitialTimestamp = genesisState.InitialTimestamp / 3 * 3
	genesisState.InitialParameters.MaintenanceSkipSlots = 2
	genesisState.InitialMediatorCandidates = initialMediatorCandidates(mcs, nodeStr)

	//配置测试的基金会地址及密码
	//account, _, err = createExampleAccount(ctx)
	//if err != nil {
	//	return err
	//}
	//genesisState.SystemConfig.FoundationAddress = account

	var genesisJson []byte
	genesisJson, err = json.MarshalIndent(genesisState, "", "  ")
	if err != nil {
		utils.Fatalf("%v", err)
		return err
	}

	genesisOut := getGenesisPath(ctx)

	err = os.MkdirAll(filepath.Dir(genesisOut), os.ModePerm)
	if err != nil {
		utils.Fatalf("%v", err)
		return err
	}

	genesisFile, err = os.Create(genesisOut)
	if err != nil {
		utils.Fatalf("%v", err)
		return err
	}
	defer genesisFile.Close()

	_, err = genesisFile.Write(genesisJson)
	if err != nil {
		utils.Fatalf("%v", err)
		return err
	}

	fmt.Println("Creating example genesis state in file: " + genesisOut)

	// 修改本节点的一些特殊配置
	modifyConfig(ctx, mcs)

	return nil
}
func initSysContracts() []core.SysContract {
	list := make([]core.SysContract, 0)
	list = append(list, core.SysContract{Address: syscontract.CreateTokenContractAddress, Name: "PRC20", Active: true})
	list = append(list, core.SysContract{Address: syscontract.CreateToken721ContractAddress,
		Name: "PRC721", Active: true})
	list = append(list, core.SysContract{Address: syscontract.SysConfigContractAddress,
		Name: "System Config Manager", Active: true})
	list = append(list, core.SysContract{Address: syscontract.PartitionContractAddress,
		Name: "Partition Manager", Active: true})
	list = append(list, core.SysContract{Address: syscontract.DepositContractAddress,
		Name: "Deposit Manager", Active: true})
	list = append(list, core.SysContract{Address: syscontract.DigitalIdentityContractAddress,
		Name: "Digital Identity", Active: true})
	list = append(list, core.SysContract{Address: syscontract.VoteTokenContractAddress, Name: "Vote", Active: true})
	list = append(list, core.SysContract{Address: syscontract.TestContractAddress, Name: "Test", Active: true})

	return list
}
func modifyConfig(ctx *cli.Context, mediators []*mp.MediatorConf) error {
	cfg := new(FullConfig)
	configPath := getConfigPath(ctx)

	// 加载配置文件中的配置信息到 cfg中
	err := loadConfig(configPath, cfg)
	if err != nil {
		utils.Fatalf("%v", err)
		return err
	}

	// 修改本节点中mediator的一些特殊配置
	cfg.MediatorPlugin.EnableProducing = true
	cfg.MediatorPlugin.EnableStaleProduction = true
	cfg.MediatorPlugin.EnableConsecutiveProduction = true
	cfg.MediatorPlugin.RequiredParticipation = 0
	cfg.MediatorPlugin.EnableGroupSigning = true
	cfg.MediatorPlugin.Mediators = mediators

	// change log
	isOpenSTD := ctx.Args().Get(1)
	isOpenSTD = strings.ToUpper(isOpenSTD)
	if isOpenSTD == "FALSE" {
		newLogOutPath := make([]string, 0)
		for _, p := range cfg.Log.OutputPaths {
			if p == log.LogStdout {
				continue
			}
			newLogOutPath = append(newLogOutPath, p)
		}
		cfg.Log.OutputPaths = newLogOutPath
	}
	// 修改默认的Jury配置
	if len(mediators) > 0 {
		cfg.Jury.Accounts[0].Address = mediators[0].Address
		cfg.Jury.Accounts[0].Password = mediators[0].Password
	}

	err = makeConfigFile(cfg, configPath)
	if err != nil {
		utils.Fatalf("%v", err)
		return err
	}

	fmt.Println("Rewriting config file at: ", configPath)

	return nil
}

func getGenesisPath(ctx *cli.Context) string {
	genesisOut := ctx.Args().First()

	// If no path is specified, the default path is used
	if len(genesisOut) == 0 {
		// utils.Fatalf("Must supply path to genesis JSON file")
		genesisOut = defaultGenesisJsonPath
	}

	if files.IsDir(genesisOut) {
		genesisOut = filepath.Join(genesisOut, filepath.Base(defaultGenesisJsonPath))
	}

	return common.GetAbsPath(genesisOut)
}

// initialAccount, create a initial account for a new account
func initialAccount(ctx *cli.Context) (string, error) {
	address, err := newAccount(ctx)
	if err != nil {
		utils.Fatalf("%v", err)
		return address.Str(), err
	}

	fmt.Printf("Initial token holder's account address: %s\n", address.String())

	return address.Str(), nil
}

func createExampleAccount(ctx *cli.Context) (addrStr, password string, err error) {
	password = mp.DefaultPassword
	address, err := createAccount(ctx, password)
	addrStr = address.Str()
	return
}

// createExampleGenesis, create the genesis state of new chain with the specified account
func createExampleGenesis() *core.Genesis {
	//SystemConfig := core.SystemConfig{
	//	DepositRate:               core.DefaultDepositRate,
	//	TxCoinYearRate:            core.DefaultTxCoinYearRate,
	//	GenerateUnitReward:        core.DefaultGenerateUnitReward,
	//	RewardHeight:              core.DefaultRewardHeight,
	//	FoundationAddress:         core.DefaultFoundationAddress,
	//	DepositAmountForMediator:  core.DefaultDepositAmountForMediator,
	//	DepositAmountForJury:      core.DefaultDepositAmountForJury,
	//	DepositAmountForDeveloper: core.DefaultDepositAmountForDeveloper,
	//	DepositPeriod:             core.DefaultDepositPeriod,
	//	UccMemory:                 core.DefaultUccMemory,
	//	UccMemorySwap:             core.DefaultUccMemorySwap,
	//	UccCpuShares:              core.DefaultUccCpuShares,
	//	UccCpuPeriod:              core.DefaultCpuPeriod,
	//	UccCpuQuota:               core.DefaultUccCpuQuota,
	//	UccCpuSetCpus:             core.DefaultUccCpuSetCpus,
	//	TempUccMemory:             core.DefaultTempUccMemory,
	//	TempUccMemorySwap:         core.DefaultTempUccMemorySwap,
	//	TempUccCpuShares:          core.DefaultTempUccCpuShares,
	//	TempUccCpuQuota:           core.DefaultTempUccCpuQuota,
	//	ContractSignatureNum:      core.DefaultContractSignatureNum,
	//	ContractElectionNum:       core.DefaultContractElectionNum,
	//
	//	ActiveMediatorCount: strconv.FormatUint(core.DefaultMediatorCount, 10),
	//}

	//DigitalIdentityConfig := core.DigitalIdentityConfig{
	//	// default root ca holder, 默认是基金会地址
	//	RootCAHolder: core.DefaultFoundationAddress,
	//	RootCABytes:  core.DefaultRootCABytes,
	//}

	initParams := core.NewChainParams()
	mediators := []*mp.MediatorConf{mp.DefaultMediatorConf()}

	return &core.Genesis{
		Version:     configure.Version,
		GasToken:    dagconfig.DefaultToken, //core.DefaultAlias,
		TokenAmount: core.DefaultTokenAmount,
		//TokenDecimal:              core.DefaultTokenDecimal,
		ChainID:     core.DefaultChainID,
		TokenHolder: core.DefaultTokenHolder,
		Text:        core.DefaultText,
		//SystemConfig:          SystemConfig,
		DigitalIdentityConfig: core.DefaultDigitalIdentityConfig(),
		ParentUnitHeight:      -1,
		InitialParameters:     initParams,
		ImmutableParameters:   core.NewImmutChainParams(),
		InitialTimestamp:      gen.InitialTimestamp(initParams.MediatorInterval),
		//InitialActiveMediators:    core.DefaultMediatorCount,
		InitialMediatorCandidates: initialMediatorCandidates(mediators, core.DefaultNodeInfo),
		SystemContracts:           initSysContracts(),
	}
}

func initialMediatorCandidates(mediators []*mp.MediatorConf, nodeInfo string) []*core.InitialMediator {
	mcLen := len(mediators)
	initialMediators := make([]*core.InitialMediator, mcLen)
	for i := 0; i < mcLen; i++ {
		im := core.NewInitialMediator()
		im.AddStr = mediators[i].Address
		im.RewardAdd = mediators[i].Address
		im.InitPubKey = mediators[i].InitPubKey
		im.Node = nodeInfo
		initialMediators[i] = im
	}

	return initialMediators
}

func dumpJson(ctx *cli.Context) error {
	genesis := createExampleGenesis()

	genesisJson, err := json.MarshalIndent(*genesis, "", "  ")
	if err != nil {
		utils.Fatalf("%v", err)
		return err
	}

	filePath := getGenesisPath(ctx)

	err = os.MkdirAll(filepath.Dir(filePath), os.ModePerm)
	if err != nil {
		utils.Fatalf("%v", err)
		return err
	}

	file, err1 := os.Create(filePath)
	if err1 != nil {
		utils.Fatalf("%v", err1)
		return err1
	}
	defer file.Close()

	_, err = file.Write(genesisJson)
	if err != nil {
		utils.Fatalf("%v", err)
		return err
	}

	fmt.Println("Creating example genesis state in file: " + filePath)
	return nil
}
