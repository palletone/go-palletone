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
 *  * @date 2018
 *
 */

//SysConfig来自于系统合约SysConfigContractAddress的状态数据

package storage

import (
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/contracts/syscontract"
	"github.com/palletone/go-palletone/dag/modules"
)

//var CONF_PREFIX = append(constants.CONTRACT_STATE_PREFIX, scc.SysConfigContractAddress.Bytes()...)

/**
获取配置信息
get config information
*/
func (statedb *StateDb) GetConfig(name string) ([]byte, *modules.StateVersion, error) {
	id := syscontract.SysConfigContractAddress.Bytes21()
	return statedb.GetContractState(id, name)
}

/**
存储配置信息
*/
func (statedb *StateDb) SaveConfig(confs []modules.ContractWriteSet, stateVersion *modules.StateVersion) error {
	id := syscontract.SysConfigContractAddress.Bytes()
	log.Debugf("Save config into contract[%x]'s statedb", id)
	return statedb.SaveContractStates(id, confs, stateVersion)
}
func (statedb *StateDb) GetMinFee() (*modules.AmountAsset, error) {
	return &modules.AmountAsset{0, modules.NewPTNAsset()}, nil
}
