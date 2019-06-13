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

package modules

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/big"
	"strconv"
	"time"

	"bytes"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/award"
	"github.com/palletone/go-palletone/common/crypto"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/obj"
	"github.com/palletone/go-palletone/common/util"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/dag/errors"
	"github.com/palletone/go-palletone/dag/parameter"
)

var (
	TXFEE       = big.NewInt(100000000) // transaction fee =1ptn
	TX_MAXSIZE  = (256 * 1024)
	TX_BASESIZE = (100 * 1024) //100kb
)
var DepositContractLockScript = common.Hex2Bytes("140000000000000000000000000000000000000001c8")

// TxOut defines a bitcoin transaction output.
type TxOut struct {
	Value    int64  `json:"value"`
	PkScript []byte `json:"pk_script"`
	Asset    *Asset `json:"asset_info"`
}

// TxIn defines a bitcoin transaction input.
type TxIn struct {
	PreviousOutPoint *OutPoint `json:"pre_outpoint"`
	SignatureScript  []byte    `json:"signature_script"`
	Sequence         uint32    `json:"sequence"`
}

func NewTransaction(msg []*Message) *Transaction {
	return newTransaction(msg)
}

func NewContractCreation(msg []*Message) *Transaction {
	return newTransaction(msg)
}

func newTransaction(msg []*Message) *Transaction {
	tx := new(Transaction)
	for _, m := range msg {
		tx.TxMessages = append(tx.TxMessages, m)
	}
	return tx
}

// AddTxIn adds a transaction input to the message.
func (tx *Transaction) AddMessage(msg *Message) {
	tx.TxMessages = append(tx.TxMessages, msg)
}

// AddTxIn adds a transaction input to the message.
func (pld *PaymentPayload) AddTxIn(ti *Input) {
	pld.Inputs = append(pld.Inputs, ti)
}

// AddTxOut adds a transaction output to the message.
func (pld *PaymentPayload) AddTxOut(to *Output) {
	pld.Outputs = append(pld.Outputs, to)
}

type TransactionWithUnitInfo struct {
	*Transaction
	UnitHash  common.Hash
	UnitIndex uint64
	Timestamp uint64
	TxIndex   uint64
}

type TxPoolTransaction struct {
	Tx *Transaction

	From         []*OutPoint
	CreationDate time.Time `json:"creation_date"`
	Priority_lvl string    `json:"priority_lvl"` // 打包的优先级
	UnitHash     common.Hash
	UnitIndex    uint64
	Pending      bool
	Confirmed    bool
	IsOrphan     bool
	Discarded    bool // will remove
	//TxFee        *AmountAsset `json:"tx_fee"`
	TxFee      []*Addition `json:"tx_fee"`
	Index      uint64      `json:"index"` // index 是该Unit位置。
	Extra      []byte
	Tag        uint64
	Expiration time.Time
	//该Tx依赖于哪些TxId作为先决条件
	DependOnTxs []common.Hash
}

func (tx *TxPoolTransaction) Less(otherTx interface{}) bool {
	ap, _ := strconv.ParseFloat(tx.Priority_lvl, 64)
	bp, _ := strconv.ParseFloat(otherTx.(*TxPoolTransaction).Priority_lvl, 64)
	return ap < bp
}

func (tx *TxPoolTransaction) GetPriorityLvl() string {
	// priority_lvl=  fee/size*(1+(time.Now-CreationDate)/24)
	level, _ := strconv.ParseFloat(tx.Priority_lvl, 64)
	if level > 0 {
		return tx.Priority_lvl
	}
	var priority_lvl float64
	if txfee := tx.GetTxFee(); txfee.Int64() > 0 {
		// t0, _ := time.Parse(TimeFormatString, tx.CreationDate)
		if tx.CreationDate.Unix() <= 0 {
			tx.CreationDate = time.Now()
		}
		priority_lvl, _ = strconv.ParseFloat(fmt.Sprintf("%f", float64(txfee.Int64())/tx.Tx.Size().Float64()*(1+float64(time.Now().Second()-tx.CreationDate.Second())/(24*3600))), 64)
	}
	tx.Priority_lvl = strconv.FormatFloat(priority_lvl, 'E', -1, 64)
	return tx.Priority_lvl
}
func (tx *TxPoolTransaction) GetPriorityfloat64() float64 {
	level, _ := strconv.ParseFloat(tx.Priority_lvl, 64)
	if level > 0 {
		return level
	}
	var priority_lvl float64
	if txfee := tx.GetTxFee(); txfee.Int64() > 0 {
		// t0, _ := time.Parse(TimeFormatString, tx.CreationDate)
		if tx.CreationDate.Unix() <= 0 {
			tx.CreationDate = time.Now()
		}
		priority_lvl, _ = strconv.ParseFloat(fmt.Sprintf("%f", float64(txfee.Int64())/tx.Tx.Size().Float64()*(1+float64(time.Now().Second()-tx.CreationDate.Second())/(24*3600))), 64)
	}
	return priority_lvl
}
func (tx *TxPoolTransaction) SetPriorityLvl(priority float64) {
	tx.Priority_lvl = strconv.FormatFloat(priority, 'E', -1, 64)
}
func (tx *TxPoolTransaction) GetTxFee() *big.Int {
	var fee uint64
	if tx.TxFee != nil {
		for _, ad := range tx.TxFee {
			fee += ad.Amount
		}
	} else {
		fee = 20 // 20dao
		//tx.TxFee = &AmountAsset{Amount: 20, Asset: tx.Tx.Asset()}
	}
	return big.NewInt(int64(fee))
}

