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
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/contracts/shim"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"github.com/palletone/go-palletone/dag/errors"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/shopspring/decimal"

	"github.com/palletone/adaptor"
)

type FabSample struct {
}

func (p *FabSample) Init(stub shim.ChaincodeStubInterface) pb.Response {
	args := stub.GetStringArgs()
	if len(args) < 1 {
		return shim.Error("need 1 args (MapContractAddr)")
	}

	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get invoke address\"}"
		return shim.Error(jsonResp)
	}
	err = stub.PutState(symbolsOwner, []byte(invokeAddr.String()))
	if err != nil {
		return shim.Error("write symbolsOwner failed: " + err.Error())
	}

	err = stub.PutState(symbolsFabChaincode, []byte(args[0]))
	if err != nil {
		return shim.Error("write symbolsFabChaincode failed: " + err.Error())
	}

	return shim.Success(nil)
}

func (p *FabSample) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	f, args := stub.GetFunctionAndParameters()

	switch f {
	case "setOwner":
		if len(args) < 1 {
			return shim.Error("need 1 args (PTNAddr)")
		}
		return p.SetOwner(args[0], stub)
	case "setFabChaincodeID":
		if len(args) < 1 {
			return shim.Error("need 1 args (FabricChaincodeID)")
		}
		return p.SetFabContract(args[0], stub)

	case "payoutPTNByTxID":
		if len(args) < 2 {
			return shim.Error("need 1 args (FabTransferTxID, PTNAddr)")
		}
		ptnAddr, err := common.StringToAddress(args[1])
		if err != nil {
			return shim.Error("Invalid address string:" + args[0])
		}
		return p.PayoutPTNByTxID(args[0], ptnAddr, stub)

	case "invokeFabChaincode":
		return p.InvokeFabChaincode(stub)

	case "withdrawAmount":
		if len(args) < 2 {
			return shim.Error("need 2  args (PTNAddress,PTNAmount)")
		}
		withdrawAddr, err := common.StringToAddress(args[0])
		if err != nil {
			return shim.Error("Invalid address string:" + args[0])
		}
		amount, err := decimal.NewFromString(args[1])
		if err != nil {
			return shim.Error("Invalid amount:" + args[1])
		}
		return p.WithdrawAmount(stub, withdrawAddr, amount)

	case "getPayout":
		if len(args) < 1 {
			return shim.Error("need 1 args (FabTransferTxID)")
		}
		return p.GetPayout(args[0], stub)

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

//todo modify conforms 15
const Confirms = uint(15)

const symbolsOwner = "owner_"
const symbolsFabChaincode = "fabchaincode_"

const symbolsPayout = "payout_"
const symbolsInvoke = "invoke_"

func (p *FabSample) SetOwner(ptnAddr string, stub shim.ChaincodeStubInterface) pb.Response {
	//only owner can set
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get invoke address\"}"
		return shim.Error(jsonResp)
	}
	owner, err := getOwner(stub)
	if err != nil {
		return shim.Error(err.Error())
	}
	if owner != invokeAddr.String() {
		return shim.Error("Only owner can set")
	}
	err = stub.PutState(symbolsOwner, []byte(ptnAddr))
	if err != nil {
		return shim.Error("write symbolsOwner failed: " + err.Error())
	}
	return shim.Success([]byte("Success"))
}

func getOwner(stub shim.ChaincodeStubInterface) (string, error) {
	result, _ := stub.GetState(symbolsOwner)
	if len(result) == 0 {
		return "", errors.New("Need set Owner")
	}

	return string(result), nil
}

func (p *FabSample) SetFabContract(contractAddr string, stub shim.ChaincodeStubInterface) pb.Response {
	//only owner can set
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get invoke address\"}"
		return shim.Error(jsonResp)
	}
	owner, err := getOwner(stub)
	if err != nil {
		return shim.Error(err.Error())
	}
	if owner != invokeAddr.String() {
		return shim.Error("Only owner can set")
	}

	err = stub.PutState(symbolsFabChaincode, []byte(contractAddr))
	if err != nil {
		return shim.Error("write symbolsFabChaincode failed: " + err.Error())
	}

	return shim.Success([]byte("Success"))
}

func getFabContract(stub shim.ChaincodeStubInterface) (string, error) {
	result, _ := stub.GetState(symbolsFabChaincode)
	if len(result) == 0 {
		return "", errors.New("Need set MapContractAddr")
	}

	return string(result), nil
}

