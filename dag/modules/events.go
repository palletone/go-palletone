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
	//Logs []*Log
}
type ChainSideEvent struct {
	Unit *Unit
}

type ChainHeadEvent struct{ Unit *Unit }

// 活跃 mediators 更新事件
type ActiveMediatorsUpdatedEvent struct {
	IsChanged bool // 标记活跃 mediators 是否有改变
}

//系统合约被调用，导致状态数据库改变
type SysContractStateChangeEvent struct {
	ContractId []byte
	WriteSet   []ContractWriteSet
}

type ChainMaintenanceEvent struct {
}

type ToGroupSignEvent struct {
}
