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
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/crypto"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/tokenengine"
	"time"

	dagcommon "github.com/palletone/go-palletone/dag/common"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/storage"
	"github.com/stretchr/testify/assert"

	"github.com/coocood/freecache"
	"github.com/golang/mock/gomock"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/palletcache"
	"github.com/palletone/go-palletone/dag/txspool"
	"github.com/palletone/go-palletone/validator"
	"testing"
)

func cache() palletcache.ICache {
	return freecache.NewCache(1000 * 1024)
}
func TestMemDag_AddUnit(t *testing.T) {
	lastHeader := newTestUnit(common.Hash{}, 0, key1)
	db, _ := ptndb.NewMemDatabase()
	dagDb := storage.NewDagDb(db)
	utxoDb := storage.NewUtxoDb(db, tokenengine.Instance)
	stateDb := storage.NewStateDb(db)
	idxDb := storage.NewIndexDb(db)
	propDb := storage.NewPropertyDb(db)
	propDb.SetNewestUnit(lastHeader.UnitHeader)

	unitRep := dagcommon.NewUnitRepository(dagDb, idxDb, utxoDb, stateDb, propDb, tokenengine.Instance)
	unitRep.SaveUnit(lastHeader, false)
	propRep := dagcommon.NewPropRepository(propDb)
	propRep.StoreGlobalProp(modules.NewGlobalProp())
	stateRep := dagcommon.NewStateRepository(stateDb)
	gasToken := dagconfig.DagConfig.GetGasToken()
	memdag := NewMemDag(gasToken, 2, false,
		db, unitRep, propRep, stateRep, cache(), tokenengine.Instance)
	parent := common.HexToHash("0x2c30cd5b06c4c6d184aae3e1ed76492f16a0fa335673dba99c8efa813c1a1e30")
	_, _, _, _, _, err := memdag.AddUnit(newTestUnit(parent, 1, key2), nil, true)
	assert.Nil(t, err)
}
func BenchmarkMemDag_AddUnit(b *testing.B) {
	//mockCtrl := gomock.NewController(t)
	//defer mockCtrl.Finish()
	lastHeader := newTestUnit(common.Hash{}, 0, key1)
	db, _ := ptndb.NewMemDatabase()
	dagDb := storage.NewDagDb(db)
	utxoDb := storage.NewUtxoDb(db, tokenengine.Instance)
	stateDb := storage.NewStateDb(db)
	idxDb := storage.NewIndexDb(db)
	propDb := storage.NewPropertyDb(db)
	propDb.SetNewestUnit(lastHeader.UnitHeader)
	propDb.SetNewestUnit(lastHeader.Header())

	unitRep := dagcommon.NewUnitRepository(dagDb, idxDb, utxoDb, stateDb, propDb, tokenengine.Instance)
	unitRep.SaveUnit(lastHeader, false)
	propRep := dagcommon.NewPropRepository(propDb)
	propRep.StoreGlobalProp(modules.NewGlobalProp())
	stateRep := dagcommon.NewStateRepository(stateDb)
	gasToken := modules.PTNCOIN
	memdag := NewMemDag(gasToken, 2, false,
		db, unitRep, propRep, stateRep, cache(), tokenengine.Instance)

	parentHash := lastHeader.Hash()
	for i := 0; i < b.N; i++ {
		unit := newTestUnit(parentHash, uint64(i+1), key1)
		_, _, _, _, _, err := memdag.AddUnit(unit, nil, true)
		assert.Nil(b, err)
		parentHash = unit.Hash()
	}
}
func newTestUnit(parentHash common.Hash, height uint64, key []byte) *modules.Unit {
	h := newTestHeader(parentHash, height, key)
	txs := make(modules.Transactions, 0)
	msgs := make([]*modules.Message, 0)
	payload := new(modules.DataPayload)
	payload.MainData = []byte("test_main")
	payload.ExtraData = []byte("test_extra")
	sig := new(modules.SignaturePayload)
	sig.Signatures = append(sig.Signatures, modules.SignatureSet{Signature: []byte("sig")})
	msgs = append(msgs, &modules.Message{App: modules.APP_DATA, Payload: payload})
	msgs = append(msgs, &modules.Message{App: modules.APP_SIGNATURE, Payload: sig})
	txs = append(txs, modules.NewTransaction(msgs))
	h.TxRoot = core.DeriveSha(txs)
	return modules.NewUnit(h, txs)
}

