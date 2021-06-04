package auctioncc

import (
	"github.com/palletone/go-palletone/contracts/syscontract"
	"github.com/palletone/go-palletone/contracts/shim"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"github.com/shopspring/decimal"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/errors"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/ptnjson"
	"encoding/json"
	"fmt"
	"time"
)

var myContractAddr = syscontract.AuctionContractAddress.String()

//PCGTta3M4t3yXu8uRgkKvaWd2d8DSVFQsbL
type AuctionMgr struct {
}

func (p *AuctionMgr) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

func (p *AuctionMgr) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	f, args := stub.GetFunctionAndParameters()

	switch f {
	case "maker_fix": //create a maker
		if len(args) != 3 {
			return shim.Error("must input 2 args: [WantAsset][WantAmount][RewardAddress]")
		}
		wantToken, err := modules.StringToAsset(args[0])
		if err != nil {
			return shim.Error("Invalid wantToken string:" + args[0])
		}
		wantAmount, err := decimal.NewFromString(args[1])
		if err != nil {
			return shim.Error("Invalid wantAmount string:" + args[1])
		}
		rewardAddress, err := common.StringToAddress(args[2])
		if err != nil {
			return shim.Error("Invalid rewardAddress string:" + args[2])
		}
		err = p.MakerFix(stub, wantToken, wantAmount, rewardAddress)
		if err != nil {
			return shim.Error("AddAuctionOrder error:" + err.Error())
		}
		return shim.Success(nil)
	case "taker_fix": //获取挂单信息,token互换
		if len(args) != 1 {
			return shim.Error("must input 1 args: [AuctionSN]")
		}
		err := p.TakerFix(stub, args[0])
		if err != nil {
			return shim.Error(err.Error())
		}
		return shim.Success(nil)
	case "maker_auction": //挂单
		if len(args) != 7 {
			return shim.Error("must input 2 args: [WantAsset][StartAmount][TargetAmount][StepAmount][StartTime][EndTime][RewardAddress]")
		}
		wantToken, err := modules.StringToAsset(args[0])
		if err != nil {
			return shim.Error("Invalid WantAsset string:" + args[0])
		}
		startAmount, err := decimal.NewFromString(args[1])
		if err != nil {
			return shim.Error("Invalid StartAmount string:" + args[1])
		}
		targetAmount, err := decimal.NewFromString(args[2]) //todo  可以为空
		if err != nil {
			return shim.Error("Invalid targetAmount string:" + args[2])
		}
		stepAmount, err := decimal.NewFromString(args[3]) //todo  可以为空
		if err != nil {
			return shim.Error("Invalid StepAmount string:" + args[3])
		}
		startTime, err := time.Parse("2016-06-01 12:12:12", args[4]) //todo  可以为空
		if err != nil {
			return shim.Error("Invalid StartTime string:" + args[4])
		}
		endTime, err := time.Parse("2016-06-01 12:12:12", args[5]) //todo  可以为空
		if err != nil {
			return shim.Error("Invalid StartTime string:" + args[5])
		}
		rewardAddress, err := common.StringToAddress(args[6]) //todo  可以为空
		if err != nil {
			return shim.Error("Invalid rewardAddress string:" + args[6])
		}
		err = p.MakerAuction(stub, wantToken, startAmount, targetAmount, stepAmount, startTime, endTime, rewardAddress)
		if err != nil {
			return shim.Error("AddAuctionOrder error:" + err.Error())
		}
		return shim.Success(nil)

	case "taker_auction": //参加竞拍
		if len(args) != 2 {
			return shim.Error("must input 3 args: [AuctionSN][BidAmount]")
		}
		err := p.TakerAuction(stub, args[0], args[1])
		if err != nil {
			return shim.Error(err.Error())
		}
		return shim.Success(nil)

	case "stop_auction": //拍卖结束
		return shim.Success(nil)

	case "cancel": //撤销订单
		if len(args) != 1 {
			return shim.Error("must input 1 args: [AuctionSN]")
		}
		err := p.Cancel(stub, args[0])
		if err != nil {
			return shim.Error(err.Error())
		}
		return shim.Success(nil)
	case "cancelall": //撤销订单
		result, err := p.GetActiveOrderList(stub)
		if err != nil {
			return shim.Error(err.Error())
		}
		for _, addr := range result {
			err := p.Cancel(stub, addr.AuctionSn)
			if err != nil {
				return shim.Error(err.Error())
			}
		}
		return shim.Success(nil)
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

	case "getActiveOrderList": //列出订单列表
		result, err := p.GetActiveOrderList(stub)
		if err != nil {
			return shim.Error(err.Error())
		}
		data, _ := json.Marshal(result)
		return shim.Success(data)
	case "getActiveOrdersByMaker": //列出订单列表
		if len(args) != 1 {
			return shim.Error("must input 1 args: [maker address]")
		}
		addr, err := common.StringToAddress(args[0])
		if err != nil {
			return shim.Error("Invalid address:" + err.Error())
		}
		result, err := p.GetActiveOrdersByMaker(stub, addr)
		if err != nil {
			return shim.Error(err.Error())
		}
		data, _ := json.Marshal(result)
		return shim.Success(data)
	case "getOrderMatchList": //列出订单的成交记录
		if len(args) != 1 {
			return shim.Error("must input 1 args: [AuctionSN]")
		}
		result, err := p.GetOrderMatchList(stub, args[0])
		if err != nil {
			return shim.Error(err.Error())
		}
		data, _ := json.Marshal(result)
		return shim.Success(data)
	case "getAllMatchList": //列出订单的成交记录
		if len(args) != 0 {
			return shim.Error("must input 0 arg")
		}
		result, err := p.GetAllMatchList(stub)
		if err != nil {
			return shim.Error(err.Error())
		}
		data, _ := json.Marshal(result)
		return shim.Success(data)
	default:
		return shim.Error("no case")
	}

	return shim.Error("no case")
}

