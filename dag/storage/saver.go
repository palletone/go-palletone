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
	"fmt"
	"log"

	"github.com/palletone/go-palletone/common"
	palletdb "github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/common/rlp"
	"github.com/palletone/go-palletone/dag/constants"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/modules"
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

/**
key: [HEADER_PREFIX][chain index number]_[chain index]_[unit hash]
value: unit header rlp encoding bytes
*/
// save header
func SaveHeader(uHash common.Hash, h *modules.Header) error {
	key := fmt.Sprintf("%s%v_%s_%s", HEADER_PREFIX, h.Number.Index, h.Number.String(), uHash.Bytes())
	return Store(Dbconn, key, *h)
}

func SaveHashNumber(uHash common.Hash, height modules.ChainIndex) error {
	key := fmt.Sprintf("%s%s", UNIT_HASH_NUMBER_Prefix, uHash)
	return Store(Dbconn, key, height)
}

// height and assetid can get a unit key.
func SaveUHashIndex(cIndex modules.ChainIndex, uHash common.Hash) error {
	key := fmt.Sprintf("%s_%s_%d", UNIT_NUMBER_PREFIX, cIndex.AssetID.String(), cIndex.Index)
	return Store(Dbconn, key, uHash)
}

/**
key: [BODY_PREFIX][merkle root]
value: all transactions hash set's rlp encoding bytes
*/
func SaveBody(unitHash common.Hash, txsHash []common.Hash) error {
	// Dbconn.Put(append())
	key := fmt.Sprintf("%s%s", BODY_PREFIX, unitHash.String())
	return Store(Dbconn, key, txsHash)
}

func GetBody(unitHash common.Hash) ([]common.Hash, error) {
	key := fmt.Sprintf("%s%s", BODY_PREFIX, unitHash.String())
	data, err := Get([]byte(key))
	if err != nil {
		return nil, err
	}
	var txHashs []common.Hash
	if err := rlp.DecodeBytes(data, &txHashs); err != nil {
		return nil, err
	}
	return txHashs, nil
}

func SaveTransactions(txs *modules.Transactions) error {
	key := fmt.Sprintf("%s%s", TRANSACTIONS_PREFIX, txs.Hash())
	return Store(Dbconn, key, *txs)
}

/**
key: [TRANSACTION_PREFIX][tx hash]
value: transaction struct rlp encoding bytes
*/
func SaveTransaction(tx *modules.Transaction) error {
	// save transaction
	if err := StoreBytes(Dbconn, append(TRANSACTION_PREFIX, tx.TxHash.Bytes()...), tx); err != nil {
		return err
	}

	if err := StoreBytes(Dbconn, append(Transaction_Index, tx.TxHash.Bytes()...), tx); err != nil {
		return err
	}
	return nil
}
func SaveTxLookupEntry(unit *modules.Unit) error {
	for i, tx := range unit.Transactions() {
		in := modules.TxLookupEntry{
			UnitHash:  unit.Hash(),
			UnitIndex: unit.NumberU64(),
			Index:     uint64(i),
		}
		data, err := rlp.EncodeToBytes(in)
		if err != nil {
			return err
		}
		if err := StoreBytes(Dbconn, append(LookupPrefix, tx.TxHash.Bytes()...), data); err != nil {
			return err
		}
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
		if err := rlp.DecodeBytes(keys_byte[:], &keys); err != nil {
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

	return Store(Dbconn, "array_units", keys)
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

func SaveContract(contract *modules.Contract) (common.Hash, error) {
	if common.EmptyHash(contract.CodeHash) {
		contract.CodeHash = rlp.RlpHash(contract.Code)
	}
	// key = cs+ rlphash(contract)
	if common.EmptyHash(contract.Id) {
		ids := rlp.RlpHash(contract)
		if len(ids) > len(contract.Id) {
			id := ids[len(ids)-common.HashLength:]
			copy(contract.Id[common.HashLength-len(id):], id)
		} else {
			//*contract.Id = new(common.Hash)
			copy(contract.Id[common.HashLength-len(ids):], ids[:])
		}

	}
	if Dbconn == nil {
		Dbconn = ReNewDbConn(dagconfig.DefaultConfig.DbPath)
	}
	return contract.Id, StoreBytes(Dbconn, append(CONTRACT_PTEFIX, contract.Id[:]...), contract)
}

//  get  unit chain version
// GetUnitChainVersion reads the version number from db.
func GetUnitChainVersion() int {
	var vsn uint
	if Dbconn == nil {
		Dbconn = ReNewDbConn(dagconfig.DefaultConfig.DbPath)
	}
	enc, _ := Dbconn.Get([]byte("UnitchainVersion"))
	rlp.DecodeBytes(enc, &vsn)
	return int(vsn)
}

// SaveUnitChainVersion writes vsn as the version number to db.
func SaveUnitChainVersion(vsn int) error {
	enc, _ := rlp.EncodeToBytes(uint(vsn))
	if Dbconn == nil {
		Dbconn = ReNewDbConn(dagconfig.DefaultConfig.DbPath)
	}
	return Dbconn.Put([]byte("UnitchainVersion"), enc)
}
