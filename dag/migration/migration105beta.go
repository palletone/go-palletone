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
 *  * @author PalletOne core developer albert <dev@pallet.one>
 *  * @date 2019-2020
 *
 */
package migration

import (
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/dag/constants"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/storage"
)

type Migration105alpha_105beta struct {
	dagdb   ptndb.Database
	idxdb   ptndb.Database
	utxodb  ptndb.Database
	statedb ptndb.Database
	propdb  ptndb.Database
}

func (m *Migration105alpha_105beta) FromVersion() string {
	return "1.0.5-alpha"
}

func (m *Migration105alpha_105beta) ToVersion() string {
	return "1.0.5-beta"
}

func (m *Migration105alpha_105beta) ExecuteUpgrade() error {
	// 转换mediator结构体
	if err := m.upgradeMediatorInfo(); err != nil {
		return err
	}

	return nil
}

func (m *Migration105alpha_105beta) upgradeMediatorInfo() error {
	oldMediatorsIterator := m.statedb.NewIteratorWithPrefix(constants.MEDIATOR_INFO_PREFIX)
	for oldMediatorsIterator.Next() {
		oldMediator := &MediatorInfo105alpha{}
		err := rlp.DecodeBytes(oldMediatorsIterator.Value(), oldMediator)
		if err != nil {
			log.Debugf(err.Error())
			return err
		}

		mie := core.NewMediatorInfoExpand()
		mie.MediatorInfoExpand105alpha = *oldMediator.MediatorInfoExpand105alpha

		newMediator := &modules.MediatorInfo{
			MediatorInfoBase:   oldMediator.MediatorInfoBase,
			MediatorApplyInfo:  oldMediator.MediatorApplyInfo,
			MediatorInfoExpand: mie,
		}

		err = storage.StoreToRlpBytes(m.statedb, oldMediatorsIterator.Key(), newMediator)
		if err != nil {
			log.Debugf(err.Error())
			return err
		}
	}

	return nil
}

type MediatorInfo105alpha struct {
	*core.MediatorInfoBase
	*core.MediatorApplyInfo
	*core.MediatorInfoExpand105alpha
}
