/*
 *
 *    This file is part of go-palletone.
 *    go-palletone is free software: you can redistribute it and/or modify
 *    it under the terms of the GNU General Public License as published by
 *    the Free Software Foundation, either version 3 of the License, or
 *    (at your option) any later version.
 *    go-palletone is distributed in the hope that it will be useful,
 *    but WITHOUT ANY WARRANTY; without even the implied warranty of
 *    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *    GNU General Public License for more details.
 *    You should have received a copy of the GNU General Public License
 *    along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
 * /
 *
 *  * @author PalletOne core developer <dev@pallet.one>
 *  * @date 2018
 *
 */

package ptnjson

import (
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/contracts/syscontract"
	"time"

	"encoding/hex"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/dag/modules"
)

type TxJson struct {
	TxHash             string              `json:"tx_hash"`
	RequestHash        string              `json:"request_hash"`
	TxSize             float64             `json:"tx_size"`
	Payment            []*PaymentJson      `json:"payment"`
	Fee                uint64              `json:"fee"`
	AccountStateUpdate *AccountStateJson   `json:"account_state_update"`
	Data               []*DataJson         `json:"data"`
	ContractTpl        *TplJson            `json:"contract_tpl"`
	Deploy             *DeployJson         `json:"contract_deploy"`
	Invoke             *InvokeJson         `json:"contract_invoke"`
	Stop               *StopJson           `json:"contract_stop"`
	Signature          *SignatureJson      `json:"signature"`
	InstallRequest     *InstallRequestJson `json:"install_request"`
	DeployRequest      *DeployRequestJson  `json:"deploy_request"`
	InvokeRequest      *InvokeRequestJson  `json:"invoke_request"`
	StopRequest        *StopRequestJson    `json:"stop_request"`
}
type TxWithUnitInfoJson struct {
	*TxJson
	UnitHash   string    `json:"unit_hash"`
	UnitHeight uint64    `json:"unit_height"`
	Timestamp  time.Time `json:"timestamp"`
	TxIndex    uint64    `json:"tx_index"`
}
type TplJson struct {
	Number       int    `json:"row_number"`
	TemplateId   string `json:"template_id"`
	Bytecode     []byte `json:"bytecode"`      // contract bytecode
	BytecodeSize int    `json:"bytecode_size"` // contract bytecode
	ErrorCode    uint32 `json:"error_code"`
	ErrorMessage string `json:"error_message"`
}
type DeployJson struct {
	Number       int    `json:"row_number"`
	ContractId   string `json:"contract_id"`
	Name         string `json:"name"`
	EleNode      string `json:"election_node"`
	ReadSet      string `json:"read_set"`
	WriteSet     string `json:"write_set"`
	DuringTime   uint64 `json:"during_time"`
	ErrorCode    uint32 `json:"error_code"`
	ErrorMessage string `json:"error_message"`
}
type InvokeJson struct {
	Number       int    `json:"row_number"`
	ContractId   string `json:"contract_id"` // contract id
	ReadSet      string `json:"read_set"`    // the set data of read, and value could be any type
	WriteSet     string `json:"write_set"`   // the set data of write, and value could be any type
	Payload      string `json:"payload"`     // the contract execution result
	ErrorCode    uint32 `json:"error_code"`
	ErrorMessage string `json:"error_message"`
}
type StopJson struct {
	Number     int    `json:"row_number"`
	ContractId string `json:"contract_id"`
	//Jury         []string `json:"jury"`
	ReadSet      string `json:"read_set"`
	WriteSet     string `json:"write_set"`
	ErrorCode    uint32 `json:"error_code"`
	ErrorMessage string `json:"error_message"`
}
type SignatureJson struct {
	Number     int      `json:"row_number"`
	Signatures []string `json:"signature_set"` // the array of signature
}

type InvokeRequestJson struct {
	Number       int           `json:"row_number"`
	ContractAddr string        `json:"contract_addr"`
	Args         []string      `json:"arg_set"`
	Timeout      time.Duration `json:"timeout"`
}

