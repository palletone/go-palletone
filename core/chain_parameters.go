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

import (
	"fmt"
	"reflect"
	"strconv"
)

type ImmutableChainParameters struct {
	MinimumMediatorCount uint8 `json:"minMediatorCount"`
	MinMediatorInterval  uint8 `json:"minMediatorInterval"`
}

func NewImmutChainParams() ImmutableChainParameters {
	return ImmutableChainParameters{
		MinimumMediatorCount: DefaultMinMediatorCount,
		MinMediatorInterval:  DefaultMinMediatorInterval,
	}
}

func NewChainParametersBase() ChainParametersBase {
	return ChainParametersBase{
		GenerateUnitReward:        DefaultGenerateUnitReward,
		RewardHeight:              DefaultRewardHeight,
		FoundationAddress:         DefaultFoundationAddress,
		DepositAmountForMediator:  DefaultDepositAmountForMediator,
		DepositAmountForJury:      DefaultDepositAmountForJury,
		DepositAmountForDeveloper: DefaultDepositAmountForDeveloper,
		UccCpuSetCpus:             DefaultUccCpuSetCpus,
		ActiveMediatorCount:       DefaultActiveMediatorCount,
		MaximumMediatorCount:      DefaultMaxMediatorCount,
		MediatorInterval:          DefaultMediatorInterval,
		MaintenanceInterval:       DefaultMaintenanceInterval,
		MaintenanceSkipSlots:      DefaultMaintenanceSkipSlots,
		MediatorCreateFee:         DefaultMediatorCreateFee,
		AccountUpdateFee:          DefaultAccountUpdateFee,
		TransferPtnBaseFee:        DefaultTransferPtnBaseFee,
		TransferPtnPricePerKByte:  DefaultTransferPtnPricePerKByte,
		//CurrentFees:               newFeeSchedule(),
	}
}

type ChainParametersBase struct {
	GenerateUnitReward uint64 `json:"generateUnitReward"` //每生产一个单元，奖励多少Dao的PTN
	RewardHeight       uint64 `json:"reward_height"`      //每多少高度进行一次奖励的派发

	FoundationAddress string `json:"foundationAddress"` //基金会地址，该地址具有一些特殊权限，比如发起参数修改的投票，发起罚没保证金等

	DepositAmountForMediator  uint64 `json:"depositAmountForMediator"` //保证金的数量
	DepositAmountForJury      uint64 `json:"depositAmountForJury"`
	DepositAmountForDeveloper uint64 `json:"depositAmountForDeveloper"`

	UccCpuSetCpus string `json:"ucc_cpu_set_cpus"` //限制使用某些CPUS  "1,3"  "0-2"

	// 活跃mediator的数量。 number of active mediators
	ActiveMediatorCount uint8 `json:"activeMediatorCount"`

	// 用户可投票mediator的最大数量。the maximum number of mediator users can vote for
	MaximumMediatorCount uint8 `json:"maxMediatorCount"`

	// unit生产之间的间隔时间，以秒为单元。 interval in seconds between Units
	MediatorInterval uint8 `json:"mediatorInterval"`

	// 区块链维护事件之间的间隔，以秒为单元。 interval in sections between unit maintenance events
	MaintenanceInterval uint32 `json:"maintenanceInterval"`

	// 在维护时跳过的MediatorInterval数量。 number of MediatorInterval to skip at maintenance time
	MaintenanceSkipSlots uint8 `json:"maintenanceSkipSlots"`

	// 目前的操作交易费，current schedule of fees
	MediatorCreateFee        uint64 `json:"mediatorCreateFee"`
	AccountUpdateFee         uint64 `json:"accountUpdateFee"`
	TransferPtnBaseFee       uint64 `json:"transferPtnBaseFee"`
	TransferPtnPricePerKByte uint64 `json:"transferPtnPricePerKByte"`
	//CurrentFees              FeeSchedule `json:"currentFees"`
}

