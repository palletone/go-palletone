package modules

import (
	"github.com/palletone/go-palletone/common"
)

// TxPreEvent is posted when a transaction enters the transaction pool.
type TxPreEvent struct{ Tx *Transaction }

// PendingLogsEvent is posted pre mining and notifies of pending logs.
type PendingLogsEvent struct {
	Logs []*Log
}

// RemovedTransactionEvent is posted when a reorg happens
type RemovedTransactionEvent struct{ Txs Transactions }

// RemovedLogsEvent is posted when a reorg happens
type RemovedLogsEvent struct{ Logs []*Log }

type ChainEvent struct {
	Unit *Unit
	Hash common.Hash
	Logs []*Log
}
type ChainSideEvent struct {
	Unit *Unit
}

type ChainHeadEvent struct{ Unit *Unit }
