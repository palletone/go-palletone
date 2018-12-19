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
	"crypto/sha256"
	"encoding/json"
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
	"github.com/palletone/go-palletone/dag/txspool"
	"github.com/palletone/go-palletone/dag/vote"
	"github.com/palletone/go-palletone/tokenengine"
)

type IUnitRepository interface {
	//设置稳定单元的Hash
	SetStableUnitHash(hash common.Hash)
	//设置最新单元的Hash
	SetLastUnitHash(hash common.Hash)
	//清空Unstable数据，回滚到稳定点状态
	RollbackToStableUnit()
	//批量增加多个Unit，主要用于主链切换的情形
	BatchSaveUnit(units []*modules.Unit)
	GetGenesisUnit(index uint64) (*modules.Unit, error)
	GenesisHeight() modules.ChainIndex
	SaveUnit(unit *modules.Unit, txpool txspool.ITxPool, isGenesis bool, passed bool) error
	CreateUnit(mAddr *common.Address, txpool txspool.ITxPool, ks *keystore.KeyStore, t time.Time) ([]modules.Unit, error)
	IsGenesis(hash common.Hash) bool
}
type UnitRepository struct {
	dagdb          storage.IDagDb
	idxdb          storage.IIndexDb
	uxtodb         storage.IUtxoDb
	statedb        storage.IStateDb
	validate       Validator
	utxoRepository IUtxoRepository
	logger         log.ILogger
}

func NewUnitRepository(dagdb storage.IDagDb, idxdb storage.IIndexDb, utxodb storage.IUtxoDb, statedb storage.IStateDb, l log.ILogger) *UnitRepository {
	utxoRep := NewUtxoRepository(utxodb, idxdb, statedb, l)
	val := NewValidate(dagdb, utxodb, utxoRep, statedb, l)
	return &UnitRepository{dagdb: dagdb, idxdb: idxdb, uxtodb: utxodb, statedb: statedb, validate: val, utxoRepository: utxoRep}
}

func NewUnitRepository4Db(db ptndb.Database, l log.ILogger) *UnitRepository {
	dagdb := storage.NewDagDb(db, l)
	utxodb := storage.NewUtxoDb(db, l)
	statedb := storage.NewStateDb(db, l)
	idxdb := storage.NewIndexDb(db, l)
	utxoRep := NewUtxoRepository(utxodb, idxdb, statedb, l)
	val := NewValidate(dagdb, utxodb, utxoRep, statedb, l)
	return &UnitRepository{dagdb: dagdb, idxdb: idxdb, uxtodb: utxodb, statedb: statedb, validate: val, utxoRepository: utxoRep}
}

func RHashStr(x interface{}) string {
	x_byte, err := json.Marshal(x)
	if err != nil {
		return ""
	}
	s256 := sha256.New()
	s256.Write(x_byte)
	return fmt.Sprintf("%x", s256.Sum(nil))

}

/**
生成创世单元，需要传入创世单元的配置信息以及coinbase交易
generate genesis unit, need genesis unit configure fields and transactions list
*/
func NewGenesisUnit(txs modules.Transactions, time int64, asset *modules.Asset) (*modules.Unit, error) {
	gUnit := modules.Unit{}

	// genesis unit height
	chainIndex := modules.ChainIndex{AssetID: asset.AssetId, IsMain: true, Index: 0}

	// transactions merkle root
	root := core.DeriveSha(txs)

	// generate genesis unit header
	header := modules.Header{
		AssetIDs:     []modules.IDType16{asset.AssetId},
		Number:       chainIndex,
		TxRoot:       root,
		Creationdate: time,
	}

	gUnit.UnitHeader = &header
	// copy txs
	gUnit.CopyBody(txs)
	// set unit size
	gUnit.UnitSize = gUnit.Size()
	// set unit hash
	gUnit.UnitHash = gUnit.Hash()
	return &gUnit, nil
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

	r := sign[:32]
	s := sign[32:64]
	v := sign[64:]
	if len(v) != 1 {
		return unit, errors.New("error.")
	}

	unit.UnitHeader.Authors = modules.Authentifier{
		Address: signer,
		R:       r,
		S:       s,
		V:       v,
	}
	// to set witness list, should be creator himself
	// var authentifier modules.Authentifier
	// authentifier.Address = signer
	// unit.UnitHeader.Witness = append(unit.UnitHeader.Witness, &authentifier)
	// unit.UnitHeader.GroupSign = sign
	return unit, nil
}
func (rep *UnitRepository) SetStableUnitHash(hash common.Hash) {
	rep.dagdb.SetStableUnitHash(hash)
}

