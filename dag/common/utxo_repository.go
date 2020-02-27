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
	"errors"
	"unsafe"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/parameter"
	"github.com/palletone/go-palletone/dag/storage"
	"github.com/palletone/go-palletone/tokenengine"
)

type UtxoRepository struct {
	txUtxodb    storage.IUtxoDb
	reqUtxodb   storage.IUtxoDb
	idxdb       storage.IIndexDb
	statedb     storage.IStateDb
	propDb      storage.IPropertyDb
	tokenEngine tokenengine.ITokenEngine
}

func NewUtxoRepository(txutxodb storage.IUtxoDb, requtxodb storage.IUtxoDb, idxdb storage.IIndexDb,
	statedb storage.IStateDb, propDb storage.IPropertyDb,
	tokenEngine tokenengine.ITokenEngine) *UtxoRepository {
	return &UtxoRepository{
		txUtxodb:    txutxodb,
		reqUtxodb:   requtxodb,
		idxdb:       idxdb,
		statedb:     statedb,
		propDb:      propDb,
		tokenEngine: tokenEngine,
	}
}
func NewUtxoRepository4Db(db ptndb.Database, tokenEngine tokenengine.ITokenEngine) *UtxoRepository {
	requtxodb := storage.NewUtxoDb(db, tokenEngine, true)
	txutxodb := storage.NewUtxoDb(db, tokenEngine, false)
	statedb := storage.NewStateDb(db)
	idxdb := storage.NewIndexDb(db)
	propDb := storage.NewPropertyDb(db)

	return &UtxoRepository{
		txUtxodb:    txutxodb,
		reqUtxodb:   requtxodb,
		idxdb:       idxdb,
		statedb:     statedb,
		propDb:      propDb,
		tokenEngine: tokenEngine,
	}
}

type IUtxoRepository interface {
	GetUtxoEntry(outpoint *modules.OutPoint) (*modules.Utxo, error)
	GetStxoEntry(outpoint *modules.OutPoint) (*modules.Stxo, error)
	GetTxOutput(outpoint *modules.OutPoint) (*modules.Utxo, error)
	GetAllUtxos() (map[modules.OutPoint]*modules.Utxo, error)
	GetAddrOutpoints(addr common.Address) ([]modules.OutPoint, error)
	GetAddrUtxos(addr common.Address, asset *modules.Asset) (map[modules.OutPoint]*modules.Utxo, error)
	GetAddrUtxoAndReqMapping(addr common.Address, asset *modules.Asset) (map[modules.OutPoint]*modules.Utxo, map[common.Hash]common.Hash, error)
	GetUxto(txin modules.Input) *modules.Utxo
	UpdateUtxo(unitTime int64, txHash, reqHash common.Hash, payment *modules.PaymentPayload, msgIndex uint32) error
	IsUtxoSpent(outpoint *modules.OutPoint) (bool, error)
	//ComputeTxFee(tx *modules.Transaction) (*modules.AmountAsset, error)
	GetUxtoSetByInputs(txins []modules.Input) (map[modules.OutPoint]*modules.Utxo, uint64)
	//GetAccountTokens(addr common.Address) (map[string]*modules.AccountToken, error)
	//WalletBalance(addr common.Address, asset modules.Asset) uint64
	// ComputeAwards(txs []*txspool.TxPoolTransaction, dagdb storage.IDagDb) (*modules.Addition, error)
	// ComputeTxAward(tx *modules.Transaction, dagdb storage.IDagDb) (uint64, error)
	ClearUtxo() error
	ClearAddrUtxo(addr common.Address) error
	SaveUtxoView(view map[modules.OutPoint]*modules.Utxo) error
	//SaveUtxoEntity(outpoint *modules.OutPoint, utxo *modules.Utxo) error
}

