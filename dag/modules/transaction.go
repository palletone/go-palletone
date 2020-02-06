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
	"sync/atomic"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/crypto"
	"github.com/palletone/go-palletone/common/obj"
	"github.com/palletone/go-palletone/common/util"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/dag/constants"
	"github.com/palletone/go-palletone/dag/errors"
	"github.com/palletone/go-palletone/dag/parameter"
)

var (
	//TXFEE       = big.NewInt(100000000) // transaction fee =1ptn
	TX_MAXSIZE  = 256 * 1024 //256kb
	TX_BASESIZE = 100 * 1024 //100kb
)

//一个交易的状态
type TxStatus byte

const (
	TxStatus_NotFound TxStatus = iota //找不到该交易
	TxStatus_InPool                   //未打包
	TxStatus_Unstable                 //已打包未稳定
	TxStatus_Stable                   //已打包，已稳定
)

func (s TxStatus) String() string {
	switch s {
	case TxStatus_NotFound:
		return "NotFound"
	case TxStatus_InPool:
		return "InPool"
	case TxStatus_Unstable:
		return "Unstable"
	case TxStatus_Stable:
		return "Stable"
	}
	return "Unknown"
}

//var DepositContractLockScript = common.Hex2Bytes("140000000000000000000000000000000000000001c8")

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

//func NewContractCreation(msg []*Message) *Transaction {
//	return newTransaction(msg)
//}

func newTransaction(msg []*Message) *Transaction {
	tx := transaction_sdw{}
	if len(msg) > 0 {
		tx.TxMessages = make([]*Message, len(msg))
		copy(tx.TxMessages, msg)
	}
	return &Transaction{txdata: tx}
}

// AddTxIn adds a transaction input to the message.
func (tx *Transaction) AddMessage(msg *Message) {
	msgs := tx.Messages()
	if msg != nil {
		msgs = append(msgs, CopyMessage(msg))
	}
	tx.SetMessages(msgs)
}

type TransactionWithUnitInfo struct {
	*Transaction
	UnitHash  common.Hash
	UnitIndex uint64
	Timestamp uint64
	TxIndex   uint64
}

// Hash hashes the RLP encoding of tx.
// It uniquely identifies the transaction.
func (tx *Transaction) Hash() common.Hash {
	if hash := tx.hash.Load(); hash != nil {
		return hash.(common.Hash)
	}
	oldFlag := tx.Illegal()
	if oldFlag {
		tx.txdata.Illegal = false
		v := util.RlpHash(tx)
		tx.hash.Store(v)
		tx.txdata.Illegal = true
		return v
	}
	v := util.RlpHash(tx)
	tx.hash.Store(v)
	return v
}

func (tx *Transaction) RequestHash() common.Hash {
	d := transaction_sdw{}
	for _, msg := range tx.TxMessages() {
		d.TxMessages = append(d.TxMessages, msg)
		if msg.App >= APP_CONTRACT_TPL_REQUEST { //100以上的APPCode是请求
			break
		}
	}
	return util.RlpHash(&Transaction{txdata: d})
}

func (tx *Transaction) GetContractId() []byte {
	for _, msg := range tx.txdata.TxMessages {
		switch msg.App {
		case APP_CONTRACT_DEPLOY_REQUEST:
			addr := crypto.RequestIdToContractAddress(tx.RequestHash())
			return addr.Bytes()
		case APP_CONTRACT_INVOKE_REQUEST:
			payload := msg.Payload.(*ContractInvokeRequestPayload)
			return common.CopyBytes(payload.ContractId)
		case APP_CONTRACT_STOP_REQUEST:
			payload := msg.Payload.(*ContractStopRequestPayload)
			return common.CopyBytes(payload.ContractId)
		}
	}
	return nil
}

//浅拷贝
func (tx *Transaction) Messages() []*Message {
	msgs := make([]*Message, len(tx.txdata.TxMessages))
	copy(msgs, tx.txdata.TxMessages)
	return msgs
}

// 深拷贝
func (tx *Transaction) TxMessages() []*Message {
	temp_msgs := make([]*Message, 0)
	for _, msg := range tx.txdata.TxMessages {
		temp_msgs = append(temp_msgs, CopyMessage(msg))
	}

	return temp_msgs
}

