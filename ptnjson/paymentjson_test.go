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
	"encoding/json"
	"github.com/shopspring/decimal"
	"testing"
)

type DecimalTest struct {
	Amount   decimal.Decimal
	LockTime uint32
}

func TestDecimalJson(t *testing.T) {
	d := DecimalTest{Amount: decimal.RequireFromString("123.456"), LockTime: 123}
	js, _ := json.Marshal(d)
	t.Log(string(js))
	jsonStr := []byte("{\"Amount\":\"666.456\",\"LockTime\":123}")
	d2 := &DecimalTest{}
	json.Unmarshal(jsonStr, d2)
	t.Logf("%+v", d2)
}
