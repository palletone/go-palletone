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
	depositAmountsForJury      uint64
	depositAmountsForMediator  uint64
	depositAmountsForDeveloper uint64
)

type DepositChaincode struct {
}

func (d *DepositChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("*** DepositChaincode system contract init ***")
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
	depositAmountsForDeveloperStr, err := stub.GetSystemConfig("DepositAmountForJury")
	if err != nil {
		return shim.Error("GetSystemConfig with DepositAmount error: " + err.Error())
	}
	//转换
	depositAmountsForDeveloper, err = strconv.ParseUint(depositAmountsForDeveloperStr, 10, 64)
	if err != nil {
		return shim.Error("String transform to uint64 error: " + err.Error())
	}
	fmt.Println("需要的Developer保证金数量=", depositAmountsForDeveloper)
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
	//第一个参数：合约地址；第二个参数：保证金；第三个参数：角色（Mediator Jury ContractDeveloper)
	if len(args) < 3 {
		return shim.Error("Input parameter error,need three parameters.")
	}
	//获取 请求 调用 地址
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		return shim.Error("GetInvokeFromAddr error: " + err.Error())
	}
	fmt.Println("invokeFromAddr address = ", invokeAddr.String())
	//获取 请求 ptn 数量
	invokeTokens, err := stub.GetInvokeTokens()
	if err != nil {
		return shim.Error("GetPayToContractPtnTokens error: " + err.Error())
	}
	//获取退保证金数量，将 string 转 uint64
	ptnAccount, _ := strconv.ParseUint(args[1], 10, 64)
	invokeTokens.Amount = ptnAccount
	fmt.Println("invokeTokens ", invokeTokens.Amount)
	fmt.Printf("invokeTokens %#v\n", invokeTokens.Asset)
	role := args[2]
	var res pb.Response
	switch {
	case role == "Mediator":
		//
		res = d.handleMediatorDepositWitnessPay(stub, invokeAddr, invokeTokens)
	case role == "Jury":
		//
		res = d.handleJuryDepositWitnessPay(stub, invokeAddr, invokeTokens)
	case role == "Developer":
		//
		res = d.handleDeveloperDepositWitnessPay(stub, invokeAddr, invokeTokens)
	default:
		return shim.Error("role error.")
	}
	return res
}

//处理 Mediator
func (d *DepositChaincode) handleMediatorDepositWitnessPay(stub shim.ChaincodeStubInterface, invokeAddr common.Address, invokeTokens *modules.InvokeTokens) pb.Response {
	//第一步：判断数据库里是否存在该节点信息
	//	是：证明第一次交或者以前交过但是被退出过列表
	//		第二步：判断提交保证金是否足够
	//			是：
	//				第三步：加入列表
	//				第四步：更新数据库
	//			否：报错
	//	否：证明是在规定范围之上增加保证金数量
	//		第二步：更新数据库
	stateValue := new(modules.DepositStateValue)
	//获取一下该用户下的账簿情况
	stateValueBytes, err := stub.GetState(invokeAddr.String())
	if err != nil {
		return shim.Error("Get account balance from ledger error: " + err.Error())
	}
	//账户不存在，第一次参与
	if stateValueBytes == nil {
		if invokeTokens.Amount < depositAmountsForMediator {
			return shim.Success([]byte("Payment amount is insufficient."))
		}
		addList("Mediator", invokeAddr, stub)
		//写入写集
		stateValue.DepositBalance.Amount = invokeTokens.Amount
		stateValue.DepositBalance.Asset = invokeTokens.Asset
		stateValue.Time = time.Now() //第一次交付保证金的时间，并且加入列表
		stateValue.Extra = "这是第一次参与或者以前交过但是被退出过列表"
	} else {
		err = json.Unmarshal(stateValueBytes, stateValue)
		if err != nil {
			return shim.Error("Unmarshal stateValueBytes error " + err.Error())
		}
		//判断资产类型是否一致
		//err = assetIsEqual(invokeTokens.Asset, stateValue.Asset)
		//if err != nil {
		//	return shim.Error("InvokeAsset is not equal with stateAsset error: " + err.Error())
		//}
		//账户已存在，进行信息的更新操作
		//更新stateValue
		stateValue.DepositBalance.Amount += invokeTokens.Amount
		//stateValue.Time = time.Now()
		stateValue.Extra = "这是第二次向合约支付保证金"
	}
	stateValueMarshalBytes, err := json.Marshal(stateValue)
	if err != nil {
		return shim.Error("Marshal valueState error " + err.Error())
	}
	stub.PutState(invokeAddr.String(), stateValueMarshalBytes)
	return shim.Success([]byte("ok"))
}

