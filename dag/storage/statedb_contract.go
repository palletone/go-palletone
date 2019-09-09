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
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/contracts/syscontract"
	"github.com/palletone/go-palletone/dag/constants"
	"github.com/palletone/go-palletone/dag/modules"
)

func (statedb *StateDb) SaveContract(contract *modules.Contract) error {
	//保存一个新合约的状态信息
	//如果数据库中已经存在同样的合约ID，则报错
	key := append(constants.CONTRACT_PREFIX, contract.ContractId...)
	//count := getCountByPrefix(statedb.db, key)
	//if count > 0 {
	//	return errors.New("Contract[" + common.Bytes2Hex(contract.ContractId) + "]'s state existed!")
	//}
	log.Debugf("Save contract[%x]", contract.ContractId)
	err := StoreToRlpBytes(statedb.db, key, contract)
	if err != nil {
		return err
	}
	return statedb.saveTplToContractMapping(contract)
}

func (statedb *StateDb) GetContract(id []byte) (*modules.Contract, error) {
	key := append(constants.CONTRACT_PREFIX, id...)
	contract := new(modules.Contract)
	err := RetrieveFromRlpBytes(statedb.db, key, contract)
	return contract, err

}

func (statedb *StateDb) GetAllContracts() ([]*modules.Contract, error) {
	rows := getprefix(statedb.db, constants.CONTRACT_PREFIX)
	result := make([]*modules.Contract, 0, len(rows))
	for _, v := range rows {
		contract := &modules.Contract{}
		rlp.DecodeBytes(v, contract)
		result = append(result, contract)
	}
	return result, nil
}

func (statedb *StateDb) saveTplToContractMapping(contract *modules.Contract) error {
	key := append(constants.CONTRACT_TPL_INSTANCE_MAP, contract.TemplateId...)
	key = append(key, contract.ContractId...)
	return statedb.db.Put(key, contract.ContractId)
}

func (statedb *StateDb) GetContractIdsByTpl(tplId []byte) ([][]byte, error) {
	key := append(constants.CONTRACT_TPL_INSTANCE_MAP, tplId...)
	rows := getprefix(statedb.db, key)
	result := [][]byte{}
	for _, v := range rows {
		result = append(result, v)
	}
	return result, nil
}

func MediatorDepositKey(medAddr string) string {
	return string(constants.MEDIATOR_INFO_PREFIX) + string(constants.DEPOSIT_BALANCE_PREFIX) + medAddr
}

func (statedb *StateDb) SaveContractState(contractId []byte, ws *modules.ContractWriteSet,
	version *modules.StateVersion) error {
	cid := contractId
	if ws.ContractId != nil {
		cid = ws.ContractId
	}
	key := getContractStateKey(cid, ws.Key)
	if ws.IsDelete {
		log.Debugf("Delete contract state by key:[%s]", ws.Key)
		return statedb.db.Delete(key)
	}
	if err := storeBytesWithVersion(statedb.db, key, version, ws.Value); err != nil {
		return err
	}
	return nil
}

func getContractStateKey(id []byte, field string) []byte {
	key := append(constants.CONTRACT_STATE_PREFIX, id...)
	return append(key, field...)
}

func (statedb *StateDb) GetContractJury(contractId []byte) (*modules.ElectionNode, error) {
	log.Debugf("GetContractJury contractId %x", contractId)
	key := append(constants.CONTRACT_JURY_PREFIX, contractId...)
	data, _, err := retrieveWithVersion(statedb.db, key)
	if err != nil {
		return nil, err
	}
	jury := modules.ElectionNode{}
	err = rlp.DecodeBytes(data, &jury)
	if err != nil {
		return nil, err
	}
	return &jury, nil
}
func (statedb *StateDb) SaveContractJury(contractId []byte, jury modules.ElectionNode,
	version *modules.StateVersion) error {
	log.Debugf("SaveContractJury contractId %x", contractId)
	key := append(constants.CONTRACT_JURY_PREFIX, contractId...)
	juryb, err := rlp.EncodeToBytes(jury)
	if err != nil {
		return err
	}
	return storeBytesWithVersion(statedb.db, key, version, juryb)
}

/**
保存合约属性信息,合约属性有CONTRACT_STATE_PREFIX+contractId+key 作为Key
To save contract
*/
//func saveContractState(db ptndb.Putter, id []byte, field string, value []byte, version *modules.StateVersion) error {
//	key := getContractStateKey(id, field)
//
//	log.Debugf("Try to save contract state with key:%v, version:%x", field, version.Bytes())
//	if err := storeBytesWithVersion(db, key, version, value); err != nil {
//		log.Error("Save contract state error", err.Error())
//		return err
//	}
//	return nil
//}

