/*
 *
 *    This file is part of go-palletone.
 *    go-palletone is free software: you can redistribute it and/or modify
 *    it under the terms of the GNU General Public License as published by
 *    the Free Software Foundation, either version 3 of the License, or
 *    (at your option) any later version.
 *    go-palletone is distributed in the hope that it will be useful,
 *    but WITHOUT ANY WARRANTY; without even the implied warranty of
 *    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *    GNU General Public License for more details.
 *    You should have received a copy of the GNU General Public License
 *    along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
 * /
 *
 *  * @author PalletOne core developer <dev@pallet.one>
 *  * @date 2018
 *
 */

package modules

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"

	"github.com/martinlindhe/base36"
	"github.com/palletone/go-palletone/common/bitutil"
)

var (
	TimeFormatString = "2006/01/02 15:04:05"
	PTNCOIN          = AssetId{0x40, 0x00, 0x82, 0xBB, 0x08, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00}
	BTCCOIN          = AssetId{'b', 't', 'c', 'c', 'o', 'i', 'n'}
)

const (
	ID_LENGTH = 16
)

//AssetId 资产类别,前26bit是symbol的base36编码，27-29是Symbol编码后字节长度，30-32bit为AssetType，剩下的是Txid的前12字节
type AssetId [ID_LENGTH]byte

func ZeroIdType16() AssetId {
	return AssetId{}
}

func (it AssetId) String() string {
	if it == PTNCOIN {
		return "PTN"
	}
	symbol, assetType, decimal, txHash, uidType := it.ParseAssetId()
	if bitutil.IsZero(txHash) { //不是合约创建的Token，而是创世单元生成的Token
		if assetType == AssetType_FungibleToken && decimal == 8 {
			return symbol
		}
	}

	type2 := byte(assetType)<<3 | byte(uidType)
	rst := symbol + "+" + base36.EncodeBytes([]byte{decimal})
	rst += base36.EncodeBytes([]byte{type2})
	rst += base36.EncodeBytes(txHash)
	return rst
}

func String2AssetId(str string) (AssetId, UniqueIdType, error) {
	str = strings.ToUpper(str)
	if str == "PTN" {
		return PTNCOIN, UniqueIdType_Null, nil
	}
	strArray := strings.Split(str, "+")
	if len(strArray) < 2 {
		asset, err := NewAssetId(strArray[0], AssetType_FungibleToken, 8,
			[]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, UniqueIdType_Null)
		return asset, UniqueIdType_Null, err
	}
	symbol := strArray[0]
	type2 := base36.DecodeToBytes(string(strArray[1][1]))[0]
	assetType := AssetType(type2 >> 3)
	uniqueIdType := UniqueIdType(type2 & 7)
	decimal := base36.DecodeToBytes(strArray[1][0:1])
	tx12 := base36.DecodeToBytes(strArray[1][2:])
	assetId, err := NewAssetId(symbol, assetType, decimal[0], tx12, uniqueIdType)
	return assetId, uniqueIdType, err
}

func NewAssetId(symbol string, assetType AssetType, decimal byte, requestId []byte,
	uniqueIdType UniqueIdType) (AssetId, error) {
	if len(symbol) > 5 || len(symbol) == 0 {
		return AssetId{}, errors.New("Symbol must less than 5 characters")
	}
	if decimal > 18 {
		return AssetId{}, errors.New("Decimal must less than 19")
	}
	if len(requestId) < 11 {
		return AssetId{}, errors.New("requestId must more than 10")
	}
	assetId := AssetId{}
	assetSymbol := base36.DecodeToBytes(symbol)
	copy(assetId[4-len(assetSymbol):4], assetSymbol)
	firstByte := assetId[0] | (byte(len(assetSymbol) << 5))
	firstByte = firstByte | byte(assetType)<<2
	assetId[0] = firstByte
	assetId[4] = byte(uniqueIdType)<<5 | decimal
	copy(assetId[5:], requestId[0:11])
	return assetId, nil
}

func (id AssetId) ParseAssetId() (string, AssetType, byte, []byte, UniqueIdType) {
	var assetId [16]byte
	copy(assetId[:], id[:])
	assetId0 := id[0]
	len := assetId0 >> 5
	t := (assetId0 & 0xc) >> 2
	assetId[0] = assetId0 & 3
	symbol := base36.EncodeBytes(assetId[4-len : 4])
	return symbol, AssetType(t), assetId[4] & 0x1f, assetId[5:], UniqueIdType(assetId[4] >> 5)
}
func (id AssetId) GetSymbol() string {
	var assetId [16]byte
	copy(assetId[:], id[:])
	assetId0 := id[0]
	len := assetId0 >> 5
	assetId[0] = assetId0 & 3
	symbol := base36.EncodeBytes(assetId[4-len : 4])
	return symbol
}
func (id AssetId) GetAssetType() AssetType {
	t := (id[0] & 0xc) >> 2
	return AssetType(t)
}
func (id AssetId) GetDecimal() byte {
	return id[4] & 0x1f
}
func (id AssetId) ToAsset() *Asset {
	return &Asset{AssetId: id}
}
func (asset AssetId) Equal(another AssetId) bool {
	return bytes.Equal(asset.Bytes(), another.Bytes())
}

func (it AssetId) Bytes() []byte {
	return it[:]
}

func (it *AssetId) SetBytes(b []byte) {
	if len(b) > len(it) {
		b = b[len(b)-ID_LENGTH:]
	}

	copy(it[ID_LENGTH-len(b):], b)
}

func SetIdTypeByHex(id string) (AssetId, error) {
	bytes, err := hex.DecodeString(id)
	if err != nil {
		return AssetId{}, err
	}
	var id_type AssetId
	copy(id_type[0:], bytes)
	return id_type, nil
}

func (assetId AssetId) MarshalJSON() ([]byte, error) {
	return json.Marshal(assetId.String())
}

func (assetId *AssetId) UnmarshalJSON(data []byte) error {
	var str string
	err := json.Unmarshal(data, &str)
	if err != nil {
		return err
	}
	a, _, err := String2AssetId(str)
	if err != nil {
		return err
	}
	assetId.SetBytes(a.Bytes())
	return nil
}
