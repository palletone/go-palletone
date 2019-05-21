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

	"testing"
	"github.com/palletone/go-palletone/common/log"
)

func TestMemDag_AddUnit(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	lastHeader := newTestUnit(common.Hash{},0)
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
	stateRep:=dagcommon.NewStateRepository(stateDb)
	gasToken := modules.PTNCOIN
	memdag := NewMemDag(gasToken, false, db, unitRep, propRep,stateRep)
	//tunitRep, tutxoRep, tstateRep := unstableChain.GetUnstableRepositories()

	err := memdag.AddUnit(newTestUnit(common.Hash{},0), txpool)
	assert.Nil(t, err)
}
func BenchmarkMemDag_AddUnit(b *testing.B) {
	mockCtrl := gomock.NewController(b)
	defer mockCtrl.Finish()
	lastHeader := newTestUnit(common.Hash{},0)
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
	stateRep:=dagcommon.NewStateRepository(stateDb)
	gasToken := modules.PTNCOIN
	memdag := NewMemDag(gasToken, false, db, unitRep, propRep,stateRep)
	//tunitRep, tutxoRep, tstateRep := unstableChain.GetUnstableRepositories()
	parentHash := lastHeader.Hash()
	for i := 0; i < b.N; i++ {
		unit := newTestUnit(parentHash, uint64(i+1))
		err := memdag.AddUnit(unit, txpool)
		assert.Nil(b, err)
		parentHash = unit.Hash()
	}
}
func newTestUnit(parentHash common.Hash,height uint64) *modules.Unit {
	h := newTestHeader(parentHash,height)
	return modules.NewUnit(h, []*modules.Transaction{})
}
var (
	key1, _ = crypto.GenerateKey()
	addr1=crypto.PubkeyToAddress(&key1.PublicKey)
	key2, _ = crypto.GenerateKey()
	addr2=crypto.PubkeyToAddress(&key2.PublicKey)
)
func newTestHeader(parentHash common.Hash,height uint64) *modules.Header {


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
	key, _ := crypto.GenerateKey()
	sig, _ := crypto.Sign(h.TxRoot[:], key)
	au.Signature = sig
	au.PubKey = crypto.CompressPubkey(&key.PublicKey)
	h.Authors = au
	h.Time = 123
	return h
}

//添加一个正常Unit，最新单元会更新，添加孤儿Unit，最新单元不会更新；补上了孤儿遗失的单元，那么孤儿单元不再是孤儿，会加到链上
func TestMemDag_AddOrphanUnit(t *testing.T){
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	lastHeader := newTestUnit(common.Hash{},0)
	txpool := txspool.NewMockITxPool(mockCtrl)
	db, _ := ptndb.NewMemDatabase()
	dagDb := storage.NewDagDb(db)
	utxoDb := storage.NewUtxoDb(db)
	stateDb := storage.NewStateDb(db)
	idxDb := storage.NewIndexDb(db)
	propDb := storage.NewPropertyDb(db)
	propDb.SetNewestUnit(lastHeader.Header())
	gp:=modules.NewGlobalProp()
	gp.ActiveMediators=make( map[common.Address]bool)
	gp.ActiveMediators[addr1]=true
	gp.ActiveMediators[addr2]=true
	propDb.StoreGlobalProp(gp)
	unitRep := dagcommon.NewUnitRepository(dagDb, idxDb, utxoDb, stateDb, propDb)
	unitRep.SaveUnit(lastHeader, false)
	propRep := dagcommon.NewPropRepository(propDb)
	stateRep:=dagcommon.NewStateRepository(stateDb)
	gasToken := modules.PTNCOIN
	memdag := NewMemDag(gasToken, true, db, unitRep, propRep,stateRep)
	u1:=newTestUnit(lastHeader.Hash(),1)
	log.Debugf("Try add unit[%x] to memdag",u1.Hash())
	err := memdag.AddUnit(u1, txpool)
	assert.Nil(t, err)
	assert.EqualValues(t,1, memdag.GetLastMainchainUnit().NumberU64())

	u2:=newTestUnit(u1.Hash(),2)
	u3:=newTestUnit(u2.Hash(),3)
	log.Debugf("Try add orphan unit[%x] to memdag",u3.Hash())
	err = memdag.AddUnit(u3, txpool)
	assert.Nil(t, err)
	assert.EqualValues(t,1, memdag.GetLastMainchainUnit().NumberU64())
	log.Debugf("Try add missed unit[%x] to memdag",u2.Hash())
	err = memdag.AddUnit(u2, txpool)
	assert.Nil(t, err)
	assert.EqualValues(t,3, memdag.GetLastMainchainUnit().NumberU64())
}
//添加1,2单元后，再次添加2'最新单元不变，再添加3‘ 则主链切换，最新单元更新为3’
func TestMemDag_SwitchMainChain(t *testing.T){
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	u0 := newTestUnit(common.Hash{},1)
	txpool := txspool.NewMockITxPool(mockCtrl)
	db, _ := ptndb.NewMemDatabase()
	dagDb := storage.NewDagDb(db)
	utxoDb := storage.NewUtxoDb(db)
	stateDb := storage.NewStateDb(db)
	idxDb := storage.NewIndexDb(db)
	propDb := storage.NewPropertyDb(db)
	propDb.SetNewestUnit(u0.UnitHeader)
	gp:=modules.NewGlobalProp()
	gp.ActiveMediators=make( map[common.Address]bool)
	gp.ActiveMediators[addr1]=true
	gp.ActiveMediators[addr2]=true
	propDb.StoreGlobalProp(gp)
	unitRep := dagcommon.NewUnitRepository(dagDb, idxDb, utxoDb, stateDb, propDb)
	unitRep.SaveUnit(u0, false)
	propRep := dagcommon.NewPropRepository(propDb)
	stateRep:=dagcommon.NewStateRepository(stateDb)
	gasToken := modules.PTNCOIN
	memdag := NewMemDag(gasToken, true, db, unitRep, propRep,stateRep)
	u1:=newTestUnit(u0.Hash(),2)
	log.Debugf("Try add unit[%x] to memdag",u1.Hash())
	err := memdag.AddUnit(u1, txpool)
	assert.Nil(t, err)
	assert.EqualValues(t,2, memdag.GetLastMainchainUnit().NumberU64())

	u22:=newTestUnit(u0.Hash(),2)
	log.Debugf("Try add side unit[%x] to memdag",u22.Hash())
	err = memdag.AddUnit(u22, txpool)
	assert.Nil(t, err)
	assert.EqualValues(t,u1.Hash(), memdag.GetLastMainchainUnit().Hash())

	u33:=newTestUnit(u22.Hash(),3)
	log.Debugf("Try add new longest chain unit[%x] to memdag",u33.Hash())

	err = memdag.AddUnit(u33, txpool)
	assert.Nil(t, err)
	assert.EqualValues(t,3, memdag.GetLastMainchainUnit().NumberU64())
}