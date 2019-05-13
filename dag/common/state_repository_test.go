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

package common

import (
	"testing"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/stretchr/testify/assert"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/storage"
)

func TestStateRepository_GetContractsByTpl(t *testing.T) {
	db,_:=ptndb.NewMemDatabase()
	statedb:=storage.NewStateDb(db)
	rep:=NewStateRepository(statedb)
	tplId:=[]byte("111")
	c1:=&modules.Contract{ContractId:[]byte("1"),TemplateId:tplId,Name:"C1"}
	statedb.SaveContract(c1)
	c2:=&modules.Contract{ContractId:[]byte("2"),TemplateId:tplId,Name:"C2"}
	statedb.SaveContract(c2)

	c3:=&modules.Contract{ContractId:[]byte("3"),TemplateId:tplId,Name:"C3"}
	statedb.SaveContract(c3)

	contracts,err:= rep.GetContractsByTpl(tplId)
	assert.Nil(t,err)
	assert.Equal(t,3,len(contracts))
	for _,c:=range contracts{
		t.Logf("Contract:%#v",c)
	}
}
