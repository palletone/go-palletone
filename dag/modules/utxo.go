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

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/util"
	"github.com/palletone/go-palletone/dag/constants"
	"time"
)

var DAO uint64 = 100000000

type txoFlags uint8

const (
	tfCoinBase txoFlags = 1 << iota

	tfSpent

	tfModified
)

type Utxo struct {
	Amount   uint64 `json:"amount"`    // 数量
	Asset    *Asset `json:"asset"`     // 资产类别
	PkScript []byte `json:"pk_script"` // 锁定脚本
	LockTime uint32 `json:"lock_time"`
	//VoteResult common.Address `json:"vote_info"` //这个字段删掉
	// flags contains additional info about output such as whether it is spent, and whether is has
	// been modified since is was loaded.
	Timestamp uint64 `json:"timestamp"` //Unit's Timestamp
	Flags     txoFlags
}

func NewUtxo(output *Output, lockTime uint32, timestamp int64) *Utxo {
	return &Utxo{
		Amount:    output.Value,
		Asset:     output.Asset,
		PkScript:  output.PkScript,
		LockTime:  lockTime,
		Timestamp: uint64(timestamp),
	}
}
func (u *Utxo) GetTimestamp() int64 {
	return int64(u.Timestamp)
}
func (u *Utxo) Bytes() []byte {
	data, _ := rlp.EncodeToBytes(u)
	return data
}
func (utxo *Utxo) GetCoinDays() uint64 {
	if utxo.Timestamp == 0 {
		return 0
	}
	holdSecond := time.Now().Unix() - utxo.GetTimestamp()

	holdDays := holdSecond / 86400 //24*60*60
	return uint64(holdDays) * utxo.Amount
}

type UtxoWithOutPoint struct {
	*Utxo
	OutPoint
}

func NewUtxoWithOutPoint(utxo *Utxo, o OutPoint) *UtxoWithOutPoint {
	uto := &UtxoWithOutPoint{Utxo: utxo, OutPoint: o}
	return uto
}
func (u *UtxoWithOutPoint) GetAmount() uint64 {
	return u.Amount
}
func (u *UtxoWithOutPoint) Set(utxo *Utxo, o *OutPoint) {
	u.Utxo = utxo
	u.OutPoint = *o
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
		PkScript:  utxo.PkScript,
		Asset:     utxo.Asset,
		Amount:    utxo.Amount,
		LockTime:  utxo.LockTime,
		Flags:     utxo.Flags,
		Timestamp: utxo.Timestamp,
	}
}
func (utxo *Utxo) Flag2Str() string {
	return UtxoFlags2String(utxo.Flags)
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
		constants.UTXO_INDEX_PREFIX,
		utxoIndex.AccountAddr.String(),
		utxoIndex.Asset.String())
	return []byte(key)
}

func (utxoIndex *UtxoIndex) AccountKey() []byte {
	key := fmt.Sprintf("%s%s",
		constants.UTXO_INDEX_PREFIX,
		utxoIndex.AccountAddr.String())
	return []byte(key)
}

func (utxoIndex *UtxoIndex) QueryFields(key []byte) error {
	preLen := len(constants.UTXO_INDEX_PREFIX)
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
	key := append(constants.UTXO_INDEX_PREFIX, utxoIndex.AccountAddr.Bytes()...)
	key = append(key, utxoIndex.Asset.Bytes()...)
	key = append(key, utxoIndex.OutPoint.Bytes()...)
	return key[:]
	//	key := fmt.Sprintf("%s%s||%s||%s",
	//	constants.UTXO_INDEX_PREFIX,
	//	utxoIndex.AccountAddr.String(),
	//	utxoIndex.Asset.String(),
	//	utxoIndex.OutPoint.String())
	//return []byte(key)
}

func (outpoint *OutPoint) ToKey() []byte {
	// key: [UTXO_PREFIX][TxHash][MessageIndex][OutIndex]
	key := append(constants.UTXO_PREFIX, outpoint.Bytes()...)
	//key = append(key, common.EncodeNumberUint32(outpoint.MessageIndex)...)
	//key = append(key, common.EncodeNumberUint32(outpoint.OutIndex)...)
	return key[:]

}

//func (outpoint *OutPoint) ToKeyStr() string {
//	b := outpoint.ToKey()
//	return string(b)
//}

func (outpoint *OutPoint) SetString(data string) error {
	rs := []rune(data)
	data = string(rs[len(constants.UTXO_PREFIX):])
	if err := rlp.DecodeBytes([]byte(data), outpoint); err != nil {
		return err
	}
	return nil
}

func (outpoint *OutPoint) Bytes() []byte {

	data := append(outpoint.TxHash.Bytes(), common.EncodeNumberUint32(outpoint.MessageIndex)...)
	data = append(data, common.EncodeNumberUint32(outpoint.OutIndex)...)
	//data, err := rlp.EncodeToBytes(outpoint)
	//if err != nil {
	//	return nil
	//}
	return data
}
func (outpoint *OutPoint) Hash() common.Hash {
	v := util.RlpHash(outpoint)
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
	preLen := len(constants.UTXO_PREFIX)
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
	Value    uint64 `json:"value,string"`
	PkScript []byte `json:"pk_script"`
	Asset    *Asset `json:"asset"`
}

type Input struct {
	SignatureScript  []byte    `json:"signature_script"`
	Extra            []byte    `json:"extra" rlp:"nil"` // if user creating a new asset, this field should be it's config data. Otherwise it is null.
	PreviousOutPoint *OutPoint `json:"pre_outpoint"`
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
//type AssetInfo struct {
//	GasToken          string         `json:"alias"`           // asset name
//	AssetID        *Asset         `json:"asset_id"`        // asset id
//	InitialTotal   uint64         `json:"initial_total"`   // total circulation
//	Decimal        uint32         `json:"deciaml"`         // asset accuracy
//	DecimalUnit    string         `json:"unit"`            // asset unit
//	OriginalHolder common.Address `json:"original_holder"` // holder address when creating the asset
//}
//
//func (assetInfo *AssetInfo) Tokey() []byte {
//	key := fmt.Sprintf("%s%s",
//		constants.ASSET_INFO_PREFIX,
//		assetInfo.AssetID.AssetId.String())
//	return []byte(key)
//}
//
//func (assetInfo *AssetInfo) Print() {
//	fmt.Println("Asset alias", assetInfo.GasToken)
//	fmt.Println("Asset Assetid", assetInfo.AssetID.AssetId)
//	fmt.Println("Asset UniqueId", assetInfo.AssetID.UniqueId)
//	//fmt.Println("Asset ChainId", assetInfo.AssetID.ChainId)
//	fmt.Println("Asset Decimal", assetInfo.Decimal)
//	fmt.Println("Asset DecimalUnit", assetInfo.DecimalUnit)
//	fmt.Println("Asset OriginalHolder", assetInfo.OriginalHolder.String())
//}

//type AccountToken struct {
//	GasToken   string `json:"alias"`
//	AssetID *Asset `json:"asset_id"`
//	Balance uint64 `json:"balance"`
//}

func UtxoFlags2String(flag txoFlags) string {
	var str string
	switch flag {
	case tfCoinBase:
		str = "coin_base"
	case tfSpent:
		str = "spent"
	case tfModified:
		str = "modified"
	default:
		str = "normal"
	}
	return str
}
