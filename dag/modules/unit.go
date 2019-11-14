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
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
	"unsafe"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/crypto"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/util"
	"github.com/palletone/go-palletone/core"
	"go.dedis.ch/kyber/v3"
	"io"
)

// unit state
const (
	U_STATE_NO_GROUPSIGN = 0x20
	U_STATE_NO_CONFIRMED = 0x21
	U_STATE_CONFIRMED    = 0x22
)

type Header struct {
	header       *header_sdw `json:"header"`
	hash         common.Hash `json:"hash"`
	group_sign   []byte      `json:"group_sign"`   // 群签名, 用于加快单元确认速度
	group_pubKey []byte      `json:"group_pubKey"` // 群公钥, 用于验证群签名
}
type header_sdw struct {
	ParentsHash []common.Hash `json:"parents_hash"`
	Authors     Authentifier  `json:"mediator"` // the unit creation authors
	//GroupSign   []byte        `json:"group_sign"`   // 群签名, 用于加快单元确认速度
	//GroupPubKey []byte        `json:"group_pubKey"` // 群公钥, 用于验证群签名
	TxRoot     common.Hash `json:"root"`
	TxsIllegal []uint16    `json:"txs_illegal"` //Unit中非法交易索引
	Number     *ChainIndex `json:"index"`
	Extra      []byte      `json:"extra"`
	Time       int64       `json:"creation_time"` // unit create time
	CryptoLib  []byte      `json:"crypto_lib"`    //该区块使用的加解密算法和哈希算法，0位表示非对称加密算法，1位表示Hash算法
}

func new_header_sdw() *header_sdw {
	return &header_sdw{ParentsHash: make([]common.Hash, 0),
		Authors:    Authentifier{},
		TxRoot:     common.Hash{},
		TxsIllegal: make([]uint16, 0),
		Number:     new(ChainIndex),
		Extra:      make([]byte, 0),
		CryptoLib:  make([]byte, 0)}
}
func (h *Header) NumberU64() uint64 {
	return h.header.Number.Index
}

func (h *Header) Timestamp() int64 {
	return h.header.Time
}

func (h *Header) GetGroupPubKeyByte() []byte {
	return h.group_pubKey
}

func (h *Header) GetGroupPubKey() (kyber.Point, error) {
	pubKeyB := h.group_pubKey
	if len(pubKeyB) == 0 {
		return nil, errors.New("group public key is null")
	}

	pubKey := core.Suite.Point()
	err := pubKey.UnmarshalBinary(pubKeyB)

	return pubKey, err
}
func (h *Header) SetNumber(number *ChainIndex) {
	if h.header == nil {
		h.header = new_header_sdw()
	}
	h.header.Number = number
}
func (h *Header) SetParentHash(parents []common.Hash) {
	if h.header == nil {
		h.header = new_header_sdw()
	}
	if len(parents) > 0 {
		h.header.ParentsHash = make([]common.Hash, 0)
		h.header.ParentsHash = append(h.header.ParentsHash, parents...)
	}
}
func (h *Header) SetTxRoot(txroot common.Hash) {
	if h.header == nil {
		h.header = new_header_sdw()
	}
	h.header.TxRoot.Set(txroot)
}
func (h *Header) SetAuthor(author Authentifier) {
	if h.header == nil {
		h.header = new_header_sdw()
	}
	h.header.Authors.Signature = author.Signature
	h.header.Authors.PubKey = author.PubKey
}
func (h *Header) SetTime(timestamp int64) {
	if h.header == nil {
		h.header = new_header_sdw()
	}
	h.header.Time = timestamp
}
func (h *Header) SetTxsIllegal(txsillegal []uint16) {
	if h.header == nil {
		h.header = new_header_sdw()
	}
	if len(txsillegal) > 0 {
		h.header.TxsIllegal = make([]uint16, 0)
		h.header.TxsIllegal = append(h.header.TxsIllegal, txsillegal...)
	}
}
func (h *Header) SetExtra(extra []byte) {
	if h.header == nil {
		h.header = new_header_sdw()
	}
	if len(extra) > 0 {
		h.header.Extra = make([]byte, 0)
		h.header.Extra = append(h.header.Extra, extra...)
	}
}
func (h *Header) SetCryptoLib(cryptolib []byte) {
	if h.header == nil {
		h.header = new_header_sdw()
	}
	if len(cryptolib) > 0 {
		h.header.CryptoLib = make([]byte, 0)
		h.header.CryptoLib = append(h.header.CryptoLib, cryptolib...)
	}
}
func (cpy *Header) CopyHeader(h *Header) {
	index := new(ChainIndex)
	if h.header == nil {
		h.header = new_header_sdw()
	}
	index.Index = h.header.Number.Index
	index.AssetID = h.header.Number.AssetID
	*cpy = *h
	cpy.header.Number = index
}

