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
 * Copyright IBM Corp. All Rights Reserved.
 * @author PalletOne core developers <dev@pallet.one>
 * @date 2018
 */

package core

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/contracts/shim"
	"github.com/palletone/go-palletone/core/vmContractPub/ccprovider"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"github.com/palletone/go-palletone/core/vmContractPub/util"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
	"time"
)

//create a chaincode invocation spec
func createCIS(ccname string, args [][]byte) (*pb.ChaincodeInvocationSpec) {
	spec := &pb.ChaincodeInvocationSpec{ChaincodeSpec: &pb.ChaincodeSpec{Type: pb.ChaincodeSpec_Type(pb.ChaincodeSpec_Type_value["GOLANG"]), ChaincodeId: &pb.ChaincodeID{Name: ccname}, Input: &pb.ChaincodeInput{Args: args}}}
	return spec
}

// GetCDS retrieves a chaincode deployment spec for the required chaincode
func GetCDS(contractid []byte, ctxt context.Context, txid string, signedProp *pb.SignedProposal, prop *pb.Proposal, chainID string, chaincodeID string) ([]byte, error) {
	version := util.GetSysCCVersion()
	log.Infof("chainID[%s] txid[%s]", chainID, txid)

	cccid := ccprovider.NewCCContext(contractid, chainID, "lscc", version, txid, true, signedProp, prop)
	res, _, err := ExecuteChaincode(ctxt, cccid, [][]byte{[]byte("getdepspec"), []byte(chainID), []byte(chaincodeID)}, 0)
	if err != nil {
		return nil, errors.WithMessage(err, fmt.Sprintf("execute getdepspec(%s, %s) of LSCC error", chainID, chaincodeID))
	}
	if res.Status != shim.OK {
		return nil, errors.Errorf("get ChaincodeDeploymentSpec for %s/%s from LSCC error: %s", chaincodeID, chainID, res.Message)
	}

	return res.Payload, nil
}

//glh
//add
// ChaincodeDefinition describes all of the necessary information for a peer to decide whether to endorse
// a proposal and whether to validate a transaction, for a particular chaincode.
type ChaincodeDefinition interface {
	// CCName returns the name of this chaincode (the name it was put in the ChaincodeRegistry with).
	CCName() string

	// Hash returns the hash of the chaincode.
	Hash() []byte

	// CCVersion returns the version of the chaincode.
	CCVersion() string

	// Validation returns how to validate transactions for this chaincode.
	// The string returned is the name of the validation method (usually 'vscc')
	// and the bytes returned are the argument to the validation (in the case of
	// 'vscc', this is a marshaled pb.VSCCArgs message).
	Validation() (string, []byte)

	// Endorsement returns how to endorse proposals for this chaincode.
	// The string returns is the name of the endorsement method (usually 'escc').
	Endorsement() string
}

// GetChaincodeDefinition returns resourcesconfig.ChaincodeDefinition for the chaincode with the supplied name
func GetChaincodeDefinition(ctxt context.Context, txid string, signedProp *pb.SignedProposal, prop *pb.Proposal, chainID string, chaincodeID string) (ChaincodeDefinition, error) {
	version := util.GetSysCCVersion()
	//cccid := ccprovider.NewCCContext(chainID, "lscc", version, txid, true, signedProp, prop)
	cccid := ccprovider.NewCCContext(nil, chainID, "lscc", version, txid, true, signedProp, prop)
	res, _, err := ExecuteChaincode(ctxt, cccid, [][]byte{[]byte("getccdata"), []byte(chainID), []byte(chaincodeID)}, 0)
	if err == nil {
		if res.Status != shim.OK {
			return nil, errors.New(res.Message)
		}
		cd := &ccprovider.ChaincodeData{}
		err = proto.Unmarshal(res.Payload, cd)
		if err != nil {
			return nil, err
		}
		return cd, nil
	}

	return nil, err
}

// ExecuteChaincode executes a given chaincode given chaincode name and arguments
func ExecuteChaincode(ctxt context.Context, cccid *ccprovider.CCContext, args [][]byte, timeout time.Duration) (*pb.Response, *pb.ChaincodeEvent, error) {
	log.Debugf("execute chain code")
	var spec *pb.ChaincodeInvocationSpec
	var err error
	var res *pb.Response
	var ccevent *pb.ChaincodeEvent

	spec= createCIS(cccid.Name, args)
	res, ccevent, err = Execute(ctxt, cccid, spec, timeout)
	if err != nil {
		err = errors.WithMessage(err, "error executing chaincode")
		log.Errorf("execute %+v", err)
		return nil, nil, err
	}

	return res, ccevent, err
}
