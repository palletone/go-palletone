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
	"github.com/palletone/go-palletone/common/log"
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
			return shim.Error("must input 2 args: [AuctionSN][BidAmount]")
		}
		err := p.TakerAuction(stub, args[0], args[1])
		if err != nil {
			return shim.Error(err.Error())
		}
		return shim.Success(nil)
	case "stop_auction": //拍卖结束
		if len(args) != 2 {
			return shim.Error("must input 2 args: [AuctionSN]")
		}
		err := p.StopAuction(stub, args[0])
		if err != nil {
			return shim.Error(err.Error())
		}
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

	case "setAuctionMgrAddressList": //设置管理地址
		if len(args) != 0 {
			return shim.Error("must input 0 arg")
		}
		ads := common.Addresses{}
		for _, arg := range args {
			addr, err := common.StringToAddress(arg)
			if err == nil {
				ads = append(ads, addr)
			}
		}
		err := setAuctionContractMgrAddress(stub, ads)
		if err != nil {
			return shim.Error(err.Error())
		}
		data, _ := json.Marshal(ads)
		return shim.Success(data)

	case "getAuctionMgrAddressList": //获取管理地址
		if len(args) != 0 {
			return shim.Error("must input 0 arg")
		}
		result, err := getAuctionContractMgrAddress(stub)
		if err != nil {
			return shim.Error(err.Error())
		}
		data, _ := json.Marshal(result)
		return shim.Success(data)

	case "setRewardRate": //设置拍卖资金费率--奖励
		if len(args) != 1 {
			return shim.Error("must input 1 arg")
		}
		rate, err := decimal.NewFromString(args[0])
		if err != nil {
			return shim.Error("Invalid amount:" + args[0])
		}
		err = setAuctionFeeRate(stub, 0, rate)
		if err != nil {
			return shim.Error(err.Error())
		}
		data, _ := json.Marshal(rate)
		return shim.Success(data)

	case "setDestructionRate": //设置拍卖资金费率--销毁
		if len(args) != 1 {
			return shim.Error("must input 1 arg")
		}
		rate, err := decimal.NewFromString(args[0])
		if err != nil {
			return shim.Error("Invalid amount:" + args[0])
		}
		err = setAuctionFeeRate(stub, 1, rate)
		if err != nil {
			return shim.Error(err.Error())
		}
		data, _ := json.Marshal(rate)
		return shim.Success(data)

	case "getRewardRate": //获取拍卖资金费率--奖励
		if len(args) != 0 {
			return shim.Error("must input 0 arg")
		}
		rate := getAuctionFeeRate(stub, 0)
		data, _ := json.Marshal(rate)
		return shim.Success(data)
	case "getDestructionRate": //获取拍卖资金费率--奖励
		if len(args) != 0 {
			return shim.Error("must input 0 arg")
		}
		rate := getAuctionFeeRate(stub, 1)
		data, _ := json.Marshal(rate)
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
	t, _ := stub.GetTxTimestamp(10)
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
	order.CreateTime = time.Unix(t.Seconds, 0)
	//return order.CreateTime.UTC().Format(modules.Layout2)

	log.Debugf("MakerFix:\n"+
		"AuctionType:%d\n"+
		"Address:%s\n"+
		"SaleAsset:%s\n"+
		"SaleAmount:%d\n"+
		"WantAsset:%s\n"+
		"WantAmount:%d\n"+
		"RewardAddress:%s\n"+
		"AuctionSn:%s\n"+
		"Status:%d\n",
		order.AuctionType, order.Address.String(), order.SaleAsset.String(), order.SaleAmount, order.WantAsset.String(),
		order.WantAmount, order.RewardAddress.String(), order.AuctionSn, order.Status)

	return p.AddAuctionOrder(stub, order)
}

