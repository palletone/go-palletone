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
	MinMaintSkipSlots    uint8    `json:"min_maint_skip_slots"`  // 最小区块链维护间隔
	MinimumMediatorCount uint8    `json:"min_mediator_count"`    // 最小活跃mediator数量
	MinMediatorInterval  uint8    `json:"min_mediator_interval"` // 最小的生产槽间隔时间
	UccPrivileged        bool     `json:"ucc_privileged"`        // 防止容器以root权限运行
	UccCapDrop           []string `json:"ucc_cap_drop"`          // 确保容器以最小权限运行
	UccNetworkMode       string   `json:"ucc_network_mode"`      // 容器运行网络模式
	UccOOMKillDisable    bool     `json:"ucc_oom_kill_disable"`  // 是否内存使用量超过上限时系统杀死进程
}

func NewImmutChainParams() ImmutableChainParameters {
	return ImmutableChainParameters{
		MinMaintSkipSlots:    DefaultMinMaintSkipSlots,
		MinimumMediatorCount: DefaultMinMediatorCount,
		MinMediatorInterval:  DefaultMinMediatorInterval,
		UccPrivileged:        DefaultUccPrivileged,
		UccCapDrop: []string{"mknod", "setfcap", "audit_write", "net_bind_service", "net_raw",
			"kill", "setgid", "setuid", "setpcap", "chown", "fowner", "sys_chroot"},
		UccNetworkMode:    DefaultUccNetworkMode,
		UccOOMKillDisable: DefaultUccOOMKillDisable,
	}
}

func NewChainParametersBase() ChainParametersBase {
	return ChainParametersBase{
		GenerateUnitReward:        DefaultGenerateUnitReward,
		RewardHeight:              DefaultRewardHeight,
		PledgeDailyReward:         DefaultPledgeDailyReward,
		FoundationAddress:         DefaultFoundationAddress,
		RmExpConFromSysParam:      DefaultRmExpConFromSysParam,
		DepositAmountForMediator:  DefaultDepositAmountForMediator,
		DepositAmountForJury:      DefaultDepositAmountForJury,
		DepositAmountForDeveloper: DefaultDepositAmountForDeveloper,
		ActiveMediatorCount:       DefaultActiveMediatorCount,
		MaximumMediatorCount:      DefaultMaxMediatorCount,
		MediatorInterval:          DefaultMediatorInterval,
		MaintenanceInterval:       DefaultMaintenanceInterval,
		MaintenanceSkipSlots:      DefaultMaintenanceSkipSlots,
		MediatorCreateFee:         DefaultMediatorCreateFee,
		AccountUpdateFee:          DefaultAccountUpdateFee,
		TransferPtnBaseFee:        DefaultTransferPtnBaseFee,
		TransferPtnPricePerKByte:  DefaultTransferPtnPricePerKByte,
		// ContractInvokeFee:         DefaultContractInvokeFee,
		UnitMaxSize: DefaultUnitMaxSize,
	}
}

