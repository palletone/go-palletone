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
 *  * @author PalletOne core developers <dev@pallet.one>
 *  * @date 2018
 *
 */

package ptnapi

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/hexutil"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/ptnjson"
	"github.com/shopspring/decimal"
	"strings"
)

type PublicDagAPI struct {
	b Backend
	// d *dag
}

func NewPublicDagAPI(b Backend) *PublicDagAPI {
	return &PublicDagAPI{b}
}

//所有可能大数据量的查询只能在本地查询
type PrivateDagAPI struct {
	b Backend
}

func NewPrivateDagAPI(b Backend) *PrivateDagAPI {
	return &PrivateDagAPI{b}
}
func (s *PublicDagAPI) GetHexCommon(ctx context.Context, key string) (string, error) {
	key_bytes, err0 := hexutil.Decode(key)
	if err0 != nil {
		log.Info("getCommon by Hex error", "error", err0)
		return "", err0
	}
	//log.Info("GetCommon by hex info.", "key", string(key_bytes), "bytes", key_bytes)
	items, err := s.b.GetCommon(key_bytes[:])
	if err != nil {
		return "", err
	}
	hex := hexutil.Encode(items)
	return hex, nil
}
func (s *PublicDagAPI) GetCommon(ctx context.Context, key string) ([]byte, error) {
	// key to bytes
	if key == "" {
		return nil, fmt.Errorf("参数为空")
	}
	return s.b.GetCommon([]byte(key))
}

func (s *PrivateDagAPI) GetCommonByPrefix(ctx context.Context, prefix string) (string, error) {
	if prefix == "" {
		return "", fmt.Errorf("参数为空")
	}
	result := s.b.GetCommonByPrefix([]byte(prefix))
	if len(result) == 0 {
		return "all_items:null", nil
	}

	info := NewPublicReturnInfo("all_items", result)
	result_json, err := json.Marshal(info)
	return string(result_json), err
}
func (s *PublicDagAPI) GetHeaderByHash(ctx context.Context, condition string) (string, error) {
	hash := common.Hash{}
	if err := hash.SetHexString(condition); err != nil {
		log.Info("PublicBlockChainAPI", "GetUnitByHash SetHexString err:", err, "condition:", condition)
		return "", err
	}
	header, err := s.b.GetHeaderByHash(hash)
	if err != nil {
		log.Info("PublicBlockChainAPI", "GetHeaderByHash err:", err, "hash", hash.String())
		return "", err
	}
	headerJson := ptnjson.ConvertUnitHeader2Json(header)
	headerRlp, _ := rlp.EncodeToBytes(header)
	info := NewPublicReturnInfoWithHex("header", headerJson, headerRlp)
	content, err := json.Marshal(info)
	if err != nil {
		log.Info("PublicBlockChainAPI", "GetHeaderByHash Marshal err:", err, "hash", hash.String())
		return "info Marshal err", err
	}
	return string(content), nil
}
func (s *PublicDagAPI) GetHeaderByNumber(ctx context.Context, condition string) (string, error) {
	number := &modules.ChainIndex{}
	index, err := strconv.ParseInt(condition, 10, 64)
	if err != nil {
		log.Info("PublicBlockChainAPI", "GetHeaderByNumber strconv.ParseInt err:", err, "condition:", condition)
		return "", err
	}
	number.Index = uint64(index)
	number.AssetID = dagconfig.DagConfig.GetGasToken()
	header, err := s.b.GetHeaderByNumber(number)
	if err != nil {
		return "", err
	}
	headerRlp, err := rlp.EncodeToBytes(header)
	if err != nil {
		log.Info("PublicBlockChainAPI", "GetHeaderByNumber err:", err, "number", number.String())
	}
	headerJson := ptnjson.ConvertUnitHeader2Json(header)
	info := NewPublicReturnInfoWithHex("header", headerJson, headerRlp)
	content, err := json.Marshal(info)
	if err != nil {
		log.Info("PublicBlockChainAPI", "GetHeaderByNumber Marshal err:", err, "number", number.String())
	}
	return string(content), nil
}

