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
	"sort"
	"strconv"
	"time"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/contracts/shim"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"github.com/palletone/go-palletone/dag/modules"
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
	//case "getAllSysParamsConf":
	//	log.Info("Start getAllSysParamsConf Invoke")
	//	resultByte, err := s.getAllSysParamsConf(stub)
	//	if err != nil {
	//		jsonResp := "{\"Error\":\"getAllSysParamsConf err: " + err.Error() + "\"}"
	//		return shim.Error(jsonResp)
	//	}
	//	resut := ptnjson.ConvertAllSysConfigToJson(resultByte)
	//	res, err := json.Marshal(resut)
	//	if err != nil {
	//		jsonResp := "{\"Error\":\"getAllSysParamsConf err: " + err.Error() + "\"}"
	//		return shim.Error(jsonResp)
	//	}
	//	return shim.Success(res)
	//case "getSysParamValByKey":
	//	log.Info("Start getSysParamValByKey Invoke")
	//	resultByte, err := s.getSysParamValByKey(stub, args)
	//	if err != nil {
	//		jsonResp := "{\"Error\":\"getSysParamValByKey err: " + err.Error() + "\"}"
	//		return shim.Error(jsonResp)
	//	}
	//	return shim.Success(resultByte)
	case UpdateSysParamWithoutVote:
		if len(args) != 2 {
			err := "args len not equal 2"
			log.Debugf(err)
			jsonResp := "{\"Error\":\"updateSysParamWithoutVote err: " + err + "\"}"
			return shim.Error(jsonResp)
		}
		log.Info("Start updateSysParamWithoutVote Invoke")
		resultByte, err := s.UpdateSysParamWithoutVote(stub, args[0], args[1])
		if err != nil {
			jsonResp := "{\"Error\":\"updateSysParamWithoutVote err: " + err.Error() + "\"}"
			return shim.Error(jsonResp)
		}
		return shim.Success(resultByte)
	case "getWithoutVoteResult":
		log.Info("Start getWithoutVoteResult Invoke")
		resultByte, err := s.GetWithoutVoteResult(stub)
		if err != nil {
			jsonResp := "{\"Error\":\"getWithoutVoteResult err: " + err.Error() + "\"}"
			return shim.Success([]byte(jsonResp))
		}
		return shim.Success(resultByte)
	case "getVotesResult":
		log.Info("Start getVotesResult Invoke")
		result, err := s.GetVotesResult(stub /*, args*/)
		if err != nil {
			jsonResp := "{\"Error\":\"getVotesResult err: " + err.Error() + "\"}"
			return shim.Error(jsonResp)
		}
		data, _ := json.Marshal(result)
		return shim.Success(data)
	case CreateVotesTokens:
		if len(args) != 5 {
			err := "need 5 args (Name,TotalSupply,LeastNum,VoteEndTime,VoteContentJson)"
			jsonResp := "{\"Error\":\"createVotesTokens err: " + err + "\"}"
			return shim.Success([]byte(jsonResp))
		}
		log.Info("Start createVotesTokens Invoke")
		totalSupply, err := strconv.ParseUint(args[1], 10, 64)
		if err != nil {
			jsonResp := "{\"Error\":\"Failed to convert total supply\"}"
			return shim.Error(jsonResp)
		}
		if totalSupply == 0 {
			jsonResp := "{\"Error\":\"Can't be zero\"}"
			return shim.Error(jsonResp)
		}
		leastNum, err := strconv.ParseUint(args[2], 10, 64)
		if err != nil {
			jsonResp := "{\"Error\":\"Failed to convert least numbers\"}"
			return shim.Error(jsonResp)
		}
		if leastNum == 0 {
			jsonResp := "{\"Error\":\"Can't be zero\"}"
			return shim.Error(jsonResp)
		}
		resultByte, err := s.CreateVotesTokens(stub, args[0], totalSupply, leastNum, args[3], args[4])
		if err != nil {
			jsonResp := "{\"Error\":\"createVotesTokens err: " + err.Error() + "\"}"
			return shim.Success([]byte(jsonResp))
		}
		return shim.Success(resultByte)
	case "nodesVote":
		if len(args) < 1 {
			err := "need 1 args (SupportRequestJson)"
			jsonResp := "{\"Error\":\"nodesVote err: " + err + "\"}"
			return shim.Success([]byte(jsonResp))
		}
		log.Info("Start nodesVote Invoke")
		resultByte, err := s.NodesVote(stub, args[0])
		if err != nil {
			jsonResp := "{\"Error\":\"nodesVote err: " + err.Error() + "\"}"
			return shim.Success([]byte(jsonResp))
		}
		return shim.Success(resultByte)
	default:
		log.Error("Invoke funcName err: ", "error", funcName)
		jsonResp := "{\"Error\":\"Invoke funcName err: " + funcName + "\"}"
		return shim.Error(jsonResp)
	}
}
func (s *SysConfigChainCode) GetWithoutVoteResult(stub shim.ChaincodeStubInterface) ([]byte, error) {
	return stub.GetState(modules.DesiredSysParamsWithoutVote)
}

