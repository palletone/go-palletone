package tokentradecc

import (
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/shopspring/decimal"
	"github.com/palletone/go-palletone/contracts/shim"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/dag/errors"
)

const TRADELIST_RECORD = "TradeOrder-"
const TRADELIST_HISTORY = "TradeOrderHistory-"

type TradeOrder struct {
	TradeType     byte           //1:买, 2:卖, 3:其他
	Address       common.Address //挂单地址
	InAsset       *modules.Asset //转入Asset
	InAmount      uint64         //挂单的金额
	InUnCount     uint64         //未成交数量---want amount
	Scale         string         //兑换比例
	WantAsset     *modules.Asset //交易asset
	WantAmount    uint64         //挂单时想要多少金额//竞拍时起拍价
	UnAmount      uint64         //未成交数量---want amount
	StartTime     string         //time.Time //开始时间   //int64 rlp失败
	EndTime       string         //time.Time //结束时间
	RewardAddress common.Address //奖励地址
	TradeSn       string         //挂单ID
	CreateTime    string         //time.Time
	Status        byte           //0 撤销，1 挂单中，2 成交完毕
}

type TradeFeeUse struct {
	Asset              *modules.Asset
	RewardAddress      common.Address
	RewardAmount       uint64 //decimal.Decimal //奖励
	DestructionAddress common.Address
	DestructionAmount  uint64 //decimal.Decimal //销毁
}

type TradeOrderJson struct {
	TradeType     string          `json:"trade_type"` //
	Address       common.Address  `json:"address"`    //挂单地址
	InAsset       string          `json:"in_asset"`
	InAmount      decimal.Decimal `json:"in_amount"`
	InUnCount     decimal.Decimal `json:"in_uncount"`
	WantAsset     string          `json:"want_asset"`
	WantAmount    decimal.Decimal `json:"want_amount"`
	Scale         string          `json:"scale"`          //兑换比例
	UnAmount      decimal.Decimal `json:"un_trade_count"` //未成交数量---want amount
	StartTime     string          `json:"start_time"`
	EndTime       string          `json:"end_time"`
	RewardAddress string          `json:"reward_address"`
	TradeSn       string          `json:"trade_sn"`
	Status        string          `json:"status"`
	CreateTime    string          `json:"create_time"`
}

func convertSheet(exm TradeOrder) *TradeOrderJson {
	newSheet := TradeOrderJson{}
	newSheet.TradeType = string(exm.TradeType)
	newSheet.Address = exm.Address
	newSheet.InAsset = exm.InAsset.String()
	newSheet.InAmount = exm.InAsset.DisplayAmount(exm.InAmount)
	newSheet.InUnCount = exm.InAsset.DisplayAmount(exm.InUnCount)
	newSheet.Scale = exm.Scale
	newSheet.WantAsset = exm.WantAsset.String()
	newSheet.WantAmount = exm.WantAsset.DisplayAmount(exm.WantAmount)
	newSheet.UnAmount = exm.WantAsset.DisplayAmount(exm.UnAmount)
	newSheet.StartTime = exm.StartTime
	newSheet.EndTime = exm.EndTime
	newSheet.RewardAddress = exm.RewardAddress.String()
	newSheet.TradeSn = exm.TradeSn
	newSheet.Status = string(exm.Status)
	newSheet.CreateTime = exm.CreateTime
	return &newSheet
}

//增加一个新订单
func SaveTradeOrder(stub shim.ChaincodeStubInterface, order *TradeOrder) error {
	data, _ := rlp.EncodeToBytes(order)
	key := TRADELIST_RECORD + order.TradeSn
	return stub.PutState(key, data)
}

func deleteTradeOrder(stub shim.ChaincodeStubInterface, orderSn string) error {
	return stub.DelState(TRADELIST_RECORD + orderSn)
}

//更新一个已有的订单
func UpdateTradeOrder(stub shim.ChaincodeStubInterface, order *TradeOrder) error {
	if order.Status == 1 { //更新未成交订单
		err := SaveTradeOrder(stub, order)
		if err != nil {
			return err
		}
		return nil
	} else if order.Status == 2 || order.Status == 0 { //成交 或者 取消
		err := deleteTradeOrder(stub, order.TradeSn)
		if err != nil {
			return err
		}
		err = saveTradeOrderHistory(stub, order)
		if err != nil {
			return err
		}
		return nil
	}
	return nil
}

//取消一个订单
func cancelTradeOrder(stub shim.ChaincodeStubInterface, order *TradeOrder) error {
	order.Status = 0
	err := deleteTradeOrder(stub, order.TradeSn)
	if err != nil {
		return err
	}
	err = saveTradeOrderHistory(stub, order)
	if err != nil {
		return err
	}
	return nil
}

func getActiveOrderCount(stub shim.ChaincodeStubInterface) (int, error) {
	kvs, err := stub.GetStateByPrefix(TRADELIST_RECORD)
	if err != nil {
		return 0, err
	}
	return len(kvs), nil
}

