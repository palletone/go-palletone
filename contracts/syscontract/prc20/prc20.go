/*
	This file is part of go-palletone.
	go-palletone is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.
	go-palletone is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.
	You should have received a copy of the GNU General Public License
	along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
*/
/*
 * Copyright IBM Corp. All Rights Reserved.
 * @author PalletOne core developers <dev@pallet.one>
 * @date 2018
 */

package prc20

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/math"
	"github.com/palletone/go-palletone/contracts/shim"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	dm "github.com/palletone/go-palletone/dag/modules"
)

const symbolsKey = "symbol_"

type PRC20 struct {
}

type TokenInfo struct {
	Symbol      string
	CreateAddr  string
	TotalSupply uint64
	Decimals    uint64
	SupplyAddr  string
	AssetID     dm.AssetId
}

func (p *PRC20) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

func (p *PRC20) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	f, args := stub.GetFunctionAndParameters()

	switch f {
	case "createToken":
		return createToken(args, stub)
	case "supplyToken":
		return supplyToken(args, stub)
	case "getTokenInfo":
		return oneToken(args, stub)
	case "getAllTokenInfo":
		return allToken(args, stub)
	default:
		jsonResp := "{\"Error\":\"Unknown function " + f + "\"}"
		return shim.Error(jsonResp)
	}
}

func setSymbols(stub shim.ChaincodeStubInterface, tkInfo *TokenInfo) error {
	val, err := json.Marshal(tkInfo)
	if err != nil {
		return err
	}
	err = stub.PutState(symbolsKey+tkInfo.Symbol, val)
	return err
}
func getSymbols(stub shim.ChaincodeStubInterface, symbol string) *TokenInfo {
	//
	tkInfo := TokenInfo{}
	tkInfoBytes, _ := stub.GetState(symbolsKey + symbol)
	if len(tkInfoBytes) == 0 {
		return nil
	}
	//
	err := json.Unmarshal(tkInfoBytes, &tkInfo)
	if err != nil {
		return nil
	}

	return &tkInfo
}

func getSymbolsAll(stub shim.ChaincodeStubInterface) []TokenInfo {
	KVs, _ := stub.GetStateByPrefix(symbolsKey)
	var tkInfos []TokenInfo
	for _, oneKV := range KVs {
		tkInfo := TokenInfo{}
		err := json.Unmarshal(oneKV.Value, &tkInfo)
		if err != nil {
			continue
		}
		tkInfos = append(tkInfos, tkInfo)
	}
	return tkInfos
}

func createToken(args []string, stub shim.ChaincodeStubInterface) pb.Response {
	//params check
	if len(args) < 4 {
		return shim.Error("need 4 args (Name,Symbol,Decimals,TotalSupply,[SupplyAddress])")
	}

	//==== convert params to token information
	var fungible dm.FungibleToken
	//name symbol
	fungible.Name = args[0]
	fungible.Symbol = strings.ToUpper(args[1])
	if fungible.Symbol == "PTN" {
		jsonResp := "{\"Error\":\"Can't use PTN\"}"
		return shim.Error(jsonResp)
	}
	if len(fungible.Symbol) > 5 {
		jsonResp := "{\"Error\":\"Symbol must less than 5 characters\"}"
		return shim.Error(jsonResp)
	}
	//decimals
	decimals, _ := strconv.ParseUint(args[2], 10, 64)
	if decimals > 18 {
		jsonResp := "{\"Error\":\"Can't big than 18\"}"
		return shim.Error(jsonResp)
	}
	fungible.Decimals = byte(decimals)
	//total supply
	totalSupply, err := strconv.ParseUint(args[3], 10, 64)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to convert total supply\"}"
		return shim.Error(jsonResp)
	}
	if totalSupply == 0 {
		jsonResp := "{\"Error\":\"Can't be zero\"}"
		return shim.Error(jsonResp)
	}
	fungible.TotalSupply = totalSupply
	//address of supply
	if len(args) > 4 {
		fungible.SupplyAddress = args[4]
	}

	//check name is only or not
	tkInfo := getSymbols(stub, fungible.Symbol)
	if tkInfo != nil {
		jsonResp := "{\"Error\":\"The symbol have been used\"}"
		return shim.Error(jsonResp)
	}

	//convert to json
	createJson, err := json.Marshal(fungible)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to generate token Json\"}"
		return shim.Error(jsonResp)
	}
	//get creator
	createAddr, err := stub.GetInvokeAddress()
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get invoke address\"}"
		return shim.Error(jsonResp)
	}

	//last put state
	txid := stub.GetTxID()
	assetID, _ := dm.NewAssetId(fungible.Symbol, dm.AssetType_FungibleToken,
		fungible.Decimals, common.Hex2Bytes(txid[2:]), dm.UniqueIdType_Null)
	info := TokenInfo{fungible.Symbol, createAddr.String(), totalSupply, decimals,
		fungible.SupplyAddress, assetID}

	err = setSymbols(stub, &info)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to set symbols\"}"
		return shim.Error(jsonResp)
	}

	//set token define
	err = stub.DefineToken(byte(dm.AssetType_FungibleToken), createJson, createAddr.String())
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to call stub.DefineToken\"}"
		return shim.Error(jsonResp)
	}

	return shim.Success(createJson)
}

