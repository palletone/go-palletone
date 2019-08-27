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
 * @author PalletOne core developer Jiyou Wang <dev@pallet.one>
 * @date 2018
 */
package light

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/dag"
	"github.com/palletone/go-palletone/dag/modules"
)

const (
	OKUTXOsSync   = 0
	ERRUTXOsExist = 3
)

type utxosRespData struct {
	addr  string
	utxos map[modules.OutPoint]*modules.Utxo
}

func NewUtxosRespData() *utxosRespData {
	return &utxosRespData{utxos: make(map[modules.OutPoint]*modules.Utxo)}
}

func (u *utxosRespData) encode() ([][][]byte, error) {
	addrarr := [][]byte{}
	arrs := [][][]byte{}
	addrarr = append(addrarr, []byte(u.addr))
	arrs = append(arrs, addrarr)

	for outpoint, utxo := range u.utxos {
		var data [][]byte
		d1, err := json.Marshal(outpoint)
		if err != nil {
			return arrs, err
		}
		//log.Debug("Light PalletOne","utxosRespData encode outpoint",string(d1))
		data = append(data, d1)

		d2, err := json.Marshal(utxo)
		if err != nil {
			return arrs, err
		}
		//log.Debug("Light PalletOne","utxosRespData encode utxo",string(d2))
		data = append(data, d2)
		arrs = append(arrs, data)
	}
	return arrs, nil
}

func (u *utxosRespData) decode(arrs [][][]byte) error {
	u.addr = string(arrs[0][0])

	for _, arr := range arrs[1:] {
		var outpoint modules.OutPoint
		var utxo *modules.Utxo

		if err := json.Unmarshal(arr[0], &outpoint); err != nil {
			return err
		}
		if err := json.Unmarshal(arr[1], &utxo); err != nil {
			return err
		}
		u.utxos[outpoint] = utxo
	}
	return nil
}

type utxosReq struct {
	addr     string
	time     time.Time // Timestamp of the announcement
	step     chan int  //0:ok   1:err  2:timeout
	utxosync *UtxosSync
}

type UtxosSync struct {
	reqs map[string]*utxosReq //key:addr
	lock sync.RWMutex
	dag  dag.IDag
}

func NewUTXOsReq(addr string, utxosync *UtxosSync) *utxosReq {
	return &utxosReq{addr: addr, time: time.Now(), step: make(chan int), utxosync: utxosync}
}

func (req *utxosReq) Wait() int {
	timeout := time.NewTicker(spvReqTimeout)
	defer timeout.Stop()
	for {
		select {
		case result := <-req.step:
			//req.valid.forgetHash(req.strindex)
			req.utxosync.forgetHash(req.addr)
			return result
		case <-timeout.C:
			req.utxosync.forgetHash(req.addr)
			return ERRSPVTIMEOUT
		}
	}
}

func NewUtxosSync(dag dag.IDag) *UtxosSync {
	return &UtxosSync{
		dag:  dag,
		reqs: make(map[string]*utxosReq),
	}
}

func (u *UtxosSync) AddUtxoSyncReq(addr string) (*utxosReq, error) {
	u.lock.RLock()
	if req, ok := u.reqs[addr]; ok {
		u.lock.RUnlock()

		req.step <- ERRUTXOsExist
		log.Debug("Light PalletOne", "StartSyncByAddr key is exist. addr:", addr)
		return nil, errors.New("Key is not exist")
	}
	u.lock.RUnlock()

	req := NewUTXOsReq(addr, u)
	u.lock.Lock()
	u.reqs[addr] = req
	u.lock.Unlock()
	return req, nil
}

func (u *UtxosSync) forgetHash(addr string) {
	u.lock.Lock()
	delete(u.reqs, addr)
	u.lock.Unlock()
}

func (u *UtxosSync) SaveUtxoView(respdata *utxosRespData) error {
	u.lock.RLock()
	req, ok := u.reqs[respdata.addr]
	if !ok {
		u.lock.RUnlock()
		log.Debug("Light PalletOne", "SaveUtxoView key is not exist. addr:", respdata.addr)
		return fmt.Errorf("addr(%v) is not exist", respdata.addr)
	}
	u.lock.RUnlock()

	address, err := common.StringToAddress(respdata.addr)
	if err != nil {
		log.Debug("Light PalletOne", "SaveUtxoView err:", err, "addr", respdata.addr)
		return err
	}

	if err := u.dag.ClearUtxo(address); err != nil {
		log.Debug("Light PalletOne", "SaveUtxoView ClearUtxo err:", err, "addr", respdata.addr)
		return err
	}
	if err := u.dag.SaveUtxoView(respdata.utxos); err != nil {
		log.Debug("Light PalletOne", "SaveUtxoView failed,error:", err, "addr:", respdata.addr)
		return err
	}
	req.step <- OKUTXOsSync
	return nil
}
