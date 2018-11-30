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
		return shim.Success([]byte("Invoke error"))
	}
}

//交付保证金
//handle witness pay
func (d *DepositChaincode) depositWitnessPay(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	//第一个参数：合约地址；第二个参数：保证金；第三个参数：角色（Mediator Jury ContractDeveloper)
	if len(args) < 3 {
		return shim.Success([]byte("Input parameter Success,need three parameters."))
	}
	//获取 请求 调用 地址
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		return shim.Success([]byte("GetInvokeFromAddr error:"))
	}
	fmt.Println("invokeFromAddr address = ", invokeAddr.String())
	//获取 请求 ptn 数量
	invokeTokens, err := stub.GetInvokeTokens()
	if err != nil {
		return shim.Success([]byte("GetPayToContractPtnTokens error:"))
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
		//处理Mediator交付保证金
		res = d.handleMediatorDepositWitnessPay(stub, invokeAddr, invokeTokens)
	case role == "Jury":
		//处理Jury交付保证金
		res = d.handleJuryDepositWitnessPay(stub, invokeAddr, invokeTokens)
	case role == "Developer":
		//处理Developer交付保证金
		res = d.handleDeveloperDepositWitnessPay(stub, invokeAddr, invokeTokens)
	default:
		return shim.Success([]byte("role error."))
	}
	return res
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
		//判断保证金是否足够
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
	//	return shim.Success("InvokeAsset is not equal with stateAsset Success:"))
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

//保证金退还
//handle cashback rewards
func (d *DepositChaincode) depositCashback(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	//第一个参数：数量；第二个参数：角色（角色（Mediator Jury ContractDeveloper)
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
	//比较退款数量和数据库数量
	//Asset判断
	//数量比较
	if values.TotalAmount < invokeTokens.Amount {
		return shim.Success([]byte("Your delivery amount with ptn token is insufficient."))
	}
	//获取角色
	role := args[1]
	var res pb.Response
	switch {
	case role == "Mediator":
		//处理Mediator退还保证金
		res = d.handleMediatorDepositCashback(stub, invokeAddr, invokeTokens, values)
	case role == "Jury":
		//处理Jury退还保证金
		res = d.handleJuryDepositCashback(stub, invokeAddr, invokeTokens, values)
	case role == "Developer":
		//处理Developer退还保证金
		res = d.handleDeveloperDepositCashback(stub, invokeAddr, invokeTokens, values)
	default:
		//角色错误
		return shim.Success([]byte("role error."))
	}
	return res
}

func (d *DepositChaincode) handleMediatorDepositCashback(stub shim.ChaincodeStubInterface, invokeAddr common.Address, invokeTokens *modules.InvokeTokens, values *modules.DepositStateValues) pb.Response {
	var err error
	//规定mediator 退款要么全部退，要么退款后，剩余数量在mediator数量范围内，
	values.TotalAmount -= invokeTokens.Amount
	//判断是否全部退
	if values.TotalAmount == 0 {
		//加入候选列表的时的时间
		startTime := values.Time.YearDay()
		//当前时间
		endTime := time.Now().UTC().YearDay()
		//判断是否已超过规定周期
		if endTime-startTime >= 0 {
			//退出全部，即删除cashback
			err = d.cashbackOrForfeitureAllDeposit("Mediator", stub, invokeAddr, invokeTokens, invokeAddr, values)
			if err != nil {
				return shim.Success([]byte(err.Error()))
			}
			return shim.Success([]byte("成功退出"))
		} else {
			//没有超过周期，不能退出
			return shim.Success([]byte("还在规定周期之内，不得退出列表"))
		}
	} else if values.TotalAmount < depositAmountsForMediator {
		//说明退款后，余额少于规定数量
		return shim.Success([]byte("说明退款后，余额少于规定数量"))
	} else {
		//TODO 这是只退一部分钱，剩下余额还是在规定范围之内
		d.cashbackOrForfeitureSomeDeposit("Mediator", stub, invokeAddr, invokeTokens, invokeAddr, values)
		if err != nil {
			return shim.Success([]byte(err.Error()))
		}
		return shim.Success([]byte("成功退出一部分"))
	}
}

