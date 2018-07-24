package peer

import "golang.org/x/net/context"

type EndorserServer interface {
	ProcessProposal(context.Context, *SignedProposal, *Proposal, string, *ChaincodeID) (*ProposalResponse, *ContractInvokePayload, error)
}
