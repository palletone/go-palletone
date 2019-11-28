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
package migration

import (
	"fmt"
	"testing"

	"github.com/palletone/go-palletone/common/hexutil"
)

func TestStrToHex(t *testing.T) {
	str := "testmigration_key"
	val := "testmigration_value"
	hex := hexutil.Encode([]byte(str))
	val_hex := hexutil.Encode([]byte(val))
	fmt.Println(hex)
	fmt.Println(val_hex)
}
