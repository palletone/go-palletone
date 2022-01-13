/*
	This file is part of go-palletone.
	go-palletone is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.
	go-palletone is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.
	You should have received a copy of the GNU General Public License
	along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
*/

/*
 * Copyright IBM Corp. All Rights Reserved.
 * @author PalletOne core developers <dev@pallet.one>
 * @date 2018
 */
package tokentradecc

import (
	"github.com/shopspring/decimal"
	"github.com/palletone/go-palletone/contracts/shim"
	"github.com/palletone/go-palletone/common"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"github.com/palletone/go-palletone/dag/modules"
	"strconv"
	"encoding/json"
	"github.com/palletone/go-palletone/common/log"
)

//PCGTta3M4t3yXu8uRgkKvaWd2d8DSt3GNej
type TokenTrade struct {
}

func (p *TokenTrade) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

func (p *TokenTrade) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	f, args := stub.GetFunctionAndParameters()
	switch f {
	case "maker_trade": //create a maker
		if len(args) != 5 {
			return shim.Error("tokenTrade, must input 5 args: [TradeType][WantAsset][WantAmount][Scale][RewardAddress]")
		}
		tradeType := common.Str2Uint64(args[0])
		wantToken, err := modules.StringToAsset(args[1])
		if err != nil {
			return shim.Error("tokenTrade, Invalid wantToken string:" + args[1])
		}
		//wantAmount := common.Str2Uint64(args[2])
		wantAmount, err := decimal.NewFromString(args[2])
		if err != nil {
			return shim.Error("tokenTrade, Invalid wantAmount string:" + args[2])
		}
		scale := args[3]
		rewardAddress := common.Address{}
		if len(args[4]) > 0 {
			rewardAddress, err = common.StringToAddress(args[4]) //todo  可以为空
			if err != nil {
				return shim.Error("tokenTrade, Invalid rewardAddress string:" + args[4])
			}
		}
		err = p.MakerTrade(stub, byte(tradeType), wantToken, wantAmount, scale, rewardAddress)
		if err != nil {
			return shim.Error("tokenTrade, MakerFix error:" + err.Error())
		}
		return shim.Success(nil)
	case "taker_trade": //获取挂单信息,token转账
		if len(args) != 2 {
			return shim.Error("tokenTrade, taker_fix must input 2 args:  [TradeType][TradeSN]")
		}
		tradeType := common.Str2Uint64(args[0])
		orderSn := args[1]
		if "0x" != orderSn[0:2] && "0X" != orderSn[0:2] {
			orderSn = "0x" + orderSn
		}
		err := p.TakerTrade(stub, byte(tradeType), orderSn)
		if err != nil {
			return shim.Error(err.Error())
		}
		return shim.Success(nil)

	case "cancel": //撤销订单
		if len(args) != 1 {
		return shim.Error("tokenTrade, must input 1 args: [TradeSN]")
	}
		orderSn := args[0]
		if "0x" != orderSn[0:2] && "0X" != orderSn[0:2] {
			orderSn = "0x" + orderSn
		}
		err := p.Cancel(stub, orderSn)
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
			err := p.Cancel(stub, addr.TradeSn)
			if err != nil {
				log.Debugf("tokenTrade, cancelall, cancel %s fail:%s", addr.TradeSn, err.Error())
			}
		}
		return shim.Success(nil)
	case "payout": //付出Token
		if len(args) != 3 {
			return shim.Error("tokenTrade, must input 3 args: Address,Amount,Asset")
		}
		if !isFoundationInvoke(stub) {
			return shim.Error("tokenTrade, Foundation only")
		}
		addr, err := common.StringToAddress(args[0])
		if err != nil {
			return shim.Error("tokenTrade, Invalid address string:" + args[0])
		}
		amount, err := decimal.NewFromString(args[1])
		if err != nil {
			return shim.Error("tokenTrade, Invalid amount:" + args[1])
		}
		asset, err := modules.StringToAsset(args[2])
		if err != nil {
			return shim.Error("tokenTrade, Invalid asset string:" + args[2])
		}
		err = p.Payout(stub, addr, amount, asset)
		if err != nil {
			return shim.Error(err.Error())
		}
		return shim.Success(nil)

	case "get_active_order_list": //列出挂单列表
		start, end := 0, 0
		if len(args) >= 3 {
			return shim.Error("tokenTrade, must input < 3 arg")
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
			return shim.Error("tokenTrade, must input0 arg")
		}
		data, err := p.GetActiveOrderCount(stub)
		if err != nil {
			return shim.Error(err.Error())
		}
		return shim.Success([]byte(strconv.Itoa(data)))

	case "get_active_orders_by_maker": //列出挂单列表
		if len(args) != 1 {
			return shim.Error("tokenTrade, must input 1 args: [maker address]")
		}
		addr, err := common.StringToAddress(args[0])
		if err != nil {
			return shim.Error("tokenTrade, Invalid address:" + err.Error())
		}
		result, err := p.GetActiveOrdersByMaker(stub, addr)
		if err != nil {
			return shim.Error(err.Error())
		}
		data, _ := json.Marshal(result)
		return shim.Success(data)

	case "get_active_order_by_sn": //列出指定ID挂单
		if len(args) != 1 {
			return shim.Error("tokenTrade, must input 1 args: [Order ID]")
		}
		orderSn := args[0]
		if "0x" != orderSn[0:2] && "0X" != orderSn[0:2] {
			orderSn = "0x" + orderSn
		}
		result, err := p.GetActiveOrdersByID(stub, orderSn)
		if err != nil {
			return shim.Error(err.Error())
		}
		data, _ := json.Marshal(result)
		return shim.Success(data)

	case "get_history_order_count":
		if len(args) != 0 {
			return shim.Error("tokenTrade, must input0 arg")
		}
		data, err := p.GetHistoryOrderCount(stub)
		if err != nil {
			return shim.Error(err.Error())
		}
		return shim.Success([]byte(strconv.Itoa(data)))

	case "get_history_order_list": //列出所有历史挂单
		start, end := 0, 0
		if len(args) >= 3 {
			return shim.Error("tokenTrade, must input < 3 arg")
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
			return shim.Error("tokenTrade, must input 1 args[Order ID]")
		}
		orderSn := args[0]
		if "0x" != orderSn[0:2] && "0X" != orderSn[0:2] {
			orderSn = "0x" + orderSn
		}
		result, err := p.GetHistoryOrderBySn(stub, orderSn)
		if err != nil {
			return shim.Error(err.Error())
		}
		data, _ := json.Marshal(result)
		return shim.Success(data)

	case "get_history_order_by_maker": //列出指定ID历史挂单
		if len(args) != 1 {
			return shim.Error("tokenTrade, must input 1 args[MakerAddr]")
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
			return shim.Error("tokenTrade, must input 1 args: [TradeSN]")
		}
		orderSn := args[0]
		if "0x" != orderSn[0:2] && "0X" != orderSn[0:2] {
			orderSn = "0x" + orderSn
		}
		result, err := p.GetMatchListByOrderSn(stub, orderSn)
		if err != nil {
			return shim.Error(err.Error())
		}
		data, _ := json.Marshal(result)
		return shim.Success(data)
	case "get_all_match_list": //列出订单的成交记录
		start, end := 0, 0
		if len(args) >= 3 {
			return shim.Error("tokenTrade, must input < 3 arg")
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
			return shim.Error("tokenTrade, must input0 arg")
		}
		data, err := p.GetMatchCount(stub)
		if err != nil {
			return shim.Error(err.Error())
		}
		return shim.Success([]byte(strconv.Itoa(data)))

	case "set_trade_mgr_address_list": //设置管理地址
		if len(args) <= 0 {
			return shim.Error("tokenTrade, must input > 0 arg")
		}
		ads := common.Addresses{}
		for _, arg := range args {
			addr, err := common.StringToAddress(arg)
			if err == nil {
				ads = append(ads, addr)
			}
		}
		err := setTradeContractMgrAddress(stub, ads)
		if err != nil {
			return shim.Error(err.Error())
		}
		data, _ := json.Marshal(ads)
		return shim.Success(data)

	case "get_trade_mgr_address_list": //获取管理地址
		if len(args) != 0 {
			return shim.Error("tokenTrade, must input 0 arg")
		}
		result, err := getTradeContractMgrAddress(stub)
		if err != nil {
			return shim.Error(err.Error())
		}
		data, _ := json.Marshal(result)
		return shim.Success(data)

	default:
		return shim.Error("tokenTrade, no case..")
	}
	return shim.Error("tokenTrade, no case..")
}
