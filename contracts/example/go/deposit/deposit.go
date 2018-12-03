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
	"github.com/palletone/go-palletone/common/award"
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
		return shim.Success([]byte("GetSystemConfig with DepositAmount error: "))
	}
	//转换
	depositAmountsForMediator, err = strconv.ParseUint(depositAmountsForMediatorStr, 10, 64)
	if err != nil {
		return shim.Success([]byte("String transform to uint64 error:"))
	}
	fmt.Println("需要的mediator保证金数量=", depositAmountsForMediator)
	depositAmountsForJuryStr, err := stub.GetSystemConfig("DepositAmountForJury")
	if err != nil {
		return shim.Success([]byte("GetSystemConfig with DepositAmount error:"))
	}
	//转换
	depositAmountsForJury, err = strconv.ParseUint(depositAmountsForJuryStr, 10, 64)
	if err != nil {
		return shim.Success([]byte("String transform to uint64 error:"))
	}
	fmt.Println("需要的jury保证金数量=", depositAmountsForJury)
	depositAmountsForDeveloperStr, err := stub.GetSystemConfig("DepositAmountForDeveloper")
	if err != nil {
		return shim.Success([]byte("GetSystemConfig with DepositAmount error:"))
	}
	//转换
	depositAmountsForDeveloper, err = strconv.ParseUint(depositAmountsForDeveloperStr, 10, 64)
	if err != nil {
		return shim.Success([]byte("String transform to uint64 error:"))
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
		//return d.depositWitnessPay(stub, args)
	case "ApplyForDepositCashback":
		//申请保证金退还
		//handle cashback rewards
		//void deposit_cashback(const account_object& acct, token_type amount, bool require_vesting = true)
		return d.applyForDepositCashback(stub, args)
	case "ApplyForForfeitureDeposit":
		//申请保证金没收
		//void forfeiture_deposit(const witness_object& wit, token_type amount)
		return d.applyForForfeitureDeposit(stub, args)
	case "HandleApplications":
		//基金会对申请做相应的处理
		return d.handleApplications(stub, args)
	}
	return shim.Success([]byte("Invoke error"))
}

//交付保证金
//handle witness pay
func (d *DepositChaincode) depositWitnessPay(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	//第一个参数：合约地址；第二个参数：保证金；第三个参数：角色（Mediator Jury ContractDeveloper)
	//Deposit("contractAddr","2000","Mediator")
	if len(args) < 3 {
		return shim.Success([]byte("Input parameter Success,need three parameters."))
	}
	//获取 请求 调用 地址（即交付保证节点地址）
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		return shim.Success([]byte("GetInvokeFromAddr error:"))
	}
	fmt.Println("invokeFromAddr address = ", invokeAddr.String())
	//获取 请求 ptn 数量（即交付保证金数量）
	invokeTokens, err := stub.GetInvokeTokens()
	if err != nil {
		return shim.Success([]byte("GetPayToContractPtnTokens error:"))
	}
	//获取退保证金数量，将 string 转 uint64
	//TODO test
	ptnAccount, _ := strconv.ParseUint(args[1], 10, 64)
	invokeTokens.Amount = ptnAccount
	fmt.Println("invokeTokens ", invokeTokens.Amount)
	fmt.Printf("invokeTokens %#v\n", invokeTokens.Asset)
	//获取角色
	role := args[2]
	switch {
	case role == "Mediator":
		//处理Mediator交付保证金
		return d.handleMediatorDepositWitnessPay(stub, invokeAddr, invokeTokens)
	case role == "Jury":
		//处理Jury交付保证金
		return d.handleJuryDepositWitnessPay(stub, invokeAddr, invokeTokens)
	case role == "Developer":
		//处理Developer交付保证金
		return d.handleDeveloperDepositWitnessPay(stub, invokeAddr, invokeTokens)
	default:
		return shim.Success([]byte("role error."))
	}
}

//处理 Mediator
func (d *DepositChaincode) handleMediatorDepositWitnessPay(stub shim.ChaincodeStubInterface, invokeAddr common.Address, invokeTokens *modules.InvokeTokens) pb.Response {
	//获取一下该用户下的账簿情况
	stateValueBytes, err := stub.GetState(invokeAddr.String())
	if err != nil {
		return shim.Success([]byte("Get account balance from ledger error:"))
	}
	stateValues := new(modules.DepositStateValues)
	stateValue := new(modules.DepositStateValue)
	//账户不存在，第一次参与
	if stateValueBytes == nil {
		//判断保证金是否足够(Mediator第一次交付必须足够)
		if invokeTokens.Amount < depositAmountsForMediator {
			return shim.Success([]byte("Payment amount is insufficient."))
		}
		//加入列表
		addList("Mediator", invokeAddr, stub)
		//处理数据
		stateValues.TotalAmount = invokeTokens.Amount
		stateValues.Time = time.Now().UTC()
		stateValue.DepositBalance.Amount = invokeTokens.Amount
		stateValue.DepositBalance.Asset = invokeTokens.Asset
		stateValue.Time = time.Now().UTC() //第一次交付保证金的时间，并且加入列表
		stateValue.Extra = "这是第一次参与"
		stateValues.Values = append(stateValues.Values, stateValue)
	} else {
		//已经是mediator了
		err = json.Unmarshal(stateValueBytes, stateValues)
		if err != nil {
			return shim.Success([]byte("Unmarshal stateValueBytes error"))
		}
		//处理数据
		stateValues.TotalAmount += invokeTokens.Amount
		stateValue.DepositBalance.Amount = invokeTokens.Amount
		stateValue.DepositBalance.Asset = invokeTokens.Asset
		stateValue.Time = time.Now().UTC()
		stateValue.Extra = "这是再次向合约增加保证金数量"
		stateValues.Values = append(stateValues.Values, stateValue)
	}
	//序列化
	stateValueMarshalBytes, err := json.Marshal(stateValues)
	if err != nil {
		return shim.Success([]byte("Marshal valueState error"))
	}
	//更新数据
	stub.PutState(invokeAddr.String(), stateValueMarshalBytes)
	str := strconv.FormatUint(invokeTokens.Amount, 10)
	return shim.Success([]byte(str))
}