func (p *AuctionMgr) TakerFix(stub shim.ChaincodeStubInterface, orderSn string) error {
	takerAddress, _ := stub.GetInvokeAddress()
	//takerGasFee, _ := stub.GetInvokeFees()
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
		rate := getAuctionFeeRate(stub, 0)
		amount := decimal.NewFromFloat(float64(auction.WantAmount)).Mul(rate).IntPart()
		rewardAmount = uint64(amount)
		//rewardAmount = auction.WantAmount * rate //todo   2 / 10
	}
	rate := getAuctionFeeRate(stub, 1)
	amount := decimal.NewFromFloat(float64(auction.WantAmount)).Mul(rate).IntPart()
	destructionAmount := uint64(amount)

	now, err := stub.GetTxTimestamp(10)
	if err != nil {
		return errors.New("takerAuction, GetTxTimestamp err:" + err.Error())
	}

	//destructionAmount := auction.WantAmount * 1 / 10 //todo
	//remainAmount := auction.WantAmount - rewardAmount - destructionAmount
	desAddr, _ := common.StringToAddress(DestructionAddress)
	feeUse := AuctionFeeUse{
		Asset:              auction.WantAsset,
		RewardAddress:      auction.RewardAddress,
		RewardAmount:       rewardAmount, //auction.WantAsset.DisplayAmount(10), //auction.WantAsset.DisplayAmount(auction.SaleAmount)
		DestructionAddress: desAddr,
		DestructionAmount:  destructionAmount, //auction.WantAsset.DisplayAmount(5),
	}

	//更新状态数据
	matchRecord := &MatchRecord{
		AuctionType:      auction.AuctionType,
		AuctionOrderSn:   auction.AuctionSn,
		TakerReqId:       stub.GetTxID(),
		MakerAddress:     auction.Address,
		MakerAsset:       auction.SaleAsset,
		MakerAssetAmount: auction.SaleAmount,
		TakerAddress:     takerAddress,
		TakerAsset:       takerPayAsset,
		TakerAssetAmount: takerPayAmount, //todo   - takerGasFee.Amount
		FeeUse:           feeUse,
		recordTime:       now.Seconds,
	}

	log.Debugf("TakerFix:\n"+
		"AuctionType:%d\n"+
		"AuctionOrderSn:%s\n"+
		"TakerReqId:%s\n"+
		"MakerAddress:%s\n"+
		"MakerAsset:%s\n"+
		"MakerAssetAmount:%d\n"+
		"TakerAddress:%s\n"+
		"TakerAsset:%s\n"+
		"TakerAssetAmount:%d\n"+
		"FeeUse.asset:%s\n"+
		"FeeUse.RewardAddress:%s\n"+
		"FeeUse.RewardAmount:%d\n"+
		"FeeUse.DestructionAddress:%s\n"+
		"FeeUse.DestructionAmount:%d\n",
		matchRecord.AuctionType, matchRecord.AuctionOrderSn, matchRecord.TakerReqId, matchRecord.MakerAddress.String(), matchRecord.MakerAsset.String(), matchRecord.MakerAssetAmount,
		matchRecord.TakerAddress.String(), matchRecord.TakerAsset.String(), matchRecord.TakerAssetAmount,
		matchRecord.FeeUse.Asset.String(), matchRecord.FeeUse.RewardAddress.String(), matchRecord.FeeUse.RewardAmount, matchRecord.FeeUse.DestructionAddress.String(), matchRecord.FeeUse.DestructionAmount)

	err = saveMatchRecord(stub, matchRecord)
	if err != nil {
		return err
	}
	auction.Status = 2
	err = UpdateAuctionOrder(stub, auction)
	if err != nil {
		return err
	}

	return transferTokensProcess(stub, auction, matchRecord)
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
	t, _ := stub.GetTxTimestamp(10)
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
	order.CreateTime = time.Unix(t.Seconds, 0)

	log.Debugf("MakerAuction:\n"+
		"AuctionType:%d\n"+
		"Address:%s\n"+
		"SaleAsset:%s\n"+
		"SaleAmount:%d\n"+
		"WantAsset:%s\n"+
		"WantAmount:%d\n"+
		"RewardAddress:%s\n"+
		"AuctionSn:%s\n"+
		"Status:%d\n",
		order.AuctionType, order.Address.String(), order.SaleAsset.String(), order.SaleAmount, order.WantAsset.String(),
		order.WantAmount, order.RewardAddress.String(), order.AuctionSn, order.Status)

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

	//if auction.AuctionType == 2 { //英式拍卖
	//if auction.AuctionType == 3 { //荷兰式拍卖

	//计算奖励和销毁的费用
	var rewardAmount uint64 = 0
	if !auction.RewardAddress.IsZero() {
		rewardAmount = auction.WantAmount * 1 / 10 //todo
	}
	destructionAmount := auction.WantAmount * 1 / 10 //todo
	desAddr, _ := common.StringToAddress(DestructionAddress)
	feeUse := AuctionFeeUse{
		Asset:              auction.WantAsset,
		RewardAddress:      auction.RewardAddress,
		RewardAmount:       rewardAmount, //auction.WantAsset.DisplayAmount(10), //auction.WantAsset.DisplayAmount(auction.SaleAmount)
		DestructionAddress: desAddr,
		DestructionAmount:  destructionAmount, //auction.WantAsset.DisplayAmount(5),
	}

	//更新状态数据
	submitRecord := &MatchRecord{
		AuctionType:      auction.AuctionType,
		AuctionOrderSn:   auction.AuctionSn,
		TakerReqId:       stub.GetTxID(),
		MakerAddress:     auction.Address,
		MakerAsset:       auction.SaleAsset,
		MakerAssetAmount: auction.SaleAmount,
		TakerAddress:     takerAddress,
		TakerAsset:       takerPayAsset,
		TakerAssetAmount: takerPayAmount,
		FeeUse:           feeUse,
		recordTime:       now.Seconds,
	}

	err = saveMatchRecord(stub, submitRecord)
	if err != nil {
		return err
	}

	return nil
}