func (rep *UnitRepository) SetLastUnitHash(hash common.Hash) {
	rep.dagdb.SetLastUnitHash(hash)
}
func (rep *UnitRepository) RollbackToStableUnit() {
	//TODO Devin
}

//批量增加多个Unit，主要用于主链切换的情形
func (rep *UnitRepository) BatchSaveUnit(units []*modules.Unit) {
	//TODO Devin
}

/**
创建单元
create common unit
@param mAddr is minner addr
return: correct if error is nil, and otherwise is incorrect
*/
func (unitOp *UnitRepository) CreateUnit(mAddr *common.Address, txpool txspool.ITxPool, ks *keystore.KeyStore, t time.Time) ([]modules.Unit, error) {
	if txpool == nil || !common.IsValidAddress(mAddr.String()) || ks == nil {
		log.Debug("UnitRepository", "CreateUnit txpool:", txpool, "mdAddr:", mAddr.String(), "ks:", ks)
		return nil, fmt.Errorf("Create unit: nil address or txspool is not allowed")
	}
	units := []modules.Unit{}
	// step1. get mediator responsible for asset (for now is ptn)
	// bAsset, _, _ := unitOp.statedb.GetConfig([]byte(modules.FIELD_GENESIS_ASSET))
	// if len(bAsset) <= 0 {
	// 	return nil, fmt.Errorf("Create unit error: query asset info empty")
	// }
	// var asset modules.Asset
	// if err := rlp.DecodeBytes(bAsset, &asset); err != nil {
	// 	return nil, fmt.Errorf("Create unit: %s", err.Error())
	// }

	// @jay
	//var asset modules.Asset
	//assetId, _ := modules.SetIdTypeByHex(dagconfig.DefaultConfig.PtnAssetHex)
	//asset.AssetId = assetId
	//asset.UniqueId = assetId
	asset := modules.NewPTNAsset()
	// step2. compute chain height
	// get current world_state index.

	index := uint64(1)
	isMain := true
	// chainIndex := modules.ChainIndex{AssetID: asset.AssetId, IsMain: isMain, Index: index}
	chainIndex, err := unitOp.statedb.GetCurrentChainIndex(asset.AssetId)
	if err != nil {
		chainIndex = &modules.ChainIndex{AssetID: asset.AssetId, IsMain: isMain, Index: index + 1}
		unitOp.logger.Error("GetCurrentChainIndex is failed.", "error", err)
	} else {
		chainIndex.Index += 1
	}
	// step3. generate genesis unit header
	header := modules.Header{
		AssetIDs: []modules.IDType16{},
		Number:   *chainIndex,
		//TxRoot:   root,
		//		Creationdate: time.Now().Unix(),
	}
	header.AssetIDs = append(header.AssetIDs, asset.AssetId)
	h_hash := header.HashWithOutTxRoot()

	// step4. get transactions from txspool
	poolTxs, _ := txpool.GetSortedTxs(h_hash)

	// step5. compute minner income: transaction fees + interest
	fees, err := unitOp.utxoRepository.ComputeFees(poolTxs)
	if err != nil {
		log.Error("ComputeFees is failed.", "error", err.Error())
		return nil, err
	} else {
		log.Debug("THE unit transactions fee is here. ", "fees", fees)
	}
	additions := make(map[common.Address]*modules.Addition)
	//TODO 附加利息收益
	//获取保证金利息
	contractAddition, err := unitOp.utxoRepository.ComputeAwards(poolTxs, unitOp.dagdb)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	if contractAddition != nil {
		addr, _ := common.StringToAddress("PCGTta3M4t3yXu8uRgkKvaWd2d8DR32W9vM")
		additions[addr] = contractAddition
	}
	//coinbase, err := CreateCoinbase(mAddr, fees+awards, asset, t)
	coinbase, err := CreateCoinbase(mAddr, fees, additions, asset, t)

	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	txs := modules.Transactions{coinbase}
	// step6 get unit's txs in txpool's txs
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

	// step9. generate genesis unit header
	header.TxRoot = root
	unit := modules.Unit{}
	unit.UnitHeader = &header
	unit.UnitHash = header.Hash()

	// step10. copy txs
	unit.CopyBody(txs)

	// step11. set size
	unit.UnitSize = unit.Size()
	units = append(units, unit)
	return units, nil
}

