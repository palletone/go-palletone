package auctioncc

import (
	"github.com/palletone/go-palletone/contracts/syscontract"
	"github.com/palletone/go-palletone/contracts/shim"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"github.com/shopspring/decimal"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/common"
	"encoding/json"
	"time"
	"github.com/palletone/go-palletone/common/log"
	"strconv"
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
		rewardAddress := common.Address{}
		if len(args[2]) > 0 {
			rewardAddress, err = common.StringToAddress(args[2]) //todo  可以为空
			if err != nil {
				return shim.Error("Invalid rewardAddress string:" + args[2])
			}
		}
		err = p.MakerFix(stub, wantToken, wantAmount, rewardAddress)
		if err != nil {
			return shim.Error("MakerFix error:" + err.Error())
		}
		return shim.Success(nil)
	case "taker_fix": //获取挂单信息,token互换
		if len(args) != 1 {
			return shim.Error("taker_fix must input 1 args: [AuctionSN]")
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
		startTime := ""
		endTime := ""
		if len(args[4]) > 0 {
			sTime, err := time.Parse("2006-01-02 15:04:05", args[4]) //todo  可以kon
			if err == nil {
				startTime = sTime.Format(TimeFormt)
			} else {
				return shim.Error("Invalid startTime string:" + args[4])
			}
		}
		if args[5] != "" {
			eTime, err := time.Parse("2006-01-02 15:04:05", args[5]) //todo  可以为空
			if err == nil {
				endTime = eTime.Format(TimeFormt)
			} else {
				return shim.Error("Invalid endTime string:" + args[5])
			}
		}
		log.Debugf("maker_auction startTime :[%s], endTime[%s]", startTime, endTime)
		rewardAddress := common.Address{}
		if len(args[6]) > 0 {
			rewardAddress, err = common.StringToAddress(args[6]) //todo  可以为空
			if err != nil {
				return shim.Error("Invalid rewardAddress string:" + args[6])
			}
		}
		err = p.MakerAuction(stub, wantToken, startAmount, targetAmount, stepAmount, startTime, endTime, rewardAddress)
		if err != nil {
			return shim.Error("AddAuctionOrder error:" + err.Error())
		}
		return shim.Success(nil)
	case "taker_auction": //参加竞拍
		if len(args) != 1 {
			return shim.Error("must input 1 args: [AuctionSN]")
		}
		err := p.TakerAuction(stub, args[0])
		if err != nil {
			return shim.Error(err.Error())
		}
		return shim.Success(nil)
	case "update_taker_auction": //增加竞拍token数量
		if len(args) != 1 {
			return shim.Error("must input 1 args: [AuctionSN]")
		}
		err := p.UpdateTakerAuction(stub, args[0])
		if err != nil {
			return shim.Error(err.Error())
		}
		return shim.Success(nil)
	case "stop_auction": //拍卖结束
		if len(args) != 1 {
			return shim.Error("must input 1 args: [AuctionSN]")
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
	case "cancel_all": //撤销订单
		result, err := p.GetActiveOrderList(stub, 0, 0)
		if err != nil {
			return shim.Error(err.Error())
		}
		for _, addr := range result {
			err := p.Cancel(stub, addr.AuctionSn)
			if err != nil {
				log.Debugf("cancelall, cancel %s fail:%s", addr.AuctionSn, err.Error())
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

	case "get_active_order_list": //列出挂单列表
		start, end := 0, 0
		if len(args) >= 3 {
			return shim.Error("must input < 3 arg")
		}
		if len(args) > 0 {
			s, err := strconv.Atoi(args[0])
			if err == nil {
				start = s
			}
		}
		if len(args) > 1 {
			e, err := strconv.Atoi(args[1])
			if err == nil {
				end = e
			}
		}
		result, err := p.GetActiveOrderList(stub, start, end)
		if err != nil {
			return shim.Error(err.Error())
		}
		data, _ := json.Marshal(result)
		return shim.Success(data)

	case "get_active_order_count": //列出挂单列表
		if len(args) != 0 {
			return shim.Error("must input0 arg")
		}
		data, err := p.GetActiveOrderCount(stub)
		if err != nil {
			return shim.Error(err.Error())
		}
		return shim.Success([]byte(strconv.Itoa(data)))

	case "get_active_orders_by_maker": //列出挂单列表
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

	case "get_active_order_by_sn": //列出指定ID挂单
		if len(args) != 1 {
			return shim.Error("must input 1 args: [Order ID]")
		}
		result, err := p.GetActiveOrdersByID(stub, args[0])
		if err != nil {
			return shim.Error(err.Error())
		}
		data, _ := json.Marshal(result)
		return shim.Success(data)

	case "get_history_order_count":
		if len(args) != 0 {
			return shim.Error("must input0 arg")
		}
		data, err := p.GetHistoryOrderCount(stub)
		if err != nil {
			return shim.Error(err.Error())
		}
		return shim.Success([]byte(strconv.Itoa(data)))

	case "get_history_order_list": //列出所有历史挂单
		start, end := 0, 0
		if len(args) >= 3 {
			return shim.Error("must input < 3 arg")
		}
		if len(args) > 0 {
			s, err := strconv.Atoi(args[0])
			if err == nil {
				start = s
			}
		}
		if len(args) > 1 {
			e, err := strconv.Atoi(args[1])
			if err == nil {
				end = e
			}
		}
		result, err := p.GetHistoryOrderList(stub, start, end)
		if err != nil {
			return shim.Error(err.Error())
		}
		data, _ := json.Marshal(result)
		return shim.Success(data)

	case "get_history_order_by_sn": //列出指定ID历史挂单
		if len(args) != 1 {
			return shim.Error("must input 1 args[Order ID]")
		}
		result, err := p.GetHistoryOrderBySn(stub, args[0])
		if err != nil {
			return shim.Error(err.Error())
		}
		data, _ := json.Marshal(result)
		return shim.Success(data)

	case "get_history_order_by_maker": //列出指定ID历史挂单
		if len(args) != 1 {
			return shim.Error("must input 1 args[MakerAddr]")
		}
		addr, err := common.StringToAddress(args[0])
		if err != nil {
			return shim.Error("Invalid address:" + err.Error())
		}
		result, err := p.GetHistoryOrderByMakerAddr(stub, addr)
		if err != nil {
			return shim.Error(err.Error())
		}
		data, _ := json.Marshal(result)
		return shim.Success(data)

	case "get_match_list_by_order_sn": //列出订单的成交记录
		if len(args) != 1 {
			return shim.Error("must input 1 args: [AuctionSN]")
		}
		result, err := p.GetMatchListByOrderSn(stub, args[0])
		if err != nil {
			return shim.Error(err.Error())
		}
		data, _ := json.Marshal(result)
		return shim.Success(data)
	case "get_all_match_list": //列出订单的成交记录
		start, end := 0, 0
		if len(args) >= 3 {
			return shim.Error("must input < 3 arg")
		}
		if len(args) > 0 {
			s, err := strconv.Atoi(args[0])
			if err == nil {
				start = s
			}
		}
		if len(args) > 1 {
			e, err := strconv.Atoi(args[1])
			if err == nil {
				end = e
			}
		}
		result, err := p.GetAllMatchList(stub, start, end)
		if err != nil {
			return shim.Error(err.Error())
		}
		data, _ := json.Marshal(result)
		return shim.Success(data)

	case "get_match_count":
		if len(args) != 0 {
			return shim.Error("must input0 arg")
		}
		data, err := p.GetMatchCount(stub)
		if err != nil {
			return shim.Error(err.Error())
		}
		return shim.Success([]byte(strconv.Itoa(data)))

	case "set_auction_mgr_address_list": //设置管理地址
		if len(args) <= 0 {
			return shim.Error("must input > 0 arg")
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

	case "get_auction_mgr_address_list": //获取管理地址
		if len(args) != 0 {
			return shim.Error("must input 0 arg")
		}
		result, err := getAuctionContractMgrAddress(stub)
		if err != nil {
			return shim.Error(err.Error())
		}
		data, _ := json.Marshal(result)
		return shim.Success(data)

	case "set_reward_rate": //设置拍卖资金费率--奖励
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

	case "set_destruction_rate": //设置拍卖资金费率--销毁
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

	case "get_reward_rate": //获取拍卖资金费率--奖励
		if len(args) != 0 {
			return shim.Error("must input 0 arg")
		}
		rate := getAuctionFeeRate(stub, 0)
		data, _ := json.Marshal(rate)
		return shim.Success(data)
	case "get_destruction_rate": //获取拍卖资金费率--奖励
		if len(args) != 0 {
			return shim.Error("must input 0 arg")
		}
		rate := getAuctionFeeRate(stub, 1)
		data, _ := json.Marshal(rate)
		return shim.Success(data)

	case "set_first_reward_fee_rate_level": //设置第一次交易奖励费率级别  第一交易奖励=正常交易金额*奖励费率*第一次奖励级别
		if len(args) != 1 {
			return shim.Error("must input 1 arg")
		}
		rate, err := decimal.NewFromString(args[0])
		if err != nil {
			return shim.Error("Invalid amount:" + args[0])
		}
		err = setFirstFeeRateLevel(stub, 0, rate)
		if err != nil {
			return shim.Error(err.Error())
		}
		data, _ := json.Marshal(rate)
		return shim.Success(data)

	case "set_first_destruction_fee_rate_level": //设置第一次交易销毁费率级别  第一交易销毁=正常交易金额*销毁费率*第一次销毁级别
		if len(args) != 1 {
			return shim.Error("must input 1 arg")
		}
		rate, err := decimal.NewFromString(args[0])
		if err != nil {
			return shim.Error("Invalid amount:" + args[0])
		}
		err = setFirstFeeRateLevel(stub, 1, rate)
		if err != nil {
			return shim.Error(err.Error())
		}
		data, _ := json.Marshal(rate)
		return shim.Success(data)

	case "get_first_reward_fee_rate_level":
		if len(args) != 0 {
			return shim.Error("must input 0 arg")
		}
		rate := getFirstFeeRateLevel(stub, 0)
		data, _ := json.Marshal(rate)
		return shim.Success(data)

	case "get_first_destruction_fee_rate_level":
		if len(args) != 0 {
			return shim.Error("must input 0 arg")
		}
		rate := getFirstFeeRateLevel(stub, 1)
		data, _ := json.Marshal(rate)
		return shim.Success(data)

	default:
		return shim.Error("no case")
	}

	return shim.Error("no case")
}

//
//func main() {
//	err := shim.Start(new(AuctionMgr))
//	if err != nil {
//		fmt.Printf("Error starting AuctionMgr chaincode: %s", err)
//	}
//}