//获得订单列表
func getAllTradeOrder(stub shim.ChaincodeStubInterface, start, end int) ([]*TradeOrderJson, error) {
	if start > end {
		return nil, errors.New("input num err")
	}
	kvs, err := stub.GetStateByPrefix(TRADELIST_RECORD)
	if err != nil {
		return nil, err
	}
	from, to := 0, len(kvs)
	if start > 0 {
		from = start
	}
	if end > 0 {
		to = end
	}
	result := make([]*TradeOrderJson, 0, len(kvs))
	for i, kv := range kvs {
		if i >= from && i <= to {
			record := &TradeOrder{}
			err = rlp.DecodeBytes(kv.Value, record)
			if err != nil {
				return nil, err
			}
			log.Debugf("getAllAuctionOrder, record:%v", record)
			jsSheet := convertSheet(*record)
			result = append(result, jsSheet)
		}
	}
	return result, nil
}

//查询某地址的订单，
//TODO 全表扫描，大量订单时，性能有问题
func getTradeOrderByAddress(stub shim.ChaincodeStubInterface, addr common.Address) ([]*TradeOrderJson, error) {
	kvs, err := stub.GetStateByPrefix(TRADELIST_RECORD)
	if err != nil {
		return nil, err
	}
	result := make([]*TradeOrderJson, 0, len(kvs))
	for _, kv := range kvs {
		record := &TradeOrder{}
		err = rlp.DecodeBytes(kv.Value, record)
		if err != nil {
			return nil, err
		}
		if record.Address == addr {
			jsSheet := convertSheet(*record)
			result = append(result, jsSheet)
		}
	}
	return result, nil
}

//根据订单号获得订单
func getTradeRecordBySn(stub shim.ChaincodeStubInterface, auctionSn string) (*TradeOrder, error) {
	key := TRADELIST_RECORD + auctionSn
	value, err := stub.GetState(key)
	if err != nil {
		return nil, err
	}
	record := &TradeOrder{}
	if len(value) > 0 {
		err = rlp.DecodeBytes(value, record)
		if err != nil {
			return nil, err
		}
		return record, nil
	} else {
		return nil, errors.New("not find id")
	}
}

//将一个Order移动到History有两种情况，
//1,该订单全部成交完毕
//2,该订单被Maker取消了
func saveTradeOrderHistory(stub shim.ChaincodeStubInterface, order *TradeOrder) error {
	data, _ := rlp.EncodeToBytes(order)
	key := TRADELIST_HISTORY + order.TradeSn
	return stub.PutState(key, data)
}

func getHistoryOrderCount(stub shim.ChaincodeStubInterface) (int, error) {
	kvs, err := stub.GetStateByPrefix(TRADELIST_HISTORY)
	if err != nil {
		return 0, err
	}
	return len(kvs), nil
}

func getAllHistoryOrder(stub shim.ChaincodeStubInterface, start, end int) ([]*TradeOrderJson, error) {
	if start > end {
		return nil, errors.New("input num err")
	}
	kvs, err := stub.GetStateByPrefix(TRADELIST_HISTORY)
	if err != nil {
		return nil, err
	}
	from, to := 0, len(kvs)
	if start > 0 {
		from = start
	}
	if end > 0 {
		to = end
	}
	result := make([]*TradeOrderJson, 0, len(kvs))
	for i, kv := range kvs {
		if i >= from && i <= to {
			record := &TradeOrder{}
			err = rlp.DecodeBytes(kv.Value, record)
			if err != nil {
				return nil, err
			}
			jsSheet := convertSheet(*record)
			result = append(result, jsSheet)
		}
	}
	return result, nil
}
func getHistoryOrderBySn(stub shim.ChaincodeStubInterface, sn string) (*TradeOrderJson, error) {
	key := TRADELIST_HISTORY + sn
	value, err := stub.GetState(key)
	if err != nil {
		return nil, err
	}
	record := &TradeOrder{}
	if len(value) > 0 {
		err = rlp.DecodeBytes(value, record)
		if err != nil {
			return nil, err
		}
		jsSheet := convertSheet(*record)
		return jsSheet, nil
	} else {
		return nil, errors.New("not find id")
	}
}

func getHistoryOrderByMakerAddr(stub shim.ChaincodeStubInterface, addr common.Address) ([]*TradeOrderJson, error) {
	kvs, err := stub.GetStateByPrefix(TRADELIST_HISTORY)
	if err != nil {
		return nil, err
	}
	result := make([]*TradeOrderJson, 0, len(kvs))
	for _, kv := range kvs {
		record := &TradeOrder{}
		err = rlp.DecodeBytes(kv.Value, record)
		if err != nil {
			return nil, err
		}
		if record.Address.Equal(addr) {
			jsSheet := convertSheet(*record)
			result = append(result, jsSheet)
		}
	}
	return result, nil
}