type InstallRequestJson struct {
	Number         int      `json:"row_number"`
	TplName        string   `json:"tpl_name"`
	TplDescription string   `json:"tpl_description"`
	Path           string   `json:"path"`
	Version        string   `json:"version"`
	Abi            string   `json:"abi"`
	Language       string   `json:"language"`
	AddrHash       []string `json:"addr_hash"`
}

type DeployRequestJson struct {
	Number  int           `json:"row_number"`
	TplId   string        `json:"tpl_id"`
	Args    []string      `json:"arg_set"`
	Timeout time.Duration `json:"timeout"`
	ExtData string        `json:"extend_data"`
}

type StopRequestJson struct {
	Number      int    `json:"row_number"`
	ContractId  string `json:"contract_id"`
	DeleteImage bool   `json:"delete_image"`
}
type DataJson struct {
	Number    int    `json:"row_number"`
	MainData  string `json:"main_data"`
	ExtraData string `json:"extra_data"`
	Reference string `json:"reference"`
}
type AccountStateJson struct {
	Number   int    `json:"row_number"`
	WriteSet string `json:"write_set"`
}

func ConvertTxWithUnitInfo2FullJson(tx *modules.TransactionWithUnitInfo,
	utxoQuery modules.QueryUtxoFunc) *TxWithUnitInfoJson {
	txjson := &TxWithUnitInfoJson{
		UnitHash:   tx.UnitHash.String(),
		UnitHeight: tx.UnitIndex,
		Timestamp:  time.Unix(int64(tx.Timestamp), 0),
		TxIndex:    tx.TxIndex,
	}
	txjson.TxJson = ConvertTx2FullJson(tx.Transaction, utxoQuery)

	return txjson
}
func ConvertTx2FullJson(tx *modules.Transaction,
	utxoQuery modules.QueryUtxoFunc) *TxJson {
	txjson := &TxJson{}
	txjson.Payment = []*PaymentJson{}
	txjson.Data = []*DataJson{}
	txjson.TxHash = tx.Hash().String()
	txjson.RequestHash = tx.RequestHash().String()
	txjson.TxSize = float64(tx.Size())
	for i, m := range tx.TxMessages {
		if m.App == modules.APP_PAYMENT {
			pay := m.Payload.(*modules.PaymentPayload)
			if utxoQuery == nil {
				payJson := ConvertPayment2Json(pay)
				payJson.Number = i
				txjson.Payment = append(txjson.Payment, payJson)
			} else {
				payJson := ConvertPayment2JsonIncludeFromAddr(pay, utxoQuery)
				payJson.Number = i
				txjson.Payment = append(txjson.Payment, payJson)
			}
		} else if m.App == modules.APP_DATA {
			data := m.Payload.(*modules.DataPayload)
			dataJson := &DataJson{
				MainData:  string(data.MainData),
				ExtraData: string(data.ExtraData),
				Reference: string(data.Reference),
			}
			dataJson.Number = i
			txjson.Data = append(txjson.Data, dataJson)
		} else if m.App == modules.APP_CONTRACT_TPL_REQUEST {
			req := m.Payload.(*modules.ContractInstallRequestPayload)
			txjson.InstallRequest = convertInstallRequest2Json(req)
			txjson.InstallRequest.Number = i
		} else if m.App == modules.APP_CONTRACT_DEPLOY_REQUEST {
			req := m.Payload.(*modules.ContractDeployRequestPayload)
			txjson.DeployRequest = convertDeployRequest2Json(req)
			txjson.DeployRequest.Number = i
		} else if m.App == modules.APP_CONTRACT_INVOKE_REQUEST {
			req := m.Payload.(*modules.ContractInvokeRequestPayload)
			txjson.InvokeRequest = convertInvokeRequest2Json(req)
			txjson.InvokeRequest.Number = i
		} else if m.App == modules.APP_CONTRACT_STOP_REQUEST {
			req := m.Payload.(*modules.ContractStopRequestPayload)
			txjson.StopRequest = convertStopRequest2Json(req)
			txjson.StopRequest.Number = i
		} else if m.App == modules.APP_CONTRACT_TPL {
			tpl := m.Payload.(*modules.ContractTplPayload)
			txjson.ContractTpl = convertTpl2Json(tpl)
			txjson.ContractTpl.Number = i
		} else if m.App == modules.APP_CONTRACT_DEPLOY {
			deploy := m.Payload.(*modules.ContractDeployPayload)
			txjson.Deploy = convertDeploy2Json(deploy)
			txjson.Deploy.Number = i
		} else if m.App == modules.APP_CONTRACT_INVOKE {
			invoke := m.Payload.(*modules.ContractInvokePayload)
			txjson.Invoke = convertInvoke2Json(invoke)
			txjson.Invoke.Number = i
		} else if m.App == modules.APP_CONTRACT_STOP {
			stop := m.Payload.(*modules.ContractStopPayload)
			txjson.Stop = convertStop2Json(stop)
			txjson.Stop.Number = i
		} else if m.App == modules.APP_SIGNATURE {
			sig := m.Payload.(*modules.SignaturePayload)
			txjson.Signature = convertSig2Json(sig)
			txjson.Signature.Number = i
		} else if m.App == modules.APP_ACCOUNT_UPDATE {
			acc := m.Payload.(*modules.AccountStateUpdatePayload)
			txjson.AccountStateUpdate = convertAccountState2Json(acc)
			txjson.AccountStateUpdate.Number = i
		}
	}
	if utxoQuery != nil {
		fee, err := tx.GetTxFee(utxoQuery)
		if err == nil {
			txjson.Fee = fee.Amount
		}
	}
	return txjson
}
func ConvertJson2Tx(json *TxJson) *modules.Transaction {
	tx := &modules.Transaction{}
	for _, payjson := range json.Payment {
		pay := ConvertJson2Payment(payjson)
		tx.AddMessage(modules.NewMessage(modules.APP_PAYMENT, pay))
	}
	return tx
}
func convertTpl2Json(tpl *modules.ContractTplPayload) *TplJson {
	tpljson := new(TplJson)
	tpljson.TemplateId = hex.EncodeToString(tpl.TemplateId)
	tpljson.Bytecode = tpl.ByteCode[:]
	tpljson.BytecodeSize = len(tpl.ByteCode[:])
	tpljson.ErrorCode = tpl.ErrMsg.Code
	tpljson.ErrorMessage = tpl.ErrMsg.Message
	return tpljson
}
func convertDeploy2Json(deploy *modules.ContractDeployPayload) *DeployJson {
	djson := new(DeployJson)
	djson.Name = deploy.Name
	djson.ContractId = hex.EncodeToString(deploy.ContractId)
	ele, _ := json.Marshal(deploy.EleNode)
	djson.EleNode = string(ele)
	rset, _ := json.Marshal(deploy.ReadSet)
	djson.ReadSet = string(rset)
	wset, _ := json.Marshal(deploy.WriteSet)
	djson.WriteSet = string(wset)
	djson.DuringTime = deploy.DuringTime
	djson.ErrorCode = deploy.ErrMsg.Code
	djson.ErrorMessage = deploy.ErrMsg.Message
	return djson
}

