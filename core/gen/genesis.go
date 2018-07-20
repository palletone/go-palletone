// Copyright 2014 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package gen

import (
	//"crypto/ecdsa"
	"errors"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/rlp"
	"github.com/palletone/go-palletone/common/txscript"
	"github.com/palletone/go-palletone/configure"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/core/accounts"
	"github.com/palletone/go-palletone/core/accounts/keystore"
	asset2 "github.com/palletone/go-palletone/dag/asset"
	dagCommon "github.com/palletone/go-palletone/dag/common"
	"github.com/palletone/go-palletone/dag/modules"
	"fmt"
)

const (
	DefaultMediatorInterval = 5
	DefaultMediatorCount    = 21
	DefaultTokenAmount      = 1000000000
	DefaultTokenDecimal     = 8
	DefaultDepositRate      = 0.02
	defaultTokenHolder      = "P1Kp2hcLhGEP45Xgx7vmSrE37QXunJUd8gJ"
)

// SetupGenesisBlock writes or updates the genesis block in db.
// The block that will be used is:
//
//                          genesis == nil       genesis != nil
//                       +------------------------------------------
//     db has no genesis |  main-net default  |  genesis
//     db has genesis    |  from DB           |  genesis (if compatible)
//
// The stored chain configuration will be updated if it is compatible (i.e. does not
// specify a fork block below the local head block). In case of a conflict, the
// error is a *configure.ConfigCompatError and the new, unwritten config is returned.
//
// The returned chain configuration is never nil.
func SetupGenesisUnit(genesis *core.Genesis, ks *keystore.KeyStore, account accounts.Account) error {
	//	privateKey, err := ks.GetPrivateKey(account)
	//	if err != nil {
	//		log.Info("SetupGenesisUnit GetPrivateKey err:", err.Error())
	//		return err
	//	}

	unit, err := setupGenesisUnit(genesis, ks)
	if err != nil && unit != nil {
		log.Info("Genesis is Exist")
		return nil
	}
	if err != nil {
		log.Error("Failed to write genesis block:", err.Error())
		return err
	}

	sign, err1 := ks.SigUnit(unit, account.Address)
	if err1 != nil {
		msg := fmt.Sprintf("Failed to write genesis block:%v", err1.Error())
		log.Error(msg)
		return err1
	}
	publicKey, err2 := ks.GetPublicKey(account.Address)
	if err2 != nil {
		msg:=fmt.Sprintf("Failed to Get Public Key:%v", err2.Error())
		log.Error(msg)
		return err2
	}
	r := sign[:32]
	s := sign[32:64]
	v := append(sign, sign[64]+27)
	unit.UnitHeader.Authors = &modules.Authentifier{
		Address:account.Address.String(),
		R:r,
		S:s,
		V:v,
	}
	//	publicKey := crypto.FromECDSAPub(&privateKey.PublicKey)
	log.Info("Successfully SIG Genesis Block")
	pass := keystore.VerifyUnitWithPK(sign, unit, publicKey)
	if pass {
		log.Info("Valid signature")
	} else {
		log.Info("Invalid signature")
	}
	CommitDB(unit, publicKey, sign, common.Address{} /*account.Address*/)
	return nil
}

func setupGenesisUnit(genesis *core.Genesis, ks *keystore.KeyStore) (*modules.Unit, error) {

	// Just commit the new block if there is no stored genesis block.
	stored := dagCommon.GetGenesisUnit(0)
	// Check whether the genesis block is already written.
	if stored != nil {
		return stored, errors.New("the genesis block is already written")
	}

	if genesis == nil {
		log.Info("Writing default main-net genesis block")
		genesis = DefaultGenesisBlock()
	} else {
		log.Info("Writing custom genesis block")
	}
	txs := GetGensisTransctions(ks, genesis)
	log.Info("-> Genesis transactions:")
	for i, tx := range txs{
		msg := fmt.Sprintf("Tx[%d]: %s\n", i, tx.TxHash.String())
		log.Info(msg)
	}
	//return modules.NewGenesisUnit(genesis, txs)
	return dagCommon.NewGenesisUnit(txs)
}

