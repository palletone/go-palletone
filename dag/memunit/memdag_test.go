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

package memunit

import (
	"github.com/golang/mock/gomock"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/crypto"
	"github.com/palletone/go-palletone/common/ptndb"

	dagcommon "github.com/palletone/go-palletone/dag/common"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/storage"
	"github.com/palletone/go-palletone/dag/txspool"
	"github.com/stretchr/testify/assert"

	"crypto/ecdsa"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/core"
	"testing"
)

func TestMemDag_AddUnit(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	lastHeader := newTestUnit(common.Hash{}, 0, key1)
	txpool := txspool.NewMockITxPool(mockCtrl)
	db, _ := ptndb.NewMemDatabase()
	dagDb := storage.NewDagDb(db)
	utxoDb := storage.NewUtxoDb(db)
	stateDb := storage.NewStateDb(db)
	idxDb := storage.NewIndexDb(db)
	propDb := storage.NewPropertyDb(db)
	propDb.SetNewestUnit(lastHeader.UnitHeader)

	//utxoRep := dagcommon.NewUtxoRepository(utxoDb, idxDb, stateDb)
	unitRep := dagcommon.NewUnitRepository(dagDb, idxDb, utxoDb, stateDb, propDb)
	unitRep.SaveUnit(lastHeader, false)
	propRep := dagcommon.NewPropRepository(propDb)
	propRep.StoreGlobalProp(modules.NewGlobalProp())
	stateRep := dagcommon.NewStateRepository(stateDb)
	gasToken := modules.PTNCOIN
	memdag := NewMemDag(gasToken, 2, false, db, unitRep, propRep, stateRep)
	//tunitRep, tutxoRep, tstateRep := unstableChain.GetUnstableRepositories()

	err := memdag.AddUnit(newTestUnit(common.Hash{}, 0, key2), txpool)
	assert.Nil(t, err)
}
func BenchmarkMemDag_AddUnit(b *testing.B) {
	mockCtrl := gomock.NewController(b)
	defer mockCtrl.Finish()
	lastHeader := newTestUnit(common.Hash{}, 0, key1)
	txpool := txspool.NewMockITxPool(mockCtrl)
	db, _ := ptndb.NewMemDatabase()
	dagDb := storage.NewDagDb(db)
	utxoDb := storage.NewUtxoDb(db)
	stateDb := storage.NewStateDb(db)
	idxDb := storage.NewIndexDb(db)
	propDb := storage.NewPropertyDb(db)
	propDb.SetNewestUnit(lastHeader.UnitHeader)
	propDb.SetNewestUnit(lastHeader.Header())
	//utxoRep := dagcommon.NewUtxoRepository(utxoDb, idxDb, stateDb)
	unitRep := dagcommon.NewUnitRepository(dagDb, idxDb, utxoDb, stateDb, propDb)
	unitRep.SaveUnit(lastHeader, false)
	propRep := dagcommon.NewPropRepository(propDb)
	propRep.StoreGlobalProp(modules.NewGlobalProp())
	stateRep := dagcommon.NewStateRepository(stateDb)
	gasToken := modules.PTNCOIN
	memdag := NewMemDag(gasToken, 2, false, db, unitRep, propRep, stateRep)
	//tunitRep, tutxoRep, tstateRep := unstableChain.GetUnstableRepositories()
	parentHash := lastHeader.Hash()
	for i := 0; i < b.N; i++ {
		unit := newTestUnit(parentHash, uint64(i+1), key1)
		err := memdag.AddUnit(unit, txpool)
		assert.Nil(b, err)
		parentHash = unit.Hash()
	}
}
func newTestUnit(parentHash common.Hash, height uint64, key *ecdsa.PrivateKey) *modules.Unit {
	h := newTestHeader(parentHash, height, key)
	return modules.NewUnit(h, []*modules.Transaction{})
}