type CoinbaseWriteSet struct {
	IsDelete bool   `json:"is_delete"`
	Key      string `json:"key"`
	Value    string `json:"value"`
}

func convertInvoke2Json(invoke *modules.ContractInvokePayload) *InvokeJson {
	injson := new(InvokeJson)
	injson.ContractId = contractId2AddrString(invoke.ContractId)
	rset, _ := json.Marshal(invoke.ReadSet)
	injson.ReadSet = string(rset)
	//Speical for coinbase
	if injson.ContractId == syscontract.CoinbaseContractAddress.String() {
		wsjs := []CoinbaseWriteSet{}
		for _, w := range invoke.WriteSet {
			wsj := CoinbaseWriteSet{IsDelete: w.IsDelete, Key: w.Key}
			aa := []modules.AmountAsset{}
			rlp.DecodeBytes(w.Value, &aa)
			value := ""
			for _, a := range aa {
				value += a.String() + ";"
			}
			wsj.Value = value
			wsjs = append(wsjs, wsj)
		}
		wset, _ := json.Marshal(wsjs)
		injson.WriteSet = string(wset)

	} else {
		wset, _ := json.Marshal(invoke.WriteSet)
		injson.WriteSet = string(wset)
	}
	injson.Payload = string(invoke.Payload)
	injson.ErrorCode = invoke.ErrMsg.Code
	injson.ErrorMessage = invoke.ErrMsg.Message

	return injson
}
func contractId2AddrString(contractId []byte) string {
	addr := common.NewAddress(contractId, common.ContractHash)
	return addr.String()
}
func convertStop2Json(stop *modules.ContractStopPayload) *StopJson {
	sjson := new(StopJson)

	sjson.ContractId = contractId2AddrString(stop.ContractId)
	rset, _ := json.Marshal(stop.ReadSet)
	sjson.ReadSet = string(rset)
	wset, _ := json.Marshal(stop.WriteSet)
	sjson.WriteSet = string(wset)
	sjson.ErrorCode = stop.ErrMsg.Code
	sjson.ErrorMessage = stop.ErrMsg.Message
	return sjson
}
func convertSig2Json(sig *modules.SignaturePayload) *SignatureJson {
	sigjson := new(SignatureJson)
	for _, sig := range sig.Signatures {
		set := fmt.Sprintf("pubkey:%x,signature:%x", sig.PubKey, sig.Signature)
		sigjson.Signatures = append(sigjson.Signatures, set)
	}
	return sigjson
}

