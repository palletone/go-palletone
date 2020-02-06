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
 * @author PalletOne core developer Albert·Gou <dev@pallet.one>
 * @date 2018
 */

package mediatorplugin

import (
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/event"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/hexutil"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/dag/modules"
	"go.dedis.ch/kyber/v3/share"
	"go.dedis.ch/kyber/v3/share/dkg/pedersen"
	"go.dedis.ch/kyber/v3/sign/tbls"
)

func (mp *MediatorPlugin) signUnitsTBLS(localMed common.Address) {
	mp.toTBLSBufLock.RLock()
	defer mp.toTBLSBufLock.RUnlock()

	medUnitsBuf, ok := mp.toTBLSSignBuf[localMed]
	if !ok {
		log.Debugf("the mediator(%v) has no units to sign TBLS yet", localMed.Str())
		return
	}

	for unitHash := range medUnitsBuf {
		go mp.signUnitTBLS(localMed, unitHash)
	}
}

func (mp *MediatorPlugin) recoverUnitsTBLS(localMed common.Address) {
	mp.toTBLSBufLock.RLock()
	defer mp.toTBLSBufLock.RUnlock()

	sigSharesBuf, ok := mp.toTBLSRecoverBuf[localMed]
	if !ok {
		log.Debugf("the mediator(%v) has no signature shares to recover group sign yet", localMed.Str())
		return
	}

	for unitHash := range sigSharesBuf {
		go mp.recoverUnitTBLS(localMed, unitHash)
	}
}

func (mp *MediatorPlugin) AddToTBLSSignBufs(newHash common.Hash) {
	if !mp.groupSigningEnabled {
		return
	}

	header, err := mp.dag.GetHeaderByHash(newHash)
	if header == nil {
		err = fmt.Errorf("fail to get header by hash: %v, err: %v", newHash.TerminalString(), err.Error())
		log.Errorf(err.Error())
		return
	}

	var ms []common.Address
	// 严格要求换届unix时间是产块间隔的整数倍
	if header.Timestamp() > mp.dag.LastMaintenanceTime() {
		ms = mp.GetLocalActiveMediators()
	} else {
		ms = mp.GetLocalPrecedingMediators()
	}

	for _, localMed := range ms {
		log.Debugf("the mediator(%v) received a unit(hash: %v , # %v) to be group-signed",
			localMed.Str(), newHash.TerminalString(), header.NumberU64())
		go mp.addToTBLSSignBuf(localMed, newHash)
	}
}

func (mp *MediatorPlugin) addToTBLSSignBuf(localMed common.Address, unitHash common.Hash) {
	mp.toTBLSBufLock.Lock()
	//log.Debugf("toTBLSBufLock.Lock()")
	if _, ok := mp.toTBLSSignBuf[localMed]; !ok {
		mp.toTBLSSignBuf[localMed] = make(map[common.Hash]bool)
	}

	mp.toTBLSSignBuf[localMed][unitHash] = true
	//log.Debugf("toTBLSBufLock.Unlock()")
	mp.toTBLSBufLock.Unlock()

	go mp.signUnitTBLS(localMed, unitHash)

	// 当 unit 过了确认时间后，及时删除待群签名的 unit，防止内存溢出
	expiration := mp.dag.UnitIrreversibleTime()
	deleteBuf := time.NewTimer(expiration)

	select {
	case <-mp.quit:
		return
	case <-deleteBuf.C:
		mp.toTBLSBufLock.Lock()
		//log.Debugf("toTBLSBufLock.Lock()")
		if _, ok := mp.toTBLSSignBuf[localMed][unitHash]; ok {
			log.Debugf("the unit(%v) has expired confirmation time, no longer need the mediator(%v)"+
				" to sign-group", unitHash.TerminalString(), localMed.Str())
			delete(mp.toTBLSSignBuf[localMed], unitHash)
		}
		//log.Debugf("toTBLSBufLock.Unlock()")
		mp.toTBLSBufLock.Unlock()
	}
}

func (mp *MediatorPlugin) SubscribeSigShareEvent(ch chan<- SigShareEvent) event.Subscription {
	return mp.sigShareScope.Track(mp.sigShareFeed.Subscribe(ch))
}

