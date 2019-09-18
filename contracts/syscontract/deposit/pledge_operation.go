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

package deposit

import (
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/math"
	"github.com/palletone/go-palletone/contracts/syscontract"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"regexp"
	"strconv"
	"strings"

	"encoding/json"
	"github.com/palletone/go-palletone/contracts/shim"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"github.com/palletone/go-palletone/dag/modules"
)

//  质押PTN
func processPledgeDeposit(stub shim.ChaincodeStubInterface) pb.Response {
	//  获取是否是保证金合约
	invokeTokens, err := isContainDepositContractAddr(stub)
	if err != nil {
		return shim.Error(err.Error())
	}
	pledgeAmount := invokeTokens.Amount
	//  获取请求地址
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		return shim.Error(err.Error())
	}
	//  添加进入质押记录
	err = pledgeDepositRep(stub, invokeAddr, pledgeAmount)
	if err != nil {
		return shim.Error(err.Error())
	}
	//记录投票情况
	//err = saveMediatorVote(stub, invokeAddr.String(), args)
	//if err != nil {
	//	return shim.Error(err.Error())
	//}
	return shim.Success(nil)
}

//  每天计算各节点收益
func handlePledgeReward(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 0 {
		return shim.Error("need 0 args")
	}
	gp, err := stub.GetSystemConfig()
	if err != nil {
		return shim.Error(err.Error())
	}
	depositDailyReward := gp.ChainParameters.PledgeDailyReward

	err = handleRewardAllocation(stub, depositDailyReward)
	if err != nil {
		return shim.Error(err.Error())
	}
	if depositDailyReward > 0 {
		//增发到合约
		log.Debugf("Create coinbase %d to pledge contract", depositDailyReward)
		err = stub.SupplyToken(dagconfig.DagConfig.GetGasToken().Bytes(),
			[]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, depositDailyReward, syscontract.DepositContractAddress.String())
		if err != nil {
			return shim.Error(err.Error())
		}
	}
	return shim.Success(nil)

}

//  普通节点申请提取PTN
func processPledgeWithdraw(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("need 1 arg, withdraw Dao amount")
	}
	//  获取请求地址
	inAddr, err := stub.GetInvokeAddress()
	if err != nil {
		return shim.Error(err.Error())
	}

	amount := args[0]
	ptnAccount := uint64(0)
	if strings.ToLower(amount) == "all" {
		ptnAccount = math.MaxUint64
	} else {
		ptnAccount, err = strconv.ParseUint(amount, 10, 64)
		if err != nil {
			return shim.Error(err.Error())
		}
	}
	//  保存质押提取
	err = pledgeWithdrawRep(stub, inAddr, ptnAccount)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

func queryPledgeStatusByAddr(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("need 1 arg, Address")
	}
	status, err := getPledgeStatus(stub, args[0])
	if err != nil {
		return shim.Error(err.Error())
	}
	pjson := convertPledgeStatus2Json(status)
	data, _ := json.Marshal(pjson)
	return shim.Success(data)
}

func queryAllPledgeHistory(stub shim.ChaincodeStubInterface) pb.Response {

	history, err := getAllPledgeRewardHistory(stub)
	if err != nil {
		return shim.Error(err.Error())
	}
	data, _ := json.Marshal(history)
	return shim.Success(data)
}

func queryPledgeList(stub shim.ChaincodeStubInterface) pb.Response {
	list, err := getLastPledgeList(stub)
	if err != nil {
		return shim.Error(err.Error())
	}
	result, _ := json.Marshal(list)
	return shim.Success(result)
}

func queryPledgeListByDate(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	date := args[0]
	reg := regexp.MustCompile(`[\d]{8}`)
	if !reg.Match([]byte(date)) {
		return shim.Error("must use YYYYMMDD format")
	}
	list, err := getPledgeListByDate(stub, date)
	if err != nil {
		return shim.Error(err.Error())
	}
	result, _ := json.Marshal(list)
	return shim.Success(result)
}

func convertPledgeStatus2Json(p *modules.PledgeStatus) *modules.PledgeStatusJson {
	data := &modules.PledgeStatusJson{}
	gasToken := dagconfig.DagConfig.GetGasToken().ToAsset()
	data.NewDepositAmount = gasToken.DisplayAmount(p.NewDepositAmount)
	data.PledgeAmount = gasToken.DisplayAmount(p.PledgeAmount)
	data.OtherAmount = gasToken.DisplayAmount(p.OtherAmount)
	if p.WithdrawApplyAmount == math.MaxUint64 {
		data.WithdrawApplyAmount = "all"
	} else {
		data.WithdrawApplyAmount = gasToken.DisplayAmount(p.WithdrawApplyAmount).String()
	}
	return data
}
