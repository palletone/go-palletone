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
	"math/big"
	"strings"

	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/contracts/shim"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"github.com/palletone/go-palletone/dag/errors"
	"github.com/palletone/go-palletone/dag/modules"
)

type PTNMain struct {
}

func (p *PTNMain) Init(stub shim.ChaincodeStubInterface) pb.Response {
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

	err = stub.PutState(symbolsContractMap, []byte(args[0]))
	if err != nil {
		return shim.Error("write symbolsContractMap failed: " + err.Error())
	}

	return shim.Success(nil)
}

func (p *PTNMain) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	f, args := stub.GetFunctionAndParameters()

	switch f {
	case "setOwner":
		return _setOwner(args, stub)
	case "setETHContract":
		return _setETHContract(args, stub)

	case "payoutPTNByETHAddr":
		return _payoutPTNByETHAddr(args, stub)
	case "payoutPTNByTxID":
		return _payoutPTNByTxID(args, stub)

	case "get":
		return get(args, stub)
	default:
		jsonResp := "{\"Error\":\"Unknown function " + f + "\"}"
		return shim.Error(jsonResp)
	}
}

//todo modify PTN ERC20 address
const PTN_ERC20Addr = "0xa54880da9a63cdd2ddacf25af68daf31a1bcc0c9"

//todo modify conforms 15
const Confirms = uint(1)

const symbolsOwner = "owner_"
const symbolsContractMap = "eth_map"

const symbolsPayout = "payout_"

