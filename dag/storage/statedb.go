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

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/contracts/syscontract"
	"github.com/palletone/go-palletone/dag/constants"
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

func (statedb *StateDb) DeleteState(key []byte) error {
	return statedb.db.Delete(key)
}

// get prefix: return maps
func (db *StateDb) GetPrefix(prefix []byte) map[string][]byte {
	return getprefix(db.db, prefix)
}

func (statedb *StateDb) GetJuryCandidateList() (map[string]bool, error) {
	depositeContractAddress := syscontract.DepositContractAddress
	val, _, err := statedb.GetContractState(depositeContractAddress.Bytes(), modules.JuryList)
	if err != nil {
		return nil, fmt.Errorf("jury candidate list is nil.")
	}

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
	return false
}

func (statedb *StateDb) GetAllJuror() (map[string]*modules.Juror, error) {
	depositeContractAddress := syscontract.DepositContractAddress
	juryvalues, err := statedb.GetContractStatesByPrefix(depositeContractAddress.Bytes(), string(constants.DEPOSIT_JURY_BALANCE_PREFIX))
	if err != nil {
		return nil, err
	}
	if len(juryvalues) > 0 {
		jurynode := make(map[string]*modules.Juror)
		for a, v := range juryvalues {
			n := modules.Juror{}
			err := json.Unmarshal(v.Value, &n)
			if err != nil {
				return nil, err
			}
			jurynode[a] = &n
		}
		return jurynode, nil
	}
	return nil, nil
}

func (statedb *StateDb) GetJurorByAddr(addr string) (*modules.Juror, error) {
	depositeContractAddress := syscontract.DepositContractAddress
	key := string(constants.DEPOSIT_JURY_BALANCE_PREFIX) + addr
	val, _, err := statedb.GetContractState(depositeContractAddress.Bytes(), key)
	if err != nil {
		return nil, err
	}
	juror := &modules.Juror{}
	err = json.Unmarshal(val, juror)
	if err != nil {
		return nil, err
	}
	return juror, nil
}

func (statedb *StateDb) GetContractDeveloperList() ([]common.Address, error) {
	depositeContractAddress := syscontract.DepositContractAddress
	val, _, err := statedb.GetContractState(depositeContractAddress.Bytes(), modules.DeveloperList)
	if err != nil {
		return nil, fmt.Errorf("devCc candidate list is nil.")
	}
	depList := make(map[string]bool)
	err = json.Unmarshal(val, &depList)
	if err != nil {
		return nil, err
	}
	res := make([]common.Address, len(depList))
	for addStr := range depList {
		add, err := common.StringToAddress(addStr)
		if err != nil {
			log.Debugf(err.Error())
			continue
		}
		res = append(res, add)
	}
	return res, nil
}

func (statedb *StateDb) IsInContractDeveloperList(address common.Address) bool {
	list, err := statedb.GetContractDeveloperList()
	if err != nil {
		return false
	}
	for _, d := range list {
		if d.Equal(address) {
			return true
		}
	}
	return false
}

func (statedb *StateDb) GetDataVersion() (*modules.DataVersion, error) {
	data, err := statedb.db.Get(constants.DATA_VERSION_KEY)
	if err != nil {
		return nil, err
	}
	data_version := new(modules.DataVersion)
	if err := rlp.DecodeBytes(data, data_version); err != nil {
		return nil, err
	}
	return data_version, nil
}

func (statedb *StateDb) SaveDataVersion(dv *modules.DataVersion) error {
	data, err := rlp.EncodeToBytes(dv)
	if err != nil {
		return err
	}
	return statedb.db.Put(constants.DATA_VERSION_KEY, data)
}
