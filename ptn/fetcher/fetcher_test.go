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

package fetcher

import (
	"fmt"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/ptndb"
	//dag2 "github.com/palletone/go-palletone/dag"
	"github.com/palletone/go-palletone/dag/constants"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/storage"
	"log"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

var (
	memdb, _    = ptndb.NewMemDatabase()
	genesisUnit = newGenesisForTest(memdb)
)

func newGenesisForTest(db ptndb.Database) *modules.Unit {
	header := modules.NewHeader([]common.Hash{}, 1, []byte{})

	header.Number.AssetID = modules.PTNCOIN
	//header.Number.IsMain = true
	header.Number.Index = 0
	header.Authors = modules.Authentifier{[]byte{}, []byte{}}
	header.GroupSign = []byte{}
	header.GroupPubKey = []byte{}
	tx, _ := NewCoinbaseTransaction()
	txs := modules.Transactions{tx}
	genesisUnit := modules.NewUnit(header, txs)
	//fmt.Println("genesisUnit=", genesisUnit.Hash())
	//fmt.Println("genesisUTx=", genesisUnit.Transactions()[0])
	err := SaveGenesis(db, genesisUnit)
	if err != nil {
		log.Println("SaveGenesis, err", err)
		return nil
	}
	return genesisUnit
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
func newDag(db ptndb.Database, gunit *modules.Unit, number int) (modules.Units, error) {
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
		tx, _ := NewCoinbaseTransaction()
		txs := modules.Transactions{tx}
		unit := modules.NewUnit(header, txs)
		err := SaveUnit(db, unit, true)
		if err != nil {
			log.Println("save genesis error", err)
			return nil, err
		}
		units[i] = unit
		par = unit
	}
	return units, nil
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
	//l := log2.NewTestLog()
	dagDb := storage.NewDagDb(db)
	// step4. save unit header
	// key is like "[HEADER_PREFIX][chain index number]_[chain index]_[unit hash]"
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
func saveHashByIndex(db ptndb.Database, hash common.Hash, index uint64) error {
	key := fmt.Sprintf("%s%v_", constants.HEADER_PREFIX, index)
	err := db.Put([]byte(key), hash.Bytes())
	return err
}

// makeChain creates a chain of n blocks starting at and including parent.
// the returned hash chain is ordered head->parent. In addition, every 3rd block
// contains a transaction and every 5th an uncle to allow testing correct block reassembly
func makeDag(n int, parent *modules.Unit) ([]common.Hash, map[common.Hash]*modules.Unit) {
	hashes := make([]common.Hash, n+1)
	hashes[len(hashes)-1] = parent.Hash()
	dags := make(map[common.Hash]*modules.Unit, n+1)
	dags[parent.Hash()] = parent
	memdb, _ := ptndb.NewMemDatabase()
	//dag, _ := dag2.NewDagForTest(memdb)
	units, err := newDag(memdb, parent, n)
	if err != nil {
		log.Println("new dag err", err.Error())
	}
	for i, u := range units {
		hashes[len(hashes)-i-2] = u.Hash()
		dags[u.Hash()] = u
	}
	return hashes, dags
}

// fetcherTester is a test simulator for mocking out local block chain.
type fetcherTester struct {
	fetcher *Fetcher

	hashes []common.Hash                 // Hash chain belonging to the tester
	blocks map[common.Hash]*modules.Unit // Blocks belonging to the tester
	drops  map[string]bool               // Map of peers dropped by the fetcher

	lock sync.RWMutex
}

// newTester creates a new fetcher test mocker.
func newTester() *fetcherTester {

	tester := &fetcherTester{
		hashes: []common.Hash{genesisUnit.Hash()},
		blocks: map[common.Hash]*modules.Unit{genesisUnit.Hash(): genesisUnit},
		drops:  make(map[string]bool),
	}
	tester.fetcher = New(tester.getBlock, tester.verifyHeader, tester.broadcastBlock, tester.chainHeight, tester.insertChain, tester.dropPeer)
	tester.fetcher.Start()

	return tester
}

// isHeaderExist retrieves a block from the tester's block chain.
func (f *fetcherTester) getBlock(hash common.Hash) bool {
	f.lock.RLock()
	defer f.lock.RUnlock()
	_, exist := f.blocks[hash]
	return exist
}

// verifyHeader is a nop placeholder for the block header verification.
func (f *fetcherTester) verifyHeader(header *modules.Unit) error {
	return nil
}

// broadcastBlock is a nop placeholder for the block broadcasting.
func (f *fetcherTester) broadcastBlock(block *modules.Unit, propagate bool) {
}

// chainHeight retrieves the current height (block number) of the chain.
func (f *fetcherTester) chainHeight(assetId modules.AssetId) uint64 {
	f.lock.RLock()
	defer f.lock.RUnlock()
	//mem, _ := ptndb.NewMemDatabase()
	//dag, _ := dag2.NewDag(memdb)
	//unit := dag.GetCurrentUnit(assetId)
	//if unit != nil {
	//	return unit.NumberU64()
	//}
	//return uint64(0)
	return f.blocks[f.hashes[len(f.hashes)-1]].NumberU64()
}

// insertChain injects a new blocks into the simulated chain.
func (f *fetcherTester) insertChain(units modules.Units) (int, error) {
	f.lock.Lock()
	defer f.lock.Unlock()
	//fmt.Println(len(units))
	for i, unit := range units {
		// Make sure the parent in known
		//if _, ok := f.blocks[unit.ParentHash()[0]]; !ok {
		//	return i, errors.New("unknown parent")
		//}
		// Discard any new blocks if the same height already exists
		if unit.NumberU64() <= f.blocks[f.hashes[len(f.hashes)-1]].NumberU64() {
			return i, nil
		}
		// Otherwise build our current chain
		f.hashes = append(f.hashes, unit.Hash())
		f.blocks[unit.Hash()] = unit
	}
	return 0, nil
}

// dropPeer is an emulator for the peer removal, simply accumulating the various
// peers dropped by the fetcher.
func (f *fetcherTester) dropPeer(peer string) {
	f.lock.Lock()
	defer f.lock.Unlock()

	f.drops[peer] = true
}

// makeHeaderFetcher retrieves a block header fetcher associated with a simulated peer.
func (f *fetcherTester) makeHeaderFetcher(peer string, blocks map[common.Hash]*modules.Unit, drift time.Duration) headerRequesterFn {
	closure := make(map[common.Hash]*modules.Unit)
	for hash, block := range blocks {
		closure[hash] = block
	}
	// Create a function that return a header from the closure
	return func(hash common.Hash) error {
		// Gather the blocks to return
		headers := make([]*modules.Header, 0, 1)
		if block, ok := closure[hash]; ok {
			headers = append(headers, block.Header())
		}
		// Return on a new thread
		go f.fetcher.FilterHeaders(peer, headers, time.Now().Add(drift))

		return nil
	}
}

// makeBodyFetcher retrieves a block body fetcher associated with a simulated peer.
func (f *fetcherTester) makeBodyFetcher(peer string, blocks map[common.Hash]*modules.Unit, drift time.Duration) bodyRequesterFn {
	closure := make(map[common.Hash]*modules.Unit)
	for hash, block := range blocks {
		closure[hash] = block
	}
	// Create a function that returns blocks from the closure
	return func(hashes []common.Hash) error {
		// Gather the block bodies to return
		transactions := make([][]*modules.Transaction, 0, len(hashes))

		for _, hash := range hashes {
			if block, ok := closure[hash]; ok {
				transactions = append(transactions, block.Transactions())
			}
		}
		// Return on a new thread
		go f.fetcher.FilterBodies(peer, transactions, time.Now().Add(drift))

		return nil
	}
}

// verifyFetchingEvent verifies that one single event arrive on a fetching channel.
func verifyFetchingEvent(t *testing.T, fetching chan []common.Hash, arrive bool) {
	if arrive {
		select {
		case <-fetching:
		case <-time.After(time.Second):
			t.Fatalf("fetching timeout")
		}
	} else {
		select {
		case <-fetching:
			t.Fatalf("fetching invoked")
		case <-time.After(10 * time.Millisecond):
		}
	}
}

// verifyCompletingEvent verifies that one single event arrive on an completing channel.
func verifyCompletingEvent(t *testing.T, completing chan []common.Hash, arrive bool) {
	if arrive {
		select {
		case <-completing:
		case <-time.After(time.Second):
			t.Fatalf("completing timeout")
		}
	} else {
		select {
		case <-completing:
			t.Fatalf("completing invoked")
		case <-time.After(10 * time.Millisecond):
		}
	}
}

// verifyImportEvent verifies that one single event arrive on an import channel.
func verifyImportEvent(t *testing.T, imported chan *modules.Unit, arrive bool) {
	if arrive {
		select {
		case <-imported:
		case <-time.After(time.Second):
			t.Fatalf("import timeout")
		}
	} else {
		select {
		case <-imported:
			t.Fatalf("import invoked")
		case <-time.After(10 * time.Millisecond):
		}
	}
}

// verifyImportCount verifies that exactly count number of events arrive on an
// import hook channel.
func verifyImportCount(t *testing.T, imported chan *modules.Unit, count int) {
	for i := 0; i < count; i++ {
		select {
		case <-imported:
		case <-time.After(3 * time.Second):
			t.Fatalf("block %d: import timeout", i+1)
		}
	}
	verifyImportDone(t, imported)
}

// verifyImportDone verifies that no more events are arriving on an import channel.
func verifyImportDone(t *testing.T, imported chan *modules.Unit) {
	select {
	case <-imported:
		t.Fatalf("extra block imported")
	case <-time.After(50 * time.Millisecond):
	}
}

// Tests that a fetcher accepts block announcements and initiates retrievals for
// them, successfully importing into the local chain.
func TestSequentialAnnouncements1(t *testing.T) { testSequentialAnnouncements(t, 1) }
func testSequentialAnnouncements(t *testing.T, protocol int) {
	// Create a chain of blocks to import
	targetunits := 4 * hashLimit
	hashes, blocks := makeDag(targetunits, genesisUnit)
	tester := newTester()
	headerFetcher := tester.makeHeaderFetcher("valid", blocks, -gatherSlack)
	bodyFetcher := tester.makeBodyFetcher("valid", blocks, 0)
	// Iteratively announce blocks until all are imported
	imported := make(chan *modules.Unit)
	tester.fetcher.importedHook = func(block *modules.Unit) { imported <- block }

	for i := len(hashes) - 2; i >= 0; i-- {
		chain := &modules.ChainIndex{
			AssetID: modules.PTNCOIN,
			//IsMain:  true,
			Index: uint64(len(hashes) - i - 1),
		}
		tester.fetcher.Notify("valid", hashes[i], chain, time.Now().Add(-arriveTimeout), headerFetcher, bodyFetcher)
		verifyImportEvent(t, imported, true)
	}
	verifyImportDone(t, imported)
}

// Tests that if blocks are announced by multiple peers (or even the same buggy
// peer), they will only get downloaded at most once.
//func TestConcurrentAnnouncements1(t *testing.T) { testConcurrentAnnouncements(t, 1) }
func testConcurrentAnnouncements(t *testing.T, protocol int) {
	// Create a chain of blocks to import
	targetunits := 4 * hashLimit
	hashes, blocks := makeDag(targetunits, genesisUnit)
	// Assemble a tester with a built in counter for the requests
	tester := newTester()
	firstHeaderFetcher := tester.makeHeaderFetcher("first", blocks, -gatherSlack)
	firstBodyFetcher := tester.makeBodyFetcher("first", blocks, 0)
	secondHeaderFetcher := tester.makeHeaderFetcher("second", blocks, -gatherSlack)
	secondBodyFetcher := tester.makeBodyFetcher("second", blocks, 0)
	counter := uint32(0)
	firstHeaderWrapper := func(hash common.Hash) error {
		atomic.AddUint32(&counter, 1)
		return firstHeaderFetcher(hash)
	}
	secondHeaderWrapper := func(hash common.Hash) error {
		atomic.AddUint32(&counter, 1)
		return secondHeaderFetcher(hash)
	}
	// Iteratively announce blocks until all are imported
	imported := make(chan *modules.Unit)
	tester.fetcher.importedHook = func(block *modules.Unit) { imported <- block }
	for i := len(hashes) - 2; i >= 0; i-- {
		chain := &modules.ChainIndex{
			AssetID: modules.PTNCOIN,
			//IsMain:  true,
			Index: uint64(len(hashes) - i - 1),
		}
		tester.fetcher.Notify("first", hashes[i], chain, time.Now().Add(-arriveTimeout), firstHeaderWrapper, firstBodyFetcher)
		tester.fetcher.Notify("second", hashes[i], chain, time.Now().Add(-arriveTimeout+time.Millisecond), secondHeaderWrapper, secondBodyFetcher)
		tester.fetcher.Notify("second", hashes[i], chain, time.Now().Add(-arriveTimeout-time.Millisecond), secondHeaderWrapper, secondBodyFetcher)
		verifyImportEvent(t, imported, true)
	}
	verifyImportDone(t, imported)
	// Make sure no blocks were retrieved twice
	if int(counter) != targetunits {
		t.Fatalf("retrieval count mismatch: have %v, want %v", counter, targetunits)
	}
}

// Tests that announcements retrieved in a random order are cached and eventually
// imported when all the gaps are filled in.
//func TestRandomArrivalImport1(t *testing.T) { testRandomArrivalImport(t, 1) }
func testRandomArrivalImport(t *testing.T, protocol int) {
	// Create a chain of blocks to import, and choose one to delay
	targetunits := maxQueueDist
	hashes, blocks := makeDag(targetunits, genesisUnit)
	skip := targetunits / 2
	tester := newTester()
	headerFetcher := tester.makeHeaderFetcher("valid", blocks, -gatherSlack)
	bodyFetcher := tester.makeBodyFetcher("valid", blocks, 0)

	// Iteratively announce blocks, skipping one entry
	imported := make(chan *modules.Unit, len(hashes)-1)
	tester.fetcher.importedHook = func(block *modules.Unit) { imported <- block }

	for i := len(hashes) - 1; i >= 0; i-- {
		if i != skip {
			chain := &modules.ChainIndex{
				AssetID: modules.PTNCOIN,
				//IsMain:  true,
				Index: uint64(len(hashes) - i - 1),
			}
			tester.fetcher.Notify("valid", hashes[i], chain, time.Now().Add(-arriveTimeout), headerFetcher, bodyFetcher)
			time.Sleep(time.Millisecond)
		}
	}
	// Finally announce the skipped entry and check full import
	chainskip := &modules.ChainIndex{
		AssetID: modules.PTNCOIN,
		//IsMain:  true,
		Index: uint64(len(hashes) - skip - 1),
	}
	tester.fetcher.Notify("valid", hashes[skip], chainskip, time.Now().Add(-arriveTimeout), headerFetcher, bodyFetcher)
	verifyImportCount(t, imported, len(hashes)-1)
}

// Tests that direct block enqueues (due to block propagation vs. hash announce)
// are correctly schedule, filling and import queue gaps.
//func TestQueueGapFill1(t *testing.T) { testQueueGapFill(t, 1) }
func testQueueGapFill(t *testing.T, protocol int) {
	// Create a chain of blocks to import, and choose one to not announce at all
	targetunits := maxQueueDist
	hashes, blocks := makeDag(targetunits, genesisUnit)
	skip := targetunits / 2

	tester := newTester()
	headerFetcher := tester.makeHeaderFetcher("valid", blocks, -gatherSlack)
	bodyFetcher := tester.makeBodyFetcher("valid", blocks, 0)

	// Iteratively announce blocks, skipping one entry
	imported := make(chan *modules.Unit, len(hashes)-1)
	tester.fetcher.importedHook = func(block *modules.Unit) { imported <- block }

	for i := len(hashes) - 1; i >= 0; i-- {
		if i != skip {
			chain := &modules.ChainIndex{
				AssetID: modules.PTNCOIN,
				//IsMain:  true,
				Index: uint64(len(hashes) - i - 1),
			}
			tester.fetcher.Notify("valid", hashes[i], chain, time.Now().Add(-arriveTimeout), headerFetcher, bodyFetcher)
			time.Sleep(time.Millisecond)
		}
	}
	// Fill the missing block directly as if propagated
	tester.fetcher.Enqueue("valid", blocks[hashes[skip]])
	verifyImportCount(t, imported, len(hashes)-1)
}

// Tests that blocks arriving from various sources (multiple propagations, hash
//announces, etc) do not get scheduled for import multiple times.
func TestImportDeduplication1(t *testing.T) { testImportDeduplication(t, 1) }
func testImportDeduplication(t *testing.T, protocol int) {
	// Create two blocks to import (one for duplication, the other for stalling)
	hashes, blocks := makeDag(2, genesisUnit)

	// Create the tester and wrap the importer with a counter
	tester := newTester()
	headerFetcher := tester.makeHeaderFetcher("valid", blocks, -gatherSlack)
	bodyFetcher := tester.makeBodyFetcher("valid", blocks, 0)

	counter := uint32(0)
	tester.fetcher.insertChain = func(blocks modules.Units) (int, error) {
		atomic.AddUint32(&counter, uint32(len(blocks)))
		return tester.insertChain(blocks)
	}
	// Instrument the fetching and imported events
	fetching := make(chan []common.Hash)
	imported := make(chan *modules.Unit, len(hashes)-1)
	tester.fetcher.fetchingHook = func(hashes []common.Hash) { fetching <- hashes }
	tester.fetcher.importedHook = func(block *modules.Unit) { imported <- block }

	// Announce the duplicating block, wait for retrieval, and also propagate directly
	chain := &modules.ChainIndex{
		AssetID: modules.PTNCOIN,
		//IsMain:  true,
		Index: 1,
	}
	tester.fetcher.Notify("valid", hashes[0], chain, time.Now().Add(-arriveTimeout), headerFetcher, bodyFetcher)
	<-fetching

	tester.fetcher.Enqueue("valid", blocks[hashes[0]])
	tester.fetcher.Enqueue("valid", blocks[hashes[0]])
	tester.fetcher.Enqueue("valid", blocks[hashes[0]])

	// Fill the missing block directly as if propagated, and check import uniqueness
	tester.fetcher.Enqueue("valid", blocks[hashes[1]])
	verifyImportCount(t, imported, 2)

	if counter != 2 {
		t.Fatalf("import invocation count mismatch: have %v, want %v", counter, 2)
	}
}

// Tests that if a block is empty (i.e. header only), no body request should be
// made, and instead the header should be assembled into a whole block in itself.
func TestEmptyBlockShortCircuit1(t *testing.T) { testEmptyBlockShortCircuit(t, 1) }
func testEmptyBlockShortCircuit(t *testing.T, protocol int) {
	// Create a chain of blocks to import
	hashes, blocks := makeDag(32, genesisUnit)

	tester := newTester()
	headerFetcher := tester.makeHeaderFetcher("valid", blocks, -gatherSlack)
	bodyFetcher := tester.makeBodyFetcher("valid", blocks, 0)

	// Add a monitoring hook for all internal events
	fetching := make(chan []common.Hash)
	tester.fetcher.fetchingHook = func(hashes []common.Hash) { fetching <- hashes }

	completing := make(chan []common.Hash)
	tester.fetcher.completingHook = func(hashes []common.Hash) { completing <- hashes }

	imported := make(chan *modules.Unit)
	tester.fetcher.importedHook = func(block *modules.Unit) { imported <- block }

	// Iteratively announce blocks until all are imported
	for i := len(hashes) - 2; i >= 0; i-- {
		chain := &modules.ChainIndex{
			AssetID: modules.PTNCOIN,
			//IsMain:  true,
			Index: uint64(len(hashes) - i - 1),
		}
		tester.fetcher.Notify("valid", hashes[i], chain, time.Now().Add(-arriveTimeout), headerFetcher, bodyFetcher)

		// All announces should fetch the header
		verifyFetchingEvent(t, fetching, true)

		// Only blocks with data contents should request bodies
		verifyCompletingEvent(t, completing, len(blocks[hashes[i]].Transactions()) > 0)

		// Irrelevant of the construct, import should succeed
		verifyImportEvent(t, imported, true)
	}
	verifyImportDone(t, imported)
}

/*
// Tests that a peer is unable to use unbounded memory with sending infinite block announcements to a node, but that even in the face of such an attack, the fetcher remains operational.
func TestHashMemoryExhaustionAttack1(t *testing.T) { testHashMemoryExhaustionAttack(t, 1) }
func testHashMemoryExhaustionAttack(t *testing.T, protocol int) {
	// Create a tester with instrumented import hooks
	tester := newTester()

	imported, announces := make(chan *modules.Unit), int32(0)
	tester.fetcher.importedHook = func(block *modules.Unit) { imported <- block }
	tester.fetcher.announceChangeHook = func(hash common.Hash, added bool) {
		if added {
			atomic.AddInt32(&announces, 1)
		} else {
			atomic.AddInt32(&announces, -1)
		}
	}
	// Create a valid chain and an infinite junk chain
	targetUnits := hashLimit + 2*maxQueueDist
	hashes, blocks := makeDag(targetUnits, genesisUnit)
	validHeaderFetcher := tester.makeHeaderFetcher("valid", blocks, -gatherSlack)
	validBodyFetcher := tester.makeBodyFetcher("valid", blocks, 0)

	attack, _ := makeDag(targetUnits, genesisUnit)
	attackerHeaderFetcher := tester.makeHeaderFetcher("attacker", nil, -gatherSlack)
	attackerBodyFetcher := tester.makeBodyFetcher("attacker", nil, 0)

	// Feed the tester a huge hashset from the attacker, and a limited from the valid peer
	for i := 0; i < len(attack); i++ {
		if i < maxQueueDist {
			chain := modules.ChainIndex{
				AssetID: modules.PTNCOIN,
				IsMain:  true,
				Index:   uint64(i) + 1,
			}
			tester.fetcher.Notify("valid", hashes[len(hashes)-2-i], chain, time.Now(), validHeaderFetcher, validBodyFetcher)
		}
		chain := modules.ChainIndex{
			AssetID: modules.PTNCOIN,
			IsMain:  true,
			Index:   1,
		}
		//tester.fetcher.Notify("attacker", attack[i], chain  don't distance drop , time.Now(), attackerHeaderFetcher, attackerBodyFetcher)
		tester.fetcher.Notify("attacker", attack[i], chain , time.Now(), attackerHeaderFetcher, attackerBodyFetcher)
	}
	if count := atomic.LoadInt32(&announces); count != hashLimit+maxQueueDist {
		t.Fatalf("queued announce count mismatch: have %d, want %d", count, hashLimit+maxQueueDist)
	}
	// Wait for fetches to complete
	verifyImportCount(t, imported, maxQueueDist)

	// Feed the remaining valid hashes to ensure DOS protection state remains clean
	for i := len(hashes) - maxQueueDist - 2; i >= 0; i-- {
		chain := modules.ChainIndex{
			AssetID: modules.PTNCOIN,
			IsMain:  true,
			Index:   uint64(len(hashes) - i - 1),
		}
		tester.fetcher.Notify("valid", hashes[i], chain, time.Now().Add(-arriveTimeout), validHeaderFetcher, validBodyFetcher)
		verifyImportEvent(t, imported, true)
	}
	verifyImportDone(t, imported)
}

// Tests that blocks sent to the fetcher (either through propagation or via hash announces and retrievals) don't pile up indefinitely, exhausting available system memory.

func TestBlockMemoryExhaustionAttack(t *testing.T) {
	// Create a tester with instrumented import hooks
	tester := newTester()

	imported, enqueued := make(chan *modules.Unit), int32(0)
	tester.fetcher.importedHook = func(block *modules.Unit) { imported <- block }
	tester.fetcher.queueChangeHook = func(hash common.Hash, added bool) {
		if added {
			atomic.AddInt32(&enqueued, 1)
		} else {
			atomic.AddInt32(&enqueued, -1)
		}
	}
	// Create a valid chain and a batch of dangling (but in range) blocks
	targetBlocks := hashLimit + 2*maxQueueDist
	hashes, blocks := makeDag(targetBlocks, genesisUnit)
	attack := make(map[common.Hash]*modules.Unit)
	for i := byte(0); len(attack) < blockLimit+2*maxQueueDist; i++ {
		//hashes, blocks := makeChain(maxQueueDist-1, i, unknownBlock)
		hashes, blocks := makeDag(maxQueueDist-1, genesisUnit)
		for _, hash := range hashes[:maxQueueDist-2] {
			attack[hash] = blocks[hash]
		}
	}
	// Try to feed all the attacker blocks make sure only a limited batch is accepted
	for _, block := range attack {
		tester.fetcher.Enqueue("attacker", block)
	}
	time.Sleep(200 * time.Millisecond)
	if queued := atomic.LoadInt32(&enqueued); queued != blockLimit {
		t.Fatalf("queued block count mismatch: have %d, want %d", queued, blockLimit)
	}
	// Queue up a batch of valid blocks, and check that a new peer is allowed to do so
	for i := 0; i < maxQueueDist-1; i++ {
		tester.fetcher.Enqueue("valid", blocks[hashes[len(hashes)-3-i]])
	}
	time.Sleep(100 * time.Millisecond)
	if queued := atomic.LoadInt32(&enqueued); queued != blockLimit+maxQueueDist-1 {
		t.Fatalf("queued block count mismatch: have %d, want %d", queued, blockLimit+maxQueueDist-1)
	}
	// Insert the missing piece (and sanity check the import)
	tester.fetcher.Enqueue("valid", blocks[hashes[len(hashes)-2]])
	verifyImportCount(t, imported, maxQueueDist)

	// Insert the remaining blocks in chunks to ensure clean DOS protection
	for i := maxQueueDist; i < len(hashes)-1; i++ {
		tester.fetcher.Enqueue("valid", blocks[hashes[len(hashes)-2-i]])
		verifyImportEvent(t, imported, true)
	}
	verifyImportDone(t, imported)
}
*/
