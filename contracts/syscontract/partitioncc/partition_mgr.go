/*
 *
 *    This file is part of go-palletone.
 *    go-palletone is free software: you can redistribute it and/or modify
 *    it under the terms of the GNU General Public License as published by
 *    the Free Software Foundation, either version 3 of the License, or
 *    (at your option) any later version.
 *    go-palletone is distributed in the hope that it will be useful,
 *    but WITHOUT ANY WARRANTY; without even the implied warranty of
 *    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *    GNU General Public License for more details.
 *    You should have received a copy of the GNU General Public License
 *    along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
 * /
 *
 *  * @author PalletOne core developer <dev@pallet.one>
 *  * @date 2018-2019
 *
 */

package partitioncc

import (
	"encoding/json"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/contracts/shim"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"github.com/palletone/go-palletone/dag/errors"
	dm "github.com/palletone/go-palletone/dag/modules"
	"strconv"
)

type PartitionMgr struct {
}

func (p *PartitionMgr) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

func (p *PartitionMgr) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	f, args := stub.GetFunctionAndParameters()

	switch f {
	case "registerPartition":
		return registerPartition(args, stub)
	case "listPartition":
		return listPartition(stub)
	case "updatePartition":
		return updatePartition(args, stub)
	case "setMainChain":
		return setMainChain(args, stub)
	case "getMainChain":
		return getMainChain(stub)
	default:
		jsonResp := "{\"Error\":\"Unknown function " + f + "\"}"
		return shim.Error(jsonResp)
	}
}

const PartitionChainPrefix = "PC"
const MainChainKey = "MainChain"
const ErrorForbiddenAccess = "Forbidden access"

func getPartitionChains(stub shim.ChaincodeStubInterface) ([]*dm.PartitionChain, error) {
	list, err := stub.GetStateByPrefix(PartitionChainPrefix)
	if err != nil {
		return nil, err
	}
	chains := []*dm.PartitionChain{}
	for _, kv := range list {
		data := kv.Value
		var partitionChain *dm.PartitionChain
		err = json.Unmarshal(data, &partitionChain)
		if err != nil {
			return nil, err
		}
		chains = append(chains, partitionChain)
	}

	return chains, nil
}
func addPartitionChain(stub shim.ChaincodeStubInterface, chain *dm.PartitionChain) error {
	key := PartitionChainPrefix + chain.GasToken.String()
	value, err := json.Marshal(chain)
	if err != nil {
		return err
	}
	return stub.PutState(key, value)
}
func buildPartitionChain(args []string) (*dm.PartitionChain, error) {
	if len(args) < 8 {
		return nil, errors.New("need 8 args (GenesisHash,GenesisHeight,ForkUnitHash,ForkUnitHeight,GasToken,Status,SyncModel,[Peers])")
	}
	var err error
	partitionChain := &dm.PartitionChain{}
	partitionChain.GenesisHash = common.HexToHash(args[0])
	partitionChain.GenesisHeight, _ = strconv.ParseUint(args[1], 10, 64)
	partitionChain.ForkUnitHash = common.HexToHash(args[2])
	partitionChain.ForkUnitHeight, _ = strconv.ParseUint(args[3], 10, 64)
	partitionChain.GasToken, _, err = dm.String2AssetId(args[4])
	if err != nil {
		return nil, err
	}
	partitionChain.Status = args[5][0] - '0'
	partitionChain.SyncModel = args[6][0] - '0'
	if len(args[7]) > 0 {
		peers := []string{}
		err = json.Unmarshal([]byte(args[7]), &peers)
		if err != nil {
			return nil, err
		}
		partitionChain.Peers = peers
	}
	return partitionChain, nil
}
func hasPermission(stub shim.ChaincodeStubInterface) bool {
	requester, _, _, _, _, _ := stub.GetInvokeParameters()
	foundationAddress, _ := stub.GetSystemConfig("FoundationAddress")
	return foundationAddress == requester.String()
}
func registerPartition(args []string, stub shim.ChaincodeStubInterface) pb.Response {
	if !hasPermission(stub) {
		return shim.Error(ErrorForbiddenAccess)
	}
	//params check
	partitionChain, err := buildPartitionChain(args)
	if err != nil {
		return shim.Error(err.Error())
	}
	err = addPartitionChain(stub, partitionChain)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

func listPartition(stub shim.ChaincodeStubInterface) pb.Response {

	chains, err := getPartitionChains(stub)
	if err != nil {
		return shim.Error(err.Error())
	}
	data, _ := json.Marshal(chains)
	return shim.Success(data)

}

func updatePartition(args []string, stub shim.ChaincodeStubInterface) pb.Response {
	if !hasPermission(stub) {
		return shim.Error(ErrorForbiddenAccess)
	}
	partitionChain, err := buildPartitionChain(args)
	if err != nil {
		return shim.Error(err.Error())
	}
	err = addPartitionChain(stub, partitionChain)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}
func buildMainChain(args []string) (*dm.MainChain, error) {
	if len(args) < 5 {
		return nil, errors.New("need 5 args (GenesisHash,GasToken,Status,SyncModel,[Peers])")
	}
	var err error
	mainChain := &dm.MainChain{}
	mainChain.GenesisHash = common.HexToHash(args[0])
	mainChain.GasToken, _, err = dm.String2AssetId(args[1])
	if err != nil {
		return nil, err
	}
	mainChain.Status = args[2][0] - '0'
	mainChain.SyncModel = args[3][0] - '0'
	if len(args[4]) > 0 {
		peers := []string{}
		err = json.Unmarshal([]byte(args[4]), &peers)
		if err != nil {
			return nil, err
		}
		mainChain.Peers = peers
	}
	return mainChain, nil
}
func setMainChain(args []string, stub shim.ChaincodeStubInterface) pb.Response {
	if !hasPermission(stub) {
		return shim.Error(ErrorForbiddenAccess)
	}
	mainChain, err := buildMainChain(args)
	if err != nil {
		return shim.Error(err.Error())
	}
	value, err := json.Marshal(mainChain)
	if err != nil {
		return shim.Error(err.Error())
	}
	err = stub.PutState(MainChainKey, value)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}
func getMainChain(stub shim.ChaincodeStubInterface) pb.Response {
	data, err := stub.GetState(MainChainKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	var mainChain *dm.MainChain
	err = json.Unmarshal(data, &mainChain)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(data)
}
