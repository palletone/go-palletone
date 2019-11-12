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
 *  * @author PalletOne core developers <dev@pallet.one>
 *  * @date 2018-2019
 *
 */
package migration

import (
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/dag/constants"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/storage"
	"strconv"
)

type Migration104alpha_104beta struct {
	dagdb   ptndb.Database
	idxdb   ptndb.Database
	utxodb  ptndb.Database
	statedb ptndb.Database
	propdb  ptndb.Database
}

func (m *Migration104alpha_104beta) FromVersion() string {
	return "1.0.4-alpha"
}

func (m *Migration104alpha_104beta) ToVersion() string {
	return "1.0.4-beta"
}

func (m *Migration104alpha_104beta) ExecuteUpgrade() error {
	// 增加两个系统参数
	if err := m.upgradeGP(); err != nil {
		return err
	}

	return nil
}

func (m *Migration104alpha_104beta) upgradeGP() error {
	oldGp := &GlobalProperty104beta{}
	err := storage.RetrieveFromRlpBytes(m.propdb, constants.GLOBALPROPERTY_KEY, oldGp)
	if err != nil {
		log.Errorf(err.Error())
		return err
	}

	newData := &modules.GlobalPropertyTemp{}

	newData.ActiveJuries = oldGp.ActiveJuries
	newData.ActiveMediators = oldGp.ActiveMediators
	newData.PrecedingMediators = oldGp.PrecedingMediators
	newData.ImmutableParameters = oldGp.ImmutableParameters

	newData.ChainParameters.GenerateUnitReward = oldGp.ChainParameters.GenerateUnitReward
	newData.ChainParameters.PledgeDailyReward = oldGp.ChainParameters.PledgeDailyReward
	newData.ChainParameters.RewardHeight = oldGp.ChainParameters.RewardHeight
	newData.ChainParameters.UnitMaxSize = oldGp.ChainParameters.UnitMaxSize
	newData.ChainParameters.FoundationAddress = oldGp.ChainParameters.FoundationAddress
	newData.ChainParameters.DepositAmountForMediator = oldGp.ChainParameters.DepositAmountForMediator
	newData.ChainParameters.DepositAmountForJury = oldGp.ChainParameters.DepositAmountForJury
	newData.ChainParameters.DepositAmountForDeveloper = oldGp.ChainParameters.DepositAmountForDeveloper
	newData.ChainParameters.ActiveMediatorCount = oldGp.ChainParameters.ActiveMediatorCount
	newData.ChainParameters.MaximumMediatorCount = oldGp.ChainParameters.MaximumMediatorCount
	newData.ChainParameters.MediatorInterval = oldGp.ChainParameters.MediatorInterval
	newData.ChainParameters.MaintenanceInterval = oldGp.ChainParameters.MaintenanceInterval
	newData.ChainParameters.MaintenanceSkipSlots = oldGp.ChainParameters.MaintenanceSkipSlots
	newData.ChainParameters.MediatorCreateFee = oldGp.ChainParameters.MediatorCreateFee
	newData.ChainParameters.AccountUpdateFee = oldGp.ChainParameters.AccountUpdateFee
	newData.ChainParameters.TransferPtnBaseFee = oldGp.ChainParameters.TransferPtnBaseFee
	newData.ChainParameters.TransferPtnPricePerKByte = oldGp.ChainParameters.TransferPtnPricePerKByte

	// =======================chainParameters=============================
	UccMemory, err := strconv.ParseInt(oldGp.ChainParameters.UccMemory, 10, 64)
	if err != nil {
		return err
	}
	newData.ChainParameters.UccMemory = UccMemory

	UccCpuShares, err := strconv.ParseInt(oldGp.ChainParameters.UccCpuShares, 10, 64)
	if err != nil {
		return err
	}
	newData.ChainParameters.UccCpuShares = UccCpuShares

	UccCpuQuota, err := strconv.ParseInt(oldGp.ChainParameters.UccCpuQuota, 10, 64)
	if err != nil {
		return err
	}
	newData.ChainParameters.UccCpuQuota = UccCpuQuota

	UccDisk, err := strconv.ParseInt(oldGp.ChainParameters.UccDisk, 10, 64)
	if err != nil {
		return err
	}
	newData.ChainParameters.UccDisk = UccDisk

	TempUccMemory, err := strconv.ParseInt(oldGp.ChainParameters.TempUccMemory, 10, 64)
	if err != nil {
		return err
	}
	newData.ChainParameters.TempUccMemory = TempUccMemory

	TempUccCpuShares, err := strconv.ParseInt(oldGp.ChainParameters.TempUccCpuShares, 10, 64)
	if err != nil {
		return err
	}
	newData.ChainParameters.TempUccCpuShares = TempUccCpuShares

	TempUccCpuQuota, err := strconv.ParseInt(oldGp.ChainParameters.TempUccCpuQuota, 10, 64)
	if err != nil {
		return err
	}
	newData.ChainParameters.TempUccCpuQuota = TempUccCpuQuota

	ContractSignatureNum, err := strconv.ParseInt(oldGp.ChainParameters.ContractSignatureNum, 10, 64)
	if err != nil {
		return err
	}
	newData.ChainParameters.ContractSignatureNum = int(ContractSignatureNum)

	ContractElectionNum, err := strconv.ParseInt(oldGp.ChainParameters.ContractElectionNum, 10, 64)
	if err != nil {
		return err
	}
	newData.ChainParameters.ContractElectionNum = int(ContractElectionNum)

	ContractTxTimeoutUnitFee, err := strconv.ParseUint(oldGp.ChainParameters.ContractTxTimeoutUnitFee, 10, 64)
	if err != nil {
		return err
	}
	newData.ChainParameters.ContractTxTimeoutUnitFee = ContractTxTimeoutUnitFee

	ContractTxSizeUnitFee, err := strconv.ParseUint(oldGp.ChainParameters.ContractTxSizeUnitFee, 10, 64)
	if err != nil {
		return err
	}
	newData.ChainParameters.ContractTxSizeUnitFee = ContractTxSizeUnitFee

	ContractTxInstallFeeLevel, err := strconv.ParseFloat(oldGp.ChainParameters.ContractTxInstallFeeLevel, 64)
	if err != nil {
		return err
	}
	newData.ChainParameters.ContractTxInstallFeeLevel = ContractTxInstallFeeLevel

	ContractTxDeployFeeLevel, err := strconv.ParseFloat(oldGp.ChainParameters.ContractTxDeployFeeLevel, 64)
	if err != nil {
		return err
	}
	newData.ChainParameters.ContractTxDeployFeeLevel = ContractTxDeployFeeLevel

	ContractTxInvokeFeeLevel, err := strconv.ParseFloat(oldGp.ChainParameters.ContractTxInvokeFeeLevel, 64)
	if err != nil {
		return err
	}
	newData.ChainParameters.ContractTxInvokeFeeLevel = ContractTxInvokeFeeLevel

	ContractTxStopFeeLevel, err := strconv.ParseFloat(oldGp.ChainParameters.ContractTxStopFeeLevel, 64)
	if err != nil {
		return err
	}
	newData.ChainParameters.ContractTxStopFeeLevel = ContractTxStopFeeLevel


	//
	newData.ChainParameters.ContractSystemVersion = oldGp.ChainParameters.ContractSystemVersion
	newData.ChainParameters.RmExpConFromSysParam = oldGp.ChainParameters.RmExpConFromSysParam

	UccDuringTime, err := strconv.ParseInt(oldGp.ChainParameters.UccDuringTime, 10, 64)
	if err != nil {
		return err
	}
	newData.ChainParameters.UccDuringTime = UccDuringTime

	//新加的

	newData.ChainParameters.PledgeAllocateThreshold = core.DefaultPledgeAllocateThreshold
	newData.ChainParameters.PledgeRecordsThreshold = core.DefaultPledgeRecordsThreshold

	err = storage.StoreToRlpBytes(m.propdb, constants.GLOBALPROPERTY_KEY, newData)
	if err != nil {
		log.Errorf(err.Error())
		return err
	}

	return nil
}

