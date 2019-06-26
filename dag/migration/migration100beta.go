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
 *  * @date 2018-2019
 *
 */
package migration

import (
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/dag/modules"
)

type Migration100_100 struct {
	//mdag  dag.IDag
	dagdb   ptndb.Database
	idxdb   ptndb.Database
	utxodb  ptndb.Database
	statedb ptndb.Database
	propdb  ptndb.Database
}

func (m *Migration100_100) FromVersion() string {
	return "1.0.0-beta"
}
func (m *Migration100_100) ToVersion() string {
	return "1.0.1-beta"
}
func (m *Migration100_100) ExecuteUpgrade() error {
	// gp 只做了添加和删除的修改
	log.Info("exec migration ", "version", m.FromVersion())
	newGp := modules.NewGlobalProp()
	newData, err := rlp.EncodeToBytes(newGp)
	if err != nil {
		log.Error("ExecuteUpgrade error:" + err.Error())
		return err
	}
	m.statedb.Put([]byte("gpGlobalProperty"), newData)
	return nil
}
