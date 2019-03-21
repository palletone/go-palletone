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

package ptnjson

import (
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAssetAmt2JsonAmt(t *testing.T) {
	asset, _ := modules.NewAsset("TEST", modules.AssetType_FungibleToken, 4, []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 11, 12, 13}, modules.UniqueIdType_Null, modules.UniqueId{})
	jsonAmt := AssetAmt2JsonAmt(asset, 12000)
	assert.Equal(t, jsonAmt.String(), "1.2")
	t.Log(jsonAmt)

	devin, _ := modules.StringToAsset("DEVIN+40U1RRBWTSW9LIXC58D")
	jsonAmt = AssetAmt2JsonAmt(devin, 1000000)
	t.Logf("DevinAmount:%s", jsonAmt.String())
}
func TestJsonAmt2AssetAmt(t *testing.T) {
	asset, _ := modules.NewAsset("TEST", modules.AssetType_FungibleToken, 5, []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 11, 12, 13}, modules.UniqueIdType_Null, modules.UniqueId{})

	jsonAmt, _ := decimal.NewFromString("1.2345")
	amt := JsonAmt2AssetAmt(asset, jsonAmt)
	assert.Equal(t, amt, uint64(123450))
	t.Log(amt)
}