// Hash hashes the RLP encoding of tx.
// It uniquely identifies the transaction.
func (tx *Transaction) Hash() common.Hash {
	//	b, err := json.Marshal(tx)
	//	if err != nil {
	//		log.Error("json marshal error", "error", err)
	//		return common.Hash{}
	//	}
	//	v := rlp.RlpHash(b[:])
	//	return v
	//}
	//func (tx *Transaction) Hash_old() common.Hash {
	tx.Illegal = false

	v := util.RlpHash(tx)
	return v
}

func (tx *Transaction) RequestHash() common.Hash {
	req := &Transaction{}
	for _, msg := range tx.TxMessages {
		req.AddMessage(msg)
		if msg.App >= APP_CONTRACT_TPL_REQUEST { //100以上的APPCode是请求
			break
		}
	}
	return util.RlpHash(req)
}

func (tx *Transaction) ContractIdBytes() []byte {
	for _, msg := range tx.TxMessages {
		switch msg.App {
		case APP_CONTRACT_DEPLOY_REQUEST:
			//tmp := common.BytesToAddress(tx.RequestHash().Bytes())
			//out := common.NewAddress(tmp.Bytes(), common.ContractHash)
			addr := crypto.RequestIdToContractAddress(tx.RequestHash())
			return addr.Bytes()
		case APP_CONTRACT_INVOKE_REQUEST:
			payload := msg.Payload.(*ContractInvokeRequestPayload)
			return payload.ContractId
		case APP_CONTRACT_STOP_REQUEST:
			payload := msg.Payload.(*ContractStopRequestPayload)
			return payload.ContractId
		}
	}
	return nil
}

func (tx *Transaction) Messages() []*Message {
	return tx.TxMessages[:]
}

// Size returns the true RLP encoded storage UnitSize of the transaction, either by
// encoding and returning it, or returning a previsouly cached value.
func (tx *Transaction) Size() common.StorageSize {
	return CalcDateSize(tx)
}

func (tx *Transaction) CreateDate() string {
	n := time.Now()
	return n.Format(TimeFormatString)
}

// address return the tx's original address  of from and to
//func (tx *Transaction) GetAddressInfo() ([]*OutPoint, [][]byte) {
//	froms := make([]*OutPoint, 0)
//	tos := make([][]byte, 0)
//	if len(tx.Messages()) > 0 {
//		msg := tx.Messages()[0]
//		if msg.App == APP_PAYMENT {
//			payment, ok := msg.Payload.(*PaymentPayload)
//			if ok {
//				for _, input := range payment.Inputs {
//					if input.PreviousOutPoint != nil {
//						froms = append(froms, input.PreviousOutPoint)
//					}
//				}
//
//				for _, out := range payment.Outputs {
//					tos = append(tos, out.PkScript[:])
//				}
//			}
//		}
//	}
//	return froms, tos
//}
func (tx *Transaction) Asset() *Asset {
	if tx == nil {
		return nil
	}
	asset := new(Asset)
	msg := tx.Messages()[0]
	if msg.App == APP_PAYMENT {
		pay := msg.Payload.(*PaymentPayload)
		for _, out := range pay.Outputs {
			if out.Asset != nil {
				asset.AssetId = out.Asset.AssetId
				asset.UniqueId = out.Asset.UniqueId
				break
			}
		}
	}
	return asset
}
func (tx *Transaction) CopyFrTransaction(cpy *Transaction) {
	obj.DeepCopy(&tx, cpy)
}

// Len returns the length of s.
func (s Transactions) Len() int { return len(s) }

// Swap swaps the i'th and the j'th element in s.
func (s Transactions) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// GetRlp implements Rlpable and returns the i'th element of s in rlp.
func (s Transactions) GetRlp(i int) []byte {
	enc, _ := rlp.EncodeToBytes(s[i])
	return enc
}
func (s Transactions) Hash() common.Hash {
	b, err := json.Marshal(s)
	if err != nil {
		log.Error("json marshal error", "error", err)
		return common.Hash{}
	}

	v := util.RlpHash(b[:])
	return v
}

