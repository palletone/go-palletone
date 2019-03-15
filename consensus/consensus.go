// This file is part of go-palletone
// go-palletone is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-palletone is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with go-palletone. If not, see <http://www.gnu.org/licenses/>.
//
// @author PalletOne DevTeam dev@pallet.one
// @date 2018

// Package consensus implements different PalletOne consensus engines.
package consensus

import (
	//"fmt"
	//"sync"

	"encoding/json"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/event"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/dag"
	"github.com/palletone/go-palletone/dag/txspool"
	"time"
)

// Engine is an algorithm agnostic consensus engine.
type Engine interface {
}

type DPOSEngine struct {
	scope    event.SubscriptionScope
	dposFeed event.Feed
	dag      dag.IDag
	txpool   txspool.ITxPool
}

func (engine *DPOSEngine) SubscribeCeEvent(ch chan<- core.ConsensusEvent) event.Subscription {
	return engine.scope.Track(engine.dposFeed.Subscribe(ch))
}

func (engine *DPOSEngine) SendEvents(content []byte) {
	engine.dposFeed.Send(core.ConsensusEvent{content})
}

func (engine *DPOSEngine) Stop() {
	// Unsubscribe all subscriptions registered from dops
	engine.scope.Close()
	log.Info("DPOSEngine stopped")
}

func (engine *DPOSEngine) Engine() int {
	return 0
	address, err := common.StringToAddress("P19QMdx59PDYRxJpR2T9c2r5F5VhxxnkoRe")
	if err != nil {
		log.Debug("Test P2P", "DPOSEngine->Engine err", err)
		return -1
	}
	when := time.Time{}

	newUnits, err1 := engine.dag.CreateUnit(&address, engine.txpool, when)
	if err1 != nil {
		log.Debug("Test P2P", "DPOSEngine->Engine CreateUnit err", err1)
		return -2
	}
	data, err2 := json.Marshal(newUnits[0])
	if err2 != nil {
		log.Debug("Test P2P", "DPOSEngine->Engine CreateUnit json marshal err:", err2)
		return -3
	}
	log.Debug("Test P2P", "DPOSEngine->Engine SendEvents data", string(data))
	engine.SendEvents(data)
	return 0
}
func New(dag dag.IDag, txpool txspool.ITxPool) *DPOSEngine {
	return &DPOSEngine{
		dag:    dag,
		txpool: txpool,
	}
}

//var engine ConsensusEngine = DPOSEngine{}
