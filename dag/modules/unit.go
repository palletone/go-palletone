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
	"strings"
	"time"
	"unsafe"

	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"go.dedis.ch/kyber/v3"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/crypto"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/util"
	"github.com/palletone/go-palletone/core"
)

// validate unit state
//const (
//	UNIT_STATE_VALIDATED                = 0x00
//	UNIT_STATE_AUTHOR_SIGNATURE_PASSED  = 0x01
//	UNIT_STATE_EMPTY                    = 0x02
//	UNIT_STATE_INVALID_AUTHOR_SIGNATURE = 0x03
//	UNIT_STATE_INVALID_GROUP_SIGNATURE  = 0x04
//	UNIT_STATE_HAS_INVALID_TRANSACTIONS = 0x05
//	UNIT_STATE_INVALID_SIZE             = 0x06
//	UNIT_STATE_INVALID_EXTRA_DATA       = 0x07
//	UNIT_STATE_INVALID_HEADER           = 0x08
//	UNIT_STATE_CHECK_HEADER_PASSED      = 0x09
//	UNIT_STATE_INVALID_HEADER_WITNESS   = 0x10
//	UNIT_STATE_OTHER_ERROR              = 0xFF
//)

// unit state
const (
	U_STATE_NO_GROUPSIGN = 0x20
	U_STATE_NO_CONFIRMED = 0x21
	U_STATE_CONFIRMED    = 0x22
)

type Header struct {
	ParentsHash []common.Hash `json:"parents_hash"`
	Authors     Authentifier  `json:"mediator"`     // the unit creation authors
	GroupSign   []byte        `json:"group_sign"`   // 群签名, 用于加快单元确认速度
	GroupPubKey []byte        `json:"group_pubKey"` // 群公钥, 用于验证群签名
	TxRoot      common.Hash   `json:"root"`
	TxsIllegal  []uint16      `json:"txs_illegal"` //Unit中非法交易索引
	Number      *ChainIndex   `json:"index"`
	Extra       []byte        `json:"extra"`
	Time        int64         `json:"creation_time"` // unit create time
	CryptoLib   []byte        `json:"crypto_lib"`    //该区块使用的加解密算法和哈希算法，0位表示非对称加密算法，1位表示Hash算法
}

func (cpy *Header) CopyHeader(h *Header) {
	cpy = h
	if len(h.ParentsHash) > 0 {
		cpy.ParentsHash = make([]common.Hash, len(h.ParentsHash))
		for i := 0; i < len(h.ParentsHash); i++ {
			cpy.ParentsHash[i] = h.ParentsHash[i]
		}
	}
}

func NewHeader(parents []common.Hash, used uint64, extra []byte) *Header {
	hashs := make([]common.Hash, 0)
	hashs = append(hashs, parents...) // 切片指针传递的问题，这里得再review一下。
	var b []byte
	number := &ChainIndex{}
	return &Header{ParentsHash: hashs, Number: number, Extra: append(b, extra...)}
}

func HeaderEqual(oldh, newh *Header) bool {
	if oldh.Hash() == newh.Hash() {
		return true
	}
	pars := len(oldh.ParentsHash)
	// 两个parents hash
	if pars == 2 && 2 == len(newh.ParentsHash) {
		if oldh.ParentsHash[0] == newh.ParentsHash[1] && oldh.ParentsHash[1] == newh.ParentsHash[0] {
			return true
		}
	}
	return false
}

func (h *Header) Index() uint64 {
	return h.Number.Index
}
func (h *Header) ChainIndex() *ChainIndex {
	return h.Number
}

func (h *Header) Hash() common.Hash {
	emptyHeader := CopyHeader(h)
	// 计算header’hash时 剔除群签
	//emptyHeader.Authors = Authentifier{} Hash必须包含Mediator签名
	emptyHeader.GroupSign = nil
	emptyHeader.GroupPubKey = nil
	return util.RlpHash(emptyHeader)
}
func (h *Header) HashWithoutAuthor() common.Hash {
	emptyHeader := CopyHeader(h)
	// 计算header’hash时 剔除群签
	emptyHeader.Authors = Authentifier{}
	emptyHeader.GroupSign = nil
	emptyHeader.GroupPubKey = nil
	return util.RlpHash(emptyHeader)
}

