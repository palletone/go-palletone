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

package common

import (
	"fmt"
	"math/big"
	"reflect"
	"strings"
	"time"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/hexutil"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/core"

	"github.com/palletone/go-palletone/core/accounts/keystore"
	"github.com/palletone/go-palletone/dag/constants"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/errors"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/storage"

	"github.com/palletone/go-palletone/contracts/syscontract"
	"github.com/palletone/go-palletone/dag/txspool"
	"github.com/palletone/go-palletone/tokenengine"
	//"github.com/palletone/go-palletone/validator"
	"encoding/json"
	"sync"
)

type IUnitRepository interface {
	GetGenesisUnit() (*modules.Unit, error)
	//GenesisHeight() modules.ChainIndex
	SaveUnit(unit *modules.Unit, isGenesis bool) error
	CreateUnit(mAddr *common.Address, txpool txspool.ITxPool, t time.Time) (*modules.Unit, error)
	IsGenesis(hash common.Hash) bool
	GetAddrTransactions(addr common.Address) ([]*modules.TransactionWithUnitInfo, error)
	GetHeaderByHash(hash common.Hash) (*modules.Header, error)
	GetHeaderList(hash common.Hash, parentCount int) ([]*modules.Header, error)
	SaveHeader(header *modules.Header) error
	SaveHeaders(headers []*modules.Header) error
	GetHeaderByNumber(index *modules.ChainIndex) (*modules.Header, error)
	IsHeaderExist(uHash common.Hash) (bool, error)
	GetHashByNumber(number *modules.ChainIndex) (common.Hash, error)

	GetUnitTransactions(hash common.Hash) (modules.Transactions, error)
	GetUnit(hash common.Hash) (*modules.Unit, error)

	GetBody(unitHash common.Hash) ([]common.Hash, error)
	GetTransaction(hash common.Hash) (*modules.TransactionWithUnitInfo, error)
	GetTransactionOnly(hash common.Hash) (*modules.Transaction, error)
	GetTxLookupEntry(hash common.Hash) (*modules.TxLookupEntry, error)
	GetCommon(key []byte) ([]byte, error)
	GetCommonByPrefix(prefix []byte) map[string][]byte
	//GetReqIdByTxHash(hash common.Hash) (common.Hash, error)
	GetTxHashByReqId(reqid common.Hash) (common.Hash, error)
	//GetAddrOutput(addr string) ([]modules.Output, error)
	GetTrieSyncProgress() (uint64, error)
	//GetHeadHeaderHash() (common.Hash, error)
	//GetHeadUnitHash() (common.Hash, error)
	//GetHeadFastUnitHash() (common.Hash, error)
	GetNumberWithUnitHash(hash common.Hash) (*modules.ChainIndex, error)
	//GetCanonicalHash(number uint64) (common.Hash, error)
	GetAssetTxHistory(asset *modules.Asset) ([]*modules.TransactionWithUnitInfo, error)
	//SaveNumberByHash(uHash common.Hash, number modules.ChainIndex) error
	//SaveHashByNumber(uHash common.Hash, number modules.ChainIndex) error
	//UpdateHeadByBatch(hash common.Hash, number uint64) error

	//GetHeaderRlp(hash common.Hash, index uint64) rlp.RawValue
	GetFileInfo(filehash []byte) ([]*modules.FileInfo, error)

	//获得某个分区上的最新不可逆单元
	GetLastIrreversibleUnit(assetID modules.AssetId) (*modules.Unit, error)

	GetTxFromAddress(tx *modules.Transaction) ([]common.Address, error)
}
type UnitRepository struct {
	dagdb storage.IDagDb
	idxdb storage.IIndexDb
	//uxtodb         storage.IUtxoDb
	statedb storage.IStateDb
	propdb  storage.IPropertyDb
	//validate       validator.Validator
	utxoRepository IUtxoRepository
	lock           sync.RWMutex
}

func NewUnitRepository(dagdb storage.IDagDb, idxdb storage.IIndexDb, utxodb storage.IUtxoDb, statedb storage.IStateDb, propdb storage.IPropertyDb) *UnitRepository {
	utxoRep := NewUtxoRepository(utxodb, idxdb, statedb)
	//val := validator.NewValidate(dagdb, utxoRep, statedb)
	return &UnitRepository{dagdb: dagdb, idxdb: idxdb, statedb: statedb, utxoRepository: utxoRep, propdb: propdb}
}

func NewUnitRepository4Db(db ptndb.Database) *UnitRepository {
	dagdb := storage.NewDagDb(db)
	utxodb := storage.NewUtxoDb(db)
	statedb := storage.NewStateDb(db)
	idxdb := storage.NewIndexDb(db)
	propdb := storage.NewPropertyDb(db)
	utxoRep := NewUtxoRepository(utxodb, idxdb, statedb)
	//val := validator.NewValidate(dagdb, utxoRep, statedb)
	return &UnitRepository{dagdb: dagdb, idxdb: idxdb, statedb: statedb, propdb: propdb, utxoRepository: utxoRep}
}

func (rep *UnitRepository) GetHeaderByHash(hash common.Hash) (*modules.Header, error) {
	return rep.dagdb.GetHeaderByHash(hash)
}
func (rep *UnitRepository) GetHeaderList(hash common.Hash, parentCount int) ([]*modules.Header, error) {
	rep.lock.RLock()
	defer rep.lock.RUnlock()
	result := []*modules.Header{}
	uhash := hash
	for i := 0; i < parentCount; i++ {
		h, err := rep.GetHeaderByHash(uhash)
		if err != nil {
			return nil, err
		}
		result = append(result, h)
		if len(h.ParentsHash) == 0 { //Genesis unit
			break
		}
		uhash = h.ParentsHash[0]
	}
	return result, nil
}
func (rep *UnitRepository) SaveHeader(header *modules.Header) error {
	return rep.dagdb.SaveHeader(header)
}
func (rep *UnitRepository) SaveHeaders(headers []*modules.Header) error {
	return rep.dagdb.SaveHeaders(headers)
}
func (rep *UnitRepository) GetHeaderByNumber(index *modules.ChainIndex) (*modules.Header, error) {
	rep.lock.RLock()
	defer rep.lock.RUnlock()
	hash, err := rep.dagdb.GetHashByNumber(index)
	if err != nil {
		return nil, err
	}
	return rep.dagdb.GetHeaderByHash(hash)
}
func (rep *UnitRepository) IsHeaderExist(uHash common.Hash) (bool, error) {
	return rep.dagdb.IsHeaderExist(uHash)
}
func (rep *UnitRepository) GetUnit(hash common.Hash) (*modules.Unit, error) {
	rep.lock.RLock()
	defer rep.lock.RUnlock()
	return rep.getUnit(hash)
}
func (rep *UnitRepository) getUnit(hash common.Hash) (*modules.Unit, error) {
	uHeader, err := rep.dagdb.GetHeaderByHash(hash)
	if err != nil {
		log.Error("getChainUnit when GetHeaderByHash failed ", "error", err, "hash", hash.String())
		return nil, err
	}
	txs, err := rep.dagdb.GetUnitTransactions(hash)
	if err != nil {
		log.Error("getChainUnit when GetUnitTransactions failed ", "error", err, "hash", hash.String())
		return nil, err
	}
	// generate unit
	unit := &modules.Unit{
		UnitHeader: uHeader,
		UnitHash:   hash,
		Txs:        txs,
	}
	unit.UnitSize = unit.Size()
	return unit, nil
}