var (
	key1, _    = crypto.MyCryptoLib.KeyGen()
	pubKey1, _ = crypto.MyCryptoLib.PrivateKeyToPubKey(key1)
	addr1      = crypto.PubkeyBytesToAddress(pubKey1)
	key2, _    = crypto.MyCryptoLib.KeyGen()
	pubKey2, _ = crypto.MyCryptoLib.PrivateKeyToPubKey(key2)
	addr2      = crypto.PubkeyBytesToAddress(pubKey2)
	key3, _    = crypto.MyCryptoLib.KeyGen()
	key4, _    = crypto.MyCryptoLib.KeyGen()
)

func newTestHeader(parentHash common.Hash, height uint64, key []byte) *modules.Header {

	h := new(modules.Header)
	au := modules.Authentifier{}

	h.GroupSign = []byte("group_sign")
	h.GroupPubKey = []byte("group_pubKey")
	h.Number = &modules.ChainIndex{}
	h.Number.AssetID = dagconfig.DagConfig.GetGasToken()
	h.Number.Index = height
	h.Extra = make([]byte, 20)
	h.ParentsHash = []common.Hash{parentHash}
	h.TxRoot = common.HexToHash("0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421")

	sig, _ := crypto.MyCryptoLib.Sign(key, h.TxRoot[:])
	au.Signature = sig
	au.PubKey, _ = crypto.MyCryptoLib.PrivateKeyToPubKey(key)
	h.Authors = au
	h.Time = int64(1536451200) + 1000
	return h
}