func (d *DepositChaincode) handleJuryOrDeveloperDepositWitnessPay(stub shim.ChaincodeStubInterface, invokeAddr common.Address, invokeTokens *modules.InvokeTokens, who string) pb.Response {
	//第一步：判断数据库里是否存在该节点信息
	//	是：证明第一次交或者以前交过但是被退出过列表
	//		第二步：判断是 Jury 还是 Developer
	//			Jury:
	//				第三步：判断保证金是否足够
	//					否：
	//						第四部：更新数据库
	//					是：
	//						第四步：加入列表
	//						第五步：更新数据库
	//			Developer:
	//					同理 Jury
	//	否：证明该节点已在数据库里，但是不知道是否在列表中
	//		Jury and Developer 一样逻辑
	//		第二步：判断是否在列表中
	//			是：证明在担任成员之上，增加保证金
	//				第三步：更新数据库
	//			否：
	//				第三步：判断第二次交了之后是否能达到规定数量
	//					是：
	//						第四步：加入列表
	//						第五步：更新数据库
	//					否：
	//						第四步：更新数据库
	isJury := false
	isDeveloper := false
	stateValue := new(modules.DepositStateValue)
	//获取一下该用户下的账簿情况
	stateValueBytes, err := stub.GetState(invokeAddr.String())
	if err != nil {
		return shim.Error("Get account balance from ledger error: " + err.Error())
	}
	//账户不存在，第一次参与
	if stateValueBytes == nil {
		//判断
		if who == "Jury" {
			//fmt.Println("la", invokeTokens.Amount)
			//fmt.Println(depositAmountsForJury)
			if invokeTokens.Amount >= depositAmountsForJury {
				addList("Jury", invokeAddr, stub)
			}
		} else if who == "Developer" {
			if invokeTokens.Amount >= depositAmountsForDeveloper {
				addList("Developer", invokeAddr, stub)
			}
		}
		//写入写集
		stateValue.DepositBalance.Amount = invokeTokens.Amount
		stateValue.DepositBalance.Asset = invokeTokens.Asset
		stateValue.Time = time.Now()
		stateValue.Extra = "这是第一次参与"
		//stub.PutState(invokeAddr.String(), stateValueMarshalBytes)
		//return shim.Success([]byte("ok"))
	} else {
		//账户已存在，进行信息的更新操作
		err = json.Unmarshal(stateValueBytes, stateValue)
		if err != nil {
			return shim.Error("Unmarshal stateValueBytes error " + err.Error())
		}
		if who == "Jury" {
			if stateValue.DepositBalance.Amount >= depositAmountsForJury {
				//原来就是jury
				isJury = true
			}
		} else if who == "Developer" {
			if stateValue.DepositBalance.Amount >= depositAmountsForDeveloper {
				//原来就是 Developer
				isDeveloper = true
			}
		}
		//更新stateValue
		stateValue.DepositBalance.Amount += invokeTokens.Amount
		stateValue.Extra = "这是第二次向合约支付保证金"
	}
	//判断资产类型是否一致
	//err = assetIsEqual(invokeTokens.Asset, stateValue.Asset)
	//if err != nil {
	//	return shim.Error("InvokeAsset is not equal with stateAsset error: " + err.Error())
	//}
	if !isJury && who == "Jury" {
		//判断交了保证金后是否超过了jury
		if stateValue.DepositBalance.Amount >= depositAmountsForJury {
			addList("Jury", invokeAddr, stub)
			stateValue.Time = time.Now()
		}
	}
	if !isDeveloper && who == "Developer" {
		//判断交了保证金后是否超过了jury
		if stateValue.DepositBalance.Amount >= depositAmountsForDeveloper {
			addList("Developer", invokeAddr, stub)
			stateValue.Time = time.Now()
		}
	}
	stateValueMarshalBytes, err := json.Marshal(stateValue)
	if err != nil {
		return shim.Error("Marshal valueState error " + err.Error())
	}
	stub.PutState(invokeAddr.String(), stateValueMarshalBytes)
	return shim.Success([]byte("ok"))
}

