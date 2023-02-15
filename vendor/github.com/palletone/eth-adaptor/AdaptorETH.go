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

	"github.com/palletone/adaptor"
)

type RPCParams struct {
	Rawurl     string `json:"rawurl"`
	TxQueryUrl string
}

type AdaptorETH struct {
	NetID int
	RPCParams
	//lockContractAddress string
}

func NewAdaptorETHTestnet() *AdaptorETH {
	return &AdaptorETH{
		NetID: NETID_TEST,
		RPCParams: RPCParams{Rawurl: "https://ropsten.infura.io",
			TxQueryUrl: "https://api-ropsten.etherscan.io/api?apikey=VYSBPQ383RJXM7HBQVTIK5NGIG8ZYVV6T6"},
		//lockContractAddress:"0x4d736ed88459b2db85472aab13a9d0ce2a6ea676",
	}
}
func NewAdaptorETHMainnet() *AdaptorETH {
	return &AdaptorETH{
		NetID: NETID_MAIN,
		RPCParams: RPCParams{Rawurl: "https://mainnet.infura.io",
			TxQueryUrl: "https://api.etherscan.io/api?apikey=VYSBPQ383RJXM7HBQVTIK5NGIG8ZYVV6T6"},
		//lockContractAddress:"0x1989a21eb0f28063e47e6b448e8d76774bc9b493",
	}
}

const (
	NETID_MAIN = iota
	NETID_TEST
)

/*IUtility*/
//创建一个新的私钥
func (aeth *AdaptorETH) NewPrivateKey(input *adaptor.NewPrivateKeyInput) (*adaptor.NewPrivateKeyOutput, error) {
	prikey, err := NewPrivateKey(aeth.NetID)
	if err != nil {
		return nil, err
	}
	result := adaptor.NewPrivateKeyOutput{PrivateKey: prikey}
	return &result, nil
}

//根据私钥创建公钥
func (aeth *AdaptorETH) GetPublicKey(input *adaptor.GetPublicKeyInput) (
	*adaptor.GetPublicKeyOutput, error) {
	pubkey, err := GetPublicKey(input.PrivateKey)
	if err != nil {
		return nil, err
	}
	result := adaptor.GetPublicKeyOutput{PublicKey: pubkey}
	return &result, nil
}

//根据Key创建地址
func (aeth *AdaptorETH) GetAddress(key *adaptor.GetAddressInput) (
	*adaptor.GetAddressOutput, error) {
	addr, err := PubKeyToAddress(key.Key)
	if err != nil {
		return nil, err
	}
	result := adaptor.GetAddressOutput{Address: addr}
	return &result, nil
}
func (aeth *AdaptorETH) GetPalletOneMappingAddress(addrInput *adaptor.GetPalletOneMappingAddressInput) (
	*adaptor.GetPalletOneMappingAddressOutput, error) {
	if len(addrInput.MappingDataSource) == 0 {
		return nil, errors.New("you must set mapping contract address in MappingDataSource of addrInput")
	}
	return GetMappAddr(addrInput, &aeth.RPCParams)
}

func (aerc20 *AdaptorETH) HashMessage(input *adaptor.HashMessageInput) (*adaptor.HashMessageOutput, error) {
	return HashMessage(input)
}

//对一条消息进行签名
func (aeth *AdaptorETH) SignMessage(input *adaptor.SignMessageInput) (
	*adaptor.SignMessageOutput, error) {
	return SignMessage(input)
}

//对签名进行验证
func (aeth *AdaptorETH) VerifySignature(input *adaptor.VerifySignatureInput) (
	*adaptor.VerifySignatureOutput, error) {
	return VerifySignature(input)
}

