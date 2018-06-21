// Copyright 2017 The go-palletone Authors
// This file is part of the go-palletone library.
//
// The go-palletone library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-palletone library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-palletone library. If not, see <http://www.gnu.org/licenses/>.

// Package consensus implements different Ethereum consensus engines.
package consensus

import (
	//"fmt"
	//"sync"

	"github.com/palletone/go-palletone/common/event"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/core"
)

// Engine is an algorithm agnostic consensus engine.
type Engine interface {
}

type Mode uint

const (
	ModeNormal Mode = iota
	ModeShared
	ModeTest
	ModeFake
	ModeFullFake
)

type Config struct {
	CacheDir       string
	CachesInMem    int
	CachesOnDisk   int
	DatasetDir     string
	DatasetsInMem  int
	DatasetsOnDisk int
	PowMode        Mode
}

type DPOSEngine struct {
	config   Config
	scope    event.SubscriptionScope
	dposFeed event.Feed
}

func (engine *DPOSEngine) SubscribeCeEvent(ch chan<- core.ConsensusEvent) event.Subscription {
	return engine.scope.Track(engine.dposFeed.Subscribe(ch))
}

func (engine *DPOSEngine) SendEvents(content string) {
	engine.dposFeed.Send(core.ConsensusEvent{content})
}

func (engine *DPOSEngine) Stop() {
	// Unsubscribe all subscriptions registered from dops
	engine.scope.Close()
	log.Info("DPOSEngine stopped")
}

func (engine *DPOSEngine) Engine() int {
	engine.SendEvents("A")
	return 0
}
func New() *DPOSEngine {
	return &DPOSEngine{}
}

//var engine ConsensusEngine = DPOSEngine{}
