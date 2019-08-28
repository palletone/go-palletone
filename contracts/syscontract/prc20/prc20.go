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
	"fmt"
	"strconv"
	"strings"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/math"
	"github.com/palletone/go-palletone/contracts/shim"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	dm "github.com/palletone/go-palletone/dag/modules"

	"github.com/shopspring/decimal"
)

const symbolsKey = "symbol_"
const jsonResp1 = "{\"Error\":\"Failed to get invoke address\"}"
const jsonResp2 = "{\"Error\":\"Token not exist\"}"
const jsonResp3 = "{\"Error\":\"Failed to set symbols\"}"
const jsonResp4 = "{\"Error\":\"Failed to add global state\"}"

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

func paramCheckValid(args []string) (bool, string) {
	if len(args) > 32 {
		return false, "args number out of range 32"
	}
	for _, arg := range args {
		if len(arg) > 2048 {
			return false, "arg length out of range 2048"
		}
	}
	return true, ""
}

func (p *PRC20) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

func (p *PRC20) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	f, args := stub.GetFunctionAndParameters()

	valid, errMsg := paramCheckValid(args)
	if !valid {
		return shim.Error(errMsg)
	}
	switch f {
	case "createToken":
		return createToken(args, stub)
	case "supplyToken":
		return supplyToken(args, stub)
	case "getTokenInfo":
		return oneToken(args, stub)
	case "getAllTokenInfo":
		return allToken(stub)
	case "changeSupplyAddr":
		return changeSupplyAddr(args, stub)
	case "frozenToken":
		return frozenToken(args, stub)
	default:
		jsonResp := "{\"Error\":\"Unknown function " + f + "\"}"
		return shim.Error(jsonResp)
	}
}

func setGlobal(stub shim.ChaincodeStubInterface, tkInfo *TokenInfo) error {
	gTkInfo := dm.GlobalTokenInfo{Symbol: tkInfo.Symbol, TokenType: 1, Status: 0, CreateAddr: tkInfo.CreateAddr,
		TotalSupply: tkInfo.TotalSupply, SupplyAddr: tkInfo.SupplyAddr, AssetID: tkInfo.AssetID}
	val, err := json.Marshal(gTkInfo)
	if err != nil {
		return err
	}
	err = stub.PutGlobalState(dm.GlobalPrefix+gTkInfo.Symbol, val)
	return err
}

func getGlobal(stub shim.ChaincodeStubInterface, symbol string) *dm.GlobalTokenInfo {
	//
	gTkInfo := dm.GlobalTokenInfo{}
	tkInfoBytes, _ := stub.GetGlobalState(dm.GlobalPrefix + symbol)
	if len(tkInfoBytes) == 0 {
		return nil
	}
	//
	err := json.Unmarshal(tkInfoBytes, &gTkInfo)
	if err != nil {
		return nil
	}

	return &gTkInfo
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
	tkInfos := make([]TokenInfo, 0, len(KVs))
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

func checkAddr(addr string) error {
	if addr == "" {
		return nil
	}
	_, err := common.StringToAddress(addr)
	return err
}

func getSupply(supplyStr string, decimals uint64) (uint64, error) {
	//totalSupply, err := strconv.ParseUint(args[3], 10, 64)
	totalSupply, err := decimal.NewFromString(supplyStr)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to convert total supply\"}"
		return 0, fmt.Errorf(jsonResp)
	}
	if !totalSupply.IsPositive() {
		jsonResp := "{\"Error\":\"Must be positive\"}"
		return 0, fmt.Errorf(jsonResp)
	}
	totalSupply = totalSupply.Mul(decimal.New(1, int32(decimals)))
	return uint64(totalSupply.IntPart()), nil
}

func createToken(args []string, stub shim.ChaincodeStubInterface) pb.Response {
	//params check
	if len(args) < 4 {
		return shim.Error("need 4 args (Name,Symbol,Decimals,TotalSupply,[SupplyAddress])")
	}

	//==== convert params to token information
	var fungible dm.FungibleToken
	//name symbol
	if len(args[0]) > 1024 {
		jsonResp := "{\"Error\":\"Name length should not be greater than 1024\"}"
		return shim.Error(jsonResp)
	}
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
	totalSupply, err := getSupply(args[3], decimals)
	if err != nil {
		return shim.Error(err.Error())
	}
	fungible.TotalSupply = totalSupply

	//address of supply
	if len(args) > 4 {
		fungible.SupplyAddress = args[4]
		err := checkAddr(fungible.SupplyAddress)
		if err != nil {
			jsonResp := "{\"Error\":\"The SupplyAddress is invalid\"}"
			return shim.Error(jsonResp)
		}
	}

	//check name is only or not
	gTkInfo := getGlobal(stub, fungible.Symbol)
	if gTkInfo != nil {
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
		return shim.Error(jsonResp1)
	}

	//last put state
	txid := stub.GetTxID()
	assetID, _ := dm.NewAssetId(fungible.Symbol, dm.AssetType_FungibleToken,
		fungible.Decimals, common.Hex2Bytes(txid[2:]), dm.UniqueIdType_Null)
	info := TokenInfo{fungible.Symbol, createAddr.String(), totalSupply, decimals,
		fungible.SupplyAddress, assetID}

	err = setSymbols(stub, &info)
	if err != nil {
		return shim.Error(jsonResp3)
	}

	//set token define
	err = stub.DefineToken(byte(dm.AssetType_FungibleToken), createJson, createAddr.String())
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to call stub.DefineToken\"}"
		return shim.Error(jsonResp)
	}
	//add global state
	err = setGlobal(stub, &info)
	if err != nil {
		return shim.Error(jsonResp4)
	}
	return shim.Success(createJson)
}

