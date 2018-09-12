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
	"bytes"
	"errors"
	"fmt"
	"log"
	"math/big"
	"reflect"
	"unsafe"

	"strings"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/hexutil"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/common/rlp"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/syndtr/goleveldb/leveldb/iterator"
)

// DatabaseReader wraps the Get method of a backing data store.
type DatabaseReader interface {
	Get(key []byte) (value []byte, err error)
	NewIterator() iterator.Iterator
	NewIteratorWithPrefix(prefix []byte) iterator.Iterator
}

// @author Albert·Gou
func Retrieve(db ptndb.Database, key string, v interface{}) error {
	//rv := reflect.ValueOf(v)
	//if rv.Kind() != reflect.Ptr || rv.IsNil() {
	//	return errors.New("an invalid argument, the argument must be a non-nil pointer")
	//}

	data, err := db.Get( []byte(key))
	if err != nil {
		return err
	}

	err = rlp.DecodeBytes(data, v)
	if err != nil {
		return err
	}

	return nil
}


// get string
func GetString(db ptndb.Database, key []byte) (string, error) {
	if re, err := db.Get(key); err != nil {
		return "", err
	} else {
		return *(*string)(unsafe.Pointer(&re)), nil
	}
}

// get prefix: return maps
func (dagdb *DagDatabase)GetPrefix(prefix []byte) map[string][]byte {
	return getprefix(dagdb.db, prefix)

}
// get prefix: return maps
func (db *UtxoDatabase)GetPrefix(prefix []byte) map[string][]byte {
	return getprefix(db.db, prefix)
}
// get prefix: return maps
func (db *WorldStateDatabase)GetPrefix(prefix []byte) map[string][]byte {
	return getprefix(db.db, prefix)
}
func (db *IndexDatabase)GetPrefix(prefix []byte) map[string][]byte {
	return getprefix(db.db, prefix)
}
// get prefix
func getprefix(db DatabaseReader, prefix []byte) map[string][]byte {
	iter := db.NewIteratorWithPrefix(prefix)
	result := make(map[string][]byte)
	for iter.Next() {
		key := iter.Key()
		value := make([]byte, 0)
		// 请注意： 直接赋值取得iter.Value()的最后一个指针
		result[string(key)] = append(value, iter.Value()...)
	}
	return result
}

