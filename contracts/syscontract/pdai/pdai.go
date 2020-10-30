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

package pdai

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"

	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/contracts/shim"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"github.com/palletone/go-palletone/dag/errors"
	dm "github.com/palletone/go-palletone/dag/modules"
	"github.com/shopspring/decimal"
)

type Dai struct {
}

var Dai_decimal = 0
var Dai_symbol = "PDAI"

const symbolsCDP = "cdp_"
const symbolsBalance = "bal_"

func (p *Dai) Init(stub shim.ChaincodeStubInterface) pb.Response {
	//invokeAddr, err := stub.GetInvokeAddress()
	//if err != nil {
	//	jsonResp := "{\"Error\":\"Failed to get invoke address\"}"
	//	return shim.Error(jsonResp)
	//}
	err := initGlobal(stub)
	if err != nil {
		return shim.Error("setGlobal failed: " + err.Error())
	}
	//err = stub.PutState(symbolsAdmin, []byte(invokeAddr.String()))
	//if err != nil {
	//	return shim.Error("write symbolsAdmin failed: " + err.Error())
	//}
	return shim.Success(nil)
}

func (p *Dai) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	f, args := stub.GetFunctionAndParameters()

	switch f {
	case "initDai":
		return p.InitDai(stub)
	case "createCDP":
		if len(args) < 2 {
			return shim.Error("need 2 args (wantPDAIAmount, ptnPrice)")
		}
		if len(args[0]) == 0 {
			return shim.Error("wantPDAIAmount is empty")
		}
		if len(args[1]) == 0 {
			return shim.Error("ptnPrice is empty")
		}
		amount, err := getAmount(args[0], 0)
		if err != nil {
			return shim.Error(err.Error())
		}
		return p.CreateCDP(stub, amount, args[1])
	case "deleteCDP":
		if len(args) < 1 {
			return shim.Error("need 1 args (createCDPReqID)")
		}
		if len(args[0]) == 0 {
			return shim.Error("createCDPReqID is empty")
		}
		return p.DeleteCDP(stub, args[0])

	case "savePDai":
		return p.SavePDai(stub)

	case "getAdmin":
		return p.Get(stub, symbolsAdmin)
	case "setAdmin":
		if len(args) < 1 {
			return shim.Error("need 1 args (PTNAddr)")
		}
		return p.SetAdmin(args[0], stub)
	case "Set":
		if len(args) < 2 {
			return shim.Error("need 2 args (Key, Value)")
		}
		return p.Set(stub, args[0], args[1])
	case "get":
		if len(args) < 1 {
			return shim.Error("need 1 args (Key)")
		}
		return p.Get(stub, args[0])

	default:
		jsonResp := "{\"Error\":\"Unknown function " + f + "\"}"
		return shim.Error(jsonResp)
	}
}

func getAmount(amountStr string, decimals uint64) (uint64, error) {
	amount, err := decimal.NewFromString(amountStr)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to convert amount\"}"
		return 0, fmt.Errorf(jsonResp)
	}
	if !amount.IsPositive() {
		jsonResp := "{\"Error\":\"amount must be positive\"}"
		return 0, fmt.Errorf(jsonResp)
	}
	amount = amount.Mul(decimal.New(1, int32(decimals)))
	uMaxDecimal, _ := decimal.NewFromString("18446744073709551615")
	if amount.GreaterThan(uMaxDecimal) {
		jsonResp := "{\"Error\":\"TotalSupply * decimals is too big\"}"
		return 0, fmt.Errorf(jsonResp)
	}
	return uint64(amount.IntPart()), nil
}

type GlobalTokenInfo struct {
	Symbol      string
	TokenType   uint8 //1:prc20 2:prc721 3:vote 4:SysVote
	Status      uint8
	CreateAddr  string
	TotalSupply uint64
	SupplyAddr  string
	AssetID     dm.AssetId
}

func initGlobal(stub shim.ChaincodeStubInterface) error {
	assetID, _ := dm.NewAssetId(Dai_symbol, dm.AssetType_FungibleToken,
		byte(Dai_decimal), []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		dm.UniqueIdType_Null)
	gTkInfo := GlobalTokenInfo{Symbol: Dai_symbol, TokenType: 1, Status: 0, CreateAddr: "",
		TotalSupply: 0, SupplyAddr: "", AssetID: assetID}
	val, err := json.Marshal(gTkInfo)
	if err != nil {
		return err
	}
	err = stub.PutGlobalState(dm.GlobalPrefix+gTkInfo.Symbol, val)
	return err
}

