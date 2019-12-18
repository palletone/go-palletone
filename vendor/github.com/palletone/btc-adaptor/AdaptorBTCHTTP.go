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
package btcadaptor

import (
	"errors"

	"github.com/palletone/adaptor"
)

type AdaptorBTCHTTP struct {
	NetID int
	RPCParams
}

/*IUtility*/
//创建一个新的私钥
func (abtc *AdaptorBTCHTTP) NewPrivateKey(input *adaptor.NewPrivateKeyInput) (*adaptor.NewPrivateKeyOutput, error) {
	prikey, err := NewPrivateKey(abtc.NetID)
	if err != nil {
		return nil, err
	}
	result := adaptor.NewPrivateKeyOutput{PrivateKey: prikey}
	return &result, nil
}

//根据私钥创建公钥
func (abtc *AdaptorBTCHTTP) GetPublicKey(input *adaptor.GetPublicKeyInput) (*adaptor.GetPublicKeyOutput, error) {
	pubkey, err := GetPublicKey(input.PrivateKey, abtc.NetID)
	if err != nil {
		return nil, err
	}
	result := adaptor.GetPublicKeyOutput{PublicKey: pubkey}
	return &result, nil
}

//根据Key创建地址
func (abtc *AdaptorBTCHTTP) GetAddress(key *adaptor.GetAddressInput) (*adaptor.GetAddressOutput, error) {
	addr, err := PubKeyToAddress(key.Key, abtc.NetID)
	if err != nil {
		return nil, err
	}
	result := adaptor.GetAddressOutput{Address: addr}
	return &result, nil
}

//获得原链的地址和PalletOne的地址的映射
func (abtc *AdaptorBTCHTTP) GetPalletOneMappingAddress(addr *adaptor.GetPalletOneMappingAddressInput) (*adaptor.GetPalletOneMappingAddressOutput, error) { //todo
	return nil, errors.New("todo")
}

func (abtc *AdaptorBTCHTTP) HashMessage(input *adaptor.HashMessageInput) (*adaptor.HashMessageOutput, error) {
	return HashMessage(input)
}

//对一条消息进行签名
func (abtc *AdaptorBTCHTTP) SignMessage(input *adaptor.SignMessageInput) (*adaptor.SignMessageOutput, error) {
	return SignMessage(input)
}

//对签名进行验证
func (abtc *AdaptorBTCHTTP) VerifySignature(input *adaptor.VerifySignatureInput) (*adaptor.VerifySignatureOutput, error) {
	return VerifySignature(input)
}

//对一条交易进行签名，并返回签名结果
func (abtc *AdaptorBTCHTTP) SignTransaction(input *adaptor.SignTransactionInput) (*adaptor.SignTransactionOutput, error) {
	return SignTransaction(input, abtc.NetID)
}

//将未签名的原始交易与签名进行绑定，返回一个签名后的交易
func (abtc *AdaptorBTCHTTP) BindTxAndSignature(input *adaptor.BindTxAndSignatureInput) (*adaptor.BindTxAndSignatureOutput, error) {
	return BindTxAndSignature(input, abtc.NetID)
}

//根据交易内容，计算交易Hash
func (abtc *AdaptorBTCHTTP) CalcTxHash(input *adaptor.CalcTxHashInput) (*adaptor.CalcTxHashOutput, error) {
	return CalcTxHash(input)
}

//将签名后的交易广播到网络中,如果发送交易需要手续费，指定最多支付的手续费
func (abtc *AdaptorBTCHTTP) SendTransaction(input *adaptor.SendTransactionInput) (*adaptor.SendTransactionOutput, error) { //todo zxl
	return nil, errors.New("todo")
}

//根据交易ID获得交易的基本信息
func (abtc *AdaptorBTCHTTP) GetTxBasicInfo(input *adaptor.GetTxBasicInfoInput) (*adaptor.GetTxBasicInfoOutput, error) {
	return GetTxBasicInfoHttp(input, abtc.NetID)
}

//查询获得一个区块的信息
func (abtc *AdaptorBTCHTTP) GetBlockInfo(input *adaptor.GetBlockInfoInput) (*adaptor.GetBlockInfoOutput, error) { //todo zxl
	return nil, errors.New("todo")
}

