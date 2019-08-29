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

package main

import (
	"fmt"
	"github.com/palletone/adaptor"
	"github.com/palletone/go-palletone/contracts/shim"
	"github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
)

type SimpleChaincode struct {
}

var is adaptor.ISmartContract

// 测试客户端参数，以及msg0的参数
// 获取客户端两个参数，第一个作为key,第二个作为value
// 获取部署请求地址，作为key，value为客户端第二个参数
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) peer.Response {
	fmt.Println("Init=========================================================")
	args := stub.GetStringArgs()
	_, err := strconv.Atoi(args[1])
	if err != nil {
		return shim.Success([]byte(err.Error()))
	}
	err = stub.PutState(args[0], []byte(args[1]))
	if err != nil {
		return shim.Error(err.Error())
	}

	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		return shim.Error(err.Error())
	}
	err = stub.PutState(invokeAddr.String(), []byte(args[1]))
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

//进行各种合约测试
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	function, args := stub.GetFunctionAndParameters()
	switch function {
	//  获取合约初始化第一个参数的值
	//  ["GetValueWithKey","A"]
	case "GetValueWithKey":
		fmt.Println("GetValueWithKey")
		byte, err := stub.GetState(args[0])
		if err != nil {
			return shim.Error(err.Error())
		}
		return shim.Success(byte)
		//  获取合约初始化发布者的初始化的值
		//  ["GetValueWithInvokeAddress"]
	case "GetValueWithInvokeAddress":
		invokeAddr, err := stub.GetInvokeAddress()
		if err != nil {
			return shim.Error(err.Error())
		}
		byte, err := stub.GetState(invokeAddr.String())
		if err != nil {
			return shim.Error(err.Error())
		}
		return shim.Success(byte)
		//  通过申请大内存导致合约触发OOM
		//  ["OutOfMemory","10000000000"]
	case "OutOfMemory":
		a, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Println("strconv.Atoi err: ", err.Error())
			return shim.Error(err.Error())
		}
		outOfMemory := make([]string, a)
		fmt.Println("OutOfMemory: ", outOfMemory)
		return shim.Success([]byte(outOfMemory[a]))
		//  除以 0 导致异常
		//  ["DivideByZero","10","0"]
	case "DivideByZero":
		a, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Println("strconv.Atoi err: ", err.Error())
			return shim.Error(err.Error())
		}
		b, err := strconv.Atoi(args[1])
		if err != nil {
			fmt.Println("strconv.Atoi err: ", err.Error())
			return shim.Error(err.Error())
		}
		v := a / b
		fmt.Println("DivideByZero: ", v)
		return shim.Success([]byte(strconv.Itoa(v)))
		//  获取百度首页
		//  ["GetBAIDUHomePage"]
	case "GetBAIDUHomePage":
		res, err := http.Get("https://www.baidu.com")
		if err != nil {
			fmt.Println("http.Get err: ", err.Error())
			return shim.Error(err.Error())
		}
		defer res.Body.Close()
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			fmt.Println("ioutil.ReadAll err: ", err.Error())
			return shim.Error(err.Error())
		}
		return shim.Success(body)
		//  越界异常
		//  ["IndexOutOfRange","10","10"]
	case "IndexOutOfRange":
		a, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Println("strconv.Atoi err: ", err.Error())
			return shim.Error(err.Error())
		}
		b, err := strconv.Atoi(args[1])
		if err != nil {
			fmt.Println("strconv.Atoi err: ", err.Error())
			return shim.Error(err.Error())
		}
		slice10 := make([]int, a)
		value := slice10[b]
		fmt.Println("value = ", value)
		return shim.Success([]byte(strconv.Itoa(value)))
		//  无限循环
		//  ["ForLoop"]
	case "ForLoop":
		i := 0
		for {
			i++
			//fmt.Println("for loop")
		}
		fmt.Println("=========Println=========", i)
		return shim.Success(nil)
		//  将一个网页内容保存到容器中
		//  ["WriteHomePageToContainer","https://www.baidu.com"]
	case "WriteHomePageToContainer":
		res, err := http.Get(args[0])
		if err != nil {
			fmt.Println("http.Get err: ", err.Error())
			return shim.Error(err.Error())
		}
		defer res.Body.Close()
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			fmt.Println("ioutil.ReadAll err: ", err.Error())
			return shim.Error(err.Error())
		}
		fileObj, err := os.OpenFile("homepage.html", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			fmt.Println("os.OpenFile err: ", err.Error())
			return shim.Error(err.Error())
		}
		defer fileObj.Close()
		_, err = fileObj.Write(body)
		if err != nil {
			fmt.Println("fileObj.Write err: ", err.Error())
			return shim.Error(err.Error())
		}
		return shim.Success(body)
	default:
		return shim.Error("Please enter valid function name.")
	}
	//return shim.Success([]byte("Please enter valid function name."))
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Test_Simple chaincode: %s", err)
	}
}
