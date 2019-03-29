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

package sysconfigcc

import (
	"encoding/json"
	"fmt"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/contracts/shim"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"github.com/palletone/go-palletone/dag/modules"
	"sort"
	"strconv"
	"strings"
	"time"
)

type SysConfigChainCode struct {
}

func (s *SysConfigChainCode) Init(stub shim.ChaincodeStubInterface) peer.Response {
	log.Info("*** SysConfigChainCode system contract init ***")
	return shim.Success([]byte("Success"))
}

func (s *SysConfigChainCode) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	funcName, args := stub.GetFunctionAndParameters()
	switch funcName {
	case "getAllSysParamsConf":
		log.Info("Start getAllSysParamsConf Invoke")
		resultByte, err := s.getAllSysParamsConf(stub)
		if err != nil {
			jsonResp := "{\"Error\":\"getAllSysParamsConf err: " + err.Error() + "\"}"
			return shim.Error(jsonResp)
		}
		return shim.Success(resultByte)
	case "getSysParamValByKey":
		log.Info("Start getSysParamValByKey Invoke")
		resultByte, err := s.getSysParamValByKey(stub, args)
		if err != nil {
			jsonResp := "{\"Error\":\"getSysParamValByKey err: " + err.Error() + "\"}"
			return shim.Error(jsonResp)
		}
		return shim.Success(resultByte)
	case "updateSysParamWithoutVote":
		log.Info("Start updateSysParamWithoutVote Invoke")
		resultByte, err := s.updateSysParamWithoutVote(stub, args)
		if err != nil {
			jsonResp := "{\"Error\":\"updateSysParamWithoutVote err: " + err.Error() + "\"}"
			return shim.Error(jsonResp)
		}
		return shim.Success(resultByte)
	case "getVotesResult":
		log.Info("Start getVotesResult Invoke")
		resultByte, err := s.getVotesResult(stub, args)
		if err != nil {
			jsonResp := "{\"Error\":\"getVotesResult err: " + err.Error() + "\"}"
			return shim.Success([]byte(jsonResp))
		}
		return shim.Success(resultByte)
	case "createVotesTokens":
		log.Info("Start createVotesTokens Invoke")
		resultByte, err := s.createVotesTokens(stub, args)
		if err != nil {
			jsonResp := "{\"Error\":\"createVotesTokens err: " + err.Error() + "\"}"
			return shim.Success([]byte(jsonResp))
		}
		return shim.Success(resultByte)
	default:
		log.Error("Invoke funcName err: ", "error", funcName)
		jsonResp := "{\"Error\":\"Invoke funcName err: " + funcName + "\"}"
		return shim.Error(jsonResp)
	}
}

func (s *SysConfigChainCode) getVotesResult(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	//params check
	if len(args) < 1 {
		return nil, fmt.Errorf("need 1 args (AssetID String)")
	}

	//assetIDStr
	assetIDStr := strings.ToUpper(args[0])
	//check name is exist or not
	tkInfo := getSymbols(stub, assetIDStr)
	if tkInfo == nil {
		jsonResp := "{\"Error\":\"Token not exist\"}"
		return nil, fmt.Errorf(jsonResp)
	}

	//get token information
	var topicSupports []TopicSupports
	err := json.Unmarshal(tkInfo.VoteContent, &topicSupports)
	if err != nil {
		jsonResp := "{\"Error\":\"Results format invalid, Error!!!\"}"
		return nil, fmt.Errorf(jsonResp)
	}

	//
	isVoteEnd := false
	headerTime, err := stub.GetTxTimestamp(10)
	if err != nil {
		jsonResp := "{\"Error\":\"GetTxTimestamp invalid, Error!!!\"}"
		return nil, fmt.Errorf(jsonResp)
	}
	if headerTime.Seconds > tkInfo.VoteEndTime.Unix() {
		isVoteEnd = true
	}
	//calculate result
	var supportResults []SupportResult
	for i, oneTopicSupport := range topicSupports {
		var oneResult SupportResult
		oneResult.TopicIndex = uint64(i) + 1
		oneResult.TopicTitle = oneTopicSupport.TopicTitle
		oneResultSort := sortSupportByCount(oneTopicSupport.VoteResults)
		oneResult.VoteResults = append(oneResult.VoteResults, oneResultSort...)
		//for i := uint64(0); i < oneTopicSupport.SelectMax; i++ {
		//	oneResult.VoteResults = append(oneResult.VoteResults, oneResultSort[i])
		//}
		supportResults = append(supportResults, oneResult)
	}

	//token
	asset := tkInfo.AssetID
	tkID := TokenIDInfo{IsVoteEnd: isVoteEnd, CreateAddr: tkInfo.CreateAddr, TotalSupply: tkInfo.TotalSupply,
		SupportResults: supportResults, AssetID: asset.String()}

	//return json
	tkJson, err := json.Marshal(tkID)
	if err != nil {
		jsonResp := "{\"Error\":\"" + err.Error() + "\"}"

		return nil, fmt.Errorf(jsonResp)
	}
	return tkJson, nil //test
}

