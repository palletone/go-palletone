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

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"

	"github.com/ethereum/go-ethereum/rlp"
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
	//log.Debugf("Try to save tx[%s]", txHash.String())
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
	log.Debugf("Save RequestId[%s] map to TxId[%s]", reqid.String(), txhash.String())
	key := append(constants.REQID_TXID_PREFIX, reqid.Bytes()...)
	return dagdb.db.Put(key, txhash.Bytes())
}
func (dagdb *DagDb) GetAllTxs() ([]*modules.Transaction, error) {
	kvs := getprefix(dagdb.db, constants.TRANSACTION_PREFIX)
	result := make([]*modules.Transaction, len(kvs))
	for _, v := range kvs {
		tx := new(modules.Transaction)
		rlp.DecodeBytes(v, tx)
		result = append(result, tx)
	}
	return result, nil
}

//
//func (dagdb *DagDb) saveOutputByAddr(addr string, hash common.Hash, msgindex int, output *modules.Output) error {
//	if hash == (common.Hash{}) {
//		return errors.New("empty tx hash.")
//	}
//	key := append(constants.AddrOutput_Prefix, []byte(addr)...)
//	key = append(key, []byte(hash.String())...)
//	err := StoreToRlpBytes(dagdb.db, append(key, new(big.Int).SetInt64(int64(msgindex)).Bytes()...), output)
//	return err
//}
//
//func (dagdb *DagDb) updateAddrTransactions(tx *modules.Transaction, hash common.Hash) error {
//
//	if hash == (common.Hash{}) {
//		return errors.New("empty tx hash.")
//	}
//	froms, err := dagdb.GetTxFromAddress(tx)
//	if err != nil {
//		return err
//	}
//	// 1. save from_address
//	for _, addr := range froms {
//		go dagdb.saveAddrTxHashByKey(constants.AddrTx_From_Prefix, addr, hash)
//	}
//	// 2. to_address 已经在上层接口处理了。
//	// for _, addr := range tos { // constants.AddrTx_To_Prefix
//	// 	go dagdb.saveAddrTxHashByKey(constants.AddrTx_To_Prefix, addr, hash)
//	// }
//	return nil
//}
//func (dagdb *DagDb) saveAddrTxHashByKey(key []byte, addr string, hash common.Hash) error {
//
//	hashs := make([]common.Hash, 0)
//	data, err := dagdb.db.Get(append(key, []byte(addr)...))
//	if err != nil {
//		if err.Error() != "leveldb: not found" {
//			return err
//		} else { // first store the addr
//			hashs = append(hashs, hash)
//			if err := StoreToRlpBytes(dagdb.db, append(key, []byte(addr)...), hashs); err != nil {
//				return err
//			}
//			return nil
//		}
//	}
//	if err := rlp.DecodeBytes(data, &hashs); err != nil {
//		return err
//	}
//	hashs = append(hashs, hash)
//	if err := StoreToRlpBytes(dagdb.db, append(key, []byte(addr)...), hashs); err != nil {
//		return err
//	}
//	return nil
//}
//
//// Get income transactions
//func (dagdb *DagDb) GetAddrOutput(addr string) ([]modules.Output, error) {
//
//	data := dagdb.GetPrefix(append(constants.AddrOutput_Prefix, []byte(addr)...))
//	outputs := make([]modules.Output, 0)
//	var err error
//	for _, b := range data {
//		out := new(modules.Output)
//		if err := rlp.DecodeBytes(b, out); err == nil {
//			outputs = append(outputs, *out)
//		} else {
//			err = err
//		}
//	}
//	return outputs, err
//}
//
//func (dagdb *DagDb) GetTxFromAddress(tx *modules.Transaction) ([]string, error) {
//
//	froms := make([]string, 0)
//	if tx == nil {
//		return froms, errors.New("tx is nil, not exist address.")
//	}
//	outpoints, _ := tx.GetAddressInfo()
//	for _, op := range outpoints {
//		addr, err := dagdb.getOutpointAddr(op)
//		if err == nil {
//			froms = append(froms, addr)
//		} else {
//			log.Info("get out address is failed.", "error", err)
//		}
//	}
//
//	return froms, nil
//}
//func (dagdb *DagDb) getOutpointAddr(outpoint *modules.OutPoint) (string, error) {
//	if outpoint == nil {
//		return "", fmt.Errorf("outpoint_key is nil ")
//	}
//	out_key := append(constants.OUTPOINT_ADDR_PREFIX, outpoint.ToKey()...)
//	data, err := dagdb.db.Get(out_key[:])
//	if len(data) <= 0 {
//		return "", fmt.Errorf("address is null. outpoint_key(%s)", outpoint.ToKey())
//	}
//	if err != nil {
//		return "", err
//	}
//	var str string
//	err0 := rlp.DecodeBytes(data, &str)
//	return str, err0
//}

// GetTransaction retrieves a specific transaction from the database , along with its added positional metadata
// p2p 同步区块 分为同步header 和body。 GetBody可以省掉节点包装交易块的过程。
//func (dagdb *DagDb) GetTransaction(hash common.Hash) (*modules.TransactionWithUnitInfo, error) {
//	unitHash, unitNumber, txIndex, err1 := dagdb.GetTxLookupEntry(hash)
//	if err1 != nil {
//		log.Info("dag db GetTransaction,GetTxLookupEntry failed.", "error", err1, "tx_hash:", hash)
//		return nil, err1
//	}
//	tx, err := dagdb.GetTransactionOnly(hash)
//	if err != nil {
//		log.Info("GetTransactionOnly error:", err.Error())
//		return nil, err
//	}
//	resultTx := modules.TransactionWithUnitInfo{Transaction: tx, UnitHash: unitHash, UnitHeight: unitNumber, TxIndex: txIndex}
//	return resultTx, nil
//}

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

//func (dagdb *DagDb) GetTransactionByHash(hash common.Hash) (*modules.Transaction, common.Hash, error) {
//	unitHash, _, _, err := dagdb.GetTxLookupEntry(hash)
//	if err != nil {
//		log.Info("dag db GetTransaction,GetTxLookupEntry failed.", "error", err, "tx_hash:", hash)
//		return nil, unitHash, err
//	}
//
//	tx, err1 := dagdb.GetTransactionOnly(hash)
//	return tx, unitHash, err1
//}
