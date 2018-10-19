/*
   This file is part of go-palletone.
   go-palletone is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.
   go-palletone is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.
   You should have received a copy of the GNU General Public License
   along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
*/

/*
 * @author PalletOne core developers <dev@pallet.one>
 * @date 2018
 */

package storage

import (
	"fmt"
	"testing"

	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/dag/modules"
)

func TestUnitNumberIndex(t *testing.T) {
	key1 := fmt.Sprintf("%s_%s_%d", modules.UNIT_NUMBER_PREFIX, modules.BTCCOIN.String(), 10000)
	key2 := fmt.Sprintf("%s_%s_%d", modules.UNIT_NUMBER_PREFIX, modules.PTNCOIN.String(), 678934)

	if key1 != "nh_btcoin_10000" {
		log.Debug("not equal.", key1)
	}
	if key2 != "nh_ptncoin_678934" {
		log.Debug("not equal.", key2)
	}
}
