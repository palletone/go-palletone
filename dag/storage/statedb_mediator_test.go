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
 *  * @author PalletOne core developer  <dev@pallet.one>
 *  * @date 2018-2019
 *
 */
package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/common"
)

func TestSaveAndGetMediator(t *testing.T) {
	db := MockStateMemDb()
	addr,_:=common.StringToAddress("P1JJkSss3dEPCBJD6V759MgkB4vtj2EiUFX")
	point,_:=core.StrToPoint("gxjMgTWsVV7KVgSWHu6YiDQKknA58Sa4df23pTJphfgcLMqtHWdemW29BNEQFxjRvpjh7AhpW79sbju4DBQNVHBwhNwM9a624Qb4RdTYJd7RuaXgciJ2nFKDgSRRa351BhSXPyJiD96zoMub4rMVPEwXigYzvC7bPPFayAGxM9eQFUV")
	node,_:=core.StrToMedNode( "pnode://f056aca66625c286ae444add82f44b9eb74f18a8a96572360cb70df9b6d64d9bd2c58a345e570beb2bcffb037cd0a075f548b73083d31c12f1f4564865372534@127.0.0.1:30303")
	med:=&core.Mediator{Address:addr,InitPubKey:point,Node:node, MediatorApplyInfo:&core.MediatorApplyInfo{},MediatorInfoExpand:&core.MediatorInfoExpand{}}
	med.Content="Test Content"
	err:=db.StoreMediator(med)
	assert.Nil(t, err)
	dbMed,err:= db.RetrieveMediator(addr)
	assert.Nil(t,err)
	assert.Equal(t,med,dbMed)
}
