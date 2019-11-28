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

package validator

import (
	"testing"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/coocood/freecache"
	"github.com/palletone/go-palletone/common"
	"github.com/stretchr/testify/assert"
	"encoding/json"
)

func TestValidatorCache_HasTxValidateResult(t *testing.T) {
	add:=[]*modules.Addition{}
	add=append(add,&modules.Addition{Amount:123,Asset:modules.NewPTNAsset()})
	cache:=NewValidatorCache(freecache.NewCache( 1024 * 1024))
	txId:=common.HexToHash("0x123456789")
	cache.AddTxValidateResult(txId,add)
	has,cadd:= cache.HasTxValidateResult(txId)
	assert.True(t,has)
	data,_:=json.Marshal(cadd)
	t.Logf("%s",(data))
	newId:=common.HexToHash("0x66666")
	has,cadd= cache.HasTxValidateResult(newId)
	assert.False(t,has)
	assert.Nil(t,cadd)
}
