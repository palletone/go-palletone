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
	"fmt"
	"strconv"
	"time"
	//	"errors"
	//	"fmt"
	//	"io"
	"math/big"
	//	"sync/atomic"
	//	"time"
	//
	"github.com/palletone/go-palletone/common"
	//	"github.com/palletone/go-palletone/common/crypto"
	//	"github.com/palletone/go-palletone/common/crypto/sha3"
	//	"github.com/palletone/go-palletone/common/hexutil"
	//"github.com/Re-volution/sizestruct"
	"github.com/palletone/go-palletone/common/rlp"
)

var (
	TXFEE = big.NewInt(5) // transaction fee =5ptn
)

func NewTransaction(nonce uint64, fee *big.Int, sig []byte) *Transaction {
	return newTransaction(nonce, fee, sig)
}

func NewContractCreation(nonce uint64, fee *big.Int, sig []byte) *Transaction {
	return newTransaction(nonce, fee, sig)
}

func newTransaction(nonce uint64, fee *big.Int, sig []byte) *Transaction {
	if len(sig) == 65 {
		sig = common.CopyBytes(sig)
	}
	// var f *big.Int
	// if fee != nil {
	// 	f.Set(fee)
	// }
	au_from := &Authentifier{R: sig[:32], S: sig[32:64], V: sig[64:]}

	tx := new(Transaction)
	tx.AccountNonce = nonce
	tx.From = au_from
	tx.TxFee = fee
	tx.CreationDate = time.Now().Format(TimeFormatString)
	tx.TxHash = tx.Hash()
	tx.Txsize = tx.Size()

	tx.Priority_lvl = tx.GetPriorityLvl()
	return tx
}

//// ChainId returns which chain id this transaction was signed for (if at all)
//func (tx Transaction) ChainId() *big.Int {
//	return deriveChainId(tx.data.V)
//}
//
//// Protected returns whether the transaction is protected from replay protection.
//func (tx Transaction) Protected() bool {
//	return isProtectedV(tx.data.V)
//}
//
//func isProtectedV(V *big.Int) bool {
//	if V.BitLen() <= 8 {
//		v := V.Uint64()
//		return v != 27 && v != 28
//	}
//	// anything not 27 or 28 are considered unprotected
//	return true
//}
//
//// EncodeRLP implements rlp.Encoder
//func (tx *Transaction) EncodeRLP(w io.Writer) error {
//	return rlp.Encode(w, &tx.data)
//}
//
//// DecodeRLP implements rlp.Decoder
//func (tx *Transaction) DecodeRLP(s *rlp.Stream) error {
//	_, UnitSize, _ := s.Kind()
//	err := s.Decode(&tx.data)
//	if err == nil {
//		tx.UnitSize.Store(common.StorageSize(rlp.ListSize(UnitSize)))
//	}
//
//	return err
//}
//
//// MarshalJSON encodes the web3 RPC transaction format.
//func (tx *Transaction) MarshalJSON() ([]byte, error) {
//	UnitHash := tx.Hash()
//	data := tx.data
//	data.Hash = &UnitHash
//	return data.MarshalJSON()
//}
//
//// UnmarshalJSON decodes the web3 RPC transaction format.
//func (tx *Transaction) UnmarshalJSON(input []byte) error {
//	var dec txdata
//	if err := dec.UnmarshalJSON(input); err != nil {
//		return err
//	}
//	var V byte
//	if isProtectedV(dec.V) {
//		chainID := deriveChainId(dec.V).Uint64()
//		V = byte(dec.V.Uint64() - 35 - 2*chainID)
//	} else {
//		V = byte(dec.V.Uint64() - 27)
//	}
//	if !crypto.ValidateSignatureValues(V, dec.R, dec.S, false) {
//		return errors.New("invalid transaction v, r, s values")
//	}
//	*tx = Transaction{data: dec}
//	return nil
//}
//
//func (tx Transaction) Data() []byte { return common.CopyBytes(tx.data.Payload) }
//
func (tx Transaction) PriorityLvl() float64 {
	return tx.Priority_lvl
}
func (tx Transaction) GetPriorityLvl() float64 {
	// priority_lvl=  fee/size*(1+(time.Now-CreationDate)/24)
	var priority_lvl float64
	if txfee := tx.TxFee.Int64(); txfee > 0 {
		t0, _ := time.Parse(TimeFormatString, tx.CreationDate)
		priority_lvl, _ = strconv.ParseFloat(fmt.Sprintf("%f", float64(txfee)/tx.Txsize.Float64()*(1+float64(time.Now().Hour()-t0.Hour())/24)), 64)
	}
	return priority_lvl
}
func (tx Transaction) SetPriorityLvl(priority float64) {
	tx.Priority_lvl = priority
}
func (tx Transaction) Nonce() uint64 { return tx.AccountNonce }