func convertInvokeRequest2Json(req *modules.ContractInvokeRequestPayload) *InvokeRequestJson {
	reqJson := &InvokeRequestJson{}
	reqJson.ContractAddr = contractId2AddrString(req.ContractId)
	reqJson.Args = []string{}
	for _, arg := range req.Args {
		reqJson.Args = append(reqJson.Args, string(arg))
	}
	reqJson.Timeout = time.Duration(req.Timeout) * time.Second
	return reqJson
}

func convertInstallRequest2Json(req *modules.ContractInstallRequestPayload) *InstallRequestJson {
	reqJson := &InstallRequestJson{}
	reqJson.TplName = req.TplName
	reqJson.Path = req.Path
	reqJson.Version = req.Version
	reqJson.Abi = req.Abi
	reqJson.Language = req.Language
	reqJson.TplDescription = req.TplDescription

	reqJson.AddrHash = []string{}
	for _, aHash := range req.AddrHash {
		reqJson.AddrHash = append(reqJson.AddrHash, hex.EncodeToString(aHash[:]))
	}

	return reqJson
}

func convertDeployRequest2Json(req *modules.ContractDeployRequestPayload) *DeployRequestJson {
	reqJson := &DeployRequestJson{}

	reqJson.TplId = hex.EncodeToString(req.TemplateId)
	reqJson.Args = []string{}
	for _, arg := range req.Args {
		reqJson.Args = append(reqJson.Args, string(arg))
	}
	reqJson.Timeout = time.Duration(req.Timeout) * time.Second
	reqJson.ExtData = hex.EncodeToString(req.ExtData)
	return reqJson
}

func convertStopRequest2Json(req *modules.ContractStopRequestPayload) *StopRequestJson {
	reqJson := &StopRequestJson{}

	reqJson.ContractId = contractId2AddrString(req.ContractId)
	//reqJson.Txid = req.Txid

	return reqJson
}
func convertAccountState2Json(accountState *modules.AccountStateUpdatePayload) *AccountStateJson {
	jsonAcc := &AccountStateJson{}
	writeSet, _ := json.Marshal(accountState.WriteSet)
	jsonAcc.WriteSet = string(writeSet)
	return jsonAcc
}
