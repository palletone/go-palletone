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

const symbolsKey = "symbols"

type PRC721 struct {
}

type TokenInfo struct {
	Symbol      string
	TokenType   uint8
	TokenMax    uint64 //only use when TokenType is Sequence
	CreateAddr  string
	TotalSupply uint64
	SupplyAddr  string
	AssetID     dm.IDType16
}

type Symbols struct {
	TokenInfos map[string]TokenInfo `json:"tokeninfos"`
}

func (p *PRC721) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

func (p *PRC721) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	f, args := stub.GetFunctionAndParameters()

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
		return nil, nil
	}
	// variant bits; see section 4.1.1
	uuid[8] = uuid[8]&^0xc0 | 0x80
	// version 4 (pseudo-random); see section 4.1.3
	uuid[6] = uuid[6]&^0xf0 | 0x40
	return uuid, nil
}

func createToken(args []string, stub shim.ChaincodeStubInterface) pb.Response {
	//params check
	if len(args) < 4 {
		return shim.Error("need 5 args (Name,Symbol,Type,TotalSupply,[TokenIDs,SupplyAddress])")
	}

	//==== convert params to token information
	var nonFungible dm.NonFungibleToken
	//name symbol
	nonFungible.Name = args[0]
	nonFungible.Symbol = strings.ToUpper(args[1])
	if nonFungible.Symbol == "PTN" {
		jsonResp := "{\"Error\":\"Can't use PTN\"}"
		return shim.Success([]byte(jsonResp))
	}
	if len(nonFungible.Symbol) > 5 {
		jsonResp := "{\"Error\":\"Symbol must less than 5 characters\"}"
		return shim.Success([]byte(jsonResp))
	}
	//type
	idType := dm.UniqueIdType_Null
	if args[2] == "1" {
		idType = dm.UniqueIdType_Sequence
	} else if args[2] == "2" {
		idType = dm.UniqueIdType_Uuid
	} else if args[2] == "3" {
		idType = dm.UniqueIdType_UserDefine
	} else {
		jsonResp := "{\"Error\":\"Only string, 1(Seqence) or 2(UUID) or 3(Custom)\"}"
		return shim.Success([]byte(jsonResp))
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
		return shim.Success([]byte(jsonResp))
	}
	if totalSupply > 1000 {
		jsonResp := "{\"Error\":\"Not allow bigger than 1000 NonFungibleToken when create\"}"
		return shim.Success([]byte(jsonResp))
	}
	nonFungible.TotalSupply = totalSupply
	//tokenIDs
	var tokenIDStrs []string
	if idType == dm.UniqueIdType_UserDefine {
		if len(args) < 5 {
			jsonResp := "{\"Error\":\"Your tokeType is 2(Custom), need tokenIDs\"}"
			return shim.Success([]byte(jsonResp))
		}
		err = json.Unmarshal([]byte(args[4]), &tokenIDStrs)
		if err != nil {
			jsonResp := "{\"Error\":\"tokenIDs format invalid, must be hex strings\"}"
			return shim.Success([]byte(jsonResp))
		}
		if uint64(len(tokenIDStrs)) != totalSupply {
			if err != nil {
				jsonResp := "{\"Error\":\"tokenIDs and totalSupply is not match\"}"
				return shim.Success([]byte(jsonResp))
			}
		}
	}
	//address of supply
	if len(args) > 5 {
		nonFungible.SupplyAddress = args[5]
	}

	//check name is only or not
	symbols, err := getSymbols(stub)
	if _, ok := symbols.TokenInfos[nonFungible.Symbol]; ok {
		jsonResp := "{\"Error\":\"The symbol have been used\"}"
		return shim.Success([]byte(jsonResp))
	}
	//generate nonFungibleData
	if idType == dm.UniqueIdType_Sequence {
		start := uint64(1)
		for i := uint64(0); i < totalSupply; i++ {
			seqByte := convertToByte(start + i)
			nFdata := dm.NonFungibleMetaData{seqByte}
			nonFungible.NonFungibleData = append(nonFungible.NonFungibleData, nFdata)
		}
	} else if idType == dm.UniqueIdType_Uuid {
		for i := uint64(0); i < totalSupply; i++ {
			UUID, _ := generateUUID()
			if len(UUID) < 16 {
				jsonResp := "{\"Error\":\"generateUUID() failed\"}"
				return shim.Success([]byte(jsonResp))
			}
			nFdata := dm.NonFungibleMetaData{UUID}
			nonFungible.NonFungibleData = append(nonFungible.NonFungibleData, nFdata)
		}
	} else if idType == dm.UniqueIdType_UserDefine {
		for _, oneTokenID := range tokenIDStrs {
			oneTokenIDByte, _ := hex.DecodeString(oneTokenID)
			if len(oneTokenID) < 16 {
				jsonResp := "{\"Error\":\"tokenIDs format invalid, must be hex string\"}"
				return shim.Success([]byte(jsonResp))
			}
			nFdata := dm.NonFungibleMetaData{oneTokenIDByte}
			nonFungible.NonFungibleData = append(nonFungible.NonFungibleData, nFdata)
		}
	}

	//convert to json
	createJson, err := json.Marshal(nonFungible)
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
	err = stub.DefineToken(byte(dm.AssetType_NonFungibleToken), createJson, createAddr)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to call stub.DefineToken\"}"
		return shim.Error(jsonResp)
	}

	//last put state
	txid := stub.GetTxID()
	assetID, _ := dm.NewAssetId(nonFungible.Symbol, dm.AssetType_NonFungibleToken,
		0, common.Hex2Bytes(txid[2:]), idType)

	//
	newAsset := &dm.Asset{}
	newAsset.AssetId = assetID
	for _, nFdata := range nonFungible.NonFungibleData {
		newAsset.UniqueId.SetBytes(nFdata.UniqueBytes)
		key := newAsset.String()
		err = stub.PutState(key, []byte("0"))
		if err != nil {
			jsonResp := "{\"Error\":\"Failed to set Asset\"}"
			return shim.Error(jsonResp)
		}
	}

	info := TokenInfo{nonFungible.Symbol, byte(idType), totalSupply, createAddr, totalSupply,
		nonFungible.SupplyAddress, assetID}
	symbols.TokenInfos[nonFungible.Symbol] = info

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
		return shim.Error("need 2 args (Symbol,SupplyAmout,[TokenIDs])")
	}

	//symbol
	symbol := strings.ToUpper(args[0])
	//check name is exist or not
	symbols, err := getSymbols(stub)
	if _, ok := symbols.TokenInfos[symbol]; !ok {
		jsonResp := "{\"Error\":\"Token not exist\"}"
		return shim.Success([]byte(jsonResp))
	}
	tokenInfo := symbols.TokenInfos[symbol]

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
	if supplyAmount > 1000 {
		jsonResp := "{\"Error\":\"Not allow bigger than 1000 NonFungibleToken when create\"}"
		return shim.Success([]byte(jsonResp))
	}
	if math.MaxInt64-tokenInfo.TotalSupply < supplyAmount {
		jsonResp := "{\"Error\":\"Too big, overflow\"}"
		return shim.Success([]byte(jsonResp))
	}

	//tokenIDs
	var tokenIDStrs []string
	if len(args) > 2 && tokenInfo.TokenType == byte(dm.UniqueIdType_UserDefine) {
		if len(args) < 2 {
			jsonResp := "{\"Error\":\"Your tokeType is 2(Custom), need tokenIDs\"}"
			return shim.Success([]byte(jsonResp))
		}
		err = json.Unmarshal([]byte(args[2]), &tokenIDStrs)
		if err != nil {
			jsonResp := "{\"Error\":\"tokenIDs format invalid, must be hex strings\"}"
			return shim.Success([]byte(jsonResp))
		}
		if uint64(len(tokenIDStrs)) != supplyAmount {
			if err != nil {
				jsonResp := "{\"Error\":\"tokenIDs and supplyAmount is not match\"}"
				return shim.Success([]byte(jsonResp))
			}
		}
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
	idType := dm.UniqueIdType(symbols.TokenInfos[symbol].TokenType)
	nFdatas := []dm.NonFungibleMetaData{}
	if idType == dm.UniqueIdType_Sequence {
		start := symbols.TokenInfos[symbol].TokenMax + 1
		for i := uint64(0); i < supplyAmount; i++ {
			seqByte := convertToByte(start + i)
			err = stub.SupplyToken(assetID.Bytes(), seqByte, 1, invokeAddr)
			if err != nil {
				jsonResp := "{\"Error\":\"Failed to call stub.SupplyToken\"}"
				return shim.Error(jsonResp)
			}
			nFdatas = append(nFdatas, dm.NonFungibleMetaData{seqByte})
		}
	} else if idType == dm.UniqueIdType_Uuid {
		for i := uint64(0); i < supplyAmount; i++ {
			UUID, _ := generateUUID()
			if len(UUID) < 16 {
				jsonResp := "{\"Error\":\"generateUUID() failed\"}"
				return shim.Success([]byte(jsonResp))
			}
			err = stub.SupplyToken(assetID.Bytes(), UUID, 1, invokeAddr)
			if err != nil {
				jsonResp := "{\"Error\":\"Failed to call stub.SupplyToken\"}"
				return shim.Error(jsonResp)
			}
			nFdatas = append(nFdatas, dm.NonFungibleMetaData{UUID})
		}
	} else if idType == dm.UniqueIdType_UserDefine {
		for _, oneTokenID := range tokenIDStrs {
			oneTokenIDByte, _ := hex.DecodeString(oneTokenID)
			if len(oneTokenID) < 16 {
				jsonResp := "{\"Error\":\"tokenIDs format invalid, must be hex string\"}"
				return shim.Success([]byte(jsonResp))
			}
			err = stub.SupplyToken(assetID.Bytes(), oneTokenIDByte, 1, invokeAddr)
			if err != nil {
				jsonResp := "{\"Error\":\"Failed to call stub.SupplyToken\"}"
				return shim.Error(jsonResp)
			}
			nFdatas = append(nFdatas, dm.NonFungibleMetaData{oneTokenIDByte})
		}
	}

	//
	newAsset := &dm.Asset{}
	newAsset.AssetId = assetID
	for _, nFdata := range nFdatas {
		newAsset.UniqueId.SetBytes(nFdata.UniqueBytes)
		key := newAsset.String()
		err = stub.PutState(key, []byte("0"))
		if err != nil {
			jsonResp := "{\"Error\":\"Failed to set Asset\"}"
			return shim.Error(jsonResp)
		}
	}

	//add supply
	tokenInfo.TotalSupply += supplyAmount
	if idType == dm.UniqueIdType_Sequence {
		tokenInfo.TokenMax += supplyAmount
	}
	symbols.TokenInfos[symbol] = tokenInfo
	err = setSymbols(symbols, stub)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to set symbols\"}"
		return shim.Error(jsonResp)
	}
	return shim.Success([]byte("")) //test
}

