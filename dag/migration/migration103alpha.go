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

type Migration102delta_103alpha struct {
	dagdb   ptndb.Database
	idxdb   ptndb.Database
	utxodb  ptndb.Database
	statedb ptndb.Database
	propdb  ptndb.Database
}

func (m *Migration102delta_103alpha) FromVersion() string {
	return "1.0.2-release"
}

func (m *Migration102delta_103alpha) ToVersion() string {
	return "1.0.3-alpha"
}

func (m *Migration102delta_103alpha) ExecuteUpgrade() error {
	//添加两个系统参数，转换GLOBALPROPERTY结构体
	if err := m.upgradeGP(); err != nil {
		return err
	}

	return nil
}

func (m *Migration102delta_103alpha) upgradeGP() error {
	oldGp := &GlobalProperty102delta{}
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

	//新加的
	newData.ChainParameters.RmExpConFromSysParam = core.DefaultRmExpConFromSysParam
	newData.ChainParameters.ContractSystemVersion = core.DefaultContractSystemVersion

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

	//新加的
	newData.ChainParameters.UccDuringTime = core.DefaultContainerDuringTime

	err = storage.StoreToRlpBytes(m.propdb, constants.GLOBALPROPERTY_KEY, newData)
	if err != nil {
		log.Errorf(err.Error())
		return err
	}

	return nil
}

type GlobalProperty102delta struct {
	GlobalPropBase102delta
	ActiveJuries       []common.Address
	ActiveMediators    []common.Address
	PrecedingMediators []common.Address
}

type GlobalPropBase102delta struct {
	ImmutableParameters core.ImmutableChainParameters // 不可改变的区块链网络参数
	ChainParameters     ChainParameters102delta       // 区块链网络参数
}

type ChainParameters102delta struct {
	ChainParametersBase102delta

	UccMemory string
	UccCpuShares string
	UccCpuQuota  string
	UccDisk      string

	TempUccMemory    string
	TempUccCpuShares string
	TempUccCpuQuota  string

	ContractSignatureNum string
	ContractElectionNum  string

	ContractTxTimeoutUnitFee  string
	ContractTxSizeUnitFee     string

	ContractTxInstallFeeLevel string
	ContractTxDeployFeeLevel  string
	ContractTxInvokeFeeLevel  string
	ContractTxStopFeeLevel    string
}

type ChainParametersBase102delta struct {
	GenerateUnitReward uint64 `json:"generate_unit_reward"` //每生产一个单元，奖励多少Dao的PTN
	PledgeDailyReward  uint64 `json:"pledge_daily_reward"`  //质押金的日奖励额
	RewardHeight       uint64 `json:"reward_height"`        //每多少高度进行一次奖励的派发
	UnitMaxSize        uint64 `json:"unit_max_size"`        //一个单元最大允许多大
	FoundationAddress  string `json:"foundation_address"`   //基金会地址，该地址具有一些特殊权限，比如发起参数修改的投票，发起罚没保证金等

	DepositAmountForMediator  uint64 `json:"deposit_amount_for_mediator"` //保证金的数量
	DepositAmountForJury      uint64 `json:"deposit_amount_for_jury"`
	DepositAmountForDeveloper uint64 `json:"deposit_amount_for_developer"`

	// 活跃mediator的数量。 number of active mediators
	ActiveMediatorCount uint8 `json:"active_mediator_count"`

	// 用户可投票mediator的最大数量。the maximum number of mediator users can vote for
	MaximumMediatorCount uint8 `json:"max_mediator_count"`

	// unit生产之间的间隔时间，以秒为单元。 interval in seconds between Units
	MediatorInterval uint8 `json:"mediator_interval"`

	// 区块链维护事件之间的间隔，以秒为单元。 interval in sections between unit maintenance events
	MaintenanceInterval uint32 `json:"maintenance_interval"`

	// 在维护时跳过的MediatorInterval数量。 number of MediatorInterval to skip at maintenance time
	MaintenanceSkipSlots uint8 `json:"maintenance_skip_slots"`

	// 目前的操作交易费，current schedule of fees
	MediatorCreateFee        uint64 `json:"mediator_create_fee"`
	AccountUpdateFee         uint64 `json:"account_update_fee"`
	TransferPtnBaseFee       uint64 `json:"transfer_ptn_base_fee"`
	TransferPtnPricePerKByte uint64 `json:"transfer_ptn_price_per_KByte"`
}
