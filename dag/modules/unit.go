/* This file is part of go-palletone.
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

   @author PalletOne core developers <dev@pallet.one>
   @date 2018
*/

// unit package, unit structure and storage api
package modules

import (
	"crypto/ecdsa"
	"strings"
	"time"
	"unsafe"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/crypto"
	"github.com/palletone/go-palletone/common/rlp"
)

// validate unit state
const (
	UNIT_STATE_VALIDATED                = 0x00
	UNIT_STATE_AUTHOR_SIGNATURE_PASSED  = 0x01
	UNIT_STATE_EMPTY                    = 0x02
	UNIT_STATE_INVALID_AUTHOR_SIGNATURE = 0x03
	UNIT_STATE_INVALID_GROUP_SIGNATURE  = 0x04
	UNIT_STATE_HAS_INVALID_TRANSACTIONS = 0x05
	UNIT_STATE_INVALID_SIZE             = 0x06
	UNIT_STATE_INVALID_EXTRA_DATA       = 0x07
	UNIT_STATE_INVALID_HEADER           = 0x08
	UNIT_STATE_CHECK_HEADER_PASSED      = 0x09
	UNIT_STATE_INVALID_HEADER_WITNESS   = 0x10
	UNIT_STATE_OTHER_ERROR              = 0xFF
)

type Header struct {
	ParentsHash  []common.Hash `json:"parents_hash"`
	AssetIDs     []IDType16    `json:"assets"`
	Authors      Authentifier  `json:"author" rlp:"-"` // the unit creation authors
	GroupSign    []byte        `json:"witness"`        // 群签名
	TxRoot       common.Hash   `json:"root"`
	Number       ChainIndex    `json:"index"`
	Extra        []byte        `json:"extra"`
	Creationdate int64         `json:"creation_time"` // unit create time
}

func (cpy *Header) CopyHeader(h *Header) {
	cpy = h
	if len(h.ParentsHash) > 0 {
		cpy.ParentsHash = make([]common.Hash, len(h.ParentsHash))
		for i := 0; i < len(h.ParentsHash); i++ {
			cpy.ParentsHash[i] = h.ParentsHash[i]
		}
	}

	if len(h.AssetIDs) > 0 {
		cpy.AssetIDs = make([]IDType16, len(h.AssetIDs))
		for i := 0; i < len(h.AssetIDs); i++ {
			cpy.AssetIDs[i] = h.AssetIDs[i]
		}
	}

}

func NewHeader(parents []common.Hash, asset []IDType16, used uint64, extra []byte) *Header {
	hashs := make([]common.Hash, 0)
	hashs = append(hashs, parents...) // 切片指针传递的问题，这里得再review一下。
	var b []byte
	return &Header{ParentsHash: hashs, AssetIDs: asset, Extra: append(b, extra...)}
}

func HeaderEqual(oldh, newh *Header) bool {
	if oldh.ParentsHash[0] == newh.ParentsHash[0] && oldh.ParentsHash[1] == newh.ParentsHash[1] {
		return true
	} else if oldh.ParentsHash[0] == newh.ParentsHash[1] && oldh.ParentsHash[1] == newh.ParentsHash[0] {
		return true
	}
	return false
}

func (h *Header) Index() uint64 {
	return h.Number.Index
}
func (h *Header) ChainIndex() *ChainIndex {
	return &h.Number
}

func (h *Header) Hash() common.Hash {
	emptyHeader := CopyHeader(h)
	//emptyHeader.Authors = nil
	emptyHeader.GroupSign = make([]byte, 0)
	return rlp.RlpHash(emptyHeader)
}

func (h *Header) Size() common.StorageSize {
	return common.StorageSize(unsafe.Sizeof(*h)) + common.StorageSize(len(h.Extra)/8)
}

// CopyHeader creates a deep copy of a block header to prevent side effects from
// modifying a header variable.
func CopyHeader(h *Header) *Header {
	cpy := *h

	if len(h.ParentsHash) > 0 {
		cpy.ParentsHash = make([]common.Hash, len(h.ParentsHash))
		for i := 0; i < len(h.ParentsHash); i++ {
			cpy.ParentsHash[i].Set(h.ParentsHash[i])
		}
	}

	if len(h.AssetIDs) > 0 {
		copy(cpy.AssetIDs, h.AssetIDs)
	}

	if len(h.GroupSign) > 0 {
		copy(cpy.GroupSign, h.GroupSign)
	}

	if len(h.TxRoot) > 0 {
		cpy.TxRoot.Set(h.TxRoot)
	}

	return &cpy
}

