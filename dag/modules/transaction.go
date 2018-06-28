package modules

import (
	"errors"
	"fmt"
	"io"
	"math/big"
	"sync/atomic"
	"time"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/crypto"
	"github.com/palletone/go-palletone/common/crypto/sha3"
	"github.com/palletone/go-palletone/common/hexutil"
	"github.com/palletone/go-palletone/common/rlp"
)

type PoolTransaction struct {
	data txdata
	// caches
	hash         atomic.Value
	size         atomic.Value
	memery       atomic.Value
	from         atomic.Value
	excutiontime uint `json:"excution_time"`

	creationdate time.Time `json:"creation_date"`
}
type txdata struct {
	AccountNonce      uint64          `json:"account_nonce"`
	From              *common.Address `json:"from"`
	Price             *big.Int        `json:"price"`              // 交易费
	Recipient         *common.Address `json:"to"       rlp:"nil"` // nil means contract creation
	Amount            *big.Int        `json:"value"`
	Payload           []byte          `json:"input"`
	TranReceiptStatus string          `json:"tran_receipt_status"`

	// Signature values
	V *big.Int `json:"v"`
	R *big.Int `json:"r"`
	S *big.Int `json:"s"`

	// This is only used when marshaling to JSON.
	Hash *common.Hash `json:"hash" rlp:"-"`
}

type txdataMarshaling struct {
	AccountNonce hexutil.Uint64
	Price        *hexutil.Big
	Amount       *hexutil.Big
	Payload      hexutil.Bytes
	V            *hexutil.Big
	R            *hexutil.Big
	S            *hexutil.Big
}

func NewPoolTransaction(from, to common.Address, amount *big.Int, data []byte) *PoolTransaction {
	return newTransaction(&from, &to, amount, data)
}

func NewContractCreation(from common.Address, amount *big.Int, data []byte) *PoolTransaction {
	return newTransaction(&from, nil, amount, data)
}

func newTransaction(from, to *common.Address, amount *big.Int, data []byte) *PoolTransaction {
	if len(data) > 0 {
		data = common.CopyBytes(data)
	}
	d := txdata{
		From:      from,
		Recipient: to,
		Payload:   data,
		Amount:    new(big.Int),
		Price:     new(big.Int),
		V:         new(big.Int),
		R:         new(big.Int),
		S:         new(big.Int),
	}
	if amount != nil {
		d.Amount.Set(amount)
	}

	return &PoolTransaction{data: d}
}

// ChainId returns which chain id this transaction was signed for (if at all)
func (tx *PoolTransaction) ChainId() *big.Int {
	return deriveChainId(tx.data.V)
}

// Protected returns whether the transaction is protected from replay protection.
func (tx *PoolTransaction) Protected() bool {
	return isProtectedV(tx.data.V)
}

func isProtectedV(V *big.Int) bool {
	if V.BitLen() <= 8 {
		v := V.Uint64()
		return v != 27 && v != 28
	}
	// anything not 27 or 28 are considered unprotected
	return true
}

// EncodeRLP implements rlp.Encoder
func (tx *PoolTransaction) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, &tx.data)
}

// DecodeRLP implements rlp.Decoder
func (tx *PoolTransaction) DecodeRLP(s *rlp.Stream) error {
	_, size, _ := s.Kind()
	err := s.Decode(&tx.data)
	if err == nil {
		tx.size.Store(common.StorageSize(rlp.ListSize(size)))
	}

	return err
}

// MarshalJSON encodes the web3 RPC transaction format.
func (tx *PoolTransaction) MarshalJSON() ([]byte, error) {
	hash := tx.Hash()
	data := tx.data
	data.Hash = &hash
	return data.MarshalJSON()
}

// UnmarshalJSON decodes the web3 RPC transaction format.
func (tx *PoolTransaction) UnmarshalJSON(input []byte) error {
	var dec txdata
	if err := dec.UnmarshalJSON(input); err != nil {
		return err
	}
	var V byte
	if isProtectedV(dec.V) {
		chainID := deriveChainId(dec.V).Uint64()
		V = byte(dec.V.Uint64() - 35 - 2*chainID)
	} else {
		V = byte(dec.V.Uint64() - 27)
	}
	if !crypto.ValidateSignatureValues(V, dec.R, dec.S, false) {
		return errors.New("invalid transaction v, r, s values")
	}
	*tx = PoolTransaction{data: dec}
	return nil
}