func (repository *UtxoRepository) getUtxoEntry(outpoint *modules.OutPoint) (*modules.Utxo, error) {
	data, err := repository.txUtxodb.GetUtxoEntry(outpoint)
	if err == nil {
		return data, nil
	}
	log.Debugf("GetUtxoEntry(%s) not in TxUtxo, try ReqUtxo", outpoint.String())
	//Tx UTXO找不到，试着去Req UTXO 找
	return repository.reqUtxodb.GetUtxoEntry(outpoint)
}
func (repository *UtxoRepository) GetUtxoEntry(outpoint *modules.OutPoint) (*modules.Utxo, error) {
	data, err := repository.getUtxoEntry(outpoint)
	if err != nil {
		log.Warnf("GetUtxoEntry(%s) also not in ReqUtxo", outpoint.String())
	} else {
		log.Debugf("GetUtxoEntry(%s) from ReqUtxo, return value", outpoint.String())
	}
	return data, err
	//mapHash,err:= repository.txUtxodb.GetRequestAndTxMapping(outpoint.TxHash)
	//if err != nil {//找不到Mapping
	//	log.Warnf("retrieve utxo[%s] get error:%s", outpoint.String(), err.Error())
	//	return nil,err
	//}
	////去Req UTXO表重新查找
	//outpoint2:=modules.NewOutPoint(mapHash,outpoint.MessageIndex,outpoint.OutIndex)
	//log.DebugDynamic(func() string {
	//	return fmt.Sprintf("try to retrieve utxo by new outpoint:[%s],old outpoint:%s",
	//		outpoint2.String(),outpoint.String())
	//})
	//return repository.reqUtxodb.GetUtxoEntry(outpoint2)
}
func (repository *UtxoRepository) GetStxoEntry(outpoint *modules.OutPoint) (*modules.Stxo, error) {
	data, err := repository.txUtxodb.GetStxoEntry(outpoint)
	if err == nil {
		return data, nil
	}
	//Tx STXO找不到，试着去Req STXO 找
	return repository.reqUtxodb.GetStxoEntry(outpoint)
	//mapHash,err:= repository.txUtxodb.GetRequestAndTxMapping(outpoint.TxHash)
	//if err != nil {//找不到Mapping
	//	log.Warnf("retrieve utxo[%s] get error:%s", outpoint.String(), err.Error())
	//	return nil,err
	//}
	////去Req STXO表重新查找
	//outpoint2:=modules.NewOutPoint(mapHash,outpoint.MessageIndex,outpoint.OutIndex)
	//log.DebugDynamic(func() string {
	//	return fmt.Sprintf("try to retrieve stxo by new outpoint:[%s],old outpoint:%s",
	//		outpoint2.String(),outpoint.String())
	//})
	//return repository.reqUtxodb.GetStxoEntry(outpoint2)
}

//获得消费的和未消费的交易输出
func (repository *UtxoRepository) GetTxOutput(outpoint *modules.OutPoint) (*modules.Utxo, error) {
	utxo, err := repository.getUtxoEntry(outpoint)
	if err != nil {
		stxo, err := repository.GetStxoEntry(outpoint)
		if err != nil {
			return nil, err
		}
		return &modules.Utxo{
			Amount:    stxo.Amount,
			Asset:     stxo.Asset,
			PkScript:  stxo.PkScript,
			LockTime:  stxo.LockTime,
			Timestamp: stxo.Timestamp,
			Flags:     2,
		}, nil
	}
	return utxo, nil
}

func (repository *UtxoRepository) IsUtxoSpent(outpoint *modules.OutPoint) (bool, error) {
	return repository.txUtxodb.IsUtxoSpent(outpoint)
}
func (repository *UtxoRepository) GetAllUtxos() (map[modules.OutPoint]*modules.Utxo, error) {
	return repository.txUtxodb.GetAllUtxos()
}
func (repository *UtxoRepository) GetAddrOutpoints(addr common.Address) ([]modules.OutPoint, error) {
	return repository.txUtxodb.GetAddrOutpoints(addr)
}

//活动一个地址的所有UTXO，包括完整交易的和Request的，存证交叉，需要通过Mapping过滤
func (repository *UtxoRepository) GetAddrUtxos(addr common.Address, asset *modules.Asset) (
	map[modules.OutPoint]*modules.Utxo, error) {
	utxo1, err := repository.txUtxodb.GetAddrUtxos(addr, asset)
	if err != nil {
		return nil, err
	}
	mappingHashs := make(map[common.Hash]bool)
	for o := range utxo1 {
		mappingHash, err := repository.txUtxodb.GetRequestAndTxMapping(o.TxHash)
		if err == nil {
			mappingHashs[mappingHash] = true
		}
	}
	utxo2, err := repository.reqUtxodb.GetAddrUtxos(addr, asset)
	if err != nil {
		return nil, err
	}
	for o, u := range utxo2 {
		if _, has := mappingHashs[o.TxHash]; !has {
			utxo1[o] = u
		}
	}
	return utxo1, nil
}