//处理 Jury
func (d *DepositChaincode) handleJuryDepositWitnessPay(stub shim.ChaincodeStubInterface, invokeAddr common.Address, invokeTokens *modules.InvokeTokens) pb.Response {
	//获取一下该用户下的账簿情况
	stateValueBytes, err := stub.GetState(invokeAddr.String())
	if err != nil {
		return shim.Success([]byte("Get account balance from ledger error:"))
	}
	stateValues := new(modules.DepositStateValues)
	stateValue := new(modules.DepositStateValue)
	isJury := false
	if stateValueBytes == nil {
		if invokeTokens.Amount >= depositAmountsForJury {
			addList("Jury", invokeAddr, stub)
			stateValues.Time = time.Now().UTC()
			isJury = true
		}
		stateValues.TotalAmount = invokeTokens.Amount
		//写入写集
		stateValue.DepositBalance.Amount = invokeTokens.Amount
		stateValue.DepositBalance.Asset = invokeTokens.Asset
		stateValue.Time = time.Now().UTC()
		stateValue.Extra = "这是第一次参与"
		stateValues.Values = append(stateValues.Values, stateValue)
	} else {
		//账户已存在，进行信息的更新操作
		err = json.Unmarshal(stateValueBytes, stateValues)
		if err != nil {
			return shim.Success([]byte("Unmarshal stateValueBytes error"))
		}
		if stateValue.DepositBalance.Amount >= depositAmountsForJury {
			//原来就是jury
			isJury = true
		}
		//更新stateValue
		stateValues.TotalAmount += invokeTokens.Amount
		stateValue.DepositBalance.Amount = invokeTokens.Amount
		stateValue.DepositBalance.Asset = invokeTokens.Asset
		stateValue.Time = time.Now().UTC()
		stateValue.Extra = "这是再次向合约支付保证金"
		stateValues.Values = append(stateValues.Values, stateValue)
	}
	if !isJury {
		//判断交了保证金后是否超过了jury
		if stateValue.DepositBalance.Amount >= depositAmountsForJury {
			addList("Jury", invokeAddr, stub)
			stateValues.Time = time.Now().UTC()
		}
	}
	stateValueMarshalBytes, err := json.Marshal(stateValues)
	if err != nil {
		return shim.Success([]byte("Marshal valueState error"))
	}
	stub.PutState(invokeAddr.String(), stateValueMarshalBytes)
	str := strconv.FormatUint(invokeTokens.Amount, 10)
	return shim.Success([]byte(str))
}

//处理 ContractDeveloper
func (d *DepositChaincode) handleDeveloperDepositWitnessPay(stub shim.ChaincodeStubInterface, invokeAddr common.Address, invokeTokens *modules.InvokeTokens) pb.Response {
	//获取一下该用户下的账簿情况
	stateValueBytes, err := stub.GetState(invokeAddr.String())
	if err != nil {
		return shim.Success([]byte("Get account balance from ledger error:"))
	}
	stateValues := new(modules.DepositStateValues)
	stateValue := new(modules.DepositStateValue)
	isDeveloper := false
	if stateValueBytes == nil {
		if invokeTokens.Amount >= depositAmountsForDeveloper {
			addList("Developer", invokeAddr, stub)
			stateValues.Time = time.Now().UTC()
			isDeveloper = true
		}
		//写入写集
		stateValues.TotalAmount = invokeTokens.Amount
		stateValue.DepositBalance.Amount = invokeTokens.Amount
		stateValue.DepositBalance.Asset = invokeTokens.Asset
		stateValue.Time = time.Now().UTC()
		stateValue.Extra = "这是第一次参与"
		stateValues.Values = append(stateValues.Values, stateValue)
	} else {
		//账户已存在，进行信息的更新操作
		err = json.Unmarshal(stateValueBytes, stateValues)
		if err != nil {
			return shim.Success([]byte("Unmarshal stateValueBytes error"))
		}
		if stateValue.DepositBalance.Amount >= depositAmountsForDeveloper {
			//原来就是 Developer
			isDeveloper = true
		}
		//更新stateValue
		stateValues.TotalAmount += invokeTokens.Amount
		stateValue.DepositBalance.Amount = invokeTokens.Amount
		stateValue.DepositBalance.Asset = invokeTokens.Asset
		stateValue.Time = time.Now().UTC()
		stateValue.Extra = "这是再次向合约支付保证金"
		stateValues.Values = append(stateValues.Values, stateValue)
	}
	//判断资产类型是否一致
	//err = assetIsEqual(invokeTokens.Asset, stateValue.Asset)
	//if err != nil {
	// return shim.Success("InvokeAsset is not equal with stateAsset Success:"))
	//}
	if !isDeveloper {
		//判断交了保证金后是否超过了jury
		if stateValue.DepositBalance.Amount >= depositAmountsForDeveloper {
			addList("Developer", invokeAddr, stub)
			stateValues.Time = time.Now().UTC()
		}
	}
	stateValueMarshalBytes, err := json.Marshal(stateValues)
	if err != nil {
		return shim.Success([]byte("Marshal valueState error"))
	}
	stub.PutState(invokeAddr.String(), stateValueMarshalBytes)
	str := strconv.FormatUint(invokeTokens.Amount, 10)
	return shim.Success([]byte(str))
}

