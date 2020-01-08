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
 *  * @author PalletOne core developers <dev@pallet.one>
 *  * @date 2018-2019
 *
 */
package migration

import (
	"github.com/palletone/go-palletone/common/ptndb"
)

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

	/* version: 1.0.0-beta end */
	/* version: 1.0.0-beta */
	//m_101_beta := NewNothingMigration("1.0.1-beta", "1.0.2-beta")
	//migrations[m_101_beta.FromVersion()] = m_101_beta
	/* version: 1.0.0-beta end */

	m_100_beta := NewMigration100_101(db)
	migrations[m_100_beta.FromVersion()] = m_100_beta

	m_101_beta := NewMigration101_102(db)
	migrations[m_101_beta.FromVersion()] = m_101_beta

	m_102_gamma := NewMigration102beta_102gamma(db)
	migrations[m_102_gamma.FromVersion()] = m_102_gamma

	m_102_delta := NewMigration102gamma_102delta(db)
	migrations[m_102_delta.FromVersion()] = m_102_delta

	m_102_delta_release := NewNothingMigration("1.0.2-delta", "1.0.2-release")
	migrations[m_102_delta_release.FromVersion()] = m_102_delta_release

	m_103_alpha := NewMigration102delta_103alpha(db)
	migrations[m_103_alpha.FromVersion()] = m_103_alpha

	m_103_beta := NewMigration103alpha_103beta(db)
	migrations[m_103_beta.FromVersion()] = m_103_beta

	m_103_gamma := NewMigration103beta_103gamma(db)
	migrations[m_103_gamma.FromVersion()] = m_103_gamma

	m_103_gamma_release := NewNothingMigration("1.0.3-gamma", "1.0.3-release")
	migrations[m_103_gamma_release.FromVersion()] = m_103_gamma_release

	m_104_alpha := NewNothingMigration("1.0.3-release", "1.0.4-alpha")
	migrations[m_104_alpha.FromVersion()] = m_104_alpha

	m_104_beta := NewMigration104alpha_104beta(db)
	migrations[m_104_beta.FromVersion()] = m_104_beta

	m_104_release := NewNothingMigration("1.0.4-beta", "1.0.4-release")
	migrations[m_104_release.FromVersion()] = m_104_release

	m_104_rc1 := NewNothingMigration("1.0.4-release", "1.0.4-rc1")
	migrations[m_104_rc1.FromVersion()] = m_104_rc1

	m_105_alpha := NewMigration104beta_105alpha(db)
	migrations[m_105_alpha.FromVersion()] = m_105_alpha

	m_105_beta := NewMigration105alpha_105beta(db)
	migrations[m_105_beta.FromVersion()] = m_105_beta

	m_105_gamma := NewMigration105beta_105gamma(db)
	migrations[m_105_gamma.FromVersion()] = m_105_gamma

	m_105_delta := NewMigration105gamma_105delta(db)
	migrations[m_105_delta.FromVersion()] = m_105_delta

	return migrations
}

// func NewMigration0615_100(db ptndb.Database) *Migration0615_100 {
// 	return &Migration0615_100{dagdb: db, idxdb: db, utxodb: db, statedb: db, propdb: db}
// }

func NewMigration100_101(db ptndb.Database) *Migration100_101 {
	return &Migration100_101{dagdb: db, idxdb: db, utxodb: db, statedb: db, propdb: db}
}

func NewMigration101_102(db ptndb.Database) *Migration101_102 {
	return &Migration101_102{dagdb: db, idxdb: db, utxodb: db, statedb: db, propdb: db}
}

func NewMigration102beta_102gamma(db ptndb.Database) *Migration102beta_102gamma {
	return &Migration102beta_102gamma{dagdb: db, idxdb: db, utxodb: db, statedb: db, propdb: db}
}

func NewMigration102gamma_102delta(db ptndb.Database) *Migration102gamma_102delta {
	return &Migration102gamma_102delta{dagdb: db, idxdb: db, utxodb: db, statedb: db, propdb: db}
}

func NewMigration102delta_103alpha(db ptndb.Database) *Migration102delta_103alpha {
	return &Migration102delta_103alpha{dagdb: db, idxdb: db, utxodb: db, statedb: db, propdb: db}
}

func NewMigration103alpha_103beta(db ptndb.Database) *Migration103alpha_103beta {
	return &Migration103alpha_103beta{dagdb: db, idxdb: db, utxodb: db, statedb: db, propdb: db}
}

func NewMigration103beta_103gamma(db ptndb.Database) *Migration103beta_103gamma {
	return &Migration103beta_103gamma{dagdb: db, idxdb: db, utxodb: db, statedb: db, propdb: db}
}

func NewMigration104alpha_104beta(db ptndb.Database) *Migration104alpha_104beta {
	return &Migration104alpha_104beta{dagdb: db, idxdb: db, utxodb: db, statedb: db, propdb: db}
}

func NewMigration104beta_105alpha(db ptndb.Database) *Migration104beta_105alpha {
	return &Migration104beta_105alpha{dagdb: db, idxdb: db, utxodb: db, statedb: db, propdb: db}
}

func NewMigration105alpha_105beta(db ptndb.Database) *Migration105alpha_105beta {
	return &Migration105alpha_105beta{dagdb: db, idxdb: db, utxodb: db, statedb: db, propdb: db}
}

func NewMigration105beta_105gamma(db ptndb.Database) *Migration105beta_105gamma {
	return &Migration105beta_105gamma{dagdb: db, idxdb: db, utxodb: db, statedb: db, propdb: db}
}

func NewMigration105gamma_105delta(db ptndb.Database) *Migration105gamma_105delta {
	return &Migration105gamma_105delta{dagdb: db, idxdb: db, utxodb: db, statedb: db, propdb: db}
}
