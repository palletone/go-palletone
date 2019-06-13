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

import ()

//如果从一个版本升级到另一个版本，数据库不需要做任何更改时，使用该实例
type NothingMigration struct {
	from, to string
}

func NewNothingMigration(from, to string) *NothingMigration {
	return &NothingMigration{from: from, to: to}
}
func (m *NothingMigration) FromVersion() string {
	return m.from
}
func (m *NothingMigration) ToVersion() string {
	return m.to
}
func (m *NothingMigration) ExecuteUpgrade() error {

	return nil
}
