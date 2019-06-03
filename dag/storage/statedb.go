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
 *  * @author PalletOne core developer  <dev@pallet.one>
 *  * @date 2018
 *
 */

package storage

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/contracts/syscontract"
	"github.com/palletone/go-palletone/dag/modules"
)

//保存了对合约写集、Config、Asset信息
type StateDb struct {
	db ptndb.Database
}

func NewStateDb(db ptndb.Database) *StateDb {
	return &StateDb{db: db}
}

func storeBytesWithVersion(db ptndb.Putter, key []byte, version *modules.StateVersion, val []byte) error {
	v := append(version.Bytes(), val...)
	if err := db.Put(key, v); err != nil {
		return err
	}
	return nil
}

func retrieveWithVersion(db ptndb.Database, key []byte) ([]byte, *modules.StateVersion, error) {
	data, err := db.Get(key)
	if err != nil {
		return nil, nil, err
	}

	return splitValueAndVersion(data)
}

//将Statedb里的Value分割为Version和用户数据
func splitValueAndVersion(data []byte) ([]byte, *modules.StateVersion, error) {
	if len(data) < 28 {
		return nil, nil, errors.New("the data is irregular.")
	}
	verBytes := data[:28]
	objData := data[28:]
	version := &modules.StateVersion{}
	version.SetBytes(verBytes)
	return objData, version, nil
}

// ######################### SAVE IMPL START ###########################

//func (statedb *StateDb) SaveAssetInfo(assetInfo *modules.AssetInfo) error {
//	key := assetInfo.Tokey()
//	return StoreToRlpBytes(statedb.db, key, assetInfo)
//}

func (statedb *StateDb) DeleteState(key []byte) error {
	return statedb.db.Delete(key)
}

// ######################### SAVE IMPL END ###########################

// ######################### GET IMPL START ###########################

//func (statedb *StateDb) GetAssetInfo(assetId *modules.Asset) (*modules.AssetInfo, error) {
//	key := append(constants.ASSET_INFO_PREFIX, assetId.AssetId.String()...)
//	data, err := statedb.db.Get(key)
//	if err != nil {
//		return nil, err
//	}
//
//	var assetInfo modules.AssetInfo
//	err = rlp.DecodeBytes(data, &assetInfo)
//
//	if err != nil {
//		return nil, err
//	}
//	return &assetInfo, nil
//}

// get prefix: return maps
func (db *StateDb) GetPrefix(prefix []byte) map[string][]byte {
	return getprefix(db.db, prefix)
}

// ######################### GET IMPL END ###########################

func (statedb *StateDb) GetJuryCandidateList() (map[string]bool, error) {
	depositeContractAddress := syscontract.DepositContractAddress
	val, _, err := statedb.GetContractState(depositeContractAddress.Bytes(), modules.JuryList)
	if err != nil {
		return nil, fmt.Errorf("jury candidate list is nil.")
	}
	//var candidateList []common.Address
	candidateList := make(map[string]bool)
	err = json.Unmarshal(val, &candidateList)
	if err != nil {
		return nil, err
	}
	return candidateList, nil
}

func (statedb *StateDb) IsInJuryCandidateList(address common.Address) bool {
	list, err := statedb.GetJuryCandidateList()
	if err != nil {
		return false
	}
	if _, ok := list[address.String()]; ok {
		return true
	}
	//for _, v := range list {
	//	if strings.Compare(v.String(), address.String()) == 0 {
	//		return true
	//	}
	//}
	return false
}

func (statedb *StateDb) UpdateSysParams(version *modules.StateVersion) error {
	//基金会单独修改的
	var err error
	modifies, err := statedb.GetSysParamWithoutVote()
	if err != nil {
		return err
	}
	//基金会发起投票的
	info, err := statedb.GetSysParamsWithVotes()
	if err != nil {
		return err
	}
	if modifies == nil && info == nil {
		return nil
	}
	//获取当前的version
	if len(modifies) > 0 {
		for k, v := range modifies {
			err = statedb.SaveSysConfig(k, []byte(v), version)
			if err != nil {
				return err
			}
		}
		//将基金会当前单独修改的重置为nil
		err = statedb.SaveSysConfig(modules.DesiredSysParamsWithoutVote, nil, version)
		if err != nil {
			return err
		}
	}
	if info == nil {
		return nil
	}
	//foundAddr, _, err := statedb.GetSysConfig(modules.FoundationAddress)
	//if err != nil {
	//	return err
	//}
	//if info.CreateAddr != string(foundAddr) {
	//	return fmt.Errorf("only foundation can call this function")
	//}
	if !info.IsVoteEnd {
		return nil
	}
	for _, v1 := range info.SupportResults {
		for _, v2 := range v1.VoteResults {
			//TODO
			if v2.Num >= info.LeastNum {
				err = statedb.SaveSysConfig(v1.TopicTitle, []byte(v2.SelectOption), version)
				if err != nil {
					return err
				}
				break
			}
		}
	}
	//将基金会当前投票修改的重置为nil
	err = statedb.SaveSysConfig(modules.DesiredSysParamsWithVote, nil, version)
	if err != nil {
		return err
	}
	return nil
}

func (statedb *StateDb) GetSysParamWithoutVote() (map[string]string, error) {
	var res map[string]string

	val, _, err := statedb.GetSysConfig(modules.DesiredSysParamsWithoutVote)
	if err != nil {
		log.Debugf(err.Error())
		return nil, err
	}

	err = json.Unmarshal(val, &res)
	if err != nil {
		log.Debugf(err.Error())
		return nil, err
	}

	return res, nil
}

func (statedb *StateDb) GetSysParamsWithVotes() (*modules.SysTokenIDInfo, error) {
	val, _, err := statedb.GetSysConfig(modules.DesiredSysParamsWithVote)
	if err != nil {
		return nil, err
	}
	info := &modules.SysTokenIDInfo{}
	if val == nil {
		return nil, err
	} else if len(val) > 0 {
		err := json.Unmarshal(val, info)
		if err != nil {
			return nil, err
		}
		return info, nil
	} else {
		return nil, nil
	}
}
