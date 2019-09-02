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
	"encoding/hex"
	"fmt"
	"io"
	"math"
	"math/big"
	"strconv"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common"
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
	//for _, m := range msg {
	//	tx.TxMessages = append(tx.TxMessages, m)
	//}
	tx.TxMessages = append(tx.TxMessages, msg...)
	return tx
}

// AddTxIn adds a transaction input to the message.
func (tx *Transaction) AddMessage(msg *Message) {
	tx.TxMessages = append(tx.TxMessages, msg)
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
	Discarded    bool        // will remove
	TxFee        []*Addition `json:"tx_fee"`
	Index        uint64      `json:"index"` // index 是该Unit位置。
	Extra        []byte
	Tag          uint64
	Expiration   time.Time
	//该Tx依赖于哪些TxId作为先决条件
	DependOnTxs []common.Hash
}

func (tx *TxPoolTransaction) Less(otherTx interface{}) bool {
	ap, _ := strconv.ParseFloat(tx.Priority_lvl, 64)
	bp, _ := strconv.ParseFloat(otherTx.(*TxPoolTransaction).Priority_lvl, 64)
	return ap < bp
}

func (tx *TxPoolTransaction) GetPriorityLvl() string {
	if tx.Priority_lvl != "" && tx.Priority_lvl > "0" {
		return tx.Priority_lvl
	}
	var priority_lvl float64
	if txfee := tx.GetTxFee(); txfee.Int64() > 0 {
		if tx.CreationDate.Unix() <= 0 {
			tx.CreationDate = time.Now()
		}
		priority_lvl, _ = strconv.ParseFloat(fmt.Sprintf("%f", float64(txfee.Int64())/
			tx.Tx.Size().Float64()*(1+float64(time.Now().Second()-tx.CreationDate.Second())/(24*3600))), 64)
	}
	tx.Priority_lvl = strconv.FormatFloat(priority_lvl, 'f', -1, 64)
	return tx.Priority_lvl
}
func (tx *TxPoolTransaction) GetPriorityfloat64() float64 {
	level, _ := strconv.ParseFloat(tx.Priority_lvl, 64)
	if level > 0 {
		return level
	}
	var priority_lvl float64
	if txfee := tx.GetTxFee(); txfee.Int64() > 0 {
		if tx.CreationDate.Unix() <= 0 {
			tx.CreationDate = time.Now()
		}
		priority_lvl, _ = strconv.ParseFloat(fmt.Sprintf("%f", float64(txfee.Int64())/
			tx.Tx.Size().Float64()*(1+float64(time.Now().Second()-tx.CreationDate.Second())/(24*3600))), 64)
	}
	return priority_lvl
}
func (tx *TxPoolTransaction) SetPriorityLvl(priority float64) {
	tx.Priority_lvl = strconv.FormatFloat(priority, 'f', -1, 64)
}
func (tx *TxPoolTransaction) GetTxFee() *big.Int {
	var fee uint64
	if tx.TxFee != nil {
		for _, ad := range tx.TxFee {
			fee += ad.Amount
		}
	} else {
		fee = 20 // 20dao
	}
	return big.NewInt(int64(fee))
}