func (unitOp *UnitRepository) GetCurrentChainIndex(assetId modules.IDType16) (*modules.ChainIndex, error) {
	return unitOp.statedb.GetCurrentChainIndex(assetId)
}

/**
从leveldb中查询GenesisUnit信息
To get genesis unit info from leveldb
*/
func (unitOp *UnitRepository) GetGenesisUnit(index uint64) (*modules.Unit, error) {
	// unit key: [HEADER_PREFIX][chain index number]_[chain index]_[unit hash]
	//key := fmt.Sprintf("%s%v_", constants.HEADER_PREFIX, index)

	// data := unitOp.dagdb.GetPrefix([]byte(key))
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
	// 	txs, err := unitOp.dagdb.GetUnitTransactions(unit.UnitHash)
	// 	if err != nil {
	// 		return nil, fmt.Errorf("Get genesis unit transactions: %s", err.Error())
	// 	}
	// 	unit.Txs = txs
	// 	unit.UnitSize = unit.Size()
	// 	return &unit, nil
	// 	//}
	// }
	// return nil, nil
	number := modules.ChainIndex{}
	number.Index = index
	number.IsMain = true

	//number.AssetID, _ = modules.SetIdTypeByHex(dagconfig.DefaultConfig.PtnAssetHex) //modules.PTNCOIN
	//asset := modules.NewPTNAsset()
	number.AssetID = modules.CoreAsset.AssetId
	hash, err := unitOp.dagdb.GetHashByNumber(number)
	if err != nil {
		log.Debug("unitOp: getgenesis by number , current error.", "error", err)
		return nil, err
	}
	log.Debug("unitOp: get genesis(hash):", "geneseis_hash", hash)
	return unitOp.dagdb.GetUnit(hash)
}

/**
获取创世单元的高度
To get genesis unit height
*/
func (unitRep *UnitRepository) GenesisHeight() modules.ChainIndex {
	unit, err := unitRep.GetGenesisUnit(0)
	if unit == nil || err != nil {
		return modules.ChainIndex{}
	}
	return unit.UnitHeader.Number
}
func (unitRep *UnitRepository) IsGenesis(hash common.Hash) bool {
	unit, err := unitRep.GetGenesisUnit(0)
	if unit == nil || err != nil {
		return false
	}
	return hash == unit.Hash()
}

func (unitOp *UnitRepository) GetUnitTransactions(unitHash common.Hash) (modules.Transactions, error) {
	txs := modules.Transactions{}
	// get body data: transaction list.
	// if getbody return transactions list, then don't range txHashlist.
	txHashList, err := unitOp.dagdb.GetBody(unitHash)
	if err != nil {
		return nil, err
	}
	// get transaction by tx'hash.
	for _, txHash := range txHashList {
		tx, _, _, _ := unitOp.dagdb.GetTransaction(txHash)
		if err != nil {
			txs = append(txs, tx)
		}
	}
	return txs, nil
}

