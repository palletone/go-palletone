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
 *  * @date 2018
 *
 */

package storage

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/common/rlp"
	"github.com/palletone/go-palletone/dag/constants"
	"github.com/palletone/go-palletone/dag/modules"
)

func (statedb *StateDb) SaveContract(contract *modules.Contract) error {
	//保存一个新合约的状态信息
	//如果数据库中已经存在同样的合约ID，则报错
	prefix := append(constants.CONTRACT_PREFIX, contract.Id...)
	count := getCountByPrefix(statedb.db, prefix)
	if count > 0 {
		return errors.New("Contract[" + common.Bytes2Hex(contract.Id) + "]'s state existed!")
	}
	return StoreBytes(statedb.db, prefix, contract)
}
func (statedb *StateDb) SaveContractState(id []byte, name string, value interface{}, version *modules.StateVersion) error {
	return SaveContractState(statedb, constants.CONTRACT_STATE_PREFIX, id, name, value, version)
}

/**
保存合约属性信息
To save contract
*/
func SaveContractState(statedb *StateDb, prefix []byte, id []byte, field string, value interface{}, version *modules.StateVersion) error {
	key := []byte{}
	key = append(prefix, id...)
	//key = append(key, []byte(modules.FIELD_SPLIT_STR)...)
	key = append(key, []byte(field)...)
	//key = append(key, []byte(modules.FIELD_SPLIT_STR)...)
	//key = append(key, version.Bytes()...)
	log.Debug(fmt.Sprintf("Try to save contract state with version:%x", version.Bytes()))
	if err := StoreBytesWithVersion(statedb.db, key, version, value); err != nil {
		log.Error("Save contract template", err.Error())
		return err
	}
	return nil
}

func SaveContract(db ptndb.Database, contract *modules.Contract) error {
	if common.EmptyHash(contract.CodeHash) {
		contract.CodeHash = rlp.RlpHash(contract.Code)
	}
	// key = cs+ rlphash(contract)
	//if common.EmptyHash(contract.Id) {
	//	ids := rlp.RlpHash(contract)
	//	if len(ids) > len(contract.Id) {
	//		id := ids[len(ids)-common.HashLength:]
	//		copy(contract.Id[common.HashLength-len(id):], id)
	//	} else {
	//		//*contract.Id = new(common.Hash)
	//		copy(contract.Id[common.HashLength-len(ids):], ids[:])
	//	}
	//
	//}

	return StoreBytes(db, append(constants.CONTRACT_PREFIX, contract.Id[:]...), contract)
}

// Get contract key's value
func GetContractKeyValue(db DatabaseReader, id common.Hash, key string) (interface{}, error) {
	var val interface{}
	if common.EmptyHash(id) {
		return nil, errors.New("the filed not defined")
	}
	con_bytes, err := db.Get(append(constants.CONTRACT_PREFIX, id[:]...))
	if err != nil {
		return nil, err
	}
	contract := new(modules.Contract)
	err = rlp.DecodeBytes(con_bytes, contract)
	if err != nil {
		log.Errorf("err:", err)
		return nil, err
	}
	obj := reflect.ValueOf(contract)
	myref := obj.Elem()
	typeOftype := myref.Type()

	for i := 0; i < myref.NumField(); i++ {
		filed := myref.Field(i)
		if typeOftype.Field(i).Name == key {
			val = filed.Interface()
			log.Errorf("", i, ". ", typeOftype.Field(i).Name, " ", filed.Type(), "=: ", filed.Interface())
			break
		} else if i == myref.NumField()-1 {
			val = nil
		}
	}
	return val, nil
}

func (statedb *StateDb) SaveContractTemplateState(id []byte, name string, value interface{}, version *modules.StateVersion) error {
	return SaveContractState(statedb, constants.CONTRACT_TPL, id, name, value, version)
}

/**
获取合约（或模板）所有属性
To get contract or contract template all fields and return
*/
func (statedb *StateDb) GetContractAllState() []*modules.ContractReadSet {
	// key format: [PREFIX 2][ID 20][field]
	// key := append(modules.CONTRACT_STATE_PREFIX, id...)
	data := getprefix(statedb.db, constants.CONTRACT_STATE_PREFIX)
	if data == nil || len(data) <= 0 {
		return nil
	}
	allState := []*modules.ContractReadSet{}
	for k, v := range data {
		if len(k) <= 22 {
			//Contract本身的状态，而不是某个Field的值
			continue
		}
		sKey := string(k[22:])
		data, version, err := splitValueAndVersion(v)
		if err != nil {
			log.Error("Invalid state data, cannot parse and split version")
			continue
		}
		rdSet := &modules.ContractReadSet{
			Key:     sKey,
			Value:   data,
			Version: version,
		}
		allState = append(allState, rdSet)
	}
	return allState
}

