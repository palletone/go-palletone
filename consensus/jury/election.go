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

	"github.com/palletone/go-palletone/dag/errors"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/util"
	"github.com/palletone/go-palletone/consensus/jury/vrfEc"
	"github.com/palletone/go-palletone/consensus/jury/vrfEs"
	"github.com/palletone/go-palletone/core/accounts/keystore"
	"github.com/palletone/go-palletone/core/accounts"
	alg "github.com/palletone/go-palletone/consensus/jury/algorithm"
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

func (p *Processor) electionEventIsProcess(event *ElectionEvent, addr *common.Address) bool {
	if event == nil {
		return false
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
		return true
	}
	//jury node
	if p.isLocalActiveJury(*addr) {
		return true
	}
	return false
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
		if e, ok := p.mel[evt.ReqId]; ok {
			if e.req {
				return true, nil
			} else {
				e.req = true
			}
		} else {
			p.mel[evt.ReqId] = &electionVrf{}
			p.mel[evt.ReqId].req = true
			p.mel[evt.ReqId].tm = time.Now()
		}
	case ELECTION_EVENT_RESULT:
		evt := event.Event.(*ElectionResultEvent)
		if e, ok := p.mel[evt.ReqId]; ok {
			for _, a := range e.rst {
				if a == evt.Ele.AddrHash {
					return true, nil
				}
			}
			e.rst = append(e.rst, evt.Ele.AddrHash)
		} else {
			p.mel[evt.ReqId] = &electionVrf{rst: make([]common.Hash, 0)}
			p.mel[evt.ReqId].rst = append(p.mel[evt.ReqId].rst, evt.Ele.AddrHash)
			p.mel[evt.ReqId].tm = time.Now()
		}
	}
	go p.ptn.ElectionBroadcast(*event)
	return false, nil
}

func (p *Processor) processElectionRequestEvent(ele *elector, reqEvt *ElectionRequestEvent) (result *ElectionEvent, err error) {
	//产生vrf证明
	//计算二项式分步，确定自己是否选中
	//如果选中，则对请求结果返回
	if len(p.local) < 1 {
		return nil, errors.New("ProcessElectionRequestEvent, local jury addr is nil")
	}
	addr := common.Address{}
	addrHash := common.Hash{}
	for addr, _ = range p.local {
		addrHash = util.RlpHash(addr)
		break //only first one
	}
	proof, err := ele.checkElected(conversionElectionSeedData(reqEvt.ReqId[:]))
	if err != nil {
		log.Error("ProcessElectionRequestEvent", "reqHash", reqEvt.ReqId, "checkElected err", err)
		return nil, err
	}
	pubKey, err := p.ptn.GetKeyStore().GetPublicKey(addr)
	if err != nil {
		log.Error("ProcessElectionRequestEvent", "get pubKey err, address:", addr)
		return nil, err
	}

	if proof != nil {
		rstEvt := &ElectionResultEvent{
			ReqId: reqEvt.ReqId,
			Ele:   modules.ElectionInf{AddrHash: addrHash, Proof: proof, PublicKey: pubKey},
		}
		log.Debug("ProcessElectionRequestEvent", "ok, reqId", reqEvt.ReqId.String())
		evt := &ElectionEvent{EType: ELECTION_EVENT_RESULT, Event: rstEvt}
		return evt, nil
	}
	return nil, nil
}