func (u *Unit) CopyBody(txs Transactions) Transactions {
	if len(txs) > 0 {
		u.Txs = make([]*Transaction, len(txs))
		for i, pTx := range txs {
			tx := Transaction{
				TxHash: pTx.TxHash,
			}
			if len(pTx.TxMessages) > 0 {
				tx.TxMessages = make([]*Message, len(pTx.TxMessages))
				for j := 0; j < len(pTx.TxMessages); j++ {
					tx.TxMessages[j] = pTx.TxMessages[j]
				}
			}
			u.Txs[i] = &tx
		}
	}
	return u.Txs
}

//wangjiyou add for ptn/fetcher.go
type Units []*Unit

// key: unit.UnitHash(unit)
type Unit struct {
	UnitHeader *Header            `json:"unit_header"`  // unit header
	Txs        Transactions       `json:"transactions"` // transaction list
	UnitHash   common.Hash        `json:"unit_hash"`    // unit hash
	UnitSize   common.StorageSize `json:"unit_size"`    // unit size
	// These fields are used by package ptn to track
	// inter-peer block relay.
	ReceivedAt   time.Time
	ReceivedFrom interface{}
}

func (unit *Unit) UnitAuthor() *common.Address {
	if unit != nil {
		return &unit.UnitHeader.Authors.Address
	}
	return nil
}

//type OutPoint struct {
//	TxHash       common.Hash // reference Utxo struct key field
//	MessageIndex uint32      // message index in transaction
//	OutIndex     uint32
//}

func (unit *Unit) IsEmpty() bool {
	if unit == nil || len(unit.Txs) <= 0 {
		return true
	}
	return false
}

//type Transactions []*Transaction
type TxPoolTxs []*TxPoolTransaction

//type Transaction struct {
//	TxHash     common.Hash `json:"txhash"`
//	TxMessages []Message   `json:"messages"`
//	Locktime   uint32      `json:"lock_time"`
//}
//出于DAG和基于Token的分区共识的考虑，设计了该ChainIndex，
type ChainIndex struct {
	AssetID IDType16
	IsMain  bool
	Index   uint64
}

func (height ChainIndex) String() string {
	return common.Bytes2Hex(height.Bytes())
}
func (height ChainIndex) Bytes() []byte {
	data, err := rlp.EncodeToBytes(height)
	if err != nil {
		return nil
	}
	return data[:]
}

//type Author struct {
//	Address        common.Address `json:"address"`
//	Pubkey         []byte/*common.Hash*/ `json:"pubkey"`
//	TxAuthentifier *Authentifier `json:"authentifiers"`
//}

type Authentifier struct {
	Address common.Address `json:"address"`
	R       []byte         `json:"r"`
	S       []byte         `json:"s"`
	V       []byte         `json:"v"`
}

func (au *Authentifier) Empty() bool {
	if common.IsValidAddress(au.Address.String()) || len(au.R) == 0 || len(au.S) == 0 || len(au.V) == 0 {
		return true
	}
	return false
}

func NewUnit(header *Header, txs Transactions) *Unit {
	u := &Unit{
		UnitHeader: CopyHeader(header),
		Txs:        CopyTransactions(txs),
	}
	u.UnitSize = u.Size()
	u.UnitHash = u.Hash()
	return u
}

func CopyTransactions(txs Transactions) Transactions {
	cpy := txs
	return cpy
}

type UnitNonce [8]byte

/************************** Unit Members  *****************************/
func (u *Unit) Header() *Header {
	return CopyHeader(u.UnitHeader)
}

// transactions
func (u *Unit) Transactions() []*Transaction {
	return u.Txs
}

// return transaction
func (u *Unit) Transaction(hash common.Hash) *Transaction {
	for _, transaction := range u.Txs {
		if transaction.TxHash == hash {
			return transaction
		}
	}
	return nil
}

// function Hash, return the unit's hash.
func (u *Unit) Hash() common.Hash {
	if u.UnitHash != u.UnitHeader.Hash() {
		u.UnitHash = common.Hash{}
		u.UnitHash.Set(u.UnitHeader.Hash())
	}
	return u.UnitHash
}

