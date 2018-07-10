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
	"os"
	"unsafe"

	"github.com/palletone/go-palletone/cmd/utils"
	"github.com/palletone/go-palletone/common/crypto"
	"github.com/palletone/go-palletone/common/hexutil"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/util"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/core/accounts"
	"github.com/palletone/go-palletone/core/accounts/keystore"
	"github.com/palletone/go-palletone/core/gen"
	"github.com/palletone/go-palletone/core/node"
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
			utils.LightModeFlag,
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

// initGenesis will initialise the given JSON format genesis file and writes it as
// the zero'd block (i.e. genesis) or will fail hard if it can't succeed.
func initGenesis(ctx *cli.Context) error {
	// Make sure we have a valid genesis JSON
	genesisPath := ctx.Args().First()
	if len(genesisPath) == 0 {
		utils.Fatalf("Must supply path to genesis JSON file")
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

	node := makeFullNode(ctx)
	txs, account, _ := getTransctions(ctx, node, genesis)
	unit, err := gen.SetupGenesisBlock(genesis, txs)
	if err != nil {
		utils.Fatalf("Failed to write genesis block: %v", err)
		return err
	}
	log.Info("Successfully Get Genesis Block")

	sig, err := sigGenesis(ctx, node, unit, account)
	if err != nil {
		utils.Fatalf("Failed to write genesis block: %v", err)
		return err
	}
	log.Info("Successfully SIG Genesis Block")
	return commitDB(sig)
}

func getTransctions(ctx *cli.Context, node *node.Node, genesis *core.Genesis) (modules.Transactions, accounts.Account, string) {
	ks := node.AccountManager().Backends(keystore.KeyStoreType)[0].(*keystore.KeyStore)
	addr := genesis.TokenHolder
	account, password := unlockAccount(ctx, ks, addr, 0, nil)
	//log.Info("======sigGenesis account:", account.Address.String())
	//log.Info("======sigGenesis password:", password)
	tx := &modules.Transaction{AccountNonce: 1}
	txs := []*modules.Transaction{}
	txs = append(txs, tx)
	return txs, account, password
}

func sigGenesis(ctx *cli.Context, node *node.Node, unit *modules.Unit, account accounts.Account) ([]byte, error) {
	//-------------

	ks := node.AccountManager().Backends(keystore.KeyStoreType)[0].(*keystore.KeyStore)
	hash := crypto.Keccak256Hash(util.RHashBytes(unit))
	//unit signature
	sign, err := ks.SignHash(account, hash.Bytes())
	if err != nil {
		utils.Fatalf("Sign error:%s", err)
		return []byte{}, err
	}
	strSign := hexutil.Encode(sign)
	return *(*[]byte)(unsafe.Pointer(&strSign)), nil
}

func commitDB(sign []byte) error {
	return nil
}