func (p *AuctionMgr) MakerFix(stub shim.ChaincodeStubInterface, wantAsset *modules.Asset, wantAmount decimal.Decimal, rewardAddress common.Address) error {
	addr, err := stub.GetInvokeAddress()
	if err != nil {
		return errors.New("Invalid address string:" + err.Error())
	}

	saleToken, saleAmount, err := getPayToContract(stub)
	if err != nil {
		return err
	}
	txid := stub.GetTxID()
	order := &AuctionOrder{}
	order.AuctionType = 1 //fix
	order.Address = addr
	order.SaleAsset = saleToken
	order.SaleAmount = saleAmount
	order.WantAsset = wantAsset
	order.WantAmount = wantAsset.Uint64Amount(wantAmount)
	order.RewardAddress = rewardAddress
	order.AuctionSn = txid
	order.Status = 1

	return p.AddAuctionOrder(stub, order)
}

func (p *AuctionMgr) TakerFix(stub shim.ChaincodeStubInterface, orderSn string) error {
	takerAddress, _ := stub.GetInvokeAddress()
	auction, err := getAuctionRecordBySn(stub, orderSn)
	if err != nil {
		return errors.New("invalid/sold out/canceled auction SN:" + orderSn)
	}
	takerPayAsset, takerPayAmount, err := getPayToContract(stub)
	if err != nil {
		return err
	}
	//检查assert是否是否相同
	if !takerPayAsset.Equal(auction.WantAsset) {
		return errors.New("current asset not match takerFix order want asset")
	}

	//检查金额是否满足
	if takerPayAmount < auction.WantAmount {
		return errors.New("TakerFix, takerPayAmount < auction.WantAmount")
	}

	//计算奖励和销毁的费用
	var rewardAmount uint64 = 0
	if !auction.RewardAddress.IsZero() {
		rewardAmount = auction.WantAmount * 1 / 10 //todo   2 / 10
	}
	destructionAmount := auction.WantAmount * 1 / 10 //todo
	remainAmount := auction.WantAmount - rewardAmount - destructionAmount

	feeUse := AuctionFeeUse{
		Asset:             auction.WantAsset,
		RewardAmount:      rewardAmount,      //auction.WantAsset.DisplayAmount(10), //auction.WantAsset.DisplayAmount(auction.SaleAmount)
		DestructionAmount: destructionAmount, //auction.WantAsset.DisplayAmount(5),
	}

	//更新状态数据
	matchRecord := &MatchRecord{
		AuctionOrderSn:   auction.AuctionSn,
		TakerReqId:       stub.GetTxID(),
		MakerAddress:     auction.Address,
		MakerAsset:       auction.SaleAsset,
		MakerAssetAmount: auction.SaleAmount,
		TakerAddress:     takerAddress,
		TakerAsset:       auction.WantAsset,
		TakerAssetAmount: takerPayAmount,
		FeeUse:           feeUse,
	}
	err = saveMatchRecord(stub, matchRecord)
	if err != nil {
		return err
	}
	auction.Status = 2
	err = UpdateAuctionOrder(stub, auction)
	if err != nil {
		return err
	}
	//交换给taker的
	err = stub.PayOutToken(takerAddress.String(), &modules.AmountAsset{
		Amount: auction.SaleAmount,
		Asset:  auction.SaleAsset,
	}, 0)
	if err != nil {
		return err
	}

	//返还taker多余的
	takerChangeAmount := takerPayAmount - auction.WantAmount
	if takerChangeAmount > 0 { //剩余的打回
		err = stub.PayOutToken(takerAddress.String(), &modules.AmountAsset{
			Amount: takerChangeAmount,
			Asset:  takerPayAsset,
		}, 0)
		if err != nil {
			return err
		}
	}
	//卖出的
	if remainAmount > 0 {
		err = stub.PayOutToken(auction.Address.String(), &modules.AmountAsset{
			Amount: remainAmount,
			Asset:  auction.WantAsset,
		}, 0)
		if err != nil {
			return err
		}
	}
	//奖励
	if feeUse.RewardAmount > 0 {
		err = stub.PayOutToken(auction.RewardAddress.String(), &modules.AmountAsset{
			Amount: feeUse.RewardAmount,
			Asset:  feeUse.Asset,
		}, 0)
		if err != nil {
			return err
		}
	}
	//销毁
	if feeUse.DestructionAmount > 0 {
		err = stub.PayOutToken(DestructionAddress, &modules.AmountAsset{
			Amount: feeUse.DestructionAmount,
			Asset:  feeUse.Asset,
		}, 0)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *AuctionMgr) MakerAuction(stub shim.ChaincodeStubInterface, wantAsset *modules.Asset, wantAmount, targetAmount, stepAmount decimal.Decimal, startTime, endTime time.Time, rewardAddress common.Address) error {
	addr, err := stub.GetInvokeAddress()
	if err != nil {
		return errors.New("Invalid address string:" + err.Error())
	}
	saleToken, saleAmount, err := getPayToContract(stub)
	if err != nil {
		return err
	}
	txid := stub.GetTxID()
	order := &AuctionOrder{}
	order.AuctionType = 2
	order.Address = addr
	order.SaleAsset = saleToken
	order.SaleAmount = saleAmount
	order.WantAsset = wantAsset
	order.WantAmount = wantAsset.Uint64Amount(wantAmount)
	order.TargetAmount = wantAsset.Uint64Amount(targetAmount)
	order.StepAmount = wantAsset.Uint64Amount(stepAmount)
	order.StartTime = startTime
	order.EndTime = endTime
	order.RewardAddress = rewardAddress
	order.AuctionSn = txid
	order.Status = 1

	return p.AddAuctionOrder(stub, order)
}

func (p *AuctionMgr) TakerAuction(stub shim.ChaincodeStubInterface, orderSn, bidAmount string) error {
	takerAddress, _ := stub.GetInvokeAddress()
	auction, err := getAuctionRecordBySn(stub, orderSn)
	if err != nil {
		return errors.New("invalid/sold out/canceled takerAuction SN:" + orderSn)
	}

	//检查挂单是否有效
	if auction.Status != 1 {
		return errors.New("takerAuction order is invalid")
	}

	//检查时间是否有效
	now, err := stub.GetTxTimestamp(10)
	if err != nil {
		return errors.New("takerAuction, GetTxTimestamp err:" + err.Error())
	}
	//nowTime := time.Unix(now.Seconds, 0)
	//if nowTime.Before(auction.StartTime) || nowTime.After(auction.EndTime) {
	//	//	log.Debugf("")
	//	//	return errors.New("takerAuction, Not in the auction time zone now")
	//	//}

	if now.Seconds < auction.StartTime.Unix() {
		return errors.New("takerAuction, now.Seconds < auction.StartTime")
	} else if now.Seconds > auction.EndTime.Unix() {
		return errors.New("takerAuction, now.Seconds > auction.EndTime")
	}

	takerPayAsset, takerPayAmount, err := getPayToContract(stub)
	if err != nil {
		return err
	}
	//检查assert是否相同
	if !takerPayAsset.Equal(auction.WantAsset) {
		return errors.New("takerAuction, current asset not match auction order want asset")
	}

	//检查提交金额数量是否满足
	if takerPayAmount < auction.WantAmount {
		return fmt.Errorf("takerAuction, auction[%s] takerPayAmount[%d] < WantAmount[%d]", auction.AuctionSn, takerPayAmount, auction.WantAmount)
	}

	//
	//if auction.AuctionType == 2 { //英式拍卖
	//	//金额检查
	//	if takerPayAmount < auction.WantAmount {
	//		return fmt.Errorf("takerAuction, auction[%s] takerPayAmount[%d] < WantAmount[%d]", auction.AuctionSn, takerPayAmount, auction.WantAmount)
	//	}
	//
	//	//记录订单
	//}
	//if auction.AuctionType == 3 { //荷兰式拍卖
	//	//检查费用
	//}

	//计算奖励和销毁的费用
	var rewardAmount uint64 = 0
	if !auction.RewardAddress.IsZero() {
		rewardAmount = auction.WantAmount * 1 / 10 //todo
	}
	destructionAmount := auction.WantAmount * 1 / 10 //todo

	feeUse := AuctionFeeUse{
		Asset:             auction.WantAsset,
		RewardAmount:      rewardAmount,      //auction.WantAsset.DisplayAmount(10), //auction.WantAsset.DisplayAmount(auction.SaleAmount)
		DestructionAmount: destructionAmount, //auction.WantAsset.DisplayAmount(5),
	}
	//更新状态数据
	submitRecord := &MatchRecord{
		AuctionOrderSn:   auction.AuctionSn,
		TakerReqId:       stub.GetTxID(),
		MakerAddress:     auction.Address,
		MakerAsset:       auction.SaleAsset,
		MakerAssetAmount: auction.SaleAmount,
		TakerAddress:     takerAddress,
		TakerAsset:       takerPayAsset,
		TakerAssetAmount: takerPayAmount,
		FeeUse:           feeUse,
	}

	err = saveMatchRecord(stub, submitRecord)
	if err != nil {
		return err
	}

	return nil
}

func (p *AuctionMgr) GetActiveOrderList(stub shim.ChaincodeStubInterface) ([]*AuctionOrderJson, error) {
	return getAllAuctionOrder(stub)
}
func (p *AuctionMgr) GetActiveOrdersByMaker(stub shim.ChaincodeStubInterface, addr common.Address) ([]*AuctionOrderJson, error) {
	return getAuctionOrderByAddress(stub, addr)
}

func (p *AuctionMgr) GetHistoryOrderList(stub shim.ChaincodeStubInterface) ([]*AuctionOrderJson, error) {
	return getAllHistoryOrder(stub)
}

func (p *AuctionMgr) GetOrderMatchList(stub shim.ChaincodeStubInterface, orderSn string) ([]*MatchRecordJson, error) {
	return getMatchRecordByOrderSn(stub, orderSn)
}
func (p *AuctionMgr) GetAllMatchList(stub shim.ChaincodeStubInterface) ([]*MatchRecordJson, error) {
	return getAllMatchRecord(stub)
}

func (p *AuctionMgr) Cancel(stub shim.ChaincodeStubInterface, orderSn string) error {
	auction, err := getAuctionRecordBySn(stub, orderSn)
	if err != nil {
		return errors.New("invalid/sold out/canceled exchange SN:" + orderSn)
	}
	addr, err := stub.GetInvokeAddress()
	if addr != auction.Address {
		return errors.New("you are not the owner")
	}
	err = cancelAuctionOrder(stub, auction)
	if err != nil {
		return err
	}
	//未成交的金额退回
	err = stub.PayOutToken(auction.Address.String(), &modules.AmountAsset{
		Amount: auction.SaleAmount,
		Asset:  auction.SaleAsset,
	}, 0)
	if err != nil {
		return err
	}
	return nil
}
func (p *AuctionMgr) Payout(stub shim.ChaincodeStubInterface, addr common.Address, amount decimal.Decimal, asset *modules.Asset) error {
	uint64Amt := ptnjson.JsonAmt2AssetAmt(asset, amount)
	return stub.PayOutToken(addr.String(), &modules.AmountAsset{
		Amount: uint64Amt,
		Asset:  asset,
	}, 0)
}

func (p *AuctionMgr) AddAuctionOrder(stub shim.ChaincodeStubInterface, sheet *AuctionOrder) error {
	addr := sheet.Address
	if !KycUser(addr) {
		return errors.New("Please verify your ID")
	}
	err := SaveAuctionOrder(stub, sheet)
	if err != nil {
		return errors.New("saveRecord error:" + err.Error())
	}
	return nil
}
