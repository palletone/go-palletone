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
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/dag/constants"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/dag/storage"
)

type Migration105beta_105gamma struct {
	dagdb   ptndb.Database
	idxdb   ptndb.Database
	utxodb  ptndb.Database
	statedb ptndb.Database
	propdb  ptndb.Database
}

func (m *Migration105beta_105gamma) FromVersion() string {
	return "1.0.5-beta"
}

func (m *Migration105beta_105gamma) ToVersion() string {
	return "1.0.5-gamma"
}

func (m *Migration105beta_105gamma) ExecuteUpgrade() error {
	// 统计每个mediator的 TotalProduct
	if err := m.countMediatorTotalProduct(); err != nil {
		return err
	}

	return nil
}

func (m *Migration105beta_105gamma) countMediatorTotalProduct() error {
	// 遍历所有的 unit header
	headers := m.dagdb.NewIteratorWithPrefix(constants.HEADER_PREFIX)
	for headers.Next()  {
		header := new(modules.Header)
		err := rlp.DecodeBytes(headers.Value(), header)
		if err != nil {
			log.Errorf(err.Error())
			return err
		}

		// 获取生产该 unit 的mediator
		key := append(constants.MEDIATOR_INFO_PREFIX, header.Author().Bytes()...)
		mi := modules.NewMediatorInfo()
		err = storage.RetrieveFromRlpBytes(m.statedb, key, mi)
		if err != nil {
			log.Errorf(err.Error())
			return err
		}

		// 将 mediator的 TotalProduct 加1
		mi.TotalProduct++
		err = storage.StoreToRlpBytes(m.statedb, key, mi)
		if err != nil {
			log.Errorf(err.Error())
			return err
		}

		//stateDb := storage.NewStateDb(m.statedb)
		//med := header.Author()
		//mi, err := stateDb.RetrieveMediatorInfo(med)
		//if err != nil {
		//	log.Errorf(err.Error())
		//	return err
		//}
		//
		//mi.TotalProduct++
		//err = stateDb.StoreMediatorInfo(med, mi)
		//if err != nil {
		//	return err
		//}
	}

	return nil
}
