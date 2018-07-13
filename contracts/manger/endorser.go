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
	"fmt"

	"github.com/palletone/go-palletone/core/vmContractPub/util"
	"github.com/palletone/go-palletone/core/vmContractPub/flogging"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	putils "github.com/palletone/go-palletone/core/vmContractPub/protos/utils"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
	"github.com/palletone/go-palletone/contracts/shim"
	"github.com/palletone/go-palletone/contracts/rwset"
	"github.com/palletone/go-palletone/contracts/core"
)

type chaincodeError struct {
	status int32
	msg    string
}

func (ce chaincodeError) Error() string {
	return fmt.Sprintf("chaincode error (status: %d, message: %s)", ce.status, ce.msg)
}

var logger = flogging.MustGetLogger("ccmanger")

// The Jira issue that documents Endorser flow along with its relationship to
// the lifecycle chaincode - https://jira.hyperledger.org/browse/FAB-181
//type privateDataDistributor func(channel string, txID string, privateData *rwset.TxPvtReadWriteSet) error

// Support contains functions that the endorser requires to execute its tasks
type Support interface {
	IsSysCCAndNotInvokableExternal(name string) bool
// GetTxSimulator returns the transaction simulator for the specified ledger
	// a client may obtain more than one such simulator; they are made unique
	// by way of the supplied txid
	GetTxSimulator(chainid string, txid string) (rwset.TxSimulator, error)

	// GetTransactionByID retrieves a transaction by id
	//GetTransactionByID(chid, txID string) (*pb.ProcessedTransaction, error)
	IsSysCC(name string) bool

	Execute(ctxt context.Context, cid, name, version, txid string, syscc bool, signedProp *pb.SignedProposal, prop *pb.Proposal, spec interface{}) (*pb.Response, *pb.ChaincodeEvent, error)
	//GetChaincodeDefinition(ctx context.Context, chainID string, txid string, signedProp *pb.SignedProposal, prop *pb.Proposal, chaincodeID string, txsim ledger.TxSimulator) (resourcesconfig.ChaincodeDefinition, error)
}

// Endorser provides the Endorser service ProcessProposal
type Endorser struct {
	//distributePrivateData privateDataDistributor
	s                     Support
}

// validateResult provides the result of endorseProposal verification
type validateResult struct {
	prop    *pb.Proposal
	hdrExt  *pb.ChaincodeHeaderExtension
	chainID string
	txid    string
	resp    *pb.ProposalResponse
}

// NewEndorserServer creates and returns a new Endorser server instance.
func NewEndorserServer(s Support) pb.EndorserServer {
	e := &Endorser{
		s: s,
	}
	return e
}

//call specified chaincode (system or user)
func (e *Endorser) callChaincode(ctxt context.Context, chainID string, version string, txid string, signedProp *pb.SignedProposal, prop *pb.Proposal, cis *pb.ChaincodeInvocationSpec, cid *pb.ChaincodeID, txsim rwset.TxSimulator) (*pb.Response, *pb.ChaincodeEvent, error) {
	logger.Debugf("[%s][%s] Entry chaincode: %s version: %s", chainID, shorttxid(txid), cid, version)
	defer logger.Debugf("[%s][%s] Exit", chainID, shorttxid(txid))
	var err error
	var res *pb.Response
	var ccevent *pb.ChaincodeEvent

	if txsim != nil {
		ctxt = context.WithValue(ctxt, core.TXSimulatorKey, txsim)
	}

	scc := e.s.IsSysCC(cid.Name)
	res, ccevent, err = e.s.Execute(ctxt, chainID, cid.Name, version, txid, scc, signedProp, prop, cis)
	if err != nil {
		return nil, nil, err
	}

	if res.Status >= shim.ERRORTHRESHOLD {
		return res, nil, nil
	}

	return res, ccevent, err
}

