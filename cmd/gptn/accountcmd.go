// Copyright 2016 The go-ethereum Authors
// This file is part of go-palletone.
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
	"fmt"
	"gopkg.in/urfave/cli.v1"
	"strings"

	"github.com/palletone/go-palletone/cmd/console"
	"github.com/palletone/go-palletone/cmd/utils"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/crypto"
	"github.com/palletone/go-palletone/common/hexutil"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/core/accounts"
	"github.com/palletone/go-palletone/core/accounts/keystore"
	"github.com/palletone/go-palletone/internal/ptnapi"
	"github.com/palletone/go-palletone/ptnjson"
	//"github.com/btcsuite/btcd/btcjson"
	"github.com/shopspring/decimal"
)

var (
	// 	walletCommand = cli.Command{
	// 		Name:      "wallet",
	// 		Usage:     "Manage PalletOne presale wallets",
	// 		ArgsUsage: "",
	// 		Category:  "ACCOUNT COMMANDS",
	// 		Description: `
	//     gptn wallet import /path/to/my/presale.wallet

	// will prompt for your password and imports your ether presale account.
	// It can be used non-interactively with the --password option taking a
	// passwordfile as argument containing the wallet password in plaintext.`,
	// 		Subcommands: []cli.Command{
	// 			{

	// 				Name:      "import",
	// 				Usage:     "Import PalletOne presale wallet",
	// 				ArgsUsage: "<keyFile>",
	// 				Action:    utils.MigrateFlags(importWallet),
	// 				Category:  "ACCOUNT COMMANDS",
	// 				Flags: []cli.Flag{
	// 					utils.DataDirFlag,
	// 					utils.KeyStoreDirFlag,
	// 					utils.PasswordFileFlag,
	// 					utils.LightKDFFlag,
	// 				},
	// 				Description: `
	// 	gptn wallet [options] /path/to/my/presale.wallet

	// will prompt for your password and imports your ether presale account.
	// It can be used non-interactively with the --password option taking a
	// passwordfile as argument containing the wallet password in plaintext.`,
	// 			},
	// 		},
	// 	}

	accountCommand = cli.Command{
		Name:     "account",
		Usage:    "Manage accounts",
		Category: "ACCOUNT COMMANDS",
		Description: `

Manage accounts, list all existing accounts, import a private key into a new
account, create a new account or update an existing account.

It supports interactive mode, when you are prompted for password as well as
non-interactive mode where passwords are supplied via a given password file.
Non-interactive mode is only meant for scripted use on test networks or known
safe environments.

Make sure you remember the password you gave when creating a new account (with
either new or import). Without it you are not able to unlock your account.

Note that exporting your key in unencrypted format is NOT supported.

Keys are stored under <DATADIR>/keystore.
It is safe to transfer the entire directory or the individual keys therein
between ethereum nodes by simply copying.

Make sure you backup your keys regularly.`,
		Subcommands: []cli.Command{
			{
				Name:   "list",
				Usage:  "Print summary of existing accounts",
				Action: utils.MigrateFlags(accountList),
				Flags: []cli.Flag{
					utils.DataDirFlag,
					utils.KeyStoreDirFlag,
				},
				Description: `
Print a short summary of all accounts`,
			},
			{
				Name:   "new",
				Usage:  "Create a new account",
				Action: utils.MigrateFlags(accountCreate),
				Flags: []cli.Flag{
					utils.DataDirFlag,
					utils.KeyStoreDirFlag,
					utils.PasswordFileFlag,
					utils.LightKDFFlag,
				},
				Description: `
    gptn account new

Creates a new account and prints the address.

The account is saved in encrypted format, you are prompted for a passphrase.

You must remember this passphrase to unlock your account in the future.

For non-interactive use the passphrase can be specified with the --password flag:

Note, this is meant to be used for testing only, it is a bad idea to save your
password to file or expose in any other way.
`,
			},
			{
				Name:      "update",
				Usage:     "Update an existing account",
				Action:    utils.MigrateFlags(accountUpdate),
				ArgsUsage: "<address>",
				Flags: []cli.Flag{
					utils.DataDirFlag,
					utils.KeyStoreDirFlag,
					utils.LightKDFFlag,
				},
				Description: `
    gptn account update <address>

Update an existing account.

The account is saved in the newest version in encrypted format, you are prompted
for a passphrase to unlock the account and another to save the updated file.

This same command can therefore be used to migrate an account of a deprecated
format to the newest format or change the password for an account.

For non-interactive use the passphrase can be specified with the --password flag:

    gptn account update [options] <address>

Since only one password can be given, only format update can be performed,
changing your password is only possible interactively.
`,
			},

			{
				Name:      "sign",
				Usage:     "sign a string",
				Action:    utils.MigrateFlags(accountSignString),
				ArgsUsage: "<address> <string>",
				Flags: []cli.Flag{
					utils.DataDirFlag,
					utils.KeyStoreDirFlag,
					utils.PasswordFileFlag,
				},
				Description: `
    gptn account sign <address> <string>
Sign a text by one account and return Signature.
`,
			},
			{
				Name:      "verify",
				Usage:     "verify a signature",
				Action:    utils.MigrateFlags(accountSignVerify),
				ArgsUsage: "<address> <message> <signature>",
				Flags: []cli.Flag{
					utils.DataDirFlag,
					utils.KeyStoreDirFlag,
					utils.PasswordFileFlag,
				},
				Description: `
    gptn account verify <address> <message> <signature>
verify the message signature.
`,
			},
			{
				Name:      "dumpkey",
				Usage:     "Dump the private key",
				Action:    utils.MigrateFlags(accountDumpKey),
				ArgsUsage: "<address>",
				Flags: []cli.Flag{
					utils.DataDirFlag,
					utils.KeyStoreDirFlag,
					utils.PasswordFileFlag,
				},
				Description: `
    gptn account dumpkey <address>
	Dump the private key.
`,
			},
			{
				Name:      "createtx",
				Usage:     "createtx sendfrom sendto ptncount",
				Action:    utils.MigrateFlags(accountCreateTx),
				ArgsUsage: "createtx <fromaddress> <toaddress> ptncount",
				Flags: []cli.Flag{
					utils.DataDirFlag,
					utils.KeyStoreDirFlag,
					utils.PasswordFileFlag,
				},
				Description: `
    gptn account createtx <fromaddress> <toaddress> count
	Dump the private key.
`,
			},
			{
				Name:      "signtx",
				Usage:     "signtx a raw transcaction",
				Action:    utils.MigrateFlags(accountSignTx),
				ArgsUsage: "signtx <rawtx> <script> privkey",
				Flags: []cli.Flag{
					utils.DataDirFlag,
					utils.KeyStoreDirFlag,
					utils.PasswordFileFlag,
				},
				Description: `
    gptn account createtx <fromaddress> <toaddress> count
	Dump the private key.
`,
			},
			{
				Name:   "import",
				Usage:  "Import a private key into a new account",
				Action: utils.MigrateFlags(accountImport),
				Flags: []cli.Flag{
					utils.DataDirFlag,
					utils.KeyStoreDirFlag,
					utils.PasswordFileFlag,
					utils.LightKDFFlag,
				},
				ArgsUsage: "<keyFile>",
				Description: `
    gptn account import <keyfile>

Imports an unencrypted private key from <keyfile> and creates a new account.
Prints the address.

The keyfile is assumed to contain an unencrypted private key in hexadecimal format.

The account is saved in encrypted format, you are prompted for a passphrase.

You must remember this passphrase to unlock your account in the future.

For non-interactive use the passphrase can be specified with the -password flag:

    gptn account import [options] <keyfile>

Note:
As you can directly copy your encrypted accounts to another ethereum instance,
this import mechanism is not needed when you transfer an account between
nodes.
`,
			},
		},
	}
)

