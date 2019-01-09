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
	"sort"
	"strings"
	"sync"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/errors"
	"github.com/palletone/go-palletone/dag/modules"
)

// non validated units set
type MemUnitInfo map[common.Hash]*modules.Unit

type MemUnit struct {
	//memUnitInfo *MemUnitInfo
	memUnitInfo *sync.Map
	memLock     sync.RWMutex

	numberToHash     map[modules.ChainIndex]common.Hash
	numberToHashLock sync.RWMutex
}

func InitMemUnit() *MemUnit {
	memUnitInfo := new(sync.Map)
	numberToHash := map[modules.ChainIndex]common.Hash{}
	memUnit := MemUnit{
		memUnitInfo:  memUnitInfo,
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

func (mu *MemUnit) DelHashByNumber(chainIndex modules.ChainIndex) error {
	mu.numberToHashLock.Lock()
	defer mu.numberToHashLock.Unlock()
	if _, ok := mu.numberToHash[chainIndex]; !ok {
		return errors.New("the hash is not exist")
	}
	delete(mu.numberToHash, chainIndex)
	return nil
}
func (mu *MemUnit) Add(u *modules.Unit) error {
	if mu == nil {
		mu = InitMemUnit()
	}
	// _, ok := mu.memUnitInfo.Load(u.Hash())
	// //_, ok := (*mu.memUnitInfo)[u.UnitHash]
	// if !ok {
	// 	mu.memUnitInfo.Store(u.Hash(), u)
	// 	// (*mu.memUnitInfo)[u.UnitHash] = u
	// }

	_, ok := mu.memUnitInfo.LoadOrStore(u.Hash(), u)
	if !ok {
		mu.memUnitInfo.Store(u.Hash(), u)
	}
	log.Info("insert memUnit success.", "hashHex", u.Hash().String())
	return nil
}

func (mu *MemUnit) Get(hash common.Hash) (*modules.Unit, error) {
	// mu.memLock.RLock()
	// defer mu.memLock.RUnlock()
	data, ok := mu.memUnitInfo.Load(hash)
	if !ok {
		return nil, fmt.Errorf("Get mem unit: unit does not be found.")
	}
	// unit, ok := (*mu.memUnitInfo)[hash]
	// if !ok || unit == nil {
	// 	return nil, fmt.Errorf("Get mem unit: unit does not be found.")
	// }
	unit := data.(*modules.Unit)
	return unit, nil
}

func (mu *MemUnit) Exists(hash common.Hash) bool {
	_, ok := mu.memUnitInfo.Load(hash)
	return ok
}
func (mu *MemUnit) Refresh(hash common.Hash) error {
	// 删除该hash在memUnit的记录。
	if hash == (common.Hash{}) {
		return errors.New("hash is null.")
	}
	_, ok := mu.memUnitInfo.Load(hash)
	if ok {
		mu.memUnitInfo.Delete(hash)
	} else {
		log.Debug(fmt.Sprintf("the hash(%s) is not exist", hash.String()))
	}

	mu.memLock.Lock()
	for index, h := range mu.numberToHash {
		if h == hash {
			delete(mu.numberToHash, index)
			break
		}
	}
	mu.memLock.Unlock()
	return nil
	//return errors.New(fmt.Sprintf("the hash(%s) is not exist", hash.String()))
}

func (mu *MemUnit) Lenth() uint64 {
	var count uint64
	mu.memUnitInfo.Range(func(key, val interface{}) bool {
		fmt.Println(key, val)
		count++
		return true
	})
	return count
}

/*********************************************************************/
// fork data
type ForkData []*Data
type Data struct {
	hash common.Hash
	addr string
}

func (f *ForkData) Add(hash common.Hash, addr string) error {
	if f.Exists(hash) {
		return fmt.Errorf("Save fork data: unit is already existing.")
	}
	*f = append(*f, &Data{hash: hash, addr: addr})
	return nil
}

func (f *ForkData) Exists(hash common.Hash) bool {
	for _, data := range *f {
		if strings.Compare(data.hash.String(), hash.String()) == 0 {
			return true
		}
	}
	return false
}

func (f *ForkData) GetLast() common.Hash {

	for i := len(*f) - 1; i >= 0; i-- {
		d := f.get_last(i)
		if d != nil {
			return d.hash
		}
	}
	return common.Hash{}
}
func (f *ForkData) get_last(index int) *Data {
	if index > len(*f) || index < 0 {
		return nil
	}
	return (*f)[index]
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

func (forkIndex *ForkIndex) AddData(unitHash common.Hash, parentsHash []common.Hash, index uint64, address string) (int64, error) {
	forkIndexLock.Lock()
	defer forkIndexLock.Unlock()

	in, err := forkIndex.addDate(unitHash, parentsHash, index, address)
	return in, err
}
func (forkIndex *ForkIndex) addDate(hash common.Hash, parentsHash []common.Hash, index uint64, addr string) (int64, error) {
	data1 := make(ForkData, 0)
	data, has := (*forkIndex)[index]
	if has {
		if data.Exists(hash) {
			return int64(index), nil
		}
		// index++
		// forkIndex.addDate(hash, parentsHash, index)
		if err := data.Add(hash, addr); err != nil {
			return -1, err
		}
	} else {
		// add hash into ForkData and return index.
		if err := data1.Add(hash, addr); err != nil {
			return -1, err
		}
	}

	if data1.Exists(hash) {
		(*forkIndex)[index] = data1
	} else {
		(*forkIndex)[index] = data
	}

	h := (*forkIndex)[index-1]
	// TODO   验证后续再加
	if h != nil && len(h) > 0 {
		for _, v := range h {
			if common.CheckExists(v.hash, parentsHash) >= 0 {
				log.Debug("checkExists  success  =================", "index", index)
				return int64(index), nil
			}
		}
	} else {
		hh := (*forkIndex)[uint64(0)] // 重启后第一个稳定的unit hash
		for _, v := range hh {
			if common.CheckExists(v.hash, parentsHash) >= 0 {
				log.Debug("checkExists first hash success  =================", "index", index)
				return int64(index), nil
			}
		}
	}

	return -2, fmt.Errorf(" =================== Unit(%x) is not continuously", hash)
}

// the  index of parameter is fork's index
func (forkIndex *ForkIndex) IsReachedIrreversibleHeight(index uint64, main_index uint64, threshold int) bool {
	forkIndexLock.RLock()
	defer forkIndexLock.RUnlock()
	if int(index) <= threshold {
		return false
	}

	if data, has := (*forkIndex)[index]; has { //dagconfig.DefaultConfig.IrreversibleHeight {
		if data == nil {
			return false
		}
		// TODO  超过2/3个mediator生产的单元，fork里的第一个单元才能被确认为已不可逆（已稳定）。
		// ...

		if s_index := index - uint64(dagconfig.DefaultConfig.IrreversibleHeight); s_index >= main_index {
			if data := (*forkIndex)[s_index+1]; data != nil {
				return true
			}
		}
	}
	return false
}

type UInt64Slice []uint64

func (c UInt64Slice) Len() int {
	return len(c)
}
func (c UInt64Slice) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}
func (c UInt64Slice) Less(i, j int) bool {
	return c[i] < c[j]
}
func (forkIndex *ForkIndex) GetStableUnitHash(index int64) []common.Hash {
	// 计数器， 确保最少规模的mediator生产的单元才能确认为稳定的
	var countMediators map[string]struct{}
	if index < int64(dagconfig.DefaultConfig.IrreversibleHeight) {
		return nil
	}
	countMediators = make(map[string]struct{})
	all_index := make(UInt64Slice, 0)
	var min_index uint64
	forkIndexLock.RLock()
	defer forkIndexLock.RUnlock()
	for index, hashs := range *forkIndex {
		for _, data := range hashs {
			if data != nil {
				countMediators[data.addr] = struct{}{}

			}
		}
		all_index = append(all_index, index)
	}
	// 判断够不够最小规模mediator数，不够则返回，否则返回高度最小且最老的hash值。
	if len(countMediators) <= dagconfig.DefaultConfig.IrreversibleHeight {
		return nil
	}

	if len(all_index) > 0 {
		sort.Sort(all_index)
		min_index = all_index[0]
	}

	s_index := uint64(index - int64(dagconfig.DefaultConfig.IrreversibleHeight-1))
	if min_index > 0 {
		s_index = min_index
	}

	hashs, has := (*forkIndex)[s_index]

	if !has {
		return nil
	}
	if len(hashs) <= 0 {
		return nil
	}
	//hash := (hashs)[0]
	delHashs := make([]common.Hash, 0)
	if len(hashs) > 0 {
		for i := 0; i < len(hashs); i++ {
			delHashs = append(delHashs, hashs[i].hash)
		}
	}
	// forkIndex.RemoveStableIndex(s_index)
	return delHashs
}
func (forkIndex *ForkIndex) RemoveStableIndex(index uint64) {
	if forkIndex == nil {
		return
	}
	forkIndexLock.Lock()
	defer forkIndexLock.Unlock()
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
	return (*forkIndex)[index][0].hash
}

func (forkIndex *ForkIndex) Lenth() int {
	forkIndexLock.RLock()
	defer forkIndexLock.RUnlock()
	return len(*forkIndex)
}
