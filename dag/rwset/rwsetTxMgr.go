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
 * @date 2018-2020
 */

package rwset

import (
	"errors"

	"sync"

	"github.com/palletone/go-palletone/common/log"
)

var RwM TxManager
var ChainId = "palletone"

type RwSetTxMgr struct {
	name      string
	baseTxSim map[string]TxSimulator //key: unitHash or Height
	closed    bool
	rwLock    *sync.RWMutex
	wg        sync.WaitGroup
}

func NewRwSetMgr(name string) (*RwSetTxMgr, error) {
	return &RwSetTxMgr{name: name, baseTxSim: make(map[string]TxSimulator), rwLock: new(sync.RWMutex)}, nil
}

// NewTxSimulator implements method in interface `txmgmt.TxMgr`
func (m *RwSetTxMgr) NewTxSimulator(idag IDataQuery, unitId string) (TxSimulator, error) {

		m.rwLock.RLock()
		ts, ok := m.baseTxSim[unitId]
		m.rwLock.RUnlock()
		if ok {
			log.Infof("unitId[%s] already exit, don't create txsimulator again.", unitId)
			return ts, nil
		}
		t := NewBasedTxSimulator(idag)
		if t == nil {
			return nil, errors.New("NewBaseTxSimulator is failed.")
		}
		m.rwLock.Lock()
		m.baseTxSim[unitId] = t
		m.wg.Add(1)
		m.rwLock.Unlock()
		log.Debugf("creat sys rwSetTx [%s]", unitId)

		return t, nil

}

//func (m *RwSetTxMgr) BaseTxSim() map[string]TxSimulator {
//	return m.baseTxSim
//}

// 每次产块结束后，需要关闭该chainId的txsimulator.
func (m *RwSetTxMgr) CloseTxSimulator(unitId string) error {
	m.rwLock.Lock()
	defer m.rwLock.Unlock()
	if ts, ok := m.baseTxSim[unitId]; ok {
		ts.Done()
		delete(m.baseTxSim, unitId)
		m.wg.Done()
	}
	return nil
}
func (m *RwSetTxMgr) Close() {
	m.rwLock.Lock()
	defer m.rwLock.Unlock()
	if m.closed {
		return
	}
	for _, ts := range m.baseTxSim {
		if ts.CheckDone() != nil {
			continue
		}
	}
	m.baseTxSim = make(map[string]TxSimulator)
	m.closed = true
}

//func Init() {
//	var err error
//	RwM, err = NewRwSetMgr("default")
//	if err != nil {
//		log.Error("fail!")
//	}
//}

func init() {
	RwM, _ = NewRwSetMgr("default")
}