func (mp *MediatorPlugin) signUnitTBLS(localMed common.Address, unitHash common.Hash) {
	mp.toTBLSBufLock.Lock()
	//log.Debugf("toTBLSBufLock.Lock()")
	//defer log.Debugf("toTBLSBufLock.Unlock()")
	defer mp.toTBLSBufLock.Unlock()

	medUnitsBuf, ok := mp.toTBLSSignBuf[localMed]
	if !ok {
		log.Debugf("the mediator(%v) has no units to sign TBLS yet", localMed.Str())
		return
	}

	dag := mp.dag
	mp.dkgLock.RLock()
	defer mp.dkgLock.RUnlock()
	var (
		dkgr   *dkg.DistKeyGenerator
		header *modules.Header
		err    error
	)

	// 1. 获取群签名所需数据
	{
		_, ok := medUnitsBuf[unitHash]
		if !ok {
			log.Debugf("the mediator(%v) has no unit(%v) to sign TBLS, or the unit has been signed by mediator",
				localMed.Str(), unitHash.TerminalString())
			return
		}

		header, err = dag.GetHeaderByHash(unitHash)
		if header == nil {
			err = fmt.Errorf("fail to get header by hash: %v, err: %v", unitHash.TerminalString(), err.Error())
			log.Errorf(err.Error())
			return
		}

		// 判断是否是换届前的单元
		if header.Timestamp() > mp.lastMaintenanceTime {
			dkgr, ok = mp.activeDKGs[localMed]
		} else {
			dkgr, ok = mp.precedingDKGs[localMed]
		}

		if !ok {
			log.Debugf("the mediator(%v)'s dkg is not existed", localMed.Str())
			return
		}
	}

	// 2. 判断群签名的相关条件
	{
		// 如果单元没有群公钥， 则跳过群签名
		pkb := header.GetGroupPubKeyByte()
		if len(pkb) == 0 {
			err := fmt.Errorf("this unit(hash: %v , # %v )'s group public key is null",
				unitHash.TerminalString(), header.NumberU64())
			log.Debug(err.Error())
			return
		}

		// 判断父 unit 是否不可逆
		parentHash := header.ParentHash()[0]
		isStable, err := dag.IsIrreversibleUnit(parentHash)
		if err != nil {
			return
		}
		if !isStable {
			log.Debugf("the unit(hash: %v , # %v )'s parent unit(%v) is not irreversible",
				unitHash.TerminalString(), header.NumberU64(), parentHash.TerminalString())
			return
		}
	}

	// 3. 群签名
	dks, err := dkgr.DistKeyShare()
	if err != nil {
		log.Debugf("the mediator(%v)'s dkg get dks err:%v", localMed.Str(), err.Error())
		return
	}

	sigShare, err := tbls.Sign(mp.suite, dks.PriShare(), unitHash[:])
	if err != nil {
		log.Debugf("the mediator(%v)'s TBLS sign the unit(%v) err:%v",
			localMed.Str(), unitHash.TerminalString(), err.Error())
		return
	}

	// 4. 群签名成功后的处理
	log.Debugf("the mediator(%v) signed-group the unit(hash: %v , # %v)", localMed.Str(),
		unitHash.TerminalString(), header.NumberU64())
	delete(mp.toTBLSSignBuf[localMed], unitHash)

	event := SigShareEvent{
		UnitHash: unitHash,
		SigShare: sigShare,
		Deadline: mp.getGroupSignMessageDeadline(),
	}

	go mp.sigShareFeed.Send(event)
}

// 收集签名分片
func (mp *MediatorPlugin) AddToTBLSRecoverBuf(event *SigShareEvent) {
	if !mp.groupSigningEnabled {
		return
	}

	newUnitHash := event.UnitHash
	log.Debugf("received the sign shares of the unit(%v)", newUnitHash.TerminalString())

	dag := mp.dag
	header, err := dag.GetHeaderByHash(newUnitHash)
	if header == nil {
		err = fmt.Errorf("fail to get unit(%v), err: %v", newUnitHash.TerminalString(), err.Error())
		log.Errorf(err.Error())
		return
	}

	localMed := header.Author()
	mp.toTBLSBufLock.Lock()
	//log.Debugf("toTBLSBufLock.Lock()")
	//defer log.Debugf("toTBLSBufLock.Unlock()")
	defer mp.toTBLSBufLock.Unlock()

	medSigSharesBuf, ok := mp.toTBLSRecoverBuf[localMed]
	if !ok {
		// 不是本地mediator生产的 unit
		errStr := fmt.Errorf("the mediator(%v) of the unit(hash: %v, # %v ) is not local",
			localMed.Str(), newUnitHash.TerminalString(), header.NumberU64())
		log.Debugf(errStr.Error())
		return
	}

	// 当buf不存在时，说明已经成功recover出群签名, 或者已经过了unit确认时间，不需要群签名，忽略该签名分片
	sigShareSet, ok := medSigSharesBuf[newUnitHash]
	if !ok {
		errStr := fmt.Errorf("the unit(hash: %v, # %v ) need not to recover group-sign by the mediator(%v)",
			newUnitHash.TerminalString(), header.NumberU64(), localMed.Str())
		log.Debugf(errStr.Error())
		return
	}

	sigShareSet.append(event.SigShare)

	// recover群签名
	go mp.recoverUnitTBLS(localMed, newUnitHash)
}

func (mp *MediatorPlugin) SubscribeGroupSigEvent(ch chan<- GroupSigEvent) event.Subscription {
	return mp.groupSigScope.Track(mp.groupSigFeed.Subscribe(ch))
}

