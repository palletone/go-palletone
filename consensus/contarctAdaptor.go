package consensus

import (
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/consensus/jury"
	"time"
)

type AdapterJury struct {
	Processor *jury.Processor
}

func (a *AdapterJury) AdapterFunRequest(reqId common.Hash, contractId common.Address, timeOut time.Duration) (interface{}, error) {
	return a.Processor.AdapterFunRequest(reqId, contractId, timeOut)
}
