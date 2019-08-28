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

package prc721

import (
	"crypto/rand"
	"encoding/json"
	"io"
	"sort"
	"strconv"
	"strings"

	"encoding/binary"
	"encoding/hex"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/math"
	"github.com/palletone/go-palletone/contracts/shim"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	dm "github.com/palletone/go-palletone/dag/modules"
)

const symbolsKey = "symbol_"
const jsonResp1 = "{\"Error\":\"Failed to add global state\"}"
const jsonResp2 = "{\"Error\":\"Failed to set symbols\"}"
const jsonResp3 = "{\"Error\":\"Token not exist\"}"
const jsonResp4 = "{\"Error\":\"Asset_TokenID invalid\"}"
const jsonResp5 = "{\"Error\":\"Failed to get invoke address\"}"

type PRC721 struct {
}

type TokenInfo struct {
	Symbol      string
	TokenType   uint8
	TokenMax    uint64 //only use when TokenType is Sequence
	CreateAddr  string
	TotalSupply uint64
	SupplyAddr  string
	AssetID     dm.AssetId
}

type Symbols struct {
	TokenInfos map[string]TokenInfo `json:"tokeninfos"`
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
func (p *PRC721) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

func (p *PRC721) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
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
	case "existTokenID":
		return existTokenID(args, stub)
	case "setTokenURI":
		return setTokenURI(args, stub)
	case "getTokenURI":
		return getTokenURI(args, stub)
	case "getTokenInfo":
		return oneToken(args, stub)
	case "getAllTokenInfo":
		return allToken(stub)
	case "changeSupplyAddr":
		return changeSupplyAddr(args, stub)
	default:
		jsonResp := "{\"Error\":\"Unknown function " + f + "\"}"
		return shim.Error(jsonResp)
	}
}

