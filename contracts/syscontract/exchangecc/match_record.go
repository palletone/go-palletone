package exchangecc

import (
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/contracts/shim"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/shopspring/decimal"
)

const MatchRecordPrefix = "MatchRecord-"

//订单成交记录
type MatchRecord struct {
	ExchangeOrderSn  string
	TakerReqId       string
	MakerMatchAmount uint64 //Maker成交了多少金额
	MakerMatchAsset  *modules.Asset
	MakerAddress     common.Address
	TakerMatchAmount uint64 //Taker成交了多少金额
	TakerMatchAsset  *modules.Asset
	TakerAddress     common.Address
}
type MatchRecordJson struct {
	ExchangeOrderSn  string
	TakerReqId       string
	MakerMatchAmount decimal.Decimal //Maker成交了多少金额
	MakerMatchAsset  string
	MakerAddress     string
	TakerMatchAmount decimal.Decimal //Taker成交了多少金额
	TakerMatchAsset  string
	TakerAddress     string
}

func convertMatchRecord(record *MatchRecord) *MatchRecordJson {
	return &MatchRecordJson{
		ExchangeOrderSn:  record.ExchangeOrderSn,
		TakerReqId:       record.TakerReqId,
		MakerMatchAmount: record.MakerMatchAsset.DisplayAmount(record.MakerMatchAmount),
		MakerMatchAsset:  record.MakerMatchAsset.String(),
		MakerAddress:     record.MakerAddress.String(),
		TakerMatchAmount: record.TakerMatchAsset.DisplayAmount(record.TakerMatchAmount),
		TakerMatchAsset:  record.TakerMatchAsset.String(),
		TakerAddress:     record.TakerAddress.String(),
	}
}

//保存一笔成交记录
func saveMatchRecord(stub shim.ChaincodeStubInterface, record *MatchRecord) error {
	data, _ := rlp.EncodeToBytes(record)
	key := MatchRecordPrefix + record.ExchangeOrderSn + "-" + record.TakerReqId
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
