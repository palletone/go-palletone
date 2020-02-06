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
	"encoding/json"
	"fmt"
	"strings"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/contracts/shim"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"github.com/palletone/go-palletone/dag/errors"
)

type Trace struct {
}

func (p *Trace) Init(stub shim.ChaincodeStubInterface) pb.Response {
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get invoke address\"}"
		return shim.Error(jsonResp)
	}
	err = stub.PutState(symbolsAdmin, []byte(invokeAddr.String()))
	if err != nil {
		return shim.Error("write symbolsAdmin failed: " + err.Error())
	}
	return shim.Success(nil)
}

func (p *Trace) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	f, args := stub.GetFunctionAndParameters()

	switch f {
	case "addProof":
		if len(args) < 5 {
			return shim.Error("need 5 args (Category,Key,Value,Reference,OwnerAddress)")
		}
		if len(args[0]) == 0 {
			return shim.Error("Category is empty")
		}
		if len(args[1]) == 0 {
			return shim.Error("Key is empty")
		}
		if len(args[3]) == 0 {
			return shim.Error("Reference is empty")
		}
		ownerAddr, err := common.StringToAddress(args[4])
		if err != nil {
			return shim.Error("Invalid address string:" + args[4])
		}
		return p.AddProof(stub, args[0], args[1], args[2], args[3], ownerAddr)
	case "delProof":
		if len(args) < 2 {
			return shim.Error("need 2 args (Category,Key)")
		}
		if len(args[0]) == 0 {
			return shim.Error("Category is empty")
		}
		if len(args[1]) == 0 {
			return shim.Error("Key is empty")
		}
		return p.DelProof(stub, args[0], args[1])

	case "getProof":
		if len(args) < 2 {
			return shim.Error("need 2 args (Category,Key)")
		}
		if len(args[0]) == 0 {
			return shim.Error("Category is empty")
		}
		if len(args[1]) == 0 {
			return shim.Error("Key is empty")
		}
		result, err := p.GetProof(stub, args[0], args[1])
		if err != nil {
			return shim.Error(err.Error())
		}
		data, _ := json.Marshal(result)
		return shim.Success(data)

	case "getProofByCategory":
		if len(args) < 1 {
			return shim.Error("need 1 args (Category)")
		}
		if len(args[0]) == 0 {
			return shim.Error("Category is empty")
		}
		result, err := p.GetProofByCategory(stub, args[0])
		if err != nil {
			return shim.Error(err.Error())
		}
		data, _ := json.Marshal(result)
		return shim.Success(data)

	case "getProofByOwner":
		if len(args) < 1 {
			return shim.Error("need 1 args (OwnerAddress)")
		}
		ownerAddr, err := common.StringToAddress(args[0])
		if err != nil {
			return shim.Error("Invalid address string:" + args[0])
		}
		result, err := p.GetProofByOwner(stub, ownerAddr)
		if err != nil {
			return shim.Error(err.Error())
		}
		data, _ := json.Marshal(result)
		return shim.Success(data)

	case "getProofByReference":
		if len(args) < 1 {
			return shim.Error("need 1 args (Reference)")
		}
		if len(args[0]) == 0 {
			return shim.Error("Reference is empty")
		}
		result, err := p.GetProofByReference(stub, args[0])
		if err != nil {
			return shim.Error(err.Error())
		}
		data, _ := json.Marshal(result)
		return shim.Success(data)

	case "setAdmin":
		if len(args) < 1 {
			return shim.Error("need 1 args (PTNAddr)")
		}
		return p.SetAdmin(args[0], stub)

	case "Set":
		if len(args) < 2 {
			return shim.Error("need 2 args (Key, Value)")
		}
		return p.Set(stub, args[0], args[1])
	case "get":
		if len(args) < 1 {
			return shim.Error("need 1 args (Key)")
		}
		return p.Get(stub, args[0])

	default:
		jsonResp := "{\"Error\":\"Unknown function " + f + "\"}"
		return shim.Error(jsonResp)
	}
}

const symbolsAdmin = "admin_"
const symbolsOwner = "owner_"
const SEP = "@"
const OWNER = "owner"

type proof struct {
	Value     string
	Reference string
	OwnerAddr string
}