func (statedb *StateDb) SaveContractStates(id []byte, wset []modules.ContractWriteSet,
	version *modules.StateVersion) error {
	batch := statedb.db.NewBatch()
	for _, write := range wset {
		cid := id
		if len(write.ContractId) != 0 {
			cid = write.ContractId
		}
		key := getContractStateKey(cid, write.Key)

		if write.IsDelete {
			batch.Delete(key)
			log.Debugf("Delete contract state by key:[%s]", write.Key)
		} else {
			if err := storeBytesWithVersion(batch, key, version, write.Value); err != nil {
				return err
			}
		}
	}
	err := batch.Write()
	if err != nil {
		contractAddress := common.NewAddress(id, common.ContractHash)
		log.Errorf("batch write contract(%v) state error:%s", contractAddress.Str(), err)
		return err
	}

	return nil
}

/**
获取合约全部属性
To get contract or contract template all fields
*/
func (statedb *StateDb) GetContractStatesById(id []byte) (map[string]*modules.ContractStateValue, error) {
	key := append(constants.CONTRACT_STATE_PREFIX, id...)
	data := getprefix(statedb.db, key)
	if len(data) == 0 {
		return nil, fmt.Errorf("the contract %x state is null.", id)
	}
	var err error
	result := make(map[string]*modules.ContractStateValue)
	for dbkey, state_version := range data {
		state, version, err0 := splitValueAndVersion(state_version)
		if err0 != nil {
			err = err0
		}
		realKey := dbkey[len(key):]
		if realKey != "" {
			result[realKey] = &modules.ContractStateValue{Value: state, Version: version}
			log.Debug("the contract's state get info.", "key", realKey)
		}
	}
	return result, err
}

/**
获取合约全部属性 by Prefix
To get contract or contract template all fields
*/
func (statedb *StateDb) GetContractStatesByPrefix(id []byte,
	prefix string) (map[string]*modules.ContractStateValue, error) {
	key := append(constants.CONTRACT_STATE_PREFIX, id...)
	data := getprefix(statedb.db, append(key, []byte(prefix)...))
	if len(data) == 0 {
		return nil, fmt.Errorf("the contract %x state is null.", id)
	}
	var err error
	result := make(map[string]*modules.ContractStateValue)
	for dbkey, state_version := range data {
		state, version, err0 := splitValueAndVersion(state_version)
		if err0 != nil {
			err = err0
		}
		realKey := dbkey[len(key):]
		if realKey != "" {
			result[realKey] = &modules.ContractStateValue{Value: state, Version: version}
			log.Debug("the contract's state get info.", "key", realKey)
		}
	}
	return result, err
}

/**
获取合约某一个属性
To get contract or contract template one field
*/
func (statedb *StateDb) GetContractState(id []byte, field string) ([]byte, *modules.StateVersion, error) {

	key := getContractStateKey(id, field)
	log.Debugf("DB[%s] GetContractState for key:%x. field:%s ", reflect.TypeOf(statedb.db).String(), key, field)
	data, version, err := retrieveWithVersion(statedb.db, key)
	//log.Debugf("GetContractState Result:%x,version:%s", data, version.String())
	return data, version, err
}

// GetContract can get a Contract by the contract hash

