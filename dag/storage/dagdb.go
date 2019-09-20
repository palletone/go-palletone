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
	"math/big"
	"reflect"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/dag/constants"
	"github.com/palletone/go-palletone/dag/modules"
)

//对DAG对象的操作，包括：Unit，Tx等
type DagDb struct {
	db ptndb.Database
}

func NewDagDb(db ptndb.Database) *DagDb {
	return &DagDb{db: db}
}

type IDagDb interface {
	GetGenesisUnitHash() (common.Hash, error)
	SaveGenesisUnitHash(hash common.Hash) error

	SaveHeader(h *modules.Header) error
	SaveHeaders(headers []*modules.Header) error
	SaveTransaction(tx *modules.Transaction) error
	GetAllTxs() ([]*modules.Transaction, error)
	SaveBody(unitHash common.Hash, txsHash []common.Hash) error
	GetBody(unitHash common.Hash) ([]common.Hash, error)
	SaveTxLookupEntry(unit *modules.Unit) error

	PutTrieSyncProgress(count uint64) error

	GetUnitTransactions(hash common.Hash) (modules.Transactions, error)
	GetTransactionOnly(hash common.Hash) (*modules.Transaction, error)
	GetTxLookupEntry(hash common.Hash) (*modules.TxLookupEntry, error)
	GetPrefix(prefix []byte) map[string][]byte
	GetHeaderByHash(hash common.Hash) (*modules.Header, error)
	IsHeaderExist(uHash common.Hash) (bool, error)
	IsTransactionExist(txHash common.Hash) (bool, error)
	GetHashByNumber(number *modules.ChainIndex) (common.Hash, error)

	GetTrieSyncProgress() (uint64, error)

	// common geter
	GetCommon(key []byte) ([]byte, error)
	GetCommonByPrefix(prefix []byte) map[string][]byte
	SaveCommon(key, val []byte) error
	// get txhash  and save index
	//GetReqIdByTxHash(hash common.Hash) (common.Hash, error)
	GetTxHashByReqId(reqid common.Hash) (common.Hash, error)
	ForEachAllTxDo(txAction func(key []byte, transaction *modules.Transaction) error) error
}

func (dagdb *DagDb) IsHeaderExist(uHash common.Hash) (bool, error) {
	key := append(constants.HEADER_PREFIX, uHash.Bytes()...)
	return dagdb.db.Has(key)
}

/* ----- common saver ----- */
func (dagdb *DagDb) SaveCommon(key, val []byte) error {
	return dagdb.db.Put(key, val)
}

/* ----- common geter ----- */
func (dagdb *DagDb) GetCommon(key []byte) ([]byte, error) {
	return dagdb.db.Get(key)
}
func (dagdb *DagDb) GetCommonByPrefix(prefix []byte) map[string][]byte {
	iter := dagdb.db.NewIteratorWithPrefix(prefix)
	result := make(map[string][]byte)
	for iter.Next() {
		key := iter.Key()
		value := make([]byte, 0)
		// 请注意： 直接赋值取得iter.Value()的最后一个指针
		// result[*(*string)(unsafe.Pointer(&key))] = append(value, iter.Value()...)
		result[string(key)] = append(value, iter.Value()...)
	}
	return result
}
func (dagdb *DagDb) GetGenesisUnitHash() (common.Hash, error) {
	hash := common.Hash{}
	hashb, err := dagdb.db.Get(constants.GenesisUnitHash)
	if err != nil {
		return hash, err
	}
	hash.SetBytes(hashb)
	return hash, nil
}
func (dagdb *DagDb) SaveGenesisUnitHash(hash common.Hash) error {
	log.Debugf("Save GenesisUnitHash:%#x", hash.Bytes())
	return dagdb.db.Put(constants.GenesisUnitHash, hash.Bytes())

}

// ###################### SAVE IMPL START ######################
/**
key: [HEADER_PREFIX][chain index number]_[chain index]_[unit hash]
value: unit header rlp encoding bytes
*/
// save header
func (dagdb *DagDb) SaveHeader(h *modules.Header) error {

	err := dagdb.saveHeader(dagdb.db, h)
	if err != nil {
		return err
	}
	return dagdb.saveHeaderChainIndex(dagdb.db, h)
}
func (dagdb *DagDb) SaveHeaders(headers []*modules.Header) error {
	batch := dagdb.db.NewBatch()
	for _, h := range headers {
		err := dagdb.saveHeader(batch, h)
		if err != nil {
			return err
		}
		err = dagdb.saveHeaderChainIndex(batch, h)
		if err != nil {
			return err
		}
	}
	return batch.Write()
}
func (dagdb *DagDb) saveHeader(putter ptndb.Putter, h *modules.Header) error {
	uHash := h.Hash()
	key := append(constants.HEADER_PREFIX, uHash.Bytes()...)
	err := StoreToRlpBytes(putter, key[:], h)
	if err != nil {
		log.Error("Save Header error", err.Error())
		return err
	}
	log.Debugf("DB[%s](%p) Save header for unit: %#x,key:%x", reflect.TypeOf(dagdb.db).String(), dagdb, uHash.Bytes(), key)
	return nil
}