type SysTokenIDInfo struct {
	CreateAddr     string
	TotalSupply    uint64
	LeastNum       uint64
	AssetID        string
	CreateTime     int64
	IsVoteEnd      bool
	SupportResults []*SysSupportResult
}
type SysSupportResult struct {
	TopicIndex  uint64
	TopicTitle  string
	VoteResults []*SysVoteResult
}
type SysVoteResult struct {
	SelectOption string
	Num          uint64
}

func (s *SysConfigChainCode) GetVotesResult(stub shim.ChaincodeStubInterface /*, args []string*/) (*SysTokenIDInfo, error) {
	//check name is exist or not
	tkInfo := getSymbols(stub)
	if tkInfo == nil {
		return nil, fmt.Errorf("Token not exist")
	}

	//get token information
	var topicSupports []SysTopicSupports
	err := json.Unmarshal(tkInfo.VoteContent, &topicSupports)
	if err != nil {
		return nil, fmt.Errorf("Results format invalid, Error!!!")
	}

	//
	isVoteEnd := false
	headerTime, err := stub.GetTxTimestamp(10)
	if err != nil {
		return nil, fmt.Errorf("GetTxTimestamp invalid, Error!!!")
	}
	if headerTime.Seconds > tkInfo.VoteEndTime.Unix() {
		isVoteEnd = true
	}
	//calculate result
	supportResults := make([]*SysSupportResult, 0, len(topicSupports))
	for i, oneTopicSupport := range topicSupports {
		oneResult := &SysSupportResult{}
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
	tkID := SysTokenIDInfo{IsVoteEnd: isVoteEnd, CreateAddr: tkInfo.CreateAddr, TotalSupply: tkInfo.TotalSupply,
		SupportResults: supportResults, AssetID: asset.String(), CreateTime: tkInfo.VoteEndTime.UTC().Unix(), LeastNum: tkInfo.LeastNum}
	return &tkID, nil
}

func (s *SysConfigChainCode) CreateVotesTokens(stub shim.ChaincodeStubInterface, name string, totalSupply uint64,
	leastNum uint64, voteEndTime string, voteContentJSON string) ([]byte, error) {
	//get creator
	createAddr, err := stub.GetInvokeAddress()
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get invoke address\"}"
		return nil, fmt.Errorf(jsonResp)
	}

	gp, err := stub.GetSystemConfig()
	if err != nil {
		return nil, fmt.Errorf("fail to get system config err")
	}
	if createAddr.Str() != gp.ChainParameters.FoundationAddress {
		jsonResp := "{\"Error\":\"Only foundation can call this function\"}"
		return nil, fmt.Errorf(jsonResp)
	}

	//==== convert params to token information
	var vt modules.VoteToken
	//name symbol
	if len(name) > 1024 {
		jsonResp := "{\"Error\":\"Name length should not be greater than 1024\"}"
		return []byte{}, fmt.Errorf(jsonResp)
	}
	vt.Name = name
	vt.Symbol = "SVOTE"

	//total supply
	vt.TotalSupply = totalSupply

	//endTime
	endTime, err := time.Parse("2006-01-02 15:04:05", voteEndTime)
	if err != nil {
		jsonResp := "{\"Error\":\"No vote end time\"}"
		return nil, fmt.Errorf(jsonResp)
	}
	vt.VoteEndTime = endTime
	//VoteContent
	var voteTopics []SysVoteTopic
	err = json.Unmarshal([]byte(voteContentJSON), &voteTopics)
	if err != nil {
		jsonResp := "{\"Error\":\"VoteContent format invalid\"}"
		return nil, fmt.Errorf(jsonResp)
	}
	//init support
	supports := make([]SysTopicSupports, 0, len(voteTopics))
	for _, oneTopic := range voteTopics {
		oneSupport := SysTopicSupports{TopicTitle: oneTopic.TopicTitle}
		for _, oneOption := range oneTopic.SelectOptions {
			// 检查参数
			err := core.CheckSysConfigArgType(oneSupport.TopicTitle, oneOption)
			if err != nil {
				log.Debugf(err.Error())
				return nil, err
			}

			err = core.CheckChainParameterValue(oneSupport.TopicTitle, oneOption, &gp.ImmutableParameters,
				&gp.ChainParameters, func() int { return getMediatorCount(stub) })
			if err != nil {
				log.Debugf(err.Error())
				return nil, err
			}

			oneResult := &SysVoteResult{}
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
	//assetIDStr := assetID.String()
	//check name is only or not
	tkInfo := getSymbols(stub)
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
	info := SysTokenInfo{vt.Name, vt.Symbol, createAddr.String(), leastNum, totalSupply,
		endTime, voteContentJson, assetID}

	err = setSymbols(stub, &info)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to set symbols\"}"
		return nil, fmt.Errorf(jsonResp)
	}

	//set token define
	err = stub.DefineToken(byte(modules.AssetType_VoteToken), createJson, createAddr.String())
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to call stub.DefineToken\"}"
		return nil, fmt.Errorf(jsonResp)
	}

	//add global state
	err = setGlobal(stub, &info)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to add global state\"}"
		return nil, fmt.Errorf(jsonResp)
	}

	return createJson, nil //test
}