func setGlobal(stub shim.ChaincodeStubInterface, tkInfo *TokenInfo) error {
	gTkInfo := dm.GlobalTokenInfo{Symbol: tkInfo.Symbol, TokenType: 2, Status: 0, CreateAddr: tkInfo.CreateAddr,
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

func convertToByte(n uint64) []byte {
	by8 := make([]byte, 8)
	binary.BigEndian.PutUint64(by8, n)
	by16 := make([]byte, 16)
	copy(by16[16-len(by8):], by8)
	return by16
}

func generateUUID() ([]byte, error) {
	uuid := make([]byte, 16)
	_, err := io.ReadFull(rand.Reader, uuid)
	if err != nil {
		return nil, err
	}
	// variant bits; see section 4.1.1
	uuid[8] = uuid[8]&^0xc0 | 0x80
	// version 4 (pseudo-random); see section 4.1.3
	uuid[6] = uuid[6]&^0xf0 | 0x40
	return uuid, nil
}

type TokenIDMeta struct {
	TokenID  string
	MetaData string
}

func chekcTokenIDRepeat(tokenIDMetas []TokenIDMeta) bool {
	tokenIDs := make(map[string]bool)
	for _, oneTokenMeta := range tokenIDMetas {
		if _, ok := tokenIDs[oneTokenMeta.TokenID]; ok {
			return true
		}
		tokenIDs[oneTokenMeta.TokenID] = true
	}
	return false
}

func genNFData(idType dm.UniqueIdType, totalSupply uint64, start uint64, tokenIDMetas []TokenIDMeta) ([]dm.NonFungibleMetaData, string) {
	nfDatas := []dm.NonFungibleMetaData{}
	if idType == dm.UniqueIdType_Sequence {
		for i := uint64(0); i < totalSupply; i++ {
			seqByte := convertToByte(start + i)
			nFdata := dm.NonFungibleMetaData{UniqueBytes: seqByte}
			nfDatas = append(nfDatas, nFdata)
		}
	} else if idType == dm.UniqueIdType_Uuid {
		for i := uint64(0); i < totalSupply; i++ {
			UUID, _ := generateUUID()
			if len(UUID) < 16 {
				jsonResp := "{\"Error\":\"generateUUID() failed\"}"
				return nil, jsonResp
			}
			nFdata := dm.NonFungibleMetaData{UniqueBytes: UUID}
			nfDatas = append(nfDatas, nFdata)
		}
	} else if idType == dm.UniqueIdType_UserDefine {
		for _, oneTokenMeta := range tokenIDMetas {
			oneTokenIDByte, _ := hex.DecodeString(oneTokenMeta.TokenID)
			if len(oneTokenIDByte) != 16 {
				jsonResp := "{\"Error\":\"tokenIDMetas format invalid, must be 32 len hex string\"}"
				return nil, jsonResp
			}
			nFdata := dm.NonFungibleMetaData{UniqueBytes: oneTokenIDByte}
			nfDatas = append(nfDatas, nFdata)
		}
	} else if idType == dm.UniqueIdType_Ascii {
		for _, oneTokenMeta := range tokenIDMetas {
			if len(oneTokenMeta.TokenID) != 16 {
				jsonResp := "{\"Error\":\"tokenIDMetas format invalid, len must be 16 ascii string\"}"
				return nil, jsonResp
			}
			nFdata := dm.NonFungibleMetaData{UniqueBytes: []byte(oneTokenMeta.TokenID)}
			nfDatas = append(nfDatas, nFdata)
		}
	}
	return nfDatas, ""
}

func checkAddr(addr string) error {
	if addr == "" {
		return nil
	}
	_, err := common.StringToAddress(addr)
	return err
}

func createToken(args []string, stub shim.ChaincodeStubInterface) pb.Response {
	//params check
	if len(args) < 5 {
		return shim.Error("need 5 args (Name,Symbol,Type,TotalSupply,TokenIDMetas,[SupplyAddress])")
	}

	//==== convert params to token information
	var nonFungible dm.NonFungibleToken
	//name symbol
	if len(args[0]) > 1024 {
		jsonResp := "{\"Error\":\"Name length should not be greater than 1024\"}"
		return shim.Error(jsonResp)
	}
	nonFungible.Name = args[0]
	nonFungible.Symbol = strings.ToUpper(args[1])
	if nonFungible.Symbol == "PTN" {
		jsonResp := "{\"Error\":\"Can't use PTN\"}"
		return shim.Error(jsonResp)
	}
	if len(nonFungible.Symbol) > 5 {
		jsonResp := "{\"Error\":\"Symbol must less than 5 characters\"}"
		return shim.Error(jsonResp)
	}
	//type
	var idType dm.UniqueIdType
	if args[2] == "1" {
		idType = dm.UniqueIdType_Sequence
	} else if args[2] == "2" {
		idType = dm.UniqueIdType_Uuid
	} else if args[2] == "3" {
		idType = dm.UniqueIdType_UserDefine
	} else if args[2] == "4" {
		idType = dm.UniqueIdType_Ascii
	} else {
		jsonResp := "{\"Error\":\"Only string, 1(Seqence) or 2(UUID) or 3(Custom) or 4(Assii)\"}"
		return shim.Error(jsonResp)
	}
	nonFungible.Type = byte(idType)

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
	if totalSupply > 1000 {
		jsonResp := "{\"Error\":\"Not allow bigger than 1000 NonFungibleToken when create\"}"
		return shim.Error(jsonResp)
	}
	nonFungible.TotalSupply = totalSupply
	//tokenIDMetas
	var tokenIDMetas []TokenIDMeta
	err = json.Unmarshal([]byte(args[4]), &tokenIDMetas)
	if err != nil {
		jsonResp := "{\"Error\":\"tokenIDMetas format invalid, must be hex strings\"}"
		return shim.Error(jsonResp)
	}
	if uint64(len(tokenIDMetas)) != totalSupply {
		jsonResp := "{\"Error\":\"tokenIDMetas and totalSupply is not match\"}"
		return shim.Error(jsonResp)
	}
	if idType == dm.UniqueIdType_UserDefine && chekcTokenIDRepeat(tokenIDMetas) {
		jsonResp := "{\"Error\":\"tokenIDMetas have repeat tokenID\"}"
		return shim.Error(jsonResp)
	}
	//address of supply
	if len(args) > 5 {
		nonFungible.SupplyAddress = args[5]
		err := checkAddr(nonFungible.SupplyAddress)
		if err != nil {
			jsonResp := "{\"Error\":\"The SupplyAddress is invalid\"}"
			return shim.Error(jsonResp)
		}
	}

	//check name is only or not
	tkInfo := getSymbols(stub, nonFungible.Symbol)
	if tkInfo != nil {
		jsonResp := "{\"Error\":\"The symbol have been used\"}"
		return shim.Error(jsonResp)
	}

	//generate nonFungibleData
	nFdatas, errStr := genNFData(idType, totalSupply, 1, tokenIDMetas)
	if errStr != "" {
		return shim.Error(errStr)
	}
	nonFungible.NonFungibleData = nFdatas

	//convert to json
	createJson, err := json.Marshal(nonFungible)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to generate token Json\"}"
		return shim.Error(jsonResp)
	}
	//get creator
	createAddr, err := stub.GetInvokeAddress()
	if err != nil {
		return shim.Error(jsonResp5)
	}

	//last put state
	txid := stub.GetTxID()
	assetID, _ := dm.NewAssetId(nonFungible.Symbol, dm.AssetType_NonFungibleToken,
		0, common.Hex2Bytes(txid[2:]), idType)

	//
	newAsset := &dm.Asset{}
	newAsset.AssetId = assetID
	for i, nFdata := range nonFungible.NonFungibleData {
		newAsset.UniqueId.SetBytes(nFdata.UniqueBytes)
		key := newAsset.String()
		err = stub.PutState(key, []byte(tokenIDMetas[i].MetaData))
		if err != nil {
			jsonResp := "{\"Error\":\"Failed to set Asset\"}"
			return shim.Error(jsonResp)
		}
	}
	info := TokenInfo{nonFungible.Symbol, byte(idType), totalSupply, createAddr.String(), totalSupply,
		nonFungible.SupplyAddress, assetID}
	err = setSymbols(stub, &info)
	if err != nil {
		return shim.Error(jsonResp2)
	}
	//set token define
	err = stub.DefineToken(byte(dm.AssetType_NonFungibleToken), createJson, createAddr.String())
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to call stub.DefineToken\"}"
		return shim.Error(jsonResp)
	}
	//add global state
	err = setGlobal(stub, &info)
	if err != nil {
		return shim.Error(jsonResp1)
	}
	return shim.Success(createJson)
}