type ChainParametersBase struct {
	GenerateUnitReward uint64 `json:"generate_unit_reward"` //每生产一个单元，奖励多少Dao的PTN
	PledgeDailyReward  uint64 `json:"pledge_daily_reward"`  //质押金的日奖励额
	RewardHeight       uint64 `json:"reward_height"`        //每多少高度进行一次奖励的派发
	UnitMaxSize        uint64 `json:"unit_max_size"`        //一个单元最大允许多大
	FoundationAddress  string `json:"foundation_address"`   //基金会地址，该地址具有一些特殊权限，比如发起参数修改的投票，发起罚没保证金等

	DepositAmountForMediator  uint64 `json:"deposit_amount_for_mediator"` //保证金的数量
	DepositAmountForJury      uint64 `json:"deposit_amount_for_jury"`
	DepositAmountForDeveloper uint64 `json:"deposit_amount_for_developer"`
	RmExpConFromSysParam      bool   `json:"remove_expired_container_from_system_parameter"`
	//UccCpuSetCpus string `json:"ucc_cpu_set_cpus"` //限制使用某些CPUS  "1,3"  "0-2"

	// 活跃mediator的数量。 number of active mediators
	ActiveMediatorCount uint8 `json:"active_mediator_count"`

	// 用户可投票mediator的最大数量。the maximum number of mediator users can vote for
	MaximumMediatorCount uint8 `json:"max_mediator_count"`

	// unit生产之间的间隔时间，以秒为单元。 interval in seconds between Units
	MediatorInterval uint8 `json:"mediator_interval"`

	// MaintenanceInterval 必须是 MediatorInterval的整数倍
	// 区块链维护事件之间的间隔，以秒为单元。 interval in sections between unit maintenance events
	MaintenanceInterval uint32 `json:"maintenance_interval"`

	// 在维护时跳过的MediatorInterval数量。 number of MediatorInterval to skip at maintenance time
	MaintenanceSkipSlots uint8 `json:"maintenance_skip_slots"`

	// 目前的操作交易费，current schedule of fees
	MediatorCreateFee        uint64 `json:"mediator_create_fee"`
	AccountUpdateFee         uint64 `json:"account_update_fee"`
	TransferPtnBaseFee       uint64 `json:"transfer_ptn_base_fee"`
	TransferPtnPricePerKByte uint64 `json:"transfer_ptn_price_per_KByte"`
	// ContractInvokeFee        uint64 `json:"contract_invoke_fee"`
}

func NewChainParams() ChainParameters {
	return ChainParameters{
		ChainParametersBase: NewChainParametersBase(),
		// TxCoinYearRate:       DefaultTxCoinYearRate,
		//DepositPeriod:        DefaultDepositPeriod,
		UccMemory:     DefaultUccMemory,
		UccCpuShares:  DefaultUccCpuShares,
		UccCpuQuota:   DefaultUccCpuQuota,
		UccDisk:       DefaultUccDisk,
		UccDuringTime: DefaultContainerDuringTime,

		TempUccMemory:         DefaultTempUccMemory,
		TempUccCpuShares:      DefaultTempUccCpuShares,
		TempUccCpuQuota:       DefaultTempUccCpuQuota,
		ContractSystemVersion: DefaultContractSystemVersion,
		ContractSignatureNum:  DefaultContractSignatureNum,
		ContractElectionNum:   DefaultContractElectionNum,

		ContractTxTimeoutUnitFee:  DefaultContractTxTimeoutUnitFee,
		ContractTxSizeUnitFee:     DefaultContractTxSizeUnitFee,
		ContractTxInstallFeeLevel: DefaultContractTxInstallFeeLevel,
		ContractTxDeployFeeLevel:  DefaultContractTxDeployFeeLevel,
		ContractTxInvokeFeeLevel:  DefaultContractTxInvokeFeeLevel,
		ContractTxStopFeeLevel:    DefaultContractTxStopFeeLevel,
	}
}

// ChainParameters 区块链网络参数结构体的定义
//变量名一定要大写，否则外部无法访问，导致无法进行json编码和解码
type ChainParameters struct {
	ChainParametersBase

	// TxCoinYearRate float64 `json:"tx_coin_year_rate"` //交易币天的年利率
	//DepositRate   float64 `json:"deposit_rate"`   //保证金的年利率
	//DepositPeriod int     `json:"deposit_period"` //保证金周期

	//对启动用户合约容器的相关资源的限制
	UccMemory     int64 `json:"ucc_memory"`
	UccCpuShares  int64 `json:"ucc_cpu_shares"`
	UccCpuQuota   int64 `json:"ucc_cpu_quota"`
	UccDisk       int64 `json:"ucc_disk"`
	UccDuringTime int64 `json:"ucc_during_time"`

	//对中间容器的相关资源限制
	TempUccMemory    int64 `json:"temp_ucc_memory"`
	TempUccCpuShares int64 `json:"temp_ucc_cpu_shares"`
	TempUccCpuQuota  int64 `json:"temp_ucc_cpu_quota"`

	//contract about
	ContractSystemVersion string `json:"contract_system_version"`
	ContractSignatureNum  int    `json:"contract_signature_num"`
	ContractElectionNum   int    `json:"contract_election_num"`

	ContractTxTimeoutUnitFee  uint64  `json:"contract_tx_timeout_unit_fee"`
	ContractTxSizeUnitFee     uint64  `json:"contract_tx_size_unit_fee"`
	ContractTxInstallFeeLevel float64 `json:"contract_tx_install_fee_level"`
	ContractTxDeployFeeLevel  float64 `json:"contract_tx_deploy_fee_level"`
	ContractTxInvokeFeeLevel  float64 `json:"contract_tx_invoke_fee_level"`
	ContractTxStopFeeLevel    float64 `json:"contract_tx_stop_fee_level"`
}

