package core

import (
	"github.com/palletone/go-palletone/common/event"
)

type ConsensusEngine interface {
	Engine() int
	Stop()
	SubscribeCeEvent(chan<- ConsensusEvent) event.Subscription
	//SubscribeTxPreEvent(chan<- core.TxPreEvent) event.Subscription
}

//type TxPreEvent struct{ Tx *types.Transaction }
type ConsensusEvent struct {
	Ce string
}