//保证金退还，只申请，当然符合要求了才能申请成功，并且加入申请列表
//handle cashback rewards
func (d *DepositChaincode) applyForDepositCashback(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	//第一个参数：数量；第二个参数：角色（角色（Mediator Jury ContractDeveloper)
	//depositCashback("保证金数量","Mediator")
	if len(args) < 2 {
		return shim.Success([]byte("Input parameter Success,need two parameters."))
	}
	//获取 请求 调用 地址
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		return shim.Success([]byte("GetInvokeFromAddr error:"))
	}
	fmt.Println("invokeAddr address ", invokeAddr.String())
	//获取退保证金数量，将 string 转 uint64
	ptnAccount, err := strconv.ParseUint(args[0], 10, 64)
	if err != nil {
		return shim.Success([]byte("String transform to uint64 error:"))
	}
	fmt.Println("ptnAccount  args[0] ", ptnAccount)
	asset := modules.NewPTNAsset()
	invokeTokens := &modules.InvokeTokens{
		Amount: ptnAccount,
		Asset:  asset,
	}
	//
	//先获取数据库信息
	stateValueBytes, err := stub.GetState(invokeAddr.String())
	if err != nil {
		return shim.Success([]byte("Get account balance from ledger error:"))
	}
	//判断数据库是否为空
	if stateValueBytes == nil {
		return shim.Success([]byte("Your account does not exist."))
	}
	values := new(modules.DepositStateValues)
	//如果不为空，反序列化数据库信息
	err = json.Unmarshal(stateValueBytes, values)
	if err != nil {
		return shim.Success([]byte("Unmarshal stateValueBytes error:"))
	}
	//TODO 判断该节点是否正在担任执行任务
	if values.IsRunning == true {
		return shim.Success([]byte("正在执行任务，不能退保证金。"))
	}
	//比较退款数量和数据库数量
	//Asset判断
	//数量比较
	if values.TotalAmount < invokeTokens.Amount {
		return shim.Success([]byte("Your delivery amount with ptn token is insufficient."))
	}
	err = d.addListForCashback(args[1], stub, invokeAddr, invokeTokens)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success([]byte("申请成功"))
}

//加入退款申请列表
func (d *DepositChaincode) addListForCashback(role string, stub shim.ChaincodeStubInterface, invokeAddr common.Address, invokeTokens *modules.InvokeTokens) error {
	//先获取申请列表
	listForCashbackByte, err := stub.GetState("ListForCashback")
	if err != nil {
		return err
	}
	listForCashback := new(modules.ListForCashback)
	//序列化
	cashback := new(modules.Cashback)
	cashback.InvokeAddress = invokeAddr
	cashback.InvokeTokens = *invokeTokens
	cashback.Role = role
	cashback.ApplyTime = time.Now().UTC().Unix()
	if listForCashbackByte == nil {
		listForCashback.Cashback = append(listForCashback.Cashback, cashback)
	} else {
		err = json.Unmarshal(listForCashbackByte, listForCashback)
		if err != nil {
			return err
		}
		listForCashback.Cashback = append(listForCashback.Cashback, cashback)
	}
	//反序列化
	listForCashbackByte, err = json.Marshal(listForCashback)
	if err != nil {
		return err
	}
	err = stub.PutState("ListForCashback", listForCashbackByte)
	if err != nil {
		return err
	}
	return nil
}

//这里是基金会处理保证金提取的请求
func (d *DepositChaincode) handleDepositCashbackApplication(stub shim.ChaincodeStubInterface, foundationAddr, forfeitureAddr common.Address, applyTime int64, values *modules.DepositStateValues, check string) pb.Response {
	//提取保证金节点地址，申请时间
	if check == "ok" {
		//return d.agreeForApplyCashback(stub, foundationAddr, forfeitureAddr, applyTime, values)
	} else {
		//return d.disagreeForApplyCashback(stub, forfeitureAddr, applyTime)
	}
	return shim.Success([]byte("ok"))
}

//同意申请退保证金请求

