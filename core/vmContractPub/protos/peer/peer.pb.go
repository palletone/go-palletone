package peer

import (
	"github.com/palletone/go-palletone/dag"
	unit "github.com/palletone/go-palletone/dag/modules"
	"golang.org/x/net/context"
	"time"
)

type EndorserServer interface {
	ProcessProposal([]byte, dag.IDag, []byte, context.Context, *SignedProposal, *Proposal, string, *ChaincodeID, time.Duration) (*ProposalResponse, *unit.ContractInvokePayload, error)
}
