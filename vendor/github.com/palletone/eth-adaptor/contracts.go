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
package adaptoreth

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	//	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	//"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/eth-adaptor/bind"

	"github.com/palletone/adaptor"
)

func CreateMultiSigAddress(createMultiSigAddressParams *adaptor.CreateMultiSigAddressParams) (string, error) {
	//	//redeem format : A B 1 2 3 4 utc
	//	utc := time.Now().UTC().Unix()
	//redeem format : A B 1 2 3 4
	var redeem string
	for _, address := range createMultiSigAddressParams.Addresses {
		if "0x" == address[0:2] {
			address = address[2:]
		}
		redeem = fmt.Sprintf("%s%s", redeem, address)
	}
	//	redeem = fmt.Sprintf("%s%x", redeem, utc)

	//save result
	var result adaptor.CreateMultiSigAddressResult
	result.RedeemHex = redeem

	//
	jsonResult, err := json.Marshal(result)
	if err != nil {
		return "", err
	}

	return string(jsonResult), nil
}

type CalculateSigParams struct {
	PrivateKeyHex      string `json:"privatekeyhex"`
	PalletContractAddr string `json:"palletcontractaddr"`
	TokenAddr          string `json:"tokenaddr"`
	RedeemHex          string `json:"redeemhex"`
	RecverAddr         string `json:"recveraddr"`
	Amount             string `json:"amount"`
	Nonece             string `json:"nonece"`
}
type CalculateSigResult struct {
	Signature string `json:"signature"`
}

func CalculateSig(params string) string {
	//convert params from json format
	var calculateSigParams CalculateSigParams
	err := json.Unmarshal([]byte(params), &calculateSigParams)
	if err != nil {
		return err.Error()
	}

	//remove 0x, then convert to ecdsa private key
	if "0x" == calculateSigParams.PrivateKeyHex[0:2] {
		calculateSigParams.PrivateKeyHex = calculateSigParams.PrivateKeyHex[2:]
	}
	privateKey, err := crypto.HexToECDSA(calculateSigParams.PrivateKeyHex)
	if err != nil {
		return err.Error()
	}

	//redeem bytes
	redeem, err := hexutil.Decode(calculateSigParams.RedeemHex)
	if err != nil {
		return err.Error()
	}

	//address
	recvAddr := common.HexToAddress(calculateSigParams.RecverAddr)

	//contract address
	contractAddr := common.HexToAddress(calculateSigParams.PalletContractAddr)

	//amount
	amountBigInt := new(big.Int)
	amountBigInt.SetString(calculateSigParams.Amount, 10)
	paddedAmount := common.LeftPadBytes(amountBigInt.Bytes(), 32)

	//nonece
	nonece := new(big.Int)
	nonece.SetString(calculateSigParams.Nonece, 10)
	paddedNonece := common.LeftPadBytes(nonece.Bytes(), 32)

	//0x0 or empty is eth multisig, otherwise ERC20
	var hash common.Hash
	if calculateSigParams.TokenAddr == "" || calculateSigParams.TokenAddr == "0x0" {
		hash = crypto.Keccak256Hash(redeem, recvAddr.Bytes(),
			contractAddr.Bytes(), paddedAmount, paddedNonece)
	} else {
		//
		tokenAddr := common.HexToAddress(calculateSigParams.TokenAddr)
		hash = crypto.Keccak256Hash(tokenAddr.Bytes(), redeem, recvAddr.Bytes(),
			contractAddr.Bytes(), paddedAmount, paddedNonece)
	}
	//	fmt.Println("hash : ", hash.Hex())

	//sign the hash
	signature, err := crypto.Sign(hash.Bytes(), privateKey)
	if err != nil {
		return err.Error()
	}
	//	fmt.Println(hexutil.Encode(signature))

	//save result
	var result CalculateSigResult
	result.Signature = hexutil.Encode(signature)

	//
	jsonResult, err := json.Marshal(result)
	if err != nil {
		return err.Error()
	}

	return string(jsonResult)
}

