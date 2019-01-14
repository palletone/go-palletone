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
	"encoding/json"
	"fmt"
	"math/big"
	"unsafe"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/common/rlp"
	// "github.com/palletone/go-palletone/common/util"
	"github.com/palletone/go-palletone/dag/constants"
	"github.com/palletone/go-palletone/dag/errors"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/tokenengine"
)

//对DAG对象的操作，包括：Unit，Tx等
type DagDb struct {
	db ptndb.Database
	//logger log.ILogger
}

func NewDagDb(db ptndb.Database) *DagDb {
	return &DagDb{db: db}
}

type IDagDb interface {
	GetGenesisUnitHash() (common.Hash, error)
	SaveGenesisUnitHash(hash common.Hash) error

	SaveHeader(h *modules.Header) error
	SaveTransaction(tx *modules.Transaction) error
	SaveBody(unitHash common.Hash, txsHash []common.Hash) error
	GetBody(unitHash common.Hash) ([]common.Hash, error)
	SaveTransactions(txs *modules.Transactions) error
	//SaveNumberByHash(uHash common.Hash, number modules.ChainIndex) error
	//SaveHashByNumber(uHash common.Hash, number modules.ChainIndex) error
	SaveTxLookupEntry(unit *modules.Unit) error
	//SaveTokenInfo(token_info *modules.TokenInfo) (*modules.TokenInfo, error)
	//SaveAllTokenInfo(token_itmes *modules.AllTokenInfo) error

	PutCanonicalHash(hash common.Hash, number uint64) error
	PutHeadHeaderHash(hash common.Hash) error
	PutHeadUnitHash(hash common.Hash) error
	PutHeadFastUnitHash(hash common.Hash) error
	PutTrieSyncProgress(count uint64) error
	UpdateHeadByBatch(hash common.Hash, number uint64) error

	//GetUnit(hash common.Hash) (*modules.Unit, error)
	GetUnitTransactions(hash common.Hash) (modules.Transactions, error)
	GetTransaction(hash common.Hash) (*modules.Transaction, common.Hash, uint64, uint64)
	GetTxLookupEntry(hash common.Hash) (common.Hash, uint64, uint64, error)
	GetPrefix(prefix []byte) map[string][]byte
	GetHeader(hash common.Hash) (*modules.Header, error)
	IsHeaderExist(uHash common.Hash) (bool, error)
	//GetUnitFormIndex(number modules.ChainIndex) (*modules.Unit, error)
	//GetHeaderByNumber(index *modules.ChainIndex) (*modules.Header, error)
	//GetNumberWithUnitHash(hash common.Hash) (*modules.ChainIndex, error)
	GetHashByNumber(number *modules.ChainIndex) (common.Hash, error)
	//GetHeaderRlp(hash common.Hash, index uint64) rlp.RawValue
	GetCanonicalHash(number uint64) (common.Hash, error)
	GetAddrOutput(addr string) ([]modules.Output, error)

	GetHeadHeaderHash() (common.Hash, error)
	GetHeadUnitHash() (common.Hash, error)
	GetHeadFastUnitHash() (common.Hash, error)
	//GetAllLeafNodes() ([]*modules.Header, error)
	GetTrieSyncProgress() (uint64, error)
	//GetLastIrreversibleUnit(assetID modules.IDType16) (*modules.Unit, error)
	//GetTokenInfo(key string) (*modules.TokenInfo, error)
	//GetAllTokenInfo() (*modules.AllTokenInfo, error)

	// common geter
	GetCommon(key []byte) ([]byte, error)
	GetCommonByPrefix(prefix []byte) map[string][]byte

	// get txhash  and save index
	GetReqIdByTxHash(hash common.Hash) (common.Hash, error)
	GetTxHashByReqId(reqid common.Hash) (common.Hash, error)
	SaveReqIdByTx(tx *modules.Transaction) error
	GetTxFromAddress(tx *modules.Transaction) ([]string, error)
}

func (dagdb *DagDb) IsHeaderExist(uHash common.Hash) (bool, error) {
	key := append(constants.HEADER_PREFIX, uHash.Bytes()...)
	return dagdb.db.Has(key)
}

