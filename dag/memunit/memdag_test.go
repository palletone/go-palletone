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
	"crypto/ecdsa"
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
)

func TestMemDag_AddUnit(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	lastHeader := newTestUnit()
	txpool := txspool.NewMockITxPool(mockCtrl)
	db, _ := ptndb.NewMemDatabase()
	dagDb := storage.NewDagDb(db)
	utxoDb := storage.NewUtxoDb(db)
	stateDb := storage.NewStateDb(db)
	idxDb := storage.NewIndexDb(db)
	propDb := storage.NewPropertyDb(db)
	propDb.SetLastStableUnit(lastHeader.Hash(), modules.NewChainIndex(modules.PTNCOIN, 1))

	//utxoRep := dagcommon.NewUtxoRepository(utxoDb, idxDb, stateDb)
	unitRep := dagcommon.NewUnitRepository(dagDb, idxDb, utxoDb, stateDb, propDb)
	unitRep.SaveUnit(lastHeader, false)
	propRep := dagcommon.NewPropRepository(propDb)
	propRep.StoreGlobalProp(modules.NewGlobalProp())
	//stateRep := dagcommon.NewStateRepository(stateDb)
	//hash, idx, _ := propRep.GetLastStableUnit(modules.PTNCOIN)
	gasToken := modules.PTNCOIN
	memdag := NewMemDag(gasToken, false, db, unitRep, propRep)
	//tunitRep, tutxoRep, tstateRep := unstableChain.GetUnstableRepositories()

	err := memdag.AddUnit(newTestUnit(), txpool)
	assert.Nil(t, err)
}
func BenchmarkMemDag_AddUnit(b *testing.B) {
	mockCtrl := gomock.NewController(b)
	defer mockCtrl.Finish()
	lastHeader := newTestUnit()
	txpool := txspool.NewMockITxPool(mockCtrl)
	db, _ := ptndb.NewMemDatabase()
	dagDb := storage.NewDagDb(db)
	utxoDb := storage.NewUtxoDb(db)
	stateDb := storage.NewStateDb(db)
	idxDb := storage.NewIndexDb(db)
	propDb := storage.NewPropertyDb(db)
	propDb.SetLastStableUnit(lastHeader.Hash(), modules.NewChainIndex(modules.PTNCOIN, 0))
	propDb.SetNewestUnit(lastHeader.Header())
	//utxoRep := dagcommon.NewUtxoRepository(utxoDb, idxDb, stateDb)
	unitRep := dagcommon.NewUnitRepository(dagDb, idxDb, utxoDb, stateDb, propDb)
	unitRep.SaveUnit(lastHeader, false)
	propRep := dagcommon.NewPropRepository(propDb)
	propRep.StoreGlobalProp(modules.NewGlobalProp())
	//stateRep := dagcommon.NewStateRepository(stateDb)
	//hash, idx, _ := propRep.GetLastStableUnit(modules.PTNCOIN)
	gasToken := modules.PTNCOIN
	memdag := NewMemDag(gasToken, false, db, unitRep, propRep)
	//tunitRep, tutxoRep, tstateRep := unstableChain.GetUnstableRepositories()
	parentHash := lastHeader.Hash()
	for i := 0; i < b.N; i++ {
		unit := newTestChildUnit(parentHash, i+1)
		err := memdag.AddUnit(unit, txpool)
		assert.Nil(b, err)
		parentHash = unit.Hash()
	}
}
func newTestUnit() *modules.Unit {
	h := newTestHeader()
	return modules.NewUnit(h, []*modules.Transaction{})
}
func newTestChildUnit(parent common.Hash, height int) *modules.Unit {
	h := newTestHeader()
	h.ParentsHash = []common.Hash{parent}
	h.Number.Index = uint64(height)
	return modules.NewUnit(h, []*modules.Transaction{})
}
func newTestHeader() *modules.Header {
	key := new(ecdsa.PrivateKey)
	key, _ = crypto.GenerateKey()
	h := new(modules.Header)
	//h.AssetIDs = append(h.AssetIDs, PTNCOIN)
	au := modules.Authentifier{}

	h.GroupSign = []byte("group_sign")
	h.GroupPubKey = []byte("group_pubKey")
	h.Number = &modules.ChainIndex{}
	h.Number.AssetID = modules.PTNCOIN
	h.Number.Index = uint64(0)
	h.Extra = make([]byte, 20)
	h.ParentsHash = append(h.ParentsHash, h.TxRoot)
	//tr := common.Hash{}
	//tr = tr.SetString("c35639062e40f8891cef2526b387f42e353b8f403b930106bb5aa3519e59e35f")
	h.TxRoot = common.HexToHash("c35639062e40f8891cef2526b387f42e353b8f403b930106bb5aa3519e59e35f")
	sig, _ := crypto.Sign(h.TxRoot[:], key)
	au.Signature = sig
	au.PubKey = crypto.CompressPubkey(&key.PublicKey)
	h.Authors = au
	h.Time = 123
	return h
}
