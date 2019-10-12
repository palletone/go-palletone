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
	"math"
	"strconv"
	"strings"

	"github.com/shopspring/decimal"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/contracts/shim"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	dm "github.com/palletone/go-palletone/dag/modules"
)

const symbolsKey = "symbol_"
const jsonResp1 = "{\"Error\":\"Failed to get invoke address\"}"
const jsonResp2 = "{\"Error\":\"Token not exist\"}"
const jsonResp3 = "{\"Error\":\"Failed to set symbols\"}"
const jsonResp4 = "{\"Error\":\"Failed to add global state\"}"

//PRC20 chainCode name
type PRC20 struct {
}

type tokenInfo struct {
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

//Init chainCode when deploy a instance
func (p *PRC20) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

//Invoke functions of chainCode
func (p *PRC20) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	f, args := stub.GetFunctionAndParameters()

	valid, errMsg := paramCheckValid(args)
	if !valid {
		return shim.Error(errMsg)
	}
	switch f {
	case "createToken":
		if len(args) < 4 {
			return shim.Error("need 4 args (Name,Symbol,Decimals,TotalSupply,[SupplyAddress])")
		}
		decimals, _ := strconv.ParseUint(args[2], 10, 64)
		if decimals > 18 {
			jsonResp := "{\"Error\":\"Can't big than 18\"}"
			return shim.Error(jsonResp)
		}
		//total supply
		totalSupply, err := getSupply(args[3], decimals)
		if err != nil {
			return shim.Error(err.Error())
		}
		supplyAddr := ""
		if len(args) > 4 {
			supplyAddr = args[4]
		}
		result, err := p.CreateToken(stub, args[0], args[1], int(decimals), totalSupply, supplyAddr)
		if err != nil {
			return shim.Error(err.Error())
		}
		return shim.Success(result)
	case "supplyToken":
		if len(args) < 2 {
			return shim.Error("need 2 args (Symbol,SupplyAmout)")
		}
		supplyDecimal, err := decimal.NewFromString(args[1])
		if err != nil {
			jsonResp := "{\"Error\":\"Failed to convert total supply\"}"
			return shim.Error(jsonResp)
		}
		err = p.SupplyToken(stub, args[0], supplyDecimal)
		if err != nil {
			return shim.Error(err.Error())
		}
		return shim.Success([]byte(""))
	case "getTokenInfo":
		if len(args) < 1 {
			return shim.Error("need 1 args (Symbol)")
		}
		tkIDInfo, err := p.GetTokenInfo(stub, args[0])
		if err != nil {
			return shim.Error(err.Error())
		}
		tkJSON, err := json.Marshal(tkIDInfo)
		if err != nil {
			return shim.Error(err.Error())
		}
		return shim.Success(tkJSON)
	case "getAllTokenInfo":
		tkIDInfo := p.GetAllTokenInfo(stub)
		tkJSON, err := json.Marshal(tkIDInfo)
		if err != nil {
			return shim.Error(err.Error())
		}
		return shim.Success(tkJSON)
	case "changeSupplyAddr":
		if len(args) < 2 {
			return shim.Error("need 2 args (Symbol,NewSupplyAddr)")
		}
		err := p.ChangeSupplyAddr(stub, args[0], args[1])
		if err != nil {
			return shim.Error(err.Error())
		}
		return shim.Success([]byte(""))
	case "frozenToken":
		if len(args) < 1 {
			return shim.Error("need 1 args (Symbol)")
		}
		err := p.FrozenToken(stub, args[0])
		if err != nil {
			return shim.Error(err.Error())
		}
		return shim.Success([]byte(""))
	default:
		jsonResp := "{\"Error\":\"Unknown function " + f + "\"}"
		return shim.Error(jsonResp)
	}
}

