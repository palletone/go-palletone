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
 * @author PalletOne core developer Albert·Gou <dev@pallet.one>
 * @date 2018
 *
 */

package core

import (
	"testing"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/stretchr/testify/assert"
)

func Test_ChainParameters_Rlp(t *testing.T) {
	cp := NewChainParams()
	data, err := rlp.EncodeToBytes(&cp)
	assert.Nil(t, err)
	t.Logf("%x", data)

	cp2 := &ChainParameters{}
	err = rlp.DecodeBytes(data, cp2)
	assert.Nil(t, err)
	// assert.Equal(t, cp.TxCoinYearRate, cp2.TxCoinYearRate)
	assert.Equal(t, cp.ContractElectionNum, cp2.ContractElectionNum)
	assert.Equal(t, cp.UccCpuShares, cp2.UccCpuShares)
}
