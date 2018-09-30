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
	"bytes"
	"errors"
	"fmt"
	"github.com/palletone/go-palletone/common"
	ptnLog "github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/common/rlp"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/tokenengine"
	"log"
	"math/big"
	"strings"
)

//对DAG对象的操作，包括：Unit，Tx等
type DagDatabase struct {
	db ptndb.Database
}

func NewDagDatabase(db ptndb.Database) *DagDatabase {
	return &DagDatabase{db: db}
}

type DagDb interface {
	//GetGenesisUnit() (*modules.Unit, error)
	//SaveUnit(unit *modules.Unit, isGenesis bool) error
	SaveHeader(uHash common.Hash, h *modules.Header) error
	SaveTransaction(tx *modules.Transaction) error
	SaveBody(unitHash common.Hash, txsHash []common.Hash) error
	GetBody(unitHash common.Hash) ([]common.Hash, error)
	SaveTransactions(txs *modules.Transactions) error
	SaveNumberByHash(uHash common.Hash, number modules.ChainIndex) error
	SaveHashByNumber(uHash common.Hash, number modules.ChainIndex) error
	SaveTxLookupEntry(unit *modules.Unit) error

	PutCanonicalHash(hash common.Hash, number uint64) error
	PutHeadHeaderHash(hash common.Hash) error
	PutHeadUnitHash(hash common.Hash) error
	PutHeadFastUnitHash(hash common.Hash) error
	PutTrieSyncProgress(count uint64) error
	UpdateHeadByBatch(hash common.Hash, number uint64) error

	GetUnit(hash common.Hash) *modules.Unit
	GetUnitTransactions(hash common.Hash) (modules.Transactions, error)
	GetTransaction(hash common.Hash) (*modules.Transaction, common.Hash, uint64, uint64)
	GetPrefix(prefix []byte) map[string][]byte
	GetHeader(hash common.Hash, index *modules.ChainIndex) (*modules.Header, error)
	GetUnitFormIndex(number modules.ChainIndex) *modules.Unit
	GetHeaderByHeight(index modules.ChainIndex) (*modules.Header, error)
	GetNumberWithUnitHash(hash common.Hash) (modules.ChainIndex, error)
	GetHeaderRlp(hash common.Hash, index uint64) rlp.RawValue
	GetCanonicalHash(number uint64) (common.Hash, error)
	GetAddrOutput(addr string) ([]modules.Output, error)
	GetAddrTransactions(addr string) (modules.Transactions, error)
	GetHeadHeaderHash() (common.Hash, error)
	GetHeadUnitHash() (common.Hash, error)
	GetHeadFastUnitHash() (common.Hash, error)
	GetTrieSyncProgress() (uint64, error)
	GetLastIrreversibleUnit(assetID modules.IDType16) *modules.Unit
}

// ###################### SAVE IMPL START ######################
/**
key: [HEADER_PREFIX][chain index number]_[chain index]_[unit hash]
value: unit header rlp encoding bytes
*/
// save header
func (dagdb *DagDatabase) SaveHeader(uHash common.Hash, h *modules.Header) error {
	// encNum := encodeBlockNumber(h.Number.Index)
	// key := append(HEADER_PREFIX, encNum...)
	// key = append(key, h.Number.Bytes()...)
	// return StoreBytes(dagdb.db, append(key, uHash.Bytes()...), h)
	key := fmt.Sprintf("%s%v_%s_%s", HEADER_PREFIX, h.Number.Index, h.Number.String(), uHash.String())
	log.Println("xxxxxxxxxxxxxxxxxxxxxxxxxx--- Header's key in leveldb ---xxxxxxxxxxxxxxxxxxxxxxxxxxxx ", key)
	return StoreBytes(dagdb.db, []byte(key), h)
}

//這是通過modules.ChainIndex存儲hash
func (dagdb *DagDatabase) SaveNumberByHash(uHash common.Hash, number modules.ChainIndex) error {
	return StoreBytes(dagdb.db, append(UNIT_HASH_NUMBER_Prefix, []byte(uHash.String())...), number)
}

