/*
 *
 *     This file is part of go-palletone.
 *     go-palletone is free software: you can redistribute it and/or modify
 *     it under the terms of the GNU General Public License as published by
 *     the Free Software Foundation, either version 3 of the License, or
 *     (at your option) any later version.
 *     go-palletone is distributed in the hope that it will be useful,
 *     but WITHOUT ANY WARRANTY; without even the implied warranty of
 *     MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *     GNU General Public License for more details.
 *     You should have received a copy of the GNU General Public License
 *     along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
 * /
 *
 *  * @author PalletOne core developers <dev@pallet.one>
 *  * @date 2018
 *
 */

package memunit

import (
	"fmt"
	"strings"
	"sync"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	dagCommon "github.com/palletone/go-palletone/dag/common"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/errors"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/storage"
	"github.com/palletone/go-palletone/dag/txspool"
)

/*********************************************************************/
// TODO MemDag
type MemDag struct {
	dagdb             storage.IDagDb
	statedb           storage.IStateDb
	unitRep           dagCommon.IUnitRepository
	lastValidatedUnit map[string]*modules.Unit // the key is asset id
	forkIndex         map[string]ForkIndex     // the key is asset id
	mainChain         map[string]*MainData     // the key is asset id, value is fork index
	currentUnit       map[string]*modules.Unit // current added  unit in memdag
	memUnit           *MemUnit
	memSize           uint
	chainLock         sync.RWMutex
	delhashs          chan common.Hash
}

func NewMemDag(db storage.IDagDb, sdb storage.IStateDb, unitRep dagCommon.IUnitRepository) *MemDag {
	//fork_index := make(ForkIndex)
	memdag := &MemDag{
		lastValidatedUnit: make(map[string]*modules.Unit),
		forkIndex:         make(map[string]ForkIndex),
		memUnit:           InitMemUnit(),
		memSize:           dagconfig.DefaultConfig.MemoryUnitSize,
		dagdb:             db,
		unitRep:           unitRep,
		mainChain:         make(map[string]*MainData),
		currentUnit:       make(map[string]*modules.Unit),
		statedb:           sdb,
		delhashs:          make(chan common.Hash, 15),
	}

	// get genesis Last Irreversible Unit
	genesisUnit, err := unitRep.GetGenesisUnit(0)
	if err != nil {
		log.Error("NewMemDag when GetGenesisUnit", "error", err.Error())
		return nil
	}
	if genesisUnit == nil {
		log.Error("Get genesis unit failed, unit of genesis is nil.")
		return nil
	}
	assetid := genesisUnit.UnitHeader.Number.AssetID
	lastIrreUnit, _ := db.GetLastIrreversibleUnit(assetid)
	if lastIrreUnit != nil {
		memdag.lastValidatedUnit[assetid.String()] = lastIrreUnit
	}
	memdag.currentUnit[assetid.String()] = lastIrreUnit
	main_data := new(MainData)
	main_data.Index = lastIrreUnit.UnitHeader.ChainIndex()
	main_data.Hash = &lastIrreUnit.UnitHash
	main_data.Number = lastIrreUnit.UnitHeader.Index()
	memdag.mainChain[assetid.String()] = main_data

	data0 := make(ForkData, 0)
	if err := data0.Add(lastIrreUnit.UnitHash, lastIrreUnit.UnitHeader.Authors.Address.String()); err == nil {
		fork_index := make(ForkIndex)
		fork_index[uint64(0)] = data0
		memdag.forkIndex[assetid.String()] = fork_index
	}

	return memdag
}