//返回一个地址的TxUtxo和该ReqHash对应的TxHash
func (repository *UtxoRepository) GetAddrUtxoAndReqMapping(addr common.Address, asset *modules.Asset) (
	map[modules.OutPoint]*modules.Utxo, map[common.Hash]common.Hash, error) {
	utxo1, err := repository.txUtxodb.GetAddrUtxos(addr, asset)
	if err != nil {
		return nil, nil, err
	}
	mappingHashs := make(map[common.Hash]common.Hash)
	for o := range utxo1 {
		mappingHash, err := repository.txUtxodb.GetRequestAndTxMapping(o.TxHash)
		if err == nil {
			mappingHashs[mappingHash] = o.TxHash
		}
	}

	return utxo1, mappingHashs, nil
}
func (repository *UtxoRepository) SaveUtxoView(view map[modules.OutPoint]*modules.Utxo) error {
	return repository.txUtxodb.SaveUtxoView(view)
}

//func (repository *UtxoRepository) SaveUtxoEntity(outpoint *modules.OutPoint, utxo *modules.Utxo) error {
//	return repository.txUtxodb.SaveUtxoEntity(outpoint, utxo)
//}
func (repository *UtxoRepository) ClearUtxo() error {
	err := repository.txUtxodb.ClearUtxo()
	if err != nil {
		return err
	}
	return repository.reqUtxodb.ClearUtxo()
}
func (repository *UtxoRepository) ClearAddrUtxo(addr common.Address) error {
	err := repository.txUtxodb.ClearAddrUtxo(addr)
	if err != nil {
		return err
	}
	return repository.reqUtxodb.ClearAddrUtxo(addr)
}

/**
根据用户地址、选择的资产类型、转账金额、手续费返回utxo
Return utxo struct and his total amount according to user's address, asset type
*/
//func (repository *UtxoRepository) ReadUtxos(addr common.Address, asset modules.Asset) (
// map[modules.OutPoint]*modules.Utxo, uint64) {
//
//	if dagconfig.DagConfig.UtxoIndex {
//		return repository.readUtxosByIndex(addr, asset)
//	} else {
//		return repository.readUtxosFrAll(addr, asset)
//	}
//}

/**
根据UTXO索引数据库读取UTXO信息
To get utxo info by utxo index db
*/
//func (repository *UtxoRepository) readUtxosByIndex(addr common.Address, asset modules.Asset) (
// map[modules.OutPoint]*modules.Utxo, uint64) {
//	// step1. read outpoint from utxo index
//	utxoIndex := modules.UtxoIndex{
//		AccountAddr: addr,
//		Asset:       &asset,
//	}
//	key := utxoIndex.AssetKey()
//	data := repository.txUtxodb.GetPrefix([]byte(key))
//	if data == nil {
//		return nil, 0
//	}
//	// step2. get utxo
//	vout := map[modules.OutPoint]*modules.Utxo{}
//	balance := uint64(0)
//	for k := range data {
//		if err := utxoIndex.QueryFields([]byte(k)); err != nil {
//			continue
//		}
//		utxo, err := repository.txUtxodb.GetUtxoEntry(utxoIndex.OutPoint)
//		if err != nil {
//			log.Error("Get utxo error by outpoint", "error:", err.Error())
//			continue
//		}
//		//var utxo modules.Utxo
//		//if err := rlp.DecodeBytes([]byte(udata), &utxo); err != nil {
//		//	log.Error("Decode utxo data :", "error:", err.Error())
//		//	continue
//		//}
//		vout[*utxoIndex.OutPoint] = utxo
//		balance += utxo.Amount
//	}
//	return vout, balance
//}