type GlobalProperty104beta struct {
	GlobalPropBase104beta
	ActiveJuries       []common.Address
	ActiveMediators    []common.Address
	PrecedingMediators []common.Address
}


type GlobalPropBase104beta struct {
	ImmutableParameters core.ImmutableChainParameters // 不可改变的区块链网络参数
	ChainParameters     ChainParameters104beta       // 区块链网络参数
}

type ChainParameters104beta struct {
	ChainParametersBase104beta

	//对启动用户合约容器的相关资源的限制
	UccMemory     string
	UccCpuShares  string
	UccCpuQuota   string
	UccDisk       string
	UccDuringTime string

	//对中间容器的相关资源限制
	TempUccMemory    string
	TempUccCpuShares string
	TempUccCpuQuota  string

	//contract about
	ContractSystemVersion string
	ContractSignatureNum  string
	ContractElectionNum   string

	ContractTxTimeoutUnitFee string
	ContractTxSizeUnitFee     string
	ContractTxInstallFeeLevel string
	ContractTxDeployFeeLevel  string
	ContractTxInvokeFeeLevel  string
	ContractTxStopFeeLevel    string
}

type ChainParametersBase104beta struct {
	GenerateUnitReward uint64 `json:"generate_unit_reward"` //每生产一个单元，奖励多少Dao的PTN
	PledgeDailyReward  uint64 `json:"pledge_daily_reward"`  //质押金的日奖励额
	RewardHeight       uint64 `json:"reward_height"`        //每多少高度进行一次奖励的派发
	UnitMaxSize        uint64 `json:"unit_max_size"`        //一个单元最大允许多大
	//TransactionMaxSize uint64 `json:"tx_max_size"`          //一个交易最大允许多大
	FoundationAddress string `json:"foundation_address"` //基金会地址，该地址具有一些特殊权限，比如发起参数修改的投票，发起罚没保证金等

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
	MediatorCreateFee        uint64 `json:"mediator_create_fee"` //no use, delete
	AccountUpdateFee         uint64 `json:"account_update_fee"`
	TransferPtnBaseFee       uint64 `json:"transfer_ptn_base_fee"`
	TransferPtnPricePerKByte uint64 `json:"transfer_ptn_price_per_KByte"` //APP_DATA
}