func _setOwner(args []string, stub shim.ChaincodeStubInterface) pb.Response {
	//params check
	if len(args) < 1 {
		return shim.Error("need 1 args (PTNAddr)")
	}
	_, err := getOwner(stub)
	if err == nil {
		return shim.Error("Owner has been set")
	}
	err = stub.PutState(symbolsOwner, []byte(args[0]))
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

func _setETHContract(args []string, stub shim.ChaincodeStubInterface) pb.Response {
	//params check
	if len(args) < 1 {
		return shim.Error("need 1 args (MapContractAddr)")
	}

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

	err = stub.PutState(symbolsContractMap, []byte(args[0]))
	if err != nil {
		return shim.Error("write symbolsContractMap failed: " + err.Error())
	}

	return shim.Success([]byte("Success"))
}

func getMapAddr(stub shim.ChaincodeStubInterface) (string, error) {
	result, _ := stub.GetState(symbolsContractMap)
	if len(result) == 0 {
		return "", errors.New("Need set MapContractAddr")
	}

	return string(result), nil
}

func _payoutPTNByTxID(args []string, stub shim.ChaincodeStubInterface) pb.Response {
	//params check
	if len(args) < 1 {
		return shim.Error("need 1  args (transferTxID)")
	}

	//
	txIDHex := args[0]
	if "0x" == txIDHex[0:2] && "0X" == txIDHex[0:2] {
		txIDHex = txIDHex[2:]
	}
	result, _ := stub.GetState(symbolsPayout + txIDHex)
	if len(result) != 0 {
		log.Debugf("The tx has been payout")
		return shim.Error("The tx has been payout")
	}

	//get sender receiver amount
	txID, err := hex.DecodeString(txIDHex)
	if err != nil {
		log.Debugf("txid invalid: %s", err.Error())
		return shim.Error(fmt.Sprintf("txid invalid: %s", err.Error()))
	}
	txResult, err := GetErc20Tx(txID, stub)
	if err != nil {
		log.Debugf("GetErc20Tx failed: %s", err.Error())
		return shim.Error(err.Error())
	}

	//check tx status
	if !txResult.Tx.IsSuccess {
		log.Debugf("The tx is failed")
		return shim.Error("The tx is failed")
	}
	//check contract address, must be ptn erc20 contract address
	if strings.ToLower(txResult.Tx.ToAddress) != PTN_ERC20Addr {
		log.Debugf("The tx is't PTN contract")
		return shim.Error("The tx is't PTN contract")
	}
	//check receiver, must be ptnmap contract address
	mapAddr, err := getMapAddr(stub)
	if err != nil {
		log.Debugf("getMapAddr failed: %s", err.Error())
		return shim.Error(err.Error())
	}

	if strings.ToLower(txResult.Tx.ToAddress) != mapAddr {
		log.Debugf("strings.ToLower(txResult.To): %s, mapAddr: %s ", strings.ToLower(txResult.Tx.ToAddress), mapAddr)
		return shim.Error("Not send token to the Map contract")
	}
	//check token amount
	bigIntAmout := txResult.Tx.Amount.Amount.Div(&txResult.Tx.Amount.Amount, big.NewInt(1e10)) //Token's decimal is 18, PTN's decimal is 8
	amt := txResult.Tx.Amount.Amount.Uint64()
	if amt == 0 {
		log.Debugf("Amount is 0")
		return shim.Error("Amount is 0")
	}

	//check confirms
	curHeight, err := getHight(stub)
	if curHeight == 0 || err != nil {
		return shim.Error("getHeight failed")
	}
	if curHeight-txResult.Tx.BlockHeight < Confirms {
		log.Debugf("Need more confirms")
		return shim.Error("Need more confirms")
	}

	//get the mapping ptnAddr
	ptnAddr, err := getPTNMapAddr(mapAddr, txResult.Tx.FromAddress, stub)
	if err != nil {
		log.Debugf("getPTNMapAddr failed: %s", err.Error())
		return shim.Error(err.Error())
	}

	//get addrPTN
	//ptnAddr := common.HexToAddress(ptnHex).String()
	//ptnAddr := "P" + base58.CheckEncode(common.FromHex(ptnHex), 0)
	if ptnAddr == "" {
		log.Debugf("Need transfer 1 PTNMap for bind address")
		return shim.Error("Need transfer 1 PTNMap for bind address")
	}
	//save payout history
	err = stub.PutState(symbolsPayout+txIDHex, []byte(ptnAddr+"-"+bigIntAmout.String()))
	if err != nil {
		log.Debugf("write symbolsPayout failed: %s", err.Error())
		return shim.Error("write symbolsPayout failed: " + err.Error())
	}

	//payout
	asset := modules.NewPTNAsset()
	amtToken := &modules.AmountAsset{Amount: amt, Asset: asset}
	err = stub.PayOutToken(ptnAddr, amtToken, 0)
	if err != nil {
		log.Debugf("PayOutToken failed: %s", err.Error())
		return shim.Error("PayOutToken failed: " + err.Error())
	}

	return shim.Success([]byte("Success"))
}

func _payoutPTNByETHAddr(args []string, stub shim.ChaincodeStubInterface) pb.Response {
	//params check
	if len(args) < 1 {
		return shim.Error("need 1  args (ETHAddr)")
	}

	//
	ethAddr := args[0]

	//
	mapAddr, err := getMapAddr(stub)
	if err != nil {
		log.Debugf("getMapAddr failed: %s", err.Error())
		return shim.Error(err.Error())
	}

	//get the mapping ptnAddr
	ptnAddr, err := getPTNMapAddr(mapAddr, ethAddr, stub)
	if err != nil {
		log.Debugf("getPTNMapAddr failed: %s", err.Error())
		return shim.Error(err.Error())
	}

	//get addrPTN
	if ptnAddr == "" {
		log.Debugf("Need transfer 1 PTNMap for bind address")
		return shim.Error("Need transfer 1 PTNMap for bind address")
	}

	txResults, err := GetAddrHistory(ethAddr, mapAddr, stub)
	if err != nil {
		log.Debugf("GetAddrHistory failed: %s", err.Error())
		return shim.Error(err.Error())
	}

	curHeight, err := getHight(stub)
	if curHeight == 0 || err != nil {
		return shim.Error("getHeight failed")
	}

	var amt uint64
	for _, txResult := range txResults.Txs {
		txIDHex := hex.EncodeToString(txResult.TxID)
		//check confirms
		if curHeight-txResult.BlockHeight < Confirms {
			log.Debugf("Need more confirms %s", txIDHex)
			continue
		}
		//
		result, _ := stub.GetState(symbolsPayout + txIDHex)
		if len(result) != 0 {
			log.Debugf("The tx has been payout")
			continue
		}

		//check token amount
		bigIntAmout := txResult.Amount.Amount.Div(&txResult.Amount.Amount, big.NewInt(1e10)) //Token's decimal is 18, PTN's decimal is 8
		amt += txResult.Amount.Amount.Uint64()

		//save payout history
		err = stub.PutState(symbolsPayout+txIDHex, []byte(ptnAddr+"-"+bigIntAmout.String()))
		if err != nil {
			log.Debugf("write symbolsPayout failed: %s", err.Error())
			return shim.Error("write symbolsPayout failed: " + err.Error())
		}

	}

	if amt == 0 {
		log.Debugf("Amount is 0")
		return shim.Error("Amount is 0")
	}

	//payout
	asset := modules.NewPTNAsset()
	amtToken := &modules.AmountAsset{Amount: amt, Asset: asset}
	err = stub.PayOutToken(ptnAddr, amtToken, 0)
	if err != nil {
		log.Debugf("PayOutToken failed: %s", err.Error())
		return shim.Error("PayOutToken failed: " + err.Error())
	}

	return shim.Success([]byte("Success"))
}

//refer to the struct GetTransferTxInput in "github.com/palletone/adaptor/ICryptoCurrency.go",
type GetTransferTxInput struct {
	TxID  []byte `json:"tx_id"`
	Extra []byte `json:"extra"`
}

//TxBasicInfo 一个交易的基本信息
type TxBasicInfo struct {
	TxID           []byte `json:"tx_id"`           //交易的ID，Hash
	TxRawData      []byte `json:"tx_raw"`          //交易的二进制数据
	CreatorAddress string `json:"creator_address"` //交易的发起人
	TargetAddress  string `json:"target_address"`  //交易的目标地址（被调用的合约、收款人）
	IsInBlock      bool   `json:"is_in_block"`     //是否已经被打包到区块链中
	IsSuccess      bool   `json:"is_success"`      //是被标记为成功执行
	IsStable       bool   `json:"is_stable"`       //是否已经稳定不可逆
	BlockID        []byte `json:"block_id"`        //交易被打包到了哪个区块ID
	BlockHeight    uint   `json:"block_height"`    //交易被打包到的区块的高度
	TxIndex        uint   `json:"tx_index"`        //Tx在区块中的位置
	Timestamp      uint64 `json:"timestamp"`       //交易被打包的时间戳
}

//AmountAsset Token的金额和资产标识
type AmountAsset struct {
	Amount big.Int `json:"amount"` //金额，最小单位
	Asset  string  `json:"asset"`  //资产标识
}

//SimpleTransferTokenTx 一个简单的Token转账交易
type SimpleTransferTokenTx struct {
	TxBasicInfo
	FromAddress string       `json:"from_address"` //转出地址
	ToAddress   string       `json:"to_address"`   //转入地址
	Amount      *AmountAsset `json:"amount"`       //转账金额
	Fee         *AmountAsset `json:"fee"`          //转账交易费
	AttachData  []byte       `json:"attach_data"`  //附加的数据（备注之类的）
}
type GetTransferTxOutput struct {
	Tx    SimpleTransferTokenTx `json:"transaction"`
	Extra []byte                `json:"extra"`
}

func GetErc20Tx(txID []byte, stub shim.ChaincodeStubInterface) (*GetTransferTxOutput, error) {
	input := GetTransferTxInput{TxID: txID}
	inputBytes, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	//
	result, err := stub.OutChainCall("erc20", "GetTransferTx", inputBytes)
	if err != nil {
		return nil, errors.New("GetTransferTx error")
	}
	log.Debugf("result : %s", string(result))

	//
	var output GetTransferTxOutput
	err = json.Unmarshal(result, &output)
	if err != nil {
		return nil, err
	}
	return &output, nil
}

type GetAddrTxHistoryInput struct {
	FromAddress       string `json:"from_address"`         //转账的付款方地址
	ToAddress         string `json:"to_address"`           //转账的收款方地址
	Asset             string `json:"asset"`                //资产标识
	PageSize          uint32 `json:"page_size"`            //分页大小，0表示不分页
	PageIndex         uint32 `json:"page_index"`           //分页后的第几页数据
	AddressLogicAndOr bool   `json:"address_logic_and_or"` //付款地址,收款地址是And=1关系还是Or=0关系
	Asc               bool   `json:"asc"`                  //按时间顺序从老到新
	Extra             []byte `json:"extra"`
}

type GetAddrTxHistoryOutput struct {
	Txs   []*SimpleTransferTokenTx `json:"transactions"` //返回的交易列表
	Count uint32                   `json:"count"`        //忽略分页，有多少条记录
	Extra []byte                   `json:"extra"`
}

func GetAddrHistory(ethAddrFrom, mapAddrTo string, stub shim.ChaincodeStubInterface) (*GetAddrTxHistoryOutput, error) {
	input := GetAddrTxHistoryInput{FromAddress: ethAddrFrom, ToAddress: mapAddrTo, Asset: PTN_ERC20Addr,
		AddressLogicAndOr: true}
	inputBytes, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	//
	result, err := stub.OutChainCall("erc20", "GetAddrTxHistory", inputBytes)
	if err != nil {
		return nil, errors.New("GetTransferTx error")
	}
	//
	var output GetAddrTxHistoryOutput
	err = json.Unmarshal(result, &output)
	if err != nil {
		return nil, err
	}
	return &output, nil
}

//refer to the struct GetBlockInfoInput in "github.com/palletone/adaptor/IUtility.go",
type GetBlockInfoInput struct {
	Latest  bool   `json:"latest"`   //true表示查询最新区块
	Height  uint64 `json:"height"`   //根据高度查询区块
	BlockID []byte `json:"block_id"` //根据Hash查询区块
	Extra   []byte `json:"extra"`
}

type BlockInfo struct {
	BlockID         []byte `json:"block_id"`         //交易被打包到了哪个区块ID
	BlockHeight     uint   `json:"block_height"`     //交易被打包到的区块的高度
	Timestamp       uint64 `json:"timestamp"`        //交易被打包的时间戳
	ParentBlockID   []byte `json:"parent_block_id"`  //父区块ID
	HeaderRawData   []byte `json:"header_raw_data"`  //区块头的原始信息
	TxsRoot         []byte `json:"txs_root"`         //默克尔根
	ProducerAddress string `json:"producer_address"` //生产者地址
	IsStable        bool   `json:"is_stable"`        //是否已经稳定不可逆
}
type GetBlockInfoOutput struct {
	Block BlockInfo `json:"block"`
	Extra []byte    `json:"extra"`
}

func getHight(stub shim.ChaincodeStubInterface) (uint, error) {
	//
	input := GetBlockInfoInput{Latest: true} //get best hight
	//
	inputBytes, err := json.Marshal(input)
	if err != nil {
		return 0, err
	}
	//
	result, err := stub.OutChainCall("erc20", "GetBlockInfo", inputBytes)
	if err != nil {
		return 0, err
	}
	//
	var output GetBlockInfoOutput
	err = json.Unmarshal(result, &output)
	if err != nil {
		return 0, err
	}

	if output.Block.BlockHeight == 0 {
		return 0, errors.New("{\"Error\":\"Failed to get eth height\"}")
	}

	return output.Block.BlockHeight, nil
}

//refer to the struct GetPalletOneMappingAddressInput in "github.com/palletone/adaptor/IUtility.go",
type GetPalletOneMappingAddressInput struct {
	PalletOneAddress  string `json:"palletone_address"`
	ChainAddress      string `json:"chain_address"`
	MappingDataSource string `json:"mapping_data_source"` //映射地址数据查询的地方，以太坊就是一个合约地址
	Extra             []byte `json:"extra"`
}
type GetPalletOneMappingAddressOutput struct {
	PalletOneAddress string `json:"palletone_address"`
	ChainAddress     string `json:"chain_address"`
	Extra            []byte `json:"extra"`
}

func getPTNMapAddr(mapAddr, fromAddr string, stub shim.ChaincodeStubInterface) (string, error) {
	var input GetPalletOneMappingAddressInput
	input.MappingDataSource = mapAddr
	input.ChainAddress = fromAddr

	inputBytes, err := json.Marshal(input)
	if err != nil {
		return "", err
	}
	//
	result, err := stub.OutChainCall("erc20", "GetPalletOneMappingAddress", inputBytes)
	if err != nil {
		return "", errors.New("GetPalletOneMappingAddress failed")
	}
	//
	var output GetPalletOneMappingAddressOutput
	err = json.Unmarshal(result, &output)
	if err != nil {
		return "", err
	}
	if output.PalletOneAddress == "" {
		return "", errors.New("GetPalletOneMappingAddress result empty")
	}

	return output.PalletOneAddress, nil
}

func get(args []string, stub shim.ChaincodeStubInterface) pb.Response {
	if len(args) > 0 {
		result, _ := stub.GetState(args[0])
		return shim.Success(result) //test
	}
	result, _ := stub.GetState("result")
	return shim.Success(result)
}

func main() {
	err := shim.Start(new(PTNMain))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
