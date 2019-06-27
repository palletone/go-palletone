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
	"bytes"
	"github.com/palletone/go-palletone/common/crypto"
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

func (p *Processor) selectElectionInf(local []modules.ElectionInf, recv []modules.ElectionInf, num int) []modules.ElectionInf {
	if len(local)+len(recv) < num {
		return nil
	}
	eles := make([]modules.ElectionInf, 0)
	if len(local) >= num {
		eles = local[:num]
	} else {
		less := num - len(local)
		eles = append(eles, local...)
		eles = append(eles, recv[:less]...)
	}
	return eles
}

func (p *Processor) electionEventIsProcess(event *ElectionEvent, addr *common.Address) (common.Hash, bool) {
	if event == nil {
		return common.Hash{}, false
	}
	reqId := common.Hash{}
	haveJ := p.localHaveActiveJury()
	haveM := p.ptn.LocalHaveActiveMediator()
	switch event.EType {
	case ELECTION_EVENT_VRF_REQUEST:
		reqId = event.Event.(*ElectionRequestEvent).ReqId
		if haveJ { //localHaveActiveJury()
			return reqId, true
		}
	case ELECTION_EVENT_VRF_RESULT:
		reqId = event.Event.(*ElectionResultEvent).ReqId
		if haveJ { //localHaveActiveJury()
			return reqId, true
		}
	case ELECTION_EVENT_SIG_REQUEST:
		reqId = event.Event.(*ElectionSigRequestEvent).ReqId
		if haveM {
			return reqId, true
		}
	case ELECTION_EVENT_SIG_RESULT:
		reqId = event.Event.(*ElectionSigResultEvent).ReqId
		if haveM {
			return reqId, true
		}
	}
	return reqId, false
}
func (p *Processor) electionEventBroadcast(event *ElectionEvent) (recved bool, invalid bool, err error) {
	if event == nil {
		return true, false, errors.New("electionEventBroadcast event is nil")
	}
	var reqId common.Hash
	p.locker.Lock()
	defer p.locker.Unlock()

	switch event.EType {
	case ELECTION_EVENT_VRF_REQUEST:
		evt := event.Event.(*ElectionRequestEvent)
		reqId = evt.ReqId
		if e, ok := p.mel[reqId]; ok {
			if e.vrfReqEd {
				return true, p.mel[reqId].invalid, nil
			}
			e.vrfReqEd = true
		}
	case ELECTION_EVENT_VRF_RESULT:
		evt := event.Event.(*ElectionResultEvent)
		reqId = evt.ReqId
		if e, ok := p.mel[reqId]; ok {
			for _, a := range e.rcvEle {
				if a.AddrHash == evt.Ele.AddrHash {
					return true, p.mel[reqId].invalid, nil
				}
			}
		}
	case ELECTION_EVENT_SIG_REQUEST:
		evt := event.Event.(*ElectionSigRequestEvent)
		reqId = evt.ReqId
		if e, ok := p.mel[reqId]; ok {
			if e.sigReqEd {
				return true, p.mel[reqId].invalid, nil
			}
			e.sigReqEd = true
		}
	case ELECTION_EVENT_SIG_RESULT:
		evt := event.Event.(*ElectionSigResultEvent)
		reqId = evt.ReqId
		if e, ok := p.mel[reqId]; ok {
			for _, sig := range e.sigs {
				if bytes.Equal(sig.PubKey, evt.Sig.PubKey) {
					return true, p.mel[reqId].invalid, nil
				}
			}
		}
	}
	if _, ok := p.mel[reqId]; !ok {
		p.mel[reqId] = &electionVrf{
			invalid: false,
			rcvEle:  make([]modules.ElectionInf, 0),
			sigs:    make([]modules.SignatureSet, 0),
		}
		p.mel[reqId].vrfReqEd = true
		p.mel[reqId].tm = time.Now()
	}
	go p.ptn.ElectionBroadcast(*event, false)
	return false, p.mel[reqId].invalid, nil
}