func accountList(ctx *cli.Context) error {
	stack, _ := makeConfigNode(ctx)
	var index int
	for _, wallet := range stack.AccountManager().Wallets() {
		for _, account := range wallet.Accounts() {
			fmt.Printf("Account #%d: {%s} %s\n", index, account.Address.Str(), &account.URL)
			index++
		}
	}
	return nil
}

// tries unlocking the specified account a few times.
func unlockAccount(ctx *cli.Context, ks *keystore.KeyStore, address string, i int, passwords []string) (accounts.Account, string) {
	account, err := utils.MakeAddress(ks, address)
	if err != nil {
		utils.Fatalf("Could not list accounts: %v", err)
	}
	for trials := 0; trials < 3; trials++ {
		prompt := fmt.Sprintf("Unlocking account %s | Attempt %d/%d", address, trials+1, 3)
		password := getPassPhrase(prompt, false, i, passwords)
		err = ks.Unlock(account, password)
		if err == nil {
			log.Info("Unlocked account", "address", account.Address.Str())
			return account, password
		}
		if err, ok := err.(*keystore.AmbiguousAddrError); ok {
			log.Info("Unlocked account", "address", account.Address.Str())
			return ambiguousAddrRecovery(ks, err, password), password
		}
		if err != keystore.ErrDecrypt {
			// No need to prompt again if the error is not decryption-related.
			log.Info("Unlocked account err:", err.Error())
			break
		}
	}
	// All trials expended to unlock account, bail out
	utils.Fatalf("Failed to unlock account %s (%v)", address, err)

	return accounts.Account{}, ""
}