func (rep *UnitRepository) GetHashByNumber(number *modules.ChainIndex) (common.Hash, error) {
	return rep.dagdb.GetHashByNumber(number)
}
func (rep *UnitRepository) GetBody(unitHash common.Hash) ([]common.Hash, error) {
	return rep.dagdb.GetBody(unitHash)
}
func (rep *UnitRepository) GetTransaction(hash common.Hash) (*modules.TransactionWithUnitInfo, error) {
	tx, err := rep.dagdb.GetTransactionOnly(hash)
	if err != nil {
		return nil, err
	}
	txlookup, err1 := rep.dagdb.GetTxLookupEntry(hash)
	if err1 != nil {
		log.Info("dag db GetTransaction,GetTxLookupEntry failed.", "error", err1, "tx_hash:", hash)
		return nil, err1
	}
	resultTx := &modules.TransactionWithUnitInfo{Transaction: tx, UnitHash: txlookup.UnitHash, UnitIndex: txlookup.UnitIndex, TxIndex: txlookup.Index, Timestamp: txlookup.Timestamp}
	return resultTx, nil
}
func (rep *UnitRepository) GetTransactionOnly(hash common.Hash) (*modules.Transaction, error) {
	return rep.dagdb.GetTransactionOnly(hash)
}
func (rep *UnitRepository) GetTxLookupEntry(hash common.Hash) (*modules.TxLookupEntry, error) {
	return rep.dagdb.GetTxLookupEntry(hash)
}
func (rep *UnitRepository) GetCommon(key []byte) ([]byte, error) { return rep.dagdb.GetCommon(key) }
func (rep *UnitRepository) GetCommonByPrefix(prefix []byte) map[string][]byte {
	return rep.dagdb.GetCommonByPrefix(prefix)
}

//func (rep *UnitRepository) GetReqIdByTxHash(hash common.Hash) (common.Hash, error) {
//	return rep.dagdb.GetReqIdByTxHash(hash)
//}
func (rep *UnitRepository) GetTxHashByReqId(reqid common.Hash) (common.Hash, error) {
	return rep.dagdb.GetTxHashByReqId(reqid)
}

//func (rep *UnitRepository) GetAddrOutput(addr string) ([]modules.Output, error) {
//	return rep.dagdb.GetAddrOutput(addr)
//}
func (rep *UnitRepository) GetTrieSyncProgress() (uint64, error) {
	return rep.dagdb.GetTrieSyncProgress()
}

//func (rep *UnitRepository) GetHeadHeaderHash() (common.Hash, error) {
//	return rep.dagdb.GetHeadHeaderHash()
//}
//func (rep *UnitRepository) GetHeadUnitHash() (common.Hash, error) { return rep.dagdb.GetHeadUnitHash() }
//func (rep *UnitRepository) GetHeadFastUnitHash() (common.Hash, error) {
//	return rep.dagdb.GetHeadFastUnitHash()
//}
func (rep *UnitRepository) GetNumberWithUnitHash(hash common.Hash) (*modules.ChainIndex, error) {
	header, err := rep.dagdb.GetHeaderByHash(hash)
	if err != nil {
		return nil, err
	}
	return header.Number, nil
}

func (rep *UnitRepository) GetAssetTxHistory(asset *modules.Asset) ([]*modules.TransactionWithUnitInfo, error) {
	txIds, err := rep.idxdb.GetTokenTxIds(asset)
	if err != nil {
		return nil, err
	}
	var result []*modules.TransactionWithUnitInfo
	for _, txId := range txIds {
		tx, err := rep.GetTransaction(txId)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Retrieve Tx by hash[%s] get error:%s", txId.String(), err.Error()))
		}
		result = append(result, tx)
	}
	return result, nil
}

//func (rep *UnitRepository) SaveNumberByHash(uHash common.Hash, number modules.ChainIndex) error {
//	return rep.dagdb.SaveNumberByHash(uHash, number)
//}
//func (rep *UnitRepository) SaveHashByNumber(uHash common.Hash, number modules.ChainIndex) error {
//	return rep.dagdb.SaveHashByNumber(uHash, number)
//}
//func (rep *UnitRepository) UpdateHeadByBatch(hash common.Hash, number uint64) error {
//	return rep.dagdb.UpdateHeadByBatch(hash, number)
//}

//func (rep *UnitRepository) GetHeaderRlp(hash common.Hash, index uint64) rlp.RawValue {
//	return rep.dagdb.GetHeaderRlp(hash, index)
//}

//func RHashStr(x interface{}) string {
//	x_byte, err := json.Marshal(x)
//	if err != nil {
//		return ""
//	}
//	s256 := sha256.New()
//	s256.Write(x_byte)
//	return fmt.Sprintf("%x", s256.Sum(nil))
//
//}

/**
生成创世单元，需要传入创世单元的配置信息以及coinbase交易
generate genesis unit, need genesis unit configure fields and transactions list
parentUnitHeight=-1,means don't have parent unit
*/
func NewGenesisUnit(txs modules.Transactions, time int64, asset *modules.Asset, parentUnitHeight int64, parentUnitHash common.Hash) (*modules.Unit, error) {
	gUnit := &modules.Unit{}

	// genesis unit height
	chainIndex := &modules.ChainIndex{AssetID: asset.AssetId, IsMain: true, Index: uint64(parentUnitHeight + 1)}

	// transactions merkle root
	root := core.DeriveSha(txs)

	// generate genesis unit header
	header := &modules.Header{
		Number: chainIndex,
		TxRoot: root,
		Time:   time,
	}
	if parentUnitHeight >= 0 { //has parent unit
		header.ParentsHash = []common.Hash{parentUnitHash}
	}

	gUnit.UnitHeader = header
	// copy txs
	gUnit.CopyBody(txs)
	// set unit size
	gUnit.UnitSize = gUnit.Size()
	// set unit hash
	gUnit.UnitHash = gUnit.Hash()
	return gUnit, nil
}

// WithSignature, returns a new unit with the given signature.
// @author Albert·Gou
func GetUnitWithSig(unit *modules.Unit, ks *keystore.KeyStore, signer common.Address) (*modules.Unit, error) {
	// signature unit: only sign header data(without witness and authors fields)
	sign, err1 := ks.SigUnit(unit.UnitHeader, signer)
	if err1 != nil {
		msg := fmt.Sprintf("Failed to write genesis block:%v", err1.Error())
		log.Error(msg)
		return unit, err1
	}
	pubKey, err := ks.GetPublicKey(signer)
	if err != nil {
		return nil, err
	}
	//r := sign[:32]
	//s := sign[32:64]
	//v := sign[64:]
	//if len(v) != 1 {
	//	return unit, errors.New("error.")
	//}
	log.Debugf("Unit[%s] signed by address:%s", unit.Hash().String(), signer.String())
	unit.UnitHeader.Authors = modules.Authentifier{
		PubKey:    pubKey,
		Signature: sign,
	}
	// to set witness list, should be creator himself
	// var authentifier modules.Authentifier
	// authentifier.Address = signer
	// unit.UnitHeader.Witness = append(unit.UnitHeader.Witness, &authentifier)
	// unit.UnitHeader.GroupSign = sign
	return unit, nil
}