//func (d *DepositChaincode) handleMediatorDepositCashback(stub shim.ChaincodeStubInterface, invokeAddr common.Address, invokeTokens *modules.InvokeTokens, values *modules.DepositStateValues) pb.Response {
//	var err error
//	//规定mediator 退款要么全部退，要么退款后，剩余数量在mediator数量范围内，
//	values.TotalAmount -= invokeTokens.Amount
//	//判断是否全部退
//	if values.TotalAmount == 0 {
//		//加入候选列表的时的时间
//		startTime := values.Time.YearDay()
//		//当前时间
//		endTime := time.Now().UTC().YearDay()
//		//判断是否已超过规定周期
//		if endTime-startTime >= 0 {
//			//退出全部，即删除cashback
//			err = d.cashbackAllDeposit("Mediator", stub, invokeAddr, invokeTokens, values)
//			if err != nil {
//				return shim.Success([]byte(err.Error()))
//			}
//			return shim.Success([]byte("成功退出"))
//		} else {
//			//没有超过周期，不能退出
//			return shim.Success([]byte("还在规定周期之内，不得退出列表"))
//		}
//	} else if values.TotalAmount < depositAmountsForMediator {
//		//说明退款后，余额少于规定数量
//		return shim.Success([]byte("说明退款后，余额少于规定数量"))
//	} else {
//		//TODO 这是只退一部分钱，剩下余额还是在规定范围之内
//		err = d.addListForCashback("Mediator", stub, invokeAddr, invokeTokens)
//		if err != nil {
//			return shim.Success([]byte(err.Error()))
//		}
//		return shim.Success([]byte("成功退出一部分"))
//	}
//}

//对Jury退保证金的处理
//func (d *DepositChaincode) handleJuryDepositCashback(stub shim.ChaincodeStubInterface, invokeAddr common.Address, invokeTokens *modules.InvokeTokens, values *modules.DepositStateValues) pb.Response {
//	var res pb.Response
//	if values.TotalAmount >= depositAmountsForJury {
//		//已在列表中
//		res = d.handleJuryFromList(stub, invokeAddr, invokeTokens, values)
//	} else {
//		//TODO 不在列表中,没有奖励，直接退
//		//res = handleCommonJury(stub, invokeAddr, invokeTokens, values)
//	}
//	return res
//}

//Jury已在列表中
//func (d *DepositChaincode) handleJuryFromList(stub shim.ChaincodeStubInterface, invokeAddr common.Address, invokeTokens *modules.InvokeTokens, values *modules.DepositStateValues) pb.Response {
//	//退出列表
//	var err error
//	//计算余额
//	values.TotalAmount -= invokeTokens.Amount
//	//判断是否退出列表
//	if values.TotalAmount == 0 {
//		//加入列表时的时间
//		startTime := values.Time.YearDay()
//		//当前退出时间
//		endTime := time.Now().UTC().YearDay()
//		//判断是否已到期
//		if endTime-startTime >= 0 {
//			//退出全部，即删除cashback，利息计算好了
//			err = d.cashbackAllDeposit("Jury", stub, invokeAddr, invokeTokens, values)
//			if err != nil {
//				return shim.Success([]byte(err.Error()))
//			}
//			return shim.Success([]byte("成功退出"))
//		} else {
//			return shim.Success([]byte("未到期，不能退出列表"))
//		}
//	} else {
//		//TODO 退出一部分，且退出该部分金额后还在列表中，还没有计算利息
//		d.addListForCashback("Jury", stub, invokeAddr, invokeTokens)
//		if err != nil {
//			return shim.Success([]byte(err.Error()))
//		}
//		return shim.Success([]byte("成功退出一部分"))
//	}
//}

//对Developer退保证金的处理
//func (d *DepositChaincode) handleDeveloperDepositCashback(stub shim.ChaincodeStubInterface, invokeAddr common.Address, invokeTokens *modules.InvokeTokens, values *modules.DepositStateValues) pb.Response {
//	var res pb.Response
//	if values.TotalAmount >= depositAmountsForDeveloper {
//		//已在列表中
//		res = d.handleDeveloperFromList(stub, invokeAddr, invokeTokens, values)
//	} else {
//		////TODO 不在列表中,没有奖励，直接退
//		//res = handleCommonDeveloper(stub, invokeAddr, invokeTokens, values)
//	}
//	return res
//}

//Developer已在列表中
//func (d *DepositChaincode) handleDeveloperFromList(stub shim.ChaincodeStubInterface, invokeAddr common.Address, invokeTokens *modules.InvokeTokens, values *modules.DepositStateValues) pb.Response {
//	//退出列表
//	var err error
//	//计算余额
//	values.TotalAmount -= invokeTokens.Amount
//	//判断是否退出列表
//	if values.TotalAmount == 0 {
//		//加入列表时的时间
//		startTime := values.Time.YearDay()
//		//当前退出时间
//		endTime := time.Now().UTC().YearDay()
//		//判断是否已到期
//		if endTime-startTime >= 0 {
//			//退出全部，即删除cashback
//			err = d.cashbackAllDeposit("Developer", stub, invokeAddr, invokeTokens, values)
//			if err != nil {
//				return shim.Success([]byte(err.Error()))
//			}
//			return shim.Success([]byte("成功退出"))
//		} else {
//			return shim.Success([]byte("未到期，不能退出列表"))
//		}
//	} else {
//		//TODO 退出一部分，是否还在列表里需要基金会执行时才确定
//		d.addListForCashback("Developer", stub, invokeAddr, invokeTokens)
//		if err != nil {
//			return shim.Success([]byte(err.Error()))
//		}
//		return shim.Success([]byte("成功退出一部分"))
//	}
//}