func (e *Endorser) simulateProposal(ctx context.Context, chainID string, txid string, signedProp *pb.SignedProposal, prop *pb.Proposal, cid *pb.ChaincodeID, txsim rwset.TxSimulator) (*pb.Response, []byte, *pb.ChaincodeEvent, error) {
	logger.Debugf("[%s][%s] Entry chaincode: %s", chainID, shorttxid(txid), cid)
	defer logger.Debugf("[%s][%s] Exit", chainID, shorttxid(txid))

	cis, err := putils.GetChaincodeInvocationSpec(prop)
	if err != nil {
		return nil, nil, nil, err
	}
	logger.Infof("spec=%v", cis)

	//var cdLedger resourcesconfig.ChaincodeDefinition
	//
	//if !e.s.IsSysCC(cid.Name) {
	//	cdLedger, err = e.s.GetChaincodeDefinition(ctx, chainID, txid, signedProp, prop, cid.Name, txsim)
	//	if err != nil {
	//		return nil, nil, nil, nil, errors.WithMessage(err, fmt.Sprintf("make sure the chaincode %s has been successfully instantiated and try again", cid.Name))
	//	}
	//	version = cdLedger.CCVersion()
	//
	//	err = e.s.CheckInstantiationPolicy(cid.Name, version, cdLedger)
	//	if err != nil {
	//		return nil, nil, nil, nil, err
	//	}
	//} else {
	//	version = util.GetSysCCVersion()
	//}

	//---3. execute the proposal and get simulation results
	//var simResult *ledger.TxSimulationResults
	var simResBytes []byte
	var res *pb.Response
	var ccevent *pb.ChaincodeEvent
	res, ccevent, err = e.callChaincode(ctx, chainID, cid.Version, txid, signedProp, prop, cis, cid, txsim)
	if err != nil {
		logger.Errorf("[%s][%s] failed to invoke chaincode %s, error: %+v", chainID, shorttxid(txid), cid, err)
		return  nil, nil, nil, err
	}

	if txsim != nil {
		//if simResult, err = txsim.GetTxSimulationResults(); err != nil {
		//	return  nil, nil, nil, err
		//}
	}
	return res, simResBytes, ccevent, nil
}

//endorse the proposal by calling the ESCC
func (e *Endorser) endorseProposal(ctx context.Context, chainID string, txid string, signedProp *pb.SignedProposal, proposal *pb.Proposal, response *pb.Response, simRes []byte, event *pb.ChaincodeEvent, visibility []byte, ccid *pb.ChaincodeID, txsim rwset.TxSimulator) (*pb.ProposalResponse, error) {
	logger.Debugf("[%s][%s] Entry chaincode: %s", chainID, shorttxid(txid), ccid)
	defer logger.Debugf("[%s][%s] Exit", chainID, shorttxid(txid))

	return nil, nil
}

// ProcessProposal process the Proposal
//func (e *Endorser) ProcessProposal(ctx context.Context, signedProp *pb.SignedProposal) (*pb.ProposalResponse, error) {
func (e *Endorser) ProcessProposal(ctx context.Context, signedProp *pb.SignedProposal, prop *pb.Proposal, chainID string, txid string, cid *pb.ChaincodeID) (*pb.ProposalResponse, error) {
	addr := util.ExtractRemoteAddress(ctx)
	logger.Debug("Entering: Got request from", addr)
	defer logger.Debugf("Exit: request from", addr)

	//0 -- check and validate
	var txsim rwset.TxSimulator
	var err error
	if chainID != "" {
		if txsim, err = e.s.GetTxSimulator(chainID, txid); err != nil {
			return &pb.ProposalResponse{Response: &pb.Response{Status: 500, Message: err.Error()}}, err
		}
		//defer txsim.Done()
	}

	if  err != nil {
		return &pb.ProposalResponse{Response: &pb.Response{Status: 500, Message: err.Error()}}, err
	}

	//1 -- simulate
	res, _, ccevent, err := e.simulateProposal(ctx, chainID, txid, signedProp, prop, cid, txsim)
	if err != nil {
		return &pb.ProposalResponse{Response: &pb.Response{Status: 500, Message: err.Error()}}, err
	}

	if res != nil {
		if res.Status >= shim.ERROR {
			logger.Errorf("[%s][%s] simulateProposal() resulted in chaincode[] response status %d for txid: %s",
				chainID, shorttxid(txid),  res.Status, txid)

			resp := &pb.ProposalResponse{
				Payload:  nil,
				Response: &pb.Response{Status: 500, Message: "Chaincode Error"}}
			return resp, err
		}
	}else {
		logger.Error("simulateProposal response is nil")
		return &pb.ProposalResponse{
			Payload:  nil, Response: &pb.Response{Status: 500, Message: "Chaincode Error"}}, nil
	}

	//2 -- endorse and get a marshalled ProposalResponse message
	pResp := &pb.ProposalResponse{Response: res}
	unit, err := converRwTxResult2DagUnit(txsim)
	if err != nil {
		logger.Errorf("chainID[%s] converRwTxResult2DagUnit failed", chainID)
		return nil, errors.New("Conver RwSet to dag unit fail")
	}
	logger.Info("write dag unit[%v], ccevent[%v]", unit, ccevent)
	// todo

	pResp.Response.Payload = res.Payload

	return pResp, nil
}
