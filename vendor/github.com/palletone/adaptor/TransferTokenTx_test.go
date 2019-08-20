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
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSimpleTransferTokenTx_String(t *testing.T) {
	tx := &SimpleTransferTokenTx{
		TxBasicInfo: TxBasicInfo{TxID: []byte("123456"), BlockHeight: 123, IsStable: true},
		FromAddress: "P15c2tpiRj7AZgQi3i8SHUZGwwDNF7zZSD8",
		ToAddress:   "P1NzevLMVCFJKWr4KAcHxyyh9xXaVU8yv3N",
		Amount:      NewAmountAssetUint64(1234, "BTC"),
		Fee:         nil,
		AttachData:  []byte("Hello"),
	}
	data, err := json.Marshal(tx)
	assert.Nil(t, err)
	t.Logf("Json tx:%s", string(data))
	newTx := SimpleTransferTokenTx{}
	err = json.Unmarshal(data, &newTx)
	assert.Nil(t, err)
	t.Logf("Unmarshal tx:%s", newTx.String())
}
