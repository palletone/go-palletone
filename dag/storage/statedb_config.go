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
	"encoding/json"
	"github.com/palletone/go-palletone/contracts/syscontract"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/modules"
)

//var CONF_PREFIX = append(constants.CONTRACT_STATE_PREFIX, scc.SysConfigContractAddress.Bytes()...)
func (statedb *StateDb) SaveSysConfig(key string, val []byte, ver *modules.StateVersion) error {
	//SaveContractState(id []byte, name string, value interface{}, version *modules.StateVersion)
	id := syscontract.SysConfigContractAddress.Bytes()
	err := saveContractState(statedb.db, id, key, val, ver)
	if err != nil {
		return err
	}
	return nil
}

/**
获取配置信息
get config information
*/
func (statedb *StateDb) GetSysConfig(name string) ([]byte, *modules.StateVersion, error) {
	id := syscontract.SysConfigContractAddress.Bytes()
	return statedb.GetContractState(id, name)
}
func (statedb *StateDb) GetAllSysConfig() (map[string]*modules.ContractStateValue, error) {
	id := syscontract.SysConfigContractAddress.Bytes()
	return statedb.GetContractStatesById(id)
}

/**
存储配置信息
*/
//func (statedb *StateDb) SaveConfig(confs []modules.ContractWriteSet, stateVersion *modules.StateVersion) error {
//	id := syscontract.SysConfigContractAddress.Bytes21()
//	log.Debugf("Save config into contract[%x]'s statedb", id)
//	return statedb.SaveContractStates(id, confs, stateVersion)
//}
func (statedb *StateDb) GetMinFee() (*modules.AmountAsset, error) {
	assetId := dagconfig.DagConfig.GetGasToken()
	return &modules.AmountAsset{Amount: 0, Asset: assetId.ToAsset()}, nil
}
func (statedb *StateDb) GetPartitionChains() ([]*modules.PartitionChain, error) {
	id := syscontract.PartitionContractAddress.Bytes()
	rows, err := statedb.GetContractStatesByPrefix(id, "PC")
	result := []*modules.PartitionChain{}
	if err != nil {
		return result, nil
	}

	for _, v := range rows {
		partition := &modules.PartitionChain{}
		json.Unmarshal(v.Value, &partition)
		result = append(result, partition)
	}
	return result, nil
}
func (statedb *StateDb) GetMainChain() (*modules.MainChain, error) {
	id := syscontract.PartitionContractAddress.Bytes()
	data, _, err := statedb.GetContractState(id, "MainChain")
	if err != nil {
		return nil, err
	}
	mainChain := &modules.MainChain{}
	err = json.Unmarshal(data, mainChain)
	if err != nil {
		return nil, err
	}
	return mainChain, nil
}