/**
为创世单元生成ConfigPayload
To generate config payload for genesis unit
*/
func GenGenesisConfigPayload(genesisConf *core.Genesis, asset *modules.Asset) (modules.ConfigPayload, error) {
	var confPay modules.ConfigPayload
	confPay.ConfigSet = []modules.ContractWriteSet{}

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

				//confPay.ConfigSet = append(confPay.ConfigSet,
				//	modules.ContractWriteSet{Key: sk, Value: modules.ToPayloadMapValueBytes(v.Field(k).Interface())})
				confPay.ConfigSet = append(confPay.ConfigSet,
					modules.ContractWriteSet{Key: sk, Value: v.Field(k).Interface()})
			}
		} else {
			sk := tt.Field(i).Name
			if strings.Contains(sk, "Initial") {
				sk = strings.Replace(sk, "Initial", "", -1)
			}
			confPay.ConfigSet = append(confPay.ConfigSet,
				modules.ContractWriteSet{Key: sk, Value: modules.ToPayloadMapValueBytes(vv.Field(i).Interface())})
		}
	}

	confPay.ConfigSet = append(confPay.ConfigSet,
		modules.ContractWriteSet{Key: modules.FIELD_GENESIS_ASSET, Value: modules.ToPayloadMapValueBytes(*asset)})

	return confPay, nil
}

//Yiran
func (unitOp *UnitRepository) SaveVote(msg *modules.Message, voter common.Address) error {

	// type deduct
	VotePayLoad, ok := msg.Payload.(*vote.VoteInfo)
	if !ok {
		return errors.New("not a valid vote payload")
	}

	// save by type
	switch {
	case VotePayLoad.VoteType == vote.TypeMediator:
		//Addresses := common.BytesListToAddressList(VotePayLoad.Contents)
		mediator := common.BytesToAddress(VotePayLoad.Contents)

		if err := unitOp.statedb.AppendVotedMediator(voter, mediator); err != nil {
			return err
		}

	}
	return nil

}

//Get who send this transaction
func (unitOp *UnitRepository) getRequesterAddress(tx *modules.Transaction) (common.Address, error) {
	msg0 := tx.TxMessages[0]
	if msg0.App != modules.APP_PAYMENT {
		return common.Address{}, errors.New("Invalid Tx, first message must be a payment")
	}
	pay := msg0.Payload.(*modules.PaymentPayload)
	utxo, err := unitOp.uxtodb.GetUtxoEntry(pay.Inputs[0].PreviousOutPoint)
	if err != nil {
		return common.Address{}, err
	}
	return tokenengine.GetAddressFromScript(utxo.PkScript)

}

