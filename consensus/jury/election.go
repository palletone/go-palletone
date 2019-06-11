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
	"crypto/ecdsa"
	"fmt"
	"time"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/util"
	alg "github.com/palletone/go-palletone/consensus/jury/algorithm"
	"github.com/palletone/go-palletone/consensus/jury/vrfEc"
	"github.com/palletone/go-palletone/consensus/jury/vrfEs"
	"github.com/palletone/go-palletone/core/accounts"
	"github.com/palletone/go-palletone/core/accounts/keystore"
	"github.com/palletone/go-palletone/dag/errors"
	"github.com/palletone/go-palletone/dag/modules"
)

type vrfAccount struct {
	pubKey *ecdsa.PublicKey  //vrfEc
	priKey *ecdsa.PrivateKey //vrfEc
}

type elector struct {
	num    uint
	weight uint64
	total  uint64
	vrfAct vrfAccount //vrf ec

	addr     common.Address
	password string
	ks       *keystore.KeyStore
}

func (e *elector) checkElectedEc(data []byte) (proof []byte, err error) {
	if e.num < 0 || e.weight < 1 || data == nil {
		errs := fmt.Sprintf("checkElected param error, num[%d], weight[%d]", e.num, e.weight)
		return nil, errors.New(errs)
	}
	proof, err = vrfEc.VrfProve(e.vrfAct.priKey, data)
	if err != nil {
		return nil, err
	}
	vrfValue := vrfEc.VrfProof2Value(e.vrfAct.priKey.Params(), proof)
	if len(vrfValue) > 0 {
		if alg.Selected(e.num, e.weight, uint64(e.total), vrfValue) > 0 {
			return proof, nil
		}
	}
	return nil, nil
}

func (e *elector) checkElected(data []byte) (proof []byte, err error) {
	if e.num < 0 || e.weight < 1 || data == nil {
		errs := fmt.Sprintf("checkElected param error, num[%d], weight[%d]", e.num, e.weight)
		return nil, errors.New(errs)
	}
	a := accounts.Account{
		Address: e.addr,
	}
	privateKey, err := e.ks.DumpPrivateKey(a, e.password)
	if err != nil {
		return nil, err
	}
	proof, err = vrfEs.VrfProve(privateKey, data)
	if err != nil {
		return nil, err
	}
	vrfValue := proof
	if len(vrfValue) > 0 {
		if alg.Selected(e.num, e.weight, uint64(e.total), vrfValue) > 0 {
			return proof, nil
		}
	}
	return nil, nil
}

func (e *elector) verifyVrfEc(proof, data []byte) (bool, error) {
	ok, err := vrfEc.VrfVerify(e.vrfAct.pubKey, data, proof)
	if err != nil {
		return false, err
	}
	if ok {
		vrfValue := vrfEc.VrfProof2Value(e.vrfAct.pubKey.Params(), proof)
		if len(vrfValue) > 0 {
			if alg.Selected(e.num, e.weight, uint64(e.total), vrfValue) > 0 {
				return true, nil
			}
		}
	}
	return false, nil
}

func (e *elector) verifyVrf(proof, data []byte, pubKey []byte) (bool, error) {
	ok, err := vrfEs.VrfVerify(pubKey, data, proof)
	if err != nil {
		log.Error("verifyVrf fail", "ok?", ok)
		return false, err
	}
	if ok {
		vrfValue := proof
		if len(vrfValue) > 0 {
			if alg.Selected(e.num, e.weight, uint64(e.total), vrfValue) > 0 {
				return true, nil
			}
		}
	}
	return false, nil
}

