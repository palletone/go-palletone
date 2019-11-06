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
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/crypto"
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

//PRC721 chainCode name
type PRC721 struct {
}

type tokenInfo struct {
	Symbol      string
	TokenType   uint8
	TokenMax    uint64 //only use when TokenType is Sequence
	CreateAddr  string
	TotalSupply uint64
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

//Init
func (p *PRC721) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

//Invoke
func (p *PRC721) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	f, args := stub.GetFunctionAndParameters()

	valid, errMsg := paramCheckValid(args)
	if !valid {
		return shim.Error(errMsg)
	}
	switch f {
	case "createToken":
		if len(args) < 5 {
			return shim.Error("need 5 args (Name,Symbol,Type,TotalSupply,TokenIDMetas,[SupplyAddress])")
		}
		//total supply
		totalSupply, err := getSupply(args[3])
		if err != nil {
			return shim.Error(err.Error())
		}
		supplyAddr := ""
		if len(args) > 5 {
			supplyAddr = args[5]
		}
		return p.CreateToken(stub, args[0], args[1], args[2], totalSupply, args[4], supplyAddr)
	case "supplyToken":
		if len(args) < 3 {
			return shim.Error("need 2 args (Symbol,SupplyAmount,[TokenIDMetas])")
		}
		//total supply
		supplyAmount, err := getSupply(args[1])
		if err != nil {
			return shim.Error(err.Error())
		}
		tokenIDMetas := ""
		if len(args) > 2 {
			tokenIDMetas = args[2]
		}
		return p.SupplyToken(stub, args[0], supplyAmount, tokenIDMetas)
	case "existTokenID":
		if len(args) < 1 {
			return shim.Error("need 1 args (Asset_TokenID)")
		}
		exist, err := p.ExistTokenID(stub, args[0])
		if err != nil {
			return shim.Error(err.Error())
		}
		return shim.Success([]byte(exist))
	case "setTokenURI":
		if len(args) < 2 {
			return shim.Error("need 1 args (Asset_TokenID,TokenURI)")
		}
		return p.SetTokenURI(stub, args[0], args[1])
	case "getTokenURI":
		if len(args) < 1 {
			return shim.Error("need 1 args (Asset_TokenID)")
		}
		//asset
		asset := &dm.Asset{}
		err := asset.SetString(args[0])
		if err != nil {
			return shim.Error(jsonResp4)
		}
		tokenURI := p.GetTokenURI(stub, args[0])
		if len(tokenURI) == 0 {
			jsonResp := "{\"Error\":\"No this tokenID\"}"
			return shim.Error(jsonResp)
		}
		//
		return shim.Success([]byte(tokenURI))
	case "getTokenInfo":
		if len(args) < 1 {
			return shim.Error("need 1 args (Symbol)")
		}
		tkIDInfo, err := p.GetOneTokenInfo(stub, args[0])
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
		return p.ChangeSupplyAddr(stub, args[0], args[1])
	default:
		jsonResp := "{\"Error\":\"Unknown function " + f + "\"}"
		return shim.Error(jsonResp)
	}
}