func setGlobal(stub shim.ChaincodeStubInterface, gTkInfo *GlobalTokenInfo) error {
	val, err := json.Marshal(gTkInfo)
	if err != nil {
		return err
	}
	err = stub.PutGlobalState(dm.GlobalPrefix+gTkInfo.Symbol, val)
	return err
}

func getGlobal(stub shim.ChaincodeStubInterface, symbol string) *GlobalTokenInfo {
	//
	gTkInfo := GlobalTokenInfo{}
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

const symbolsAdmin = "admin_"

func (p *Dai) InitDai(stub shim.ChaincodeStubInterface) pb.Response {
	err := initGlobal(stub)
	if err != nil {
		return shim.Error("setGlobal failed: " + err.Error())
	}
	return shim.Success(nil)
}

func (p *Dai) CreateCDP(stub shim.ChaincodeStubInterface, wantAmount uint64, ptnPrice string) pb.Response {
	reqID := stub.GetTxID()
	if "0x" == reqID[0:2] || "0X" == reqID[0:2] {
		reqID = reqID[2:]
	}

	invokeTokens, err := stub.GetInvokeTokens()
	if err != nil {
		jsonResp := "{\"Error\":\"GetInvokeTokens failed\"}"
		return shim.Error(jsonResp)
	}
	ptnAmount := uint64(0)
	for i := 0; i < len(invokeTokens); i++ {
		if invokeTokens[i].Asset.AssetId == dm.PTNCOIN {
			ptnAmount += invokeTokens[i].Amount
			break
		}
	}
	if ptnAmount == 0 { //no ptn
		jsonResp := "{\"Error\":\"PTN amount is zero\"}"
		return shim.Error(jsonResp)
	}

	gTkInfo := getGlobal(stub, Dai_symbol)
	if gTkInfo == nil {
		return shim.Error("PDAI not exist")
	}

	ptnValue, _ := decimal.NewFromString(ptnPrice)
	pntAsset := dm.NewPTNAsset()
	ptnAmountAsset := pntAsset.DisplayAmount(ptnAmount)
	ptnValue = ptnValue.Mul(ptnAmountAsset)

	oneDaiValue, _ := decimal.NewFromString("0.01")
	percent, _ := decimal.NewFromString("1.5")
	percentAmount := ptnValue.Div(oneDaiValue).Div(percent)
	pdaiAsset := gTkInfo.AssetID.ToAsset()
	wantAssetAmount := pdaiAsset.DisplayAmount(wantAmount)
	log.Debugf("wantAssetAmount is %s,percentAmount is %s",
		wantAssetAmount.String(), percentAmount.String())
	if wantAssetAmount.Cmp(percentAmount) > 0 {
		return shim.Error("mortgage must bigger than 150%")
	}

	assetAmount := pdaiAsset.Uint64Amount(wantAssetAmount)
	if math.MaxUint64-gTkInfo.TotalSupply < assetAmount {
		return shim.Error("PDAI totalSupply is too big, overflow")
	}

	//get invoke address
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		return shim.Error("Failed to get invoke address")
	}

	//call SupplyToken
	assetID := gTkInfo.AssetID
	err = stub.SupplyToken(assetID.Bytes(),
		[]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		assetAmount, invokeAddr.String())
	if err != nil {
		return shim.Error("Failed to call stub.SupplyToken")
	}

	//add supply
	gTkInfo.TotalSupply += assetAmount

	err = setGlobal(stub, gTkInfo)
	if err != nil {
		return shim.Error("Failed to add global state")
	}

	err = stub.PutState(symbolsCDP+reqID, []byte(fmt.Sprintf("%d", assetAmount)))
	if err != nil {
		return shim.Error("Failed to PutState CDP" + err.Error())
	}

	return shim.Success([]byte("Success"))
}

