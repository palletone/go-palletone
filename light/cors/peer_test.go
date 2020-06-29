/*
 *  This file is part of go-palletone.
 *  go-palletone is free software: you can redistribute it and/or modify
 *  it under the terms of the GNU General Public License as published by
 *  the Free Software Foundation, either version 3 of the License, or
 *  (at your option) any later version.
 *  go-palletone is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *  GNU General Public License for more details.
 *  You should have received a copy of the GNU General Public License
 *  along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
 *
 *  @author PalletOne core developer <dev@pallet.one>
 *  @date 2018-2020
 */

package cors

import (
	"testing"

	"github.com/palletone/go-palletone/dag/modules"
	"github.com/stretchr/testify/assert"
)

func TestHandshakeRlp(t *testing.T) {
	number := modules.NewChainIndex(modules.PTNCOIN, 12345)
	var send keyValueList
	send = send.add("headNum", number)
	t.Logf("*ChainIndex:%x", send[0].Value)
	assert.NotNil(t, send[0].Value)
	send = send.add("headNum", *number)
	t.Logf("ChainIndex:%x", send[1].Value)
}
