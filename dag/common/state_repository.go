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

package common

import (
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/dag/storage"
)

type IStateRepository interface {
	GetCandidateMediators() []*core.MediatorInfo
}

type StateRepository struct {
	statedb storage.IStateDb
	logger  log.ILogger
}

func NewStateRepository(statedb storage.IStateDb, l log.ILogger) *StateRepository {
	return &StateRepository{statedb: statedb, logger: l}
}
func (rep *StateRepository) GetCandidateMediators() []*core.MediatorInfo {
	addrs, err := rep.statedb.GetCandidateMediatorAddrList()
	result := []*core.MediatorInfo{}
	if err != nil {
		return result
	}

	for _, addr := range addrs {
		minfo, err := rep.statedb.GetAccountMediatorInfo(addr)
		if err != nil {
			rep.logger.Errorf("GetMediator info from address:{%s} has an error:%s", addr.String(), err)
			continue
		}
		result = append(result, minfo)
	}
	return result
}