/**
扫描UTXO全表，获取对应账户的指定token可用的utxo列表
To get utxo info by scanning all utxos.
*/
//func (repository *UtxoRepository) readUtxosFrAll(addr common.Address, asset modules.Asset) (
// map[modules.OutPoint]*modules.Utxo, uint64) {
//	// key: [UTXO_PREFIX][addr]_[asset]_[msgindex]_[out index]
//	data := repository.txUtxodb.GetPrefix(constants.UTXO_PREFIX)
//	if data == nil {
//		return nil, 0
//	}
//
//	// value: utxo rlp bytes
//	vout := map[modules.OutPoint]*modules.Utxo{}
//	var balance uint64
//	balance = 0
//	for k, v := range data {
//		var utxo modules.Utxo
//		if err := rlp.DecodeBytes([]byte(v), &utxo); err != nil {
//			log.Error("Decode utxo data :", "error:", err.Error())
//			continue
//		}
//		// check asset
//		if strings.Compare(asset.AssetId.String(), utxo.Asset.AssetId.String()) != 0 ||
//			strings.Compare(asset.UniqueId.String(), utxo.Asset.UniqueId.String()) != 0 {
//			continue
//		}
//		// get addr
//		sAddr, _ := tokenengine.GetAddressFromScript(utxo.PkScript)
//		// check address
//		if strings.Compare(sAddr.String(), addr.String()) != 0 {
//			continue
//		}
//		outpoint := modules.KeyToOutpoint([]byte(k))
//		vout[*outpoint] = &utxo
//		balance += utxo.Amount
//	}
//
//	return vout, balance
//}

/**
获取某个input对应的utxo信息
To get utxo struct according to it's input information
*/
func (repository *UtxoRepository) GetUxto(txin modules.Input) *modules.Utxo {
	if txin.PreviousOutPoint == nil {
		return nil
	}
	data, err := repository.GetUtxoEntry(txin.PreviousOutPoint)
	if err != nil {
		return nil
	}
	return data
}

// GetUtosOutPoint
//func (repository *UtxoRepository) GetUtxoByOutpoint(outpoint *modules.OutPoint) (*modules.Utxo, error) {
//	return repository.txUtxodb.GetUtxoEntry(outpoint)
//}

/**
根据交易信息中的outputs创建UTXO， 根据交易信息中的inputs销毁UTXO
To create utxo according to outpus in transaction, and destroy utxo according to inputs in transaction
*/
func (repository *UtxoRepository) UpdateUtxo(unitTime int64, txHash, reqHash common.Hash,
	payment *modules.PaymentPayload, msgIndex uint32) error {
	// update utxo
	delBy := txHash
	if txHash.IsZero() {
		delBy = reqHash
	}
	err := repository.destroyUtxo(delBy, uint64(unitTime), payment.Inputs)
	if err != nil {
		return err
	}
	// create utxo
	errs := repository.writeUtxo(unitTime, txHash, reqHash, msgIndex, payment.Outputs, payment.LockTime)
	if len(errs) > 0 {
		log.Error("error occurred on updated utxos, check the log file to find details.")
		return errors.New("error occurred on updated utxos, check the log file to find details.")
	}

	return nil

}

