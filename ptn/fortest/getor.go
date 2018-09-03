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

package fortest

import (
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"unsafe"
	"log"
	"fmt"
	"bytes"
	"math/big"
	"github.com/palletone/go-palletone/common/rlp"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/dag/storage"
	"errors"
	"github.com/palletone/go-palletone/common/ptndb"
)

// DatabaseReader wraps the Get method of a backing data store.
type DatabaseReader interface {
	Get(key []byte) (value []byte, err error)
	NewIterator() iterator.Iterator
	NewIteratorWithPrefix(prefix []byte) iterator.Iterator
}

// @author Albert·Gou
func Retrieve(Dbconn ptndb.Database,key string, v interface{}) error {
	//rv := reflect.ValueOf(v)
	//if rv.Kind() != reflect.Ptr || rv.IsNil() {
	//	return errors.New("an invalid argument, the argument must be a non-nil pointer")
	//}

	data, err := Get(Dbconn,[]byte(key))
	if err != nil {
		return err
	}
	err = rlp.DecodeBytes(data, v)
	if err != nil {
		return err
	}

	return nil
}

// get bytes
func Get(Dbconn ptndb.Database,key []byte) ([]byte, error) {
	// return Dbconn.Get(key)
	b, err := Dbconn.Get(key)
	return b, err
}

// get string
func GetString(Dbconn ptndb.Database,key []byte) (string, error) {
	if re, err := Dbconn.Get(key); err != nil {
		return "", err
	} else {
		return *(*string)(unsafe.Pointer(&re)), nil
	}
}

