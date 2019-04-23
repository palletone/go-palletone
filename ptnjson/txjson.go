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
	"time"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/dag/modules"
)

type TxJson struct {
	TxHash         string              `json:"tx_hash"`
	TxSize         float64             `json:"tx_size"`
	Payment        *PaymentJson        `json:"payment"`
	Vote           *VoteJson           `json:"vote"`
	Data           *DataJson           `json:"data"`
	ContractTpl    *TplJson            `json:"contract_tpl"`
	Deploy         *DeployJson         `json:"contract_deploy"`
	Invoke         *InvokeJson         `json:"contract_invoke"`
	Stop           *StopJson           `json:"contract_stop"`
	Signature      *SignatureJson      `json:"signature"`
	InstallRequest *InstallRequestJson `json:"install_request"`
	DeployRequest  *DeployRequestJson  `json:"deploy_request"`
	InvokeRequest  *InvokeRequestJson  `json:"invoke_request"`
	StopRequest    *StopRequestJson    `json:"stop_request"`
}
type TplJson struct {
	TemplateId string `json:"template_id"`
	Name       string `json:"name"`
	Path       string `json:"path"`
	Version    string `json:"version"`
	Memory     uint16 `json:"memory"`
	Bytecode   []byte `json:"bytecode"` // contract bytecode
}
type DeployJson struct {
	TemplateId string   `json:"template_id"`
	ContractId string   `json:"contract_id"`
	Name       string   `json:"name"`
	Args       [][]byte `json:"args"` // contract arguments list
	Jury       []string `json:"jury"`
	EleList    string   `json:"election_list"`
	ReadSet    string   `json:"read_set"`
	WriteSet   string   `json:"write_set"`
}
type InvokeJson struct {
	ContractId   string   `json:"contract_id"` // contract id
	FunctionName string   `json:"function_name"`
	Args         [][]byte `json:"args"`      // contract arguments list
	ReadSet      string   `json:"read_set"`  // the set data of read, and value could be any type
	WriteSet     string   `json:"write_set"` // the set data of write, and value could be any type
	Payload      []byte   `json:"payload"`   // the contract execution result
}
type StopJson struct {
	ContractId string   `json:"contract_id"`
	Jury       []string `json:"jury"`
	ReadSet    string   `json:"read_set"`
	WriteSet   string   `json:"write_set"`
}
type SignatureJson struct {
	Signatures []string `json:"signature_set"` // the array of signature
}

type VoteJson struct {
	Content string `json:"vote_content"`
}

type InvokeRequestJson struct {
	ContractAddr string   `json:"contract_addr"`
	FunctionName string   `json:"function_name"`
	Args         []string `json"arg_set"`
}

type InstallRequestJson struct {
	TplName string `json:"tpl_name"`
	Path    string `json:"path"`
	Version string `json:"version"`
}

type DeployRequestJson struct {
	TplId   string        `json:"tpl_id"`
	TxId    string        `json:"tx_id"`
	Args    []string      `json:"arg_set"`
	Timeout time.Duration `json:"timeout"`
}

type StopRequestJson struct {
	ContractId  string `json:"contract_id"`
	Txid        string `json:"tx_id"`
	DeleteImage bool   `json:"delete_image"`
}
type DataJson struct {
	MainData  string `json:"main_data"`
	ExtraData string `json:"extra_data"`
}

