/*
 *
 *    This file is part of go-palletone.
 *    go-palletone is free software: you can redistribute it and/or modify
 *    it under the terms of the GNU General Public License as published by
 *    the Free Software Foundation, either version 3 of the License, or
 *    (at your option) any later version.
 *    go-palletone is distributed in the hope that it will be useful,
 *    but WITHOUT ANY WARRANTY; without even the implied warranty of
 *    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *    GNU General Public License for more details.
 *    You should have received a copy of the GNU General Public License
 *    along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
 * /
 *
 *  * @author PalletOne core developer <dev@pallet.one>
 *  * @date 2018-2019
 *
 */

package parameter

type SysParameters struct {
	TxCoinDayInterest       float64 //一个币天产生多少利息
	DepositContractInterest float64 //保证金合约一天产生多少利息
	GenerateUnitReward      uint64  //每产生一个Unit奖励多少Dao的Token
	RewardHeight            uint64
	ContractFeeJuryPercent  float64 //合约执行的手续费中，有多少比例是分给Mediator
}

var CurrentSysParameters = &SysParameters{
	TxCoinDayInterest:       0.01 / 365,
	DepositContractInterest: 0.02 / 365,
	GenerateUnitReward:      100000000,
	RewardHeight:            50,
	ContractFeeJuryPercent:  0.6,
}
