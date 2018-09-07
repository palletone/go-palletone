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

// Unspent Transaction Output module.
package modules

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/rlp"
)

type txoFlags uint8

const (
	tfSpent txoFlags = 1 << iota

	tfModified
)

var (
	// state storage
	CONTRACT_ATTRI    = []byte("contract") // like contract_[contract address]_[key]
	UTXO_PREFIX       = []byte("uo")
	UTXO_INDEX_PREFIX = []byte("ui")
	ASSET_INFO_PREFIX = []byte("ai")
)

//Asset to identify token
//By default, system asset id=0,UniqueId=0,ChainId=1
//默认的PTN资产，则AssetId=0，UniqueId=0,ChainId是当前链的ID
type Asset struct {
	AssertId IDType16 `json:"assert_id"` // 资产类别
	UniqueId IDType16 `json:"unique_id"` // every token has its unique id
	ChainId  uint64   `json:"chain_id"`  // main chain id or sub-chain id
}

func (asset *Asset) String() string {
	data, err := rlp.EncodeToBytes(asset)
	if err != nil {
		return ""
	}
	return string(data)
}

func (asset *Asset) SetString(data string) error {
	if err := rlp.DecodeBytes([]byte(data), asset); err != nil {
		return err
	}
	return nil
}

func (asset *Asset) IsEmpty() bool {
	if len(asset.AssertId) <= 0 || len(asset.UniqueId) <= 0 {
		return true
	}
	return false
}

func (asset *Asset) Bytes() []byte {
	data, err := rlp.EncodeToBytes(asset)
	if err != nil {
		return nil
	}
	return data
}

func (asset *Asset) SetBytes(data []byte) error {
	if err := rlp.DecodeBytes(data, asset); err != nil {
		return err
	}
	return nil
}

type Utxo struct {
	Amount   uint64 `json:"amount"`  // 数量
	Asset    *Asset `json:"Asset"`   // 资产类别
	PkScript []byte `json:"program"` // 要执行的代码段
	LockTime uint32 `json:"lock_time"`

	// falgs contains additional info about output such as whether it is spent, and whether is has
	// been modified since is was loaded.
	Flags txoFlags
}

func (utxo *Utxo) IsEmpty() bool {
	if len(utxo.PkScript) != 0 || utxo.Amount > 0 || utxo.LockTime > 0 || utxo.Asset != nil {
		return false
	}
	return true
}
func (utxo *Utxo) IsModified() bool {
	return utxo.Flags*tfModified == tfModified
}
func (utxo *Utxo) IsSpent() bool {
	return utxo.Flags&tfSpent == tfSpent
}
func (utxo *Utxo) Spend() {
	if utxo.IsSpent() {
		return
	}
	utxo.Flags |= tfSpent | tfModified
}
func (utxo *Utxo) Clone() *Utxo {
	if utxo == nil {
		return nil
	}
	return &Utxo{
		PkScript: utxo.PkScript,
		Asset:    utxo.Asset,
		Amount:   utxo.Amount,
		LockTime: utxo.LockTime,
		Flags:    utxo.Flags,
	}
}

// UtxoIndex is key
// utxo index db value: amount
type UtxoIndex struct {
	AccountAddr common.Address `json:"account_id"` // 所属人id
	Asset       Asset
	OutPoint    OutPoint
}

type UtxoIndexValue struct {
	Amount   uint64 `json:"amount"`
	LockTime uint32 `json:"lock_time"`
}

func (utxoIndex *UtxoIndex) AssetKey() []byte {
	key := fmt.Sprintf("%s%s||%s",
		UTXO_INDEX_PREFIX,
		utxoIndex.AccountAddr.String(),
		utxoIndex.Asset.String())
	return []byte(key)
}

func (utxoIndex *UtxoIndex) AccountKey() []byte {
	key := fmt.Sprintf("%s%s",
		UTXO_INDEX_PREFIX,
		utxoIndex.AccountAddr.String())
	//fmt.Println("Account prefix:", key)
	return []byte(key)
}

func (utxoIndex *UtxoIndex) QueryFields(key []byte) error {
	preLen := len(UTXO_INDEX_PREFIX)
	s := string(key[preLen:])
	ss := strings.Split(s, "||")
	if len(ss) != 3 {
		return fmt.Errorf("Query UtxoIndex Fields error: len=%d, ss=%v", len(ss), ss)
	}
	sAddr := ss[0]
	sAsset := ss[1]
	sOutpoint := ss[2]

	utxoIndex.AccountAddr.SetString(sAddr)
	if err := utxoIndex.Asset.SetString(sAsset); err != nil {
		return err
	}
	if err := utxoIndex.OutPoint.SetString(sOutpoint); err != nil {
		return err
	}
	return nil
}