func (p *FabSample) PayoutPTNByTxID(txID string, ptnAddr common.Address, stub shim.ChaincodeStubInterface) pb.Response {
	//
	if "0x" == txID[0:2] || "0X" == txID[0:2] {
		txID = txID[2:]
	}
	result, _ := stub.GetState(symbolsPayout + txID)
	if len(result) != 0 {
		log.Debugf("The tx has been payout")
		return shim.Error("The tx has been payout")
	}

	//get sender receiver amount
	txIDByte, err := hex.DecodeString(txID)
	if err != nil {
		log.Debugf("txid invalid: %s", err.Error())
		return shim.Error(fmt.Sprintf("txid invalid: %s", err.Error()))
	}
	txResult, err := getTxDetails(txIDByte, stub)//todo
	if err != nil {
		log.Debugf("GetTxDetails failed: %s", err.Error())
		return shim.Error(err.Error())
	}

	//check tx status
	if !txResult.IsSuccess {
		log.Debugf("The tx is failed")
		return shim.Error("The tx is failed")
	}
	//check receiver, must be ptnmap contract address
	fabContract, err := getFabContract(stub)
	if err != nil {
		log.Debugf("getFabContract failed: %s", err.Error())
		return shim.Error(err.Error())
	}
	if strings.ToLower(txResult.TargetAddress) != fabContract {
		log.Debugf("strings.ToLower(txResult.To): %s, fabContract: %s ", strings.ToLower(txResult.TargetAddress), fabContract)
		return shim.Error("Not send to the TargetAddress")
	}

	//check token amount
	var args []string
	err = json.Unmarshal(txResult.TxRawData, &args)
	if err != nil {
		log.Debugf("Unmarshal args failed: %s", err.Error())
		return shim.Error(err.Error())
	}
	if len(args) != 4 {
		log.Debugf("args'len is not 4: %d", len(args))
		return shim.Error(fmt.Sprintf("args'len is not 4: %d", len(args)))
	}
	if args[0] != "invoke"{
		log.Debugf("func is not invoke: %s", args[0])
		return shim.Error(fmt.Sprintf("func is not invoke: %s", args[0]))
	}
	if args[2] != "B" {
		log.Debugf("not send to B: %s", args[2])
		return shim.Error(fmt.Sprintf("not send to B: %s", args[2]))
	}
	amount := args[3]
	amt,_ := strconv.ParseUint(amount, 10, 64) //Token's decimal is 8, PTN's decimal is 8
	if amt == 0 {
		log.Debugf("Amount is 0")
		return shim.Error("Amount is 0")
	}

	//save payout history
	err = stub.PutState(symbolsPayout+txID, []byte(ptnAddr.String()+"-"+amount))
	if err != nil {
		log.Debugf("write symbolsPayout failed: %s", err.Error())
		return shim.Error("write symbolsPayout failed: " + err.Error())
	}

	//payout
	//asset := modules.NewPTNAsset()
	asset, _ := modules.StringToAsset("PTN")
	amtToken := &modules.AmountAsset{Amount: amt, Asset: asset}
	err = stub.PayOutToken(ptnAddr.String(), amtToken, 0)
	if err != nil {
		log.Debugf("PayOutToken failed: %s", err.Error())
		return shim.Error("PayOutToken failed: " + err.Error())
	}

	return shim.Success([]byte("Success"))
}

func getTxDetails(txID []byte, stub shim.ChaincodeStubInterface) (*adaptor.GetContractInvokeTxOutput, error) {
	input := adaptor.GetContractInvokeTxInput{TxID: txID}
	inputBytes, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	//
	result, err := stub.OutChainCall("fabric", "GetContractInvokeTx", inputBytes)
	if err != nil {
		return nil, errors.New("GetTransferTx error: " + err.Error())
	}
	log.Debugf("result : %s", string(result))

	//
	var output adaptor.GetContractInvokeTxOutput
	err = json.Unmarshal(result, &output)
	if err != nil {
		return nil, err
	}
	return &output, nil
}
func (p *FabSample) InvokeFabChaincode(stub shim.ChaincodeStubInterface) pb.Response {
	invokeTokens, err := stub.GetInvokeTokens()
	if err != nil {
		jsonResp := "{\"Error\":\"GetInvokeTokens failed\"}"
		return shim.Error(jsonResp)
	}
	_, contractAddr := stub.GetContractID()
	ptnNum := uint64(0)
	for i := 0; i < len(invokeTokens); i++ {
		if invokeTokens[i].Asset.AssetId == modules.PTNCOIN {
			if invokeTokens[i].Address == contractAddr {
				ptnNum += invokeTokens[i].Amount
			}
		}
	}
	if ptnNum == 0 {
		log.Debugf("send contract ptnNum is 0")
		return shim.Error("send contract ptnNum is 0")
	}

	//check receiver, must be ptnmap contract address
	fabContract, err := getFabContract(stub)
	if err != nil {
		log.Debugf("getFabContract failed: %s", err.Error())
		return shim.Error(err.Error())
	}

	amount := fmt.Sprintf("%d", ptnNum)
	//query
	balance, err := queryFabChaincode(fabContract,stub)
	if err != nil {
		log.Debugf("queryFabChaincode failed: %s", err.Error())
		return shim.Error(err.Error())
	}
	if ptnNum > balance {
		log.Debugf("balance is not enough, ptnNum: %d", ptnNum)
		return shim.Error(fmt.Sprintf("balance is not enough, ptnNum: %d", ptnNum))
	}

	//invoke
	txID,err:= invokeFabChaincode(fabContract, amount, stub)
	log.Debugf("txID: %s", txID)

	err = stub.PutState(symbolsInvoke+stub.GetTxID(), []byte(amount))//todo
	if err != nil {
		log.Debugf("write symbolsInvoke failed: %s", err.Error())
		return shim.Error("write symbolsInvoke failed: " + err.Error())
	}

	return shim.Success([]byte("Success"))
}