func (tx *PoolTransaction) Data() []byte { return common.CopyBytes(tx.data.Payload) }

func (tx *PoolTransaction) Nonce() uint64            { return tx.data.AccountNonce }
func (tx *PoolTransaction) Price() *big.Int          { return new(big.Int).Set(tx.data.Price) }
func (tx *PoolTransaction) Value() *big.Int          { return new(big.Int).Set(tx.data.Amount) }
func (tx *PoolTransaction) Account() *common.Address { return tx.data.From }
func (tx *PoolTransaction) CheckNonce() bool         { return true }

// To returns the recipient address of the PoolTransaction .
// It returns nil if the transaction is a contract creation.
func (tx *PoolTransaction) ToAddress() *common.Address {
	if tx.data.Recipient == nil {
		return nil
	}
	to := *tx.data.Recipient
	return &to
}

// Hash hashes the RLP encoding of tx.
// It uniquely identifies the transaction.
func (tx *PoolTransaction) Hash() common.Hash {
	if hash := tx.hash.Load(); hash != nil {
		return hash.(common.Hash)
	}
	v := rlpHash(tx)
	tx.hash.Store(v)
	return v
}

// Size returns the true RLP encoded storage size of the transaction, either by
// encoding and returning it, or returning a previsouly cached value.
func (tx *PoolTransaction) Size() common.StorageSize {
	if size := tx.size.Load(); size != nil {
		return size.(common.StorageSize)
	}
	c := writeCounter(0)
	rlp.Encode(&c, &tx.data)
	tx.size.Store(common.StorageSize(c))
	return common.StorageSize(c)
}

func (tx *PoolTransaction) String() string {
	var from, to string
	if tx.data.V != nil {
		// make a best guess about the signer and use that to derive
		// the sender.
		signer := deriveSigner(tx.data.V)
		if f, err := Sender(signer, tx); err != nil { // derive but don't cache
			from = "[invalid sender: invalid sig]"
		} else {
			from = fmt.Sprintf("%x", f[:])
		}
	} else {
		from = "[invalid sender: nil V field]"
	}

	if tx.data.Recipient == nil {
		to = "[contract creation]"
	} else {
		to = fmt.Sprintf("%x", tx.data.Recipient[:])
	}
	enc, _ := rlp.EncodeToBytes(&tx.data)
	return fmt.Sprintf(`
	TX(%x)
	Contract: %v
	From:     %s
	To:       %s
	Nonce:    %v
	Price: %#x
	Value:    %#x
	Data:     0x%x
	V:        %#x
	R:        %#x
	S:        %#x
	Hex:      %x
`,
		tx.Hash(),
		tx.data.Recipient == nil,
		from,
		to,
		tx.data.From,
		tx.data.Price,
		tx.data.Amount,
		tx.data.Payload,
		tx.data.V,
		tx.data.R,
		tx.data.S,
		enc,
	)
}

func (tx *PoolTransaction) WithSignature(signer Signer, sig []byte) (*PoolTransaction, error) {
	r, s, v, err := signer.SignatureValues(tx, sig)
	if err != nil {
		return nil, err
	}
	cpy := &PoolTransaction{data: tx.data}
	cpy.data.R, cpy.data.S, cpy.data.V = r, s, v
	return cpy, nil
}

// Cost returns amount + price
func (tx *PoolTransaction) Cost() *big.Int {
	total := new(big.Int) // .mul(tx.data.Price, new(big.Int).SetUint64(tx.data.GasLimit))
	total.Add(tx.data.Price, tx.data.Amount)
	return total
}

// AsMessage returns the transaction as a core.Message.
//
// AsMessage requires a signer to derive the sender.
//
// XXX Rename message to something less arbitrary?
func (tx *PoolTransaction) AsMessage(s Signer) (PoolMessage, error) {
	msg := PoolMessage{
		from:       *tx.data.From,
		gasPrice:   new(big.Int).Set(tx.data.Price),
		to:         tx.data.Recipient,
		amount:     tx.data.Amount,
		data:       tx.data.Payload,
		checkNonce: true,
	}

	var err error
	msg.from, err = Sender(s, tx)
	return msg, err
}

