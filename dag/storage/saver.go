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
	"encoding/binary"
	"encoding/json"
	"errors"
	"log"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common/ptndb"
)

// encodeBlockNumber encodes a block number as big endian uint64
func Uint64ToBytes(number uint64) []byte {
	enc := make([]byte, 8)
	binary.BigEndian.PutUint64(enc, number)
	return enc
}

func BytesToUint64(b []byte) uint64 {

	return binary.BigEndian.Uint64(b)
}

func ConvertBytes(val interface{}) (re []byte) {
	var err error
	if re, err = json.Marshal(val); err != nil {
		log.Println("json.marshal error:", err)
	}
	return re[:]
}

func GetKeysWithTag(db ptndb.Database, tag string) []string {
	var keys []string

	if keys_byte, err := db.Get([]byte(tag)); err != nil {
		log.Println("get keys error:", err)
	} else {
		if err := json.Unmarshal(keys_byte, &keys); err != nil {
			log.Println("error:", err)
		}
	}
	return keys
}
func AddKeysWithTag(db ptndb.Database, key, tag string) error {
	keys := GetKeysWithTag(db, tag)
	if len(keys) <= 0 {
		return errors.New("null keys.")
	}
	for _, v := range keys {
		if v == key {
			return errors.New("key is already exist.")
		}
	}
	keys = append(keys, key)

	if err := db.Put([]byte(tag), ConvertBytes(keys)); err != nil {
		return err
	}

	return nil
}

//  get  unit chain version
// GetUnitChainVersion reads the version number from db.
func GetUnitChainVersion(db ptndb.Database) int {
	var vsn uint

	enc, _ := db.Get([]byte("UnitchainVersion"))
	rlp.DecodeBytes(enc, &vsn)
	return int(vsn)
}

// SaveUnitChainVersion writes vsn as the version number to db.
func SaveUnitChainVersion(db ptndb.Putter, vsn int) error {
	enc, _ := rlp.EncodeToBytes(uint(vsn))
	return db.Put([]byte("UnitchainVersion"), enc)
}
