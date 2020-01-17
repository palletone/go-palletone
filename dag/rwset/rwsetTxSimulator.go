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

package rwset

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"

	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/modules"
)

type RwSetTxSimulator struct {
	chainIndex   *modules.ChainIndex
	requestIds         []common.Hash
	rwsetBuilder *RWSetBuilder
	write_cache  map[string][]byte //本合约的写缓存
	global_state map[string][]byte //全局合约的写缓存
	changedData []*ContractChangedData //每个Request造成的合约数据变化
	dag          IDataQuery
	//writePerformed          bool   // 没用到，注释掉
	//pvtdataQueriesPerformed bool
	doneInvoked             bool
}

//用于缓存一个合约调用结束后的状态和Token变化
type ContractChangedData struct{
	WriteState map[string]map[string][]byte //map[ContractIdHex]map[Key]value
	UsedUtxo []*modules.OutPoint
	NewUtxo map[modules.OutPoint]*modules.Utxo //因为TxHash还没有生成，所以OutPoint里面的TxHash是RequestId
}
func NewContractChangedData() *ContractChangedData{
	return &ContractChangedData{
		WriteState:make(map[string]map[string][]byte),
		UsedUtxo:make([]*modules.OutPoint,0,0),
		NewUtxo:make(map[modules.OutPoint]*modules.Utxo),
	}
}
func (d *ContractChangedData) GetState(contractid []byte, key string) []byte{
	contractHex:=hex.EncodeToString(contractid)
	kvMap,ok:= d.WriteState[contractHex]
	if ok{
		return kvMap[key]
	}
	return nil
}

