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
	"fmt"
	"strings"
	"unsafe"

	"github.com/palletone/go-palletone/common"
	award2 "github.com/palletone/go-palletone/common/award"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/rlp"
	"github.com/palletone/go-palletone/dag/constants"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/storage"
	"github.com/palletone/go-palletone/tokenengine"
)

type UtxoRepository struct {
	utxodb  storage.IUtxoDb
	idxdb   storage.IIndexDb
	statedb storage.IStateDb
}

func NewUtxoRepository(utxodb storage.IUtxoDb, idxdb storage.IIndexDb, statedb storage.IStateDb) *UtxoRepository {
	return &UtxoRepository{utxodb: utxodb, idxdb: idxdb, statedb: statedb}
}

type IUtxoRepository interface {
	ReadUtxos(addr common.Address, asset modules.Asset) (map[modules.OutPoint]*modules.Utxo, uint64)
	GetUxto(txin modules.Input) modules.Utxo
	UpdateUtxo(txHash common.Hash, msg *modules.Message, msgIndex uint32) error
	ComputeFees(txs []*modules.TxPoolTransaction) (uint64, error)
	ComputeTxFee(tx *modules.Transaction) (*modules.InvokeFees, error)
	GetUxtoSetByInputs(txins []modules.Input) (map[modules.OutPoint]*modules.Utxo, uint64)
	GetAccountTokens(addr common.Address) (map[string]*modules.AccountToken, error)
	WalletBalance(addr common.Address, asset modules.Asset) uint64
	ComputeAwards(txs []*modules.TxPoolTransaction, dagdb storage.IDagDb) (*modules.Addition, error)
	ComputeTxAward(tx *modules.Transaction, dagdb storage.IDagDb) (uint64, error)
}

/**
根据用户地址、选择的资产类型、转账金额、手续费返回utxo
Return utxo struct and his total amount according to user's address, asset type
*/
func (repository *UtxoRepository) ReadUtxos(addr common.Address, asset modules.Asset) (map[modules.OutPoint]*modules.Utxo, uint64) {

	if dagconfig.DefaultConfig.UtxoIndex {
		return repository.readUtxosByIndex(addr, asset)
	} else {
		return repository.readUtxosFrAll(addr, asset)
	}
}

/**
根据UTXO索引数据库读取UTXO信息
To get utxo info by utxo index db
*/
func (repository *UtxoRepository) readUtxosByIndex(addr common.Address, asset modules.Asset) (map[modules.OutPoint]*modules.Utxo, uint64) {
	// step1. read outpoint from utxo index
	utxoIndex := modules.UtxoIndex{
		AccountAddr: addr,
		Asset:       &asset,
	}
	key := utxoIndex.AssetKey()
	data := repository.utxodb.GetPrefix([]byte(key))
	if data == nil {
		return nil, 0
	}
	// step2. get utxo
	vout := map[modules.OutPoint]*modules.Utxo{}
	balance := uint64(0)
	for k := range data {
		if err := utxoIndex.QueryFields([]byte(k)); err != nil {
			continue
		}
		utxo, err := repository.utxodb.GetUtxoEntry(utxoIndex.OutPoint)
		if err != nil {
			log.Error("Get utxo error by outpoint", "error:", err.Error())
			continue
		}
		//var utxo modules.Utxo
		//if err := rlp.DecodeBytes([]byte(udata), &utxo); err != nil {
		//	log.Error("Decode utxo data :", "error:", err.Error())
		//	continue
		//}
		vout[*utxoIndex.OutPoint] = utxo
		balance += utxo.Amount
	}
	return vout, balance
}

/**
扫描UTXO全表，获取对应账户的指定token可用的utxo列表
To get utxo info by scanning all utxos.
*/
func (repository *UtxoRepository) readUtxosFrAll(addr common.Address, asset modules.Asset) (map[modules.OutPoint]*modules.Utxo, uint64) {
	// key: [UTXO_PREFIX][addr]_[asset]_[msgindex]_[out index]
	data := repository.utxodb.GetPrefix(constants.UTXO_PREFIX)
	if data == nil {
		return nil, 0
	}

	// value: utxo rlp bytes
	vout := map[modules.OutPoint]*modules.Utxo{}
	var balance uint64
	balance = 0
	for k, v := range data {
		var utxo modules.Utxo
		if err := rlp.DecodeBytes([]byte(v), &utxo); err != nil {
			log.Error("Decode utxo data :", "error:", err.Error())
			continue
		}
		// check asset
		if strings.Compare(asset.AssetId.String(), utxo.Asset.AssetId.String()) != 0 ||
			strings.Compare(asset.UniqueId.String(), utxo.Asset.UniqueId.String()) != 0 {
			continue
		}
		// get addr
		sAddr, _ := tokenengine.GetAddressFromScript(utxo.PkScript)
		// check address
		if strings.Compare(sAddr.String(), addr.String()) != 0 {
			continue
		}
		outpoint := modules.KeyToOutpoint([]byte(k))
		vout[*outpoint] = &utxo
		balance += utxo.Amount
	}

	return vout, balance
}