func NewHeader(parents []common.Hash, used uint64, extra []byte) *Header {
	//hashs := make([]common.Hash, 0)
	//hashs = append(hashs, parents...) // 切片指针传递的问题，这里得再review一下。
	number := &ChainIndex{}
	h := new(Header)
	h.header.ParentsHash = make([]common.Hash, 0)
	h.header.ParentsHash = append(h.header.ParentsHash, parents...)
	h.header.Number = number
	h.header.Extra = make([]byte, 0)
	h.header.Extra = append(h.header.Extra, extra...)
	return h
}

func (h *Header) Index() uint64 {
	return h.header.Number.Index
}
func (h *Header) ChainIndex() *ChainIndex {
	return h.header.Number
}
func (h *Header) TxRoot() common.Hash {
	return h.header.TxRoot
}
func (h *Header) Hash() common.Hash {
	// 计算header’hash时 剔除群签
	//groupSign := h.GroupSign
	//groupPubKey := h.GroupPubKey
	//h.GroupSign = make([]byte, 0)
	//h.GroupPubKey = make([]byte, 0)
	//hash := util.RlpHash(h)
	//h.GroupSign = append(h.GroupSign, groupSign...)
	//h.GroupPubKey = append(h.GroupPubKey, groupPubKey...)

	if h.hash == (common.Hash{}) {
		h.hash.Set(util.RlpHash(h.header))
	}
	return h.hash
}
func (h *Header) HashWithoutAuthor() common.Hash {
	//groupSign := h.header.GroupSign
	//groupPubKey := h.header.GroupPubKey
	author := h.header.Authors
	//h.header.GroupSign = make([]byte, 0)
	//h.header.GroupPubKey = make([]byte, 0)
	h.header.Authors = Authentifier{}
	hash := util.RlpHash(h.header)
	//h.header.GroupSign = append(h.header.GroupSign, groupSign...)
	//h.header.GroupPubKey = append(h.header.GroupPubKey, groupPubKey...)
	h.header.Authors.PubKey = author.PubKey[:]
	h.header.Authors.Signature = author.Signature[:]
	return hash
}

// HashWithOutTxRoot return  header's hash without txs root.
func (h *Header) HashWithOutTxRoot() common.Hash {
	//groupSign := h.header.GroupSign
	//groupPubKey := h.header.GroupPubKey
	author := h.header.Authors
	txroot := h.header.TxRoot
	//h.header.GroupSign = make([]byte, 0)
	//h.header.GroupPubKey = make([]byte, 0)
	h.header.Authors = Authentifier{}
	h.header.TxRoot = common.Hash{}

	b, err := json.Marshal(h.header)
	if err != nil {
		log.Error("json marshal error", "error", err)
		return common.Hash{}
	}
	hash := util.RlpHash(b[:])
	//h.header.GroupSign = append(h.header.GroupSign, groupSign...)
	//h.header.GroupPubKey = append(h.header.GroupPubKey, groupPubKey...)
	h.header.Authors.PubKey = author.PubKey[:]
	h.header.Authors.Signature = author.Signature[:]
	h.header.TxRoot = txroot
	return hash
}

