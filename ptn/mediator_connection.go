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

import "time"

func (pm *ProtocolManager) chainMaintainEventRecvLoop() {
	for {
		select {
		case <-pm.chainMaintainCh:
			go pm.switchMediatorConnect()

			// Err() channel will be closed when unsubscribing.
		case <-pm.chainMaintainSub.Err():
			return
		}
	}
}

func (pm *ProtocolManager) switchMediatorConnect() {
	// 1. 若干数据还没同步完成，则忽略本次切换，继续同步
	if !pm.dag.IsSynced() {
		return
	}

	// 2. 和新的活跃mediator节点相连
	go pm.connectWitchActiveMediators()

	// 3. 检查是否连接完成并发送事件
	go pm.checkActiveMediatorConnection()

	// 4. 延迟关闭和旧活跃mediator节点的连接
	go pm.delayDiscPrecedingMediator()
}

func (pm *ProtocolManager) connectWitchActiveMediators() {
	// 1. 判断本节点是否是活跃mediator
	if !pm.producer.LocalHaveActiveMediator() {
		return
	}

	// 2. 和其他活跃mediator节点相连
	peers := pm.dag.GetActiveMediatorNodes()
	for id, peer := range peers {
		// 仅当不是本节点，并还未连接时，才进行连接
		if peer.ID != pm.srvr.Self().ID && pm.peers.Peer(id) == nil {
			pm.srvr.AddTrustedPeer(peer)
		}
	}
}

func (pm *ProtocolManager) checkActiveMediatorConnection() {
	// 2. 是否和所有其他活跃mediator节点相连完成
	checkFn := func() bool {
		peers := pm.dag.GetActiveMediatorNodes()
		for id, peer := range peers {
			// 仅当不是本节点，并还未连接完成时，返回false
			if peer.ID != pm.srvr.Self().ID && pm.peers.Peer(id) == nil {
				return false
			}
		}

		return true
	}

	// 3. 发送mediator连接完成事件
	sendEventFn := func() {

	}

	// 1. 设置Ticker, 每隔一段时间检查一次
	checkTick := time.NewTicker(500 * time.Millisecond)
	for {
		select {
		case <-pm.quitSync:
			return
		case <-checkTick.C:
			if checkFn() {
				checkTick.Stop()
				sendEventFn()
			}
		}
	}
}

func (pm *ProtocolManager) delayDiscPrecedingMediator() {
	//todo
}
