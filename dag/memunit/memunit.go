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
	"strings"
	"sync"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/modules"
)

// non validated units set
type MemUnitInfo map[common.Hash]*modules.Unit

type MemUnit struct {
	memUnitInfo *MemUnitInfo
	memLock     sync.RWMutex
}

func InitMemUnit() *MemUnit {
	memUnitInfo := make(MemUnitInfo)
	memUnit := MemUnit{memUnitInfo: &memUnitInfo}
	return &memUnit
}

//set the mapping relationship
//key:number  value:unit hash
func (mu *MemUnit) SetHashByNumber() error {
	return nil
}

//get the mapping relationship
//key:number  result:unit hash
func (mu *MemUnit) GetHashByNumber() (common.Hash, error) {
	return common.Hash{}, nil
}

func (mu *MemUnit) Add(u *modules.Unit) error {
	mu.memLock.Lock()
	defer mu.memLock.Unlock()
	if mu == nil {
		mu = InitMemUnit()
	}
	_, ok := (*mu.memUnitInfo)[u.UnitHash]
	if !ok {
		(*mu.memUnitInfo)[u.UnitHash] = u
	}
	return nil
}

func (mu *MemUnit) Get(hash common.Hash) (*modules.Unit, error) {
	mu.memLock.RLock()
	defer mu.memLock.RUnlock()
	unit, ok := (*mu.memUnitInfo)[hash]
	if !ok || unit == nil {
		return nil, fmt.Errorf("Get mem unit: unit does not be found.")
	}
	return unit, nil
}

func (mu *MemUnit) Exists(hash common.Hash) bool {
	mu.memLock.RLock()
	defer mu.memLock.RUnlock()
	_, ok := (*mu.memUnitInfo)[hash]
	if ok {
		return true
	}
	return false
}

func (mu *MemUnit) Lenth() uint64 {
	mu.memLock.RLock()
	defer mu.memLock.RUnlock()
	return uint64(len(*mu.memUnitInfo))
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
