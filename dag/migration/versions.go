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

import "github.com/palletone/go-palletone/common/ptndb"

func NewMigrations(db ptndb.Database) map[string]IMigration {
	// 将所有待升级的migration版本，在这里实例化。
	migrations := make(map[string]IMigration)
	/* version: 0615 */
	// m_0615 := NewMigration0615_100(db)
	// if ver := m_0615.FromVersion(); ver != "" {
	// 	migrations[ver] = m_0615
	// }
	// /* version: 0615 end*/

	/* version: 1.0.0-beta */
	m_100_beta := NewMigration100_101(db)
	migrations[m_100_beta.FromVersion()] = m_100_beta
	/* version: 1.0.0-beta end */
	/* version: 1.0.0-beta */
	//m_101_beta := NewNothingMigration("1.0.1-beta", "1.0.2-beta")
	//migrations[m_101_beta.FromVersion()] = m_101_beta
	/* version: 1.0.0-beta end */
	return migrations
}

// func NewMigration0615_100(db ptndb.Database) *Migration0615_100 {
// 	return &Migration0615_100{dagdb: db, idxdb: db, utxodb: db, statedb: db, propdb: db}
// }

func NewMigration100_101(db ptndb.Database) *Migration100_101 {
	return &Migration100_101{dagdb: db, idxdb: db, utxodb: db, statedb: db, propdb: db}
}