/*ICryptoCurrency*/
//获取某地址下持有某资产的数量,返回数量为该资产的最小单位
func (abtc *AdaptorBTCHTTP) GetBalance(input *adaptor.GetBalanceInput) (*adaptor.GetBalanceOutput, error) { //todo zxl
	return nil, errors.New("todo")
}

//获取某资产的小数点位数
func (abtc *AdaptorBTCHTTP) GetAssetDecimal(asset *adaptor.GetAssetDecimalInput) (*adaptor.GetAssetDecimalOutput, error) {
	result := adaptor.GetAssetDecimalOutput{Decimal: 8}
	return &result, nil
}

//创建一个转账交易，但是未签名
func (abtc *AdaptorBTCHTTP) CreateTransferTokenTx(input *adaptor.CreateTransferTokenTxInput) (*adaptor.CreateTransferTokenTxOutput, error) { //todo
	return nil, errors.New("todo")
}

//获取某个地址对某种Token的交易历史,支持分页和升序降序排列
func (abtc *AdaptorBTCHTTP) GetAddrTxHistory(input *adaptor.GetAddrTxHistoryInput) (*adaptor.GetAddrTxHistoryOutput, error) { //todo
	return nil, errors.New("todo")
}

//根据交易ID获得对应的转账交易
func (abtc *AdaptorBTCHTTP) GetTransferTx(input *adaptor.GetTransferTxInput) (*adaptor.GetTransferTxOutput, error) { //todo zxl
	return nil, errors.New("todo")
}

//创建一个多签地址，该地址必须要满足signCount个签名才能解锁
func (abtc *AdaptorBTCHTTP) CreateMultiSigAddress(input *adaptor.CreateMultiSigAddressInput) (*adaptor.CreateMultiSigAddressOutput, error) {
	return CreateMultiSigAddress(input, abtc.NetID)
}

func (abtc *AdaptorBTCHTTP) CreateMultiSigPayoutTx(input *adaptor.CreateMultiSigPayoutTxInput) (*adaptor.CreateMultiSigPayoutTxOutput, error) {
	return nil, errors.New("todo")
}

//func (abtc AdaptorBTCHTTP) GetUTXO(params *adaptor.GetUTXOParams) (*adaptor.GetUTXOResult, error) {
//	return GetUTXO(params, &abtc.RPCParams, abtc.NetID)
//}
//func (abtc AdaptorBTCHTTP) GetUTXOHttp(params *adaptor.GetUTXOHttpParams) (*adaptor.GetUTXOHttpResult, error) {
//	return GetUTXOHttp(params, abtc.NetID)
//}

//func (abtc AdaptorBTCHTTP) GetTransactionByHash(params *adaptor.GetTransactionByHashParams) (*adaptor.GetTransactionByHashResult, error) {
//	return GetTransactionByHash(params, &abtc.RPCParams)
//}
//func (abtc AdaptorBTCHTTP) GetTransactionHttp(params *adaptor.GetTransactionHttpParams) (*adaptor.GetTransactionHttpResult, error) {
//	return GetTransactionHttp(params, abtc.NetID)
//}

//func (abtc AdaptorBTCHTTP) GetBalance(params *adaptor.GetBalanceParams) (*adaptor.GetBalanceResult, error) {
//	return GetBalance(params, &abtc.RPCParams, abtc.NetID)
//}
//func (abtc AdaptorBTCHTTP) GetBalanceHttp(params *adaptor.GetBalanceHttpParams) (*adaptor.GetBalanceHttpResult, error) {
//	return GetBalanceHttp(params, abtc.NetID)
//}

//func (abtc AdaptorBTCHTTP) SendTransaction(params *adaptor.SendTransactionParams) (*adaptor.SendTransactionResult, error) {
//	return SendTransaction(params, &abtc.RPCParams)
//}
//func (abtc AdaptorBTCHTTP) SendTransactionHttp(params *adaptor.SendTransactionHttpParams) (*adaptor.SendTransactionHttpResult, error) {
//	return SendTransactionHttp(params, abtc.NetID)
//}
