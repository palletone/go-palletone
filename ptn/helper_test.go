// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

// This file contains some shares testing functionality, common to  multiple
// different files and modules being tested.

package ptn

import (
	"crypto/ecdsa"
	"crypto/rand"
	//"math/big"
	"fmt"
	"log"
	"sync"
	"testing"

	"github.com/palletone/go-palletone/common"

	//"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/common/event"
	"github.com/palletone/go-palletone/common/p2p"
	"github.com/palletone/go-palletone/common/p2p/discover"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/ptn/downloader"

	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/consensus/jury"
	"github.com/palletone/go-palletone/consensus/mediatorplugin"
	"github.com/palletone/go-palletone/dag"
	"github.com/palletone/go-palletone/dag/constants"
	"github.com/palletone/go-palletone/dag/storage"
	"github.com/palletone/go-palletone/dag/txspool"
	"github.com/palletone/go-palletone/tokenengine"
	"math/big"
	"time"
)

//var (
//	testBankKey, _ = crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
//	testBank       = crypto.PubkeyToAddress(testBankKey.PublicKey)
//)

// newTestProtocolManager creates a new protocol manager for testing purposes,
// with the given number of blocks already known, and potential notification
// channels for different events.
func newTestProtocolManager(mode downloader.SyncMode, blocks int, idag dag.IDag, pro producer, newtx chan<- []*modules.Transaction) (*ProtocolManager, ptndb.Database, error) {
	memdb, _ := ptndb.NewMemDatabase()
	if idag == nil {
		idag, _ = MakeDags(memdb, blocks)
	}
	//dag, _ := MakeDags(memdb, blocks)
	//uu := dag.CurrentUnit()
	//log.Printf("--------newTestProtocolManager--CurrentUnit--unit.UnitHeader-----%#v\n", uu.UnitHeader)
	//log.Printf("--------newTestProtocolManager--CurrentUnit--unit.UnitHash-------%#v\n", uu.UnitHash)
	//log.Printf("--------newTestProtocolManager--CurrentUnit--unit.UnitHeader.ParentsHash-----%#v\n", uu.UnitHeader.ParentsHash)
	//log.Printf("--------newTestProtocolManager--CurrentUnit--unit.UnitHeader.Number.Index-----%#v\n", uu.UnitHeader.Number.Index)
	//index := modules.ChainIndex{
	//	modules.PTNCOIN,
	//	true,
	//	0,
	//}
	//uu = dag.GetUnitByNumber(index)
	//log.Printf("--------newTestProtocolManager--index=0--unit.UnitHeader-----%#v\n", uu.UnitHeader)
	//log.Printf("--------newTestProtocolManager--index=0--unit.UnitHash-------%#v\n", uu.UnitHash)
	//log.Printf("--------newTestProtocolManager--index=0--unit.UnitHeader.ParentsHash-----%#v\n", uu.UnitHeader.ParentsHash)
	//log.Printf("--------newTestProtocolManager--index=0--unit.UnitHeader.Number.Index-----%#v\n", uu.UnitHeader.Number.Index)
	//engine := new(consensus.DPOSEngine)
	typemux := new(event.TypeMux)
	//producer mediatorplugin.MediatorPlugin
	if pro == nil {
		pro = new(mediatorplugin.MediatorPlugin)
	}
	//producer := new(mediatorplugin.MediatorPlugin)
	index0 := &modules.ChainIndex{
		modules.PTNCOIN,
		0,
	}
	genesisUint, _ := idag.GetUnitByNumber(index0)

	pm, err := NewProtocolManager(mode, DefaultConfig.NetworkId, modules.NewPTNIdType(), &testTxPool{added: newtx}, idag, typemux, pro, genesisUint, nil, nil)
	if err != nil {
		return nil, nil, err
	}
	//pm.SetForTest()
	config := p2p.DefaultConfig
	running := &p2p.Server{Config: config}
	pm.Start(running, 1000, nil)
	return pm, memdb, nil
}

