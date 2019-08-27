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
 * @author PalletOne core developers <dev@pallet.one>
 * @date 2018
 */
package ethadaptor

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/palletone/eth-adaptor/bind"

	"github.com/palletone/adaptor"
)

func convertContractParams(paramsNew *[]interface{}, parsed *abi.ABI, method string, params [][]byte) {
	var theMethod abi.Method
	if "" == method {
		theMethod = parsed.Constructor
	} else {
		theMethod = parsed.Methods[method]
	}

	if len(params) == len(theMethod.Inputs) {
		for i := range params {
			switch theMethod.Inputs[i].Type.T {
			case abi.IntTy:
				fallthrough
			case abi.UintTy:
				paramInt := new(big.Int)
				paramInt.SetString(string(params[i]), 10)
				*paramsNew = append(*paramsNew, paramInt)
			case abi.BoolTy:
				paramBool, _ := strconv.ParseBool(string(params[i]))
				*paramsNew = append(*paramsNew, paramBool)
			case abi.StringTy:
				*paramsNew = append(*paramsNew, string(params[i]))

			case abi.SliceTy: //#zxl#
				//				//client 	[]--->string--->[][i]byte...[]byte(arg)
				//				//chaincode []byte--->string 			agrs[i]
				//				//adaptor	string--->[]string 			json
				//				fmt.Println(parsed.Methods[method].Inputs[i].Type.Elem.T)
				//				strArray := arg.([]string)
				//				for i := range strArray {
				//					fmt.Println(strArray[i])
				//				}
				fmt.Println("Not support")
			case abi.ArrayTy: //#zxl#
			case abi.AddressTy:
				paramBytes := common.HexToAddress(string(params[i]))
				*paramsNew = append(*paramsNew, paramBytes)
			case abi.FixedBytesTy:
				str := string(params[i])
				if "0x" == str[0:2] {
					str = str[2:]
				}
				paramBytes := common.Hex2Bytes(str)
				inputSize := parsed.Methods[method].Inputs[i].Type.Size
				if len(paramBytes) == inputSize && inputSize == 32 {
					//switch inputSize {
					//case 32:
					//byte32 := new([32]byte)
					var byte32 [32]byte
					for j := 0; j < len(paramBytes); j++ {
						byte32[j] = paramBytes[j]
					}
					*paramsNew = append(*paramsNew, byte32)
					//}
				}
			case abi.BytesTy:
				fallthrough
			case abi.HashTy:
				str := string(params[i])
				if "0x" == str[0:2] {
					str = str[2:]
				}
				*paramsNew = append(*paramsNew, common.Hex2Bytes(str))
			case abi.FixedPointTy: //#zxl#
			case abi.FunctionTy: //#zxl#

			}
		}
	}
}

//func prepareResults(outs *[]interface{}, parsed *abi.ABI, method string) {
//	for i, output := range parsed.Methods[method].Outputs {
//		switch output.Type.T {
//		case abi.IntTy:
//			fallthrough
//		case abi.UintTy:
//			paramInt := new(*big.Int)
//			*outs = append(*outs, paramInt)
//
//		case abi.BoolTy:
//			paramBool := new(bool)
//			*outs = append(*outs, paramBool)
//
//		case abi.StringTy:
//			paramStr := new(string)
//			*outs = append(*outs, paramStr)
//
//		case abi.SliceTy: //#zxl#
//		case abi.ArrayTy: //#zxl#
//		case abi.AddressTy:
//			paramAddress := new(common.Address)
//			*outs = append(*outs, paramAddress)
//
//		case abi.FixedBytesTy: //#zxl
//			fallthrough
//		case abi.BytesTy:
//			inputSize := parsed.Methods[method].Inputs[i].Type.Size
//			switch inputSize {
//			case 0:
//				paramBytes := new([]uint8)
//				*outs = append(*outs, paramBytes)
//			case 32:
//				paramBytes32 := new([32]byte)
//				*outs = append(*outs, paramBytes32)
//			}
//		case abi.HashTy:
//			paramAddress := new(common.Hash)
//			*outs = append(*outs, paramAddress)
//
//		case abi.FixedPointTy: //#zxl#
//		case abi.FunctionTy: //#zxl#
//		}
//	}
//}
//
//func parseResults(outs *[]interface{}) []interface{} {
//	results := []interface{}{}
//	for _, out := range *outs {
//		switch out.(type) {
//		case **big.Int:
//			bigIntResult := **(out.(**big.Int))
//			results = append(results, bigIntResult.String())
//		case *bool:
//			boolResult := *(out.(*bool))
//			results = append(results, strconv.FormatBool(boolResult))
//		case *string:
//			strResult := *(out.(*string))
//			results = append(results, strResult)
//		case *common.Address:
//			addrResult := *(out.(*common.Address))
//			results = append(results, addrResult.String())
//		case *[]uint8:
//			bytesResult := *out.(*[]byte)
//			results = append(results, common.Bytes2Hex(bytesResult))
//		case *[32]uint8:
//			bytesResult := *out.(*[32]byte)
//			results = append(results, common.Bytes2Hex(bytesResult[:]))
//		}
//	}
//	return results
//}