//這是通過hash存儲modules.ChainIndex
func (dagdb *DagDatabase) SaveHashByNumber(uHash common.Hash, number modules.ChainIndex) error {
	i := 0
	if number.IsMain {
		i = 1
	}
	key := fmt.Sprintf("%s_%s_%d_%d", UNIT_NUMBER_PREFIX, number.AssetID.String(), i, number.Index)
	//fmt.Println("SaveHashByNumber=[]byte(key)=>", key)
	//fmt.Println("====",uHash)
	return StoreBytes(dagdb.db, []byte(key), uHash)
}

// height and assetid can get a unit key.
func (dagdb *DagDatabase) SaveUHashIndex(cIndex modules.ChainIndex, uHash common.Hash) error {
	key := fmt.Sprintf("%s_%s_%d", UNIT_NUMBER_PREFIX, cIndex.AssetID.String(), cIndex.Index)
	return Store(dagdb.db, key, uHash.Bytes())
}

/**
key: [BODY_PREFIX][unit hash]
value: all transactions hash set's rlp encoding bytes
*/
func (dagdb *DagDatabase) SaveBody(unitHash common.Hash, txsHash []common.Hash) error {
	// db.Put(append())
	return StoreBytes(dagdb.db, append(BODY_PREFIX, []byte(unitHash.String())...), txsHash)
}

func (dagdb *DagDatabase) GetBody(unitHash common.Hash) ([]common.Hash, error) {
	data, err := dagdb.db.Get(append(BODY_PREFIX, []byte(unitHash.String())...))
	if err != nil {
		return nil, err
	}
	var txHashs []common.Hash
	if err := rlp.DecodeBytes(data, &txHashs); err != nil {
		return nil, err
	}
	return txHashs, nil
}

func (dagdb *DagDatabase) SaveTransactions(txs *modules.Transactions) error {
	key := fmt.Sprintf("%s%s", TRANSACTIONS_PREFIX, txs.Hash())
	return Store(dagdb.db, key, *txs)
}

/**
key: [TRANSACTION_PREFIX][tx hash]
value: transaction struct rlp encoding bytes
*/
func (dagdb *DagDatabase) SaveTransaction(tx *modules.Transaction) error {
	// save transaction
	if err := StoreBytes(dagdb.db, append(TRANSACTION_PREFIX, []byte(tx.TxHash.String())...), tx); err != nil {
		return err
	}

	if err := StoreBytes(dagdb.db, append(Transaction_Index, []byte(tx.TxHash.String())...), tx); err != nil {
		return err
	}
	dagdb.updateAddrTransactions(tx.Address().String(), tx.TxHash)
	// store output by addr
	for i, msg := range tx.TxMessages {
		payload, ok := msg.Payload.(modules.PaymentPayload)
		if ok {
			for _, output := range payload.Output {
				//  pkscript to addr
				addr, _ := tokenengine.GetAddressFromScript(output.PkScript[:])
				dagdb.saveOutputByAddr(addr.String(), tx.TxHash, i, *output)
			}
		}
	}

	return nil
}

func (dagdb *DagDatabase) saveOutputByAddr(addr string, hash common.Hash, msgindex int, output modules.Output) error {
	if hash == (common.Hash{}) {
		return errors.New("empty tx hash.")
	}
	key := append(AddrOutput_Prefix, []byte(addr)...)
	key = append(key, []byte(hash.String())...)
	if err := StoreBytes(dagdb.db, append(key, new(big.Int).SetInt64(int64(msgindex)).Bytes()...), output); err != nil {
		return err
	}
	return nil
}

func (dagdb *DagDatabase) updateAddrTransactions(addr string, hash common.Hash) error {
	if hash == (common.Hash{}) {
		return errors.New("empty tx hash.")
	}
	hashs := make([]common.Hash, 0)
	data, err := dagdb.db.Get(append(AddrTransactionsHash_Prefix, []byte(addr)...))
	if err != nil {
		if err.Error() != "leveldb: not found" {
			return err
		} else { // first store the addr
			hashs = append(hashs, hash)
			if err := StoreBytes(dagdb.db, append(AddrTransactionsHash_Prefix, []byte(addr)...), hashs); err != nil {
				return err
			}
			return nil
		}
	}
	if err := rlp.DecodeBytes(data, hashs); err != nil {
		return err
	}
	hashs = append(hashs, hash)
	if err := StoreBytes(dagdb.db, append(AddrTransactionsHash_Prefix, []byte(addr)...), hashs); err != nil {
		return err
	}
	return nil
}

