package peer

import (
	"golang.org/x/net/context"
	"time"
	unit "github.com/palletone/go-palletone/dag/modules"
)

type EndorserServer interface {
	ProcessProposal([]byte, context.Context, *SignedProposal, *Proposal, string, *ChaincodeID, time.Duration) (*ProposalResponse, *unit.ContractInvokePayload, error)
}
