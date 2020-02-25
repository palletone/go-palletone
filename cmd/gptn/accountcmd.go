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
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/palletone/go-palletone/cmd/console"
	"github.com/palletone/go-palletone/cmd/utils"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/crypto"
	"github.com/palletone/go-palletone/common/hexutil"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/core/accounts"
	"github.com/palletone/go-palletone/core/accounts/keystore"
	"github.com/palletone/go-palletone/dag/errors"
	"github.com/palletone/go-palletone/internal/ptnapi"
	"github.com/palletone/go-palletone/ptnjson"
	"github.com/palletone/go-palletone/tokenengine"
	"gopkg.in/urfave/cli.v1"

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
				Name:   "convert",
				Usage:  "Convert account address to hex format address",
				Action: utils.MigrateFlags(accountConvert),
				Flags:  []cli.Flag{},
				Description: `
Convert account address to hex format address`,
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
				Name:   "newHd",
				Usage:  "Create a new HD account",
				Action: utils.MigrateFlags(hdAccountCreate),
				Flags: []cli.Flag{
					utils.DataDirFlag,
					utils.KeyStoreDirFlag,
					utils.PasswordFileFlag,
					utils.LightKDFFlag,
				},
				Description: `
    gptn account newHd
`,
			},
			{
				Name:   "getHdAccount",
				Usage:  "get HD account by account index",
				Action: utils.MigrateFlags(hdAccountGet),
				Flags: []cli.Flag{
					utils.DataDirFlag,
					utils.KeyStoreDirFlag,
					utils.PasswordFileFlag,
					utils.LightKDFFlag,
				},
				Description: `
    gptn account getHdAccount P1XXXXX accountIndex
`,
			},
			{
				Name:      "multi",
				Usage:     "Create a new multisign account",
				Action:    utils.MigrateFlags(accountMultiCreate),
				ArgsUsage: "<string>",
				Flags: []cli.Flag{
					utils.DataDirFlag,
					utils.KeyStoreDirFlag,
					utils.PasswordFileFlag,
					utils.LightKDFFlag,
				},
				Description: `
    gptn account multi pubKeyCount pubKey1 pubKey2 ... needSignCount

Creates a new multisign account and prints the address and redeemScript.

pubKeyCount must less than 15.
You can use "dumppubkey" command to get account public key.
needSignCount must less or equal pubKeyCount.
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
				ArgsUsage: "key hex data",
				Description: `
    gptn account import hex

Imports an unencrypted private key from hex and creates a new account.
Prints the address.

The account is saved in encrypted format, you are prompted for a passphrase.

You must remember this passphrase to unlock your account in the future.

For non-interactive use the passphrase can be specified with the -password flag:

    gptn account import [options] <key hex>

Note:
As you can directly copy your encrypted accounts to another ethereum instance,
this import mechanism is not needed when you transfer an account between
nodes.
`,
			},
			{
				Name:   "importHd",
				Usage:  "Import a mnemonic into a new HD wallet account",
				Action: utils.MigrateFlags(accountImportHd),
				Flags: []cli.Flag{
					utils.DataDirFlag,
					utils.KeyStoreDirFlag,
					utils.PasswordFileFlag,
					utils.LightKDFFlag,
				},
				ArgsUsage: "mnemonic string",
				Description: `
    gptn account importhd "mnemonic"
`,
			},

			{
				Name:      "dumppubkey",
				Usage:     "Dump the public key",
				Action:    utils.MigrateFlags(accountDumpPubKey),
				ArgsUsage: "<address>",
				Flags: []cli.Flag{
					utils.DataDirFlag,
					utils.KeyStoreDirFlag,
					utils.PasswordFileFlag,
				},
				Description: `
    gptn account dumppubkey <address>
    Dump the public key.
`,
			},
		},
	}
)

func accountList(ctx *cli.Context) error {
	stack, _ := makeConfigNode(ctx, false)
	var index int
	for _, wallet := range stack.AccountManager().Wallets() {
		for _, account := range wallet.Accounts() {
			fmt.Printf("Account #%d: {%s} %s\n", index, account.Address.Str(), &account.URL)
			index++
		}
	}
	return nil
}
func accountConvert(ctx *cli.Context) error {
	addrStr := ctx.Args().First()
	if len(addrStr) == 0 {
		utils.Fatalf("address must be given as argument")
	}
	addr, err := common.StringToAddress(addrStr)
	if err != nil {
		return err
	}
	hexAddr := hexutil.Encode(addr.Bytes())
	fmt.Printf("Account hex format: %s\n", hexAddr)
	return nil
}

// tries unlocking the specified account a few times.
func unlockAccount( /*ctx *cli.Context*/ ks *keystore.KeyStore, address string, i int,
	passwords []string) (accounts.Account, string) {
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
	var match accounts.Account
	for _, a := range err.Matches {
		if err := ks.Unlock(a, auth); err == nil {
			match = a
			break
		}
	}
	if match.Address.IsZero() {
		utils.Fatalf("None of the listed files could be unlocked.")
	}
	fmt.Printf("Your passphrase unlocked %s\n", match.URL)
	fmt.Println("In order to avoid this warning, you need to remove the following duplicate key files:")
	for _, a := range err.Matches {
		if a != match {
			fmt.Println("  ", a.URL)
		}
	}
	return match
}

// accountCreate creates a new account into the keystore defined by the CLI flags.
func createAccount(ctx *cli.Context, password string) (common.Address, error) {
	var err error
	var cfg FullConfig
	var configDir string
	// Load config file.
	if cfg, configDir, err = maybeLoadConfig(ctx); err != nil {
		utils.Fatalf("%v", err)
		return common.Address{}, err
	}

	cfg.Node.P2P = cfg.P2P
	utils.SetNodeConfig(ctx, &cfg.Node, configDir)
	scryptN, scryptP, keydir, _ := cfg.Node.AccountConfig()

	address, err := keystore.StoreKey(keydir, password, scryptN, scryptP)
	if err != nil {
		utils.Fatalf("Failed to create account: %v", err)
		return common.Address{}, err
	}

	return address, nil
}
func createHdAccount(ctx *cli.Context, password string) (common.Address, error) {
	var err error
	var cfg FullConfig
	var configDir string
	// Load config file.
	if cfg, configDir, err = maybeLoadConfig(ctx); err != nil {
		utils.Fatalf("%v", err)
		return common.Address{}, err
	}

	cfg.Node.P2P = cfg.P2P
	utils.SetNodeConfig(ctx, &cfg.Node, configDir)
	scryptN, scryptP, keydir, _ := cfg.Node.AccountConfig()

	address, mnemonic, err := keystore.StoreHdSeed(keydir, password, scryptN, scryptP)
	if err != nil {
		utils.Fatalf("Failed to create account: %v", err)
		return common.Address{}, err
	}
	fmt.Println("Please remember your mnemonic is: " + mnemonic)
	return address, nil
}

// accountCreate creates a new account into the keystore defined by the CLI flags.
func createMultiAccount(ctx *cli.Context, pubkey [][]byte, check int) ([]byte, []byte, common.Address, error) {
	var err error
	var cfg FullConfig
	var configDir string
	// Load config file.
	if cfg, configDir, err = maybeLoadConfig(ctx); err != nil {
		utils.Fatalf("%v", err)
		return []byte{}, []byte{}, common.Address{}, err
	}

	cfg.Node.P2P = cfg.P2P
	utils.SetNodeConfig(ctx, &cfg.Node, configDir)
	bcheck := IntToByte(int64(check))
	redeemScript := tokenengine.Instance.GenerateRedeemScript(bcheck[7], pubkey)
	lockScript := tokenengine.Instance.GenerateP2SHLockScript(crypto.Hash160(redeemScript))
	addressMulti, _ := tokenengine.Instance.GetAddressFromScript(lockScript)

	return lockScript, redeemScript, addressMulti, nil
}

func newAccount(ctx *cli.Context) (common.Address, error) {
	password := getPassPhrase("Your new account is locked with a password. Please give a password. "+
		"Do not forget this password.", true, 0, utils.MakePasswordList(ctx))

	address, err := createAccount(ctx, password)
	if err != nil {
		return common.Address{}, err
	}

	return address, nil
}
func newHdAccount(ctx *cli.Context) (common.Address, error) {
	password := getPassPhrase("Your new account is locked with a password. Please give a password. "+
		"Do not forget this password.", true, 0, utils.MakePasswordList(ctx))

	address, err := createHdAccount(ctx, password)
	if err != nil {
		return common.Address{}, err
	}

	return address, nil
}

func IntToByte(num int64) []byte {
	var buffer bytes.Buffer
	err := binary.Write(&buffer, binary.BigEndian, num)
	if err != nil {
		return []byte{}
	}
	return buffer.Bytes()
}

/*func BytesToInt(bys []byte) int {
    bytebuff := bytes.NewBuffer(bys)
    var data int64
    binary.Read(bytebuff, binary.BigEndian, &data)
    return int(data)
}*/
func newMultiAccount(ctx *cli.Context) ([]byte, []byte, common.Address, error) {
	if len(ctx.Args()) == 0 {
		utils.Fatalf("No pubkey specified to create multi account")
	}
	var pki []byte
	var pk [][]byte
	totalstring := ctx.Args().First()
	total, err := strconv.Atoi(totalstring)
	if err != nil || total < 0 || total > 15 {
		utils.Fatalf("Pubkey specified to create multi account cannot more than 15")
		return []byte{}, []byte{}, common.Address{}, err
	}
	for arg_s := 1; arg_s < total+1; arg_s++ {
		pki, err = hex.DecodeString(ctx.Args()[arg_s])
		if err != nil || total < 0 || total > 15 {
			utils.Fatalf("Pubkey specified to create multi account cannot more than 15")
			return []byte{}, []byte{}, common.Address{}, err
		}
		pk = append(pk, pki)
	}
	s_check := ctx.Args()[total+1]
	check, err := strconv.Atoi(s_check)
	if err != nil || check < 0 || check > 15 {
		utils.Fatalf("Pubkey specified to create multi account cannot more than 15")
		return []byte{}, []byte{}, common.Address{}, err
	}
	lockscript, redeemScript, address, err := createMultiAccount(ctx, pk, check)
	if err != nil {
		return []byte{}, []byte{}, common.Address{}, err
	}
	return lockscript, redeemScript, address, nil
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
func hdAccountCreate(ctx *cli.Context) error {
	address, err := newHdAccount(ctx)
	if err != nil {
		utils.Fatalf("%v", err)
	}

	//	fmt.Printf("Address Hex: {%x}\n", address)
	fmt.Printf("Address: %s\n", address.String())
	return nil
}
func hdAccountGet(ctx *cli.Context) error {
	if len(ctx.Args()) == 0 {
		utils.Fatalf("No accounts specified to update")
	}

	stack, _ := makeConfigNode(ctx, false)
	ks := stack.GetKeyStore()
	addr := ctx.Args().First()
	account, _ := utils.MakeAddress(ks, addr)
	pwd := getPassPhrase("Please give a password to unlock your account", false, 0, nil)

	userId := ctx.Args()[1]
	accountIndex, err := strconv.Atoi(userId)
	if err != nil {
		return errors.New("invalid argument, args 1 must be a number")
	}

	hdAccount, err := ks.GetHdAccountWithPassphrase(account, pwd, uint32(accountIndex))
	if err != nil {
		utils.Fatalf("get HD account error:%s", err)
	}
	fmt.Println(hdAccount.Address.String())
	return nil
}
func accountMultiCreate(ctx *cli.Context) error {
	lockscript, redeem, address, err := newMultiAccount(ctx)
	if err != nil {
		utils.Fatalf("%v", err)
	}

	//	fmt.Printf("Address Hex: {%x}\n", address)
	fmt.Printf("Address: %s lockscript:%x redeem : %x\n", address.String(), lockscript, redeem)
	return nil
}

// accountUpdate transitions an account from a previous format to the current
// one, also providing the possibility to change the pass-phrase.
func accountUpdate(ctx *cli.Context) error {
	if len(ctx.Args()) == 0 {
		utils.Fatalf("No accounts specified to update")
	}
	stack, _ := makeConfigNode(ctx, false)
	ks := stack.GetKeyStore()

	for _, addr := range ctx.Args() {
		account, oldPassword := unlockAccount(ks, addr, 0, nil)
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

	stack, _ := makeConfigNode(ctx, false)
	ks := stack.GetKeyStore()
	addr := ctx.Args().First()
	account, _ := utils.MakeAddress(ks, addr)
	data := []byte(ctx.Args()[1])
	fmt.Printf("%s Data:%#x", addr, data)
	pwd := getPassPhrase("Please give a password to unlock your account", false, 0, utils.MakePasswordList(ctx))
	sign, err := ks.SignMessageWithPassphrase(account, pwd, data)
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
	stack, _ := makeConfigNode(ctx, false)
	ks := stack.GetKeyStore()
	addr := ctx.Args().First()
	account, _ := utils.MakeAddress(ks, addr)
	pwd := getPassPhrase("Please give a password to unlock your account", false, 0, nil)
	prvKey, keyType, err := ks.DumpKey(account, pwd)
	if err != nil {
		return err
	}
	if keyType == keystore.KeyType_ECDSA_KEY {

		wif := crypto.ToWIF(prvKey)
		fmt.Printf("Your private key hex is : {%x}, WIF is {%s}\n", prvKey, wif)
		//pK, _ := crypto.ToECDSA(prvKey)
		pubBytes, _ := crypto.MyCryptoLib.PrivateKeyToPubKey(prvKey)
		fmt.Printf("Compressed public key hex is {%x}", pubBytes)
	}
	if keyType == keystore.KeyType_HD_Seed {
		fmt.Printf("Your HD wallet seed hex:%x\r\n", prvKey)
	}
	return nil
}

func accountDumpPubKey(ctx *cli.Context) error {
	if len(ctx.Args()) == 0 {
		utils.Fatalf("No accounts specified to dump public key")
	}
	stack, _ := makeConfigNode(ctx, false)
	ks := stack.GetKeyStore()
	addr := ctx.Args().First()
	account, _ := utils.MakeAddress(ks, addr)
	pwd := getPassPhrase("Please give a password to unlock your account", false, 0, nil)
	prvKey, keyType, _ := ks.DumpKey(account, pwd)
	if keyType == keystore.KeyType_ECDSA_KEY {
		b, _ := crypto.MyCryptoLib.PrivateKeyToPubKey(prvKey)
		fmt.Println(hex.EncodeToString(b))
	}
	if keyType == keystore.KeyType_HD_Seed {
		_, pubKey, err := keystore.NewAccountKey(prvKey, 0)
		if err != nil {
			return err
		}
		fmt.Println(hex.EncodeToString(pubKey))
	}
	return nil
}

func accountSignVerify(ctx *cli.Context) error {
	if len(ctx.Args()) == 0 {
		utils.Fatalf("No accounts specified to update")
	}

	stack, _ := makeConfigNode(ctx, false)
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

// 	stack, _ := makeConfigNode(ctx, false)
// 	passphrase := getPassPhrase("", false, 0, utils.MakePasswordList(ctx))

// 	ks := stack.GetKeyStore()
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

//type RawTransactionGenResult struct {
//	Rawtx string `json:"rawtx"`
//}

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
	inputs := make([]ptnjson.TransactionInput, 0, len(rawTransactionGenParams.Inputs))
	for _, inputOne := range rawTransactionGenParams.Inputs {
		input := ptnjson.TransactionInput{Txid: inputOne.Txid, Vout: inputOne.Vout, MessageIndex: inputOne.MessageIndex}
		inputs = append(inputs, input)
	}
	if len(inputs) == 0 {
		return nil
	}
	//realNet := &chaincfg.MainNetParams
	amounts := []ptnjson.AddressAmt{}
	for _, outOne := range rawTransactionGenParams.Outputs {
		if len(outOne.Address) == 0 || outOne.Amount.LessThanOrEqual(decimal.New(0, 0)) {
			continue
		}
		amounts = append(amounts, ptnjson.AddressAmt{Address: outOne.Address, Amount: outOne.Amount})
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
	rawinputs := make([]ptnjson.RawTxInput, 0, len(signTransactionParams.Inputs))
	for _, inputOne := range signTransactionParams.Inputs {
		input := ptnjson.RawTxInput{Txid: inputOne.Txid, Vout: inputOne.Vout, MessageIndex: inputOne.MessageIndex,
			ScriptPubKey: inputOne.ScriptPubKey, RedeemScript: inputOne.RedeemScript}
		rawinputs = append(rawinputs, input)
	}
	if len(rawinputs) == 0 {
		return nil
	}
	keys := make([]string, 0, len(signTransactionParams.PrivKeys))
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

	return nil
}
func accountImport(ctx *cli.Context) error {
	keyHex := ctx.Args().First()
	if len(keyHex) == 0 {
		utils.Fatalf("keyHex must be given as argument")
	}
	key, err := hexutil.Decode(keyHex)
	if err != nil {
		utils.Fatalf("Failed to load the private key: %v", err)
	}
	stack, _ := makeConfigNode(ctx, false)
	passphrase := getPassPhrase("Your new account is locked with a password. "+
		"Please give a password. Do not forget this password.", true, 0, utils.MakePasswordList(ctx))

	ks := stack.GetKeyStore()
	acct, err := ks.ImportECDSA(key, passphrase)
	if err != nil {
		utils.Fatalf("Could not create the account: %v", err)
	}
	fmt.Printf("Address: {%s}\n", acct.Address.String())
	return nil
}
func accountImportHd(ctx *cli.Context) error {
	mnemonic := ctx.Args().First()
	if len(mnemonic) == 0 {
		utils.Fatalf("mnemonic must be given as argument")
	}

	stack, _ := makeConfigNode(ctx, false)
	passphrase := getPassPhrase("Your new account is locked with a password. "+
		"Please give a password. Do not forget this password.", true, 0, utils.MakePasswordList(ctx))

	ks := stack.GetKeyStore()
	acct, err := ks.ImportHdSeedFromMnemonic(mnemonic, passphrase)
	if err != nil {
		utils.Fatalf("Could not create the account: %v", err)
	}
	fmt.Printf("HD wallet account0 address: {%s}\n", acct.Address.String())
	return nil
}
