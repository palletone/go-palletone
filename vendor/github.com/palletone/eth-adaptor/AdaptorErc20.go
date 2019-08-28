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
 * @date 2018-2019
 */
package ethadaptor

import (
	"errors"
	"fmt"
	"math/big"
	"strconv"

	"github.com/btcsuite/btcutil/base58"

	"github.com/palletone/adaptor"
)

type AdaptorErc20 struct {
	NetID int
	RPCParams
	//lockContractAddress string
}

func NewAdaptorErc20(netID int, rPCParams RPCParams) *AdaptorErc20 {
	return &AdaptorErc20{netID, rPCParams}
}
func NewAdaptorErc20Testnet() *AdaptorErc20 {
	return &AdaptorErc20{
		NetID: NETID_TEST,
		RPCParams: RPCParams{Rawurl: "https://ropsten.infura.io",
			TxQueryUrl: "https://api-ropsten.etherscan.io/api"},
		//lockContractAddress: "0x4d736ed88459b2db85472aab13a9d0ce2a6ea676",
	}
}
func NewAdaptorErc20Mainnet() *AdaptorErc20 {
	return &AdaptorErc20{
		NetID: NETID_MAIN,
		RPCParams: RPCParams{Rawurl: "https://mainnet.infura.io",
			TxQueryUrl: "https://api.etherscan.io/api"},
		//lockContractAddress: "0x1989a21eb0f28063e47e6b448e8d76774bc9b493",
	}
}

/*IUtility*/
//创建一个新的私钥
func (aerc20 *AdaptorErc20) NewPrivateKey(input *adaptor.NewPrivateKeyInput) (*adaptor.NewPrivateKeyOutput, error) {
	prikey, err := NewPrivateKey(aerc20.NetID)
	if err != nil {
		return nil, err
	}
	result := adaptor.NewPrivateKeyOutput{PrivateKey: prikey}
	return &result, nil
}

//根据私钥创建公钥
func (aerc20 *AdaptorErc20) GetPublicKey(input *adaptor.GetPublicKeyInput) (*adaptor.GetPublicKeyOutput, error) {
	pubkey, err := GetPublicKey(input.PrivateKey)
	if err != nil {
		return nil, err
	}
	result := adaptor.GetPublicKeyOutput{PublicKey: pubkey}
	return &result, nil
}

//根据Key创建地址
func (aerc20 *AdaptorErc20) GetAddress(key *adaptor.GetAddressInput) (*adaptor.GetAddressOutput, error) {
	addr, err := PubKeyToAddress(key.Key)
	if err != nil {
		return nil, err
	}
	result := adaptor.GetAddressOutput{Address: addr}
	return &result, nil
}

func GetMappAddr(addr *adaptor.GetPalletOneMappingAddressInput,
	rpcParams *RPCParams, queryContractAddr string) (
	*adaptor.GetPalletOneMappingAddressOutput, error) {
	const MapAddrABI = `[
	{
		"constant": true,
		"inputs": [
			{
				"name": "ptnAddr",
				"type": "address"
			}
		],
		"name": "getMapEthAddr",
		"outputs": [
			{
				"name": "",
				"type": "address"
			}
		],
		"payable": false,
		"stateMutability": "view",
		"type": "function"
	},
	{
		"constant": true,
		"inputs": [
			{
				"name": "addr",
				"type": "address"
			}
		],
		"name": "getMapPtnAddr",
		"outputs": [
			{
				"name": "",
				"type": "string"
			}
		],
		"payable": false,
		"stateMutability": "view",
		"type": "function"
	}
]`

	var input adaptor.QueryContractInput
	input.ContractAddress = queryContractAddr
	if len(addr.ChainAddress) != 0 { //ETH地址
		input.Function = "getMapPtnAddr"
		input.Args = append(input.Args, []byte(addr.ChainAddress))
	} else { //PTN地址 P开头
		input.Function = "getMapEthAddr"
		if addr.PalletOneAddress[0] != byte('P') {
			return nil, fmt.Errorf("PalletOne address must start with 'P'")
		}
		addrBytes, _, err := base58.CheckDecode(addr.PalletOneAddress[1:])
		if err != nil {
			return nil, err
		}
		addrHex := fmt.Sprintf("%x", addrBytes)
		input.Args = append(input.Args, []byte(addrHex))
	}
	input.Extra = []byte(MapAddrABI)

	//
	resultQuery, err := QueryContract(&input, rpcParams)
	if err != nil {
		return nil, err
	}
	resultStr := string(resultQuery.QueryResult)
	fmt.Println("address map:", resultStr)
	if len(resultStr) == 0 || resultStr == "0x0000000000000000000000000000000000000000" {
		return nil, adaptor.ErrNotFound
	}

	var result adaptor.GetPalletOneMappingAddressOutput
	result.PalletOneAddress = resultStr[2 : len(resultStr)-2]

	return &result, nil
}
func (aerc20 *AdaptorErc20) GetPalletOneMappingAddress(addr *adaptor.GetPalletOneMappingAddressInput) (
	*adaptor.GetPalletOneMappingAddressOutput, error) {
	if len(addr.MappingDataSource) == 0 {
		return nil, errors.New("you must define mapping contract address in MappingDataSource")
	}
	return GetMappAddr(addr, &aerc20.RPCParams, addr.MappingDataSource)
}

//对一条交易进行签名，并返回签名结果
func (aerc20 *AdaptorErc20) SignTransaction(input *adaptor.SignTransactionInput) (
	*adaptor.SignTransactionOutput, error) {
	if 'm' == input.Transaction[0] {
		inputNew := &adaptor.SignMessageInput{}
		inputNew.PrivateKey = input.PrivateKey
		inputNew.Message = input.Transaction[1:]
		result, err := SignMessage(inputNew)
		if err != nil {
			return nil, err
		}
		return &adaptor.SignTransactionOutput{Signature: result.Signature}, nil
	}
	return SignTransaction(input)
}