func (s *SysConfigChainCode) createVotesTokens(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	//params check
	if len(args) < 5 {
		return nil, fmt.Errorf("need 5 args (Name,VoteType,TotalSupply,VoteEndTime,VoteContentJson)")
	}
	//get creator
	createAddr, err := stub.GetInvokeAddress()
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get invoke address\"}"
		return nil, fmt.Errorf(jsonResp)
	}
	//TODO 基金会地址
	//foundationAddress, _ := stub.GetSystemConfig("FoundationAddress")
	//if createAddr != foundationAddress {
	//	jsonResp := "{\"Error\":\"Only foundation can call this function\"}"
	//	return nil, fmt.Errorf(jsonResp)
	//}
	//==== convert params to token information
	var vt modules.VoteToken
	//name symbol
	vt.Name = args[0]
	vt.Symbol = "VOTE"

	//vote type
	//if args[1] == "0" {
	//	vt.VoteType = byte(0)
	//} else if args[1] == "1" {
	//	vt.VoteType = byte(1)
	//} else if args[1] == "2" {
	//	vt.VoteType = byte(2)
	//} else {
	//	jsonResp := "{\"Error\":\"Only string, 0 or 1 or 2\"}"
	//	return shim.Success([]byte(jsonResp))
	//}
	//total supply
	totalSupply, err := strconv.ParseUint(args[2], 10, 64)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to convert total supply\"}"
		return nil, fmt.Errorf(jsonResp)
	}
	if totalSupply == 0 {
		jsonResp := "{\"Error\":\"Can't be zero\"}"
		return nil, fmt.Errorf(jsonResp)
	}
	vt.TotalSupply = totalSupply
	//VoteEndTime
	VoteEndTime, err := time.Parse("2006-01-02 15:04:05", args[3])
	if err != nil {
		jsonResp := "{\"Error\":\"No vote end time\"}"
		return nil, fmt.Errorf(jsonResp)
	}
	vt.VoteEndTime = VoteEndTime
	//VoteContent
	var voteTopics []VoteTopic
	err = json.Unmarshal([]byte(args[4]), &voteTopics)
	if err != nil {
		jsonResp := "{\"Error\":\"VoteContent format invalid\"}"
		return nil, fmt.Errorf(jsonResp)
	}
	//init support
	var supports []TopicSupports
	for _, oneTopic := range voteTopics {
		var oneSupport TopicSupports
		oneSupport.TopicTitle = oneTopic.TopicTitle
		for _, oneOption := range oneTopic.SelectOptions {
			var oneResult VoteResult
			oneResult.SelectOption = oneOption
			oneSupport.VoteResults = append(oneSupport.VoteResults, oneResult)
		}
		//oneResult.SelectOptionsNum = uint64(len(oneRequest.SelectOptions))
		oneSupport.SelectMax = oneTopic.SelectMax
		supports = append(supports, oneSupport)
	}
	voteContentJson, err := json.Marshal(supports)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to generate voteContent Json\"}"
		return nil, fmt.Errorf(jsonResp)
	}
	vt.VoteContent = voteContentJson

	txid := stub.GetTxID()
	assetID, _ := modules.NewAssetId(vt.Symbol, modules.AssetType_VoteToken,
		0, common.Hex2Bytes(txid[2:]), modules.UniqueIdType_Null)
	assetIDStr := assetID.String()
	//check name is only or not
	tkInfo := getSymbols(stub, assetIDStr)
	if tkInfo != nil {
		jsonResp := "{\"Error\":\"Repeat AssetID\"}"
		return nil, fmt.Errorf(jsonResp)
	}

	//convert to json
	createJson, err := json.Marshal(vt)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to generate token Json\"}"
		return nil, fmt.Errorf(jsonResp)
	}

	//last put state
	info := TokenInfo{vt.Name, vt.Symbol, createAddr, vt.VoteType, totalSupply,
		VoteEndTime, voteContentJson, assetID}

	err = setSymbols(stub, &info)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to set symbols\"}"
		return nil, fmt.Errorf(jsonResp)
	}

	//set token define
	err = stub.DefineToken(byte(modules.AssetType_VoteToken), createJson, createAddr)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to call stub.DefineToken\"}"
		return nil, fmt.Errorf(jsonResp)
	}

	return createJson, nil //test
}