/**
保存单元数据，如果单元的结构基本相同
save genesis unit data
*/
func (unitOp *UnitRepository) SaveUnit(unit *modules.Unit, txpool txspool.ITxPool, isGenesis bool, passed bool) error {

	if unit.UnitSize == 0 || unit.Size() == 0 {
		log.Error("Unit is null")
		return fmt.Errorf("Unit is null")
	}
	// step1 验证 群签名
	// if passed == true , don't validate group sign
	if !passed {
		if no := unitOp.validate.ValidateUnitGroupSign(unit.Header(), isGenesis); no != modules.UNIT_STATE_INVALID_GROUP_SIGNATURE {
			return fmt.Errorf("Validate unit's group sign failed, err number=%d", no)
		}
	}

	// step2. check unit signature, should be compare to mediator list
	if dagconfig.DefaultConfig.WhetherValidateUnitSignature {
		errno := unitOp.validate.ValidateUnitSignature(unit.UnitHeader, isGenesis)
		if int(errno) != modules.UNIT_STATE_VALIDATED && int(errno) != modules.UNIT_STATE_AUTHOR_SIGNATURE_PASSED {
			return fmt.Errorf("Validate unit signature, errno=%d", errno)
		}
	}

	// step3. check unit size
	if unit.UnitSize != unit.Size() {
		log.Info("Validate size", "error", "Size is invalid")
		return modules.ErrUnit(-1)
	}
	// log.Info("===dag ValidateTransactions===")
	// step4. check transactions in unit
	// TODO must recover
	_, isSuccess, err := unitOp.validate.ValidateTransactions(&unit.Txs, isGenesis)
	if err != nil || !isSuccess {
		return fmt.Errorf("Validate unit(%s) transactions failed: %v", unit.UnitHash.String(), err)
	}

	// step5. traverse transactions and save them
	txHashSet := []common.Hash{}
	for txIndex, tx := range unit.Txs {
		var requester common.Address
		var err error
		if txIndex > 0 { //coinbase don't have requester
			requester, err = unitOp.getRequesterAddress(tx)
			if err != nil {
				return err
			}
		}
		txHash := tx.Hash()
		// traverse messages
		for msgIndex, msg := range tx.TxMessages {
			// handle different messages
			switch msg.App {
			case modules.APP_PAYMENT:
				if ok := unitOp.savePaymentPayload(txHash, msg, uint32(msgIndex)); ok != true {
					return fmt.Errorf("Save payment payload error.")
				}
			case modules.APP_CONTRACT_TPL:
				if ok := unitOp.saveContractTpl(unit.UnitHeader.Number, uint32(txIndex), msg); ok != true {
					return fmt.Errorf("Save contract template error.")
				}
			case modules.APP_CONTRACT_DEPLOY:
				if ok := unitOp.saveContractInitPayload(unit.UnitHeader.Number, uint32(txIndex), msg); ok != true {
					return fmt.Errorf("Save contract init payload error.")
				}
			case modules.APP_CONTRACT_INVOKE:
				if ok := unitOp.saveContractInvokePayload(tx, unit.UnitHeader.Number, uint32(txIndex), msg); ok != true {
					return fmt.Errorf("Save contract invode payload error.")
				}
			case modules.APP_CONFIG:
				if ok := unitOp.saveConfigPayload(txHash, msg, unit.UnitHeader.Number, uint32(txIndex)); ok == false {
					return fmt.Errorf("Save contract invode payload error.")
				}
			case modules.APP_VOTE:
				if err = unitOp.SaveVote(msg, requester); err != nil {
					return fmt.Errorf("Save vote payload error.")
				}
			case modules.OP_MEDIATOR_CREATE:
				if ok := unitOp.ApplyOperation(msg, true); ok == false {
					return fmt.Errorf("Apply Mediator Creating Operation error.")
				}

			case modules.APP_CONTRACT_TPL_REQUEST:
				//todo
			case modules.APP_CONTRACT_DEPLOY_REQUEST:
				//todo
			case modules.APP_CONTRACT_STOP_REQUEST:
				//todo
			case modules.APP_CONTRACT_INVOKE_REQUEST:
				//todo
			case modules.APP_SIGNATURE:
				//todo

			case modules.APP_TEXT:
			default:
				return fmt.Errorf("Message type is not supported now: %v", msg.App)
			}
		}
		// step6. save transaction
		if err := unitOp.dagdb.SaveTransaction(tx); err != nil {
			log.Info("Save transaction:", "error", err.Error())
			return err
		}
		txHashSet = append(txHashSet, txHash)
	}

	// step7  send unitHash set to txpool, to update txpool's pending
	if txpool != nil {
		go txpool.SendStoredTxs(txHashSet[:])
	}

	// step8. save unit body, the value only save txs' hash set, and the key is merkle root
	if err := unitOp.dagdb.SaveBody(unit.UnitHash, txHashSet); err != nil {
		log.Info("SaveBody", "error", err.Error())
		return err
	}

	// step 9  save txlookupEntry
	if err := unitOp.dagdb.SaveTxLookupEntry(unit); err != nil {
		log.Info("SaveTxLookupEntry", "error", err.Error())
		return err
	}

	// step10. save unit header
	// key is like "[HEADER_PREFIX][chain index number]_[chain index]_[unit hash]"
	if err := unitOp.dagdb.SaveHeader(unit.UnitHash, unit.UnitHeader); err != nil {
		log.Info("SaveHeader:", "error", err.Error())
		return modules.ErrUnit(-3)
	}
	// step11. save unit hash and chain index relation
	// key is like "[UNIT_HASH_NUMBER][unit_hash]"
	if err := unitOp.dagdb.SaveNumberByHash(unit.UnitHash, unit.UnitHeader.Number); err != nil {
		log.Info("SaveHashNumber:", "error", err.Error())
		return fmt.Errorf("Save unit number hash error, %s", err)
	}
	// step12 SaveHashByNumber
	if err := unitOp.dagdb.SaveHashByNumber(unit.UnitHash, unit.UnitHeader.Number); err != nil {
		log.Info("SaveNumberByHash:", "error", err.Error())
		return fmt.Errorf("Save unit number error, %s", err)
	}

	// step13 update state
	unitOp.dagdb.PutCanonicalHash(unit.UnitHash, unit.NumberU64())
	unitOp.dagdb.PutHeadHeaderHash(unit.UnitHash)
	unitOp.dagdb.PutHeadUnitHash(unit.UnitHash)
	unitOp.dagdb.PutHeadFastUnitHash(unit.UnitHash)
	// todo send message to transaction pool to delete unit's transactions
	return nil
}

