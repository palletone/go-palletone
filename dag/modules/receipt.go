// Copyright 2018 PalletOne

package modules

import (
	"github.com/palletone/go-palletone/common"
)

// Receipt represents the results of a transaction.
type Receipt struct {
	// Consensus fields
	PostState []byte `json:"root"`
	Status    uint   `json:"status"`
	// CumulativeGasUsed uint64 `json:"cumulativeGasUsed" gencodec:"required"`
	// Bloom             Bloom  `json:"logsBloom"         gencodec:"required"`
	Logs []*Log `json:"logs"              gencodec:"required"`

	// Implementation fields (don't reorder!)
	TxHash          common.Hash    `json:"transactionHash" gencodec:"required"`
	ContractAddress common.Address `json:"contractAddress"`
	// GasUsed         uint64         `json:"gasUsed" gencodec:"required"`
}