//==== ==== ==== ==== ==== ==== ==== ==== ==== ==== ==== ==== ==== ==== ==== ===

func Keccak256HashPackedSig(sigParams *adaptor.Keccak256HashPackedSigParams) (string, error) {
	//remove 0x, then convert to ecdsa private key
	if "0x" == sigParams.PrivateKeyHex[0:2] {
		sigParams.PrivateKeyHex = sigParams.PrivateKeyHex[2:]
	}
	privateKey, err := crypto.HexToECDSA(sigParams.PrivateKeyHex)
	if err != nil {
		return "", err
	}

	var paramTypes []string
	err = json.Unmarshal([]byte(sigParams.ParamTypes), &paramTypes)
	if err != nil {
		return "", err
	}

	var params []string
	err = json.Unmarshal([]byte(sigParams.Params), &params)
	if err != nil {
		return "", err
	}

	var allBytes [][]byte
	if len(paramTypes) == len(params) {
		for i, arg := range params {
			//fmt.Println("=== ==== ==== ==== ", arg)
			switch paramTypes[i] {
			case "Int":
				fallthrough
			case "Uint":
				paramBigInt := new(big.Int)
				paramBigInt.SetString(arg, 10)
				paddedBytes := common.LeftPadBytes(paramBigInt.Bytes(), 32)
				allBytes = append(allBytes, paddedBytes)

			case "Bool":
				paramBool, _ := strconv.ParseBool(arg)
				if paramBool {
					paddedBytes := common.LeftPadBytes(common.Big1.Bytes(), 32)
					allBytes = append(allBytes, paddedBytes)
				} else {
					paddedBytes := common.LeftPadBytes(common.Big0.Bytes(), 32)
					allBytes = append(allBytes, paddedBytes)
				}

			case "String":
				paddedBytes := common.LeftPadBytes([]byte(arg), 32)
				allBytes = append(allBytes, paddedBytes)

			case "Slice": //#zxl#
			case "Array": //#zxl#
			case "Address":
				addr := common.HexToAddress(arg)
				allBytes = append(allBytes, addr.Bytes())
				//fmt.Println(addr.Bytes())

			case "FixedBytes":
				fallthrough
			case "Hash":
				fallthrough
			case "Bytes":
				if "0x" == arg[0:2] {
					arg = arg[2:]
				}
				paramBytes := common.Hex2Bytes(arg)
				allBytes = append(allBytes, paramBytes)
				//fmt.Println(paramBytes)

			case "FixedPoint": //#zxl#
			case "Function": //#zxl#
			}
		}
	} else {
		return "", errors.New("Params error : ParamTypes and params not match.")
	}

	if len(allBytes) != len(params) {
		return "", errors.New("Params error : Process params error.")
	}

	//
	hash := crypto.Keccak256Hash(allBytes...)

	//sign the hash
	signature, err := crypto.Sign(hash.Bytes(), privateKey)
	if err != nil {
		return "", err
	}

	//save result
	var result adaptor.Keccak256HashPackedSigResult
	result.Hash = hash.String()
	result.Signature = hexutil.Encode(signature)

	//
	jsonResult, err := json.Marshal(result)
	if err != nil {
		return "", err
	}

	return string(jsonResult), nil
}