/* ----- common geter ----- */
func (dagdb *DagDb) GetCommon(key []byte) ([]byte, error) {
	return dagdb.db.Get(key)
}
func (dagdb *DagDb) GetCommonByPrefix(prefix []byte) map[string][]byte {
	iter := dagdb.db.NewIteratorWithPrefix(prefix)
	result := make(map[string][]byte, 0)
	for iter.Next() {
		key := iter.Key()
		value := make([]byte, 0)
		// 请注意： 直接赋值取得iter.Value()的最后一个指针
		// result[*(*string)(unsafe.Pointer(&key))] = append(value, iter.Value()...)
		result[string(key)] = append(value, iter.Value()...)
	}
	for k, val := range result {
		fmt.Println("key:::::::::  ", k, string(val))
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
	log.Debugf("Save GenesisUnitHash:%x", hash.Bytes())
	return dagdb.db.Put(constants.GenesisUnitHash, hash.Bytes())

}

// ###################### SAVE IMPL START ######################
/**
key: [HEADER_PREFIX][chain index number]_[chain index]_[unit hash]
value: unit header rlp encoding bytes
*/
// save header
func (dagdb *DagDb) SaveHeader(h *modules.Header) error {
	uHash := h.Hash()
	key := append(constants.HEADER_PREFIX, uHash.Bytes()...)
	err := StoreBytes(dagdb.db, (key), h)
	if err != nil {
		log.Error("Save Header error", err.Error())
		return err
	}
	log.Debugf("Save header for unit: %x", uHash.Bytes())
	return dagdb.saveHeaderHeightIndex(h)
}

//为Unit的Height建立索引,这个索引是必须的，所以在dagdb中实现，而不是在indexdb实现。
func (dagdb *DagDb) saveHeaderHeightIndex(h *modules.Header) error {
	idxKey := append(constants.HEADER_HEIGTH_PREFIX, h.Number.Bytes()...)
	uHash := h.Hash()
	err := StoreBytes(dagdb.db, idxKey, uHash)
	if err != nil {
		log.Error("Save Header height index error", err.Error())
		return err
	}
	log.Debugf("Save header number %s for unit: %x", h.Number.String(), uHash.Bytes())
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

//
////這是通過modules.ChainIndex存儲hash
//func (dagdb *DagDb) SaveNumberByHash(uHash common.Hash, number modules.ChainIndex) error {
//	if number == (modules.ChainIndex{}) {
//		return errors.New("the saving chain_index is null.")
//	}
//	key := fmt.Sprintf("%s%s", constants.UNIT_HASH_NUMBER_Prefix, uHash.String())
//	index := new(modules.ChainIndex)
//	index.AssetID = number.AssetID
//	index.Index = number.Index
//	index.IsMain = number.IsMain
//	if _, err := GetBytes(dagdb.db, []byte(key)); err == nil {
//		if !index.IsMain { // 若index不在主链，则不更新。
//			return nil
//		}
//	}
//	return StoreBytes(dagdb.db, []byte(key), index)
//}
//
////這是通過hash存儲modules.ChainIndex
//func (dagdb *DagDb) SaveHashByNumber(uHash common.Hash, number modules.ChainIndex) error {
//	i := 0
//	if number.IsMain {
//		i = 1
//	}
//	key := fmt.Sprintf("%s_%s_%d_%d", constants.UNIT_NUMBER_PREFIX, number.AssetID.String(), i, number.Index)
//	if m_data, err := GetBytes(dagdb.db, *(*[]byte)(unsafe.Pointer(&key))); err == nil {
//		// step1. 若uHash.String()==str 则无需再次存储。
//		var str string
//		err1 := rlp.DecodeBytes(m_data, &str)
//		if err1 == nil {
//			if str == uHash.String() {
//				return nil // 无需重复存储
//			}
//		}
//		// step2. 若不相等，则更新
//		// 确定前一个为主链单元，保存该number到侧链上。
//		i = 0
//	}
//	key = fmt.Sprintf("%s_%s_%d_%d", constants.UNIT_NUMBER_PREFIX, number.AssetID.String(), i, number.Index)
//	//log.Info("*****************DagDB SaveHashByNumber info.", "SaveHashByNumber_key", string(key), "hash:", uHash.Hex())
//	return StoreBytes(dagdb.db, *(*[]byte)(unsafe.Pointer(&key)), uHash.Hex())
//}

//func (dagdb *DagDb) UpdateParentChainIndexByHash(hash common.Hash, index modules.ChainIndex) error {
//	index.IsMain = true
//	err := dagdb.SaveNumberByHash(hash, index)
//	if err != nil {
//		return fmt.Errorf("update parent chainindex by hash error: %s", err.Error())
//	}
//	err1 := dagdb.SaveHashByNumber(hash, index)
//	if err1 != nil {
//		return fmt.Errorf("update parent hash by chainindex error: %s", err1)
//	}
//	return nil
//}

//	i := 0
//	if number.IsMain {
//		i = 1
//	}
//	key := fmt.Sprintf("%s_%s_%d_%d", constants.UNIT_NUMBER_PREFIX, number.AssetID.String(), i, number.Index)
//	ha, err := GetBytes(dagdb.db, *(*[]byte)(unsafe.Pointer(&key)))
//	log.Debug("DagDB GetHashByNumber info.", "error", err, "GetHashByNumber_key", string(key), "hash:", fmt.Sprintf("%x", ha))
//	if err != nil {
//		return common.Hash{}, err
//	}
//
//	strhash := ""
//	err1 := rlp.DecodeBytes(ha, &strhash)
//	if err1 != nil {
//		log.Debug("GetHashByNumber", "DecodeBytes_err", err1)
//		return common.Hash{}, err1
//	}
//	hash := common.Hash{}
//	if err := hash.SetHexString(strhash); err != nil {
//		log.Debug("GetHashByNumber", "SetHexString err:", err)
//		return common.Hash{}, err
//	}
//
//	return hash, nil
//}
//
//// height and assetid can get a unit key.
//func (dagdb *DagDb) SaveUHashIndex(cIndex modules.ChainIndex, uHash common.Hash) error {
//	key := fmt.Sprintf("%s_%s_%d", constants.UNIT_NUMBER_PREFIX, cIndex.AssetID.String(), cIndex.Index)
//	return Store(dagdb.db, key, uHash.Hex())
//}

/**
key: [BODY_PREFIX][unit hash]
value: all transactions hash set's rlp encoding bytes
*/
func (dagdb *DagDb) SaveBody(unitHash common.Hash, txsHash []common.Hash) error {
	// db.Put(append())
	log.Debugf("Save body of unit[%s], include txs:%x", unitHash.String(), txsHash)
	return StoreBytes(dagdb.db, append(constants.BODY_PREFIX, unitHash.Bytes()...), txsHash)
}

func (dagdb *DagDb) GetBody(unitHash common.Hash) ([]common.Hash, error) {
	log.Debug("get unit body info", "unitHash", unitHash.String())
	data, err := dagdb.db.Get(append(constants.BODY_PREFIX, unitHash.Bytes()...))
	if err != nil {
		return nil, err
	}

	var txHashs []common.Hash
	if err := rlp.DecodeBytes(data, &txHashs); err != nil {
		return nil, err
	}
	return txHashs, nil
}

func (dagdb *DagDb) SaveTransactions(txs *modules.Transactions) error {
	key := fmt.Sprintf("%s%s", constants.TRANSACTIONS_PREFIX, txs.Hash())
	return Store(dagdb.db, key, *txs)
}

/**
key: [TRANSACTION_PREFIX][tx hash]
value: transaction struct rlp encoding bytes
*/
func (dagdb *DagDb) SaveTransaction(tx *modules.Transaction) error {
	// save transaction
	bytes, err := json.Marshal(tx)
	if err != nil {
		return err
	}
	txHash := tx.Hash()
	str := *(*string)(unsafe.Pointer(&bytes))
	key0 := string(constants.TRANSACTION_PREFIX) + txHash.String()
	if err := StoreString(dagdb.db, key0, str); err != nil {
		return err
	}
	key1 := string(constants.Transaction_Index) + txHash.String()
	if err := StoreString(dagdb.db, key1, str); err != nil {
		return err
	}
	dagdb.updateAddrTransactions(tx, txHash)
	// store output by addr
	for i, msg := range tx.TxMessages {
		if msg.App >= modules.APP_CONTRACT_TPL_REQUEST && msg.App <= modules.APP_CONTRACT_STOP_REQUEST {
			if err := dagdb.SaveReqIdByTx(tx); err != nil {
				log.Error("SaveReqIdByTx is failed,", "error", err)
			}
			continue
		}
		payload, ok := msg.Payload.(*modules.PaymentPayload)
		if ok {
			for _, output := range payload.Outputs {
				//  pkscript to addr
				addr, err := tokenengine.GetAddressFromScript(output.PkScript[:])
				if err != nil {
					log.Error("GetAddressFromScript is failed,", "error", err)
				}
				dagdb.saveOutputByAddr(addr.String(), txHash, i, output)
			}
		}
	}
	return nil
}

func (dagdb *DagDb) saveOutputByAddr(addr string, hash common.Hash, msgindex int, output *modules.Output) error {
	if hash == (common.Hash{}) {
		return errors.New("empty tx hash.")
	}
	key := append(constants.AddrOutput_Prefix, []byte(addr)...)
	key = append(key, []byte(hash.String())...)
	err := StoreBytes(dagdb.db, append(key, new(big.Int).SetInt64(int64(msgindex)).Bytes()...), output)
	return err
}

func (dagdb *DagDb) updateAddrTransactions(tx *modules.Transaction, hash common.Hash) error {

	if hash == (common.Hash{}) {
		return errors.New("empty tx hash.")
	}
	froms, err := dagdb.GetTxFromAddress(tx)
	if err != nil {
		return err
	}
	// 1. save from_address
	for _, addr := range froms {
		go dagdb.saveAddrTxHashByKey(constants.AddrTx_From_Prefix, addr, hash)
	}
	// 2. to_address 已经在上层接口处理了。
	// for _, addr := range tos { // constants.AddrTx_To_Prefix
	// 	go dagdb.saveAddrTxHashByKey(constants.AddrTx_To_Prefix, addr, hash)
	// }
	return nil
}
func (dagdb *DagDb) saveAddrTxHashByKey(key []byte, addr string, hash common.Hash) error {

	hashs := make([]common.Hash, 0)
	data, err := dagdb.db.Get(append(key, []byte(addr)...))
	if err != nil {
		if err.Error() != "leveldb: not found" {
			return err
		} else { // first store the addr
			hashs = append(hashs, hash)
			if err := StoreBytes(dagdb.db, append(key, []byte(addr)...), hashs); err != nil {
				return err
			}
			return nil
		}
	}
	if err := rlp.DecodeBytes(data, &hashs); err != nil {
		return err
	}
	hashs = append(hashs, hash)
	if err := StoreBytes(dagdb.db, append(key, []byte(addr)...), hashs); err != nil {
		return err
	}
	return nil
}

func (dagdb *DagDb) SaveTxLookupEntry(unit *modules.Unit) error {
	log.Debugf("Save tx lookup entry, tx count:%d", len(unit.Txs))
	for i, tx := range unit.Transactions() {
		in := &modules.TxLookupEntry{
			UnitHash:  unit.Hash(),
			UnitIndex: unit.NumberU64(),
			Index:     uint64(i),
		}

		if err := StoreBytes(dagdb.db, append(constants.LookupPrefix, tx.Hash().Bytes()...), in); err != nil {
			return err
		}
	}
	return nil
}

//func (dagdb *DagDb) SaveTokenInfo(token_info *modules.TokenInfo) (*modules.TokenInfo, error) {
//	if token_info == nil {
//		return token_info, errors.New("token info is null.")
//	}
//	// id, _ := modules.SetIdTypeByHex(token_info.TokenHex)
//
//	key := string(constants.TOKENTYPE) + token_info.TokenHex
//	log.Info("================save token info =========", "key", key)
//	if err := StoreBytes(dagdb.db, *(*[]byte)(unsafe.Pointer(&key)), token_info); err != nil {
//		return token_info, err
//	}
//	// 更新all token_info table.
//	infos, _ := dagdb.GetAllTokenInfo()
//	if infos == nil {
//		infos = new(modules.AllTokenInfo)
//	}
//
//	infos.Add(token_info)
//	dagdb.SaveAllTokenInfo(infos)
//	return token_info, nil
//}
//
//func (dagdb *DagDb) SaveAllTokenInfo(token_itmes *modules.AllTokenInfo) error {
//	if err := StoreString(dagdb.db, string(constants.TOKENINFOS), token_itmes.String()); err != nil {
//		return err
//	}
//	return nil
//}

// ###################### SAVE IMPL END ######################
// ###################### GET IMPL START ######################

// Get income transactions
func (dagdb *DagDb) GetAddrOutput(addr string) ([]modules.Output, error) {

	data := dagdb.GetPrefix(append(constants.AddrOutput_Prefix, []byte(addr)...))
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

func (dagdb *DagDb) GetTxFromAddress(tx *modules.Transaction) ([]string, error) {

	froms := make([]string, 0)
	if tx == nil {
		return froms, errors.New("tx is nil, not exist address.")
	}
	outpoints, _ := tx.GetAddressInfo()
	for _, op := range outpoints {
		addr, err := dagdb.getOutpointAddr(op)
		if err == nil {
			froms = append(froms, addr)
		} else {
			log.Info("get out address is failed.", "error", err)
		}
	}

	return froms, nil
}
func (dagdb *DagDb) getOutpointAddr(outpoint *modules.OutPoint) (string, error) {
	if outpoint == nil {
		return "", fmt.Errorf("outpoint_key is nil ")
	}
	out_key := append(constants.OutPointAddr_Prefix, outpoint.ToKey()...)
	data, err := dagdb.db.Get(out_key[:])
	if len(data) <= 0 {
		return "", fmt.Errorf("address is null. outpoint_key(%s)", outpoint.ToKey())
	}
	if err != nil {
		return "", err
	}
	var str string
	err0 := rlp.DecodeBytes(data, &str)
	return str, err0
}

//func (dagdb *DagDb) GetNumberWithUnitHash(hash common.Hash) (*modules.ChainIndex, error) {
//	key := fmt.Sprintf("%s%s", constants.UNIT_HASH_NUMBER_Prefix, hash.String())
//
//	data, err := dagdb.db.Get([]byte(key))
//	if err != nil {
//		return nil, err
//	}
//	if len(data) <= 0 {
//		return nil, fmt.Errorf("chainIndex is null. hash(%s)", hash.String())
//	}
//	number := new(modules.ChainIndex)
//	if err := rlp.DecodeBytes(data, number); err != nil {
//		return nil, fmt.Errorf("Get unit number when rlp decode error:%s", err.Error())
//	}
//
//	return number, nil
//}

//  GetCanonicalHash get

func (dagdb *DagDb) GetCanonicalHash(number uint64) (common.Hash, error) {
	key := append(constants.HEADER_PREFIX, encodeBlockNumber(number)...)
	data, err := dagdb.db.Get(append(key, constants.NumberSuffix...))
	if err != nil {
		return common.Hash{}, err
	}
	if len(data) == 0 {
		return common.Hash{}, err
	}
	return common.BytesToHash(data), nil
}
func (dagdb *DagDb) GetHeadHeaderHash() (common.Hash, error) {
	data, err := dagdb.db.Get(constants.HeadHeaderKey)
	if err != nil {
		return common.Hash{}, err
	}
	if len(data) != 8 {
		return common.Hash{}, errors.New("data's len is error.")
	}
	return common.BytesToHash(data), nil
}

// GetHeadUnitHash stores the head unit's hash.
func (dagdb *DagDb) GetHeadUnitHash() (common.Hash, error) {
	data, err := dagdb.db.Get(constants.HeadUnitHash)
	if err != nil {
		return common.Hash{}, err
	}
	return common.BytesToHash(data), nil
}

// GetHeadFastUnitHash stores the fast head unit's hash.
func (dagdb *DagDb) GetHeadFastUnitHash() (common.Hash, error) {
	data, err := dagdb.db.Get(constants.HeadFastKey)
	if err != nil {
		return common.Hash{}, err
	}
	return common.BytesToHash(data), nil
}

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
		log.Info("GetUnitTransactions when get body error", "error", err.Error(), "unit_hash", hash.String())
		return nil, err
	}
	// get transaction by tx'hash.
	for _, txHash := range txHashList {
		tx, _, _, _ := dagdb.GetTransaction(txHash)
		if tx != nil {
			txs = append(txs, tx)
		}
	}
	return txs, nil
}

//func (dagdb *DagDb) GetUnitFormIndex(number modules.ChainIndex) (*modules.Unit, error) {
//	i := 0
//	if number.IsMain {
//		i = 1
//	}
//	key := fmt.Sprintf("%s_%s_%d_%d", constants.UNIT_NUMBER_PREFIX, number.AssetID.String(), i, number.Index)
//	hash, err := dagdb.db.Get([]byte(key))
//	if err != nil {
//		return nil, err
//	}
//	var hex string
//	rlp.DecodeBytes(hash, &hex)
//	h := common.HexToHash(hex)
//
//	return dagdb.GetUnit(h)
//}
//
//func (dagdb *DagDb) GetLastIrreversibleUnit(assetID modules.IDType16) (*modules.Unit, error) {
//	key := fmt.Sprintf("%s_%s_1_", constants.UNIT_NUMBER_PREFIX, assetID.String())
//
//	data := dagdb.GetPrefix([]byte(key))
//	var irreKey string
//	var irreIndex []string
//	for k := range data {
//		// get the key of max index
//		if sts := strings.Split(k, key); len(sts) == 2 {
//			irreIndex = append(irreIndex, sts[1])
//		}
//	}
//	var max int64
//	for i, v := range irreIndex {
//		if index, err := strconv.ParseInt(v, 10, 64); err == nil {
//			if i == 0 {
//				max = index
//			} else {
//				if max < index {
//					max = index
//				}
//			}
//
//		}
//	}
//	irreKey = fmt.Sprintf(key+"%d", max)
//	rlpUnitHash := data[irreKey]
//	log.Info("============== GetLastIrreversibleUnit max index key is ===================== ", "irreKey", irreKey, "hash", rlpUnitHash)
//	if len(rlpUnitHash) > 0 {
//		var hex string
//		err := rlp.DecodeBytes(rlpUnitHash, &hex)
//		if err != nil {
//			log.Error("GetLastIrreversibleUnit error:" + err.Error())
//			return nil, err
//		}
//		unitHash := common.HexToHash(hex)
//		return dagdb.GetUnit(unitHash)
//	}
//	return nil, errors.New(fmt.Sprintf("the irrekey :%s ,is not found unit's hash.", irreKey))
//}

func (dagdb *DagDb) GetHeader(hash common.Hash) (*modules.Header, error) {
	key := append(constants.HEADER_PREFIX, hash.Bytes()...)
	header := new(modules.Header)
	err := retrieve(dagdb.db, key, header)
	if err != nil {
		return nil, err
	}
	return header, nil

}

// GetHeaderByNumber ,first :get hash  , return header.
//func (dagdb *DagDb) GetHeaderByNumber(index *modules.ChainIndex) (*modules.Header, error) {
//	hash, err := dagdb.GetHashByNumber(index)
//	if err != nil {
//		return nil, err
//	}
//	return dagdb.GetHeader(hash)
//}

//func (dagdb *DagDb) GetHeaderRlp(hash common.Hash, index uint64) rlp.RawValue {
//	encNum := encodeBlockNumber(index)
//	key := append(constants.HEADER_PREFIX, encNum...)
//	header_bytes, err := dagdb.db.Get(append(key, hash.Bytes()...))
//	// rlp  to  Header struct
//	if err != nil {
//		log.Error("GetHeaderRlp error", "error", err)
//	}
//	return header_bytes
//}

//func (dagdb *DagDb) GetHeaderFormIndex(number modules.ChainIndex) *modules.Header {
//	unit, err := dagdb.GetUnitFormIndex(number)
//	if err != nil {
//		return unit.UnitHeader
//	}
//	return nil
//}

// GetTxLookupEntry return unit's hash ,number
func (dagdb *DagDb) GetTxLookupEntry(hash common.Hash) (common.Hash, uint64, uint64, error) {
	data, err0 := dagdb.db.Get(append(constants.LookupPrefix, hash.Bytes()...))
	if len(data) == 0 {
		return common.Hash{}, 0, 0, errors.New("not found legal data.")
	}
	if err0 != nil {
		return common.Hash{}, 0, 0, err0
	}
	var entry *modules.TxLookupEntry
	if err := rlp.DecodeBytes(data, &entry); err != nil {
		log.Info("get entry structure failed ===================", "error", err, "tx_entry", entry)
	}

	return entry.UnitHash, entry.UnitIndex, entry.Index, err0
}

// GetTransaction retrieves a specific transaction from the database , along with its added positional metadata
// p2p 同步区块 分为同步header 和body。 GetBody可以省掉节点包装交易块的过程。
func (dagdb *DagDb) GetTransaction(hash common.Hash) (*modules.Transaction, common.Hash, uint64, uint64) {
	unitHash, unitNumber, txIndex, err1 := dagdb.GetTxLookupEntry(hash)
	if err1 != nil {
		log.Error("dag db GetTransaction", "GetTxLookupEntry err:", err1, "hash:", hash)
		return nil, unitHash, unitNumber, txIndex
	}
	// if unitHash != (common.Hash{}) {
	// 	body, _ := dagdb.GetBody(unitHash)
	// 	if body == nil || len(body) <= int(txIndex) {
	// 		return nil, common.Hash{}, 0, 0
	// 	}
	// 	tx, err := dagdb.gettrasaction(body[txIndex])
	// 	if err == nil {
	// 		return tx, unitHash, unitNumber, txIndex
	// 	}
	// }
	tx, err := dagdb.gettrasaction(hash)
	if err != nil {
		fmt.Println("gettrasaction error:", err.Error())
		return nil, unitHash, unitNumber, txIndex
	}

	return tx, unitHash, unitNumber, txIndex
}

// gettrasaction can get a transaction by hash.
func (dagdb *DagDb) gettrasaction(hash common.Hash) (*modules.Transaction, error) {
	if hash == (common.Hash{}) {
		return nil, errors.New("hash is not exist.")
	}
	key := string(constants.TRANSACTION_PREFIX) + hash.String()

	data, err := getString(dagdb.db, []byte(key))
	if err != nil {
		log.Error("get transaction failed......", "error", err)
		return nil, err
	}
	tx := new(modules.Transaction)
	if err := json.Unmarshal([]byte(data), &tx); err != nil {
		log.Error("tx Unmarshal failed......", "error", err, "data:", data)
		return nil, err
	}
	// TODO ---- 将不同msg‘s app 反序列化后赋值给payload interface{}.
	//log.Debug("================== transaction_info======================", "error", err, "transaction_info", tx)
	//msgs, err1 := ConvertMsg(tx)
	//if err1 != nil {
	//	log.Error("tx comvertmsg failed......", "err:", err1, "tx:", tx)
	//	return nil, err1
	//}
	//
	//tx.TxMessages = msgs
	return tx, err
}

//func ConvertMsg(tx *modules.Transaction) ([]*modules.Message, error) {
//	if tx == nil {
//		return nil, errors.New("convert msg tx is nil")
//	}
//	msgs := make([]*modules.Message, 0)
//	for _, msg := range tx.Messages() {
//		//fmt.Println("msg ", msg)
//
//		data1, err1 := json.Marshal(msg.Payload)
//		if err1 != nil {
//			return nil, err1
//		}
//		switch msg.App {
//		default:
//			//case APP_PAYMENT, APP_CONTRACT_TPL, APP_DATA, APP_VOTE:
//			// payment := new(modules.PaymentPayload)
//			// err2 := json.Unmarshal(data1, &payment)
//			// if err2 != nil {
//			// 	return nil, err2
//			// }
//			// msg.Payload = payment
//			msgs = append(msgs, msg)
//
//		case modules.APP_PAYMENT: //0
//			payment := new(modules.PaymentPayload)
//			err2 := json.Unmarshal(data1, &payment)
//			if err2 != nil {
//				return nil, err2
//			}
//			msg.Payload = payment
//			msgs = append(msgs, msg)
//		case modules.APP_CONTRACT_TPL: //1
//			payment := new(modules.ContractTplPayload)
//			err2 := json.Unmarshal(data1, &payment)
//			if err2 != nil {
//				return nil, err2
//			}
//			msg.Payload = payment
//			msgs = append(msgs, msg)
//
//		case modules.APP_CONTRACT_DEPLOY: //2
//			payment := new(modules.ContractDeployPayload)
//			err2 := json.Unmarshal(data1, &payment)
//			if err2 != nil {
//				return nil, err2
//			}
//			msg.Payload = payment
//			msgs = append(msgs, msg)
//
//		case modules.APP_CONTRACT_INVOKE: //3
//			payment := new(modules.ContractInvokePayload)
//			err2 := json.Unmarshal(data1, &payment)
//			if err2 != nil {
//				return nil, err2
//			}
//			// decode WriteSet interface
//			//for i, cw := range payment.WriteSet {
//			//	fmt.Println("lal========================ala", i, cw.Value)
//			//	//fmt.Printf("lalalalala%#v\n\n",cw.Value)
//			//	val_byte, _ := json.Marshal(cw.Value)
//			//	var item []byte
//			//	json.Unmarshal(val_byte, &item)
//			//	fmt.Println("===========", item)
//			//	payment.WriteSet[i].Value = item
//			//
//			//}
//
//			msg.Payload = payment
//			msgs = append(msgs, msg)
//		case modules.APP_CONTRACT_INVOKE_REQUEST: //4
//			payment := new(modules.ContractInvokeRequestPayload)
//			err2 := json.Unmarshal(data1, &payment)
//			if err2 != nil {
//				return nil, err2
//			}
//			msg.Payload = payment
//			msgs = append(msgs, msg)
//		case modules.APP_CONFIG: //5
//			payment := new(modules.ConfigPayload)
//			err2 := json.Unmarshal(data1, &payment)
//			if err2 != nil {
//				return nil, err2
//			}
//			msg.Payload = payment
//			msgs = append(msgs, msg)
//		case modules.APP_DATA: //6
//			payment := new(modules.DataPayload)
//			err2 := json.Unmarshal(data1, &payment)
//			if err2 != nil {
//				return nil, err2
//			}
//			msg.Payload = payment
//			msgs = append(msgs, msg)
//		case modules.APP_VOTE: //7
//			payment := new(vote.VoteInfo)
//			err2 := json.Unmarshal(data1, &payment)
//			if err2 != nil {
//				return nil, err2
//			}
//			msg.Payload = payment
//			msgs = append(msgs, msg)
//		case modules.APP_SIGNATURE: //8
//			payment := new(modules.SignaturePayload)
//			err2 := json.Unmarshal(data1, &payment)
//			if err2 != nil {
//				return nil, err2
//			}
//			msg.Payload = payment
//			msgs = append(msgs, msg)
//
//		case modules.OP_MEDIATOR_CREATE:
//			payment := new(modules.MediatorCreateOperation)
//			err2 := json.Unmarshal(data1, &payment)
//			if err2 != nil {
//				return nil, err2
//			}
//			msg.Payload = payment
//			msgs = append(msgs, msg)
//
//		}
//	}
//	return msgs, nil
//}

func (dagdb *DagDb) GetContractNoReader(db ptndb.Database, id common.Hash) (*modules.Contract, error) {
	if common.EmptyHash(id) {
		return nil, errors.New("the filed not defined")
	}
	con_bytes, err := dagdb.db.Get(append(constants.CONTRACT_PREFIX, id[:]...))
	if err != nil {
		log.Error(fmt.Sprintf("getContract error: %s", err.Error()))
		return nil, err
	}
	contract := new(modules.Contract)
	err = rlp.DecodeBytes(con_bytes, contract)
	if err != nil {
		log.Error("getContract failed,decode error:" + err.Error())
		return nil, err
	}
	return contract, nil
}

//batch put HeaderCanon & HeaderKey & HeadUnitHash & HeadFastKey
func (dagdb *DagDb) UpdateHeadByBatch(hash common.Hash, number uint64) error {
	batch := dagdb.db.NewBatch()
	errorList := &[]error{}

	key := append(constants.HeaderCanon_Prefix, encodeBlockNumber(number)...)
	BatchErrorHandler(batch.Put(append(key, constants.NumberSuffix...), hash.Bytes()), errorList) //PutCanonicalHash
	BatchErrorHandler(batch.Put(constants.HeadHeaderKey, hash.Bytes()), errorList)                //PutHeadHeaderHash
	BatchErrorHandler(batch.Put(constants.HeadUnitHash, hash.Bytes()), errorList)                 //PutHeadUnitHash
	BatchErrorHandler(batch.Put(constants.HeadFastKey, hash.Bytes()), errorList)                  //PutHeadFastUnitHash
	if len(*errorList) == 0 {                                                                     //each function call succeed.
		return batch.Write()
	}
	return fmt.Errorf("UpdateHeadByBatch, at least one sub function call failed.")
}

func (dagdb *DagDb) PutCanonicalHash(hash common.Hash, number uint64) error {
	key := append(constants.HeaderCanon_Prefix, encodeBlockNumber(number)...)
	if err := dagdb.db.Put(append(key, constants.NumberSuffix...), hash.Bytes()); err != nil {
		return err
	}
	return nil
}
func (dagdb *DagDb) PutHeadHeaderHash(hash common.Hash) error {
	if err := dagdb.db.Put(constants.HeadHeaderKey, hash.Bytes()); err != nil {
		return err
	}
	return nil
}

// PutHeadUnitHash stores the head unit's hash.
func (dagdb *DagDb) PutHeadUnitHash(hash common.Hash) error {
	if err := dagdb.db.Put(constants.HeadUnitHash, hash.Bytes()); err != nil {
		return err
	}
	return nil
}

// PutHeadFastUnitHash stores the fast head unit's hash.
func (dagdb *DagDb) PutHeadFastUnitHash(hash common.Hash) error {
	if err := dagdb.db.Put(constants.HeadFastKey, hash.Bytes()); err != nil {
		return err
	}
	return nil
}

// PutTrieSyncProgress stores the fast sync trie process counter to support
// retrieving it across restarts.
func (dagdb *DagDb) PutTrieSyncProgress(count uint64) error {
	if err := dagdb.db.Put(constants.TrieSyncKey, new(big.Int).SetUint64(count).Bytes()); err != nil {
		return err
	}
	return nil
}

//
//// GetTokenInfo
//func (dagdb *DagDb) GetAllTokenInfo() (*modules.AllTokenInfo, error) {
//	data, err := dagdb.db.Get(constants.TOKENINFOS)
//	if err != nil {
//		log.Info("123123123123", "all", data)
//		return nil, err
//	}
//	if all, err := modules.Jsonbytes2AllTokenInfo(data); err != nil {
//		log.Info("78787878", "all", all)
//		return nil, err
//	} else {
//		log.Info("56565656", "all", all)
//		return all, nil
//	}
//}
//func (dagdb *DagDb) GetTokenInfo(key string) (*modules.TokenInfo, error) {
//	log.Info("================get token info =========", "key", string(key))
//	key = *(*string)(unsafe.Pointer(&constants.TOKENTYPE)) + key
//	data, err := dagdb.db.Get([]byte(key))
//	if err != nil {
//		return nil, err
//	}
//
//	info := new(modules.TokenInfo)
//	if err := rlp.DecodeBytes(data, &info); err != nil {
//		return nil, err
//	}
//	return info, nil
//}
//
//func (dagdb *DagDb) GetAllLeafNodes() ([]*modules.Header, error) {
//	all_token, err := dagdb.GetAllTokenInfo()
//	if err != nil {
//		return nil, err
//	}
//	if all_token == nil {
//		return nil, errors.New("all token info is nil.")
//	}
//	if all_token.Items == nil {
//		return nil, errors.New("items's map is nil.")
//	}
//	headers := make([]*modules.Header, 0)
//	for _, tokenInfo := range all_token.Items {
//		unit, err := dagdb.GetLastIrreversibleUnit(tokenInfo.Token)
//		if err != nil {
//			log.Error("GetLastIrreversibleUnit failed,", "error", err)
//		}
//		if unit != nil {
//			headers = append(headers, unit.UnitHeader)
//		}
//	}
//	return headers, nil
//}

// ###################### GET IMPL END ######################

func (dagdb *DagDb) GetReqIdByTxHash(hash common.Hash) (common.Hash, error) {
	key := fmt.Sprintf("%s_%s", string(constants.TxHash2ReqPrefix), hash.String())
	str, err := GetString(dagdb.db, key)
	return common.HexToHash(str), err
}

func (dagdb *DagDb) GetTxHashByReqId(reqid common.Hash) (common.Hash, error) {
	key := fmt.Sprintf("%s_%s", string(constants.ReqIdPrefix), reqid.String())
	str, err := GetString(dagdb.db, key)
	return common.HexToHash(str), err
}
func (dagdb *DagDb) SaveReqIdByTx(tx *modules.Transaction) error {
	txhash := tx.Hash()
	reqid := tx.RequestHash()
	key1 := fmt.Sprintf("%s_%s", string(constants.ReqIdPrefix), reqid.String())
	key2 := fmt.Sprintf("%s_%s", string(constants.TxHash2ReqPrefix), txhash.String())
	err1 := StoreString(dagdb.db, key1, txhash.String())
	err2 := StoreString(dagdb.db, key2, reqid.String())
	if err1 != nil {
		return err1
	}
	return err2
}
