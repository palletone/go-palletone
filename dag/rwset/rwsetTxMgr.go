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

package rwset

import (
	"errors"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/dag"
	"sync"
)

type RwSetTxMgr struct {
	name      string
	baseTxSim map[string]TxSimulator
	closed    bool
	rwLock    sync.RWMutex
	wg        sync.WaitGroup
}

func NewRwSetMgr(name string) (*RwSetTxMgr, error) {
	return &RwSetTxMgr{name: name, baseTxSim: make(map[string]TxSimulator)}, nil
}

// NewTxSimulator implements method in interface `txmgmt.TxMgr`
func (m *RwSetTxMgr) NewTxSimulator(idag dag.IDag, chainid string, txid string, is_sys bool) (TxSimulator, error) {
	log.Debugf("constructing new tx simulator")
	hash := common.HexToHash(txid)
	if is_sys {
		m.rwLock.RLock()
		ts, ok := m.baseTxSim[chainid]
		m.rwLock.RUnlock()
		if ok {
			if ts.(*RwSetTxSimulator).txid == hash {
				log.Infof("chainid[%s] , txid[%s]already exit, return.", chainid, txid)
				return ts, nil
			}
		}
	}
	t := NewBasedTxSimulator(idag, hash)
	if t == nil {
		return nil, errors.New("NewBaseTxSimulator is failed.")
	}
	m.rwLock.Lock()
	m.baseTxSim[chainid] = t
	m.wg.Add(1)
	m.rwLock.Unlock()
	log.Infof("creat new rwSetTx")

	return t, nil
}
func (m *RwSetTxMgr) CloseTxSimulator(chainid string) error {
	m.rwLock.Lock()
	defer m.rwLock.Unlock()
	if ts, ok := m.baseTxSim[chainid]; ok {
		if ts.CheckDone() != nil {
			return errors.New("this txsimulator isnot done.")
		}
		delete(m.baseTxSim, chainid)
		m.wg.Done()
	}
	return nil
}
func (m *RwSetTxMgr) Close() {
	m.rwLock.Lock()
	if m.closed {
		return
	}
	for _, ts := range m.baseTxSim {
		if ts.CheckDone() == nil {
			continue
		}
		// todo
		//ts.Done()
	}
	m.wg.Wait()
	m.closed = true
	m.rwLock.Unlock()
	return
}