// TxDifference returns a new set t which is the difference between a to b.
func TxDifference(a, b Transactions) (keep Transactions) {
	keep = make(Transactions, 0, len(a))

	remove := make(map[common.Hash]struct{})
	for _, tx := range b {
		remove[tx.Hash()] = struct{}{}
	}

	for _, tx := range a {
		if _, ok := remove[tx.Hash()]; !ok {
			keep = append(keep, tx)
		}
	}

	return keep
}

// TxByPrice implements both the sort and the heap interface, making it useful
// for all at once sorting as well as individually adding and removing elements.
type TxByPrice TxPoolTxs

func (s TxByPrice) Len() int      { return len(s) }
func (s TxByPrice) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s *TxByPrice) Push(x interface{}) {
	*s = append(*s, x.(*TxPoolTransaction))
}
func (s *TxByPrice) Pop() interface{} {
	old := *s
	n := len(old)
	x := old[n-1]
	*s = old[0 : n-1]
	return x
}

type TxByPriority []*TxPoolTransaction

func (s TxByPriority) Len() int           { return len(s) }
func (s TxByPriority) Less(i, j int) bool { return s[i].Priority_lvl > s[j].Priority_lvl }
func (s TxByPriority) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func (s *TxByPriority) Push(x interface{}) {
	*s = append(*s, x.(*TxPoolTransaction))
}

func (s *TxByPriority) Pop() interface{} {
	old := *s
	n := len(old)
	x := old[n-1]
	*s = old[0 : n-1]
	return x
}

type TxByCreationDate []*TxPoolTransaction

func (tc TxByCreationDate) Len() int           { return len(tc) }
func (tc TxByCreationDate) Less(i, j int) bool { return tc[i].Priority_lvl > tc[j].Priority_lvl }
func (tc TxByCreationDate) Swap(i, j int)      { tc[i], tc[j] = tc[j], tc[i] }

type WriteCounter common.StorageSize

func (c *WriteCounter) Write(b []byte) (int, error) {
	*c += WriteCounter(len(b))
	return len(b), nil
}

func CalcDateSize(data interface{}) common.StorageSize {
	c := WriteCounter(0)
	rlp.Encode(&c, data)
	return common.StorageSize(c)
}

var (
	EmptyRootHash = core.DeriveSha(Transactions{})
)

type TxLookupEntry struct {
	UnitHash  common.Hash `json:"unit_hash"`
	UnitIndex uint64      `json:"unit_index"`
	Index     uint64      `json:"index"`
	Timestamp uint64      `json:"timestamp"`
}
type Transactions []*Transaction

func (txs Transactions) GetTxIds() []common.Hash {
	ids := make([]common.Hash, len(txs))
	for i, tx := range txs {
		ids[i] = tx.Hash()
	}
	return ids
}

type Transaction struct {
	TxMessages []*Message `json:"messages"`
	CertId     []byte     `json:"cert_id"` // should be big.Int byte
	Illegal    bool       `json:"Illegal"` // not hash, 1:no valid, 0:ok
}
type QueryUtxoFunc func(outpoint *OutPoint) (*Utxo, error)
type GetScriptSignersFunc func(tx *Transaction, msgIdx, inputIndex int) ([]common.Address, error)

//计算该交易的手续费，基于UTXO，所以传入查询UTXO的函数指针
func (tx *Transaction) GetTxFee(queryUtxoFunc QueryUtxoFunc, unitTime int64) (*AmountAsset, error) {
	log.Infof("Calculate tx fee,tx[%s]", tx.Hash().String())
	for _, msg := range tx.TxMessages {
		payload, ok := msg.Payload.(*PaymentPayload)
		if !ok {
			continue
		}
		if payload.IsCoinbase() {
			continue
		}
		inAmount := uint64(0)
		outAmount := uint64(0)
		for _, txin := range payload.Inputs {
			utxo, _ := queryUtxoFunc(txin.PreviousOutPoint)
			if utxo == nil {
				return nil, fmt.Errorf("Txin(txhash=%s, msgindex=%v, outindex=%v)'s utxo is empty:",
					txin.PreviousOutPoint.TxHash.String(),
					txin.PreviousOutPoint.MessageIndex,
					txin.PreviousOutPoint.OutIndex)
			}
			// check overflow
			if inAmount+utxo.Amount > (1<<64 - 1) {
				return nil, fmt.Errorf("Compute fees: txin total overflow")
			}
			inAmount += utxo.Amount
			if unitTime > 0 {
				//计算币龄利息
				rate := parameter.CurrentSysParameters.TxCoinDayInterest
				if bytes.Equal(utxo.PkScript, DepositContractLockScript) {
					rate = parameter.CurrentSysParameters.DepositContractInterest
				}

				interest := award.GetCoinDayInterest(utxo.GetTimestamp(), unitTime, utxo.Amount, rate)
				if interest > 0 {
					log.Infof("Calculate tx fee,Add interest value:%d to tx[%s] fee", interest, tx.Hash().String())
					inAmount += interest
				}
			}
		}

		for _, txout := range payload.Outputs {
			// check overflow
			if outAmount+txout.Value > (1<<64 - 1) {
				return nil, fmt.Errorf("Compute fees: txout total overflow")
			}
			outAmount += txout.Value
		}
		if inAmount < outAmount {

			return nil, fmt.Errorf("Compute fees: tx %s txin amount less than txout amount. amount:%d ,outAmount:%d ", tx.Hash().String(), inAmount, outAmount)
		}
		fees := inAmount - outAmount
		return &AmountAsset{Amount: fees, Asset: payload.Outputs[0].Asset}, nil

	}
	return nil, fmt.Errorf("Compute fees: no payment payload")
}