func setGlobal(stub shim.ChaincodeStubInterface, tkInfo *tokenInfo) error {
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

func setSymbols(stub shim.ChaincodeStubInterface, tkInfo *tokenInfo) error {
	val, err := json.Marshal(tkInfo)
	if err != nil {
		return err
	}
	err = stub.PutState(symbolsKey+tkInfo.Symbol, val)
	return err
}
func getSymbols(stub shim.ChaincodeStubInterface, symbol string) *tokenInfo {
	//
	tkInfo := tokenInfo{}
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

func getSymbolsAll(stub shim.ChaincodeStubInterface) []tokenInfo {
	KVs, _ := stub.GetStateByPrefix(symbolsKey)
	tkInfos := make([]tokenInfo, 0, len(KVs))
	for _, oneKV := range KVs {
		tkInfo := tokenInfo{}
		err := json.Unmarshal(oneKV.Value, &tkInfo)
		if err != nil {
			continue
		}
		tkInfos = append(tkInfos, tkInfo)
	}
	return tkInfos
}

func checkAddr(addr string) error {
	if addr == "" { //if empty, not need check
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
	//uMaxStr := fmt.Sprintf("%d", uint64(math.MaxUint64))
	uMaxDecimal, _ := decimal.NewFromString("18446744073709551615")
	if totalSupply.GreaterThan(uMaxDecimal) {
		jsonResp := "{\"Error\":\"TotalSupply * decimals is too big\"}"
		return 0, fmt.Errorf(jsonResp)
	}
	return uint64(totalSupply.IntPart()), nil
}

func getSupplydecimals(supplyDecimal decimal.Decimal, decimals uint64) (uint64, error) {
	if !supplyDecimal.IsPositive() {
		jsonResp := "{\"Error\":\"Must be positive\"}"
		return 0, fmt.Errorf(jsonResp)
	}
	supplyDecimal = supplyDecimal.Mul(decimal.New(1, int32(decimals)))
	return uint64(supplyDecimal.IntPart()), nil
}

//CreateToken create token implement
func (p *PRC20) CreateToken(stub shim.ChaincodeStubInterface, name string, symbol string, decimals int,
	totalSupply uint64, supplyAddress string) ([]byte, error) {
	//==== convert params to token information
	var fungible dm.FungibleToken
	//name symbol
	if len(name) > 1024 {
		jsonResp := "{\"Error\":\"Name length should not be greater than 1024\"}"
		return []byte{}, fmt.Errorf(jsonResp)
	}
	fungible.Name = name
	fungible.Symbol = strings.ToUpper(symbol)
	if fungible.Symbol == "PTN" {
		jsonResp := "{\"Error\":\"Can't use PTN\"}"
		return []byte{}, fmt.Errorf(jsonResp)
	}
	if len(fungible.Symbol) > 5 {
		jsonResp := "{\"Error\":\"Symbol must less than 5 characters\"}"
		return []byte{}, fmt.Errorf(jsonResp)
	}
	//decimals
	fungible.Decimals = byte(decimals)

	fungible.TotalSupply = totalSupply

	//address of supply
	if len(supplyAddress) > 0 {
		fungible.SupplyAddress = supplyAddress
		err := checkAddr(fungible.SupplyAddress)
		if err != nil {
			jsonResp := "{\"Error\":\"The SupplyAddress is invalid\"}"
			return []byte{}, fmt.Errorf(jsonResp)
		}
	}

	//check name is only or not
	gTkInfo := getGlobal(stub, fungible.Symbol)
	if gTkInfo != nil {
		jsonResp := "{\"Error\":\"The symbol have been used\"}"
		return []byte{}, fmt.Errorf(jsonResp)
	}

	//convert to json
	createJSON, err := json.Marshal(fungible)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to generate token Json\"}"
		return []byte{}, fmt.Errorf(jsonResp)
	}
	//get creator
	createAddr, err := stub.GetInvokeAddress()
	if err != nil {
		return []byte{}, fmt.Errorf(jsonResp1)
	}

	//last put state
	txid := stub.GetTxID()
	assetID, _ := dm.NewAssetId(fungible.Symbol, dm.AssetType_FungibleToken,
		fungible.Decimals, common.Hex2Bytes(txid[2:]), dm.UniqueIdType_Null)
	info := tokenInfo{fungible.Symbol, createAddr.String(), totalSupply, uint64(decimals),
		fungible.SupplyAddress, assetID}

	err = setSymbols(stub, &info)
	if err != nil {
		return []byte{}, fmt.Errorf(jsonResp3)
	}

	//set token define
	err = stub.DefineToken(byte(dm.AssetType_FungibleToken), createJSON, createAddr.String())
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to call stub.DefineToken\"}"
		return []byte{}, fmt.Errorf(jsonResp)
	}
	//add global state
	err = setGlobal(stub, &info)
	if err != nil {
		return []byte{}, fmt.Errorf(jsonResp4)
	}
	return createJSON, nil
}

//SupplyToken supply token implement
func (p *PRC20) SupplyToken(stub shim.ChaincodeStubInterface, symbol string, supplyDecimal decimal.Decimal) error {
	symbolAsset := symbol
	var symbolOnly string
	if index := strings.IndexRune(symbolAsset, '+'); index != -1 {
		asset, err := dm.StringToAsset(symbolAsset)
		if err != nil {
			jsonResp := "{\"Error\":\"Asset is \"}"
			return fmt.Errorf(jsonResp)
		}
		symbolOnly = asset.AssetId.GetSymbol()
	} else {
		symbolOnly = strings.ToUpper(symbolAsset)
	}
	//check name is exist or not
	gTkInfo := getGlobal(stub, symbolOnly)
	if gTkInfo == nil {
		return fmt.Errorf(jsonResp2)
	}

	//check status
	if gTkInfo.Status != 0 {
		jsonResp := "{\"Error\":\"Status is frozen\"}"
		return fmt.Errorf(jsonResp)
	}

	tkInfo := getSymbols(stub, symbolOnly)
	if tkInfo == nil {
		return fmt.Errorf(jsonResp2)
	}

	//supply amount
	supplyAmount, err := getSupplydecimals(supplyDecimal, uint64(tkInfo.AssetID.GetDecimal()))
	if err != nil {
		return fmt.Errorf(err.Error())
	}
	if math.MaxUint64-tkInfo.TotalSupply < supplyAmount {
		jsonResp := "{\"Error\":\"Too big, overflow\"}"
		return fmt.Errorf(jsonResp)
	}

	//get invoke address
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		return fmt.Errorf(jsonResp1)
	}
	//check supply address
	if invokeAddr.String() != tkInfo.SupplyAddr {
		jsonResp := "{\"Error\":\"Not the supply address\"}"
		return fmt.Errorf(jsonResp)
	}

	//call SupplyToken
	assetID := tkInfo.AssetID
	err = stub.SupplyToken(assetID.Bytes(),
		[]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, supplyAmount, tkInfo.SupplyAddr)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to call stub.SupplyToken\"}"
		return fmt.Errorf(jsonResp)
	}

	//add supply
	tkInfo.TotalSupply += supplyAmount

	err = setSymbols(stub, tkInfo)
	if err != nil {
		return fmt.Errorf(jsonResp3)
	}

	err = setGlobal(stub, tkInfo)
	if err != nil {
		return fmt.Errorf(jsonResp4)
	}

	return nil
}

