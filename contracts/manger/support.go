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
	"golang.org/x/net/context"
	"github.com/golang/protobuf/proto"
	"github.com/palletone/go-palletone/contracts/rwset"

	"github.com/palletone/go-palletone/core/vmContractPub/ccprovider"
	"github.com/palletone/go-palletone/contracts/scc"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	ut "github.com/palletone/go-palletone/dag/modules"
	chaincode "github.com/palletone/go-palletone/contracts/core"
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

		logger.Infof("===cis:%v", cis)

		//decorate the chaincode input

		//decorators := library.InitRegistry(library.Config{}).Lookup(library.Decoration).([]decoration.Decorator)
		//cis.ChaincodeSpec.Input.Decorations = make(map[string][]byte)
		//cis.ChaincodeSpec.Input = decoration.Apply(prop, cis.ChaincodeSpec.Input, decorators...)
		//cccid.ProposalDecorations = cis.ChaincodeSpec.Input.Decorations
		return chaincode.ExecuteChaincode(ctxt, cccid, cis.ChaincodeSpec.Input.Args)
	default:
		panic("programming error, unkwnown spec type")
	}
}

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

func RwTxResult2DagInvokeUnit(tx rwset.TxSimulator, txid string, nm string, fun []byte) (*ut.ContractInvokePayload, error) {
	logger.Debug("enter")

	invokeData := ut.ContractInvokePayload{}
	invokeData.ContractId = txid

	rd, wt, err := tx.GetRwData(nm)
	if err != nil {
		return nil, err
	}

	logger.Infof("txid=%s, nm=%s, rd=%v, wt=%v", txid, nm, rd, wt)
	dag := ut.ContractInvokePayload{ContractId:txid, Function: fun, ReadSet:make(map[string]interface{}), WriteSet:make(map[string]interface{})}

	for key, val:= range rd {
		dag.ReadSet[key] = val
		logger.Infof("readSet: fun[%s], key[%s], val[%v]", dag.Function, key, val)
	}
	for key, val:= range wt {
		dag.WriteSet[key] = val
		logger.Infof("WriteSet: fun[%s], key[%s], val[%v]", dag.Function, key, dag.WriteSet[key])
	}

	return &dag, nil
}

var rwM *rwset.RwSetTxMgr

func init() {
	var err error
	rwM, err =rwset.NewRwSetMgr("default")
	if err != nil {
		logger.Error("fail!")
	}
}