func CheckSysConfigArgType(field, value string) error {
	var err error
	vn := reflect.ValueOf(ChainParameters{}).FieldByName(field)

	switch vn.Kind() {
	case reflect.Invalid:
		err = fmt.Errorf("no such field: %v", field)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		_, err = strconv.ParseInt(value, 10, 64)
	case reflect.Bool:
		_, err = strconv.ParseBool(value)
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

type GetMediatorCountFn func() int

func CheckChainParameterValue(field, value string, icp *ImmutableChainParameters, cp *ChainParameters,
	fn GetMediatorCountFn) error {
	var err error

	switch field {
	case "MediatorInterval":
		newMediatorInterval, _ := strconv.ParseUint(value, 10, 64)
		if newMediatorInterval < uint64(icp.MinMediatorInterval) {
			err = fmt.Errorf("new mediator interval(%v) cannot less than min interval(%v)",
				newMediatorInterval, icp.MinMediatorInterval)
		}
	case "MaintenanceSkipSlots":
		newMaintenanceSkipSlots, _ := strconv.ParseUint(value, 10, 64)
		if newMaintenanceSkipSlots < uint64(icp.MinMaintSkipSlots) {
			err = fmt.Errorf("new MaintenanceSkipSlots(%v) cannot less than MinMaintSkipSlots(%v)",
				newMaintenanceSkipSlots, icp.MinMaintSkipSlots)
		}
	case "ActiveMediatorCount":
		newActiveMediatorCount, _ := strconv.ParseUint(value, 10, 16)
		if (newActiveMediatorCount & 1) == 0 {
			// 保证活跃mediator数量为奇数
			err = fmt.Errorf("new ActiveMediatorCount(%v) must be odd", newActiveMediatorCount)
		} else if newActiveMediatorCount < uint64(icp.MinimumMediatorCount) {
			// 保证活跃mediator数量不小于MinimumMediatorCount
			err = fmt.Errorf("new ActiveMediatorCount(%v) cannot less than MinimumMediatorCount(%v)",
				newActiveMediatorCount, icp.MinimumMediatorCount)
		} else {
			// 保证活跃mediator数量不大于mediator总数
			mediatorCount := uint64(fn())
			if newActiveMediatorCount > mediatorCount {
				err = fmt.Errorf("new ActiveMediatorCount(%v) cannot more than mediator count(%v)",
					newActiveMediatorCount, mediatorCount)
			}
		}
	case "MaintenanceInterval":
		newMaintenanceInterval, _ := strconv.ParseUint(value, 10, 64)
		minMaintenanceInterval := cp.MediatorInterval * cp.MaintenanceSkipSlots
		if !(newMaintenanceInterval > uint64(minMaintenanceInterval)) {
			// 保证MaintenanceInterval大于必要的时长
			err = fmt.Errorf("new MaintenanceInterval(%v) must be larger than %v",
				newMaintenanceInterval, minMaintenanceInterval)
		} else if newMaintenanceInterval%uint64(cp.MediatorInterval) != 0 {
			// 保证MaintenanceInterval能被MediatorInterval整除
			err = fmt.Errorf("new MaintenanceInterval(%v) must be divisible by mediator interval(%v)",
				newMaintenanceInterval, cp.MediatorInterval)
		}

	default:
		err = nil
	}

	return err
}