func (s *PublicDagAPI) GetUnitByHash(ctx context.Context, condition string) string {
	log.Info("PublicDagAPI", "GetUnitByHash condition:", condition)
	hash := common.Hash{}
	if err := hash.SetHexString(condition); err != nil {
		log.Info("PublicBlockChainAPI", "GetUnitByHash SetHexString err:", err, "condition:", condition)
		return ""
	}
	unit := s.b.GetUnitByHash(hash)
	if unit == nil {
		log.Info("PublicBlockChainAPI", "GetUnitByHash GetUnitByHash is nil hash:", hash)
		return "GetUnitByHash nil"
	}
	jsonUnit := ptnjson.ConvertUnit2Json(unit, s.b.Dag().GetTxOutput)
	content, err := json.Marshal(jsonUnit)
	if err != nil {
		log.Info("PublicBlockChainAPI", "GetUnitByHash Marshal err:", err, "unit:", *unit)
		return "jsonUnit Marshal err"
	}
	return string(content)
}

func (s *PublicDagAPI) GetUnitByNumber(ctx context.Context, condition string) string {
	log.Info("PublicDagAPI", "GetUnitByNumber condition:", condition)

	number := &modules.ChainIndex{}
	index, err := strconv.ParseInt(condition, 10, 64)
	if err != nil {
		log.Info("PublicBlockChainAPI", "GetUnitByNumber strconv.ParseInt err:", err, "condition:", condition)
		return ""
	}
	number.Index = uint64(index)

	number.AssetID = dagconfig.DagConfig.GetGasToken()
	log.Info("PublicBlockChainAPI info", "GetUnitByNumber_number.Index:", number.Index, "number:", number.String())

	unit := s.b.GetUnitByNumber(number)
	if unit == nil {
		log.Info("PublicBlockChainAPI", "GetUnitByNumber GetUnitByNumber is nil number:", number)
		return "GetUnitByNumber nil"
	}
	jsonUnit := ptnjson.ConvertUnit2Json(unit, s.b.Dag().GetTxOutput)
	content, err := json.Marshal(jsonUnit)
	if err != nil {
		log.Info("PublicBlockChainAPI", "GetUnitByNumber Marshal err:", err, "unit:", *unit)
		return "json UnitMarshal err"
	}
	return string(content)
}

// getUnitsByIndex
func (s *PublicDagAPI) GetUnitsByIndex(ctx context.Context, start, end decimal.Decimal, asset string) string {
	log.Info("PublicDagAPI ,GetUnitsByIndexs:", "start", start, "end", end)
	units := s.b.GetUnitsByIndex(start, end, asset)
	jsonUnits := make([]*ptnjson.UnitJson, 0)

	for _, u := range units {
		jsonu := ptnjson.ConvertUnit2Json(u, s.b.Dag().GetTxOutput)
		jsonUnits = append(jsonUnits, jsonu)
	}
	info := NewPublicReturnInfo("units", jsonUnits)
	content, err := json.Marshal(info)
	if err != nil {
		log.Info("PublicBlockChainAPI", "GetUnitsByIndexs Marshal err:", err)
	}
	return string(content)
}

