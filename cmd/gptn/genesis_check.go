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
	"time"

	"github.com/palletone/go-palletone/cmd/utils"
	"github.com/palletone/go-palletone/core"
	"gopkg.in/urfave/cli.v1"
)

// regulateGenesisTimestamp, regulate initial timestamp
// @author Albert·Gou
func regulateGenesisTimestamp(ctx *cli.Context, genesis *core.Genesis) {
	if ctx.GlobalIsSet(GenesisTimestampFlag.Name) {
		secFromNow := ctx.GlobalInt64(GenesisTimestampFlag.Name)
		mi := int64(genesis.InitialParameters.MediatorInterval)
		genesis.InitialTimestamp = time.Now().Unix() + mi + secFromNow
		genesis.InitialTimestamp -= genesis.InitialTimestamp % mi
	}
}

// validateGenesis, determine if the settings in genesis meet the security check, and if not, terminate the program
// 判断genesis中的设置是否符合安全性检查，如果不满足，则终止程序
// @author Albert·Gou
func validateGenesis(genesis *core.Genesis) {
	initialTime := genesis.InitialTimestamp
	fcAssert(initialTime != 0, "Must initialize genesis timestamp.")

	mediatorInterv := int64(genesis.InitialParameters.MediatorInterval)
	fcAssert(mediatorInterv > 0, "mediator interval must be larger than zero.")

	fcAssert(initialTime%mediatorInterv == 0,
		"Genesis timestamp must be divisible by mediator interval.")

	//minMediatorInterval := int64(genesis.ImmutableParameters.MinMediatorInterval)
	//fcAssert(mediatorInterv > minMediatorInterval, "mediator interval must be larger than min interval.")

	mediatorCandidateCount := uint16(len(genesis.InitialMediatorCandidates))
	fcAssert(mediatorCandidateCount != 0, "Cannot start a chain with zero mediators.")

	initialActiveMediator := genesis.InitialActiveMediators
	fcAssert(initialActiveMediator <= mediatorCandidateCount,
		"initial active mediators is larger than the number of candidate mediators.")

	fcAssert((initialActiveMediator&1) == 1, "min mediator count must be odd")

	minMediatorCount := uint16(genesis.ImmutableParameters.MinMediatorCount)
	fcAssert(initialActiveMediator >= minMediatorCount,
		"initial active mediators must be larger than min mediator count.")
}

// fcAssert, determine if the expectation is true, if not, the program terminates and prompts
// @author Albert·Gou
func fcAssert(expectation bool, errTip string, args ...interface{}) {
	if !expectation {
		if len(args) > 0 {
			utils.Fatalf(errTip, args)
		} else {
			utils.Fatalf(errTip)
		}
	}
}
