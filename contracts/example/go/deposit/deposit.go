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

//Package deposit implements some functions for deposit contract.
package deposit

import (
	"encoding/json"
	"fmt"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/contracts/shim"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"github.com/palletone/go-palletone/dag/modules"
	"strconv"
	"time"
)

var (
	depositAmountsForJury     uint64
	depositAmountsForMediator uint64
)

type DepositChaincode struct {
}

func (d *DepositChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("*** DepositChaincode system contract init ***")
	depositAmountsForJuryStr, err := stub.GetSystemConfig("DepositAmountForJury")
	if err != nil {
		return shim.Error("GetSystemConfig with DepositAmount error: " + err.Error())
	}
	//转换
	depositAmountsForJury, err = strconv.ParseUint(depositAmountsForJuryStr, 10, 64)
	if err != nil {
		return shim.Error("String transform to uint64 error: " + err.Error())
	}
	fmt.Println("需要的jury保证金数量=", depositAmountsForJury)
	depositAmountsForMediatorStr, err := stub.GetSystemConfig("DepositAmountForMediator")
	if err != nil {
		return shim.Error("GetSystemConfig with DepositAmount error: " + err.Error())
	}
	//转换
	depositAmountsForMediator, err = strconv.ParseUint(depositAmountsForMediatorStr, 10, 64)
	if err != nil {
		return shim.Error("String transform to uint64 error: " + err.Error())
	}
	fmt.Println("需要的mediator保证金数量=", depositAmountsForMediator)
	return shim.Success([]byte("ok"))
}

func (d *DepositChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	funcName, args := stub.GetFunctionAndParameters()
	switch funcName {
	case "DepositWitnessPay":
		//交付保证金
		//handle witness pay
		//void deposit_witness_pay(const witness_object& wit, token_type amount)
		return d.depositWitnessPay(stub, args)
	case "DepositCashback":
		//保证金退还
		//handle cashback rewards
		//void deposit_cashback(const account_object& acct, token_type amount, bool require_vesting = true)
		return d.depositCashback(stub, args)
	case "ForfeitureDeposit":
		//保证金没收
		//void forfeiture_deposit(const witness_object& wit, token_type amount)
		return d.forfeitureDeposit(stub, args)
	default:
		return shim.Error("Invoke error")
	}
}

//交付保证金
//handle witness pay
func (d *DepositChaincode) depositWitnessPay(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	//获取 请求 调用 地址
	invokeFromAddr, err := stub.GetInvokeAddress()
	if err != nil {
		return shim.Error("GetInvokeFromAddr error: " + err.Error())
	}
	fmt.Println("invokeFromAddr address = ", invokeFromAddr.String())
	//获取 请求 ptn 数量
	invokeTokens, err := stub.GetInvokeTokens()
	if err != nil {
		return shim.Error("GetPayToContractPtnTokens error: " + err.Error())
	}
	fmt.Println("invokeTokens ", invokeTokens.Amount)
	fmt.Printf("invokeTokens %#v\n", invokeTokens.Asset)
	stateValue := new(modules.DepositStateValue)
	//获取一下该用户下的账簿情况
	stateValueBytes, err := stub.GetState(invokeFromAddr.String())
	if err != nil {
		return shim.Error("Get account balance from ledger error: " + err.Error())
	}
	//账户不存在，第一次参与
	if stateValueBytes == nil {
		//写入写集
		stateValue.DepositBalance.Amount = invokeTokens.Amount
		stateValue.DepositBalance.Asset = invokeTokens.Asset
		stateValue.Time = time.Now()
		stateValue.Extra = "这是第一次参与陪审团"
		stateValueMarshalBytes, err := json.Marshal(stateValue)
		if err != nil {
			return shim.Error("Marshal valueState error " + err.Error())
		}
		//判断想成为是 Jury 还是 Mediator
		addList(invokeFromAddr, stateValue.DepositBalance.Amount, stub)
		stub.PutState(invokeFromAddr.String(), stateValueMarshalBytes)
		return shim.Success([]byte("ok"))
	}
	//账户已存在，进行信息的更新操作
	err = json.Unmarshal(stateValueBytes, stateValue)
	if err != nil {
		return shim.Error("Unmarshal stateValueBytes error " + err.Error())
	}
	//先判断原来是jury还是mediator，还是什么都不是
	who := whoIs(stateValue.DepositBalance.Amount)
	//who := whoIs(uint64(0))
	//判断资产类型是否一致
	//err = assetIsEqual(invokeTokens.Asset, stateValue.Asset)
	//if err != nil {
	//	return shim.Error("InvokeAsset is not equal with stateAsset error: " + err.Error())
	//}
	//更新stateValue
	stateValue.DepositBalance.Amount += invokeTokens.Amount
	stateValue.Time = time.Now()
	stateValue.Extra = "这是第二次向合约支付保证金，这里的时间是否需要修改为最新的？"
	stateValueMarshalBytes, err := json.Marshal(stateValue)
	if err != nil {
		return shim.Error("Marshal valueState error " + err.Error())
	}
	//判断第二次交保证金的逻辑
	handleMember(who, invokeFromAddr, stateValue.DepositBalance.Amount, stub)
	stub.PutState(invokeFromAddr.String(), stateValueMarshalBytes)
	return shim.Success([]byte("ok"))
}

