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

package storage

import (
	"errors"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/dag/constants"
	"github.com/palletone/go-palletone/dag/modules"
)

/**
key: [TRANSACTION_PREFIX][tx hash]
value: transaction struct rlp encoding bytes
*/
func (dagdb *DagDb) SaveTransaction(tx *modules.Transaction) error {
	// save transaction
	txHash := tx.Hash()
	//Save tx to db
	key := append(constants.TRANSACTION_PREFIX, txHash.Bytes()...)
	err := StoreToRlpBytes(dagdb.db, key, tx)
	if err != nil {
		log.Errorf("Save tx[%s] error:%s", txHash.Str(), err.Error())
		return err
	}
	//Save reqid
	if tx.IsContractTx() {
		if err := dagdb.saveReqIdByTx(tx); err != nil {
			log.Error("SaveReqIdByTx is failed,", "error", err)
		}
	}
	return nil
}
func (dagdb *DagDb) saveReqIdByTx(tx *modules.Transaction) error {
	txhash := tx.Hash()
	reqid := tx.RequestHash()
	if txhash == reqid {
		return nil
	}
	log.Debugf("Save RequestId[%s] map to TxId[%s]", reqid.String(), txhash.String())
	key := append(constants.REQID_TXID_PREFIX, reqid.Bytes()...)
	return dagdb.db.Put(key, txhash.Bytes())
}
func (dagdb *DagDb) GetAllTxs() ([]*modules.Transaction, error) {
	kvs := getprefix(dagdb.db, constants.TRANSACTION_PREFIX)
	result := make([]*modules.Transaction, 0, len(kvs))
	for k, v := range kvs {
		tx := new(modules.Transaction)
		err := rlp.DecodeBytes(v, tx)
		if err != nil {
			log.Errorf("Cannot decode key[%s] rlp tx:%x", k, v)
			return nil, err
		}
		result = append(result, tx)
	}
	return result, nil
}

// GetTransactionOnly can get a transaction by hash.
func (dagdb *DagDb) GetTransactionOnly(hash common.Hash) (*modules.Transaction, error) {
	if hash == (common.Hash{}) {
		return nil, errors.New("hash is not exist.")
	}
	tx := new(modules.Transaction)
	key := append(constants.TRANSACTION_PREFIX, hash.Bytes()...)
	err := RetrieveFromRlpBytes(dagdb.db, key, tx)
	if err != nil {
		log.Warn("get transaction failed.", "tx_hash", hash.String(), "error", err)
		return nil, err
	}
	return tx, nil
}

func (dagdb *DagDb) IsTransactionExist(hash common.Hash) (bool, error) {
	key := append(constants.TRANSACTION_PREFIX, hash.Bytes()...)
	exist, err := dagdb.db.Has(key)
	if err != nil {
		log.Warnf("Check tx is exist throw error:%s", err.Error())
		return false, err
	}
	return exist, nil
}

func (dagdb *DagDb) GetTxHashByReqId(reqid common.Hash) (common.Hash, error) {
	key := append(constants.REQID_TXID_PREFIX, reqid.Bytes()...)
	txid := common.Hash{}
	val, err := dagdb.db.Get(key)
	if err != nil {
		return txid, err
	}
	txid.SetBytes(val)

	return txid, err
}
func (dagdb *DagDb) ForEachAllTxDo(txAction func(key []byte, transaction *modules.Transaction) error) error {
	iter := dagdb.db.NewIteratorWithPrefix(constants.TRANSACTION_PREFIX)
	for iter.Next() {
		key := iter.Key()
		value := iter.Value()
		tx := new(modules.Transaction)
		err := rlp.DecodeBytes(value, tx)
		if err != nil {
			log.Errorf("Cannot decode key[%s] rlp tx:%x", key, value)
			return err
		}
		err = txAction(key, tx)
		if err != nil {
			log.Errorf("tx[%s] action error:%s", tx.Hash().String(), err.Error())
			return err
		}
	}
	return nil
}
