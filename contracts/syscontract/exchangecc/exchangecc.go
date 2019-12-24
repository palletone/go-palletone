package exchangecc

import (
	"encoding/json"
	"errors"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/contracts/shim"
	"github.com/palletone/go-palletone/contracts/syscontract"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/ptnjson"
	"github.com/shopspring/decimal"
)

var myContractAddr = syscontract.ExchangeContractAddress.String()

//PCGTta3M4t3yXu8uRgkKvaWd2d8DS36t3ba
type ExchangeMgr struct {
}

func (p *ExchangeMgr) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}
func getPayToContract(stub shim.ChaincodeStubInterface) (*modules.Asset, uint64, error) {
	payAssets, _ := stub.GetInvokeTokens()
	for _, invokeAA := range payAssets {
		if invokeAA.Address == myContractAddr {
			return invokeAA.Asset, invokeAA.Amount, nil
		}
	}
	return nil, 0, errors.New("No pay to contract token")
}
func (p *ExchangeMgr) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	f, args := stub.GetFunctionAndParameters()

	switch f {
	case "maker": //create a maker
		if len(args) != 2 {
			return shim.Error("must input 2 args: [WantAsset][WantAmount]")
		}

		wantToken, err := modules.StringToAsset(args[0])
		if err != nil {
			return shim.Error("Invalid address string:" + args[0])
		}
		wantAmount, err := decimal.NewFromString(args[1])
		if err != nil {
			return shim.Error("Invalid address string:" + args[0])
		}
		err = p.Maker(stub, wantToken, wantAmount)
		if err != nil {
			return shim.Error("AddExchangeOrder error:" + err.Error())
		}

		return shim.Success(nil)
	case "getActiveOrderList": //列出订单列表
		result, err := p.GetActiveOrderList(stub)
		if err != nil {
			return shim.Error(err.Error())
		}
		data, _ := json.Marshal(result)
		return shim.Success(data)
	case "taker": //获取挂单信息,token互换
		if len(args) != 1 {
			return shim.Error("must input 1 args: [ExchangeSN]")
		}
		err := p.Taker(stub, args[0])
		if err != nil {
			return shim.Error(err.Error())
		}
		return shim.Success(nil)
	case "cancel": //撤销订单
		if len(args) != 1 {
			return shim.Error("must input 1 args: [ExchangeSN]")
		}
		err := p.Cancel(stub, args[0])
		if err != nil {
			return shim.Error(err.Error())
		}
		return shim.Success(nil)
	case "getOrderMatchList": //列出订单的成交记录
		if len(args) != 1 {
			return shim.Error("must input 1 args: [ExchangeSN]")
		}
		result, err := p.GetOrderMatchList(stub, args[0])
		if err != nil {
			return shim.Error(err.Error())
		}
		data, _ := json.Marshal(result)
		return shim.Success(data)
	case "payout": //付出Token
		if len(args) != 3 {
			return shim.Error("must input 3 args: Address,Amount,Asset")
		}
		if !isFoundationInvoke(stub) {
			return shim.Error("Foundation only")
		}
		addr, err := common.StringToAddress(args[0])
		if err != nil {
			return shim.Error("Invalid address string:" + args[0])
		}
		amount, err := decimal.NewFromString(args[1])
		if err != nil {
			return shim.Error("Invalid amount:" + args[1])
		}
		asset, err := modules.StringToAsset(args[2])
		if err != nil {
			return shim.Error("Invalid asset string:" + args[2])
		}
		err = p.Payout(stub, addr, amount, asset)
		if err != nil {
			return shim.Error(err.Error())
		}
		return shim.Success(nil)
	default:
		jsonResp := "{\"Error\":\"Unknown function " + f + "\"}"
		return shim.Error(jsonResp)
	}
}
func (p *ExchangeMgr) Maker(stub shim.ChaincodeStubInterface, wantAsset *modules.Asset, wantAmount decimal.Decimal) error {
	addr, err := stub.GetInvokeAddress()
	if err != nil {
		return errors.New("Invalid address string:" + err.Error())
	}

	saleToken, saleAmount, err := getPayToContract(stub)

	if err != nil {
		return err
	}

	newsheet := &ExchangeOrder{}
	newsheet.Address = addr
	newsheet.SaleAsset = saleToken
	newsheet.SaleAmount = saleAmount
	newsheet.WantAsset = wantAsset
	newsheet.WantAmount = wantAsset.Uint64Amount(wantAmount)
	txid := stub.GetTxID()
	newsheet.ExchangeSn = txid
	newsheet.Status = 1
	newsheet.CurrentWantAmount = newsheet.WantAmount
	newsheet.CurrentSaleAmount = newsheet.SaleAmount
	return p.AddExchangeOrder(stub, newsheet)
}