//添加一个正常Unit，最新单元会更新，添加孤儿Unit，最新单元不会更新；补上了孤儿遗失的单元，那么孤儿单元不再是孤儿，会加到链上
func TestMemDag_AddOrphanUnit(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	txpool := txspool.NewMockITxPool(mockCtrl)
	lastHeader := newTestUnit(common.Hash{}, 0, key1)
	txpool.EXPECT().SetPendingTxs(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	txpool.EXPECT().ResetPendingTxs(gomock.Any()).Return(nil).AnyTimes()
	db, _ := ptndb.NewMemDatabase()
	dagDb := storage.NewDagDb(db)
	utxoDb := storage.NewUtxoDb(db, tokenengine.Instance)
	stateDb := storage.NewStateDb(db)
	idxDb := storage.NewIndexDb(db)
	propDb := storage.NewPropertyDb(db)
	propDb.SetNewestUnit(lastHeader.Header())
	mockMediatorInit(stateDb, propDb)
	unitRep := dagcommon.NewUnitRepository(dagDb, idxDb, utxoDb, stateDb, propDb, tokenengine.Instance)
	unitRep.SaveUnit(lastHeader, false)
	propRep := dagcommon.NewPropRepository(propDb)
	stateRep := dagcommon.NewStateRepository(stateDb)
	gasToken := modules.PTNCOIN
	memdag := NewMemDag(gasToken, 2, false,
		db, unitRep, propRep, stateRep, cache(), tokenengine.Instance)
	u1 := newTestUnit(lastHeader.Hash(), 1, key2)
	log.Debugf("Try add unit[%x] to memdag, index: %d", u1.Hash(), u1.NumberU64())

	log.Debugf("unit txs:[%#x]", u1.Txs.GetTxIds())
	_, _, _, _, _, err := memdag.AddUnit(u1, txpool, true)
	assert.Nil(t, err)
	assert.EqualValues(t, 1, memdag.GetLastMainChainUnit().NumberU64())

	u2 := newTestUnit(u1.Hash(), 2, key1)
	u3 := newTestUnit(u2.Hash(), 3, key2)
	log.Debugf("Try add orphan unit[%x] to memdag, index: %d", u3.Hash(), u3.NumberU64())
	_, _, _, _, _, err = memdag.AddUnit(u3, txpool, true)
	assert.Nil(t, err)
	assert.EqualValues(t, 1, memdag.GetLastMainChainUnit().NumberU64())
	log.Debugf("Try add missed unit[%x] to memdag, index: %d", u2.Hash(), u2.NumberU64())
	_, _, _, _, _, err = memdag.AddUnit(u2, txpool, true)
	assert.Nil(t, err)
	time.Sleep(1 * time.Second)
	assert.EqualValues(t, 3, memdag.GetLastMainChainUnit().NumberU64())
}

//添加1,2单元后，再次添加2'最新单元不变，再添加3'则主链切换，最新单元更新为3'
func TestMemDag_SwitchMainChain(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	txpool := txspool.NewMockITxPool(mockCtrl)
	txpool.EXPECT().SetPendingTxs(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	txpool.EXPECT().ResetPendingTxs(gomock.Any()).Return(nil).AnyTimes()
	u0 := newTestUnit(common.Hash{}, 1, key1)

	db, _ := ptndb.NewMemDatabase()
	dagDb := storage.NewDagDb(db)
	utxoDb := storage.NewUtxoDb(db, tokenengine.Instance)
	stateDb := storage.NewStateDb(db)
	idxDb := storage.NewIndexDb(db)
	propDb := storage.NewPropertyDb(db)
	propDb.SetNewestUnit(u0.UnitHeader)
	mockMediatorInit(stateDb, propDb)
	unitRep := dagcommon.NewUnitRepository(dagDb, idxDb, utxoDb, stateDb, propDb, tokenengine.Instance)
	unitRep.SaveUnit(u0, false)
	propRep := dagcommon.NewPropRepository(propDb)
	stateRep := dagcommon.NewStateRepository(stateDb)
	gasToken := modules.PTNCOIN
	memdag := NewMemDag(gasToken, 2,
		false, db, unitRep, propRep, stateRep, cache(), tokenengine.Instance)

	u1 := newTestUnit(u0.Hash(), 2, key2)
	log.Debugf("Try add unit[%x] to memdag", u1.Hash())
	_, _, _, _, _, err := memdag.AddUnit(u1, txpool, true)
	assert.Nil(t, err)
	assert.EqualValues(t, 2, memdag.GetLastMainChainUnit().NumberU64())

	u22 := newTestUnit(u0.Hash(), 2, key1)
	log.Debugf("Try add side unit[%x] to memdag", u22.Hash())
	_, _, _, _, _, err = memdag.AddUnit(u22, txpool, true)
	assert.Nil(t, err)
	assert.EqualValues(t, u1.Hash(), memdag.GetLastMainChainUnit().Hash())

	u33 := newTestUnit(u22.Hash(), 3, key2)
	log.Debugf("Try add new longest chain unit[%x] to memdag", u33.Hash())

	_, _, _, _, _, err = memdag.AddUnit(u33, txpool, true)
	assert.Nil(t, err)
	time.Sleep(1 * time.Second)
	assert.EqualValues(t, 3, memdag.GetLastMainChainUnit().NumberU64())
}

func mockMediatorInit(statedb storage.IStateDb, propDb storage.IPropertyDb) {
	point, _ := core.StrToPoint("Dsn4gF2xpsM79R6kBfsR1joZD4BoPfBGREJGStCAz1bFfUnB5QXBGbNfudxyCWz6uWZZ8c43BYWkxiezyF5uifhv1diiykrxzgFhLMSAvppx34RjJwzjmXAXnYMuQX3Jy2P3ygehcKmATAyXQCVoXde6Xo3tkA2Jv8Zb8zDcdGjbFyd")
	node, _ := core.StrToMedNode("pnode://f056aca66625c286ae444add82f44b9eb74f18a8a96572360cb70df9b6d64d9bd2c58a345e570beb2bcffb037cd0a075f548b73083d31c12f1f4564865372534@127.0.0.1:30303")
	m1 := &core.Mediator{Address: addr1, InitPubKey: point, Node: node,
		MediatorApplyInfo: core.NewMediatorApplyInfo(), MediatorInfoExpand: core.NewMediatorInfoExpand()}

	statedb.StoreMediator(m1)
	m2 := &core.Mediator{Address: addr2, InitPubKey: point, Node: node,
		MediatorApplyInfo: core.NewMediatorApplyInfo(), MediatorInfoExpand: core.NewMediatorInfoExpand()}
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
func mockValidator() validator.Validator {
	return &mockValidate{}
}

type mockValidate struct {
}

func (v mockValidate) ValidateTx(tx *modules.Transaction, isFullTx bool) ([]*modules.Addition, validator.ValidationCode, error) {
	return nil, validator.TxValidationCode_VALID, nil
}

func (v mockValidate) ValidateUnitExceptGroupSig(unit *modules.Unit) validator.ValidationCode {
	return validator.TxValidationCode_VALID
}
func (v mockValidate) ValidateUnitExceptPayment(unit *modules.Unit) error {
	return nil
}

//验证一个Header是否合法（Mediator签名有效）
func (v mockValidate) ValidateHeader(h *modules.Header) validator.ValidationCode {
	return validator.TxValidationCode_VALID
}
func (v mockValidate) ValidateUnitGroupSign(h *modules.Header) error {
	return nil
}
func (v mockValidate) CheckTxIsExist(tx *modules.Transaction) bool {
	return false
}