func (tx Transaction) Fee() *big.Int { return tx.TxFee }

//func (tx Transaction) Account() *common.Address { return tx.data.From }

//func (tx Transaction) CheckNonce() bool         { return true }
//
//// To returns the recipient address of theTransaction .
//// It returns nil if the transaction is a contract creation.
//func (tx Transaction) ToAddress() *common.Address {
//	if tx.data.Recipient == nil {
//		return nil
//	}
//	to := *tx.data.Recipient
//	return &to
//}
//
//// Hash hashes the RLP encoding of tx.
//// It uniquely identifies the transaction.
func (tx Transaction) Hash() common.Hash {
	v := rlp.RlpHash(tx)
	tx.TxHash.Set(v)
	return v
}

// Size returns the true RLP encoded storage UnitSize of the transaction, either by
// encoding and returning it, or returning a previsouly cached value.
func (tx *Transaction) Size() common.StorageSize {
	if size := tx.Txsize.Float64(); size > 0 {
		return tx.Txsize
	}
	c := writeCounter(0)
	rlp.Encode(&c, &tx)
	return common.StorageSize(c)
}

//func (tx *Transaction) String() string {
//	var from, to string
//	if tx.data.V != nil {
//		// make a best guess about the signer and use that to derive
//		// the sender.
//		signer := deriveSigner(tx.data.V)
//		if f, err := Sender(signer, tx); err != nil { // derive but don't cache
//			from = "[invalid sender: invalid sig]"
//		} else {
//			from = fmt.Sprintf("%x", f[:])
//		}
//	} else {
//		from = "[invalid sender: nil V field]"
//	}
//
//	if tx.data.Recipient == nil {
//		to = "[contract creation]"
//	} else {
//		to = fmt.Sprintf("%x", tx.data.Recipient[:])
//	}
//	enc, _ := rlp.EncodeToBytes(&tx.data)
//	return fmt.Sprintf(`
//	TX(%x)
//	Contract: %v
//	From:     %s
//	To:       %s
//	Nonce:    %v
//	Price: %#x
//	Value:    %#x
//	Data:     0x%x
//	V:        %#x
//	R:        %#x
//	S:        %#x
//	Hex:      %x
//`,
//		tx.Hash(),
//		tx.data.Recipient == nil,
//		from,
//		to,
//		tx.data.From,
//		tx.data.Price,
//		tx.data.Amount,
//		tx.data.Payload,
//		tx.data.V,
//		tx.data.R,
//		tx.data.S,
//		enc,
//	)
//}

// func (tx *Transaction) WithSignature(signer Signer, sig []byte) (*Transaction, error) {
// 	r, s, v, err := signer.SignatureValues(tx, sig)
// 	if err != nil {
// 		return nil, err
// 	}
// 	cpy := &Transaction{data: tx.data}
// 	cpy.data.R, cpy.data.S, cpy.data.V = r, s, v
// 	return cpy, nil
// }

// Cost returns amount + price
func (tx *Transaction) Cost() *big.Int {
	if tx.TxFee.Cmp(TXFEE) < 0 {
		tx.TxFee = TXFEE
	}
	return tx.TxFee
}