/**
获取某个input对应的utxo信息
To get utxo struct according to it's input information
*/
func (repository *UtxoRepository) GetUxto(txin modules.Input) modules.Utxo {

	data, err := repository.utxodb.GetUtxoEntry(txin.PreviousOutPoint)
	if err != nil {
		return modules.Utxo{}
	}
	return *data
}

// GetUtosOutPoint
func (repository *UtxoRepository) GetUtxoByOutpoint(outpoint *modules.OutPoint) (*modules.Utxo, error) {
	return repository.utxodb.GetUtxoEntry(outpoint)
}

/**
根据交易信息中的outputs创建UTXO， 根据交易信息中的inputs销毁UTXO
To create utxo according to outpus in transaction, and destory utxo according to inputs in transaction
*/
func (repository *UtxoRepository) UpdateUtxo(txHash common.Hash, msg *modules.Message, msgIndex uint32) error {
	var payload interface{}

	payload = msg.Payload
	payment, ok := payload.(*modules.PaymentPayload)
	if ok == true {
		// create utxo
		errs := repository.writeUtxo(txHash, msgIndex, payment.Outputs, payment.LockTime)
		if len(errs) > 0 {
			log.Error("error occurred on updated utxos, check the log file to find details.")
			return errors.New("error occurred on updated utxos, check the log file to find details.")
		}
		// update utxo
		repository.destoryUtxo(payment.Inputs)
		return nil
	}
	return errors.New("UpdateUtxo: the transaction payload is not payment.")
}

