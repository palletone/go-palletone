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
	"github.com/pkg/errors"
	"golang.org/x/net/context"
	"time"
)

//Execute - execute proposal, return original response of chaincode
func Execute(ctxt context.Context, cccid *ccprovider.CCContext, spec interface{}, timeout time.Duration) (*pb.Response, *pb.ChaincodeEvent, error) {
	log.Debugf("execute enter")
	var err error
	var cds *pb.ChaincodeDeploymentSpec
	var ci *pb.ChaincodeInvocationSpec

	//init will call the Init method of a on a chain
	cctyp := pb.ChaincodeMessage_INIT
	if cds, _ = spec.(*pb.ChaincodeDeploymentSpec); cds == nil {
		if ci, _ = spec.(*pb.ChaincodeInvocationSpec); ci == nil {
			log.Error("Execute, Execute should be called with deployment or invocation spec")
			return nil, nil, errors.New("Execute should be called with deployment or invocation spec")
		}
		cctyp = pb.ChaincodeMessage_TRANSACTION
	}

	_, cMsg, err := theChaincodeSupport.Launch(ctxt, cccid, spec)
	if err != nil {
		log.Errorf("Execute %s error: %+v", cccid.Name, err)
		return nil, nil, err
	}

	cMsg.Decorations = cccid.ProposalDecorations
	var ccMsg *pb.ChaincodeMessage
	ccMsg, err = createCCMessage(cccid.ContractId, cctyp, cccid.ChainID, cccid.TxID, cMsg)
	if err != nil {
		return nil, nil, errors.WithMessage(err, "failed to create chaincode message")
	}

	resp, err := theChaincodeSupport.Execute(ctxt, cccid, ccMsg, timeout) //theChaincodeSupport.executetimeout
	log.Debugf("resp")
	if err != nil {
		// Rollback transaction
		return nil, nil, errors.WithMessage(err, "failed to execute transaction")
	} else if resp == nil {
		// Rollback transaction
		return nil, nil, errors.Errorf("failed to receive a response for txid (%s)", cccid.TxID)
	}

	if resp.ChaincodeEvent != nil {
		resp.ChaincodeEvent.ChaincodeId = cccid.Name
		resp.ChaincodeEvent.TxId = cccid.TxID
	}

	if resp.Type == pb.ChaincodeMessage_COMPLETED {
		res := &pb.Response{}
		unmarshalErr := proto.Unmarshal(resp.Payload, res)
		if unmarshalErr != nil {
			return nil, nil, errors.Wrap(unmarshalErr, fmt.Sprintf("failed to unmarshal response for txid (%s)", cccid.TxID))
		}

		// Success
		return res, resp.ChaincodeEvent, nil
	} else if resp.Type == pb.ChaincodeMessage_ERROR {
		// Rollback transaction
		return nil, resp.ChaincodeEvent, errors.Errorf("transaction returned with failure: %s", string(resp.Payload))
	}

	//TODO - this should never happen ... a panic is more appropriate but will save that for future
	return nil, nil, errors.Errorf("receive a response for txid (%s) but in invalid state (%d)", cccid.TxID, resp.Type)
}

// ExecuteWithErrorFilter is similar to Execute, but filters error contained in chaincode response and returns Payload of response only.
// Mostly used by unit-test.
func ExecuteWithErrorFilter(ctxt context.Context, cccid *ccprovider.CCContext, spec interface{}, timeout time.Duration) ([]byte, *pb.ChaincodeEvent, error) {
	res, event, err := Execute(ctxt, cccid, spec, timeout)
	if err != nil {
		log.Errorf("ExecuteWithErrorFilter %s error: %+v", cccid.Name, err)
		return nil, nil, err
	}

	if res == nil {
		log.Errorf("ExecuteWithErrorFilter %s get nil response without error", cccid.Name)
		return nil, nil, err
	}

	if res.Status != shim.OK {
		return nil, nil, errors.New(res.Message)
	}

	return res.Payload, event, nil
}