// newTestProtocolManagerMust creates a new protocol manager for testing purposes,
// with the given number of blocks already known, and potential notification
// channels for different events. In case of an error, the constructor force-
// fails the test.
func newTestProtocolManagerMust(t *testing.T, mode downloader.SyncMode, blocks int, dag dag.IDag, pro producer, newtx chan<- []*modules.Transaction, ju *jury.Processor) (*ProtocolManager, ptndb.Database) {
	pm, db, err := newTestProtocolManager(mode, blocks /*generator,*/, dag, pro, newtx)
	if err != nil {
		t.Fatalf("Failed to create protocol manager: %v", err)
	}
	return pm, db
}

// testTxPool is a fake, helper transaction pool for testing purposes
type testTxPool struct {
	txFeed     event.Feed
	pool       []*modules.Transaction        // Collection of all transactions
	added      chan<- []*modules.Transaction // Notification channel for new transactions
	sequenPool *modules.SequeueTxPoolTxs

	lock sync.RWMutex // Protects the transaction pool
}

// AddRemotes appends a batch of transactions to the pool, and notifies any
// listeners if the addition channel is non nil

func (p *testTxPool) OutPointIsSpend(outPoint *modules.OutPoint) (bool, error) {
	return true, nil
}

func (p *testTxPool) AddLocal(tx *modules.Transaction) error {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.pool = append(p.pool, tx)
	if p.added != nil {
		p.added <- p.pool
	}
	return nil
}

func (p *testTxPool) AddLocals(txs []*modules.Transaction) []error {
	errs := make([]error, 0)
	for _, tx := range txs {
		errs = append(errs, p.AddLocal(tx))
	}
	return errs
}

func (p *testTxPool) AllHashs() []*common.Hash {
	hashs := make([]*common.Hash, 0)
	p.lock.RLock()
	defer p.lock.RUnlock()
	for _, tx := range p.pool {
		hash := tx.Hash()
		hashs = append(hashs, &hash)
	}
	return hashs
}

func (p *testTxPool) AddRemote(tx *modules.Transaction) error {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.pool = append(p.pool, tx)
	if p.added != nil {
		p.added <- p.pool
	}
	return nil

}

func (p *testTxPool) AddRemotes(txs []*modules.Transaction) []error {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.pool = append(p.pool, txs...)
	if p.added != nil {
		p.added <- txs
	}
	return make([]error, len(txs))
}
func (p *testTxPool) AddSequenTx(tx *modules.Transaction) error {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.sequenPool.Add(txspool.TxtoTxpoolTx(tx))
	return nil
}
func (p *testTxPool) AddSequenTxs(txs []*modules.Transaction) error {
	p.lock.Lock()
	defer p.lock.Unlock()
	for _, tx := range txs {
		p.sequenPool.Add(txspool.TxtoTxpoolTx(tx))
	}
	return nil
}

func (p *testTxPool) Content() (map[common.Hash]*modules.TxPoolTransaction, map[common.Hash]*modules.TxPoolTransaction) {
	return nil, nil
}

func (p *testTxPool) Get(hash common.Hash) (*modules.TxPoolTransaction, common.Hash) {
	return nil, (common.Hash{})
}

func (p *testTxPool) GetPoolTxsByAddr(addr string) ([]*modules.TxPoolTransaction, error) {
	return nil, nil
}

func (p *testTxPool) GetNonce(hash common.Hash) uint64 {
	return 0
}
func (p *testTxPool) GetUtxoEntry(outpoint *modules.OutPoint) (*modules.Utxo, error) {
	return nil, nil
}
func (p *testTxPool) GetSortedTxs(hash common.Hash, index uint64) ([]*modules.TxPoolTransaction, common.StorageSize) {
	return nil, 0
}
func (p *testTxPool) SendStoredTxs(hashs []common.Hash) error {
	return nil
}
func (p *testTxPool) Stats() (int, int, int) {
	return 0, 0, 0
}