func NewChainParams() ChainParameters {
	return ChainParameters{
		ChainParametersBase:  NewChainParametersBase(),
		DepositRate:          DefaultDepositRate,
		TxCoinYearRate:       DefaultTxCoinYearRate,
		DepositPeriod:        DefaultDepositPeriod,
		UccMemory:            DefaultUccMemory,
		UccMemorySwap:        DefaultUccMemorySwap,
		UccCpuShares:         DefaultUccCpuShares,
		UccCpuPeriod:         DefaultCpuPeriod,
		UccCpuQuota:          DefaultUccCpuQuota,
		TempUccMemory:        DefaultTempUccMemory,
		TempUccMemorySwap:    DefaultTempUccMemorySwap,
		TempUccCpuShares:     DefaultTempUccCpuShares,
		TempUccCpuQuota:      DefaultTempUccCpuQuota,
		ContractSignatureNum: DefaultContractSignatureNum,
		ContractElectionNum:  DefaultContractElectionNum,
	}
}

// ChainParameters 区块链网络参数结构体的定义
//变量名一定要大写，否则外部无法访问，导致无法进行json编码和解码
type ChainParameters struct {
	ChainParametersBase

	TxCoinYearRate float64 `json:"txCoinYearRate"` //交易币天的年利率
	DepositRate    float64 `json:"depositRate"`    //保证金的年利率
	DepositPeriod  int     `json:"depositPeriod"`  //保证金周期

	//对启动用户合约容器的相关资源的限制
	UccMemory     int64 `json:"ucc_memory"`      //物理内存  104857600  100m
	UccMemorySwap int64 `json:"ucc_memory_swap"` //内存交换区，不设置默认为memory的两倍
	UccCpuShares  int64 `json:"ucc_cpu_shares"`  //CPU占用率，相对的  CPU 利用率权重，默认为 1024
	UccCpuQuota   int64 `json:"ucc_cpu_quota"`   // 限制CPU --cpu-period=50000 --cpu-quota=25000
	UccCpuPeriod  int64 `json:"ucc_cpu_period"`  //限制CPU 周期设为 50000，将容器在每个周期内的 CPU 配额设置为 25000，表示该容器每 50ms 可以得到 50% 的 CPU 运行时间

	//对中间容器的相关资源限制
	TempUccMemory     int64 `json:"temp_ucc_memory"`
	TempUccMemorySwap int64 `json:"temp_ucc_memory_swap"`
	TempUccCpuShares  int64 `json:"temp_ucc_cpu_shares"`
	TempUccCpuQuota   int64 `json:"temp_ucc_cpu_quota"`

	//contract about
	ContractSignatureNum int `json:"contract_signature_num"`
	ContractElectionNum  int `json:"contract_election_num"`
}

func CheckSysConfigArgs(field, value string) error {
	var err error
	vn := reflect.ValueOf(ChainParameters{}).FieldByName(field)

	switch vn.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		_, err = strconv.ParseInt(value, 10, 64)
	case reflect.Invalid:
		err = fmt.Errorf("no such field: %v", field)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		_, err = strconv.ParseUint(value, 10, 64)
	case reflect.String:
		err = nil
	case reflect.Float64, reflect.Float32:
		_, err = strconv.ParseFloat(value, 64)
	default:
		err = fmt.Errorf("unexpected type: %v", vn.Type().String())
	}

	return err
}

// 操作交易费计划
//type FeeSchedule struct {
//	// mediator 创建费用
//	MediatorCreateFee uint64                `json:"mediatorCreateFee"`
//	AccountUpdateFee  uint64                `json:"accountUpdateFee"`
//	TransferFee       TransferFeeParameters `json:"transferPtnFee"`
//}

//func newFeeSchedule() (f FeeSchedule) {
//	f.MediatorCreateFee = DefaultMediatorCreateFee
//	f.AccountUpdateFee = DefaultAccountUpdateFee
//	f.TransferFee = newTransferFeeParameters()
//
//	return
//}

// 转账交易费
//type TransferFeeParameters struct {
//	BaseFee       uint64 `json:"baseFee"`
//	PricePerKByte uint64 `json:"pricePerKByte"`
//}

//func newTransferFeeParameters() (tf TransferFeeParameters) {
//	tf.BaseFee = DefaultTransferPtnBaseFee
//	tf.PricePerKByte = DefaultTransferPtnPricePerKByte
//
//	return
//}