func Keccak256HashVerify(verifyParams *adaptor.Keccak256HashVerifyParams) (string, error) {
	//
	if "0x" == verifyParams.PublicKeyHex[0:2] {
		verifyParams.PublicKeyHex = verifyParams.PublicKeyHex[2:]
	}
	pubkey := common.Hex2Bytes(verifyParams.PublicKeyHex)
	if "0x" == verifyParams.Hash[0:2] {
		verifyParams.Hash = verifyParams.Hash[2:]
	}
	hash := common.Hex2Bytes(verifyParams.Hash)
	if "0x" == verifyParams.Signature[0:2] {
		verifyParams.Signature = verifyParams.Signature[2:]
	}
	sig := common.Hex2Bytes(verifyParams.Signature)
	if len(sig) == 65 {
		sig = sig[0:64]
	}

	//sign the hash
	valid := crypto.VerifySignature(pubkey, hash, sig)

	//save result
	var result adaptor.Keccak256HashVerifyResult
	result.Valid = valid

	//
	jsonResult, err := json.Marshal(result)
	if err != nil {
		return "", err
	}

	return string(jsonResult), nil
}

func RecoverAddr(recoverParams *adaptor.RecoverParams) (string, error) {
	//
	if "0x" == recoverParams.Hash[0:2] {
		recoverParams.Hash = recoverParams.Hash[2:]
	}
	hash := common.Hex2Bytes(recoverParams.Hash)
	if "0x" == recoverParams.Signature[0:2] {
		recoverParams.Signature = recoverParams.Signature[2:]
	}
	sig := common.Hex2Bytes(recoverParams.Signature)

	//
	pubkeyByte, err := crypto.Ecrecover(hash, sig)
	if err != nil {
		return "", err
	}

	pubkey, err := crypto.UnmarshalPubkey(pubkeyByte)
	if err != nil {
		return "", err
	}
	addr := crypto.PubkeyToAddress(*pubkey)

	//save result
	var result adaptor.RecoverResult
	result.Addr = addr.String()

	//
	jsonResult, err := json.Marshal(result)
	if err != nil {
		return "", err
	}

	return string(jsonResult), nil
}

func convertContractParams(paramsNew *[]interface{}, parsed *abi.ABI,
	method string, paramsJson string) error {

	var params []string
	err := json.Unmarshal([]byte(paramsJson), &params)
	if err != nil {
		return err
	}

	var theMethod abi.Method
	if "" == method {
		theMethod = parsed.Constructor
	} else {
		theMethod = parsed.Methods[method]
	}

	if len(params) == len(theMethod.Inputs) {
		for i, arg := range params {
			switch theMethod.Inputs[i].Type.T {
			case abi.IntTy:
				fallthrough
			case abi.UintTy:
				paramInt := new(big.Int)
				paramInt.SetString(arg, 10)
				*paramsNew = append(*paramsNew, paramInt)
			case abi.BoolTy:
				paramBool, _ := strconv.ParseBool(arg)
				*paramsNew = append(*paramsNew, paramBool)
			case abi.StringTy:
				*paramsNew = append(*paramsNew, arg)

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
				paramBytes := common.HexToAddress(arg)
				*paramsNew = append(*paramsNew, paramBytes)
			case abi.FixedBytesTy:
				if "0x" == arg[0:2] {
					arg = arg[2:]
				}
				paramBytes := common.Hex2Bytes(arg)
				inputSize := parsed.Methods[method].Inputs[i].Type.Size
				if len(paramBytes) == inputSize {
					switch inputSize {
					case 32:
						//byte32 := new([32]byte)
						var byte32 [32]byte
						for i := 0; i < len(paramBytes); i++ {
							byte32[i] = paramBytes[i]
						}
						*paramsNew = append(*paramsNew, byte32)
					}
				}
			case abi.BytesTy:
				fallthrough
			case abi.HashTy:
				if "0x" == arg[0:2] {
					arg = arg[2:]
				}
				paramBytes := common.Hex2Bytes(arg)
				*paramsNew = append(*paramsNew, paramBytes[:])
			case abi.FixedPointTy: //#zxl#
			case abi.FunctionTy: //#zxl#

			}
		}
	}
	return nil
}