func (p *testTxPool) Stop() {}

// Pending returns all the transactions known to the pool
func (p *testTxPool) Pending() (map[common.Hash][]*modules.TxPoolTransaction, error) {
	p.lock.RLock()
	defer p.lock.RUnlock()

	batches := make(map[common.Hash][]*modules.TxPoolTransaction)
	//for _, tx := range p.pool {
	// from, _ := types.Sender(types.HomesteadSigner{}, tx)
	// batches[from] = append(batches[from], tx)
	//}
	//for _, batch := range batches {
	// sort.Sort(types.TxByNonce(batch))
	//}
	return batches, nil
}
func (p *testTxPool) Queued() ([]*modules.TxPoolTransaction, error) {
	p.lock.RLock()
	defer p.lock.RUnlock()
	batches := make([]*modules.TxPoolTransaction, 0)
	return batches, nil
}

func (p *testTxPool) SubscribeTxPreEvent(ch chan<- modules.TxPreEvent) event.Subscription {
	return p.txFeed.Subscribe(ch)
}

func (p *testTxPool) ProcessTransaction(tx *modules.Transaction, allowOrphan bool, rateLimit bool, tag txspool.Tag) ([]*txspool.TxDesc, error) {
	return []*txspool.TxDesc{}, nil
}
func (p *testTxPool) AllTxpoolTxs() map[common.Hash]*modules.TxPoolTransaction {
	return nil
}
func (p *testTxPool) GetTxFee(tx *modules.Transaction) (*modules.AmountAsset, error) {
	return &modules.AmountAsset{}, nil
}
func (p *testTxPool) ValidateOrphanTx(tx *modules.Transaction) (bool, error) {
	return false, nil
}
func (p *testTxPool) DiscardTxs(hashs []common.Hash) error {
	return nil
}
func (p *testTxPool) ResetPendingTxs(txs []*modules.Transaction) error {
	return nil
}
func (p *testTxPool) SetPendingTxs(unit_hash common.Hash, num uint64, txs []*modules.Transaction) error {
	return nil
}

//func (p *testTxPool) DiscardTx(hash common.Hash) error {
//	return nil
//}

// newTestTransaction create a new dummy transaction.
func newTestTransaction(from *ecdsa.PrivateKey, nonce uint64, datasize int) *modules.Transaction {
	msg := &modules.Message{
		App: modules.APP_PAYMENT,
		//PayloadHash: common.HexToHash("xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"),
		Payload: "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
	}
	//tx := modules.NewTransaction(nonce, big.NewInt(0), []byte("abc"))
	tx := modules.NewTransaction(
		[]*modules.Message{msg, msg, msg},
	)

	return tx
}

// testPeer is a simulated peer to allow testing direct network calls.
type testPeer struct {
	net     p2p.MsgReadWriter // Network layer reader/writer to simulate remote messaging
	app     *p2p.MsgPipeRW    // Application layer reader/writer to simulate the local side
	version int
	*peer
}

// newTestPeer creates a new peer registered at the given protocol manager.
func newTestPeer(name string, version int, pm *ProtocolManager, shake bool, dag dag.IDag) (*testPeer, <-chan error) {
	// Create a message pipe to communicate through
	app, net := p2p.MsgPipe()

	// Generate a random id and create the peer
	var id discover.NodeID
	rand.Read(id[:])

	peer := pm.newPeer(version, p2p.NewPeer(id, name, nil), net)

	// Start the peer on a new thread
	errc := make(chan error, 1)
	go func() {
		select {
		case pm.newPeerCh <- peer:
			errc <- pm.handle(peer)
		case <-pm.quitSync:
			errc <- p2p.DiscQuitting
		}
	}()
	tp := &testPeer{app: app, net: net, peer: peer, version: version}
	// Execute any implicitly requested handshakes and return
	if shake {
		var (
			//number = modules.ChainIndex{
			//	modules.PTNCOIN,
			//	true,
			//	0,
			//}
			//genesis = pm.dag.GetUnitByNumber(number)
			head   = pm.dag.CurrentHeader(modules.PTNCOIN)
			index  = head.Number
			stable = pm.dag.GetStableChainIndex(modules.PTNCOIN)
		)
		fmt.Println("==========================================index:", index.Index)
		//fmt.Println("	if shake {===》》》",td)
		//genesis, err := dag.GetGenesisUnit(0)
		////fmt.Println("genesis unti if shake {===》》》", genesis.UnitHash)
		//if err != nil {
		//	fmt.Println("GetGenesisUnit===error:=", err)
		//}
		tp.handshake(nil, index, stable, head.Hash(), pm.genesis.Hash())
	}
	return tp, errc
}

