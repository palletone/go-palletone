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
	utxodb  storage.IUtxoDb
	idxdb   storage.IIndexDb
	statedb storage.IStateDb
	propDb  storage.IPropertyDb
	tokenEngine tokenengine.ITokenEngine
}

func NewUtxoRepository(utxodb storage.IUtxoDb, idxdb storage.IIndexDb,
	statedb storage.IStateDb,propDb storage.IPropertyDb,
	tokenEngine tokenengine.ITokenEngine) *UtxoRepository {
	return &UtxoRepository{
		utxodb: utxodb,
		idxdb: idxdb,
		statedb: statedb,
		propDb: propDb,
		tokenEngine:tokenEngine,
	}
}
func NewUtxoRepository4Db(db ptndb.Database,tokenEngine tokenengine.ITokenEngine) *UtxoRepository {
	utxodb := storage.NewUtxoDb(db,tokenEngine)
	statedb := storage.NewStateDb(db)
	idxdb := storage.NewIndexDb(db)
	propDb := storage.NewPropertyDb(db)

	return &UtxoRepository{
		utxodb: utxodb,
		idxdb: idxdb,
		statedb: statedb,
		propDb: propDb,
		tokenEngine:tokenEngine,
	}
}

type IUtxoRepository interface {
	GetUtxoEntry(outpoint *modules.OutPoint) (*modules.Utxo, error)
	GetStxoEntry(outpoint *modules.OutPoint) (*modules.Stxo, error)
	GetAllUtxos() (map[modules.OutPoint]*modules.Utxo, error)
	GetAddrOutpoints(addr common.Address) ([]modules.OutPoint, error)
	GetAddrUtxos(addr common.Address, asset *modules.Asset) (map[modules.OutPoint]*modules.Utxo, error)
	GetUxto(txin modules.Input) *modules.Utxo
	UpdateUtxo(unitTime int64, txHash common.Hash, payment *modules.PaymentPayload, msgIndex uint32) error
	IsUtxoSpent(outpoint *modules.OutPoint) (bool, error)
	ComputeTxFee(tx *modules.Transaction) (*modules.AmountAsset, error)
	GetUxtoSetByInputs(txins []modules.Input) (map[modules.OutPoint]*modules.Utxo, uint64)
	//GetAccountTokens(addr common.Address) (map[string]*modules.AccountToken, error)
	//WalletBalance(addr common.Address, asset modules.Asset) uint64
	// ComputeAwards(txs []*modules.TxPoolTransaction, dagdb storage.IDagDb) (*modules.Addition, error)
	// ComputeTxAward(tx *modules.Transaction, dagdb storage.IDagDb) (uint64, error)
	ClearUtxo() error
	SaveUtxoView(view map[modules.OutPoint]*modules.Utxo) error
	SaveUtxoEntity(outpoint *modules.OutPoint, utxo *modules.Utxo) error
}

func (repository *UtxoRepository) GetUtxoEntry(outpoint *modules.OutPoint) (*modules.Utxo, error) {
	return repository.utxodb.GetUtxoEntry(outpoint)
}
func (repository *UtxoRepository) GetStxoEntry(outpoint *modules.OutPoint) (*modules.Stxo, error) {
	return repository.utxodb.GetStxoEntry(outpoint)
}
func (repository *UtxoRepository) IsUtxoSpent(outpoint *modules.OutPoint) (bool, error) {
	return repository.utxodb.IsUtxoSpent(outpoint)
}
func (repository *UtxoRepository) GetAllUtxos() (map[modules.OutPoint]*modules.Utxo, error) {
	return repository.utxodb.GetAllUtxos()
}
func (repository *UtxoRepository) GetAddrOutpoints(addr common.Address) ([]modules.OutPoint, error) {
	return repository.utxodb.GetAddrOutpoints(addr)
}
func (repository *UtxoRepository) GetAddrUtxos(addr common.Address, asset *modules.Asset) (
	map[modules.OutPoint]*modules.Utxo, error) {
	return repository.utxodb.GetAddrUtxos(addr, asset)
}
func (repository *UtxoRepository) SaveUtxoView(view map[modules.OutPoint]*modules.Utxo) error {
	return repository.utxodb.SaveUtxoView(view)
}
func (repository *UtxoRepository) SaveUtxoEntity(outpoint *modules.OutPoint, utxo *modules.Utxo) error {
	return repository.utxodb.SaveUtxoEntity(outpoint, utxo)
}
func (repository *UtxoRepository) ClearUtxo() error {
	return repository.utxodb.ClearUtxo()
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
//	data := repository.utxodb.GetPrefix([]byte(key))
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
//		utxo, err := repository.utxodb.GetUtxoEntry(utxoIndex.OutPoint)
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
//	data := repository.utxodb.GetPrefix(constants.UTXO_PREFIX)
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
	data, err := repository.utxodb.GetUtxoEntry(txin.PreviousOutPoint)
	if err != nil {
		return nil
	}
	return data
}

