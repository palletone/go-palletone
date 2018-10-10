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
	"github.com/palletone/go-palletone/common/rlp"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/core/accounts/keystore"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/errors"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/storage"
	"github.com/palletone/go-palletone/dag/txspool"
	"github.com/palletone/go-palletone/tokenengine"
)

type IUnitRepository interface {
	GetGenesisUnit(index uint64) (*modules.Unit, error)
	GenesisHeight() modules.ChainIndex
	SaveUnit(unit modules.Unit, isGenesis bool) error
	CreateUnit(mAddr *common.Address, txpool *txspool.TxPool, ks *keystore.KeyStore, t time.Time) ([]modules.Unit, error)
}
type UnitRepository struct {
	dagdb          storage.IDagDb
	idxdb          storage.IIndexDb
	uxtodb         storage.IUtxoDb
	statedb        storage.IStateDb
	validate       Validator
	utxoRepository IUtxoRepository
	logger log.ILogger
}

func NewUnitRepository(dagdb storage.IDagDb, idxdb storage.IIndexDb, utxodb storage.IUtxoDb, statedb storage.IStateDb,l log.ILogger) *UnitRepository {
	val := NewValidate(dagdb, utxodb, statedb)
	utxoRep := NewUtxoRepository(utxodb, idxdb, statedb,l)
	return &UnitRepository{dagdb: dagdb, idxdb: idxdb, uxtodb: utxodb, statedb: statedb, validate: val, utxoRepository: utxoRep}
}
func NewUnitRepository4Db(db ptndb.Database,l log.ILogger) *UnitRepository {
	dagdb := storage.NewDagDb(db)
	utxodb := storage.NewUtxoDb(db,l)
	statedb := storage.NewStateDb(db)
	idxdb := storage.NewIndexDb(db)
	val := NewValidate(dagdb, utxodb, statedb)
	utxoRep := NewUtxoRepository(utxodb, idxdb, statedb,l)
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

	unit.UnitHeader.Authors = &modules.Authentifier{
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

/**
创建单元
create common unit
@param mAddr is minner addr
return: correct if error is nil, and otherwise is incorrect
*/
func (unitOp *UnitRepository) CreateUnit(mAddr *common.Address, txpool *txspool.TxPool, ks *keystore.KeyStore, t time.Time) ([]modules.Unit, error) {
	if txpool == nil || mAddr == nil || ks == nil {
		return nil, fmt.Errorf("Create unit: nil address or txspool is not allowed")
	}
	units := []modules.Unit{}
	// step1. get mediator responsible for asset (for now is ptn)
	bAsset := unitOp.statedb.GetConfig([]byte(modules.FIELD_GENESIS_ASSET))
	if len(bAsset) <= 0 {
		return nil, fmt.Errorf("Create unit error: query asset info empty")
	}
	var asset modules.Asset
	if err := rlp.DecodeBytes(bAsset, &asset); err != nil {
		return nil, fmt.Errorf("Create unit: %s", err.Error())
	}
	// step2. compute chain height
	index := uint64(1)
	isMain := true
	chainIndex := modules.ChainIndex{AssetID: asset.AssetId, IsMain: isMain, Index: index}

	// step3. get transactions from txspool
	poolTxs, _ := txpool.GetSortedTxs()
	// step4. compute minner income: transaction fees + interest
	//txs := txspool.PoolTxstoTxs(poolTxs)
	fees, err := unitOp.utxoRepository.ComputeFees(poolTxs)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	coinbase, err := CreateCoinbase(mAddr, fees, &asset, t)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	txs := modules.Transactions{coinbase}
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

	// step5. transactions merkle root
	root := core.DeriveSha(txs)

	// step6. generate genesis unit header
	header := modules.Header{
		AssetIDs: []modules.IDType16{},
		Number:   chainIndex,
		TxRoot:   root,
		//		Creationdate: time.Now().Unix(),
	}
	header.AssetIDs = append(header.AssetIDs, asset.AssetId)
	unit := modules.Unit{}
	unit.UnitHeader = &header
	// step7. copy txs
	unit.CopyBody(txs)
	// step8. set size
	unit.UnitSize = unit.Size()
	units = append(units, unit)
	return units, nil
}

/**
从leveldb中查询GenesisUnit信息
To get genesis unit info from leveldb
*/
func (unitOp *UnitRepository) GetGenesisUnit(index uint64) (*modules.Unit, error) {
	// unit key: [HEADER_PREFIX][chain index number]_[chain index]_[unit hash]
	key := fmt.Sprintf("%s%v_", storage.HEADER_PREFIX, index)
	// encNum := ptndb.EncodeBlockNumber(index)
	// key := append(storage.HEADER_PREFIX, encNum...)

	//if memdb, ok := db.(*ptndb.MemDatabase); ok {
	//	hash, err := memdb.Get([]byte(key))
	//	if err != nil {
	//		return nil, err
	//	}
	//	var h common.Hash
	//	h.SetBytes(hash)
	//	unit := unitOp.dagdb.GetUnit( h)
	//	return unit, nil
	//} else if _, ok := db.(*ptndb.LDBDatabase); ok {
	data := unitOp.dagdb.GetPrefix([]byte(key))
	if len(data) > 1 {
		return nil, fmt.Errorf("multiple genesis unit")
	} else if len(data) <= 0 {
		return nil, errors.ErrNotFound
	}
	for _, v := range data {
		// get unit header
		var uHeader modules.Header
		if err := rlp.DecodeBytes([]byte(v), &uHeader); err != nil {
			return nil, fmt.Errorf("Get genesis unit header:%s", err.Error())
		}
		// generate unit
		unit := modules.Unit{
			UnitHeader: &uHeader,
		}
		// compute unit hash
		unit.UnitHash = unit.Hash()
		// get transaction list
		txs, err := unitOp.dagdb.GetUnitTransactions(unit.UnitHash)
		if err != nil {
			//TODO xiaozhi
			return nil, fmt.Errorf("Get genesis unit transactions: %s", err.Error())
		}
		unit.Txs = txs
		unit.UnitSize = unit.Size()
		return &unit, nil
		//}
	}
	return nil, nil
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

	confPay.ConfigSet = []modules.PayloadMapStruct{}

	tt := reflect.TypeOf(*genesisConf)
	vv := reflect.ValueOf(*genesisConf)

	for i := 0; i < tt.NumField(); i++ {
		if strings.Compare(tt.Field(i).Name, "SystemConfig") == 0 {
			t := reflect.TypeOf(genesisConf.SystemConfig)
			v := reflect.ValueOf(genesisConf.SystemConfig)
			for k := 0; k < t.NumField(); k++ {
				sk := t.Field(k).Name
				if strings.Contains(sk, "Initial") {
					sk = strings.Replace(sk, "Initial", "", -1)
				}

				confPay.ConfigSet = append(confPay.ConfigSet, modules.PayloadMapStruct{Key: sk, Value: modules.ToPayloadMapValueBytes(v.Field(k).Interface())})
			}
		} else {
			sk := tt.Field(i).Name
			if strings.Contains(sk, "Initial") {
				sk = strings.Replace(sk, "Initial", "", -1)
			}
			confPay.ConfigSet = append(confPay.ConfigSet, modules.PayloadMapStruct{Key: sk, Value: modules.ToPayloadMapValueBytes(vv.Field(i).Interface())})
		}
	}

	confPay.ConfigSet = append(confPay.ConfigSet, modules.PayloadMapStruct{Key: modules.FIELD_GENESIS_ASSET, Value: modules.ToPayloadMapValueBytes(*asset)})

	return confPay, nil
}

/**
保存单元数据，如果单元的结构基本相同
save genesis unit data
*/
func (unitOp *UnitRepository) SaveUnit(unit modules.Unit, isGenesis bool) error {

	if unit.UnitSize == 0 || unit.Size() == 0 {
		log.Error("Unit is null")
		return fmt.Errorf("Unit is null")
	}
	// step1. check unit signature, should be compare to mediator list
	if dagconfig.DefaultConfig.WhetherValidateUnitSignature {
		errno := unitOp.validate.ValidateUnitSignature(unit.UnitHeader, isGenesis)
		if int(errno) != modules.UNIT_STATE_VALIDATED && int(errno) != modules.UNIT_STATE_AUTHOR_SIGNATURE_PASSED {
			return fmt.Errorf("Validate unit signature, errno=%d", errno)
		}
	}

	// step2. check unit size
	if unit.UnitSize != unit.Size() {
		log.Info("Validate size", "error", "Size is invalid")
		return modules.ErrUnit(-1)
	}
	// step3. check transactions in unit
	_, isSuccess, err := unitOp.validate.ValidateTransactions(&unit.Txs, isGenesis)
	if isSuccess != true {
		return fmt.Errorf("Validate unit(%s) transactions failed: %v", unit.UnitHash.String(), err)
	}
	// step4. save unit header
	// key is like "[HEADER_PREFIX][chain index number]_[chain index]_[unit hash]"
	if err := unitOp.dagdb.SaveHeader(unit.UnitHash, unit.UnitHeader); err != nil {
		log.Info("SaveHeader:", "error", err.Error())
		return modules.ErrUnit(-3)
	}
	// step5. save unit hash and chain index relation
	// key is like "[UNIT_HASH_NUMBER][unit_hash]"
	if err := unitOp.dagdb.SaveNumberByHash(unit.UnitHash, unit.UnitHeader.Number); err != nil {
		log.Info("SaveHashNumber:", "error", err.Error())
		return fmt.Errorf("Save unit number hash error, %s", err)
	}
	if err := unitOp.dagdb.SaveHashByNumber(unit.UnitHash, unit.UnitHeader.Number); err != nil {
		log.Info("SaveNumberByHash:", "error", err.Error())
		return fmt.Errorf("Save unit number error, %s", err)
	}
	// step6. traverse transactions and save them
	txHashSet := []common.Hash{}
	for txIndex, tx := range unit.Txs {
		// traverse messages
		for msgIndex, msg := range tx.TxMessages {
			// handle different messages
			switch msg.App {
			case modules.APP_PAYMENT:
				if ok := unitOp.savePaymentPayload(tx.TxHash, msg, uint32(msgIndex)); ok != true {
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
				if ok := unitOp.saveContractInvokePayload(unit.UnitHeader.Number, uint32(txIndex), msg); ok != true {
					return fmt.Errorf("Save contract invode payload error.")
				}
			case modules.APP_CONFIG:
				if ok := unitOp.saveConfigPayload(tx.TxHash, msg, unit.UnitHeader.Number, uint32(txIndex)); ok == false {
					return fmt.Errorf("Save contract invode payload error.")
				}
			case modules.APP_TEXT:
			default:
				return fmt.Errorf("Message type is not supported now: %v", msg.App)
			}
		}
		// step7. save transaction
		if err := unitOp.dagdb.SaveTransaction(tx); err != nil {
			log.Info("Save transaction:", "error", err.Error())
			return err
		}
		txHashSet = append(txHashSet, tx.TxHash)
	}

	// step8. save unit body, the value only save txs' hash set, and the key is merkle root
	if err := unitOp.dagdb.SaveBody(unit.UnitHash, txHashSet); err != nil {
		log.Info("SaveBody", "error", err.Error())
		return err
	}
	// step 10  save txlookupEntry
	if err := unitOp.dagdb.SaveTxLookupEntry(&unit); err != nil {
		return err
	}
	// update state
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
func (unitOp *UnitRepository) saveContractInvokePayload(height modules.ChainIndex, txIndex uint32, msg *modules.Message) bool {
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
		// save new state to database
		if unitOp.updateState(payload.ContractId, ws.Key, version, ws.Value) != true {
			continue
		}
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

func CreateCoinbase(addr *common.Address, income uint64, asset *modules.Asset, t time.Time) (*modules.Transaction, error) {
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
	payload := modules.PaymentPayload{
		Input:  []*modules.Input{&input},
		Output: []*modules.Output{&output},
	}
	// step3. create message
	msg := &modules.Message{
		App:     modules.APP_PAYMENT,
		Payload: &payload,
	}
	// step4. create coinbase
	var coinbase modules.Transaction
	//coinbase := modules.Transaction{
	//	TxMessages: []modules.Message{msg},
	//}
	coinbase.TxMessages = append(coinbase.TxMessages, msg)
	// coinbase.CreationDate = coinbase.CreateDate()
	coinbase.TxHash = coinbase.Hash()

	return &coinbase, nil
}

/**
删除合约状态
To delete contract state
*/
func (unitOp *UnitRepository) deleteContractState(contractID []byte, field string) {
	oldKeyPrefix := fmt.Sprintf("%s%s^*^%s",
		storage.CONTRACT_STATE_PREFIX,
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
			storage.CONTRACT_STATE_PREFIX,
			hexutil.Encode(contractID[:]),
			key,
			version.String())
		if err := unitOp.statedb.SaveContractState(contractID, key, val, version); err != nil {
			log.Error("Save state", "error", err.Error())
			return false
		}
	}
	return true
}