func (s *PublicDagAPI) GetFastUnitIndex(ctx context.Context, assetid string) string {
	log.Debug("PublicDagAPI", "GetUnitByNumber condition:", assetid)
	if assetid == "" {
		assetid = "PTN"
	}
	assetid = strings.ToUpper(assetid)
	token, _, err := modules.String2AssetId(assetid)
	if err != nil {
		return "unknow assetid:" + assetid + ". " + err.Error()
	}
	if assetid != "PTN" {
		GlobalStateContractId := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
		val, _, err := s.b.GetContractState(GlobalStateContractId, modules.GlobalPrefix+strings.ToUpper(token.GetSymbol()))
		if err != nil || len(val) == 0 {
			return "unknow assetid: " + assetid + ", " + err.Error()
		}
	}
	stableUnit := s.b.Dag().CurrentUnit(token)
	ustabeUnit := s.b.Dag().GetCurrentMemUnit(token, 0)
	result := new(ptnjson.FastUnitJson)
	if ustabeUnit != nil {
		result.FastHash = ustabeUnit.UnitHash
		result.FastIndex = ustabeUnit.NumberU64()
	}
	if stableUnit != nil {
		result.StableHash = stableUnit.UnitHash
		result.StableIndex = stableUnit.NumberU64()
	}
	content, err := json.Marshal(result)
	if err != nil {
		log.Info("PublicDagAPI", "GetFastUnitIndex Marshal err:", err)
		return "result Marshal err"
	}
	return string(content)
}
func (s *PublicDagAPI) GetUnitSummaryByNumber(ctx context.Context, condition string) string {
	log.Info("PublicBlockChainAPI", "GetUnitByNumber condition:", condition)

	number := &modules.ChainIndex{}
	index, err := strconv.ParseInt(condition, 10, 64)
	if err != nil {
		log.Info("PublicBlockChainAPI", "GetUnitByNumber strconv.ParseInt err:", err, "condition:", condition)
		return ""
	}
	number.Index = uint64(index)

	number.AssetID = dagconfig.DagConfig.GetGasToken()
	log.Info("PublicBlockChainAPI info", "GetUnitByNumber_number.Index:", number.Index, "number:", number.String())

	unit := s.b.GetUnitByNumber(number)
	if unit == nil {
		log.Info("PublicBlockChainAPI", "GetUnitByNumber GetUnitByNumber is nil number:", number)
		return "GetUnitByNumber nil"
	}
	jsonUnit := ptnjson.ConvertUnit2SummaryJson(unit)
	content, err := json.Marshal(jsonUnit)
	if err != nil {
		log.Info("PublicBlockChainAPI", "GetUnitByNumber Marshal err:", err, "unit:", *unit)
		return "json Unit Marshal err"
	}
	return string(content)
}

func (s *PublicDagAPI) GetUnstableUnits() []*ptnjson.UnitSummaryJson {
	return s.b.GetUnstableUnits()
}
func (s *PublicDagAPI) GetUnitTxsInfo(ctx context.Context, hashHex string) (string, error) {
	hash := common.HexToHash(hashHex)
	if item, err := s.b.GetUnitTxsInfo(hash); err != nil {
		return "unit_txs:null", err
	} else {
		info := NewPublicReturnInfo("unit_txs", item)
		result_json, _ := json.Marshal(info)
		return string(result_json), nil
	}
}

func (s *PublicDagAPI) GetUnitTxsHashHex(ctx context.Context, hashHex string) (string, error) {
	hash := common.HexToHash(hashHex)

	if item, err := s.b.GetUnitTxsHashHex(hash); err != nil {
		return "unit_txs_hash:null", err
	} else {
		info := NewPublicReturnInfo("unit_txs_hash", item)
		result_json, _ := json.Marshal(info)
		return string(result_json), nil
	}
}

func (s *PublicDagAPI) GetTxByHash(ctx context.Context, hashHex string) (string, error) {
	hash := common.HexToHash(hashHex)
	if item, err := s.b.GetTxByHash(hash); err != nil {
		return "transaction_info:null", err
	} else {
		info := NewPublicReturnInfo("transaction_info", item)
		result_json, _ := json.Marshal(info)
		return string(result_json), nil
	}
}
func (s *PublicDagAPI) GetTxByReqId(ctx context.Context, hashHex string) (string, error) {
	hash := common.HexToHash(hashHex)
	if item, err := s.b.GetTxByReqId(hash); err != nil {
		return "transaction_info:null", err
	} else {
		info := NewPublicReturnInfo("transaction_info", item)
		result_json, _ := json.Marshal(info)
		return string(result_json), nil
	}
}
func (s *PublicDagAPI) GetTxSearchEntry(ctx context.Context, hashHex string) (string, error) {
	hash := common.HexToHash(hashHex)
	item, err := s.b.GetTxSearchEntry(hash)

	info := NewPublicReturnInfo("tx_entry", item)
	result_json, _ := json.Marshal(info)
	return string(result_json), err
}

