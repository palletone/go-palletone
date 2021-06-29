package auctioncc

import (
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/shopspring/decimal"
	"github.com/palletone/go-palletone/contracts/shim"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common/log"
)

const AUCTIONLIST_RECORD = "AuctionOrder-"
const AUCTIONLIST_HISTORY = "AuctionOrderHistory-"

type AuctionOrder struct {
	AuctionType byte //1:一口价, 2:英式拍卖, 3:荷兰式拍卖
	Address     common.Address
	SaleAsset   *modules.Asset
	SaleAmount  uint64 //挂单的金额  ，NFT的话为1
	WantAsset   *modules.Asset
	WantAmount  uint64 //挂单时想要多少金额//竞拍时起拍价

	TargetAmount uint64 //最大价格,可以不设置，默认0     //auction
	StepAmount   uint64 //阶梯数量,可以不设置，默认0     //auction
	StartTime    string //time.Time //开始时间           //int64 rlp失败
	EndTime      string //time.Time //结束时间                    //auction

	RewardAddress common.Address
	AuctionSn     string
	CreateTime    string //time.Time
	Status        byte   //0 撤销， 1 挂单中，2 成交完毕，
}

type AuctionFeeUse struct {
	Asset              *modules.Asset
	RewardAddress      common.Address
	RewardAmount       uint64 //decimal.Decimal //奖励
	DestructionAddress common.Address
	DestructionAmount  uint64 //decimal.Decimal //销毁
}

type AuctionOrderJson struct {
	AuctionType   string          `json:"auction_type"` //
	Address       common.Address  `json:"address"`      //挂单地址
	SaleAsset     string          `json:"sale_asset"`
	SaleAmount    decimal.Decimal `json:"sale_amount"`
	WantAsset     string          `json:"want_asset"`
	WantAmount    decimal.Decimal `json:"want_amount"`
	TargetAmount  decimal.Decimal `json:"target_amount"`
	StepAmount    decimal.Decimal `json:"step_amount"`
	StartTime     string          `json:"start_time"`
	EndTime       string          `json:"end_time"`
	RewardAddress string          `json:"reward_address"`
	AuctionSn     string          `json:"auction_sn"`
	Status        string          `json:"status"`
	CreateTime    string          `json:"create_time"`
}

func convertSheet(exm AuctionOrder) *AuctionOrderJson {
	newSheet := AuctionOrderJson{}
	newSheet.AuctionType = string(exm.AuctionType)
	newSheet.Address = exm.Address
	newSheet.SaleAsset = exm.SaleAsset.String()
	newSheet.SaleAmount = exm.SaleAsset.DisplayAmount(exm.SaleAmount)
	newSheet.WantAsset = exm.WantAsset.String()
	newSheet.WantAmount = exm.WantAsset.DisplayAmount(exm.WantAmount)
	newSheet.TargetAmount = exm.WantAsset.DisplayAmount(exm.TargetAmount)
	newSheet.StepAmount = exm.WantAsset.DisplayAmount(exm.StepAmount)
	newSheet.StartTime = exm.StartTime
	newSheet.EndTime = exm.EndTime
	newSheet.RewardAddress = exm.RewardAddress.String()
	newSheet.AuctionSn = exm.AuctionSn
	newSheet.CreateTime = exm.CreateTime
	return &newSheet
}

//增加一个新订单
func SaveAuctionOrder(stub shim.ChaincodeStubInterface, order *AuctionOrder) error {
	data, _ := rlp.EncodeToBytes(order)
	key := AUCTIONLIST_RECORD + order.AuctionSn
	return stub.PutState(key, data)
}

func deleteAuctionOrder(stub shim.ChaincodeStubInterface, orderSn string) error {
	return stub.DelState(AUCTIONLIST_RECORD + orderSn)
}

//更新一个已有的订单
func UpdateAuctionOrder(stub shim.ChaincodeStubInterface, order *AuctionOrder) error {
	if order.Status == 1 { //更新未成交订单
		err := SaveAuctionOrder(stub, order)
		if err != nil {
			return err
		}
		return nil
	} else if order.Status == 2 || order.Status == 0 { //成交 或者 取消
		err := deleteAuctionOrder(stub, order.AuctionSn)
		if err != nil {
			return err
		}
		err = saveAuctionOrderHistory(stub, order)
		if err != nil {
			return err
		}
		return nil
	}
	return nil
}

//取消一个订单
func cancelAuctionOrder(stub shim.ChaincodeStubInterface, order *AuctionOrder) error {
	order.Status = 0
	err := deleteAuctionOrder(stub, order.AuctionSn)
	if err != nil {
		return err
	}
	err = saveAuctionOrderHistory(stub, order)
	if err != nil {
		return err
	}
	return nil
}

//获得订单列表
func getAllAuctionOrder(stub shim.ChaincodeStubInterface) ([]*AuctionOrderJson, error) {
	kvs, err := stub.GetStateByPrefix(AUCTIONLIST_RECORD)
	if err != nil {
		return nil, err
	}
	result := make([]*AuctionOrderJson, 0, len(kvs))
	for _, kv := range kvs {
		record := &AuctionOrder{}
		err = rlp.DecodeBytes(kv.Value, record)
		if err != nil {
			return nil, err
		}
		log.Debugf("getAllAuctionOrder, record:%v", record)
		jsSheet := convertSheet(*record)
		result = append(result, jsSheet)
	}
	return result, nil
}

//查询某地址的订单，
//TODO 全表扫描，大量订单时，性能有问题
func getAuctionOrderByAddress(stub shim.ChaincodeStubInterface, addr common.Address) ([]*AuctionOrderJson, error) {
	kvs, err := stub.GetStateByPrefix(AUCTIONLIST_RECORD)
	if err != nil {
		return nil, err
	}
	result := make([]*AuctionOrderJson, 0, len(kvs))
	for _, kv := range kvs {
		record := &AuctionOrder{}
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
func getAuctionRecordBySn(stub shim.ChaincodeStubInterface, auctionSn string) (*AuctionOrder, error) {
	key := AUCTIONLIST_RECORD + auctionSn
	value, err := stub.GetState(key)
	if err != nil {
		return nil, err
	}
	record := &AuctionOrder{}
	err = rlp.DecodeBytes(value, record)
	if err != nil {
		return nil, err
	}
	return record, nil
}

//将一个Order移动到History有两种情况，
//1,该订单全部成交完毕
//2,该订单被Maker取消了
func saveAuctionOrderHistory(stub shim.ChaincodeStubInterface, order *AuctionOrder) error {
	data, _ := rlp.EncodeToBytes(order)
	key := AUCTIONLIST_HISTORY + order.AuctionSn
	return stub.PutState(key, data)
}
func getAllHistoryOrder(stub shim.ChaincodeStubInterface) ([]*AuctionOrderJson, error) {
	kvs, err := stub.GetStateByPrefix(AUCTIONLIST_HISTORY)
	if err != nil {
		return nil, err
	}
	result := make([]*AuctionOrderJson, 0, len(kvs))
	for _, kv := range kvs {
		record := &AuctionOrder{}
		err = rlp.DecodeBytes(kv.Value, record)
		if err != nil {
			return nil, err
		}
		jsSheet := convertSheet(*record)
		result = append(result, jsSheet)
	}
	return result, nil
}