func (statedb *StateDb) SaveContractDeploy(reqid []byte, deploy *modules.ContractDeployPayload) error {
	// key: requestId
	key := append(constants.CONTRACT_DEPLOY, reqid...)
	return StoreToRlpBytes(statedb.db, key, deploy)
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
	// key : requestId
	key := append(constants.CONTRACT_DEPLOY_REQ, reqid...)
	return StoreToRlpBytes(statedb.db, key, deploy)
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

func (statedb *StateDb) GetContractInvoke(reqId []byte) (*modules.ContractInvokePayload, error) {
	key := append(constants.CONTRACT_INVOKE, reqId...)
	data, err := statedb.db.Get(key)
	if err != nil {
		return nil, err
	}
	invoke := new(modules.ContractInvokePayload)
	if err := rlp.DecodeBytes(data, &invoke); err != nil {
		return nil, err
	}
	return invoke, nil
}

func (statedb *StateDb) UpdateStateByContractInvoke(invoke *modules.ContractInvokeRequestPayload) error {
	contractAddress := common.NewAddress(invoke.ContractId, common.ContractHash)

	if contractAddress == syscontract.DepositContractAddress {
		log.Debugf("Save Deposit Contract Invoke Req")

		if string(invoke.Args[0]) == modules.ApplyMediator {
			//log.Debugf("ApplyMediator args:%s", string(invoke.Args[1]))
			mco := modules.NewMediatorCreateArgs()

			err := json.Unmarshal(invoke.Args[1], &mco)
			if err == nil {
				log.Debugf("Save Apply Mediator Invoke Req for account: (%v)", mco.AddStr)

				mi := modules.NewMediatorInfo()
				mi.MediatorInfoBase = mco.MediatorInfoBase
				mi.MediatorApplyInfo = mco.MediatorApplyInfo

				addr, err := mco.Validate()
				if err == nil {
					statedb.StoreMediatorInfo(addr, mi)
				} else {
					log.Warnf("Validate MediatorCreateArgs err: %v", err.Error())
				}
			} else {
				log.Warnf("ApplyMediator Args Unmarshal: %v", err.Error())
			}
		} else if string(invoke.Args[0]) == modules.UpdateMediatorInfo {
			//log.Debugf("UpdateMediatorInfo args:%s", string(invoke.Args[1]))
			var mua modules.MediatorUpdateArgs

			err := json.Unmarshal(invoke.Args[1], &mua)
			if err == nil {
				log.Debugf("Save Update Mediator(%v) Invoke Req", mua.AddStr)

				addr, err := mua.Validate()
				if err == nil {
					mi, err := statedb.RetrieveMediatorInfo(addr)
					if err == nil {
						if mua.Logo != nil {
							mi.Logo = *mua.Logo
						}
						if mua.Name != nil {
							mi.Name = *mua.Name
						}
						if mua.Location != nil {
							mi.Location = *mua.Location
						}
						if mua.Url != nil {
							mi.Url = *mua.Url
						}
						if mua.Description != nil {
							mi.Description = *mua.Description
						}
						if mua.Node != nil {
							mi.Node = *mua.Node
						}
						if mua.RewardAdd != nil {
							mi.RewardAdd = *mua.RewardAdd
						}
						if mua.InitPubKey != nil {
							mi.InitPubKey = *mua.InitPubKey
						}
						statedb.StoreMediatorInfo(addr, mi)
					} else {
						log.Warnf("RetrieveMediatorInfo error: %v", err.Error())
					}
				} else {
					log.Warnf("StrToMedAdd err: %v", err.Error())
				}
			} else {
				log.Warnf("UpdateMediatorInfo Args Unmarshal: %v", err.Error())
			}
		}
	}

	return nil
}

func (statedb *StateDb) SaveContractInvokeReq(reqid []byte, invoke *modules.ContractInvokeRequestPayload) error {
	contractAddress := common.NewAddress(invoke.ContractId, common.ContractHash)
	log.Debugf("save contract invoke req id(%v) contractAddress: %v, timeout: %v",
		hex.EncodeToString(reqid), contractAddress.Str(), invoke.Timeout)

	// key: reqid
	key := append(constants.CONTRACT_INVOKE_REQ, reqid...)
	return StoreToRlpBytes(statedb.db, key, invoke)
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

func (statedb *StateDb) SaveContractStop(reqid []byte, stop *modules.ContractStopPayload) error {
	// key: reqid
	key := append(constants.CONTRACT_STOP, reqid...)
	return StoreToRlpBytes(statedb.db, key, stop)
}

func (statedb *StateDb) GetContractStop(reqId []byte) (*modules.ContractStopPayload, error) {
	key := append(constants.CONTRACT_STOP, reqId...)
	data, err := statedb.db.Get(key)
	if err != nil {
		return nil, err
	}
	stop := new(modules.ContractStopPayload)
	if err := rlp.DecodeBytes(data, &stop); err != nil {
		return nil, err
	}
	return stop, nil
}

func (statedb *StateDb) SaveContractStopReq(reqid []byte, stopr *modules.ContractStopRequestPayload) error {
	// key: reqid
	key := append(constants.CONTRACT_STOP_REQ, reqid...)
	return StoreToRlpBytes(statedb.db, key, stopr)
}

func (statedb *StateDb) GetContractStopReq(reqId []byte) (*modules.ContractStopRequestPayload, error) {
	key := append(constants.CONTRACT_STOP_REQ, reqId...)
	data, err := statedb.db.Get(key)
	if err != nil {
		return nil, err
	}
	stopr := new(modules.ContractStopRequestPayload)
	if err := rlp.DecodeBytes(data, &stopr); err != nil {
		return nil, err
	}
	return stopr, nil
}
func (statedb *StateDb) SaveContractSignature(reqid []byte, sig *modules.SignaturePayload) error {
	// key: reqid
	key := append(constants.CONTRACT_SIGNATURE, reqid...)
	return StoreToRlpBytes(statedb.db, key, sig)
}

func (statedb *StateDb) GetContractSignature(reqId []byte) (*modules.SignaturePayload, error) {
	key := append(constants.CONTRACT_SIGNATURE, reqId...)
	data, err := statedb.db.Get(key)
	if err != nil {
		return nil, err
	}
	sig := new(modules.SignaturePayload)
	if err := rlp.DecodeBytes(data, &sig); err != nil {
		return nil, err
	}
	return sig, nil
}
