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
 *  * @date 2018-2019
 *
 */

package modules

import (
	"fmt"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/shopspring/decimal"
)

//金额和资产
type AmountAsset struct {
	Amount uint64 `json:"amount"`
	Asset  *Asset `json:"asset"`
}

func NewAmountAsset(amount uint64, asset *Asset) *AmountAsset {
	return &AmountAsset{Amount: amount, Asset: asset}
}
func (aa *AmountAsset) Bytes() []byte {
	data, _ := rlp.EncodeToBytes(aa)
	return data
}
func BytesToAmountAsset(b []byte) *AmountAsset {
	var a AmountAsset
	a.SetBytes(b)
	return &a
}
func (aa *AmountAsset) SetBytes(data []byte) error {
	err := rlp.DecodeBytes(data, aa)
	return err
}
func (aa *AmountAsset) String() string {

	number := assetAmt2DecimalAmt(aa.Asset, aa.Amount)
	return number.String() + aa.Asset.String()
}
func assetAmt2DecimalAmt(asset *Asset, amount uint64) decimal.Decimal {
	dec := asset.GetDecimal()
	d, _ := decimal.NewFromString(fmt.Sprintf("%d", amount))
	for i := 0; i < int(dec); i++ {
		d = d.Div(decimal.New(10, 0))
	}
	return d
}