// HashWithOutTxRoot return  header's hash without txs root.
func (h *Header) HashWithOutTxRoot() common.Hash {
	emptyHeader := CopyHeader(h)
	// 计算header’hash时 剔除签名和群签
	emptyHeader.Authors = Authentifier{}
	emptyHeader.GroupSign = nil
	emptyHeader.GroupPubKey = nil
	emptyHeader.TxRoot = common.Hash{}
	b, err := json.Marshal(emptyHeader)
	if err != nil {
		log.Error("json marshal error", "error", err)
		return common.Hash{}
	}
	return util.RlpHash(b[:])

}

func (h *Header) Size() common.StorageSize {
	return common.StorageSize(unsafe.Sizeof(*h)) + common.StorageSize(len(h.Extra)/8)
}

// CopyHeader creates a deep copy of a block header to prevent side effects from
// modifying a header variable.
func CopyChainIndex(index *ChainIndex) *ChainIndex {
	cop := new(ChainIndex)
	cop.AssetID = index.AssetID
	//cop.IsMain = index.IsMain
	cop.Index = index.Index
	return cop
}
func CopyHeader(h *Header) *Header {
	if h == nil {
		return nil
	}
	cpy := Header{}
	//	cpy.Number = h.Number
	if h.Number != nil {
		cpy.Number = CopyChainIndex(h.Number)
	}
	cpy.Extra = h.Extra[:]
	cpy.Time = h.Time
	cpy.Authors = h.Authors

	if len(h.ParentsHash) > 0 {
		cpy.ParentsHash = make([]common.Hash, len(h.ParentsHash))
		for i := 0; i < len(h.ParentsHash); i++ {
			cpy.ParentsHash[i].Set(h.ParentsHash[i])
		}
	}

	if len(h.GroupSign) > 0 {
		copy(cpy.GroupSign, h.GroupSign)
	}

	if len(h.GroupPubKey) > 0 {
		copy(cpy.GroupPubKey, h.GroupPubKey)
	}

	if len(h.TxRoot) > 0 {
		cpy.TxRoot.Set(h.TxRoot)
	}

	if len(h.TxsIllegal) > 0 {
		cpy.TxsIllegal = make([]uint16, 0)
		for _, txsI := range h.TxsIllegal {
			cpy.TxsIllegal = append(cpy.TxsIllegal, txsI)
		}
	}

	return &cpy
}