func NewMemDagForTest(db storage.IDagDb, sdb storage.IStateDb, unitRep dagCommon.IUnitRepository, txpool txspool.ITxPool) *MemDag {
	memdag := &MemDag{
		lastValidatedUnit: make(map[string]*modules.Unit),
		forkIndex:         make(map[string]ForkIndex),
		memUnit:           InitMemUnit(),
		memSize:           dagconfig.DefaultConfig.MemoryUnitSize,
		dagdb:             db,
		unitRep:           unitRep,
		mainChain:         make(map[string]*MainData),
		currentUnit:       make(map[string]*modules.Unit),
		statedb:           sdb,
		delhashs:          make(chan common.Hash, 15),
	}

	// get genesis Last Irreversible Unit
	unit, err := createUnitForTest()

	unitRep.SaveUnit(unit, txpool, true)

	genesisUnit, err := unitRep.GetGenesisUnit(0)
	if err != nil {
		log.Error("NewMemDag when GetGenesisUnit", "error", err.Error())
		return nil
	}
	if genesisUnit == nil {
		log.Error("Get genesis unit failed, unit of genesis is nil.")
		return nil
	}
	assetid := genesisUnit.UnitHeader.Number.AssetID
	lastIrreUnit, _ := db.GetLastIrreversibleUnit(assetid)
	if lastIrreUnit != nil {
		memdag.lastValidatedUnit[assetid.String()] = lastIrreUnit
	}
	memdag.currentUnit[assetid.String()] = lastIrreUnit
	main_data := new(MainData)
	main_data.Index = lastIrreUnit.UnitHeader.ChainIndex()
	main_data.Hash = &lastIrreUnit.UnitHash
	main_data.Number = lastIrreUnit.UnitHeader.Index()
	memdag.mainChain[assetid.String()] = main_data

	data0 := make(ForkData, 0)
	if err := data0.Add(lastIrreUnit.UnitHash, lastIrreUnit.UnitHeader.Authors.Address.String()); err == nil {
		fork_index := make(ForkIndex)
		fork_index[uint64(0)] = data0
		memdag.forkIndex[assetid.String()] = fork_index
	}

	return memdag
}
func createUnitForTest() (*modules.Unit, error) {
	asset := new(modules.Asset)
	asset.AssetId = modules.PTNCOIN
	asset.UniqueId = modules.PTNCOIN
	asset.ChainId = 1
	// new payload tpl payload
	inputs := make([]*modules.Input, 0)
	in := new(modules.Input)
	in.Extra = []byte("jay")
	inputs = append(inputs, in)
	outputs := make([]*modules.Output, 0)
	out := new(modules.Output)
	out.Value = 1100000000
	out.Asset = asset
	outputs = append(outputs, out)
	payment := modules.NewPaymentPayload(inputs, outputs)
	msg0 := modules.NewMessage(modules.APP_PAYMENT, payment)
	tplPayload := modules.NewContractTplPayload([]byte("contract_template0000"),
		"TestContractTpl", "./contract", "1.1.1", 1024,
		[]byte{175, 52, 23, 180, 156, 109, 17, 232, 166, 226, 84, 225, 173, 184, 229, 159})
	// new msg
	msg := modules.NewMessage(modules.APP_CONTRACT_TPL, tplPayload)
	msgs := []*modules.Message{msg0}
	// new transactions
	tx := modules.NewTransaction(msgs[:])
	tx1 := modules.NewTransaction(append(msgs, msg))
	tx1 = tx1
	txs := modules.Transactions{tx}
	// new unit

	unit, err := dagCommon.NewGenesisUnit(txs, 1536451201, asset)
	log.Info("create unit success.", "error", err, "hash", unit.Hash().String())
	return unit, err
}

func (chain *MemDag) GetDelhashs() chan common.Hash {
	if chain.delhashs == nil {
		chain.delhashs = make(chan common.Hash, 15)
	}

	return chain.delhashs
}
func (chain *MemDag) PushDelHashs(hashs []common.Hash) {
	if chain.delhashs == nil {
		chain.delhashs = make(chan common.Hash, 15)
	}
	for _, hash := range hashs {
		chain.delhashs <- hash
	}
}

func (chain *MemDag) validateMemory() bool {
	chain.chainLock.RLock()
	defer chain.chainLock.RUnlock()
	length := chain.memUnit.Lenth()
	//log.Debug("MemDag", "validateMemory unit length:", length, "chain.memSize:", chain.memSize)
	if length >= uint64(chain.memSize) {
		return false
	}

	return true
}