func GetGensisTransctions(ks *keystore.KeyStore, genesis *core.Genesis) modules.Transactions {
	// step1, generate payment payload message: coin creation
	holder := common.Address{}
	holder.SetString(genesis.TokenHolder)
	if txscript.CheckP2PKHAddress(holder) == false {
		log.Error("Genesis holder address is an invalid p2pkh address.")
		return nil
	}
	assetInfo := modules.AssetInfo{
		Alias:          genesis.Alias,
		InitialTotal:   genesis.TokenAmount,
		Decimal:        genesis.TokenDecimal,
		DecimalUnit:    genesis.DecimalUnit,
		OriginalHolder: holder,
	}
	extra, err := rlp.EncodeToBytes(assetInfo)
	if err != nil {
		log.Error("Get genesis assetinfo bytes error.")
		return nil
	}
	txin := modules.Input{
		Extra: extra, // save asset info
	}
	// get new asset id
	assetId := asset2.NewAsset()
	asset := modules.Asset{
		AssertId: assetId,
		UniqueId: assetId,
		ChainId:  genesis.ChainID,
	}
	// generate p2pkh bytes
	publicKey, err := ks.GetPublicKey(holder)
	if err != nil {
		log.Error("Failed to Get Public Key:", err.Error())
		return nil
	}
	pkscript := txscript.PayToPubkeyHashScript(publicKey)
	txout := modules.Output{
		Value:    genesis.TokenAmount,
		Asset:    asset,
		PkScript: pkscript,
	}
	pay := modules.PaymentPayload{
		Inputs:  []modules.Input{txin},
		Outputs: []modules.Output{txout},
	}
	msg0 := modules.Message{
		App:     modules.APP_PAYMENT,
		Payload: pay,
	}
	msg0.PayloadHash = rlp.RlpHash(pay)
	// step2, generate global config payload message
	configPayload, err := dagCommon.GenGenesisConfigPayload(genesis)
	if err!=nil{
		log.Error("Generate genesis unit config payload error.")
		return nil
	}
	msg1 := modules.Message{
		App:     modules.APP_CONFIG,
		Payload: configPayload,
	}
	msg1.PayloadHash = rlp.RlpHash(configPayload)
	// step3, genesis transaction
	tx := &modules.Transaction{
		AccountNonce: 1,
		TxMessages: []modules.Message{msg0, msg1},
	}
	txHash, err := rlp.EncodeToBytes(tx.TxMessages)
	if err!=nil{
		msg := fmt.Sprintf("Get genesis transactions hash error: %s", err)
		log.Error(msg)
		return nil
	}
	tx.TxHash.SetBytes(txHash)
	// step4, sign tx
	R, S, V, err := ks.SigTX(tx.TxHash, holder)
	if err!=nil {
		msg := fmt.Sprintf("Sign transaction error: %s", err)
		log.Error(msg)
		return nil
	}
	tx.From = &modules.Authentifier{
		Address:holder.String(),
		R:R,
		S:S,
		V:V,
	}
	fmt.Println("tx.From", tx.From)
	txs := []*modules.Transaction{tx}
	return txs
}

func CommitDB(unit *modules.Unit, publicKey, addr []byte, address common.Address) error {
	var authentifier modules.Authentifier
	authentifier.Address = string(addr)
	author := new(modules.Author)
	author.Address = address
	author.Pubkey = publicKey
	author.TxAuthentifier = authentifier
	unit.UnitHeader.Witness = append(unit.UnitHeader.Witness, *author)
	if err:=dagCommon.SaveUnit(*unit); err!=nil{
		log.Error("Save genesis unit error.")
	}
	return nil
}

// DefaultGenesisBlock returns the PalletOne main net genesis block.
func DefaultGenesisBlock() *core.Genesis {
	SystemConfig := core.SystemConfig{
		MediatorInterval: DefaultMediatorInterval,
		DepositRate:      DefaultDepositRate,
	}
	return &core.Genesis{
		Version:                   configure.Version,
		TokenAmount:               DefaultTokenAmount,
		TokenDecimal:              DefaultTokenDecimal,
		ChainID:                   1,
		TokenHolder:               defaultTokenHolder,
		SystemConfig:              SystemConfig,
		InitialActiveMediators:    DefaultMediatorCount,
		InitialMediatorCandidates: InitialMediatorCandidates(DefaultMediatorCount, defaultTokenHolder),
	}
}

// DefaultTestnetGenesisBlock returns the Ropsten network genesis block.
func DefaultTestnetGenesisBlock() *core.Genesis {
	SystemConfig := core.SystemConfig{
		MediatorInterval: DefaultMediatorInterval,
		DepositRate:      DefaultDepositRate,
	}
	return &core.Genesis{
		Version:                   configure.Version,
		TokenAmount:               DefaultTokenAmount,
		TokenDecimal:              DefaultTokenDecimal,
		ChainID:                   1,
		TokenHolder:               defaultTokenHolder,
		SystemConfig:              SystemConfig,
		InitialActiveMediators:    DefaultMediatorCount,
		InitialMediatorCandidates: InitialMediatorCandidates(DefaultMediatorCount, defaultTokenHolder),
	}
}

func InitialMediatorCandidates(len int, address string) []string {
	initialMediatorSet := make([]string, len)
	for i := 0; i < len; i++ {
		initialMediatorSet[i] = address
	}

	return initialMediatorSet
}
