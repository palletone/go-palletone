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
	"errors"
	"fmt"
	"sort"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/dag"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/modules"
)

type RwSetTxSimulator struct {
	chainIndex   *modules.ChainIndex
	txid         common.Hash
	rwsetBuilder *RWSetBuilder
	write_cache  map[string][]byte
	dag          dag.IDag
	//writePerformed          bool   // 没用到，注释掉
	pvtdataQueriesPerformed bool
	doneInvoked             bool
}

type VersionedValue struct {
	Value   []byte
	Version *Version
}

func NewBasedTxSimulator(idag dag.IDag, hash common.Hash) *RwSetTxSimulator {
	rwsetBuilder := NewRWSetBuilder()
	gasToken := dagconfig.DagConfig.GetGasToken()
	unit := idag.GetCurrentUnit(gasToken)
	cIndex := unit.Header().Number
	log.Debugf("constructing new tx simulator txid = [%s]", hash.String())
	return &RwSetTxSimulator{chainIndex: cIndex, txid: hash, rwsetBuilder: rwsetBuilder,
		write_cache: make(map[string][]byte), dag: idag}
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
	if value, has := s.write_cache[key]; has {
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

	log.Debugf("RW:GetStatesByPrefix,ns[%s]--contractid[%x]---prefix[%s]", ns, contractid, prefix)

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
	timeIndex := header.Number.Index / uint64(rangeNumber) * uint64(rangeNumber)
	timeHeader, err := s.dag.GetHeaderByNumber(&modules.ChainIndex{AssetID: header.Number.AssetID, Index: timeIndex})
	if err != nil {
		return nil, errors.New("GetHeaderByNumber failed" + err.Error())
	}

	return []byte(fmt.Sprintf("%d", timeHeader.Time)), nil
}
func (s *RwSetTxSimulator) SetState(contractId []byte, ns string, key string, value []byte) error {
	if err := s.CheckDone(); err != nil {
		return err
	}
	if s.pvtdataQueriesPerformed {
		return errors.New("pvtdata Queries Performed")
	}
	//todo ValidateKeyValue
	s.rwsetBuilder.AddToWriteSet(contractId, ns, key, value)
	s.write_cache[key] = value
	return nil
}

// DeleteState implements method in interface `ledger.TxSimulator`
func (s *RwSetTxSimulator) DeleteState(contractId []byte, ns string, key string) error {
	return s.SetState(contractId, ns, key, nil)
}

func (s *RwSetTxSimulator) GetRwData(ns string) ([]*KVRead, []*KVWrite, error) {
	rd := make(map[string]*KVRead)
	wt := make(map[string]*KVWrite)
	log.Info("GetRwData", "ns info", ns)

	if s.rwsetBuilder != nil {
		if s.rwsetBuilder.pubRwBuilderMap != nil {
			if s.rwsetBuilder.pubRwBuilderMap[ns] != nil {
				pubRwBuilderMap, ok := s.rwsetBuilder.pubRwBuilderMap[ns]
				if ok {
					rd = pubRwBuilderMap.readMap
					wt = pubRwBuilderMap.writeMap
				} else {
					return nil, nil, errors.New("rw_data not found.")
				}
			}
		}
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
	s.txid = item.txid
	s.rwsetBuilder = item.rwsetBuilder
	s.write_cache = item.write_cache
	s.dag = item.dag
}

func (h *RwSetTxSimulator) GetTxSimulationResults() ([]byte, error) {

	return nil, nil
}

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
		str += ("key:" + k)
		for rk, rv := range v.readMap {
			//str += fmt.Sprintf("val__[key:%s],[value:%s]", rk, rv.String())
			log.Debug("RwSetTxSimulator) String", "key", rk, "val-", rv)
		}
	}
	return str
}
