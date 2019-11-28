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

package storage

import (
	"testing"

	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/stretchr/testify/assert"
)

var tplId1 = []byte("1234567")

func TestStatedb_SaveAndGetTemplate(t *testing.T) {
	tpl := &modules.ContractTemplate{TplId: tplId1, TplName: "Test", TplDescription: "Descr", Version: "v1.0", Abi: "abi content", Language: "go"}
	db, _ := ptndb.NewMemDatabase()
	statedb := NewStateDb(db)
	err := statedb.SaveContractTpl(tpl)
	assert.Nil(t, err)
	dbTemplate, err := statedb.GetContractTpl(tplId1)
	assert.Nil(t, err)
	assert.NotNil(t, dbTemplate)
	assert.Equal(t, tpl.TplName, dbTemplate.TplName)
}