func (dagdb *DagDatabase)GetUnit(hash common.Hash) *modules.Unit {
	// 1. get chainindex
	height, err := dagdb.GetNumberWithUnitHash( hash)
	//fmt.Printf("height=%#v\n", height)
	if err != nil {
		log.Println("GetUnit when GetUnitNumber failed , error:", err)
		return nil
	}
	// 2. unit header
	uHeader, err := dagdb.GetHeader( hash, &height)
	if err != nil {
		log.Println("GetUnit when GetHeader failed , error:", err)
		return nil
	}
	// get unit hash
	uHash := common.Hash{}
	uHash.SetBytes(hash.Bytes())
	// get transaction list
	txs, err := dagdb.GetUnitTransactions(uHash)
	if err != nil {
		log.Println("GetUnit when GetUnitTransactions failed , error:", err)
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
func (dagdb *DagDatabase)GetUnitTransactions(hash common.Hash) (modules.Transactions, error) {
	txs := modules.Transactions{}
	txHashList, err := dagdb.GetBody(hash)
	if err != nil {
		return nil, err
	}
	// get transaction by tx'hash.
	for _, txHash := range txHashList {
		tx, _, _, _ := dagdb.GetTransaction(txHash)
		if err != nil {
			txs = append(txs, tx)
		}
	}
	return txs, nil
}
func (dagdb *DagDatabase)GetUnitFormIndex(number modules.ChainIndex) *modules.Unit {
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
	return dagdb.GetUnit( h)
}

func (dagdb *DagDatabase)GetLastIrreversibleUnit(assetID modules.IDType16) *modules.Unit {
	key := fmt.Sprintf("%s_%s_1_", UNIT_NUMBER_PREFIX, assetID.String())

	data := dagdb.GetPrefix( []byte(key))
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
		return dagdb.GetUnit( unitHash)
	}
	return nil
}

func (dagdb *DagDatabase)GetHeader(hash common.Hash, index *modules.ChainIndex) (*modules.Header, error) {
	encNum := encodeBlockNumber(index.Index)
	key := append(HEADER_PREFIX, encNum...)
	key = append(key, index.Bytes()...)
	header_bytes, err := dagdb.db.Get(append(key, hash.Bytes()...))
	//key := fmt.Sprintf("%s%v_%s_%s", HEADER_PREFIX, index.Index, index.String(), hash.Bytes())
	//header_bytes, err := db.Get([]byte(key))
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

func (dagdb *DagDatabase)GetHeaderByHeight( index modules.ChainIndex) (*modules.Header, error) {
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

func (dagdb *DagDatabase)GetHeaderRlp(hash common.Hash, index uint64) rlp.RawValue {
	encNum := encodeBlockNumber(index)
	key := append(HEADER_PREFIX, encNum...)
	header_bytes, err := dagdb.db.Get(append(key, hash.Bytes()...))
	// rlp  to  Header struct
	log.Println(err)
	return header_bytes
}

func (dagdb *DagDatabase)GetHeaderFormIndex( number modules.ChainIndex) *modules.Header {
	unit :=dagdb. GetUnitFormIndex( number)
	return unit.UnitHeader
}

// GetTxLookupEntry
func (dagdb *DagDatabase)GetTxLookupEntry( hash common.Hash) (common.Hash, uint64, uint64) {
	data, _ := dagdb.db.Get(append(LookupPrefix, hash.Bytes()...))
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
func (dagdb *DagDatabase)GetTransaction(hash common.Hash) (*modules.Transaction, common.Hash, uint64, uint64) {
	unitHash, unitNumber, txIndex :=dagdb. GetTxLookupEntry(hash)
	if unitHash != (common.Hash{}) {
		body, _ :=dagdb. GetBody( unitHash)
		if body == nil || len(body) <= int(txIndex) {
			return nil, common.Hash{}, 0, 0
		}
		tx, err := dagdb. gettrasaction(body[txIndex])
		if err == nil {
			return tx, unitHash, unitNumber, txIndex
		}
	}
	tx, err :=dagdb. gettrasaction( hash)
	if err != nil {
		return nil, unitHash, unitNumber, txIndex
	}
	return tx, unitHash, unitNumber, txIndex
}

// gettrasaction can get a transaction by hash.
func (dagdb *DagDatabase)gettrasaction( hash common.Hash) (*modules.Transaction, error) {
	if hash == (common.Hash{}) {
		return nil, errors.New("hash is not exist.")
	}
	data, err := dagdb.db.Get( append(TRANSACTION_PREFIX, hash.Bytes()...))
	if err != nil {
		return nil, err
	}
	tx := new(modules.Transaction)
	if err := rlp.DecodeBytes(data, &tx); err != nil {
		return nil, err
	}
	return tx, nil
}

func (dagdb *DagDatabase)GetContractNoReader(db ptndb.Database, id common.Hash) (*modules.Contract, error) {
	if common.EmptyHash(id) {
		return nil, errors.New("the filed not defined")
	}
	con_bytes, err := dagdb.db.Get( append(CONTRACT_PTEFIX, id[:]...))
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

// GetContract can get a Contract by the contract hash
func (statedb *WorldStateDatabase)GetContract( id common.Hash) (*modules.Contract, error) {
	if common.EmptyHash(id) {
		return nil, errors.New("the filed not defined")
	}
	con_bytes, err :=statedb.db.Get(append(CONTRACT_PTEFIX, id[:]...))
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

/**
获取合约模板
To get contract template
*/
func (statedb *WorldStateDatabase)GetContractTpl(templateID []byte) (version *modules.StateVersion, bytecode []byte, name string, path string) {
	key := fmt.Sprintf("%s%s^*^bytecode",
		CONTRACT_TPL,
		hexutil.Encode(templateID[:]),
	)
	data :=statedb. GetPrefix( []byte(key))
	if len(data) == 1 {
		for k, v := range data {
			if !version.ParseStringKey(k) {
				return
			}
			if err := rlp.DecodeBytes([]byte(v), &bytecode); err != nil {
				log.Println("GetContractTpl when get bytecode", "error", err.Error())
				return
			}
			break
		}
	}
	_, nameByte :=statedb.GetTplState( templateID, "ContractName")
	if nameByte == nil {
		return
	}
	if err := rlp.DecodeBytes(nameByte, &name); err != nil {
		log.Println("GetContractTpl when get name", "error", err.Error())
		return
	}

	_, pathByte := statedb.GetTplState( templateID, "ContractPath")
	if err := rlp.DecodeBytes(pathByte, &path); err != nil {
		log.Println("GetContractTpl when get path", "error", err.Error())
		return
	}
	return
}

func GetContractRlp(db DatabaseReader, id common.Hash) (rlp.RawValue, error) {
	if common.EmptyHash(id) {
		return nil, errors.New("the filed not defined")
	}
	con_bytes, err := db.Get(append(CONTRACT_PTEFIX, id[:]...))
	if err != nil {
		return nil, err
	}
	return con_bytes, nil
}

// Get contract key's value
func GetContractKeyValue(db DatabaseReader, id common.Hash, key string) (interface{}, error) {
	var val interface{}
	if common.EmptyHash(id) {
		return nil, errors.New("the filed not defined")
	}
	con_bytes, err := db.Get(append(CONTRACT_PTEFIX, id[:]...))
	if err != nil {
		return nil, err
	}
	contract := new(modules.Contract)
	err = rlp.DecodeBytes(con_bytes, contract)
	if err != nil {
		log.Println("err:", err)
		return nil, err
	}
	obj := reflect.ValueOf(contract)
	myref := obj.Elem()
	typeOftype := myref.Type()

	for i := 0; i < myref.NumField(); i++ {
		filed := myref.Field(i)
		if typeOftype.Field(i).Name == key {
			val = filed.Interface()
			log.Println(i, ". ", typeOftype.Field(i).Name, " ", filed.Type(), "=: ", filed.Interface())
			break
		} else if i == myref.NumField()-1 {
			val = nil
		}
	}
	return val, nil
}

const missingNumber = uint64(0xffffffffffffffff)

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
func (dagdb *DagDatabase)GetNumberWithUnitHash( hash common.Hash) (modules.ChainIndex, error) {
	data, _ := dagdb.db.Get(append(UNIT_HASH_NUMBER_Prefix, hash.Bytes()...))
	if len(data) <= 0 {
		return modules.ChainIndex{}, fmt.Errorf("Get from unit number rlp data none")
	}
	var number modules.ChainIndex
	if err := rlp.DecodeBytes(data, &number); err != nil {
		return modules.ChainIndex{}, fmt.Errorf("Get unit number when rlp decode error:%s", err.Error())
	}
	return number, nil
}

//  GetCanonicalHash get

func (dagdb *DagDatabase)GetCanonicalHash(number uint64) (common.Hash, error) {
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
func (dagdb *DagDatabase)GetHeadHeaderHash() (common.Hash, error) {
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
func(dagdb *DagDatabase) GetHeadUnitHash() (common.Hash, error) {
	data, err := dagdb.db.Get(HeadUnitKey)
	if err != nil {
		return common.Hash{}, err
	}
	return common.BytesToHash(data), nil
}

// GetHeadFastUnitHash stores the fast head unit's hash.
func (dagdb *DagDatabase)GetHeadFastUnitHash() (common.Hash, error) {
	data, err := dagdb.db.Get(HeadFastKey)
	if err != nil {
		return common.Hash{}, err
	}
	return common.BytesToHash(data), nil
}

// GetTrieSyncProgress stores the fast sync trie process counter to support
// retrieving it across restarts.
func (dagdb *DagDatabase)GetTrieSyncProgress() (uint64, error) {
	data, err := dagdb.db.Get(TrieSyncKey)
	if err != nil {
		return 0, err
	}
	return new(big.Int).SetBytes(data).Uint64(), nil
}

//  dbFetchUtxoEntry
func (utxodb *UtxoDatabase)GetUtxoEntry( key []byte) (*modules.Utxo, error) {
	utxo := new(modules.Utxo)
	data, err := utxodb.db.Get(key)
	if err != nil {
		return nil, err
	}

	if err := rlp.DecodeBytes(data, &utxo); err != nil {
		return nil, err
	}
	return utxo, nil
}
func (utxodb *UtxoDatabase)GetUtxoByIndex(indexKey []byte) ([]byte,error){
	return utxodb.db.Get(indexKey)
}

// GetAdddrTransactionsHash
func GetAddrTransactionsHash(db DatabaseReader, addr string) ([]common.Hash, error) {
	data, err := db.Get(append(AddrTransactionsHash_Prefix, []byte(addr)...))
	if err != nil {
		return []common.Hash{}, err
	}
	hashs := make([]common.Hash, 0)
	if err := rlp.DecodeBytes(data, hashs); err != nil {
		return []common.Hash{}, err
	}
	return hashs, nil
}

// GetAddrTransactions
func(dagdb *DagDatabase) GetAddrTransactions(addr string) (modules.Transactions, error) {
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
		tx, _, _, _ :=dagdb. GetTransaction( hash)
		txs = append(txs, tx)
	}
	return txs, nil
}

// Get income transactions
func (dagdb *DagDatabase)GetAddrOutput(addr string) ([]modules.Output, error) {

	data :=dagdb. GetPrefix(append(AddrOutput_Prefix, []byte(addr)...))
	outputs := make([]modules.Output, 0)
	var err error
	for _, b := range data {
		out := new(modules.Output)
		if err := rlp.DecodeBytes(b, &out); err == nil {
			outputs = append(outputs, *out)
		} else {
			err = err
		}
	}
	return outputs, err
}

/**
获取模板所有属性
To get contract or contract template all fields and return
*/
func (statedb *WorldStateDatabase)GetTplAllState( id string) map[modules.ContractReadSet][]byte {
	// key format: [PREFIX][ID]_[field]_[version]
	key := fmt.Sprintf("%s%s^*^", CONTRACT_TPL, id)
	data := getprefix(statedb.db, []byte(key))
	if data == nil || len(data) <= 0 {
		return nil
	}
	allState := map[modules.ContractReadSet][]byte{}
	for k, v := range data {
		sKey := strings.Split(k, "^*^")
		if len(sKey) != 3 {
			continue
		}
		var version modules.StateVersion
		if !version.ParseStringKey(key) {
			continue
		}
		rdSet := modules.ContractReadSet{
			Key:   sKey[1],
			Value: &version,
		}
		allState[rdSet] = v
	}
	return allState
}

/**
获取合约（或模板）所有属性
To get contract or contract template all fields and return
*/
func (statedb *WorldStateDatabase)GetContractAllState(id []byte) map[modules.ContractReadSet][]byte {
	// key format: [PREFIX][ID]_[field]_[version]
	key := fmt.Sprintf("%s%s^*^", CONTRACT_STATE_PREFIX, hexutil.Encode(id))
	data := getprefix(statedb.db, []byte(key))
	if data == nil || len(data) <= 0 {
		return nil
	}
	allState := map[modules.ContractReadSet][]byte{}
	for k, v := range data {
		sKey := strings.Split(k, "^*^")
		if len(sKey) != 3 {
			continue
		}
		var version modules.StateVersion
		if !version.ParseStringKey(key) {
			continue
		}
		rdSet := modules.ContractReadSet{
			Key:   sKey[1],
			Value: &version,
		}
		allState[rdSet] = v
	}
	return allState
}

/**
获取合约（或模板）某一个属性
To get contract or contract template one field
*/
func(statedb *WorldStateDatabase) GetTplState( id []byte, field string) (modules.StateVersion, []byte) {
	key := fmt.Sprintf("%s%s^*^%s^*^", CONTRACT_TPL, hexutil.Encode(id[:]), field)
	data := getprefix(statedb.db, []byte(key))
	if data == nil || len(data) != 1 {
		return modules.StateVersion{}, nil
	}
	for k, v := range data {
		var version modules.StateVersion
		if !version.ParseStringKey(k) {
			return modules.StateVersion{}, nil
		}
		return version, v
	}
	return modules.StateVersion{}, nil
}

/**
获取合约（或模板）某一个属性
To get contract or contract template one field
*/
func(statedb *WorldStateDatabase) GetContractState(id string, field string) (modules.StateVersion, []byte) {
	key := fmt.Sprintf("%s%s^*^%s^*^", CONTRACT_STATE_PREFIX, id, field)
	data := getprefix(statedb.db, []byte(key))
	if data == nil || len(data) != 1 {
		return modules.StateVersion{}, nil
	}
	for k, v := range data {
		var version modules.StateVersion
		if !version.ParseStringKey(k) {
			return modules.StateVersion{}, nil
		}
		return version, v
	}
	log.Println("11111111")
	return modules.StateVersion{}, nil
}
func (statedb *WorldStateDatabase) GetAssetInfo( assetId *modules.Asset) (modules.AssetInfo, error){
	key := append(modules.ASSET_INFO_PREFIX, assetId.AssetId.String()...)
	data, err := statedb.db.Get(key)
	if err != nil {
		return modules.AssetInfo{}, err
	}

	var assetInfo modules.AssetInfo
	err = rlp.DecodeBytes(data, &assetInfo)

	if err != nil {
		return assetInfo, err
	}
	return assetInfo, nil
}