func supplyToken(args []string, stub shim.ChaincodeStubInterface) pb.Response {
	//params check
	if len(args) < 2 {
		return shim.Error("need 2 args (Symbol,SupplyAmout)")
	}

	//symbol
	symbolAsset := args[0]
	var symbol string
	if index := strings.IndexRune(symbolAsset, '+'); index != -1 {
		asset, err := dm.StringToAsset(symbolAsset)
		if err != nil {
			jsonResp := "{\"Error\":\"Asset is \"}"
			return shim.Error(jsonResp)
		}
		symbol = asset.AssetId.GetSymbol()
	} else {
		symbol = strings.ToUpper(symbolAsset)
	}
	//check name is exist or not
	gTkInfo := getGlobal(stub, symbol)
	if gTkInfo == nil {
		return shim.Error(jsonResp2)
	}

	//check status
	if gTkInfo.Status != 0 {
		jsonResp := "{\"Error\":\"Status is frozen\"}"
		return shim.Error(jsonResp)
	}

	tkInfo := getSymbols(stub, symbol)
	if tkInfo == nil {
		return shim.Error(jsonResp2)
	}

	//supply amount
	supplyAmount, err := getSupply(args[1], uint64(tkInfo.AssetID.GetDecimal()))
	if err != nil {
		return shim.Error(err.Error())
	}
	if math.MaxInt64-tkInfo.TotalSupply < supplyAmount {
		jsonResp := "{\"Error\":\"Too big, overflow\"}"
		return shim.Error(jsonResp)
	}

	//get invoke address
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		return shim.Error(jsonResp1)
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
		return shim.Error(jsonResp3)
	}

	err = setGlobal(stub, tkInfo)
	if err != nil {
		return shim.Error(jsonResp4)
	}

	return shim.Success([]byte(""))
}

func changeSupplyAddr(args []string, stub shim.ChaincodeStubInterface) pb.Response {
	//params check
	if len(args) < 2 {
		return shim.Error("need 2 args (Symbol,NewSupplyAddr)")
	}

	//symbol
	symbol := strings.ToUpper(args[0])
	//check name is exist or not
	gTkInfo := getGlobal(stub, symbol)
	if gTkInfo == nil {
		return shim.Error(jsonResp2)
	}

	//check status
	if gTkInfo.Status != 0 {
		jsonResp := "{\"Error\":\"Status is frozen\"}"
		return shim.Error(jsonResp)
	}

	tkInfo := getSymbols(stub, symbol)
	if tkInfo == nil {
		return shim.Error(jsonResp2)
	}

	//new supply address
	newSupplyAddr := args[1]
	err := checkAddr(newSupplyAddr)
	if err != nil {
		jsonResp := "{\"Error\":\"The SupplyAddress is invalid\"}"
		return shim.Error(jsonResp)
	}

	//get invoke address
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		return shim.Error(jsonResp1)
	}
	//check supply address
	if invokeAddr.String() != tkInfo.SupplyAddr {
		jsonResp := "{\"Error\":\"Not the supply address\"}"
		return shim.Error(jsonResp)
	}

	//set supply address
	tkInfo.SupplyAddr = newSupplyAddr

	err = setSymbols(stub, tkInfo)
	if err != nil {
		return shim.Error(jsonResp3)
	}

	err = setGlobal(stub, tkInfo)
	if err != nil {
		return shim.Error(jsonResp4)
	}

	return shim.Success([]byte(""))
}

func frozenToken(args []string, stub shim.ChaincodeStubInterface) pb.Response {
	//params check
	if len(args) < 1 {
		return shim.Error("need 1 args (Symbol)")
	}

	//symbol
	symbol := strings.ToUpper(args[0])
	//check name is exist or not
	gTkInfo := getGlobal(stub, symbol)
	if gTkInfo == nil {
		return shim.Error(jsonResp2)
	}

	//get invoke address
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get invoke address\"}"
		return shim.Error(jsonResp)
	}
	//check address
	invokeAddrStr := invokeAddr.String()
	ownerAddr := gTkInfo.SupplyAddr
	if len(ownerAddr) == 0 {
		ownerAddr = gTkInfo.CreateAddr
	}
	if invokeAddrStr != ownerAddr {
		gp, err := stub.GetSystemConfig()
		if err != nil {
			jsonResp := "{\"Error\":\"GetSystemConfig() failed\"}"
			return shim.Error(jsonResp)
		}
		if invokeAddrStr != gp.ChainParameters.FoundationAddress {
			jsonResp := "{\"Error\":\"Only the FoundationAddress or Owner can frozen token\"}"
			return shim.Error(jsonResp)
		}
	}

	//set status
	gTkInfo.Status = 1
	val, err := json.Marshal(gTkInfo)
	if err != nil {
		jsonResp := "{\"Error\":\"Marshal gTkInfo failed\"}"
		return shim.Error(jsonResp)
	}
	err = stub.PutGlobalState(dm.GlobalPrefix+gTkInfo.Symbol, val)
	if err != nil {
		return shim.Error(jsonResp4)
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
		return shim.Error(jsonResp2)
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

func allToken(stub shim.ChaincodeStubInterface) pb.Response {
	tkInfos := getSymbolsAll(stub)

	tkIDs := make([]TokenIDInfo, 0, len(tkInfos))
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
