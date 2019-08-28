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
	"sync"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/dag"
)

var RwM *RwSetTxMgr
var ChainId = "palletone"

type RwSetTxMgr struct {
	name      string
	baseTxSim map[string]TxSimulator
	closed    bool
	rwLock    *sync.RWMutex
	wg        sync.WaitGroup
}

func NewRwSetMgr(name string) (*RwSetTxMgr, error) {
	return &RwSetTxMgr{name: name, baseTxSim: make(map[string]TxSimulator), rwLock: new(sync.RWMutex)}, nil
}

// NewTxSimulator implements method in interface `txmgmt.TxMgr`
func (m *RwSetTxMgr) NewTxSimulator(idag dag.IDag, chainid string, txid string, is_sys bool) (TxSimulator, error) {
	log.Debugf("constructing new tx simulator")
	hash := common.HexToHash(txid)
	if !is_sys { // 用户合约
		m.rwLock.RLock()
		ts, ok := m.baseTxSim[chainid+txid]
		m.rwLock.RUnlock()
		if ok {
			if ts.(*RwSetTxSimulator).txid == hash {
				log.Infof("chainid[%s] , txid[%s]already exit, don't create user txsimulator again.", chainid, txid)
				return ts, nil
			}
		}
		// new txsimulator
		t0 := NewBasedTxSimulator(idag, hash)
		m.rwLock.Lock()
		m.baseTxSim[chainid+txid] = t0
		m.wg.Add(1)
		m.rwLock.Unlock()
		log.Infof("create user rwSetTx [%s]", hash.String())
		return t0, nil
	} else {
		m.rwLock.RLock()
		ts, ok := m.baseTxSim[chainid]
		m.rwLock.RUnlock()
		if ok {
			log.Infof("chainid[%s] , txid[%s]already exit, don't create sys txsimulator again.", chainid, txid)
			return ts, nil
		}
		t := NewBasedTxSimulator(idag, hash)
		if t == nil {
			return nil, errors.New("NewBaseTxSimulator is failed.")
		}
		m.rwLock.Lock()
		m.baseTxSim[chainid] = t
		m.wg.Add(1)
		m.rwLock.Unlock()
		log.Infof("creat sys rwSetTx [%s]", hash.String())

		return t, nil
	}
}

func (m *RwSetTxMgr) BaseTxSim() map[string]TxSimulator {
	return m.baseTxSim
}

// 每次产块结束后，需要关闭该chainId的txsimulator.
func (m *RwSetTxMgr) CloseTxSimulator(chainid, txid string) error {
	m.rwLock.Lock()
	defer m.rwLock.Unlock()
	if ts, ok := m.baseTxSim[chainid+txid]; ok {
		ts.Done()
		delete(m.baseTxSim, chainid+txid)
		m.wg.Done()
	}
	if ts, ok := m.baseTxSim[chainid]; ok {
		ts.Done()
		delete(m.baseTxSim, chainid)
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

func Init() {
	var err error
	RwM, err = NewRwSetMgr("default")
	if err != nil {
		log.Error("fail!")
	}
}

func init() {
	RwM, _ = NewRwSetMgr("default")
}