//ChangeSupplyAddr change supply address
func (p *PRC20) ChangeSupplyAddr(stub shim.ChaincodeStubInterface, symbol string, newSupplyAddr string) error {
	//check name is exist or not
	gTkInfo := getGlobal(stub, strings.ToUpper(symbol))
	if gTkInfo == nil {
		return fmt.Errorf(jsonResp2)
	}

	//check status
	if gTkInfo.Status != 0 {
		jsonResp := "{\"Error\":\"Status is frozen\"}"
		return fmt.Errorf(jsonResp)
	}

	tkInfo := getSymbols(stub, symbol)
	if tkInfo == nil {
		return fmt.Errorf(jsonResp2)
	}

	//new supply address
	err := checkAddr(newSupplyAddr)
	if err != nil {
		jsonResp := "{\"Error\":\"The SupplyAddress is invalid\"}"
		return fmt.Errorf(jsonResp)
	}

	//get invoke address
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		return fmt.Errorf(jsonResp1)
	}
	//check supply address
	if invokeAddr.String() != tkInfo.SupplyAddr {
		jsonResp := "{\"Error\":\"Not the supply address\"}"
		return fmt.Errorf(jsonResp)
	}

	//set supply address
	tkInfo.SupplyAddr = newSupplyAddr

	err = setSymbols(stub, tkInfo)
	if err != nil {
		return fmt.Errorf(jsonResp3)
	}

	err = setGlobal(stub, tkInfo)
	if err != nil {
		return fmt.Errorf(jsonResp4)
	}

	return nil
}

//FrozenToken frozen one token
func (p *PRC20) FrozenToken(stub shim.ChaincodeStubInterface, symbol string) error {
	//check name is exist or not
	gTkInfo := getGlobal(stub, strings.ToUpper(symbol))
	if gTkInfo == nil {
		return fmt.Errorf(jsonResp2)
	}

	//get invoke address
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get invoke address\"}"
		return fmt.Errorf(jsonResp)
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
			return fmt.Errorf(jsonResp)
		}
		if invokeAddrStr != gp.ChainParameters.FoundationAddress {
			jsonResp := "{\"Error\":\"Only the FoundationAddress or Owner can frozen token\"}"
			return fmt.Errorf(jsonResp)
		}
	}

	//set status
	gTkInfo.Status = 1
	val, err := json.Marshal(gTkInfo)
	if err != nil {
		jsonResp := "{\"Error\":\"Marshal gTkInfo failed\"}"
		return fmt.Errorf(jsonResp)
	}
	err = stub.PutGlobalState(dm.GlobalPrefix+gTkInfo.Symbol, val)
	if err != nil {
		return fmt.Errorf(jsonResp4)
	}
	return nil
}

type tokenIDInfo struct {
	Symbol      string
	CreateAddr  string
	TotalSupply uint64
	Decimals    uint64
	SupplyAddr  string
	AssetID     string
}

//GetTokenInfo get one token information
func (p *PRC20) GetTokenInfo(stub shim.ChaincodeStubInterface, symbol string) (*tokenIDInfo, error) {
	//check name is exist or not
	symbol = strings.ToUpper(symbol)
	tkInfo := getSymbols(stub, symbol)
	if tkInfo == nil {
		return nil, fmt.Errorf(jsonResp2)
	}

	//token
	asset := tkInfo.AssetID
	tkID := tokenIDInfo{symbol, tkInfo.CreateAddr, tkInfo.TotalSupply,
		tkInfo.Decimals, tkInfo.SupplyAddr, asset.String()}
	return &tkID, nil
}

//GetAllTokenInfo get all token information
func (p *PRC20) GetAllTokenInfo(stub shim.ChaincodeStubInterface) []tokenIDInfo {
	tkInfos := getSymbolsAll(stub)

	tkIDs := make([]tokenIDInfo, 0, len(tkInfos))
	for _, tkInfo := range tkInfos {
		asset := tkInfo.AssetID
		tkID := tokenIDInfo{tkInfo.Symbol, tkInfo.CreateAddr, tkInfo.TotalSupply,
			tkInfo.Decimals, tkInfo.SupplyAddr, asset.String()}
		tkIDs = append(tkIDs, tkID)
	}

	//return json
	return tkIDs
}
