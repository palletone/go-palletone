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
	"sync"
	"time"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/event"
	"github.com/palletone/go-palletone/common/hexutil"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/dag/modules"
	"go.dedis.ch/kyber/v3/share"
	"go.dedis.ch/kyber/v3/share/dkg/pedersen"
	"go.dedis.ch/kyber/v3/sign/tbls"
)

func (mp *MediatorPlugin) signUnitsTBLS(localMed common.Address) {
	medUnitsBuf, ok := mp.toTBLSSignBuf[localMed]
	if !ok {
		log.Debugf("the mediator(%v) has no units to sign TBLS yet", localMed.Str())
		return
	}

	rangeFn := func(key, value interface{}) bool {
		newUnitHash, ok := key.(common.Hash)
		if !ok {
			log.Debugf("key converted to Hash failed")
			return true
		}
		go mp.signUnitTBLS(localMed, newUnitHash)

		return true
	}

	medUnitsBuf.Range(rangeFn)
}

func (mp *MediatorPlugin) recoverUnitsTBLS(localMed common.Address) {
	mp.recoverBufLock.RLock()
	defer mp.recoverBufLock.RUnlock()

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

	newHeader, err := mp.dag.GetHeaderByHash(newHash)
	if newHeader == nil || err != nil {
		err = fmt.Errorf("fail to get header by hash in dag: %v", newHash.TerminalString())
		log.Debugf(err.Error())
		return
	}

	var ms []common.Address
	// 严格要求换届unix时间是产块间隔的整数倍
	if newHeader.Timestamp() > mp.dag.LastMaintenanceTime() {
		ms = mp.GetLocalActiveMediators()
	} else {
		ms = mp.GetLocalPrecedingMediators()
	}

	for _, localMed := range ms {
		log.Debugf("the mediator(%v) received a unit(%v) to be group-signed",
			localMed.Str(), newHash.TerminalString())
		go mp.addToTBLSSignBuf(localMed, newHash)
	}
}

func (mp *MediatorPlugin) addToTBLSSignBuf(localMed common.Address, unitHash common.Hash) {
	if _, ok := mp.toTBLSSignBuf[localMed]; !ok {
		mp.toTBLSSignBuf[localMed] = new(sync.Map)
	}

	mp.toTBLSSignBuf[localMed].LoadOrStore(unitHash, true)
	go mp.signUnitTBLS(localMed, unitHash)

	// 当 unit 过了确认时间后，及时删除待群签名的 unit，防止内存溢出
	expiration := mp.dag.UnitIrreversibleTime()
	deleteBuf := time.NewTimer(expiration)

	select {
	case <-mp.quit:
		return
	case <-deleteBuf.C:
		if _, ok := mp.toTBLSSignBuf[localMed].Load(unitHash); ok {
			log.Debugf("the unit(%v) has expired confirmation time, no longer need the mediator(%v)"+
				" to sign-group", unitHash.TerminalString(), localMed.Str())
			mp.toTBLSSignBuf[localMed].Delete(unitHash)
		}
	}
}

func (mp *MediatorPlugin) SubscribeSigShareEvent(ch chan<- SigShareEvent) event.Subscription {
	return mp.sigShareScope.Track(mp.sigShareFeed.Subscribe(ch))
}

