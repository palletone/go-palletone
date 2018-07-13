/*
	This file is part of go-palletone.
	go-palletone is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.
	go-palletone is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.
	You should have received a copy of the GNU General Public License
	along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
*/
/*
 * @author PalletOne core developers <dev@pallet.one>
 * @date 2018
 */

package manger

import (

	chaincode "github.com/palletone/go-palletone/contracts/core"
	"github.com/palletone/go-palletone/core/vmContractPub/ccprovider"
	"github.com/palletone/go-palletone/contracts/scc"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"golang.org/x/net/context"
	"github.com/golang/protobuf/proto"
	"github.com/palletone/go-palletone/contracts/rwset"
	ut "github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/common/log"
)

// SupportImpl provides an implementation of the endorser.Support interface
// issuing calls to various static methods of the peer
type SupportImpl struct{
}

// IsSysCCAndNotInvokableExternal returns true if the supplied chaincode is
// ia system chaincode and it NOT invokable
func (s *SupportImpl) IsSysCCAndNotInvokableExternal(name string) bool {
	return scc.IsSysCCAndNotInvokableExternal(name)
}

// GetTxSimulator returns the transaction simulator for the specified ledger
// a client may obtain more than one such simulator; they are made unique
// by way of the supplied txid
func (s *SupportImpl) GetTxSimulator(chainid string, txid string) (rwset.TxSimulator, error) {
	return rwM.NewTxSimulator(chainid, txid)
}

// GetTransactionByID retrieves a transaction by id
//func (s *SupportImpl) GetTransactionByID(chid, txID string) (*pb.ProcessedTransaction, error) {
//	lgr := peer.GetLedger(chid)
//	if lgr == nil {
//		return nil, errors.Errorf("failed to look up the ledger for channel %s", chid)
//	}
//	tx, err := lgr.GetTransactionByID(txID)
//	if err != nil {
//		return nil, errors.WithMessage(err, "GetTransactionByID failed")
//	}
//	return tx, nil
//}

//IsSysCC returns true if the name matches a system chaincode's
//system chaincode names are system, chain wide
func (s *SupportImpl) IsSysCC(name string) bool {
	return scc.IsSysCC(name)
}

//Execute - execute proposal, return original response of chaincode
func (s *SupportImpl) Execute(ctxt context.Context, cid, name, version, txid string, syscc bool, signedProp *pb.SignedProposal, prop *pb.Proposal, spec interface{}) (*pb.Response, *pb.ChaincodeEvent, error) {
	cccid := ccprovider.NewCCContext(cid, name, version, txid, syscc, signedProp, prop)

	switch spec.(type) {
	case *pb.ChaincodeDeploymentSpec:
		return chaincode.Execute(ctxt, cccid, spec)
	case *pb.ChaincodeInvocationSpec:
		cis := spec.(*pb.ChaincodeInvocationSpec)

		log.Info("===cis:%v", cis)
		//decorate the chaincode input
		//glh
		//decorators := library.InitRegistry(library.Config{}).Lookup(library.Decoration).([]decoration.Decorator)
		//cis.ChaincodeSpec.Input.Decorations = make(map[string][]byte)
		//cis.ChaincodeSpec.Input = decoration.Apply(prop, cis.ChaincodeSpec.Input, decorators...)
		//cccid.ProposalDecorations = cis.ChaincodeSpec.Input.Decorations
		return chaincode.ExecuteChaincode(ctxt, cccid, cis.ChaincodeSpec.Input.Args)
	default:
		panic("programming error, unkwnown spec type")
	}
}

// GetChaincodeDefinition returns resourcesconfig.ChaincodeDefinition for the chaincode with the supplied name
//func (s *SupportImpl) GetChaincodeDefinition(ctx context.Context, chainID string, txid string, signedProp *pb.SignedProposal, prop *pb.Proposal, chaincodeID string, txsim rwset.TxSimulator) (resourcesconfig.ChaincodeDefinition, error) {
//	ctxt := ctx
//	if txsim != nil {
//		ctxt = context.WithValue(ctx, chaincode.TXSimulatorKey, txsim)
//	}
//	return chaincode.GetChaincodeDefinition(ctxt, txid, signedProp, prop, chainID, chaincodeID)
//}

//CheckACL checks the ACL for the resource for the channel using the
//SignedProposal from which an id can be extracted for testing against a policy
//func (s *SupportImpl) CheckACL(signedProp *pb.SignedProposal, chdr *common.ChannelHeader, shdr *common.SignatureHeader, hdrext *pb.ChaincodeHeaderExtension) error {
//	return aclmgmt.GetACLProvider().CheckACL(resources.PROPOSE, chdr.ChannelId, signedProp)
//}

// CheckInstantiationPolicy returns an error if the instantiation in the supplied
// ChaincodeDefinition differs from the instantiation policy stored on the ledger
//func (s *SupportImpl) CheckInstantiationPolicy(name, version string, cd resourcesconfig.ChaincodeDefinition) error {
//	return ccprovider.CheckInstantiationPolicy(name, version, cd.(*ccprovider.ChaincodeData))
//}

// shorttxid replicates the chaincode package function to shorten txids.
func shorttxid(txid string) string {
	if len(txid) < 8 {
		return txid
	}
	return txid[0:8]
}

func GetBytesChaincodeEvent(event *pb.ChaincodeEvent) ([]byte, error) {
	eventBytes, err := proto.Marshal(event)
	return eventBytes, err
}

func converRwTxResult2DagUnit(ts rwset.TxSimulator) (*ut.ContractInvokePayload, error) {
	logger.Info("enter")
	return nil, nil
}

var rwM *rwset.RwSetTxMgr

func init() {
	var err error
	rwM, err =rwset.NewRwSetMgr("default")
	if err != nil {
		logger.Error("fail!")
	}
}