//对一条消息进行签名
func (aerc20 *AdaptorErc20) SignMessage(input *adaptor.SignMessageInput) (*adaptor.SignMessageOutput, error) {
	return SignMessage(input)
}

//对签名进行验证
func (aerc20 *AdaptorErc20) VerifySignature(input *adaptor.VerifySignatureInput) (
	*adaptor.VerifySignatureOutput, error) {
	return VerifySignature(input)
}

//将未签名的原始交易与签名进行绑定，返回一个签名后的交易
func (aerc20 *AdaptorErc20) BindTxAndSignature(input *adaptor.BindTxAndSignatureInput) (
	*adaptor.BindTxAndSignatureOutput, error) {
	return BindTxAndSignature(input)
}

//根据交易内容，计算交易Hash
func (aerc20 *AdaptorErc20) CalcTxHash(input *adaptor.CalcTxHashInput) (*adaptor.CalcTxHashOutput, error) {
	return CalcTxHash(input)
}

//将签名后的交易广播到网络中,如果发送交易需要手续费，指定最多支付的手续费
func (aerc20 *AdaptorErc20) SendTransaction(input *adaptor.SendTransactionInput) (
	*adaptor.SendTransactionOutput, error) {
	return SendTransaction(input, &aerc20.RPCParams)
}

//根据交易ID获得交易的基本信息
func (aerc20 *AdaptorErc20) GetTxBasicInfo(input *adaptor.GetTxBasicInfoInput) (
	*adaptor.GetTxBasicInfoOutput, error) {
	return GetTxBasicInfo(input, &aerc20.RPCParams, aerc20.NetID)
}

/*ICryptoCurrency*/
//获取某地址下持有某资产的数量,返回数量为该资产的最小单位
func (aerc20 *AdaptorErc20) GetBalance(input *adaptor.GetBalanceInput) (*adaptor.GetBalanceOutput, error) {
	const ERC20ABI = "[{\"constant\":true,\"inputs\":[{\"name\":\"who\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"}]"

	var inputQuery adaptor.QueryContractInput
	inputQuery.ContractAddress = input.Asset
	inputQuery.Function = "balanceOf"
	inputQuery.Args = append(inputQuery.Args, []byte(input.Address))
	inputQuery.Extra = []byte(ERC20ABI)

	//
	resultQuery, err := QueryContract(&inputQuery, &aerc20.RPCParams)
	if err != nil {
		return nil, err
	}
	balanceStr := string(resultQuery.QueryResult)
	balanceAmt := new(big.Int)
	balanceAmt.SetString(balanceStr[1:len(balanceStr)-1], 10)
	var result = &adaptor.GetBalanceOutput{}
	result.Balance = adaptor.AmountAsset{Amount: balanceAmt, Asset: input.Asset}
	return result, nil
}

//获取某资产的小数点位数
func (aerc20 *AdaptorErc20) GetAssetDecimal(asset *adaptor.GetAssetDecimalInput) (
	*adaptor.GetAssetDecimalOutput, error) {
	const ERC20ABI = "[{\"constant\":true,\"inputs\":[],\"name\":\"decimals\",\"outputs\":[{\"name\":\"\",\"type\":\"uint8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"}]"

	var input adaptor.QueryContractInput
	input.ContractAddress = asset.Asset
	input.Function = "decimals"
	input.Extra = []byte(ERC20ABI)

	//
	resultQuery, err := QueryContract(&input, &aerc20.RPCParams)
	if err != nil {
		return nil, err
	}
	decimalStr := string(resultQuery.QueryResult)
	decimal, _ := strconv.ParseUint(decimalStr[1:len(decimalStr)-1], 10, 64)
	var result adaptor.GetAssetDecimalOutput
	result.Decimal = uint(decimal)

	return &result, nil
}

//创建一个转账交易，但是未签名
func (aerc20 *AdaptorErc20) CreateTransferTokenTx(input *adaptor.CreateTransferTokenTxInput) (
	*adaptor.CreateTransferTokenTxOutput, error) {
	return CreateTx(input) //add m and pack, return msg
}

//获取某个地址对某种Token的交易历史,支持分页和升序降序排列
func (aerc20 *AdaptorErc20) GetAddrTxHistory(input *adaptor.GetAddrTxHistoryInput) (*adaptor.GetAddrTxHistoryOutput,
	error) {
	return GetAddrErc20TxHistoryHTTP(aerc20.TxQueryUrl, input) // use web api
}

//根据交易ID获得对应的转账交易
func (aerc20 *AdaptorErc20) GetTransferTx(input *adaptor.GetTransferTxInput) (*adaptor.GetTransferTxOutput, error) {
	return GetTransferTx(input, &aerc20.RPCParams, aerc20.NetID)
}

//创建一个多签地址，该地址必须要满足signCount个签名才能解锁
func (aerc20 *AdaptorErc20) CreateMultiSigAddress(input *adaptor.CreateMultiSigAddressInput) (
	*adaptor.CreateMultiSigAddressOutput, error) {
	//return &adaptor.CreateMultiSigAddressOutput{Address: aerc20.lockContractAddress}, nil
	return nil, errors.New("please deploy multi-sign contract yourself")
}

//获取最新区块头
func (aerc20 *AdaptorErc20) GetBlockInfo(input *adaptor.GetBlockInfoInput) (*adaptor.GetBlockInfoOutput, error) {
	return GetBlockInfo(input, &aerc20.RPCParams)
}
