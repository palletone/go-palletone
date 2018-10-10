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
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/dag/modules"
)

type IndexDb struct{
	db ptndb.Database
}
func NewIndexDb(db ptndb.Database) *IndexDb {
	return &IndexDb{db:db,}
}

type IIndexDb interface {
	GetPrefix(prefix []byte) map[string][]byte
	SaveIndexValue(key []byte,value interface{}) error
	GetUtxoByIndex(idx *modules.UtxoIndex) (*modules.Utxo, error)
	DeleteUtxoByIndex(idx *modules.UtxoIndex) error
}
// ###################### SAVE IMPL START ######################
func (idxdb *IndexDb) SaveIndexValue(key []byte, value interface{}) error {
	return StoreBytes(idxdb.db, key, value)
}

// ###################### SAVE IMPL END ######################
// ###################### GET IMPL START ######################
func (db *IndexDb) GetPrefix(prefix []byte) map[string][]byte {
	return getprefix(db.db, prefix)
}
// ###################### GET IMPL END ######################
func(db *IndexDb) GetUtxoByIndex(idx *modules.UtxoIndex) (*modules.Utxo, error){
	key:=idx.ToKey()
	utxo:=new(modules.Utxo)
	err:= retrieve(db.db,key,utxo)
	return utxo,err
}
func(db *IndexDb)DeleteUtxoByIndex(idx *modules.UtxoIndex) error{
	return db.db.Delete(idx.ToKey())
}