/**
创建单元
create common unit
@param mAddr is minner addr
return: correct if error is nil, and otherwise is incorrect
*/
func (rep *UnitRepository) CreateUnit(mAddr *common.Address, txpool txspool.ITxPool, t time.Time) (*modules.Unit, error) {
	log.Debug("Start create unit...")
	rep.lock.RLock()
	begin := time.Now()

	defer func() {
		rep.lock.RUnlock()
		log.Infof("CreateUnit cost time %s", time.Since(begin))
	}()
	//if txpool == nil || !common.IsValidAddress(mAddr.String()) || ks == nil {
	//	log.Debug("UnitRepository", "CreateUnit txpool:", txpool, "mdAddr:", mAddr.String(), "ks:", ks)
	//	return nil, fmt.Errorf("Create unit: nil address or txspool is not allowed")
	//}

	// step1. get mediator responsible for asset (for now is ptn)
	assetId := dagconfig.DagConfig.GetGasToken()

	// step2. compute chain height
	// get current world_state index.
	index := uint64(1)
	isMain := true
	// chainIndex := modules.ChainIndex{AssetID: asset.AssetId, IsMain: isMain, Index: index}
	phash, chainIndex, _, err := rep.propdb.GetNewestUnit(assetId)
	if err != nil {
		chainIndex = &modules.ChainIndex{AssetID: assetId, IsMain: isMain, Index: index + 1}
		log.Error("GetCurrentChainIndex is failed.", "error", err)
	} else {
		chainIndex.Index += 1
	}

	// step3. generate genesis unit header
	header := modules.Header{
		Number:      chainIndex,
		ParentsHash: []common.Hash{},
	}
	header.ParentsHash = append(header.ParentsHash, phash)
	h_hash := header.HashWithOutTxRoot()
	log.Debug("Start txpool.GetSortedTxs...")

	// step4. get transactions from txspool
	poolTxs, _ := txpool.GetSortedTxs(h_hash)

	txIds := []common.Hash{}
	for _, tx := range poolTxs {
		txIds = append(txIds, tx.Tx.Hash())
	}
	log.Infof("txpool.GetSortedTxs cost time %s, include txs:[%#x]", time.Since(begin), txIds)
	// step5. compute minner income: transaction fees + interest

	//交易费用
	ads, err := ComputeTxFees(mAddr, poolTxs)
	if err != nil {
		log.Error("ComputeContractProcessorFees is failed.", "error", err.Error())
		return nil, err
	}
	//保证金利息
	addr, _ := common.StringToAddress("PCGTta3M4t3yXu8uRgkKvaWd2d8DR32W9vM")
	awardAd, err := rep.ComputeAwardsFees(&addr, poolTxs)
	if err != nil && awardAd != nil {
		ads = append(ads, awardAd)
	}

	outAds := arrangeAdditionFeeList(ads)
	coinbase, rewards, err := CreateCoinbase(outAds, t)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	// 若配置增发，或者该单元包含有效交易（rewards>0），则将增发奖励和交易费全发给该mediator。
	txs := make(modules.Transactions, 0)
	if rewards > 0 || dagconfig.DagConfig.IsRewardCoin {
		log.Debug("=======================Is rewards && coinbase tx info ================", "IsReward", dagconfig.DagConfig.IsRewardCoin, "amount", rewards, "hash", coinbase.Hash().String())
		txs = append(txs, coinbase)
	}

	// step6 get unit's txs in txpool's txs
	//TODO must recover
	if len(poolTxs) > 0 {
		for _, tx := range poolTxs {
			t := txspool.PooltxToTx(tx)
			txs = append(txs, t)
		}
	}

	/**
	todo 需要根据交易中涉及到的token类型来确定交易打包到哪个区块
	todo 如果交易中涉及到其他币种的交易，则需要将交易费的单独打包
	*/

	// step8. transactions merkle root
	root := core.DeriveSha(txs)
	log.Infof("core.DeriveSha cost time %s", time.Since(begin))
	// step9. generate genesis unit header
	header.TxRoot = root
	unit := &modules.Unit{}
	unit.UnitHeader = &header
	unit.UnitHash = header.Hash()

	// step10. copy txs
	unit.CopyBody(txs)

	// step11. set size
	unit.UnitSize = unit.Size()
	//units = append(units, unit)
	return unit, nil
}
func ComputeFees(txs []*modules.TxPoolTransaction) (uint64, error) {
	fee := uint64(0)
	for _, tx := range txs {
		fee += tx.TxFee.Amount
	}
	return fee, nil
}

func ComputeTxFees(m *common.Address, txs []*modules.TxPoolTransaction) ([]*modules.Addition, error) {
	if m == nil {
		return nil, errors.New("ComputeTxFees param is nil")
	}
	ads := make([]*modules.Addition, 0)
	for _, tx := range txs {
		if tx.Tx == nil || tx.TxFee == nil {
			continue
		}
		a := &modules.Addition{
			Amount: tx.TxFee.Amount,
			Asset:  *tx.TxFee.Asset,
		}

		if !tx.Tx.IsContractTx() {
			a.Addr = *m
			ads = append(ads, a)
			continue
		} else if !tx.Tx.IsSystemContract(){
			a.Addr = *m
			ads = append(ads, a)
			continue
		}

		addrs := tx.Tx.GetContractTxSignatureAddress()
		nm := len(addrs)
		if nm <= 0 {
			a.Addr = *m
			ads = append(ads, a)
			continue
		}
		t := a.Amount * 6 / 10
		for _, add := range addrs {
			a.Amount = t / uint64(nm) //jury fee= all * 0.6/nm
			a.Addr = add
			ads = append(ads, a)
		}
		a.Amount = a.Amount - t //mediator fee = all * 0.4
		a.Addr = *m
		ads = append(ads, a)
	}
	return ads, nil
}

//获取保证金利息
func (rep *UnitRepository) ComputeAwardsFees(addr *common.Address, poolTxs []*modules.TxPoolTransaction) (*modules.Addition, error) {
	if poolTxs == nil {
		return nil, errors.New("ComputeAwardsFees param is nil")
	}
	awardAd, err := rep.utxoRepository.ComputeAwards(poolTxs, rep.dagdb)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	if awardAd != nil {
		awardAd.Addr = *addr
	}
	return awardAd, nil
}

func arrangeAdditionFeeList(ads []*modules.Addition) []*modules.Addition {
	if len(ads) <= 0 {
		return nil
	}
	out := make([]*modules.Addition, 0)
	for _, a := range ads {
		ok := false
		b := &modules.Addition{}
		for _, b = range out {
			if ok, _ = a.IsEqualStyle(b); ok {
				break
			}
		}
		if ok {
			b.Amount += a.Amount
			continue
		}
		out = append(out, a)
	}
	if len(out) < 1 {
		return nil
	} else {
		return out
	}
}

func (rep *UnitRepository) GetCurrentChainIndex(assetId modules.AssetId) (*modules.ChainIndex, error) {
	_, idx, _, err := rep.propdb.GetNewestUnit(assetId)
	if err != nil {
		return nil, err
	}
	return idx, nil
}