//对一条交易进行签名，并返回签名结果 //call SignMessage
func (aeth *AdaptorETH) SignTransaction(input *adaptor.SignTransactionInput) (
	*adaptor.SignTransactionOutput, error) {
	if 'm' == input.Transaction[0] && 's' == input.Transaction[1] && 'g' == input.Transaction[2] {
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

//将未签名的原始交易与签名进行绑定，返回一个签名后的交易 //产生 contractTx input data
func (aeth *AdaptorETH) BindTxAndSignature(input *adaptor.BindTxAndSignatureInput) (
	*adaptor.BindTxAndSignatureOutput,
	error) {
	return BindTxAndSignature(input) //first append methodID, then msg[33:], then pack signatures
}

//根据交易内容，计算交易Hash //need implement simple hash
func (aeth *AdaptorETH) CalcTxHash(input *adaptor.CalcTxHashInput) (*adaptor.CalcTxHashOutput, error) {
	return CalcTxHash(input)
}

//将签名后的交易广播到网络中,如果发送交易需要手续费，指定最多支付的手续费
func (aeth *AdaptorETH) SendTransaction(input *adaptor.SendTransactionInput) (*adaptor.SendTransactionOutput, error) {
	return SendTransaction(input, &aeth.RPCParams)
}

//根据交易ID获得交易的基本信息
func (aeth *AdaptorETH) GetTxBasicInfo(input *adaptor.GetTxBasicInfoInput) (*adaptor.GetTxBasicInfoOutput, error) {
	return GetTxBasicInfo(input, &aeth.RPCParams, aeth.NetID)
}

//获取最新区块头
func (aeth *AdaptorETH) GetBlockInfo(input *adaptor.GetBlockInfoInput) (*adaptor.GetBlockInfoOutput, error) {
	return GetBlockInfo(input, &aeth.RPCParams)
}

/*ICryptoCurrency*/
//获取某地址下持有某资产的数量,返回数量为该资产的最小单位
func (aeth *AdaptorETH) GetBalance(input *adaptor.GetBalanceInput) (*adaptor.GetBalanceOutput, error) {
	return GetBalanceETH(input, &aeth.RPCParams)
}

//获取某资产的小数点位数
func (aeth *AdaptorETH) GetAssetDecimal(asset *adaptor.GetAssetDecimalInput) (
	*adaptor.GetAssetDecimalOutput, error) {
	result := adaptor.GetAssetDecimalOutput{Decimal: 18}
	return &result, nil
}

//创建一个转账交易，但是未签名 //send eth tx, not sign
func (aeth *AdaptorETH) CreateTransferTokenTx(input *adaptor.CreateTransferTokenTxInput) (
	*adaptor.CreateTransferTokenTxOutput, error) {
	return CreateETHTx(input, &aeth.RPCParams)
}

//获取某个地址对某种Token的交易历史,支持分页和升序降序排列
func (aeth *AdaptorETH) GetAddrTxHistory(input *adaptor.GetAddrTxHistoryInput) (
	*adaptor.GetAddrTxHistoryOutput, error) {
	return GetAddrTxHistoryHTTP(aeth.TxQueryUrl, input) // use web api
}

//根据交易ID获得对应的转账交易
func (aeth *AdaptorETH) GetTransferTx(input *adaptor.GetTransferTxInput) (*adaptor.GetTransferTxOutput, error) {
	return GetTransferTx(input, &aeth.RPCParams, aeth.NetID, false)
}

//创建一个多签地址，该地址必须要满足signCount个签名才能解锁 //eth没有多签，需要自己部署多签合约
func (aeth *AdaptorETH) CreateMultiSigAddress(input *adaptor.CreateMultiSigAddressInput) (
	*adaptor.CreateMultiSigAddressOutput, error) {
	return nil, errors.New("please deploy multi-sign contract yourself")
}

//构造一个从多签地址付出Token的交易
func (aerc20 *AdaptorETH) CreateMultiSigPayoutTx(input *adaptor.CreateMultiSigPayoutTxInput) (
	*adaptor.CreateMultiSigPayoutTxOutput, error) {
	//return &adaptor.CreateMultiSigAddressOutput{Address: aerc20.lockContractAddress}, nil
	return nil, errors.New("to do") //todo
}

/*ISmartContract*/
//创建一个安装合约的交易，未签名 //使用eth合约创建交易，没有安装， not implement
func (aeth *AdaptorETH) CreateContractInstallTx(input *adaptor.CreateContractInstallTxInput) (
	*adaptor.CreateContractInstallTxOutput, error) {
	return nil, errors.New("not implement")
}

//查询合约安装的结果的交易 //使用eth合约创建交易，没有安装， not implement
func (aeth *AdaptorETH) GetContractInstallTx(input *adaptor.GetContractInstallTxInput) (
	*adaptor.GetContractInstallTxOutput, error) {
	return nil, errors.New("not implement")
}

//初始化合约实例 //eth合约创建交易的生成
func (aeth *AdaptorETH) CreateContractInitialTx(input *adaptor.CreateContractInitialTxInput) (
	*adaptor.CreateContractInitialTxOutput, error) {
	return CreateContractInitialTx(input, &aeth.RPCParams, aeth.NetID)
}

//查询初始化合约实例的交易 //查询eth合约创建交易，是合约查询的封装
func (aeth *AdaptorETH) GetContractInitialTx(input *adaptor.GetContractInitialTxInput) (
	*adaptor.GetContractInitialTxOutput, error) {
	return GetContractInitialTx(input, &aeth.RPCParams, aeth.NetID)
}

//调用合约方法 //eth合约调用交易的生成
func (aeth *AdaptorETH) CreateContractInvokeTx(input *adaptor.CreateContractInvokeTxInput) (
	*adaptor.CreateContractInvokeTxOutput, error) {
	return CreateContractInvokeTx(input, &aeth.RPCParams, aeth.NetID)
}

//查询调用合约方法的交易 //eth合约查询不产生交易， not implement
func (aeth *AdaptorETH) GetContractInvokeTx(input *adaptor.GetContractInvokeTxInput) (
	*adaptor.GetContractInvokeTxOutput, error) {
	return nil, errors.New("not implement")
}

//调用合约的查询方法 //eth合约查询的生成
func (aeth *AdaptorETH) QueryContract(input *adaptor.QueryContractInput) (
	*adaptor.QueryContractOutput, error) {
	return QueryContract(input, &aeth.RPCParams)
}