func (mp *MediatorPlugin) recoverUnitTBLS(localMed common.Address, unitHash common.Hash) {
	mp.toTBLSBufLock.Lock()
	//log.Debugf("toTBLSBufLock.Lock()")
	//defer log.Debugf("toTBLSBufLock.Unlock()")
	defer mp.toTBLSBufLock.Unlock()

	// 1. 获取所有的签名分片
	sigSharesBuf, ok := mp.toTBLSRecoverBuf[localMed]
	if !ok {
		log.Debugf("the mediator((%v) has not units to recover group sign yet", localMed.Str())
		return
	}

	sigShareSet, ok := sigSharesBuf[unitHash]
	if !ok {
		log.Debugf("the mediator(%v) need not to recover group-sign of the unit(%v), "+
			"or this unit has been recovered to group-sign by this mediator",
			localMed.Str(), unitHash.TerminalString())
		return
	}

	sigShareSet.lock()
	//log.Debugf("sigShareSet.lock()")
	//defer log.Debugf("sigShareSet.unlock()")
	defer sigShareSet.unlock()

	// 2. 获取阈值、mediator数量、DKG
	mp.dkgLock.RLock()
	defer mp.dkgLock.RUnlock()
	var (
		mSize     int
		threshold int
		dkgr      *dkg.DistKeyGenerator
	)

	dag := mp.dag
	header, err := dag.GetHeaderByHash(unitHash)
	if header == nil {
		err = fmt.Errorf("fail to get header by hash: %v, err: %v", unitHash.TerminalString(), err.Error())
		log.Errorf(err.Error())
		return
	}

	// 判断是否是换届前的单元
	if header.Timestamp() > mp.lastMaintenanceTime {
		mSize = dag.ActiveMediatorsCount()
		threshold = dag.ChainThreshold()
		dkgr, ok = mp.activeDKGs[localMed]
	} else {
		mSize = dag.PrecedingMediatorsCount()
		threshold = dag.PrecedingThreshold()
		dkgr, ok = mp.precedingDKGs[localMed]
	}

	if !ok {
		log.Debugf("the mediator(%v)'s dkg is not existed", localMed.Str())
		return
	}

	// 3. 判断是否达到群签名的各种条件
	count := sigShareSet.len()
	if count < threshold {
		log.Debugf("the count(%v) of sign shares of the unit(hash: %v , # %v) does not reach the threshold(%v)",
			count, unitHash.TerminalString(), header.NumberU64(), threshold)
		return
	}

	dks, err := dkgr.DistKeyShare()
	if err != nil {
		log.Debugf("the mediator(%v)'s dkg get dks err:%v", localMed.Str(), err.Error())
		return
	}

	// 4. recover群签名
	suite := mp.suite
	pubPoly := share.NewPubPoly(suite, suite.Point().Base(), dks.Commitments())
	groupSig, err := tbls.Recover(suite, pubPoly, unitHash[:], sigShareSet.popSigShares(), threshold, mSize)
	if err != nil {
		log.Debugf("the mediator(%v)'s TBLS recover the unit(%v) group-sign err:%v",
			localMed.Str(), unitHash.TerminalString(), err.Error())
		return
	}

	log.Debugf("Recovered the Unit(hash: %v , # %v )'s the Group-sign: %v",
		unitHash.TerminalString(), header.NumberU64(), hexutil.Encode(groupSig))

	// 5. recover后的相关处理
	// recover后 删除buf
	delete(mp.toTBLSRecoverBuf[localMed], unitHash)

	deadline := time.Now().Add(mp.dag.UnitIrreversibleTime())
	event := GroupSigEvent{
		UnitHash: unitHash,
		GroupSig: groupSig,
		Deadline: uint64(deadline.Unix()),
	}
	go mp.groupSigFeed.Send(event)
}

func (mp *MediatorPlugin) groupSignUnit(localMed common.Address, unitHash common.Hash) {
	if !mp.groupSigningEnabled {
		return
	}

	// 1. 初始化签名unit相关的签名分片的buf
	mp.toTBLSBufLock.Lock()
	//log.Debugf("toTBLSBufLock.Lock()")
	if _, ok := mp.toTBLSRecoverBuf[localMed]; !ok {
		mp.toTBLSRecoverBuf[localMed] = make(map[common.Hash]*sigShareSet)
	}
	aSize := mp.dag.ActiveMediatorsCount()
	mp.toTBLSRecoverBuf[localMed][unitHash] = newSigShareSet(aSize)
	//log.Debugf("toTBLSBufLock.Unlock()")
	mp.toTBLSBufLock.Unlock()

	// 2. 过了 unit 确认时间后，及时删除群签名分片的相关数据，防止内存溢出
	go func() {
		expiration := mp.dag.UnitIrreversibleTime()
		deleteBuf := time.NewTimer(expiration)

		select {
		case <-mp.quit:
			return
		case <-deleteBuf.C:
			mp.toTBLSBufLock.Lock()
			//log.Debugf("toTBLSBufLock.Lock()")
			if _, ok := mp.toTBLSRecoverBuf[localMed][unitHash]; ok {
				log.Debugf("the unit(%v) has expired confirmation time, no longer need the mediator(%v) "+
					"to recover group-sign", unitHash.TerminalString(), localMed.Str())
				delete(mp.toTBLSRecoverBuf[localMed], unitHash)
			}
			//log.Debugf("toTBLSBufLock.Unlock()")
			mp.toTBLSBufLock.Unlock()
		}
	}()
}