func (p *AuctionMgr) StopAuction(stub shim.ChaincodeStubInterface, orderSn string) error {
	invokeAddress, _ := stub.GetInvokeAddress()
	auction, err := getAuctionRecordBySn(stub, orderSn)
	if err != nil {
		return errors.New("invalid/sold out/canceled StopAuction SN:" + orderSn)
	}
	//检查合约状态是否有效
	if auction.Status != 1 {
		return errors.New("StopAuction order is invalid")
	}
	if !invokeAddress.Equal(auction.Address) || !isFoundationInvoke(stub) {
		return errors.New("StopAuction addr are not the owner or not foundation")
	}

	//按金额、时间获取成交记录
	matchRecords, err := getMatchRecordByOrderSn(stub, orderSn)
	if err != nil {
		return errors.New("StopAuction getMatchRecordByOrderSn err:" + err.Error())
	}
	maxRecord := getMaxAmountRecord(matchRecords)

	//费用扣除、退款
	for _, record := range matchRecords {
		if record.TakerReqId == maxRecord.TakerReqId {
			err = transferTokensProcess(stub, auction, maxRecord)
			if err != nil {
				return errors.New("StopAuction,TransferTokensProcess err:" + err.Error())
			}
			continue
		}
		if record.TakerAssetAmount > 0 { //剩余的打回
			err = stub.PayOutToken(record.TakerAddress.String(), &modules.AmountAsset{
				Amount: record.TakerAssetAmount,
				Asset:  record.TakerAsset,
			}, 0)
			if err != nil {
				return err
			}
		}
	}

	//更新auctionOrder
	auction.Status = 2
	err = UpdateAuctionOrder(stub, auction)
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
	return getMatchRecordJsonByOrderSn(stub, orderSn)
}
func (p *AuctionMgr) GetAllMatchList(stub shim.ChaincodeStubInterface) ([]*MatchRecordJson, error) {
	return getAllMatchRecordJson(stub)
}

func (p *AuctionMgr) Cancel(stub shim.ChaincodeStubInterface, orderSn string) error {
	auction, err := getAuctionRecordBySn(stub, orderSn)
	if err != nil {
		return errors.New("invalid/sold out/canceled auction SN:" + orderSn)
	}
	//检查挂单状态是否有效
	if auction.Status != 1 {
		return errors.New(" auction.Status is invalid/ auction SN:" + orderSn)
	}
	//检查是否有成交记录
	//match, err := getAllMatchRecordJson(stub )
	//if isInMatchRecord(orderSn){
	//
	//}

	//检查是否是Maker或者基金会
	addr, err := stub.GetInvokeAddress()
	if addr != auction.Address || !isFoundationInvoke(stub) {
		return errors.New("you are not the owner or not foundation")
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