// Size returns the true RLP encoded storage UnitSize of the transaction, either by
// encoding and returning it, or returning a previsouly cached value.
func (tx *Transaction) Size() common.StorageSize {
	if size := tx.size.Load(); size != nil {
		return size.(common.StorageSize)
	}
	size := CalcDateSize(tx)
	tx.size.Store(size)
	return size
}

func (tx *Transaction) Asset() *Asset {
	if tx == nil {
		return nil
	}
	asset := new(Asset)
	msg := tx.txdata.TxMessages[0]
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
	txdata transaction_sdw
	hash   atomic.Value
	size   atomic.Value
}
type transaction_sdw struct {
	TxMessages []*Message `json:"messages"`
	CertId     []byte     `json:"cert_id"` // should be big.Int byte
	Illegal    bool       `json:"Illegal"` // not hash, 1:no valid, 0:ok
}
type QueryUtxoFunc func(outpoint *OutPoint) (*Utxo, error)
type GetAddressFromScriptFunc func(lockScript []byte) (common.Address, error)
type GetScriptSignersFunc func(tx *Transaction, msgIdx, inputIndex int) ([]common.Address, error)
type QueryStateByVersionFunc func(id []byte, field string, version *StateVersion) ([]byte, error)
type GetJurorRewardAddFunc func(jurorAdd common.Address) common.Address

