package auctioncc

import (
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/contracts/shim"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/shopspring/decimal"
)

const MatchRecordPrefix = "AuctionMatchRecord-"

//订单成交记录/提交记录
type MatchRecord struct {
	AuctionOrderSn   string
	TakerReqId       string
	MakerAddress     common.Address
	MakerAsset       *modules.Asset
	MakerAssetAmount uint64 //Makers  NFT数量为1
	TakerAddress     common.Address
	TakerAsset       *modules.Asset
	TakerAssetAmount uint64        //Taker提交多少金额
	FeeUse           AuctionFeeUse //消耗的费用：奖励和燃烧
}

type MatchRecordJson struct {
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
	return &MatchRecordJson{
		AuctionOrderSn:   record.AuctionOrderSn,
		TakerReqId:       record.TakerReqId,

		MakerAddress:     record.MakerAddress.String(),
		MakerAsset:  record.MakerAsset.String(),
		MakerAssetAmount: record.MakerAsset.DisplayAmount(record.MakerAssetAmount),

		TakerAddress:     record.TakerAddress.String(),
		TakerAsset:  record.TakerAsset.String(),
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
func getMatchRecordByOrderSn(stub shim.ChaincodeStubInterface, orderSn string) ([]*MatchRecordJson, error) {
	key := MatchRecordPrefix + orderSn
	rows, err := stub.GetStateByPrefix(key)
	if err != nil {
		return nil, err
	}
	result := []*MatchRecordJson{}
	for _, row := range rows {
		record := &MatchRecord{}
		rlp.DecodeBytes(row.Value, record)
		result = append(result, convertMatchRecord(record))
	}
	return result, nil
}
func getAllMatchRecord(stub shim.ChaincodeStubInterface) ([]*MatchRecordJson, error) {
	key := MatchRecordPrefix
	rows, err := stub.GetStateByPrefix(key)
	if err != nil {
		return nil, err
	}
	result := []*MatchRecordJson{}
	for _, row := range rows {
		record := &MatchRecord{}
		rlp.DecodeBytes(row.Value, record)
		result = append(result, convertMatchRecord(record))
	}
	return result, nil
}
