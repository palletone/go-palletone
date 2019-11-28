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
 *  * @author PalletOne core developers <dev@pallet.one>
 *  * @date 2018-2019
 *
 *
 */

package modules

import (
	"testing"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common"
	"github.com/stretchr/testify/assert"
)

func TestCalcThreshold(t *testing.T) {
	t2 := calcThreshold(2)
	assert.Equal(t, 2, t2)
	t3 := calcThreshold(3)
	assert.Equal(t, 3, t3)
	t4 := calcThreshold(4)
	assert.Equal(t, 3, t4)
}

func TestGlobalProperty_Rlp(t *testing.T) {
	gp := NewGlobalProp()
	addr1, _ := common.StringToAddress("P1Kp2hcLhGEP45Xgx7vmSrE37QXunJUd8gj")
	addr2, _ := common.StringToAddress("P124gB1bXHDTXmox58g4hd4u13HV3e5vKie")
	gp.ChainParameters.MaximumMediatorCount = 21
	gp.ActiveMediators[addr1] = true
	gp.ActiveMediators[addr2] = false
	gp.ActiveJuries[addr1] = true
	data, err := rlp.EncodeToBytes(gp)
	assert.Nil(t, err)
	t.Logf("%x", data)

	gp2 := &GlobalProperty{}
	err = rlp.DecodeBytes(data, gp2)
	assert.Nil(t, err)
	assert.Equal(t, uint8(21), gp2.ChainParameters.MaximumMediatorCount)
	assert.Equal(t, 2, len(gp2.ActiveMediators))
}
