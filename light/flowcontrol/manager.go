// Copyright 2016 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

// Package flowcontrol implements a client side flow control mechanism
package flowcontrol

import (
	"sync"
	"time"

	"github.com/palletone/go-palletone/common/mclock"
)

const rcConst = 1000000

type cmNode struct {
	node                         *ClientNode
	lastUpdate                   mclock.AbsTime
	serving, recharging          bool
	rcWeight                     uint64
	rcValue, rcDelta, startValue int64
	finishRecharge               mclock.AbsTime
}

func (node *cmNode) update(time mclock.AbsTime) {
	dt := int64(time - node.lastUpdate)
	node.rcValue += node.rcDelta * dt / rcConst
	node.lastUpdate = time
	if node.recharging && time >= node.finishRecharge {
		node.recharging = false
		node.rcDelta = 0
		node.rcValue = 0
	}
}

func (node *cmNode) set(serving bool, simReqCnt, sumWeight uint64) {
	if node.serving && !serving {
		node.recharging = true
		sumWeight += node.rcWeight
	}
	node.serving = serving
	if node.recharging && serving {
		node.recharging = false
		sumWeight -= node.rcWeight
	}

	node.rcDelta = 0
	if serving {
		node.rcDelta = int64(rcConst / simReqCnt)
	}
	if node.recharging {
		node.rcDelta = -int64(node.node.cm.rcRecharge * node.rcWeight / sumWeight)
		node.finishRecharge = node.lastUpdate + mclock.AbsTime(node.rcValue*rcConst/(-node.rcDelta))
	}
}

type ClientManager struct {
	lock                             sync.Mutex
	nodes                            map[*cmNode]struct{}
	simReqCnt, sumWeight, rcSumValue uint64
	maxSimReq, maxRcSum              uint64
	rcRecharge                       uint64
	resumeQueue                      chan chan bool
	time                             mclock.AbsTime
}

func NewClientManager(rcTarget, maxSimReq, maxRcSum uint64) *ClientManager {
	cm := &ClientManager{
		nodes:       make(map[*cmNode]struct{}),
		resumeQueue: make(chan chan bool),
		rcRecharge:  rcConst * rcConst / (100*rcConst/rcTarget - rcConst),
		maxSimReq:   maxSimReq,
		maxRcSum:    maxRcSum,
	}
	go cm.queueProc()
	return cm
}

func (pm *ClientManager) Stop() {
	pm.lock.Lock()
	defer pm.lock.Unlock()

	// signal any waiting accept routines to return false
	pm.nodes = make(map[*cmNode]struct{})
	close(pm.resumeQueue)
}

func (pm *ClientManager) addNode(cnode *ClientNode) *cmNode {
	time := mclock.Now()
	node := &cmNode{
		node:           cnode,
		lastUpdate:     time,
		finishRecharge: time,
		rcWeight:       1,
	}
	pm.lock.Lock()
	defer pm.lock.Unlock()

	pm.nodes[node] = struct{}{}
	pm.update(mclock.Now())
	return node
}

func (pm *ClientManager) removeNode(node *cmNode) {
	pm.lock.Lock()
	defer pm.lock.Unlock()

	time := mclock.Now()
	pm.stop(node, time)
	delete(pm.nodes, node)
	pm.update(time)
}

// recalc sumWeight
func (pm *ClientManager) updateNodes(time mclock.AbsTime) (rce bool) {
	var sumWeight, rcSum uint64
	for node := range pm.nodes {
		rc := node.recharging
		node.update(time)
		if rc && !node.recharging {
			rce = true
		}
		if node.recharging {
			sumWeight += node.rcWeight
		}
		rcSum += uint64(node.rcValue)
	}
	pm.sumWeight = sumWeight
	pm.rcSumValue = rcSum
	return
}

func (pm *ClientManager) update(time mclock.AbsTime) {
	for {
		firstTime := time
		for node := range pm.nodes {
			if node.recharging && node.finishRecharge < firstTime {
				firstTime = node.finishRecharge
			}
		}
		if pm.updateNodes(firstTime) {
			for node := range pm.nodes {
				if node.recharging {
					node.set(node.serving, pm.simReqCnt, pm.sumWeight)
				}
			}
		} else {
			pm.time = time
			return
		}
	}
}

func (pm *ClientManager) canStartReq() bool {
	return pm.simReqCnt < pm.maxSimReq && pm.rcSumValue < pm.maxRcSum
}

func (pm *ClientManager) queueProc() {
	for rc := range pm.resumeQueue {
		for {
			time.Sleep(time.Millisecond * 10)
			pm.lock.Lock()
			pm.update(mclock.Now())
			cs := pm.canStartReq()
			pm.lock.Unlock()
			if cs {
				break
			}
		}
		close(rc)
	}
}

func (pm *ClientManager) accept(node *cmNode, time mclock.AbsTime) bool {
	pm.lock.Lock()
	defer pm.lock.Unlock()

	pm.update(time)
	if !pm.canStartReq() {
		resume := make(chan bool)
		pm.lock.Unlock()
		pm.resumeQueue <- resume
		<-resume
		pm.lock.Lock()
		if _, ok := pm.nodes[node]; !ok {
			return false // reject if node has been removed or manager has been stopped
		}
	}
	pm.simReqCnt++
	node.set(true, pm.simReqCnt, pm.sumWeight)
	node.startValue = node.rcValue
	pm.update(pm.time)
	return true
}

func (pm *ClientManager) stop(node *cmNode, time mclock.AbsTime) {
	if node.serving {
		pm.update(time)
		pm.simReqCnt--
		node.set(false, pm.simReqCnt, pm.sumWeight)
		pm.update(time)
	}
}

func (pm *ClientManager) processed(node *cmNode, time mclock.AbsTime) (rcValue, rcCost uint64) {
	pm.lock.Lock()
	defer pm.lock.Unlock()

	pm.stop(node, time)
	return uint64(node.rcValue), uint64(node.rcValue - node.startValue)
}
