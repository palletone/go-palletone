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
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/hexutil"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/common/rlp"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/core/accounts/keystore"
	"github.com/palletone/go-palletone/dag/asset"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/storage"
	"github.com/palletone/go-palletone/dag/txspool"
	"github.com/palletone/go-palletone/tokenengine"
	"math/big"
	"time"
)

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
func NewGenesisUnit(txs modules.Transactions, time int64) (*modules.Unit, error) {
	gUnit := modules.Unit{}

	// genesis unit asset id
	gAssetID := asset.NewAsset()

	// genesis unit height
	chainIndex := modules.ChainIndex{AssetID: gAssetID, IsMain: true, Index: 0}

	// transactions merkle root
	root := core.DeriveSha(txs)

	// generate genesis unit header
	header := modules.Header{
		AssetIDs:     []modules.IDType16{gAssetID},
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

// @author Albert·Gou
func StoreUnit(db ptndb.Database, unit *modules.Unit) error {
	err := SaveUnit(db, *unit, false)

	if err != nil {
		log.Error(fmt.Sprintf("%v", err))
		return err
	}

	// 此处应当更新DB中的全局属性
	//	go storage.StoreDynGlobalProp(dgp)

	return nil
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
		Address: signer.String(),
		R:       r,
		S:       s,
		V:       v,
	}
	// to set witness list, should be creator himself
	var authentifier modules.Authentifier
	authentifier.Address = signer.String()
	unit.UnitHeader.Witness = append(unit.UnitHeader.Witness, &authentifier)

	return unit, nil
}

/**
创建单元
create common unit
@param mAddr is minner addr
return: correct if error is nil, and otherwise is incorrect
*/
func CreateUnit(db ptndb.Database, mAddr *common.Address, txpool *txspool.TxPool, ks *keystore.KeyStore, t time.Time) ([]modules.Unit, error) {
	if txpool == nil || mAddr == nil || ks == nil {
		return nil, fmt.Errorf("Create unit: nil address or txspool is not allowed")
	}

	units := []modules.Unit{}
	// step1. get mediator responsible for asset (for now is ptn)
	bAsset := GetConfig(db, []byte("GenesisAsset"))
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
	chainIndex := modules.ChainIndex{AssetID: asset.AssertId, IsMain: isMain, Index: index}

	// step3. get transactions from txspool
	poolTxs, _ := txpool.GetSortedTxs()
	// step4. compute minner income: transaction fees + interest
	//txs := txspool.PoolTxstoTxs(poolTxs)
	fees, err := ComputeFees(db, poolTxs)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	coinbase, err := createCoinbase(mAddr, fees, &asset, ks, t)
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
		AssetIDs: []modules.IDType16{asset.AssertId},
		Number:   chainIndex,
		TxRoot:   root,
		//		Creationdate: time.Now().Unix(),
	}

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
func GetGenesisUnit(db ptndb.Database, index uint64) (*modules.Unit, error) {
	// unit key: [HEADER_PREFIX][chain index number]_[chain index]_[unit hash]
	key := fmt.Sprintf("%s%v_", storage.HEADER_PREFIX, index)
	data := storage.GetPrefix(db, []byte(key))
	if len(data) > 1 {
		return nil, fmt.Errorf("multiple genesis unit")
	} else if len(data) <= 0 {
		return nil, nil
	}
	for k, v := range data {
		sk := string(k[len(storage.HEADER_PREFIX):])
		// get index
		skArr := strings.Split(sk, "_")
		if len(skArr) != 3 {
			return nil, fmt.Errorf("split genesis key error")
		}
		// get unit hash
		uHash := common.Hash{}
		uHash.SetString(skArr[2])
		// get unit header
		var uHeader modules.Header
		if err := rlp.DecodeBytes([]byte(v), &uHeader); err != nil {
			return nil, fmt.Errorf("Get genesis unit header:%s", err.Error())
		}
		// get transaction list
		txs, err := GetUnitTransactions(db, uHash)
		if err != nil {
			return nil, fmt.Errorf("Get genesis unit transactions: %s", err.Error())
		}
		// generate unit
		unit := modules.Unit{
			UnitHeader: &uHeader,
			UnitHash:   uHash,
			Txs:        txs,
		}
		unit.UnitSize = unit.Size()
		return &unit, nil
	}
	return nil, nil
}

/**
获取创世单元的高度
To get genesis unit height
*/
func GenesisHeight(db ptndb.Database) modules.ChainIndex {
	unit, err := GetGenesisUnit(db, 0)
	if unit == nil || err != nil {
		return modules.ChainIndex{}
	}
	return unit.UnitHeader.Number
}

func GetUnitTransactions(db ptndb.Database, unitHash common.Hash) (modules.Transactions, error) {
	txs := modules.Transactions{}
	// get body data: transaction list.
	// if getbody return transactions list, then don't range txHashlist.
	txHashList, err := storage.GetBody(db, unitHash)
	if err != nil {
		return nil, err
	}
	// get transaction by tx'hash.
	for _, txHash := range txHashList {
		tx, _, _, _ := storage.GetTransaction(db, txHash)
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

	confPay.ConfigSet = append(confPay.ConfigSet, modules.PayloadMapStruct{Key: "GenesisAsset", Value: modules.ToPayloadMapValueBytes(asset)})

	return confPay, nil
}

/**
保存单元数据，如果单元的结构基本相同
save genesis unit data
*/
func SaveUnit(db ptndb.Database, unit modules.Unit, isGenesis bool) error {

	if unit.UnitSize == 0 || unit.Size() == 0 {
		log.Error("Unit is null")
		return fmt.Errorf("Unit is null")
	}
	// step1. check unit signature, should be compare to mediator list
	errno := ValidateUnitSignature(db, unit.UnitHeader, isGenesis)
	if int(errno) != modules.UNIT_STATE_VALIDATED && int(errno) != modules.UNIT_STATE_AUTHOR_SIGNATURE_PASSED {
		return fmt.Errorf("Validate unit signature, errno=%d", errno)
	}

	// step2. check unit size
	if unit.UnitSize != unit.Size() {
		log.Info("Validate size", "error", "Size is invalid")
		return modules.ErrUnit(-1)
	}
	// step3. check transactions in unit
	_, isSuccess, err := ValidateTransactions(db, &unit.Txs, isGenesis)
	if isSuccess != true {
		return fmt.Errorf("Validate unit(%s) transactions failed: %v", unit.UnitHash.String(), err)
	}
	// step4. save unit header
	// key is like "[HEADER_PREFIX][chain index number]_[chain index]_[unit hash]"
	if err := storage.SaveHeader(db, unit.UnitHash, unit.UnitHeader); err != nil {
		log.Info("SaveHeader:", "error", err.Error())
		return modules.ErrUnit(-3)
	}
	// step5. save unit hash and chain index relation
	// key is like "[UNIT_HASH_NUMBER][unit_hash]"
	if err := storage.SaveNumberByHash(db, unit.UnitHash, unit.UnitHeader.Number); err != nil {
		log.Info("SaveHashNumber:", "error", err.Error())
		return fmt.Errorf("Save unit hash error")
	}
	if err := storage.SaveHashByNumber(db, unit.UnitHash, unit.UnitHeader.Number); err != nil {
		log.Info("SaveNumberByHash:", "error", err.Error())
		return fmt.Errorf("Save unit number error")
	}
	// step6. traverse transactions and save them
	txHashSet := []common.Hash{}
	for txIndex, tx := range unit.Txs {
		// traverse messages
		for msgIndex, msg := range tx.TxMessages {
			// handle different messages
			switch msg.App {
			case modules.APP_PAYMENT:
				if ok := savePaymentPayload(db, tx.TxHash, msg, uint32(msgIndex)); ok != true {
					log.Info("Save payment payload error.")
					return fmt.Errorf("Save payment payload error.")
				}
			case modules.APP_CONTRACT_TPL:
				if ok := saveContractTpl(db, unit.UnitHeader.Number, uint32(txIndex), msg); ok != true {
					log.Info("Save contract template error.")
					return fmt.Errorf("Save contract template error.")
				}
			case modules.APP_CONTRACT_DEPLOY:
				if ok := saveContractInitPayload(db, unit.UnitHeader.Number, uint32(txIndex), msg); ok != true {
					log.Info("Save contract init payload error.")
					return fmt.Errorf("Save contract init payload error.")
				}
			case modules.APP_CONTRACT_INVOKE:
				if ok := saveContractInvokePayload(db, unit.UnitHeader.Number, uint32(txIndex), msg); ok != true {
					log.Info("Save contract invode payload error.")
					return fmt.Errorf("Save contract invode payload error.")
				}
			case modules.APP_CONFIG:
				if ok := saveConfigPayload(db, tx.TxHash, msg, unit.UnitHeader.Number, uint32(txIndex)); ok == false {
					log.Info("Save contract invode payload error.")
					return fmt.Errorf("Save contract invode payload error.")
				}
			case modules.APP_TEXT:
			default:
				log.Info("Message type is not supported now")
				return fmt.Errorf("Message type is not supported now: %v", msg.App)
			}
		}
		// step7. save transaction
		if err := storage.SaveTransaction(db, tx); err != nil {
			log.Info("Save transaction:", "error", err.Error())
			return err
		}
		txHashSet = append(txHashSet, tx.TxHash)
	}

	// step8. save unit body, the value only save txs' hash set, and the key is merkle root
	if err := storage.SaveBody(db, unit.UnitHash, txHashSet); err != nil {
		log.Info("SaveBody", "error", err.Error())
		return err
	}
	// step 10  save txlookupEntry
	if err := storage.SaveTxLookupEntry(db, &unit); err != nil {
		return err
	}
	// update state
	storage.PutCanonicalHash(db, unit.UnitHash, unit.NumberU64())
	storage.PutHeadHeaderHash(db, unit.UnitHash)
	storage.PutHeadUnitHash(db, unit.UnitHash)
	storage.PutHeadFastUnitHash(db, unit.UnitHash)
	// todo send message to transaction pool to delete unit's transactions
	return nil
}

/**
保存PaymentPayload
save PaymentPayload data
*/
func savePaymentPayload(db ptndb.Database, txHash common.Hash, msg *modules.Message, msgIndex uint32) bool {
	// if inputs is none then it is just a normal coinbase transaction
	// otherwise, if inputs' length is 1, and it PreviousOutPoint should be none
	// if this is a create token transaction, the Extra field should be AssetInfo struct's [rlp] encode bytes
	// if this is a create token transaction, should be return a assetid
	var pl interface{}
	pl = msg.Payload
	_, ok := pl.(modules.PaymentPayload)
	if ok == false {
		return false
	}

	// save utxo
	UpdateUtxo(db, txHash, msg, msgIndex)
	return true
}

/**
保存配置交易
save config payload
*/
func saveConfigPayload(db ptndb.Database, txHash common.Hash, msg *modules.Message, height modules.ChainIndex, txIndex uint32) bool {
	var pl interface{}
	pl = msg.Payload
	payload, ok := pl.(modules.ConfigPayload)
	if ok == false {
		return false
	}
	version := modules.StateVersion{
		Height:  height,
		TxIndex: txIndex,
	}
	if err := SaveConfig(db, payload.ConfigSet, &version); err != nil {
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
func saveContractInvokePayload(db ptndb.Database, height modules.ChainIndex, txIndex uint32, msg *modules.Message) bool {
	var pl interface{}
	pl = msg.Payload
	payload, ok := pl.(modules.ContractInvokePayload)
	if ok == false {
		return false
	}
	// save contract state
	// key: [CONTRACT_STATE_PREFIX][contract id]_[field name]_[state version]
	for _, ws := range payload.WriteSet {
		version := modules.StateVersion{
			Height:  height,
			TxIndex: txIndex,
		}
		// save new state to database
		if updateState(db, payload.ContractId, ws.Key, version, ws.Value) != true {
			continue
		}
	}
	return true
}

/**
保存合约初始化状态
To save contract init state
*/
func saveContractInitPayload(db ptndb.Database, height modules.ChainIndex, txIndex uint32, msg *modules.Message) bool {
	var pl interface{}
	pl = msg.Payload
	payload, ok := pl.(modules.ContractDeployPayload)
	if ok == false {
		return false
	}

	// save contract state
	// key: [CONTRACT_STATE_PREFIX][contract id]_[field name]_[state version]
	version := modules.StateVersion{
		Height:  height,
		TxIndex: txIndex,
	}
	for _, ws := range payload.WriteSet {
		// save new state to database
		if updateState(db, payload.ContractId, ws.Key, version, ws.Value) != true {
			continue
		}
	}
	// save contract name
	if !storage.SaveContractState(db, storage.CONTRACT_STATE_PREFIX, payload.ContractId, "ContractName", payload.Name, version) {
		return false
	}
	// save contract jury list
	if !storage.SaveContractState(db, storage.CONTRACT_STATE_PREFIX, payload.ContractId, "ContractJury", payload.Jury, version) {
		return false
	}
	return true
}

/**
保存合约模板代码
To save contract template code
*/
func saveContractTpl(db ptndb.Database, height modules.ChainIndex, txIndex uint32, msg *modules.Message) bool {
	var pl interface{}
	pl = msg.Payload
	payload, ok := pl.(modules.ContractTplPayload)
	if ok == false {
		return false
	}

	// step1. generate version for every contract template
	version := modules.StateVersion{
		Height:  height,
		TxIndex: txIndex,
	}

	// step2. save contract template bytecode data
	// key:[CONTRACT_TPL][Template id]_bytecode_[template version]
	key := fmt.Sprintf("%s%s^*^bytecode^*^%s",
		storage.CONTRACT_TPL,
		hexutil.Encode(payload.TemplateId[:]),
		version.String())

	if err := storage.Store(db, key, payload.Bytecode); err != nil {
		log.Error("Save contract template", "error", err.Error())
		return false
	}
	// step3. save contract template name, path, Memery
	if !storage.SaveContractState(db, storage.CONTRACT_TPL, payload.TemplateId, "tplname", payload.Name, version) {
		return false
	}
	if !storage.SaveContractState(db, storage.CONTRACT_TPL, payload.TemplateId, "tplpath", payload.Path, version) {
		return false
	}
	if !storage.SaveContractState(db, storage.CONTRACT_TPL, payload.TemplateId, "tplmemory", payload.Memery, version) {
		return false
	}
	return true
}

/**
从levedb中根据ChainIndex获得Unit信息
To get unit information by its ChainIndex
*/
func QueryUnitByChainIndex(db ptndb.Database, number modules.ChainIndex) *modules.Unit {
	return storage.GetUnitFormIndex(db, number)
}

/**
创建coinbase交易
To create coinbase transaction
*/

func createCoinbase(addr *common.Address, income uint64, asset *modules.Asset, ks *keystore.KeyStore, t time.Time) (*modules.Transaction, error) {
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
		Asset:    *asset,
		PkScript: script,
	}
	payload := modules.PaymentPayload{
		Input:  []*modules.Input{&input},
		Output: []*modules.Output{&output},
	}
	// step3. create message
	msg := modules.Message{
		App:     modules.APP_PAYMENT,
		Payload: payload,
	}
	// step4. create coinbase
	coinbase := modules.Transaction{
		TxMessages: []*modules.Message{&msg},
	}
	// coinbase.CreationDate = coinbase.CreateDate()
	coinbase.TxHash = coinbase.Hash()

	return &coinbase, nil
}

/**
删除合约状态
To delete contract state
*/
func deleteContractState(db ptndb.Database, contractID []byte, field string) {
	oldKeyPrefix := fmt.Sprintf("%s%s^*^%s",
		storage.CONTRACT_STATE_PREFIX,
		hexutil.Encode(contractID[:]),
		field)
	data := storage.GetPrefix(db, []byte(oldKeyPrefix))
	for k := range data {
		if err := storage.Delete(db, []byte(k)); err != nil {
			log.Error("Delete contract state", "error", err.Error())
			continue
		}
	}
}

/**
签名交易
To Sign transaction
*/
func SignTransaction(txHash common.Hash, addr *common.Address, ks *keystore.KeyStore) (*modules.Authentifier, error) {
	R, S, V, err := ks.SigTX(txHash, *addr)
	if err != nil {
		msg := fmt.Sprintf("Sign transaction error: %s", err)
		log.Error(msg)
		return nil, nil
	}
	sig := modules.Authentifier{
		Address: addr.String(),
		R:       R,
		S:       S,
		V:       V,
	}
	return &sig, nil
}

/**
保存contract state
To save contract state
*/
func updateState(db ptndb.Database, contractID []byte, key string, version modules.StateVersion, val interface{}) bool {
	delState, isDel := val.(modules.DelContractState)
	if isDel {
		if delState.IsDelete == false {
			return true
		}
		// delete old state from database
		deleteContractState(db, contractID, key)

	} else {
		// delete old state from database
		deleteContractState(db, contractID, key)
		// insert new state
		key := fmt.Sprintf("%s%s^*^%s^*^%s",
			storage.CONTRACT_STATE_PREFIX,
			hexutil.Encode(contractID[:]),
			key,
			version.String())
		if err := storage.Store(db, key, val); err != nil {
			log.Error("Save state", "error", err.Error())
			return false
		}
	}
	return true
}
