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
	"github.com/palletone/go-palletone/common/ptndb"

	"github.com/palletone/go-palletone/dag/modules"

	"encoding/json"
	"fmt"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/contracts/syscontract"
	"github.com/palletone/go-palletone/core"
	"strings"
)

//保存了对合约写集、Config、Asset信息
type StateDb struct {
	db ptndb.Database
}

func NewStateDb(db ptndb.Database) *StateDb {
	return &StateDb{db: db}
}

// ######################### SAVE IMPL START ###########################

//func (statedb *StateDb) SaveAssetInfo(assetInfo *modules.AssetInfo) error {
//	key := assetInfo.Tokey()
//	return StoreBytes(statedb.db, key, assetInfo)
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

func (statedb *StateDb) StoreMediator(med *core.Mediator) error {
	return StoreMediator(statedb.db, med)
}

func (statedb *StateDb) StoreMediatorInfo(add common.Address, mi *modules.MediatorInfo) error {
	return StoreMediatorInfo(statedb.db, add, mi)
}

func (statedb *StateDb) RetrieveMediatorInfo(address common.Address) (*modules.MediatorInfo, error) {
	return RetrieveMediatorInfo(statedb.db, address)
}

func (statedb *StateDb) RetrieveMediator(address common.Address) (*core.Mediator, error) {
	return RetrieveMediator(statedb.db, address)
}

//func (statedb *StateDb) SaveChainIndex(index *modules.ChainIndex) error {
//	bytes, err := rlp.EncodeToBytes(index)
//	if err != nil {
//		return err
//	}
//	key := constants.CURRENTCHAININDEX_PREFIX + index.AssetID.String()
//	if err := statedb.db.Put([]byte(key), bytes); err != nil {
//		return err
//	}
//	return nil
//}
//func (statedb *StateDb) GetCurrentChainIndex(assetId modules.AssetId) (*modules.ChainIndex, error) {
//	// get current chainIndex
//	key := constants.CURRENTCHAININDEX_PREFIX + assetId.String()
//	bytes, err := statedb.db.Get([]byte(key))
//	if err != nil {
//		return nil, err
//	}
//	chainIndex := new(modules.ChainIndex)
//	if err := rlp.DecodeBytes(bytes, &chainIndex); err != nil {
//		return nil, err
//	}
//	return chainIndex, nil
//}

func (statedb *StateDb) GetMediatorCount() int {
	return GetMediatorCount(statedb.db)
}

func (statedb *StateDb) IsMediator(address common.Address) bool {
	return IsMediator(statedb.db, address)
}

func (statedb *StateDb) GetMediators() map[common.Address]bool {
	return GetMediators(statedb.db)
}

func (statedb *StateDb) LookupMediator() map[common.Address]*core.Mediator {
	return LookupMediator(statedb.db)
}

//xiaozhi
func (statedb *StateDb) GetApprovedMediatorList() ([]*modules.MediatorRegisterInfo, error) {
	depositeContractAddress := syscontract.DepositContractAddress
	val, _, err := statedb.GetContractState(depositeContractAddress.Bytes(), "MediatorList")
	if err != nil {
		return nil, fmt.Errorf("mediator candidate list is nil.")
	}
	var candidateList []*modules.MediatorRegisterInfo
	err = json.Unmarshal(val, &candidateList)
	if err != nil {
		return nil, err
	}
	return candidateList, nil
}

func (statedb *StateDb) IsApprovedMediator(address common.Address) bool {
	list, err := statedb.GetApprovedMediatorList()
	if err != nil {
		return false
	}
	for _, v := range list {
		if strings.Compare(v.Address, address.String()) == 0 {
			return true
		}
	}
	return false
}

func (statedb *StateDb) GetJuryCandidateList() ([]string, error) {
	depositeContractAddress := syscontract.DepositContractAddress
	val, _, err := statedb.GetContractState(depositeContractAddress.Bytes(), "JuryList")
	if err != nil {
		return nil, fmt.Errorf("jury candidate list is nil.")
	}
	var candidateList []string
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
	for _, v := range list {
		if strings.Compare(v, address.String()) == 0 {
			return true
		}
	}
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
		for _, v := range modifies {
			err = statedb.SaveSysConfig(v.Key, []byte(v.Value), version)
			if err != nil {
				return err
			}
		}
		//将基金会当前单独修改的制为nil
		err = statedb.SaveSysConfig("sysParam", nil, version)
		if err != nil {
			return err
		}
	}
	if info == nil {
		return nil
	}
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
	//将基金会当前投票修改的制为nil
	err = statedb.SaveSysConfig("sysParams", nil, version)
	if err != nil {
		return err
	}
	return nil
}

func (statedb *StateDb) GetSysParamWithoutVote() ([]*modules.FoundModify, error) {
	val, _, err := statedb.GetConfig("sysParam")
	if err != nil {
		return nil, err
	}
	var modifies []*modules.FoundModify
	if val == nil {
		return nil, err
	} else if len(val) > 0 {
		err := json.Unmarshal(val, &modifies)
		if err != nil {
			return nil, err
		}
		return modifies, nil
	} else {
		return nil, nil
	}
}

func (statedb *StateDb) GetSysParamsWithVotes() (*modules.SysTokenIDInfo, error) {
	val, _, err := statedb.GetConfig("sysParams")
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

func (statedb *StateDb) SaveSysConfig(key string, val []byte, ver *modules.StateVersion) error {
	//SaveContractState(id []byte, name string, value interface{}, version *modules.StateVersion)
	id := syscontract.SysConfigContractAddress.Bytes21()
	err := statedb.SaveContractState(id, key, val, ver)
	if err != nil {
		return err
	}
	return nil
}
