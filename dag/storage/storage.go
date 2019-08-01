/*
   This file is part of go-palletone.
   go-palletone is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.
   go-palletone is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.
   You should have received a copy of the GNU General Public License
   along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
*/

/*
 * @author PalletOne core developers <dev@pallet.one>
 * @date 2018
 */

package storage

import (
	"encoding/json"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/common/util"
)

func StoreToJsonBytes(db ptndb.Putter, key []byte, value interface{}) error {
	val, err := json.Marshal(value)
	if err != nil {
		log.Debugf("json marshal err: %v", err.Error())
		return err
	}

	err = db.Put(key, val)
	if err != nil {
		log.Debugf("DB put err: %v", err.Error())
		return err
	}

	return nil
}

func StoreToRlpBytes(db ptndb.Putter, key []byte, value interface{}) error {
	val, err := rlp.EncodeToBytes(value)
	if err != nil {
		return err
	}
	err = db.Put(key, val)
	if err != nil {
		log.Error("StoreToRlpBytes", "key:", string(key), "err:", err)
	}
	return err
}

func GetBytes(db ptndb.Database, key []byte) ([]byte, error) {
	val, err := db.Get(key)
	if err != nil {
		return []byte{}, err
	}
	log.Debug("storage GetBytes", "key:", string(key), "value:", string(val))
	return val, nil
}

func StoreString(db ptndb.Putter, key, value string) error {
	return db.Put(util.ToByte(key), util.ToByte(value))
}
func GetString(db ptndb.Database, key string) (string, error) {
	return getString(db, util.ToByte(key))
}

func BatchErrorHandler(err error, errorList *[]error) {
	if err != nil {
		*errorList = append(*errorList, err)
	}
}