// handshake simulates a trivial handshake that expects the same state from the
// remote side as we are simulating locally.
func (p *testPeer) handshake(t *testing.T, index, stalbe *modules.ChainIndex, head common.Hash, genesis common.Hash) {
	msg := &statusData{
		ProtocolVersion: uint32(p.version),
		NetworkId:       DefaultConfig.NetworkId,
		Index:           index,
		GenesisUnit:     genesis,
		//StableIndex:     stalbe,
	}
	if err := p2p.ExpectMsg(p.app, StatusMsg, msg); err != nil {
		//log.Fatalf("status recv: %v", err)
	}
	if err := p2p.Send(p.app, StatusMsg, msg); err != nil {
		log.Fatalf("status send: %v", err)
	}
}

// close terminates the local side of the peer, notifying the remote protocol
// manager of termination.
func (p *testPeer) close() {
	p.app.Close()
}

func MakeDags(Memdb ptndb.Database, unitAccount int) (*dag.Dag, error) {
	dag, _ := dag.NewDagForTest(Memdb)
	genesisUnit := newGenesisForTest(dag.Db)
	newDag(dag.Db, genesisUnit, unitAccount)
	return dag, nil
}
func unitForTest(index int) *modules.Unit {
	header := modules.NewHeader([]common.Hash{}, 1, []byte{})
	header.Number.AssetID = modules.PTNCOIN
	//header.Number.IsMain = true
	header.Number.Index = uint64(index)
	header.Authors = modules.Authentifier{[]byte{}, []byte{}}
	header.GroupSign = []byte{}
	header.GroupPubKey = []byte{}
	//tx, _ := NewCoinbaseTransaction()
	tx, _ := CreateCoinbase()
	fmt.Printf("----------%#v\n", tx)
	txs := modules.Transactions{tx}
	genesisUnit := modules.NewUnit(header, txs)
	return genesisUnit
}

func newGenesisForTest(db ptndb.Database) *modules.Unit {
	header := modules.NewHeader([]common.Hash{}, 1, []byte{})
	header.Number.AssetID = modules.PTNCOIN
	//header.Number.IsMain = true
	header.Number.Index = 0
	header.Authors = modules.Authentifier{[]byte{}, []byte{}}
	header.GroupSign = []byte{}
	header.GroupPubKey = []byte{}
	//tx, _ := NewCoinbaseTransaction()
	tx, _ := CreateCoinbase()
	fmt.Printf("----------%#v\n", tx)
	txs := modules.Transactions{tx}
	genesisUnit := modules.NewUnit(header, txs)
	err := SaveGenesis(db, genesisUnit)
	if err != nil {
		log.Println("SaveGenesis, err", err)
		return nil
	}
	return genesisUnit
}
func newDag(memdb ptndb.Database, gunit *modules.Unit, number int) (modules.Units, error) {
	units := make(modules.Units, number)
	par := gunit
	for i := 0; i < number; i++ {
		header := modules.NewHeader([]common.Hash{par.UnitHash}, 1, []byte{})
		header.Number.AssetID = par.UnitHeader.Number.AssetID
		//header.Number.IsMain = par.UnitHeader.Number.IsMain
		header.Number.Index = par.UnitHeader.Number.Index + 1
		header.Authors = modules.Authentifier{[]byte{}, []byte{}}
		header.GroupSign = []byte{}
		header.GroupPubKey = []byte{}
		//tx, _ := NewCoinbaseTransaction()
		tx, _ := CreateCoinbase()
		txs := modules.Transactions{tx}
		unit := modules.NewUnit(header, txs)
		err := SaveUnit(memdb, unit, true)
		if err != nil {
			log.Println("save genesis error", err)
			return nil, err
		}
		units[i] = unit
		par = unit
	}
	return units, nil
}