//处理 Jury
func (d *DepositChaincode) handleJuryDepositWitnessPay(stub shim.ChaincodeStubInterface, invokeAddr common.Address, invokeTokens *modules.InvokeTokens) pb.Response {
	return d.handleJuryOrDeveloperDepositWitnessPay(stub, invokeAddr, invokeTokens, "Jury")
}

//处理 ContractDeveloper
func (d *DepositChaincode) handleDeveloperDepositWitnessPay(stub shim.ChaincodeStubInterface, invokeAddr common.Address, invokeTokens *modules.InvokeTokens) pb.Response {
	return d.handleJuryOrDeveloperDepositWitnessPay(stub, invokeAddr, invokeTokens, "Developer")
}

//保证金退还
//handle cashback rewards
func (d *DepositChaincode) depositCashback(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	//第一个参数：数量；第二个参数：角色（角色（Mediator Jury ContractDeveloper)
	if len(args) < 2 {
		return shim.Error("Input parameter error,need two parameters.")
	}
	//获取 请求 调用 地址
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		return shim.Error("GetInvokeFromAddr error: " + err.Error())
	}
	fmt.Println("invokeAddr address ", invokeAddr)
	//获取退保证金数量，将 string 转 uint64
	ptnAccount, err := strconv.ParseUint(args[0], 10, 64)
	if err != nil {
		return shim.Error("String transform to uint64 error: " + err.Error())
	}
	fmt.Println("ptnAccount  args[0] ", ptnAccount)
	asset := modules.Asset{
		modules.PTNCOIN,
		modules.PTNCOIN,
		0,
	}
	invokeTokens := &modules.InvokeTokens{
		Amount: ptnAccount,
		Asset:  asset,
	}
	//
	//第一步：先获取数据库信息
	stateValueBytes, err := stub.GetState(invokeAddr.String())
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
	//第二步：比较退款数量和数据库数量
	//Asset判断
	//数量比较
	if stateValue.DepositBalance.Amount < invokeTokens.Amount {
		return shim.Error("Your delivery amount with ptn token is insufficient.")
	}
	//获取角色
	role := args[1]
	var res pb.Response
	switch {
	case role == "Mediator":
		//
		res = d.handleMediatorDepositCashback(stub, invokeAddr, invokeTokens, stateValue)
	case role == "Jury":
		//
		res = d.handleJuryDepositCashback(stub, invokeAddr, invokeTokens, stateValue)
	case role == "Developer":
		//
		res = d.handleDeveloperDepositCashback(stub, invokeAddr, invokeTokens, stateValue)
	default:
		return shim.Error("role error.")
	}
	return res
}

