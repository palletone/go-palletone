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
	if err := data0.Add(lastIrreUnit.UnitHash); err == nil {
		fork_index := make(ForkIndex)
		fork_index[uint64(0)] = data0
		memdag.forkIndex[assetid.String()] = fork_index
	}

	return memdag
}

func (chain *MemDag) validateMemory() bool {
	chain.chainLock.RLock()
	defer chain.chainLock.RUnlock()
	length := chain.memUnit.Lenth()
	//log.Info("MemDag", "validateMemory unit length:", length, "chain.memSize:", chain.memSize)
	if length >= uint64(chain.memSize) {
		return false
	}

	return true
}

func (chain *MemDag) Save(unit *modules.Unit) error {

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
	index, err := forkIndex.AddData(unit.Hash(), unit.ParentHash(), unit.UnitHeader.Index())
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
		forkData.Add(unit.Hash())
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
		log.Info("this is test line 1111..........................................  ", "index", index, "lastIndex", irreUnit.UnitHeader.Index())
		// set unit irreversible
		// unitHash := forkIndex.GetReachedIrreversibleHeightUnitHash(index)
		// prune fork if the irreversible height has been reached

		// save the matured unit into leveldb
		// @jay
		stable_hash := forkIndex.GetStableUnitHash(index)
		if stable_hash == (common.Hash{}) {
			log.Error("stable_hash is nil ..............")
			return errors.New("stable_hash is nil ..............")
		}

		stable_unit, err := chain.memUnit.Get(stable_hash)
		if err != nil {
			return err
		}
		if err := chain.unitRep.SaveUnit(stable_unit, false); err != nil {
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
		// if err := chain.Prune(assetId, stable_hash); err != nil {
		// 	log.Error("Check if the irreversible height has been reached", "error", err.Error())
		// 	return err
		// }
		// go chain.Prune(assetId, stable_hash)
	} else {
		// TODO save unit into memUnit , update  world state index.
		if !chain.Exists(unit.Hash()) {
			err := chain.memUnit.Add(unit)
			if err != nil {
				log.Error("memUnit add unit is failed.", "error", err)
				return err
			}
		} else {
			log.Debug("--------save memunit is success------")
		}
	}
	chain.forkIndex[assetId] = forkIndex

	for key, val := range forkIndex {
		log.Debug("forkIndex Info ---->>>  ", "key", key)
		log.Debug("forkIndex Info ---->>>  ", "key", val)
	}
	return nil
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
func (chain *MemDag) Prune(assetId string, maturedUnitHash common.Hash) error {
	chain.chainLock.Lock()
	defer chain.chainLock.Unlock()
	// get fork index

	index, subindex := chain.QueryIndex(assetId, maturedUnitHash)
	if index < 0 {
		return fmt.Errorf("Prune error: matured unit is not found in memory")
	}

	// save all the units before matured unit into db
	// @jay
	// forkdata := (*(chain.forkIndex[assetId])) [index]
	fork_index := chain.forkIndex[assetId]

	forkdata, has := fork_index[index]
	if !has {
		return fmt.Errorf("memUnit get forkData is failed, error hash: %x , index%d", maturedUnitHash, index)
	}
	for i := 0; i < subindex; i++ {
		unitHash := (forkdata)[i]
		//unit := (*chain.memUnit)[unitHash]
		unit, err := chain.memUnit.Get(unitHash)
		if err != nil {
			return fmt.Errorf("memUnit get unithash(%v) error:%s ", unitHash, err.Error())
		}
		if err := chain.unitRep.SaveUnit(unit, false); err != nil {
			return fmt.Errorf("Prune error when save unit: %s", err.Error())
		}
	}

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
	if lenth := len(forkdata); lenth > 0 {
		delete(fork_index, index)
		// prune other forks

		fmt.Println("------------------------------ new forkindex lenth>0 --------------------------------------", fork_index, index, lenth, subindex)
		chain.forkIndex[assetId] = fork_index
	} else {
		fmt.Println("------------------------------ new forkindex--------------------------------------", index, lenth, subindex)
	}
	// save the matured unit
	// chain.lastValidatedUnit[assetId] = maturedUnitHash

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
		for index, forkdata := range forkindex {
			if len(forkdata) > 0 {
				// chain.mainChain[assetid] = index
				// maxLenth = len(*forkdata)
				maxIndex = index
			}
		}
		main_data.Number = maxIndex
		chain.mainChain[assetid] = main_data
	}
	return nil
}

func (chain *MemDag) QueryIndex(assetId string, maturedUnitHash common.Hash) (uint64, int) {
	chain.chainLock.RLock()

	forkindex, ok := chain.forkIndex[assetId]
	if !ok {
		return 0, -1
	}
	chain.chainLock.RUnlock()
	for index, forkdata := range forkindex {
		for subindex, unitHash := range forkdata {
			if strings.Compare(unitHash.String(), maturedUnitHash.String()) == 0 {
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