//保证金退还
//handle cashback rewards
func (d *DepositChaincode) depositCashback(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	//获取 请求 调用 地址
	invokeFromAddr, err := stub.GetInvokeAddress()
	if err != nil {
		return shim.Error("GetInvokeFromAddr error: " + err.Error())
	}
	fmt.Println("invokeFromAddr address ", invokeFromAddr)
	//获取退保证金数量，将 string 转 uint64
	ptnAccount, err := strconv.ParseUint(args[1], 10, 64)
	if err != nil {
		return shim.Error("String transform to uint64 error: " + err.Error())
	}
	fmt.Println("args[1] ", ptnAccount)
	stateValueBytes, err := stub.GetState(invokeFromAddr.String())
	if err != nil {
		return shim.Error("Get account balance from ledger error: " + err.Error())
	}
	if stateValueBytes == nil {
		return shim.Error("Your account does not exist.")
	}
	stateValue := new(modules.DepositStateValue)
	err = json.Unmarshal(stateValueBytes, stateValue)
	if err != nil {
		return shim.Error("Unmarshal stateValueBytes error: " + err.Error())
	}
	if stateValue.DepositBalance.Amount < ptnAccount {
		return shim.Error("Your delivery amount with ptn token is insufficient.")
	}
	//判断是 Jury 还是 Mediator
	who := whoIs(stateValue.DepositBalance.Amount)
	//调用从合约把token转到地址
	err = stub.PayOutToken(invokeFromAddr.String(), stateValue.DepositBalance.Asset, ptnAccount, 0)
	if err != nil {
		return shim.Error("PayOutToken error: " + err.Error())
	}
	//更新
	stateValue.DepositBalance.Amount -= ptnAccount
	stateValue.Time = time.Now()
	stateValue.Extra = "这是退出保证金，可能只退一部分钱，时间是否需要修改？"
	stateValueMarshalBytes, err := json.Marshal(stateValue)
	if err != nil {
		return shim.Error("Marshal valueState error: " + err.Error())
	}
	//判断退款后是 Jury 还是 Mediator 还是  都不是（即移除列表）
	handleMember(who, invokeFromAddr, stateValue.DepositBalance.Amount, stub)
	stub.PutState(invokeFromAddr.String(), stateValueMarshalBytes)
	return shim.Success([]byte("ok"))
}

//保证金没收
//handle forfeiture deposit
func (d DepositChaincode) forfeitureDeposit(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	//参数是陪审员的账户和罚没数量
	if len(args) != 2 {
		return shim.Error("Input error: need two arg (witnessAddr and amount)")
	}
	//获取该账户的账本信息
	stateValueBytes, err := stub.GetState(args[0])
	if err != nil {
		return shim.Error("Get account balance from ledger error: " + err.Error())
	}
	//if stateValueBytes == nil {
	//	return shim.Error("you have not depositWitnessPay for deposit.")
	//}
	stateValue := new(modules.DepositStateValue)
	err = json.Unmarshal(stateValueBytes, stateValue)
	if err != nil {
		return shim.Error("unmarshal accBalByte error " + err.Error())
	}
	//获取没收保证金数量，将 string 转 uint64
	ptnAccount, _ := strconv.ParseUint(args[1], 10, 64)
	//if err != nil {
	//	return shim.Error("String transform to uint64 error: " + err.Error())
	//}
	//if stateValue.DepositBalance.Amount < ptnAccount {
	//	return shim.Error("Your amount balance does not enough.")
	//}
	if stateValue.DepositBalance.Amount < ptnAccount {
		return shim.Error("Forfeiture too many.")
	}
	//判断是 Jury 还是 Mediator
	who := whoIs(stateValue.DepositBalance.Amount)
	//获取基金会地址
	foundationAddress, err := stub.GetSystemConfig("FoundationAddress")
	if err != nil {
		return shim.Error(err.Error())
	}
	//fmt.Println("foundationAddress", foundationAddress)
	//调用从合约把token转到地址
	err = stub.PayOutToken(foundationAddress, stateValue.DepositBalance.Asset, ptnAccount, 0)
	if err != nil {
		return shim.Error("PayOutToken error: " + err.Error())
	}
	//写入写集
	stateValue.DepositBalance.Amount -= ptnAccount
	stateValue.Time = time.Now()
	stateValue.Extra = "这是退出保证金，可能只退一部分钱，时间是否需要修改？"
	stateValueMarshalBytes, err := json.Marshal(stateValue)
	if err != nil {
		return shim.Error("Marshal valueState error " + err.Error())
	}
	addr, _ := common.StringToAddress(args[0])
	//判断罚款后是 Jury 还是 Mediator 还是  都不是（即移除列表）
	handleMember(who, addr, stateValue.DepositBalance.Amount, stub)
	stub.PutState(args[0], stateValueMarshalBytes)
	return shim.Success([]byte("ok"))
}

//func assetIsEqual(invokeAsset, stateAsset modules.Asset) error {
//	if invokeAsset != stateAsset {
//		return fmt.Errorf("asset is not equal")
//	}
//	return nil
//}