func (chain *MemDag) Save(unit *modules.Unit, txpool txspool.ITxPool) error {
	if unit == nil {
		return fmt.Errorf("Save mem unit: unit is null")
	}
	chain.chainLock.Lock()
	defer chain.chainLock.Unlock()
	if chain.memUnit.Exists(unit.Hash()) {
		return fmt.Errorf("Save mem unit: unit is already exists in memory")
	}

	//TODO must recover
	//if !chain.validateMemory() {
	//	return fmt.Errorf("Save mem unit: size is out of limit")
	//}

	assetId := unit.UnitHeader.Number.AssetID.String()

	// save fork index
	forkIndex, ok := chain.forkIndex[assetId]
	if !ok {
		// create forindex
		chain.forkIndex[assetId] = make(map[uint64]ForkData)
		forkIndex = chain.forkIndex[assetId]
	}
	if forkIndex == nil {
		forkIndex = make(map[uint64]ForkData)
	}
	// get asset chain's las irreversible unit
	irreUnit, ok := chain.lastValidatedUnit[assetId]
	if !ok {
		lastIrreUnit, _ := chain.dagdb.GetLastIrreversibleUnit(unit.UnitHeader.Number.AssetID)
		if lastIrreUnit != nil {
			irreUnit = lastIrreUnit
			chain.lastValidatedUnit[assetId] = irreUnit
		}
	}
	// save unit to index
	index, err := forkIndex.AddData(unit.Hash(), unit.ParentHash(), unit.UnitHeader.Index(), unit.UnitHeader.Authors.Address.String())
	switch index {
	case -1:
		log.Error("errrrorrrrrrrrrrrrrrrrrrrrrrrrrrrrrr", "error", err)
		return err
	case -2:
		// check last irreversible unit
		// if it is not null, check continuously
		// 测试utxo转账 ，暂时隐藏-----
		// if strings.Compare(irreUnitHash.String(), "") != 0 {
		// 	if common.CheckExists(irreUnitHash, unit.UnitHeader.ParentsHash) < 0 {
		// 		return fmt.Errorf("The unit(%s) is not continious.", unit.UnitHash.String())
		// 	}
		// }
		// add new fork into index
		if common.CheckExists(irreUnit.UnitHash, unit.UnitHeader.ParentsHash) < 0 {
			log.Info(fmt.Sprintf("xxxxxxxxxxxxxxxx   The unit(%s) is not continious. index:(%d) ", unit.Hash().String(), unit.UnitHeader.ChainIndex().Index))
		}
		forkData := make(ForkData, 0)
		forkData.Add(unit.Hash(), unit.UnitHeader.Authors.Address.String())
		// index = int64(len(forkIndex))
		forkIndex[unit.UnitHeader.Index()] = forkData
		log.Info(fmt.Sprintf(".............. The unit(%s) is not continious.%v", unit.Hash().String(), forkData))
	default:
		log.Info("forkindex add unit is success.", "index", index)
	}
	// save memory unit
	if err := chain.memUnit.Add(unit); err != nil {
		return err
	} else {
		chain.currentUnit[assetId] = unit
	}

	//save chainindex mapping unit hash
	chain.memUnit.SetHashByNumber(unit.Number(), unit.Hash())

	// Check if the irreversible height has been reached

	if forkIndex.IsReachedIrreversibleHeight(uint64(index), irreUnit.UnitHeader.Index()) {
		log.Info("IsReachedIrreversibleUnit  .......................................  ", "index", index, "lastIndex", irreUnit.UnitHeader.Index())
		// set unit irreversible
		// unitHash := forkIndex.GetReachedIrreversibleHeightUnitHash(index)
		// prune fork if the irreversible height has been reached

		// save the matured unit into leveldb
		// @jay
		hashs := forkIndex.GetStableUnitHash(index)
		var stable_hash common.Hash
		if len(hashs) > 0 {
			stable_hash = hashs[0]
		}
		if len(hashs) > 1 {
			hashs = hashs[1:]
			chain.PushDelHashs(hashs[:])
		}
		if stable_hash == (common.Hash{}) {
			log.Error("stable_hash is nil ..............")
			return errors.New("stable_hash is nil ..............")
		}

		stable_unit, err := chain.memUnit.Get(stable_hash)
		if err != nil {
			return err
		}
		if err := chain.unitRep.SaveUnit(stable_unit, txpool, false); err != nil {
			log.Error("save the matured unit into leveldb", "error", err.Error(), "hash", stable_unit.UnitHash.String(), "index", index)
			return err
		} else {
			// 更新memUnit
			chain.lastValidatedUnit[assetId] = stable_unit
			chain.memUnit.Refresh(stable_hash)
			current_index, _ := chain.statedb.GetCurrentChainIndex(stable_unit.UnitHeader.ChainIndex().AssetID)
			chain_index := unit.UnitHeader.ChainIndex()
			if chain_index.Index > current_index.Index {
				chain.statedb.SaveChainIndex(chain_index)
			}
			log.Info("+++++++++++++++++++++++ save_memDag_success +++++++++++++++++++++++", "save_memDag_Unit_hash", unit.Hash().String(), "index", index)
		}
		if err := chain.Prune(assetId, hashs); err != nil {
			log.Error("Check if the irreversible height has been reached", "error", err.Error())
			return err
		}
		// go chain.Prune(assetId, stable_hash,)
	} else {
		// TODO save unit into memUnit , update  world state index.
		if !chain.Exists(unit.Hash()) {
			err := chain.memUnit.Add(unit)
			if err != nil {
				log.Error("memUnit add unit is failed.", "error", err)
				return err
			}
		}
	}
	chain.forkIndex[assetId] = forkIndex

	for key, val := range forkIndex {
		log.Debug("forkIndex Info ---->>>  ", "key", key)
		log.Debug("forkIndex Info ---->>>  ", "key", val)
	}
	return nil
}

