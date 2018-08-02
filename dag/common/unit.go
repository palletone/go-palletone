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
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unsafe"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/crypto"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/rlp"
	"github.com/palletone/go-palletone/common/util"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/core/accounts/keystore"
	"github.com/palletone/go-palletone/dag/asset"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/storage"
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

//  last unit
func CurrentUnit() *modules.Unit {
	return &modules.Unit{}
}

// get unit
func GetUnit(hash *common.Hash) *modules.Unit {
	unit_bytes, err := storage.Get(append(storage.UNIT_PREFIX, hash.Bytes()...))
	if err != nil {
		return nil
	}
	unit := new(modules.Unit)
	if err := json.Unmarshal(unit_bytes, &unit); err == nil {
		return unit
	}
	return nil
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
	if len(txs) > 0 {
		gUnit.Txs = make([]*modules.Transaction, len(txs))
		for i, pTx := range txs {
			tx := modules.Transaction{
				AccountNonce: pTx.AccountNonce,
				TxHash:       pTx.TxHash,
				From:         pTx.From,
				Excutiontime: pTx.Excutiontime,
				Memery:       pTx.Memery,
				CreationDate: pTx.CreationDate,
				TxFee:        pTx.TxFee,
				Txsize:       pTx.Txsize,
			}
			if len(pTx.TxMessages) > 0 {
				tx.TxMessages = make([]modules.Message, len(pTx.TxMessages))
				for j := 0; j < len(pTx.TxMessages); j++ {
					tx.TxMessages[j] = pTx.TxMessages[j]
				}
			}
			gUnit.Txs[i] = &tx
		}
	}
	// set unit size
	gUnit.UnitSize = gUnit.Size()
	// set unit hash
	gUnit.UnitHash = gUnit.Hash()
	return &gUnit, nil
}

