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

package vote

import (
	"encoding/json"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/contracts/shim"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	dm "github.com/palletone/go-palletone/dag/modules"
)

const symbolsKey = "symbols"

type Vote struct {
}

//one topic
type VoteTopic struct {
	TopicTitle    string
	SelectOptions []string
	SelectMax     uint64
}

//topic support result
type TopicSupports struct {
	TopicTitle    string
	SupportCounts map[string]uint64
	SelectMax     uint64
	//SelectOptionsNum  uint64
}

//vote token information
type TokenInfo struct {
	Name        string
	Symbol      string
	CreateAddr  string
	VoteType    byte
	TotalSupply uint64
	VoteEndTime time.Time
	VoteContent []byte
	AssetID     dm.IDType16
}

//one user's support
type SupportRequest struct {
	TopicIndex    uint64
	SelectOptions []string
}

type Symbols struct {
	TokenInfos map[string]TokenInfo `json:"tokeninfos"`
}

func (v *Vote) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

func (v *Vote) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	f, args := stub.GetFunctionAndParameters()

	switch f {
	case "createToken":
		return createToken(args, stub)
	case "support":
		return support(args, stub)
	case "getVoteResult":
		return getVoteResult(args, stub)
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
	if len(args) < 6 {
		return shim.Error("need 4 args (Name,Symbol,VoteType,TotalSupply,VoteEndTime,VoteContentJson)")
	}

	//==== convert params to token information
	var vt dm.VoteToken
	//name symbol
	vt.Name = args[0]
	vt.Symbol = strings.ToUpper(args[1])
	if vt.Symbol == "PTN" {
		jsonResp := "{\"Error\":\"Can't use PTN\"}"
		return shim.Success([]byte(jsonResp))
	}
	if len(vt.Symbol) > 5 {
		jsonResp := "{\"Error\":\"Symbol must less than 5 characters\"}"
		return shim.Success([]byte(jsonResp))
	}

	//vote type
	if args[2] == "0" {
		vt.VoteType = byte(0)
	} else if args[2] == "1" {
		vt.VoteType = byte(1)
	} else if args[2] == "2" {
		vt.VoteType = byte(2)
	} else {
		jsonResp := "{\"Error\":\"Only string, 0 or 1 or 2\"}"
		return shim.Success([]byte(jsonResp))
	}
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
	vt.TotalSupply = totalSupply
	//VoteEndTime
	VoteEndTime, err := time.Parse("2006-01-02 15:04:05", args[4])
	if err != nil {
		jsonResp := "{\"Error\":\"No vote end time\"}"
		return shim.Success([]byte(jsonResp))
	}
	if VoteEndTime.Before(time.Now().UTC()) {
		jsonResp := "{\"Error\":\"Invalid time\"}"
		return shim.Success([]byte(jsonResp))
	}
	vt.VoteEndTime = VoteEndTime
	//VoteContent
	var voteTopics []VoteTopic
	err = json.Unmarshal([]byte(args[5]), &voteTopics)
	if err != nil {
		jsonResp := "{\"Error\":\"VoteContent format invalid\"}"
		return shim.Success([]byte(jsonResp))
	}
	//init support
	var supports []TopicSupports
	for _, oneTopic := range voteTopics {
		var oneSupport TopicSupports
		oneSupport.SupportCounts = make(map[string]uint64)
		oneSupport.TopicTitle = oneTopic.TopicTitle
		for _, oneOption := range oneTopic.SelectOptions {
			oneSupport.SupportCounts[oneOption] = 0
		}
		//oneResult.SelectOptionsNum = uint64(len(oneRequest.SelectOptions))
		oneSupport.SelectMax = oneTopic.SelectMax
		supports = append(supports, oneSupport)
	}
	voteContentJson, err := json.Marshal(supports)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to generate voteContent Json\"}"
		return shim.Error(jsonResp)
	}
	vt.VoteContent = voteContentJson

	//check name is only or not
	symbols, err := getSymbols(stub)
	if _, ok := symbols.TokenInfos[vt.Symbol]; ok {
		jsonResp := "{\"Error\":\"The symbol have been used\"}"
		return shim.Success([]byte(jsonResp))
	}

	//convert to json
	createJson, err := json.Marshal(vt)
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
	err = stub.DefineToken(byte(dm.AssetType_VoteToken), createJson, createAddr)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to call stub.DefineToken\"}"
		return shim.Error(jsonResp)
	}

	//last put state
	txid := stub.GetTxID()
	assetID, _ := dm.NewAssetId(vt.Symbol, dm.AssetType_VoteToken,
		0, common.Hex2Bytes(txid[2:]))
	info := TokenInfo{vt.Name, vt.Symbol, createAddr, vt.VoteType, totalSupply,
		VoteEndTime, voteContentJson, assetID}
	symbols.TokenInfos[vt.Symbol] = info

	err = setSymbols(symbols, stub)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to set symbols\"}"
		return shim.Error(jsonResp)
	}
	return shim.Success(createJson) //test
}