func SaveGenesis(db ptndb.Database, unit *modules.Unit) error {
	if unit.NumberU64() != 0 {
		return fmt.Errorf("can't commit genesis unit with number > 0")
	}
	err := SaveUnit(db, unit, true)
	if err != nil {
		log.Println("SaveGenesis==", err)
		return err
	}
	return nil
}

func CreateCoinbase() (*modules.Transaction, error) {
	sAddr := "P1NsG3kiKJc87M6Di6YriqHxqfPhdvxVj2B"
	addr, err := common.StringToAddress(sAddr)
	if err != nil {

	}
	//bAsset := []byte("GenesisAsset")
	//if len(bAsset) <= 0 {
	//	return nil, fmt.Errorf("Create unit error: query asset info empty")
	//}
	//var asset modules.Asset
	//if err := rlp.DecodeBytes(bAsset, &asset); err != nil {
	//	fmt.Println("lalall")
	//	return nil, fmt.Errorf("Create unit: %s", err.Error())
	//}
	asset := modules.NewPTNAsset()
	// setp1. create P2PKH script
	script := tokenengine.Instance.GenerateP2PKHLockScript(addr.Bytes())
	// step. compute total income
	totalIncome := int64(100000000) + int64(100000000)
	// step2. create payload
	createT := big.Int{}
	input := modules.Input{
		Extra: createT.SetInt64(time.Now().Unix()).Bytes(),
	}
	output := modules.Output{
		Value:    uint64(totalIncome),
		Asset:    asset,
		PkScript: script,
	}
	payload := modules.PaymentPayload{
		Inputs:  []*modules.Input{&input},
		Outputs: []*modules.Output{&output},
	}
	// step3. create message
	msg := &modules.Message{
		App:     modules.APP_PAYMENT,
		Payload: &payload,
	}
	// step4. create coinbase
	var coinbase modules.Transaction
	//coinbase := modules.Transaction{
	//	TxMessages: []modules.Message{msg},
	//}
	coinbase.TxMessages = append(coinbase.TxMessages, msg)
	// coinbase.CreationDate = coinbase.CreateDate()
	//coinbase.TxHash = coinbase.Hash()

	return &coinbase, nil
}

