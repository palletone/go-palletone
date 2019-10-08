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

package jury

import (
	"encoding/json"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/dag/errors"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/tokenengine"
)

//将ContractInvokeResult中合约付款出去的请求转换为UTXO对应的Payment
func resultToContractPayments(dag iDag, result *modules.ContractInvokeResult) ([]*modules.PaymentPayload, error) {
	addr := common.NewAddress(result.ContractId, common.ContractHash)
	payments := []*modules.PaymentPayload{}
	if result.TokenPayOut != nil && len(result.TokenPayOut) > 0 {
		payouts := tokenPayOutGroupByAsset(result.TokenPayOut)
		for ast, aa := range payouts {
			asset := ast
			utxos, err := dag.GetAddr1TokenUtxos(addr, &asset)
			if err != nil {
				return nil, err
			}
			utxo2 := convertMapUtxo(utxos)
			us := core.Utxos{}
			for _, u := range utxo2 {
				us = append(us, u)
			}
			totalPayAmt := uint64(0)
			for _, a := range aa {
				totalPayAmt += a.Amount
			}
			selected, change, err := core.Select_utxo_Greedy(us, totalPayAmt)
			if err != nil {
				return nil, err
			}
			payment := &modules.PaymentPayload{}
			for _, s := range selected {
				sutxo := s.(*modules.UtxoWithOutPoint)
				in := modules.NewTxIn(&sutxo.OutPoint, nil)
				payment.AddTxIn(in)
			}
			for _, a := range aa {
				out := modules.NewTxOut(a.Amount, tokenengine.Instance.GenerateLockScript(a.Address), &asset)
				payment.AddTxOut(out)
			}
			//Change
			if change > 0 {
				out2 := modules.NewTxOut(change, tokenengine.Instance.GenerateLockScript(addr), &asset)
				payment.AddTxOut(out2)
			}
			payments = append(payments, payment)
		}
	}
	return payments, nil
}

type addrAmount struct {
	Address common.Address
	Amount  uint64
}

func tokenPayOutGroupByAsset(payouts []*modules.TokenPayOut) map[modules.Asset][]*addrAmount {
	result := make(map[modules.Asset][]*addrAmount)
	for _, payout := range payouts {
		asset := *payout.Asset
		if aa, ok := result[asset]; ok {
			hasSameAddr := false
			for _, a := range aa {
				if a.Address == payout.PayTo {
					hasSameAddr = true
					a.Amount += payout.Amount
					break
				}
			}
			if !hasSameAddr {
				aa = append(aa, &addrAmount{Address: payout.PayTo, Amount: payout.Amount})
			}
			result[asset] = aa
		} else {
			result[asset] = []*addrAmount{{Address: payout.PayTo, Amount: payout.Amount}}
		}
	}
	return result
}

func resultToCoinbase(result *modules.ContractInvokeResult) ([]*modules.PaymentPayload, error) {
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
			newAsset.AssetId, _ = modules.NewAssetId(token.Symbol, modules.AssetType_FungibleToken,
				token.Decimals, result.RequestId.Bytes(), modules.UniqueIdType_Null)
			out := modules.NewTxOut(token.TotalSupply, tokenengine.Instance.GenerateLockScript(result.TokenDefine.Creator), newAsset)
			coinbase.AddTxOut(out)
		} else if result.TokenDefine.TokenType == 1 { //ERC721
			token := modules.NonFungibleToken{}
			err := json.Unmarshal(result.TokenDefine.TokenDefineJson, &token)
			if err != nil {
				log.Error("Cannot parse token define json to NonFungibleToken", result.TokenDefine.TokenDefineJson)
				return nil, err
			}

			for i := uint64(0); i < token.TotalSupply; i++ {
				if len(token.NonFungibleData[i].UniqueBytes) < 16 {
					return nil, errors.New("UniqueBytes's len must bigger than 16")
				}
				newAsset := &modules.Asset{}
				newAsset.AssetId, _ = modules.NewAssetId(token.Symbol, modules.AssetType_NonFungibleToken,
					0, result.RequestId.Bytes(), modules.UniqueIdType(token.Type))
				newAsset.UniqueId.SetBytes(token.NonFungibleData[i].UniqueBytes)
				out := modules.NewTxOut(1, tokenengine.Instance.GenerateLockScript(result.TokenDefine.Creator), newAsset)
				coinbase.AddTxOut(out)
			}
		} else if result.TokenDefine.TokenType == 2 { //VoteToken
			token := modules.VoteToken{}
			err := json.Unmarshal(result.TokenDefine.TokenDefineJson, &token)
			if err != nil {
				log.Error("Cannot parse token define json to VoteToken", result.TokenDefine.TokenDefineJson)
				return nil, err
			}
			newAsset := &modules.Asset{}
			newAsset.AssetId, _ = modules.NewAssetId(token.Symbol, modules.AssetType_VoteToken,
				0, result.RequestId.Bytes(), modules.UniqueIdType_Null)
			out := modules.NewTxOut(token.TotalSupply, tokenengine.Instance.GenerateLockScript(result.TokenDefine.Creator), newAsset)
			coinbase.AddTxOut(out)
		}
		//TODO Devin ERC721
		coinbases = append(coinbases, coinbase)
	}
	if result.TokenSupply != nil && len(result.TokenSupply) > 0 {
		coinbase := &modules.PaymentPayload{}
		for _, tokenSupply := range result.TokenSupply {
			assetId := &modules.Asset{}
			assetId.AssetId.SetBytes(tokenSupply.AssetId)
			assetId.UniqueId.SetBytes(tokenSupply.UniqueId)
			out := modules.NewTxOut(tokenSupply.Amount, tokenengine.Instance.GenerateLockScript(tokenSupply.Creator), assetId)
			//
			coinbase.AddTxOut(out)
		}
		coinbases = append(coinbases, coinbase)

	}
	return coinbases, nil
}

func convertMapUtxo(utxo map[modules.OutPoint]*modules.Utxo) []*modules.UtxoWithOutPoint {
	result := make([]*modules.UtxoWithOutPoint, 0, len(utxo))
	for o, u := range utxo {
		uo := &modules.UtxoWithOutPoint{}
		o1 := o
		uo.Set(u, &o1)
		result = append(result, uo)
	}
	return result
}