var (
	key1, _ = crypto.GenerateKey()
	addr1   = crypto.PubkeyToAddress(&key1.PublicKey)
	key2, _ = crypto.GenerateKey()
	addr2   = crypto.PubkeyToAddress(&key2.PublicKey)
	key3, _ = crypto.GenerateKey()
	key4, _ = crypto.GenerateKey()
)

func newTestHeader(parentHash common.Hash, height uint64, key *ecdsa.PrivateKey) *modules.Header {

	h := new(modules.Header)
	//h.AssetIDs = append(h.AssetIDs, PTNCOIN)
	au := modules.Authentifier{}

	h.GroupSign = []byte("group_sign")
	h.GroupPubKey = []byte("group_pubKey")
	h.Number = &modules.ChainIndex{}
	h.Number.AssetID = modules.PTNCOIN
	h.Number.Index = height
	h.Extra = make([]byte, 20)
	h.ParentsHash = []common.Hash{parentHash}
	//tr := common.Hash{}
	//tr = tr.SetString("c35639062e40f8891cef2526b387f42e353b8f403b930106bb5aa3519e59e35f")
	h.TxRoot = common.HexToHash("c35639062e40f8891cef2526b387f42e353b8f403b930106bb5aa3519e59e35f")

	sig, _ := crypto.Sign(h.TxRoot[:], key)
	au.Signature = sig
	au.PubKey = crypto.CompressPubkey(&key.PublicKey)
	h.Authors = au
	h.Time = int64(height) * 3
	return h
}

//添加一个正常Unit，最新单元会更新，添加孤儿Unit，最新单元不会更新；补上了孤儿遗失的单元，那么孤儿单元不再是孤儿，会加到链上
func TestMemDag_AddOrphanUnit(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	lastHeader := newTestUnit(common.Hash{}, 0, key1)
	txpool := txspool.NewMockITxPool(mockCtrl)
	db, _ := ptndb.NewMemDatabase()
	dagDb := storage.NewDagDb(db)
	utxoDb := storage.NewUtxoDb(db)
	stateDb := storage.NewStateDb(db)
	idxDb := storage.NewIndexDb(db)
	propDb := storage.NewPropertyDb(db)
	propDb.SetNewestUnit(lastHeader.Header())
	mockMediatorInit(stateDb, propDb)
	unitRep := dagcommon.NewUnitRepository(dagDb, idxDb, utxoDb, stateDb, propDb)
	unitRep.SaveUnit(lastHeader, false)
	propRep := dagcommon.NewPropRepository(propDb)
	stateRep := dagcommon.NewStateRepository(stateDb)
	gasToken := modules.PTNCOIN
	memdag := NewMemDag(gasToken, 2, false, db, unitRep, propRep, stateRep)
	u1 := newTestUnit(lastHeader.Hash(), 1, key2)
	log.Debugf("Try add unit[%x] to memdag", u1.Hash())
	err := memdag.AddUnit(u1, txpool)
	assert.Nil(t, err)
	assert.EqualValues(t, 1, memdag.GetLastMainChainUnit().NumberU64())

	u2 := newTestUnit(u1.Hash(), 2, key1)
	u3 := newTestUnit(u2.Hash(), 3, key2)
	log.Debugf("Try add orphan unit[%x] to memdag", u3.Hash())
	err = memdag.AddUnit(u3, txpool)
	assert.Nil(t, err)
	assert.EqualValues(t, 1, memdag.GetLastMainChainUnit().NumberU64())
	log.Debugf("Try add missed unit[%x] to memdag", u2.Hash())
	err = memdag.AddUnit(u2, txpool)
	assert.Nil(t, err)
	assert.EqualValues(t, 3, memdag.GetLastMainChainUnit().NumberU64())
}