func (p *Trace) AddProof(stub shim.ChaincodeStubInterface, category, key,
	value, reference string, ownerAddr common.Address) pb.Response {
	//check is 'owner' & 'admin' or not
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get invoke address\"}"
		return shim.Error(jsonResp)
	}
	isAdmin := false
	result, _ := stub.GetState(symbolsOwner + invokeAddr.String())
	if len(result) == 0 {
		admin, _ := getAdmin(stub)
		if admin != invokeAddr.String() {
			return shim.Error("Only Admin or Owner can add")
		}
		isAdmin = true
	}

	ownerInput := ownerAddr.String()
	if isAdmin {
		result, _ := stub.GetState(symbolsOwner + ownerInput)
		// save new owner
		if len(result) == 0 {
			err := stub.PutState(symbolsOwner+ownerInput, []byte("owner"))
			if err != nil {
				return shim.Error("write new owner failed: " + err.Error())
			}
		}
	}

	//
	isModify := false
	var pfOld proof
	saveResult, _ := stub.GetState(category + SEP + key)
	if len(saveResult) != 0 { //modify old proof
		err = json.Unmarshal(saveResult, &pfOld)
		if err != nil {
			return shim.Error(err.Error())
		}
		if !isAdmin && pfOld.OwnerAddr != invokeAddr.String() {
			return shim.Error("Owner only can modify you own proofs")
		}
		isModify = true
	} else { //add new proof
		if !isAdmin && !ownerAddr.Equal(invokeAddr) {
			return shim.Error("Owner only can add proof by your address")
		}
	}

	//check is put state with 'owner' & 'reference' or not
	needPutOwner := true
	needPutReference := true
	if isModify {
		if ownerInput != pfOld.OwnerAddr { //transfer
			if !isAdmin {
				result, _ := stub.GetState(symbolsOwner + ownerInput)
				if len(result) == 0 {
					return shim.Error("OwnerAddress is not owner yet, transfer failed")
				}
			}
			err = stub.DelState(pfOld.OwnerAddr + SEP + category + SEP + key)
			if err != nil {
				return shim.Error("delete old ownerAddr proof failed: " + err.Error())
			}
		} else {
			needPutOwner = false
		}
		if reference != pfOld.Reference {
			err = stub.DelState(pfOld.Reference + SEP + category + SEP + key)
			if err != nil {
				return shim.Error("delete old reference proof failed: " + err.Error())
			}
		} else {
			needPutReference = false
		}
	}
	//add new proof
	pf := proof{Value: value, Reference: reference, OwnerAddr: ownerInput}
	pfJSON, _ := json.Marshal(pf)
	err = stub.PutState(category+SEP+key, pfJSON)
	if err != nil {
		return shim.Error("write category + key proof failed: " + err.Error())
	}
	if needPutOwner {
		err = stub.PutState(ownerInput+SEP+category+SEP+key, pfJSON)
		if err != nil {
			return shim.Error("write ownerAddr proof failed: " + err.Error())
		}
	}
	if needPutReference {
		err = stub.PutState(reference+SEP+category+SEP+key, pfJSON)
		if err != nil {
			return shim.Error("write reference proof failed: " + err.Error())
		}
	}

	return shim.Success([]byte("Success"))
}

func (p *Trace) DelProof(stub shim.ChaincodeStubInterface, category, key string) pb.Response {
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get invoke address\"}"
		return shim.Error(jsonResp)
	}
	result, _ := stub.GetState(symbolsOwner + invokeAddr.String())
	if len(result) == 0 {
		admin, _ := getAdmin(stub)
		if admin != invokeAddr.String() {
			return shim.Error("Only Admin or Owner can delete")
		}
	}

	//
	saveResult, _ := stub.GetState(category + SEP + key)
	if len(saveResult) == 0 {
		return shim.Success([]byte("Not exist"))
	}

	// save proof
	var pf proof
	err = json.Unmarshal(saveResult, &pf)
	if err != nil {
		return shim.Error(err.Error())
	}
	pf.Value = ""
	pfJSON, _ := json.Marshal(pf)
	err = stub.PutState(category+SEP+key, pfJSON)
	if err != nil {
		return shim.Error("write category + key proof failed: " + err.Error())
	}
	err = stub.PutState(pf.OwnerAddr+SEP+category+SEP+key, pfJSON)
	if err != nil {
		return shim.Error("write  ownerAddr proof failed: " + err.Error())
	}
	err = stub.PutState(pf.Reference+SEP+category+SEP+key, pfJSON)
	if err != nil {
		return shim.Error("write reference proof failed: " + err.Error())
	}
	return shim.Success([]byte("Success"))
}

type ProofInfo struct {
	Category  string
	Key       string
	Value     string
	Reference string
	OwnerAddr string
}