func queryFabChaincode(fabContract string, stub shim.ChaincodeStubInterface) (uint64,error)  {
	inputQuery := &adaptor.QueryContractInput{
		ContractAddress:fabContract,
		Function:"query",
		Args:[][]byte{[]byte("B")},
	}
	inputQueryBytes, err := json.Marshal(inputQuery)
	if err != nil {
		return 0, err
	}
	//
	resultQuery, err := stub.OutChainCall("fabric", "QueryContract",
		inputQueryBytes)
	if err != nil {
		return 0, errors.New("QueryContract error: " + err.Error())
	}
	log.Debugf("QueryContract is OK")

	//
	var outputQuery adaptor.QueryContractOutput
	err = json.Unmarshal(resultQuery, &outputQuery)
	if err != nil {
		return 0, err
	}
	log.Debugf("outputQuery.QueryResult %s", string(outputQuery.QueryResult))
	balance,err:= strconv.ParseUint(string(outputQuery.QueryResult), 10, 64)
	if err != nil {
		return 0, err
	}
	return balance, nil
}

func invokeFabChaincode(fabContract, amount string, stub shim.ChaincodeStubInterface) (string,error)  {
	args := [][]byte{[]byte("B"),[]byte("A"), []byte(amount)}
	inputCreateInvoke := adaptor.CreateContractInvokeTxInput{
		ContractAddress:fabContract,
		Function:"invoke",Args:args}
	inputCreateInvokeBytes, err := json.Marshal(inputCreateInvoke)
	if err != nil {
		return "", err
	}
	//
	resultCreateInvoke, err := stub.OutChainCall("fabric", "CreateContractInvokeTx",
		inputCreateInvokeBytes)
	if err != nil {
		return "", errors.New("CreateContractInvokeTx error: " + err.Error())
	}
	log.Debugf("CreateContractInvokeTx is OK")

	//
	var outputCreateInvoke adaptor.CreateContractInvokeTxOutput
	err = json.Unmarshal(resultCreateInvoke, &outputCreateInvoke)
	if err != nil {
		return "", err
	}

	//sign tx
	inputSign := adaptor.SignTransactionInput{
		Transaction:outputCreateInvoke.RawTransaction}
	inputSignBytes, err := json.Marshal(inputSign)
	if err != nil {
		return "", err
	}
	//
	resultSign, err := stub.OutChainCall("fabric", "SignTransaction",
		inputSignBytes)
	if err != nil {
		return "", errors.New("SignTransaction error: " + err.Error())
	}
	log.Debugf("SignTransaction is OK")

	//
	var outputSign adaptor.SignTransactionOutput
	err = json.Unmarshal(resultSign, &outputSign)
	if err != nil {
		return "", err
	}

	//send tx
	inputSend := adaptor.SendTransactionInput{
		Transaction:outputSign.SignedTx,
		Extra:[]byte("invoke"),//Must set
	}
	inputSendBytes, err := json.Marshal(inputSend)
	if err != nil {
		return "", err
	}
	//
	resultSend, err := stub.OutChainCall("fabric", "SendTransaction",
		inputSendBytes)
	if err != nil {
		return "", errors.New("SendTransaction error: " + err.Error())
	}
	log.Debugf("SendTransaction is OK")

	//
	var outputSend adaptor.GetContractInvokeTxOutput
	err = json.Unmarshal(resultSend, &outputSend)
	if err != nil {
		return "", err
	}
	log.Debugf("outputSend.TxID %s", string(outputSend.TxID))
	return string(outputSend.TxID), nil
}

func (p *FabSample) GetPayout(txID string, stub shim.ChaincodeStubInterface) pb.Response {
	if "0x" == txID[0:2] || "0X" == txID[0:2] {
		txID = txID[2:]
	}
	result, _ := stub.GetState(symbolsPayout + txID)
	return shim.Success(result)
}

func (p *FabSample) Get(stub shim.ChaincodeStubInterface, key string) pb.Response {
	result, _ := stub.GetState(key)
	return shim.Success(result)
}

func (p *FabSample) WithdrawAmount(stub shim.ChaincodeStubInterface, withdrawAddr common.Address, amount decimal.Decimal) pb.Response {
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get invoke address\"}"
		return shim.Error(jsonResp)
	}
	owner, err := getOwner(stub)
	if err != nil {
		return shim.Error(err.Error())
	}
	if owner != invokeAddr.String() {
		return shim.Error("Only owner can withdraw")
	}

	//contractAddr
	amount = amount.Mul(decimal.New(100000000, 0))
	amtToken := &modules.AmountAsset{Amount: uint64(amount.IntPart()), Asset: modules.NewPTNAsset()}
	err = stub.PayOutToken(withdrawAddr.String(), amtToken, 0)
	if err != nil {
		log.Debugf("PayOutToken failed: %s", err.Error())
		return shim.Error("PayOutToken failed: " + err.Error())
	}

	return shim.Success([]byte("Success"))
}

func main() {
	err := shim.Start(new(FabSample))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