func supplyToken(args []string, stub shim.ChaincodeStubInterface) pb.Response {
	//params check
	if len(args) < 3 {
		return shim.Error("need 2 args (Symbol,SupplyAmout,TokenIDMetas])")
	}

	//symbol
	symbol := strings.ToUpper(args[0])
	//check name is exist or not
	gTkInfo := getGlobal(stub, symbol)
	if gTkInfo == nil {

		return shim.Error(jsonResp3)
	}

	//check status
	if gTkInfo.Status != 0 {
		jsonResp := "{\"Error\":\"Status is frozen\"}"
		return shim.Error(jsonResp)
	}

	tkInfo := getSymbols(stub, symbol)
	if tkInfo == nil {
		jsonResp := "{\"Error\":\"Token not exist in contract\"}"
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
	if supplyAmount > 1000 {
		jsonResp := "{\"Error\":\"Not allow bigger than 1000 NonFungibleToken when create\"}"
		return shim.Error(jsonResp)
	}
	if math.MaxInt64-tkInfo.TotalSupply < supplyAmount {
		jsonResp := "{\"Error\":\"Too big, overflow\"}"
		return shim.Error(jsonResp)
	}

	//tokenIDMetas
	var tokenIDMetas []TokenIDMeta
	err = json.Unmarshal([]byte(args[2]), &tokenIDMetas)
	if err != nil {
		jsonResp := "{\"Error\":\"tokenIDMetas format invalid, must be hex strings\"}"
		return shim.Error(jsonResp)
	}
	if uint64(len(tokenIDMetas)) != supplyAmount {
		jsonResp := "{\"Error\":\"tokenIDMetas and supplyAmount is not match\"}"
		return shim.Error(jsonResp)
	}
	idType := dm.UniqueIdType(tkInfo.TokenType)
	if idType == dm.UniqueIdType_UserDefine && chekcTokenIDRepeat(tokenIDMetas) {
		jsonResp := "{\"Error\":\"tokenIDMetas have repeat tokenID\"}"
		return shim.Error(jsonResp)
	}

	//get invoke address
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		return shim.Error(jsonResp5)
	}
	//check supply address
	if invokeAddr.String() != tkInfo.SupplyAddr && tkInfo.SupplyAddr != "" {
		jsonResp := "{\"Error\":\"Not the supply address\"}"
		return shim.Error(jsonResp)
	}

	//call SupplyToken
	nFdatas, errStr := genNFData(idType, supplyAmount, tkInfo.TokenMax+1, tokenIDMetas)
	if errStr != "" {
		return shim.Error(errStr)
	}

	//
	newAsset := &dm.Asset{}
	newAsset.AssetId = tkInfo.AssetID
	for _, nFdata := range nFdatas {
		newAsset.UniqueId.SetBytes(nFdata.UniqueBytes)
		key := newAsset.String()
		valBytes, _ := stub.GetState(key)
		if len(valBytes) != 0 {
			jsonResp := "{\"Error\":\"Token's tokenID has exist\"}"
			return shim.Error(jsonResp)
		}
	}
	for i, nFdata := range nFdatas {
		newAsset.UniqueId.SetBytes(nFdata.UniqueBytes)
		key := newAsset.String()
		err = stub.PutState(key, []byte(tokenIDMetas[i].MetaData))
		if err != nil {
			jsonResp := "{\"Error\":\"Failed to set Asset\"}"
			return shim.Error(jsonResp)
		}
	}
	//add supply
	tkInfo.TotalSupply += supplyAmount
	if idType == dm.UniqueIdType_Sequence {
		tkInfo.TokenMax += supplyAmount
	}
	err = setSymbols(stub, tkInfo)
	if err != nil {
		return shim.Error(jsonResp2)
	}

	for _, nFdata := range nFdatas {
		err = stub.SupplyToken(tkInfo.AssetID.Bytes(), nFdata.UniqueBytes, 1, invokeAddr.String())
		if err != nil {
			jsonResp := "{\"Error\":\"Failed to call stub.SupplyToken\"}"
			return shim.Error(jsonResp)
		}
	}

	err = setGlobal(stub, tkInfo)
	if err != nil {
		return shim.Error(jsonResp1)
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
		return shim.Error(jsonResp3)
	}

	//check status
	if gTkInfo.Status != 0 {
		jsonResp := "{\"Error\":\"Status is frozen\"}"
		return shim.Error(jsonResp)
	}

	tkInfo := getSymbols(stub, symbol)
	if tkInfo == nil {
		return shim.Error(jsonResp3)
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
		jsonResp := "{\"Error\":\"Failed to get invoke address\"}"
		return shim.Error(jsonResp)
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
		return shim.Error(jsonResp2)
	}

	err = setGlobal(stub, tkInfo)
	if err != nil {
		return shim.Error(jsonResp1)
	}

	return shim.Success([]byte(""))
}