func prepareResults(outs *[]interface{}, parsed *abi.ABI, method string) {
	for i, output := range parsed.Methods[method].Outputs {
		switch output.Type.T {
		case abi.IntTy:
			fallthrough
		case abi.UintTy:
			paramInt := new(*big.Int)
			*outs = append(*outs, paramInt)

		case abi.BoolTy:
			paramBool := new(bool)
			*outs = append(*outs, paramBool)

		case abi.StringTy:
			paramStr := new(string)
			*outs = append(*outs, paramStr)

		case abi.SliceTy: //#zxl#
		case abi.ArrayTy: //#zxl#
		case abi.AddressTy:
			paramAddress := new(common.Address)
			*outs = append(*outs, paramAddress)

		case abi.FixedBytesTy: //#zxl
			fallthrough
		case abi.BytesTy:
			inputSize := parsed.Methods[method].Inputs[i].Type.Size
			switch inputSize {
			case 0:
				paramBytes := new([]uint8)
				*outs = append(*outs, paramBytes)
			case 32:
				paramBytes32 := new([32]byte)
				*outs = append(*outs, paramBytes32)
			}
		case abi.HashTy:
			paramAddress := new(common.Hash)
			*outs = append(*outs, paramAddress)

		case abi.FixedPointTy: //#zxl#
		case abi.FunctionTy: //#zxl#
		}
	}
}

func parseResults(outs *[]interface{}) []interface{} {
	results := []interface{}{}
	for _, out := range *outs {
		switch out.(type) {
		case **big.Int:
			bigIntResult := **(out.(**big.Int))
			results = append(results, bigIntResult.String())
		case *bool:
			boolResult := *(out.(*bool))
			results = append(results, strconv.FormatBool(boolResult))
		case *string:
			strResult := *(out.(*string))
			results = append(results, strResult)
		case *common.Address:
			addrResult := *(out.(*common.Address))
			results = append(results, addrResult.String())
		case *[]uint8:
			bytesResult := *out.(*[]byte)
			results = append(results, common.Bytes2Hex(bytesResult[:]))
		case *[32]uint8:
			bytesResult := *out.(*[32]byte)
			results = append(results, common.Bytes2Hex(bytesResult[:]))
		}
	}
	return results
}

func QueryContract(queryContractParams *adaptor.QueryContractParams, rpcParams *RPCParams, netID int) (string, error) {
	//get rpc client
	client, err := GetClient(rpcParams)
	if err != nil {
		return "", err
	}

	//
	parsed, err := abi.JSON(strings.NewReader(queryContractParams.ContractABI))
	if err != nil {
		return "", err
	}

	//
	contractAddr := common.HexToAddress(queryContractParams.ContractAddr)
	contract := bind.NewBoundContract(contractAddr, parsed, client, client, client)

	//
	var results []interface{}
	if queryContractParams.Params != "" {
		//
		var paramsNew []interface{}
		err = convertContractParams(&paramsNew, &parsed,
			queryContractParams.Method, queryContractParams.Params)
		if err != nil {
			return "", err
		}
		//
		results, err = contract.CallZXL(&bind.CallOpts{Pending: false},
			queryContractParams.Method, paramsNew...)
	} else {
		//
		results, err = contract.CallZXL(&bind.CallOpts{Pending: false},
			queryContractParams.Method, queryContractParams.ParamsArray...)
	}
	if err != nil {
		return "", err
	}

	//
	resultsJson, err := json.Marshal(results)
	if err != nil {
		return "", err
	}

	//save result
	var result adaptor.QueryContractResult
	result.Result = string(resultsJson)

	//
	jsonResult, err := json.Marshal(result)
	if err != nil {
		return "", err
	}

	return string(jsonResult), nil
}