func (p *Processor) checkElectionSigRequestEventValid(evt *ElectionSigRequestEvent) bool {
	if evt == nil {
		return false
	}
	reqId := evt.ReqId
	if len(evt.Ele) != p.electionNum {
		log.Debugf("[%s]checkElectionSigRequestEventValid, len(%d)", shortId(reqId.String()), len(evt.Ele))
		return false
	}
	etor := &elector{
		num:   uint(p.electionNum),
		total: uint64(p.dag.JuryCount()), //需要固定
	}
	etor.weight = electionWeightValue(etor.total)
	//jjhAd, _, err := p.dag.GetConfig(modules.FoundationAddress)
	for i, e := range evt.Ele {
		if e.Etype == 1 {
			//if err == nil && bytes.Equal(reqAddr[:], jjhAd) {
			//	log.Debugf("[%s]checkElectionSigRequestEventValid, e.Etype == 1, ok, contractId[%s]", shortId(reqId.String()), string(contractId))
			//	continue
			//} else {
			//	log.Debugf("[%s]checkElectionSigRequestEventValid, e.Etype == 1, but not jjh request addr, contractId[%s]", shortId(reqId.String()), string(contractId))
			//	log.Debugf("[%s]checkElectionSigRequestEventValid, reqAddr[%s], jjh[%s]", shortId(reqId.String()), string(reqAddr[:]), string(jjhAd))
			//
			//	return false
			//}
			continue
		}

		//验证proof是否通过
		isVerify, err := etor.verifyVrf(e.Proof, getElectionSeedData(reqId), e.PublicKey)
		if err != nil || !isVerify {
			log.Infof("[%s]checkElectionSigRequestEventValid, index[%d],verifyVrf fail", shortId(reqId.String()), i)
			return false
		}
	}

	return true
}

func (p *Processor) checkElectionSigResultEventValid(evt *ElectionSigResultEvent) bool {
	if evt == nil {
		return false
	}
	reqId := evt.ReqId
	if _, ok := p.mel[reqId]; !ok {
		return false
	}
	if p.mtx[reqId].eleInf == nil {
		return false
	}
	ele := p.mtx[reqId].eleInf
	reqEvt := &ElectionSigRequestEvent{
		ReqId: reqId,
		Ele:   ele,
	}
	hash := util.RlpHash(reqEvt)
	if !crypto.VerifySignature(evt.Sig.PubKey, hash.Bytes(), evt.Sig.Signature) {
		log.Debugf("[%s]checkElectionSigResultEventValid, VerifySignature fail", shortId(reqId.String()))
		log.Debug("checkElectionSigResultEventValid", "reqEvt", reqEvt, "PubKey", evt.Sig.PubKey, "Signature", evt.Sig.Signature, "hash", hash)
		return false
	}
	return true
}

func (p *Processor) processElectionRequestEvent(elr *elector, reqEvt *ElectionRequestEvent) (err error) {
	//产生vrf证明
	//计算二项式分步，确定自己是否选中
	//如果选中，则对请求结果返回
	if elr == nil || reqEvt == nil {
		return errors.New("processElectionRequestEvent, param is nil")
	}
	p.locker.Lock()
	defer p.locker.Unlock()

	reqId := reqEvt.ReqId
	if !p.localHaveActiveJury() {
		return fmt.Errorf("processElectionRequestEvent, local jury addr is nil, reqId[%s]", reqId.String())
	}
	addr := common.Address{}
	addrHash := common.Hash{}
	for addr, _ = range p.local {
		addrHash = util.RlpHash(addr)
		break //only first one
	}
	proof, err := elr.checkElected(getElectionSeedData(reqEvt.ReqId))
	if err != nil {
		log.Errorf("[%s]processElectionRequestEvent, checkElected err, %s", shortId(reqId.String()), err.Error())
		return fmt.Errorf("processElectionRequestEvent, checkElected err, reqId[%s]", shortId(reqId.String()))
	}
	pubKey, err := p.ptn.GetKeyStore().GetPublicKey(addr)
	if err != nil {
		log.Errorf("[%s]processElectionRequestEvent, get pubKey err, address[%s]", shortId(reqId.String()), addr.String())
		return fmt.Errorf("processElectionRequestEvent, get pubKey err,reqId[%s]", shortId(reqId.String()))
	}
	if proof != nil {
		rstEvt := &ElectionResultEvent{
			ReqId: reqEvt.ReqId,
			Ele:   modules.ElectionInf{AddrHash: addrHash, Proof: proof, PublicKey: pubKey},
		}
		log.Debugf("[%s]processElectionRequestEvent, ok", shortId(reqId.String()))
		go p.ptn.ElectionBroadcast(ElectionEvent{EType: ELECTION_EVENT_VRF_RESULT, Event: rstEvt}, true)
		return nil
	}
	return nil
}