type TokenIDInfo struct {
	Symbol      string
	CreateAddr  string
	TokenType   uint8 //no
	TotalSupply uint64
	SupplyAddr  string
	AssetID     string
	TokenIDs    []string //no
}

func existTokenID(args []string, stub shim.ChaincodeStubInterface) pb.Response {
	//params check
	if len(args) < 1 {
		return shim.Error("need 1 args (Asset_TokenID)")
	}

	//asset
	assetStr := args[0]
	asset := &dm.Asset{}
	err := asset.SetString(assetStr)
	if err != nil {
		return shim.Error(jsonResp4)
	}
	//
	valBytes, _ := stub.GetState(assetStr)
	if len(valBytes) == 0 {
		return shim.Success([]byte("False"))
	}
	return shim.Success([]byte("True"))
}

func setTokenURI(args []string, stub shim.ChaincodeStubInterface) pb.Response {
	//params check
	if len(args) < 2 {
		return shim.Error("need 1 args (Asset_TokenID,TokenURI)")
	}

	//asset
	assetStr := args[0]
	tokenURI := args[1]
	asset := &dm.Asset{}
	err := asset.SetString(assetStr)
	if err != nil {
		return shim.Error(jsonResp4)
	}
	//get invoke address
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		return shim.Error(jsonResp5)
	}
	tokens, _ := stub.GetTokenBalance(invokeAddr.String(), asset)
	if len(tokens) == 0 {
		jsonResp := "{\"Error\":\"Failed to get the balance of invoke address\"}"
		return shim.Error(jsonResp)
	}

	//
	valBytes, _ := stub.GetState(assetStr)
	if len(valBytes) == 0 {
		jsonResp := "{\"Error\":\"No this tokenID\"}"
		return shim.Error(jsonResp)
	}

	err = stub.PutState(assetStr, []byte(tokenURI))
	if err != nil {
		return shim.Error("Failed to set tokenURI")
	}

	return shim.Success([]byte("True"))
}