func (p *Dai) DeleteCDP(stub shim.ChaincodeStubInterface, reqID string) pb.Response {
	if "0x" == reqID[0:2] || "0X" == reqID[0:2] {
		reqID = reqID[2:]
	}

	result, _ := stub.GetState(symbolsCDP + reqID)
	if len(result) == 0 {
		return shim.Error("get CDP by reqID failed")
	}
	daiNum, _ := strconv.ParseUint(string(result), 10, 64)

	//get invoke address
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		return shim.Error("Failed to get invoke address")
	}

	amount := uint64(0) //todo
	resultBal, _ := stub.GetState(symbolsBalance + invokeAddr.String())
	if len(result) != 0 {
		amount, _ = strconv.ParseUint(string(resultBal), 10, 64)
	}
	if amount < daiNum {
		return shim.Error(fmt.Sprintf("amount of CDP is not match, amount is %d", amount))
	}
	amount -= daiNum
	err = stub.PutState(symbolsBalance+invokeAddr.String(), []byte(fmt.Sprintf("%d", amount)))
	if err != nil {
		return shim.Error("Failed to PutState balance" + err.Error())
	}
	err = stub.DelState(symbolsCDP + reqID)
	if err != nil {
		return shim.Error("Failed to DelState CDP" + err.Error())
	}
	assetID, _ := dm.NewAssetId(Dai_symbol, dm.AssetType_FungibleToken,
		byte(Dai_decimal), []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		dm.UniqueIdType_Null)

	gTkInfo := getGlobal(stub, Dai_symbol)
	if gTkInfo == nil {
		return shim.Error("PDAI not exist")
	}
	//sub supply
	gTkInfo.TotalSupply -= daiNum

	err = setGlobal(stub, gTkInfo)
	if err != nil {
		return shim.Error("Failed to add global state")
	}

	amountAsset := &dm.AmountAsset{
		Amount: daiNum,
		Asset:  assetID.ToAsset(),
	}
	err = stub.PayOutToken("P1111111111111111111114oLvT2", amountAsset, 0)
	if err != nil {
		return shim.Error("Failed to call stub.PayOutToken")
	}

	return shim.Success([]byte("Success"))
}

func (p *Dai) SavePDai(stub shim.ChaincodeStubInterface) pb.Response {
	invokeTokens, err := stub.GetInvokeTokens()
	if err != nil {
		jsonResp := "{\"Error\":\"GetInvokeTokens failed\"}"
		return shim.Error(jsonResp)
	}

	daiNum := uint64(0)
	assetID, _ := dm.NewAssetId(Dai_symbol, dm.AssetType_FungibleToken,
		byte(Dai_decimal), []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		dm.UniqueIdType_Null)

	for i := 0; i < len(invokeTokens); i++ {
		if invokeTokens[i].Address == "PCGTta3M4t3yXu8uRgkKvaWd2d8DSRr1tUD" &&
			invokeTokens[i].Asset.AssetId == assetID {
			daiNum += invokeTokens[i].Amount
			break
		}
	}
	if daiNum == 0 { //no vote token
		jsonResp := "{\"Error\":\"PDAI amount is zero\"}"
		return shim.Error(jsonResp)
	}

	//get invoke address
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		return shim.Error("Failed to get invoke address")
	}

	amount := uint64(0) //todo
	result, _ := stub.GetState(symbolsBalance + invokeAddr.String())
	if len(result) != 0 {
		amount, _ = strconv.ParseUint(string(result), 10, 64)
	}
	if amount+daiNum > math.MaxUint64 { //todo
		return shim.Error(fmt.Sprintf("balannce is overflow, amount is %d", amount))
	}
	amount += daiNum

	err = stub.PutState(symbolsBalance+invokeAddr.String(), []byte(fmt.Sprintf("%d", amount)))
	if err != nil {
		return shim.Error("Failed to PutState balance" + err.Error())
	}

	return shim.Success([]byte("Success"))
}

func (p *Dai) SetAdmin(ptnAddr string, stub shim.ChaincodeStubInterface) pb.Response {
	//only admin can set
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get invoke address\"}"
		return shim.Error(jsonResp)
	}
	admin, err := getAdmin(stub)
	if err != nil {
		return shim.Error(err.Error())
	}
	if admin != invokeAddr.String() {
		return shim.Error("Only admin can set")
	}
	err = stub.PutState(symbolsAdmin, []byte(ptnAddr))
	if err != nil {
		return shim.Error("write symbolsAdmin failed: " + err.Error())
	}

	return shim.Success([]byte("Success"))
}

func getAdmin(stub shim.ChaincodeStubInterface) (string, error) {
	result, _ := stub.GetState(symbolsAdmin)
	if len(result) == 0 {
		return "", errors.New("Need set Owner")
	}

	return string(result), nil
}

func (p *Dai) Get(stub shim.ChaincodeStubInterface, key string) pb.Response {
	result, _ := stub.GetState(key)
	return shim.Success(result)
}
func (p *Dai) Set(stub shim.ChaincodeStubInterface, key string, value string) pb.Response {
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get invoke address\"}"
		return shim.Error(jsonResp)
	}
	admin, err := getAdmin(stub)
	if err != nil {
		return shim.Error(err.Error())
	}
	if admin != invokeAddr.String() {
		return shim.Error("Only admin can set")
	}

	err = stub.PutState(key, []byte(value))
	if err != nil {
		return shim.Error(fmt.Sprintf("PutState failed: %s", err.Error()))
	}
	return shim.Success([]byte("Success"))
}