func (p *Processor) processElectionResultEvent(elr *elector, rstEvt *ElectionResultEvent) error {
	//验证vrf证明
	//收集vrf地址并添加缓存
	//检查缓存地址数量
	if elr == nil || rstEvt == nil {
		return errors.New("processElectionResultEvent, param is nil")
	}
	p.locker.Lock()
	defer p.locker.Unlock()

	reqId := rstEvt.ReqId
	mel := p.mel[reqId]
	if mel == nil {
		mel = &electionVrf{
			rcvEle: make([]modules.ElectionInf, 0),
			sigs:   make([]modules.SignatureSet, 0),
		}
	}
	if len(mel.rcvEle) >= p.electionNum*2 {
		log.Infof("[%s]processElectionResultEvent, The quantity has reached the requirement", shortId(reqId.String()))
		return nil
	}
	log.Infof("[%s]processElectionResultEvent, ele addrHash[%s]", shortId(reqId.String()), rstEvt.Ele.AddrHash.String())
	//检查是否重复收到
	for _, ele := range p.mel[reqId].rcvEle {
		if bytes.Equal(rstEvt.Ele.AddrHash.Bytes(), ele.AddrHash.Bytes()) {
			log.Infof("[%s]processElectionResultEvent, ele already add, addrHash[%s]", shortId(reqId.String()), rstEvt.Ele.AddrHash.String())
			return nil
		}
	}
	//验证vrf
	ok, err := elr.verifyVrf(rstEvt.Ele.Proof, getElectionSeedData(reqId), rstEvt.Ele.PublicKey) //rstEvt.ReqId[:]
	if err != nil {
		log.Errorf("[%s]processElectionResultEvent, verify VRF fail", shortId(reqId.String()))
		return fmt.Errorf("processElectionResultEvent, verify VRF fail, reqId[%s]", shortId(reqId.String()))
	}
	//接收vrf
	if ok {
		log.Debugf("[%s]processElectionResultEvent, add ele, addHash[%s]", shortId(reqId.String()), rstEvt.Ele.AddrHash.String())
		mel.rcvEle = append(mel.rcvEle, rstEvt.Ele)
	}
	return nil
}

func (p *Processor) processElectionSigRequestEvent(evt *ElectionSigRequestEvent) error {
	//检查是否接收过
	//检查ele有效性
	if evt == nil {
		return errors.New("processElectionSigRequestEvent, param is nil")
	}
	if !p.ptn.LocalHaveActiveMediator() {
		return errors.New("processElectionSigRequestEvent, local no active mediator")
	}
	reqId := evt.ReqId
	if !p.checkElectionSigRequestEventValid(evt) {
		log.Debugf("[%s]processElectionSigRequestEvent, evt is invalid", shortId(reqId.String()))
		return nil
	}
	mAddrs := p.ptn.GetLocalActiveMediators()
	if len(mAddrs) < 1 {
		log.Debugf("[%s]processElectionSigRequestEvent,LocalActiveMediators < 1", shortId(reqId.String()))
		return  nil
	}
	mAddr := mAddrs[0] //first
	ks := p.ptn.GetKeyStore()
	pk, err := ks.GetPublicKey(mAddr)
	if err != nil {
		log.Debugf("[%s]processElectionSigRequestEvent, GetPublicKey fail", shortId(reqId.String()))
		return nil
	}
	sig, err := ks.SigData(evt, mAddr)
	if err != nil {
		log.Debugf("[%s]processElectionSigRequestEvent, SigData fail", shortId(reqId.String()))
		return nil
	}
	//todo
	//hash := util.RlpHash(evt)
	//if !crypto.VerifySignature(pk, hash.Bytes(), sig) {
	//	log.Debugf("[%s]processElectionSigRequestEvent, VerifySignature fail", shortId(reqId.String()))
	//}
	log.Debug("processElectionSigRequestEvent", "reqId", shortId(reqId.String()), "evt", evt, "PubKey", pk, "Signature", sig, "hash", hash)
	if e, ok := p.mel[reqId]; ok {
		e.brded = true //关闭签名广播请求
		e.sigReqEd = true
	}
	resultEvt := &ElectionSigResultEvent{
		ReqId: reqId,
		Sig:   modules.SignatureSet{PubKey: pk, Signature: sig},
	}
	//广播resultEvt
	go p.ptn.ElectionBroadcast(ElectionEvent{EType: ELECTION_EVENT_SIG_RESULT, Event: resultEvt}, true)
	return nil
}

