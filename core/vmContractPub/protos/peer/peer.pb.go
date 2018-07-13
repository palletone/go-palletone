package peer

import "golang.org/x/net/context"

type EndorserServer interface {
	ProcessProposal(context.Context, *SignedProposal, *Proposal, string, string, *ChaincodeID) (*ProposalResponse, error)
}
