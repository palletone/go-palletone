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
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/dag/modules"
	"unsafe"
)

type TransactionJson struct {
	TxHash     string       `json:"txhash"`
	UnitHash   string       `json:"unithash"`
	Payment    *PaymentJson `json:"payment"`
	TxMessages string       `json:"txmessages"`
}
type MessageJson struct {
	messages []string
}

func ConvertTx02Json(tx *modules.Transaction, hash common.Hash) *TransactionJson {
	var hexHash string
	pay := tx.TxMessages[0].Payload.(*modules.PaymentPayload)
	if hash != (common.Hash{}) {
		hexHash = hash.String()
	}
	payment := ConvertPayment2Json(pay)
	return &TransactionJson{
		TxHash:     tx.Hash().String(),
		UnitHash:   hexHash,
		Payment:    &payment,
		TxMessages: ConvertMegs2Json(tx.TxMessages),
	}
}

func ConvertMegs2Json(msgs []*modules.Message) string {
	data, err := json.Marshal(msgs)
	if err != nil {
		return ""
	}
	return *(*string)(unsafe.Pointer(&data))
}

// TODO
func isCoinBase(tx *modules.Transaction) bool {
	if len(tx.TxMessages) != 1 {
		return false
	}
	msg, ok := tx.TxMessages[0].Payload.(*modules.PaymentPayload)
	if !ok {
		return false
	}
	prevOut := msg.Inputs[0].PreviousOutPoint
	if prevOut.TxHash != (common.Hash{}) {
		return false
	}
	return true
}

type GetTranscationOut struct {
	Addr  string `json:"address"`
	Value uint64 `json:"vout"`
	Asset string `json:"asset"`
}
type GetTransactions struct {
	Txid    string              `json:"txid"`
	Inputs  []string            `json:"inputs"`
	Outputs []GetTranscationOut `json:"outputs"`
}

func ConvertGetTransactions2Json(gets []GetTransactions) string {
	data, err := json.Marshal(gets)
	if err != nil {
		return ""
	}
	return *(*string)(unsafe.Pointer(&data))
}