func (u *Unit) CopyBody(txs Transactions) Transactions {
	if len(txs) > 0 {
		u.Txs = make([]*Transaction, len(txs))
		for i, pTx := range txs {
			//hash := pTx.Hash()

			tx := Transaction{}
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

func (s Units) Len() int {
	return len(s)
}

func (s Units) Less(i, j int) bool {
	return s[i].NumberU64() < s[j].NumberU64()
}

func (s Units) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

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

func (h *Header) Author() common.Address {
	if h == nil {
		log.Error("the Unit Header pointer is nil!")
	}
	return crypto.PubkeyBytesToAddress(h.Authors.PubKey)
}

func (unit *Unit) Author() common.Address {
	if unit == nil {
		log.Error("the Unit pointer is nil!")
	}
	return unit.UnitHeader.Author()
}

func (unit *Unit) GroupPubKey() (kyber.Point, error) {
	pubKeyB := unit.UnitHeader.GroupPubKey
	if len(pubKeyB) == 0 {
		return nil, errors.New("group public key is null")
	}

	pubKey := core.Suite.Point()
	err := pubKey.UnmarshalBinary(pubKeyB)

	return pubKey, err
}

//type OutPoint struct {
//	TxHash       common.Hash // reference Utxo struct key field
//	MessageIndex uint32      // message index in transaction
//	OutIndex     uint32
//}

func (unit *Unit) IsEmpty() bool {
	if unit == nil || unit.Hash() == (common.Hash{}) {
		return true
	}
	return false
}
func (unit *Unit) String4Log() string {
	txs := []common.Hash{}
	for _, tx := range unit.Txs {
		txs = append(txs, tx.Hash())
	}
	return fmt.Sprintf("Hash:%s,Index:%d,Txs:%x", unit.Hash().String(), unit.NumberU64(), txs)
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
	AssetID AssetId `json:"asset_id"`
	//IsMain  bool    `json:"is_main"`
	Index uint64 `json:"index"`
}

func NewChainIndex(assetId AssetId, idx uint64) *ChainIndex {
	return &ChainIndex{AssetID: assetId, Index: idx}
}
func (height *ChainIndex) String() string {
	return fmt.Sprintf("%s-%d", height.AssetID.GetSymbol(), height.Index)
}

//Index 8Bytes + AssetID 16Bytes
func (height *ChainIndex) Bytes() []byte {
	idx := make([]byte, 8)
	littleEndian.PutUint64(idx, height.Index)
	return append(idx, height.AssetID.Bytes()...)
}

func (height *ChainIndex) SetBytes(data []byte) {
	height.Index = littleEndian.Uint64(data[:8])
	height.AssetID.SetBytes(data[8:])
}

func (height *ChainIndex) Equal(in *ChainIndex) bool {
	if in == nil {
		return false
	}
	if !bytes.Equal(height.AssetID[:], in.AssetID[:]) {
		return false
	}
	if height.Index != in.Index {
		return false
	}
	return true
}

//type Author struct {
//	Address        common.Address `json:"address"`
//	Pubkey         []byte/*common.Hash*/ `json:"pubkey"`
//	TxAuthentifier *Authentifier `json:"authentifiers"`
//}

type Authentifier struct {
	PubKey    []byte `json:"pubkey"`
	Signature []byte `json:"signature"`
}

func (au *Authentifier) Empty() bool {
	return len(au.PubKey) == 0 || len(au.Signature) == 0
}
func (au *Authentifier) Address() common.Address {

	return crypto.PubkeyBytesToAddress(au.PubKey)
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
		if transaction.Hash() == hash {
			return transaction
		}
	}
	return nil
}

// function Hash, return the unit's hash.
func (u *Unit) Hash() common.Hash {
	headerHash := u.UnitHeader.Hash()
	if u.UnitHash != headerHash {
		u.UnitHash = common.Hash{}
		u.UnitHash.Set(headerHash)
	}
	return u.UnitHash
}

// function Size, return the unit's StorageSize.
func (u *Unit) Size() common.StorageSize {
	if u.UnitSize > 0 {
		return u.UnitSize
	}
	emptyUnit := &Unit{}
	emptyUnit.UnitHeader = CopyHeader(u.UnitHeader)
	//emptyUnit.UnitHeader.Authors = nil
	emptyUnit.UnitHeader.GroupSign = make([]byte, 0)
	emptyUnit.CopyBody(u.Txs[:])

	b, err := rlp.EncodeToBytes(emptyUnit)
	if err != nil {
		log.Errorf("rlp encode Unit error:%s", err.Error())
		return common.StorageSize(0)
	} else {
		if len(b) > 0 {
			u.UnitSize = common.StorageSize(len(b))
		}
		return common.StorageSize(len(b))
	}
}

//func (u *Unit) NumberU64() uint64 { return u.Head.Number.Uint64() }
func (u *Unit) Number() *ChainIndex {
	return u.UnitHeader.Number
}

func (u *Unit) NumberU64() uint64 {
	return u.UnitHeader.Number.Index
}

func (u *Unit) Timestamp() int64 {
	return u.UnitHeader.Time
}

// return unit's parents UnitHash
func (u *Unit) ParentHash() []common.Hash {
	return u.UnitHeader.ParentsHash
}

func (u *Unit) SetGroupSign(sign []byte) {
	if len(sign) > 0 {
		u.UnitHeader.GroupSign = sign
	}
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

/*
func RSVtoPublicKey(hash, r, s, v []byte) (*ecdsa.PublicKey, error) {
	sig := make([]byte, 65)
	copy(sig[32-len(r):32], r)
	copy(sig[64-len(s):64], s)
	copy(sig[64:], v)
	return crypto.SigToPub(hash, sig)
}
*/
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
