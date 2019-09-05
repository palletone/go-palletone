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

package deposit

import (
	"fmt"
	"github.com/btcsuite/btcutil/base58"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/contracts/shim"
	"github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"github.com/palletone/go-palletone/dag/modules"
	"go.dedis.ch/kyber/v3/pairing/bn256"
)

func juryPayToDepositContract(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 1 {
		return shim.Error("need 1 parameter")
	}
	pubB := base58.Decode(args[0])
	pub := bn256.NewSuiteG2().Point()
	err := pub.UnmarshalBinary(pubB)
	if err != nil {
		err = fmt.Errorf("invalid init jury public key \"%v\" : %v", args[0], err)
		return shim.Error(err.Error())
	}
	return nodePayToDepositContract(stub, modules.Jury, args)
}

func juryApplyQuit(stub shim.ChaincodeStubInterface) peer.Response {
	log.Debug("juryApplyQuit")
	err := applyQuitList(modules.Jury, stub)
	if err != nil {
		log.Error("applyQuitList err: ", "error", err)
		return shim.Error(err.Error())
	}
	return shim.Success([]byte(nil))
}

//  处理
func handleJury(stub shim.ChaincodeStubInterface, quitAddr common.Address) error {
	return handleNode(stub, quitAddr, modules.Jury)
}
