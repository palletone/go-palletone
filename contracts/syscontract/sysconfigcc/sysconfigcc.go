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
	"github.com/palletone/go-palletone/contracts/shim"
	"github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/core"
	"encoding/json"
	"fmt"
)

type SysConfigChainCode struct {
}

func (s *SysConfigChainCode) Init(stub shim.ChaincodeStubInterface) peer.Response {
	log.Info("*** SysConfigChainCode system contract init ***")
	return shim.Success([]byte("success"))
}

func (s *SysConfigChainCode) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	funcName, args := stub.GetFunctionAndParameters()
	switch funcName {
	case "getAllSysParamsConf":
		log.Info("Start getAllSysParamsConf Invoke")
		allValBytes,err := s.getAllSysParamsConf(stub)
		if err != nil {
			return shim.Error("getAllSysParamsConf err: "+err.Error())
		}
		return shim.Success(allValBytes)
	//case "getOldSysParamByKey":
	//	log.Info("Start getOldSysParamByKey Invoke")
	//	oldValByte, err :=s.getOldSysParamByKey(stub,args[0])
	//	if err != nil {
	//		return shim.Error("getOldSysParamByKey err: "+err.Error())
	//	}
	//	return shim.Success(oldValByte)
	//case "getCurrSysParamByKey":
	//	log.Info("Start getCurrSysParamByKey")
	//	currValByte,err := s.getCurrSysParamByKey(stub,args[0])
	//	if err != nil {
	//		return shim.Error("getCurrSysParamByKey err: "+err.Error())
	//	}
	//	return shim.Success(currValByte)
	case "getSysParamValByKey":
		log.Info("Start getSysParamValByKey Invoke")
		val, err := s.getSysParamValByKey(stub,args[0])
		if err != nil {
			return shim.Error("getSysParamValByKey err: "+err.Error())
		}
		return shim.Success(val)
	case "updateSysParamWithoutVote":
		log.Info("Start updateSysParamWithoutVote Invoke")
		resultByte, err := s.updateSysParamWithoutVote(stub,args)
		if err != nil {
			return shim.Error("updateSysParamWithoutVote err: "+err.Error())
		}
		return shim.Success(resultByte)
	//case "updateSysParamWithVote":
	//	log.Info("Start updateSysParamWithVote Invoke")
	//	resultByte,err := s.updateSysParamWithVote(stub,args)
	//	if err != nil {
	//		return shim.Error("updateSysParamWithVote err: "+err.Error())
	//	}
	//	return shim.Success(resultByte)
	default:
		log.Error("Invoke funcName err: ","error",funcName)
		return shim.Error("Invoke funcName err: " + funcName)
	}
}

func (s *SysConfigChainCode) getAllSysParamsConf(stub shim.ChaincodeStubInterface) ([]byte,error){
	sysVal,err := stub.GetState("sysConf")
	if err != nil {
		return nil,err
	}
	return sysVal,nil
}

func (s *SysConfigChainCode) getOldSysParamByKey(stub shim.ChaincodeStubInterface,key string) ([]byte, error){
	return []byte("one value"),nil
}

func (s *SysConfigChainCode) getCurrSysParamByKey(stub shim.ChaincodeStubInterface,key string) ([]byte, error){
	return []byte("curr value"),nil
}

func (s *SysConfigChainCode) updateSysParamWithoutVote(stub shim.ChaincodeStubInterface,args []string) ([]byte, error){
	invokeFromAddr,err := stub.GetInvokeAddress()
	if err != nil {
		return nil,err
	}
	//基金会地址
	foundationAddress, _ := stub.GetSystemConfig("FoundationAddress")
	if invokeFromAddr != foundationAddress {
		return nil, fmt.Errorf("Only foundation can call this function")
	}
	key := args[0]
	newValue := args[1]
	oldValue, err := stub.GetState(args[0])
	if err != nil {
		return nil,err
	}
	err = stub.PutState(key,[]byte(newValue))
	if err != nil {
		return nil,err
	}
	sysValByte,err := stub.GetState("sysConf")
	if err != nil {
		return nil,err
	}
	sysVal := &core.SystemConfig{}
	err = json.Unmarshal(sysValByte,sysVal)
	if err != nil {
		return nil,err
	}
	if key ==  "DepositAmountForJury"{
		sysVal.DepositAmountForJury = newValue
	}
	sysValByte,err = json.Marshal(sysVal)
	if err != nil {
		return nil,err
	}
	err = stub.PutState("sysConf",sysValByte)
	if err != nil {
		return nil,err
	}
	return []byte("update value from "+string(oldValue)+" to "+newValue),nil
}

func (s *SysConfigChainCode) updateSysParamWithVote(stub shim.ChaincodeStubInterface,args []string) ([]byte, error){
	return []byte("with vote"),nil
}

func (s *SysConfigChainCode) getSysParamValByKey(stub shim.ChaincodeStubInterface,key string) ([]byte, error){
	val, err := stub.GetState(key)
	if err != nil {
		return nil,err
	}
	return val,nil
}