func (d *DepositChaincode) handleJuryOrDeveloperDepositCashback(who string, stub shim.ChaincodeStubInterface, invokeAddr common.Address, invokeTokens *modules.InvokeTokens, stateValue *modules.DepositStateValue) pb.Response {
	//第一步：从合约把token转到地址
	//第二步：判断是 Jury 还是 Mediator
	//	Jury:
	//		第三步：判断原来是否在列表中
	//			是：
	//				第四步：判断退保证金后，余额是否还在规定数量中
	//					否：
	//						第五步：移除列表
	//						第六步：更新数据库
	//			否：
	//				第四步：更新列表
	//	Developer:
	//		同理 Jury
	var err error
	//调用从合约把token转到地址
	err = stub.PayOutToken(invokeAddr.String(), stateValue.DepositBalance.Asset, invokeTokens.Amount, 0)
	if err != nil {
		return shim.Error("PayOutToken error: " + err.Error())
	}
	if who == "Jury" {
		//需要判断是在在列表中
		isJury := false
		if stateValue.DepositBalance.Amount >= depositAmountsForJury {
			isJury = true
		}
		stateValue.DepositBalance.Amount -= invokeTokens.Amount
		if isJury {
			//判断退保证金后，是否还在规定数量之内
			if stateValue.DepositBalance.Amount < depositAmountsForJury {
				//移除出列表
				handleMember("Jury", invokeAddr, stub)
				stateValue.Time = time.Now()
				stateValue.Extra = "这是退出保证金，且不在列表中了"
			} else {
				stateValue.Extra = "这是退出保证金，但余额还够规定范围之内"
			}
		} else {
			//一开始就不在列表中
			stateValue.Extra = "一开始就不在列表中,退剩余余额而已"
		}
	} else if who == "Developer" {
		//需要判断是在在列表中
		isDev := false
		if stateValue.DepositBalance.Amount >= depositAmountsForDeveloper {
			isDev = true
		}
		stateValue.DepositBalance.Amount -= invokeTokens.Amount
		if isDev {
			//判断退保证金后，是否还在规定数量之内
			if stateValue.DepositBalance.Amount < depositAmountsForDeveloper {
				//移除出列表
				handleMember("Developer", invokeAddr, stub)
				stateValue.Time = time.Now()
				stateValue.Extra = "这是退出保证金，且不在列表中了"
			} else {
				stateValue.Extra = "这是退出保证金，但余额还够规定范围之内"
			}
		} else {
			//一开始就不在列表中
			stateValue.Extra = "一开始就不在列表中,退剩余余额而已"
		}
	}
	stateValueMarshalBytes, err := json.Marshal(stateValue)
	if err != nil {
		return shim.Error("Marshal valueState error " + err.Error())
	}
	stub.PutState(invokeAddr.String(), stateValueMarshalBytes)
	return shim.Success([]byte("ok"))
}

func (d *DepositChaincode) handleJuryDepositCashback(stub shim.ChaincodeStubInterface, invokeAddr common.Address, invokeTokens *modules.InvokeTokens, stateValue *modules.DepositStateValue) pb.Response {
	return d.handleJuryOrDeveloperDepositCashback("Jury", stub, invokeAddr, invokeTokens, stateValue)
}
func (d *DepositChaincode) handleDeveloperDepositCashback(stub shim.ChaincodeStubInterface, invokeAddr common.Address, invokeTokens *modules.InvokeTokens, stateValue *modules.DepositStateValue) pb.Response {
	return d.handleJuryOrDeveloperDepositCashback("Developer", stub, invokeAddr, invokeTokens, stateValue)
}

func (d *DepositChaincode) handleMediatorDepositCashback(stub shim.ChaincodeStubInterface, invokeAddr common.Address, invokeTokens *modules.InvokeTokens, stateValue *modules.DepositStateValue) pb.Response {
	//第一步：判断是否全部退还保证金
	//是
	//	第二步：判断是否已超过规定周期
	//	第三步：从合约把保证金转到该请求地址
	//	第四步：移除列表
	//	第五步：更新数据库
	//否
	//	第二步：判断退款后，余额是否超过规定数量
	//		否
	//			出错
	//		是
	//			正常退款
	//			第三步：从合约把保证金转到该请求地址
	//			第四步：更新数据库
	var err error
	//规定mediator 退款要么全部退，要么退款后，剩余数量在mediator数量范围内，
	stateValue.DepositBalance.Amount -= invokeTokens.Amount
	//第三步：判断是否全部退
	if stateValue.DepositBalance.Amount == 0 {
		startTime := stateValue.Time.Unix()
		endTime := time.Now().Unix()
		//第四步：判断是否已超过规定周期
		if endTime-startTime >= 10000 { //TODO
			//调用从合约把token转到地址
			//第五步：从合约把token转到地址
			err = stub.PayOutToken(invokeAddr.String(), stateValue.DepositBalance.Asset, invokeTokens.Amount, 0)
			if err != nil {
				return shim.Error("PayOutToken error: " + err.Error())
			}
			//从列表移除，并在状态数据库删除
			//第六步：从列表移除
			handleMember("Mediator", invokeAddr, stub)
			stateValue.Time = time.Now()
			stateValue.Extra = "这是退出全部保证金,并移除出列表"
			//stateValueMarshalBytes, err := json.Marshal(stateValue)
			//if err != nil {
			//	return shim.Error("Marshal valueState error: " + err.Error())
			//}
			////第七步：并在状态数据库删除
			//stub.PutState(invokeAddr.String(), stateValueMarshalBytes)
		}
		//return shim.Success([]byte("ok"))
	} else if stateValue.DepositBalance.Amount < depositAmountsForMediator {
		//说明退款后，余额少于规定数量
		return shim.Error("说明退款后，余额少于规定数量")
	} else {
		stateValue.Extra = "这是退出保证金，只退一部分钱，剩下余额还是在规定范围之内"
		//调用从合约把token转到地址
		//第四步：从合约把token转到地址
		err = stub.PayOutToken(invokeAddr.String(), stateValue.DepositBalance.Asset, invokeTokens.Amount, 0)
		if err != nil {
			return shim.Error("PayOutToken error: " + err.Error())
		}
	}
	//更新
	stateValueMarshalBytes, err := json.Marshal(stateValue)
	if err != nil {
		return shim.Error("Marshal valueState error: " + err.Error())
	}
	//第五步：更新状态数据库
	stub.PutState(invokeAddr.String(), stateValueMarshalBytes)
	return shim.Success([]byte("ok"))
}

