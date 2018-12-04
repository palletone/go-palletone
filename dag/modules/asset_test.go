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
	"github.com/martinlindhe/base36"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAsset_String(t *testing.T) {
	s := base36.DecodeToBytes("DEVIN")
	t.Logf("Data:%08b", s)
	t.Logf("Data:%08b", (byte(5)<<5)|s[0])
	//t.Logf("Data:%08b", base36.DecodeToBytes("00112"))
	//t.Logf("Data:%08b", base36.DecodeToBytes("ZZZZ"))
	//t.Logf("Data:%08b", base36.DecodeToBytes("ZZZ"))
	//t.Logf("Data:%08b", base36.DecodeToBytes("ZZ"))
	//t.Logf("Data:%08b", base36.DecodeToBytes("Z"))
	//symbol := base36.DecodeToBytes("Z")
	//id := IDType16{}
	//copy(id[4-len(symbol):4], symbol)
	//t.Logf("Data:%08b", id)
	asset, err := NewAsset("DEVIN", AssetType_FungibleToken, []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 11, 12, 13, 14, 15, 16}, IDType16{})
	assert.Nil(t, err)
	t.Log(asset.String())
	t.Logf("AssetId:%08b", asset.AssetId)
	asset2, err := NewAsset("ABC", AssetType_FungibleToken, []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 11, 12, 13, 14, 15, 16}, IDType16{})
	assetStr := asset2.String()
	t.Log(assetStr)
	t.Logf("AssetId:%08b", asset2.AssetId)
	a := Asset{}
	a.SetString(assetStr)
	t.Logf("Asset:%08b", a.AssetId)
	assert.Equal(t, asset2.Bytes(), a.Bytes())
}
func TestAsset_SetString(t *testing.T) {
	asset := &Asset{}
	asset.SetString("PTN")
	t.Logf("%08b", asset.AssetId)
	t.Logf("ptn string:%s", asset.String())
	assert.Equal(t, asset.String(), "PTN")
}
func TestPTNAsset(t *testing.T) {
	asset, err := NewAssetId("PTN", AssetType_FungibleToken, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	assert.Nil(t, err)
	t.Logf("PTN hex:%X", asset.Bytes())
	assert.Equal(t, asset, PTNCOIN)
}
