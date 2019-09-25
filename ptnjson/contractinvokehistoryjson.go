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
 *  * @date 2018-2019
 *
 */

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
 *  * @date 2018-2019
 *
 */

package ptnjson

import (
	"time"

	"github.com/palletone/go-palletone/dag/modules"
)

type ContractInvokeHistoryJson struct {
	TxHash     string              `json:"tx_hash"`
	RequestId  string              `json:"request_id"`
	UnitHash   string              `json:"unit_hash"`
	TxSize     float64             `json:"tx_size"`
	Payment    *PaymentSummaryJson `json:"payment"`
	InvokeJson *InvokeJson         `json:"invoke_result"`

	InvokeRequest *InvokeRequestJson `json:"invoke_request"`
	Timestamp     string             `json:"timestamp"`
}

func ConvertTx2ContractInvokeHistoryJson(tx *modules.TransactionWithUnitInfo,
	utxoQuery modules.QueryUtxoFunc) *ContractInvokeHistoryJson {
	json := &ContractInvokeHistoryJson{
		TxHash:    tx.Hash().String(),
		UnitHash:  tx.UnitHash.String(),
		TxSize:    float64(tx.Size()),
		RequestId: tx.RequestHash().String(),
	}
	for _, m := range tx.TxMessages {
		if m.App == modules.APP_PAYMENT {
			pay := m.Payload.(*modules.PaymentPayload)

			payJson := ConvertPayment2SummaryJson(pay, utxoQuery)
			json.Payment = payJson

		} else if m.App == modules.APP_CONTRACT_INVOKE_REQUEST {
			req := m.Payload.(*modules.ContractInvokeRequestPayload)
			json.InvokeRequest = convertInvokeRequest2Json(req)
		} else if m.App == modules.APP_CONTRACT_INVOKE {
			req := m.Payload.(*modules.ContractInvokePayload)
			json.InvokeJson = convertInvoke2Json(req)
		}
	}
	t := time.Unix(int64(tx.Timestamp), 0)
	json.Timestamp = t.String()
	return json
}