//计算该交易的手续费，基于UTXO，所以传入查询UTXO的函数指针
func (tx *Transaction) GetTxFee(queryUtxoFunc QueryUtxoFunc) (*AmountAsset, error) {
	msg0 := tx.txdata.TxMessages[0]
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

func (tx *Transaction) CertId() []byte { return common.CopyBytes(tx.txdata.CertId) }
func (tx *Transaction) Illegal() bool  { return tx.txdata.Illegal }
func (tx *Transaction) SetMessages(msgs []*Message) {
	if len(msgs) > 0 {
		d := transaction_sdw{}
		d.TxMessages = make([]*Message, len(msgs))
		copy(d.TxMessages, msgs)
		d.CertId = tx.CertId()
		d.Illegal = tx.Illegal()
		temp := &Transaction{txdata: d}
		tx.txdata = d
		tx.hash.Store(temp.Hash())
		tx.size.Store(temp.Size())
	}
}
func (tx *Transaction) SetCertId(certid []byte) {
	d := transaction_sdw{}
	d.CertId = common.CopyBytes(certid)
	d.Illegal = tx.Illegal()
	d.TxMessages = append(d.TxMessages, tx.txdata.TxMessages...)
	temp := &Transaction{txdata: d}
	tx.txdata = d
	tx.hash.Store(temp.Hash())
	tx.size.Store(temp.Size())
}
func (tx *Transaction) SetIllegal(illegal bool) {
	d := transaction_sdw{}
	d.Illegal = illegal
	d.CertId = common.CopyBytes(tx.CertId())
	d.TxMessages = append(d.TxMessages, tx.txdata.TxMessages...)
	temp := &Transaction{txdata: d}
	tx.txdata = d
	tx.hash.Store(temp.Hash())
	tx.size.Store(temp.Size())
}
func (tx *Transaction) ModifiedMsg(index int, msg *Message) {
	if len(tx.Messages()) < index {
		return
	}
	sdw := transaction_sdw{}
	for i, m := range tx.Messages() {
		if i == index {
			sdw.TxMessages = append(sdw.TxMessages, msg)
		} else {
			sdw.TxMessages = append(sdw.TxMessages, m)
		}
	}
	sdw.Illegal = tx.Illegal()
	sdw.CertId = tx.CertId()
	temp := &Transaction{txdata: sdw}
	tx.txdata = sdw
	tx.hash.Store(temp.Hash())
	tx.size.Store(temp.Size())
}

func (tx *Transaction) GetCoinbaseReward(versionFunc QueryStateByVersionFunc,
	scriptFunc GetAddressFromScriptFunc) (*AmountAsset, error) {
	writeMap := make(map[string][]AmountAsset)
	readMap := make(map[string][]AmountAsset)
	msgs := tx.TxMessages()
	if len(msgs) == 2 && msgs[0].App == APP_PAYMENT &&
		msgs[1].App == APP_CONTRACT_INVOKE { //进行了汇总付款
		invoke := msgs[1].Payload.(*ContractInvokePayload)
		for _, read := range invoke.ReadSet {
			readResult, err := versionFunc(read.ContractId, read.Key, read.Version)
			if err != nil {
				return nil, err
			}
			var aa []AmountAsset
			err = rlp.DecodeBytes(readResult, &aa)
			if err != nil {
				return nil, err
			}
			addr := read.Key[len(constants.RewardAddressPrefix):]
			readMap[addr] = aa
		}
		payment := msgs[0].Payload.(*PaymentPayload)
		for _, out := range payment.Outputs {
			aa := AmountAsset{
				Amount: out.Value,
				Asset:  out.Asset,
			}
			addr, _ := scriptFunc(out.PkScript)
			writeMap[addr.String()] = []AmountAsset{aa}
		}
	} else if msgs[0].App == APP_CONTRACT_INVOKE { //进行了记账
		invoke := msgs[0].Payload.(*ContractInvokePayload)
		for _, write := range invoke.WriteSet {
			var aa []AmountAsset
			err := rlp.DecodeBytes(write.Value, &aa)
			if err != nil {
				return nil, err
			}
			addr := write.Key[len(constants.RewardAddressPrefix):]
			writeMap[addr] = aa
		}

		for _, read := range invoke.ReadSet {
			readResult, err := versionFunc(read.ContractId, read.Key, read.Version)
			if err != nil {
				return nil, err
			}
			var aa []AmountAsset
			err = rlp.DecodeBytes(readResult, &aa)
			if err != nil {
				return nil, err
			}
			addr := read.Key[len(constants.RewardAddressPrefix):]
			readMap[addr] = aa
		}
	} else {
		return &AmountAsset{Asset: NewPTNAsset()}, nil
	}

	//计算Write Map和Read Map的差，获得Reward值
	reward := &AmountAsset{}
	for writeAddr, writeAA := range writeMap {
		reward.Asset = writeAA[0].Asset
		if readAA, ok := readMap[writeAddr]; ok {
			readAmt := uint64(0)
			if len(readAA) != 0 { //上一次没有清空
				readAmt = readAA[0].Amount
			}
			reward.Amount += writeAA[0].Amount - readAmt
		} else {
			reward.Amount += writeAA[0].Amount
		}
	}
	return reward, nil
}

//该Tx如果保存后，会产生的新的Utxo
func (tx *Transaction) GetNewUtxos() map[OutPoint]*Utxo {
	result := map[OutPoint]*Utxo{}
	txHash := tx.Hash()
	for msgIndex, msg := range tx.txdata.TxMessages {
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

//获取一个交易中花费了哪些OutPoint
func (tx *Transaction) GetSpendOutpoints() []*OutPoint {
	result := []*OutPoint{}
	for _, msg := range tx.txdata.TxMessages {
		if msg.App != APP_PAYMENT {
			continue
		}
		pay := msg.Payload.(*PaymentPayload)
		inputs := pay.Inputs
		for _, input := range inputs {
			if input.PreviousOutPoint != nil {
				if input.PreviousOutPoint.TxHash.IsSelfHash() { //合约Payback的情形
					op := NewOutPoint(tx.Hash(), input.PreviousOutPoint.MessageIndex, input.PreviousOutPoint.OutIndex)
					result = append(result, op)
				} else {
					result = append(result, input.PreviousOutPoint)
				}
			}
		}
	}
	return result
}

//获得合约交易的签名对应的陪审员地址
func (tx *Transaction) GetContractTxSignatureAddress() []common.Address {
	if !tx.IsContractTx() {
		return nil
	}
	addrs := make([]common.Address, 0)
	for _, msg := range tx.txdata.TxMessages {
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
	msgs := tx.TxMessages()
	request := transaction_sdw{}
	for _, msg := range msgs {
		request.TxMessages = append(request.TxMessages, msg)
		if msg.App.IsRequest() {

			break
		}

	}
	request.CertId = tx.CertId()
	return &Transaction{txdata: request}

}

//获取一个被Jury执行完成后，但是还没有进行陪审员签名的交易
func (tx *Transaction) GetResultRawTx() *Transaction {

	sdw := transaction_sdw{}
	isResultMsg := false
	for _, msg := range tx.TxMessages() {
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
		sdw.TxMessages = append(sdw.TxMessages, msg)
	}
	sdw.CertId = tx.CertId()
	sdw.Illegal = tx.Illegal()
	return &Transaction{txdata: sdw}
}

func (tx *Transaction) GetResultTx() *Transaction {
	sdw := transaction_sdw{}
	for _, msg := range tx.TxMessages() {
		if msg.App == APP_SIGNATURE {
			continue //移除SignaturePayload
		}
		sdw.TxMessages = append(sdw.TxMessages, msg)
	}
	sdw.CertId = tx.CertId()
	return &Transaction{txdata: sdw}
}

//Request 这条Message的Index是多少
func (tx *Transaction) GetRequestMsgIndex() int {
	for idx, msg := range tx.txdata.TxMessages {
		if msg.App.IsRequest() {
			return idx
		}
	}
	return -1
}

//这个交易是否包含了从合约付款出去的结果,有则返回该Payment
func (tx *Transaction) HasContractPayoutMsg() (bool, int, *Message) {
	isInvokeResult := false
	for i, msg := range tx.txdata.TxMessages {
		if msg.App.IsRequest() {
			isInvokeResult = true
			continue
		}
		if isInvokeResult && msg.App == APP_PAYMENT {
			pay := msg.Payload.(*PaymentPayload)
			if !pay.IsCoinbase() {
				return true, i, msg
			}
		}
	}
	return false, 0, nil
}

//获取该交易的所有From地址
func (tx *Transaction) GetFromAddrs(queryUtxoFunc QueryUtxoFunc, getAddrFunc GetAddressFromScriptFunc) (
	[]common.Address, error) {
	addrMap := map[common.Address]bool{}
	for _, msg := range tx.txdata.TxMessages {
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
	msg0 := tx.txdata.TxMessages[0]
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

func (tx *Transaction) GetContractTxType() (MessageType, error) {
	for _, msg := range tx.Messages() {
		if msg.App >= APP_CONTRACT_TPL_REQUEST && msg.App <= APP_CONTRACT_STOP_REQUEST {
			return msg.App, nil
		}
	}
	return APP_UNKNOW, fmt.Errorf("GetContractTxType, not contract Tx, txHash[%s]", tx.Hash().String())
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
func (tx *Transaction) DataPayloadSize() int {
	size := 0
	for _, msg := range tx.txdata.TxMessages {
		if msg.App == APP_DATA {
			data := msg.Payload.(*DataPayload)
			size += len(data.MainData) + len(data.ExtraData) + len(data.Reference)
		}
	}
	return size
}

//Deep copy transaction to a new object
func (tx *Transaction) Clone() *Transaction {
	newTx := new(Transaction)
	data, _ := rlp.EncodeToBytes(tx)
	rlp.DecodeBytes(data, newTx)

	return newTx
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
	for _, m := range tx.txdata.TxMessages {
		if m.App >= APP_CONTRACT_TPL && m.App <= APP_SIGNATURE {
			return true
		}
	}
	return false
}

func (tx *Transaction) IsSystemContract() bool {
	for _, msg := range tx.txdata.TxMessages {
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
	lastMsg := tx.txdata.TxMessages[len(tx.txdata.TxMessages)-1]
	return lastMsg.App.IsRequest()

}

//获得合约请求Msg的Index
func (tx *Transaction) GetContractInvokeReqMsgIdx() int {
	for idx, msg := range tx.txdata.TxMessages {
		if msg.App == APP_CONTRACT_INVOKE_REQUEST {
			return idx
		}
	}
	return -1
}

//之前的费用分配有Bug，在ContractInstall的时候会分配错误。在V2中解决了这个问题，但是由于测试网已经有历史数据了，所以需要保留历史计算方法。
func (tx *Transaction) GetTxFeeAllocateLegacyV1(queryUtxoFunc QueryUtxoFunc, getSignerFunc GetScriptSignersFunc,
	mediatorReward common.Address) ([]*Addition, error) {
	fee, err := tx.GetTxFee(queryUtxoFunc)
	result := make([]*Addition, 0)
	if err != nil {
		return nil, err
	}
	if fee.Amount == 0 {
		return result, nil
	}

	isResultMsg := false
	jury := []common.Address{}
	for msgIdx, msg := range tx.TxMessages() {
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

	juryAllocatedAmt := uint64(0)
	if isResultMsg { //合约执行，Fee需要分配给Jury
		juryAmount := float64(fee.Amount) * parameter.CurrentSysParameters.ContractFeeJuryPercent
		juryCount := float64(len(jury))
		for _, juror := range jury {
			jIncome := &Addition{
				Addr:   juror,
				Amount: uint64(juryAmount / juryCount),
				Asset:  fee.Asset,
			}
			juryAllocatedAmt += jIncome.Amount
			result = append(result, jIncome)
		}
		//	mediatorIncome := &Addition{
		//		Addr:   mediatorAddr,
		//		Amount: fee.Amount - juryAllocatedAmt,
		//		Asset:  fee.Asset,
		//	}
		//	result = append(result, mediatorIncome)
		//} else { //没有合约执行，全部分配给Mediator
		//	mediatorIncome := &Addition{
		//		Addr:   mediatorAddr,
		//		Amount: fee.Amount,
		//		Asset:  fee.Asset,
		//	}
		//	result = append(result, mediatorIncome)
	}

	mediatorIncome := &Addition{
		Addr:   mediatorReward,
		Amount: fee.Amount - juryAllocatedAmt,
		Asset:  fee.Asset,
	}
	result = append(result, mediatorIncome)

	return result, nil
}

//获得一笔交易的手续费分配情况,包括Mediator的打包费，Juror的合约执行费
func (tx *Transaction) GetTxFeeAllocate(queryUtxoFunc QueryUtxoFunc, getSignerFunc GetScriptSignersFunc,
	mediatorReward common.Address, getJurorRewardFunc GetJurorRewardAddFunc) ([]*Addition, error) {
	fee, err := tx.GetTxFee(queryUtxoFunc)
	result := make([]*Addition, 0)
	if err != nil {
		return nil, err
	}
	if fee.Amount == 0 {
		return result, nil
	}

	isJuryInside := false
	jury := []common.Address{}
	for msgIdx, msg := range tx.TxMessages() {
		if msg.App == APP_CONTRACT_INVOKE_REQUEST ||
			msg.App == APP_CONTRACT_DEPLOY_REQUEST ||
			msg.App == APP_CONTRACT_STOP_REQUEST {
			isJuryInside = true
			//只有合约部署和调用的时候会涉及到Jury，才会分手续费给Jury
			continue
		}
		if isJuryInside && msg.App == APP_SIGNATURE {
			payload := msg.Payload.(*SignaturePayload)
			for _, sig := range payload.Signatures {
				jury = append(jury, crypto.PubkeyBytesToAddress(sig.PubKey))
			}
		}
		if isJuryInside && msg.App == APP_PAYMENT {
			payment := msg.Payload.(*PaymentPayload)
			if !payment.IsCoinbase() {
				jury, err = getSignerFunc(tx, msgIdx, 0)
				if err != nil {
					return nil, errors.New("Parse unlock script to get signers error:" + err.Error())
				}
			}
		}
	}

	juryAllocatedAmt := uint64(0)
	if isJuryInside { //合约执行，Fee需要分配给Jury
		juryAmount := float64(fee.Amount) * parameter.CurrentSysParameters.ContractFeeJuryPercent
		juryCount := float64(len(jury))
		for _, juror := range jury {
			jIncome := &Addition{
				Addr:   getJurorRewardFunc(juror),
				Amount: uint64(juryAmount / juryCount),
				Asset:  fee.Asset,
			}
			juryAllocatedAmt += jIncome.Amount
			result = append(result, jIncome)
		}
		//	mediatorIncome := &Addition{
		//		Addr:   mediatorAddr,
		//		Amount: fee.Amount - juryAllocatedAmt,
		//		Asset:  fee.Asset,
		//	}
		//	result = append(result, mediatorIncome)
		//} else { //没有合约部署或者执行，全部分配给Mediator
		//	mediatorIncome := &Addition{
		//		Addr:   mediatorAddr,
		//		Amount: fee.Amount,
		//		Asset:  fee.Asset,
		//	}
		//	result = append(result, mediatorIncome)
	}

	mediatorIncome := &Addition{
		Addr:   mediatorReward,
		Amount: fee.Amount - juryAllocatedAmt,
		Asset:  fee.Asset,
	}
	result = append(result, mediatorIncome)

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
	if a.Asset != nil {
		return hex.EncodeToString(append(a.Addr.Bytes21(), a.Asset.Bytes()...))
	}
	return hex.EncodeToString(a.Addr.Bytes21())
}
