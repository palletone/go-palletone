package peer

import (
	"github.com/palletone/go-palletone/dag"
	"github.com/palletone/go-palletone/dag/modules"
	"golang.org/x/net/context"
	"time"
)

type EndorserServer interface {
	ProcessProposal(dag.IDag, []byte, context.Context, *SignedProposal, *Proposal, string, *ChaincodeID, time.Duration) (*ProposalResponse, *modules.ContractInvokeResult, error)
}
