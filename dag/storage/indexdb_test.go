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

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/stretchr/testify/assert"
)

func TestIndexDb_QueryProofOfExistenceByReference(t *testing.T) {
	ref := []byte("APP1-News-123")
	poe1 := &modules.ProofOfExistence{MainData: []byte("News Hash1"), ExtraData: []byte("News metadata json1"), Reference: ref, TxId: common.BytesToHash([]byte("txid")), UnitHash: common.Hash{}, Timestamp: 123}
	poe2 := &modules.ProofOfExistence{MainData: []byte("News Hash1"), ExtraData: []byte("News op1"), Reference: ref, TxId: common.BytesToHash([]byte("txid1")), UnitHash: common.Hash{}, Timestamp: 333}
	poe3 := &modules.ProofOfExistence{MainData: []byte("News Hash1"), ExtraData: []byte("News op2"), Reference: ref, TxId: common.BytesToHash([]byte("txid2")), UnitHash: common.Hash{}, Timestamp: 222}
	db, _ := ptndb.NewMemDatabase()
	idxdb := NewIndexDb(db)
	err := idxdb.SaveProofOfExistence(poe1)
	assert.Nil(t, err)
	idxdb.SaveProofOfExistence(poe2)
	idxdb.SaveProofOfExistence(poe3)
	result, err := idxdb.QueryProofOfExistenceByReference(ref)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(result))
	for _, poe := range result {
		t.Logf("%v", poe)
	}
}


