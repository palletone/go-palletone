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
package modules

import (
	"fmt"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"strings"
)

// non validated units set
type MemUnit map[common.Hash]*Unit

func InitMemUnit() *MemUnit {
	return &MemUnit{}
}

func (mu *MemUnit) Add(u *Unit) error {
	_, ok := (*mu)[u.UnitHash]
	if !ok {
		(*mu)[u.UnitHash] = u
	}
	return nil
}

func (mu *MemUnit) Get(u *Unit) (*Unit, error) {
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

func initForkData() *ForkData {
	forkdata := ForkData{}
	return &forkdata
}
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

func (forkIndex *ForkIndex) AddData(index int, unitHash common.Hash) error {
	if index < 0 || index > forkIndex.Lenth() {
		return fmt.Errorf("Add fork data error: index is invalid")
	}
	forkData := (*forkIndex)[index]
	*forkData = append(*forkData, unitHash)
	return nil
}

func (forkIndex *ForkIndex) Lenth() int {
	return len(*forkIndex)
}

/*********************************************************************/
type MemDag struct {
	forkId            map[string]int8 // fork chain id
	lastValidatedUnit *Unit
	forkIndex         map[string]*ForkIndex
	forkData          *ForkData
	memUnit           *MemUnit
	memSize           uint8
}

func InitMemDag() *MemDag {
	memdag := MemDag{
		forkId:            map[string]int8{},
		lastValidatedUnit: nil,
		forkIndex:         map[string]*ForkIndex{},
		forkData:          initForkData(),
		memUnit:           InitMemUnit(),
		memSize:           dagconfig.DefaultConfig.MemoryUnitSize,
	}
	return &memdag
}

func (chain *MemDag) validateMemory() bool {
	if chain.memUnit.Lenth() >= uint64(chain.memSize) {
		return false
	}
	return true
}

func (chain *MemDag) Save(unit *Unit) error {
	if unit == nil {
		return fmt.Errorf("Save mem unit: unit is null")
	}
	if chain.memUnit.Exists(unit.UnitHash) {
		return fmt.Errorf("Save mem unit: unit is already exists in memory")
	}
	if !chain.validateMemory() {
		return fmt.Errorf("Save mem unit: size is out of limit")
	}

	// save fork index
	assetId := unit.UnitHeader.Number.AssetID.String()
	forkIndex, ok := chain.forkIndex[assetId]
	if ok {
		forkIndex = &ForkIndex{}
		if err := forkIndex.AddData(0, unit.UnitHash); err != nil {
			return err
		}
		chain.forkId[assetId] = 0
	} else {
		id := chain.forkId[assetId]
		if err := forkIndex.AddData(int(id), unit.UnitHash); err != nil {
			return err
		}
	}
	// save memory unit
	if err := chain.memUnit.Add(unit); err != nil {
		return err
	}
	// save fork data

	// prune fork
	return nil
}

func (chain *MemDag) Exists(uHash common.Hash) bool {
	if chain.memUnit.Exists(uHash) {
		return true
	}
	return false
}
