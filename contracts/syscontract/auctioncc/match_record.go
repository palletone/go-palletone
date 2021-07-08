package auctioncc

import (
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/contracts/shim"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/shopspring/decimal"
	"github.com/palletone/go-palletone/common/log"
	"fmt"
	"github.com/palletone/go-palletone/dag/errors"
)

const MatchRecordPrefix = "AuctionMatchRecord-"
const AuctionLastAmountPrefix = "AuctionLastAmountRecord-"

//订单成交记录/提交记录
type MatchRecord struct {
	AuctionType      byte //1:一口价, 2:英式拍卖, 3:荷兰式拍卖
	AuctionOrderSn   string
	TakerReqId       string
	MakerAddress     common.Address
	MakerAsset       *modules.Asset
	MakerAssetAmount uint64 //Makers  NFT数量为1
	TakerAddress     common.Address
	TakerAsset       *modules.Asset
	TakerAssetAmount uint64        //Taker提交多少金额
	FeeUse           AuctionFeeUse //消耗的费用：奖励和燃烧
	RecordTime       string        //成交时间
}

type MatchRecordJson struct {
	AuctionType        string          `json:"auction_type"`
	AuctionOrderSn     string          `json:"auction_sn"`
	TakerReqId         string          `json:"taker_reqId"`
	MakerAddress       string          `json:"maker_address"`
	MakerAsset         string          `json:"maker_asset"`
	MakerAssetAmount   decimal.Decimal `json:"maker_asset_amount"` //Makers  NFT数量为1
	TakerAddress       string          `json:"taker_address"`
	TakerAsset         string          `json:"taker_asset"`
	TakerAssetAmount   decimal.Decimal `json:"taker_asset_amount"` //Taker成交或者提交了多少金额
	FeeAsset           string          `json:"fee_asset"`
	RewardAddress      string          `json:"reward_address"`
	RewardAmount       decimal.Decimal `json:"reward_amount"`
	DestructionAddress string          `json:"destruction_address"`
	DestructionAmount  decimal.Decimal `json:"destruction_amount"`
	RecordTime         string          `json:"record_time"`
}

type AuctionLastAmount struct {
	AuctionOrderSn string
	TakerReqId     string
	TakerAddress   common.Address
	TakerAsset     *modules.Asset
	TakerAmount    uint64
}

func convertMatchRecord(record *MatchRecord) *MatchRecordJson {
	if record.AuctionOrderSn == "" {
		return nil
	}

	return &MatchRecordJson{
		AuctionType:        string(record.AuctionType),
		AuctionOrderSn:     record.AuctionOrderSn,
		TakerReqId:         record.TakerReqId,
		MakerAddress:       record.MakerAddress.String(),
		MakerAsset:         record.MakerAsset.String(),
		MakerAssetAmount:   record.MakerAsset.DisplayAmount(record.MakerAssetAmount),
		TakerAddress:       record.TakerAddress.String(),
		TakerAsset:         record.TakerAsset.String(),
		TakerAssetAmount:   record.TakerAsset.DisplayAmount(record.TakerAssetAmount),
		FeeAsset:           record.FeeUse.Asset.String(),
		RewardAddress:      record.FeeUse.RewardAddress.String(),
		RewardAmount:       record.FeeUse.Asset.DisplayAmount(record.FeeUse.RewardAmount),
		DestructionAddress: record.FeeUse.DestructionAddress.String(),
		DestructionAmount:  record.FeeUse.Asset.DisplayAmount(record.FeeUse.DestructionAmount),
		RecordTime:         record.RecordTime,
	}
}

//保存一笔成交记录
func saveMatchRecord(stub shim.ChaincodeStubInterface, record *MatchRecord) error {
	data, _ := rlp.EncodeToBytes(record)
	key := MatchRecordPrefix + record.AuctionOrderSn + "-" + record.TakerReqId
	stub.DelState(key)
	return stub.PutState(key, data)
}