func (p *Processor) electionEventIsProcess(event *ElectionEvent, addr *common.Address) (common.Hash, bool) {
	if event == nil {
		return common.Hash{}, false
	}
	reqId := common.Hash{}
	switch event.EType {
	case ELECTION_EVENT_REQUEST:
		reqId = event.Event.(*ElectionRequestEvent).ReqId
	case ELECTION_EVENT_RESULT:
		reqId = event.Event.(*ElectionResultEvent).ReqId
	}
	//request node
	if _, ok := p.mel[reqId]; ok && p.mel[reqId].eChan != nil {
		return reqId, true
	}
	//jury node
	if p.isLocalActiveJury(*addr) {//localHaveActiveJury()
		return reqId, true
	}
	return reqId, false
}
func (p *Processor) electionEventBroadcast(event *ElectionEvent) (recved bool, err error) {
	if event == nil {
		return false, errors.New("electionEventBroadcast event is nil")
	}
	p.locker.Lock()
	defer p.locker.Unlock()

	switch event.EType {
	case ELECTION_EVENT_REQUEST:
		evt := event.Event.(*ElectionRequestEvent)
		reqId := evt.ReqId
		if e, ok := p.mel[reqId]; ok {
			if e.req {
				return true, nil
			} else {
				e.req = true
			}
		} else {
			p.mel[reqId] = &electionVrf{}
			p.mel[reqId].req = true
			p.mel[reqId].tm = time.Now()
		}
	case ELECTION_EVENT_RESULT:
		evt := event.Event.(*ElectionResultEvent)
		reqId := evt.ReqId
		if e, ok := p.mel[reqId]; ok {
			for _, a := range e.rst {
				if a == evt.Ele.AddrHash {
					return true, nil
				}
			}
			e.rst = append(e.rst, evt.Ele.AddrHash)
		} else {
			p.mel[reqId] = &electionVrf{rst: make([]common.Hash, 0)}
			p.mel[reqId].rst = append(p.mel[reqId].rst, evt.Ele.AddrHash)
			p.mel[reqId].tm = time.Now()
		}
	}
	go p.ptn.ElectionBroadcast(*event)
	return false, nil
}

func (p *Processor) processElectionRequestEvent(ele *elector, reqEvt *ElectionRequestEvent) (result *ElectionEvent, err error) {
	//产生vrf证明
	//计算二项式分步，确定自己是否选中
	//如果选中，则对请求结果返回
	reqId := reqEvt.ReqId
	if len(p.local) < 1 {
		return nil, fmt.Errorf("processElectionRequestEvent, local jury addr is nil, reqId[%s]", reqId.String())
	}
	addr := common.Address{}
	addrHash := common.Hash{}
	for addr, _ = range p.local {
		addrHash = util.RlpHash(addr)
		break //only first one
	}
	proof, err := ele.checkElected(getElectionSeedData(reqEvt.ReqId))
	if err != nil {
		log.Errorf("[%s]processElectionRequestEvent, checkElected err, %s", shortId(reqId.String()), err.Error())
		return nil, fmt.Errorf("processElectionRequestEvent, checkElected err, reqId[%s]", shortId(reqId.String()))
	}
	pubKey, err := p.ptn.GetKeyStore().GetPublicKey(addr)
	if err != nil {
		log.Errorf("[%s]processElectionRequestEvent, get pubKey err, address[%s]", shortId(reqId.String()), addr.String())
		return nil, fmt.Errorf("processElectionRequestEvent, get pubKey err,reqId[%s]", shortId(reqId.String()))
	}

	if proof != nil {
		rstEvt := &ElectionResultEvent{
			ReqId: reqEvt.ReqId,
			Ele:   modules.ElectionInf{AddrHash: addrHash, Proof: proof, PublicKey: pubKey},
		}
		log.Debugf("[%s]processElectionRequestEvent, ok", shortId(reqId.String()))
		evt := &ElectionEvent{EType: ELECTION_EVENT_RESULT, Event: rstEvt}
		return evt, nil
	}
	return nil, nil
}

func (p *Processor) processElectionResultEvent(ele *elector, rstEvt *ElectionResultEvent) error {
	//验证vrf证明
	//收集vrf地址并添加缓存
	//检查缓存地址数量
	reqId := rstEvt.ReqId
	if _, ok := p.mtx[reqId]; !ok {
		log.Debugf("[%s]processElectionResultEvent, This node does not need to process the election message", shortId(reqId.String()))
		return nil
	}
	mel := p.mel[reqId]
	if mel.eChan == nil {
		//not request node
		return nil
	}
	if len(mel.eInf) >= p.electionNum {
		log.Infof("[%s]processElectionResultEvent, The quantity has reached the requirement", shortId(reqId.String()))
		return nil
	}
	log.Infof("[%s]processElectionResultEvent, ele addrHash[%s]", shortId(reqId.String()), rstEvt.Ele.AddrHash.String())

	tmpReqId := common.BytesToAddress(reqId.Bytes())
	contractId := common.NewAddress(tmpReqId.Bytes(), common.ContractHash)
	log.Debugf("[%s]processElectionResultEvent, contractId[%s] contractIdBytes[%x]", shortId(reqId.String()), contractId.String(), contractId.Bytes())

	ok, err := ele.verifyVrf(rstEvt.Ele.Proof, getElectionSeedData(reqId), rstEvt.Ele.PublicKey) //rstEvt.ReqId[:]
	if err != nil {
		log.Errorf("[%s]processElectionResultEvent, verify VRF fail", shortId(reqId.String()))
		return fmt.Errorf("processElectionResultEvent, verify VRF fail, reqId[%s]", shortId(reqId.String()))
	}
	if ok {
		p.locker.Lock()
		mel.eInf = append(mel.eInf, rstEvt.Ele)
		p.lockVrf[contractId] = append(p.lockVrf[contractId], rstEvt.Ele) //add lock vrf election info
		p.locker.Unlock()
		if len(mel.eInf) >= p.electionNum {
			//通知接收数量达到要求
			log.Infof("[%s]processElectionResultEvent,VRF address number is enough, Ok, contractId[%s]", shortId(reqId.String()), contractId.String())
			mel.eChan <- true
		}
	}
	return nil
}