/**
从leveldb中查询GenesisUnit信息
To get genesis unit info from leveldb
*/
func (rep *UnitRepository) GetGenesisUnit() (*modules.Unit, error) {
	rep.lock.RLock()
	defer rep.lock.RUnlock()
	ghash, err := rep.dagdb.GetGenesisUnitHash()
	if err != nil {
		log.Debug("rep: getgenesis by number , current error.", "error", err)
		return nil, err
	}
	return rep.getUnit(ghash)
	// unit key: [HEADER_PREFIX][chain index number]_[chain index]_[unit hash]
	//key := fmt.Sprintf("%s%v_", constants.HEADER_PREFIX, index)

	// data := rep.dagdb.GetPrefix([]byte(key))
	// if len(data) > 1 {
	// 	return nil, fmt.Errorf("multiple genesis unit")
	// } else if len(data) <= 0 {
	// 	return nil, errors.ErrNotFound
	// }
	// for _, v := range data {
	// 	// get unit header
	// 	var uHeader modules.Header
	// 	if err := rlp.DecodeBytes([]byte(v), &uHeader); err != nil {
	// 		return nil, fmt.Errorf("Get genesis unit header:%s", err.Error())
	// 	}
	// 	// generate unit
	// 	unit := modules.Unit{
	// 		UnitHeader: &uHeader,
	// 	}
	// 	// compute unit hash
	// 	unit.UnitHash = unit.Hash()
	// 	// get transaction list
	// 	txs, err := rep.dagdb.GetUnitTransactions(unit.UnitHash)
	// 	if err != nil {
	// 		return nil, fmt.Errorf("Get genesis unit transactions: %s", err.Error())
	// 	}
	// 	unit.Txs = txs
	// 	unit.UnitSize = unit.Size()
	// 	return &unit, nil
	// 	//}
	// }
	// return nil, nil
	//number := modules.ChainIndex{}
	//number.Index = index
	//number.IsMain = true
	//
	////number.AssetID, _ = modules.SetIdTypeByHex(dagconfig.DefaultConfig.PtnAssetHex) //modules.PTNCOIN
	////asset := modules.NewPTNAsset()
	//number.AssetID = modules.CoreAsset.AssetId
	//hash, err := rep.dagdb.GetHashByNumber(number)
	//if err != nil {
	//	log.Debug("rep: getgenesis by number , current error.", "error", err)
	//	return nil, err
	//}
	//log.Debug("rep: get genesis(hash):", "geneseis_hash", hash)
	//return rep.dagdb.getChainUnit(hash)
}

/**
获取创世单元的高度
To get genesis unit height
*/
//func (unitRep *UnitRepository) GenesisHeight() modules.ChainIndex {
//	unit, err := unitRep.GetGenesisUnit()
//	if unit == nil || err != nil {
//		return modules.ChainIndex{}
//	}
//	return unit.UnitHeader.Number
//}
func (unitRep *UnitRepository) IsGenesis(hash common.Hash) bool {
	unit, err := unitRep.dagdb.GetGenesisUnitHash()
	if err != nil {
		return false
	}
	return hash == unit
}

func (rep *UnitRepository) GetUnitTransactions(unitHash common.Hash) (modules.Transactions, error) {
	rep.lock.RLock()
	defer rep.lock.RUnlock()
	txs := modules.Transactions{}
	// get body data: transaction list.
	// if getbody return transactions list, then don't range txHashlist.
	txHashList, err := rep.dagdb.GetBody(unitHash)
	if err != nil {
		return nil, err
	}
	// get transaction by tx'hash.
	for _, txHash := range txHashList {
		tx, err := rep.dagdb.GetTransactionOnly(txHash)
		if err != nil {
			return nil, err
		}
		txs = append(txs, tx)
	}
	return txs, nil
}

/**
为创世单元生成ConfigPayload
To generate config payload for genesis unit
*/
func GenGenesisConfigPayload(genesisConf *core.Genesis, asset *modules.Asset) ([]*modules.ContractInvokePayload, error) {
	writeSets := []modules.ContractWriteSet{}
	digitalWriteSets := []modules.ContractWriteSet{}

	tt := reflect.TypeOf(*genesisConf)
	vv := reflect.ValueOf(*genesisConf)

	for i := 0; i < tt.NumField(); i++ {
		// modified by Albert·Gou, 不是交易，已在其他地方处理
		if strings.Contains(tt.Field(i).Name, "Initial") ||
			strings.Contains(tt.Field(i).Name, "Immutable") {
			continue
		}

		if strings.Compare(tt.Field(i).Name, "SystemConfig") == 0 {
			t := reflect.TypeOf(genesisConf.SystemConfig)
			v := reflect.ValueOf(genesisConf.SystemConfig)
			for k := 0; k < t.NumField(); k++ {
				sk := t.Field(k).Name
				if strings.Contains(sk, "Initial") {
					sk = strings.Replace(sk, "Initial", "", -1)
				}

				//writeSets.ConfigSet = append(writeSets.ConfigSet,
				//	modules.ContractWriteSet{Key: sk, Value: modules.ToPayloadMapValueBytes(v.Field(k).Interface())})
				writeSets = append(writeSets,
					modules.ContractWriteSet{Key: sk, Value: []byte(v.Field(k).String())})
			}
			sysConfByte, _ := json.Marshal(genesisConf.SystemConfig)
			writeSets = append(writeSets, modules.ContractWriteSet{Key: "sysConf", Value: []byte(sysConfByte)})

		} else if strings.Compare(tt.Field(i).Name, "DigitalIdentityConfig") == 0 {
			// 2019.4.12
			t := reflect.TypeOf(genesisConf.DigitalIdentityConfig)
			v := reflect.ValueOf(genesisConf.DigitalIdentityConfig)
			for k := 0; k < t.NumField(); k++ {
				sk := t.Field(k).Name
				digitalWriteSets = append(digitalWriteSets,
					modules.ContractWriteSet{Key: sk, Value: []byte(v.Field(k).String())})
				log.Debugf(">>>> save digital contract state, key:%s", sk)
			}
			digitalConfByte, _ := json.Marshal(genesisConf.DigitalIdentityConfig)
			digitalWriteSets = append(digitalWriteSets, modules.ContractWriteSet{Key: "digitalConf", Value: []byte(digitalConfByte)})
		} else {
			sk := tt.Field(i).Name
			if strings.Contains(sk, "Initial") {
				sk = strings.Replace(sk, "Initial", "", -1)
			}
			writeSets = append(writeSets,
				modules.ContractWriteSet{Key: sk, Value: modules.ToPayloadMapValueBytes(vv.Field(i).Interface())})
		}
	}

	writeSets = append(writeSets,
		modules.ContractWriteSet{Key: modules.FIELD_GENESIS_ASSET, Value: modules.ToPayloadMapValueBytes(*asset)})

	contractInvokePayloads := []*modules.ContractInvokePayload{}
	// generate systemcontract invoke payload
	sysconfigPayload := &modules.ContractInvokePayload{}
	sysconfigPayload.ContractId = syscontract.SysConfigContractAddress.Bytes21()
	sysconfigPayload.WriteSet = writeSets
	contractInvokePayloads = append(contractInvokePayloads, sysconfigPayload)

	// generate digital identity contract invoke pyaload
	digitalPayload := &modules.ContractInvokePayload{
		ContractId: syscontract.DigitalIdentityContractAddress.Bytes21(),
		WriteSet:   digitalWriteSets,
	}
	contractInvokePayloads = append(contractInvokePayloads, digitalPayload)
	return contractInvokePayloads, nil
}