func support(args []string, stub shim.ChaincodeStubInterface) pb.Response {
	//params check
	if len(args) < 2 {
		return shim.Error("need 2 args (Symbol,SupportRequestJson)")
	}

	//symbol
	symbol := strings.ToUpper(args[0])
	//check name is exist or not
	symbols, err := getSymbols(stub)
	if _, ok := symbols.TokenInfos[symbol]; !ok {
		jsonResp := "{\"Error\":\"Token not exist\"}"
		return shim.Success([]byte(jsonResp))
	}

	//parse support requests
	var supportRequests []SupportRequest
	err = json.Unmarshal([]byte(args[1]), &supportRequests)
	if err != nil {
		jsonResp := "{\"Error\":\"SupportRequestJson format invalid\"}"
		return shim.Success([]byte(jsonResp))
	}
	//get token information
	tokenInfo := symbols.TokenInfos[symbol]
	var topicSupports []TopicSupports
	err = json.Unmarshal(tokenInfo.VoteContent, &topicSupports)
	if err != nil {
		jsonResp := "{\"Error\":\"Results format invalid, Error!!!\"}"
		return shim.Success([]byte(jsonResp))
	}
	//check time
	if tokenInfo.VoteEndTime.Before(time.Now().UTC()) {
		jsonResp := "{\"Error\":\"Vote is over\"}"
		return shim.Success([]byte(jsonResp))
	}
	//check token
	invokeTokens, err := stub.GetInvokeTokens()
	if err != nil {
		jsonResp := "{\"Error\":\"GetInvokeTokens failed\"}"
		return shim.Success([]byte(jsonResp))
	}
	voteNum := uint64(0)
	for i := 0; i < len(invokeTokens); i++ {
		if invokeTokens[i].Asset.AssetId == tokenInfo.AssetID && invokeTokens[i].Address == "P1111111111111111111114oLvT2" {
			voteNum += invokeTokens[i].Amount
		}
	}
	if voteNum == 0 { //no vote token
		jsonResp := "{\"Error\":\"Vote token empty\"}"
		return shim.Success([]byte(jsonResp))
	}
	if voteNum < uint64(len(supportRequests)) { //vote token more than request
		jsonResp := "{\"Error\":\"Vote token more than support request\"}"
		return shim.Success([]byte(jsonResp))
	}

	//save support
	indexHistory := make(map[uint64]uint8)
	indexRepeat := false
	for _, oneSupport := range supportRequests {
		selectIndex := oneSupport.TopicIndex - 1
		if _, ok := indexHistory[selectIndex]; ok { //check select repeat
			indexRepeat = true
			break
		}
		indexHistory[selectIndex] = 1
		if selectIndex < uint64(len(topicSupports)) { //1.check index, must not out of total
			if uint64(len(oneSupport.SelectOptions)) <= topicSupports[selectIndex].SelectMax { //2.check one select's options, must not out of select's max
				for _, oneSelectOption := range oneSupport.SelectOptions {
					if _, ok := topicSupports[selectIndex].SupportCounts[oneSelectOption]; ok { //3.check select option, must be real select options
						topicSupports[selectIndex].SupportCounts[oneSelectOption] += 1
					}
				}
			}
		}
	}
	if indexRepeat {
		jsonResp := "{\"Error\":\"Repeat index of select option \"}"
		return shim.Error(jsonResp)
	}
	voteContentJson, err := json.Marshal(topicSupports)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to generate voteContent Json\"}"
		return shim.Error(jsonResp)
	}
	tokenInfo.VoteContent = voteContentJson

	//save token information
	symbols.TokenInfos[symbol] = tokenInfo
	err = setSymbols(symbols, stub)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to set symbols\"}"
		return shim.Error(jsonResp)
	}
	return shim.Success([]byte("")) //test
}

type TokenIDInfo struct {
	Symbol         string
	CreateAddr     string
	TotalSupply    uint64
	SupportResults []SupportResult
	AssetID        string
}
type SupportResult struct {
	TopicIndex   uint64
	TopicTitle   string
	TopicResults []TopicResult
}

// A data structure to hold a key/value pair.
type TopicResult struct {
	TopicOption      string
	OptionSupportNum uint64
}

// A slice of TopicResult that implements sort.Interface to sort by Value.
type TopicResultList []TopicResult

func (p TopicResultList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p TopicResultList) Len() int           { return len(p) }
func (p TopicResultList) Less(i, j int) bool { return p[i].OptionSupportNum > p[j].OptionSupportNum }

// A function to turn a map into a TopicResultList, then sort and return it.
func sortSupportByCount(selectCounts map[string]uint64) TopicResultList {
	var tpl TopicResultList
	for k, v := range selectCounts {
		op := TopicResult{k, v}
		tpl = append(tpl, op)
	}
	sort.Stable(tpl) //sort.Sort(tpl)
	return tpl
}
func getVoteResult(args []string, stub shim.ChaincodeStubInterface) pb.Response {
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

	//get token information
	tokenInfo := symbols.TokenInfos[symbol]
	var topicSupports []TopicSupports
	err = json.Unmarshal(tokenInfo.VoteContent, &topicSupports)
	if err != nil {
		jsonResp := "{\"Error\":\"Results format invalid, Error!!!\"}"
		return shim.Success([]byte(jsonResp))
	}

	//calculate result
	var supportResults []SupportResult
	for i, oneTopicSupport := range topicSupports {
		var oneResult SupportResult
		oneResult.TopicIndex = uint64(i) + 1
		oneResult.TopicTitle = oneTopicSupport.TopicTitle
		oneResultSort := sortSupportByCount(oneTopicSupport.SupportCounts)
		for i := uint64(0); i < oneTopicSupport.SelectMax; i++ {
			oneResult.TopicResults = append(oneResult.TopicResults, oneResultSort[i])
		}
		supportResults = append(supportResults, oneResult)
	}

	//token
	asset := symbols.TokenInfos[symbol].AssetID
	tkID := TokenIDInfo{Symbol: symbol, CreateAddr: symbols.TokenInfos[symbol].CreateAddr,
		TotalSupply: symbols.TokenInfos[symbol].TotalSupply, SupportResults: supportResults, AssetID: asset.ToAssetId()}

	//return json
	tkJson, err := json.Marshal(tkID)
	if err != nil {
		return shim.Success([]byte(err.Error()))
	}
	return shim.Success(tkJson) //test
}
