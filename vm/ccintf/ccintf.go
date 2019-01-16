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

package ccintf

//This package defines the interfaces that support runtime and
//communication between chaincode and peer (chaincode support).
//Currently inproccontroller uses it. dockercontroller does not.

import (
	"encoding/hex"

	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"github.com/palletone/go-palletone/core/vmContractPub/util"
	"golang.org/x/net/context"
)

//ChaincodeStream interface for stream between Peer and chaincode instance.
type ChaincodeStream interface {
	Send(*pb.ChaincodeMessage) error
	Recv() (*pb.ChaincodeMessage, error)
}

//CCSupport must be implemented by the chaincode support side in peer
//(such as chaincode_support)
type CCSupport interface {
	HandleChaincodeStream(context.Context, ChaincodeStream) error
}

// GetCCHandlerKey is used to pass CCSupport via context
func GetCCHandlerKey() string {
	return "CCHANDLER"
}

//CCID encapsulates chaincode ID
type CCID struct {
	ChaincodeSpec *pb.ChaincodeSpec
	NetworkID     string
	PeerID        string
	ChainID       string
	Version       string
}

//GetName returns canonical chaincode name based on chain name
func (ccid *CCID) GetName() string {
	if ccid.ChaincodeSpec == nil {
		panic("nil chaincode spec")
	}

	name := ccid.ChaincodeSpec.ChaincodeId.Name
	if ccid.Version != "" {
		name = name + "-" + ccid.Version
	}

	//this better be chainless system chaincode!
	if ccid.ChainID != "" {
		hash := util.ComputeSHA256([]byte(ccid.ChainID))
		hexstr := hex.EncodeToString(hash[:])
		name = name + "-" + hexstr
	}

	return name
}