//退出列表处理
//func (d *DepositChaincode) cashbackAllDeposit(who string, stub shim.ChaincodeStubInterface, invokeAddr common.Address, invokeTokens *modules.InvokeTokens, values *modules.DepositStateValues) error {
//	//计算总奖励
//	totalAward := calculateAllAwards(values.Values)
//	//退款总额：本金+利息
//	invokeTokens.Amount += totalAward
//	//TODO 该节点申请退出列表，需要得到基金会的同意，暂时加入申请列表中
//	return d.addListForCashback(who, stub, invokeAddr, invokeTokens)
//}

//节点还在列表中，计算这部分提款的余额的奖励
//func (d *DepositChaincode) cashbackSomeDeposit(who string, stub shim.ChaincodeStubInterface, invokeAddr common.Address, invokeTokens *modules.InvokeTokens, forfeitureAddr common.Address, values *modules.DepositStateValues) error {
//	//TODO 计算这部分余额的奖励
//	someAward := calculateSomeAwards(values.Values, invokeTokens)
//	//退款+利息
//	invokeTokens.Amount += someAward
//	//TODO 该节点申请退出列表，需要得到基金会的同意，暂时加入申请列表中
//	return d.addListForCashback(who, stub, invokeAddr, invokeTokens)
//}

//社区申请没收某节点的保证金数量
func (d DepositChaincode) applyForForfeitureDeposit(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	//谁请求的 没收地址 数量 角色 额外说明
	//forfeiture common.Address, invokeTokens modules.InvokeTokens, role, extra string
	if len(args) != 4 {
		return shim.Error("需要4个参数")
	}
	//申请地址
	invokeAddr, _ := stub.GetInvokeAddress()
	forfeiture := new(modules.Forfeiture)
	forfeiture.InvokeAddress = invokeAddr
	forfeitureAddr, err := common.StringToAddress(args[0])
	//获取没收节点地址
	if err != nil {
		return shim.Success([]byte("string to address error"))
	}
	fmt.Println(forfeitureAddr.String())
	//获取没收节点的账本信息
	stateValueBytes, err := stub.GetState(forfeitureAddr.String())
	if err != nil {
		return shim.Success([]byte("Get account balance from ledger error:"))
	}
	//判断没收节点账户是否为空
	if stateValueBytes == nil {
		return shim.Success([]byte("you have not depositWitnessPay for deposit."))
	}
	values := new(modules.DepositStateValues)
	//将没收节点账户序列化
	err = json.Unmarshal(stateValueBytes, values)
	if err != nil {
		return shim.Success([]byte("unmarshal accBalByte error"))
	}
	//获取没收保证金数量，将 string 转 uint64
	ptnAccount, err := strconv.ParseUint(args[1], 10, 64)
	if err != nil {
		return shim.Success([]byte("String transform to uint64 error:"))
	}
	fmt.Println("ptnAccount  args[1] ", ptnAccount)
	//判断账户余额和没收请求数量
	if values.TotalAmount < ptnAccount {
		return shim.Success([]byte("Forfeiture too many."))
	}
	forfeiture.ForfeitureAddress = forfeitureAddr
	asset := modules.NewPTNAsset()
	invokeTokens := modules.InvokeTokens{
		Amount: ptnAccount,
		Asset:  asset,
	}
	forfeiture.InvokeTokens = invokeTokens
	forfeiture.Role = args[2]
	forfeiture.Extra = args[3]
	forfeiture.ApplyTime = time.Now().UTC().Unix()
	//先获取列表，再更新列表
	listForForfeitureByte, err := stub.GetState("ListForForfeiture")
	if err != nil {
		return shim.Error(err.Error())
	}
	listForForfeiture := new(modules.ListForForfeiture)
	if listForForfeitureByte == nil {
		listForForfeiture.Forfeiture = append(listForForfeiture.Forfeiture, forfeiture)
	} else {
		err = json.Unmarshal(listForForfeitureByte, listForForfeiture)
		if err != nil {
			return shim.Error(err.Error())
		}
		listForForfeiture.Forfeiture = append(listForForfeiture.Forfeiture, forfeiture)
	}
	listForForfeitureByte, err = json.Marshal(listForForfeiture)
	if err != nil {
		return shim.Error(err.Error())
	}
	err = stub.PutState("ListForForfeiture", listForForfeitureByte)
	return shim.Success([]byte("申请成功"))
}

//基金会处理没收请求
func (d *DepositChaincode) handleForfeitureDepositApplication(stub shim.ChaincodeStubInterface, foundationAddr, forfeitureAddr common.Address, applyTime int64, values *modules.DepositStateValues, check string) pb.Response {
	//check 如果为ok，则同意此申请，如果为no，则不同意此申请
	if check == "ok" {
		return d.agreeForApplyForfeiture(stub, foundationAddr, forfeitureAddr, applyTime, values)
	} else {
		//移除申请列表，不做处理
		return d.disagreeForApplyForfeiture(stub, forfeitureAddr, applyTime)
	}
}

//不同意这样没收请求，则直接从没收列表中移除该节点
func (d *DepositChaincode) disagreeForApplyForfeiture(stub shim.ChaincodeStubInterface, forfeiture common.Address, applyTime int64) pb.Response {
	//获取没收列表
	listForForfeitureByte, err := stub.GetState("ListForForfeiture")
	if err != nil {
		return shim.Error(err.Error())
	}
	if listForForfeitureByte == nil {
		return shim.Error("列表为空")
	}
	listForForfeiture := new(modules.ListForForfeiture)
	err = json.Unmarshal(listForForfeitureByte, listForForfeiture)
	if err != nil {
		return shim.Error(err.Error())
	}
	node := moveInApplyForForfeitureList(stub, listForForfeiture.Forfeiture, forfeiture, applyTime)
	if node != nil {
		return shim.Error("列表里没有该申请")
	}
	return shim.Success([]byte("移除列表成功"))
}

