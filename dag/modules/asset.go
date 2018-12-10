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

import "strings"
import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/martinlindhe/base36"
	"github.com/palletone/go-palletone/common/rlp"
)

//Asset to identify token
//By default, system asset id=0,UniqueId=0
//默认的PTN资产，则AssetId=0，UniqueId=0
type Asset struct {
	//AssetId 资产类别,前26bit是symbol的base36编码，27-29是Symbol编码后字节长度，30-32bit为AssetType，剩下的是Txid的前12字节
	AssetId  IDType16 `json:"asset_id"`
	UniqueId IDType16 `json:"unique_id"` // every token has its unique id
	//ChainId  uint64   `json:"chain_id"`  // main chain id or sub-chain id,read from toml config NetworkId
}
type AssetType byte

const (
	AssetType_FungibleToken AssetType = iota
	AssetType_NonFungibleToken
	AssetType_VoteToken
)

func NewPTNAsset() *Asset {
	//return &Asset{AssetId: PTNCOIN}
	asset, err := NewAsset("PTN", AssetType_FungibleToken, 8, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, IDType16{})
	if err != nil {
		return nil
	}
	return asset
}
func NewAsset(symbol string, assetType AssetType, decimal byte, requestId []byte, uniqueId IDType16) (*Asset, error) {
	asset := &Asset{}
	assetId, err := NewAssetId(symbol, assetType, decimal, requestId)
	if err != nil {
		return nil, err
	}
	asset.AssetId = assetId
	asset.UniqueId = uniqueId
	return asset, nil
}

func NewPTNIdType() IDType16 {
	ptn, _ := NewAssetId("PTN", AssetType_FungibleToken, 8, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	return ptn
}

func NewAssetId(symbol string, assetType AssetType, decimal byte, requestId []byte) (IDType16, error) {
	if len(symbol) > 5 {
		return IDType16{}, errors.New("Symbol must less than 5 characters")
	}
	assetId := IDType16{}
	assetSymbol := base36.DecodeToBytes(symbol)
	//fmt.Printf(base36.EncodeBytes(assetSymbol))
	copy(assetId[4-len(assetSymbol):4], assetSymbol)
	firstByte := assetId[0] | (byte(len(assetSymbol) << 5))
	firstByte = firstByte | byte(assetType)<<2
	assetId[0] = firstByte
	assetId[4] = decimal
	copy(assetId[5:], requestId[0:11])
	return assetId, nil
}

func (asset *Asset) String() string {
	if asset.AssetId == PTNCOIN {
		return "PTN"
	}
	//symbol, t, tx12 := asset.AssetId.ParseAssetId()
	assetIdStr := asset.AssetId.ToAssetId()
	if asset.UniqueId == ZeroIdType16() {
		return assetIdStr
	}

	return fmt.Sprintf("%s-%s", assetIdStr, asset.UniqueId.String())
}

func (asset *Asset) SetString(str string) error {
	if str == "PTN" {
		asset.AssetId = PTNCOIN
		return nil
	}
	if !strings.Contains(str, "-") {
		//ERC20, AssetID only
		a, err := string2AssetId(str)
		if err != nil {
			return err
		}
		asset.AssetId = a
	} else {
		//ERC721
		strArray := strings.Split(str, "-")
		a, err := string2AssetId(strArray[0])
		if err != nil {
			return err
		}
		asset.AssetId = a
		uniqueId, err := hex.DecodeString(strArray[1])
		if err != nil {
			return err
		}
		asset.UniqueId.SetBytes(uniqueId)
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