type TokenIDInfo struct {
	Symbol      string
	CreateAddr  string
	TokenType   uint8
	TotalSupply uint64
	SupplyAddr  string
	AssetID     string
	TokenIDs    []string
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
		jsonResp := "{\"Error\":\"Asset_TokenID invalid\"}"
		return shim.Success([]byte(jsonResp))
	}
	//
	valBytes, err := stub.GetState(assetStr)
	if len(valBytes) == 0 {
		return shim.Success([]byte("False")) //
	}
	return shim.Success([]byte("True")) //test
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
		jsonResp := "{\"Error\":\"Asset_TokenID invalid\"}"
		return shim.Success([]byte(jsonResp))
	}
	//
	valBytes, _ := stub.GetState(assetStr)
	if len(valBytes) == 0 {
		jsonResp := "{\"Error\":\"No this tokenID\"}"
		return shim.Success([]byte(jsonResp))
	}

	err = stub.PutState(assetStr, []byte(tokenURI))
	if err != nil {
		return shim.Success([]byte("Failed to set tokenURI")) //test
	}

	return shim.Success([]byte("True")) //test
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
		jsonResp := "{\"Error\":\"Asset_TokenID invalid\"}"
		return shim.Success([]byte(jsonResp))
	}
	//
	valBytes, _ := stub.GetState(assetStr)
	if len(valBytes) == 0 {
		jsonResp := "{\"Error\":\"No this tokenID\"}"
		return shim.Success([]byte(jsonResp))
	}
	//
	return shim.Success(valBytes) //test
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

	//token assetID
	assetID := symbols.TokenInfos[symbol].AssetID

	//
	var tkIDs []string
	KVs, _ := stub.GetStateByPrefix(assetID.ToAssetId())
	for _, oneKV := range KVs {
		assetTkID := strings.SplitN(oneKV.Key, "-", 2)
		if len(assetTkID) == 2 {
			tkIDs = append(tkIDs, assetTkID[1])
		}
	}
	sort.Strings(tkIDs)

	//
	tkIDInfo := TokenIDInfo{symbol, symbols.TokenInfos[symbol].CreateAddr,
		symbols.TokenInfos[symbol].TokenType, symbols.TokenInfos[symbol].TotalSupply,
		symbols.TokenInfos[symbol].SupplyAddr, assetID.ToAssetId(), tkIDs}
	//return json
	tkJson, err := json.Marshal(tkIDInfo)
	if err != nil {
		return shim.Success([]byte(err.Error()))
	}
	return shim.Success(tkJson) //test
}

func allToken(args []string, stub shim.ChaincodeStubInterface) pb.Response {
	symbols, err := getSymbols(stub)

	var tkIDInfos []TokenIDInfo
	tkIDs := []string{"Only return simple information"}
	for symbol := range symbols.TokenInfos {
		asset := symbols.TokenInfos[symbol].AssetID
		tkIDInfo := TokenIDInfo{symbol, symbols.TokenInfos[symbol].CreateAddr,
			symbols.TokenInfos[symbol].TokenType, symbols.TokenInfos[symbol].TotalSupply,
			symbols.TokenInfos[symbol].SupplyAddr, asset.ToAssetId(), tkIDs}
		tkIDInfos = append(tkIDInfos, tkIDInfo)
	}

	//return json
	tksJson, err := json.Marshal(tkIDInfos)
	if err != nil {
		return shim.Success([]byte(err.Error()))
	}
	return shim.Success(tksJson) //test
}