func (s *SysConfigChainCode) nodesVote(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	return nil, nil
}

func (s *SysConfigChainCode) getAllSysParamsConf(stub shim.ChaincodeStubInterface) ([]byte, error) {
	sysVal, err := stub.GetState("sysConf")
	if err != nil {
		return nil, err
	}
	return sysVal, nil
}

func (s *SysConfigChainCode) updateSysParamWithoutVote(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	//invokeFromAddr, err := stub.GetInvokeAddress()
	//if err != nil {
	//	return nil, err
	//}
	//TODO 基金会地址
	//foundationAddress, _ := stub.GetSystemConfig("FoundationAddress")
	//if invokeFromAddr != foundationAddress {
	//	jsonResp := "{\"Error\":\"Only foundation can call this function\"}"
	//	return nil, fmt.Errorf(jsonResp)
	//}
	key := args[0]
	newValue := args[1]
	oldValue, err := stub.GetState(args[0])
	if err != nil {
		return nil, err
	}
	err = stub.PutState(key, []byte(newValue))
	if err != nil {
		return nil, err
	}
	sysValByte, err := stub.GetState("sysConf")
	if err != nil {
		return nil, err
	}
	sysVal := &core.SystemConfig{}
	err = json.Unmarshal(sysValByte, sysVal)
	if err != nil {
		return nil, err
	}
	switch key {
	case "DepositAmountForJury":
		sysVal.DepositAmountForJury = newValue
	case "DepositRate":
		sysVal.DepositRate = newValue
	case "FoundationAddress":
		sysVal.FoundationAddress = newValue
	case "DepositAmountForMediator":
		sysVal.DepositAmountForMediator = newValue
	case "DepositAmountForDeveloper":
		sysVal.DepositAmountForDeveloper = newValue
	case "DepositPeriod":
		sysVal.DepositPeriod = newValue
	case "RootCaHolder":
		sysVal.RootCaHolder = newValue
	}
	sysValByte, err = json.Marshal(sysVal)
	if err != nil {
		return nil, err
	}
	err = stub.PutState("sysConf", sysValByte)
	if err != nil {
		return nil, err
	}
	jsonResp := "{\"Success\":\"update value from " + string(oldValue) + " to " + newValue + "\"}"
	return []byte(jsonResp), nil
}

func (s *SysConfigChainCode) getSysParamValByKey(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	if len(args) != 1 {
		jsonResp := "{\"Error\":\" need 1 args (AssetID String)\"}"
		return nil, fmt.Errorf(jsonResp)
	}
	val, err := stub.GetState(args[0])
	if err != nil {
		return nil, err
	}
	jsonResp := "{\"" + args[0] + "\":\"" + string(val) + "\"}"
	return []byte(jsonResp), nil
}

func getSymbols(stub shim.ChaincodeStubInterface, assetID string) *TokenInfo {
	//
	tkInfo := TokenInfo{}
	tkInfoBytes, _ := stub.GetState(symbolsKey + assetID)
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
func setSymbols(stub shim.ChaincodeStubInterface, tkInfo *TokenInfo) error {
	val, err := json.Marshal(tkInfo)
	if err != nil {
		return err
	}
	err = stub.PutState(symbolsKey+tkInfo.AssetID.String(), val)
	return err
}

// A slice of TopicResult that implements sort.Interface to sort by Value.
type VoteResultList []VoteResult

func (p VoteResultList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p VoteResultList) Len() int           { return len(p) }
func (p VoteResultList) Less(i, j int) bool { return p[i].Num > p[j].Num }

// A function to turn a map into a TopicResultList, then sort and return it.
func sortSupportByCount(tpl VoteResultList) VoteResultList {
	sort.Stable(tpl) //sort.Sort(tpl)
	return tpl
}
