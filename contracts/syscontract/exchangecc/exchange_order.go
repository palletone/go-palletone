package exchangecc

import (
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/contracts/shim"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/shopspring/decimal"
)

const EXCHANGELIST_RECORD = "ExchangeOrder-"
const EXCHANGELIST_HISTORY = "ExchangeOrderHistory-"

type ExchangeOrder struct {
	Address           common.Address
	SaleAsset         *modules.Asset
	SaleAmount        uint64 //挂单时卖多少金额
	CurrentSaleAmount uint64 //当前还有多少Amount在卖
	WantAsset         *modules.Asset
	WantAmount        uint64 //挂单时需要多少金额
	CurrentWantAmount uint64 //当前还需要多少Amount
	ExchangeSn        string
	Status            byte //1挂单中，2成交完毕，0撤销
}

func (exo *ExchangeOrder) getDealSaleAmount() uint64 {
	return exo.SaleAmount - exo.CurrentSaleAmount
}
func (exo *ExchangeOrder) getDealWantAmount() uint64 {
	return exo.WantAmount - exo.CurrentWantAmount
}

type ExchangeOrderJson struct {
	Address           common.Address
	SaleAsset         string
	SaleAmount        decimal.Decimal
	CurrentSaleAmount decimal.Decimal
	DealSaleAmount    decimal.Decimal
	WantAsset         string
	WantAmount        decimal.Decimal
	DealWantAmount    decimal.Decimal
	CurrentWantAmount decimal.Decimal
	ExchangeSn        string
	Status            string
}

func convertSheet(exm ExchangeOrder) *ExchangeOrderJson {
	newSheet := ExchangeOrderJson{}
	newSheet.Address = exm.Address
	newSheet.SaleAsset = exm.SaleAsset.String()
	newSheet.CurrentSaleAmount = exm.SaleAsset.DisplayAmount(exm.CurrentSaleAmount)
	newSheet.DealSaleAmount = exm.SaleAsset.DisplayAmount(exm.getDealSaleAmount())
	newSheet.SaleAmount = exm.SaleAsset.DisplayAmount(exm.SaleAmount)
	newSheet.WantAsset = exm.WantAsset.String()
	newSheet.WantAmount = exm.WantAsset.DisplayAmount(exm.WantAmount)
	newSheet.CurrentWantAmount = exm.WantAsset.DisplayAmount(exm.CurrentWantAmount)
	newSheet.DealWantAmount = exm.WantAsset.DisplayAmount(exm.getDealWantAmount())
	newSheet.ExchangeSn = exm.ExchangeSn
	return &newSheet
}

//增加一个新订单
func SaveExchangeOrder(stub shim.ChaincodeStubInterface, order *ExchangeOrder) error {
	data, _ := rlp.EncodeToBytes(order)
	key := EXCHANGELIST_RECORD + order.ExchangeSn
	return stub.PutState(key, data)
}
func deleteExchangeOrder(stub shim.ChaincodeStubInterface, orderSn string) error {
	return stub.DelState(EXCHANGELIST_RECORD + orderSn)
}

//更新一个已有的订单
func UpdateExchangeOrder(stub shim.ChaincodeStubInterface, order *ExchangeOrder) error {
	if order.CurrentSaleAmount == 0 { //全部成交
		order.Status = 2
		err := deleteExchangeOrder(stub, order.ExchangeSn)
		if err != nil {
			return err
		}
		err = saveExchangeOrderHistory(stub, order)
		if err != nil {
			return err
		}
		return nil
	} else { //部分成交
		return SaveExchangeOrder(stub, order)
	}
}

//取消一个订单
func cancelExchangeOrder(stub shim.ChaincodeStubInterface, order *ExchangeOrder) error {
	order.Status = 0
	err := deleteExchangeOrder(stub, order.ExchangeSn)
	if err != nil {
		return err
	}
	err = saveExchangeOrderHistory(stub, order)
	if err != nil {
		return err
	}
	return nil
}

//获得订单列表
func getExchangeRecords(stub shim.ChaincodeStubInterface) ([]*ExchangeOrderJson, error) {
	kvs, err := stub.GetStateByPrefix(EXCHANGELIST_RECORD)
	if err != nil {
		return nil, err
	}
	result := make([]*ExchangeOrderJson, 0, len(kvs))
	for _, kv := range kvs {
		record := &ExchangeOrder{}
		err = rlp.DecodeBytes(kv.Value, record)
		if err != nil {
			return nil, err
		}
		jsSheet := convertSheet(*record)
		result = append(result, jsSheet)
	}
	return result, nil
}

//根据订单号获得订单
func getExchangeRecordBySn(stub shim.ChaincodeStubInterface, exchangeSn string) (*ExchangeOrder, error) {
	key := EXCHANGELIST_RECORD + exchangeSn
	value, err := stub.GetState(key)
	if err != nil {
		return nil, err
	}
	record := &ExchangeOrder{}
	err = rlp.DecodeBytes(value, record)
	if err != nil {
		return nil, err
	}
	return record, nil
}

//将一个Order移动到History有两种情况，
//1,该订单全部成交完毕
//2,该订单被Maker取消了
func saveExchangeOrderHistory(stub shim.ChaincodeStubInterface, order *ExchangeOrder) error {
	data, _ := rlp.EncodeToBytes(order)
	key := EXCHANGELIST_HISTORY + order.ExchangeSn
	return stub.PutState(key, data)
}
