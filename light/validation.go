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
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/trie"
	"github.com/palletone/go-palletone/dag"
	"github.com/palletone/go-palletone/dag/errors"
	"gopkg.in/karalabe/cookiejar.v2/collections/prque"
	"strconv"
	"sync"
	"time"
)

const (
	spvReqTimeout = 30 * time.Second // Maximum allotted time to return an explicitly requested block
	ERRSPVOTHERS  = 1
	ERRSPVTIMEOUT = 2
)

type proofReq struct {
	strindex string
	txhash   common.Hash // Hash of the block being announced
	time     time.Time   // Timestamp of the announcement
	step     chan int    //0:ok   1:err  2:timeout
	valid    *Validation
}

func NewProofReq(txhash common.Hash, valid *Validation) *proofReq {
	t := time.Now().UnixNano()
	str := strconv.FormatInt(t, 10)
	return &proofReq{strindex: str + txhash.String(), txhash: txhash, time: time.Now(), step: make(chan int), valid: valid}
}

func (req *proofReq) Wait() int {
	timeout := time.NewTicker(spvReqTimeout)
	defer timeout.Stop()
	for {
		select {
		case result := <-req.step:
			req.valid.forgetHash(req.strindex)
			return result
		case <-timeout.C:
			req.valid.forgetHash(req.strindex)
			return ERRSPVTIMEOUT
		}
	}
}

type Validation struct {
	preq     map[string]*proofReq //key:txhash  request queue
	preqLock sync.RWMutex

	queue *prque.Prque //recv validation path
	dag   dag.IDag

	//quit chan struct{}
}

func NewValidation(dag dag.IDag) *Validation {
	return &Validation{
		preq:  make(map[string]*proofReq),
		queue: prque.New(),
		//quit:  make(chan struct{}),
		dag: dag,
	}
}

func (v *Validation) Start() {
	//go v.loop()
}

// Stop terminates the announcement based synchroniser, canceling all pending
// operations.
func (v *Validation) Stop() {
	//close(v.quit)
}

func (v *Validation) forgetHash(index string) {
	v.preqLock.Lock()
	delete(v.preq, index)
	v.preqLock.Unlock()
}

func (v *Validation) Check(resp *proofsRespData) (int, error) {
	header, err := v.dag.GetHeaderByHash(resp.headerhash)
	if err != nil {
		log.Debug("Light PalletOne", "Validation->Check GetHeaderByHash err", err, "header hash", resp.headerhash)
		return 0, err
	}
	//TODO recover
	//if header.TxRoot.String() != resp.txroothash.String() {
	//	return 0, errors.New("txroothash not equal")
	//}
	log.Debug("Light PalletOne", "key", resp.key, "proof", resp.pathData)
	nodeSet := resp.pathData.NodeSet()
	_, err, _ = trie.VerifyProof(header.TxRoot, resp.key, nodeSet)
	if err != nil {
		log.Debug("Light PalletOne", "Validation->Check VerifyProof err", err)
		return 0, err
	}
	return 0, nil
}

func (v *Validation) AddSpvResp(resp *proofsRespData) error {
	v.preqLock.RLock()
	vreq, ok := v.preq[resp.index]
	if !ok {
		v.preqLock.RUnlock()

		//vreq.step <- ERRSPVOTHERS
		log.Debug("Light PalletOne", "Validation->Check key is not exist.key", resp.index)
		return errors.New("Key is not exist")
	}
	v.preqLock.RUnlock()
	_, err := v.Check(resp)
	if err != nil {
		vreq.step <- ERRSPVOTHERS
		return err
	}
	vreq.step <- 0
	return nil
}

func (v *Validation) AddSpvReq(strhash string) (*proofReq, error) {
	hash := common.Hash{}
	hash.SetHexString(strhash)
	log.Debug("Light PalletOne ProtocolManager ReqProof", "strhash", strhash, "common hash", hash.String())

	//TODO add limit console visit times
	//v.preqLock.RLock()
	//if _, ok := v.preq[hash]; ok {
	//	v.preqLock.RUnlock()
	//	return nil, errors.New("Key is exist")
	//}
	//v.preqLock.RUnlock()

	req := NewProofReq(hash, v)
	v.preqLock.Lock()
	v.preq[req.strindex] = req
	v.preqLock.Unlock()
	return req, nil
}
