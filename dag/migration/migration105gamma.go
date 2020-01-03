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
	"fmt"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/dag/constants"
	"github.com/palletone/go-palletone/dag/modules"
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
	// 获取所有的 mediator信息
	mediatorInfoMap := make(map[[22]byte]*modules.MediatorInfo)
	mediatorInfoIter := m.statedb.NewIteratorWithPrefix(constants.MEDIATOR_INFO_PREFIX)

	for mediatorInfoIter.Next() {
		mediatorInfo := &modules.MediatorInfo{}
		err := rlp.DecodeBytes(mediatorInfoIter.Value(), mediatorInfo)
		if err != nil {
			log.Errorf(err.Error())
			return err
		}

		var key [22]byte
		copy(key[:], mediatorInfoIter.Key()[0:22])
		mediatorInfoMap[key] = mediatorInfo
	}

	// 遍历所有的 unit header
	headers := m.dagdb.NewIteratorWithPrefix(constants.HEADER_PREFIX)
	count := 0 // 计数器打印日志
	for headers.Next() {
		// 解析 header
		header := new(modules.Header)
		err := rlp.DecodeBytes(headers.Value(), header)
		if err != nil {
			log.Errorf(err.Error())
			return err
		}

		// 获取生产该 unit 的mediator
		var key [22]byte
		med := header.Author()
		copy(key[:], storage.GetMmediatorKey(med)[0:22])
		mi, found := mediatorInfoMap[key]
		if !found {
			if header.NumberU64() == 0 {
				continue
			}

			errStr := fmt.Sprintf("cannot find mediator info: %v", med.Str())
			log.Errorf(errStr)
			return fmt.Errorf(errStr)
		}

		// 将 mediator的 TotalProduct 加1
		mi.TotalProduct++

		count++
		if count%100000 == 0 {
			log.Infof("mediator info of %v units has been counted", count)
		}
	}

	// 存储所有mediator信息
	for key, mi := range mediatorInfoMap {
		err := storage.StoreToRlpBytes(m.statedb, key[:], mi)
		if err != nil {
			log.Errorf(err.Error())
			return err
		}
	}

	return nil
}