func (chain *MemDag) updateMemdag(unit *modules.Unit, txpool txspool.ITxPool) error {
	asstid := unit.UnitHeader.ChainIndex().AssetID
	if !chain.Exists(unit.Hash()) {
		return nil
	}

	// 保存单元
	err := chain.unitRep.SaveUnit(unit, txpool, false)
	// 将该unit 从memdag中剔除
	if err == nil {
		// 1. refresh memUnit
		go chain.memUnit.Refresh(unit.Hash())

		// 2. update lastValidatedUnit
		old := chain.lastValidatedUnit[asstid.String()]
		if old.UnitHeader.Index() < unit.UnitHeader.Index() {
			chain.lastValidatedUnit[asstid.String()] = unit

			// update mainData
			index := unit.UnitHeader.ChainIndex()
			mainHash := unit.Hash()
			chain.mainChain[asstid.String()] = &MainData{Index: index, Hash: &mainHash, Number: index.Index}
		}

		// 3.update fork(对分叉数据剪枝)

	}
	// 更新某些状态
	// 当前最新区块高度是否小于此unit高度。
	// get currentUnit

	curChainIndex, err := chain.statedb.GetCurrentChainIndex(unit.UnitHeader.ChainIndex().AssetID)
	if err == nil {
		if curChainIndex.Index < unit.UnitHeader.Index() {
			// update state
			chain.dagdb.PutCanonicalHash(unit.UnitHash, unit.NumberU64())
			chain.dagdb.PutHeadHeaderHash(unit.UnitHash)
			chain.dagdb.PutHeadUnitHash(unit.UnitHash)
			chain.dagdb.PutHeadFastUnitHash(unit.UnitHash)
		}
	}
	// 分支修剪
	go chain.Prune(unit.UnitHeader.ChainIndex().AssetID.String(), []common.Hash{unit.Hash()})

	// 将该单元的父单元一并确认群签
	if list := unit.ParentHash(); len(list) > 0 {
		for _, h := range list {
			if chain.Exists(h) {
				chain.UpdateMemDag(h, unit.UnitHeader.GroupSign[:], txpool)
			}
		}
	}

	return nil
}

func (chain *MemDag) UpdateMemDag(hash common.Hash, sign []byte, txpool txspool.ITxPool) error {
	chain.chainLock.Lock()
	defer chain.chainLock.Unlock()

	unit, err := chain.GetUnit(hash)
	if err != nil {
		return err
	}
	unit.SetGroupSign(sign[:])
	return chain.updateMemdag(unit, txpool)
}

func (chain *MemDag) Exists(uHash common.Hash) bool {
	if chain.memUnit.Exists(uHash) {
		return true
	}
	return false
}

/**
对分叉数据进行剪支
Prune fork data
*/
func (chain *MemDag) Prune(assetId string, delhashs []common.Hash) error {
	// @jay
	if len(delhashs) < 1 {
		return nil
	}
	// get fork index
	maturedUnitHash := delhashs[0]

	chain.chainLock.Lock()
	defer chain.chainLock.Unlock()

	index, subindex := chain.QueryIndex(assetId, maturedUnitHash)
	if index < 0 {
		return fmt.Errorf("Prune error: matured unit is not found in memory")
	}

	fork_index := chain.forkIndex[assetId]

	forkdata, has := fork_index[index]
	if !has {
		return fmt.Errorf("memUnit get forkData is failed, error hash: %x , index%d", maturedUnitHash, index)
	}
	data1 := make(ForkData, 0)
	for index, data := range forkdata {
		if index == subindex {
			data1.Add(data.hash, data.addr)
			// forkdata[index] = data1
			break
		}
	}
	fork_index[index] = data1

	// rollback transaction pool
	for j := subindex; j < len(forkdata); j++ {

	}

	if unit, err := chain.memUnit.Get(maturedUnitHash); err == nil && unit != nil {
		main_data := new(MainData)
		main_data.Hash = &maturedUnitHash
		main_data.Index = unit.UnitHeader.ChainIndex()
		main_data.Number = unit.UnitHeader.Index()
		chain.mainChain[assetId] = main_data
	}

	// refresh forkindex
	for _, hash := range delhashs {
		if unit, err := chain.memUnit.Get(hash); err == nil {
			// 1. update forkindex
			fork_index.RemoveStableIndex(unit.UnitHeader.Index())
			chain.forkIndex[unit.UnitHeader.ChainIndex().AssetID.String()] = fork_index
			// 2. memUnit chainIndex
			go chain.memUnit.DelHashByNumber(*unit.UnitHeader.ChainIndex())
		}
	}

	return nil
}

