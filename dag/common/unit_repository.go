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
	"reflect"
	"strings"
	"time"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/core"

	"github.com/palletone/go-palletone/core/accounts/keystore"
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

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/dag/constants"
	"github.com/palletone/go-palletone/dag/parameter"
)

type IUnitRepository interface {
	GetGenesisUnit() (*modules.Unit, error)
	//GenesisHeight() modules.ChainIndex
	SaveUnit(unit *modules.Unit, isGenesis bool) error
	CreateUnit(mAddr common.Address, txpool txspool.ITxPool, propdb IPropRepository, t time.Time) (*modules.Unit, error)
	IsGenesis(hash common.Hash) bool
	GetAddrTransactions(addr common.Address) ([]*modules.TransactionWithUnitInfo, error)
	GetHeaderByHash(hash common.Hash) (*modules.Header, error)
	GetHeaderList(hash common.Hash, parentCount int) ([]*modules.Header, error)
	SaveHeader(header *modules.Header) error
	SaveNewestHeader(header *modules.Header) error
	SaveHeaders(headers []*modules.Header) error
	GetHeaderByNumber(index *modules.ChainIndex) (*modules.Header, error)
	IsHeaderExist(uHash common.Hash) (bool, error)
	GetHashByNumber(number *modules.ChainIndex) (common.Hash, error)

	GetUnitTransactions(hash common.Hash) (modules.Transactions, error)
	GetUnit(hash common.Hash) (*modules.Unit, error)

	GetBody(unitHash common.Hash) ([]common.Hash, error)
	GetTransaction(hash common.Hash) (*modules.TransactionWithUnitInfo, error)
	GetTransactionOnly(hash common.Hash) (*modules.Transaction, error)
	IsTransactionExist(txHash common.Hash) (bool, error)
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
	GetTxRequesterAddress(tx *modules.Transaction) (common.Address, error)
	//根据现有Tx数据，重新构建地址和Tx的关系索引
	RefreshAddrTxIndex() error
	GetAssetReference(asset []byte) ([]*modules.ProofOfExistence, error)
	QueryProofOfExistenceByReference(ref []byte) ([]*modules.ProofOfExistence, error)
	SubscribeSysContractStateChangeEvent(ob AfterSysContractStateChangeEventFunc)
	SaveCommon(key, val []byte) error
	RebuildAddrTxIndex() error
}
type UnitRepository struct {
	dagdb          storage.IDagDb
	idxdb          storage.IIndexDb
	statedb        storage.IStateDb
	propdb         storage.IPropertyDb
	utxoRepository IUtxoRepository
	tokenEngine    tokenengine.ITokenEngine
	lock           sync.RWMutex
	observers      []AfterSysContractStateChangeEventFunc
}

//type Observer interface {
//	//更新事件
//	AfterSysContractStateChangeEvent(event *modules.SysContractStateChangeEvent)
//}
type AfterSysContractStateChangeEventFunc func(event *modules.SysContractStateChangeEvent)

func NewUnitRepository(dagdb storage.IDagDb, idxdb storage.IIndexDb,
	utxodb storage.IUtxoDb, statedb storage.IStateDb,
	propdb storage.IPropertyDb,
	engine tokenengine.ITokenEngine) *UnitRepository {
	utxoRep := NewUtxoRepository(utxodb, idxdb, statedb, propdb, engine)
	return &UnitRepository{
		dagdb:          dagdb,
		idxdb:          idxdb,
		statedb:        statedb,
		utxoRepository: utxoRep,
		propdb:         propdb,
		tokenEngine:    engine,
	}
}

func NewUnitRepository4Db(db ptndb.Database, tokenEngine tokenengine.ITokenEngine) *UnitRepository {
	dagdb := storage.NewDagDb(db)
	utxodb := storage.NewUtxoDb(db, tokenEngine)
	statedb := storage.NewStateDb(db)
	idxdb := storage.NewIndexDb(db)
	propdb := storage.NewPropertyDb(db)
	utxoRep := NewUtxoRepository(utxodb, idxdb, statedb, propdb, tokenEngine)
	return &UnitRepository{
		dagdb:          dagdb,
		idxdb:          idxdb,
		statedb:        statedb,
		propdb:         propdb,
		utxoRepository: utxoRep,
		tokenEngine:    tokenEngine,
	}
}

func (rep *UnitRepository) SubscribeSysContractStateChangeEvent(ob AfterSysContractStateChangeEventFunc) {
	if rep.observers == nil {
		rep.observers = []AfterSysContractStateChangeEventFunc{}
	}

	rep.observers = append(rep.observers, ob)
}