//该Tx如果保存后，会产生的新的Utxo
func (tx *Transaction) GetNewUtxos() map[OutPoint]*Utxo {
	result := map[OutPoint]*Utxo{}
	txHash := tx.Hash()
	for msgIndex, msg := range tx.TxMessages {
		if msg.App != APP_PAYMENT {
			continue
		}
		pay := msg.Payload.(*PaymentPayload)
		txouts := pay.Outputs
		for outIndex, txout := range txouts {
			utxo := &Utxo{
				Amount:   txout.Value,
				Asset:    txout.Asset,
				PkScript: txout.PkScript,
				LockTime: pay.LockTime,
			}

			// write to database
			outpoint := OutPoint{
				TxHash:       txHash,
				MessageIndex: uint32(msgIndex),
				OutIndex:     uint32(outIndex),
			}
			result[outpoint] = utxo
		}
	}
	return result
}

func (tx *Transaction) GetContractTxSignatureAddress() []common.Address {
	if !tx.IsContractTx() {
		return nil
	}
	addrs := make([]common.Address, 0)
	for _, msg := range tx.TxMessages {
		switch msg.App {
		case APP_SIGNATURE:
			payload := msg.Payload.(*SignaturePayload)
			for _, sig := range payload.Signatures {
				addrs = append(addrs, crypto.PubkeyBytesToAddress(sig.PubKey))
			}
		}
	}
	return addrs
}

//如果是合约调用交易，Copy其中的Msg0到ContractRequest的部分，如果不是请求，那么返回完整Tx
func (tx *Transaction) GetRequestTx() *Transaction {
	request := &Transaction{}
	request.CertId = tx.CertId
	for _, msg := range tx.TxMessages {
		if msg.App.IsRequest() {

			if msg.App == APP_CONTRACT_TPL_REQUEST {
				payload := new(ContractInstallRequestPayload)
				obj.DeepCopy(payload, msg.Payload)
				request.AddMessage(NewMessage(msg.App, payload))

			} else if msg.App == APP_CONTRACT_DEPLOY_REQUEST {
				payload := new(ContractDeployRequestPayload)
				obj.DeepCopy(payload, msg.Payload)
				request.AddMessage(NewMessage(msg.App, payload))
			} else if msg.App == APP_CONTRACT_INVOKE_REQUEST {
				payload := new(ContractInvokeRequestPayload)
				obj.DeepCopy(payload, msg.Payload)
				request.AddMessage(NewMessage(msg.App, payload))

			} else if msg.App == APP_CONTRACT_STOP_REQUEST {
				payload := new(ContractStopRequestPayload)
				obj.DeepCopy(payload, msg.Payload)
				request.AddMessage(NewMessage(msg.App, payload))
			}
			return request
		} else {
			if msg.App == APP_PAYMENT {
				payload := new(PaymentPayload)
				obj.DeepCopy(payload, msg.Payload)
				request.AddMessage(NewMessage(msg.App, payload))
				// } else if msg.App == APP_CONTRACT_TPL {
				// 	payload := new(ContractTplPayload)
				// 	obj.DeepCopy(payload, msg.Payload)
				// 	request.AddMessage(NewMessage(msg.App, payload))
				// } else if msg.App == APP_CONTRACT_DEPLOY {
				// 	payload := new(ContractDeployPayload)
				// 	obj.DeepCopy(payload, msg.Payload)
				// 	request.AddMessage(NewMessage(msg.App, payload))
				// } else if msg.App == APP_CONTRACT_INVOKE {
				// 	payload := new(ContractInvokePayload)
				// 	obj.DeepCopy(payload, msg.Payload)
				// 	request.AddMessage(NewMessage(msg.App, payload))
				// } else if msg.App == APP_CONTRACT_STOP {
				// 	payload := new(ContractStopPayload)
				// 	obj.DeepCopy(payload, msg.Payload)
				// 	request.AddMessage(NewMessage(msg.App, payload))
				// } else if msg.App == APP_SIGNATURE {
				// 	payload := new(SignaturePayload)
				// 	obj.DeepCopy(payload, msg.Payload)
				// 	request.AddMessage(NewMessage(msg.App, payload))
				//} else if msg.App == APP_CONFIG {
				//	payload := new(ConfigPayload)
				//	obj.DeepCopy(payload, msg.Payload)
				//	request.AddMessage(NewMessage(msg.App, payload))
			} else if msg.App == APP_DATA {
				payload := new(DataPayload)
				obj.DeepCopy(payload, msg.Payload)
				request.AddMessage(NewMessage(msg.App, payload))
			} else if msg.App == APP_ACCOUNT_UPDATE {
				payload := new(AccountStateUpdatePayload)
				obj.DeepCopy(payload, msg.Payload)
				request.AddMessage(NewMessage(msg.App, payload))
			} else {
				log.Error("Invalid tx message")
				return nil
			}
		}
	}
	return request
}