func SaveUnit(db ptndb.Database, unit *modules.Unit, isGenesis bool) error {
	if unit.UnitSize == 0 || unit.Size() == 0 {
		log.Println("Unit is null")
		return fmt.Errorf("Unit is null")
	}
	if unit.UnitSize != unit.Size() {
		log.Println("Validate size", "error", "Size is invalid")
		return modules.ErrUnit(-1)
	}
	//_, isSuccess, err := dag.ValidateTransactions(&unit.Txs, isGenesis)
	//if isSuccess != true {
	//	fmt.Errorf("Validate unit(%s) transactions failed: %v", unit.UnitHash.String(), err)
	//	return fmt.Errorf("Validate unit(%s) transactions failed: %v", unit.UnitHash.String(), err)
	//}
	// step4. save unit header
	// key is like "[HEADER_PREFIX][chain index number]_[chain index]_[unit hash]"
	//l := log2.NewTestLog()
	//dagDb := storage.NewDagDb(db, l)
	dagDb := storage.NewDagDb(db)

	if err := dagDb.SaveHeader(unit.UnitHeader); err != nil {
		log.Println("SaveHeader:", "error", err.Error())
		return modules.ErrUnit(-3)
	}
	// step5. save unit hash and chain index relation
	// key is like "[UNIT_HASH_NUMBER][unit_hash]"
	//if err := dagDb.SaveNumberByHash(unit.UnitHash, unit.UnitHeader.Number); err != nil {
	//	log.Println("SaveHashNumber:", "error", err.Error())
	//	return fmt.Errorf("Save unit hash and number error")
	//}
	//if err := dagDb.SaveHashByNumber(unit.UnitHash, unit.UnitHeader.Number); err != nil {
	//	log.Println("SaveNumberByHash:", "error", err.Error())
	//	return fmt.Errorf("Save unit hash and number error")
	//}

	// step6. traverse transactions and save them
	txHashSet := []common.Hash{}
	for _, tx := range unit.Txs {
		// traverse messages

		//fmt.Println("tx==", tx.TxHash)
		// step7. save transaction
		if err := dagDb.SaveTransaction(tx); err != nil {
			log.Println("Save transaction:", "error", err.Error())
			return err
		}
		txHashSet = append(txHashSet, tx.Hash())
	}

	// step8. save unit body, the value only save txs' hash set, and the key is merkle root
	if err := dagDb.SaveBody(unit.UnitHash, txHashSet); err != nil {
		log.Println("SaveBody", "error", err.Error())
		return err
	}
	if err := dagDb.SaveTxLookupEntry(unit); err != nil {
		return err
	}
	if err := dagDb.SaveTxLookupEntry(unit); err != nil {
		return err
	}
	if err := saveHashByIndex(db, unit.UnitHash, unit.UnitHeader.Number.Index); err != nil {
		return err
	}
	// update state
	//dagDb.PutCanonicalHash(unit.UnitHash, unit.NumberU64())
	//dagDb.PutHeadHeaderHash(unit.UnitHash)
	//dagDb.PutHeadUnitHash(unit.UnitHash)
	//dagDb.PutHeadFastUnitHash(unit.UnitHash)
	// todo send message to transaction pool to delete unit's transactions
	return nil
}
func NewUnit(header *modules.Header, txs modules.Transactions) *modules.Unit {
	u := &modules.Unit{
		UnitHeader: header,
		Txs:        txs,
	}
	u.ReceivedAt = time.Now()
	u.UnitSize = u.Size()
	u.UnitHash = u.Hash()
	return u
}
func NewHeader(parents []common.Hash, asset []modules.AssetId, extra []byte) *modules.Header {
	hashs := make([]common.Hash, 0)
	hashs = append(hashs, parents...) // 切片指针传递的问题，这里得再review一下。
	var b []byte
	//return &modules.Header{ParentsHash: hashs, AssetIDs: asset, Extra: append(b, extra...), Time: time.Now().Unix()}
	return &modules.Header{ParentsHash: hashs, Extra: append(b, extra...), Time: time.Now().Unix()}
}
func NewCoinbaseTransaction() (*modules.Transaction, error) {
	input := &modules.Input{}
	output := &modules.Output{}
	payload := modules.PaymentPayload{
		Inputs:  []*modules.Input{input},
		Outputs: []*modules.Output{output},
	}
	msg := modules.Message{
		App:     modules.APP_PAYMENT,
		Payload: payload,
	}
	var coinbase modules.Transaction
	coinbase.TxMessages = append(coinbase.TxMessages, &msg)
	//coinbase.TxHash = coinbase.Hash()
	return &coinbase, nil
}

func saveHashByIndex(db ptndb.Database, hash common.Hash, index uint64) error {
	key := fmt.Sprintf("%s%v_", constants.HEADER_PREFIX, index)
	err := db.Put([]byte(key), hash.Bytes())
	return err
}
