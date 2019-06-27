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
	"github.com/palletone/go-palletone/contracts/shim"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"github.com/palletone/go-palletone/dag/modules"
	"strconv"
)

//  质押PTN投票mediator
func normalNodePledgeVote(stub shim.ChaincodeStubInterface, args []string) pb.Response {
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
	err = pledgeDeposit(stub, invokeAddr, pledgeAmount)
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

//  普通节点修改质押mediator
//func normalNodeChangeVote(stub shim.ChaincodeStubInterface, args []string) pb.Response {
//
//	//  获取请求地址
//	inAddr, err := stub.GetInvokeAddress()
//	if err != nil {
//		return shim.Error(err.Error())
//	}
//	////  获取是否存在
//	//nor, err := getNor(stub, inAddr.String())
//	//if err != nil {
//	//	return shim.Error(err.Error())
//	//}
//	//if nor == nil {
//	//	return shim.Error("node was nil")
//	//}
//
//	//mediatorAddr := args[0]
//	//nor.MediatorAddr = mediatorAddr
//	//  保存
//	err = saveMediatorVote(stub, inAddr.String(), args)
//	if err != nil {
//		return shim.Error(err.Error())
//	}
//	return shim.Success(nil)
//}

//  普通节点申请提取PTN
func normalNodeExtractVote(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("need 1 arg, withdraw Dao amount")
	}
	//  获取请求地址
	inAddr, err := stub.GetInvokeAddress()
	if err != nil {
		return shim.Error(err.Error())
	}
	//  获取是否存在
	// nor, err := getNor(stub, inAddr.String())
	// if err != nil {
	// 	return shim.Error(err.Error())
	// }
	// if nor == nil {
	// 	return shim.Error("node was nil")
	// }
	amount := args[0]
	ptnAccount, err := strconv.ParseUint(amount, 10, 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	// if ptnAccount > nor.AmountAsset.Amount {
	// 	return shim.Error("PTN was not enough")
	// }
	//  保存质押提取
	err = savePledgeWithdrawRecord(stub, modules.NewAddressAmount(inAddr.String(), ptnAccount))
	//extPtnLis, err := getExtPtn(stub)
	//if err != nil {
	//	return shim.Error(err.Error())
	//}
	//if extPtnLis == nil {
	//	extPtnLis = make(map[string]*extractPtn)
	//}
	//if extPtn, ok := extPtnLis[inAddr.String()]; ok {
	//	extPtn.Amount = ptnAccount
	//} else {
	//	extPtnLis[inAddr.String()] = &extractPtn{Amount: ptnAccount}
	//}
	////fees, err := stub.GetInvokeFees()
	////if err != nil {
	////	return shim.Error(err.Error())
	////}
	////extPtn.Time = TimeStr()
	////  保存
	//err = saveExtPtn(stub, extPtnLis)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}
