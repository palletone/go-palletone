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
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/dag/constants"
	"github.com/palletone/go-palletone/dag/modules"
)

func accountKey(address common.Address) []byte {
	key := append(constants.ACCOUNT_INFO_PREFIX, address.Bytes21()...)

	return key
}

func (statedb *StateDb) UpdateAccountBalance(address common.Address, addAmount int64) error {
	key := append(constants.ACCOUNT_PTN_BALANCE_PREFIX, address.Bytes21()...)
	balance := uint64(0)
	data, err := statedb.db.Get(key)
	if err != nil {
		// 第一次更新时， 数据库没有该账户的相关数据
		log.Debugf("Account balance for [%s] don't exist,create it first", address.String())
	} else {
		balance = BytesToUint64(data)
	}
	//log.Debugf("Update Ptn Balance for address:%s, add Amount:%d", address.String(), addAmount)
	balance = uint64(int64(balance) + addAmount)
	return statedb.db.Put(key, Uint64ToBytes(balance))
}

func (statedb *StateDb) GetAccountBalance(address common.Address) uint64 {
	key := append(constants.ACCOUNT_PTN_BALANCE_PREFIX, address.Bytes21()...)
	balance := uint64(0)
	data, err := statedb.db.Get(key)
	if err == nil {

		balance = BytesToUint64(data)
	}
	return balance
}

func (statedb *StateDb) LookupAccount() map[common.Address]*modules.AccountInfo {
	result := make(map[common.Address]*modules.AccountInfo)

	iter := statedb.db.NewIteratorWithPrefix(constants.ACCOUNT_PTN_BALANCE_PREFIX)
	for iter.Next() {
		key := iter.Key()
		balance := BytesToUint64(iter.Value())

		addB := bytes.TrimPrefix(key, constants.ACCOUNT_PTN_BALANCE_PREFIX)
		add := common.BytesToAddress(addB)

		acc := &modules.AccountInfo{Balance: balance}

		//get address vote mediator
		accKey := append(accountKey(add), []byte(constants.VOTED_MEDIATORS)...)
		data, _, err := retrieveWithVersion(statedb.db, accKey)
		if err == nil {
			votedMediators := make([]common.Address, 0)
			err = rlp.DecodeBytes(data, votedMediators)
			if err == nil {
				acc.VotedMediators = votedMediators
			}
		}

		log.Debugf("Found account[%v] balance:%v,vote mediator:%v", add.String(), balance, acc.VotedMediators)
		result[add] = acc
	}

	return result
}
func (statedb *StateDb) SaveAccountState(address common.Address, write *modules.ContractWriteSet, version *modules.StateVersion) error {
	key := accountKey(address)
	key = append(key, []byte(write.Key)...)
	if write.IsDelete {
		err := statedb.db.Delete(key)
		return err
	}
	return storeBytesWithVersion(statedb.db, key, version, write.Value)

}
func (statedb *StateDb) SaveAccountStates(address common.Address, writeset []modules.ContractWriteSet, version *modules.StateVersion) error {
	batch := statedb.db.NewBatch()
	keyPrefix := accountKey(address)
	for _, write := range writeset {
		key := []byte{}
		key = append(key, keyPrefix...)
		key = append(key, []byte(write.Key)...)
		if write.IsDelete {
			log.Infof("Account[%s] try to delete account state by key:%s", address.String(), write.Key)
			err := batch.Delete(key)
			return err
		}
		log.Debugf("Account[%s] try to set account state key:%s,value:%s", address.String(), write.Key, string(write.Value))
		err := storeBytesWithVersion(batch, key, version, write.Value)
		if err != nil {
			return err
		}
	}
	return batch.Write()
}

func (statedb *StateDb) GetAllAccountStates(address common.Address) (map[string]*modules.ContractStateValue, error) {
	key := accountKey(address)
	data := getprefix(statedb.db, key)
	var err error
	result := make(map[string]*modules.ContractStateValue, 0)
	for dbkey, state_version := range data {
		state, version, err0 := splitValueAndVersion(state_version)
		if err0 != nil {
			err = err0
		}
		realKey := dbkey[len(key):]
		if realKey != "" {
			result[realKey] = &modules.ContractStateValue{Value: state, Version: version}
			log.Info("the contract's state get info.", "key", realKey)
		}
	}
	return result, err
}
func (statedb *StateDb) GetAccountState(address common.Address, statekey string) (*modules.ContractStateValue, error) {
	key := accountKey(address)
	key = append(key, []byte(statekey)...)
	data, version, err := retrieveWithVersion(statedb.db, key)
	if err != nil {
		return nil, err
	}
	return &modules.ContractStateValue{Value: data, Version: version}, nil
}
