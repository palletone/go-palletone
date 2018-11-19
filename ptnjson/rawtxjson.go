package ptnjson

import (
	"fmt"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/dag/errors"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/tokenengine"
	"github.com/shopspring/decimal"
)

// TransactionInput represents the inputs to a transaction.  Specifically a
// transaction hash and output number pair.
type TransactionInput struct {
	Txid         string `json:"txid"`
	Vout         uint32 `json:"vout"`
	MessageIndex uint32 `json:"messageindex"`
}

// CreateRawTransactionCmd defines the createrawtransaction JSON-RPC command.
type CreateRawTransactionCmd struct {
	Inputs   []TransactionInput
	Amounts  map[string]decimal.Decimal `jsonrpcusage:"{\"address\":amount,...}"` // In BTC
	LockTime *int64
}

// CreateVoteTransactionCmd defines the createrawtransaction JSON-RPC command.
type CreateVoteTransactionCmd struct {
	Inputs          []TransactionInput
	Amounts         map[string]decimal.Decimal `jsonrpcusage:"{\"address\":amount,...}"` // In BTC
	LockTime        *int64
	MediatorAddress string
	ExpiredTerm     uint16
}

// NewCreateRawTransactionCmd returns a new instance which can be used to issue
// a createrawtransaction JSON-RPC command.
//
// Amounts are in BTC.
func NewCreateRawTransactionCmd(inputs []TransactionInput, amounts map[string]decimal.Decimal,
	lockTime *int64) *CreateRawTransactionCmd {

	return &CreateRawTransactionCmd{
		Inputs:   inputs,
		Amounts:  amounts,
		LockTime: lockTime,
	}
}

func NewCreateVoteTransactionCmd(inputs []TransactionInput, amounts map[string]decimal.Decimal,
	lockTime *int64, mediatorAddress string, expiredTerm uint16) *CreateVoteTransactionCmd {

	return &CreateVoteTransactionCmd{
		Inputs:          inputs,
		Amounts:         amounts,
		LockTime:        lockTime,
		MediatorAddress: mediatorAddress,
		ExpiredTerm:     expiredTerm,
	}
}

type CmdTransactionGenParams struct {
	Address string `json:"address"`
	Outputs []struct {
		Address string          `json:"address"`
		Amount  decimal.Decimal `json:"amount"`
	} `json:"outputs"`
	Locktime int64 `json:"locktime"`
}

type RawTransactionGenParams struct {
	Inputs []struct {
		Txid         string `json:"txid"`
		Vout         uint32 `json:"vout"`
		MessageIndex uint32 `json:"messageindex"`
	} `json:"inputs"`
	Outputs []struct {
		Address string          `json:"address"`
		Amount  decimal.Decimal `json:"amount"`
	} `json:"outputs"`
	Locktime int64 `json:"locktime"`
}

type VoteTransactionGenParams struct {
	Inputs []struct {
		Txid         string `json:"txid"`
		Vout         uint32 `json:"vout"`
		MessageIndex uint32 `json:"messageindex"`
	} `json:"inputs"`
	Outputs []struct {
		Address string  `json:"address"`
		Amount  float64 `json:"amount"`
	} `json:"outputs"`
	Locktime int64 `json:"locktime"`
	// Additional fields
	MediatorAddress string `json:"mediatoraddress"`
	ExpiredTerm     uint16 `json:"expiredterm"`
}

func ConvertRawTxJson2Paymsg(rawTxJson RawTransactionGenParams) (*modules.PaymentPayload, error) {

	pay := modules.NewPaymentPayload([]*modules.Input{}, []*modules.Output{})
	for _, input := range rawTxJson.Inputs {
		preTxId := &common.Hash{}
		preTxId.SetHexString(input.Txid)
		txin := modules.NewTxIn(modules.NewOutPoint(preTxId, input.MessageIndex, input.Vout), nil)
		pay.AddTxIn(txin)
	}

	for _, out := range rawTxJson.Outputs {
		addr, err := common.StringToAddress(out.Address)
		if err != nil {
			return nil, errors.New("Invalid address:" + out.Address)
		}
		lockScript := tokenengine.GenerateLockScript(addr)
		txout := modules.NewTxOut(Ptn2Dao(out.Amount), lockScript, nil)
		pay.AddTxOut(txout)
	}

	return pay, nil
}
func Ptn2Dao(ptnAmount decimal.Decimal) uint64 {
	return uint64(ptnAmount.Mul(decimal.New(100000000, 0)).IntPart())
}
func Dao2Ptn(amount uint64) decimal.Decimal {
	d, _ := decimal.NewFromString(fmt.Sprintf("%d", amount))
	return d.Div(decimal.New(100000000, 0))
}

func ConvertRawTxJson2Tx(rawTxJson RawTransactionGenParams) *modules.Transaction {
	pay, _ := ConvertRawTxJson2Paymsg(rawTxJson)
	tx := modules.NewTransaction([]*modules.Message{})
	tx.AddMessage(&modules.Message{App: modules.APP_PAYMENT, Payload: pay})
	return tx
}
