package peer

import (
	"time"

	"github.com/palletone/go-palletone/dag/dboperation"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/rwset"
	"golang.org/x/net/context"
)

type EndorserServer interface {
	ProcessProposal(rwset.TxManager, dboperation.IContractDag, []byte, context.Context, *SignedProposal, *Proposal, string, *ChaincodeID, time.Duration) (*ProposalResponse, *modules.ContractInvokeResult, error)
}