//// AsMessage returns the transaction as a core.Message.
////
//// AsMessage requires a signer to derive the sender.
////
//// XXX Rename message to something less arbitrary?
//func (tx *Transaction) AsMessage(s Signer) (Message, error) {
//	msg := Message{
//		from:       *tx.data.From,
//		gasPrice:   new(big.Int).Set(tx.data.Price),
//		to:         tx.data.Recipient,
//		amount:     tx.data.Amount,
//		data:       tx.data.Payload,
//		checkNonce: true,
//	}
//
//	var err error
//	msg.from, err = Sender(s, tx)
//	return msg, err
//}
//

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
	v := rlp.RlpHash(s)
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

// single account, otherwise a nonce comparison doesn't make much sense.
type TxByNonce Transactions

func (s TxByNonce) Len() int           { return len(s) }
func (s TxByNonce) Less(i, j int) bool { return s[i].AccountNonce < s[j].AccountNonce }
func (s TxByNonce) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

// TxByPrice implements both the sort and the heap interface, making it useful
// for all at once sorting as well as individually adding and removing elements.
type TxByPrice Transactions

func (s TxByPrice) Len() int           { return len(s) }
func (s TxByPrice) Less(i, j int) bool { return s[i].TxFee.Cmp(s[j].TxFee) < 0 }
func (s TxByPrice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func (s *TxByPrice) Push(x interface{}) {
	*s = append(*s, x.(*Transaction))
}

func (s *TxByPrice) Pop() interface{} {
	old := *s
	n := len(old)
	x := old[n-1]
	*s = old[0 : n-1]
	return x
}

type TxByPriority Transactions

func (s TxByPriority) Len() int           { return len(s) }
func (s TxByPriority) Less(i, j int) bool { return s[i].Priority_lvl > s[j].Priority_lvl }
func (s TxByPriority) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func (s *TxByPriority) Push(x interface{}) {
	*s = append(*s, x.(*Transaction))
}

func (s *TxByPriority) Pop() interface{} {
	old := *s
	n := len(old)
	x := old[n-1]
	*s = old[0 : n-1]
	return x
}

//// Message is a fully derived transaction and implements core.Message
////
//// NOTE: In a future PR this will be removed.
//
//func NewMessage(from, to *common.Address, nonce uint64, amount *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte, checkNonce bool) Message {
//	return Message{
//		from:       *from,
//		to:         to,
//		nonce:      nonce,
//		amount:     amount,
//		gasLimit:   gasLimit,
//		gasPrice:   gasPrice,
//		data:       data,
//		checkNonce: checkNonce,
//	}
//}
//
//func (m Message) From() *common.Address { return &m.from }
//func (m Message) To() *common.Address   { return m.to }
//func (m Message) GasPrice() *big.Int    { return m.gasPrice }
//func (m Message) Value() *big.Int       { return m.amount }
//func (m Message) Gas() uint64           { return m.gasLimit }
//func (m Message) Nonce() uint64         { return m.nonce }
//func (m Message) Data() []byte          { return m.data }
//func (m Message) CheckNonce() bool      { return m.checkNonce }
//
//// deriveChainId derives the chain id from the given v parameter
//func deriveChainId(v *big.Int) *big.Int {
//	if v.BitLen() <= 64 {
//		v := v.Uint64()
//		if v == 27 || v == 28 {
//			return new(big.Int)
//		}
//		return new(big.Int).SetUint64((v - 35) / 2)
//	}
//	v = new(big.Int).Sub(v, big.NewInt(35))
//	return v.Div(v, big.NewInt(2))
//}
//func rlpHash(x interface{}) (h common.Hash) {
//	hw := sha3.NewKeccak256()
//	rlp.Encode(hw, x)
//	hw.Sum(h[:0])
//	return h
//}
//
//// deriveSigner makes a *best* guess about which signer to use.
//func deriveSigner(V *big.Int) Signer {
//	if V.Sign() != 0 && isProtectedV(V) {
//		return NewEIP155Signer(deriveChainId(V))
//	} else {
//		return HomesteadSigner{}
//	}
//}
//
type writeCounter common.StorageSize

func (c *writeCounter) Write(b []byte) (int, error) {
	*c += writeCounter(len(b))
	return len(b), nil
}

var (
	EmptyRootHash = DeriveSha(Transactions{})
)

type TxLookupEntry struct {
	UnitHash  common.Hash
	UnitIndex uint64
	Index     uint64
}