/**
切换主链：将最长链作为主链
Switch to the longest fork
*/
func (chain *MemDag) SwitchMainChain() error {
	chain.chainLock.Lock()
	defer chain.chainLock.Unlock()
	// chose the longest fork as the main chain
	for assetid, forkindex := range chain.forkIndex {
		var maxIndex uint64
		main_data := new(MainData)
		var hash common.Hash
		for index, forkdata := range forkindex {
			if len(forkdata) > 0 {
				if maxIndex < index {
					maxIndex = index
					hash = forkdata.GetLast()
				}
			}
		}
		main_data.Number = maxIndex
		main_data.Hash = &hash
		chain.mainChain[assetid] = main_data
	}
	return nil
}

func (chain *MemDag) QueryIndex(assetId string, maturedUnitHash common.Hash) (uint64, int) {
	chain.chainLock.RLock()
	defer chain.chainLock.RUnlock()

	forkindex, ok := chain.forkIndex[assetId]
	if !ok {
		return 0, -1
	}

	for index, forkdata := range forkindex {
		for subindex, data := range forkdata {
			if strings.Compare(data.hash.String(), maturedUnitHash.String()) == 0 {
				return index, subindex
			}
		}
	}
	return 0, -1
}

func (chain *MemDag) GetCurrentUnit(assetid modules.IDType16, index uint64) (*modules.Unit, error) {
	sAssetID := assetid.String()
	// chain.chainLock.RLock()
	// defer chain.chainLock.RUnlock()
	// to get from lastValidatedUnit
	lastValidatedUnit, has := chain.lastValidatedUnit[sAssetID]
	if !has {
		log.Debug("memdag's lastValidated Unit is null.")
	}

	currentUnit, ok := chain.currentUnit[sAssetID]

	if ok {
		if currentUnit.UnitHeader.Index() >= index {
			return currentUnit, nil
		}
	}
	fork, has := chain.forkIndex[sAssetID]
	if !has {
		return nil, fmt.Errorf("MemDag.GetCurrentUnit currented error, forkIndex has no asset(%s) info.", assetid.String())
	}

	forkdata := fork[index]
	if forkdata != nil {
		curHash := forkdata.GetLast()
		if curHash == (common.Hash{}) {
			return nil, fmt.Errorf("forkdata getLast failed,curHash is null,the index(%d)", index)
		}
		curUnit, err := chain.memUnit.Get(curHash)
		if err != nil {
			return nil, fmt.Errorf("MemDag.GetCurrentUnit error: get no unit hash(%s) in memUnit,error(%s)", curHash.String(), err.Error())
		}
		if curUnit.UnitHeader.Index() >= index {
			return curUnit, nil
		} else {
			return nil, fmt.Errorf("memdag's current unit is old， cur_index(%d), index(%d)", curUnit.UnitHeader.Index(), index)
		}

	}
	// return lastValidatedUnit
	if currentUnit == nil {
		return lastValidatedUnit, nil
	}
	return currentUnit, nil
}

func (chain *MemDag) GetCurrentUnitChainIndex(assetid modules.IDType16, index uint64) (*modules.ChainIndex, error) {
	chain.chainLock.RLock()
	unit, err := chain.GetCurrentUnit(assetid, index)

	chain.chainLock.RUnlock()
	if err != nil {
		return nil, err
	}
	if unit == nil {
		return nil, errors.New("GetCurrentUnitChainIndex failed.")
	}
	chainIndex := unit.UnitHeader.ChainIndex()

	return chainIndex, nil
}

func (chain *MemDag) GetUnit(hash common.Hash) (*modules.Unit, error) {
	return chain.memUnit.Get(hash)
}
