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

package ptn

import (
	"time"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/event"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/p2p/discover"
	mp "github.com/palletone/go-palletone/consensus/mediatorplugin"
)

// @author Albert·Gou
type producer interface {
	// SubscribeNewProducedUnitEvent should return an event subscription of
	// NewProducedUnitEvent and send events to the given channel.
	SubscribeNewProducedUnitEvent(ch chan<- mp.NewProducedUnitEvent) event.Subscription

	// AddToTBLSSignBufs is to TBLS sign the unit
	AddToTBLSSignBufs(newHash common.Hash)

	SubscribeSigShareEvent(ch chan<- mp.SigShareEvent) event.Subscription
	AddToTBLSRecoverBuf(newUnitHash common.Hash, sigShare []byte)

	SubscribeVSSDealEvent(ch chan<- mp.VSSDealEvent) event.Subscription
	AddToDealBuf(deal *mp.VSSDealEvent)

	SubscribeVSSResponseEvent(ch chan<- mp.VSSResponseEvent) event.Subscription
	AddToResponseBuf(resp *mp.VSSResponseEvent)

	LocalHaveActiveMediator() bool
	LocalHavePrecedingMediator() bool

	SubscribeGroupSigEvent(ch chan<- mp.GroupSigEvent) event.Subscription
	UpdateMediatorsDKG(isRenew bool)
}

func (pm *ProtocolManager) activeMediatorsUpdatedEventRecvLoop() {
	log.Debugf("activeMediatorsUpdatedEventRecvLoop")
	for {
		select {
		case event := <-pm.activeMediatorsUpdatedCh:
			go pm.switchMediatorConnect(event.IsChanged)

			// Err() channel will be closed when unsubscribing.
		case <-pm.activeMediatorsUpdatedSub.Err():
			return
		}
	}
}

func (pm *ProtocolManager) switchMediatorConnect(isChanged bool) {
	log.Debug("switchMediatorConnect", "isChanged", isChanged)

	// 若干数据还没同步完成，则忽略本次切换，继续同步
	if !pm.dag.IsSynced() {
		log.Debugf("this node is not synced")
		return
	}

	// todo albert 待优化
	//if !isChanged {
	//	go pm.producer.UpdateMediatorsDKG(false)
	//	return
	//}

	// 和新的活跃mediator节点相连
	go pm.connectWitchActiveMediators()

	// 检查是否连接和同步，并更新DKG和VSS
	go pm.checkConnectedAndSynced()

	// 延迟关闭和旧活跃mediator节点的连接
	go pm.delayDiscPrecedingMediator()
}

func (pm *ProtocolManager) connectWitchActiveMediators() {
	// 1. 判断本节点是否是活跃mediator
	log.Debugf("to connected with all active mediator nodes")
	if !pm.producer.LocalHaveActiveMediator() {
		return
	}

	// 2. 和其他活跃mediator节点相连
	peers := pm.dag.GetActiveMediatorNodes()
	for _, peer := range peers {
		// 仅当不是本节点，才做处理
		if peer.ID != pm.srvr.Self().ID {
			pm.srvr.AddTrustedPeer(peer) // 加入Trusted列表
			pm.srvr.AddPeer(peer)        // 建立连接
		}
	}
}

func (pm *ProtocolManager) checkConnectedAndSynced() {
	log.Debugf("check if it is connected to all active mediator peers")
	if !pm.producer.LocalHaveActiveMediator() {
		return
	}

	// 2. 是否和所有其他活跃mediator节点相连完成
	//headNum := pm.dag.HeadUnitNum()
	//gasToken := dagconfig.DagConfig.GetGasToken()
	checkFn := func() bool {
		nodes := pm.dag.GetActiveMediatorNodes()
		for id, node := range nodes {
			// 仅当不是本节点，并还未连接完成时，或者未同步，返回false
			if node.ID == pm.srvr.Self().ID {
				continue
			}

			peer := pm.peers.Peer(id)
			if peer == nil {
				return false
			}

			// todo Albert 待使用
			//_, pHeadNum := peer.Head(gasToken)
			//if pHeadNum == nil || pHeadNum.Index < headNum {
			//	return false
			//}
		}

		log.Debugf("connected with all active mediator peers")
		//log.Debugf("connected with all active mediator peers, all all peers synced")
		return true
	}

	// 3. 更新DKG和VSS
	processFn := func() {
		go pm.producer.UpdateMediatorsDKG(true)
	}

	// 1. 设置Ticker, 每隔一段时间检查一次
	checkTick := time.NewTicker(200 * time.Millisecond)

	defer checkTick.Stop()
	// 设置检查期限，防止死循环
	expiration := pm.dag.UnitIrreversibleTime()
	killLoop := time.NewTimer(expiration)

	for {
		select {
		case <-pm.quitSync:
			return
		case <-killLoop.C:
			return
		case <-checkTick.C:
			if checkFn() {
				processFn()
				return
			}
		}
	}
}

func (pm *ProtocolManager) delayDiscPrecedingMediator() {
	// 1. 判断当前节点是否是上一届活跃mediator
	if !pm.producer.LocalHavePrecedingMediator() {
		return
	}

	// 如果当前节点不是活跃mediator，则删除全部之前的mediator节点
	isActive := pm.producer.LocalHaveActiveMediator()

	// 2. 统计出需要断开连接的mediator节点
	delayDiscNodes := make(map[string]*discover.Node)

	activePeers := pm.dag.GetActiveMediatorNodes()
	precedingPeers := pm.dag.GetPrecedingMediatorNodes()
	for id, peer := range precedingPeers {
		// 仅当上一届mediator 不是本届活跃mediator，或者本节点不是活跃mediator
		if _, ok := activePeers[id]; !isActive || !ok /*&& pm.peers.Peer(id) != nil*/ {
			delayDiscNodes[id] = peer
		}
	}

	// 3. 设置定时器延迟 将上一届的活跃mediator节点从Trusted列表中移除
	disconnectFn := func() {
		for _, peer := range delayDiscNodes {
			pm.srvr.RemoveTrustedPeer(peer)
		}
	}

	expiration := pm.dag.UnitIrreversibleTime()
	delayDisc := time.NewTimer(expiration)

	select {
	case <-pm.quitSync:
		return
	case <-delayDisc.C:
		log.Debugf("disconnect with preceding mediator nodes")
		disconnectFn()
	}
}