//同意申请没收请求
func (d *DepositChaincode) agreeForApplyForfeiture(stub shim.ChaincodeStubInterface, foundationAddr, forfeitureAddr common.Address, applyTime int64, values *modules.DepositStateValues) pb.Response {
	//获取列表
	listForForfeitureByte, err := stub.GetState("ListForForfeiture")
	if err != nil {
		return shim.Error(err.Error())
	}
	if listForForfeitureByte == nil {
		return shim.Error("listForForfeitureByte is nil")
	}
	listForForfeiture := new(modules.ListForForfeiture)
	err = json.Unmarshal(listForForfeitureByte, listForForfeiture)
	if err != nil {
		return shim.Error("json unmarshal error " + err.Error())
	}
	//在列表中移除，并获取没收情况
	forfeiture := moveInApplyForForfeitureList(stub, listForForfeiture.Forfeiture, forfeitureAddr, applyTime)
	//判断节点类型
	switch {
	case forfeiture.Role == "Mediator":
		return d.handleMediatorForfeitureDeposit(foundationAddr, forfeitureAddr, &forfeiture.InvokeTokens, values, stub)
	case forfeiture.Role == "Jury":
		return d.handleJuryForfeitureDeposit(foundationAddr, forfeitureAddr, &forfeiture.InvokeTokens, values, stub)
	case forfeiture.Role == "Developer":
		return d.handleDeveloperForfeitureDeposit(foundationAddr, forfeitureAddr, &forfeiture.InvokeTokens, values, stub)
	default:
		return shim.Error("role error")
	}
}

func (d *DepositChaincode) handleApplications(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	//地址，申请时间，是否同意，类型（提款，没收，错误）
	if len(args) != 4 {
		return shim.Success([]byte("Input parameter error,need four parameters."))
	}

	//基金会地址
	invokeAddr, _ := stub.GetInvokeAddress()
	fmt.Println("invokeAddr==", invokeAddr.String())
	//获取系统配置基金会地址
	foundationAddress, err := stub.GetSystemConfig("FoundationAddress")
	if err != nil {
		return shim.Success([]byte("获取基金会地址错误"))
	}
	fmt.Println("foundationAddress==", foundationAddress)
	//判断没收请求地址是否是基金会地址
	if invokeAddr.String() != foundationAddress {
		return shim.Success([]byte("请求地址不正确，请使用基金会的地址"))
	}

	//获取没收节点地址
	nodeAddr, err := common.StringToAddress(args[0])
	if err != nil {
		return shim.Success([]byte("string to address error"))
	}
	fmt.Println(nodeAddr.String())

	//获取没收节点的账本信息
	stateValueBytes, err := stub.GetState(nodeAddr.String())
	if err != nil {
		return shim.Success([]byte("Get account balance from ledger error:"))
	}
	//判断没收节点账户是否为空
	if stateValueBytes == nil {
		return shim.Success([]byte("you have not depositWitnessPay for deposit."))
	}
	values := new(modules.DepositStateValues)
	//将没收节点账户序列化
	err = json.Unmarshal(stateValueBytes, values)
	if err != nil {
		return shim.Success([]byte("unmarshal accBalByte error"))
	}

	//获取申请时间戳
	applyTime, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return shim.Error("string to int64 error " + err.Error())
	}

	//获取是否同意
	check := args[2]

	//获取处理类别
	style := args[3]
	switch {
	case style == "Cashback":
		return d.handleDepositCashbackApplication(stub, invokeAddr, nodeAddr, applyTime, values, check)
	case style == "Forfeiture":
		return d.handleForfeitureDepositApplication(stub, invokeAddr, nodeAddr, applyTime, values, check)
	default:
		return shim.Error("类别错误")
	}
}