//获取一个被Jury执行完成后，但是还没有进行陪审员签名的交易
func (tx *Transaction) GetResultRawTx() *Transaction {

	txCopy := tx.Clone()
	result := &Transaction{}
	result.CertId = tx.CertId
	isResultMsg := false
	for _, msg := range txCopy.TxMessages {
		if msg.App.IsRequest() {
			isResultMsg = true
		}
		if msg.App == APP_SIGNATURE {
			continue //移除SignaturePayload
		}
		if isResultMsg && msg.App == APP_PAYMENT { //移除ContractPayout中的解锁脚本
			pay := msg.Payload.(*PaymentPayload)
			for _, in := range pay.Inputs {
				in.SignatureScript = nil
			}
		}
		result.TxMessages = append(result.TxMessages, msg)
	}
	return result
}

func (tx *Transaction) GetResultTx() *Transaction {
	txCopy := tx.Clone()
	result := &Transaction{}
	result.CertId = tx.CertId
	for _, msg := range txCopy.TxMessages {
		if msg.App == APP_SIGNATURE {
			continue //移除SignaturePayload
		}
		result.TxMessages = append(result.TxMessages, msg)
	}
	return result
}

//Request 这条Message的Index是多少
func (tx *Transaction) GetRequestMsgIndex() int {
	for idx, msg := range tx.TxMessages {
		if msg.App.IsRequest() {
			return idx
		}
	}
	return -1
}

//这个交易是否包含了从合约付款出去的结果,有则返回该Payment
func (tx *Transaction) HasContractPayoutMsg() (bool, *PaymentPayload) {
	isInvokeResult := false
	for _, msg := range tx.TxMessages {
		if msg.App.IsRequest() {
			isInvokeResult = true
			continue
		}
		if isInvokeResult && msg.App == APP_PAYMENT {
			pay := msg.Payload.(*PaymentPayload)
			if !pay.IsCoinbase() {
				return true, pay
			}
		}
	}
	return false, nil
}

func (tx *Transaction) InvokeContractId() []byte {
	for _, msg := range tx.TxMessages {
		if msg.App == APP_CONTRACT_INVOKE_REQUEST {
			contractId := msg.Payload.(*ContractInvokeRequestPayload).ContractId
			return contractId
		}
	}
	return nil
}

type Addition struct {
	Addr   common.Address `json:"address"`
	Amount uint64         `json:"amount"`
	Asset  *Asset         `json:"asset"`
}

type OutPoint struct {
	TxHash       common.Hash `json:"txhash"`        // reference Utxo struct key field
	MessageIndex uint32      `json:"message_index"` // message index in transaction
	OutIndex     uint32      `json:"out_index"`
}

func (outpoint *OutPoint) String() string {
	return fmt.Sprintf("Outpoint[TxId:{%#x},MsgIdx:{%d},OutIdx:{%d}]", outpoint.TxHash, outpoint.MessageIndex, outpoint.OutIndex)
}

func NewOutPoint(hash common.Hash, messageindex uint32, outindex uint32) *OutPoint {
	return &OutPoint{
		TxHash:       hash,
		MessageIndex: messageindex,
		OutIndex:     outindex,
	}
}

// VarIntSerializeSize returns the number of bytes it would take to serialize
// val as a variable length integer.
func VarIntSerializeSize(val uint64) int {
	// The value is small enough to be represented by itself, so it's
	// just 1 byte.
	if val < 0xfd {
		return 1
	}
	// Discriminant 1 byte plus 2 bytes for the uint16.
	if val <= math.MaxUint16 {
		return 3
	}
	// Discriminant 1 byte plus 4 bytes for the uint32.
	if val <= math.MaxUint32 {
		return 5
	}
	// Discriminant 1 byte plus 8 bytes for the uint64.
	return 9
}

// SerializeSize returns the number of bytes it would take to serialize the
// the transaction output.
func (t *Output) SerializeSize() int {
	// Value 8 bytes + serialized varint size for the length of PkScript +
	// PkScript bytes.
	return 8 + VarIntSerializeSize(uint64(len(t.PkScript))) + len(t.PkScript)
}
func (t *Input) SerializeSize() int {
	// Outpoint Hash 32 bytes + Outpoint Index 4 bytes + Sequence 4 bytes +
	// serialized varint size for the length of SignatureScript +
	// SignatureScript bytes.
	return 40 + VarIntSerializeSize(uint64(len(t.SignatureScript))) +
		len(t.SignatureScript)
}