func getTokenURI(args []string, stub shim.ChaincodeStubInterface) pb.Response {
	//params check
	if len(args) < 1 {
		return shim.Error("need 1 args (Asset_TokenID)")
	}

	//asset
	assetStr := args[0]
	asset := &dm.Asset{}
	err := asset.SetString(assetStr)
	if err != nil {
		return shim.Error(jsonResp4)
	}
	//
	valBytes, _ := stub.GetState(assetStr)
	if len(valBytes) == 0 {
		jsonResp := "{\"Error\":\"No this tokenID\"}"
		return shim.Error(jsonResp)
	}
	//
	return shim.Success(valBytes)
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
		return shim.Error(jsonResp3)
	}

	var tkIDs []string
	KVs, _ := stub.GetStateByPrefix(tkInfo.AssetID.String())
	for _, oneKV := range KVs {
		assetTkID := strings.SplitN(oneKV.Key, "-", 2)
		if len(assetTkID) == 2 {
			tkIDs = append(tkIDs, assetTkID[1])
		}
	}
	sort.Strings(tkIDs)

	//
	tkIDInfo := TokenIDInfo{symbol, tkInfo.CreateAddr, tkInfo.TokenType,
		tkInfo.TotalSupply, tkInfo.SupplyAddr, tkInfo.AssetID.String(), tkIDs}
	//return json
	tkJson, err := json.Marshal(tkIDInfo)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(tkJson)
}

func allToken(stub shim.ChaincodeStubInterface) pb.Response {
	tkInfos := getSymbolsAll(stub)
	tkIDInfos := make([]TokenIDInfo, 0, len(tkInfos))
	tkIDs := []string{"Only return simple information"}
	for _, tkInfo := range tkInfos {
		tkIDInfo := TokenIDInfo{tkInfo.Symbol, tkInfo.CreateAddr,
			tkInfo.TokenType, tkInfo.TotalSupply,
			tkInfo.SupplyAddr, tkInfo.AssetID.String(), tkIDs}
		tkIDInfos = append(tkIDInfos, tkIDInfo)
	}

	//return json
	tksJson, err := json.Marshal(tkIDInfos)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(tksJson)
}