//保证金没收
//handle forfeiture deposit
//func (d *DepositChaincode) handlForfeitureApplication(stub shim.ChaincodeStubInterface, args []string) pb.Response {
//	//第一个参数：没收地址；第二个参数：没收数量；第三个参数：角色 (Mediator  Jury  Developer)
//	//forfeitureDeposit("没收地址","申请时间","是否同意")
//	if len(args) != 3 {
//		return shim.Success([]byte("Input parameter error,need three parameters."))
//	}
//	//基金会地址
//	invokeAddr, _ := stub.GetInvokeAddress()
//	fmt.Println("invokeAddr==", invokeAddr.String())
//	//获取系统配置基金会地址
//	foundationAddress, err := stub.GetSystemConfig("FoundationAddress")
//	if err != nil {
//		return shim.Success([]byte("获取基金会地址错误"))
//	}
//	fmt.Println("foundationAddress==", foundationAddress)
//	//判断没收请求地址是否是基金会地址
//	if invokeAddr.String() != foundationAddress {
//		return shim.Success([]byte("请求地址不正确，请使用基金会的地址"))
//	}
//	//获取没收节点地址
//	forfeitureAddr, err := common.StringToAddress(args[0])
//	if err != nil {
//		return shim.Success([]byte("string to address error"))
//	}
//	fmt.Println(forfeitureAddr.String())
//	//获取没收节点的账本信息
//	stateValueBytes, err := stub.GetState(forfeitureAddr.String())
//	if err != nil {
//		return shim.Success([]byte("Get account balance from ledger error:"))
//	}
//	//判断没收节点账户是否为空
//	if stateValueBytes == nil {
//		return shim.Success([]byte("you have not depositWitnessPay for deposit."))
//	}
//	values := new(modules.DepositStateValues)
//	//将没收节点账户序列化
//	err = json.Unmarshal(stateValueBytes, values)
//	if err != nil {
//		return shim.Success([]byte("unmarshal accBalByte error"))
//	}
//	//获取申请时间戳
//	applyTime, err := strconv.ParseInt(args[1], 10, 64)
//	if err != nil {
//		return shim.Error("string to int64 error " + err.Error())
//	}
//
//	////从列表中找到这个请求，支付给基金会地址
//	//err = d.forfeitureDeposit(stub, invokeAddr, forfeitureAddr, applyTime, values, args[2])
//	//
//	////获取没收保证金数量，将 string 转 uint64
//	//ptnAccount, _ := strconv.ParseUint(args[1], 10, 64)
//	//if err != nil {
//	//	return shim.Success([]byte("String transform to uint64 error:"))
//	//}
//	////判断账户余额和没收请求数量
//	//if values.TotalAmount < ptnAccount {
//	//	return shim.Success([]byte("Forfeiture too many."))
//	//}
//	//asset := modules.NewPTNAsset()
//	//invokeTokens := &modules.InvokeTokens{
//	//	Amount: ptnAccount,
//	//	Asset:  asset,
//	//}
//	////判断是 Jury 还是 Mediator 还是 developer
//	//role := args[2]
//	//res := pb.Response{}
//	//switch {
//	//case role == "Mediator":
//	//	//处理没收Mediator保证金
//	//	res = d.handleMediatorForfeitureDeposit(invokeAddr, forfeitureAddr, invokeTokens, values, stub)
//	//case role == "Jury":
//	//	//处理没收Jury保证金
//	//	res = d.handleJuryForfeitureDeposit(invokeAddr, forfeitureAddr, invokeTokens, values, stub)
//	//case role == "Developer":
//	//	//处理没收Developer保证金
//	//	res = d.handleDeveloperForfeitureDeposit(invokeAddr, forfeitureAddr, invokeTokens, values, stub)
//	//}
//	//return res
//}

func (d *DepositChaincode) forfeitureAllDeposit(role string, stub shim.ChaincodeStubInterface, foundationAddr, forfeitureAddr common.Address, invokeTokens *modules.InvokeTokens) error {
	//调用从合约把token转到请求地址
	err := stub.PayOutToken(foundationAddr.String(), invokeTokens, 0)
	if err != nil {
		return err
	}
	//移除出列表
	handleMember(role, forfeitureAddr, stub)
	//删除节点
	err = stub.DelState(forfeitureAddr.String())
	if err != nil {
		return err
	}
	return nil
}

//处理没收Mediator保证金
func (d *DepositChaincode) handleMediatorForfeitureDeposit(foundationAddr, forfeitureAddr common.Address, tokens *modules.InvokeTokens, values *modules.DepositStateValues, stub shim.ChaincodeStubInterface) pb.Response {
	var err error
	//计算余额
	values.TotalAmount -= tokens.Amount
	//判断是否没收全部，即在列表中移除该节点
	if values.TotalAmount == 0 {
		//没收不考虑是否在规定周期内,其实它肯定是在列表中并已在周期内
		//没收全部，即删除,已经是计算好奖励了
		err = d.forfeitureAllDeposit("Mediator", stub, foundationAddr, forfeitureAddr, tokens)
		if err != nil {
			return shim.Success([]byte(err.Error()))
		}
		return shim.Success([]byte("成功退出"))
	} else {
		//TODO 对于mediator，要么全没收，要么退出一部分，且退出该部分金额后还在列表中
		d.forfeitureSomeDeposit("Mediator", stub, foundationAddr, forfeitureAddr, tokens, values)
		if err != nil {
			return shim.Success([]byte(err.Error()))
		}
		return shim.Success([]byte("成功没收一部分"))
	}
}

func (d *DepositChaincode) forfertureAndMoveList(role string, stub shim.ChaincodeStubInterface, foundationAddr, forfeitureAddr common.Address, tokens *modules.InvokeTokens, values *modules.DepositStateValues) error {
	//调用从合约把token转到请求地址
	err := stub.PayOutToken(foundationAddr.String(), tokens, 0)
	if err != nil {
		return err
	}
	handleMember(role, forfeitureAddr, stub)
	//从账户中移除相应没收余额
	handleValues(values.Values, tokens)
	//更新数据库
	//序列化
	stateValuesMarshalByte, err := json.Marshal(values)
	if err != nil {
		return err
	}
	//更新数据
	err = stub.PutState(forfeitureAddr.String(), stateValuesMarshalByte)
	if err != nil {
		return err
	}
	return nil
}