//func (msg *PaymentPayload) SerializeSize() int {
//	n := msg.baseSize()
//	return n
//}
func (msg *Transaction) SerializeSize() int {
	n := msg.baseSize()
	return n
}

//Deep copy transaction to a new object
func (tx *Transaction) Clone() Transaction {
	newTx := &Transaction{}
	data, _ := rlp.EncodeToBytes(tx)
	rlp.DecodeBytes(data, newTx)
	return *newTx
}

const defaultTxInOutAlloc = 15

// SerializeNoWitness encodes the transaction to w in an identical manner to
// Serialize, however even if the source transaction has inputs with witness
// data, the old serialization format will still be used.
func (msg *PaymentPayload) SerializeNoWitness(w io.Writer) error {
	//return msg.BtcEncode(w, 0, BaseEncoding)
	return nil
}

func (msg *Transaction) baseSize() int {
	b, _ := rlp.EncodeToBytes(msg)
	return len(b)
}
func (tx *Transaction) IsContractTx() bool {
	for _, m := range tx.TxMessages {
		if m.App >= APP_CONTRACT_TPL && m.App <= APP_SIGNATURE {
			return true
		}
	}
	return false
}

func (tx *Transaction) IsSystemContract() bool {
	for _, msg := range tx.TxMessages {
		if msg.App == APP_CONTRACT_INVOKE_REQUEST {
			contractId := msg.Payload.(*ContractInvokeRequestPayload).ContractId
			contractAddr := common.NewAddress(contractId, common.ContractHash)
			log.Debug("isSystemContract", "contract id", contractAddr, "len", len(contractAddr))
			return contractAddr.IsSystemContractAddress() //, nil

		} else if msg.App == APP_CONTRACT_TPL_REQUEST {
			return true //todo  先期将install作为系统合约处理，只有Mediator可以安装，后期在扩展到所有节点
		} else if msg.App >= APP_CONTRACT_DEPLOY_REQUEST {
			return false //, nil
		}
	}
	return true //, errors.New("isSystemContract not find contract type")
}

//判断一个交易是否是一个合约请求交易，并且还没有被执行
func (tx *Transaction) IsNewContractInvokeRequest() bool {
	lastMsg := tx.TxMessages[len(tx.TxMessages)-1]
	return lastMsg.App.IsRequest()

}

//获得合约请求Msg的Index
func (tx *Transaction) GetContractInvokeReqMsgIdx() int {
	for idx, msg := range tx.TxMessages {
		if msg.App == APP_CONTRACT_INVOKE_REQUEST {
			return idx
		}
	}
	return -1
}
func (tx *Transaction) GetTxFeeAllocate(queryUtxoFunc QueryUtxoFunc, unitTime int64, getSignerFunc GetScriptSignersFunc, mediatorAddr common.Address) ([]*Addition, error) {
	fee, err := tx.GetTxFee(queryUtxoFunc, unitTime)
	result := []*Addition{}
	if err != nil {
		return nil, err
	}
	if fee.Amount == 0 {
		return result, nil
	}
	isResultMsg := false
	jury := []common.Address{}
	for msgIdx, msg := range tx.TxMessages {
		if msg.App.IsRequest() {
			isResultMsg = true
			continue
		}
		if isResultMsg && msg.App == APP_SIGNATURE {
			payload := msg.Payload.(*SignaturePayload)
			for _, sig := range payload.Signatures {
				jury = append(jury, crypto.PubkeyBytesToAddress(sig.PubKey))
			}
		}
		if isResultMsg && msg.App == APP_PAYMENT {
			payment := msg.Payload.(*PaymentPayload)
			if !payment.IsCoinbase() {
				jury, err = getSignerFunc(tx, msgIdx, 0)
				if err != nil {
					return nil, errors.New("Parse unlock script to get signers error:" + err.Error())
				}
			}
		}
	}
	if isResultMsg { //合约执行，Fee需要分配给Jury
		juryAmount := float64(fee.Amount) * parameter.CurrentSysParameters.ContractFeeJuryPercent
		juryAllocatedAmt := uint64(0)
		juryCount := float64(len(jury))
		for _, jurior := range jury {
			jIncome := &Addition{
				Addr:   jurior,
				Amount: uint64(juryAmount / juryCount),
				Asset:  fee.Asset,
			}
			juryAllocatedAmt += jIncome.Amount
			result = append(result, jIncome)
		}
		mediatorIncome := &Addition{
			Addr:   mediatorAddr,
			Amount: fee.Amount - juryAllocatedAmt,
			Asset:  fee.Asset,
		}
		result = append(result, mediatorIncome)
	} else { //没有合约执行，全部分配给Mediator
		mediatorIncome := &Addition{
			Addr:   mediatorAddr,
			Amount: fee.Amount,
			Asset:  fee.Asset,
		}
		result = append(result, mediatorIncome)
	}
	return result, nil
}

