package consensus

import (
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/consensus/jury"
	"time"
)

type AdapterJury struct {
	Processor *jury.Processor
}

func (a *AdapterJury) AdapterFunRequest(reqId common.Hash, contractId common.Address, msgType uint32,
	consultContent []byte, myAnswer []byte) ([]byte, error) {
	return a.Processor.AdapterFunRequest(reqId, contractId, msgType, consultContent, myAnswer)
}
func (a *AdapterJury) AdapterFunResult(reqId common.Hash, contractId common.Address, msgType uint32,
	consultContent []byte, timeOut time.Duration) ([]byte, error) {
	return a.Processor.AdapterFunResult(reqId, contractId, msgType, consultContent, timeOut)
}
