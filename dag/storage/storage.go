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
	"fmt"
	"math/big"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/common/rlp"
	"github.com/palletone/go-palletone/common/util"
	"github.com/syndtr/goleveldb/leveldb/errors"
)

func (dagdb *DagDatabase) PutCanonicalHash(hash common.Hash, number uint64) error {
	key := append(HeaderCanon_Prefix, encodeBlockNumber(number)...)
	if err := dagdb.db.Put(append(key, NumberSuffix...), hash.Bytes()); err != nil {
		return err
	}
	return nil
}
func (dagdb *DagDatabase) PutHeadHeaderHash(hash common.Hash) error {
	if err := dagdb.db.Put(HeadHeaderKey, hash.Bytes()); err != nil {
		return err
	}
	return nil
}

// PutHeadUnitHash stores the head unit's hash.
func (dagdb *DagDatabase) PutHeadUnitHash(hash common.Hash) error {
	if err := dagdb.db.Put(HeadUnitKey, hash.Bytes()); err != nil {
		return err
	}
	return nil
}

// PutHeadFastUnitHash stores the fast head unit's hash.
func (dagdb *DagDatabase) PutHeadFastUnitHash(hash common.Hash) error {
	if err := dagdb.db.Put(HeadFastKey, hash.Bytes()); err != nil {
		return err
	}
	return nil
}

// PutTrieSyncProgress stores the fast sync trie process counter to support
// retrieving it across restarts.
func (dagdb *DagDatabase) PutTrieSyncProgress(count uint64) error {
	if err := dagdb.db.Put(TrieSyncKey, new(big.Int).SetUint64(count).Bytes()); err != nil {
		return err
	}
	return nil
}

// value will serialize to rlp encoding bytes
func Store(db ptndb.Database, key string, value interface{}) error {
	return StoreBytes(db, []byte(key), value)
}

func StoreBytes(db ptndb.Database, key []byte, value interface{}) error {
	val, err := rlp.EncodeToBytes(value)
	if err != nil {
		return err
	}

	_, err = db.Get(key)
	if err != nil {
		if err.Error() == errors.ErrNotFound.Error() {
			//	if err == errors.New("not found") {
			if err := db.Put(key, val); err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		if err = db.Delete(key); err != nil {
			return err
		}
		if err := db.Put(key, val); err != nil {
			return err
		}
	}
	return nil
}

func StoreString(db ptndb.Putter, key, value string) error {
	return db.Put(util.ToByte(key), util.ToByte(value))
}

//batch put HeaderCanon & HeaderKey & HeadUnitKey & HeadFastKey
func (dagdb *DagDatabase) UpdateHeadByBatch(hash common.Hash, number uint64) error {
	batch := dagdb.db.NewBatch()
	errorList := &[]error{}

	key := append(HeaderCanon_Prefix, encodeBlockNumber(number)...)
	BatchErrorHandler(batch.Put(append(key, NumberSuffix...), hash.Bytes()), errorList) //PutCanonicalHash
	BatchErrorHandler(batch.Put(HeadHeaderKey, hash.Bytes()), errorList)                //PutHeadHeaderHash
	BatchErrorHandler(batch.Put(HeadUnitKey, hash.Bytes()), errorList)                  //PutHeadUnitHash
	BatchErrorHandler(batch.Put(HeadFastKey, hash.Bytes()), errorList)                  //PutHeadFastUnitHash
	if len(*errorList) == 0 {                                                           //each function call succeed.
		return batch.Write()
	}
	return fmt.Errorf("UpdateHeadByBatch, at least one sub function call failed.")
}

func BatchErrorHandler(err error, errorList *[]error) {
	if err != nil {
		*errorList = append(*errorList, err)
	}
}