//判断一个交易是否是完整交易，如果是普通转账交易就是完整交易，
//如果是合约请求交易，那么带了结果Msg的就是完整交易
//func (tx *Transaction) IsFullTx() bool{
//	if
//}

//	// Version 4 bytes + LockTime 4 bytes + Serialized varint size for the
//	// number of transaction inputs and outputs.
//	n := 16 + VarIntSerializeSize(uint64(len(msg.TxMessages))) +
//		VarIntSerializeSize(uint64(len(msg.TxHash)))
//	for _, mtx := range msg.TxMessages {
//		payload := mtx.Payload
//		payment, ok := payload.(PaymentPayload)
//		if ok == true {
//			for _, txIn := range payment.Inputs {
//				n += txIn.SerializeSize()
//			}
//			for _, txOut := range payment.Outputs {
//				n += txOut.SerializeSize()
//			}
//		}
//	}
//	return n
//}

// SerializeSizeStripped returns the number of bytes it would take to serialize
// the transaction, excluding any included witness data.
//func (msg *PaymentPayload) SerializeSizeStripped() int {
//	return msg.baseSize()
//}

// SerializeSizeStripped returns the number of bytes it would take to serialize
// the transaction, excluding any included witness data.
func (tx *Transaction) SerializeSizeStripped() int {
	return tx.baseSize()
}

// WriteVarBytes serializes a variable length byte array to w as a varInt
// containing the number of bytes, followed by the bytes themselves.
func WriteVarBytes(w io.Writer, pver uint32, bytes []byte) error {
	slen := uint64(len(bytes))
	err := WriteVarInt(w, pver, slen)
	if err != nil {
		return err
	}
	_, err = w.Write(bytes)
	return err
}

const binaryFreeListMaxItems = 1024

type binaryFreeList chan []byte

var binarySerializer binaryFreeList = make(chan []byte, binaryFreeListMaxItems)

// WriteVarInt serializes val to w using a variable number of bytes depending
// on its value.
func WriteVarInt(w io.Writer, pver uint32, val uint64) error {
	if val < 0xfd {
		return binarySerializer.PutUint8(w, uint8(val))
	}
	if val <= math.MaxUint16 {
		err := binarySerializer.PutUint8(w, 0xfd)
		if err != nil {
			return err
		}
		return binarySerializer.PutUint16(w, littleEndian, uint16(val))
	}
	if val <= math.MaxUint32 {
		err := binarySerializer.PutUint8(w, 0xfe)
		if err != nil {
			return err
		}
		return binarySerializer.PutUint32(w, littleEndian, uint32(val))
	}
	err := binarySerializer.PutUint8(w, 0xff)
	if err != nil {
		return err
	}
	return binarySerializer.PutUint64(w, littleEndian, val)
}

// Borrow returns a byte slice from the free list with a length of 8.  A new
// buffer is allocated if there are not any available on the free list.
func (l binaryFreeList) Borrow() []byte {
	var buf []byte
	select {
	case buf = <-l:
	default:
		buf = make([]byte, 8)
	}
	return buf[:8]
}

// Return puts the provided byte slice back on the free list.  The buffer MUST
// have been obtained via the Borrow function and therefore have a cap of 8.
func (l binaryFreeList) Return(buf []byte) {
	select {
	case l <- buf:
	default:
		// Let it go to the garbage collector.
	}
}

// Uint8 reads a single byte from the provided reader using a buffer from the
// free list and returns it as a uint8.
func (l binaryFreeList) Uint8(r io.Reader) (uint8, error) {
	buf := l.Borrow()[:1]
	if _, err := io.ReadFull(r, buf); err != nil {
		l.Return(buf)
		return 0, err
	}
	rv := buf[0]
	l.Return(buf)
	return rv, nil
}

// Uint16 reads two bytes from the provided reader using a buffer from the
// free list, converts it to a number using the provided byte order, and returns
// the resulting uint16.
func (l binaryFreeList) Uint16(r io.Reader, byteOrder binary.ByteOrder) (uint16, error) {
	buf := l.Borrow()[:2]
	if _, err := io.ReadFull(r, buf); err != nil {
		l.Return(buf)
		return 0, err
	}
	rv := byteOrder.Uint16(buf)
	l.Return(buf)
	return rv, nil
}

// Uint32 reads four bytes from the provided reader using a buffer from the
// free list, converts it to a number using the provided byte order, and returns
// the resulting uint32.
func (l binaryFreeList) Uint32(r io.Reader, byteOrder binary.ByteOrder) (uint32, error) {
	buf := l.Borrow()[:4]
	if _, err := io.ReadFull(r, buf); err != nil {
		l.Return(buf)
		return 0, err
	}
	rv := byteOrder.Uint32(buf)
	l.Return(buf)
	return rv, nil
}

