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
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/hexutil"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/util"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/ptnjson"
	"github.com/shopspring/decimal"
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
	items, err := s.b.GetCommon(key_bytes[:], false)
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
	return s.b.GetCommon([]byte(key), false)
}
func (s *PublicDagAPI) GetLdbCommon(ctx context.Context, key string) ([]byte, error) {
	// key to bytes
	if key == "" {
		return nil, fmt.Errorf("参数为空")
	}
	return s.b.GetCommon([]byte(key), true)
}
func (s *PrivateDagAPI) GetCommonByPrefix(ctx context.Context, prefix string) (string, error) {
	if prefix == "" {
		return "", fmt.Errorf("参数为空")
	}
	result := s.b.GetCommonByPrefix([]byte(prefix), false)
	if len(result) == 0 {
		return "all_items:null", nil
	}

	info := NewPublicReturnInfo("all_items", result)
	result_json, err := json.Marshal(info)
	return string(result_json), err
}

func (s *PublicDagAPI) GetGenesisData(ctx context.Context) (*GenesisData, error) {
	data := new(GenesisData)
	keys_byte, values_byte := s.b.GetAllData()
	data.Count = len(keys_byte)
	log.Debugf("count:%d, keys:%v", data.Count, keys_byte)
	log.Debugf("count:%d, values:%v", len(values_byte), values_byte)
	if data.Count != len(values_byte) {
		return nil, fmt.Errorf("the keys count[%d] not match the values[%d].", data.Count, len(values_byte))
	}
	for i := 0; i < data.Count; i++ {
		data.Keys = append(data.Keys, util.Bytes2Hex(keys_byte[i]))
		data.Values = append(data.Values, util.Bytes2Hex(values_byte[i]))
	}
	return data, nil
}