// getPassPhrase retrieves the password associated with an account, either fetched
// from a list of preloaded passphrases, or requested interactively from the user.
func getPassPhrase(prompt string, confirmation bool, i int, passwords []string) string {
	// If a list of passwords was supplied, retrieve from them
	if len(passwords) > 0 {
		if i < len(passwords) {
			return passwords[i]
		}
		return passwords[len(passwords)-1]
	}
	// Otherwise prompt the user for the password
	if prompt != "" {
		fmt.Println(prompt)
	}
	password, err := console.Stdin.PromptPassword("Passphrase: ")
	if err != nil {
		utils.Fatalf("Failed to read passphrase: %v", err)
	}
	if confirmation {
		confirm, err := console.Stdin.PromptPassword("Repeat passphrase: ")
		if err != nil {
			utils.Fatalf("Failed to read passphrase confirmation: %v", err)
		}
		if password != confirm {
			utils.Fatalf("Passphrases do not match")
		}
	}
	return password
}

func ambiguousAddrRecovery(ks *keystore.KeyStore, err *keystore.AmbiguousAddrError, auth string) accounts.Account {
	fmt.Printf("Multiple key files exist for address %x:\n", err.Addr)
	for _, a := range err.Matches {
		fmt.Println("  ", a.URL)
	}
	fmt.Println("Testing your passphrase against all of them...")
	var match *accounts.Account
	for _, a := range err.Matches {
		if err := ks.Unlock(a, auth); err == nil {
			match = &a
			break
		}
	}
	if match == nil {
		utils.Fatalf("None of the listed files could be unlocked.")
	}
	fmt.Printf("Your passphrase unlocked %s\n", match.URL)
	fmt.Println("In order to avoid this warning, you need to remove the following duplicate key files:")
	for _, a := range err.Matches {
		if a != *match {
			fmt.Println("  ", a.URL)
		}
	}
	return *match
}

