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
	"fmt"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/shopspring/decimal"
)

func Ptn2Dao(ptnAmount decimal.Decimal) uint64 {
	return uint64(ptnAmount.Mul(decimal.New(100000000, 0)).IntPart())
}
func Dao2Ptn(amount uint64) decimal.Decimal {
	d, _ := decimal.NewFromString(fmt.Sprintf("%d", amount))
	return d.Div(decimal.New(100000000, 0))
}
func AssetAmt2JsonAmt(asset *modules.Asset, amount uint64) decimal.Decimal {
	dec := asset.GetDecimal()
	return FormatAssetAmountByDecimal(amount, dec)
}
func FormatAssetAmountByDecimal(amount uint64, dec byte) decimal.Decimal {
	d, _ := decimal.NewFromString(fmt.Sprintf("%d", amount))
	for i := 0; i < int(dec); i++ {
		d = d.Div(decimal.New(10, 0))
	}
	return d
}
func JsonAmt2AssetAmt(asset *modules.Asset, jsonAmt decimal.Decimal) uint64 {
	dec := asset.GetDecimal()
	amt := jsonAmt
	for i := 0; i < int(dec); i++ {
		amt = amt.Mul(decimal.New(10, 0))
	}
	return uint64(amt.IntPart())
}