func CreateContractInitialTx(input *adaptor.CreateContractInitialTxInput, rpcParams *RPCParams, netID int) (
	*adaptor.CreateContractInitialTxOutput, error) {
	//get rpc client
	client, err := GetClient(rpcParams)
	if err != nil {
		return nil, err
	}

	parsed, err := abi.JSON(strings.NewReader(string(input.Contract)))
	if err != nil {
		return nil, err
	}

	//
	value := new(big.Int)
	//value.SetString(input.Value, 10)
	gasLimitU64 := uint64(2100000)
	gasLimit := big.NewInt(2100000)
	gasPrice := input.Fee.Amount.Div(input.Fee.Amount, gasLimit)

	//
	deployerAddr := common.HexToAddress(input.Address)

	//
	var tx *types.Transaction
	if len(input.Args) != 0 {
		//
		var paramsNew []interface{}
		convertContractParams(&paramsNew, &parsed,
			"", input.Args)

		//
		_, tx, _, err = bind.DeployContractZXL(&bind.TransactOpts{From: deployerAddr, Value: value, GasPrice: gasPrice,
			GasLimit: gasLimitU64}, parsed, input.Extra, client, paramsNew...)
	} else {
		_, tx, _, err = bind.DeployContractZXL(&bind.TransactOpts{From: deployerAddr, Value: value, GasPrice: gasPrice,
			GasLimit: gasLimitU64}, parsed, input.Extra, client)
	}
	if err != nil {
		return nil, err
	}

	rlpTXBytes, err := rlp.EncodeToBytes(tx)
	if err != nil {
		return nil, err
	}

	//save result
	var result adaptor.CreateContractInitialTxOutput
	result.RawTransaction = rlpTXBytes
	//result.ContractAddr = address.String()

	return &result, nil

}

func CreateContractInvokeTx(input *adaptor.CreateContractInvokeTxInput, rpcParams *RPCParams, netID int) (
	*adaptor.CreateContractInvokeTxOutput, error) {
	//get rpc client
	client, err := GetClient(rpcParams)
	if err != nil {
		return nil, err
	}

	parsed, err := abi.JSON(strings.NewReader(string(input.Extra)))
	if err != nil {
		return nil, err
	}

	//
	var addrContract common.Address
	if "0x" == input.ContractAddress[0:2] || "0X" == input.ContractAddress[0:2] {
		addrContract = common.HexToAddress(input.ContractAddress[2:])
	} else {
		addrContract = common.HexToAddress(input.ContractAddress)
	}
	var addrCallFrom common.Address
	if "0x" == input.Address[0:2] || "0X" == input.Address[0:2] {
		addrCallFrom = common.HexToAddress(input.Address[2:])
	} else {
		addrCallFrom = common.HexToAddress(input.Address)
	}

	//
	value := new(big.Int)
	gasLimitU64 := uint64(2100000)
	gasLimit := big.NewInt(2100000)
	gasPrice := input.Fee.Amount.Div(input.Fee.Amount, gasLimit)

	//
	var tx *types.Transaction
	if len(input.Args) != 0 {
		//
		var paramsNew []interface{}
		convertContractParams(&paramsNew, &parsed,
			input.Function, input.Args)

		//
		tx, err = bind.InvokeZXL(&bind.TransactOpts{From: addrCallFrom, Value: value, GasPrice: gasPrice,
			GasLimit: gasLimitU64}, parsed, client, addrContract, input.Function, paramsNew...)
	} else {
		tx, err = bind.InvokeZXL(&bind.TransactOpts{From: addrCallFrom, Value: value, GasPrice: gasPrice,
			GasLimit: gasLimitU64}, parsed, client, addrContract, input.Function)
	}
	if err != nil {
		return nil, err
	}

	rlpTXBytes, err := rlp.EncodeToBytes(tx)
	if err != nil {
		return nil, err
	}

	//save result
	var result adaptor.CreateContractInvokeTxOutput
	result.RawTransaction = rlpTXBytes

	return &result, nil
}

