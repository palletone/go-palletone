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
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/dag"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/modules"
	"runtime"
)

type RwSetTxSimulator struct {
	chainIndex              *modules.ChainIndex
	txid                    common.Hash
	rwsetBuilder            *RWSetBuilder
	dag                     dag.IDag
	writePerformed          bool
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
	return &RwSetTxSimulator{cIndex, hash, rwsetBuilder, idag, false, false, false}
}

func (s *RwSetTxSimulator) GetConfig(name string) ([]byte, error) {
	val, _, err := s.dag.GetConfig(name)
	if err != nil {
		return nil, err
	}
	return val, nil
}

// GetState implements method in interface `ledger.TxSimulator`
func (s *RwSetTxSimulator) GetState(contractid []byte, ns string, key string) ([]byte, error) {
	//testValue := []byte("abc")
	if err := s.CheckDone(); err != nil {
		return nil, err
	}
	//TODO Devin
	val, ver, err := s.dag.GetContractState(contractid, key)
	//TODO 这里证明数据库里面没有该账户信息，需要返回nil,nil
	if err != nil {
		log.Debugf("get value from db[%s] failed,key:%s", ns, key)
		return nil, nil
		//errstr := fmt.Sprintf("GetContractState [%s]-[%s] failed", ns, key)
		//		//return nil, errors.New(errstr)
	}
	//val, ver := decomposeVersionedValue(versionedValue)
	if s.rwsetBuilder != nil {
		s.rwsetBuilder.AddToReadSet(ns, key, ver)
	}
	log.Debugf("RW:GetState,ns[%s]--key[%s]---value[%s]", ns, key, val)

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
		log.Debugf("get value from db[%s] failed,prefix:%s", ns, prefix)
		return nil, nil
		//errstr := fmt.Sprintf("GetContractState [%s]-[%s] failed", ns, key)
		//		//return nil, errors.New(errstr)
	}
	result := []*modules.KeyValue{}
	for key, row := range data {
		kv := &modules.KeyValue{Key: key, Value: row.Value}
		result = append(result, kv)
		if s.rwsetBuilder != nil {
			s.rwsetBuilder.AddToReadSet(ns, key, row.Version)
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
func (s *RwSetTxSimulator) SetState(ns string, key string, value []byte) error {
	//log.Debugf("RW:SetState,ns[%s]--key[%s]---value[%s]", ns, key, value)
	//fmt.Println("SetState(ns string, key string, value []byte)===>>>\n\n", ns, key, value)
	//balance := &modules.DepositBalance{}
	//_ = json.Unmarshal(value, balance)
	//fmt.Printf("llllllll   %#v\n", stateValue)
	if err := s.CheckDone(); err != nil {
		return err
	}
	if s.pvtdataQueriesPerformed {
		return errors.New("pvtdata Queries Performed")
	}
	//todo ValidateKeyValue
	s.rwsetBuilder.AddToWriteSet(ns, key, value)
	return nil
}

// DeleteState implements method in interface `ledger.TxSimulator`
func (s *RwSetTxSimulator) DeleteState(ns string, key string) error {
	return s.SetState(ns, key, nil)
}

func (s *RwSetTxSimulator) GetRwData(ns string) (map[string]*KVRead, map[string]*KVWrite, error) {
	var rd map[string]*KVRead
	var wt map[string]*KVWrite

	log.Info("GetRwData", "ns info", ns)

	if s.rwsetBuilder != nil {
		if s.rwsetBuilder.pubRwBuilderMap != nil {
			if s.rwsetBuilder.pubRwBuilderMap[ns] != nil {
				if s.rwsetBuilder.pubRwBuilderMap[ns].readMap != nil {
					rd = s.rwsetBuilder.pubRwBuilderMap[ns].readMap
				}
				if s.rwsetBuilder.pubRwBuilderMap[ns].writeMap != nil {
					wt = s.rwsetBuilder.pubRwBuilderMap[ns].writeMap
				}
				pubRwBuilderMap, ok := s.rwsetBuilder.pubRwBuilderMap[ns]
				if ok {
					rd = pubRwBuilderMap.readMap
					wt = pubRwBuilderMap.writeMap
				} else {
					rd = nil
					wt = nil
				}
			}
		}
	}

	return rd, wt, nil
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
	h.doneInvoked = true
	h.Close()
}
func (s *RwSetTxSimulator) Close() {
	s.dag.Close()
	runtime.SetFinalizer(s, func(item *RwSetTxSimulator) {
		if len(item.txid) > 0 || item.txid != (common.Hash{}) {
			log.Infof("free txid[%s]", item.txid.String())
			for _, b := range item.txid {
				common.Free(&b)
			}
		}
		if item.chainIndex != nil {
			for _, b := range item.chainIndex.AssetID.Bytes() {
				common.Free(&b)
			}
			//common.Free(byte(&item.chainIndex.Index))
		}

	})

	return
}

func decomposeVersionedValue(versionedValue *VersionedValue) ([]byte, *Version) {
	var value []byte
	var ver *Version
	if versionedValue != nil {
		value = versionedValue.Value
		ver = versionedValue.Version
	}
	return value, ver
}

func (h *RwSetTxSimulator) GetTxSimulationResults() ([]byte, error) {

	return nil, nil
}

func (s *RwSetTxSimulator) GetTokenBalance(ns string, addr common.Address, asset *modules.Asset) (map[modules.Asset]uint64, error) {
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
func (s *RwSetTxSimulator) PayOutToken(ns string, address string, token *modules.Asset, amount uint64, lockTime uint32) error {
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
