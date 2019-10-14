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
	"encoding/hex"
	"encoding/json"
	"strconv"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/contracts/shim"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	dm "github.com/palletone/go-palletone/dag/modules"
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
		if len(args) < 11 {
			return shim.Error("need 11 args (GenesisHeaderRlp,ForkUnitHash,ForkUnitHeight,GasToken,Status," +
				"SyncModel,NetworkId,Version,StableThreshold,CrossChainToken,[]Peers)")
		}
		peers := []string{}
		err := json.Unmarshal([]byte(args[10]), &peers)
		if err != nil {
			return shim.Error(err.Error())
		}
		return p.RegisterPartition(stub, args[0], args[1], args[2], args[3], args[4], args[5], args[6], args[7],
			args[8], args[9], peers)
	case "listPartition":
		result, err := p.ListPartition(stub)
		if err != nil {
			return shim.Error(err.Error())
		}
		data, _ := json.Marshal(result)
		return shim.Success(data)
	case "updatePartition":
		if len(args) < 11 {
			return shim.Error("need 11 args (GenesisHeaderRlp,ForkUnitHash,ForkUnitHeight,GasToken,Status," +
				"SyncModel,NetworkId,Version,StableThreshold,CrossChainToken,[]Peers)")
		}
		peers := []string{}
		err := json.Unmarshal([]byte(args[10]), &peers)
		if err != nil {
			return shim.Error(err.Error())
		}
		return p.UpdatePartition(stub, args[0], args[1], args[2], args[3], args[4], args[5], args[6], args[7],
			args[8], args[9], peers)
	case "setMainChain":
		if len(args) < 9 {
			return shim.Error("need 9 args (GenesisHeaderHex,GasToken,Status,SyncModel,NetworkId,Version," +
				"StableThreshold,CrossChainToken,[]Peers)")
		}
		peers := []string{}
		err := json.Unmarshal([]byte(args[10]), &peers)
		if err != nil {
			return shim.Error(err.Error())
		}
		return p.SetMainChain(stub, args[0], args[1], args[2], args[3], args[4], args[5], args[6], args[7], peers)
	case "getMainChain":
		result, err := p.GetMainChain(stub)
		if err != nil {
			return shim.Error(err.Error())
		}
		data, _ := json.Marshal(result)
		return shim.Success(data)
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
func buildPartitionChain(genesisHeaderRlp, forkUnitHash, forkUnitHeight, gasToken, status, syncModel, networkId, version,
	stableThreshold, crossChainToken string, peers []string) (*dm.PartitionChain, error) {
	var err error
	gbytes, err := hex.DecodeString(genesisHeaderRlp)
	if err != nil {
		return nil, err
	}
	header := &dm.Header{}
	err = rlp.DecodeBytes(gbytes, header)
	if err != nil {
		return nil, err
	}
	partitionChain := &dm.PartitionChain{}
	partitionChain.GenesisHeaderRlp = gbytes
	//partitionChain.GenesisHeight, _ = strconv.ParseUint(args[1], 10, 64)
	partitionChain.ForkUnitHash = common.HexToHash(forkUnitHash)
	partitionChain.ForkUnitHeight, _ = strconv.ParseUint(forkUnitHeight, 10, 64)
	partitionChain.GasToken, _, err = dm.String2AssetId(gasToken)
	if err != nil {
		return nil, err
	}
	partitionChain.Status = status[0] - '0'
	partitionChain.SyncModel = syncModel[0] - '0'
	partitionChain.NetworkId, _ = strconv.ParseUint(networkId, 10, 64)
	partitionChain.Version, _ = strconv.ParseUint(version, 10, 64)
	threshold, _ := strconv.ParseUint(stableThreshold, 10, 32)
	partitionChain.StableThreshold = uint32(threshold)
	partitionChain.Peers = peers

	tokens := []dm.AssetId{}
	err = json.Unmarshal([]byte(crossChainToken), &tokens)
	if err != nil {
		return nil, err
	}
	partitionChain.CrossChainTokens = tokens
	return partitionChain, nil
}
func hasPermission(stub shim.ChaincodeStubInterface) bool {
	requester, _, _, _, _, _ := stub.GetInvokeParameters()
	//foundationAddress, _ := stub.GetSystemConfig(dm.FoundationAddress)
	gp, err := stub.GetSystemConfig()
	if err != nil {
		//log.Error("strconv.ParseUint err:", "error", err)
		return false
	}
	foundationAddress := gp.ChainParameters.FoundationAddress
	return foundationAddress == requester.String()
}
func (p *PartitionMgr) RegisterPartition(stub shim.ChaincodeStubInterface, genesisHeaderRlp, forkUnitHash,
	forkUnitHeight, gasToken, status, syncModel, networkId, version, stableThreshold, crossChainToken string,
	peers []string) pb.Response {
	partitionChain, err := buildPartitionChain(genesisHeaderRlp, forkUnitHash, forkUnitHeight, gasToken, status,
		syncModel, networkId, version, stableThreshold, crossChainToken, peers)
	if err != nil {
		return shim.Error(err.Error())
	}

	if !hasPermission(stub) {
		return shim.Error(ErrorForbiddenAccess)
	}
	err = addPartitionChain(stub, partitionChain)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

func (p *PartitionMgr) ListPartition(stub shim.ChaincodeStubInterface) ([]*dm.PartitionChain, error) {
	return getPartitionChains(stub)
}

func (p *PartitionMgr) UpdatePartition(stub shim.ChaincodeStubInterface, genesisHeaderRlp, forkUnitHash,
	forkUnitHeight, gasToken, status, syncModel, networkId, version, stableThreshold, crossChainToken string,
	peers []string) pb.Response {
	partitionChain, err := buildPartitionChain(genesisHeaderRlp, forkUnitHash, forkUnitHeight, gasToken, status,
		syncModel, networkId, version, stableThreshold, crossChainToken, peers)
	if err != nil {
		return shim.Error(err.Error())
	}

	if !hasPermission(stub) {
		return shim.Error(ErrorForbiddenAccess)
	}
	err = addPartitionChain(stub, partitionChain)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}
func buildMainChain(genesisHeaderHex, gasToken, status, syncModel, networkId, version, stableThreshold,
	crossChainToken string, peers []string) (*dm.MainChain, error) {
	gbytes, _ := hex.DecodeString(genesisHeaderHex)
	var err error
	header := &dm.Header{}
	err = rlp.DecodeBytes(gbytes, header)
	if err != nil {
		return nil, err
	}

	mainChain := &dm.MainChain{}
	mainChain.GenesisHeaderRlp = gbytes
	mainChain.GasToken, _, err = dm.String2AssetId(gasToken)
	if err != nil {
		return nil, err
	}
	mainChain.Status = status[0] - '0'
	mainChain.SyncModel = syncModel[0] - '0'
	mainChain.NetworkId, _ = strconv.ParseUint(networkId, 10, 64)
	mainChain.Version, _ = strconv.ParseUint(version, 10, 64)
	threshold, _ := strconv.ParseUint(stableThreshold, 10, 32)
	mainChain.StableThreshold = uint32(threshold)
	mainChain.Peers = peers

	tokens := []dm.AssetId{}
	err = json.Unmarshal([]byte(crossChainToken), &tokens)
	if err != nil {
		return nil, err
	}
	mainChain.CrossChainTokens = tokens

	return mainChain, nil
}
func (p *PartitionMgr) SetMainChain(stub shim.ChaincodeStubInterface, genesisHeaderHex, gasToken, status, syncModel,
	networkId, version, stableThreshold, crossChainToken string, peers []string) pb.Response {
	mainChain, err := buildMainChain(genesisHeaderHex, gasToken, status, syncModel,
		networkId, version, stableThreshold, crossChainToken, peers)
	if err != nil {
		return shim.Error(err.Error())
	}

	if !hasPermission(stub) {
		return shim.Error(ErrorForbiddenAccess)
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
func (p *PartitionMgr) GetMainChain(stub shim.ChaincodeStubInterface) (*dm.MainChain, error) {
	data, err := stub.GetState(MainChainKey)
	if err != nil {
		return nil, err
	}
	var mainChain *dm.MainChain
	err = json.Unmarshal(data, &mainChain)
	if err != nil {
		return nil, err
	}
	return mainChain, nil
}
