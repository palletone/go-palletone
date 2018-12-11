package peer

import (
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag"
	"golang.org/x/net/context"
	"time"
)

type EndorserServer interface {
	ProcessProposal([]byte, dag.IDag, []byte, context.Context, *SignedProposal, *Proposal, string, *ChaincodeID, time.Duration) (*ProposalResponse, *modules.ContractInvokeResult, error)
}