// Uint64 reads eight bytes from the provided reader using a buffer from the
// free list, converts it to a number using the provided byte order, and returns
// the resulting uint64.
func (l binaryFreeList) Uint64(r io.Reader, byteOrder binary.ByteOrder) (uint64, error) {
	buf := l.Borrow()[:8]
	if _, err := io.ReadFull(r, buf); err != nil {
		l.Return(buf)
		return 0, err
	}
	rv := byteOrder.Uint64(buf)
	l.Return(buf)
	return rv, nil
}

// PutUint8 copies the provided uint8 into a buffer from the free list and
// writes the resulting byte to the given writer.
func (l binaryFreeList) PutUint8(w io.Writer, val uint8) error {
	buf := l.Borrow()[:1]
	buf[0] = val
	_, err := w.Write(buf)
	l.Return(buf)
	return err
}

var (
	// littleEndian is a convenience variable since binary.LittleEndian is
	// quite long.
	littleEndian = binary.LittleEndian
	// bigEndian is a convenience variable since binary.BigEndian is quite
	// long.
	bigEndian = binary.BigEndian
)

// PutUint16 serializes the provided uint16 using the given byte order into a
// buffer from the free list and writes the resulting two bytes to the given
// writer.
func (l binaryFreeList) PutUint16(w io.Writer, byteOrder binary.ByteOrder, val uint16) error {
	buf := l.Borrow()[:2]
	byteOrder.PutUint16(buf, val)
	_, err := w.Write(buf)
	l.Return(buf)
	return err
}

// PutUint32 serializes the provided uint32 using the given byte order into a
// buffer from the free list and writes the resulting four bytes to the given
// writer.
func (l binaryFreeList) PutUint32(w io.Writer, byteOrder binary.ByteOrder, val uint32) error {
	buf := l.Borrow()[:4]
	byteOrder.PutUint32(buf, val)
	_, err := w.Write(buf)
	l.Return(buf)
	return err
}

// PutUint64 serializes the provided uint64 using the given byte order into a
// buffer from the free list and writes the resulting eight bytes to the given
// writer.
func (l binaryFreeList) PutUint64(w io.Writer, byteOrder binary.ByteOrder, val uint64) error {
	buf := l.Borrow()[:8]
	byteOrder.PutUint64(buf, val)
	_, err := w.Write(buf)
	l.Return(buf)
	return err
}
func WriteTxOut(w io.Writer, pver uint32, version int32, to *Output) error {
	err := binarySerializer.PutUint64(w, littleEndian, to.Value)
	if err != nil {
		return err
	}
	return WriteVarBytes(w, pver, to.PkScript)
}

func (a *Addition) IsEqualStyle(b *Addition) (bool, error) {
	if b == nil {
		return false, errors.New("Addition isEqual err, param is nil")
	}
	if a.Addr == b.Addr && a.Asset == b.Asset {
		return true, nil
	}
	return false, nil
}
func (a *Addition) Key() string {
	b := append(a.Addr.Bytes21(), a.Asset.Bytes()...)
	return hex.EncodeToString(b)
}

type SequeueTxPoolTxs []*TxPoolTransaction

// add
func (seqTxs *SequeueTxPoolTxs) Add(newPoolTx *TxPoolTransaction) {
	*seqTxs = append(*seqTxs, newPoolTx)
}

// add priority
func (seqTxs *SequeueTxPoolTxs) AddPriority(newPoolTx *TxPoolTransaction) {
	if len(*seqTxs) == 0 {
		*seqTxs = append(*seqTxs, newPoolTx)
	} else {
		added := false
		for i, item := range *seqTxs {
			if newPoolTx.GetPriorityfloat64() > item.GetPriorityfloat64() {
				*seqTxs = append((*seqTxs)[:i], append(SequeueTxPoolTxs{newPoolTx}, (*seqTxs)[i:]...)...)
				added = true
				break
			}
		}
		if !added {
			*seqTxs = append(*seqTxs, newPoolTx)
		}
	}
}

// get
func (seqTxs *SequeueTxPoolTxs) Get() *TxPoolTransaction {
	if len(*seqTxs) <= 0 {
		return nil
	}
	if len(*seqTxs) == 1 {
		first := (*seqTxs)[0]
		*seqTxs = make([]*TxPoolTransaction, 0)
		return first
	}
	first, rest := (*seqTxs)[0], (*seqTxs)[1:]
	*seqTxs = rest
	return first
}

// get all
func (seqTxs *SequeueTxPoolTxs) All() []*TxPoolTransaction {
	items := (*seqTxs)[:]
	*seqTxs = make([]*TxPoolTransaction, 0)
	return items
}
