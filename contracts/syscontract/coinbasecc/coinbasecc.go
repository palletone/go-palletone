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

package coinbasecc

import (
	"encoding/json"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/contracts/shim"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"github.com/palletone/go-palletone/dag/constants"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/shopspring/decimal"
)

type CoinbaseChainCode struct {
}

func (d *CoinbaseChainCode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success([]byte("ok"))
}

func (d *CoinbaseChainCode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	funcName, args := stub.GetFunctionAndParameters()
	switch funcName {
	case "queryReward":
		return d.queryGenerateUnitReward(stub, args)
	default:
		return shim.Error("coinbase cc Invoke error" + funcName)
	}
	return shim.Error("coinbase cc Invoke error" + funcName)
}

//出块奖励记录查询
func (d *CoinbaseChainCode) queryGenerateUnitReward(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	kvs, err := stub.GetStateByPrefix(constants.RewardAddressPrefix)
	if err != nil {
		return shim.Error(err.Error())
	}
	log.Debugf("queryGenerateUnitReward, return count:%d", len(kvs))
	result := []*RewardRecord{}
	for _, kv := range kvs {
		aa := []modules.AmountAsset{}
		rlp.DecodeBytes(kv.Value, &aa)
		addr := kv.Key[len(constants.RewardAddressPrefix):]
		for _, a := range aa {
			record := &RewardRecord{
				Address: addr,
				Token:   a.Asset,
				Amount:  a.Asset.DisplayAmount(a.Amount),
			}
			result = append(result, record)
		}
	}
	data, _ := json.Marshal(result)
	return shim.Success(data)
}

type RewardRecord struct {
	Address string
	Amount  decimal.Decimal
	Token   *modules.Asset
}