func (p *Processor) processElectionSigResultEvent(evt *ElectionSigResultEvent) error {
	if evt == nil {
		return errors.New("processElectionSigResultEvent, param is nil")
	}
	p.locker.Lock()
	defer p.locker.Unlock()
	reqId := evt.ReqId
	mel := p.mel[reqId]
	if mel == nil || mel.nType != 1 {
		return nil
	}
	//检查签名是否重复，收集签名，数量满足则广播请求交易
	for _, sig := range mel.sigs {
		if bytes.Equal(sig.PubKey, evt.Sig.PubKey) {
			log.Debugf("[%s]processElectionSigResultEvent, event already receive", shortId(reqId.String()))
			return nil
		}
	}
	//验证签名有效性
	if !p.checkElectionSigResultEventValid(evt) {
		log.Infof("[%s]processElectionSigResultEvent, checkElectionSigResultEventValid fail", shortId(reqId.String()))
		return errors.New("processElectionSigResultEvent SigResultEvent is invalid")
	}
	//验证签名者是否为Mediator
	addr := crypto.PubkeyBytesToAddress(evt.Sig.PubKey)
	if !p.dag.IsActiveMediator(addr) {
		log.Debugf("[%s]processElectionSigResultEvent, not mediator, addr[%s]", shortId(reqId.String()), addr.String())
		return nil
	}
	mel.sigs = append(mel.sigs, evt.Sig)
	log.Debugf("[%s]processElectionSigResultEvent,sig num=%d, add sig[%s], Threshold=%d", shortId(reqId.String()), len(mel.sigs), evt.Sig.String(), p.dag.ChainThreshold())
	if len(mel.sigs) >= p.dag.ChainThreshold() {
		event := ContractEvent{
			CType: CONTRACT_EVENT_EXEC,
			Ele:   p.mtx[reqId].eleInf,
			Tx:    p.mtx[reqId].reqTx,
		}
		log.Infof("[%s]processElectionSigResultEvent, CONTRACT_EVENT_EXEC", shortId(reqId.String()))
		log.Info("processElectionSigResultEvent======================================================ok")
		log.Info("processElectionSigResultEvent, CONTRACT_EVENT_EXEC", "reqId", shortId(reqId.String()), "event", event)
		go p.ptn.ContractBroadcast(event, true)
		return nil
	}
	return nil
}

func (p *Processor) BroadcastElectionSigRequestEvent() {
	p.locker.Lock()
	defer p.locker.Unlock()
	for reqId, ele := range p.mel {
		if ele.brded {
			continue
		}
		mtx := p.mtx[reqId]
		if mtx == nil || mtx.reqTx == nil {
			continue
		}
		if (len(mtx.eleInf) + len(ele.rcvEle)) >= p.electionNum {
			se := p.selectElectionInf(mtx.eleInf, ele.rcvEle, p.electionNum)
			mtx.eleInf = se
			event := &ElectionSigRequestEvent{
				ReqId: reqId,
				Ele:   se,
			}
			ele.brded = true
			ele.nType = 1
			log.Infof("[%s]BroadcastElectionSigRequestEvent ", shortId(reqId.String()))
			log.Debug("BroadcastElectionSigRequestEvent", "event", event, "len(mtx.eleInf)", len(mtx.eleInf), "len(ele.rcvEle)", len(ele.rcvEle))
			go p.ptn.ElectionBroadcast(ElectionEvent{EType: ELECTION_EVENT_SIG_REQUEST, Event: event}, true)
		}
	}
}

func (p *Processor) ProcessElectionEvent(event *ElectionEvent) (result *ElectionEvent, err error) {
	if event == nil {
		return nil, errors.New("ProcessElectionEvent, event is nil")
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

	//log.Infof("[%s]ProcessElectionEvent--, event type[%v] ", shortId(reqId.String()), event.EType) //del
	recved, invalid, err := p.electionEventBroadcast(event)
	if err != nil {
		return nil, err
	} else if recved || invalid {
		log.Debugf("[%s]ProcessElectionEvent, recved=%v, invalid=%v", shortId(reqId.String()), recved, invalid)
		return nil, nil
	}
	log.Infof("[%s]ProcessElectionEvent, event type[%v] ", shortId(reqId.String()), event.EType)
	//go p.ptn.ElectionBroadcast(*event, false)

	if event.EType == ELECTION_EVENT_VRF_REQUEST {
		err = p.processElectionRequestEvent(ele, event.Event.(*ElectionRequestEvent))
	} else if event.EType == ELECTION_EVENT_VRF_RESULT {
		err = p.processElectionResultEvent(ele, event.Event.(*ElectionResultEvent))
	} else if event.EType == ELECTION_EVENT_SIG_REQUEST {
		err = p.processElectionSigRequestEvent(event.Event.(*ElectionSigRequestEvent))
	} else if event.EType == ELECTION_EVENT_SIG_RESULT {
		err = p.processElectionSigResultEvent(event.Event.(*ElectionSigResultEvent))
	}

	return nil, err
}
