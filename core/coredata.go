package core

import (
	"github.com/palletone/go-palletone/common/event"
)

type ConsensusEngine interface {
	Engine() int
	Stop()
	SubscribeCeEvent(chan<- ConsensusEvent) event.Subscription
}

type ConsensusEvent struct {
	Ce string
}
