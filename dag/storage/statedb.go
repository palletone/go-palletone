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

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/dag/constants"
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
	key := append(modules.ASSET_INFO_PREFIX, assetId.AssetId.String()...)
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
func (statedb *StateDb) GetCandidateMediatorAddrList() ([]common.Address, error) {
	key := constants.STATE_CANDIDATE_MEDIATOR_LIST
	data, _, err := retrieveWithVersion(statedb.db, key)
	if err != nil {
		return nil, err
	}
	result := []common.Address{}
	rlp.DecodeBytes(data, result)
	return result, nil
}
func (statedb *StateDb) SaveCandidateMediatorAddrList(addrs []common.Address, v *modules.StateVersion) error {
	key := constants.STATE_CANDIDATE_MEDIATOR_LIST
	addrsStr := ""
	for _, addr := range addrs {
		addrsStr += addr.String() + ","
	}
	statedb.logger.Debugf("Try to save candidate mediator address list:%s", addrsStr)
	return StoreBytesWithVersion(statedb.db, key, v, addrs)
}
func (statedb *StateDb) GetActiveMediatorAddrList() ([]common.Address, error) {

	key := constants.STATE_ACTIVE_MEDIATOR_LIST
	data, _, err := retrieveWithVersion(statedb.db, key)
	if err != nil {
		return nil, err
	}
	result := []common.Address{}
	rlp.DecodeBytes(data, result)
	return result, nil
}

// ######################### GET IMPL END ###########################
