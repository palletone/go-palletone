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
	"errors"
	"fmt"
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
	"sync/atomic"
)

// unit state
const (
	U_STATE_NO_GROUPSIGN = 0x20
	U_STATE_NO_CONFIRMED = 0x21
	U_STATE_CONFIRMED    = 0x22
)

type Header struct {
	header       *header_sdw
	hash         common.Hash
	group_sign   []byte // 群签名, 用于加快单元确认速度
	group_pubKey []byte // 群公钥, 用于验证群签名
}
type header_sdw struct {
	ParentsHash []common.Hash `json:"parents_hash"`
	Authors     Authentifier  `json:"mediator"` // the unit creation authors
	TxRoot      common.Hash   `json:"root"`
	TxsIllegal  []uint16      `json:"txs_illegal"` //Unit中非法交易索引
	Number      *ChainIndex   `json:"index"`
	Extra       []byte        `json:"extra"`
	Time        int64         `json:"creation_time"` // unit create time
	CryptoLib   []byte        `json:"crypto_lib"`    //该区块使用的加解密算法和哈希算法，0位表示非对称加密算法，1位表示Hash算法
}

func initHeaderSdw(parents []common.Hash, tx_root common.Hash, pubkey, sig, extra, crypto_lib []byte,
	txs_illgal []uint16, asset_id AssetId, index uint64, t int64) *header_sdw {
	h := header_sdw{ParentsHash: parents, TxRoot: tx_root, CryptoLib: crypto_lib, Extra: extra, Time: t}
	h.Authors.PubKey = pubkey
	h.Authors.Signature = sig
	h.Number = &ChainIndex{
		AssetID: asset_id,
		Index:   index,
	}
	h.TxsIllegal = txs_illgal
	return &h

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
func NewHeader(parents []common.Hash, tx_root common.Hash, pubkey, sig, extra, crypto_lib []byte,
	txs_illgal []uint16, asset_id AssetId, index uint64, t int64) *Header {
	// init header
	sdw := initHeaderSdw(parents, tx_root, pubkey, sig, extra, crypto_lib, txs_illgal, asset_id, index, t)
	h := Header{header: sdw}

	return &h
}
func (h *Header) Index() uint64 {
	if h.header == nil {
		log.Error("the Unit Header pointer is nil!")
	}
	return h.header.Number.Index
}
func (h *Header) ChainIndex() *ChainIndex {
	if h.header == nil {
		log.Error("the Unit Header pointer is nil!")
	}
	return h.header.Number
}
func (h *Header) TxRoot() common.Hash {
	if h.header == nil {
		log.Error("the Unit Header pointer is nil!")
	}
	return h.header.TxRoot
}
func (h *Header) NumberU64() uint64 {
	if h.header == nil {
		log.Error("the Unit Header pointer is nil!")
	}
	return h.header.Number.Index
}
func (h *Header) GetAssetId() AssetId {
	if h.header == nil {
		log.Error("the Unit Header pointer is nil!")
	}
	return h.header.Number.AssetID
}
func (h *Header) Extra() []byte {
	if h.header == nil {
		log.Error("the Unit Header pointer is nil!")
	}
	return h.header.Extra
}
func (h *Header) GetAuthors() Authentifier {
	if h.header == nil {
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
func (h *Header) Timestamp() int64 {
	if h.header == nil {
		log.Error("the Unit Header pointer is nil!")
	}
	return h.header.Time
}
func (h *Header) Cryptolib() []byte {
	if h.header == nil {
		log.Error("the Unit Header pointer is nil!")
	}
	return h.header.CryptoLib
}
func (h *Header) GetGroupPubKeyByte() []byte {
	return h.group_pubKey
}

func (h *Header) GetGroupPubkey() []byte {
	return common.CopyBytes(h.group_pubKey)
}

func (h *Header) SetGroupPubkey(key []byte) {
	h.group_pubKey = make([]byte, 0)
	if len(key) > 0 {
		h.group_pubKey = append(h.group_pubKey, key...)
	}
}

func (h *Header) ParentHash() []common.Hash {
	return h.header.ParentsHash
}

func (h *Header) SetGroupSign(sign []byte) {
	h.group_sign = make([]byte, 0)
	if len(sign) > 0 {
		h.group_sign = append(h.group_sign, sign...)
	}
}

func (h *Header) GetGroupSign() []byte {
	return common.CopyBytes(h.group_sign)
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
func (h *Header) ResetHash() {
	h.hash = common.Hash{}
}
func (h *Header) SetTxRoot(txroot common.Hash) {
	// init header
	//author := h.GetAuthors()
	//sdw := initHeaderSdw(h.header.ParentsHash, txroot, author.PubKey, author.Signature, h.Extra(),
	//	h.header.CryptoLib, h.header.TxsIllegal, h.GetNumber().AssetID, h.GetNumber().Index, h.Timestamp())
	//h.header = sdw
	h.header.TxRoot = txroot
	h.ResetHash()
}
func (h *Header) SetAuthor(author Authentifier) {
	//sdw := initHeaderSdw(h.header.ParentsHash, h.header.TxRoot, author.PubKey, author.Signature, h.Extra(),
	//	h.header.CryptoLib, h.header.TxsIllegal, h.GetNumber().AssetID, h.GetNumber().Index, h.Timestamp())
	//
	//h.header = sdw
	h.header.Authors = author
	h.ResetHash()
}

func (h *Header) SetTxsIllegal(txsillegal []uint16) {
	//author := h.GetAuthors()
	//sdw := initHeaderSdw(h.header.ParentsHash, h.header.TxRoot, author.PubKey, author.Signature, h.Extra(),
	//	h.header.CryptoLib, txsillegal, h.GetNumber().AssetID, h.GetNumber().Index, h.Timestamp())
	//
	//h.header = sdw
	h.header.TxsIllegal = txsillegal
	h.ResetHash()
}

func (h *Header) Hash() common.Hash {
	if h.hash == (common.Hash{}) {
		// 计算header’hash时 剔除群签
		groupSign := h.group_sign
		groupPubKey := h.group_pubKey
		h.group_sign = make([]byte, 0)
		h.group_pubKey = make([]byte, 0)
		h.hash = util.RlpHash(h)
		h.group_sign = append(h.group_sign, groupSign...)
		h.group_pubKey = append(h.group_pubKey, groupPubKey...)
	}
	return h.hash

}
func (h *Header) HashWithoutAuthor() common.Hash {
	groupSign := h.group_sign
	groupPubKey := h.group_pubKey
	author := h.header.Authors
	h.group_sign = make([]byte, 0)
	h.group_pubKey = make([]byte, 0)
	h.header.Authors = Authentifier{}
	hash := util.RlpHash(h)
	h.group_sign = append(h.group_sign, groupSign...)
	h.group_pubKey = append(h.group_pubKey, groupPubKey...)
	h.header.Authors.PubKey = author.PubKey[:]
	h.header.Authors.Signature = author.Signature[:]
	return hash
}

// HashWithOutTxRoot return  header's hash without txs root.
func (h *Header) HashWithOutTxRoot() common.Hash {
	groupSign := h.group_sign
	groupPubKey := h.group_pubKey
	txroot := h.header.TxRoot
	h.group_sign = make([]byte, 0)
	h.group_pubKey = make([]byte, 0)
	h.header.TxRoot = common.Hash{}

	hash := util.RlpHash(h)
	h.group_sign = append(h.group_sign, groupSign...)
	h.group_pubKey = append(h.group_pubKey, groupPubKey...)
	h.header.TxRoot = txroot
	return hash
}

func (h *Header) Size() common.StorageSize {
	header := h.header
	return common.StorageSize(unsafe.Sizeof(*header)) + common.StorageSize(len(header.Extra)/8)
}

// CopyHeader creates a deep copy of a block header to prevent side effects from
// modifying a header variable.

func (cpy *Header) CopyHeader(h *Header) {
	if cpy.header == nil {
		cpy.header = new_header_sdw()
	}
	author := h.GetAuthors()
	sdw := initHeaderSdw(h.ParentHash(), h.TxRoot(), author.PubKey, author.Signature, h.Extra(), h.Cryptolib(),
		h.GetTxsIllegal(), h.ChainIndex().AssetID, h.ChainIndex().Index, h.Timestamp())
	cpy.header = sdw
	cpy.group_sign = make([]byte, len(h.group_sign))
	if len(h.group_sign) > 0 {
		copy(cpy.group_sign, h.group_sign)
	}
	cpy.group_pubKey = make([]byte, len(h.group_pubKey))
	if len(h.group_pubKey) > 0 {
		copy(cpy.group_pubKey, h.group_pubKey)
	}

}
func CopyHeader(h *Header) *Header {
	if h == nil {
		return nil
	}
	author := h.GetAuthors()
	sdw := initHeaderSdw(h.ParentHash(), h.TxRoot(), author.PubKey, author.Signature, h.Extra(), h.Cryptolib(),
		h.GetTxsIllegal(), h.ChainIndex().AssetID, h.ChainIndex().Index, h.Timestamp())
	cpy := Header{header: sdw}

	if len(h.group_sign) > 0 {
		cpy.group_sign = make([]byte, len(h.group_sign))
		copy(cpy.group_sign, h.group_sign)
	}

	if len(h.group_pubKey) > 0 {
		cpy.group_pubKey = make([]byte, len(h.group_pubKey))
		copy(cpy.group_pubKey, h.group_pubKey)
	}

	return &cpy
}

func (u *Unit) CopyBody(txs Transactions) Transactions {
	if len(txs) > 0 {
		u.Txs = make([]*Transaction, len(txs))
		for i, pTx := range txs {
			msgs := pTx.TxMessages()
			sdw := transaction_sdw{}
			if len(msgs) > 0 {
				sdw.TxMessages = make([]*Message, len(msgs))
				copy(sdw.TxMessages, msgs)
			}
			sdw.CertId = pTx.CertId()
			sdw.Illegal = pTx.Illegal()
			u.Txs[i] = &Transaction{txdata: sdw}
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
	UnitHeader *Header      `json:"unit_header"`  // unit header
	Txs        Transactions `json:"transactions"` // transaction list
	unit_size  atomic.Value
	// These fields are used by package ptn to track
	// inter-peer block relay.
	ReceivedAt   time.Time   `json:"received_at"`
	ReceivedFrom interface{} `json:"received_from"`
}

func (unit *Unit) GetAssetId() AssetId {
	return unit.UnitHeader.GetAssetId()
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
		UnitHeader: header,
		Txs:        CopyTransactions(txs),
	}
	u.Size()
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
	return u.UnitHeader.Hash()
}
func (u *Unit) DisplayId() string {
	return fmt.Sprintf("%s-%d", u.Hash().String(), u.NumberU64())
}

// function Size, return the unit's StorageSize.
func (u *Unit) Size() common.StorageSize {
	if hash := u.unit_size.Load(); hash != nil {
		return hash.(common.StorageSize)
	}

	emptyUnit := &Unit{}
	emptyUnit.UnitHeader = new(Header)
	emptyUnit.UnitHeader.CopyHeader(u.Header())
	emptyUnit.UnitHeader.group_sign = make([]byte, 0)
	emptyUnit.UnitHeader.group_pubKey = make([]byte, 0)
	emptyUnit.CopyBody(u.Txs[:])

	b, err := rlp.EncodeToBytes(emptyUnit)
	if err != nil {
		log.Errorf("rlp encode Unit error:%s", err.Error())
		return common.StorageSize(0)
	} else {
		size := common.StorageSize(len(b))
		if len(b) > 0 {
			u.unit_size.Store(size)
		}
		return size
	}
}

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
func (u *Unit) GetGroupSign() []byte {
	return u.UnitHeader.GetGroupSign()
}
func (u *Unit) GetGroupPubkey() []byte {
	return u.UnitHeader.GetGroupPubkey()
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
	b.Size()
	return b
}

func (u *Unit) ContainsParent(pHash common.Hash) bool {
	for _, hash := range u.UnitHeader.header.ParentsHash {
		if pHash == hash {
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
	if input != nil {
		temp.AssetID = input.AssetID
		temp.Index = input.Index
		return rlp.Encode(w, temp)
	}
	return nil
}