func (p *Processor) processElectionResultEvent(ele *elector, rstEvt *ElectionResultEvent) error {
	//验证vrf证明
	//收集vrf地址并添加缓存
	//检查缓存地址数量
	if _, ok := p.mtx[rstEvt.ReqId]; !ok {
		return errors.New("ProcessElectionResultEvent, reqHash not find")
	}
	mel := p.mel[rstEvt.ReqId]
	if mel.eChan == nil {
		//not request node
		return nil
	}
	if len(mel.eInf) >= p.electionNum {
		log.Info("ProcessElectionResultEvent, The quantity has reached the requirement")
		return nil
	}
	log.Info("ProcessElectionResultEvent", "reqHash", rstEvt.ReqId.String(), "addrHash", rstEvt.Ele.AddrHash.String())

	tmpReqId := common.BytesToAddress(rstEvt.ReqId.Bytes())
	contractId := common.NewAddress(tmpReqId.Bytes(), common.ContractHash)
	log.Debug("ProcessElectionResultEvent", "reqId", rstEvt.ReqId.Bytes(), "reqIdStr", rstEvt.ReqId.String(), "contractId", contractId, "tmpReqId", tmpReqId)
	log.Debug("ProcessElectionResultEvent", "contractIdBytes", contractId.Bytes(), "contractIdStr", contractId.String())

	ok, err := ele.verifyVrf(rstEvt.Ele.Proof, conversionElectionSeedData(rstEvt.ReqId[:]), rstEvt.Ele.PublicKey) //rstEvt.ReqId[:]
	if err != nil {
		log.Error("ProcessElectionResultEvent", "verify VRF fail, ReqId is", rstEvt.ReqId.Bytes())
		return err
	}
	if ok {
		p.locker.Lock()
		mel.eInf = append(mel.eInf, rstEvt.Ele)
		p.lockArf[contractId] = append(p.lockArf[contractId], rstEvt.Ele) //add lock vrf election info
		p.locker.Unlock()
		if len(mel.eInf) >= p.electionNum {
			//通知接收数量达到要求
			log.Info("ProcessElectionResultEvent,VRF address number is enough, Ok", "contractId", contractId)
			mel.eChan <- true
		}
	}
	return nil
}

func (p *Processor) ElectionRequest(reqId common.Hash, timeOut time.Duration) error {
	//return nil //todo

	if reqId == (common.Hash{}) {
		return errors.New("ElectionRequest param is nil")
	}
	seedData, err := getElectionSeedData(reqId)
	if err != nil {
		return err
	}
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
	log.Debug("ElectionRequest", "reqId", reqId.String(), "seedData", seedData)
	go p.ptn.ElectionBroadcast(ElectionEvent{EType: ELECTION_EVENT_REQUEST, Event: reqEvent})

	//超时等待选举结果
	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(timeOut)
		timeout <- true
	}()

	select {
	case <-p.mel[reqId].eChan:
		log.Debug("ElectionRequest, election Ok")
		return nil
	case <-timeout:
		log.Debug("ElectionRequest, election time out")
		return errors.New("ElectionRequest, election time out")
	}
	return errors.New("ElectionRequest, election fail")
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
	if !p.electionEventIsProcess(event, &account.Address) {
		log.Info("ProcessElectionEvent", "electionEventIsProcess if false, addr", account.Address, "event type", event.EType)
		return nil, nil
	}
	ele := &elector{
		num:      uint(p.electionNum),
		total:    uint64(p.dag.JuryCount()), //todo dynamic acquisition
		addr:     account.Address,
		password: account.Password,
		ks:       p.ptn.GetKeyStore(),
		//vrfAct: p.vrfAct,
	}
	ele.weight = electionWeightValue(ele.total)
	//log.Info("ProcessElectionEvent0", "event", event.EType)

	recved, err := p.electionEventBroadcast(event)
	if err != nil {
		return nil, err
	} else if recved {
		return nil, nil
	}
	log.Info("===ElectionMsg===", "event ", event)
	log.Info("ProcessElectionEvent", "event", event.EType)

	if event.EType == ELECTION_EVENT_REQUEST {
		return p.processElectionRequestEvent(ele, event.Event.(*ElectionRequestEvent))
	} else if event.EType == ELECTION_EVENT_RESULT {
		return nil, p.processElectionResultEvent(ele, event.Event.(*ElectionResultEvent))
	}
	return nil, errors.New("ProcessElectionEvent, fail")
}