func (s *SysConfigChainCode) NodesVote(stub shim.ChaincodeStubInterface, supportRequestJson string) ([]byte, error) {
	//check token
	invokeTokens, err := stub.GetInvokeTokens()
	if err != nil {
		return nil, fmt.Errorf("GetInvokeTokens failed")
	}
	voteNum := uint64(0)
	assetIDStr := ""
	for i := 0; i < len(invokeTokens); i++ {
		if invokeTokens[i].Asset.AssetId == modules.PTNCOIN {
			continue
		} else if invokeTokens[i].Address == "P1111111111111111111114oLvT2" {
			if assetIDStr == "" {
				assetIDStr = invokeTokens[i].Asset.String()
				voteNum += invokeTokens[i].Amount
			} else if invokeTokens[i].Asset.AssetId.String() == assetIDStr {
				voteNum += invokeTokens[i].Amount
			}
		}
	}
	if voteNum == 0 || assetIDStr == "" { //no vote token
		return nil, fmt.Errorf("Vote token empty")
	}

	//check name is exist or not
	tkInfo := getSymbols(stub)
	if tkInfo == nil {
		return nil, fmt.Errorf("Token not exist")
	}

	//parse support requests
	var supportRequests []SysSupportRequest
	err = json.Unmarshal([]byte(supportRequestJson), &supportRequests)
	if err != nil {
		return nil, fmt.Errorf("SupportRequestJson format invalid")
	}
	//get token information
	var topicSupports []SysTopicSupports
	err = json.Unmarshal(tkInfo.VoteContent, &topicSupports)
	if err != nil {
		return nil, fmt.Errorf("Results format invalid, Error!!!")

	}

	//if voteNum < uint64(len(supportRequests)) { //vote token more than request
	//	return nil, fmt.Errorf("Vote token more than support request")
	//}

	//check time
	headerTime, err := stub.GetTxTimestamp(10)
	if err != nil {
		return nil, fmt.Errorf("GetTxTimestamp invalid, Error!!!")

	}
	if headerTime.Seconds > tkInfo.VoteEndTime.Unix() {
		return nil, fmt.Errorf("Vote is over")

	}

	//save support
	indexHistory := make(map[uint64]uint8)
	indexRepeat := false
	for _, oneSupport := range supportRequests {
		topicIndex := oneSupport.TopicIndex - 1
		if _, ok := indexHistory[topicIndex]; ok { //check select repeat
			indexRepeat = true
			break
		}
		indexHistory[topicIndex] = 1
		if topicIndex < uint64(len(topicSupports)) { //1.check index, must not out of total
			if uint64(len(oneSupport.SelectIndexs)) <= topicSupports[topicIndex].SelectMax { //2.check one select's options, must not out of select's max
				lenOfVoteResult := uint64(len(topicSupports[topicIndex].VoteResults))
				selIndexHistory := make(map[uint64]uint8)
				for _, index := range oneSupport.SelectIndexs {
					selectIndex := index - 1
					if _, ok := selIndexHistory[selectIndex]; ok { //check select repeat
						break
					}
					selIndexHistory[selectIndex] = 1
					if selectIndex < lenOfVoteResult { //3.index must be real select options
						topicSupports[topicIndex].VoteResults[selectIndex].Num += voteNum //1
					}
				}
			}
		}
	}
	if indexRepeat {
		return nil, fmt.Errorf("Repeat index of select option ")

	}
	voteContentJson, err := json.Marshal(topicSupports)
	if err != nil {
		return nil, fmt.Errorf("Failed to generate voteContent Json")

	}
	tkInfo.VoteContent = voteContentJson

	//save token information
	err = setSymbols(stub, tkInfo)
	if err != nil {
		return nil, fmt.Errorf("Failed to set symbols")

	}
	return []byte("NodesVote success."), nil
}

