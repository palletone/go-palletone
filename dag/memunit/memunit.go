/*
   This file is part of go-palletone.
   go-palletone is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.
   go-palletone is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.
   You should have received a copy of the GNU General Public License
   along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
*/
/*
 * @author PalletOne core developers <dev@pallet.one>
 * @date 2018
 */
package memunit

import (
	"fmt"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/ptndb"
	dagCommon "github.com/palletone/go-palletone/dag/common"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/storage"
	"strings"
)

// non validated units set
type MemUnit map[common.Hash]*modules.Unit

func InitMemUnit() *MemUnit {
	return &MemUnit{}
}

func (mu *MemUnit) Add(u *modules.Unit) error {
	_, ok := (*mu)[u.UnitHash]
	if !ok {
		(*mu)[u.UnitHash] = u
	}
	return nil
}

func (mu *MemUnit) Get(u *modules.Unit) (*modules.Unit, error) {
	unit, ok := (*mu)[u.UnitHash]
	if !ok {
		return nil, fmt.Errorf("Get mem unit: unit does not be found.")
	}
	return unit, nil
}

func (mu *MemUnit) Exists(hash common.Hash) bool {
	_, ok := (*mu)[hash]
	if ok {
		return true
	}
	return false
}

func (mu *MemUnit) Lenth() uint64 {
	return uint64(len(*mu))
}

/*********************************************************************/
// fork data
type ForkData []common.Hash

func (f *ForkData) Add(hash common.Hash) error {
	if f.Exists(hash) {
		return fmt.Errorf("Save fork data: unit is already existing.")
	}
	*f = append(*f, hash)
	return nil
}

func (f *ForkData) Exists(hash common.Hash) bool {
	for _, uid := range *f {
		if strings.Compare(uid.String(), hash.String()) == 0 {
			return true
		}
	}
	return false
}

/*********************************************************************/
// forkIndex
type ForkIndex []*ForkData

func (forkIndex *ForkIndex) AddData(unitHash common.Hash, parentsHash []common.Hash) (int, error) {
	for index, fi := range *forkIndex {
		lenth := len(*fi)
		if lenth <= 0 {
			continue
		}
		if common.CheckExists((*fi)[lenth-1], parentsHash) >= 0 {
			if err := (*fi).Add(unitHash); err != nil {
				return -1, err
			}
			return int(index), nil
		}
	}
	return -2, fmt.Errorf("Unit(%s) is not continuously", unitHash)
}

func (forkIndex *ForkIndex) IsReachedIrreversibleHeight(index int) bool {
	if index < 0 {
		return false
	}
	if len(*(*forkIndex)[index]) >= dagconfig.DefaultConfig.IrreversibleHeight {
		return true
	}
	return false
}

func (forkIndex *ForkIndex) GetReachedIrreversibleHeightUnitHash(index int) common.Hash {
	if index < 0 {
		return common.Hash{}
	}
	return (*(*forkIndex)[index])[0]
}

func (forkIndex *ForkIndex) Lenth() int {
	return len(*forkIndex)
}

/*********************************************************************/
// TODO MemDag
type MemDag struct {
	db                ptndb.Database
	lastValidatedUnit map[string]common.Hash // the key is asset id
	forkIndex         map[string]*ForkIndex  // the key is asset id
	mainChain         map[string]int         // the key is asset id
	memUnit           *MemUnit
	memSize           uint8
}

func InitMemDag(db ptndb.Database) *MemDag {
	memdag := MemDag{
		lastValidatedUnit: nil,
		forkIndex:         map[string]*ForkIndex{},
		memUnit:           InitMemUnit(),
		memSize:           dagconfig.DefaultConfig.MemoryUnitSize,
	}
	memdag.db = db
	return &memdag
}

func (chain *MemDag) validateMemory() bool {
	if chain.memUnit.Lenth() >= uint64(chain.memSize) {
		return false
	}
	return true
}

