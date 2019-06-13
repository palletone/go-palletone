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
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAmountAsset_Bytes(t *testing.T) {
	aa := NewAmountAsset(123400000000, NewPTNAsset())
	t.Log(aa.String())
	data := aa.Bytes()
	t.Logf("%x", data)
	aa2 := BytesToAmountAsset(data)
	assert.Equal(t, aa, aa2)
}
func TestAmountAssetRlp(t *testing.T) {
	array := []AmountAsset{}
	data, err := rlp.EncodeToBytes(array)
	assert.Nil(t, err)
	t.Logf("%x", data)
}