func (p *Processor) ElectionRequest(reqId common.Hash, timeOut time.Duration) error {
	if reqId == (common.Hash{}) {
		return errors.New("ElectionRequest param is nil")
	}
	seedData := getElectionSeedData(reqId)
	p.locker.Lock()
	p.mel[reqId] = &electionVrf{
		eChan: make(chan bool, 1),
		eInf:  make([]modules.ElectionInf, 0),
		req:   true,
		tm:    time.Now(),
	}
	p.locker.Unlock()
	reqEvent := &ElectionRequestEvent{
		ReqId: reqId,
		//Data:  ele.seedData,
	}
	log.Debugf("[%s]ElectionRequest, reqId[%s], seedData[%v]", shortId(reqId.String()), reqId.String(), seedData)
	go p.ptn.ElectionBroadcast(ElectionEvent{EType: ELECTION_EVENT_REQUEST, Event: reqEvent})

	//超时等待选举结果
	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(timeOut)
		timeout <- true
	}()
	select {
	case <-p.mel[reqId].eChan:
		log.Debugf("[%s]ElectionRequest, election Ok", shortId(reqId.String()))
		return nil
	case <-timeout:
		log.Debugf("[%s]ElectionRequest, election time out", shortId(reqId.String()))
		return fmt.Errorf("ElectionRequest, election time out, reqId[%s]", reqId.String())
	}
	log.Debugf("[%s]ElectionRequest, election fail", shortId(reqId.String()))
	return fmt.Errorf("ElectionRequest, election fail, reqId[%s]", reqId.String())
}

func (p *Processor) ProcessElectionEvent(event *ElectionEvent) (result *ElectionEvent, err error) {
	if event == nil {
		return nil, errors.New("ProcessElectionRequestEvent, event is nil")
	}
	var account JuryAccount
	for _, a := range p.local {
		account.Address = a.Address
		account.Password = a.Password
		break //first one
	}
	reqId, isP := p.electionEventIsProcess(event, &account.Address)
	if !isP {
		log.Infof("[%s]ProcessElectionEvent, electionEventIsProcess is false, addr[%s], event type[%v]", shortId(reqId.String()), account.Address.String(), event.EType)
		return nil, nil
	}
	ele := &elector{
		num:      uint(p.electionNum),
		total:    uint64(p.dag.JuryCount()), // 100 todo dynamic acquisition
		addr:     account.Address,
		password: account.Password,
		ks:       p.ptn.GetKeyStore(),
	}
	ele.weight = electionWeightValue(ele.total)
	log.Debugf("[%s]ProcessElectionEvent, ele info[%v]", shortId(reqId.String()), ele)
	recved, err := p.electionEventBroadcast(event)
	if err != nil {
		return nil, err
	} else if recved {
		return nil, nil
	}
	log.Infof("[%s]ElectionMsg, event type[%v] event[%v]", shortId(reqId.String()), event.EType, event)

	if event.EType == ELECTION_EVENT_REQUEST {
		return p.processElectionRequestEvent(ele, event.Event.(*ElectionRequestEvent))
	} else if event.EType == ELECTION_EVENT_RESULT {
		return nil, p.processElectionResultEvent(ele, event.Event.(*ElectionResultEvent))
	}
	return nil, fmt.Errorf("ProcessElectionEvent fail, reqId[%s]", reqId.String())
}
