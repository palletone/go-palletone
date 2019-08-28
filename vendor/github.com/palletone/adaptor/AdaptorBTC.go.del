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

type adapterbtc interface {
	NewPrivateKey() (wifPriKey string)
	GetPublicKey(wifPriKey string) (pubKey string)
	GetAddress(wifPriKey string) (address string)
	GetAddressByPubkey(pubKeyHex string) (string, error)
	CreateMultiSigAddress(params *CreateMultiSigParams) (string, error)

	RawTransactionGen(params *RawTransactionGenParams) (string, error)
	DecodeRawTransaction(params *DecodeRawTransactionParams) (string, error)
	GetTransactionByHash(params *GetTransactionByHashParams) (string, error)
	GetTransactionHttp(getTransactionByHashParams *GetTransactionHttpParams) (string, error)

	SignTransaction(params *SignTransactionParams) (string, error)
	SignTxSend(params *SignTxSendParams) (string, error)
	MergeTransaction(params *MergeTransactionParams) (string, error)

	SignMessage(signMessageParams *SignMessageParams) (string, error)
	VerifyMessage(verifyMessageParams *VerifyMessageParams) (string, error)

	GetUTXO(params *GetUTXOParams) string
	GetUTXOHttp(params *GetUTXOHttpParams) (string, error)

	GetBalance(params *GetBalanceParams) (string, error)
	GetTransactions(params *GetTransactionsParams) (string, error)
	SendTransaction(params *SendTransactionParams) string
	SendTransactionHttp(sendTransactionParams *SendTransactionHttpParams) (string, error)
}

//
type CreateMultiSigParams struct {
	PublicKeys []string `json:"publicKeys"`
	N          int      `json:"n"`
	M          int      `json:"m"`
}
type CreateMultiSigResult struct {
	P2ShAddress  string   `json:"p2sh_address"`
	RedeemScript string   `json:"redeem_script"`
	Addresses    []string `json:"addresses"`
}

//
type RawTransactionGenParams struct {
	Inputs   []Input  `json:"inputs"`
	Outputs  []Output `json:"outputs"`
	Locktime int64    `json:"locktime"`
}
type RawTransactionGenResult struct {
	Rawtx string `json:"rawtx"`
}

//
type DecodeRawTransactionParams struct {
	Rawtx string `json:"rawtx"`
}
type Input struct {
	Txid string `json:"txid"`
	Vout uint32 `json:"vout"`
	Addr string `json:"addr"`
}
type Output struct {
	Address string  `json:"address"`
	Amount  float64 `json:"amount"` //btc
}
type DecodeRawTransactionResult struct {
	Inputs   []Input  `json:"inputs"`
	Outputs  []Output `json:"outputs"`
	Locktime uint32   `json:"locktime"`
}

//
type GetTransactionByHashParams struct {
	TxHash string `json:"txhash"`
}
type GetTransactionByHashResult struct {
	Txid     string        `json:"txid"`
	Confirms uint64        `json:"confirms"`
	Inputs   []Input       `json:"inputs"`
	Outputs  []OutputIndex `json:"outputs"`
}

//
type GetTransactionHttpParams struct {
	TxHash string `json:"txhash"`
}
type GetTransactionHttpResult struct {
	Txid     string        `json:"txid"`
	Confirms uint64        `json:"confirms"`
	Inputs   []Input       `json:"inputs"`
	Outputs  []OutputIndex `json:"outputs"`
}

//
type SignTransactionParams struct {
	TransactionHex   string   `json:"transactionhex"`
	InputRedeemIndex []int    `json:"inputredeemindex"`
	RedeemHex        []string `json:"redeemhex"`
	FromAddr         string   `json:"fromaddr"` //empty when spend from MultiAddr
	Privkeys         []string `json:"privkeys"` //wif private keys
}
type SignTransactionResult struct {
	Complete       bool   `json:"complete"`
	TransactionHex string `json:"transactionhex"`
}

//
type SendTransactionParams struct {
	TransactionHex string `json:"transactionhex"`
}
type SendTransactionResult struct {
	TransactionHah string `json:"transactionhash"`
}

