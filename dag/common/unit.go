package common

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"time"
	"unsafe"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/crypto"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/rlp"
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
	gUnit := modules.Unit{Creationdate: time.Now().UTC()}

	// genesis unit asset id
	gAssetID := asset.NewAsset()

	// genesis unit height
	chainIndex := modules.ChainIndex{AssetID: gAssetID, IsMain: true, Index: 0}

	// transactions merkle root
	root := core.DeriveSha(txs)

	// generate genesis unit header
	header := modules.Header{
		AssetIDs: []modules.IDType16{gAssetID},
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
从leveldb中查询GenesisUnit信息
To get genesis unit info from leveldb
*/
func GetGenesisUnit(index uint64) *modules.Unit {
	// unit key: [HEADER_PREFIX][index]_[chainindex struct]_[unit Bytes]
	key := fmt.Sprintf("%s%v_", storage.HEADER_PREFIX, index)
	data := storage.GetPrefix([]byte(key))
	for k, v := range data {
		var chainIndex modules.ChainIndex
		if err := rlp.DecodeBytes([]byte(k), &chainIndex); err != nil {
			msg := fmt.Sprintf("Chainindex get error: %s", err)
			log.Error(msg)
			continue
		}

		if chainIndex.IsMain == true {
			var unit modules.Unit
			if err := rlp.DecodeBytes([]byte(v), &unit); err != nil {
				msg := fmt.Sprintf("Chainindex get error: %s", err)
				log.Error(msg)
				return nil
			}
			return &unit
		}
	}
	return nil
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
保存单元数据，如果单元的结构基本相同
save genesis unit data
*/
func SaveUnit(unit modules.Unit) error {
	if unit.UnitSize == 0 || unit.Size() == 0 {
		log.Info("Unit is null")
		return nil
	}
	// check unit signature, should be compare to mediator list
	if err := checkUnitSignature(unit.UnitHeader); err != nil {
		return err
	}

	// check unit size
	if unit.UnitSize != unit.Size() {
		return modules.ErrUnit(-1)
	}
	// check transactions in unit
	totalFee, err := checkTransactions(&unit.Txs)
	if err != nil {
		return err
	}
	// todo check coin base fee
	if totalFee <= 0 {
	}
	// save unit header, key is like "[HEADER_PREFIX][chain_index]_[unit hash]"
	if err := storage.SaveHeader(unit.UnitHash, unit.UnitHeader); err != nil {
		return modules.ErrUnit(-3)
	}

	// traverse transactions and save them
	txHashSet := []common.Hash{}
	for _, tx := range unit.Txs {
		// traverse messages
		for _, msg := range tx.TxMessages {
			// handle different messages
			switch msg.App {
			case modules.APP_PAYMENT:
				if ok := savePaymentPayload(tx.TxHash, &msg); ok != true {
					log.Error("Save payment payload error.")
					return modules.ErrUnit(-5)
				}
			case modules.APP_CONTRACT_TPL:
			case modules.APP_CONTRACT_DEPLOY:
			case modules.APP_CONTRACT_INVOKE:
			case modules.APP_CONFIG:
				if ok := saveConfigPayload(tx.TxHash, &msg); ok == false {
					return modules.ErrUnit(-6)
				}
			case modules.APP_TEXT:
			default:
				log.Error("Message type is not supported now:", msg.App)
			}
		}
		// save transaction
		if err = storage.SaveTransaction(tx); err != nil {
			return err
		}
	}

	// save unit body, the value only save txs' hash set, and the key is merkle root
	if err = storage.SaveBody(unit.UnitHeader.Root, txHashSet); err != nil {
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
func checkTransactions(txs *modules.Transactions) (uint64, error) {
	fee := uint64(0)
	for _, tx := range *txs {
		for _, msg := range tx.TxMessages {
			// check message type and payload
			if !checkMessageType(msg.App, msg.Payload) {
				return 0, fmt.Errorf("Transaction (%s) message (%s) type is not consistent with payload.", tx.TxHash, msg.PayloadHash)
			}
			// check tx size
			if tx.Size() != tx.Txsize {
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
		// check transaction fee
		txFee := modules.TXFEE
		// i := big.Int{}
		// i.SetUint64(txFee)
		if tx.TxFee.Cmp(txFee) != 0 {
			return 0, fmt.Errorf("Transaction(%s)'s fee is invalid.", tx.TxHash)
		}
	}

	// to check total fee with coinbase tx

	return fee, nil
}

/**
保存PaymentPayload
save PaymentPayload data
*/
func savePaymentPayload(txHash common.Hash, msg *modules.Message) bool {
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
	UpdateUtxo(txHash, msg)
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
验证单元的签名，需要比对见证人列表
*/
func checkUnitSignature(h *modules.Header) error {
	if h.Authors == nil || len(h.Authors.Address) <= 0 {
		return fmt.Errorf("No author info")
	}
	emptySigUnit := modules.Unit{}
	emptySigUnit.UnitHeader.Authors = nil
	emptySigUnit.UnitHeader.Witness = []modules.Author{}

	sig := make([]byte, 65)
	copy(sig[32-len(h.Authors.R):32], h.Authors.R)
	copy(sig[64-len(h.Authors.S):64], h.Authors.S)
	copy(sig[64:len(sig)], h.Authors.V)

	hash := h.Hash()
	pubKey, _ := RSVtoPublicKey(hash[:], h.Authors.R[:], h.Authors.S[:], h.Authors.V[:])
	//
	//  pubKey to pubKey_bytes
	pubKey_bytes := crypto.FromECDSAPub(pubKey)
	if keystore.VerifyUnitWithPK(sig, emptySigUnit, pubKey_bytes) == false {
		return fmt.Errorf("Verify unit signature error.")
	}

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

	// todo group signature verify

	return nil
}

func RSVtoAddress(tx *modules.Transaction) common.Address {
	sig := make([]byte, 65)
	copy(sig[32-len(tx.From.R):32], tx.From.R)
	copy(sig[64-len(tx.From.S):64], tx.From.S)
	copy(sig[64:len(sig)], tx.From.V)
	pub, _ := crypto.SigToPub(tx.TxHash[:], sig)
	address := crypto.PubkeyToAddress(*pub)
	return address
}

func RSVtoPublicKey(hash, r, s, v []byte) (*ecdsa.PublicKey, error) {
	sig := make([]byte, 65)
	copy(sig[32-len(r):32], r)
	copy(sig[64-len(s):64], s)
	copy(sig[64:len(sig)], v)
	return crypto.SigToPub(hash, sig)
}

/**
从levedb中根据ChainIndex获得Unit信息
To get unit information by its ChainIndex
*/
func QueryUnitByChainIndex(index *modules.ChainIndex) *modules.Unit {
	return nil
}
