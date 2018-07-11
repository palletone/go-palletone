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

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/dag/asset"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/storage"
	"time"
	"reflect"
	"github.com/palletone/go-palletone/common/log"
	"unsafe"
	"github.com/palletone/go-palletone/common/rlp"
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
func GetUnit(hash *common.Hash, index modules.ChainIndex) *modules.Unit {
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
func NewGenesisUnit(txs modules.Transactions) (*modules.Unit, error) {
	gUnit := modules.Unit{Gasprice:0, Gasused:0, Creationdate:time.Now().UTC()}

	// genesis unit asset id
	gAssetID := asset.NewAsset()

	// genesis unit height
	chainIndex := modules.ChainIndex{AssetID: gAssetID, IsMain: true, Index: 0}

	// transactions merkle root
	root := core.DeriveSha(txs)

	// generate genesis unit header
	header := modules.Header{
		AssetIDs: []modules.IDType36{gAssetID},
		GasLimit: 0,
		GasUsed:  0,
		Number:   chainIndex,
		Root:     root,
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
	return &gUnit, nil
}

/**
为创世单元生成ConfigPayload
To generate config payload for genesis unit
 */
func GenGenesisConfigPayload(genesisConf *core.Genesis) (modules.ConfigPayload, error) {
	var confPay modules.ConfigPayload

	confPay.ConfigSet = make(map[string]interface{})
	confPay.ConfigSet["version"] = genesisConf.Version
	confPay.ConfigSet["initialActiveMediators"] = genesisConf.InitialActiveMediators
	confPay.ConfigSet["InitialMediatorCandidates"] = genesisConf.InitialMediatorCandidates
	confPay.ConfigSet["ChainID"] = genesisConf.ChainID

	t := reflect.TypeOf(genesisConf.SystemConfig)
	v := reflect.ValueOf(genesisConf.SystemConfig)
	for k := 0; k < t.NumField(); k++ {
		confPay.ConfigSet[t.Field(k).Name] = v.Field(k).Interface()
	}

	return confPay, nil
}

/**
保存创世单元数据
save genesis unit data
 */
func SaveUnit(unit modules.Unit)  error{
	// check unit signature

	// check unit size
	if unit.UnitSize != unit.Size() {
		return modules.ErrUnit(-1)
	}
	// save unit header, key is like ""
	if err := storage.SaveHeader(unit.UnitHash, unit.UnitHeader); err!=nil {
		return modules.ErrUnit(-3)
	}

	// traverse transactions
	for _, tx := range unit.Txs {
		// check tx size
		if tx.Size() != tx.TxSize {
			return modules.ErrUnit(-4)
		}
		// traverse messages
		for _, msg := range tx.TxMessages {
			if !CheckMessageType(msg.App, msg.Payload) {
				log.Error("Message payload is not consistent with app:", msg.App)
				continue
			}
			// handle different messages
			switch msg.App {
			case modules.APP_PAYMENT:
				payload :=msg.Payload.(modules.PaymentPayload)
				if ok :=savePaymentPayload(tx, &payload); ok!=true {
					log.Error("Save payment payload error.")
					return modules.ErrUnit(-5)
				}

			case modules.APP_CONTRACT_TPL:
			case modules.APP_CONTRACT_DEPLOY:
			case modules.APP_CONTRACT_INVOKE:
			case modules.APP_CONFIG:
			case modules.APP_TEXT:
			default:
				log.Error("Message type is not supported now:", msg.App)
			}
		}
	}
	// save unit body, the value only save txs' hash set, and the key is merkle root

	return nil
}

/**
检查message的app与payload是否一致
check messaage 'app' consistent with payload type
 */
func CheckMessageType(app string, payload interface{})  bool{
	switch payload.(type) {
	case modules.PaymentPayload:
		if app == modules.APP_PAYMENT { return true}
	case modules.ContractTplPayload:
		if app == modules.APP_CONTRACT_TPL { return true}
	case modules.ContractDeployPayload:
		if app == modules.APP_CONTRACT_DEPLOY { return true }
	case modules.ContractInvokePayload:
		if app == modules.APP_CONTRACT_INVOKE { return true}
	case modules.ConfigPayload:
		if app == modules.APP_CONFIG { return true }
	case modules.TextPayload:
		if app == modules.APP_TEXT { return true}
	default:
		return false
	}
	return false
}

/**
保存PaymentPayload
save PaymentPayload data
 */
func savePaymentPayload(tx *modules.Transaction, payload *modules.PaymentPayload) bool {
	// if inputs is none then it is just a normal coinbase transaction
	// otherwise, if inputs' length is 1, and it PreviousOutPoint should be none
	// if this is a create token transaction, the Extra field should be AssetInfo struct's [rlp] encode bytes
	// if this is a create token transaction, should be return a assetid
	if len(payload.Inputs) > 0{
		if len(payload.Inputs)==1 && unsafe.Sizeof(payload.Inputs[0].PreviousOutPoint)==0 {
			// create new token
			var assetInfo modules.AssetInfo
			if err:=rlp.DecodeBytes(payload.Inputs[0].Extra, &assetInfo); err!=nil{ return false }
			// create asset id
			assetInfo.AssetID.AssertId = asset.NewAsset()
			assetInfo.AssetID.UniqueId = assetInfo.AssetID.AssertId
			data := GetConfig([]byte("ChainID"))
			chainID := common.BytesToInt(data)
			if chainID < 0 { return false }
			assetInfo.AssetID.ChainId = uint64(chainID)
			// save asset info
			if err:=SaveAssetInfo(&assetInfo); err!=nil {
				log.Error("Save asset info error")
			}
		} else {
			// use utxo transaction
			// check total balance
			//utxos, total := GetUxtoSetByInputs(payload.Inputs)
			//walletAmount := WalletBalance()
			// check acount balance
			// check double spent
		}
	} else {
		// coinbase

	}
	// save utxo
	UpdateUtxo(tx.From.Address, tx)
	return true
}