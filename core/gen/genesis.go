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
	"github.com/palletone/go-palletone/common/crypto"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/configure"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/core/accounts"
	"github.com/palletone/go-palletone/core/accounts/keystore"
	dagCommon "github.com/palletone/go-palletone/dag/common"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/storage"
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
	privateKey, err := ks.GetPrivateKey(account)
	if err != nil {
		log.Info("SetupGenesisUnit GetPrivateKey err:", err.Error())
		return err
	}

	unit, err := setupGenesisUnit(genesis, ks)
	if err != nil && unit != nil {
		log.Info("Genesis is Exist")
		return nil
	}
	if err != nil {
		log.Info("Failed to write genesis block:", err.Error())
		return err
	}

	sign, err1 := ks.SigUnit(unit, privateKey)
	if err != nil {
		log.Info("Failed to write genesis block:", err1.Error())
		return err
	}
	publicKey := crypto.FromECDSAPub(&privateKey.PublicKey)
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
	txs := GetTransctions(ks, genesis)
	// Just commit the new block if there is no stored genesis block.
	stored := storage.GetGenesisUnit(0)
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
	//return modules.NewGenesisUnit(genesis, txs)
	return dagCommon.NewGenesisUnit(txs)
}

func GetTransctions(ks *keystore.KeyStore, genesis *core.Genesis) modules.Transactions {
	tx := &modules.Transaction{AccountNonce: 1}
	txs := []*modules.Transaction{}
	txs = append(txs, tx)
	return txs
}

func CommitDB(unit *modules.Unit, publicKey []byte, sign string, address common.Address) error {
	var authentifier modules.Authentifier
	authentifier.R = sign
	author := new(modules.Author)
	author.Address = address
	author.Pubkey = publicKey
	author.TxAuthentifier = authentifier
	unit.UnitHeader.Authors = author
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