// function Size, return the unit's StorageSize.
func (u *Unit) Size() common.StorageSize {
	if u.UnitSize > 0 {
		return u.UnitSize
	}
	emptyUnit := Unit{}
	emptyUnit.UnitHeader = CopyHeader(u.UnitHeader)
	//emptyUnit.UnitHeader.Authors = nil
	emptyUnit.UnitHeader.GroupSign = make([]byte, 0)
	emptyUnit.CopyBody(u.Txs[:])

	b, err := rlp.EncodeToBytes(emptyUnit)
	if err != nil {
		return common.StorageSize(0)
	} else {
		if len(b) > 0 {
			u.UnitSize = common.StorageSize(len(b))
		}
		return common.StorageSize(len(b))
	}
}

//func (u *Unit) NumberU64() uint64 { return u.Head.Number.Uint64() }
func (u *Unit) Number() ChainIndex {
	return u.UnitHeader.Number
}

func (u *Unit) NumberU64() uint64 {
	return u.UnitHeader.Number.Index
}

func (u *Unit) Timestamp() int64 {
	return u.UnitHeader.Creationdate
}

// return unit's parents UnitHash
func (u *Unit) ParentHash() []common.Hash {
	return u.UnitHeader.ParentsHash
}

type ErrUnit float64

func (e ErrUnit) Error() string {
	switch e {
	case -1:
		return "Unit size error"
	case -2:
		return "Unit signature error"
	case -3:
		return "Unit header save error"
	case -4:
		return "Unit tx size error"
	case -5:
		return "Save create token transaction error"
	case -6:
		return "Save config transaction error"
	default:
		return ""
	}
	return ""
}

/************************** Unit Members  *****************************/

// NewBlockWithHeader creates a block with the given header data. The
// header data is copied, changes to header and to the field values
// will not affect the block.
func NewUnitWithHeader(header *Header) *Unit {
	return &Unit{UnitHeader: CopyHeader(header)}
}

// WithBody returns a new block with the given transaction and uncle contents.
func (b *Unit) WithBody(transactions []*Transaction) *Unit {
	// check transactions merkle root
	txs := b.CopyBody(transactions)
	//fmt.Printf("withbody==>%#v\n", txs[0])
	//root := core.DeriveSha(txs)
	//if strings.Compare(root.String(), b.UnitHeader.TxRoot.String()) != 0 {
	//	return nil
	//}
	// set unit body
	b.Txs = CopyTransactions(txs)
	b.UnitSize = b.Size()
	b.UnitHash = b.Hash()
	return b
}

func (u *Unit) ContainsParent(pHash common.Hash) bool {
	ps := pHash.String()
	for _, hash := range u.UnitHeader.ParentsHash {
		if strings.Compare(hash.String(), ps) == 0 {
			return true
		}
	}
	return false
}

func MsgstoAddress(msgs []*Message) common.Address {
	forms := make([]common.Address, 0)
	//payment load to address.

	for _, msg := range msgs {
		payment, ok := msg.Payload.(PaymentPayload)
		if !ok {
			break
		}
		for _, pay := range payment.Inputs {
			// 通过签名信息还原出address
			from := new(common.Address)
			from.SetBytes(pay.Extra[:])
			forms = append(forms, *from)
		}
	}
	if len(forms) > 0 {
		return forms[0]
	}
	return common.Address{}
}
func RSVtoPublicKey(hash, r, s, v []byte) (*ecdsa.PublicKey, error) {
	sig := make([]byte, 65)
	copy(sig[32-len(r):32], r)
	copy(sig[64-len(s):64], s)
	copy(sig[64:], v)
	return crypto.SigToPub(hash, sig)
}

type TxValidationCode int32