/**
创建UTXO，3种情况：
1. TxHash zero, ReqHash value	不完整的交易，只有Request，主要用于客户端的连续交易的生成
2. TxHash value, ReqHash value	完整的合约交易
3. TxHash value, ReqHash zero	非合约交易
*/
func (repository *UtxoRepository) writeUtxo(unitTime int64, txHash, reqHash common.Hash,
	msgIndex uint32, txouts []*modules.Output, lockTime uint32) []error {
	log.Debugf("try to write new utxo for tx[%s],req[%s]", txHash.String(), reqHash.String())
	var errs []error
	for outIndex, txout := range txouts {
		sAddr, _ := repository.tokenEngine.GetAddressFromScript(txout.PkScript)
		if sAddr == common.DestroyAddress { //销毁地址，不产生UTXO
			continue
		}
		utxo := &modules.Utxo{
			Amount:    txout.Value,
			Asset:     txout.Asset,
			PkScript:  txout.PkScript,
			LockTime:  lockTime,
			Timestamp: uint64(unitTime),
		}

		// write to database
		outpoint := &modules.OutPoint{
			TxHash:       txHash,
			MessageIndex: msgIndex,
			OutIndex:     uint32(outIndex),
		}
		//1 保存TxHash的UTXO
		if !txHash.IsZero() {
			if err := repository.txUtxodb.SaveUtxoEntity(outpoint, utxo); err != nil {
				log.Error("Write utxo", "error", err.Error())
				errs = append(errs, err)
				continue
			}
		}
		//2 Hash都不为空，保存映射关系
		if !txHash.IsZero() && !reqHash.IsZero() {
			err := repository.txUtxodb.SaveRequestAndTxHash(reqHash, txHash)
			if err != nil {
				errs = append(errs, err)
				continue
			}
		}
		//3 保存ReqHash的UTXO
		if !reqHash.IsZero() {
			outpoint.TxHash = reqHash
			if err := repository.reqUtxodb.SaveUtxoEntity(outpoint, utxo); err != nil {
				log.Error("Write utxo", "error", err.Error())
				errs = append(errs, err)
				continue
			}
			log.Debugf("reqHash has value, try to save request utxo:%s", outpoint.String())
		}

		//update address account info
		gasToken := dagconfig.DagConfig.GetGasToken()
		if txout.Asset.AssetId == gasToken {
			err := repository.statedb.UpdateAccountBalance(sAddr, int64(txout.Value))
			if err != nil {
				log.Error("UpdateAccountBalance", "error", err.Error())
				errs = append(errs, err)
			}
		}
	}
	return errs
}

