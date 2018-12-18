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
)

type TxJson struct {
	TxHash        string             `json:"tx_hash"`
	Payment       *PaymentJson       `json:"payment"`
	Vote          *VoteJson          `json:"vote"`
	InvokeRequest *InvokeRequestJson `json:"invoke_request"`
}
type VoteJson struct {
	Content string
}
type InvokeRequestJson struct {
	ContractAddr string
	FunctionName string
	Args         []string
}

func ConvertTx2Json(tx *modules.Transaction) TxJson {
	json := TxJson{TxHash: tx.Hash().String()}
	for _, m := range tx.TxMessages {
		if m.App == modules.APP_PAYMENT {
			pay := m.Payload.(*modules.PaymentPayload)
			payJson := ConvertPayment2Json(pay)
			json.Payment = &payJson
		} else if m.App == modules.APP_VOTE {
			v := m.Payload.(*vote.VoteInfo)
			if v.VoteType == vote.TypeMediator {
				vote := &VoteJson{Content: string(v.Contents)}
				json.Vote = vote
			}
		} else if m.App == modules.APP_CONTRACT_INVOKE_REQUEST {
			req := m.Payload.(*modules.ContractInvokeRequestPayload)
			json.InvokeRequest = convertInvokeRequest2Json(req)
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