/**
保存PaymentPayload
save PaymentPayload data
*/
func (unitOp *UnitRepository) savePaymentPayload(txHash common.Hash, msg *modules.Message, msgIndex uint32) bool {
	// if inputs is none then it is just a normal coinbase transaction
	// otherwise, if inputs' length is 1, and it PreviousOutPoint should be none
	// if this is a create token transaction, the Extra field should be AssetInfo struct's [rlp] encode bytes
	// if this is a create token transaction, should be return a assetid
	// save utxo
	err := unitOp.utxoRepository.UpdateUtxo(txHash, msg, msgIndex)
	if err != nil {
		log.Error("Update utxo failed.", "error", err)
		return false
	}

	return true
}

/**
保存配置交易
save config payload
*/
func (unitOp *UnitRepository) saveConfigPayload(txHash common.Hash, msg *modules.Message, height modules.ChainIndex, txIndex uint32) bool {
	var pl interface{}
	pl = msg.Payload
	payload, ok := pl.(*modules.ConfigPayload)
	if ok == false {
		return false
	}
	version := modules.StateVersion{
		Height:  height,
		TxIndex: txIndex,
	}
	if err := unitOp.statedb.SaveConfig(payload.ConfigSet, &version); err != nil {
		errMsg := fmt.Sprintf("To save config payload error: %s", err)
		log.Error(errMsg)
		return false
	}
	return true
}

/**
保存合约调用状态
To save contract invoke state
*/
func (unitOp *UnitRepository) saveContractInvokePayload(tx *modules.Transaction, height modules.ChainIndex, txIndex uint32, msg *modules.Message) bool {
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
		// 	unitOp.statedb.SaveContractState(addr, ws.Key, ws.Value, version)
		// }
		//@jay
		// contractId is never nil.
		if payload.ContractId != nil {
			//addr, _ := getRequesterAddress(tx)
			// contractid
			unitOp.statedb.SaveContractState(payload.ContractId, ws.Key, ws.Value, version)
		}

		// save new state to database
		// if unitOp.updateState(payload.ContractId, ws.Key, version, ws.Value) != true {
		// 	continue
		// }
	}
	return true
}

/**
保存合约初始化状态
To save contract init state
*/
func (unitOp *UnitRepository) saveContractInitPayload(height modules.ChainIndex, txIndex uint32, msg *modules.Message) bool {
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
		if unitOp.updateState(payload.ContractId, ws.Key, version, ws.Value) != true {
			continue
		}
	}
	//addr := common.NewAddress(payload.ContractId, common.ContractHash)
	// save contract name
	if unitOp.statedb.SaveContractState(payload.ContractId, "ContractName", payload.Name, version) != nil {
		return false
	}
	// save contract jury list
	if unitOp.statedb.SaveContractState(payload.ContractId, "ContractJury", payload.Jury, version) != nil {
		return false
	}
	return true
}