func (dagdb *DagDatabase) SaveTxLookupEntry(unit *modules.Unit) error {
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
		if err := StoreBytes(dagdb.db, append(LookupPrefix, []byte(tx.TxHash.String())...), data); err != nil {
			return err
		}
	}
	return nil
}

// ###################### SAVE IMPL END ######################
// ###################### GET IMPL START ######################
// GetAddrTransactions
func (dagdb *DagDatabase) GetAddrTransactions(addr string) (modules.Transactions, error) {
	data, err := dagdb.db.Get(append(AddrTransactionsHash_Prefix, []byte(addr)...))
	if err != nil {
		return modules.Transactions{}, err
	}
	hashs := make([]common.Hash, 0)
	if err := rlp.DecodeBytes(data, hashs); err != nil {
		return modules.Transactions{}, err
	}
	txs := make(modules.Transactions, 0)
	for _, hash := range hashs {
		tx, _, _, _ := dagdb.GetTransaction(hash)
		txs = append(txs, tx)
	}
	return txs, nil
}

// Get income transactions
func (dagdb *DagDatabase) GetAddrOutput(addr string) ([]modules.Output, error) {

	data := dagdb.GetPrefix(append(AddrOutput_Prefix, []byte(addr)...))
	outputs := make([]modules.Output, 0)
	var err error
	for _, b := range data {
		out := new(modules.Output)
		if err := rlp.DecodeBytes(b, out); err == nil {
			outputs = append(outputs, *out)
		} else {
			err = err
		}
	}
	return outputs, err
}

//func GetUnitNumber(db DatabaseReader, hash common.Hash) (modules.ChainIndex, error) {
//	data, _ := db.Get(append(UNIT_HASH_NUMBER_Prefix, hash.Bytes()...))
//	if len(data) <= 0 {
//		return modules.ChainIndex{}, fmt.Errorf("Get from unit number rlp data none")
//	}
//	var number modules.ChainIndex
//	if err := rlp.DecodeBytes(data, &number); err != nil {
//		return modules.ChainIndex{}, fmt.Errorf("Get unit number when rlp decode error:%s", err.Error())
//	}
//	return number, nil
//}
func (dagdb *DagDatabase) GetNumberWithUnitHash(hash common.Hash) (modules.ChainIndex, error) {
	data, _ := dagdb.db.Get(append(UNIT_HASH_NUMBER_Prefix, []byte(hash.String())...))
	if len(data) <= 0 {
		return modules.ChainIndex{}, nil
	}
	var number modules.ChainIndex
	if err := rlp.DecodeBytes(data, &number); err != nil {
		return modules.ChainIndex{}, fmt.Errorf("Get unit number when rlp decode error:%s", err.Error())
	}
	return number, nil
}

//  GetCanonicalHash get

func (dagdb *DagDatabase) GetCanonicalHash(number uint64) (common.Hash, error) {
	key := append(HEADER_PREFIX, encodeBlockNumber(number)...)
	data, err := dagdb.db.Get(append(key, NumberSuffix...))
	if err != nil {
		return common.Hash{}, err
	}
	if len(data) == 0 {
		return common.Hash{}, err
	}
	return common.BytesToHash(data), nil
}
func (dagdb *DagDatabase) GetHeadHeaderHash() (common.Hash, error) {
	data, err := dagdb.db.Get(HeadHeaderKey)
	if err != nil {
		return common.Hash{}, err
	}
	if len(data) != 8 {
		return common.Hash{}, errors.New("data's len is error.")
	}
	return common.BytesToHash(data), nil
}

// GetHeadUnitHash stores the head unit's hash.
func (dagdb *DagDatabase) GetHeadUnitHash() (common.Hash, error) {
	data, err := dagdb.db.Get(HeadUnitKey)
	if err != nil {
		return common.Hash{}, err
	}
	return common.BytesToHash(data), nil
}

// GetHeadFastUnitHash stores the fast head unit's hash.
func (dagdb *DagDatabase) GetHeadFastUnitHash() (common.Hash, error) {
	data, err := dagdb.db.Get(HeadFastKey)
	if err != nil {
		return common.Hash{}, err
	}
	return common.BytesToHash(data), nil
}

