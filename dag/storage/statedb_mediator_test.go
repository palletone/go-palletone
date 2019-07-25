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

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/core"
	"github.com/stretchr/testify/assert"
)

func TestSaveAndGetMediator(t *testing.T) {
	db := MockStateMemDb()
	addr, _ := common.StringToAddress("P1JJkSss3dEPCBJD6V759MgkB4vtj2EiUFX")
	point, _ := core.StrToPoint("Dsn4gF2xpsM79R6kBfsR1joZD4BoPfBGREJGStCAz1bFfUnB5QXBGbNfudxyCWz6uWZ" +
		"Z8c43BYWkxiezyF5uifhv1diiykrxzgFhLMSAvppx34RjJwzjmXAXnYMuQX3Jy2P3ygehcKmATAyXQCVoXde6Xo3tkA2Jv8Zb8zDcdGjbFyd")
	node, _ := core.StrToMedNode("pnode://f056aca66625c286ae444add82f44b9eb74f18a8a965723" +
		"60cb70df9b6d64d9bd2c58a345e570beb2bcffb037cd0a075f548b73083d31c12f1f4564865372534@127.0.0.1:30303")

	med := core.NewMediator()
	med.Address = addr
	med.InitPubKey = point
	med.Node = node

	err := db.StoreMediator(med)
	assert.Nil(t, err)
	dbMed, err := db.RetrieveMediator(addr)
	assert.Nil(t, err)
	assert.Equal(t, med, dbMed)
}
