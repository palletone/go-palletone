package ptn

import (
	"github.com/palletone/go-palletone/consensus/jury"
	"github.com/palletone/go-palletone/common/event"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/core/accounts/keystore"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/dag/txspool"
)

type contractInf interface {
	SubscribeContractEvent(ch chan<- jury.ContractExeEvent) event.Subscription
	ProcessContractEvent(event *jury.ContractExeEvent) error

	SubscribeContractSigEvent(ch chan<- jury.ContractSigEvent) event.Subscription
	ProcessContractSigEvent(event *jury.ContractSigEvent) error

	ProcessContractSpecialEvent(event *jury.ContractSpecialEvent) error

	AddContractLoop(txpool txspool.ITxPool, addr common.Address, ks *keystore.KeyStore) error
	CheckContractTxValid(tx *modules.Transaction) bool
}
