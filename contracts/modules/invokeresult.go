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

package modules

import (
	"encoding/json"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/dag"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/tokenengine"
	"time"
)

type ContractInvokeResult struct {
	ContractId    []byte                     `json:"contract_id"` // contract id
	RequestId     common.Hash                `json:"request_id"`
	FunctionName  string                     `json:"function_name"`
	Args          [][]byte                   `json:"args"`           // contract arguments list
	ExecutionTime time.Duration              `json:"execution_time"` // contract execution time, millisecond
	ReadSet       []modules.ContractReadSet  `json:"read_set"`       // the set data of read, and value could be any type
	WriteSet      []modules.ContractWriteSet `json:"write_set"`      // the set data of write, and value could be any type
	Payload       []byte                     `json:"payload"`        // the contract execution result
	TokenPayOut   []*modules.TokenPayOut     `json:"token_payout"`   //从合约地址付出Token
	TokenSupply   []*modules.TokenSupply     `json:"token_supply"`   //增发Token请求产生的结果
	TokenDefine   *modules.TokenDefine       `json:"token_define"`   //定义新Token
}

func (result *ContractInvokeResult) ToContractInvokePayload() *modules.ContractInvokePayload {
	return modules.NewContractInvokePayload(result.ContractId, result.Args, result.ExecutionTime, result.ReadSet, result.WriteSet, result.Payload)
}
func (result *ContractInvokeResult) ToContractPayments(dag dag.IDag) ([]*modules.PaymentPayload, error) {
	addr := common.NewAddress(result.ContractId, common.ContractHash)
	payments := []*modules.PaymentPayload{}
	if result.TokenPayOut != nil && len(result.TokenPayOut) > 0 {
		for _, payout := range result.TokenPayOut {
			utxos, err := dag.GetAddr1TokenUtxos(addr.String(), payout.Asset)
			if err != nil {
				return nil, err
			}
			utxo2 := convertMapUtxo(utxos)
			us := core.Utxos{}
			for _, u := range utxo2 {
				us = append(us, u)
			}
			selected, change, err := core.Select_utxo_Greedy(us, payout.Amount)
			if err != nil {
				return nil, err
			}
			payment := &modules.PaymentPayload{}
			for _, s := range selected {
				sutxo := s.(*modules.UtxoWithOutPoint)
				in := modules.NewTxIn(&sutxo.OutPoint, nil)
				payment.AddTxIn(in)
			}
			out := modules.NewTxOut(payout.Amount, tokenengine.GenerateLockScript(payout.PayTo), payout.Asset)
			payment.AddTxOut(out)
			//Change
			out2 := modules.NewTxOut(change, tokenengine.GenerateLockScript(addr), payout.Asset)
			payment.AddTxOut(out2)
			payments = append(payments, payment)
		}
	}
	return payments, nil
}
func convertMapUtxo(utxo map[modules.OutPoint]*modules.Utxo) []*modules.UtxoWithOutPoint {
	var result []*modules.UtxoWithOutPoint
	for o, u := range utxo {
		uo := &modules.UtxoWithOutPoint{}
		uo.Set(u, &o)
		result = append(result, uo)
	}
	return result
}
func (result *ContractInvokeResult) ToCoinbase() ([]*modules.PaymentPayload, error) {
	var coinbases []*modules.PaymentPayload
	if result.TokenDefine != nil {
		coinbase := &modules.PaymentPayload{}
		if result.TokenDefine.TokenType == 0 { //ERC20
			token := modules.FungibleToken{}
			err := json.Unmarshal(result.TokenDefine.TokenDefineJson, &token)
			if err != nil {
				log.Error("Cannot parse token define json to FungibleToken", result.TokenDefine.TokenDefineJson)
				return nil, err
			}
			newAsset := &modules.Asset{}
			newAsset.AssetId, _ = modules.NewAssetId(token.Symbol, modules.AssetType_FungibleToken, result.RequestId.Bytes())
			out := modules.NewTxOut(token.TotalSupply, tokenengine.GenerateLockScript(result.TokenDefine.Creator), newAsset)
			coinbase.AddTxOut(out)
		}
		//TODO Devin ERC721
		coinbases = append(coinbases, coinbase)
	}
	if result.TokenSupply != nil && len(result.TokenSupply) > 0 {

	}
	return coinbases, nil
}