//获得一个订单的匹配成交记录
func getMatchRecordJsonByOrderSn(stub shim.ChaincodeStubInterface, orderSn string) ([]*MatchRecordJson, error) {
	key := MatchRecordPrefix + orderSn
	rows, err := stub.GetStateByPrefix(key)
	if err != nil {
		return nil, err
	}
	result := []*MatchRecordJson{}
	for _, row := range rows {
		record := &MatchRecord{}
		rlp.DecodeBytes(row.Value, record)
		log.Debugf("getMatchRecordJsonByOrderSn:%v", record)
		result = append(result, convertMatchRecord(record))
	}
	return result, nil
}
func getAllMatchRecordJson(stub shim.ChaincodeStubInterface, start, end int) ([]*MatchRecordJson, error) {
	if start > end {
		return nil, errors.New("input num err")
	}
	key := MatchRecordPrefix
	rows, err := stub.GetStateByPrefix(key)
	if err != nil {
		return nil, err
	}
	from, to := 0, len(rows)
	if start > 0 {
		from = start
	}
	if end > 0 {
		to = end
	}

	result := []*MatchRecordJson{}
	for i, row := range rows {
		if i >= from && i <= to {
			record := &MatchRecord{}
			rlp.DecodeBytes(row.Value, record)
			log.Debugf("getAllMatchRecordJson:%v", record)
			result = append(result, convertMatchRecord(record))
		}
	}
	return result, nil
}

func getMatchRecordCount(stub shim.ChaincodeStubInterface) (int, error) {
	key := MatchRecordPrefix
	rows, err := stub.GetStateByPrefix(key)
	if err != nil {
		return 0, err
	}
	return len(rows), nil
}

func getMatchRecordByOrderSn(stub shim.ChaincodeStubInterface, orderSn string) ([]*MatchRecord, error) {
	key := MatchRecordPrefix + orderSn
	rows, err := stub.GetStateByPrefix(key)
	if err != nil {
		return nil, err
	}
	result := []*MatchRecord{}
	for _, row := range rows {
		record := &MatchRecord{}
		rlp.DecodeBytes(row.Value, record)
		result = append(result, record)
	}
	return result, nil
}

func getMatchRecordByAddress(stub shim.ChaincodeStubInterface, orderSn string, addr common.Address) (*MatchRecord, error) {
	record := &MatchRecord{}
	key := MatchRecordPrefix + orderSn
	rows, err := stub.GetStateByPrefix(key)
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		rlp.DecodeBytes(row.Value, record)
		if record.TakerAddress.Equal(addr) {
			return record, nil
		}
	}
	return nil, fmt.Errorf("getMatchRecordByAddress, not find address:%s", addr.String())
}

func getMaxAmountRecord(records []*MatchRecord) (*MatchRecord) {
	maxRecord := &MatchRecord{}
	for _, ro := range records {
		if ro.TakerAssetAmount > maxRecord.TakerAssetAmount {
			maxRecord = ro
			continue
		} else if ro.TakerAssetAmount == maxRecord.TakerAssetAmount {
			//比较时间
			roTime, err1 := getTimeFromString(ro.RecordTime)
			mxTime, err2 := getTimeFromString(maxRecord.RecordTime)
			if err1 == nil && err2 == nil {
				if roTime.Before(mxTime) {
					maxRecord = ro
				}
			}
		}
	}
	//没有找到交易订单
	if maxRecord.AuctionOrderSn == "" || maxRecord.TakerReqId == "" {
		return nil
	}

	return maxRecord
}

func isInMatchRecordByOrderSn(stub shim.ChaincodeStubInterface, orderSn string) bool {
	key := MatchRecordPrefix
	rows, err := stub.GetStateByPrefix(key)
	if err != nil {
		return false
	}
	for _, row := range rows {
		record := &MatchRecord{}
		rlp.DecodeBytes(row.Value, record)
		if record.AuctionOrderSn == orderSn {
			return true
		}
	}
	return false
}

func isInMatchRecordByAssertId(stub shim.ChaincodeStubInterface, assert *modules.Asset) bool {
	key := MatchRecordPrefix
	rows, err := stub.GetStateByPrefix(key)
	if err != nil {
		return false
	}
	for _, row := range rows {
		record := &MatchRecord{}
		rlp.DecodeBytes(row.Value, record)
		log.Debugf("isInMatchRecordByAssertId, record:%v", record)
		if record.MakerAsset.Equal(assert) {
			return true
		}
	}
	return false
}

func saveAuctionLastAmountRecord(stub shim.ChaincodeStubInterface, record *AuctionLastAmount) error {
	data, _ := rlp.EncodeToBytes(record)
	key := AuctionLastAmountPrefix + record.AuctionOrderSn
	stub.DelState(key)
	return stub.PutState(key, data)
}

func getAuctionLastAmountRecord(stub shim.ChaincodeStubInterface, orderSn string) (record *AuctionLastAmount, err error) {
	lastAmount := &AuctionLastAmount{}
	key := AuctionLastAmountPrefix + orderSn
	data, err := stub.GetState(key)
	if err != nil {
		return nil, err
	}
	err = rlp.DecodeBytes(data, lastAmount)
	if err != nil {
		return nil, err
	}

	return lastAmount, nil
}