//保证金没收
//handle forfeiture deposit
func (d DepositChaincode) forfeitureDeposit(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	//参数是陪审员的账户和罚没数量
	//if len(args) != 2 {
	//	return shim.Error("Input error: need two arg (witnessAddr and amount)")
	//}
	////获取该账户的账本信息
	//stateValueBytes, err := stub.GetState(args[0])
	//if err != nil {
	//	return shim.Error("Get account balance from ledger error: " + err.Error())
	//}
	////if stateValueBytes == nil {
	////	return shim.Error("you have not depositWitnessPay for deposit.")
	////}
	//stateValue := new(modules.DepositStateValue)
	//err = json.Unmarshal(stateValueBytes, stateValue)
	//if err != nil {
	//	return shim.Error("unmarshal accBalByte error " + err.Error())
	//}
	////获取没收保证金数量，将 string 转 uint64
	//ptnAccount, _ := strconv.ParseUint(args[1], 10, 64)
	////if err != nil {
	////	return shim.Error("String transform to uint64 error: " + err.Error())
	////}
	////if stateValue.DepositBalance.Amount < ptnAccount {
	////	return shim.Error("Your amount balance does not enough.")
	////}
	//if stateValue.DepositBalance.Amount < ptnAccount {
	//	return shim.Error("Forfeiture too many.")
	//}
	////判断是 Jury 还是 Mediator
	//who := whoIs(stateValue.DepositBalance.Amount)
	////获取基金会地址
	//foundationAddress, err := stub.GetSystemConfig("FoundationAddress")
	//if err != nil {
	//	return shim.Error(err.Error())
	//}
	////fmt.Println("foundationAddress", foundationAddress)
	////调用从合约把token转到地址
	//err = stub.PayOutToken(foundationAddress, stateValue.DepositBalance.Asset, ptnAccount, 0)
	//if err != nil {
	//	return shim.Error("PayOutToken error: " + err.Error())
	//}
	////写入写集
	//stateValue.DepositBalance.Amount -= ptnAccount
	//stateValue.Time = time.Now()
	//stateValue.Extra = "这是退出保证金，可能只退一部分钱，时间是否需要修改？"
	//stateValueMarshalBytes, err := json.Marshal(stateValue)
	//if err != nil {
	//	return shim.Error("Marshal valueState error " + err.Error())
	//}
	//addr, _ := common.StringToAddress(args[0])
	////判断罚款后是 Jury 还是 Mediator 还是  都不是（即移除列表）
	//handleMember(who, addr, stateValue.DepositBalance.Amount, stub)
	//stub.PutState(args[0], stateValueMarshalBytes)
	return shim.Success([]byte("ok"))
}

//func assetIsEqual(invokeAsset, stateAsset modules.Asset) error {
//	if invokeAsset != stateAsset {
//		return fmt.Errorf("asset is not equal")
//	}
//	return nil
//}
