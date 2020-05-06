package peer

import (
	"time"

	"github.com/palletone/go-palletone/dag/dboperation"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/rwset"
	"golang.org/x/net/context"
)

type EndorserServer interface {
	ProcessProposal(rwset.TxManager, dboperation.IContractDag, []byte, context.Context, *PtnSignedProposal, *PtnProposal, string, *PtnChaincodeID, time.Duration) (*PtnProposalResponse, *modules.ContractInvokeResult, error)
}
