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
 *  * @author PalletOne core developers <dev@pallet.one>
 *  * @date 2018
 *
 */

package storage

import (
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/dag/constants"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/vote"
)

func (statedb *StateDb) GetAccountInfo(address common.Address) (*modules.AccountInfo, error) {
	key := append(constants.ACCOUNT_INFO_PREFIX, address.Bytes21()...)
	info := &modules.AccountInfo{}
	err := retrieve(statedb.db, key, info)
	if err != nil {
		statedb.logger.Errorf("Get account[%s] info throw an error:%s", address.String(), err.Error())
		return nil, err
	}
	return info, nil
}

func (statedb *StateDb) SaveAccountInfo(address common.Address, info *modules.AccountInfo) error {
	key := append(constants.ACCOUNT_INFO_PREFIX, address.Bytes21()...)
	statedb.logger.Debugf("Save account info for address:%s", address.String())
	return StoreBytes(statedb.db, key, info)
}
func (statedb *StateDb) UpdateAccountInfoBalance(address common.Address, addAmount int64) error {
	key := append(constants.ACCOUNT_INFO_PREFIX, address.Bytes21()...)
	exist, _ := statedb.db.Has(key)
	info := &modules.AccountInfo{}
	if exist {
		err := retrieve(statedb.db, key, info)
		if err != nil {
			statedb.logger.Errorf("Get account[%s] info throw an error:%s", address.String(), err.Error())
			return err
		}
	} else {
		info.PtnBalance = 0
	}
	info.PtnBalance = uint64(int64(info.PtnBalance) + addAmount)
	statedb.logger.Debugf("Update Ptn Balance for address:%s, add Amount:%d", address.String(), addAmount)
	return StoreBytes(statedb.db, key, info)
}
func (statedb *StateDb) GetAccountVoteInfo(address common.Address, voteType uint8) [][]byte {
	accountInfo, err := statedb.GetAccountInfo(address)
	if err != nil {
		return nil
	}
	res := make([][]byte, 0)
	for _, vote := range accountInfo.Votes {
		if vote.VoteType == voteType {
			res = append(res, vote.VoteContent)
		}
	}
	return res

}

func (statedb *StateDb) AddVote2Account(address common.Address, voteInfo vote.VoteInfo) error {
	accountInfo, err := statedb.GetAccountInfo(address)
	if err != nil {
		return err
	}
	accountInfo.Votes = append(accountInfo.Votes, voteInfo)
	if err = statedb.SaveAccountInfo(address, accountInfo); err != nil {
		return err
	}
	return nil
}