//对Jury退保证金的处理
func (d *DepositChaincode) handleJuryDepositCashback(stub shim.ChaincodeStubInterface, invokeAddr common.Address, invokeTokens *modules.InvokeTokens, values *modules.DepositStateValues) pb.Response {
	var res pb.Response
	if values.TotalAmount >= depositAmountsForJury {
		//已在列表中
		res = d.handleJuryFromList(stub, invokeAddr, invokeTokens, values)
	} else {
		//TODO 不在列表中
		//res = handleCommonJury(stub, invokeAddr, invokeTokens, values)
	}
	return res
}

//Jury已在列表中
func (d *DepositChaincode) handleJuryFromList(stub shim.ChaincodeStubInterface, invokeAddr common.Address, invokeTokens *modules.InvokeTokens, values *modules.DepositStateValues) pb.Response {
	//退出列表
	var err error
	//计算余额
	values.TotalAmount -= invokeTokens.Amount
	//判断是否退出列表
	if values.TotalAmount == 0 {
		//加入列表时的时间
		startTime := values.Time.YearDay()
		//当前退出时间
		endTime := time.Now().UTC().YearDay()
		//判断是否已到期
		if endTime-startTime >= 0 {
			//退出全部，即删除cashback
			err = d.cashbackOrForfeitureAllDeposit("Jury", stub, invokeAddr, invokeTokens, invokeAddr, values)
			if err != nil {
				return shim.Success([]byte(err.Error()))
			}
			return shim.Success([]byte("成功退出"))
		} else {
			return shim.Success([]byte("未到期，不能退出列表"))
		}
	} else {
		//TODO 退出一部分，且退出该部分金额后还在列表中
		d.cashbackOrForfeitureSomeDeposit("Jury", stub, invokeAddr, invokeTokens, invokeAddr, values)
		if err != nil {
			return shim.Success([]byte(err.Error()))
		}
		return shim.Success([]byte("成功退出一部分"))
	}
}

//对Developer退保证金的处理
func (d *DepositChaincode) handleDeveloperDepositCashback(stub shim.ChaincodeStubInterface, invokeAddr common.Address, invokeTokens *modules.InvokeTokens, values *modules.DepositStateValues) pb.Response {
	var res pb.Response
	if values.TotalAmount >= depositAmountsForDeveloper {
		//已在列表中
		res = d.handleDeveloperFromList(stub, invokeAddr, invokeTokens, values)
	} else {
		//TODO 不在列表中
		//res = handleCommonDeveloper(stub, invokeAddr, invokeTokens, values)
	}
	return res
}

//Developer已在列表中
func (d *DepositChaincode) handleDeveloperFromList(stub shim.ChaincodeStubInterface, invokeAddr common.Address, invokeTokens *modules.InvokeTokens, values *modules.DepositStateValues) pb.Response {
	//退出列表
	var err error
	//计算余额
	values.TotalAmount -= invokeTokens.Amount
	//判断是否退出列表
	if values.TotalAmount == 0 {
		//加入列表时的时间
		startTime := values.Time.YearDay()
		//当前退出时间
		endTime := time.Now().UTC().YearDay()
		//判断是否已到期
		if endTime-startTime >= 0 {
			//退出全部，即删除cashback
			err = d.cashbackOrForfeitureAllDeposit("Developer", stub, invokeAddr, invokeTokens, invokeAddr, values)
			if err != nil {
				return shim.Success([]byte(err.Error()))
			}
			return shim.Success([]byte("成功退出"))
		} else {
			return shim.Success([]byte("未到期，不能退出列表"))
		}
	} else {
		//TODO 退出一部分，且退出该部分金额后还在列表中
		d.cashbackOrForfeitureSomeDeposit("Jury", stub, invokeAddr, invokeTokens, invokeAddr, values)
		if err != nil {
			return shim.Success([]byte(err.Error()))
		}
		return shim.Success([]byte("成功退出一部分"))
	}
}

