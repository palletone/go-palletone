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
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/modules"
)

// non validated units set
type MemUnitInfo map[common.Hash]*modules.Unit

type MemUnit struct {
	memUnitInfo *MemUnitInfo
	memLock     sync.RWMutex

	numberToHash     map[modules.ChainIndex]common.Hash
	numberToHashLock sync.RWMutex
}

func InitMemUnit() *MemUnit {
	memUnitInfo := make(MemUnitInfo)
	numberToHash := map[modules.ChainIndex]common.Hash{}
	memUnit := MemUnit{
		memUnitInfo:  &memUnitInfo,
		numberToHash: numberToHash,
	}
	return &memUnit
}

//set the mapping relationship
//key:number  value:unit hash
func (mu *MemUnit) SetHashByNumber(chainIndex modules.ChainIndex, hash common.Hash) {
	mu.numberToHashLock.Lock()
	defer mu.numberToHashLock.Unlock()
	if _, ok := mu.numberToHash[chainIndex]; ok {
		return
	}
	mu.numberToHash[chainIndex] = hash
	return
}

//get the mapping relationship
//key:number  result:unit hash
func (mu *MemUnit) GetHashByNumber(chainIndex modules.ChainIndex) (common.Hash, error) {
	mu.numberToHashLock.RLock()
	defer mu.numberToHashLock.RUnlock()
	if hash, ok := mu.numberToHash[chainIndex]; ok {
		return hash, nil
	}
	return common.Hash{}, errors.New("have not key")
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
	log.Info("insert memUnit success.", "hashHex", u.UnitHash.String())
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
	_, ok := (*mu.memUnitInfo)[hash]
	mu.memLock.RUnlock()
	if ok {
		return true
	}
	return false
}
func (mu *MemUnit) Refresh(hash common.Hash) error {
	// 删除该hash在memUnit的记录。
	if hash == (common.Hash{}) {
		return errors.New("hash is null.")
	}
	mu.memLock.Lock()
	if _, has := (*mu.memUnitInfo)[hash]; has {
		delete((*mu.memUnitInfo), hash)
		for index, h := range mu.numberToHash {
			if h == hash {
				delete(mu.numberToHash, index)
				break
			}
		}
		return nil
	}
	mu.memLock.Unlock()
	return errors.New(fmt.Sprintf("the hash(%s) is not exist", hash.String()))
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
// type ForkIndex []*ForkData
type ForkIndex map[uint64]ForkData

// type MainIndex map[uint64]*MainData
type MainData struct {
	Index  *modules.ChainIndex
	Hash   *common.Hash
	Number uint64
}

var forkIndexLock sync.RWMutex

func (forkIndex *ForkIndex) AddData(unitHash common.Hash, parentsHash []common.Hash, index uint64) (int64, error) {
	forkIndexLock.Lock()
	defer forkIndexLock.Unlock()
	// if index <= 0 {
	// 	index = uint64(len(*forkIndex))
	// }
	in, err := forkIndex.addDate(unitHash, parentsHash, index)
	log.Info("遍历33333  fork Index:", "index", index)
	for key, data := range *forkIndex {
		fmt.Println("index: ", key)
		fmt.Println(" data: ", data)
	}
	return in, err
}
func (forkIndex *ForkIndex) addDate(hash common.Hash, parentsHash []common.Hash, index uint64) (int64, error) {
	data1 := make(ForkData, 0)
	data, has := (*forkIndex)[index]
	if has {
		log.Info("444444444444   fork Index:")
		if data.Exists(hash) {
			log.Info("444444444444 000000000  fork Index:")
			return int64(index), nil
		}
		// index++
		// forkIndex.addDate(hash, parentsHash, index)
		if err := data.Add(hash); err != nil {
			log.Info("444444444444   11111111111  fork Index:")
			return -1, err
		}
	} else {
		log.Info("5555555555555   fork Index:")
		// add hash into ForkData and return index.
		if err := data1.Add(hash); err != nil {
			log.Info("55555555555  111111111   fork Index:")
			return -1, err
		}
	}

	if len(data1) > 0 {
		(*forkIndex)[index] = data1
	} else {
		(*forkIndex)[index] = data
	}

	h := (*forkIndex)[index-1]
	// TODO   验证后续再加
	if h != nil && len(h) > 0 {
		for _, v := range h {
			if common.CheckExists(v, parentsHash) >= 0 {
				log.Debug("666666666666  memUnit add data success  =================", "index", index)
				return int64(index), nil
			}
		}
	} else {
		hh := (*forkIndex)[uint64(0)] // 重启后第一个稳定的unit hash
		for _, v := range hh {
			if common.CheckExists(v, parentsHash) >= 0 {
				log.Debug("777777777777  first  add data success  =================", "index", index)
				return int64(index), nil
			}
		}
	}

	return -2, fmt.Errorf(" =================== Unit(%x) is not continuously", hash)
}

// the  index of parameter is fork's index
func (forkIndex *ForkIndex) IsReachedIrreversibleHeight(index uint64, main_index uint64) bool {
	forkIndexLock.RLock()
	defer forkIndexLock.RUnlock()
	if index <= 15 {
		return false
	}

	if data, has := (*forkIndex)[index]; has { //dagconfig.DefaultConfig.IrreversibleHeight {
		if data == nil {
			return false
		}
		// TODO  超过15个mediator生产的单元，fork里的第一个单元才能被确认为已不可逆（已稳定）。
		// ...

		if s_index := index - uint64(dagconfig.DefaultConfig.IrreversibleHeight); s_index >= main_index {
			if data := (*forkIndex)[s_index+1]; data != nil {
				return true
			}
		}
	}
	return false
}

func (forkIndex *ForkIndex) GetStableUnitHash(index int64) common.Hash {

	if index < int64(dagconfig.DefaultConfig.IrreversibleHeight) {
		return (common.Hash{})
	}

	s_index := uint64(index - int64(dagconfig.DefaultConfig.IrreversibleHeight-1))
	forkIndexLock.RLock()
	hashs, has := (*forkIndex)[s_index]
	forkIndexLock.RUnlock()
	if !has {
		log.Info("forkIndex cache111 :::::::::::::::", "index", index, "s_index", s_index, "hashHex", hashs[0].String())
		return (common.Hash{})
	}
	if len(hashs) <= 0 {
		log.Info("forkIndex cache222 :::::::::::::::", "index", index, "s_index", s_index, "hashHex", hashs[0].String())
		return (common.Hash{})
	} else {
		log.Debug("forkIndex cache333333  :::::::::::::::", "index", index, "s_index", s_index, "hashHex", hashs[0].String())
	}
	hash := (hashs)[0]
	forkIndex.RemoveStableIndex(s_index)
	return hash
}
func (forkIndex *ForkIndex) RemoveStableIndex(index uint64) {
	if forkIndex == nil {
		return
	}
	// forkIndexLock.Lock()
	// defer forkIndexLock.Unlock()
	_, has := (*forkIndex)[index]
	if has {
		delete((*forkIndex), index)
	}
}
func (forkIndex *ForkIndex) GetReachedIrreversibleHeightUnitHash(index uint64) common.Hash {
	forkIndexLock.RLock()
	defer forkIndexLock.RUnlock()
	if index <= 0 {
		return common.Hash{}
	}
	return (*forkIndex)[index][0]
}

func (forkIndex *ForkIndex) Lenth() int {
	forkIndexLock.RLock()
	defer forkIndexLock.RUnlock()
	return len(*forkIndex)
}