func (p *Trace) GetProof(stub shim.ChaincodeStubInterface, category, key string) (*ProofInfo, error) {
	result, _ := stub.GetState(category + SEP + key)
	if len(result) == 0 {
		return nil, fmt.Errorf("Not exist")
	}
	var pf proof
	err := json.Unmarshal(result, &pf)
	if err != nil {
		return nil, err
	}

	pfInfo := ProofInfo{Category: category, Key: key,
		Value: pf.Value, Reference: pf.Reference, OwnerAddr: pf.OwnerAddr}

	return &pfInfo, nil
}

func getProofByCategory(stub shim.ChaincodeStubInterface, key string) []*ProofInfo {
	KVs, _ := stub.GetStateByPrefix(key)
	pfInfos := make([]*ProofInfo, 0, len(KVs))
	for _, oneKV := range KVs {
		var pf proof
		err := json.Unmarshal(oneKV.Value, &pf)
		if err != nil {
			continue
		}
		keys := strings.Split(oneKV.Key, SEP)
		if len(keys) != 2 {
			continue
		}
		pfInfos = append(pfInfos, &ProofInfo{Category: keys[0], Key: keys[1],
			Value: pf.Value, Reference: pf.Reference, OwnerAddr: pf.OwnerAddr})
	}
	return pfInfos
}

func getProofByOther(stub shim.ChaincodeStubInterface, key string) []*ProofInfo {
	KVs, _ := stub.GetStateByPrefix(key)
	pfInfos := make([]*ProofInfo, 0, len(KVs))
	for _, oneKV := range KVs {
		var pf proof
		err := json.Unmarshal(oneKV.Value, &pf)
		if err != nil {
			continue
		}
		keys := strings.Split(oneKV.Key, SEP)
		if len(keys) != 3 {
			continue
		}
		pfInfos = append(pfInfos, &ProofInfo{Category: keys[1], Key: keys[2],
			Value: pf.Value, Reference: pf.Reference, OwnerAddr: pf.OwnerAddr})
	}
	return pfInfos
}

func (p *Trace) GetProofByCategory(stub shim.ChaincodeStubInterface, category string) ([]*ProofInfo, error) {
	result := getProofByCategory(stub, category)
	if len(result) == 0 {
		return nil, fmt.Errorf("Not exist")
	}
	return result, nil
}

func (p *Trace) GetProofByOwner(stub shim.ChaincodeStubInterface, ownerAddr common.Address) ([]*ProofInfo, error) {
	result := getProofByOther(stub, ownerAddr.String())
	if len(result) == 0 {
		return nil, fmt.Errorf("Not exist")
	}
	return result, nil
}

func (p *Trace) GetProofByReference(stub shim.ChaincodeStubInterface, reference string) ([]*ProofInfo, error) {
	result := getProofByOther(stub, reference)
	if len(result) == 0 {
		return nil, fmt.Errorf("Not exist")
	}
	return result, nil
}

func (p *Trace) SetAdmin(ptnAddr string, stub shim.ChaincodeStubInterface) pb.Response {
	//only admin can set
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get invoke address\"}"
		return shim.Error(jsonResp)
	}
	admin, err := getAdmin(stub)
	if err != nil {
		return shim.Error(err.Error())
	}
	if admin != invokeAddr.String() {
		return shim.Error("Only admin can set")
	}
	err = stub.PutState(symbolsAdmin, []byte(ptnAddr))
	if err != nil {
		return shim.Error("write symbolsAdmin failed: " + err.Error())
	}

	return shim.Success([]byte("Success"))
}

func getAdmin(stub shim.ChaincodeStubInterface) (string, error) {
	result, _ := stub.GetState(symbolsAdmin)
	if len(result) == 0 {
		return "", errors.New("Need set Owner")
	}

	return string(result), nil
}

func (p *Trace) Get(stub shim.ChaincodeStubInterface, key string) pb.Response {
	result, _ := stub.GetState(key)
	return shim.Success(result)
}
func (p *Trace) Set(stub shim.ChaincodeStubInterface, key string, value string) pb.Response {
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get invoke address\"}"
		return shim.Error(jsonResp)
	}
	admin, err := getAdmin(stub)
	if err != nil {
		return shim.Error(err.Error())
	}
	if admin != invokeAddr.String() {
		return shim.Error("Only admin can set")
	}

	err = stub.PutState(key, []byte(value))
	if err != nil {
		return shim.Error(fmt.Sprintf("PutState failed: %s", err.Error()))
	}
	return shim.Success([]byte("Success"))
}

func main() {
	err := shim.Start(new(Trace))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
