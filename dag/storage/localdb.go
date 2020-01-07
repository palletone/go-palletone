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
	"errors"
	"fmt"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/dag/constants"
	"github.com/palletone/go-palletone/dag/modules"
)

type LocalDb struct {
	db ptndb.Database
}

func NewLocalDb(db ptndb.Database) *LocalDb {
	return &LocalDb{db: db}
}

type ILocalDb interface {
	SaveLocalTx(tx *modules.Transaction) error
	GetLocalTx(txId common.Hash) (*modules.Transaction, error)
	SaveLocalTxStatus(txId common.Hash, status byte) error
	GetLocalTxStatus(txId common.Hash) (byte, error)
}

//通过本地RPC创建或广播的交易
func (db *LocalDb) SaveLocalTx(tx *modules.Transaction) error {
	txId := tx.Hash()
	key := append(constants.LOCAL_TX_PREFIX, txId.Bytes()...)
	err := StoreToRlpBytes(db.db, key, tx)
	if err != nil {
		log.Errorf("Save tx[%s] error:%s", txId.Str(), err.Error())
		return err
	}
	return nil
}

//查询某交易的内容和状态
func (db *LocalDb) GetLocalTx(hash common.Hash) (*modules.Transaction, error) {
	if hash == (common.Hash{}) {
		return nil, errors.New("hash is not exist.")
	}
	tx := new(modules.Transaction)
	key := append(constants.LOCAL_TX_PREFIX, hash.Bytes()...)
	err := RetrieveFromRlpBytes(db.db, key, tx)
	if err != nil {
		log.Warn("get transaction failed.", "tx_hash", hash.String(), "error", err)
		return nil, err
	}
	return tx, nil
}

//保存某交易的状态
func (db *LocalDb) SaveLocalTxStatus(txId common.Hash, status byte) error {
	key := append(constants.LOCAL_TX_STATUS_PREFIX, txId.Bytes()...)
	return db.db.Put(key, []byte{status})
}

func (db *LocalDb) GetLocalTxStatus(txId common.Hash) (byte, error) {
	key := append(constants.LOCAL_TX_STATUS_PREFIX, txId.Bytes()...)
	value, err := db.db.Get(key)
	if err != nil {
		log.Errorf("GetLocalTxStatus by tx[%s] throw error:%s", txId.String(), err.Error())
		return 0, err
	}
	if len(value) != 1 {
		return 0, fmt.Errorf("invalid tx[%s] status value:%x", txId.String(), value)
	}
	return value[0], nil
}
