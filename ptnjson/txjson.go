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
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/vote"
	"time"
)

type TxJson struct {
	TxHash         string              `json:"tx_hash"`
	TxSize         float64             `json:"tx_size"`
	Payment        *PaymentJson        `json:"payment"`
	Vote           *VoteJson           `json:"vote"`
	Data           *DataJson           `json:"data"`
	InstallRequest *InstallRequestJson `json:"install_request"`
	DeployRequest  *DeployRequestJson  `json:"deploy_request"`
	InvokeRequest  *InvokeRequestJson  `json:"invoke_request"`
	StopRequest    *StopRequestJson    `json:"stop_request"`
}
type VoteJson struct {
	Content string `json:"vote_content"`
}

type InvokeRequestJson struct {
	ContractAddr string
	FunctionName string
	Args         []string
}

type InstallRequestJson struct {
	TplName string
	Path    string
	Version string
}

type DeployRequestJson struct {
	TplId   string
	TxId    string
	Args    []string
	Timeout time.Duration
}

type StopRequestJson struct {
	ContractId  string
	Txid        string
	DeleteImage bool
}
type DataJson struct {
	MainData  string
	ExtraData string
}

func ConvertTx2Json(tx *modules.Transaction, utxoQuery modules.QueryUtxoFunc) TxJson {
	json := TxJson{TxHash: tx.Hash().String(), TxSize: float64(tx.Size())}
	for _, m := range tx.TxMessages {
		if m.App == modules.APP_PAYMENT {
			pay := m.Payload.(*modules.PaymentPayload)
			if utxoQuery == nil {
				payJson := ConvertPayment2Json(pay)
				json.Payment = payJson
			} else {
				payJson := ConvertPayment2JsonIncludeFromAddr(pay, utxoQuery)
				json.Payment = payJson
			}
		} else if m.App == modules.APP_VOTE {
			v := m.Payload.(*vote.VoteInfo)
			if v.VoteType == vote.TypeMediator {
				vote := &VoteJson{Content: string(v.Contents)}
				json.Vote = vote
			}
		} else if m.App == modules.APP_DATA {
			data := m.Payload.(*modules.DataPayload)
			json.Data = &DataJson{MainData: string(data.MainData), ExtraData: string(data.ExtraData)}
		} else if m.App == modules.APP_CONTRACT_TPL_REQUEST {
			req := m.Payload.(*modules.ContractInstallRequestPayload)
			json.InstallRequest = convertInstallRequest2Json(req)
		} else if m.App == modules.APP_CONTRACT_DEPLOY_REQUEST {
			req := m.Payload.(*modules.ContractDeployRequestPayload)
			json.DeployRequest = convertDeployRequest2Json(req)
		} else if m.App == modules.APP_CONTRACT_INVOKE_REQUEST {
			req := m.Payload.(*modules.ContractInvokeRequestPayload)
			json.InvokeRequest = convertInvokeRequest2Json(req)
		} else if m.App == modules.APP_CONTRACT_STOP_REQUEST {
			req := m.Payload.(*modules.ContractStopRequestPayload)
			json.StopRequest = convertStopRequest2Json(req)
		}
	}
	return json
}
func ConvertJson2Tx(json *TxJson) *modules.Transaction {
	tx := &modules.Transaction{}
	pay := ConvertJson2Payment(json.Payment)
	tx.AddMessage(modules.NewMessage(modules.APP_PAYMENT, pay))
	return tx
}

func convertInvokeRequest2Json(req *modules.ContractInvokeRequestPayload) *InvokeRequestJson {
	addr := common.NewAddress(req.ContractId, common.ContractHash)
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
	reqJson.TplId = string(req.TplId)
	reqJson.Args = []string{}
	for _, arg := range req.Args {
		reqJson.Args = append(reqJson.Args, string(arg))
	}
	return reqJson
}

func convertStopRequest2Json(req *modules.ContractStopRequestPayload) *StopRequestJson {
	reqJson := &StopRequestJson{}
	reqJson.ContractId = string(req.ContractId)
	reqJson.Txid = req.Txid

	return reqJson
}