/**
获取合约（或模板）全部属性
To get contract or contract template all fields
*/
func (statedb *StateDb) GetContractStatesById(id []byte) (map[string]*modules.ContractStateValue, error) {
	key := append(constants.CONTRACT_STATE_PREFIX, id...)
	data := getprefix(statedb.db, key)
	if data == nil || len(data) == 0 {
		return nil, errors.New(fmt.Sprintf("the contract %s state is null.", id))
	}
	var err error
	result := make(map[string]*modules.ContractStateValue, 0)
	for dbkey, state_version := range data {
		state, version, err0 := splitValueAndVersion(state_version)
		if err0 != nil {
			err = err0
		}
		realKey := dbkey[len(key):]
		if realKey != "" {
			result[realKey] = &modules.ContractStateValue{Value: state, Version: version}
			log.Info("the contract's state get info.", "key", realKey)
		}
	}
	return result, err
}

/**
获取合约（或模板）某一个属性
To get contract or contract template one field
*/
func (statedb *StateDb) GetContractState(id []byte, field string) (*modules.StateVersion, []byte) {
	key := append(constants.CONTRACT_STATE_PREFIX, id...)
	key = append(key, []byte(field)...)
	data, version, err := retrieveWithVersion(statedb.db, key)
	if err != nil || data == nil {

		return nil, nil
	}
	return version, data
}

// GetContract can get a Contract by the contract hash
func (statedb *StateDb) GetContract(id []byte) (*modules.Contract, error) {
	con_bytes, err := statedb.db.Get(append(constants.CONTRACT_PREFIX, id...))
	if err != nil {
		log.Errorf("err:", err)
		return nil, err
	}
	contract := new(modules.Contract)
	err = rlp.DecodeBytes(con_bytes, contract)
	if err != nil {
		log.Error("err:", err)
		return nil, err
	}
	return contract, nil
}

func (statedb *StateDb) SaveContractDeploy(reqid []byte, deploy *modules.ContractDeployPayload) error {
	// key  requestId
	key := append(constants.CONTRACT_DEPLOY, reqid...)
	return StoreBytes(statedb.db, key, deploy)
}

func (statedb *StateDb) GetContractDeploy(reqId []byte) (*modules.ContractDeployPayload, error) {
	key := append(constants.CONTRACT_DEPLOY, reqId...)
	data, err := statedb.db.Get(key)
	if err != nil {
		return nil, err
	}
	deploy := new(modules.ContractDeployPayload)
	if err := rlp.DecodeBytes(data, &deploy); err != nil {
		return nil, err
	}
	return deploy, nil
}

func (statedb *StateDb) SaveContractDeployReq(reqid []byte, deploy *modules.ContractDeployRequestPayload) error {
	// key  requestId
	key := append(constants.CONTRACT_DEPLOY_REQ, reqid...)
	return StoreBytes(statedb.db, key, deploy)
}

func (statedb *StateDb) GetContractDeployReq(reqId []byte) (*modules.ContractDeployRequestPayload, error) {
	key := append(constants.CONTRACT_DEPLOY_REQ, reqId...)
	data, err := statedb.db.Get(key)
	if err != nil {
		return nil, err
	}
	deploy := new(modules.ContractDeployRequestPayload)
	if err := rlp.DecodeBytes(data, &deploy); err != nil {
		return nil, err
	}
	return deploy, nil
}

func (statedb *StateDb) SaveContractInvokeReq(reqid []byte, invoke *modules.ContractInvokeRequestPayload) error {
	// key   contractId  + funcName
	key := append(constants.CONTRACT_INVOKE_REQ, reqid...)
	return StoreBytes(statedb.db, key, invoke)
}

func (statedb *StateDb) GetContractInvokeReq(reqId []byte) (*modules.ContractInvokeRequestPayload, error) {
	key := append(constants.CONTRACT_INVOKE_REQ, reqId...)
	data, err := statedb.db.Get(key)
	if err != nil {
		return nil, err
	}
	deploy := new(modules.ContractInvokeRequestPayload)
	if err := rlp.DecodeBytes(data, &deploy); err != nil {
		return nil, err
	}
	return deploy, nil
}