//为Unit的Height建立索引,这个索引是必须的，所以在dagdb中实现，而不是在indexdb实现。
func (dagdb *DagDb) saveHeaderChainIndex(putter ptndb.Putter, h *modules.Header) error {
	idxKey := append(constants.HEADER_HEIGTH_PREFIX, h.Number.Bytes()...)
	uHash := h.Hash()
	err := StoreToRlpBytes(putter, idxKey, uHash)
	if err != nil {
		log.Error("Save Header height index error", err.Error())
		return err
	}
	log.Debugf("Save header number %s for unit: %#x", h.Number.String(), uHash.Bytes())
	return nil
}
func (dagdb *DagDb) GetHashByNumber(number *modules.ChainIndex) (common.Hash, error) {
	idxKey := append(constants.HEADER_HEIGTH_PREFIX, number.Bytes()...)
	uHash := common.Hash{}
	data, err := dagdb.db.Get(idxKey)
	if err != nil {
		return uHash, err
	}
	uHash.SetBytes(data)
	return uHash, nil
}

/**
key: [BODY_PREFIX][unit hash]
value: all transactions hash set's rlp encoding bytes
*/
func (dagdb *DagDb) SaveBody(unitHash common.Hash, txsHash []common.Hash) error {
	key := append(constants.BODY_PREFIX, unitHash.Bytes()...)
	return StoreToRlpBytes(dagdb.db, key, txsHash)
}

func (dagdb *DagDb) GetBody(unitHash common.Hash) ([]common.Hash, error) {
	//log.Debug("get unit body info", "unitHash", unitHash.String())
	key := append(constants.BODY_PREFIX, unitHash.Bytes()...)
	var txHashs []common.Hash
	err := RetrieveFromRlpBytes(dagdb.db, key, &txHashs)

	if err != nil {
		return nil, err
	}
	return txHashs, nil
}

func (dagdb *DagDb) SaveTxLookupEntry(unit *modules.Unit) error {
	if len(unit.Txs) == 0 {
		//log.Debugf("No tx in unit[%s] need to save lookup", unit.Hash().String())
		return nil
	}
	//log.Debugf("Batch save tx lookup entry, tx count:%d", len(unit.Txs))
	batch := dagdb.db.NewBatch()
	for i, tx := range unit.Transactions() {
		in := &modules.TxLookupEntry{
			UnitHash:  unit.Hash(),
			UnitIndex: unit.NumberU64(),
			Index:     uint64(i),
			Timestamp: uint64(unit.UnitHeader.Time),
		}
		key := append(constants.LOOKUP_PREFIX, tx.Hash().Bytes()...)

		if err := StoreToRlpBytes(batch, key, in); err != nil {
			return err
		}
	}
	return batch.Write()
}
func (dagdb *DagDb) GetTxLookupEntry(txHash common.Hash) (*modules.TxLookupEntry, error) {
	key := append(constants.LOOKUP_PREFIX, txHash.Bytes()...)
	entry := &modules.TxLookupEntry{}
	err := RetrieveFromRlpBytes(dagdb.db, key, entry)
	if err != nil {
		log.Info("get entry structure info:", "error", err, "tx_entry", entry)
		return nil, err
	}
	return entry, nil
}

// ###################### SAVE IMPL END ######################

// GetTrieSyncProgress stores the fast sync trie process counter to support
// retrieving it across restarts.
func (dagdb *DagDb) GetTrieSyncProgress() (uint64, error) {
	data, err := dagdb.db.Get(constants.TrieSyncKey)
	if err != nil {
		return 0, err
	}
	return new(big.Int).SetBytes(data).Uint64(), nil
}

func (dagdb *DagDb) GetPrefix(prefix []byte) map[string][]byte {
	return getprefix(dagdb.db, prefix)

}

func (dagdb *DagDb) GetUnitTransactions(hash common.Hash) (modules.Transactions, error) {
	txs := modules.Transactions{}
	txHashList, err := dagdb.GetBody(hash)
	if err != nil {
		log.Error(reflect.TypeOf(dagdb.db).String()+": GetUnitTransactions when get body error",
			"error", err.Error(), "unit_hash", hash.String())
		return nil, err
	}
	// get transaction by tx'hash.
	for _, txHash := range txHashList {
		tx, err := dagdb.GetTransactionOnly(txHash)
		if err != nil {
			return nil, err
		}
		txs = append(txs, tx)
	}
	return txs, nil
}

func (dagdb *DagDb) GetHeaderByHash(hash common.Hash) (*modules.Header, error) {
	key := append(constants.HEADER_PREFIX, hash.Bytes()...)
	log.Debugf("DB[%s](%p) Get Header by unit hash:%s,key:%x",
		reflect.TypeOf(dagdb.db).String(), dagdb, hash.String(), key)
	header := new(modules.Header)
	err := RetrieveFromRlpBytes(dagdb.db, key, header)
	if err != nil {
		return nil, err
	}
	return header, nil

}

// PutTrieSyncProgress stores the fast sync trie process counter to support
// retrieving it across restarts.
func (dagdb *DagDb) PutTrieSyncProgress(count uint64) error {
	if err := dagdb.db.Put(constants.TrieSyncKey, new(big.Int).SetUint64(count).Bytes()); err != nil {
		return err
	}
	return nil
}
