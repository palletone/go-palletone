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
	"github.com/palletone/go-palletone/common/rlp"
	"github.com/palletone/go-palletone/dag/modules"

	"encoding/json"
	"fmt"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/dag/constants"
	"strings"
)

//保存了对合约写集、Config、Asset信息
type StateDb struct {
	db     ptndb.Database
	logger log.ILogger
}

func NewStateDb(db ptndb.Database, l log.ILogger) *StateDb {
	return &StateDb{db: db, logger: l}
}

// ######################### SAVE IMPL START ###########################

func (statedb *StateDb) SaveAssetInfo(assetInfo *modules.AssetInfo) error {
	key := assetInfo.Tokey()
	return StoreBytes(statedb.db, key, assetInfo)
}

func (statedb *StateDb) DeleteState(key []byte) error {
	return statedb.db.Delete(key)
}

// ######################### SAVE IMPL END ###########################

// ######################### GET IMPL START ###########################

func (statedb *StateDb) GetAssetInfo(assetId *modules.Asset) (*modules.AssetInfo, error) {
	key := append(constants.ASSET_INFO_PREFIX, assetId.AssetId.String()...)
	data, err := statedb.db.Get(key)
	if err != nil {
		return nil, err
	}

	var assetInfo modules.AssetInfo
	err = rlp.DecodeBytes(data, &assetInfo)

	if err != nil {
		return nil, err
	}
	return &assetInfo, nil
}

// get prefix: return maps
func (db *StateDb) GetPrefix(prefix []byte) map[string][]byte {
	return getprefix(db.db, prefix)
}

// ######################### GET IMPL END ###########################

// author albert·gou
func (statedb *StateDb) StoreMediator(med *core.Mediator) error {
	return StoreMediator(statedb.db, med)
}

// author albert·gou
func (statedb *StateDb) StoreMediatorInfo(add common.Address, mi *MediatorInfo) error {
	return StoreMediatorInfo(statedb.db, add, mi)
}

func (statedb *StateDb) RetrieveMediatorInfo(address common.Address) (*MediatorInfo, error) {
	return RetrieveMediatorInfo(statedb.db, address)
}

// author albert·gou
func (statedb *StateDb) RetrieveMediator(address common.Address) (*core.Mediator, error) {
	return RetrieveMediator(statedb.db, address)
}

func (statedb *StateDb) SaveChainIndex(index *modules.ChainIndex) error {
	bytes, err := rlp.EncodeToBytes(index)
	if err != nil {
		return err
	}
	key := constants.CURRENTCHAININDEX_PREFIX + index.AssetID.String()
	if err := statedb.db.Put([]byte(key), bytes); err != nil {
		return err
	}
	return nil
}
func (statedb *StateDb) GetCurrentChainIndex(assetId modules.IDType16) (*modules.ChainIndex, error) {
	// get current chainIndex
	key := constants.CURRENTCHAININDEX_PREFIX + assetId.String()
	bytes, err := statedb.db.Get([]byte(key))
	if err != nil {
		return nil, err
	}
	chainIndex := new(modules.ChainIndex)
	if err := rlp.DecodeBytes(bytes, &chainIndex); err != nil {
		return nil, err
	}
	return chainIndex, nil
}

// author albert·gou
func (statedb *StateDb) GetMediatorCount() int {
	return GetMediatorCount(statedb.db)
}

// author albert·gou
func (statedb *StateDb) IsMediator(address common.Address) bool {
	return IsMediator(statedb.db, address)
}

// author albert·gou
func (statedb *StateDb) GetMediators() map[common.Address]bool {
	return GetMediators(statedb.db)
}

// author albert·gou
func (statedb *StateDb) LookupMediator() map[common.Address]*core.Mediator {
	return LookupMediator(statedb.db)
}

//xiaozhi
func (statedb *StateDb) GetMediatorCandidateList() ([]string, error) {
	depositeContractAddress := common.HexToAddress("0x00000000000000000000000000000000000000011C")
	_, val := statedb.GetContractState(depositeContractAddress.Bytes(), "MediatorList")
	if val == nil {
		return nil, fmt.Errorf("mediator candidate list is nil.")
	}
	var candidateList []string
	err := json.Unmarshal(val, &candidateList)
	if err != nil {
		return nil, err
	}
	return candidateList, nil
}

func (statedb *StateDb) IsInMediatorCandidateList(address common.Address) bool {
	list, err := statedb.GetMediatorCandidateList()
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