// PoolTransactions is a Transaction slice type for basic sorting.
type PoolTransactions []*PoolTransaction

// Len returns the length of s.
func (s PoolTransactions) Len() int { return len(s) }

// Swap swaps the i'th and the j'th element in s.
func (s PoolTransactions) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// GetRlp implements Rlpable and returns the i'th element of s in rlp.
func (s PoolTransactions) GetRlp(i int) []byte {
	enc, _ := rlp.EncodeToBytes(s[i])
	return enc
}

// TxDifference returns a new set t which is the difference between a to b.
func TxDifference(a, b PoolTransactions) (keep PoolTransactions) {
	keep = make(PoolTransactions, 0, len(a))

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

// single account, otherwise a nonce comparison doesn't make much sense.
type TxByNonce PoolTransactions

func (s TxByNonce) Len() int           { return len(s) }
func (s TxByNonce) Less(i, j int) bool { return s[i].data.AccountNonce < s[j].data.AccountNonce }
func (s TxByNonce) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

// TxByPrice implements both the sort and the heap interface, making it useful
// for all at once sorting as well as individually adding and removing elements.
type TxByPrice PoolTransactions

func (s TxByPrice) Len() int           { return len(s) }
func (s TxByPrice) Less(i, j int) bool { return s[i].data.Price.Cmp(s[j].data.Price) > 0 }
func (s TxByPrice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func (s *TxByPrice) Push(x interface{}) {
	*s = append(*s, x.(*PoolTransaction))
}

func (s *TxByPrice) Pop() interface{} {
	old := *s
	n := len(old)
	x := old[n-1]
	*s = old[0 : n-1]
	return x
}

// Message is a fully derived transaction and implements core.Message
//
// NOTE: In a future PR this will be removed.
type PoolMessage struct {
	to         *common.Address
	from       common.Address
	nonce      uint64
	amount     *big.Int
	gasLimit   uint64
	gasPrice   *big.Int
	data       []byte
	checkNonce bool
}

func NewMessage(from, to *common.Address, nonce uint64, amount *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte, checkNonce bool) PoolMessage {
	return PoolMessage{
		from:       *from,
		to:         to,
		nonce:      nonce,
		amount:     amount,
		gasLimit:   gasLimit,
		gasPrice:   gasPrice,
		data:       data,
		checkNonce: checkNonce,
	}
}

func (m PoolMessage) From() *common.Address { return &m.from }
func (m PoolMessage) To() *common.Address   { return m.to }
func (m PoolMessage) GasPrice() *big.Int    { return m.gasPrice }
func (m PoolMessage) Value() *big.Int       { return m.amount }
func (m PoolMessage) Gas() uint64           { return m.gasLimit }
func (m PoolMessage) Nonce() uint64         { return m.nonce }
func (m PoolMessage) Data() []byte          { return m.data }
func (m PoolMessage) CheckNonce() bool      { return m.checkNonce }

// deriveChainId derives the chain id from the given v parameter
func deriveChainId(v *big.Int) *big.Int {
	if v.BitLen() <= 64 {
		v := v.Uint64()
		if v == 27 || v == 28 {
			return new(big.Int)
		}
		return new(big.Int).SetUint64((v - 35) / 2)
	}
	v = new(big.Int).Sub(v, big.NewInt(35))
	return v.Div(v, big.NewInt(2))
}
func rlpHash(x interface{}) (h common.Hash) {
	hw := sha3.NewKeccak256()
	rlp.Encode(hw, x)
	hw.Sum(h[:0])
	return h
}

// deriveSigner makes a *best* guess about which signer to use.
func deriveSigner(V *big.Int) Signer {
	if V.Sign() != 0 && isProtectedV(V) {
		return NewEIP155Signer(deriveChainId(V))
	} else {
		return HomesteadSigner{}
	}
}

type writeCounter common.StorageSize

func (c *writeCounter) Write(b []byte) (int, error) {
	*c += writeCounter(len(b))
	return len(b), nil
}
