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
 *
 */

package core

// ChainParameters 区块链网络参数结构体的定义
//变量名一定要大些，否则外部无法访问，导致无法进行json编码和解码
type ChainParameters struct {
	// 目前的操作交易费，current schedule of fees
	CurrentFees FeeSchedule `json:"currentFees"`

	// unit生产之间的间隔时间，以秒为单元。 interval in seconds between Units
	MediatorInterval uint8 `json:"mediatorInterval"`

	// 区块链维护事件之间的间隔，以秒为单元。 interval in sections between unit maintenance events
	MaintenanceInterval uint32 `json:"maintenanceInterval"`

	// 在维护时跳过的verifiedUnitInterval数量。 number of verifiedUnitInterval to skip at maintenance time
	//	MaintenanceSkipSlots uint8
}

func NewChainParams() (c ChainParameters) {
	c.CurrentFees = newFeeSchedule()
	c.MediatorInterval = DefaultMediatorInterval
	c.MaintenanceInterval = DefaultMaintenanceInterval

	return
}

// 操作交易费计划
type FeeSchedule struct {
	// mediator 创建费用
	MediatorCreateFee        uint64                `json:"mediatorCreateFee"`
	VoteMediatorFee          uint64                `json:"voteMediatorFee"`
	TransferFee              TransferFeeParameters `json:"transferPtnFee"`
	SetDesiredMediatorNumFee uint8                 `json:"setDesiredMediatorNumFee"`
}

func newFeeSchedule() (f FeeSchedule) {
	f.MediatorCreateFee = DefaultMediatorCreateFee
	f.VoteMediatorFee = DefaultVoteMediatorFee
	f.SetDesiredMediatorNumFee = DefaultSetDesiredMediatorNumFee
	f.TransferFee = newTransferFeeParameters()

	return
}

// 转账交易费
type TransferFeeParameters struct {
	BaseFee       uint64 `json:"baseFee"`
	PricePerKByte uint64 `json:"pricePerKByte"`
}

func newTransferFeeParameters() (tf TransferFeeParameters) {
	tf.BaseFee = DefaultTransferPtnBaseFee
	tf.PricePerKByte = DefaultTransferPtnPricePerKByte

	return
}
