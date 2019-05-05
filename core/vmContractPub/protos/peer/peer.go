package peer

import (
	"time"

	"github.com/palletone/go-palletone/dag"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/rwset"
	"golang.org/x/net/context"
)

type EndorserServer interface {
	ProcessProposal(rwset.TxManager, dag.IDag, []byte, context.Context, *SignedProposal, *Proposal, string, *ChaincodeID, time.Duration) (*ProposalResponse, *modules.ContractInvokeResult, error)
}
