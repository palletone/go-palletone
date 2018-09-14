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
 *  * @date 2018
 *
 */

package storage

import (
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/common/rlp"
	"github.com/palletone/go-palletone/dag/modules"
)

type UtxoDatabase struct{
	db ptndb.Database

}
func NewUtxoDatabase(db ptndb.Database) *UtxoDatabase {
	return &UtxoDatabase{db:db,}
}
type UtxoDb interface {
	GetPrefix(prefix []byte) map[string][]byte
	GetUtxoByIndex(indexKey []byte) ([]byte,error)
	GetUtxoEntry( key []byte) (*modules.Utxo, error)
	SaveUtxoEntity(key []byte,utxo modules.Utxo) error
	DeleteUtxo(key []byte) error
}

// ###################### SAVE IMPL START ######################

func (utxodb *UtxoDatabase) SaveUtxoEntity(key []byte, utxo modules.Utxo) error {
	return StoreBytes(utxodb.db, key, utxo)
}

func (utxodb *UtxoDatabase) DeleteUtxo(key []byte) error {
	return utxodb.db.Delete(key)
}


// ###################### SAVE IMPL END ######################

// ###################### GET IMPL START ######################
//  dbFetchUtxoEntry
func (utxodb *UtxoDatabase) GetUtxoEntry(key []byte) (*modules.Utxo, error) {
	utxo := new(modules.Utxo)
	data, err := utxodb.db.Get(key)
	if err != nil {
		return nil, err
	}

	if err := rlp.DecodeBytes(data, &utxo); err != nil {
		return nil, err
	}
	return utxo, nil
}
func (utxodb *UtxoDatabase) GetUtxoByIndex(indexKey []byte) ([]byte, error) {
	return utxodb.db.Get(indexKey)
}

// get prefix: return maps
func (db *UtxoDatabase) GetPrefix(prefix []byte) map[string][]byte {
	return getprefix(db.db, prefix)
}
// ###################### GET IMPL END ######################