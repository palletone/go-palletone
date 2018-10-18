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
	"strings"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/rlp"
)

var DAO uint64 = 100000000

type txoFlags uint8

const (
	tfCoinBase txoFlags = 1 << iota

	tfSpent

	tfModified
)

//Asset to identify token
//By default, system asset id=0,UniqueId=0,ChainId=1
//默认的PTN资产，则AssetId=0，UniqueId=0,ChainId是当前链的ID
type Asset struct {
	AssetId  IDType16 `json:"asset_id"`  // 资产类别
	UniqueId IDType16 `json:"unique_id"` // every token has its unique id
	ChainId  uint64   `json:"chain_id"`  // main chain id or sub-chain id,read from toml config NetworkId
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
	if len(asset.AssetId) <= 0 || len(asset.UniqueId) <= 0 {
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

func (asset *Asset) IsSimilar(similar *Asset) bool {
	if !strings.EqualFold(asset.AssetId.String(), similar.AssetId.String()) {
		return false
	}
	if !strings.EqualFold(asset.UniqueId.String(), similar.UniqueId.String()) {
		return false
	}
	return true
}

type Utxo struct {
	Amount     uint64         `json:"amount"`    // 数量
	Asset      *Asset         `json:"Asset"`     // 资产类别
	PkScript   []byte         `json:"pk_script"` // 锁定脚本
	LockTime   uint32         `json:"lock_time"`
	VoteResult common.Address `json:"vote_info"` //edit by Yiran
	// flags contains additional info about output such as whether it is spent, and whether is has
	// been modified since is was loaded.
	Flags txoFlags
}

func (utxo *Utxo) StrPkScript() string {
	return fmt.Sprintf("%#x", utxo.PkScript)
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
func (utxo *Utxo) IsCoinBase() bool {
	return utxo.Flags&tfCoinBase == tfCoinBase
}
func (utxo *Utxo) Spend() {
	if utxo.IsSpent() {
		return
	}
	// Mark the output as spent and modified.
	utxo.Flags |= tfSpent | tfModified

}
func (utxo *Utxo) SetCoinBase() {
	utxo.Flags |= tfCoinBase
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
	Asset       *Asset
	OutPoint    *OutPoint
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
		return fmt.Errorf("Query UtxoIndex Asset Fields error: %s", err.Error())
	}
	if err := utxoIndex.OutPoint.SetString(sOutpoint); err != nil {
		return fmt.Errorf("Query UtxoIndex OutPoint Fields error: %s", err.Error())
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

func (outpoint *OutPoint) ToKey() []byte {
	// key: [UTXO_PREFIX][TxHash][MessageIndex][OutIndex]
	key := append(UTXO_PREFIX, outpoint.TxHash.Bytes()...)
	key = append(key, common.EncodeNumberUint32(outpoint.MessageIndex)...)
	key = append(key, common.EncodeNumberUint32(outpoint.OutIndex)...)
	return key[:]
	// out := fmt.Sprintf("%s%s%d_%d",
	// 	UTXO_PREFIX,
	// 	outpoint.TxHash.String(),
	// 	outpoint.MessageIndex,
	// 	outpoint.OutIndex,
	// )
	//  return []byte(out)
}

func (outpoint *OutPoint) SetString(data string) error {
	rs := []rune(data)
	data = string(rs[len(UTXO_PREFIX):])
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
func (outpoint *OutPoint) Hash() common.Hash {
	v := rlp.RlpHash(outpoint)
	return v
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

func KeyToOutpoint(key []byte) *OutPoint {
	// key: [UTXO_PREFIX][TxHash][MessageIndex][OutIndex]
	preLen := len(UTXO_PREFIX)
	sTxHash := key[preLen : len(key)-8]
	sMessage := key[(preLen + common.HashLength) : len(key)-4]
	sIndex := key[(preLen + common.HashLength + 4):]

	vout := new(OutPoint)
	vout.TxHash.SetBytes(sTxHash)
	vout.MessageIndex = common.DecodeNumberUint32(sMessage)
	vout.OutIndex = common.DecodeNumberUint32(sIndex)

	return vout
}

type Output struct {
	Value    uint64
	PkScript []byte
	Asset    *Asset
	Vote     common.Address // 投票结果
}
type Input struct {
	PreviousOutPoint *OutPoint
	SignatureScript  []byte
	Extra            []byte // if user creating a new asset, this field should be it's config data. Otherwise it is null.
}

// NewTxIn returns a new ptn transaction input with the provided
// previous outpoint point and signature script with a default sequence of
// MaxTxInSequenceNum.
func NewTxIn(prevOut *OutPoint, signatureScript []byte) *Input {
	return &Input{
		PreviousOutPoint: prevOut,
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
	AssetID        *Asset         `json:"asset_id"`        // asset id
	InitialTotal   uint64         `json:"initial_total"`   // total circulation
	Decimal        uint32         `json:"deciaml"`         // asset accuracy
	DecimalUnit    string         `json:"unit"`            // asset unit
	OriginalHolder common.Address `json:"original_holder"` // holder address when creating the asset
}

func (assetInfo *AssetInfo) Tokey() []byte {
	key := fmt.Sprintf("%s%s",
		ASSET_INFO_PREFIX,
		assetInfo.AssetID.AssetId.String())
	return []byte(key)
}

func (assetInfo *AssetInfo) Print() {
	fmt.Println("Asset alias", assetInfo.Alias)
	fmt.Println("Asset Assetid", assetInfo.AssetID.AssetId)
	fmt.Println("Asset UniqueId", assetInfo.AssetID.UniqueId)
	fmt.Println("Asset ChainId", assetInfo.AssetID.ChainId)
	fmt.Println("Asset Decimal", assetInfo.Decimal)
	fmt.Println("Asset DecimalUnit", assetInfo.DecimalUnit)
	fmt.Println("Asset OriginalHolder", assetInfo.OriginalHolder.String())
}

type AccountToken struct {
	Alias   string `json:"alias"`
	AssetID *Asset `json:"asset_id"`
	Balance uint64 `json:"balance"`
}