//
type SendTransactionHttpParams struct {
	TransactionHex string `json:"transactionhex"`
}
type SendTransactionHttpResult struct {
	TransactionHah string `json:"transactionhash"`
}

//
type SignTxSendParams struct {
	TransactionHex   string   `json:"transactionhex"`
	InputRedeemIndex []int    `json:"inputredeemindex"`
	RedeemHex        []string `json:"redeemhex"`
	FromAddr         string   `json:"fromaddr"`
	Privkeys         []string `json:"privkeys"` //wif private keys
}
type SignTxSendResult struct {
	TransactionHah string `json:"transactionhash"`
	Complete       bool   `json:"complete"`
	TransactionHex string `json:"transactionhex"`
}

//
type MergeTransactionParams struct {
	UserTransactionHex   string   `json:"usertransactionhex"`
	MergeTransactionHexs []string `json:"mergetransactionhexs"`
	InputRedeemIndex     []int    `json:"inputredeemindex"`
	RedeemHex            []string `json:"redeemhex"`
}
type MergeTransactionResult struct {
	Complete        bool   `json:"complete"`
	TransactionHash string `json:"transactionhash"`
	TransactionHex  string `json:"transactionhex"`
}

//
type SignMessageParams struct {
	Message string `json:"message"`
	Privkey string `json:"privkey"` //wif private key
}
type SignMessageResult struct {
	Signature string `json:"signature"`
}

//
type VerifyMessageParams struct {
	Message   string `json:"message"`
	Signature string `json:"signature"`
	Address   string `json:"address"`
}
type VerifyMessageResult struct {
	Valid bool `json:"valid"`
}

//
type UTXO struct {
	TxID     string  `json:"txid"`
	Vout     uint32  `json:"vout"`
	Amount   float64 `json:"amount"`
	Confirms uint64  `json:"confirms"`
}

//
type GetUTXOParams struct {
	Address      string `json:"address"`
	Minconf      int    `json:"minconf"`
	Maxconf      int    `json:"maxconf"`
	MaximumCount int    `json:"maximumCount"`
}
type GetUTXOResult struct {
	UTXOs []UTXO `json:"utxos"`
}

//
type GetUTXOHttpParams struct {
	Address string `json:"address"`
	Txid    string `json:"txid"`
}
type GetUTXOHttpResult struct {
	UTXOs []UTXO `json:"utxos"`
}

//
type GetBalanceParams struct {
	Address string `json:"address"`
	Minconf int    `json:"minconf"`
}
type GetBalanceResult struct {
	Value float64 `json:"value"` //btc
}

//
type GetBalanceHttpParams struct {
	Address string `json:"address"`
	Minconf int    `json:"minconf"`
}
type GetBalanceHttpResult struct {
	Value float64 `json:"value"` //btc
}

//
type GetTransactionsParams struct {
	Account string `json:"account"`
	Count   int    `json:"count"`
	Skip    int    `json:"skip"`
}

type InputIndex struct {
	TxHash string  `json:"txHash"`
	Index  uint32  `json:"index"`
	Addr   string  `json:"addr"`
	Value  float64 `json:"value"` //btc
}
type OutputIndex struct {
	Index uint32  `json:"index"`
	Addr  string  `json:"addr"`
	Value float64 `json:"value"` //btc
}
type Transaction struct {
	TxHash        string        `json:"txHash"`
	BlanceChanged float64       `json:"blanceChanged"` //btc
	Confirms      uint64        `json:"confirms"`
	Inputs        []InputIndex  `json:"inputs"`
	Outputs       []OutputIndex `json:"outputs"`
}
type TransactionsResult struct {
	Transactions []Transaction `json:"transactions"`
}

/* not used current

type GetUTXOParams struct {
	Addresses    []string `json:"addresses"`
	Minconf      int      `json:"minconf"`
	Maxconf      int      `json:"maxconf"`
	MaximumCount int      `json:"maximumCount"`
}

func (abtc AdaptorBTC) GetUTXO(params string) string {
	return GetUTXO(params, &abtc.RPCParams, abtc.NetID)
}

*/
