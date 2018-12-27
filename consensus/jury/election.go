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
package jury

import (
	"fmt"
	"sync"
	"sync/atomic"
	"errors"
	"github.com/palletone/go-palletone/common"
)

func GetJurors() ([]Processor, error) {
	//todo tmp
	var jurors []Processor
	var juror Processor
	//juror.ptype = TJury

	for i := 0; i < 10; i++ {
		fmt.Sprintf(juror.name, "juror_%d", i)
		jurors = append(jurors, juror)
	}

	return jurors, nil
}

//////////////////////////
type ProList struct {
	locker *sync.Mutex
	adds   map[common.Hash][]common.Address //common.Hash is contract transaction hash
}

var prolist ProList
var inited int32

func InitProList() error {
	atomic.LoadInt32(&inited)
	if inited > 0 {
		return errors.New("InitProList already init")
	}

	prolist.locker = new(sync.Mutex)
	prolist.adds = make(map[common.Hash][]common.Address, 0)
	atomic.StoreInt32(&inited, 1)
	return nil
}

func UpdateProList(hash common.Hash, address []common.Address) error {
	if len(address) < 1 {
		return errors.New("UpdateProList param is nil")
	}

	prolist.locker.Lock()
	defer prolist.locker.Unlock()

	if _, ok := prolist.adds[hash]; ok {
		//todo, 检查当前list中是否已存在
		//todo 更新部分地址

		prolist.adds[hash] = address
	} else {
		prolist.adds[hash] = address
	}

	return nil
}

func DeleteProList(hash common.Hash) error {
	prolist.locker.Lock()
	defer prolist.locker.Unlock()

	delete(prolist.adds, hash)
	return nil
}

func GetProList(hash common.Hash) ([]common.Address, error) {
	prolist.locker.Lock()
	defer prolist.locker.Unlock()

	if _, ok := prolist.adds[hash]; ok {
		return prolist.adds[hash], nil
	}

	return nil, errors.New(fmt.Sprintf("not find [%s] addr", hash.String()))
}