func (rep *UnitRepository) UpdateAccountInfo(msg *modules.Message, account common.Address) error {
	accountUpdateOp, ok := msg.Payload.(*modules.AccountUpdateOperation)
	if !ok {
		return errors.New("not a valid mediator Count Set payload")
	}

	err := rep.statedb.UpdateAccountInfo(account, accountUpdateOp)
	if err != nil {
		return err
	}

	return nil
}

//Get who send this transaction
func (rep *UnitRepository) getRequesterAddress(tx *modules.Transaction) (common.Address, error) {
	msg0 := tx.TxMessages[0]
	if msg0.App != modules.APP_PAYMENT {
		return common.Address{}, errors.New("Invalid Tx, first message must be a payment")
	}
	pay := msg0.Payload.(*modules.PaymentPayload)
	utxo, err := rep.utxoRepository.GetUtxoEntry(pay.Inputs[0].PreviousOutPoint)
	if err != nil {
		return common.Address{}, err
	}
	return tokenengine.GetAddressFromScript(utxo.PkScript)

}

/**
保存单元数据，如果单元的结构基本相同
save genesis unit data
*/
func (rep *UnitRepository) SaveUnit(unit *modules.Unit, isGenesis bool) error {
	rep.lock.Lock()
	defer rep.lock.Unlock()
	uHash := unit.Hash()
	log.Debugf("Try to save a new unit to db:%s", uHash.String())
	// if unit.UnitSize == 0 || unit.Size() == 0 {
	// 	log.Error("Unit is null")
	// 	return fmt.Errorf("Unit is null")
	// } //Validator will check unit

	// step10. save unit header
	// key is like "[HEADER_PREFIX][chain index number]_[chain index]_[unit hash]"
	if err := rep.dagdb.SaveHeader(unit.UnitHeader); err != nil {
		log.Info("SaveHeader:", "error", err.Error())
		return modules.ErrUnit(-3)
	}
	// step5. traverse transactions and save them
	txHashSet := []common.Hash{}
	for txIndex, tx := range unit.Txs {
		err := rep.saveTx4Unit(unit, txIndex, tx)
		if err != nil {
			return err
		}
		txHashSet = append(txHashSet, tx.Hash())
	}
	// step8. save unit body, the value only save txs' hash set, and the key is merkle root
	if err := rep.dagdb.SaveBody(uHash, txHashSet); err != nil {
		log.Info("SaveBody", "error", err.Error())
		return err
	}

	// step 9  save txlookupEntry
	if err := rep.dagdb.SaveTxLookupEntry(unit); err != nil {
		log.Info("SaveTxLookupEntry", "error", err.Error())
		return err
	}
	//step12+ Special process genesis unit
	if isGenesis {
		if err := rep.propdb.SetNewestUnit(unit.Header()); err != nil {
			log.Errorf("Save ChainIndex for genesis error:%s", err.Error())
		}
		//Save StableUnit
		if err := rep.propdb.SetLastStableUnit(uHash, unit.UnitHeader.Number); err != nil {
			log.Info("Set LastStableUnit:", "error", err.Error())
			return modules.ErrUnit(-3)
		}
		rep.dagdb.SaveGenesisUnitHash(unit.Hash())
	}

	// step1 验证 群签名
	// if passed == true , don't validate group sign
	//if !passed {
	//	if state := rep.validate.ValidateUnitGroupSign(unit.Header(), isGenesis); state ==
	// 		modules.UNIT_STATE_INVALID_GROUP_SIGNATURE {
	//		return fmt.Errorf("Validate unit's group sign failed, err number=%d", state)
	//	}
	//}

	// step2. check unit signature, should be compare to mediator list
	//if dagconfig.DefaultConfig.WhetherValidateUnitSignature {
	//	errno := rep.validate.ValidateUnitSignature(unit.UnitHeader, isGenesis)
	//	if int(errno) != modules.UNIT_STATE_VALIDATED && int(errno) != modules.UNIT_STATE_AUTHOR_SIGNATURE_PASSED {
	//		return fmt.Errorf("Validate unit signature, errno=%d", errno)
	//	}
	//}
	//
	//// step3. check unit size
	//if unit.UnitSize != unit.Size() {
	//	log.Info("Validate size", "error", "Size is invalid")
	//	return modules.ErrUnit(-1)
	//}
	//// log.Info("===dag ValidateTransactions===")
	//// step4. check transactions in unit
	//// TODO must recover
	//_, isSuccess, err := rep.validate.ValidateTransactions(&unit.Txs, isGenesis)
	//if err != nil || !isSuccess {
	//	return fmt.Errorf("Validate unit(%s) transactions failed: %v", unit.UnitHash.String(), err)
	//}

	return nil
}

//Save tx in unit
func (rep *UnitRepository) saveTx4Unit(unit *modules.Unit, txIndex int, tx *modules.Transaction) error {
	var requester common.Address
	var err error
	if txIndex > 0 { //coinbase don't have requester
		requester, err = rep.getRequesterAddress(tx)
		if err != nil {
			return err
		}
	}
	txHash := tx.Hash()
	reqId := tx.RequestHash().Bytes()
	// traverse messages
	for msgIndex, msg := range tx.TxMessages {
		// handle different messages
		switch msg.App {
		case modules.APP_PAYMENT:
			if ok := rep.savePaymentPayload(txHash, msg.Payload.(*modules.PaymentPayload), uint32(msgIndex)); ok != true {
				return fmt.Errorf("Save payment payload error.")
			}
		case modules.APP_CONTRACT_TPL:
			if ok := rep.saveContractTpl(unit.UnitHeader.Number, uint32(txIndex), msg); ok != true {
				return fmt.Errorf("Save contract template error.")
			}
		case modules.APP_CONTRACT_DEPLOY:
			if ok := rep.saveContractInitPayload(unit.UnitHeader.Number, uint32(txIndex), msg); ok != true {
				return fmt.Errorf("Save contract init payload error.")
			}
		case modules.APP_CONTRACT_INVOKE:
			if ok := rep.saveContractInvokePayload(tx, unit.UnitHeader.Number, uint32(txIndex), msg); ok != true {
				return fmt.Errorf("save contract invode payload error")
			}
		case modules.APP_CONTRACT_STOP:
			if ok := rep.saveContractStop(reqId, msg); !ok {
				return fmt.Errorf("save contract stop payload failed.")
			}
			//case modules.APP_CONFIG:
			//	if ok := rep.saveConfigPayload(txHash, msg, unit.UnitHeader.Number, uint32(txIndex)); ok == false {
			//		return fmt.Errorf("Save contract invode payload error.")
			//	}
		case modules.OP_MEDIATOR_CREATE:
			if !rep.MediatorCreateApply(msg) {
				return fmt.Errorf("apply Mediator Creating Operation error")
			}
		case modules.OP_ACCOUNT_UPDATE:
			if err := rep.UpdateAccountInfo(msg, requester); err != nil {
				return fmt.Errorf("apply Account Updating Operation error")
			}
		case modules.APP_CONTRACT_TPL_REQUEST:
			// todo
		case modules.APP_CONTRACT_DEPLOY_REQUEST:
			if ok := rep.saveContractDeployReq(reqId, msg); !ok {
				return fmt.Errorf("save contract of deploy request failed.")
			}
		case modules.APP_CONTRACT_STOP_REQUEST:
			if ok := rep.saveContractStopReq(reqId, msg); !ok {
				return fmt.Errorf("save contract of stop request failed.")
			}
		case modules.APP_CONTRACT_INVOKE_REQUEST:
			if ok := rep.saveContractInvokeReq(reqId, msg); !ok {
				return fmt.Errorf("save contract of invoke request failed.")
			}
		case modules.APP_SIGNATURE:
			if ok := rep.saveSignature(reqId, msg); !ok {
				return fmt.Errorf("save contract of signature failed.")
			}
		case modules.APP_DATA:
			if ok := rep.saveDataPayload(txHash, msg); ok != true {
				return fmt.Errorf("save data payload faild.")
			}
		default:
			return fmt.Errorf("Message type is not supported now: %v", msg.App)
		}
	}
	// step6. save transaction
	if err := rep.dagdb.SaveTransaction(tx); err != nil {
		log.Info("Save transaction:", "error", err.Error())
		return err
	}
	if dagconfig.DagConfig.AddrTxsIndex {
		//Index TxId for to address
		addresses := getPayToAddresses(tx)
		for _, addr := range addresses {
			rep.idxdb.SaveAddressTxId(addr, txHash)
		}
		//Index from address to txid
		fromAddrs := rep.getPayFromAddresses(tx)
		for _, addr := range fromAddrs {
			rep.idxdb.SaveAddressTxId(addr, txHash)
		}
	}

	return nil
}
func getPayToAddresses(tx *modules.Transaction) []common.Address {
	resultMap := map[common.Address]int{}
	for _, msg := range tx.TxMessages {
		if msg.App == modules.APP_PAYMENT {
			pay := msg.Payload.(*modules.PaymentPayload)
			for _, out := range pay.Outputs {
				addr, _ := tokenengine.GetAddressFromScript(out.PkScript)
				if _, ok := resultMap[addr]; !ok {
					resultMap[addr] = 1
				}
			}
		}
	}
	keys := make([]common.Address, 0, len(resultMap))
	for k := range resultMap {
		keys = append(keys, k)
	}
	return keys
}

