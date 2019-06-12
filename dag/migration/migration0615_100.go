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
	"github.com/palletone/go-palletone/common/ptndb"
)

type Migration0615_100 struct {
	dagdb   ptndb.Database
	idxdb   ptndb.Database
	uxtodb  ptndb.Database
	statedb ptndb.Database
	propdb  ptndb.Database
}

func (m *Migration0615_100) FromVersion() string {
	return "0.6.15"
}
func (m *Migration0615_100) ToVersion() string {
	return "1.0.0"
}
func (m *Migration0615_100) ExecuteUpgrade() error {
	err := RenameKey(m.propdb, []byte("gpGlobalProperty"), []byte("GlobalProperty"))
	if err != nil {
		return err
	}
	err = RenamePrefix(m.dagdb, []byte("uht"), []byte("hh"))
	if err != nil {
		return err
	}
	return nil
}