// accountCreate creates a new account into the keystore defined by the CLI flags.
func createAccount(ctx *cli.Context, password string) (common.Address, error) {
	cfg := FullConfig{Node: defaultNodeConfig()}
	// Load config file.
	if err := maybeLoadConfig(ctx, &cfg); err != nil {
		utils.Fatalf("%v", err)
	}

	utils.SetNodeConfig(ctx, &cfg.Node)
	scryptN, scryptP, keydir, err := cfg.Node.AccountConfig()

	address, err := keystore.StoreKey(keydir, password, scryptN, scryptP)
	if err != nil {
		utils.Fatalf("Failed to create account: %v", err)
	}

	return address, nil
}

func newAccount(ctx *cli.Context) (common.Address, error) {
	password := getPassPhrase("Your new account is locked with a password. Please give a password. "+
		"Do not forget this password.", true, 0, utils.MakePasswordList(ctx))

	address, _ := createAccount(ctx, password)

	return address, nil
}

// accountCreate creates a new account into the keystore defined by the CLI flags.
func accountCreate(ctx *cli.Context) error {
	address, err := newAccount(ctx)
	if err != nil {
		utils.Fatalf("%v", err)
	}

	//	fmt.Printf("Address Hex: {%x}\n", address)
	fmt.Printf("Address: %s\n", address.String())
	return nil
}

// accountUpdate transitions an account from a previous format to the current
// one, also providing the possibility to change the pass-phrase.
func accountUpdate(ctx *cli.Context) error {
	if len(ctx.Args()) == 0 {
		utils.Fatalf("No accounts specified to update")
	}
	stack, _ := makeConfigNode(ctx)
	ks := stack.GetKeyStore()

	for _, addr := range ctx.Args() {
		account, oldPassword := unlockAccount(ctx, ks, addr, 0, nil)
		newPassword := getPassPhrase("Please give a new password. Do not forget this password.", true, 0, nil)
		if err := ks.Update(account, oldPassword, newPassword); err != nil {
			utils.Fatalf("Could not update the account: %v", err)
		}
	}
	return nil
}
func accountSignString(ctx *cli.Context) error {
	if len(ctx.Args()) == 0 {
		utils.Fatalf("No accounts specified to update")
	}

	stack, _ := makeConfigNode(ctx)
	ks := stack.GetKeyStore()
	addr := ctx.Args().First()
	account, _ := utils.MakeAddress(ks, addr)
	hash := crypto.Keccak256Hash([]byte(ctx.Args()[1]))
	fmt.Printf("%s Hash:%s", addr, hash.String())
	pwd := getPassPhrase("Please give a password to unlock your account", false, 0, utils.MakePasswordList(ctx))
	sign, err := ks.SignHashWithPassphrase(account, pwd, hash.Bytes())
	if err != nil {
		utils.Fatalf("Sign error:%s", err)
	}
	fmt.Println("Signature: " + hexutil.Encode(sign))
	return nil
}
func accountDumpKey(ctx *cli.Context) error {
	if len(ctx.Args()) == 0 {
		utils.Fatalf("No accounts specified to update")
	}
	stack, _ := makeConfigNode(ctx)
	ks := stack.GetKeyStore()
	addr := ctx.Args().First()
	account, _ := utils.MakeAddress(ks, addr)
	pwd := getPassPhrase("Please give a password to unlock your account", false, 0, nil)
	prvKey, _ := ks.DumpKey(account, pwd)
	wif := crypto.ToWIF(prvKey)
	fmt.Printf("Your private key hex is : {%x}, WIF is {%s}\n", prvKey, wif)
	pK, _ := crypto.ToECDSA(prvKey)
	pubBytes := crypto.CompressPubkey(&pK.PublicKey)
	fmt.Printf("Compressed public key hex is {%x}", pubBytes)
	return nil
}

