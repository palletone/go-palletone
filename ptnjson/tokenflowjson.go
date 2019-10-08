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
 *  * @author PalletOne core developers <dev@pallet.one>
 *  * @date 2018-2019
 *
 *
 */

package ptnjson

import (
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/tokenengine"
	"github.com/shopspring/decimal"
	"time"
)

type TokenFlowJson struct {
	TxHash        string          `json:"tx_hash"`
	UnitHash      string          `json:"unit_hash"`
	Timestamp     string          `json:"timestamp"`
	Direction     string          `json:"direction"` // + -
	FromAddresses string          `json:"from_addrs"`
	ToAddresses   string          `json:"to_addrs"`
	Amount        decimal.Decimal `json:"amount"`
	Fee           decimal.Decimal `json:"fee"`
	Asset         string          `json:"asset"`
	Balance       decimal.Decimal `json:"balance"`
}

func ConvertTx2TokenFlowJson(addr common.Address, token *modules.Asset, preBalance uint64,
	tx *modules.TransactionWithUnitInfo, utxoQuery modules.QueryUtxoFunc) ([]*TokenFlowJson, uint64) {
	result := []*TokenFlowJson{}
	for _, m := range tx.TxMessages {
		if m.App == modules.APP_PAYMENT {
			pay := m.Payload.(*modules.PaymentPayload)
			spent := uint64(0)
			income := uint64(0)
			fee := uint64(0)
			fromAddrMap := make(map[common.Address]bool)
			toAddrMap := make(map[common.Address]bool)
			totalInput := uint64(0)
			for _, input := range pay.Inputs {
				utxo, _ := utxoQuery(input.PreviousOutPoint)
				if !utxo.Asset.Equal(token) {
					break
				}
				totalInput += utxo.Amount
				fromAddr, _ := tokenengine.Instance.GetAddressFromScript(utxo.PkScript)
				if fromAddr == addr {
					spent += utxo.Amount
				}
				fromAddrMap[fromAddr] = true
			}
			totalOutput := uint64(0)
			for _, out := range pay.Outputs {
				toAddr, _ := tokenengine.Instance.GetAddressFromScript(out.PkScript)
				if !out.Asset.Equal(token) {
					break
				}
				totalOutput += out.Value
				if toAddr == addr {
					income += out.Value
				}
				toAddrMap[toAddr] = true
			}
			if spent == 0 && income == 0 { //不相关的Payment
				continue
			}
			json := &TokenFlowJson{
				TxHash:    tx.Hash().String(),
				UnitHash:  tx.UnitHash.String(),
				Direction: "+",
				Asset:     token.String(),
			}
			t := time.Unix(int64(tx.Timestamp), 0)
			json.Timestamp = t.String()

			if totalInput != 0 { //计算手续费
				fee = totalInput - totalOutput
				json.Fee = token.DisplayAmount(fee)
			}

			json.FromAddresses = joinAddress(fromAddrMap)
			preBalance += income - spent
			json.Balance = token.DisplayAmount(preBalance)
			if spent > income {
				json.Direction = "-" //花费出去
				json.Amount = token.DisplayAmount(spent - income)
				delete(toAddrMap, addr) //删除找零地址
				json.ToAddresses = joinAddress(toAddrMap)
			} else {
				json.Direction = "+" //收入
				json.Amount = token.DisplayAmount(income - spent)
				json.ToAddresses = addr.String()
			}
			result = append(result, json)
		}
	}
	return result, preBalance
}
func joinAddress(addrMap map[common.Address]bool) string {
	str := ""
	for key := range addrMap {
		str += key.String() + ";"
	}
	return str
}