func (rep *UnitRepository) GetHeaderByHash(hash common.Hash) (*modules.Header, error) {
	return rep.dagdb.GetHeaderByHash(hash)
}
func (rep *UnitRepository) GetHeaderList(hash common.Hash, parentCount int) ([]*modules.Header, error) {
	log.Debugf("GetHeaderList unitRepository lock [%s].", hash.String())
	rep.lock.RLock()
	defer log.Debugf("GetHeaderList unitRepository unlock [%s].", hash.String())
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
func (rep *UnitRepository) SaveNewestHeader(header *modules.Header) error {
	err := rep.dagdb.SaveHeader(header)
	if err != nil {
		return err
	}
	return rep.propdb.SetNewestUnit(header)
}
func (rep *UnitRepository) SaveHeaders(headers []*modules.Header) error {
	return rep.dagdb.SaveHeaders(headers)
}
func (rep *UnitRepository) GetHeaderByNumber(index *modules.ChainIndex) (*modules.Header, error) {
	log.Debugf("GetHeaderByNumber unitRepository lock [%s].", index.String())
	rep.lock.RLock()
	defer log.Debugf("GetHeaderByNumber unitRepository unlock [%s].", index.String())
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

func (rep *UnitRepository) IsTransactionExist(txHash common.Hash) (bool, error) {
	return rep.dagdb.IsTransactionExist(txHash)
}

func (rep *UnitRepository) GetUnit(hash common.Hash) (*modules.Unit, error) {
	log.Debugf("GetUnit unitRepository lock [%s].", hash.String())
	rep.lock.RLock()
	defer log.Debugf("GetUnit unitRepository unlock [%s].", hash.String())
	defer rep.lock.RUnlock()
	return rep.getUnit(hash)
}

func (rep *UnitRepository) getUnit(hash common.Hash) (*modules.Unit, error) {
	uHeader, err := rep.dagdb.GetHeaderByHash(hash)
	if err != nil {
		log.Debug("getChainUnit when GetHeaderByHash failed ", "error", err, "hash", hash.String())
		return nil, err
	}
	txs, err := rep.dagdb.GetUnitTransactions(hash)
	if err != nil {
		log.Debug("getChainUnit when GetUnitTransactions failed ", "error", err, "hash", hash.String())
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
	resultTx := &modules.TransactionWithUnitInfo{Transaction: tx, UnitHash: txlookup.UnitHash,
		UnitIndex: txlookup.UnitIndex, TxIndex: txlookup.Index, Timestamp: txlookup.Timestamp}
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
func (rep *UnitRepository) SaveCommon(key, val []byte) error {
	return rep.dagdb.SaveCommon(key, val)
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
	result := make([]*modules.TransactionWithUnitInfo, 0)
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
func NewGenesisUnit(txs modules.Transactions, time int64, asset *modules.Asset, parentUnitHeight int64,
	parentUnitHash common.Hash) (*modules.Unit, error) {
	gUnit := &modules.Unit{}

	// genesis unit height
	chainIndex := &modules.ChainIndex{AssetID: asset.AssetId, Index: uint64(parentUnitHeight + 1)}

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
		msg := fmt.Sprintf("Failed to Sig Unit:%v", err1.Error())
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
func (rep *UnitRepository) CreateUnit(mAddr common.Address, txpool txspool.ITxPool,
	propdb IPropRepository, t time.Time) (*modules.Unit, error) {
	log.Debug("create unit lock unitRepository.")
	rep.lock.RLock()
	defer rep.lock.RUnlock()
	begin := time.Now()

	// step1. get mediator responsible for asset (for now is ptn)
	assetId := dagconfig.DagConfig.GetGasToken()

	// step2. compute chain height
	// get current world_state index.
	index := uint64(1)
	phash, chainIndex, err := propdb.GetNewestUnit(assetId)
	if err != nil {
		chainIndex = &modules.ChainIndex{AssetID: assetId, Index: index + 1}
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
	log.Infof("Start txpool.GetSortedTxs..., parent hash:%s", phash.String())

	// step4. get transactions from txspool
	poolTxs, _ := txpool.GetSortedTxs(h_hash, chainIndex.Index)

	log.Debugf("txpool.GetSortedTxs cost time %s", time.Since(begin))
	// step5. compute minner income: transaction fees + interest
	tt := time.Now()
	//交易费用(包含利息)
	txs2 := []*modules.Transaction{}
	for _, tx := range poolTxs {
		txs2 = append(txs2, tx.Tx)
	}
	ads, err := rep.ComputeTxFeesAllocate(mAddr, txs2)
	if err != nil {
		txs2Ids := ""
		for _, tx := range txs2 {
			txs2Ids += tx.Hash().String() + ","
		}

		pooltxStatusStr := ""
		for txid, pooltx := range txpool.AllTxpoolTxs() {
			pooltxStatusStr += txid.String() + ":UnitHash[" + pooltx.UnitHash.String() + "];"
		}
		log.Error("CreateUnit", "ComputeTxFees is failed, error", err.Error(), "txs in this unit", txs2Ids, "pool all tx:", pooltxStatusStr)
		return nil, err
	}

	//出块奖励
	rewardAd := rep.ComputeGenerateUnitReward(mAddr, assetId.ToAsset())
	if rewardAd != nil && rewardAd.Amount > 0 {
		ads = append(ads, rewardAd)
	}

	outAds := arrangeAdditionFeeList(ads)

	coinbase, rewards, err := rep.CreateCoinbase(outAds, chainIndex.Index)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	txs := make(modules.Transactions, 0)
	if len(outAds) > 0 {
		log.Debug("=======================Is rewards && coinbase tx info ================",
			"amount", rewards, "hash", coinbase.Hash().String())
		txs = append(txs, coinbase)
	}

	illegalTxs := make([]uint16, 0)
	// step6 get unit's txs in txpool's txs
	//TODO must recover
	if len(poolTxs) > 0 {
		for idx, tx := range poolTxs {
			t := txspool.PooltxToTx(tx)
			reqId := t.RequestHash()

			//标记交易有效性
			markTxIllegal(rep.statedb, t)
			if t.Illegal {
				illegalTxs = append(illegalTxs, uint16(idx))
				log.Debugf("[%s]CreateUnit, contract is illegal, txHash[%s]", reqId.String()[0:8], t.Hash().String())
			}
			txs = append(txs, t)
		}
	}
	log.Debugf("create coinbase tx cost time %s, unit tx num[%d]", time.Since(tt), len(txs))
	/**
	todo 需要根据交易中涉及到的token类型来确定交易打包到哪个区块
	todo 如果交易中涉及到其他币种的交易，则需要将交易费的单独打包
	*/

	// step8. transactions merkle root
	root := core.DeriveSha(txs)
	// step9. generate genesis unit header
	header.TxsIllegal = illegalTxs
	header.TxRoot = root
	unit := &modules.Unit{}
	unit.UnitHeader = &header
	unit.UnitHash = header.Hash()

	// step10. copy txs
	unit.CopyBody(txs)

	// step11. set size
	unit.UnitSize = unit.Size()
	log.Debugf("CreateUnit[%s] and create unit unlock unitRepository cost time %s", unit.UnitHash.String(), time.Since(begin))
	return unit, nil
}

func checkReadSetValid(dag storage.IStateDb, contractId []byte, readSet []modules.ContractReadSet) bool {
	for _, rd := range readSet {
		_, v, err := dag.GetContractState(contractId, rd.Key)
		if err != nil {
			log.Debug("checkReadSetValid", "GetContractState fail, contractId", contractId)
			return false
		}
		if v != nil && !v.Equal(rd.Version) {
			log.Debugf("checkReadSetValid, not equal, contractId[%x], local ver1[%v], ver2[%v]", contractId, v, rd.Version)
			return false
		}
	}
	return true
}

func markTxIllegal(dag storage.IStateDb, tx *modules.Transaction) {
	if tx == nil {
		return
	}
	if !tx.IsContractTx() {
		return
	}
	if tx.IsSystemContract() {
		return
	}
	var readSet []modules.ContractReadSet
	var contractId []byte

	for _, msg := range tx.TxMessages {
		switch msg.App {
		case modules.APP_CONTRACT_DEPLOY:
			payload := msg.Payload.(*modules.ContractDeployPayload)
			readSet = payload.ReadSet
			contractId = payload.ContractId
		case modules.APP_CONTRACT_INVOKE:
			payload := msg.Payload.(*modules.ContractInvokePayload)
			readSet = payload.ReadSet
			contractId = payload.ContractId
		case modules.APP_CONTRACT_STOP:
			payload := msg.Payload.(*modules.ContractStopPayload)
			readSet = payload.ReadSet
			contractId = payload.ContractId
		}
	}
	valid := checkReadSetValid(dag, contractId, readSet)
	tx.Illegal = !valid
}

type tempTxs struct {
	allUtxo map[modules.OutPoint]*modules.Utxo
	rep     IUtxoRepository
}

func (txs *tempTxs) getUtxoEntryFromTxs(outpoint *modules.OutPoint) (*modules.Utxo, error) {
	if utxo, ok := txs.allUtxo[*outpoint]; ok {
		return utxo, nil
	}
	return txs.rep.GetUtxoEntry(outpoint)
}
func (rep *UnitRepository) ComputeTxFeesAllocate(m common.Address, txs []*modules.Transaction) (
	[]*modules.Addition, error) {

	ads := make([]*modules.Addition, 0)
	tempTxs := &tempTxs{allUtxo: make(map[modules.OutPoint]*modules.Utxo), rep: rep.utxoRepository}
	for _, tx := range txs {
		utxos := tx.GetNewUtxos()
		for o, u := range utxos {
			tempTxs.allUtxo[o] = u
		}
		allowcate, err := tx.GetTxFeeAllocate(tempTxs.getUtxoEntryFromTxs, rep.tokenEngine.GetScriptSigners, m)
		if err != nil {
			return nil, err
		}
		ads = append(ads, allowcate...)
	}

	return ads, nil
}

//,Mediator奖励
func (rep *UnitRepository) ComputeGenerateUnitReward(m common.Address, asset *modules.Asset) *modules.Addition {
	a := &modules.Addition{
		Addr:   m,
		Amount: ComputeGenerateUnitReward(),
		Asset:  asset,
	}
	return a
}

//获取保证金利息
//func (rep *UnitRepository) ComputeAwardsFees(addr *common.Address,
// poolTxs []*modules.TxPoolTransaction) (*modules.Addition, error) {
//	if poolTxs == nil {
//		return nil, errors.New("ComputeAwardsFees param is nil")
//	}
//	awardAd, err := rep.utxoRepository.ComputeAwards(poolTxs, rep.dagdb)
//	if err != nil {
//		log.Error(err.Error())
//		return nil, err
//	}
//	if awardAd != nil {
//		awardAd.Addr = *addr
//	}
//	return awardAd, nil
//}

func arrangeAdditionFeeList(ads []*modules.Addition) []*modules.Addition {
	if len(ads) <= 0 {
		return nil
	}
	out := make(map[string]*modules.Addition)
	for _, a := range ads {
		key := a.Key()
		b, ok := out[key]
		if ok {
			b.Amount += a.Amount
		} else {
			out[key] = a
		}
	}
	if len(out) < 1 {
		return nil
	}
	result := make([]*modules.Addition, 0)
	for _, v := range out {
		result = append(result, v)
	}
	return result
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
}

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
func GenGenesisConfigPayload(genesisConf *core.Genesis, asset *modules.Asset) (
	[]*modules.ContractInvokePayload, error) {
	//writeSets := []modules.ContractWriteSet{}
	digitalWriteSets := []modules.ContractWriteSet{}

	tt := reflect.TypeOf(*genesisConf)
	//vv := reflect.ValueOf(*genesisConf)

	for i := 0; i < tt.NumField(); i++ {
		// modified by Albert·Gou, 不是交易，已在其他地方处理
		//if strings.Contains(tt.Field(i).Name, "Initial") ||
		//	strings.Contains(tt.Field(i).Name, "Immutable") {
		//	continue
		//}

		//if strings.Compare(tt.Field(i).Name, "SystemConfig") == 0 {
		//	t := reflect.TypeOf(genesisConf.SystemConfig)
		//	v := reflect.ValueOf(genesisConf.SystemConfig)
		//	for k := 0; k < t.NumField(); k++ {
		//		sk := t.Field(k).Name
		//		if strings.Contains(sk, "Initial") {
		//			sk = strings.Replace(sk, "Initial", "", -1)
		//		}
		//
		//		//writeSets.ConfigSet = append(writeSets.ConfigSet,
		//		//	modules.ContractWriteSet{Key: sk, Value: modules.ToPayloadMapValueBytes(v.Field(k).Interface())})
		//		writeSets = append(writeSets,
		//			modules.ContractWriteSet{Key: sk, Value: []byte(v.Field(k).String())})
		//	}
		//	sysConfByte, _ := json.Marshal(genesisConf.SystemConfig)
		//	writeSets = append(writeSets, modules.ContractWriteSet{Key: "sysConf", Value: []byte(sysConfByte)})
		//} else
		if strings.Compare(tt.Field(i).Name, "DigitalIdentityConfig") == 0 {
			// 2019.4.12
			t := reflect.TypeOf(genesisConf.DigitalIdentityConfig)
			v := reflect.ValueOf(genesisConf.DigitalIdentityConfig)
			for k := 0; k < t.NumField(); k++ {
				sk := t.Field(k).Name
				digitalWriteSets = append(digitalWriteSets,
					modules.ContractWriteSet{Key: sk, Value: []byte(v.Field(k).String())})
			}
			digitalConfByte, _ := json.Marshal(genesisConf.DigitalIdentityConfig)
			digitalWriteSets = append(digitalWriteSets, modules.ContractWriteSet{
				Key: "digitalConf", Value: digitalConfByte})
		}
		//else {
		//	sk := tt.Field(i).Name
		//	if strings.Contains(sk, "Initial") {
		//		sk = strings.Replace(sk, "Initial", "", -1)
		//	}
		//	writeSets = append(writeSets,
		//		modules.ContractWriteSet{Key: sk, Value: modules.ToPayloadMapValueBytes(vv.Field(i).Interface())})
		//}
	}

	//writeSets = append(writeSets,
	//	modules.ContractWriteSet{Key: modules.FIELD_GENESIS_ASSET, Value: modules.ToPayloadMapValueBytes(*asset)})

	contractInvokePayloads := []*modules.ContractInvokePayload{}
	// generate systemcontract invoke payload
	//sysconfigPayload := &modules.ContractInvokePayload{}
	//sysconfigPayload.ContractId = syscontract.SysConfigContractAddress.Bytes()
	//sysconfigPayload.WriteSet = writeSets
	//contractInvokePayloads = append(contractInvokePayloads, sysconfigPayload)

	// generate digital identity contract invoke pyaload
	digitalPayload := &modules.ContractInvokePayload{
		ContractId: syscontract.DigitalIdentityContractAddress.Bytes(),
		WriteSet:   digitalWriteSets,
	}
	contractInvokePayloads = append(contractInvokePayloads, digitalPayload)
	return contractInvokePayloads, nil
}

func (rep *UnitRepository) updateAccountInfo(msg *modules.Message, account common.Address,
	index *modules.ChainIndex, txIdx uint32) error {
	accountUpdateOp, ok := msg.Payload.(*modules.AccountStateUpdatePayload)
	if !ok {
		return errors.New("not a valid AccountStateUpdatePayload")
	}

	version := &modules.StateVersion{TxIndex: txIdx, Height: index}
	err := rep.statedb.SaveAccountStates(account, accountUpdateOp.WriteSet, version)
	if err != nil {
		return err
	}

	return nil
}

//Get who send this transaction
func (rep *UnitRepository) GetTxRequesterAddress(tx *modules.Transaction) (common.Address, error) {
	msg0 := tx.TxMessages[0]
	if msg0.App != modules.APP_PAYMENT {
		return common.Address{}, errors.New("Invalid Tx, first message must be a payment")
	}
	pay := msg0.Payload.(*modules.PaymentPayload)

	utxo, err := rep.utxoRepository.GetUtxoEntry(pay.Inputs[0].PreviousOutPoint)
	if err != nil {
		return common.Address{}, err
	}
	return rep.tokenEngine.GetAddressFromScript(utxo.PkScript)

}

/**
保存单元数据，如果单元的结构基本相同
save genesis unit data
*/
func (rep *UnitRepository) SaveUnit(unit *modules.Unit, isGenesis bool) error {
	tt := time.Now()
	log.Debugf("saveUnit[%s] lock unitRepository.", unit.UnitHash.String())
	rep.lock.Lock()
	defer log.Debugf("saveUnit[%s] and unlock unitRepository cost time: %s", unit.UnitHash.String(), time.Since(tt))
	defer rep.lock.Unlock()
	uHash := unit.Hash()
	// step1. save unit header
	// key is like "[HEADER_PREFIX][chain index number]_[chain index]_[unit hash]"
	if err := rep.dagdb.SaveHeader(unit.UnitHeader); err != nil {
		log.Info("SaveHeader:", "error", err.Error())
		return modules.ErrUnit(-3)
	}
	// step2. traverse transactions and save them

	txHashSet := []common.Hash{}
	for txIndex, tx := range unit.Txs {
		err := rep.saveTx4Unit(unit, txIndex, tx)
		if err != nil {
			return err
		}
		//log.Debugf("save transaction, hash[%s] tx_index[%d]", tx.Hash().String(), txIndex)
		txHashSet = append(txHashSet, tx.Hash())
	}
	// step3. save unit body, the value only save txs' hash set, and the key is merkle root
	if err := rep.dagdb.SaveBody(uHash, txHashSet); err != nil {
		log.Info("SaveBody", "error", err.Error())
		return err
	}
	// step4  save txlookupEntry
	if err := rep.dagdb.SaveTxLookupEntry(unit); err != nil {
		log.Info("SaveTxLookupEntry", "error", err.Error())
		return err
	}
	//step5  Special process genesis unit
	if isGenesis {
		if err := rep.propdb.SetNewestUnit(unit.Header()); err != nil {
			log.Errorf("Save ChainIndex for genesis error:%s", err.Error())
		}
		rep.dagdb.SaveGenesisUnitHash(unit.Hash())
	}
	return nil
}

//Save tx in unit
func (rep *UnitRepository) saveTx4Unit(unit *modules.Unit, txIndex int, tx *modules.Transaction) error {
	var requester common.Address
	var err error
	if txIndex > 0 { //coinbase don't have requester
		requester, err = rep.GetTxRequesterAddress(tx)
		if err != nil {
			return err
		}
	}
	txHash := tx.Hash()
	reqId := tx.RequestHash().Bytes()
	unitHash := unit.Hash()
	unitTime := unit.Timestamp()
	unitHeight := unit.Header().Index()

	templateId := make([]byte, 0)
	// traverse messages
	var installReq *modules.ContractInstallRequestPayload
	reqIndex := tx.GetRequestMsgIndex()
	for msgIndex, msg := range tx.TxMessages {
		// handle different messages
		if tx.Illegal && msgIndex > reqIndex {
			break
		}
		switch msg.App {
		case modules.APP_PAYMENT:
			if ok := rep.savePaymentPayload(unit.Timestamp(), txHash, msg.Payload.(*modules.PaymentPayload),
				uint32(msgIndex)); !ok {
				return fmt.Errorf("Save payment payload error.")
			}
		case modules.APP_CONTRACT_TPL:
			tpl := msg.Payload.(*modules.ContractTplPayload)
			if ok := rep.saveContractTpl(unit.UnitHeader.Number, uint32(txIndex), installReq, tpl); !ok {
				return fmt.Errorf("Save contract template error.")
			}
		case modules.APP_CONTRACT_DEPLOY:
			deploy := msg.Payload.(*modules.ContractDeployPayload)
			if ok := rep.saveContractInitPayload(unit.UnitHeader.Number, uint32(txIndex), templateId, deploy, requester, unitTime); !ok {
				return fmt.Errorf("Save contract init payload error.")
			}
		case modules.APP_CONTRACT_INVOKE:
			if ok := rep.saveContractInvokePayload(tx, unit.UnitHeader.Number, uint32(txIndex), msg, reqIndex); !ok {
				return fmt.Errorf("save contract invode payload error")
			}
		case modules.APP_CONTRACT_STOP:
			if ok := rep.saveContractStop(reqId, msg); !ok {
				return fmt.Errorf("save contract stop payload failed.")
			}
		case modules.APP_ACCOUNT_UPDATE:
			if err := rep.updateAccountInfo(msg, requester, unit.UnitHeader.Number, uint32(txIndex)); err != nil {
				return fmt.Errorf("apply Account Updating Operation error")
			}
		case modules.APP_CONTRACT_TPL_REQUEST:
			installReq = msg.Payload.(*modules.ContractInstallRequestPayload)
		case modules.APP_CONTRACT_DEPLOY_REQUEST:
			if ok := rep.saveContractDeployReq(reqId, msg); !ok {
				return fmt.Errorf("save contract of deploy request failed.")
			}
			deployReq := msg.Payload.(*modules.ContractDeployRequestPayload)
			templateId = deployReq.TemplateId
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
			if ok := rep.saveDataPayload(requester, unitHash, unitHeight, unitTime, txHash,
				msg.Payload.(*modules.DataPayload)); !ok {
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
	//Index
	if dagconfig.DagConfig.AddrTxsIndex {
		rep.saveAddrTxIndex(txHash, tx)
	}
	return nil
}
func (rep *UnitRepository) saveAddrTxIndex(txHash common.Hash, tx *modules.Transaction) {

	//Index TxId for to address
	addresses := rep.getPayToAddresses(tx)
	for _, addr := range addresses {
		rep.idxdb.SaveAddressTxId(addr, txHash)
	}
	//Index from address to txid
	fromAddrs := rep.getPayFromAddresses(tx)
	for _, addr := range fromAddrs {
		rep.idxdb.SaveAddressTxId(addr, txHash)
	}
}

func (rep *UnitRepository) getPayToAddresses(tx *modules.Transaction) []common.Address {
	resultMap := map[common.Address]int{}
	for _, msg := range tx.TxMessages {
		if msg.App == modules.APP_PAYMENT {
			pay := msg.Payload.(*modules.PaymentPayload)
			for _, out := range pay.Outputs {
				addr, _ := rep.tokenEngine.GetAddressFromScript(out.PkScript)
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
					var lockScript []byte
					utxo, err := rep.utxoRepository.GetUtxoEntry(input.PreviousOutPoint)
					if err == nil {
						lockScript = utxo.PkScript
					}
					if err != nil {
						stxo, err := rep.utxoRepository.GetStxoEntry(input.PreviousOutPoint)
						if err != nil {
							if input.PreviousOutPoint.TxHash.IsSelfHash() {
								out := tx.TxMessages[input.PreviousOutPoint.MessageIndex].Payload.(*modules.PaymentPayload).Outputs[input.PreviousOutPoint.OutIndex]
								lockScript = out.PkScript
							} else {
								log.Errorf("Cannot find txo by:%s", input.PreviousOutPoint.String())
								return []common.Address{}
							}
						} else {
							lockScript = stxo.PkScript
						}
					}
					addr, _ := rep.tokenEngine.GetAddressFromScript(lockScript)
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

func (rep *UnitRepository) RebuildAddrTxIndex() error {
	log.Info("Star rebuild address tx index. truncate old index data...")
	rep.idxdb.TruncateAddressTxIds()
	i := 0
	err := rep.dagdb.ForEachAllTxDo(func(key []byte, tx *modules.Transaction) error {
		txHash := common.BytesToHash(key[2:])
		//TODO Devin检查Genesis Tx的Hash问题
		if txHash.String() == "0xccbb34cecf684c58ea2c44f37ef491ac40efb5cdf7952d52002a18c8ea47210c" {
			log.Warnf("tx[0xccbb34cecf684c58ea2c44f37ef491ac40efb5cdf7952d52002a18c8ea47210c],key:%x", key)
			return errors.ErrInvalidNumber
		}
		rep.saveAddrTxIndex(txHash, tx)
		i++
		if i%1000 == 0 {
			log.Infof("Build address tx index:%d", i)
		}
		return nil
	})
	if err != nil {
		return err
	}
	log.Info("Rebuild address tx index complete.")
	return nil
}

func getDataPayload(tx *modules.Transaction) *modules.DataPayload {
	dp := &modules.DataPayload{}
	for _, msg := range tx.TxMessages {
		if msg.App == modules.APP_DATA {
			pay := msg.Payload.(*modules.DataPayload)
			dp.MainData = pay.MainData
			dp.ExtraData = pay.ExtraData
			dp.Reference = pay.Reference
			return pay
		}
	}
	return nil
}

/**
保存PaymentPayload
save PaymentPayload data
*/
func (rep *UnitRepository) savePaymentPayload(unitTime int64, txHash common.Hash, msg *modules.PaymentPayload,
	msgIndex uint32) bool {
	// if inputs is none then it is just a normal coinbase transaction
	// otherwise, if inputs' length is 1, and it PreviousOutPoint should be none
	// if this is a create token transaction, the Extra field should be AssetInfo struct's [rlp] encode bytes
	// if this is a create token transaction, should be return a assetid
	// save utxo
	err := rep.utxoRepository.UpdateUtxo(unitTime, txHash, msg, msgIndex)
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

func (rep *UnitRepository) saveDataPayload(requester common.Address, unitHash common.Hash, unitHeight uint64,
	timestamp int64, txHash common.Hash, dataPayload *modules.DataPayload) bool {

	if dagconfig.DagConfig.TextFileHashIndex {

		err := rep.idxdb.SaveMainDataTxId(dataPayload.MainData, txHash)
		if err != nil {
			log.Error("error savefilehash", "err", err)
			return false
		}
		if len(dataPayload.Reference) > 0 {
			poe := &modules.ProofOfExistence{
				MainData:   dataPayload.MainData,
				ExtraData:  dataPayload.ExtraData,
				Reference:  dataPayload.Reference,
				UnitHash:   unitHash,
				UintHeight: unitHeight,
				Creator:    requester,
				TxId:       txHash,
				Timestamp:  uint64(timestamp),
			}
			err = rep.idxdb.SaveProofOfExistence(poe)
			if err != nil {
				log.Error("error SaveProofOfExistence", "err", err)
				return false
			}
			//for _, output := range msg.Outputs {
			//	asset := output.Asset
			//	if asset.AssetId.GetAssetType() == modules.AssetType_NonFungibleToken {
			//		if err = rep.idxdb.SaveTokenExistence(asset, poe); err != nil {
			//			log.Errorf("Save token and ProofOfExistence index data error:%s", err.Error())
			//		}
			//	}
			//
			//}
			//err = rep.idxdb.SaveTokenExistence()
		}
		return true
	}
	log.Debug("dagconfig textfileindex is false, don't build index for data")
	return true
}

/**
保存合约调用状态
To save contract invoke state
*/
func (rep *UnitRepository) saveContractInvokePayload(tx *modules.Transaction, height *modules.ChainIndex,
	txIndex uint32, msg *modules.Message, reqIndex int) bool {
	pl := msg.Payload
	payload, ok := pl.(*modules.ContractInvokePayload)
	if !ok {
		return false
	}

	version := &modules.StateVersion{
		Height:  height,
		TxIndex: txIndex,
	}

	err := rep.statedb.SaveContractStates(payload.ContractId, payload.WriteSet, version)
	if err != nil {
		log.Errorf("Tx[%s]Write contract state error:%s", tx.Hash().String(), err.Error())
		return false
	}

	if common.IsSystemContractAddress(payload.ContractId) && payload.ErrMsg.Code == 0 {
		eventArg := &modules.SysContractStateChangeEvent{ContractId: payload.ContractId, WriteSet: payload.WriteSet}
		for _, eventFunc := range rep.observers {
			eventFunc(eventArg)
		}

		// append by albert
		if reqIndex != -1 { // 排除创世交易中的系统合约交易没有Request的情况
			invoke, _ := tx.TxMessages[reqIndex].Payload.(*modules.ContractInvokeRequestPayload)
			rep.statedb.UpdateStateByContractInvoke(invoke)
		}
	}

	return true
}

/**
保存合约初始化状态
To save contract init state
*/
func (rep *UnitRepository) saveContractInitPayload(height *modules.ChainIndex, txIndex uint32, templateId []byte,
	payload *modules.ContractDeployPayload, requester common.Address, unitTime int64) bool {
	//编译源码时，发生错误信息，但是此时因为还没有构建chaincode容器，所以导致contractId为空
	if payload.ContractId == nil {
		log.Infof("source codes go build error")
		return true
	}
	// save contract state
	version := &modules.StateVersion{
		Height:  height,
		TxIndex: txIndex,
	}
	if len(payload.WriteSet) > 0 {
		err := rep.statedb.SaveContractStates(payload.ContractId, payload.WriteSet, version)
		if err != nil {
			log.Errorf("save contract[%x] init writeset error:%s", payload.ContractId, err.Error())
			return false
		}
	}
	contract := modules.NewContract(templateId, payload, requester, uint64(unitTime))
	err := rep.statedb.SaveContract(contract)
	if err != nil {
		log.Errorf("Save contract[%x] error:%s", payload.ContractId, err.Error())
		return false
	}
	if len(payload.EleNode.EleList) > 0 {
		//save contract election
		err = rep.statedb.SaveContractJury(payload.ContractId, payload.EleNode, version)
		if err != nil {
			log.Errorf("Save jury for contract[%x] error:%s", payload.ContractId, err.Error())
			return false
		}
	}
	return true
}

/**
保存合约模板代码
To save contract template code
*/
func (rep *UnitRepository) saveContractTpl(height *modules.ChainIndex, txIndex uint32,
	installReq *modules.ContractInstallRequestPayload, tpl *modules.ContractTplPayload) bool {

	template := modules.NewContractTemplate(installReq, tpl)
	err := rep.statedb.SaveContractTpl(template)
	if err != nil {
		log.Errorf("Save contract template fail,height:%s,txIndex:%d,error:%s", height.String(), txIndex, err.Error())
		return false
	}
	err = rep.statedb.SaveContractTplCode(tpl.TemplateId, tpl.ByteCode)
	if err != nil {
		log.Errorf("Save contract code fail,error:%s", err.Error())
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
	//  合约停止后，修改该合约的状态
	contract, err := rep.statedb.GetContract(stop.ContractId)
	if err != nil {
		log.Info("get contract with id failed,", "error", err)
		return false
	}
	contract.Status = 0
	err = rep.statedb.SaveContract(contract)
	if err != nil {
		log.Info("save contract with id failed,", "error", err)
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
func (rep *UnitRepository) CreateCoinbase(ads []*modules.Addition, height uint64) (
	*modules.Transaction, uint64, error) {
	log.DebugDynamic(func() string {
		data, _ := json.Marshal(ads)
		return "Try to create coinbase for fee allocation:" + string(data)
	})

	if height%parameter.CurrentSysParameters.RewardHeight == 0 {
		return rep.createCoinbasePayment(ads)
	} else {
		return rep.createCoinbaseState(ads)
	}
}
func (rep *UnitRepository) createCoinbaseState(ads []*modules.Addition) (*modules.Transaction, uint64, error) {
	//log.Debug("create a statedb record to write mediator and jury income")
	totalIncome := uint64(0)
	payload := modules.ContractInvokePayload{}
	contractId := syscontract.CoinbaseContractAddress.Bytes()
	payload.ContractId = contractId
	//在Coinbase合约的StateDB中保存每个Mediator和Jury的奖励值，
	//key为奖励地址，Value为[]AmountAsset
	if len(ads) != 0 {
		for _, v := range ads {
			key := constants.RewardAddressPrefix + v.Addr.String()
			data, version, err := rep.statedb.GetContractState(contractId, key)
			income := []modules.AmountAsset{}
			if err == nil { //之前有奖励
				rlp.DecodeBytes(data, &income)
				rs := modules.ContractReadSet{Key: key, Version: version}
				payload.ReadSet = append(payload.ReadSet, rs)
				log.DebugDynamic(func() string {
					jsdata, _ := json.Marshal(income)
					return "Get history reward for key:" + key + " Value:" + string(jsdata)
				})
			} else {
				log.Debugf("%s Don't have history reward create it.", key)
			}
			newValue := addIncome(income, v.Amount, v.Asset)
			newData, _ := rlp.EncodeToBytes(newValue)
			log.DebugDynamic(func() string {
				jsdata, _ := json.Marshal(newValue)
				return "Create coinbase write set for key:" + key + " Value:" + string(jsdata)
			})
			ws := modules.ContractWriteSet{IsDelete: false, Key: key, Value: newData}
			payload.WriteSet = append(payload.WriteSet, ws)
			totalIncome += v.Amount
		}
	}
	msg := &modules.Message{
		App:     modules.APP_CONTRACT_INVOKE,
		Payload: &payload,
	}
	coinbase := new(modules.Transaction)
	coinbase.TxMessages = append(coinbase.TxMessages, msg)
	return coinbase, totalIncome, nil
}
func addIncome(income []modules.AmountAsset, newAmount uint64, asset *modules.Asset) []modules.AmountAsset {
	newValue := []modules.AmountAsset{}
	hasOldValue := false
	for _, aa := range income {
		if aa.Asset.Equal(asset) {
			aa.Amount += newAmount
			hasOldValue = true
		}
		newValue = append(newValue, aa)
	}
	if !hasOldValue {
		newValue = append(newValue, modules.AmountAsset{Amount: newAmount, Asset: asset})
	}
	return newValue
}

func (rep *UnitRepository) createCoinbasePayment(ads []*modules.Addition) (*modules.Transaction, uint64, error) {
	log.Debug("create a payment to reward mediator and jury")
	totalIncome := uint64(0)

	contractId := syscontract.CoinbaseContractAddress.Bytes()

	//在Coinbase合约的StateDB中保存每个Mediator和Jury的奖励值，
	//key为奖励地址，Value为[]AmountAsset
	//读取之前的奖励统计值
	addrMap, err := rep.statedb.GetContractStatesByPrefix(contractId, constants.RewardAddressPrefix)
	if err != nil {
		log.Errorf("GetContractStates(%v) By Prefix(%v) is error",
			syscontract.CoinbaseContractAddress.Str(), constants.RewardAddressPrefix)
		return nil, 0, err
	}
	rewards := map[common.Address][]modules.AmountAsset{}
	for key, v := range addrMap {
		addr := key[len(constants.RewardAddressPrefix):]
		incomeAddr, _ := common.StringToAddress(addr)
		aa := []modules.AmountAsset{}
		rlp.DecodeBytes(v.Value, &aa)
		if len(aa) > 0 {
			rewards[incomeAddr] = aa
		}
	}
	//附加最新的奖励
	for _, ad := range ads {

		reward, ok := rewards[ad.Addr]
		if !ok {
			reward = []modules.AmountAsset{}
		}
		reward = addIncome(reward, ad.Amount, ad.Asset)
		rewards[ad.Addr] = reward
		totalIncome += ad.Amount
	}
	//所有奖励转换成PaymentPayload
	msg := rep.createCoinbasePaymentMsg(rewards)
	coinbase := new(modules.Transaction)
	coinbase.TxMessages = append(coinbase.TxMessages, msg)
	//清空历史奖励的记账值
	payload := &modules.ContractInvokePayload{}
	payload.ContractId = contractId
	empty, _ := rlp.EncodeToBytes([]modules.AmountAsset{})
	for addr := range rewards {
		key := constants.RewardAddressPrefix + addr.String()
		_, version, _ := rep.statedb.GetContractState(contractId, key)
		rs := modules.ContractReadSet{Key: key, Version: version}
		payload.ReadSet = append(payload.ReadSet, rs)

		ws := modules.ContractWriteSet{IsDelete: false, Key: key, Value: empty}
		payload.WriteSet = append(payload.WriteSet, ws)
	}
	msg1 := &modules.Message{
		App:     modules.APP_CONTRACT_INVOKE,
		Payload: payload,
	}
	coinbase.TxMessages = append(coinbase.TxMessages, msg1)
	return coinbase, totalIncome, nil
}
func (rep *UnitRepository) createCoinbasePaymentMsg(rewards map[common.Address][]modules.AmountAsset) *modules.Message {
	coinbasePayment := &modules.PaymentPayload{}
	for addr, v := range rewards {
		script := rep.tokenEngine.GenerateLockScript(addr)
		for _, reward := range v {
			additionalOutput := modules.Output{
				Value:    reward.Amount,
				Asset:    reward.Asset,
				PkScript: script,
			}
			coinbasePayment.Outputs = append(coinbasePayment.Outputs, &additionalOutput)
		}
	}
	msg := &modules.Message{
		App:     modules.APP_PAYMENT,
		Payload: coinbasePayment,
	}
	return msg
}

/**
删除合约状态
To delete contract state
*/
//func (rep *UnitRepository) deleteContractState(contractID []byte, field string) {
//	oldKeyPrefix := fmt.Sprintf("%s%s^*^%s",
//		constants.CONTRACT_STATE_PREFIX,
//		hexutil.Encode(contractID[:]),
//		field)
//	data := rep.statedb.GetPrefix([]byte(oldKeyPrefix))
//	for k := range data {
//		if err := rep.statedb.DeleteState([]byte(k)); err != nil {
//			log.Error("Delete contract state", "error", err.Error())
//			continue
//		}
//	}
//}

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

// GetAddrTransactions containing from && to address
func (rep *UnitRepository) GetAddrTransactions(address common.Address) ([]*modules.TransactionWithUnitInfo, error) {
	log.Debug("getAddrTxs unitRepository lock.")
	rep.lock.RLock()
	defer log.Debug("getAddrTxs unitRepository unlock.")
	defer rep.lock.RUnlock()
	hashs, err := rep.idxdb.GetAddressTxIds(address)
	if err != nil {
		return nil, err
	}
	txs := make([]*modules.TransactionWithUnitInfo, 0)
	for _, hash := range hashs {
		tx, err := rep.GetTransaction(hash)
		if err != nil {
			return nil, err
		}
		txs = append(txs, tx)
	}
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
	mds := make([]*modules.FileInfo, 0)
	for _, hash := range hashs {
		var md modules.FileInfo

		tx, err := rep.GetTransaction(hash)
		if err != nil {
			return nil, err
		}
		dp := getDataPayload(tx.Transaction)
		md.MainData = string(dp.MainData)
		md.ExtraData = string(dp.ExtraData)
		md.Reference = string(dp.Reference)
		md.UnitHash = tx.UnitHash
		md.UintHeight = tx.UnitIndex
		md.Txid = tx.Hash()
		md.Timestamp = tx.Timestamp
		mds = append(mds, &md)
	}
	return mds, nil
}

func (rep *UnitRepository) GetLastIrreversibleUnit(assetID modules.AssetId) (*modules.Unit, error) {
	log.Debug("GetLastIrreversibleUnit unitRepository lock.")
	rep.lock.RLock()
	defer log.Debug("GetLastIrreversibleUnit unitRepository unlock.")
	defer rep.lock.RUnlock()
	hash, _, _, err := rep.propdb.GetNewestUnit(assetID)
	if err != nil {
		return nil, err
	}
	return rep.getUnit(hash)
}

func (rep *UnitRepository) GetTxFromAddress(tx *modules.Transaction) ([]common.Address, error) {
	log.Debug("GetTxFromAddress unitRepository lock.")
	rep.lock.RLock()
	defer log.Debug("GetTxFromAddress unitRepository unlock.")
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
					addr, _ := rep.tokenEngine.GetAddressFromScript(utxo.PkScript)
					result = append(result, addr)
				}
			}
		}
	}
	return result, nil
}
func (rep *UnitRepository) RefreshAddrTxIndex() error {
	log.Debugf("RefreshAddrTxIndex unitRepository lock.")
	rep.lock.RLock()
	defer func() {
		rep.lock.RUnlock()
		log.Debug("RefreshAddrTxIndex unitRepository unlock.")
	}()
	if !dagconfig.DagConfig.AddrTxsIndex {
		return errors.New("Please enable AddrTxsIndex in toml DagConfig")
	}
	txs, err := rep.dagdb.GetAllTxs()
	if err != nil {
		return err
	}
	for _, tx := range txs {
		rep.saveAddrTxIndex(tx.Hash(), tx)
	}
	return nil
}

func (rep *UnitRepository) GetAssetReference(asset []byte) ([]*modules.ProofOfExistence, error) {
	return rep.idxdb.QueryProofOfExistenceByReference(asset)
}

func (rep *UnitRepository) QueryProofOfExistenceByReference(ref []byte) ([]*modules.ProofOfExistence, error) {
	return rep.idxdb.QueryProofOfExistenceByReference(ref)
}