func (utxoIndex *UtxoIndex) ToKey() []byte {
	key := fmt.Sprintf("%s%s||%s||%s",
		UTXO_INDEX_PREFIX,
		utxoIndex.AccountAddr.String(),
		utxoIndex.Asset.String(),
		utxoIndex.OutPoint.String())
	return []byte(key)
}

// utxo key
/*type OutPoint struct {
	TxHash       common.Hash // reference Utxo struct key field
	MessageIndex uint32      // message index in transaction
	OutIndex     uint32
}*/

func (outpoint *OutPoint) ToKey() []byte {
	out := fmt.Sprintf("%s%s_%d_%d",
		UTXO_PREFIX,
		outpoint.TxHash.String(),
		outpoint.MessageIndex,
		outpoint.OutIndex,
	)
	return []byte(out)
}

func (outpoint *OutPoint) String() string {
	data, err := rlp.EncodeToBytes(outpoint)
	if err != nil {
		return ""
	}
	return string(data)
}

func (outpoint *OutPoint) SetString(data string) error {
	if err := rlp.DecodeBytes([]byte(data), outpoint); err != nil {
		return err
	}
	return nil
}

func (outpoint *OutPoint) Bytes() []byte {
	data, err := rlp.EncodeToBytes(outpoint)
	if err != nil {
		return nil
	}
	return data
}

func (outpoint *OutPoint) IsEmpty() bool {
	emptyHash := common.Hash{}
	for i := 0; i < cap(emptyHash); i++ {
		emptyHash[i] = 0
	}
	if len(outpoint.TxHash) == 0 ||
		strings.Compare(outpoint.TxHash.String(), emptyHash.String()) == 0 {
		return true
	}
	return false
}

func KeyToOutpoint(key []byte) OutPoint {
	// key: [UTXO_PREFIX][TxHash]_[MessageIndex]_[OutIndex]
	preLen := len(UTXO_PREFIX)
	sKey := key[preLen:]
	sTxHash := sKey[:common.HashLength]
	sKey = sKey[common.HashLength:]

	data := strings.Split(string(sKey), "_")
	if len(data) != 2 {
		return OutPoint{}
	}

	var vout OutPoint
	vout.TxHash.SetBytes(sTxHash)
	i, err := strconv.Atoi(data[0])
	if err == nil {
		vout.MessageIndex = uint32(i)
	}

	i, err = strconv.Atoi(data[1])
	if err == nil {
		vout.OutIndex = uint32(i)
	}

	return vout
}

type Output struct {
	Value    uint64
	PkScript []byte
	Asset    Asset
}
type Input struct {
	PreviousOutPoint OutPoint
	SignatureScript  []byte
	Extra            []byte // if user creating a new asset, this field should be it's config data. Otherwise it is null.
}

// NewTxIn returns a new ptn transaction input with the provided
// previous outpoint point and signature script with a default sequence of
// MaxTxInSequenceNum.
func NewTxIn(prevOut *OutPoint, signatureScript []byte) *Input {
	return &Input{
		PreviousOutPoint: *prevOut,
		SignatureScript:  signatureScript,
	}
}

type SpendProof struct {
	Unit string `json:"unit"`
}

/**
保存Asset属性信息结构体
structure for saving asset property infomation
*/
type AssetInfo struct {
	Alias          string         `json:"alias"`           // asset name
	AssetID        Asset          `json:"asset_id"`        // asset id
	InitialTotal   uint64         `json:"initial_total"`   // total circulation
	Decimal        uint32         `json:"deciaml"`         // asset accuracy
	DecimalUnit    string         `json:"unit"`            // asset unit
	OriginalHolder common.Address `json:"original_holder"` // holder address when creating the asset
}

func (assetInfo *AssetInfo) Tokey() []byte {
	key := fmt.Sprintf("%s%s",
		ASSET_INFO_PREFIX,
		assetInfo.AssetID.AssertId.String())
	return []byte(key)
}

func (assetInfo *AssetInfo) Print() {
	fmt.Println("Asset alias", assetInfo.Alias)
	fmt.Println("Asset Assetid", assetInfo.AssetID.AssertId)
	fmt.Println("Asset UniqueId", assetInfo.AssetID.UniqueId)
	fmt.Println("Asset ChainId", assetInfo.AssetID.ChainId)
	fmt.Println("Asset Decimal", assetInfo.Decimal)
	fmt.Println("Asset DecimalUnit", assetInfo.DecimalUnit)
	fmt.Println("Asset OriginalHolder", assetInfo.OriginalHolder.String())
}

type AccountToken struct {
	Alias   string `json:"alias"`
	AssetID Asset  `json:"asset_id"`
	Balance uint64 `json:"balance"`
}