func (rep *UnitRepository) getPayFromAddresses(tx *modules.Transaction) []common.Address {
	resultMap := map[common.Address]int{}
	for _, msg := range tx.TxMessages {
		if msg.App == modules.APP_PAYMENT {
			pay := msg.Payload.(*modules.PaymentPayload)
			for _, input := range pay.Inputs {
				if input.PreviousOutPoint != nil {
					utxo, err := rep.utxoRepository.GetUtxoEntry(input.PreviousOutPoint)
					if err != nil {
						log.Errorf("Get utxo by [%s] throw an error:%s", input.PreviousOutPoint.String(), err.Error())
						return []common.Address{}
					}
					addr, _ := tokenengine.GetAddressFromScript(utxo.PkScript)
					if _, ok := resultMap[addr]; !ok {
						resultMap[addr] = 1
					}
				}

			}
		}
	}
	keys := make([]common.Address, 0, len(resultMap))
	for k := range resultMap {
		keys = append(keys, k)
	}
	return keys
}

func getMaindata(tx *modules.Transaction) string {
	var maindata string
	for _, msg := range tx.TxMessages {
		if msg.App == modules.APP_DATA {
			pay := msg.Payload.(*modules.DataPayload)
			maindata = string(pay.MainData)

		}
	}
	return maindata
}

func getExtradata(tx *modules.Transaction) string {
	var extradata string
	for _, msg := range tx.TxMessages {
		if msg.App == modules.APP_DATA {
			pay := msg.Payload.(*modules.DataPayload)
			extradata = string(pay.ExtraData)

		}
	}
	return extradata
}

/**
保存PaymentPayload
save PaymentPayload data
*/
func (rep *UnitRepository) savePaymentPayload(txHash common.Hash, msg *modules.PaymentPayload, msgIndex uint32) bool {
	// if inputs is none then it is just a normal coinbase transaction
	// otherwise, if inputs' length is 1, and it PreviousOutPoint should be none
	// if this is a create token transaction, the Extra field should be AssetInfo struct's [rlp] encode bytes
	// if this is a create token transaction, should be return a assetid
	// save utxo
	err := rep.utxoRepository.UpdateUtxo(txHash, msg, msgIndex)
	if err != nil {
		log.Error("Update utxo failed.", "error", err)
		return false
	}

	//对PRC721类型的通证的流转历史记录索引
	if dagconfig.DefaultConfig.Token721TxIndex {

		for _, output := range msg.Outputs {
			asset := output.Asset
			if asset.AssetId.GetAssetType() == modules.AssetType_NonFungibleToken {
				if err = rep.idxdb.SaveTokenTxId(asset, txHash); err != nil {
					log.Errorf("Save token and txid index data error:%s", err.Error())
				}
			}

		}
	}
	return true
}

/**
保存DataPayload
save DataPayload data
*/

func (rep *UnitRepository) saveDataPayload(txHash common.Hash, msg *modules.Message) bool {
	var pl interface{}
	pl = msg.Payload

	payload, ok := pl.(*modules.DataPayload)
	if ok == false {
		return false
	}

	if dagconfig.DagConfig.TextFileHashIndex {

		err := rep.idxdb.SaveMainDataTxId(payload.MainData, txHash)
		if err != nil {
			log.Error("error savefilehash", "err", err)
			return false
		}
		return true
	}
	log.Debug("dagconfig textfileindex is false, don't build index for data")
	return true
}

/**
保存配置交易
save config payload
*/
//func (rep *UnitRepository) saveConfigPayload(txHash common.Hash, msg *modules.Message, height *modules.ChainIndex, txIndex uint32) bool {
//	var pl interface{}
//	pl = msg.Payload
//	payload, ok := pl.(*modules.ConfigPayload)
//	if ok == false {
//		return false
//	}
//	version := modules.StateVersion{
//		Height:  height,
//		TxIndex: txIndex,
//	}
//	if err := rep.statedb.SaveConfig(payload.ConfigSet, &version); err != nil {
//		errMsg := fmt.Sprintf("To save config payload error: %s", err)
//		log.Error(errMsg)
//		return false
//	}
//	return true
//}

/**
保存合约调用状态
To save contract invoke state
*/
func (rep *UnitRepository) saveContractInvokePayload(tx *modules.Transaction, height *modules.ChainIndex, txIndex uint32, msg *modules.Message) bool {
	var pl interface{}
	pl = msg.Payload
	payload, ok := pl.(*modules.ContractInvokePayload)
	if ok == false {
		return false
	}
	// save contract state
	// key: [CONTRACT_STATE_PREFIX][contract id]_[field name]_[state version]
	for _, ws := range payload.WriteSet {
		version := &modules.StateVersion{
			Height:  height,
			TxIndex: txIndex,
		}
		//user just want to update it's statedb

		// if payload.ContractId == nil || len(payload.ContractId) == 0 {
		// 	addr, _ := getRequesterAddress(tx)
		// 	// contractid
		// 	rep.statedb.SaveContractState(addr, ws.Key, ws.Value, version)
		// }
		//@jay
		// contractId is never nil.
		if payload.ContractId != nil {
			//addr, _ := getRequesterAddress(tx)
			// contractid
			rep.statedb.SaveContractState(payload.ContractId, ws.Key, ws.Value, version)
		}

		// save new state to database
		// if rep.updateState(payload.ContractId, ws.Key, version, ws.Value) != true {
		// 	continue
		// }
	}
	return true
}