// get prefix: return maps
func GetPrefix(Dbconn ptndb.Database,prefix []byte) map[string][]byte {
	return getprefix(Dbconn, prefix)

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

func GetUnit(db ptndb.Database, hash common.Hash) *Unit {
	// 1. get chainindex
	height, err := GetUnitNumber(db, hash)
	if err != nil {
		//log.Println("Getunit when get unitNumber failed , error:", err)
		return nil
	}
	// 2. unit header
	uHeader, err := GetHeader(db, hash, &height)
	if err != nil {
		log.Println("Getunit when get header failed , error:", err)
		return nil
	}

	// get unit hash
	uHash := common.Hash{}
	uHash.SetBytes(hash.Bytes())

	// get transaction list
	tx, err := GetUnitTransactions(db,uHash)
	if err != nil {
		log.Println("Getunit when get transactions failed , error:", err)
		return nil
	}
	txs := Transactions{tx}
	// generate unit
	unit := &Unit{
		UnitHeader: uHeader,
		UnitHash:   uHash,
		Txs:        txs,
	}
	unit.UnitSize = unit.Size()
	return unit
}



func GetBody(Dbconn ptndb.Database,unitHash common.Hash) ([]common.Hash, error) {
	data, err := Get(Dbconn,append(storage.BODY_PREFIX, unitHash.Bytes()...))
	if err != nil {
		return nil, err
	}
	var txHashs []common.Hash
	if err := rlp.DecodeBytes(data, &txHashs); err != nil {
		return nil, err
	}
	return txHashs, nil
}
//func GetUnitTransactions(db DatabaseReader, hash common.Hash) (Transactions, error) {
//	txs := Transactions{}
//	txHashList, err := GetBody(hash)
//	if err != nil {
//		return nil, err
//	}
//	// get transaction by tx'hash.
//	for _, txHash := range txHashList {
//		tx, _, _, _ := GetTransaction(txHash)
//		if err != nil {
//			txs = append(txs, tx)
//		}
//	}
//	return txs, nil
//}
func (it *IDType16) String() string {
	result := string("")
	for _, b := range it {
		result += fmt.Sprintf("%x", b)
	}
	return result
	//var b []byte
	//length := len(it)
	//for _, v := range it {
	//	b = append(b, v)
	//}
	//count := 0
	//for i := length - 1; i >= 0; i-- {
	//	if b[i] == ' ' || b[i] == 0 {
	//		count++
	//	} else {
	//		break
	//	}
	//}
	//return util.ToString(b[:length-count])
}
func GetUnitFormIndex(db ptndb.Database, height uint64, asset IDType16) *Unit {
	key := fmt.Sprintf("%s_%s_%d", storage.UNIT_NUMBER_PREFIX, asset.String(), height)
	hash, err := db.Get([]byte(key))
	if err != nil {
		return nil
	}
	var h common.Hash
	h.SetBytes(hash)
	return GetUnit(db, h)
}

func GetHeader(db ptndb.Database, hash common.Hash, index *ChainIndex) (*Header, error) {
	encNum := encodeBlockNumber(index.Index)
	key := append(storage.HEADER_PREFIX, encNum...)
	key = append(key, index.Bytes()...)

	header_bytes, err := db.Get(append(key, hash.Bytes()...))

	if err != nil {
		return nil, err
	}
	header := new(Header)
	if err := rlp.Decode(bytes.NewReader(header_bytes), header); err != nil {
		log.Println("Invalid unit header rlp:", err)
		return nil, err
	}
	return header, nil
}

func GetHeaderByHeight(db ptndb.Database, index ChainIndex) (*Header, error) {
	encNum := encodeBlockNumber(index.Index)
	key := append(storage.HEADER_PREFIX, encNum...)
	key = append(key, index.Bytes()...)
	data := getprefix(db, key)
	if data == nil || len(data) <= 0 {
		return nil, fmt.Errorf("No such height header")
	}
	for _, v := range data {
		header := new(Header)
		if err := rlp.Decode(bytes.NewReader(v), header); err != nil {
			return nil, fmt.Errorf("Invalid unit header rlp: %s", err.Error())
		}
		return header, nil
	}
	return nil, fmt.Errorf("No such height header")
}

func GetHeaderRlp(db ptndb.Database, hash common.Hash, index uint64) rlp.RawValue {
	encNum := encodeBlockNumber(index)
	key := append(storage.HEADER_PREFIX, encNum...)
	header_bytes, err := db.Get(append(key, hash.Bytes()...))
	// rlp  to  Header struct
	log.Println(err)
	return header_bytes
}

func GetHeaderFormIndex(db ptndb.Database, height uint64, asset IDType16) *Header {
	unit := GetUnitFormIndex(db, height, asset)
	return unit.UnitHeader
}

// GetTxLookupEntry
func GetTxLookupEntry(Dbconn ptndb.Database, hash common.Hash) (common.Hash, uint64, uint64) {
	data, _ := Get(Dbconn,append(storage.LookupPrefix, hash.Bytes()...))
	if len(data) == 0 {
		return common.Hash{}, 0, 0
	}
	var entry TxLookupEntry
	if err := rlp.DecodeBytes(data, &entry); err != nil {
		return common.Hash{}, 0, 0
	}
	return entry.UnitHash, entry.UnitIndex, entry.Index

}
type TxLookupEntry struct {
	UnitHash  common.Hash
	UnitIndex uint64
	Index     uint64
}
// GetTransaction retrieves a specific transaction from the database , along with its added positional metadata
// p2p 同步区块 分为同步header 和body。 GetBody可以省掉节点包装交易块的过程。
func GetTransaction(Dbconn ptndb.Database,hash common.Hash) (*Transaction, common.Hash, uint64, uint64) {
	unitHash, unitNumber, txIndex := GetTxLookupEntry(Dbconn, hash)
	if unitHash != (common.Hash{}) {
		body, _ := GetBody(Dbconn,unitHash)
		if body == nil || len(body) <= int(txIndex) {
			return nil, common.Hash{}, 0, 0
		}
		tx, err := gettrasaction(Dbconn,body[txIndex])
		if err == nil {
			return tx, unitHash, unitNumber, txIndex
		}
	}
	tx, err := gettrasaction(Dbconn,hash)
	if err != nil {
		return nil, unitHash, unitNumber, txIndex
	}
	return tx, unitHash, unitNumber, txIndex
}

// gettrasaction can get a transaction by hash.
func gettrasaction(Dbconn ptndb.Database,hash common.Hash) (*Transaction, error) {
	if hash == (common.Hash{}) {
		return nil, errors.New("hash is not exist.")
	}

	data, err := Get(Dbconn,append(storage.TRANSACTION_PREFIX, hash.Bytes()...))
	if err != nil {
		return nil, err
	}
	tx := new(Transaction)
	if err := rlp.DecodeBytes(data, &tx); err != nil {
		return nil, err
	}
	return tx, nil
}
//
//func GetContractNoReader(id common.Hash) (*Contract, error) {
//	if common.EmptyHash(id) {
//		return nil, errors.New("the filed not defined")
//	}
//	con_bytes, err := Get(append(storage.CONTRACT_PTEFIX, id[:]...))
//	if err != nil {
//		log.Println("err:", err)
//		return nil, err
//	}
//	contract := new(Contract)
//	err = rlp.DecodeBytes(con_bytes, contract)
//	if err != nil {
//		log.Println("err:", err)
//		return nil, err
//	}
//	return contract, nil
//}
//
//// GetContract can get a Contract by the contract hash
//func GetContract(db DatabaseReader, id common.Hash) (*Contract, error) {
//	if common.EmptyHash(id) {
//		return nil, errors.New("the filed not defined")
//	}
//	con_bytes, err := db.Get(append(CONTRACT_PTEFIX, id[:]...))
//	if err != nil {
//		log.Println("err:", err)
//		return nil, err
//	}
//	contract := new(Contract)
//	err = rlp.DecodeBytes(con_bytes, contract)
//	if err != nil {
//		log.Println("err:", err)
//		return nil, err
//	}
//	return contract, nil
//}

/**
获取合约模板
To get contract template
*/
//
//func GetContractTpl(templateID []byte) (version *StateVersion, bytecode []byte, name string, path string) {
//	key := fmt.Sprintf("%s%s^*^bytecode",
//		CONTRACT_TPL,
//		hexutil.Encode(templateID[:]),
//	)
//	data := GetPrefix([]byte(key))
//	if len(data) == 1 {
//		for k, v := range data {
//			if !version.ParseStringKey(k) {
//				return
//			}
//			if err := rlp.DecodeBytes([]byte(v), &bytecode); err != nil {
//				log.Println("GetContractTpl when get bytecode", "error", err.Error())
//				return
//			}
//			break
//		}
//	}
//	_, nameByte := GetTplState(templateID, "ContractName")
//	if nameByte == nil {
//		return
//	}
//	if err := rlp.DecodeBytes(nameByte, &name); err != nil {
//		log.Println("GetContractTpl when get name", "error", err.Error())
//		return
//	}
//
//	_, pathByte := GetTplState(templateID, "ContractPath")
//	if err := rlp.DecodeBytes(pathByte, &path); err != nil {
//		log.Println("GetContractTpl when get path", "error", err.Error())
//		return
//	}
//	return
//}
//
//func GetContractRlp(db DatabaseReader, id common.Hash) (rlp.RawValue, error) {
//	if common.EmptyHash(id) {
//		return nil, errors.New("the filed not defined")
//	}
//	con_bytes, err := db.Get(append(CONTRACT_PTEFIX, id[:]...))
//	if err != nil {
//		return nil, err
//	}
//	return con_bytes, nil
//}
//
//// Get contract key's value
//func GetContractKeyValue(db DatabaseReader, id common.Hash, key string) (interface{}, error) {
//	var val interface{}
//	if common.EmptyHash(id) {
//		return nil, errors.New("the filed not defined")
//	}
//	con_bytes, err := db.Get(append(CONTRACT_PTEFIX, id[:]...))
//	if err != nil {
//		return nil, err
//	}
//	contract := new(Contract)
//	err = rlp.DecodeBytes(con_bytes, contract)
//	if err != nil {
//		log.Println("err:", err)
//		return nil, err
//	}
//	obj := reflect.ValueOf(contract)
//	myref := obj.Elem()
//	typeOftype := myref.Type()
//
//	for i := 0; i < myref.NumField(); i++ {
//		filed := myref.Field(i)
//		if typeOftype.Field(i).Name == key {
//			val = filed.Interface()
//			log.Println(i, ". ", typeOftype.Field(i).Name, " ", filed.Type(), "=: ", filed.Interface())
//			break
//		} else if i == myref.NumField()-1 {
//			val = nil
//		}
//	}
//	return val, nil
//}

const missingNumber = uint64(0xffffffffffffffff)

func GetUnitNumber(db ptndb.Database, hash common.Hash) (ChainIndex, error) {
	data,_ := db.Get(append(storage.UNIT_HASH_NUMBER_Prefix, hash.Bytes()...))
	if len(data) <= 0 {
		return ChainIndex{}, fmt.Errorf("Get from unit number rlp data none")
	}
	var number ChainIndex
	if err := rlp.DecodeBytes(data, &number); err != nil {
		return ChainIndex{}, fmt.Errorf("Get unit number when rlp decode error:%s", err.Error())
	}
	return number, nil
}

//  GetCanonicalHash get

func GetCanonicalHash(db ptndb.Database, number uint64) (common.Hash, error) {
	key := append(storage.HEADER_PREFIX, encodeBlockNumber(number)...)
	data, err := db.Get(append(key, storage.NumberSuffix...))
	if err != nil {
		return common.Hash{}, err
	}
	if len(data) == 0 {
		return common.Hash{}, err
	}
	return common.BytesToHash(data), nil
}
func GetHeadHeaderHash(db ptndb.Database) (common.Hash, error) {
	data, err := db.Get(storage.HeadHeaderKey)
	if err != nil {
		return common.Hash{}, err
	}
	if len(data) != 8 {
		return common.Hash{}, errors.New("data's len is error.")
	}
	return common.BytesToHash(data), nil
}

// GetHeadUnitHash stores the head unit's hash.
func GetHeadUnitHash(db ptndb.Database) (common.Hash, error) {
	data, err := db.Get(storage.HeadUnitKey)

	if err != nil {
		return common.Hash{}, err
	}
	return common.BytesToHash(data), nil
}

// GetHeadFastUnitHash stores the fast head unit's hash.
func GetHeadFastUnitHash(db ptndb.Database) (common.Hash, error) {
	data, err := db.Get(storage.HeadFastKey)
	if err != nil {
		return common.Hash{}, err
	}
	return common.BytesToHash(data), nil
}

// GetTrieSyncProgress stores the fast sync trie process counter to support
// retrieving it across restarts.
func GetTrieSyncProgress(db ptndb.Database) (uint64, error) {
	data, err := db.Get(storage.TrieSyncKey)
	if err != nil {
		return 0, err
	}
	return new(big.Int).SetBytes(data).Uint64(), nil
}

//  dbFetchUtxoEntry
//
//func GetUtxoEntry(db DatabaseReader, key []byte) (*Utxo, error) {
//	utxo := new(Utxo)
//	data, err := db.Get(key)
//	if err != nil {
//		return nil, err
//	}
//
//	if err := rlp.DecodeBytes(data, &utxo); err != nil {
//		return nil, err
//	}
//
//	return utxo, nil
//}

// GetAdddrTransactionsHash
func GetAddrTransactionsHash(db ptndb.Database, addr string) ([]common.Hash, error) {
	data, err := db.Get(append(storage.AddrTransactionsHash_Prefix, []byte(addr)...))
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
func GetAddrTransactions(db ptndb.Database,addr string) (Transactions, error) {
	data, err := db.Get(append(storage.AddrTransactionsHash_Prefix, []byte(addr)...))
	if err != nil {
		return Transactions{}, err
	}
	hashs := make([]common.Hash, 0)
	if err := rlp.DecodeBytes(data, hashs); err != nil {
		return Transactions{}, err
	}
	txs := make(Transactions, 0)
	for _, hash := range hashs {
		tx, _, _, _ := GetTransaction(db,hash)
		txs = append(txs, tx)
	}
	return txs, nil
}

// Get income transactions
func GetAddrOutput(Dbconn ptndb.Database, addr string) ([]Output, error) {

	data := GetPrefix(Dbconn,append(storage.AddrOutput_Prefix, []byte(addr)...))
	outputs := make([]Output, 0)
	var err error
	for _, b := range data {
		out := new(Output)
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
//
//func GetTplAllState(id string) map[ContractReadSet][]byte {
//	// key format: [PREFIX][ID]_[field]_[version]
//	key := fmt.Sprintf("%s%s^*^", storage.CONTRACT_TPL, id)
//	data := getprefix(Dbconn, []byte(key))
//	if data == nil || len(data) <= 0 {
//		return nil
//	}
//	allState := map[ContractReadSet][]byte{}
//	for k, v := range data {
//		sKey := strings.Split(k, "^*^")
//		if len(sKey) != 3 {
//			continue
//		}
//		var version StateVersion
//		if !version.ParseStringKey(key) {
//			continue
//		}
//		rdSet := ContractReadSet{
//			Key:   sKey[1],
//			Value: &version,
//		}
//		allState[rdSet] = v
//	}
//	return allState
//}

/**
获取合约（或模板）所有属性
To get contract or contract template all fields and return
*/
//
//func GetContractAllState(id []byte) map[ContractReadSet][]byte {
//	// key format: [PREFIX][ID]_[field]_[version]
//	key := fmt.Sprintf("%s%s^*^", CONTRACT_STATE_PREFIX, hexutil.Encode(id))
//	if Dbconn == nil {
//		Dbconn = ReNewDbConn(config.DbPath)
//	}
//	data := getprefix(Dbconn, []byte(key))
//	if data == nil || len(data) <= 0 {
//		return nil
//	}
//	allState := map[ContractReadSet][]byte{}
//	for k, v := range data {
//		sKey := strings.Split(k, "^*^")
//		if len(sKey) != 3 {
//			continue
//		}
//		var version StateVersion
//		if !version.ParseStringKey(key) {
//			continue
//		}
//		rdSet := ContractReadSet{
//			Key:   sKey[1],
//			Value: &version,
//		}
//		allState[rdSet] = v
//	}
//	return allState
//}
//
///**
//获取合约（或模板）某一个属性
//To get contract or contract template one field
//*/
//func GetTplState(id []byte, field string) (StateVersion, []byte) {
//	key := fmt.Sprintf("%s%s^*^%s^*^", CONTRACT_TPL, hexutil.Encode(id[:]), field)
//	if Dbconn == nil {
//		Dbconn = ReNewDbConn(config.DbPath)
//	}
//	data := getprefix(Dbconn, []byte(key))
//	if data == nil || len(data) != 1 {
//		return StateVersion{}, nil
//	}
//	for k, v := range data {
//		var version StateVersion
//		if !version.ParseStringKey(k) {
//			return StateVersion{}, nil
//		}
//		return version, v
//	}
//	return StateVersion{}, nil
//}
//
///**
//获取合约（或模板）某一个属性
//To get contract or contract template one field
//*/
//func GetContractState(id string, field string) (StateVersion, []byte) {
//	key := fmt.Sprintf("%s%s^*^%s^*^", CONTRACT_STATE_PREFIX, id, field)
//	if Dbconn == nil {
//		Dbconn = ReNewDbConn(config.DbPath)
//	}
//	data := getprefix(Dbconn, []byte(key))
//	if data == nil || len(data) != 1 {
//		return StateVersion{}, nil
//	}
//	for k, v := range data {
//		var version StateVersion
//		if !version.ParseStringKey(k) {
//			return StateVersion{}, nil
//		}
//		return version, v
//	}
//	log.Println("11111111")
//	return StateVersion{}, nil
//}
