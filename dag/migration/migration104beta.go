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
 *  * @date 2019-2020
 *
 */
package migration

import (
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/dag/constants"
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
	oldGp := &GlobalProperty104alpha{}
	err := storage.RetrieveFromRlpBytes(m.propdb, constants.GLOBALPROPERTY_KEY, oldGp)
	if err != nil {
		log.Errorf(err.Error())
		return err
	}

	newData := &GlobalProperty105delta{}

	//newData.GlobalPropExtraTemp = oldGp.GlobalPropExtraTemp
	newData.ActiveJuries = oldGp.ActiveJuries
	newData.ActiveMediators = oldGp.ActiveMediators
	newData.PrecedingMediators = oldGp.PrecedingMediators

	newData.ImmutableParameters = oldGp.ImmutableParameters
	newData.ChainParametersTemp.ChainParametersBase = oldGp.ChainParametersTemp.ChainParametersBase
	//newData.ChainParametersTemp.ChainParametersExtraTemp104alpha =
	//	oldGp.ChainParametersTemp.ChainParametersExtraTemp104alpha

	newData.ChainParametersTemp.UccMemory = oldGp.ChainParametersTemp.UccMemory
	newData.ChainParametersTemp.UccCpuShares = oldGp.ChainParametersTemp.UccCpuShares
	newData.ChainParametersTemp.UccCpuQuota = oldGp.ChainParametersTemp.UccCpuQuota
	newData.ChainParametersTemp.UccDisk = oldGp.ChainParametersTemp.UccDisk
	newData.ChainParametersTemp.UccDuringTime = oldGp.ChainParametersTemp.UccDuringTime
	newData.ChainParametersTemp.TempUccMemory = oldGp.ChainParametersTemp.TempUccMemory
	newData.ChainParametersTemp.TempUccCpuShares = oldGp.ChainParametersTemp.TempUccCpuShares
	newData.ChainParametersTemp.TempUccCpuQuota = oldGp.ChainParametersTemp.TempUccCpuQuota
	newData.ChainParametersTemp.ContractSystemVersion = oldGp.ChainParametersTemp.ContractSystemVersion
	newData.ChainParametersTemp.ContractSignatureNum = oldGp.ChainParametersTemp.ContractSignatureNum
	newData.ChainParametersTemp.ContractElectionNum = oldGp.ChainParametersTemp.ContractElectionNum
	newData.ChainParametersTemp.ContractTxTimeoutUnitFee = oldGp.ChainParametersTemp.ContractTxTimeoutUnitFee
	newData.ChainParametersTemp.ContractTxSizeUnitFee = oldGp.ChainParametersTemp.ContractTxSizeUnitFee
	newData.ChainParametersTemp.ContractTxInstallFeeLevel = oldGp.ChainParametersTemp.ContractTxInstallFeeLevel
	newData.ChainParametersTemp.ContractTxDeployFeeLevel = oldGp.ChainParametersTemp.ContractTxDeployFeeLevel
	newData.ChainParametersTemp.ContractTxInvokeFeeLevel = oldGp.ChainParametersTemp.ContractTxInvokeFeeLevel
	newData.ChainParametersTemp.ContractTxStopFeeLevel = oldGp.ChainParametersTemp.ContractTxStopFeeLevel

	//新加的
	newData.ChainParametersTemp.PledgeAllocateThreshold =
		strconv.FormatInt(int64(core.DefaultPledgeAllocateThreshold), 10)
	newData.ChainParametersTemp.PledgeRecordsThreshold =
		strconv.FormatInt(int64(core.DefaultPledgeRecordsThreshold), 10)

	// 修复在 1.0.2版本升级时初始化值的错误，重新改为0
	newData.ImmutableParameters.MinMaintSkipSlots = core.DefaultMinMaintSkipSlots

	err = storage.StoreToRlpBytes(m.propdb, constants.GLOBALPROPERTY_KEY, newData)
	if err != nil {
		log.Errorf(err.Error())
		return err
	}

	return nil
}

type GlobalProperty104alpha struct {
	GlobalPropBase104alpha
	//modules.GlobalPropExtraTemp

	ActiveJuries       []common.Address
	ActiveMediators    []common.Address
	PrecedingMediators []common.Address
}

type GlobalPropBase104alpha struct {
	ImmutableParameters core.ImmutableChainParameters // 不可改变的区块链网络参数
	ChainParametersTemp ChainParameters104alpha       // 区块链网络参数
}

type ChainParameters104alpha struct {
	core.ChainParametersBase
	//core.ChainParametersExtraTemp104alpha

	UccMemory                 string
	UccCpuShares              string
	UccCpuQuota               string
	UccDisk                   string
	UccDuringTime             string
	TempUccMemory             string
	TempUccCpuShares          string
	TempUccCpuQuota           string
	ContractSystemVersion     string
	ContractSignatureNum      string
	ContractElectionNum       string
	ContractTxTimeoutUnitFee  string
	ContractTxSizeUnitFee     string
	ContractTxInstallFeeLevel string
	ContractTxDeployFeeLevel  string
	ContractTxInvokeFeeLevel  string
	ContractTxStopFeeLevel    string
}
