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
	key := append(constants.ACCOUNT_INFO_PREFIX, address.Bytes()...)
	info := &modules.AccountInfo{}
	err := retrieve(statedb.db, key, info)
	if err != nil {
		return nil, err
	}
	return info, nil
}

func (statedb *StateDb) SaveAccountInfo(address common.Address, info *modules.AccountInfo) error {
	key := append(constants.ACCOUNT_INFO_PREFIX, address.Bytes()...)
	return StoreBytes(statedb.db, key, info)
}

func (statedb *StateDb) GetAccountVoteInfo(address common.Address, voteType uint8) ([][]byte) {
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

// todo albert·gou
//func (statedb *StateDb) GetAccountMediatorInfo(address common.Address) (*core.MediatorInfo, error) {
//	key := append(modules.ACCOUNT_INFO_PREFIX, address.Bytes()...)
//	key = append(key, []byte("MediatorInfo")...)
//	info := &core.MediatorInfo{}
//	err := retrieve(statedb.db, key, info)
//	if err != nil {
//		return nil, err
//	}
//	return info, nil
//}

// todo albert·gou
//func (statedb *StateDb) SaveAccountMediatorInfo(address common.Address, info *core.MediatorInfo, version *modules.StateVersion) error {
//	key := append(modules.ACCOUNT_INFO_PREFIX, address.Bytes()...)
//	key = append(key, []byte("MediatorInfo")...)
//	statedb.logger.Debugf("Save one mediator info for address{%s},info:{%s}", address.String(), info)
//	return StoreBytesWithVersion(statedb.db, key, version, info)
//}