//func (s *SysConfigChainCode) getAllSysParamsConf(stub shim.ChaincodeStubInterface) (map[string]*modules.ContractStateValue, error) {
//	sysVal, err := stub.GetContractAllState()
//	if err != nil {
//		return nil, err
//	}
//	return sysVal, nil
//}

func getMediatorCount(stub shim.ChaincodeStubInterface) int {
	byte, err := stub.GetState(modules.MediatorList)
	if err != nil {
		return 0
	}
	if len(byte) == 0 {
		return 0
	}

	list := make(map[string]string)
	err = json.Unmarshal(byte, &list)
	if err != nil {
		return 0
	}

	return len(list)
}

func (s *SysConfigChainCode) UpdateSysParamWithoutVote(stub shim.ChaincodeStubInterface, field, value string) ([]byte, error) {
	// 检查参数
	err := core.CheckSysConfigArgType(field, value)
	if err != nil {
		log.Debugf(err.Error())
		return nil, err
	}

	gp, err := stub.GetSystemConfig()
	if err != nil {
		return nil, fmt.Errorf("fail to get system config err")
	}

	err = core.CheckChainParameterValue(field, value, &gp.ImmutableParameters,
		&gp.ChainParameters, func() int { return getMediatorCount(stub) })
	if err != nil {
		log.Debugf(err.Error())
		return nil, err
	}

	createAddr, err := stub.GetInvokeAddress()
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get invoke address\"}"
		return nil, fmt.Errorf(jsonResp)
	}

	if createAddr.Str() != gp.ChainParameters.FoundationAddress {
		jsonResp := "{\"Error\":\"Only foundation can call this function\"}"
		return nil, fmt.Errorf(jsonResp)
	}

	resultBytes, err := stub.GetState(modules.DesiredSysParamsWithoutVote)
	if err != nil {
		log.Debugf(err.Error())
		return nil, err
	}

	var modifies map[string]string
	if resultBytes != nil && string(resultBytes) != "" {
		err := json.Unmarshal(resultBytes, &modifies)
		if err != nil {
			log.Debugf(err.Error())
			return nil, err
		}
	}

	if modifies == nil {
		modifies = make(map[string]string)
	}

	modifies[field] = value
	modifyByte, err := json.Marshal(modifies)
	if err != nil {
		return nil, err
	}
	err = stub.PutState(modules.DesiredSysParamsWithoutVote, modifyByte)
	if err != nil {
		log.Debugf(err.Error())
		return nil, err
	}

	return modifyByte, nil
}

//func (s *SysConfigChainCode) getSysParamValByKey(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
//	if len(args) != 1 {
//		jsonResp := "{\"Error\":\" need 1 args (AssetID String)\"}"
//		return nil, fmt.Errorf(jsonResp)
//	}
//	val, err := stub.GetSystemConfig(args[0])
//	//val, err := stub.GetState(args[0])
//	if err != nil {
//		return nil, err
//	}
//	// 并不是所有的配置的string类型
//	jsonResp := "{\"" + args[0] + "\":\"" + string(val) + "\"}"
//	return []byte(jsonResp), nil
//}

func setGlobal(stub shim.ChaincodeStubInterface, tkInfo *SysTokenInfo) error {
	gTkInfo := modules.GlobalTokenInfo{Symbol: tkInfo.Symbol, TokenType: 4, Status: 0, CreateAddr: tkInfo.CreateAddr,
		TotalSupply: tkInfo.TotalSupply, SupplyAddr: "", AssetID: tkInfo.AssetID}
	val, err := json.Marshal(gTkInfo)
	if err != nil {
		return err
	}
	err = stub.PutGlobalState(modules.GlobalPrefix+gTkInfo.Symbol, val)
	return err
}

func getSymbols(stub shim.ChaincodeStubInterface) *SysTokenInfo {
	//
	tkInfo := SysTokenInfo{}
	//TODO
	//tkInfoBytes, _ := stub.GetState(symbolsKey + assetID)
	tkInfoBytes, _ := stub.GetState(modules.DesiredSysParamsWithVote)
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

func setSymbols(stub shim.ChaincodeStubInterface, tkInfo *SysTokenInfo) error {
	val, err := json.Marshal(tkInfo)
	if err != nil {
		return err
	}
	err = stub.PutState(modules.DesiredSysParamsWithVote, val)
	return err
}

// A slice of TopicResult that implements sort.Interface to sort by Value.
type voteResultList []*SysVoteResult

func (p voteResultList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p voteResultList) Len() int           { return len(p) }
func (p voteResultList) Less(i, j int) bool { return p[i].Num > p[j].Num }

// A function to turn a map into a TopicResultList, then sort and return it.
func sortSupportByCount(tpl voteResultList) voteResultList {
	sort.Stable(tpl) //sort.Sort(tpl)
	return tpl
}
