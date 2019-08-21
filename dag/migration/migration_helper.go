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
	"fmt"
	"github.com/palletone/go-palletone/common/ptndb"
)

func RenameKey(db ptndb.Database, oldKey, newKey []byte) error {
	v, err := db.Get(oldKey)
	if err != nil {
		return err
	}
	err = db.Put(newKey, v)
	if err != nil {
		return err
	}
	fmt.Printf("Rename key from %s to %s", string(oldKey), string(newKey))
	return db.Delete(oldKey)
}
func RenamePrefix(db ptndb.Database, oldPrefix, newPrefix []byte) error {
	batch := db.NewBatch()
	it := db.NewIteratorWithPrefix(oldPrefix)
	count := 0
	if it.Next() {
		key := it.Key()
		value := it.Value()
		newKey := append(newPrefix, key[len(oldPrefix):]...)
		batch.Put(newKey, value)
		batch.Delete(key)
		count++
	}
	fmt.Printf("Rename prefix from %s to %s, count:%d", string(oldPrefix), string(newPrefix), count)
	return batch.Write()
}
