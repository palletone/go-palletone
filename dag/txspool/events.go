package txspool

import (
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/dag/modules"
)

// TxPreEvent is posted when a transaction enters the transaction pool.
type TxPreEvent struct{ Tx *modules.Transaction }

// PendingLogsEvent is posted pre mining and notifies of pending logs.
type PendingLogsEvent struct {
	Logs []*modules.Log
}

// RemovedTransactionEvent is posted when a reorg happens
type RemovedTransactionEvent struct{ Txs modules.Transactions }

// RemovedLogsEvent is posted when a reorg happens
type RemovedLogsEvent struct{ Logs []*modules.Log }

type ChainEvent struct {
	Unit *modules.Unit
	Hash common.Hash
	Logs []*modules.Log
}
type ChainSideEvent struct {
	Unit *modules.Unit
}

type ChainHeadEvent struct{ Unit *modules.Unit }
