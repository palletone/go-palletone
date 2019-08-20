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
package adaptor

type adaptereth interface {
	NewPrivateKey() (prikeyHex string)
	GetPublicKey(prikeyHex string) (pubKey string)
	GetAddress(prikeyHex string) (address string)

	CreateMultiSigAddress(params *CreateMultiSigAddressParams) (string, error)
	Keccak256HashPackedSig(params *Keccak256HashPackedSigParams) (string, error)
	RecoverAddr(params *RecoverParams) (string, error)

	SignTransaction(params *ETHSignTransactionParams) (string, error) //not same as btc, members
	SendTransaction(params *SendTransactionParams) (string, error)    //same as btc

	QueryContract(params *QueryContractParams) (string, error)
	GenInvokeContractTX(params *GenInvokeContractTXParams) (string, error)
	GenDeployContractTX(params *GenDeployContractTXParams) (string, error)

	GetEventByAddress(params *GetEventByAddressParams) (string, error)

	GetTransactionByHash(params *GetTransactionParams) (string, error)
	GetErc20TxByHash(params *GetErc20TxByHashParams) (string, error)

	GetBestHeader(params *GetBestHeaderParams) (string, error)
}

//
type CreateMultiSigAddressParams struct {
	Addresses []string `json:"addresses"`
	N         int      `json:"n"`
	M         int      `json:"m"`
}
type CreateMultiSigAddressResult struct {
	RedeemHex string `json:"redeemhex"`
}

//
type Keccak256HashPackedSigParams struct {
	PrivateKeyHex string `json:"privatekeyhex"`
	ParamTypes    string `json:"paramtypes"`
	Params        string `json:"params"`
}
type Keccak256HashPackedSigResult struct {
	Hash      string `json:"hash"`
	Signature string `json:"signature"`
}

//
type Keccak256HashVerifyParams struct {
	PublicKeyHex string `json:"publickeyhex"`
	Hash         string `json:"hash"`
	Signature    string `json:"signature"`
}

type Keccak256HashVerifyResult struct {
	Valid bool `json:"valid"`
}

//
type RecoverParams struct {
	Hash      string `json:"hash"`
	Signature string `json:"signature"`
}

type RecoverResult struct {
	Addr string `json:"addr"`
}

//not same as btc, members
type ETHSignTransactionParams struct {
	PrivateKeyHex  string `json:"privatekeyhex"`
	TransactionHex string `json:"transactionhex"`
}
type ETHSignTransactionResult struct {
	TransactionHex string `json:"transactionhex"`
}

////same as btc
//type SendTransactionParams struct {
//	TransactionHex string `json:"transactionhex"`
//}
//type SendTransactionResult struct {
//	TransactionHah string `json:"transactionhash"`
//}

//
type QueryContractParams struct {
	ContractABI  string        `json:"contractABI"`
	ContractAddr string        `json:"contractAddr"`
	Method       string        `json:"method"`
	Params       string        `json:"params"`
	ParamsArray  []interface{} `json:"paramsarray"`
}
type QueryContractResult struct {
	Result string `json:"result"`
}

//
type GenInvokeContractTXParams struct {
	ContractABI  string        `json:"contractabi"`
	ContractAddr string        `json:"contractaddr"`
	CallerAddr   string        `json:"calleraddr"`
	Value        string        `json:"value"`
	GasPrice     string        `json:"gasprice"`
	GasLimit     string        `json:"gaslimit"`
	Method       string        `json:"method"`
	Params       string        `json:"params"`
	ParamsArray  []interface{} `json:"paramsarray"`
}
type GenInvokeContractTXResult struct {
	TransactionHex string `json:"transactionhex"`
}

//
type GenDeployContractTXParams struct {
	ContractABI  string        `json:"contractabi"`
	ContractBin  string        `json:"contractbin"`
	DeployerAddr string        `json:"deployeraddr"`
	Value        string        `json:"value"`
	GasPrice     string        `json:"gasprice"`
	GasLimit     string        `json:"gaslimit"`
	Params       string        `json:"params"`
	ParamsArray  []interface{} `json:"paramsarray"`
}
type GenDeployContractTXResult struct {
	TransactionHex string `json:"transactionhex"`
	ContractAddr   string `json:"contractaddr"`
}

//
type GetEventByAddressParams struct {
	ContractABI  string `json:"contractABI"`
	ContractAddr string `json:"contractAddr"`
	ConcernAddr  string `json:"concernaddr"`
	StartHeight  string `json:"startheight"`
	EndHeight    string `json:"endheight"`
	EventName    string `json:"eventname"`
}
type GetEventByAddressResult struct {
	Events    []string `json:"events"`
	Txhashs   []string `json:"txhashs"`
	Blocknums []uint64 `json:"blocknums"`
}

//
type GetTransactionParams struct {
	Hash string `json:"hash"`
}
type GetTransactionResult struct {
	Hash        string `json:"hash"`
	Nonce       string `json:"nonce"`
	BlockHash   string `json:"blockHash"`
	BlockNumber string `json:"blockNumber"`
	From        string `json:"from"`
	To          string `json:"to"`
	Value       string `json:"value"`
	Gas         string `json:"gas"`
	GasPrice    string `json:"gasPrice"`
	Input       string `json:"input"`
}

//
type GetErc20TxByHashParams struct {
	Hash string `json:"hash"`
}
type GetErc20TxByHashResult struct {
	Hash         string `json:"hash"`
	Status       string `json:"status"`
	BlockHash    string `json:"blockHash"`
	BlockNumber  string `json:"blockNumber"`
	ContractAddr string `json:"contractaddr"`
	From         string `json:"from"`
	To           string `json:"to"`
	Amount       string `json:"amount"`
}

//
type GetBestHeaderParams struct {
	Number string `json:"Number"` //if empty, return the best header
}
type GetBestHeaderResult struct {
	TxHash string `json:"txhash"`
	Number string `json:"number"`
}

/* not used current
//
type GetBalanceParams struct {
	Account string `json:"account"`
}
type GetBalanceResult struct {
	Balance float64 `json:"balance"`
}
func (aeth AdaptorETH) GetBalance(params string) string {
	return GetBalance(params, &aeth.RPCParams, aeth.NetID)
}

//
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
func (aeth AdaptorETH) CalculateSig(params string) string {
	return CalculateSig(params)
}

*/
