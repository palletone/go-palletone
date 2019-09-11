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
	"fmt"
	"strings"

	"bytes"
	"encoding/json"
	"github.com/palletone/go-palletone/dag/errors"
	"github.com/shopspring/decimal"
)

const PTN string = "PTN"

//Asset to identify token
//By default, system asset id=0,UniqueId=0
//默认的PTN资产，则AssetId=0，UniqueId=0
type Asset struct {
	AssetId  AssetId  `json:"asset_id"`
	UniqueId UniqueId `json:"unique_id"` // every token has its unique id
}

type AssetType byte

const (
	AssetType_FungibleToken AssetType = iota
	AssetType_NonFungibleToken
	AssetType_VoteToken
)

func NewPTNAsset() *Asset {
	asset, err := NewAsset(PTN, AssetType_FungibleToken, 8, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		UniqueIdType_Null, UniqueId{})
	if err != nil {
		return nil
	}
	return asset
}
func NewAsset(symbol string, assetType AssetType, decimal byte, requestId []byte,
	uidType UniqueIdType, uniqueId UniqueId) (*Asset, error) {
	asset := &Asset{}
	assetId, err := NewAssetId(symbol, assetType, decimal, requestId, uidType)
	if err != nil {
		return nil, err
	}
	asset.AssetId = assetId
	asset.UniqueId = uniqueId
	return asset, nil
}

func NewPTNIdType() AssetId {
	ptn, _ := NewAssetId(PTN, AssetType_FungibleToken, 8, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		UniqueIdType_Null)
	return ptn
}
func (asset *Asset) GetDecimal() byte {
	_, _, decim_code, _, _ := asset.AssetId.ParseAssetId()
	return decim_code
}
func (asset *Asset) String() string {
	if asset.AssetId == PTNCOIN {
		return PTN
	}
	_, t, _, _, uidType := asset.AssetId.ParseAssetId()
	assetIdStr := asset.AssetId.String()
	if t != AssetType_NonFungibleToken {
		return assetIdStr
	}

	return fmt.Sprintf("%s-%s", assetIdStr, asset.UniqueId.StringFriendly(uidType))
}
func StringToAsset(str string) (*Asset, error) {
	asset := &Asset{}
	err := asset.SetString(str)
	return asset, err
}
func (asset *Asset) SetString(str string) error {
	if str == PTN {
		asset.AssetId = PTNCOIN
		return nil
	}
	if !strings.Contains(str, "-") {
		//ERC20, AssetID only
		a, _, err := String2AssetId(str)
		if err != nil {
			return err
		}
		asset.AssetId = a
	} else {
		//ERC721
		strArray := strings.Split(str, "-")
		a, uniqueIdType, err := String2AssetId(strArray[0])
		if err != nil {
			return err
		}
		asset.AssetId = a
		uidStr := str[len(strArray[0])+1:]
		uniqueId, err := String2UniqueId(uidStr, uniqueIdType)
		if err != nil {
			return err
		}
		asset.UniqueId = uniqueId
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
	b := asset.AssetId.Bytes()
	return append(b, asset.UniqueId.Bytes()...)
}

func (asset *Asset) SetBytes(data []byte) error {
	if len(data) != 32 {
		return errors.New("data length not equal 32")
	}
	asset.AssetId.SetBytes(data[:16])
	asset.UniqueId.SetBytes(data[16:])
	return nil
}
func (asset *Asset) IsSameAssetId(another *Asset) bool {
	return bytes.Equal(asset.AssetId.Bytes(), another.AssetId.Bytes())
}
func (asset *Asset) Equal(another *Asset) bool {
	return bytes.Equal(asset.Bytes(), another.Bytes())
}
func (asset *Asset) IsSimilar(similar *Asset) bool {
	if !bytes.Equal(asset.AssetId.Bytes(), similar.AssetId.Bytes()) {
		return false
	}
	if !bytes.Equal(asset.UniqueId.Bytes(), similar.UniqueId.Bytes()) {
		return false
	}
	return true
}
func (asset *Asset) MarshalJSON() ([]byte, error) {
	return json.Marshal(asset.String())
}
func (asset *Asset) UnmarshalJSON(data []byte) error {
	var str string
	err := json.Unmarshal(data, &str)
	if err != nil {
		return err
	}
	asset.SetString(str)
	return nil
}
func (asset Asset) MarshalText() ([]byte, error) {
	return []byte(asset.String()), nil
}

func (asset *Asset) DisplayAmount(amount uint64) decimal.Decimal {
	dec := asset.GetDecimal()
	d, _ := decimal.NewFromString(fmt.Sprintf("%d", amount))
	for i := 0; i < int(dec); i++ {
		d = d.Div(decimal.New(10, 0))
	}
	return d
}