const (
	TxValidationCode_VALID                        TxValidationCode = 0
	TxValidationCode_INVALID_CONTRACT_TEMPLATE    TxValidationCode = 1
	TxValidationCode_INVALID_FEE                  TxValidationCode = 2
	TxValidationCode_BAD_COMMON_HEADER            TxValidationCode = 3
	TxValidationCode_BAD_CREATOR_SIGNATURE        TxValidationCode = 4
	TxValidationCode_INVALID_ENDORSER_TRANSACTION TxValidationCode = 5
	TxValidationCode_INVALID_CONFIG_TRANSACTION   TxValidationCode = 6
	TxValidationCode_UNSUPPORTED_TX_PAYLOAD       TxValidationCode = 7
	TxValidationCode_BAD_PROPOSAL_TXID            TxValidationCode = 8
	TxValidationCode_DUPLICATE_TXID               TxValidationCode = 9
	TxValidationCode_ENDORSEMENT_POLICY_FAILURE   TxValidationCode = 10
	TxValidationCode_MVCC_READ_CONFLICT           TxValidationCode = 11
	TxValidationCode_PHANTOM_READ_CONFLICT        TxValidationCode = 12
	TxValidationCode_UNKNOWN_TX_TYPE              TxValidationCode = 13
	TxValidationCode_TARGET_CHAIN_NOT_FOUND       TxValidationCode = 14
	TxValidationCode_MARSHAL_TX_ERROR             TxValidationCode = 15
	TxValidationCode_NIL_TXACTION                 TxValidationCode = 16
	TxValidationCode_EXPIRED_CHAINCODE            TxValidationCode = 17
	TxValidationCode_CHAINCODE_VERSION_CONFLICT   TxValidationCode = 18
	TxValidationCode_BAD_HEADER_EXTENSION         TxValidationCode = 19
	TxValidationCode_BAD_CHANNEL_HEADER           TxValidationCode = 20
	TxValidationCode_BAD_RESPONSE_PAYLOAD         TxValidationCode = 21
	TxValidationCode_BAD_RWSET                    TxValidationCode = 22
	TxValidationCode_ILLEGAL_WRITESET             TxValidationCode = 23
	TxValidationCode_INVALID_WRITESET             TxValidationCode = 24
	TxValidationCode_INVALID_MSG                  TxValidationCode = 25
	TxValidationCode_INVALID_PAYMMENTLOAD         TxValidationCode = 26
	TxValidationCode_INVALID_PAYMMENT_INPUT       TxValidationCode = 27
	TxValidationCode_INVALID_PAYMMENT_OUTPUT      TxValidationCode = 28
	TxValidationCode_INVALID_PAYMMENT_LOCKTIME    TxValidationCode = 29
	TxValidationCode_INVALID_OUTPOINT             TxValidationCode = 30
	TxValidationCode_INVALID_AMOUNT               TxValidationCode = 31
	TxValidationCode_INVALID_ASSET                TxValidationCode = 32
	TxValidationCode_NOT_VALIDATED                TxValidationCode = 254
	TxValidationCode_NOT_COMPARE_SIZE             TxValidationCode = 255
	TxValidationCode_INVALID_OTHER_REASON         TxValidationCode = 256
)

var TxValidationCode_name = map[int32]string{
	0:   "VALID",
	1:   "INVALID_CONTRACT_TEMPLATE",
	2:   "INVALID_FEE",
	3:   "BAD_COMMON_HEADER",
	4:   "BAD_CREATOR_SIGNATURE",
	5:   "INVALID_ENDORSER_TRANSACTION",
	6:   "INVALID_CONFIG_TRANSACTION",
	7:   "UNSUPPORTED_TX_PAYLOAD",
	8:   "BAD_PROPOSAL_TXID",
	9:   "DUPLICATE_TXID",
	10:  "ENDORSEMENT_POLICY_FAILURE",
	11:  "MVCC_READ_CONFLICT",
	12:  "PHANTOM_READ_CONFLICT",
	13:  "UNKNOWN_TX_TYPE",
	14:  "TARGET_CHAIN_NOT_FOUND",
	15:  "MARSHAL_TX_ERROR",
	16:  "NIL_TXACTION",
	17:  "EXPIRED_CHAINCODE",
	18:  "CHAINCODE_VERSION_CONFLICT",
	19:  "BAD_HEADER_EXTENSION",
	20:  "BAD_CHANNEL_HEADER",
	21:  "BAD_RESPONSE_PAYLOAD",
	22:  "BAD_RWSET",
	23:  "ILLEGAL_WRITESET",
	24:  "INVALID_WRITESET",
	254: "NOT_VALIDATED",
	255: "NOT_COMPARE_SIZE",
	256: "INVALID_OTHER_REASON",
}

/**
根据大端规则填充字节
To full fill bytes according bigendian
*/
func FillBytes(data []byte, lenth uint8) []byte {
	newBytes := make([]byte, lenth)
	if len(data) < int(lenth) {
		len := int(lenth) - len(data)
		for i, b := range data {
			newBytes[len+i] = b
		}
	} else {
		newBytes = data[:lenth]
	}
	return newBytes
}