func supplyToken(args []string, stub shim.ChaincodeStubInterface) pb.Response {
	//params check
	if len(args) < 2 {
		return shim.Error("need 2 args (Symbol,SupplyAmout)")
	}

	//symbol
	symbol := strings.ToUpper(args[0])
	//check name is exist or not
	tkInfo := getSymbols(stub, symbol)
	if tkInfo == nil {
		jsonResp := "{\"Error\":\"Token not exist\"}"
		return shim.Error(jsonResp)
	}

	//supply amount
	supplyAmount, err := strconv.ParseUint(args[1], 10, 64)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to convert supply amount\"}"
		return shim.Error(jsonResp)
	}
	if supplyAmount == 0 {
		jsonResp := "{\"Error\":\"Can't be zero\"}"
		return shim.Error(jsonResp)
	}
	if math.MaxInt64-tkInfo.TotalSupply < supplyAmount {
		jsonResp := "{\"Error\":\"Too big, overflow\"}"
		return shim.Error(jsonResp)
	}

	//get invoke address
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get invoke address\"}"
		return shim.Error(jsonResp)
	}
	//check supply address
	if invokeAddr.String() != tkInfo.SupplyAddr {
		jsonResp := "{\"Error\":\"Not the supply address\"}"
		return shim.Error(jsonResp)
	}

	//call SupplyToken
	assetID := tkInfo.AssetID
	err = stub.SupplyToken(assetID.Bytes(),
		[]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, supplyAmount, tkInfo.SupplyAddr)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to call stub.SupplyToken\"}"
		return shim.Error(jsonResp)
	}

	//add supply
	tkInfo.TotalSupply += supplyAmount
	err = setSymbols(stub, tkInfo)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to set symbols\"}"
		return shim.Error(jsonResp)
	}
	return shim.Success([]byte(""))
}

type TokenIDInfo struct {
	Symbol      string
	CreateAddr  string
	TotalSupply uint64
	Decimals    uint64
	SupplyAddr  string
	AssetID     string
}

func oneToken(args []string, stub shim.ChaincodeStubInterface) pb.Response {
	//params check
	if len(args) < 1 {
		return shim.Error("need 1 args (Symbol)")
	}

	//symbol
	symbol := strings.ToUpper(args[0])
	//check name is exist or not
	tkInfo := getSymbols(stub, symbol)
	if tkInfo == nil {
		jsonResp := "{\"Error\":\"Token not exist\"}"
		return shim.Error(jsonResp)
	}

	//token
	asset := tkInfo.AssetID
	tkID := TokenIDInfo{symbol, tkInfo.CreateAddr, tkInfo.TotalSupply,
		tkInfo.Decimals, tkInfo.SupplyAddr, asset.String()}
	//return json
	tkJson, err := json.Marshal(tkID)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(tkJson)
}

func allToken(args []string, stub shim.ChaincodeStubInterface) pb.Response {
	tkInfos := getSymbolsAll(stub)

	var tkIDs []TokenIDInfo
	for _, tkInfo := range tkInfos {
		asset := tkInfo.AssetID
		tkID := TokenIDInfo{tkInfo.Symbol, tkInfo.CreateAddr, tkInfo.TotalSupply,
			tkInfo.Decimals, tkInfo.SupplyAddr, asset.String()}
		tkIDs = append(tkIDs, tkID)
	}

	//return json
	tksJson, err := json.Marshal(tkIDs)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(tksJson)
}