func (mp *MediatorPlugin) signUnitTBLS(localMed common.Address, unitHash common.Hash) {
	medUnitsBuf, ok := mp.toTBLSSignBuf[localMed]
	if !ok {
		log.Debugf("the mediator(%v) has no units to sign TBLS yet", localMed.Str())
		return
	}

	dag := mp.dag
	mp.dkgLock.Lock()
	defer mp.dkgLock.Unlock()
	var (
		dkgr      *dkg.DistKeyGenerator
		newHeader *modules.Header
		err       error
	)

	// 1. 获取群签名所需数据
	{
		_, ok := medUnitsBuf.Load(unitHash)
		if !ok {
			log.Debugf("the mediator(%v) has no unit(%v) to sign TBLS",
				localMed.Str(), unitHash.TerminalString())
			return
		}

		newHeader, err = mp.dag.GetHeaderByHash(unitHash)
		if newHeader == nil || err != nil {
			err = fmt.Errorf("fail to get header by hash in dag: %v", unitHash.TerminalString())
			log.Debugf(err.Error())
			return
		}

		// 判断是否是换届前的单元
		if newHeader.Timestamp() > mp.lastMaintenanceTime {
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
		pkb := newHeader.GetGroupPubKeyByte()
		if len(pkb) == 0 {
			err := fmt.Errorf("this unit(%v)'s group public key is null", unitHash.TerminalString())
			log.Debug(err.Error())
			return
		}

		// 判断父 unit 是否不可逆
		parentHash := newHeader.ParentHash()[0]
		if !dag.IsIrreversibleUnit(parentHash) {
			log.Debugf("the unit's(%v) parent unit(%v) is not irreversible",
				unitHash.TerminalString(), parentHash.TerminalString())
			return
		}
	}

	// 3. 群签名
	dks, err := dkgr.DistKeyShare()
	if err != nil {
		log.Debugf(err.Error())
		return
	}

	sigShare, err := tbls.Sign(mp.suite, dks.PriShare(), unitHash[:])
	if err != nil {
		log.Debugf(err.Error())
		return
	}

	// 4. 群签名成功后的处理
	log.Debugf("the mediator(%v) signed-group the unit(%v)", localMed.Str(),
		unitHash.TerminalString())
	mp.toTBLSSignBuf[localMed].Delete(unitHash)
	go mp.sigShareFeed.Send(SigShareEvent{UnitHash: unitHash, SigShare: sigShare})
}

// 收集签名分片
func (mp *MediatorPlugin) AddToTBLSRecoverBuf(newUnitHash common.Hash, sigShare []byte) {
	if !mp.groupSigningEnabled {
		return
	}

	log.Debugf("received the sign shares of the unit(%v)", newUnitHash.TerminalString())

	dag := mp.dag
	newUnit, err := dag.GetHeaderByHash(newUnitHash)
	if newUnit == nil || err != nil {
		err = fmt.Errorf("fail to get unit by hash in dag: %v", newUnitHash.TerminalString())
		log.Debugf(err.Error())
		return
	}

	localMed := newUnit.Author()
	mp.recoverBufLock.RLock()
	defer mp.recoverBufLock.RUnlock()

	medSigSharesBuf, ok := mp.toTBLSRecoverBuf[localMed]
	if !ok {
		err = fmt.Errorf("the mediator(%v)'s toTBLSRecoverBuf has not initialized yet", localMed.Str())
		log.Debugf(err.Error())
		return
	}

	// 当buf不存在时，说明已经recover出群签名, 或者已经过了unit确认时间，忽略该签名分片
	sigShareSet, ok := medSigSharesBuf[newUnitHash]
	if !ok {
		err = fmt.Errorf("the unit(%v) has already recovered the group signature",
			newUnitHash.TerminalString())
		log.Debugf(err.Error())
		return
	}

	sigShareSet.append(sigShare)

	// recover群签名
	go mp.recoverUnitTBLS(localMed, newUnitHash)
}

func (mp *MediatorPlugin) SubscribeGroupSigEvent(ch chan<- GroupSigEvent) event.Subscription {
	return mp.groupSigScope.Track(mp.groupSigFeed.Subscribe(ch))
}

func (mp *MediatorPlugin) recoverUnitTBLS(localMed common.Address, unitHash common.Hash) {
	mp.recoverBufLock.Lock()
	defer mp.recoverBufLock.Unlock()

	// 1. 获取所有的签名分片
	sigSharesBuf, ok := mp.toTBLSRecoverBuf[localMed]
	if !ok {
		log.Debugf("the mediator((%v) has no signature shares to recover group sign yet", localMed.Str())
		return
	}

	sigShareSet, ok := sigSharesBuf[unitHash]
	if !ok {
		log.Debugf("the mediator(%v) has no sign shares corresponding unit(%v) yet",
			localMed.Str(), unitHash.TerminalString())
		return
	}

	sigShareSet.lock()
	defer sigShareSet.unlock()

	// 为了保证多协程安全， 加锁后，再判断一次
	if _, ok = mp.toTBLSRecoverBuf[localMed][unitHash]; !ok {
		return
	}

	// 2. 获取阈值、mediator数量、DKG
	mp.dkgLock.Lock()
	defer mp.dkgLock.Unlock()
	var (
		mSize, threshold int
		dkgr             *dkg.DistKeyGenerator
	)

	{
		dag := mp.dag
		unit, err := dag.GetHeaderByHash(unitHash)
		if unit == nil || err != nil {
			err = fmt.Errorf("fail to get unit by hash in dag: %v", unitHash.TerminalString())
			log.Debugf(err.Error())
			return
		}

		// 判断是否是换届前的单元
		if unit.Timestamp() > mp.lastMaintenanceTime {
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
	}

	// 3. 判断是否达到群签名的各种条件
	if sigShareSet.len() < threshold {
		log.Debugf("the count of sign shares of the unit(%v) does not reach the threshold(%v)",
			unitHash.TerminalString(), threshold)
		return
	}

	dks, err := dkgr.DistKeyShare()
	if err != nil {
		log.Debugf(err.Error())
		return
	}

	// 4. recover群签名
	suite := mp.suite
	pubPoly := share.NewPubPoly(suite, suite.Point().Base(), dks.Commitments())
	groupSig, err := tbls.Recover(suite, pubPoly, unitHash[:], sigShareSet.popSigShares(), threshold, mSize)
	if err != nil {
		log.Debugf(err.Error())
		return
	}

	log.Debugf("Recovered the Unit(%v)'s the Group-sign: %v",
		unitHash.TerminalString(), hexutil.Encode(groupSig))

	// 5. recover后的相关处理
	// recover后 删除buf
	delete(mp.toTBLSRecoverBuf[localMed], unitHash)
	go mp.dag.SetUnitGroupSign(unitHash, groupSig, mp.ptn.TxPool())
	go mp.groupSigFeed.Send(GroupSigEvent{UnitHash: unitHash, GroupSig: groupSig})
}