func ConvertTx2Json(tx *modules.Transaction, utxoQuery modules.QueryUtxoFunc) TxJson {
	txjson := TxJson{TxHash: tx.Hash().String(), TxSize: float64(tx.Size())}
	for _, m := range tx.TxMessages {
		if m.App == modules.APP_PAYMENT {
			pay := m.Payload.(*modules.PaymentPayload)
			if utxoQuery == nil {
				payJson := ConvertPayment2Json(pay)
				txjson.Payment = payJson
			} else {
				payJson := ConvertPayment2JsonIncludeFromAddr(pay, utxoQuery)
				txjson.Payment = payJson
			}
		} else if m.App == modules.APP_DATA {
			data := m.Payload.(*modules.DataPayload)
			txjson.Data = &DataJson{MainData: string(data.MainData), ExtraData: string(data.ExtraData)}
		} else if m.App == modules.APP_CONTRACT_TPL_REQUEST {
			req := m.Payload.(*modules.ContractInstallRequestPayload)
			txjson.InstallRequest = convertInstallRequest2Json(req)
		} else if m.App == modules.APP_CONTRACT_DEPLOY_REQUEST {
			req := m.Payload.(*modules.ContractDeployRequestPayload)
			txjson.DeployRequest = convertDeployRequest2Json(req)
		} else if m.App == modules.APP_CONTRACT_INVOKE_REQUEST {
			req := m.Payload.(*modules.ContractInvokeRequestPayload)
			txjson.InvokeRequest = convertInvokeRequest2Json(req)
		} else if m.App == modules.APP_CONTRACT_STOP_REQUEST {
			req := m.Payload.(*modules.ContractStopRequestPayload)
			txjson.StopRequest = convertStopRequest2Json(req)
		} else if m.App == modules.APP_CONTRACT_TPL {
			tpl := m.Payload.(*modules.ContractTplPayload)
			txjson.ContractTpl = convertTpl2Json(tpl)
		} else if m.App == modules.APP_CONTRACT_DEPLOY {
			deploy := m.Payload.(*modules.ContractDeployPayload)
			txjson.Deploy = convertDeploy2Json(deploy)
		} else if m.App == modules.APP_CONTRACT_INVOKE {
			invoke := m.Payload.(*modules.ContractInvokePayload)
			txjson.Invoke = convertInvoke2Json(invoke)
		} else if m.App == modules.APP_CONTRACT_STOP {
			stop := m.Payload.(*modules.ContractStopPayload)
			txjson.Stop = convertStop2Json(stop)
		} else if m.App == modules.APP_SIGNATURE {
			sig := m.Payload.(*modules.SignaturePayload)
			txjson.Signature = convertSig2Json(sig)
		}
	}
	return txjson
}
func ConvertJson2Tx(json *TxJson) *modules.Transaction {
	tx := &modules.Transaction{}
	pay := ConvertJson2Payment(json.Payment)
	tx.AddMessage(modules.NewMessage(modules.APP_PAYMENT, pay))
	return tx
}
func convertTpl2Json(tpl *modules.ContractTplPayload) *TplJson {
	tpljson := new(TplJson)
	hash := common.BytesToHash(tpl.TemplateId[:])
	tpljson.TemplateId = hash.String()
	tpljson.Name = tpl.Name
	tpljson.Path = tpl.Path
	tpljson.Version = tpl.Version
	tpljson.Bytecode = tpl.Bytecode[:]
	tpljson.Memory = tpl.Memory
	return tpljson
}
func convertDeploy2Json(deploy *modules.ContractDeployPayload) *DeployJson {
	djson := new(DeployJson)
	djson.Name = deploy.Name
	hash := common.BytesToHash(deploy.TemplateId[:])
	djson.TemplateId = hash.String()
	hash = common.Hash{}
	hash.SetBytes(deploy.ContractId[:])
	djson.ContractId = hash.String()
	djson.Args = deploy.Args
	//for _, addr := range deploy.Jury {
	//	djson.Jury = append(djson.Jury, addr.String())
	//}
	ele, _ := json.Marshal(deploy.EleList)
	djson.EleList = string(ele)
	rset, _ := json.Marshal(deploy.ReadSet)
	djson.ReadSet = string(rset)
	wset, _ := json.Marshal(deploy.WriteSet)
	djson.WriteSet = string(wset)
	return djson
}
func convertInvoke2Json(invoke *modules.ContractInvokePayload) *InvokeJson {
	injson := new(InvokeJson)
	hash := common.BytesToHash(invoke.ContractId[:])
	injson.ContractId = hash.String()
	injson.FunctionName = invoke.FunctionName

	injson.Args = invoke.Args
	rset, _ := json.Marshal(invoke.ReadSet)
	injson.ReadSet = string(rset)
	wset, _ := json.Marshal(invoke.WriteSet)
	injson.WriteSet = string(wset)
	injson.Payload = invoke.Payload
	return injson
}
func convertStop2Json(stop *modules.ContractStopPayload) *StopJson {
	sjson := new(StopJson)
	hash := common.BytesToHash(stop.ContractId[:])
	sjson.ContractId = hash.String()
	rset, _ := json.Marshal(stop.ReadSet)
	sjson.ReadSet = string(rset)
	wset, _ := json.Marshal(stop.WriteSet)
	sjson.WriteSet = string(wset)
	return sjson
}
func convertSig2Json(sig *modules.SignaturePayload) *SignatureJson {
	sigjson := new(SignatureJson)
	for _, sig := range sig.Signatures {
		set, _ := json.Marshal(sig)
		sigjson.Signatures = append(sigjson.Signatures, string(set))
	}
	return sigjson
}

func convertInvokeRequest2Json(req *modules.ContractInvokeRequestPayload) *InvokeRequestJson {
	addr := common.NewAddress(req.ContractId[:], common.ContractHash)
	reqJson := &InvokeRequestJson{}
	reqJson.ContractAddr = addr.String()
	reqJson.FunctionName = req.FunctionName
	reqJson.Args = []string{}
	for _, arg := range req.Args {
		reqJson.Args = append(reqJson.Args, string(arg))
	}
	return reqJson
}

func convertInstallRequest2Json(req *modules.ContractInstallRequestPayload) *InstallRequestJson {
	reqJson := &InstallRequestJson{}
	reqJson.TplName = req.TplName
	reqJson.Path = req.Path
	reqJson.Version = req.Version
	return reqJson
}

func convertDeployRequest2Json(req *modules.ContractDeployRequestPayload) *DeployRequestJson {
	reqJson := &DeployRequestJson{}
	hash := common.BytesToHash(req.TplId[:])
	reqJson.TplId = hash.String()
	reqJson.Args = []string{}
	for _, arg := range req.Args {
		reqJson.Args = append(reqJson.Args, string(arg))
	}
	return reqJson
}

func convertStopRequest2Json(req *modules.ContractStopRequestPayload) *StopRequestJson {
	reqJson := &StopRequestJson{}
	addr := common.NewAddress(req.ContractId[:], common.ContractHash)
	reqJson.ContractId = addr.String()
	reqJson.Txid = req.Txid

	return reqJson
}
