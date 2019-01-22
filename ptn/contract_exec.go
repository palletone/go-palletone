package ptn

import (
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/event"
	"github.com/palletone/go-palletone/consensus/jury"
	"github.com/palletone/go-palletone/core/accounts/keystore"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/txspool"
)

type contractInf interface {
	SubscribeContractEvent(ch chan<- jury.ContractEvent) event.Subscription
	ProcessContractEvent(event *jury.ContractEvent) error

	AddContractLoop(txpool txspool.ITxPool, addr common.Address, ks *keystore.KeyStore) error
	CheckContractTxValid(tx *modules.Transaction, execute bool) bool
}
