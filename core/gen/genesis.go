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
	"errors"

	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/configure"
	"github.com/palletone/go-palletone/core"
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
func SetupGenesisBlock(genesis *core.Genesis, txs modules.Transactions) (*modules.Unit, error) {
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
	return modules.NewGenesisUnit(genesis, txs)
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