//退出列表处理
func (d *DepositChaincode) cashbackOrForfeitureAllDeposit(who string, stub shim.ChaincodeStubInterface, invokeAddr common.Address, invokeTokens *modules.InvokeTokens, forfeitureAddr common.Address, values *modules.DepositStateValues) error {
	//计算总奖励
	totalAward := calculateAllAwards(values.Values)
	//退款总额：本金+利息
	invokeTokens.Amount += totalAward
	//调用从合约把token转到请求地址
	err := stub.PayOutToken(invokeAddr.String(), invokeTokens, 0)
	if err != nil {
		return err
	}
	//移除出列表
	handleMember(who, invokeAddr, stub)
	//删除节点
	err = stub.DelState(forfeitureAddr.String())
	if err != nil {
		return err
	}
	return nil
}

//节点还在列表中，计算这部分提款的余额的奖励
func (d *DepositChaincode) cashbackOrForfeitureSomeDeposit(who string, stub shim.ChaincodeStubInterface, invokeAddr common.Address, invokeTokens *modules.InvokeTokens, forfeitureAddr common.Address, values *modules.DepositStateValues) error {
	//计算这部分余额的奖励
	someAward := calculateSomeAwards(values.Values, invokeTokens)
	//退款+利息
	invokeTokens.Amount += someAward
	//调用从合约把token转到请求地址
	err := stub.PayOutToken(invokeAddr.String(), invokeTokens, 0)
	if err != nil {
		return err
	}
	//序列化
	stateValueMarshalBytes, err := json.Marshal(values)
	if err != nil {
		return err
	}
	//更新状态数据库
	stub.PutState(forfeitureAddr.String(), stateValueMarshalBytes)
	if err != nil {
		return err
	}
	return nil
}