func accountSignVerify(ctx *cli.Context) error {
	if len(ctx.Args()) == 0 {
		utils.Fatalf("No accounts specified to update")
	}

	stack, _ := makeConfigNode(ctx)
	ks := stack.GetKeyStore()
	addr := ctx.Args().First()
	account, _ := utils.MakeAddress(ks, addr)

	hash := crypto.Keccak256Hash([]byte(ctx.Args()[1]))
	sign := ctx.Args()[2]
	fmt.Printf("\n%s Hash:%s\n", addr, hash.String())
	pwd := getPassPhrase("Please give a password to unlock your account", false, 0, utils.MakePasswordList(ctx))
	s, _ := hexutil.Decode(sign)
	// ss, _ := ks.SignHashWithPassphrase(account, pwd, hash.Bytes())
	// fmt.Println("Sign again:" + hexutil.Encode(ss))

	pass, err := ks.VerifySignatureWithPassphrase(account, pwd, hash.Bytes(), s)
	if err != nil {
		utils.Fatalf("Verfiy error:%s", err)
	}
	if pass {
		fmt.Println("Valid signature")
	} else {
		utils.Fatalf("Invalid signature")
	}
	return nil
}

// func importWallet(ctx *cli.Context) error {
// 	keyfile := ctx.Args().First()
// 	if len(keyfile) == 0 {
// 		utils.Fatalf("keyfile must be given as argument")
// 	}
// 	keyJson, err := ioutil.ReadFile(keyfile)
// 	if err != nil {
// 		utils.Fatalf("Could not read wallet file: %v", err)
// 	}

// 	stack, _ := makeConfigNode(ctx)
// 	passphrase := getPassPhrase("", false, 0, utils.MakePasswordList(ctx))

// 	ks := stack.AccountManager().Backends(keystore.KeyStoreType)[0].(*keystore.KeyStore)
// 	acct, err := ks.ImportPreSaleKey(keyJson, passphrase)
// 	if err != nil {
// 		utils.Fatalf("%v", err)
// 	}
// 	fmt.Printf("Address: {%x}\n", acct.Address)
// 	return nil
// }
//add by wzhyuan
//add by wzhyuan
type RawTransactionGenParams struct {
	Inputs []struct {
		Txid         string `json:"txid"`
		Vout         uint32 `json:"vout"`
		MessageIndex uint32 `json:"messageindex"`
	} `json:"inputs"`
	Outputs []struct {
		Address string          `json:"address"`
		Amount  decimal.Decimal `json:"amount"`
	} `json:"outputs"`
	Locktime int64 `json:"locktime"`
}
type RawTransactionGenResult struct {
	Rawtx string `json:"rawtx"`
}

type SignTransactionParams struct {
	RawTx  string `json:"rawtx"`
	Inputs []struct {
		Txid         string `json:"txid"`
		Vout         uint32 `json:"vout"`
		MessageIndex uint32 `json:"messageindex"`
		ScriptPubKey string `json:"scriptPubKey"`
		RedeemScript string `json:"redeemScript"`
	} `json:"rawtxinput"`
	PrivKeys []string `json:"privkeys"`
	Flags    string   `jsonrpcdefault:"\"ALL\""`
}

//type SignTransactionResult struct {
//	TransactionHex string `json:"transactionhex"`
//	Complete       bool   `json:"complete"`
//}

