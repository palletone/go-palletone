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
	"github.com/palletone/go-palletone/contracts/shim"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"strconv"
)

var depositChaincode = new(DepositChaincode)

type DepositChaincode struct {
	DepositContractAddress string
	DepositAmount          uint64
	DepositRate            float64
	FoundationAddress      string
}

func (d *DepositChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("***system contract init about DepositChaincode***")
	//获取配置文件
	depositConfigBytes, err := stub.GetDepositConfig()
	if err != nil {
		fmt.Println("deposit error: ", err.Error())
		return shim.Error(err.Error())
	}
	err = json.Unmarshal(depositConfigBytes, depositChaincode)
	if err != nil {
		fmt.Println("unmarshal depositConfigBytes error ", err)
		return shim.Error(err.Error())
	}
	//fmt.Printf("DepositChaincode=%#v\n\n", depositChaincode)
	return shim.Success([]byte("ok"))
}

func (d *DepositChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	funcName, args := stub.GetFunctionAndParameters()
	switch funcName {
	case "DepositWitnessPay":
		//交付保证金
		//handle witness pay
		//void deposit_witness_pay(const witness_object& wit, token_type amount)

		//获取用户地址
		//userAddr, err := stub.GetPayToContractAddr()
		//if err != nil {
		//	fmt.Println("GetPayToContractAddr error: ", err.Error())
		//	return shim.Error(err.Error())
		//}
		//fmt.Println("GetPayToContractAddr=", string(userAddr))
		////获取 Token 数量
		//tokenAmount, err := stub.GetPayToContractTokens()
		//if err != nil {
		//	fmt.Println("GetPayToContractTokens error: ", err.Error())
		//}
		//fmt.Println("GetPayToContractTokens=", string(tokenAmount))
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
	//判断参数是否准确，第一个参数是陪审员账户，第二个参数是Tokens
	if len(args) != 2 {
		return shim.Error("input error: need two args (witnessAddr and ptnAmount)")
	}

	//将 string 转 uint64
	ptnAccount, err := strconv.ParseUint(args[1], 10, 64)
	if err != nil {
		return shim.Error("ptnAccount input error: " + err.Error())
	}
	//与保证金合约设置的数量比较
	if ptnAccount < depositChaincode.DepositAmount {
		fmt.Println("input ptnAmount less than deposit amount.")
		return shim.Error("input ptnAmount less than deposit amount.")
	}
	//TODO 这里需要对msg0对象，获取其中的付款的数量，以来和参数比较是否大于或等于，否则返回出错

	//获取一下该用户下的账簿情况
	accBalByte, err := stub.GetState(args[0])
	if err != nil {
		return shim.Error("get account balance from ledger error: " + err.Error())
	}
	//账户不存在，第一次参与
	if accBalByte == nil {
		//写入写集
		stub.PutState(args[0], []byte(args[1]))
		return shim.Success([]byte("ok"))
	}
	//账户已存在，进行信息的更新操作
	accBalStr := string(accBalByte)
	//将 string 转 uint64
	accBal, err := strconv.ParseUint(accBalStr, 10, 64)
	if err != nil {
		return shim.Error("string parse to uint64 error: " + err.Error())
	}
	//fmt.Println("获取账户余额=", accBal)
	result := accBal + ptnAccount
	resultStr := strconv.FormatUint(result, 10)
	//写入写集
	stub.PutState(args[0], []byte(resultStr))
	return shim.Success([]byte("ok"))
}

//保证金退还
//handle cashback rewards
func (d *DepositChaincode) depositCashback(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	//有可能从mediator 退出成为 jury,把金额退出一半或者一些
	// 判断参数是否准确，第一个参数是陪审员账户，第二个参数是Tokens
	if len(args) != 2 {
		return shim.Error("input error: need two args (witnessAddr and ptnAmount)")
	}

	//TODO 这里触发退出金额的交易生成，合约到陪审员的交易，返回什么结果呢？数量，对数量进行更新或者和参数进行比较，这里可能还要涉及利率的给予
	//如果成功则进行读写集更新

	//将 string 转 uint64
	ptnAccount, err := strconv.ParseUint(args[1], 10, 64)
	if err != nil {
		return shim.Error("ptnAccount input error: " + err.Error())
	}
	//TODO 获取一下该用户下的账簿情况
	accBalByte, err := stub.GetState(args[0])
	if err != nil {
		return shim.Error("get account balance from ledger error: " + err.Error())
	}
	if accBalByte == nil {
		return shim.Error("your deposit does not exist.")
	}
	accBalStr := string(accBalByte)
	//将 string 转 uint64
	accBal, err := strconv.ParseUint(accBalStr, 10, 64)
	if err != nil {
		return shim.Error("string parse to uint64 error: " + err.Error())
	}
	//fmt.Println("accbac", accBalStr)
	//fmt.Println(ptnAccount)
	if accBal < ptnAccount {
		return shim.Error("deposit does not enough.")
	}
	//fmt.Println("lalalallal")
	result := accBal - ptnAccount
	resultStr := strconv.FormatUint(result, 10)
	//写入写集
	stub.PutState(args[0], []byte(resultStr))
	return shim.Success([]byte("ok"))
}

//保证金没收
//handle forfeiture deposit
func (d DepositChaincode) forfeitureDeposit(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	//参数是陪审员的账户
	if len(args) != 1 {
		return shim.Error("input error: only need one arg (witnessAddr)")
	}

	//TODO 这里会触发一个交易，把陪审员的保证金转给基金会的账户，可能把数量返回，写入写集

	//直接把保证金没收
	err := stub.DelState(args[0])
	if err != nil {
		return shim.Error("does not delete state")
	}
	return shim.Success([]byte("ok"))
}