func (s *PublicDagAPI) GetAddrOutpoints(ctx context.Context, addr string) (string, error) {
	items, err := s.b.GetAddrOutpoints(addr)
	if err != nil {
		return "", err
	}
	info := NewPublicReturnInfo("address_outpoints", items)
	result_json, _ := json.Marshal(info)
	return string(result_json), nil
}

func (s *PublicDagAPI) GetAddrUtxos(ctx context.Context, addr string) (string, error) {
	items, err := s.b.GetAddrUtxos(addr)

	if err != nil {
		return "", err
	}
	info := NewPublicReturnInfo("address_utxos", items)
	result_json, _ := json.Marshal(info)
	return string(result_json), nil
}

func (s *PublicDagAPI) GetTransactionsByTxid(ctx context.Context, txid string) (*ptnjson.GetTxIdResult, error) {
	tx, err := s.b.GetTxByTxid_back(txid)
	if err != nil {
		log.Error("Get transcation by hash ", "unit_hash", txid, "error", err.Error())
		return nil, err
	}
	return tx, nil
}

func (s *PublicDagAPI) GetTxHashByReqId(ctx context.Context, hashHex string) (string, error) {
	hash := common.HexToHash(hashHex)
	item, err := s.b.GetTxHashByReqId(hash)

	info := NewPublicReturnInfo("tx_hash", item)
	result_json, _ := json.Marshal(info)
	return string(result_json), err
}

// GetTxPoolTxByHash returns the pool transaction for the given hash
func (s *PublicDagAPI) GetTxPoolTxByHash(ctx context.Context, hex string) (string, error) {
	log.Debug("this is hash tx's hash hex to find tx.", "hex", hex)
	hash := common.HexToHash(hex)
	log.Debug("this is hash tx's hash  to find tx.", "hash", hash.String())
	item, err := s.b.GetTxPoolTxByHash(hash)
	if err != nil {
		return "pool_tx:null", err
	} else {
		info := NewPublicReturnInfo("txpool_tx", item)
		result_json, _ := json.Marshal(info)
		return string(result_json), nil
	}
}

func (s *PublicDagAPI) HeadUnitHash() string {
	dag := s.b.Dag()
	if dag != nil {
		return dag.HeadUnitHash().String()
	}

	return "unknown"
}

func (s *PublicDagAPI) HeadUnitTime() string {
	dag := s.b.Dag()
	if dag != nil {
		time := time.Unix(dag.HeadUnitTime(), 0)
		return time.Format("2006-01-02 15:04:05 -0700 MST")
	}

	return "1972-0-0"
}

func (s *PublicDagAPI) HeadUnitNum() uint64 {
	dag := s.b.Dag()
	if dag != nil {
		return dag.HeadUnitNum()
	}
	return uint64(0)
}

func (s *PublicDagAPI) StableUnitNum() uint64 {
	dag := s.b.Dag()
	if dag != nil {
		gasToken := dagconfig.DagConfig.GetGasToken()
		return dag.GetIrreversibleUnitNum(gasToken)
	}

	return uint64(0)
}

func (s *PublicDagAPI) IsSynced() bool {
	dag := s.b.Dag()
	if dag != nil {
		return dag.IsSynced()
	}

	return false
}

func (s *PrivateDagAPI) GetAllUtxos(ctx context.Context) (string, error) {
	items, err := s.b.GetAllUtxos()
	if err != nil {
		log.Error("Get all utxo failed.", "error", err, "result", items)
		return "", err
	}

	info := NewPublicReturnInfo("all_utxos", items)

	result_json, err := json.Marshal(info)
	if err != nil {
		log.Error("Get all utxo ,json marshal failed.", "error", err)
	}

	return string(result_json), nil
}
func (s *PrivateDagAPI) CheckHeader(ctx context.Context, number int) (bool, error) {
	dag := s.b.Dag()
	err := dag.CheckHeaderCorrect(number)
	if err != nil {
		return false, err
	}
	return true, nil
}
func (s *PrivateDagAPI) RebuildAddrTxIndex() error {
	dag := s.b.Dag()
	return dag.RebuildAddrTxIndex()
}