/**
保存合约初始化状态
To save contract init state
*/
func (rep *UnitRepository) saveContractInitPayload(height *modules.ChainIndex, txIndex uint32, msg *modules.Message) bool {
	var pl interface{}
	pl = msg.Payload
	payload, ok := pl.(*modules.ContractDeployPayload)
	if ok == false {
		return false
	}

	// save contract state
	// key: [CONTRACT_STATE_PREFIX][contract id]_[field name]_[state version]
	version := &modules.StateVersion{
		Height:  height,
		TxIndex: txIndex,
	}
	for _, ws := range payload.WriteSet {
		// save new state to database
		if rep.updateState(payload.ContractId, ws.Key, version, ws.Value) != true {
			continue
		}
	}
	//addr := common.NewAddress(payload.ContractId, common.ContractHash)
	// save contract name
	if rep.statedb.SaveContractState(payload.ContractId, "ContractName", payload.Name, version) != nil {
		return false
	}
	// save contract jury list
	if rep.statedb.SaveContractState(payload.ContractId, "ContractJury", payload.Jury, version) != nil {
		return false
	}
	return true
}

/**
保存合约模板代码
To save contract template code
*/
func (rep *UnitRepository) saveContractTpl(height *modules.ChainIndex, txIndex uint32, msg *modules.Message) bool {
	var pl interface{}
	pl = msg.Payload
	payload, ok := pl.(*modules.ContractTplPayload)
	if ok == false {
		log.Error("saveContractTpl", "error", "payload is not ContractTplPayload type")
		return false
	}

	// step1. generate version for every contract template
	version := &modules.StateVersion{
		Height:  height,
		TxIndex: txIndex,
	}
	// step2. save contract template bytecode data
	if err := rep.statedb.SaveContractTemplate(payload.TemplateId, payload.Bytecode, version.Bytes()); err != nil {
		log.Error("SaveContractTemplate", "error", err.Error())
		return false
	}
	// step3. save contract template name, path, Memory
	if err := rep.statedb.SaveContractTemplateState(payload.TemplateId, modules.FIELD_TPL_NAME, payload.Name, version); err != nil {
		log.Error("SaveContractTemplateState when save name", "error", err.Error())
		return false
	}
	if err := rep.statedb.SaveContractTemplateState(payload.TemplateId, modules.FIELD_TPL_PATH, payload.Path, version); err != nil {
		log.Error("SaveContractTemplateState when save path", "error", err.Error())
		return false
	}
	if err := rep.statedb.SaveContractTemplateState(payload.TemplateId, modules.FIELD_TPL_Memory, payload.Memory, version); err != nil {
		log.Error("SaveContractTemplateState when save memory", "error", err.Error())
		return false
	}
	if err := rep.statedb.SaveContractTemplateState(payload.TemplateId, modules.FIELD_TPL_Version, payload.Version, version); err != nil {
		log.Error("SaveContractTemplateState when save version", "error", err.Error())
		return false
	}
	return true
}

/*
保存合约请求的方法
*/

// saveContractdeployReq
func (rep *UnitRepository) saveContractDeployReq(reqid []byte, msg *modules.Message) bool {
	payload, ok := msg.Payload.(*modules.ContractDeployRequestPayload)
	if !ok {
		log.Error("saveContractDeployReq", "error", "payload is not ContractDeployReq type.")
		return false
	}
	err := rep.statedb.SaveContractDeployReq(reqid[:], payload)
	if err != nil {
		log.Info("save contract deploy req payload failed,", "error", err)
		return false
	}
	return true
}

// saveContractInvoke
func (rep *UnitRepository) saveContractInvoke(reqid []byte, msg *modules.Message) bool {
	invoke, ok := msg.Payload.(*modules.ContractInvokePayload)
	if !ok {
		log.Error("saveContractInvoke", "error", "payload is not the ContractInvoke type.")
		return false
	}
	err := rep.statedb.SaveContractInvoke(reqid[:], invoke)
	if err != nil {
		log.Info("save contract invoke payload failed,", "error", err)
		return false
	}
	return true
}

// saveContractInvokeReq
func (rep *UnitRepository) saveContractInvokeReq(reqid []byte, msg *modules.Message) bool {
	invoke, ok := msg.Payload.(*modules.ContractInvokeRequestPayload)
	if !ok {
		log.Error("saveContractInvokeReq", "error", "payload is not the ContractInvokeReq type.")
		return false
	}
	err := rep.statedb.SaveContractInvokeReq(reqid[:], invoke)
	if err != nil {
		log.Info("save contract invoke req payload failed,", "error", err)
		return false
	}
	return true
}

// saveContractStop
func (rep *UnitRepository) saveContractStop(reqid []byte, msg *modules.Message) bool {
	stop, ok := msg.Payload.(*modules.ContractStopPayload)
	if !ok {
		log.Error("saveContractStop", "error", "payload is not the ContractStop type.")
		return false
	}
	err := rep.statedb.SaveContractStop(reqid[:], stop)
	if err != nil {
		log.Info("save contract stop payload failed,", "error", err)
		return false
	}
	return true
}

// saveContractStopReq
func (rep *UnitRepository) saveContractStopReq(reqid []byte, msg *modules.Message) bool {
	stop, ok := msg.Payload.(*modules.ContractStopRequestPayload)
	if !ok {
		log.Error("saveContractStopReq", "error", "payload is not the ContractStopReq type.")
		return false
	}
	err := rep.statedb.SaveContractStopReq(reqid[:], stop)
	if err != nil {
		log.Info("save contract stopReq payload failed,", "error", err)
		return false
	}
	return true
}

// saveSignature
func (rep *UnitRepository) saveSignature(reqid []byte, msg *modules.Message) bool {
	sig, ok := msg.Payload.(*modules.SignaturePayload)
	if !ok {
		log.Error("save contract signature", "error", "payload is not the signature type.")
		return false
	}
	err := rep.statedb.SaveContractSignature(reqid[:], sig)
	if err != nil {
		log.Info("save contract signature payload failed,", "error", err)
		return false
	}
	return true
}

/**
从levedb中根据ChainIndex获得Unit信息
To get unit information by its ChainIndex
*/
//func QueryUnitByChainIndex(db ptndb.Database, number modules.ChainIndex) *modules.Unit {
//	return storage.GetUnitFormIndex(db, number)
//}