/**
销毁utxo
destroy utxo, delete from UTXO database
txid： 被哪个tx销毁
unitTime：被销毁的时间
*/
func (repository *UtxoRepository) destroyUtxo(txid common.Hash, unitTime uint64, txins []*modules.Input) error {
	for _, txin := range txins {

		if txin == nil {
			continue
		}
		if txin.PreviousOutPoint == nil { //Coinbase
			continue
		}
		outpoint := txin.PreviousOutPoint.Clone()

		if outpoint.TxHash.IsSelfHash() { //TxHash为0，表示花费当前Tx产生的UTXO
			log.Debugf("Outpoint is zero:%s,set new txid:%s", outpoint.String(), txid.String())
			outpoint.TxHash = txid
		}
		// get utxo info
		utxo, err := repository.GetUtxoEntry(outpoint)
		if err != nil {
			log.Error("Query utxo when destroy uxto", "error", err.Error(), "outpoint", outpoint.String())
			return err
		}

		// delete utxo
		if err := repository.DeleteUtxo(outpoint, txid, unitTime); err != nil {
			log.Error("Delete uxto... ", "error", err.Error())
			return err
		}
		// delete index data
		sAddr, _ := repository.tokenEngine.GetAddressFromScript(utxo.PkScript)
		if utxo.Asset.AssetId == dagconfig.DagConfig.GetGasToken() { // modules.PTNCOIN
			err := repository.statedb.UpdateAccountBalance(sAddr, -int64(utxo.Amount))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

//当一个UTXO被使用后，先去TxUTXO进行删除，然后查找ReqUtxo，如果能查到，继续进行ReqUtxo的删除
func (rep *UtxoRepository) DeleteUtxo(outpoint *modules.OutPoint, spentTxId common.Hash, spentTime uint64) error {
	err1 := rep.txUtxodb.DeleteUtxo(outpoint, spentTxId, spentTime)
	if err1 != nil {
		err1 = rep.reqUtxodb.DeleteUtxo(outpoint, spentTxId, spentTime)
		if err1 != nil { //两次删除都失败了
			log.Warnf("delete utxo by [%s] error:%s", outpoint.String(), err1.Error())
			return err1
		}
	}
	//删除成功，试着去Req UTXO 找
	mapHash, err := rep.txUtxodb.GetRequestAndTxMapping(outpoint.TxHash)
	if err != nil { //找不到Mapping
		return nil
	}
	//找到了Mapping
	outpoint2 := modules.NewOutPoint(mapHash, outpoint.MessageIndex, outpoint.OutIndex)
	err2 := rep.txUtxodb.DeleteUtxo(outpoint2, spentTxId, spentTime)
	if err2 != nil {
		err2 = rep.reqUtxodb.DeleteUtxo(outpoint2, spentTxId, spentTime)
		if err2 != nil { //两次删除都失败了
			log.Warnf("delete utxo by [%s] error:%s", outpoint2.String(), err2.Error())
			return err2
		}
	}
	return nil
}

/**
存储Asset的信息
write asset info to leveldb
*/
//func (repository *UtxoRepository)SaveAssetInfo(assetInfo *modules.AssetInfo) error {
//
//	if err := storage.Store(db, string(assetInfo.Tokey()), *assetInfo); err != nil {
//		return err
//	}
//
//	return nil
//}

/**
根据assetid从数据库中获取asset的信息
get asset information from leveldb by assetid ( Asset struct type )
*/
//func (repository *UtxoRepository) GetAssetInfo(assetId *modules.Asset) (*modules.AssetInfo, error) {
//	return repository.statedb.GetAssetInfo(assetId)
//}

/**
获得某个账户下面的余额信息
To get balance by wallet address and his/her chosen asset type
*/
//func (repository *UtxoRepository) WalletBalance(addr common.Address, asset modules.Asset) uint64 {
//	if dagconfig.DagConfig.UtxoIndex {
//		return repository.walletBalanceByIndex(addr, asset)
//	} else {
//		return repository.walletBalanceFrAll(addr, asset)
//	}
//}

/**
通过索引数据库获得账户余额
To get balance by utxo index db
*/
//func (repository *UtxoRepository) walletBalanceByIndex(addr common.Address, asset modules.Asset) uint64 {
//	balance := uint64(0)
//
//	utxoIndex := modules.UtxoIndex{
//		AccountAddr: addr,
//		Asset:       &asset,
//	}
//	preKey := utxoIndex.AssetKey()
//
//	if data := repository.txUtxodb.GetPrefix([]byte(preKey)); data != nil {
//		for _, v := range data {
//			var utxoIndexVal modules.UtxoIndexValue
//			if err := rlp.DecodeBytes(v, &utxoIndexVal); err != nil {
//				log.Error("Decode utxo data error:", err)
//				continue
//			}
//			balance += utxoIndexVal.Amount
//		}
//	}
//
//	return balance
//
//}
//
///**
//通过查询全表获得账户余额
//To get balance by query all utxo table
//*/
//func (repository *UtxoRepository) walletBalanceFrAll(addr common.Address, asset modules.Asset) uint64 {
//	balance := uint64(0)
//
//	preKey := fmt.Sprintf("%s", constants.UTXO_PREFIX)
//
//	if data := repository.txUtxodb.GetPrefix([]byte(preKey)); data != nil {
//		for _, v := range data {
//			var utxo modules.Utxo
//			if err := rlp.DecodeBytes(v, &utxo); err != nil {
//				log.Error("Decode utxo data error:", err)
//				continue
//			}
//			if !checkUtxo(&addr, &asset, &utxo) {
//				continue
//			}
//			balance += utxo.Amount
//		}
//	}
//
//	return balance
//}

/**
根据payload中的inputs获得对应的UTXO map
*/
func (repository *UtxoRepository) GetUxtoSetByInputs(txins []modules.Input) (
	map[modules.OutPoint]*modules.Utxo, uint64) {
	utxos := map[modules.OutPoint]*modules.Utxo{}
	total := uint64(0)
	for _, in := range txins {
		utxo := repository.GetUxto(in)
		if unsafe.Sizeof(utxo) == 0 {
			continue
		}
		utxos[*in.PreviousOutPoint] = utxo
		total += utxo.Amount
	}
	return utxos, total
}

/**
获得某个账户下的token名称和assetid信息
To get someone account's list of tokens and those assetids
*/
//func (repository *UtxoRepository) GetAccountTokens(addr common.Address) (map[string]*modules.AccountToken, error) {
//	if dagconfig.DefaultConfig.UtxoIndex {
//		return repository.getAccountTokensByIndex(addr)
//	} else {
//		return repository.getAccountTokensWhole(addr)
//	}
//}

/**
通过索引数据库获得用户的token信息
To get account's token info by utxo index db
*/
//func (repository *UtxoRepository) getAccountTokensByIndex(addr common.Address) (
// map[string]*modules.AccountToken, error) {
//	tokens := map[string]*modules.AccountToken{}
//	utxoIndex := modules.UtxoIndex{
//		AccountAddr: addr,
//		Asset:       &modules.Asset{},
//		OutPoint:    &modules.OutPoint{},
//	}
//	data := repository.txUtxodb.GetPrefix(utxoIndex.AccountKey())
//	if data == nil || len(data) == 0 {
//		return nil, nil
//	}
//	for k, v := range data {
//		if err := utxoIndex.QueryFields([]byte(k)); err != nil {
//			return nil, fmt.Errorf("Get account tokens by key error: data key is invalid(%s)", err.Error())
//		}
//		var utxoIndexVal modules.UtxoIndexValue
//		if err := rlp.DecodeBytes([]byte(v), &utxoIndexVal); err != nil {
//			return nil, fmt.Errorf("Get account tokens error: data value is invalid(%s)", err.Error())
//		}
//		val, ok := tokens[utxoIndex.Asset.AssetId.String()]
//		if ok {
//			val.Balance += utxoIndexVal.Amount
//		} else {
//			// get asset info
//			assetInfo, err := repository.GetAssetInfo(utxoIndex.Asset)
//			if err != nil {
//				return nil, fmt.Errorf("Get acount tokens by index error: asset info does not exist")
//			}
//			tokens[utxoIndex.Asset.AssetId.String()] = &modules.AccountToken{
//				GasToken:   assetInfo.GasToken,
//				AssetID: utxoIndex.Asset,
//				Balance: utxoIndexVal.Amount,
//			}
//		}
//	}
//	return tokens, nil
//}

/**
遍历全局utxo，获取账户token信息
To get account token info by query the whole utxo table
*/
//func (repository *UtxoRepository) getAccountTokensWhole(addr common.Address) (
// map[string]*modules.AccountToken, error) {
//	tokens := map[string]*modules.AccountToken{}
//
//	key := fmt.Sprintf("%s", string(constants.UTXO_PREFIX))
//	data := repository.txUtxodb.GetPrefix([]byte(key))
//	if data == nil {
//		return nil, nil
//	}
//
//	for _, v := range data {
//		var utxo modules.Utxo
//		if err := rlp.DecodeBytes([]byte(v), &utxo); err != nil {
//			return nil, err
//		}
//		if !checkUtxo(&addr, nil, &utxo) {
//			continue
//		}
//
//		val, ok := tokens[utxo.Asset.AssetId.String()]
//		if ok {
//			val.Balance += utxo.Amount
//		} else {
//			// get asset info
//			assetInfo, err := repository.GetAssetInfo(utxo.Asset)
//			if err != nil {
//				return nil, fmt.Errorf("Get acount tokens by whole error: asset info does not exist")
//			}
//			tokens[utxo.Asset.AssetId.String()] = &modules.AccountToken{
//				GasToken:   assetInfo.GasToken,
//				AssetID: utxo.Asset,
//				Balance: utxo.Amount,
//			}
//		}
//	}
//	return tokens, nil
//}

/**
检查该utxo是否是需要的utxo
*/
//func checkUtxo(addr *common.Address, asset *modules.Asset, utxo *modules.Utxo) bool {
//	// check asset
//	if asset != nil && (strings.Compare(asset.AssetId.String(), utxo.Asset.AssetId.String()) != 0 ||
//		strings.Compare(asset.UniqueId.String(), utxo.Asset.UniqueId.String()) != 0) {
//		return false
//	}
//	// get addr
//	sAddr, _ := tokenengine.GetAddressFromScript(utxo.PkScript)
//	// check address
//	return strings.Compare(sAddr.String(), addr.String()) == 0
//
//}

/**
根据交易列表计算交易费总和
To compute transactions' fees
*/
// func (repository *UtxoRepository) ComputeFees(txs []*txspool.TxPoolTransaction) (uint64, error) {
// 	// current time slice mediator default income is 1 ptn
// 	fees := uint64(0)
// 	unitUtxo := map[modules.OutPoint]*modules.Utxo{}
// 	for i, tx := range txs {
// 		getUtxoFromUnitAndDb := func(outpoint *modules.OutPoint) (*modules.Utxo, error) {
// 			if utxo, ok := unitUtxo[*outpoint]; ok {
// 				return utxo, nil
// 			}
// 			return repository.txUtxodb.GetUtxoEntry(outpoint)
// 		}
// 		fee, err := tx.Tx.GetTxFee(getUtxoFromUnitAndDb)
// 		if err != nil {
// 			return 0, err
// 		}
// 		tx.TxFee = fee
// 		txs[i] = tx
// 		fees += fee.Amount
// 		for outPoint, utxo := range tx.Tx.GetNewUtxos() {
// 			unitUtxo[outPoint] = utxo
// 		}
// 	}
// 	return fees, nil
// }

/**
根据交易列表计算保证金交易的收益
*/
// func (repository *UtxoRepository) ComputeAwards(txs []*txspool.TxPoolTransaction,
// dagdb storage.IDagDb) (*modules.Addition, error) {
// 	awards := uint64(0)
// 	for _, tx := range txs {
// 		award, err := repository.ComputeTxAward(tx.Tx, dagdb)
// 		if err != nil {
// 			return nil, err
// 		}
// 		awards += award
// 	}
// 	if awards == 0 {
// 		return nil, nil
// 	} else {
// 		addition := new(modules.Addition)
// 		//first tx's asset
// 		addition.Asset = txs[0].Tx.Asset()
// 		addition.Amount = awards
// 		return addition, nil
// 	}
// }

//计算一笔保证金合约的币龄收益
// func (repository *UtxoRepository) ComputeTxAward(tx *modules.Transaction, dagdb storage.IDagDb) (uint64, error) {
// 	for _, msg := range tx.TxMessages {
// 		payload, ok := msg.Payload.(*modules.PaymentPayload)
// 		if !ok {
// 			continue
// 		}
// 		if payload.IsCoinbase() {
// 			continue
// 		}
// 		//判断是否是保证金合约地址
// 		utxo := repository.GetUxto(*payload.Inputs[0])
// 		if utxo == nil {
// 			log.Infof("get utxo from db by outpoint[%s] return empty", payload.Inputs[0].PreviousOutPoint.String())
// 			return 0, nil
// 		}
// 		addr, _ := tokenengine.GetAddressFromScript(utxo.PkScript)
// 		if addr.Equal(syscontract.DepositContractAddress) {
// 			awards := uint64(0)
// 			//对每一笔input输入进行计算奖励
// 			for _, txin := range payload.Inputs {
// 				utxo = repository.GetUxto(*txin)
// 				//1.通过交易hash获取单元hash
// 				//txin.PreviousOutPoint.TxHash 获取txhash
// 				//dagdb.GetTransaction()
// 				txlookup, err := dagdb.GetTxLookupEntry(txin.PreviousOutPoint.TxHash)
// 				if err != nil {
// 					return 0, err
// 				}
// 				//2.通过单元hash获取单元信息
// 				//header, _ := dagdb.GetHeaderByHash(unitHash)
// 				//3.通过单元获取头部信息中的时间戳
// 				timestamp := int64(txlookup.Timestamp)
// 				//depositRate, _, err := repository.statedb.GetSysConfig(modules.DepositRate)
// 				//if err != nil {
// 				//	return 0, err
// 				//}
// 				t1, _ := time.Parse("2006-01-02 15:04:05", time.Unix(timestamp, 0).UTC().Format("2006-01-02 15:04:05"))
// 				depositRate := repository.propDb.GetChainParameters().PledgeDailyReward
// 				award := award2.GetAwardsWithCoins(utxo.Amount, t1, depositRate)
// 				awards += award
// 			}
// 			return awards, nil
// 		}
// 	}
// 	return 0, nil
// }

//计算一笔Tx中包含多少手续费
//func (repository *UtxoRepository) ComputeTxFee(tx *modules.Transaction) (*modules.AmountAsset, error) {
//	return tx.GetTxFee(repository.txUtxodb.GetUtxoEntry)
//}

/**
计算Mediator的出块奖励
To compute mediator interest for packaging one unit
*/
func ComputeGenerateUnitReward() uint64 {

	rewards := parameter.CurrentSysParameters.GenerateUnitReward

	return rewards
}
