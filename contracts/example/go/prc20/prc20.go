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
	"github.com/palletone/go-palletone/contracts/shim"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	dm "github.com/palletone/go-palletone/dag/modules"
)

const symbolsKey = "symbols"

type PRC20 struct {
}

type TokenInfo struct {
	SupplyAddr string
	AssetID    dm.IDType16
}

type Symbols struct {
	TokenInfos map[string]TokenInfo `json:"tokeninfos"`
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

func setSymbols(symbols *Symbols, stub shim.ChaincodeStubInterface) error {
	val, err := json.Marshal(symbols)
	if err != nil {
		return err
	}
	err = stub.PutState(symbolsKey, val)
	return err
}

func getSymbols(stub shim.ChaincodeStubInterface) (*Symbols, error) {
	//
	symbols := Symbols{TokenInfos: map[string]TokenInfo{}}
	symbolsBytes, err := stub.GetState(symbolsKey)
	if err != nil {
		return &symbols, err
	}
	//
	err = json.Unmarshal(symbolsBytes, &symbols)
	if err != nil {
		return &symbols, err
	}

	return &symbols, nil
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
		return shim.Success([]byte(jsonResp))
	}
	if len(fungible.Symbol) > 5 {
		jsonResp := "{\"Error\":\"Symbol must less than 5 characters\"}"
		return shim.Success([]byte(jsonResp))
	}
	//decimals
	decimals, _ := strconv.ParseUint(args[2], 10, 64)
	if decimals > 18 {
		jsonResp := "{\"Error\":\"Can't big than 18\"}"
		return shim.Success([]byte(jsonResp))
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
		return shim.Success([]byte(jsonResp))
	}
	fungible.TotalSupply = totalSupply
	//address of supply
	if len(args) > 4 {
		fungible.SupplyAddress = args[4]
	}

	//check name is only or not
	symbols, err := getSymbols(stub)
	if _, ok := symbols.TokenInfos[fungible.Symbol]; ok {
		jsonResp := "{\"Error\":\"The symbol have been used\"}"
		return shim.Success([]byte(jsonResp))
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

	//set token define
	err = stub.DefineToken(byte(0), createJson, createAddr)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to call stub.DefineToken\"}"
		return shim.Error(jsonResp)
	}

	//last put state
	if fungible.SupplyAddress != "" {
		txid := stub.GetTxID()
		assetID, _ := dm.NewAssetId(fungible.Symbol, dm.AssetType_FungibleToken,
			fungible.Decimals, common.Hex2Bytes(txid[2:]))
		info := TokenInfo{SupplyAddr: fungible.SupplyAddress, AssetID: assetID}
		symbols.TokenInfos[fungible.Symbol] = info
	} else {
		symbols.TokenInfos[fungible.Symbol] = TokenInfo{}
	}
	err = setSymbols(symbols, stub)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to set symbols\"}"
		return shim.Error(jsonResp)
	}
	return shim.Success(createJson) //test
}

func supplyToken(args []string, stub shim.ChaincodeStubInterface) pb.Response {
	//params check
	if len(args) < 2 {
		return shim.Error("need 2 args (Symbol,SupplyAmout)")
	}

	//symbol
	symbol := strings.ToUpper(args[0])
	//check name is exist or not
	symbols, err := getSymbols(stub)
	if _, ok := symbols.TokenInfos[symbol]; !ok {
		jsonResp := "{\"Error\":\"Token not exist\"}"
		return shim.Success([]byte(jsonResp))
	}

	//supply amount
	supplyAmount, err := strconv.ParseUint(args[1], 10, 64)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to convert supply amount\"}"
		return shim.Error(jsonResp)
	}
	if supplyAmount == 0 {
		jsonResp := "{\"Error\":\"Can't be zero\"}"
		return shim.Success([]byte(jsonResp))
	}

	//get invoke address
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get invoke address\"}"
		return shim.Error(jsonResp)
	}
	//check supply address
	if invokeAddr != symbols.TokenInfos[symbol].SupplyAddr {
		jsonResp := "{\"Error\":\"Not the supply address\"}"
		return shim.Success([]byte(jsonResp))
	}

	//call SupplyToken
	assetID := symbols.TokenInfos[symbol].AssetID
	err = stub.SupplyToken(assetID.Bytes(),
		[]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, supplyAmount, symbols.TokenInfos[symbol].SupplyAddr)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to call stub.SupplyToken\"}"
		return shim.Error(jsonResp)
	}
	return shim.Success([]byte("")) //test
}

type TokenIDInfo struct {
	Symbol     string
	AssetID    string
	SupplyAddr string
}

func oneToken(args []string, stub shim.ChaincodeStubInterface) pb.Response {
	//params check
	if len(args) < 1 {
		return shim.Error("need 1 args (Symbol)")
	}

	//symbol
	symbol := strings.ToUpper(args[0])
	//check name is exist or not
	symbols, err := getSymbols(stub)
	if _, ok := symbols.TokenInfos[symbol]; !ok {
		jsonResp := "{\"Error\":\"Token not exist\"}"
		return shim.Success([]byte(jsonResp))
	}

	//token
	asset := symbols.TokenInfos[symbol].AssetID
	tkID := TokenIDInfo{symbol, asset.ToAssetId(), symbols.TokenInfos[symbol].SupplyAddr}
	//return json
	tkJson, err := json.Marshal(tkID)
	if err != nil {
		return shim.Success([]byte(err.Error()))
	}
	return shim.Success(tkJson) //test
}

func allToken(args []string, stub shim.ChaincodeStubInterface) pb.Response {
	symbols, err := getSymbols(stub)

	var tkIDs []TokenIDInfo
	for symbol := range symbols.TokenInfos {
		asset := symbols.TokenInfos[symbol].AssetID
		tkID := TokenIDInfo{symbol, asset.ToAssetId(), symbols.TokenInfos[symbol].SupplyAddr}
		tkIDs = append(tkIDs, tkID)
	}

	//return json
	tksJson, err := json.Marshal(tkIDs)
	if err != nil {
		return shim.Success([]byte(err.Error()))
	}
	return shim.Success(tksJson) //test
}
