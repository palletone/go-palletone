package peer

import (
	"golang.org/x/net/context"
	"time"
)

type EndorserServer interface {
	ProcessProposal(context.Context, *SignedProposal, *Proposal, string, *ChaincodeID, time.Duration) (*ProposalResponse, *ContractInvokePayload, error)
}