func (h *Header) Size() common.StorageSize {
	header := h.header
	return common.StorageSize(unsafe.Sizeof(*header)) + common.StorageSize(len(h.header.Extra)/8)
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
	if h.header.Number != nil {
		cpy.header.Number = CopyChainIndex(h.header.Number)
	}
	cpy.header.Extra = h.header.Extra[:]
	cpy.header.Time = h.header.Time
	cpy.header.Authors = h.header.Authors

	if len(h.header.ParentsHash) > 0 {
		cpy.header.ParentsHash = make([]common.Hash, len(h.header.ParentsHash))
		for i := 0; i < len(h.header.ParentsHash); i++ {
			cpy.header.ParentsHash[i].Set(h.header.ParentsHash[i])
		}
	}

	if len(h.group_sign) > 0 {
		cpy.group_sign = make([]byte, len(h.group_sign))
		copy(cpy.group_sign, h.group_sign)
	}

	if len(h.group_pubKey) > 0 {
		cpy.group_pubKey = make([]byte, len(h.group_pubKey))
		copy(cpy.group_pubKey, h.group_pubKey)
	}

	if len(h.header.TxRoot) > 0 {
		cpy.header.TxRoot.Set(h.header.TxRoot)
	}

	if len(h.header.TxsIllegal) > 0 {
		cpy.header.TxsIllegal = make([]uint16, 0)
		//for _, txsI := range h.TxsIllegal {
		//	cpy.TxsIllegal = append(cpy.TxsIllegal, txsI)
		//}
		cpy.header.TxsIllegal = append(cpy.header.TxsIllegal, h.header.TxsIllegal...)
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
			tx.CertId = pTx.CertId
			tx.Illegal = pTx.Illegal
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
	ReceivedAt   time.Time   `json:"received_at"`
	ReceivedFrom interface{} `json:"received_from"`
}

func (h *Header) GetAssetId() AssetId {
	return h.header.Number.AssetID
}

func (unit *Unit) GetAssetId() AssetId {
	return unit.UnitHeader.GetAssetId()
}
func (h *Header) Extra() []byte {
	return h.header.Extra
}
func (h *Header) GetAuthors() Authentifier {
	if h == nil {
		log.Error("the Unit Header pointer is nil!")
	}
	return h.header.Authors
}
func (h *Header) GetTxsIllegal() []uint16 {
	if h.header == nil {
		log.Error("the Unit Header pointer is nil!")
	}
	return h.header.TxsIllegal
}
func (h *Header) Author() common.Address {
	if h == nil {
		log.Error("the Unit Header pointer is nil!")
	}
	return crypto.PubkeyBytesToAddress(h.header.Authors.PubKey)
}

func (unit *Unit) Author() common.Address {
	if unit == nil {
		log.Error("the Unit pointer is nil!")
	}
	return unit.UnitHeader.Author()
}

func (unit *Unit) GetGroupPubKey() (kyber.Point, error) {
	return unit.UnitHeader.GetGroupPubKey()
}

func (unit *Unit) GetGroupPubKeyByte() []byte {
	return unit.UnitHeader.GetGroupPubKeyByte()
}

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

//出于DAG和基于Token的分区共识的考虑，设计了该ChainIndex，
type ChainIndex struct {
	AssetID AssetId `json:"asset_id"`
	Index   uint64  `json:"index"`
}
type ChainIndexTemp struct {
	AssetID AssetId `json:"asset_id"`
	Index   uint64  `json:"index"`
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
	binary.LittleEndian.PutUint64(idx, height.Index)
	return append(idx, height.AssetID.Bytes()...)
}

func (height *ChainIndex) SetBytes(data []byte) {
	height.Index = binary.LittleEndian.Uint64(data[:8])
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
	return u.UnitHeader
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
	//headerHash := u.UnitHeader.Hash()
	if u.UnitHash == (common.Hash{}) {
		//u.UnitHash = common.Hash{}
		u.UnitHash.Set(u.UnitHeader.Hash())
	}
	return u.UnitHash
}

// function Size, return the unit's StorageSize.
func (u *Unit) Size() common.StorageSize {
	if u.UnitSize > 0 {
		return u.UnitSize
	}
	emptyUnit := &Unit{}
	emptyUnit.UnitHeader = u.UnitHeader
	//emptyUnit.UnitHeader.Authors = nil
	emptyUnit.UnitHeader.group_sign = make([]byte, 0)
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
	return u.UnitHeader.GetNumber()
}

func (h *Header) GetNumber() *ChainIndex {
	return h.header.Number
}

func (u *Unit) NumberU64() uint64 {
	return u.UnitHeader.NumberU64()
}

func (u *Unit) Timestamp() int64 {
	return u.UnitHeader.Timestamp()
}

// return unit's parents UnitHash
func (u *Unit) ParentHash() []common.Hash {
	return u.UnitHeader.ParentHash()
}

func (h *Header) ParentHash() []common.Hash {
	return h.header.ParentsHash
}
func (h *Header) Header() *header_sdw {
	if h.header == nil {
		h.header = new_header_sdw()
	}
	return h.header
}
func (h *Header) SetGroupSign(sign []byte) {
	if len(sign) > 0 {
		h.group_sign = make([]byte, 0)
		h.group_sign = append(h.group_sign, sign...)
	}
}

func (u *Unit) GetGroupSign() []byte {
	return u.UnitHeader.GetGroupSign()
}

func (h *Header) GetGroupSign() []byte {
	return h.group_sign
}
func (h *Header) SetGroupPubkey(key []byte) {
	if len(key) > 0 {
		h.group_pubKey = make([]byte, 0)
		h.group_pubKey = append(h.group_pubKey, key...)
	}
}

func (u *Unit) GetGroupPubkey() []byte {
	return u.UnitHeader.GetGroupPubkey()
}

func (h *Header) GetGroupPubkey() []byte {
	return h.group_pubKey
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
	for _, hash := range u.UnitHeader.header.ParentsHash {
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

func (input *ChainIndex) DecodeRLP(s *rlp.Stream) error {
	raw, err := s.Raw()
	if err != nil {
		return err
	}
	temp := &ChainIndexTemp{}
	err = rlp.DecodeBytes(raw, temp)
	if err != nil {
		return err
	}

	input.AssetID = temp.AssetID
	input.Index = temp.Index

	return nil
}
func (input *ChainIndex) EncodeRLP(w io.Writer) error {
	temp := &ChainIndexTemp{}
	temp.AssetID = input.AssetID
	temp.Index = input.Index

	return rlp.Encode(w, temp)
}