// Hash hashes the RLP encoding of tx.
// It uniquely identifies the transaction.
func (tx *Transaction) Hash() common.Hash {
	oldFlag := tx.Illegal
	tx.Illegal = false

	v := util.RlpHash(tx)
	tx.Illegal = oldFlag

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
type GetAddressFromScriptFunc func(lockScript []byte) (common.Address, error)
type GetScriptSignersFunc func(tx *Transaction, msgIdx, inputIndex int) ([]common.Address, error)

//计算该交易的手续费，基于UTXO，所以传入查询UTXO的函数指针
func (tx *Transaction) GetTxFee(queryUtxoFunc QueryUtxoFunc) (*AmountAsset, error) {
	msg0 := tx.TxMessages[0]
	if msg0.App != APP_PAYMENT {
		return nil, errors.New("Tx message 0 must a payment payload")
	}
	payload := msg0.Payload.(*PaymentPayload)

	if payload.IsCoinbase() {
		return NewAmountAsset(0, NewPTNAsset()), nil
	}
	inAmount := uint64(0)
	outAmount := uint64(0)
	var feeAsset *Asset
	for _, txin := range payload.Inputs {
		utxo, err := queryUtxoFunc(txin.PreviousOutPoint)
		if err != nil {
			return nil, fmt.Errorf("Txin(txhash=%s, msgindex=%v, outindex=%v)'s utxo is empty:%s",
				txin.PreviousOutPoint.TxHash.String(),
				txin.PreviousOutPoint.MessageIndex,
				txin.PreviousOutPoint.OutIndex,
				err.Error())
		}
		feeAsset = utxo.Asset
		// check overflow
		if inAmount+utxo.Amount > (1<<64 - 1) {
			return nil, fmt.Errorf("Compute fees: txin total overflow")
		}
		inAmount += utxo.Amount

		//if unitTime > 0 {
		//	//计算币龄利息
		//	rate := parameter.CurrentSysParameters.TxCoinDayInterest
		//	if bytes.Equal(utxo.PkScript, DepositContractLockScript) {
		//		rate = parameter.CurrentSysParameters.DepositContractInterest
		//	}
		//
		//	interest := award.GetCoinDayInterest(utxo.GetTimestamp(), unitTime, utxo.Amount, rate)
		//	if interest > 0 {
		//		//	log.Infof("Calculate tx fee,Add interest value:%d to tx[%s] fee", interest, tx.Hash().String())
		//		inAmount += interest
		//	}
		//}

	}

	for _, txout := range payload.Outputs {
		// check overflow
		if outAmount+txout.Value > (1<<64 - 1) {
			return nil, fmt.Errorf("Compute fees: txout total overflow")
		}
		outAmount += txout.Value
	}
	if inAmount < outAmount {

		return nil, fmt.Errorf("Compute fees: tx %s txin amount less than txout amount. amount:%d ,outAmount:%d ",
			tx.Hash().String(), inAmount, outAmount)
	}
	fees := inAmount - outAmount

	return &AmountAsset{Amount: fees, Asset: feeAsset}, nil

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
func (tx *Transaction) GetSpendOutpoints() []*OutPoint {
	result := []*OutPoint{}
	for _, msg := range tx.TxMessages {
		if msg.App != APP_PAYMENT {
			continue
		}
		pay := msg.Payload.(*PaymentPayload)
		inputs := pay.Inputs
		for _, input := range inputs {
			if input.PreviousOutPoint != nil {
				result = append(result, input.PreviousOutPoint)
			}
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
	//result.CertId = tx.CertId
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
	result.CertId = txCopy.CertId
	result.Illegal = txCopy.Illegal
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

//获取该交易的所有From地址
func (tx *Transaction) GetFromAddrs(queryUtxoFunc QueryUtxoFunc, getAddrFunc GetAddressFromScriptFunc) (
	[]common.Address, error) {
	addrMap := map[common.Address]bool{}
	for _, msg := range tx.TxMessages {
		if msg.App == APP_PAYMENT {
			pay := msg.Payload.(*PaymentPayload)
			for _, input := range pay.Inputs {
				if input.PreviousOutPoint != nil {
					utxo, err := queryUtxoFunc(input.PreviousOutPoint)
					if err != nil {
						return nil, errors.New("Get utxo by " + input.PreviousOutPoint.String() + " error:" + err.Error())
					}
					addr, _ := getAddrFunc(utxo.PkScript)
					addrMap[addr] = true
				}
			}
		}
	}
	result := []common.Address{}
	for k := range addrMap {
		result = append(result, k)
	}
	return result, nil
}

//获取该交易的发起人地址
func (tx *Transaction) GetRequesterAddr(queryUtxoFunc QueryUtxoFunc, getAddrFunc GetAddressFromScriptFunc) (
	common.Address, error) {
	msg0 := tx.TxMessages[0]
	if msg0.App != APP_PAYMENT {
		return common.Address{}, errors.New("Coinbase or Invalid Tx, first message must be a payment")
	}
	pay := msg0.Payload.(*PaymentPayload)

	utxo, err := queryUtxoFunc(pay.Inputs[0].PreviousOutPoint)
	if err != nil {
		return common.Address{}, err
	}
	return getAddrFunc(utxo.PkScript)

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
	return fmt.Sprintf("Outpoint[TxId:{%#x},MsgIdx:{%d},OutIdx:{%d}]",
		outpoint.TxHash, outpoint.MessageIndex, outpoint.OutIndex)
}
func (outpoint *OutPoint) Clone() *OutPoint {
	return NewOutPoint(outpoint.TxHash, outpoint.MessageIndex, outpoint.OutIndex)
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
			//log.Debug("isSystemContract", "contract id", contractAddr, "len", len(contractAddr))
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
func (tx *Transaction) GetTxFeeAllocate(queryUtxoFunc QueryUtxoFunc, getSignerFunc GetScriptSignersFunc,
	mediatorAddr common.Address) ([]*Addition, error) {
	fee, err := tx.GetTxFee(queryUtxoFunc)
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

// SerializeSizeStripped returns the number of bytes it would take to serialize
// the transaction, excluding any included witness data.
func (tx *Transaction) SerializeSizeStripped() int {
	return tx.baseSize()
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

type SequeueTxPoolTxs struct {
	seqtxs []*TxPoolTransaction
	mu     sync.RWMutex
}

// add
func (seqTxs *SequeueTxPoolTxs) Len() int {
	seqTxs.mu.RLock()
	defer seqTxs.mu.RUnlock()
	return len((*seqTxs).seqtxs)
}
func (seqTxs *SequeueTxPoolTxs) Add(newPoolTx *TxPoolTransaction) {
	seqTxs.mu.Lock()
	defer seqTxs.mu.Unlock()
	(*seqTxs).seqtxs = append((*seqTxs).seqtxs, newPoolTx)
}

// add priority
func (seqTxs *SequeueTxPoolTxs) AddPriority(newPoolTx *TxPoolTransaction) {
	seqTxs.mu.Lock()
	defer seqTxs.mu.Unlock()
	if seqTxs.Len() == 0 {
		(*seqTxs).seqtxs = append((*seqTxs).seqtxs, newPoolTx)
	} else {
		added := false
		for i, item := range (*seqTxs).seqtxs {
			if newPoolTx.GetPriorityfloat64() > item.GetPriorityfloat64() {
				(*seqTxs).seqtxs = append((*seqTxs).seqtxs[:i], append([]*TxPoolTransaction{newPoolTx}, (*seqTxs).seqtxs[i:]...)...)
				added = true
				break
			}
		}
		if !added {
			(*seqTxs).seqtxs = append((*seqTxs).seqtxs, newPoolTx)
		}
	}
}

// get
func (seqTxs *SequeueTxPoolTxs) Get() *TxPoolTransaction {
	seqTxs.mu.Lock()
	defer seqTxs.mu.Unlock()
	if seqTxs.Len() <= 0 {
		return nil
	}
	if seqTxs.Len() == 1 {
		first := (*seqTxs).seqtxs[0]
		(*seqTxs).seqtxs = make([]*TxPoolTransaction, 0)
		return first
	}
	first, rest := (*seqTxs).seqtxs[0], (*seqTxs).seqtxs[1:]
	(*seqTxs).seqtxs = rest
	return first
}

// get all
func (seqTxs *SequeueTxPoolTxs) All() []*TxPoolTransaction {
	seqTxs.mu.Lock()
	defer seqTxs.mu.Unlock()
	items := (*seqTxs).seqtxs[:]
	(*seqTxs).seqtxs = make([]*TxPoolTransaction, 0)
	return items
}