type GenesisData struct {
	Keys   []string
	Values []string
	Count  int
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
func (s *PublicDagAPI) GetHeaderByNumber(ctx context.Context, height Int) (string, error) {
	number := &modules.ChainIndex{}
	number.Index = height.Uint64()
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
func (s *PrivateDagAPI) GetHeaderByAuthor(ctx context.Context, author string, startHeight, count uint64) (string, error) {
	authorAddr, err := common.StringToAddress(author)
	if err != nil {
		return "", err
	}
	headers, err := s.b.Dag().GetHeadersByAuthor(authorAddr, startHeight, count)
	if err != nil {
		return "", err
	}
	result := []*ptnjson.HeaderJson{}
	for _, header := range headers {
		headerJson := ptnjson.ConvertUnitHeader2Json(header)
		result = append(result, headerJson)
	}
	content, err := json.Marshal(result)
	if err != nil {
		log.Info("PrivateDagAPI", "GetHeaderByAuthor Marshal err:", err, "author", author)
		return "", err
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
	jsonUnit := ptnjson.ConvertUnit2Json(unit, s.b.Dag().GetTxOutput, s.b.Dag().GetContractStateByVersion, s.b.EnableGasFee())
	content, err := json.Marshal(jsonUnit)
	if err != nil {
		log.Info("PublicBlockChainAPI", "GetUnitByHash Marshal err:", err, "unit:", *unit)
		return "jsonUnit Marshal err"
	}
	return string(content)
}

func (s *PublicDagAPI) GetUnitByNumber(ctx context.Context, height Int) string {
	log.Info("PublicDagAPI", "GetUnitByNumber height:", height.Uint64())

	number := &modules.ChainIndex{}

	number.Index = height.Uint64()

	number.AssetID = dagconfig.DagConfig.GetGasToken()
	log.Info("PublicBlockChainAPI info", "GetUnitByNumber_number.Index:", number.Index, "number:", number.String())

	unit := s.b.GetUnitByNumber(number)
	if unit == nil {
		log.Info("PublicBlockChainAPI", "GetUnitByNumber GetUnitByNumber is nil number:", number)
		return "GetUnitByNumber nil"
	}
	jsonUnit := ptnjson.ConvertUnit2Json(unit, s.b.Dag().GetTxOutput, s.b.Dag().GetContractStateByVersion, s.b.EnableGasFee())
	content, err := json.Marshal(jsonUnit)
	if err != nil {
		log.Info("PublicBlockChainAPI", "GetUnitByNumber Marshal err:", err, "unit:", *unit)
		return "json UnitMarshal err"
	}
	return string(content)
}
func (s *PublicDagAPI) GetUnitJsonByIndex(ctx context.Context, asset_id string, index uint64) string {
	number := &modules.ChainIndex{}
	number.Index = index

	assetId, _, err := modules.String2AssetId(asset_id)
	if err != nil {
		return fmt.Sprintf("the [%s] isn't unknow asset_id.", asset_id)
	}
	number.AssetID = assetId
	log.Info("PublicBlockChainAPI info", "GetUnitJsonByIndex:", index, "number:", number.String())

	unit := s.b.GetUnitByNumber(number)
	if unit == nil {
		log.Info("PublicBlockChainAPI", "GetUnitByNumber GetUnitByNumber is nil number:", number)
		return "the unit isn't exist."
	}
	content, err := json.Marshal(unit)
	if err != nil {
		log.Info("PublicBlockChainAPI", "GetUnitByNumber Marshal err:", err, "unit:", *unit)
		return "json UnitMarshal err: " + err.Error()
	}
	return string(content)
}
func (s *PublicDagAPI) GetUnitHexByIndex(ctx context.Context, asset_id string, index uint64) string {
	number := &modules.ChainIndex{}
	number.Index = index
	assetId, _, err := modules.String2AssetId(asset_id)
	if err != nil {
		return fmt.Sprintf("the [%s] isn't unknow asset_id.", asset_id)
	}
	number.AssetID = assetId
	log.Info("GetUnitHexByIndex info", "GetUnitHexByIndex:", index, "number:", number.String())

	unit := s.b.GetUnitByNumber(number)
	if unit == nil {
		log.Info("GetUnitHexByIndex is failed,", "number:", number)
		return "the unit isn't exist."
	}
	bytes, err := rlp.EncodeToBytes(unit)
	if err != nil {
		log.Info("getUnitHexByHash is failed,", "error", err.Error())
		return err.Error()
	}
	return common.Bytes2Hex(bytes)
}
func (s *PublicDagAPI) GetUnitHexByHash(ctx context.Context, condition string) string {
	log.Info("PublicDagAPI", "GetUnitHexByHash condition:", condition)
	hash := common.Hash{}
	if err := hash.SetHexString(condition); err != nil {
		log.Info("GetUnitHexByHash SetHexString err:", "error:", err.Error(), "condition:", condition)
		return "hash hex is illegal"
	}
	unit := s.b.GetUnitByHash(hash)
	if unit == nil {
		log.Info("getUnitHexByHash is failed,GetUnitHexByHash error:", "error", hash.String())
		return "GetUnitByHash nil"
	}
	bytes, err := rlp.EncodeToBytes(unit)
	if err != nil {
		log.Info("getUnitHexByHash is failed,", "error", err.Error())
		return err.Error()
	}
	return common.Bytes2Hex(bytes)
}
func (s *PublicDagAPI) InsertUnitByHex(ctx context.Context, unithex string) error {
	bytes := common.FromHex(unithex)
	unit := new(modules.Unit)
	if err := rlp.DecodeBytes(bytes, &unit); err != nil {
		log.Infof("Insert unit by hex failed, rlp decode error[%s].", err.Error())
		return err
	}
	log.Infof("rlp decode success.hash[%s], unit[%s]", unit.Hash().String(), unit.String4Log())
	return s.b.Dag().InsertUnit(unit)
}

// getUnitsByIndex
func (s *PublicDagAPI) GetUnitsByIndex(ctx context.Context, start, end decimal.Decimal, asset string) string {
	log.Info("PublicDagAPI ,GetUnitsByIndexs:", "start", start, "end", end)
	units := s.b.GetUnitsByIndex(start, end, asset)
	jsonUnits := make([]*ptnjson.UnitJson, 0)

	for _, u := range units {
		jsonu := ptnjson.ConvertUnit2Json(u, s.b.Dag().GetTxOutput, s.b.Dag().GetContractStateByVersion, s.b.EnableGasFee())
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
	gasToken := dagconfig.DagConfig.GasToken
	if assetid == "" {
		assetid = gasToken
	}

	assetid = strings.ToUpper(assetid)
	token, _, err := modules.String2AssetId(assetid)
	if err != nil {
		return "unknow assetid:" + assetid + ". " + err.Error()
	}

	if assetid != gasToken {
		GlobalStateContractId := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
		val, _, err := s.b.GetContractState(GlobalStateContractId, modules.GlobalPrefix+strings.ToUpper(token.GetSymbol()))
		if err != nil || len(val) == 0 {
			return "unknow assetid: " + assetid + ", " + err.Error()
		}
	}

	result := new(ptnjson.ChainUnitPropertyJson)
	//stableUnit := s.b.Dag().CurrentUnit(token)
	//ustabeUnit := s.b.Dag().GetCurrentMemUnit(token)
	stableUnit, _ := s.b.Dag().StableHeadUnitProperty(token)
	ustabeUnit, _ := s.b.Dag().UnstableHeadUnitProperty(token)

	if ustabeUnit != nil {
		//result.FastHash = ustabeUnit.Hash()
		//result.FastIndex = ustabeUnit.NumberU64()
		result.FastHash = ustabeUnit.Hash
		result.FastIndex = ustabeUnit.ChainIndex.Index
		result.FastTimestamp = time.Unix(int64(ustabeUnit.Timestamp),
			0).Format("2006-01-02 15:04:05 -0700 MST")
	}
	if stableUnit != nil {
		//result.StableHash = stableUnit.Hash()
		//result.StableIndex = stableUnit.NumberU64()
		result.StableHash = stableUnit.Hash
		result.StableIndex = stableUnit.ChainIndex.Index
		result.StableTimestamp = time.Unix(int64(stableUnit.Timestamp),
			0).Format("2006-01-02 15:04:05 -0700 MST")
	}

	content, err := json.Marshal(result)
	if err != nil {
		log.Info("PublicDagAPI", "GetFastUnitIndex Marshal err:", err)
		return "result Marshal err"
	}

	return string(content)
}

func (s *PublicDagAPI) GetChainInfo() (*ptnjson.ChainInfo, error) {
	gasToken := dagconfig.DagConfig.GetGasToken()

	headUnit, err := s.b.Dag().UnstableHeadUnitProperty(gasToken)
	if err != nil {
		return nil, err
	}
	stableUnit, err := s.b.Dag().StableHeadUnitProperty(gasToken)
	if err != nil {
		return nil, err
	}

	ci := new(ptnjson.ChainInfo)
	ci.HeadHash = headUnit.Hash
	ci.HeadNum = headUnit.ChainIndex.Index
	ci.HeadTime = time.Unix(int64(headUnit.Timestamp),
		0).Format("2006-01-02 15:04:05 -0700 MST")
	ci.StableHash = stableUnit.Hash
	ci.StableIndex = stableUnit.ChainIndex.Index
	ci.StableTime = time.Unix(int64(stableUnit.Timestamp),
		0).Format("2006-01-02 15:04:05 -0700 MST")

	return ci, nil
}

func (s *PublicDagAPI) GetUnitSummaryByNumber(ctx context.Context, height Int) string {
	log.Info("PublicBlockChainAPI", "GetUnitByNumber height:", height)

	number := &modules.ChainIndex{}
	number.Index = height.Uint64()

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

func (s *PublicDagAPI) GetTxPackInfo(ctx context.Context, txHash string) (*ptnjson.TxPackInfoJson, error) {
	hash := common.HexToHash(txHash)
	return s.b.GetTxPackInfo(hash)
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
	items, err := s.b.GetDagAddrUtxos(addr)

	if err != nil {
		return "", err
	}
	info := NewPublicReturnInfo("address_utxos", items)
	result_json, _ := json.Marshal(info)
	return string(result_json), nil
}

func (s *PublicDagAPI) GetAddrUtxoTxs(ctx context.Context, addr string) ([]*ptnjson.TxWithUnitInfoJson, error) {
	return s.b.GetAddrUtxoTxs(addr)
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
	item, err := s.b.GetTxPoolTxByHash(hash)
	if err != nil {
		return "pool_tx:null", err
	}
	result_json, _ := json.Marshal(item)
	return string(result_json), nil

}

//GetTxStatusByHash returns the transaction status for hash
func (s *PublicDagAPI) GetTxStatusByHash(ctx context.Context, hex string) (*ptnjson.TxPoolTxJson, error) {
	log.Debug("this is hash tx's hash hex to find tx.", "hex", hex)
	if len(hex) > 72 || len(hex) < 64 {
		return nil, fmt.Errorf("the hex[%s] is illegal.", hex)
	}
	hash := common.HexToHash(hex)

	tx_status := new(ptnjson.TxPoolTxJson)
	item, err := s.b.GetTxPoolTxByHash(hash)
	if err != nil {
		if tx_info, err := s.b.Dag().GetTxByReqId(hash); err == nil {
			return ptnjson.ConvertTxWithInfo2Json(tx_info), nil
		}
		if tx_info, err := s.b.Dag().GetTransaction(hash); err == nil {
			return ptnjson.ConvertTxWithInfo2Json(tx_info), nil
		}
		tx_status.NotExsit = true
		log.Debugf("the txhash[%s] is not exist in dag,error[%s]", hash.String(), err.Error())
		tx_status.TxHash = hex
		return tx_status, nil
	}
	return item, nil
}

// MemdagInfos returns the pool transaction for the given hash
func (s *PublicDagAPI) MemdagInfos(ctx context.Context) (string, error) {
	log.Debug("get the memdag infos...")

	item, err := s.b.MemdagInfos()
	if err != nil {
		return "memdag_infos:null", err
	} else {
		info := NewPublicReturnInfo("memdag_infos", item)
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
		return dag.StableUnitNum()
	}
	return uint64(0)
}

func (s *PublicDagAPI) GetStableUnit() (*ptnjson.UnitPropertyJson, error) {
	dag := s.b.Dag()
	if dag != nil {
		gasToken := dagconfig.DagConfig.GetGasToken()
		unitProperty, err := dag.StableHeadUnitProperty(gasToken)
		if err != nil || unitProperty == nil {
			return nil, err
		}

		return ptnjson.UnitPropertyToJson(unitProperty), nil
	}

	return nil, nil
}

func (s *PublicDagAPI) GetHeadUnit() (*ptnjson.UnitPropertyJson, error) {
	dag := s.b.Dag()
	if dag != nil {
		gasToken := dagconfig.DagConfig.GetGasToken()
		unitProperty, err := dag.UnstableHeadUnitProperty(gasToken)
		if err != nil || unitProperty == nil {
			return nil, err
		}

		return ptnjson.UnitPropertyToJson(unitProperty), nil
	}

	return nil, nil
}

func (s *PublicDagAPI) IsSynced() bool {
	dag := s.b.Dag()
	if dag != nil {
		return dag.IsSynced(false)
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

func (s *PrivateDagAPI) CheckUnits(ctx context.Context, assetId string, number int) (bool, error) {
	dag := s.b.Dag()
	err := dag.CheckUnitsCorrect(assetId, number)
	if err != nil {
		log.Errorf("check units failed, %s", err.Error())
		return false, err
	}
	log.Debugf("check units success,%s", assetId)
	return true, nil
}
func (s *PrivateDagAPI) RebuildAddrTxIndex() error {
	dag := s.b.Dag()
	return dag.RebuildAddrTxIndex()
}

type TxAndStatus struct {
	Tx     *ptnjson.TxJson
	Status string
}

func (s *PrivateDagAPI) GetLocalTx(txId string) (*TxAndStatus, error) {
	txhash := common.HexToHash(txId)
	tx, status, err := s.b.Dag().GetLocalTx(txhash)
	if err != nil {
		return nil, err
	}
	txjson := ptnjson.ConvertTx2FullJson(tx, nil)
	return &TxAndStatus{
		Tx:     txjson,
		Status: status.String(),
	}, nil
}
