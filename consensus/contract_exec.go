package consensus

import (
	"github.com/ethereum/go-ethereum/event"
	"github.com/palletone/go-palletone/consensus/jury"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/rwset"
)

type ContractInf interface {
	SubscribeContractEvent(ch chan<- jury.ContractEvent) event.Subscription

	ProcessUserContractInvokeReqTx(tx *modules.Transaction)
	ProcessContractEvent(event *jury.ContractEvent) (broadcast bool, err error)
	ProcessElectionEvent(event *jury.ElectionEvent) (err error)
	ProcessAdapterEvent(event *jury.AdapterEvent) (result *jury.AdapterEvent, err error)

	CheckContractTxValid(rwM rwset.TxManager, tx *modules.Transaction, execute bool) bool
	AddLocalTx(tx *modules.Transaction) error
}
