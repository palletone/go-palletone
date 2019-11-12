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
	oldGp := &GlobalProperty104alpha{}
	err := storage.RetrieveFromRlpBytes(m.propdb, constants.GLOBALPROPERTY_KEY, oldGp)
	if err != nil {
		log.Errorf(err.Error())
		return err
	}

	newData := &modules.GlobalPropertyTemp{}

	newData.GlobalPropExtraTemp = oldGp.GlobalPropExtraTemp
	newData.ImmutableParameters = oldGp.ImmutableParameters
	newData.ChainParametersTemp.ChainParametersBase = oldGp.ChainParameters.ChainParametersBase
	newData.ChainParametersTemp.ChainParametersExtraTemp104alpha = oldGp.ChainParameters.ChainParametersExtraTemp104alpha

	//新加的
	newData.ChainParametersTemp.PledgeAllocateThreshold =
		strconv.FormatInt(int64(core.DefaultPledgeAllocateThreshold), 10)
	newData.ChainParametersTemp.PledgeRecordsThreshold =
		strconv.FormatInt(int64(core.DefaultPledgeRecordsThreshold), 10)

	// 修复在 1.0.2版本升级的初始化值的错误，重新改为0
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
	modules.GlobalPropExtraTemp
}

type GlobalPropBase104alpha struct {
	ImmutableParameters core.ImmutableChainParameters // 不可改变的区块链网络参数
	ChainParameters     ChainParameters104alpha       // 区块链网络参数
}

type ChainParameters104alpha struct {
	core.ChainParametersBase
	core.ChainParametersExtraTemp104alpha
}
