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
	alg "github.com/palletone/go-palletone/consensus/jury/algorithm"
	"github.com/palletone/go-palletone/dag/errors"
	"github.com/palletone/go-palletone/common"
	"crypto/rand"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/util"
	"github.com/palletone/go-palletone/consensus/jury/vrf"
	"time"
)

type elector struct {
	num    int
	weight uint64
	total  uint64

	privkey alg.PrivateKey
	pubkey  alg.PublicKey
}

func (p *Processor) ElectionRequest(reqId common.Hash) error {
	return nil //todo

	if reqId == (common.Hash{}) {
		return errors.New("ElectionRequest param is nil")
	}
	//seedData= reqId + rand
	rd := make([]byte, 20)
	_, err := rand.Read(rd)
	if err != nil {
		return errors.New("ElectionRequest rand fail")
	}
	seedData := make([]byte, len(reqId)+len(rd))
	copy(seedData, reqId[:])
	copy(seedData[len(reqId):], rd)

	eleList := electionList{
		eleChan:  make(chan bool, 1),
		eleNum:   4, //todo
		seedData: seedData,
	}
	p.mtx[reqId].eleList = eleList
	reqEvent := ElectionRequestEvent{
		reqHash: reqId,
		num:     eleList.eleNum,
		data:    eleList.seedData,
	}
	log.Debug("ElectionRequest", "reqId", reqId.String(), "seedData", seedData)
	go p.ptn.ElectionBroadcast(ElectionEvent{EType: ELECTION_EVENT_REQUEST, Event: reqEvent})

	//超时等待选举结果
	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(time.Second * 5)
		timeout <- true
	}()
	select {
	case <-eleList.eleChan:
		log.Debug("ElectionRequest, election Ok")
		return nil
	case <-timeout:
		log.Debug("ElectionRequest, election time out")
	}
	return errors.New("ElectionRequest, election time out")
}

func (e *elector) checkElected(data []byte) (proof []byte, err error) {
	if e.num < 0 || e.weight < 10 || data == nil {
		return nil, errors.New("CheckElected param error")
	}
	proofHash, proof, err := e.privkey.Evaluate(data)
	if err != nil {
		return nil, err
	}
	selected := alg.Selected(e.num, e.weight, uint64(e.total), proofHash)
	if selected > 0 {
		return proof, nil
	} else {
		return nil, nil
	}
}

func (e *elector) verifyVRF(proof, data []byte) (bool, error) {
	ok, err := e.pubkey.VerifyVRF(proof, data)
	if err != nil {
		return false, err
	}
	if ok {
		proofHash := vrf.ECVRF_proof2hash(proof)
		selected := alg.Selected(e.num, e.weight, e.total, proofHash)
		if selected > 0 {
			return true, nil
		} else {
			return false, nil
		}
	}
	return false, nil
}

func (p *Processor) ProcessElectionRequestEvent(event *ElectionEvent) (result *ElectionEvent, err error) {
	//产生vrf证明
	//计算二项式分步，确定自己是否选中
	//如果选中，则对请求结果返回
	if event == nil {
		return nil, errors.New("ProcessElectionRequestEvent, event is nil")
	}
	if len(p.local) < 1 {
		return nil, errors.New("ProcessElectionRequestEvent, local jury addr is nil")
	}
	addrHash := common.Hash{}
	for addr, _ := range p.local {
		addrHash = util.RlpHash(addr)
		break //only first one
	}
	reqEvt := event.Event.(ElectionRequestEvent)
	log.Info("ProcessElectionRequestEvent", "reqHash", reqEvt.reqHash.String(), "num", reqEvt.num)
	ele := elector{
		num:    reqEvt.num,
		weight: 10,
		total:  1000,
		//privkey:, //todo
		//pubkey:,  //todo
	}
	proof, err := ele.checkElected(reqEvt.data)
	if err != nil {
		log.Error("ProcessElectionRequestEvent", "reqHash", reqEvt.reqHash, "checkElected err", err)
		return nil, err
	}
	if proof != nil {
		rstEvt := ElectionResultEvent{
			reqHash:   reqEvt.reqHash,
			addrHash:  addrHash,
			proof:     proof,
			publicKey: ele.pubkey,
		}
		log.Debug("ProcessElectionRequestEvent", "reqId", reqEvt.reqHash.String())
		evt := &ElectionEvent{EType: ELECTION_EVENT_RESULT, Event: rstEvt}
		return evt, nil
	}
	return nil, nil
}

func (p *Processor) ProcessElectionResultEvent(event *ElectionEvent) error {
	//验证vrf证明
	//收集vrf地址并添加缓存
	//检查缓存地址数量
	if event == nil {
		return errors.New("ProcessElectionResultEvent, event is nil")
	}
	rstEvt := event.Event.(ElectionResultEvent)
	log.Info("ProcessElectionResultEvent", "reqHash", rstEvt.reqHash.String(), "addrHash", rstEvt.addrHash.String())
	ele := elector{
		weight: 10,
		total:  1000,
		//privkey:, //todo
		//pubkey:,  //todo
	}
	if _, ok := p.mtx[rstEvt.reqHash]; !ok {
		return errors.New("ProcessElectionResultEvent, reqHash not find")
	}
	eList := p.mtx[rstEvt.reqHash].eleList
	if len(eList.addrHash) > eList.eleNum {
		log.Info("ProcessElectionResultEvent addrHash num > 3")
		return nil
	}
	ok, err := ele.verifyVRF(rstEvt.proof, eList.seedData)
	if err != nil {
		return err
	}
	if ok {
		eList.addrHash = append(eList.addrHash, rstEvt.addrHash)
		if len(eList.addrHash) > eList.eleNum {
			//通知接收数量达到要求
			eList.eleChan <- true
		}
	}

	return nil
}