func (p *ExchangeMgr) Taker(stub shim.ChaincodeStubInterface, orderSn string) error {
	takerAddress, _ := stub.GetInvokeAddress()
	exchange, err := getExchangeRecordBySn(stub, orderSn)
	if err != nil {
		return errors.New("invalid/sold out/canceled exchange SN:" + orderSn)
	}
	takerPayAsset, takerPayAmount, err := getPayToContract(stub)
	if err != nil {
		return err
	}
	if !takerPayAsset.Equal(exchange.WantAsset) {
		return errors.New("current asset not match exchange order want asset")
	}
	//计算成交额，取Min（订单的CurrentSaleAmount，Taker支付的Amount）
	takerDealAmount := takerPayAmount
	if exchange.CurrentWantAmount < takerDealAmount {
		takerDealAmount = exchange.CurrentWantAmount
	}
	//已知Taker的成交额，计算Maker按汇率能获得多少Amount
	makerDealAmount := uint64(float64(takerDealAmount) * float64(exchange.SaleAmount) / float64(exchange.WantAmount))
	if takerDealAmount == exchange.CurrentWantAmount { //处理精度误差
		makerDealAmount = exchange.CurrentSaleAmount
	}
	//更新状态数据
	matchRecord := &MatchRecord{
		ExchangeOrderSn:  exchange.ExchangeSn,
		TakerReqId:       stub.GetTxID(),
		MakerMatchAmount: makerDealAmount,
		MakerMatchAsset:  exchange.SaleAsset,
		MakerAddress:     exchange.Address,
		TakerMatchAmount: takerDealAmount,
		TakerMatchAsset:  exchange.WantAsset,
		TakerAddress:     takerAddress,
	}
	err = saveMatchRecord(stub, matchRecord)
	if err != nil {
		return err
	}
	exchange.CurrentSaleAmount = exchange.CurrentSaleAmount - makerDealAmount
	exchange.CurrentWantAmount = exchange.CurrentWantAmount - takerDealAmount
	err = UpdateExchangeOrder(stub, exchange)
	if err != nil {
		return err
	}
	//产生对应的Payout
	takerChangeAmount := takerPayAmount - takerDealAmount
	if takerChangeAmount > 0 { //部分成交，剩余的打回
		err = stub.PayOutToken(takerAddress.String(), &modules.AmountAsset{
			Amount: takerChangeAmount,
			Asset:  takerPayAsset,
		}, 0)
		if err != nil {
			return err
		}
	}
	//Taker成交的部分付款
	err = stub.PayOutToken(takerAddress.String(), &modules.AmountAsset{
		Amount: makerDealAmount,
		Asset:  exchange.SaleAsset,
	}, 0)
	if err != nil {
		return err
	}
	//Maker成交的部分付款
	err = stub.PayOutToken(exchange.Address.String(), &modules.AmountAsset{
		Amount: takerDealAmount,
		Asset:  takerPayAsset,
	}, 0)
	if err != nil {
		return err
	}
	return nil
}

func (p *ExchangeMgr) GetActiveOrderList(stub shim.ChaincodeStubInterface) ([]*ExchangeOrderJson, error) {
	return getExchangeRecords(stub)
}
func (p *ExchangeMgr) GetOrderMatchList(stub shim.ChaincodeStubInterface, orderSn string) ([]*MatchRecordJson, error) {
	return getMatchRecordByOrderSn(stub, orderSn)
}
func (p *ExchangeMgr) Cancel(stub shim.ChaincodeStubInterface, orderSn string) error {
	exchange, err := getExchangeRecordBySn(stub, orderSn)
	if err != nil {
		return errors.New("invalid/sold out/canceled exchange SN:" + orderSn)
	}
	addr, err := stub.GetInvokeAddress()
	if addr != exchange.Address {
		return errors.New("you are not the owner")
	}
	err = cancelExchangeOrder(stub, exchange)
	if err != nil {
		return err
	}
	//未成交的金额退回
	err = stub.PayOutToken(exchange.Address.String(), &modules.AmountAsset{
		Amount: exchange.CurrentSaleAmount,
		Asset:  exchange.SaleAsset,
	}, 0)
	if err != nil {
		return err
	}
	return nil
}
func (p *ExchangeMgr) Payout(stub shim.ChaincodeStubInterface, addr common.Address, amount decimal.Decimal, asset *modules.Asset) error {
	uint64Amt := ptnjson.JsonAmt2AssetAmt(asset, amount)
	return stub.PayOutToken(addr.String(), &modules.AmountAsset{
		Amount: uint64Amt,
		Asset:  asset,
	}, 0)
}

func (p *ExchangeMgr) AddExchangeOrder(stub shim.ChaincodeStubInterface, sheet *ExchangeOrder) error {
	addr := sheet.Address
	if !KycUser(addr) {
		return errors.New("Please verify your ID")
	}
	err := SaveExchangeOrder(stub, sheet)
	if err != nil {
		return errors.New("saveRecord error:" + err.Error())
	}
	return nil
}
