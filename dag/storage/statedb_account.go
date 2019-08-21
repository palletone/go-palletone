/*
 *
 *     This file is part of go-palletone.
 *     go-palletone is free software: you can redistribute it and/or modify
 *     it under the terms of the GNU General Public License as published by
 *     the Free Software Foundation, either version 3 of the License, or
 *     (at your option) any later version.
 *     go-palletone is distributed in the hope that it will be useful,
 *     but WITHOUT ANY WARRANTY; without even the implied warranty of
 *     MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *     GNU General Public License for more details.
 *     You should have received a copy of the GNU General Public License
 *     along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
 * /
 *
 *  * @author PalletOne core developer Albert·Gou <dev@pallet.one>
 *  * @date 2018
 *
 */

package storage

import (
	"bytes"
	"encoding/json"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/dag/constants"
	"github.com/palletone/go-palletone/dag/modules"
)

func accountKey(address common.Address) []byte {
	key := append(constants.ACCOUNT_INFO_PREFIX, address.Bytes()...)

	return key
}

func ptnBalanceKey(address common.Address) []byte {
	key := append(constants.ACCOUNT_PTN_BALANCE_PREFIX, address.Bytes()...)

	return key
}

func (statedb *StateDb) UpdateAccountBalance(address common.Address, addAmount int64) error {
	balance := statedb.GetAccountBalance(address)
	balance = uint64(int64(balance) + addAmount)

	return statedb.db.Put(ptnBalanceKey(address), Uint64ToBytes(balance))
}

func (statedb *StateDb) GetAccountBalance(address common.Address) uint64 {
	balance := uint64(0)

	data, err := statedb.db.Get(ptnBalanceKey(address))
	if err == nil {
		balance = BytesToUint64(data)
	}

	return balance
}

func (statedb *StateDb) LookupAccount() map[common.Address]*modules.AccountInfo {
	result := make(map[common.Address]*modules.AccountInfo)

	iter := statedb.db.NewIteratorWithPrefix(constants.ACCOUNT_PTN_BALANCE_PREFIX)
	for iter.Next() {
		balance := BytesToUint64(iter.Value())
		acc := &modules.AccountInfo{Balance: balance}

		//get address voted mediators
		addB := bytes.TrimPrefix(iter.Key(), constants.ACCOUNT_PTN_BALANCE_PREFIX)
		add := common.NewAddress(addB, common.PublicKeyHash)
		acc.VotedMediators = statedb.GetAccountVotedMediators(add)

		result[add] = acc
	}

	return result
}

func (statedb *StateDb) GetAccountVotedMediators(addr common.Address) map[string]bool {
	votedMediators := make(map[string]bool)

	data, err := statedb.GetAccountState(addr, constants.VOTED_MEDIATORS)
	if err != nil {
		return votedMediators
	}

	err = json.Unmarshal(data.Value, &votedMediators)
	if err != nil {
		return votedMediators
	}

	return votedMediators
}

func (statedb *StateDb) GetAccountState(address common.Address, statekey string) (*modules.ContractStateValue, error) {
	key := append(accountKey(address), statekey...)
	data, version, err := retrieveWithVersion(statedb.db, key)
	if err != nil {
		return nil, err
	}

	return &modules.ContractStateValue{Value: data, Version: version}, nil
}

func (statedb *StateDb) SaveAccountState(address common.Address, write *modules.AccountStateWriteSet,
	version *modules.StateVersion) error {
	key := append(accountKey(address), write.Key...)

	var err error
	if write.IsDelete {
		err = statedb.db.Delete(key)
	} else {
		err = storeBytesWithVersion(statedb.db, key, version, write.Value)
	}

	if err != nil {
		return err
	}

	return nil
}

func (statedb *StateDb) GetAllAccountStates(address common.Address) (map[string]*modules.ContractStateValue, error) {
	key := accountKey(address)
	data := getprefix(statedb.db, key)
	var err error
	result := make(map[string]*modules.ContractStateValue)
	for dbkey, state_version := range data {
		state, version, err0 := splitValueAndVersion(state_version)
		if err0 != nil {
			err = err0
		}
		realKey := dbkey[len(key):]
		if realKey != "" {
			result[realKey] = &modules.ContractStateValue{Value: state, Version: version}
		}
	}
	return result, err
}

func (statedb *StateDb) SaveAccountStates(address common.Address, writeset []modules.AccountStateWriteSet,
	version *modules.StateVersion) error {
	batch := statedb.db.NewBatch()

	for _, write := range writeset {
		key := append(accountKey(address), write.Key...)
		if write.IsDelete {
			err := batch.Delete(key)
			if err != nil {
				log.Debugf(err.Error())
			}
		} else {
			err := storeBytesWithVersion(batch, key, version, write.Value)
			if err != nil {
				log.Debugf(err.Error())
			}
		}
	}

	return batch.Write()
}
