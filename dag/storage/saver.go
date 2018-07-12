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
	"github.com/palletone/go-palletone/common/rlp"
	"log"

	palletdb "github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/dag/constants"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/common"
)

var (
	Dbconn             *palletdb.LDBDatabase = nil
	AssocUnstableUnits map[string]modules.Joint
	//DBPath             string = "/Users/jay/code/gocode/src/palletone/bin/leveldb"
	DBPath string = dagconfig.DefaultConfig.DbPath
)

func SaveJoint(objJoint *modules.Joint, onDone func()) (err error) {
	if objJoint.Unsigned != "" {
		return errors.New(objJoint.Unsigned)
	}
	if Dbconn == nil {
		Dbconn = ReNewDbConn(dagconfig.DefaultConfig.DbPath)
	}
	obj_unit := objJoint.Unit
	obj_unit_byte, _ := json.Marshal(obj_unit)

	if err = Dbconn.Put(append(UNIT_PREFIX, obj_unit.Hash().Bytes()...), obj_unit_byte); err != nil {
		return
	}
	// add key in  unit_keys
	log.Println("add unit key:", string(UNIT_PREFIX)+obj_unit.Hash().String(), AddUnitKeys(string(UNIT_PREFIX)+obj_unit.Hash().String()))

	if dagconfig.SConfig.Blight {
		// save  update utxo , message , transaction

	}

	if onDone != nil {
		onDone()
	}
	return
}

// save unit
func SaveUnit(unit *modules.Unit) error {
	log.Println("Start save unit... ")
	obj_unit := unit
	obj_unit_byte, err := json.Marshal(obj_unit)
	if err != nil {
		return err
	}
	if Dbconn == nil {
		Dbconn = ReNewDbConn(dagconfig.DefaultConfig.DbPath)
	}
	if err = Dbconn.Put(append(UNIT_PREFIX, obj_unit.Hash().Bytes()...), obj_unit_byte); err != nil {
		return err
	}
	AddUnitKeys(string(UNIT_PREFIX) + obj_unit.Hash().String())
	return nil
}

// save header
func SaveHeader(uHahs common.Hash, h *modules.Header) error {
	data, err := rlp.EncodeToBytes(h)
	if err != nil {
		return err
	}

	//chain_index := h.ChainIndex()
	//encNum := encodeBlockNumber(chain_index.Index)
	chain_index, err := rlp.EncodeToBytes(h.Number)
	if err != nil {
		return err
	}

	if Dbconn == nil {
		Dbconn = ReNewDbConn(dagconfig.DefaultConfig.DbPath)
	}

	// Dbconn.Put(append())
	key := append(append(HEADERPREFIX, chain_index...), uHahs.Bytes()...)
	log.Println(key)

	if err := Dbconn.Put(key, data); err != nil {
		return err
	}
	return nil
}

func SaveTransactions(txs *modules.Transactions) error {
	log.Println("Start save header... ")

	if Dbconn == nil {
		Dbconn = ReNewDbConn(dagconfig.DefaultConfig.DbPath)
	}
	data, err := rlp.EncodeToBytes(txs)
	if err != nil {
		return err
	}
	// Dbconn.Put(append())

	if err := Dbconn.Put(append(TRANSACTIONSPREFIX, txs.Hash().Bytes()...), data); err != nil {
		return err
	}
	return nil
}

// encodeBlockNumber encodes a block number as big endian uint64
func encodeBlockNumber(number uint64) []byte {
	enc := make([]byte, 8)
	binary.BigEndian.PutUint64(enc, number)
	return enc
}

func GetUnitKeys() []string {
	var keys []string
	if Dbconn == nil {
		Dbconn = ReNewDbConn(dagconfig.DefaultConfig.DbPath)
	}
	if keys_byte, err := Dbconn.Get([]byte("array_units")); err != nil {
		log.Println("get units error:", err)
	} else {
		if err := json.Unmarshal(keys_byte, &keys); err != nil {
			log.Println("error:", err)
		}
	}
	return keys
}
func AddUnitKeys(key string) error {
	keys := GetUnitKeys()
	if len(keys) <= 0 {
		return errors.New("null keys.")
	}
	for _, v := range keys {

		if v == key {
			return errors.New("key is already exist.")
		}
	}
	keys = append(keys, key)
	if Dbconn == nil {
		Dbconn = ReNewDbConn(dagconfig.DefaultConfig.DbPath)
	}

	if err := Dbconn.Put([]byte("array_units"), ConvertBytes(keys)); err != nil {
		return err
	}
	return nil

}
func ConvertBytes(val interface{}) (re []byte) {
	var err error
	if re, err = json.Marshal(val); err != nil {
		log.Println("json.marshal error:", err)
	}
	return re[:]
}
func IsGenesisUnit(unit string) bool {
	return unit == constants.GENESIS_UNIT
}

func GetKeysWithTag(tag string) []string {
	var keys []string
	if Dbconn == nil {
		Dbconn = ReNewDbConn(dagconfig.DefaultConfig.DbPath)
	}
	if keys_byte, err := Dbconn.Get([]byte(tag)); err != nil {
		log.Println("get keys error:", err)
	} else {
		if err := json.Unmarshal(keys_byte, &keys); err != nil {
			log.Println("error:", err)
		}
	}
	return keys
}
func AddKeysWithTag(key, tag string) error {
	keys := GetKeysWithTag(tag)
	if len(keys) <= 0 {
		return errors.New("null keys.")
	}
	log.Println("keys:=", keys)
	for _, v := range keys {
		if v == key {
			return errors.New("key is already exist.")
		}
	}
	keys = append(keys, key)
	if Dbconn == nil {
		Dbconn = ReNewDbConn(dagconfig.DefaultConfig.DbPath)
	}

	if err := Dbconn.Put([]byte(tag), ConvertBytes(keys)); err != nil {
		return err
	}
	return nil

}
