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
	"github.com/palletone/go-palletone/common/crypto"
	alg "github.com/palletone/go-palletone/consensus/jury/algorithm"
)

type vrfAccount struct {
	pubKey *ecdsa.PublicKey
	priKey *ecdsa.PrivateKey
}

type elector struct {
	num    uint
	weight uint64
	total  uint64
	vrfAct vrfAccount
}

func (e *elector) checkElected(data []byte) (proof []byte, err error) {
	if e.num < 0 || e.weight < 1 || data == nil {
		errs := fmt.Sprintf("checkElected param error, num[%d], weight[%d]", e.num, e.weight)
		return nil, errors.New(errs)
	}
	proof, err = vrfEc.VrfProve(e.vrfAct.priKey, data) //todo  后期调成为keystore中的vrf，先将功能调通
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

func (e *elector) verifyVRF(proof, data []byte) (bool, error) {
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

func (p *Processor) processElectionRequestEvent(ele *elector, reqEvt *ElectionRequestEvent) (result *ElectionEvent, err error) {
	//产生vrf证明
	//计算二项式分步，确定自己是否选中
	//如果选中，则对请求结果返回
	if len(p.local) < 1 {
		return nil, errors.New("ProcessElectionRequestEvent, local jury addr is nil")
	}
	addrHash := common.Hash{}
	for addr, _ := range p.local {
		addrHash = util.RlpHash(addr)
		break //only first one
	}
	log.Info("ProcessElectionRequestEvent", "reqHash", reqEvt.ReqHash.String(), "num", reqEvt.Num)
	proof, err := ele.checkElected(reqEvt.Data)
	if err != nil {
		log.Error("ProcessElectionRequestEvent", "reqHash", reqEvt.ReqHash, "checkElected err", err)
		return nil, err
	}
	if proof != nil {
		//if true { //todo for test
		rstEvt := &ElectionResultEvent{
			ReqHash:   reqEvt.ReqHash,
			AddrHash:  addrHash,
			Proof:     proof,
			PublicKey: crypto.CompressPubkey(p.vrfAct.pubKey), // *p.vrfAct.pubKey,
		}
		log.Debug("ProcessElectionRequestEvent", "reqId", reqEvt.ReqHash.String())
		evt := &ElectionEvent{EType: ELECTION_EVENT_RESULT, Event: rstEvt}
		return evt, nil
	}
	return nil, nil
}

func (p *Processor) processElectionResultEvent(ele *elector, rstEvt *ElectionResultEvent) error {
	//验证vrf证明
	//收集vrf地址并添加缓存
	//检查缓存地址数量
	log.Info("ProcessElectionResultEvent", "reqHash", rstEvt.ReqHash.String(), "addrHash", rstEvt.AddrHash.String())
	if _, ok := p.mtx[rstEvt.ReqHash]; !ok {
		return errors.New("ProcessElectionResultEvent, reqHash not find")
	}

	mtx := p.mtx[rstEvt.ReqHash]
	eleInfo := &mtx.eleInfo
	if len(mtx.addrHash) > int(eleInfo.eleNum) {
		log.Info("ProcessElectionResultEvent, The quantity has reached the requirement", "addrHash num ", eleInfo.eleNum)
		return nil
	}
	ok, err := ele.verifyVRF(rstEvt.Proof, eleInfo.seedData)
	if err != nil {
		return err
	}
	if ok {
		mtx.addrHash = append(mtx.addrHash, rstEvt.AddrHash)
		if eleInfo.contractId != (common.Address{}) {
			p.lockAddr[eleInfo.contractId] = append(p.lockAddr[eleInfo.contractId], rstEvt.AddrHash) //add addrHash
		}
		if len(mtx.addrHash) > int(eleInfo.eleNum) {
			//通知接收数量达到要求
			log.Info("ProcessElectionResultEvent,add num Ok")
			eleInfo.eleChan <- true
		}
	}
	return nil
}

func (p *Processor) ElectionRequest(reqId common.Hash, timeOut time.Duration) error {
	return nil //todo

	if reqId == (common.Hash{}) {
		return errors.New("ElectionRequest param is nil")
	}
	seedData, err := getElectionSeedData(reqId)
	if err != nil {
		return err
	}
	ele := electionInfo{
		eleChan:  make(chan bool, 1),
		eleNum:   uint(p.electionNum),
		seedData: seedData,
	}
	p.locker.Lock()
	p.mtx[reqId].eleInfo = ele
	p.locker.Unlock()
	reqEvent := &ElectionRequestEvent{
		ReqHash: reqId,
		Num:     ele.eleNum,
		Data:    ele.seedData,
	}
	log.Debug("ElectionRequest", "reqId", reqId.String(), "seedData", seedData)
	log.Debug("ElectionRequest", "reqEvent num", reqEvent.Num)
	go p.ptn.ElectionBroadcast(ElectionEvent{EType: ELECTION_EVENT_REQUEST, Event: reqEvent})

	//超时等待选举结果
	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(timeOut)
		timeout <- true
	}()

	select {
	case <-ele.eleChan:
		log.Debug("ElectionRequest, election Ok")
		return nil
	case <-timeout:
		log.Debug("ElectionRequest, election time out")
	}
	return errors.New("ElectionRequest, election time out")
}

func (p *Processor) ProcessElectionEvent(event *ElectionEvent) (result *ElectionEvent, err error) {
	if event == nil {
		return nil, errors.New("ProcessElectionRequestEvent, event is nil")
	}
	ele := &elector{
		num:    uint(p.electionNum),
		weight: 10,   //todo config
		total:  1000, //todo dynamic acquisition
		vrfAct: p.vrfAct,
	}
	if event.EType == ELECTION_EVENT_REQUEST {
		return p.processElectionRequestEvent(ele, event.Event.(*ElectionRequestEvent))
	} else if event.EType == ELECTION_EVENT_RESULT {
		return nil, p.processElectionResultEvent(ele, event.Event.(*ElectionResultEvent))
	}
	return nil, errors.New("ProcessElectionEvent, fail")
}