/**
创建coinbase交易
To create coinbase transaction
*/
func CreateCoinbase(ads []*modules.Addition, t time.Time) (*modules.Transaction, uint64, error) {
	totalIncome := uint64(0)
	payload := modules.PaymentPayload{}

	if len(ads) != 0 {
		for _, v := range ads {
			script := tokenengine.GenerateLockScript(v.Addr)
			createT := big.Int{}
			additionalInput := modules.Input{
				Extra: createT.SetInt64(t.Unix()).Bytes(),
			}
			additionalOutput := modules.Output{
				Value:    v.Amount,
				Asset:    &v.Asset,
				PkScript: script,
			}
			payload.Inputs = append(payload.Inputs, &additionalInput)
			payload.Outputs = append(payload.Outputs, &additionalOutput)

			totalIncome += v.Amount
		}
	}

	/*
	// setp1. create P2PKH script
	script := tokenengine.GenerateP2PKHLockScript(addr.Bytes())
	// step. compute total income
	totalIncome := int64(income) + int64(ComputeRewards())
	// step2. create payload
	createT := big.Int{}
	input := modules.Input{
		Extra: createT.SetInt64(t.Unix()).Bytes(),
	}
	output := modules.Output{
		Value:    uint64(totalIncome),
		Asset:    asset,
		PkScript: script,
	}
	payload.Inputs = append(payload.Inputs, &input)
	payload.Outputs = append(payload.Outputs, &output)
*/

	// step3. create message
	msg := &modules.Message{
		App:     modules.APP_PAYMENT,
		Payload: &payload,
	}
	// step4. create coinbase
	coinbase := new(modules.Transaction)
	coinbase.TxMessages = append(coinbase.TxMessages, msg)
	return coinbase, totalIncome, nil
}

/**
删除合约状态
To delete contract state
*/
func (rep *UnitRepository) deleteContractState(contractID []byte, field string) {
	oldKeyPrefix := fmt.Sprintf("%s%s^*^%s",
		constants.CONTRACT_STATE_PREFIX,
		hexutil.Encode(contractID[:]),
		field)
	data := rep.statedb.GetPrefix([]byte(oldKeyPrefix))
	for k := range data {
		if err := rep.statedb.DeleteState([]byte(k)); err != nil {
			log.Error("Delete contract state", "error", err.Error())
			continue
		}
	}
}

/**
签名交易
To Sign transaction
*/
//func SignTransaction(txHash common.Hash, addr *common.Address, ks *keystore.KeyStore) (*modules.Authentifier, error) {
//	R, S, V, err := ks.SigTX(txHash, *addr)
//	if err != nil {
//		msg := fmt.Sprintf("Sign transaction error: %s", err)
//		log.Error(msg)
//		return nil, nil
//	}
//	sig := modules.Authentifier{
//		Address: addr.String(),
//		R:       R,
//		S:       S,
//		V:       V,
//	}
//	return &sig, nil
//}

/**
保存contract state
To save contract state
*/
func (rep *UnitRepository) updateState(contractID []byte, key string, version *modules.StateVersion, val interface{}) bool {
	delState, isDel := val.(modules.DelContractState)
	if isDel {
		if delState.IsDelete == false {
			return true
		}
		// delete old state from database
		rep.deleteContractState(contractID, key)

	} else {
		// delete old state from database
		rep.deleteContractState(contractID, key)
		// insert new state
		key := fmt.Sprintf("%s%s^*^%s^*^%s",
			constants.CONTRACT_STATE_PREFIX,
			hexutil.Encode(contractID[:]),
			key,
			version.String())
		// addr := common.NewAddress(contractID, common.ContractHash)
		if err := rep.statedb.SaveContractState(contractID, key, val, version); err != nil {
			log.Error("Save state", "error", err.Error())
			return false
		}
	}
	return true
}

func IsGenesis(hash common.Hash) bool {
	genHash := common.HexToHash(dagconfig.DagConfig.GenesisHash)
	return genHash == hash
}

// GetAddrTransactions containing from && to address
func (rep *UnitRepository) GetAddrTransactions(address common.Address) ([]*modules.TransactionWithUnitInfo, error) {
	rep.lock.RLock()
	defer rep.lock.RUnlock()
	hashs, err := rep.idxdb.GetAddressTxIds(address)
	if err != nil {
		return nil, err
	}
	//alltxs := make(map[string]modules.Transactions)
	txs := make([]*modules.TransactionWithUnitInfo, 0)
	for _, hash := range hashs {
		tx, err := rep.GetTransaction(hash)
		if err != nil {
			return nil, err
		}
		txs = append(txs, tx)
	}
	//alltxs["into"] = txs

	// @yangyu 20 Feb, 2019. There is no SetFromAddressTxIds in project.
	//// from tx
	//txs = make(modules.Transactions, 0)
	//from_hashs, err1 := rep.idxdb.GetFromAddressTxIds(addr)
	//if err1 == nil {
	//	for _, hash := range from_hashs {
	//		tx, _, _, _ := rep.dagdb.GetTransaction(hash)
	//		txs = append(txs, tx)
	//	}
	//}
	//alltxs["out"] = txs
	return txs, err
}

func (rep *UnitRepository) GetFileInfo(filehash []byte) ([]*modules.FileInfo, error) {
	rep.lock.RLock()
	defer rep.lock.RUnlock()
	hashs, err := rep.idxdb.GetMainDataTxIds(filehash)
	if err != nil {
		return nil, err
	}
	mds0, err := rep.GetFileInfoByHash(hashs)
	if hashs == nil {
		hash := common.HexToHash(string(filehash))
		hashs = append(hashs, hash)
		mds1, err := rep.GetFileInfoByHash(hashs)
		return mds1, err
	}
	return mds0, err
}

func (rep *UnitRepository) GetFileInfoByHash(hashs []common.Hash) ([]*modules.FileInfo, error) {
	rep.lock.RLock()
	defer rep.lock.RUnlock()
	var mds []*modules.FileInfo
	for _, hash := range hashs {
		var md modules.FileInfo
		//txlookup, err := rep.dagdb.GetTxLookupEntry(hash)
		//if err != nil {
		//	return nil, err
		//}
		//
		//header, err := rep.dagdb.GetHeaderByHash(unithash)
		//if err != nil {
		//	return nil, err
		//}
		//for _, v := range header.ParentsHash {
		//	md.ParentsHash = v
		//}
		tx, err := rep.GetTransaction(hash)
		if err != nil {
			return nil, err
		}
		md.MainData = getMaindata(tx.Transaction)
		md.ExtraData = getExtradata(tx.Transaction)
		md.UnitHash = tx.UnitHash
		md.UintHeight = tx.UnitIndex
		md.Txid = tx.Hash()
		md.Timestamp = tx.Timestamp
		mds = append(mds, &md)
	}
	return mds, nil
}

func (rep *UnitRepository) GetLastIrreversibleUnit(assetID modules.AssetId) (*modules.Unit, error) {
	rep.lock.RLock()
	defer rep.lock.RUnlock()
	hash, _, err := rep.propdb.GetLastStableUnit(assetID)
	if err != nil {
		return nil, err
	}
	return rep.getUnit(hash)
}

func (rep *UnitRepository) GetTxFromAddress(tx *modules.Transaction) ([]common.Address, error) {
	rep.lock.RLock()
	defer rep.lock.RUnlock()
	result := []common.Address{}
	for _, msg := range tx.TxMessages {
		if msg.App == modules.APP_PAYMENT {
			pay := msg.Payload.(*modules.PaymentPayload)
			for _, input := range pay.Inputs {
				if input.PreviousOutPoint != nil {
					utxo, err := rep.utxoRepository.GetUtxoEntry(input.PreviousOutPoint)
					if err != nil {
						return nil, errors.New("Get utxo by " + input.PreviousOutPoint.String() + " error:" + err.Error())
					}
					addr, _ := tokenengine.GetAddressFromScript(utxo.PkScript)
					result = append(result, addr)
				}
			}
		}
	}
	return result, nil
}