func accountCreateTx(ctx *cli.Context) error {
	if len(ctx.Args()) == 0 {
		utils.Fatalf("No accounts specified to update")
	}
	if len(ctx.Args()) != 1 {
		utils.Fatalf("usage: [{'txid':txid,'vout':n},...] {address:amount,...}")
	}
	params := ctx.Args().First()
	var rawTransactionGenParams RawTransactionGenParams
	err := json.Unmarshal([]byte(params), &rawTransactionGenParams)
	if err != nil {
		return nil
	}
	//transaction inputs
	var inputs []ptnjson.TransactionInput
	for _, inputOne := range rawTransactionGenParams.Inputs {
		input := ptnjson.TransactionInput{inputOne.Txid, inputOne.Vout, inputOne.MessageIndex}
		inputs = append(inputs, input)
	}
	if len(inputs) == 0 {
		return nil
	}
	//realNet := &chaincfg.MainNetParams
	amounts := map[string]decimal.Decimal{}
	for _, outOne := range rawTransactionGenParams.Outputs {
		if len(outOne.Address) == 0 || outOne.Amount.LessThanOrEqual(decimal.New(0, 0)) {
			continue
		}
		amounts[outOne.Address] = (outOne.Amount)
	}
	if len(amounts) == 0 {
		return nil
	}
	arg := ptnjson.NewCreateRawTransactionCmd(inputs, amounts, &rawTransactionGenParams.Locktime)
	tx, err := ptnapi.CreateRawTransaction(arg)
	if err != nil {
		utils.Fatalf("Verfiy error:%s", err)
	}
	if tx != "" {
		fmt.Println("Create transcation success")
	} else {
		utils.Fatalf("Invalid create action")
	}
	return nil
}

func accountSignTx(ctx *cli.Context) error {
	if len(ctx.Args()) == 0 {
		utils.Fatalf("No accounts specified to update")
	}
	if len(ctx.Args()) != 1 {
		utils.Fatalf("usage :json: [{'txid':txid,'vout':n,'scriptPubKey':hex,'redeemScript':hex},...] [,...] " +
			"[uint32='ALL']")
	}
	params := ctx.Args().First()
	var signTransactionParams SignTransactionParams
	err := json.Unmarshal([]byte(params), &signTransactionParams)
	if err != nil {
		return nil
	}

	//check empty string
	if "" == signTransactionParams.RawTx {
		return nil
	}
	//transaction inputs
	var rawinputs []ptnjson.RawTxInput
	for _, inputOne := range signTransactionParams.Inputs {
		input := ptnjson.RawTxInput{inputOne.Txid, inputOne.Vout, inputOne.MessageIndex, inputOne.ScriptPubKey, inputOne.RedeemScript}
		rawinputs = append(rawinputs, input)
	}
	if len(rawinputs) == 0 {
		return nil
	}
	var keys []string
	for _, key := range signTransactionParams.PrivKeys {
		key = strings.TrimSpace(key) //Trim whitespace
		if len(key) == 0 {
			continue
		}
		keys = append(keys, key)
	}
	if len(keys) == 0 {
		return nil
	}
	//	// All returned errors (not OOM, which panics) encounted during
	//	// bytes.Buffer writes are unexpected.
	send_args := ptnjson.NewSignRawTransactionCmd(signTransactionParams.RawTx, &rawinputs, &keys, nil)
	send_args = send_args
	//signtxout, err := ptnapi.SignRawTransaction(send_args)
	//if signtxout == nil {
	//	utils.Fatalf("Invalid signature")
	//}
	//signtx := signtxout.(ptnjson.SignRawTransactionResult)
	//if err != nil {
	//	utils.Fatalf("signtx error:%s", err)
	//}
	//if signtx.Complete == true {
	//	fmt.Println("Signature success")
	//	fmt.Println(signtx.Hex)
	//} else {
	//	utils.Fatalf("Invalid signature")
	//}
	return nil
}
func accountImport(ctx *cli.Context) error {
	keyfile := ctx.Args().First()
	if len(keyfile) == 0 {
		utils.Fatalf("keyfile must be given as argument")
	}
	key, err := crypto.LoadECDSA(keyfile)
	if err != nil {
		utils.Fatalf("Failed to load the private key: %v", err)
	}
	stack, _ := makeConfigNode(ctx)
	passphrase := getPassPhrase("Your new account is locked with a password. Please give a password. Do not forget this password.", true, 0, utils.MakePasswordList(ctx))

	ks := stack.GetKeyStore()
	acct, err := ks.ImportECDSA(key, passphrase)
	if err != nil {
		utils.Fatalf("Could not create the account: %v", err)
	}
	fmt.Printf("Address: {%s}\n", acct.Address)
	return nil
}