//添加1,2单元后，再次添加2'最新单元不变，再添加3'则主链切换，最新单元更新为3'
func TestMemDag_SwitchMainChain(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	u0 := newTestUnit(common.Hash{}, 1, key1)
	txpool := txspool.NewMockITxPool(mockCtrl)
	db, _ := ptndb.NewMemDatabase()
	dagDb := storage.NewDagDb(db)
	utxoDb := storage.NewUtxoDb(db)
	stateDb := storage.NewStateDb(db)
	idxDb := storage.NewIndexDb(db)
	propDb := storage.NewPropertyDb(db)
	propDb.SetNewestUnit(u0.UnitHeader)
	mockMediatorInit(stateDb, propDb)
	unitRep := dagcommon.NewUnitRepository(dagDb, idxDb, utxoDb, stateDb, propDb)
	unitRep.SaveUnit(u0, false)
	propRep := dagcommon.NewPropRepository(propDb)
	stateRep := dagcommon.NewStateRepository(stateDb)
	gasToken := modules.PTNCOIN
	memdag := NewMemDag(gasToken, 2, false, db, unitRep, propRep, stateRep)
	u1 := newTestUnit(u0.Hash(), 2, key2)
	log.Debugf("Try add unit[%x] to memdag", u1.Hash())
	err := memdag.AddUnit(u1, txpool)
	assert.Nil(t, err)
	assert.EqualValues(t, 2, memdag.GetLastMainChainUnit().NumberU64())

	u22 := newTestUnit(u0.Hash(), 2, key1)
	log.Debugf("Try add side unit[%x] to memdag", u22.Hash())
	err = memdag.AddUnit(u22, txpool)
	assert.Nil(t, err)
	assert.EqualValues(t, u1.Hash(), memdag.GetLastMainChainUnit().Hash())

	u33 := newTestUnit(u22.Hash(), 3, key2)
	log.Debugf("Try add new longest chain unit[%x] to memdag", u33.Hash())

	err = memdag.AddUnit(u33, txpool)
	assert.Nil(t, err)
	assert.EqualValues(t, 3, memdag.GetLastMainChainUnit().NumberU64())
}

func mockMediatorInit(statedb storage.IStateDb, propDb storage.IPropertyDb) {
	point, _ := core.StrToPoint("Dsn4gF2xpsM79R6kBfsR1joZD4BoPfBGREJGStCAz1bFfUnB5QXBGbNfudxyCWz6uWZZ8c43BYWkxiezyF5uifhv1diiykrxzgFhLMSAvppx34RjJwzjmXAXnYMuQX3Jy2P3ygehcKmATAyXQCVoXde6Xo3tkA2Jv8Zb8zDcdGjbFyd")
	node, _ := core.StrToMedNode("pnode://f056aca66625c286ae444add82f44b9eb74f18a8a96572360cb70df9b6d64d9bd2c58a345e570beb2bcffb037cd0a075f548b73083d31c12f1f4564865372534@127.0.0.1:30303")
	m1 := &core.Mediator{Address: addr1, InitPubKey: point, Node: node, MediatorApplyInfo: &core.MediatorApplyInfo{}, MediatorInfoExpand: &core.MediatorInfoExpand{}}

	statedb.StoreMediator(m1)
	m2 := &core.Mediator{Address: addr2, InitPubKey: point, Node: node, MediatorApplyInfo: &core.MediatorApplyInfo{}, MediatorInfoExpand: &core.MediatorInfoExpand{}}
	statedb.StoreMediator(m2)
	gp := modules.NewGlobalProp()
	gp.ActiveMediators = make(map[common.Address]bool)
	gp.ActiveMediators[addr1] = true
	gp.ActiveMediators[addr2] = true
	gp.ChainParameters.MediatorInterval = 3
	propDb.StoreGlobalProp(gp)
	dgp := modules.NewDynGlobalProp()
	dgp.NextMaintenanceTime = 99999
	propDb.StoreDynGlobalProp(dgp)
	ms := modules.NewMediatorSchl()
	ms.CurrentShuffledMediators = append(ms.CurrentShuffledMediators, addr1)
	ms.CurrentShuffledMediators = append(ms.CurrentShuffledMediators, addr2)
	propDb.StoreMediatorSchl(ms)
}
