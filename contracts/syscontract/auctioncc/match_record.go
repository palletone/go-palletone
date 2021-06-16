package auctioncc

import (
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/contracts/shim"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/shopspring/decimal"
	"github.com/palletone/go-palletone/common/log"
)

const MatchRecordPrefix = "AuctionMatchRecord-"

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

	recordTime int64 //成交时间
}

type MatchRecordJson struct {
	AuctionType        string
	AuctionOrderSn     string
	TakerReqId         string
	MakerAddress       string
	MakerAsset         string
	MakerAssetAmount   decimal.Decimal //Makers  NFT数量为1
	TakerAddress       string
	TakerAsset         string
	TakerAssetAmount   decimal.Decimal //Taker成交或者提交了多少金额
	FeeAsset           string
	RewardAddress      string
	RewardAmount       decimal.Decimal
	DestructionAddress string
	DestructionAmount  decimal.Decimal
}

func convertMatchRecord(record *MatchRecord) *MatchRecordJson {
	if record.AuctionOrderSn == "" {
		return nil
	}
	return &MatchRecordJson{
		AuctionType:    string(record.AuctionType),
		AuctionOrderSn: record.AuctionOrderSn,
		TakerReqId:     record.TakerReqId,

		MakerAddress:     record.MakerAddress.String(),
		MakerAsset:       record.MakerAsset.String(),
		MakerAssetAmount: record.MakerAsset.DisplayAmount(record.MakerAssetAmount),

		TakerAddress:     record.TakerAddress.String(),
		TakerAsset:       record.TakerAsset.String(),
		TakerAssetAmount: record.TakerAsset.DisplayAmount(record.TakerAssetAmount),

		FeeAsset:           record.FeeUse.Asset.String(),
		RewardAddress:      record.FeeUse.RewardAddress.String(),
		RewardAmount:       record.FeeUse.Asset.DisplayAmount(record.FeeUse.RewardAmount),
		DestructionAddress: record.FeeUse.DestructionAddress.String(),
		DestructionAmount:  record.FeeUse.Asset.DisplayAmount(record.FeeUse.DestructionAmount),
	}
}

//保存一笔成交记录
func saveMatchRecord(stub shim.ChaincodeStubInterface, record *MatchRecord) error {
	data, _ := rlp.EncodeToBytes(record)
	key := MatchRecordPrefix + record.AuctionOrderSn + "-" + record.TakerReqId
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
func getAllMatchRecordJson(stub shim.ChaincodeStubInterface) ([]*MatchRecordJson, error) {
	key := MatchRecordPrefix
	rows, err := stub.GetStateByPrefix(key)
	if err != nil {
		return nil, err
	}
	result := []*MatchRecordJson{}
	for _, row := range rows {
		record := &MatchRecord{}
		rlp.DecodeBytes(row.Value, record)
		log.Debugf("getAllMatchRecordJson:%v", record)
		result = append(result, convertMatchRecord(record))
	}

	return result, nil
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

func getMaxAmountRecord(records []*MatchRecord) (*MatchRecord) {
	maxRecord := &MatchRecord{}
	for _, ro := range records {
		if ro.TakerAssetAmount > maxRecord.TakerAssetAmount {
			maxRecord = ro
			continue
		} else if ro.TakerAssetAmount == maxRecord.TakerAssetAmount {
			//比较时间
			if ro.recordTime < maxRecord.recordTime {
				maxRecord = ro
			}
		}
	}
	//没有找到交易订单
	if maxRecord.AuctionOrderSn == "" || maxRecord.TakerReqId == "" {
		return nil
	}
	return maxRecord
}

func isInMatchRecord(stub shim.ChaincodeStubInterface, orderSn string) bool {
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