func GenInvokeContractTX(invokeContractParams *adaptor.GenInvokeContractTXParams, rpcParams *RPCParams, netID int) (string, error) {
	//get rpc client
	client, err := GetClient(rpcParams)
	if err != nil {
		return "", err
	}

	parsed, err := abi.JSON(strings.NewReader(invokeContractParams.ContractABI))
	if err != nil {
		return "", err
	}

	//
	addrContract := common.HexToAddress(invokeContractParams.ContractAddr)
	addrCallFrom := common.HexToAddress(invokeContractParams.CallerAddr)

	//
	value := new(big.Int)
	value.SetString(invokeContractParams.Value, 10)
	gasPrice := new(big.Int)
	gasPrice.SetString(invokeContractParams.GasPrice, 10)
	gasLimit, err := strconv.ParseUint(invokeContractParams.GasLimit, 10, 64)
	if err != nil {
		gasLimit = 2100000
	}

	//
	var tx *types.Transaction
	if invokeContractParams.Params != "" {
		//
		var paramsNew []interface{}
		convertContractParams(&paramsNew, &parsed,
			invokeContractParams.Method, invokeContractParams.Params)

		//
		tx, err = bind.InvokeZXL(&bind.TransactOpts{From: addrCallFrom, Value: value, GasPrice: gasPrice, GasLimit: gasLimit},
			parsed, client, addrContract,
			invokeContractParams.Method, paramsNew...)
	} else {
		tx, err = bind.InvokeZXL(&bind.TransactOpts{From: addrCallFrom, Value: value, GasPrice: gasPrice, GasLimit: gasLimit},
			parsed, client, addrContract,
			invokeContractParams.Method, invokeContractParams.ParamsArray...)
	}
	if err != nil {
		return "", err
	}

	rlpTXBytes, err := rlp.EncodeToBytes(tx)
	if err != nil {
		return "", err
	}

	//save result
	var result adaptor.GenInvokeContractTXResult
	result.TransactionHex = hexutil.Encode(rlpTXBytes)

	//
	jsonResult, err := json.Marshal(result)
	if err != nil {
		return "", err
	}

	return string(jsonResult), nil
}

func GenDeployContractTX(deployContractParams *adaptor.GenDeployContractTXParams, rpcParams *RPCParams, netID int) (string, error) {
	//get rpc client
	client, err := GetClient(rpcParams)
	if err != nil {
		return "", err
	}

	parsed, err := abi.JSON(strings.NewReader(deployContractParams.ContractABI))
	if err != nil {
		return "", err
	}

	//
	value := new(big.Int)
	value.SetString(deployContractParams.Value, 10)
	gasPrice := new(big.Int)
	gasPrice.SetString(deployContractParams.GasPrice, 10)
	gasLimit, err := strconv.ParseUint(deployContractParams.GasLimit, 10, 64)
	if err != nil {
		gasLimit = 2100000
	}

	//
	deployerAddr := common.HexToAddress(deployContractParams.DeployerAddr)

	//
	var address common.Address
	var tx *types.Transaction
	if deployContractParams.Params != "" {
		//
		var paramsNew []interface{}
		convertContractParams(&paramsNew, &parsed,
			"", deployContractParams.Params)

		//
		address, tx, _, err = bind.DeployContractZXL(&bind.TransactOpts{From: deployerAddr, Value: value, GasPrice: gasPrice, GasLimit: gasLimit}, parsed,
			common.FromHex(deployContractParams.ContractBin), client, paramsNew...)
	} else {
		address, tx, _, err = bind.DeployContractZXL(&bind.TransactOpts{From: deployerAddr, Value: value, GasPrice: gasPrice, GasLimit: gasLimit}, parsed,
			common.FromHex(deployContractParams.ContractBin), client, deployContractParams.ParamsArray...)
	}
	if err != nil {
		return "", err
	}

	rlpTXBytes, err := rlp.EncodeToBytes(tx)
	if err != nil {
		return "", err
	}

	//save result
	var result adaptor.GenDeployContractTXResult
	result.TransactionHex = hexutil.Encode(rlpTXBytes)
	result.ContractAddr = address.String()

	//
	jsonResult, err := json.Marshal(result)
	if err != nil {
		return "", err
	}

	return string(jsonResult), nil

}