// GetUtosOutPoint
func (repository *UtxoRepository) GetUtxoByOutpoint(outpoint *modules.OutPoint) (*modules.Utxo, error) {
	return repository.utxodb.GetUtxoEntry(outpoint)
}

/**
根据交易信息中的outputs创建UTXO， 根据交易信息中的inputs销毁UTXO
To create utxo according to outpus in transaction, and destroy utxo according to inputs in transaction
*/
func (repository *UtxoRepository) UpdateUtxo(unitTime int64, txHash common.Hash,
	payment *modules.PaymentPayload, msgIndex uint32) error {
	// update utxo
	err := repository.destroyUtxo(txHash, uint64(unitTime), payment.Inputs)
	if err != nil {
		return err
	}
	// create utxo
	errs := repository.writeUtxo(unitTime, txHash, msgIndex, payment.Outputs, payment.LockTime)
	if len(errs) > 0 {
		log.Error("error occurred on updated utxos, check the log file to find details.")
		return errors.New("error occurred on updated utxos, check the log file to find details.")
	}

	return nil

}

/**
创建UTXO
*/
func (repository *UtxoRepository) writeUtxo(unitTime int64, txHash common.Hash,
	msgIndex uint32, txouts []*modules.Output, lockTime uint32) []error {
	var errs []error
	for outIndex, txout := range txouts {
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

		if err := repository.utxodb.SaveUtxoEntity(outpoint, utxo); err != nil {
			log.Error("Write utxo", "error", err.Error())
			errs = append(errs, err)
			continue
		}

		sAddr, _ := repository.tokenEngine.GetAddressFromScript(txout.PkScript)
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
		utxo, err := repository.utxodb.GetUtxoEntry(outpoint)
		if err != nil {
			log.Error("Query utxo when destroy uxto", "error", err.Error(), "outpoint", outpoint.String())
			return err
		}

		// delete utxo
		if err := repository.utxodb.DeleteUtxo(outpoint, txid, unitTime); err != nil {
			log.Error("Update uxto... ", "error", err.Error())
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
//	if data := repository.utxodb.GetPrefix([]byte(preKey)); data != nil {
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
//	if data := repository.utxodb.GetPrefix([]byte(preKey)); data != nil {
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
//	data := repository.utxodb.GetPrefix(utxoIndex.AccountKey())
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
//	data := repository.utxodb.GetPrefix([]byte(key))
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
// func (repository *UtxoRepository) ComputeFees(txs []*modules.TxPoolTransaction) (uint64, error) {
// 	// current time slice mediator default income is 1 ptn
// 	fees := uint64(0)
// 	unitUtxo := map[modules.OutPoint]*modules.Utxo{}
// 	for i, tx := range txs {
// 		getUtxoFromUnitAndDb := func(outpoint *modules.OutPoint) (*modules.Utxo, error) {
// 			if utxo, ok := unitUtxo[*outpoint]; ok {
// 				return utxo, nil
// 			}
// 			return repository.utxodb.GetUtxoEntry(outpoint)
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
// func (repository *UtxoRepository) ComputeAwards(txs []*modules.TxPoolTransaction,
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
func (repository *UtxoRepository) ComputeTxFee(tx *modules.Transaction) (*modules.AmountAsset, error) {
	return tx.GetTxFee(repository.utxodb.GetUtxoEntry)
}

/**
计算Mediator的出块奖励
To compute mediator interest for packaging one unit
*/
func ComputeGenerateUnitReward() uint64 {

	rewards := parameter.CurrentSysParameters.GenerateUnitReward

	return rewards
}