// GetTrieSyncProgress stores the fast sync trie process counter to support
// retrieving it across restarts.
func (dagdb *DagDatabase) GetTrieSyncProgress() (uint64, error) {
	data, err := dagdb.db.Get(TrieSyncKey)
	if err != nil {
		return 0, err
	}
	return new(big.Int).SetBytes(data).Uint64(), nil
}

func (dagdb *DagDatabase) GetPrefix(prefix []byte) map[string][]byte {
	return getprefix(dagdb.db, prefix)

}

func (dagdb *DagDatabase) GetUnit(hash common.Hash) *modules.Unit {
	// 1. get chainindex
	height, err := dagdb.GetNumberWithUnitHash(hash)
	//fmt.Printf("height=%#v\n", height)
	if err != nil {
		log.Println("GetUnit when GetUnitNumber failed , error:", err)
		return nil
	}
	// 2. unit header
	uHeader, err := dagdb.GetHeader(hash, &height)
	if err != nil {
		log.Println("GetUnit when GetHeader failed , error:", err, "hash", hash.String(), "height", height, "index", height.Index)
		return nil
	}
	// get unit hash
	uHash := common.Hash{}
	uHash.SetBytes(hash.Bytes())
	// get transaction list
	txs, err := dagdb.GetUnitTransactions(uHash)
	if err != nil {
		log.Println("GetUnit when GetUnitTransactions failed , error:", err)
		//TODO xiaozhi
		return nil
	}
	// generate unit
	unit := &modules.Unit{
		UnitHeader: uHeader,
		UnitHash:   uHash,
		Txs:        txs,
	}
	unit.UnitSize = unit.Size()
	return unit
}
func (dagdb *DagDatabase) GetUnitTransactions(hash common.Hash) (modules.Transactions, error) {
	txs := modules.Transactions{}
	txHashList, err := dagdb.GetBody(hash)
	if err != nil {
		ptnLog.Info("GetUnitTransactions when get body error", "error", err.Error())
		return nil, err
	}
	// get transaction by tx'hash.
	for _, txHash := range txHashList {
		tx, _, _, _ := dagdb.GetTransaction(txHash)
		txs = append(txs, tx)
	}
	return txs, nil
}
func (dagdb *DagDatabase) GetUnitFormIndex(number modules.ChainIndex) *modules.Unit {
	i := 0
	if number.IsMain {
		i = 1
	}
	key := fmt.Sprintf("%s_%s_%d_%d", UNIT_NUMBER_PREFIX, number.AssetID.String(), i, number.Index)
	hash, err := dagdb.db.Get([]byte(key))
	if err != nil {
		return nil
	}
	var h common.Hash
	h.SetBytes(hash)
	return dagdb.GetUnit(h)
}

func (dagdb *DagDatabase) GetLastIrreversibleUnit(assetID modules.IDType16) *modules.Unit {
	key := fmt.Sprintf("%s_%s_1_", UNIT_NUMBER_PREFIX, assetID.String())

	data := dagdb.GetPrefix([]byte(key))
	irreKey := string("")
	for k := range data {
		if strings.Compare(k, irreKey) > 0 {
			irreKey = k
		}
	}
	if strings.Compare(irreKey, "") > 0 {
		rlpUnitHash := data[irreKey]
		var unitHash common.Hash
		if err := rlp.DecodeBytes(rlpUnitHash, &unitHash); err != nil {
			log.Println("GetLastIrreversibleUnit error:", err.Error())
			return nil
		}
		return dagdb.GetUnit(unitHash)
	}
	return nil
}

func (dagdb *DagDatabase) GetHeader(hash common.Hash, index *modules.ChainIndex) (*modules.Header, error) {
	// encNum := encodeBlockNumber(index.Index)
	// key := append(HEADER_PREFIX, encNum...)
	// key = append(key, index.Bytes()...)
	// header_bytes, err := dagdb.db.Get(append(key, hash.Bytes()...))
	key := fmt.Sprintf("%s%v_%s_%s", HEADER_PREFIX, index.Index, index.String(), hash.String())
	log.Println("XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXx", key)
	header_bytes, err := dagdb.db.Get([]byte(key))
	// rlp  to  Header struct
	if err != nil {
		return nil, err
	}
	header := new(modules.Header)
	if err := rlp.Decode(bytes.NewReader(header_bytes), header); err != nil {
		log.Println("Invalid unit header rlp:", err)
		return nil, err
	}
	return header, nil
}