/**
创建UTXO
*/
func (repository *UtxoRepository) writeUtxo(txHash common.Hash, msgIndex uint32, txouts []*modules.Output, lockTime uint32) []error {
	var errs []error
	for outIndex, txout := range txouts {
		utxo := &modules.Utxo{
			Amount:   txout.Value,
			Asset:    txout.Asset,
			PkScript: txout.PkScript,
			LockTime: lockTime,
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

		// write to utxo index db
		if dagconfig.DefaultConfig.UtxoIndex == false {
			continue
		}

		// get address
		sAddr, _ := tokenengine.GetAddressFromScript(txout.PkScript)
		// save addr key index.
		outpoint_key := make([]byte, 0)
		outpoint_key = append(outpoint_key, constants.AddrOutPoint_Prefix...)
		outpoint_key = append(outpoint_key, sAddr.Bytes()...)
		repository.idxdb.SaveIndexValue(append(outpoint_key, outpoint.Hash().Bytes()...), outpoint)

		utxoIndex := modules.UtxoIndex{
			AccountAddr: sAddr,
			Asset:       txout.Asset,
			OutPoint:    outpoint,
		}
		utxoIndexVal := modules.UtxoIndexValue{
			Amount:   txout.Value,
			LockTime: lockTime,
		}
		if err := repository.idxdb.SaveIndexValue(utxoIndex.ToKey(), utxoIndexVal); err != nil {
			log.Error("Write utxo index error: %s", err.Error())
			errs = append(errs, err)
		}
		//update address account info
		if txout.Asset.AssetId == modules.PTNCOIN {
			repository.statedb.UpdateAccountInfoBalance(sAddr, int64(txout.Value))
		}
	}
	return errs
}

/**
销毁utxo
destory utxo, delete from UTXO database
*/
func (repository *UtxoRepository) destoryUtxo(txins []*modules.Input) {
	for _, txin := range txins {
		//TODO for download sync
		if txin == nil {
			continue
		}
		outpoint := txin.PreviousOutPoint
		if outpoint == nil || outpoint.IsEmpty() {
			if len(txin.Extra) > 0 {
				var assetInfo modules.AssetInfo
				if err := rlp.DecodeBytes(txin.Extra, &assetInfo); err == nil {
					// save asset info
					if err := repository.statedb.SaveAssetInfo(&assetInfo); err != nil {
						log.Error("Save asset info error")
					}
				}
			}
			continue
		}
		// get utxo info
		utxo, err := repository.utxodb.GetUtxoEntry(outpoint)
		if err != nil {
			log.Error("Query utxo when destory uxto", "error", err.Error())
			continue
		}

		// delete utxo
		if err := repository.utxodb.DeleteUtxo(outpoint); err != nil {
			log.Error("Update uxto... ", "error", err.Error())
			continue
		}
		// delete index data
		sAddr, _ := tokenengine.GetAddressFromScript(utxo.PkScript)
		// addr := common.Address{}
		// addr.SetString(sAddr.String())
		// utxoIndex := &modules.UtxoIndex{
		// 	AccountAddr: addr,
		// 	Asset:       utxo.Asset,
		// 	OutPoint:    outpoint,
		// }
		// if err := repository.idxdb.DeleteUtxoByIndex(utxoIndex); err != nil {
		// 	log.Error("Destory uxto index", "error", err.Error())
		// 	continue
		// }
		if utxo.Asset.AssetId == modules.NewPTNIdType() { // modules.PTNCOIN
			repository.statedb.UpdateAccountInfoBalance(sAddr, -int64(utxo.Amount))
		}
	}
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
get asset infomation from leveldb by assetid ( Asset struct type )
*/
func (repository *UtxoRepository) GetAssetInfo(assetId *modules.Asset) (*modules.AssetInfo, error) {
	return repository.statedb.GetAssetInfo(assetId)
}

/**
获得某个账户下面的余额信息
To get balance by wallet address and his/her chosen asset type
*/
func (repository *UtxoRepository) WalletBalance(addr common.Address, asset modules.Asset) uint64 {
	if dagconfig.DefaultConfig.UtxoIndex {
		return repository.walletBalanceByIndex(addr, asset)
	} else {
		return repository.walletBalanceFrAll(addr, asset)
	}
}

/**
通过索引数据库获得账户余额
To get balance by utxo index db
*/
func (repository *UtxoRepository) walletBalanceByIndex(addr common.Address, asset modules.Asset) uint64 {
	balance := uint64(0)

	utxoIndex := modules.UtxoIndex{
		AccountAddr: addr,
		Asset:       &asset,
	}
	preKey := utxoIndex.AssetKey()

	if data := repository.utxodb.GetPrefix([]byte(preKey)); data != nil {
		for _, v := range data {
			var utxoIndexVal modules.UtxoIndexValue
			if err := rlp.DecodeBytes(v, &utxoIndexVal); err != nil {
				log.Error("Decode utxo data error:", err)
				continue
			}
			balance += utxoIndexVal.Amount
		}
	}

	return balance

}

/**
通过查询全表获得账户余额
To get balance by query all utxo table
*/
func (repository *UtxoRepository) walletBalanceFrAll(addr common.Address, asset modules.Asset) uint64 {
	balance := uint64(0)

	preKey := fmt.Sprintf("%s", constants.UTXO_PREFIX)

	if data := repository.utxodb.GetPrefix([]byte(preKey)); data != nil {
		for _, v := range data {
			var utxo modules.Utxo
			if err := rlp.DecodeBytes(v, &utxo); err != nil {
				log.Error("Decode utxo data error:", err)
				continue
			}
			if !checkUtxo(&addr, &asset, &utxo) {
				continue
			}
			balance += utxo.Amount
		}
	}

	return balance
}

/**
根据payload中的inputs获得对应的UTXO map
*/
func (repository *UtxoRepository) GetUxtoSetByInputs(txins []modules.Input) (map[modules.OutPoint]*modules.Utxo, uint64) {
	utxos := map[modules.OutPoint]*modules.Utxo{}
	total := uint64(0)
	for _, in := range txins {
		utxo := repository.GetUxto(in)
		if unsafe.Sizeof(utxo) == 0 {
			continue
		}
		utxos[*in.PreviousOutPoint] = &utxo
		total += utxo.Amount
	}
	return utxos, total
}

/**
获得某个账户下的token名称和assetid信息
To get someone account's list of tokens and those assetids
*/
func (repository *UtxoRepository) GetAccountTokens(addr common.Address) (map[string]*modules.AccountToken, error) {
	if dagconfig.DefaultConfig.UtxoIndex {
		return repository.getAccountTokensByIndex(addr)
	} else {
		return repository.getAccountTokensWhole(addr)
	}
}

/**
通过索引数据库获得用户的token信息
To get account's token info by utxo index db
*/
func (repository *UtxoRepository) getAccountTokensByIndex(addr common.Address) (map[string]*modules.AccountToken, error) {
	tokens := map[string]*modules.AccountToken{}
	utxoIndex := modules.UtxoIndex{
		AccountAddr: addr,
		Asset:       &modules.Asset{},
		OutPoint:    &modules.OutPoint{},
	}
	data := repository.utxodb.GetPrefix(utxoIndex.AccountKey())
	if data == nil || len(data) == 0 {
		return nil, nil
	}
	for k, v := range data {
		if err := utxoIndex.QueryFields([]byte(k)); err != nil {
			return nil, fmt.Errorf("Get account tokens by key error: data key is invalid(%s)", err.Error())
		}
		var utxoIndexVal modules.UtxoIndexValue
		if err := rlp.DecodeBytes([]byte(v), &utxoIndexVal); err != nil {
			return nil, fmt.Errorf("Get account tokens error: data value is invalid(%s)", err.Error())
		}
		val, ok := tokens[utxoIndex.Asset.AssetId.String()]
		if ok {
			val.Balance += utxoIndexVal.Amount
		} else {
			// get asset info
			assetInfo, err := repository.GetAssetInfo(utxoIndex.Asset)
			if err != nil {
				return nil, fmt.Errorf("Get acount tokens by index error: asset info does not exist")
			}
			tokens[utxoIndex.Asset.AssetId.String()] = &modules.AccountToken{
				Alias:   assetInfo.Alias,
				AssetID: utxoIndex.Asset,
				Balance: utxoIndexVal.Amount,
			}
		}
	}
	return tokens, nil
}

/**
遍历全局utxo，获取账户token信息
To get account token info by query the whole utxo table
*/
func (repository *UtxoRepository) getAccountTokensWhole(addr common.Address) (map[string]*modules.AccountToken, error) {
	tokens := map[string]*modules.AccountToken{}

	key := fmt.Sprintf("%s", string(constants.UTXO_PREFIX))
	data := repository.utxodb.GetPrefix([]byte(key))
	if data == nil {
		return nil, nil
	}

	for _, v := range data {
		var utxo modules.Utxo
		if err := rlp.DecodeBytes([]byte(v), &utxo); err != nil {
			return nil, err
		}
		if !checkUtxo(&addr, nil, &utxo) {
			continue
		}

		val, ok := tokens[utxo.Asset.AssetId.String()]
		if ok {
			val.Balance += utxo.Amount
		} else {
			// get asset info
			assetInfo, err := repository.GetAssetInfo(utxo.Asset)
			if err != nil {
				return nil, fmt.Errorf("Get acount tokens by whole error: asset info does not exist")
			}
			tokens[utxo.Asset.AssetId.String()] = &modules.AccountToken{
				Alias:   assetInfo.Alias,
				AssetID: utxo.Asset,
				Balance: utxo.Amount,
			}
		}
	}
	return tokens, nil
}

/**
检查该utxo是否是需要的utxo
*/
func checkUtxo(addr *common.Address, asset *modules.Asset, utxo *modules.Utxo) bool {
	// check asset
	if asset != nil && (strings.Compare(asset.AssetId.String(), utxo.Asset.AssetId.String()) != 0 ||
		strings.Compare(asset.UniqueId.String(), utxo.Asset.UniqueId.String()) != 0) {
		return false
	}
	// get addr
	sAddr, _ := tokenengine.GetAddressFromScript(utxo.PkScript)
	// check address
	if strings.Compare(sAddr.String(), addr.String()) != 0 {
		//fmt.Printf(">>>>> Address is not compare:scriptPubKey.Address=%s, address=%s\n",
		//	sAddr, addr.String())
		return false
	}
	return true
}

/**
根据交易列表计算交易费总和
To compute transactions' fees
*/
func (repository *UtxoRepository) ComputeFees(txs []*modules.TxPoolTransaction) (uint64, error) {
	// current time slice mediator default income is 1 ptn
	fees := uint64(0)
	for i, tx := range txs {
		fee, err := repository.ComputeTxFee(tx.Tx)
		if err != nil {
			return 0, err
		}
		tx.TxFee = fee
		txs[i] = tx
		fees += fee.Amount
	}
	return fees, nil
}

/**
根据交易列表计算保证金交易的收益
*/
func (repository *UtxoRepository) ComputeAwards(txs []*modules.TxPoolTransaction, dagdb storage.IDagDb) (*modules.Addition, error) {
	awards := uint64(0)
	for _, tx := range txs {
		award, err := repository.ComputeTxAward(tx.Tx, dagdb)
		if err != nil {
			return nil, err
		}
		awards += award
	}
	if awards == 0 {
		return nil, nil
	} else {
		addition := new(modules.Addition)
		addition.Asset = *modules.NewPTNAsset()
		return addition, nil
	}
}

//计算一笔保证金合约的币龄收益
func (repository *UtxoRepository) ComputeTxAward(tx *modules.Transaction, dagdb storage.IDagDb) (uint64, error) {
	for _, msg := range tx.TxMessages {
		payload, ok := msg.Payload.(*modules.PaymentPayload)
		if !ok {
			continue
		}
		if payload.IsCoinbase() {
			continue
		}
		//判断是否是保证金合约地址
		utxo := repository.GetUxto(*payload.Inputs[0])
		addr, _ := tokenengine.GetAddressFromScript(utxo.PkScript)
		if addr.String() == "PCGTta3M4t3yXu8uRgkKvaWd2d8DR32W9vM" {
			awards := uint64(0)
			//对每一笔input输入进行计算奖励
			for _, txin := range payload.Inputs {
				utxo = repository.GetUxto(*txin)
				//1.通过交易hash获取单元hash
				//txin.PreviousOutPoint.TxHash 获取txhash
				//dagdb.GetTransaction()
				_, unitHash, _, _ := dagdb.GetTransaction(txin.PreviousOutPoint.TxHash)
				//2.通过单元hash获取单元信息
				unit, _ := dagdb.GetUnit(unitHash)
				//3.通过单元获取头部信息中的时间戳
				timestamp := unit.Header().Creationdate
				award := award2.GetAwardsWithCoins(utxo.Amount, timestamp)
				awards += award
			}
			return awards, nil
		}
	}
	return 0, nil
}

//计算一笔Tx中包含多少手续费
func (repository *UtxoRepository) ComputeTxFee(tx *modules.Transaction) (*modules.InvokeFees, error) {

	for _, msg := range tx.TxMessages {
		payload, ok := msg.Payload.(*modules.PaymentPayload)
		if ok == false {
			continue
		}
		inAmount := uint64(0)
		outAmount := uint64(0)
		for _, txin := range payload.Inputs {
			utxo := repository.GetUxto(*txin)
			if utxo.IsEmpty() {
				return nil, fmt.Errorf("Txin(txhash=%s, msgindex=%v, outindex=%v)'s utxo is empty:",
					txin.PreviousOutPoint.TxHash.String(),
					txin.PreviousOutPoint.MessageIndex,
					txin.PreviousOutPoint.OutIndex)
			}
			// check overflow
			if inAmount+utxo.Amount > (1<<64 - 1) {
				return nil, fmt.Errorf("Compute fees: txin total overflow")
			}
			inAmount += utxo.Amount
		}

		for _, txout := range payload.Outputs {
			// check overflow
			if outAmount+txout.Value > (1<<64 - 1) {
				return nil, fmt.Errorf("Compute fees: txout total overflow")
			}
			log.Debug("+++++++++++++++++++++ tx_out_amonut ++++++++++++++++++++", "tx_outAmount", txout.Value)
			outAmount += txout.Value
		}
		if inAmount < outAmount {

			return nil, fmt.Errorf("Compute fees: tx %s txin amount less than txout amount. amount:%d ,outAmount:%d ", tx.Hash().String(), inAmount, outAmount)
		}
		fees := inAmount - outAmount
		return &modules.InvokeFees{Amount: fees, Asset: payload.Outputs[0].Asset}, nil

	}
	return nil, fmt.Errorf("Compute fees: no payment payload")
}

/**
计算Mediator的利息
To compute mediator interest for packaging one unit
*/
func ComputeRewards() uint64 {
	var rewards uint64
	if dagconfig.DefaultConfig.IsRewardCoin {
		rewards = uint64(modules.DAO)
	}
	return rewards
}

func IsCoinBase(tx *modules.Transaction) bool {
	if len(tx.TxMessages) != 1 {
		return false
	}
	msg, ok := tx.TxMessages[0].Payload.(*modules.PaymentPayload)
	if !ok {
		return false
	}
	prevOut := msg.Inputs[0].PreviousOutPoint
	if prevOut.TxHash != (common.Hash{}) {
		return false
	}
	return true
}
