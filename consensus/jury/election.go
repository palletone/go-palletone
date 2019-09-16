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
	"bytes"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/util"
	"github.com/palletone/go-palletone/core/accounts"
	"github.com/palletone/go-palletone/core/accounts/keystore"
	"github.com/palletone/go-palletone/dag/errors"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/common/crypto"
	alg "github.com/palletone/go-palletone/consensus/jury/vrf/algorithm"

	"github.com/palletone/go-palletone/consensus/jury/vrf"
)

type elector struct {
	num      uint
	weight   uint64
	total    uint64
	addr     common.Address
	password string
	ks       *keystore.KeyStore
}
func newElector(num uint, total uint64, addr common.Address, password string, ks *keystore.KeyStore) *elector {
	e := &elector{
		num:      num,
		weight:   electionWeightValue(total),
		total:    total,
		addr:     addr,
		password: password,
		ks:       ks,
	}
	return e
}

func (e *elector) checkElected(data []byte) (proof []byte, err error) {
	if e.weight < 1 || data == nil {
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

	proof, sel, err := vrf.VrfProve(privateKey.(*ecdsa.PrivateKey), data)
	if err != nil {
		return nil, err
	}
	if len(sel) > 0 {
		if alg.Selected(e.num, e.weight, e.total, sel) > 0 {
			return proof, nil
		}
	}
	return nil, nil
}

func (e *elector) verifyVrf(proof, data []byte, pubKey []byte) (bool, error) {
	ok, pro, err := vrf.VrfVerify(pubKey, data, proof)
	if err != nil {
		log.Error("verifyVrf fail", "ok?", ok)
		return false, err
	}
	if ok {
		vrfValue := pro
		if len(vrfValue) > 0 {
			if alg.Selected(e.num, e.weight, e.total, vrfValue) > 0 {
				return true, nil
			}
		}
	}
	return false, nil
}

func (p *Processor) selectElectionInf(local []modules.ElectionInf,
	recv []modules.ElectionInf, num int) ([]modules.ElectionInf, bool) {
	if len(local)+len(recv) < num {
		return nil, false
	}
	eels := make([]modules.ElectionInf, 0)
	if len(local) >= num { //use local
		eels = local[:num]
	} else {
		less := num - len(local)
		eels = append(eels, local...)
		for i := 0; i < less; i++ {
			ok := true
			for _, l := range local {
				if bytes.Equal(l.AddrHash[:], recv[i].AddrHash[:]) {
					ok = false
					break
				}
			}
			if ok {
				log.Debug("selectElectionInf", "i", i, "add ele", recv[i].AddrHash.String())
				eels = append(eels, recv[i])
			}
		}
		if len(eels) < num {
			log.Debug("selectElectionInf", "len(eels):", len(eels), "< num:", num)
			return nil, false
		}
	}
	return eels, true
}

func (p *Processor) electionEventIsProcess(event *ElectionEvent) (common.Hash, bool) {
	if event == nil {
		return common.Hash{}, false
	}
	reqId := common.Hash{}
	jCnt := uint64(p.dag.JuryCount())
	haveJ := p.localHaveActiveJury()
	haveM := p.ptn.LocalHaveActiveMediator()
	switch event.EType {
	case ELECTION_EVENT_VRF_REQUEST:
		evt := event.Event.(*ElectionRequestEvent)
		reqId = evt.ReqId
		if !checkJuryCountValid(evt.JuryCount, jCnt) {
			return reqId, false
		}
		if haveJ { //localHaveActiveJury()
			return reqId, true
		}
	case ELECTION_EVENT_VRF_RESULT:
		evt := event.Event.(*ElectionResultEvent)
		reqId = evt.ReqId
		if evt.Ele.EType != 1 {
			if !checkJuryCountValid(evt.JuryCount, jCnt) {
				return reqId, false
			}
		}
		if haveJ { //localHaveActiveJury()
			return reqId, true
		}
	case ELECTION_EVENT_SIG_REQUEST:
		evt := event.Event.(*ElectionSigRequestEvent)
		reqId = evt.ReqId
		for _, e := range evt.Ele {
			if e.EType != 1 {
				if !checkJuryCountValid(evt.JuryCount, jCnt) {
					return reqId, false
				}
			}
		}
		if haveM {
			return reqId, true
		}
	case ELECTION_EVENT_SIG_RESULT:
		evt := event.Event.(*ElectionSigResultEvent)
		reqId = evt.ReqId
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
	etor := newElector(uint(p.electionNum), evt.JuryCount, common.Address{},"", nil)
	for i, e := range evt.Ele {
		if e.EType == 1 { //todo
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
	if p.mtx[reqId].eleNode == nil {
		return false
	}
	ele := p.mtx[reqId].eleNode
	reqEvt := &ElectionSigRequestEvent{
		ReqId:     reqId,
		JuryCount: evt.JuryCount,
		Ele:       ele.EleList,
	}
	hash := util.RlpHash(reqEvt)
	if !crypto.VerifySignature(evt.Sig.PubKey, hash.Bytes(), evt.Sig.Signature) {
		log.Debugf("[%s]checkElectionSigResultEventValid, VerifySignature fail", shortId(reqId.String()))
		log.Debug("checkElectionSigResultEventValid", "reqEvt", reqEvt, "PubKey", evt.Sig.PubKey,
			"Signature", evt.Sig.Signature, "hash", hash)
		return false
	}
	return true
}

func (p *Processor) processElectionRequestEvent(reqEvt *ElectionRequestEvent) (err error) {
	//产生vrf证明
	//计算二项式分步，确定自己是否选中
	//如果选中，则对请求结果返回
	if reqEvt == nil {
		return errors.New("processElectionRequestEvent, param is nil")
	}
	p.locker.Lock()
	defer p.locker.Unlock()

	reqId := reqEvt.ReqId
	if !p.localHaveActiveJury() {
		return fmt.Errorf("processElectionRequestEvent, local jury addr is nil, reqId[%s]", reqId.String())
	}
	account := p.getLocalJuryAccount()
	if account == nil {
		return errors.New("processElectionRequestEvent, getLocalJuryAccount fail")
	}
	elr := newElector(uint(p.electionNum), reqEvt.JuryCount, account.Address,account.Password, p.ptn.GetKeyStore())

	addrHash := util.RlpHash(account.Address)
	proof, err := elr.checkElected(getElectionSeedData(reqEvt.ReqId))
	if err != nil {
		log.Errorf("[%s]processElectionRequestEvent, checkElected err, %s", shortId(reqId.String()), err.Error())
		return fmt.Errorf("processElectionRequestEvent, checkElected err, reqId[%s]", shortId(reqId.String()))
	}
	pubKey, err := p.ptn.GetKeyStore().GetPublicKey(account.Address)
	if err != nil {
		log.Errorf("[%s]processElectionRequestEvent, get pubKey err, address[%s]",
			shortId(reqId.String()), account.Address.String())
		return fmt.Errorf("processElectionRequestEvent, get pubKey err,reqId[%s]", shortId(reqId.String()))
	}
	if proof != nil {
		rstEvt := &ElectionResultEvent{
			ReqId:     reqEvt.ReqId,
			JuryCount: reqEvt.JuryCount,
			Ele:       modules.ElectionInf{EType: 0, AddrHash: addrHash, Proof: proof, PublicKey: pubKey},
		}
		log.Debugf("[%s]processElectionRequestEvent, ok", shortId(reqId.String()))
		go p.ptn.ElectionBroadcast(ElectionEvent{EType: ELECTION_EVENT_VRF_RESULT, Event: rstEvt}, true)
		return nil
	}
	return nil
}

func (p *Processor) processElectionResultEvent(rstEvt *ElectionResultEvent) error {
	//验证vrf证明
	//收集vrf地址并添加缓存
	//检查缓存地址数量
	if rstEvt == nil {
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
			log.Infof("[%s]processElectionResultEvent, ele already add, addrHash[%s]",
				shortId(reqId.String()), rstEvt.Ele.AddrHash.String())
			return nil
		}
	}
	//验证vrf
	elr := newElector(uint(p.electionNum), rstEvt.JuryCount, common.Address{},"", p.ptn.GetKeyStore())
	ok, err := elr.verifyVrf(rstEvt.Ele.Proof, getElectionSeedData(reqId), rstEvt.Ele.PublicKey) //rstEvt.ReqId[:]
	if err != nil {
		log.Errorf("[%s]processElectionResultEvent, verify VRF fail", shortId(reqId.String()))
		return fmt.Errorf("processElectionResultEvent, verify VRF fail, reqId[%s]", shortId(reqId.String()))
	}
	//接收vrf
	if ok {
		log.Debugf("[%s]processElectionResultEvent, add ele, addHash[%s]",
			shortId(reqId.String()), rstEvt.Ele.AddrHash.String())
		mel.juryCnt = rstEvt.JuryCount
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
		return nil
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

	if e, ok := p.mel[reqId]; ok {
		e.brded = true //关闭签名广播请求
		e.sigReqEd = true
	}
	resultEvt := &ElectionSigResultEvent{
		ReqId:     reqId,
		JuryCount: evt.JuryCount,
		Sig:       modules.SignatureSet{PubKey: pk, Signature: sig},
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
	if len(mel.sigs) >= p.dag.ChainThreshold() {
		log.Debugf("[%s]processElectionSigResultEvent, sig  number is enough", shortId(reqId.String()))
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
	log.Debugf("[%s]processElectionSigResultEvent,sig num=%d, add sig[%s], Threshold=%d",
		shortId(reqId.String()), len(mel.sigs), evt.Sig.String(), p.dag.ChainThreshold())
	if len(mel.sigs) >= p.dag.ChainThreshold() {
		event := ContractEvent{
			CType: CONTRACT_EVENT_EXEC,
			Ele:   p.mtx[reqId].eleNode,
			Tx:    p.mtx[reqId].reqTx,
		}
		log.Infof("[%s]processElectionSigResultEvent, CONTRACT_EVENT_EXEC", shortId(reqId.String()))
		log.Info("processElectionSigResultEvent, CONTRACT_EVENT_EXEC", "reqId",
			shortId(reqId.String()), "event", event)
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
		var eList []modules.ElectionInf
		if mtx.eleNode != nil {
			eList = mtx.eleNode.EleList
		} else {
			eList = nil
			mtx.eleNode = &modules.ElectionNode{JuryCount: ele.juryCnt}
		}

		if len(eList)+len(ele.rcvEle) >= p.electionNum {
			se, valid := p.selectElectionInf(eList, ele.rcvEle, p.electionNum)
			if !valid {
				continue
			}
			mtx.eleNode.EleList = se
			event := &ElectionSigRequestEvent{
				ReqId:     reqId,
				JuryCount: ele.juryCnt,
				Ele:       se,
			}
			ele.brded = true
			ele.nType = 1
			log.Infof("[%s]BroadcastElectionSigRequestEvent ", shortId(reqId.String()))
			log.Debug("BroadcastElectionSigRequestEvent", "event", event,
				"len(mtx.eleNode)", len(mtx.eleNode.EleList), "len(ele.rcvEle)", len(ele.rcvEle))
			go p.ptn.ElectionBroadcast(ElectionEvent{EType: ELECTION_EVENT_SIG_REQUEST, Event: event}, true)
		}
	}
}

func (p *Processor) ProcessElectionEvent(event *ElectionEvent) (result *ElectionEvent, err error) {
	if event == nil {
		return nil, errors.New("ProcessElectionEvent, event is nil")
	}
	reqId, isP := p.electionEventIsProcess(event)
	if !isP {
		log.Infof("[%s]ProcessElectionEvent, electionEventIsProcess is false, event type[%v]",
			shortId(reqId.String()), event.EType)
		return nil, nil
	}
	received, invalid, err := p.electionEventBroadcast(event)
	if err != nil {
		return nil, err
	} else if received || invalid {
		log.Debugf("[%s]ProcessElectionEvent, received=%v, invalid=%v", shortId(reqId.String()), received, invalid)
		return nil, nil
	}
	log.Infof("[%s]ProcessElectionEvent, event type[%v] ", shortId(reqId.String()), event.EType)

	if event.EType == ELECTION_EVENT_VRF_REQUEST {
		err = p.processElectionRequestEvent(event.Event.(*ElectionRequestEvent))
	} else if event.EType == ELECTION_EVENT_VRF_RESULT {
		err = p.processElectionResultEvent(event.Event.(*ElectionResultEvent))
	} else if event.EType == ELECTION_EVENT_SIG_REQUEST {
		err = p.processElectionSigRequestEvent(event.Event.(*ElectionSigRequestEvent))
	} else if event.EType == ELECTION_EVENT_SIG_RESULT {
		err = p.processElectionSigResultEvent(event.Event.(*ElectionSigResultEvent))
	}

	return nil, err
}
