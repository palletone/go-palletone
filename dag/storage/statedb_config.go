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
	"fmt"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/dag/constants"

	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/contracts/syscontract"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/modules"
)

//var CONF_PREFIX = append(constants.CONTRACT_STATE_PREFIX, scc.SysConfigContractAddress.Bytes()...)
func (statedb *StateDb) SaveSysConfigContract(key string, val []byte, ver *modules.StateVersion) error {
	id := syscontract.SysConfigContractAddress.Bytes()
	write := modules.NewWriteSet(key, val)
	err := statedb.SaveContractState(id, write, ver)
	if err != nil {
		log.Error(err.Error())
		return err
	}

	return nil
}

/**
获取配置信息
get config information
*/
func (statedb *StateDb) getSysConfigContract(name string) ([]byte, *modules.StateVersion, error) {
	id := syscontract.SysConfigContractAddress.Bytes()
	return statedb.GetContractState(id, name)
}

func (statedb *StateDb) GetSysParamWithoutVote() (map[string]string, error) {
	val, _, err := statedb.getSysConfigContract(modules.DesiredSysParamsWithoutVote)
	if err != nil {
		log.Debugf(err.Error())
		return nil, err
	}

	if len(val) == 0 {
		return nil, fmt.Errorf("data is nil")
	}

	var res map[string]string
	err = json.Unmarshal(val, &res)
	if err != nil {
		log.Debugf(err.Error())
		return nil, err
	}

	return res, nil
}

func (statedb *StateDb) GetSysParamsWithVotes() (*modules.SysTokenIDInfo, error) {
	val, _, err := statedb.getSysConfigContract(modules.DesiredSysParamsWithVote)
	if err != nil {
		log.Debugf(err.Error())
		return nil, err
	}

	if len(val) == 0 {
		return nil, fmt.Errorf("data is nil")
	}

	info := &modules.SysTokenIDInfo{}
	err = json.Unmarshal(val, info)
	if err != nil {
		log.Debugf(err.Error())
		return nil, err
	}

	return info, nil
}

func (statedb *StateDb) GetMinFee() (*modules.AmountAsset, error) {
	assetId := dagconfig.DagConfig.GetGasToken()
	return &modules.AmountAsset{Amount: 1, Asset: assetId.ToAsset()}, nil
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
		err = json.Unmarshal(v.Value, &partition)
		if err != nil {
			return nil, err
		}
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
func (statedb *StateDb) GetBlacklistAddress() ([]common.Address, *modules.StateVersion, error) {
	id := syscontract.BlacklistContractAddress.Bytes()
	data, v, err := statedb.GetContractState(id, constants.BlacklistAddress)
	if err != nil { //未初始化黑名单
		log.Debug("Don't have blacklist:" + err.Error())
		return []common.Address{}, nil, nil
	}
	result := []common.Address{}
	err = rlp.DecodeBytes(data, &result)
	log.DebugDynamic(func() string {
		data, _ := json.Marshal(result)
		return "query blacklist result is:" + string(data)
	})
	return result, v, err
}