func (dagdb *DagDatabase) GetHeaderByHeight(index modules.ChainIndex) (*modules.Header, error) {
	encNum := encodeBlockNumber(index.Index)
	key := append(HEADER_PREFIX, encNum...)
	key = append(key, index.Bytes()...)

	data := getprefix(dagdb.db, key)
	if data == nil || len(data) <= 0 {
		return nil, fmt.Errorf("No such height header")
	}
	for _, v := range data {
		header := new(modules.Header)
		if err := rlp.Decode(bytes.NewReader(v), header); err != nil {
			return nil, fmt.Errorf("Invalid unit header rlp: %s", err.Error())
		}
		return header, nil
	}
	return nil, fmt.Errorf("No such height header")
}

func (dagdb *DagDatabase) GetHeaderRlp(hash common.Hash, index uint64) rlp.RawValue {
	encNum := encodeBlockNumber(index)
	key := append(HEADER_PREFIX, encNum...)
	header_bytes, err := dagdb.db.Get(append(key, hash.Bytes()...))
	// rlp  to  Header struct
	log.Println(err)
	return header_bytes
}

func (dagdb *DagDatabase) GetHeaderFormIndex(number modules.ChainIndex) *modules.Header {
	unit := dagdb.GetUnitFormIndex(number)
	return unit.UnitHeader
}

// GetTxLookupEntry
func (dagdb *DagDatabase) GetTxLookupEntry(hash common.Hash) (common.Hash, uint64, uint64) {
	data, _ := dagdb.db.Get(append(LookupPrefix, []byte(hash.String())...))
	if len(data) == 0 {
		return common.Hash{}, 0, 0
	}
	var entry modules.TxLookupEntry
	if err := rlp.DecodeBytes(data, &entry); err != nil {
		return common.Hash{}, 0, 0
	}
	return entry.UnitHash, entry.UnitIndex, entry.Index

}

// GetTransaction retrieves a specific transaction from the database , along with its added positional metadata
// p2p 同步区块 分为同步header 和body。 GetBody可以省掉节点包装交易块的过程。
func (dagdb *DagDatabase) GetTransaction(hash common.Hash) (*modules.Transaction, common.Hash, uint64, uint64) {
	unitHash, unitNumber, txIndex := dagdb.GetTxLookupEntry(hash)
	if unitHash != (common.Hash{}) {
		body, _ := dagdb.GetBody(unitHash)
		if body == nil || len(body) <= int(txIndex) {
			return nil, common.Hash{}, 0, 0
		}
		tx, err := dagdb.gettrasaction(body[txIndex])
		if err == nil {
			return tx, unitHash, unitNumber, txIndex
		}
	}
	tx, err := dagdb.gettrasaction(hash)
	if err != nil {
		fmt.Println("gettrasaction error:", err.Error())
		return nil, unitHash, unitNumber, txIndex
	}

	return tx, unitHash, unitNumber, txIndex
}

// gettrasaction can get a transaction by hash.
func (dagdb *DagDatabase) gettrasaction(hash common.Hash) (*modules.Transaction, error) {
	if hash == (common.Hash{}) {
		return nil, errors.New("hash is not exist.")
	}
	//TODO xiaozhi
	data, err := dagdb.db.Get(append(TRANSACTION_PREFIX, []byte(hash.String())...))
	if err != nil {
		return nil, err
	}

	tx := new(modules.Transaction)
	if err := rlp.DecodeBytes(data, tx); err != nil {
		return nil, err
	}
	return tx, nil
}

func (dagdb *DagDatabase) GetContractNoReader(db ptndb.Database, id common.Hash) (*modules.Contract, error) {
	if common.EmptyHash(id) {
		return nil, errors.New("the filed not defined")
	}
	con_bytes, err := dagdb.db.Get(append(CONTRACT_PTEFIX, id[:]...))
	if err != nil {
		log.Println("err:", err)
		return nil, err
	}
	contract := new(modules.Contract)
	err = rlp.DecodeBytes(con_bytes, contract)
	if err != nil {
		log.Println("err:", err)
		return nil, err
	}
	return contract, nil
}

// ###################### GET IMPL END ######################
