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

package adaptor

import (
	"fmt"
	"math/big"
)

//AmountAsset Token的金额和资产标识
type AmountAsset struct {
	Amount big.Int `json:"amount"` //金额，最小单位
	Asset  string  `json:"asset"`  //资产标识
}

func (aa *AmountAsset) String() string {
	return fmt.Sprintf("%s %s", aa.Amount.String(), aa.Asset)
}
func NewAmountAsset(amount big.Int, asset string) *AmountAsset {
	return &AmountAsset{
		Amount: amount,
		Asset:  asset,
	}
}
func NewAmountAssetUint64(amount uint64, asset string) *AmountAsset {
	a := big.Int{}
	a.SetUint64(amount)
	return &AmountAsset{
		Amount: a,
		Asset:  asset,
	}
}