/**
保存合约模板代码
To save contract template code
*/
func (unitOp *UnitRepository) saveContractTpl(height modules.ChainIndex, txIndex uint32, msg *modules.Message) bool {
	var pl interface{}
	pl = msg.Payload
	payload, ok := pl.(*modules.ContractTplPayload)
	if ok == false {
		log.Error("saveContractTpl", "error", "payload is not ContractTplPayload")
		return false
	}

	// step1. generate version for every contract template
	version := &modules.StateVersion{
		Height:  height,
		TxIndex: txIndex,
	}

	// step2. save contract template bytecode data
	if err := unitOp.statedb.SaveContractTemplate(payload.TemplateId, payload.Bytecode, version.Bytes()); err != nil {
		log.Error("SaveContractTemplate", "error", err.Error())
		return false
	}
	// step3. save contract template name, path, Memory
	if err := unitOp.statedb.SaveContractTemplateState(payload.TemplateId, modules.FIELD_TPL_NAME, payload.Name, version); err != nil {
		log.Error("SaveContractTemplateState when save name", "error", err.Error())
		return false
	}
	if err := unitOp.statedb.SaveContractTemplateState(payload.TemplateId, modules.FIELD_TPL_PATH, payload.Path, version); err != nil {
		log.Error("SaveContractTemplateState when save path", "error", err.Error())
		return false
	}
	if err := unitOp.statedb.SaveContractTemplateState(payload.TemplateId, modules.FIELD_TPL_Memory, payload.Memory, version); err != nil {
		log.Error("SaveContractTemplateState when save memory", "error", err.Error())
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

func CreateCoinbase(addr *common.Address, income uint64, addition map[common.Address]*modules.Addition, asset *modules.Asset, t time.Time) (*modules.Transaction, error) {
	//创建合约保证金币龄的奖励output
	payload := modules.PaymentPayload{}
	if len(addition) != 0 {
		for k, v := range addition {
			script := tokenengine.GenerateLockScript(k)
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
		}
	}
	// setp1. create P2PKH script
	script := tokenengine.GenerateP2PKHLockScript(addr.Bytes())
	// step. compute total income
	totalIncome := int64(income) + int64(ComputeInterest())
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
	// step3. create message
	msg := &modules.Message{
		App:     modules.APP_PAYMENT,
		Payload: &payload,
	}
	// step4. create coinbase
	coinbase := new(modules.Transaction)
	//coinbase := modules.Transaction{
	//	TxMessages: []modules.Message{msg},
	//}
	coinbase.TxMessages = append(coinbase.TxMessages, msg)
	// coinbase.CreationDate = coinbase.CreateDate()
	//coinbase.TxHash = coinbase.Hash()

	return coinbase, nil
}

/**
删除合约状态
To delete contract state
*/
func (unitOp *UnitRepository) deleteContractState(contractID []byte, field string) {
	oldKeyPrefix := fmt.Sprintf("%s%s^*^%s",
		constants.CONTRACT_STATE_PREFIX,
		hexutil.Encode(contractID[:]),
		field)
	data := unitOp.statedb.GetPrefix([]byte(oldKeyPrefix))
	for k := range data {
		if err := unitOp.statedb.DeleteState([]byte(k)); err != nil {
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
func (unitOp *UnitRepository) updateState(contractID []byte, key string, version *modules.StateVersion, val interface{}) bool {
	delState, isDel := val.(modules.DelContractState)
	if isDel {
		if delState.IsDelete == false {
			return true
		}
		// delete old state from database
		unitOp.deleteContractState(contractID, key)

	} else {
		// delete old state from database
		unitOp.deleteContractState(contractID, key)
		// insert new state
		key := fmt.Sprintf("%s%s^*^%s^*^%s",
			constants.CONTRACT_STATE_PREFIX,
			hexutil.Encode(contractID[:]),
			key,
			version.String())
		// addr := common.NewAddress(contractID, common.ContractHash)
		if err := unitOp.statedb.SaveContractState(contractID, key, val, version); err != nil {
			log.Error("Save state", "error", err.Error())
			return false
		}
	}
	return true
}

func IsGenesis(hash common.Hash) bool {
	genHash := common.HexToHash(dagconfig.DefaultConfig.GenesisHash)
	return genHash == hash
}