func QueryContract(input *adaptor.QueryContractInput, rpcParams *RPCParams) (*adaptor.QueryContractOutput, error) {
	//get rpc client
	client, err := GetClient(rpcParams)
	if err != nil {
		return nil, err
	}

	//
	parsed, err := abi.JSON(strings.NewReader(string(input.Extra)))
	if err != nil {
		return nil, err
	}

	//
	if len(input.ContractAddress) == 0 {
		return nil, fmt.Errorf("ContractAddress is empty")
	}
	var contractAddr common.Address
	if "0x" == input.ContractAddress[0:2] || "0X" == input.ContractAddress[0:2] {
		contractAddr = common.HexToAddress(input.ContractAddress[2:])
	} else {
		contractAddr = common.HexToAddress(input.ContractAddress)
	}
	contract := bind.NewBoundContract(contractAddr, parsed, client, client, client)

	//
	var results []interface{}
	if len(input.Args) != 0 {
		//
		var paramsNew []interface{}
		convertContractParams(&paramsNew, &parsed, input.Function, input.Args)

		//
		results, err = contract.CallZXL(&bind.CallOpts{Pending: false}, input.Function, paramsNew...)
	} else {
		//
		results, err = contract.CallZXL(&bind.CallOpts{Pending: false}, input.Function)
	}
	if err != nil {
		return nil, err
	}

	//
	resultsJSON, err := json.Marshal(results)
	if err != nil {
		return nil, err
	}

	//save result
	var result adaptor.QueryContractOutput
	result.QueryResult = resultsJSON

	return &result, nil
}

func QueryContractCall(input *adaptor.QueryContractInput, rpcParams *RPCParams) (*adaptor.QueryContractOutput, error) {
	//get rpc client
	client, err := GetClient(rpcParams)
	if err != nil {
		return nil, err
	}

	//
	parsed, err := abi.JSON(strings.NewReader(string(input.Extra)))
	if err != nil {
		return nil, err
	}

	//
	if len(input.ContractAddress) == 0 {
		return nil, fmt.Errorf("ContractAddress is empty")
	}
	var contractAddr common.Address
	if "0x" == input.ContractAddress[0:2] || "0X" == input.ContractAddress[0:2] {
		contractAddr = common.HexToAddress(input.ContractAddress[2:])
	} else {
		contractAddr = common.HexToAddress(input.ContractAddress)
	}
	contract := bind.NewBoundContract(contractAddr, parsed, client, client, client)

	//
	var (
		ret0 = new(string)
	)
	if len(input.Args) != 0 {
		//
		var paramsNew []interface{}
		convertContractParams(&paramsNew, &parsed, input.Function, input.Args)

		//
		err = contract.Call(&bind.CallOpts{Pending: false}, ret0, input.Function, paramsNew...)
	} else {
		//
		err = contract.Call(&bind.CallOpts{Pending: false}, ret0, input.Function)
	}
	if err != nil {
		return nil, err
	}

	//save result
	var result adaptor.QueryContractOutput
	result.QueryResult = []byte(*ret0)

	return &result, nil
}

func UnpackInput() (string, error) {
	const PANZABI = `[
	{
		"constant": false,
		"inputs": [
			{
				"name": "_ptnhex",
				"type": "address"
			},
			{
				"name": "_amt",
				"type": "uint256"
			}
		],
		"name": "transfer",
		"outputs": [
			{
				"name": "",
				"type": "bool"
			}
		],
		"payable": false,
		"stateMutability": "nonpayable",
		"type": "function"
	}
	]`
	parsed, err := abi.JSON(strings.NewReader(PANZABI))
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	methodName := "transfer"
	method, exist := parsed.Methods[methodName]
	if !exist {
		fmt.Println("Not exist method")
		return "", fmt.Errorf("not exist method")
	}
	inputData := "000000000000000000000000c5b8f9336bf26f0f931c97d17e9376c4933ab6c8" +
		"00000000000000000000000000000000000000000000001b1ae4d6e2ef500000"
	result, err := method.Inputs.UnpackValues([]byte(inputData))
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	fmt.Println(result)
	//common.LeftPadBytes()
	return "", nil
}