func setGlobal(stub shim.ChaincodeStubInterface, tkInfo *tokenInfo) error {
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

func convertToByte(n uint64) []byte {
	by8 := make([]byte, 8)
	binary.BigEndian.PutUint64(by8, n)
	by16 := make([]byte, 16)
	copy(by16[16-len(by8):], by8)
	return by16
}

func generateUUID(seed []byte) ([]byte, error) {
	newHash, err := crypto.MyCryptoLib.Hash(seed)
	if err != nil {
		return nil, err
	}
	uuid := newHash[0:16]
	// variant bits; see section 4.1.1
	uuid[8] = uuid[8]&^0xc0 | 0x80
	// version 4 (pseudo-random); see section 4.1.3
	uuid[6] = uuid[6]&^0xf0 | 0x40
	return uuid, nil
}

type tokenIDMeta struct {
	TokenID  string
	MetaData string
}

func chekcTokenIDRepeat(tokenIDMetas []tokenIDMeta) bool {
	tokenIDs := make(map[string]bool)
	for _, oneTokenMeta := range tokenIDMetas {
		if _, ok := tokenIDs[oneTokenMeta.TokenID]; ok {
			return true
		}
		tokenIDs[oneTokenMeta.TokenID] = true
	}
	return false
}

func genNFData(txid []byte, idType dm.UniqueIdType, totalSupply uint64, start uint64, tokenIDMetas []tokenIDMeta) ([]dm.NonFungibleMetaData, string) {
	nfDatas := []dm.NonFungibleMetaData{}
	if idType == dm.UniqueIdType_Sequence {
		for i := uint64(0); i < totalSupply; i++ {
			seqByte := convertToByte(start + i)
			nFdata := dm.NonFungibleMetaData{UniqueBytes: seqByte}
			nfDatas = append(nfDatas, nFdata)
		}
	} else if idType == dm.UniqueIdType_Uuid {
		seed := txid
		for i := uint64(0); i < totalSupply; i++ {
			UUID, _ := generateUUID(seed)
			if len(UUID) < 16 {
				jsonResp := "{\"Error\":\"generateUUID() failed\"}"
				return nil, jsonResp
			}
			nFdata := dm.NonFungibleMetaData{UniqueBytes: UUID}
			seed = UUID
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

func getSupply(supplyStr string) (uint64, error) {
	totalSupply, err := strconv.ParseUint(supplyStr, 10, 64)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to convert total supply\"}"
		return 0, fmt.Errorf(jsonResp)
	}
	if totalSupply == 0 {
		jsonResp := "{\"Error\":\"Can't be zero\"}"
		return 0, fmt.Errorf(jsonResp)
	}
	if totalSupply > 1000 {
		jsonResp := "{\"Error\":\"Not allow bigger than 1000 NonFungibleToken when create\"}"
		return 0, fmt.Errorf(jsonResp)
	}
	return totalSupply, nil
}

//CreateToken
func (p *PRC721) CreateToken(stub shim.ChaincodeStubInterface, name string, symbol string, UIDType string,
	totalSupply uint64, tokenIDMetas string, supplyAddress string) pb.Response {
	//==== convert params to token information
	var nonFungible dm.NonFungibleToken
	//name symbol
	if len(name) > 1024 {
		jsonResp := "{\"Error\":\"Name length should not be greater than 1024\"}"
		return shim.Error(jsonResp)
	}
	nonFungible.Name = name
	nonFungible.Symbol = strings.ToUpper(symbol)
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
	if UIDType == "1" {
		idType = dm.UniqueIdType_Sequence
	} else if UIDType == "2" {
		idType = dm.UniqueIdType_Uuid
	} else if UIDType == "3" {
		idType = dm.UniqueIdType_UserDefine
	} else if UIDType == "4" {
		idType = dm.UniqueIdType_Ascii
	} else {
		jsonResp := "{\"Error\":\"Only string, 1(Seqence) or 2(UUID) or 3(Custom) or 4(Assii)\"}"
		return shim.Error(jsonResp)
	}
	nonFungible.Type = byte(idType)

	//total supply
	nonFungible.TotalSupply = totalSupply
	//tokenIDMetas
	var tkIDMetas []tokenIDMeta
	err := json.Unmarshal([]byte(tokenIDMetas), &tkIDMetas)
	if err != nil {
		jsonResp := "{\"Error\":\"tokenIDMetas format invalid, must be hex strings\"}"
		return shim.Error(jsonResp)
	}
	if uint64(len(tkIDMetas)) != totalSupply {
		jsonResp := "{\"Error\":\"tokenIDMetas and totalSupply is not match\"}"
		return shim.Error(jsonResp)
	}
	if idType == dm.UniqueIdType_UserDefine && chekcTokenIDRepeat(tkIDMetas) {
		jsonResp := "{\"Error\":\"tokenIDMetas have repeat tokenID\"}"
		return shim.Error(jsonResp)
	}
	//address of supply
	if len(supplyAddress) != 0 {
		nonFungible.SupplyAddress = supplyAddress
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

	txid := stub.GetTxID()
	txidHash := common.Hex2Bytes(txid[2:])
	//generate nonFungibleData
	nFdatas, errStr := genNFData(txidHash, idType, totalSupply, 1, tkIDMetas)
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
	assetID, _ := dm.NewAssetId(nonFungible.Symbol, dm.AssetType_NonFungibleToken,
		0, txidHash, idType)

	//
	newAsset := &dm.Asset{}
	newAsset.AssetId = assetID
	for i, nFdata := range nonFungible.NonFungibleData {
		newAsset.UniqueId.SetBytes(nFdata.UniqueBytes)
		key := newAsset.String()
		err = stub.PutState(key, []byte(tkIDMetas[i].MetaData))
		if err != nil {
			jsonResp := "{\"Error\":\"Failed to set Asset\"}"
			return shim.Error(jsonResp)
		}
	}
	info := tokenInfo{nonFungible.Symbol, byte(idType), totalSupply, createAddr.String(), totalSupply,
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

//SupplyToken
func (p *PRC721) SupplyToken(stub shim.ChaincodeStubInterface, symbol string, supplyAmount uint64, tokenIDMetas string) pb.Response {
	//symbol
	symbol = strings.ToUpper(symbol)
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
	if math.MaxInt64-tkInfo.TotalSupply < supplyAmount {
		jsonResp := "{\"Error\":\"Too big supplyAmount, overflow\"}"
		return shim.Error(jsonResp)
	}

	//tokenIDMetas
	var tkIDMetas []tokenIDMeta
	err := json.Unmarshal([]byte(tokenIDMetas), &tkIDMetas)
	if err != nil {
		jsonResp := "{\"Error\":\"tokenIDMetas format invalid, must be hex strings\"}"
		return shim.Error(jsonResp)
	}
	if uint64(len(tkIDMetas)) != supplyAmount {
		jsonResp := "{\"Error\":\"tokenIDMetas and supplyAmount is not match\"}"
		return shim.Error(jsonResp)
	}
	idType := dm.UniqueIdType(tkInfo.TokenType)
	if idType == dm.UniqueIdType_UserDefine && chekcTokenIDRepeat(tkIDMetas) {
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

	txid := stub.GetTxID()
	txidHash := common.Hex2Bytes(txid[2:])
	//call SupplyToken
	nFdatas, errStr := genNFData(txidHash, idType, supplyAmount, tkInfo.TokenMax+1, tkIDMetas)
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
		err = stub.PutState(key, []byte(tkIDMetas[i].MetaData))
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

//ChangeSupplyAddr
func (p *PRC721) ChangeSupplyAddr(stub shim.ChaincodeStubInterface, symbol, supplyAddress string) pb.Response {
	//symbol
	symbol = strings.ToUpper(symbol)
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
	err := checkAddr(supplyAddress)
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
	tkInfo.SupplyAddr = supplyAddress
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

type tokenIDInfo struct {
	Symbol      string
	CreateAddr  string
	TokenType   uint8 //no
	TotalSupply uint64
	SupplyAddr  string
	AssetID     string
	TokenIDs    []string //no
}

//ExistTokenID
func (p *PRC721) ExistTokenID(stub shim.ChaincodeStubInterface, assetTokenID string) (string, error) {
	//asset
	asset := &dm.Asset{}
	err := asset.SetString(assetTokenID)
	if err != nil {
		return "", fmt.Errorf(jsonResp4)
	}
	//
	valBytes, _ := stub.GetState(assetTokenID)
	if len(valBytes) == 0 {
		return "False", nil
	}
	return "True", nil
}

//SetTokenURI
func (p *PRC721) SetTokenURI(stub shim.ChaincodeStubInterface, assetTokenID, tokenURI string) pb.Response {
	//asset
	asset := &dm.Asset{}
	err := asset.SetString(assetTokenID)
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
	valBytes, _ := stub.GetState(assetTokenID)
	if len(valBytes) == 0 {
		jsonResp := "{\"Error\":\"No this tokenID\"}"
		return shim.Error(jsonResp)
	}

	err = stub.PutState(assetTokenID, []byte(tokenURI))
	if err != nil {
		return shim.Error("Failed to set tokenURI")
	}

	return shim.Success([]byte("True"))
}

//GetTokenURI
func (p *PRC721) GetTokenURI(stub shim.ChaincodeStubInterface, assetTokenID string) string {
	//
	valBytes, _ := stub.GetState(assetTokenID)
	return string(valBytes)
}

//GetOneTokenInfo
func (p *PRC721) GetOneTokenInfo(stub shim.ChaincodeStubInterface, symbol string) (*tokenIDInfo, error) {
	//symbol
	symbol = strings.ToUpper(symbol)
	//check name is exist or not
	tkInfo := getSymbols(stub, symbol)
	if tkInfo == nil {
		return nil, fmt.Errorf(jsonResp3)
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
	tkIDInfo := tokenIDInfo{symbol, tkInfo.CreateAddr, tkInfo.TokenType,
		tkInfo.TotalSupply, tkInfo.SupplyAddr, tkInfo.AssetID.String(), tkIDs}
	return &tkIDInfo, nil
}

//GetAllTokenInfo
func (p *PRC721) GetAllTokenInfo(stub shim.ChaincodeStubInterface) []tokenIDInfo {
	tkInfos := getSymbolsAll(stub)
	tkIDInfos := make([]tokenIDInfo, 0, len(tkInfos))
	tkIDs := []string{"Only return simple information"}
	for _, tkInfo := range tkInfos {
		tkIDInfo := tokenIDInfo{tkInfo.Symbol, tkInfo.CreateAddr,
			tkInfo.TokenType, tkInfo.TotalSupply,
			tkInfo.SupplyAddr, tkInfo.AssetID.String(), tkIDs}
		tkIDInfos = append(tkIDInfos, tkIDInfo)
	}
	return tkIDInfos
}