//不需要移除候选列表，但是要没收一部分保证金
func (d *DepositChaincode) forfeitureSomeDeposit(role string, stub shim.ChaincodeStubInterface, foundationAddr, forfeitureAddr common.Address, tokens *modules.InvokeTokens, values *modules.DepositStateValues) error {
	//调用从合约把token转到请求地址
	err := stub.PayOutToken(foundationAddr.String(), tokens, 0)
	if err != nil {
		return err
	}
	//从账户中移除相应没收余额
	handleValues(values.Values, tokens)
	//更新数据库
	//序列化
	stateValuesMarshalByte, err := json.Marshal(values)
	if err != nil {
		return err
	}
	//更新数据
	err = stub.PutState(forfeitureAddr.String(), stateValuesMarshalByte)
	if err != nil {
		return err
	}
	return nil
}

func (d *DepositChaincode) handleJuryForfeitureDeposit(foundationAddr, forfeitureAddr common.Address, tokens *modules.InvokeTokens, values *modules.DepositStateValues, stub shim.ChaincodeStubInterface) pb.Response {
	var err error
	//计算余额
	values.TotalAmount -= tokens.Amount
	//判断是否没收全部，即在列表中移除该节点
	if values.TotalAmount == 0 {
		//没收不考虑是否在规定周期内,其实它肯定是在列表中并已在周期内
		//没收全部，即删除
		err = d.forfeitureAllDeposit("Jury", stub, foundationAddr, forfeitureAddr, tokens)
		if err != nil {
			return shim.Success([]byte(err.Error()))
		}
		return shim.Success([]byte("成功退出"))

	} else if values.TotalAmount < depositAmountsForMediator {
		//TODO 对于jury，需要移除列表
		err = d.forfertureAndMoveList("Jury", stub, foundationAddr, forfeitureAddr, tokens, values)
		if err != nil {
			return shim.Error(err.Error())
		}
		return shim.Success([]byte("成功没收一部分"))
	} else {
		//TODO 退出一部分，且退出该部分金额后还在列表中
		d.forfeitureSomeDeposit("Jury", stub, foundationAddr, forfeitureAddr, tokens, values)
		if err != nil {
			return shim.Success([]byte(err.Error()))
		}
		return shim.Success([]byte("成功没收一部分"))
	}
}

func (d *DepositChaincode) handleDeveloperForfeitureDeposit(foundationAddr, forfeitureAddr common.Address, tokens *modules.InvokeTokens, values *modules.DepositStateValues, stub shim.ChaincodeStubInterface) pb.Response {
	var err error
	//计算余额
	values.TotalAmount -= tokens.Amount
	//判断是否没收全部，即在列表中移除该节点
	if values.TotalAmount == 0 {
		//没收不考虑是否在规定周期内,其实它肯定是在列表中并已在周期内
		//没收全部，即删除
		err = d.forfeitureAllDeposit("Developer", stub, foundationAddr, forfeitureAddr, tokens)
		if err != nil {
			return shim.Success([]byte(err.Error()))
		}
		return shim.Success([]byte("成功退出"))

	} else if values.TotalAmount < depositAmountsForMediator {
		err = d.forfertureAndMoveList("Developer", stub, foundationAddr, forfeitureAddr, tokens, values)
		if err != nil {
			return shim.Error(err.Error())
		}
		return shim.Success([]byte("成功没收一部分"))
	} else {
		//TODO 退出一部分，且退出该部分金额后还在列表中
		d.forfeitureSomeDeposit("Developer", stub, foundationAddr, forfeitureAddr, tokens, values)
		if err != nil {
			return shim.Success([]byte(err.Error()))
		}
		return shim.Success([]byte("成功没收一部分"))
	}
}

//func assetIsEqual(invokeAsset, stateAsset modules.Asset) Success {
//	if invokeAsset != stateAsset {
//		return fmt.Successf("asset is not equal"))
//	}
//	return nil
//}
//计算全部的奖励
func calculateAllAwards(values []*modules.DepositStateValue) uint64 {
	allAward := uint64(0)
	for _, v := range values {
		//计算奖励
		award := award.CalculateAwardsForDepositContractNodes(v.DepositBalance.Amount, v.Time.UTC().Unix())
		allAward += award
	}
	return allAward
}

//可以说是移除数组前面的
func handleValues(values []*modules.DepositStateValue, invokeTokens *modules.InvokeTokens) {
	num := uint64(0)
	for i := 0; i < len(values); i++ {
		num += values[i].DepositBalance.Amount
		if num >= invokeTokens.Amount {
			//证明第i个是，已经超过了
			if num != invokeTokens.Amount {
				values[i].DepositBalance.Amount = num - invokeTokens.Amount
				values = values[i:]
			} else {
				values = values[i+1:]
			}
			break
		}
	}
}

//计算退出一部分钱
func calculateSomeAwards(values []*modules.DepositStateValue, invokeTokens *modules.InvokeTokens) uint64 {
	someAward := uint64(0)
	for i, value := range values {
		someAward += value.DepositBalance.Amount
		if someAward > invokeTokens.Amount {
			if i == 0 {
				values[i].DepositBalance.Amount -= invokeTokens.Amount
			} else {
				if someAward == invokeTokens.Amount {
					values = values[i+1:]
				} else {
					values = values[i:]
					values[i].DepositBalance.Amount = someAward - invokeTokens.Amount
				}
			}
			break
		}
	}
	return someAward
}