// @author Albert·Gou
func StoreUnit(unit *modules.Unit) error {
	err := SaveUnit(*unit, false)

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
// modify by Albert·Gou
func CreateUnit(mAddr *common.Address) ([]modules.Unit, error) {
	units := []modules.Unit{}
	// get mediator responsible for asset id
	assetID := modules.IDType16{}
	// get the chain last index
	index := uint64(0)
	// the unit is on main or not
	isMain := true
	// genesis unit height
	chainIndex := modules.ChainIndex{AssetID: assetID, IsMain: isMain, Index: index}

	// 交易池应该提供的是静态接口，不需要调用的时候去实例化
	txs := modules.Transactions{}

	/**
	需要根据交易中涉及到的token类型来确定交易打包到哪个区块
	如果交易中涉及到其他币种的交易，则需要将交易费的单独打包
	*/

	// transactions merkle root
	root := core.DeriveSha(txs)

	// generate genesis unit header
	header := modules.Header{
		AssetIDs: []modules.IDType16{assetID},
		Number:   chainIndex,
		TxRoot:   root,
		//		Creationdate: time.Now().Unix(),
	}

	unit := modules.Unit{}
	unit.UnitHeader = &header
	// copy txs
	if len(txs) > 0 {
		unit.Txs = make([]*modules.Transaction, len(txs))
		for i, pTx := range txs {
			tx := modules.Transaction{
				AccountNonce: pTx.AccountNonce,
				TxHash:       pTx.TxHash,
				From:         pTx.From,
				Excutiontime: pTx.Excutiontime,
				Memery:       pTx.Memery,
				CreationDate: pTx.CreationDate,
				TxFee:        pTx.TxFee,
			}
			if len(pTx.TxMessages) > 0 {
				tx.TxMessages = make([]modules.Message, len(pTx.TxMessages))
				for j := 0; j < len(pTx.TxMessages); j++ {
					tx.TxMessages[j] = pTx.TxMessages[j]
				}
			}
			unit.Txs[i] = &tx
		}
	}
	units = append(units, unit)
	return units, nil
}

/**
从leveldb中查询GenesisUnit信息
To get genesis unit info from leveldb
*/
func GetGenesisUnit(index uint64) *modules.Unit {
	// unit key: [HEADER_PREFIX][chain index number]_[chain index]_[unit hash]
	key := fmt.Sprintf("%s%v_", storage.HEADER_PREFIX, index)
	data := storage.GetPrefix([]byte(key))
	if len(data) > 1 {
		log.Info("Get genesis unit error:multiple genesis unit")
		return nil
	} else if len(data) <= 0 {
		return nil
	}
	for k, v := range data {
		sk := string(k[len(storage.HEADER_PREFIX):])
		// get index
		skArr := strings.Split(sk, "_")
		if len(skArr) != 3 {
			log.Error("Get genesis unit index and hash", "error", "split error")
			return nil
		}
		// get unit hash
		uHash := common.Hash{}
		uHash.SetString(skArr[2])
		fmt.Println("Genesis Unit header hash:", []byte(k))
		// get unit header
		fmt.Println("Unit header bytes:", []byte(v))
		var uHeader modules.Header
		if err := rlp.DecodeBytes([]byte(v), &uHeader); err != nil {
			log.Error("Get genesis unit header", "error", err.Error())
			return nil
		}
		// get transaction list
		txs, err := GetUnitTransactions(uHeader.TxRoot)
		if err != nil {
			log.Error("Get genesis unit transactions", "error", err.Error())
			return nil
		}
		// generate unit
		unit := modules.Unit{
			UnitHeader: &uHeader,
			UnitHash:   uHash,
			Txs:        txs,
		}
		unit.UnitSize = unit.Size()
		return &unit
	}
	return nil
}

/**
获取创世单元的高度
To get genesis unit height
*/
func GenesisHeight() modules.ChainIndex {
	unit := GetGenesisUnit(0)
	if unit == nil {
		return modules.ChainIndex{}
	}
	return unit.UnitHeader.Number
}

func GetUnitTransactions(root common.Hash) (modules.Transactions, error) {
	txs := modules.Transactions{}
	// get body data: transaction list
	txHashList, err := storage.GetBody(root)
	if err != nil {
		return nil, err
	}
	// get transaction data
	for _, txHash := range txHashList {
		tx, err := storage.GetTransaction(txHash)
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
func GenGenesisConfigPayload(genesisConf *core.Genesis) (modules.ConfigPayload, error) {
	var confPay modules.ConfigPayload

	confPay.ConfigSet = make(map[string]interface{})

	tt := reflect.TypeOf(*genesisConf)
	vv := reflect.ValueOf(*genesisConf)

	for i := 0; i < tt.NumField(); i++ {
		if strings.Compare(tt.Field(i).Name, "SystemConfig") == 0 {
			t := reflect.TypeOf(genesisConf.SystemConfig)
			v := reflect.ValueOf(genesisConf.SystemConfig)
			for k := 0; k < t.NumField(); k++ {
				confPay.ConfigSet[t.Field(k).Name] = v.Field(k).Interface()
			}
		} else {
			confPay.ConfigSet[tt.Field(i).Name] = vv.Field(i).Interface()
		}
	}

	return confPay, nil
}

/**
保存单元数据，如果单元的结构基本相同
save genesis unit data
*/
func SaveUnit(unit modules.Unit, isGenesis bool) error {
	if unit.UnitSize == 0 || unit.Size() == 0 {
		return fmt.Errorf("Unit is null")
	}
	// check unit signature, should be compare to mediator list
	if err := checkUnitSignature(unit.UnitHeader, isGenesis); err != nil {
		return err
	}

	// check unit size
	if unit.UnitSize != unit.Size() {
		return modules.ErrUnit(-1)
	}
	// check transactions in unit
	totalFee, err := checkTransactions(&unit.Txs, isGenesis)
	if err != nil {
		return err
	}
	// todo check coin base fee
	if totalFee <= 0 {
	}
	// save unit header
	// key is like "[HEADER_PREFIX][chain index number]_[chain index]_[unit hash]"
	if err := storage.SaveHeader(unit.UnitHash, unit.UnitHeader); err != nil {
		return modules.ErrUnit(-3)
	}
	// save unit hash and chain index relation
	// key is like "[UNIT_HASH_NUMBER][unit_hash]"
	if err := storage.SaveHashNumber(unit.UnitHash, unit.UnitHeader.Number); err != nil {
		return fmt.Errorf("Save unit hash and number error")
	}
	// traverse transactions and save them
	txHashSet := []common.Hash{}
	for txIndex, tx := range unit.Txs {
		// traverse messages
		for msgIndex, msg := range tx.TxMessages {
			// handle different messages
			switch msg.App {
			case modules.APP_PAYMENT:
				if ok := savePaymentPayload(tx.TxHash, &msg, uint32(msgIndex), tx.Locktime); ok != true {
					return fmt.Errorf("Save payment payload error.")
				}
			case modules.APP_CONTRACT_TPL:
			case modules.APP_CONTRACT_DEPLOY:
				if ok := saveContractInitPayload(unit.UnitHeader.Number, uint32(txIndex), &msg); ok != true {
					return fmt.Errorf("Save contract init payload error.")
				}
			case modules.APP_CONTRACT_INVOKE:
				if ok := saveContractInvokePayload(unit.UnitHeader.Number, uint32(txIndex), &msg); ok != true {
					return fmt.Errorf("Save contract invode payload error.")
				}
			case modules.APP_CONFIG:
				if ok := saveConfigPayload(tx.TxHash, &msg); ok == false {
					return fmt.Errorf("Save contract invode payload error.")
				}
			case modules.APP_TEXT:
			default:
				return fmt.Errorf("Message type is not supported now: %s", msg.App)
			}
		}
		// save transaction
		if err = storage.SaveTransaction(tx); err != nil {
			return err
		}
	}

	// save unit body, the value only save txs' hash set, and the key is merkle root
	if err = storage.SaveBody(unit.UnitHash, txHashSet); err != nil {
		return err
	}

	// todo send message to transaction pool to delete unit's transactions
	return nil
}

/**
检查message的app与payload是否一致
check messaage 'app' consistent with payload type
*/
func checkMessageType(app string, payload interface{}) bool {
	switch payload.(type) {
	case modules.PaymentPayload:
		if app == modules.APP_PAYMENT {
			return true
		}
	case modules.ContractTplPayload:
		if app == modules.APP_CONTRACT_TPL {
			return true
		}
	case modules.ContractDeployPayload:
		if app == modules.APP_CONTRACT_DEPLOY {
			return true
		}
	case modules.ContractInvokePayload:
		if app == modules.APP_CONTRACT_INVOKE {
			return true
		}
	case modules.ConfigPayload:
		if app == modules.APP_CONFIG {
			return true
		}
	case modules.TextPayload:
		if app == modules.APP_TEXT {
			return true
		}
	default:
		return false
	}
	return false
}

/**
检查unit中所有交易的合法性，返回所有交易的交易费总和
check all transactions in one unit
return all transactions' fee
*/
func checkTransactions(txs *modules.Transactions, isGenesis bool) (uint64, error) {
	fee := uint64(0)
	for _, tx := range *txs {
		for _, msg := range tx.TxMessages {
			// check message type and payload
			if !checkMessageType(msg.App, msg.Payload) {
				return 0, fmt.Errorf("Transaction (%s) message (%s) type is not consistent with payload.", tx.TxHash, msg.PayloadHash)
			}
			// check tx size
			if tx.Size() != tx.Txsize {
				fmt.Printf("Txsize=%v, tx.Size()=%v\n", tx.Txsize, tx.Size())
				return 0, fmt.Errorf("Transaction(%s) Size is incorrect.", tx.TxHash)
			}
			// check every type payload
			switch msg.App {
			case modules.APP_PAYMENT:

			case modules.APP_CONTRACT_TPL:

			case modules.APP_CONTRACT_DEPLOY:

			case modules.APP_CONTRACT_INVOKE:

			case modules.APP_CONFIG:

			case modules.APP_TEXT:

			default:
				return 0, fmt.Errorf("Message type(%s) is not supported now:", msg.App)
			}
		}
		if isGenesis == true {
			// check transaction fee
			txFee := modules.TXFEE
			// i := big.Int{}
			// i.SetUint64(txFee)
			if tx.TxFee.Cmp(txFee) != 0 {
				return 0, fmt.Errorf("Transaction(%s)'s fee is invalid.", tx.TxHash)
			}
		}
	}

	// to check total fee with coinbase tx

	return fee, nil
}

/**
保存PaymentPayload
save PaymentPayload data
*/
func savePaymentPayload(txHash common.Hash, msg *modules.Message, msgIndex uint32, lockTime uint32) bool {
	// if inputs is none then it is just a normal coinbase transaction
	// otherwise, if inputs' length is 1, and it PreviousOutPoint should be none
	// if this is a create token transaction, the Extra field should be AssetInfo struct's [rlp] encode bytes
	// if this is a create token transaction, should be return a assetid
	var pl interface{}
	pl = msg.Payload
	payload, ok := pl.(modules.PaymentPayload)
	if ok == false {
		return false
	}
	if len(payload.Inputs) > 0 {
		if len(payload.Inputs) == 1 && unsafe.Sizeof(payload.Inputs[0].PreviousOutPoint) == 0 {
			// create new token
			var assetInfo modules.AssetInfo
			if err := rlp.DecodeBytes(payload.Inputs[0].Extra, &assetInfo); err != nil {
				return false
			}
			// create asset id
			assetInfo.AssetID.AssertId = asset.NewAsset()
			assetInfo.AssetID.UniqueId = assetInfo.AssetID.AssertId
			data := GetConfig([]byte("ChainID"))
			chainID := common.BytesToInt(data)
			if chainID < 0 {
				return false
			}
			assetInfo.AssetID.ChainId = uint64(chainID)
			// save asset info
			if err := SaveAssetInfo(&assetInfo); err != nil {
				log.Error("Save asset info error")
			}
		}
	}
	// save utxo
	UpdateUtxo(txHash, msg, msgIndex, lockTime)
	return true
}

/**
保存配置交易
save config payload
*/
func saveConfigPayload(txHash common.Hash, msg *modules.Message) bool {
	var pl interface{}
	pl = msg.Payload
	payload, ok := pl.(modules.ConfigPayload)
	if ok == false {
		return false
	}

	if err := SaveConfig(payload.ConfigSet); err != nil {
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
func saveContractInvokePayload(height modules.ChainIndex, txIndex uint32, msg *modules.Message) bool {
	var pl interface{}
	pl = msg.Payload
	payload, ok := pl.(modules.ContractInvokePayload)
	if ok == false {
		return false
	}
	// save contract state
	// key: [CONTRACT_STATE_PREFIX][contract id]_[field name]_[state version]
	for k, v := range payload.WriteSet {
		version := modules.StateVersion{
			Height:  height,
			TxIndex: txIndex,
		}
		key := fmt.Sprintf("%s%s_%s_%s",
			storage.CONTRACT_STATE_PREFIX,
			payload.ContractId,
			k,
			version.String())
		if err := storage.Store(key, v); err != nil {
			log.Error("Save payload key", "error", err.Error())
			continue
		}
	}
	return true
}

/**
保存合约初始化状态
To save contract init state
*/
func saveContractInitPayload(height modules.ChainIndex, txIndex uint32, msg *modules.Message) bool {
	var pl interface{}
	pl = msg.Payload
	_, ok := pl.(modules.ContractDeployPayload)
	if ok == false {
		return false
	}
	/**
	涉及到合约验证和合约ID生成的问题
	*/

	return true
}

/**
验证单元的签名，需要比对见证人列表
*/
func checkUnitSignature(h *modules.Header, isGenesis bool) error {
	if h.Authors == nil || len(h.Authors.Address) <= 0 {
		return fmt.Errorf("No author info")
	}
	emptySigUnit := modules.Unit{}
	// copy unit's header
	emptySigUnit.UnitHeader = modules.CopyHeader(h)
	// signature does not contain authors and witness fields
	emptySigUnit.UnitHeader.Authors = nil
	emptySigUnit.UnitHeader.Witness = []*modules.Authentifier{}
	// recover signature
	sig := make([]byte, 65)
	copy(sig[32-len(h.Authors.R):32], h.Authors.R)
	copy(sig[64-len(h.Authors.S):64], h.Authors.S)
	copy(sig[64:], h.Authors.V)
	// recover pubkey
	hash := crypto.Keccak256Hash(util.RHashBytes(*emptySigUnit.UnitHeader))
	pubKey, err := RSVtoPublicKey(hash[:], h.Authors.R[:], h.Authors.S[:], h.Authors.V[:])
	//  pubKey to pubKey_bytes
	pubKey_bytes := crypto.FromECDSAPub(pubKey)
	if keystore.VerifyUnitWithPK(sig, *emptySigUnit.UnitHeader, pubKey_bytes) == false {
		return fmt.Errorf("Verify unit signature error.")
	}
	// if genesis unit just return
	if isGenesis == false {
		return nil
	}
	// todo group signature verify
	// get mediators
	data := GetConfig([]byte("MediatorCandidates"))
	bNum := GetConfig([]byte("ActiveMediators"))
	num, err := strconv.Atoi(string(bNum))
	if err != nil {
		return fmt.Errorf("Check unit signature error: %s", err)
	}
	if num != len(data) {
		return fmt.Errorf("Check unit signature error: mediators info error, pls update network")
	}
	// decode mediator list data
	var mediators []string
	if err := rlp.DecodeBytes(data, &mediators); err != nil {
		return fmt.Errorf("Check unit signature error: %s", err)
	}

	return nil
}

func RSVtoAddress(tx *modules.Transaction) common.Address {
	sig := make([]byte, 65)
	copy(sig[32-len(tx.From.R):32], tx.From.R)
	copy(sig[64-len(tx.From.S):64], tx.From.S)
	copy(sig[64:], tx.From.V)
	pub, _ := crypto.SigToPub(tx.TxHash[:], sig)
	address := crypto.PubkeyToAddress(*pub)
	return address
}

func RSVtoPublicKey(hash, r, s, v []byte) (*ecdsa.PublicKey, error) {
	sig := make([]byte, 65)
	copy(sig[32-len(r):32], r)
	copy(sig[64-len(s):64], s)
	copy(sig[64:], v)
	return crypto.SigToPub(hash, sig)
}

/**
从levedb中根据ChainIndex获得Unit信息
To get unit information by its ChainIndex
*/
func QueryUnitByChainIndex(index *modules.ChainIndex) *modules.Unit {
	return nil
}