func (chain *MemDag) Save(unit *modules.Unit) error {
	if unit == nil {
		return fmt.Errorf("Save mem unit: unit is null")
	}
	if chain.memUnit.Exists(unit.UnitHash) {
		return fmt.Errorf("Save mem unit: unit is already exists in memory")
	}
	if !chain.validateMemory() {
		return fmt.Errorf("Save mem unit: size is out of limit")
	}

	assetId := unit.UnitHeader.Number.AssetID.String()

	// save fork index
	forkIndex, ok := chain.forkIndex[assetId]
	if !ok {
		// create forindex
		chain.forkIndex[assetId] = &ForkIndex{}
		forkIndex = chain.forkIndex[assetId]
	}

	// get asset chain's las irreversible unit
	irreUnitHash, ok := chain.lastValidatedUnit[assetId]
	if !ok {
		lastIrreUnit := storage.GetLastIrreversibleUnit(chain.db, unit.UnitHeader.Number.AssetID)
		if lastIrreUnit != nil {
			irreUnitHash = lastIrreUnit.UnitHash
		}
	}
	// save unit to index
	index, err := forkIndex.AddData(unit.UnitHash, unit.UnitHeader.ParentsHash)
	switch index {
	case -1:
		return err
	case -2:
		// check last irreversible unit
		// if it is not null, check continuously
		if strings.Compare(irreUnitHash.String(), "") != 0 {
			if common.CheckExists(irreUnitHash, unit.UnitHeader.ParentsHash) < 0 {
				return fmt.Errorf("The unit(%s) is not continious.", unit.UnitHash)
			}
		}
		// add new fork into index
		forkData := ForkData{}
		forkData = append(forkData, unit.UnitHash)
		index = len(*forkIndex)
		*forkIndex = append(*forkIndex, &forkData)
	}
	// save memory unit
	if err := chain.memUnit.Add(unit); err != nil {
		return err
	}
	// Check if the irreversible height has been reached
	if forkIndex.IsReachedIrreversibleHeight(index) {
		// set unit irreversible
		unitHash := forkIndex.GetReachedIrreversibleHeightUnitHash(index)
		// prune fork if the irreversible height has been reached
		if err := chain.Prune(assetId, unitHash); err != nil {
			return err
		}
		// save the matured unit into leveldb
		if err := dagCommon.SaveUnit(chain.db, *unit, false); err != nil {
			return err
		}
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
	// get fork index
	index, subindex := chain.QueryIndex(assetId, maturedUnitHash)
	if index < 0 {
		return fmt.Errorf("Prune error: matured unit is not found in memory")
	}
	// save all the units before matured unit into db
	forkdata := (*(chain.forkIndex[assetId]))[index]
	for i := 0; i < subindex; i++ {
		unitHash := (*forkdata)[i]
		unit := (*chain.memUnit)[unitHash]
		if err := dagCommon.SaveUnit(chain.db, *unit, false); err != nil {
			return fmt.Errorf("Prune error when save unit: ", err.Error())
		}
	}
	// rollback transaction pool

	// refresh forkindex
	if lenth := len(*forkdata); lenth > subindex {
		newForkData := ForkData{}
		for i := subindex + 1; i < lenth; i++ {
			newForkData = append(newForkData, (*forkdata)[i])
		}
		// prune other forks
		newForkindex := ForkIndex{}
		newForkindex = append(newForkindex, &newForkData)
		chain.forkIndex[assetId] = &newForkindex
	}
	// save the matured unit
	chain.lastValidatedUnit[assetId] = maturedUnitHash

	return nil
}

/**
切换主链：将最长链作为主链
Switch to the longest fork
*/
func (chain *MemDag) SwitchMainChain() error {
	// chose the longest fork as the main chain
	for assetid, forkindex := range chain.forkIndex {
		maxLenth := 0
		for index, forkdata := range *forkindex {
			if len(*forkdata) > maxLenth {
				chain.mainChain[assetid] = index
				maxLenth = len(*forkdata)
			}
		}
	}
	return nil
}

func (chain *MemDag) QueryIndex(assetId string, maturedUnitHash common.Hash) (int, int) {
	forkindex, ok := chain.forkIndex[assetId]
	if !ok {
		return -1, -1
	}
	for index, forkdata := range *forkindex {
		for subindex, unitHash := range *forkdata {
			if strings.Compare(unitHash.String(), maturedUnitHash.String()) == 0 {
				return index, subindex
			}
		}
	}
	return -1, -1
}