//保证金没收
//handle forfeiture deposit
func (d DepositChaincode) forfeitureDeposit(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	//第一个参数：没收地址；第二个参数：没收数量；第三个参数：角色 (Mediator  Jury  Developer)
	if len(args) != 3 {
		return shim.Success([]byte("Input parameter Success,need three parameters."))
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
	//if invokeAddr.String() != foundationAddress {
	//	return shim.Success([]byte("请求地址不正确，请使用基金会的地址"))
	//}
	//获取没收节点地址
	forfeitureAddr, err := common.StringToAddress(args[0])
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
	ptnAccount, _ := strconv.ParseUint(args[1], 10, 64)
	if err != nil {
		return shim.Success([]byte("String transform to uint64 error:"))
	}
	//判断账户余额和没收请求数量
	if values.TotalAmount < ptnAccount {
		return shim.Success([]byte("Forfeiture too many."))
	}
	asset := modules.NewPTNAsset()
	invokeTokens := &modules.InvokeTokens{
		Amount: ptnAccount,
		Asset:  asset,
	}
	//判断是 Jury 还是 Mediator 还是 developer
	role := args[2]
	res := pb.Response{}
	switch {
	case role == "Mediator":
		//处理没收Mediator保证金
		res = d.handleMediatorForfeitureDeposit(invokeAddr, forfeitureAddr, invokeTokens, values, stub)
	case role == "Jury":
		//处理没收Jury保证金
		res = d.handleJuryForfeitureDeposit(invokeAddr, forfeitureAddr, invokeTokens, values, stub)
	case role == "Developer":
		//处理没收Developer保证金
		res = d.handleDeveloperForfeitureDeposit(invokeAddr, forfeitureAddr, invokeTokens, values, stub)
	}
	return res
}

//处理没收Mediator保证金
func (d *DepositChaincode) handleMediatorForfeitureDeposit(invokeAddr, forfeitureAddr common.Address, tokens *modules.InvokeTokens, values *modules.DepositStateValues, stub shim.ChaincodeStubInterface) pb.Response {
	var err error
	//计算余额
	values.TotalAmount -= tokens.Amount
	//判断是否没收全部，即在列表中移除该节点
	if values.TotalAmount == 0 {
		//没收不考虑是否在规定周期内,其实它肯定是在列表中并已在周期内
		//没收全部，即删除
		err = d.cashbackOrForfeitureAllDeposit("Mediator", stub, invokeAddr, tokens, forfeitureAddr, values)
		if err != nil {
			return shim.Success([]byte(err.Error()))
		}
		return shim.Success([]byte("成功退出"))

	} else if values.TotalAmount < depositAmountsForMediator {
		//没收后余额不在Mediator保证金之内
		return shim.Success([]byte("没收后余额不在Mediator保证金之内"))
	} else {
		//TODO 退出一部分，且退出该部分金额后还在列表中
		d.cashbackOrForfeitureSomeDeposit("Mediator", stub, invokeAddr, tokens, forfeitureAddr, values)
		if err != nil {
			return shim.Success([]byte(err.Error()))
		}
		return shim.Success([]byte("成功没收一部分"))
	}
}

func (d *DepositChaincode) handleJuryForfeitureDeposit(invokeAddr, forfeitureAddr common.Address, tokens *modules.InvokeTokens, values *modules.DepositStateValues, stub shim.ChaincodeStubInterface) pb.Response {
	var err error
	//计算余额
	values.TotalAmount -= tokens.Amount
	//判断是否没收全部，即在列表中移除该节点
	if values.TotalAmount == 0 {
		//没收不考虑是否在规定周期内,其实它肯定是在列表中并已在周期内
		//没收全部，即删除
		err = d.cashbackOrForfeitureAllDeposit("Jury", stub, invokeAddr, tokens, forfeitureAddr, values)
		if err != nil {
			return shim.Success([]byte(err.Error()))
		}
		return shim.Success([]byte("成功退出"))

	} else if values.TotalAmount < depositAmountsForMediator {
		//没收后余额不在Mediator保证金之内
		return shim.Success([]byte("没收后余额不在Mediator保证金之内"))
	} else {
		//TODO 退出一部分，且退出该部分金额后还在列表中
		d.cashbackOrForfeitureSomeDeposit("Jury", stub, invokeAddr, tokens, forfeitureAddr, values)
		if err != nil {
			return shim.Success([]byte(err.Error()))
		}
		return shim.Success([]byte("成功没收一部分"))
	}
}

func (d *DepositChaincode) handleDeveloperForfeitureDeposit(invokeAddr, forfeitureAddr common.Address, tokens *modules.InvokeTokens, values *modules.DepositStateValues, stub shim.ChaincodeStubInterface) pb.Response {
	var err error
	//计算余额
	values.TotalAmount -= tokens.Amount
	//判断是否没收全部，即在列表中移除该节点
	if values.TotalAmount == 0 {
		//没收不考虑是否在规定周期内,其实它肯定是在列表中并已在周期内
		//没收全部，即删除
		err = d.cashbackOrForfeitureAllDeposit("Developer", stub, invokeAddr, tokens, forfeitureAddr, values)
		if err != nil {
			return shim.Success([]byte(err.Error()))
		}
		return shim.Success([]byte("成功退出"))

	} else if values.TotalAmount < depositAmountsForMediator {
		//没收后余额不在Mediator保证金之内
		return shim.Success([]byte("没收后余额不在Mediator保证金之内"))
	} else {
		//TODO 退出一部分，且退出该部分金额后还在列表中
		d.cashbackOrForfeitureSomeDeposit("Developer", stub, invokeAddr, tokens, forfeitureAddr, values)
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
