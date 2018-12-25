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
 *  * @author PalletOne core developer  <dev@pallet.one>
 *  * @date 2018
 *
 */

package storage

import (
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/dag/modules"
)

type IndexDb struct {
	db     ptndb.Database
	logger log.ILogger
}

func NewIndexDb(db ptndb.Database, l log.ILogger) *IndexDb {
	return &IndexDb{db: db, logger: l}
}

type IIndexDb interface {
	GetPrefix(prefix []byte) map[string][]byte
	SaveIndexValue(key []byte, value interface{}) error
	GetUtxoByIndex(idx *modules.UtxoIndex) (*modules.Utxo, error)
	DeleteUtxoByIndex(idx *modules.UtxoIndex) error

}

// ###################### SAVE IMPL START ######################
func (idxdb *IndexDb) SaveIndexValue(key []byte, value interface{}) error {
	return StoreBytes(idxdb.db, key, value)
}


// ###################### SAVE IMPL END ######################
// ###################### GET IMPL START ######################
func (idxdb *IndexDb) GetPrefix(prefix []byte) map[string][]byte {
	return getprefix(idxdb.db, prefix)
}


// ###################### GET IMPL END ######################
func (idxdb *IndexDb) GetUtxoByIndex(idx *modules.UtxoIndex) (*modules.Utxo, error) {
	key := idx.ToKey()
	utxo := new(modules.Utxo)
	err := retrieve(idxdb.db, key, utxo)
	return utxo, err
}
func (idxdb *IndexDb) DeleteUtxoByIndex(idx *modules.UtxoIndex) error {
	return idxdb.db.Delete(idx.ToKey())
}