func (d *ContractChangedData) SetState(contractid []byte, key string,value []byte){
	contractHex:=hex.EncodeToString(contractid)
	kvMap,ok:= d.WriteState[contractHex]
	if ok{
		 kvMap[key]=value
		 return
	}
	newKv:=make(map[string][]byte)
	newKv[key]=value
	d.WriteState[contractHex]=newKv
}
func (d *ContractChangedData) SpendUtxo(op *modules.OutPoint) {
	d.UsedUtxo=append(d.UsedUtxo,op)
}
func (d *ContractChangedData) AddNewUtxo(newUtxos map[modules.OutPoint]*modules.Utxo) {
	for o,u:=range newUtxos{
		d.NewUtxo[o]=u
	}
}
//type VersionedValue struct {
//	Value   []byte
//	Version *Version
//}
func IsGlobalStateContract(contractId []byte) bool{
	return bytes.Equal(contractId,[]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
}

func NewBasedTxSimulator(idag IDataQuery) *RwSetTxSimulator {
	rwsetBuilder := NewRWSetBuilder()
	gasToken := dagconfig.DagConfig.GetGasToken()
	//unit := idag.GetCurrentUnit(gasToken)
	//cIndex := unit.Header().GetNumber()
	ustabeUnit, _ := idag.UnstableHeadUnitProperty(gasToken)
	cIndex := ustabeUnit.ChainIndex
	return &RwSetTxSimulator{chainIndex: modules.NewChainIndex(cIndex.AssetID, cIndex.Index), requestIds: []common.Hash{},
		rwsetBuilder: rwsetBuilder, write_cache: make(map[string][]byte), dag: idag}
}

//func (s *RwSetTxSimulator) GetChainParameters() ([]byte, error) {
//	cp := s.dag.GetChainParameters()
//
//	data, err := rlp.EncodeToBytes(cp)
//	if err != nil {
//		return nil, err
//	}
//
//	return data, nil
//}

func (s *RwSetTxSimulator) GetGlobalProp() ([]byte, error) {
	gp := s.dag.GetGlobalProp()

	data, err := rlp.EncodeToBytes(gp)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// GetState implements method in interface `ledger.TxSimulator`
func (s *RwSetTxSimulator) GetState(contractid []byte, ns string, key string) ([]byte, error) {
	//testValue := []byte("abc")
	if err := s.CheckDone(); err != nil {
		return nil, err
	}
	cacheState:=s.global_state
	if !IsGlobalStateContract(contractid){
		cacheState=s.write_cache
	}
		if value, has := cacheState[key]; has {
			if s.rwsetBuilder != nil {
				s.rwsetBuilder.AddToReadSet(contractid, ns, key, nil)
			}
			return value, nil
		}

	val, ver, err := s.dag.GetContractState(contractid, key)
	//TODO 这里证明数据库里面没有该账户信息，需要返回nil,nil
	if err != nil {
		log.Debugf("get value from db[%s] failed,key:%s", ns, key)
		return nil, nil
		//errstr := fmt.Sprintf("GetContractState [%s]-[%s] failed", ns, key)
		//		//return nil, errors.New(errstr)
	}
	if s.rwsetBuilder != nil {
		s.rwsetBuilder.AddToReadSet(contractid, ns, key, ver)
	}
	log.Debugf("RW:GetState,ns[%s]--key[%s]---value[%s]---ver[%v]", ns, key, val, ver)

	//TODO change.
	//return testValue, nil
	return val, nil
}
func (s *RwSetTxSimulator) GetStatesByPrefix(contractid []byte, ns string, prefix string) ([]*modules.KeyValue, error) {
	if err := s.CheckDone(); err != nil {
		return nil, err
	}

	data, err := s.dag.GetContractStatesByPrefix(contractid, prefix)

	if err != nil {
		log.Debugf("get value from db[%s] failed,prefix:%s,error:[%s]", ns, prefix, err.Error())
		return nil, nil
		//errstr := fmt.Sprintf("GetContractState [%s]-[%s] failed", ns, key)
		//		//return nil, errors.New(errstr)
	}
	result := []*modules.KeyValue{}
	for key, row := range data {
		kv := &modules.KeyValue{Key: key, Value: row.Value}
		result = append(result, kv)
		if s.rwsetBuilder != nil {
			s.rwsetBuilder.AddToReadSet(contractid, ns, key, row.Version)
		}
	}

	//log.Debugf("RW:GetStatesByPrefix,ns[%s]--contractid[%x]---prefix[%s]", ns, contractid, prefix)

	return result, nil
}

// GetState implements method in interface `ledger.TxSimulator`
func (s *RwSetTxSimulator) GetTimestamp(ns string, rangeNumber uint32) ([]byte, error) {
	//testValue := []byte("abc")
	if err := s.CheckDone(); err != nil {
		return nil, err
	}
	gasToken := dagconfig.DagConfig.GetGasToken()
	header := s.dag.CurrentHeader(gasToken)
	timeIndex := header.GetNumber().Index / uint64(rangeNumber) * uint64(rangeNumber)
	timeHeader, err := s.dag.GetHeaderByNumber(&modules.ChainIndex{AssetID: header.GetNumber().AssetID, Index: timeIndex})
	if err != nil {
		return nil, errors.New("GetHeaderByNumber failed" + err.Error())
	}

	return []byte(fmt.Sprintf("%d", timeHeader.Timestamp())), nil
}
func (s *RwSetTxSimulator) SetState(contractId []byte, ns string, key string, value []byte) error {
	if err := s.CheckDone(); err != nil {
		return err
	}

	//todo ValidateKeyValue
	s.rwsetBuilder.AddToWriteSet(contractId, ns, key, value)

	cacheState:=s.global_state
	if !IsGlobalStateContract(contractId){
		cacheState=s.write_cache
	}
	cacheState[key] = value
	return nil
}

// DeleteState implements method in interface `ledger.TxSimulator`
func (s *RwSetTxSimulator) DeleteState(contractId []byte, ns string, key string) error {
	return s.SetState(contractId, ns, key, nil)
	//TODO Devin
}

func (s *RwSetTxSimulator) GetRwData(ns string) ([]*KVRead, []*KVWrite, error) {
	rd := make(map[string]*KVRead)
	wt := make(map[string]*KVWrite)
	log.Debug("GetRwData", "ns info", ns)

	if s.rwsetBuilder != nil {
		s.rwsetBuilder.locker.RLock()
		if s.rwsetBuilder.pubRwBuilderMap != nil {
			if s.rwsetBuilder.pubRwBuilderMap[ns] != nil {
				pubRwBuilderMap, ok := s.rwsetBuilder.pubRwBuilderMap[ns]
				if ok {
					rd = pubRwBuilderMap.readMap
					wt = pubRwBuilderMap.writeMap
				} else {
					s.rwsetBuilder.locker.RUnlock()
					return nil, nil, errors.New("rw_data not found.")
				}
			}
		}
		s.rwsetBuilder.locker.RUnlock()
	}
	//sort keys and convert map to slice
	return convertReadMap2Slice(rd), convertWriteMap2Slice(wt), nil
}
func convertReadMap2Slice(rd map[string]*KVRead) []*KVRead {
	keys := make([]string, 0)
	for k := range rd {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	result := make([]*KVRead, 0)
	for _, key := range keys {
		result = append(result, rd[key])
	}
	return result
}
func convertWriteMap2Slice(rd map[string]*KVWrite) []*KVWrite {
	keys := make([]string, 0)
	for k := range rd {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	result := make([]*KVWrite, 0)
	for _, key := range keys {
		result = append(result, rd[key])
	}
	return result
}

//get all dag
func (s *RwSetTxSimulator) GetContractStatesById(contractid []byte) (map[string]*modules.ContractStateValue, error) {
	return s.dag.GetContractStatesById(contractid)
}

func (s *RwSetTxSimulator) CheckDone() error {
	if s.doneInvoked {
		return errors.New("This instance should not be used after calling Done()")
	}
	return nil
}
func (h *RwSetTxSimulator) Done() {
	if h.doneInvoked {
		return
	}
	h.Close()
	h.doneInvoked = true
}
func (s *RwSetTxSimulator) Close() {
	item := new(RwSetTxSimulator)
	s.chainIndex = item.chainIndex
	//s.txid = item.txid
	s.rwsetBuilder = item.rwsetBuilder
	s.write_cache = item.write_cache
	s.dag = item.dag
}

//func (h *RwSetTxSimulator) GetTxSimulationResults() ([]byte, error) {
//
//	return nil, nil
//}

func (s *RwSetTxSimulator) GetTokenBalance(ns string, addr common.Address, asset *modules.Asset) (
	map[modules.Asset]uint64, error) {
	var utxos map[modules.OutPoint]*modules.Utxo
	if asset == nil {
		utxos, _ = s.dag.GetAddrUtxos(addr)
	} else {
		utxos, _ = s.dag.GetAddr1TokenUtxos(addr, asset)
	}
	return convertUtxo2Balance(utxos), nil
}

func (s *RwSetTxSimulator) GetStableTransactionByHash(ns string, hash common.Hash) (*modules.Transaction, error) {
	return s.dag.GetStableTransactionOnly(hash)
}

func (s *RwSetTxSimulator) GetStableUnit(ns string, hash common.Hash, unitNumber uint64) (*modules.Unit, error) {
	if !hash.IsZero() {
		return s.dag.GetStableUnit(hash)
	}
	gasToken := dagconfig.DagConfig.GetGasToken()
	number := &modules.ChainIndex{AssetID: gasToken, Index: unitNumber}
	return s.dag.GetStableUnitByNumber(number)
}
func convertUtxo2Balance(utxos map[modules.OutPoint]*modules.Utxo) map[modules.Asset]uint64 {
	result := map[modules.Asset]uint64{}
	for _, v := range utxos {
		if val, ok := result[*v.Asset]; ok {
			result[*v.Asset] = val + v.Amount
		} else {
			result[*v.Asset] = v.Amount
		}
	}
	return result
}
func (s *RwSetTxSimulator) PayOutToken(ns string, address string, token *modules.Asset, amount uint64,
	lockTime uint32) error {
	s.rwsetBuilder.AddTokenPayOut(ns, address, token, amount, lockTime)
	return nil
}
func (s *RwSetTxSimulator) GetPayOutData(ns string) ([]*modules.TokenPayOut, error) {
	return s.rwsetBuilder.GetTokenPayOut(ns), nil
}
func (s *RwSetTxSimulator) GetTokenDefineData(ns string) (*modules.TokenDefine, error) {
	return s.rwsetBuilder.GetTokenDefine(ns), nil
}
func (s *RwSetTxSimulator) GetTokenSupplyData(ns string) ([]*modules.TokenSupply, error) {
	return s.rwsetBuilder.GetTokenSupply(ns), nil
}
func (s *RwSetTxSimulator) DefineToken(ns string, tokenType int32, define []byte, creator string) error {
	createAddr, _ := common.StringToAddress(creator)
	s.rwsetBuilder.DefineToken(ns, tokenType, define, createAddr)
	return nil
}
func (s *RwSetTxSimulator) SupplyToken(ns string, assetId, uniqueId []byte, amt uint64, creator string) error {
	createAddr, _ := common.StringToAddress(creator)
	return s.rwsetBuilder.AddSupplyToken(ns, assetId, uniqueId, amt, createAddr)
}
func (s *RwSetTxSimulator) String() string {
	str := "rwSet_txSimulator: "
	for k, v := range s.rwsetBuilder.pubRwBuilderMap {
		str += "key:" + k
		for rk, rv := range v.readMap {
			//str += fmt.Sprintf("val__[key:%s],[value:%s]", rk, rv.String())
			log.Debug("RwSetTxSimulator) String", "key", rk, "val-", rv)
		}
	}
	return str
}
