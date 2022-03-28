package tokentradecc

import (
	"time"
	"fmt"
	"github.com/palletone/go-palletone/contracts/shim"
	"github.com/shopspring/decimal"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/dag/errors"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/ptnjson"
)

func (p *TokenTrade) MakerTrade(stub shim.ChaincodeStubInterface, tradeType byte, wantAsset *modules.Asset, wantAmount decimal.Decimal, scale string, rewardAddress common.Address) error {
	addr, err := stub.GetInvokeAddress()
	if err != nil {
		return errors.New("makerTrade, Invalid address string:" + err.Error())
	}
	inToken, inAmount, err := getPayToContract(stub)
	if err != nil {
		return err
	}

	t, _ := stub.GetTxTimestamp(10)
	txid := stub.GetTxID()
	order := &TradeOrder{}
	order.TradeType = tradeType //fix
	order.Address = addr
	order.InAsset = inToken
	order.InAmount = inAmount
	order.InUnCount = inAmount
	order.Scale = scale
	order.WantAsset = wantAsset
	order.WantAmount = wantAsset.Uint64Amount(wantAmount)
	order.UnAmount = order.WantAmount
	order.RewardAddress = rewardAddress
	order.TradeSn = txid
	order.Status = 1
	order.CreateTime = time.Unix(t.Seconds, 0).Format(TimeFormt) //time.Unix(t.Seconds, 0).String()

	log.Debugf("MakerTrade:\n"+
		"TradeType:%d\n"+
		"Address:%s\n"+
		"InAsset:%s\n"+
		"InAmount:%d\n"+
		"InUnCount:%d\n"+
		"WantAsset:%s\n"+
		"WantAmount:%d\n"+
		"Scale:%s\n"+
		"UnAmount:%d\n"+
		"RewardAddress:%s\n"+
		"TradeSn:%s\n"+
		"Status:%d\n"+
		"CreateTime:%s\n",
		order.TradeType, order.Address.String(), order.InAsset.String(), order.InAmount, order.InUnCount, order.WantAsset.String(),
		order.WantAmount, order.Scale, order.UnAmount, order.RewardAddress.String(), order.TradeSn, order.Status, order.CreateTime)

	//检查scale的有效性
	sc, err := decimal.NewFromString(scale)
	if err != nil {
		log.Errorf("MakerTrade, scale to decimal err:%s", err)
		return err
	}
	mscale := sc.Mul(amt2AssetAmt(inToken, 1)).Div(amt2AssetAmt(wantAsset, 1))
	if true != mscale.Equal(decimal.New(int64(order.InAmount), 0).Div(decimal.New(int64(order.WantAmount), 0))) {
		err = fmt.Errorf("MakerTrade, scale[%s] not equal[%d][%d]", scale, order.InAmount, order.WantAmount)
		log.Error(err.Error())
		return err
	}

	return p.AddTradeOrder(stub, order)
}

func (p *TokenTrade) TakerTrade(stub shim.ChaincodeStubInterface, tradeType byte, orderSn string) error {
	takerAddress, _ := stub.GetInvokeAddress()
	trade, err := getTradeRecordBySn(stub, orderSn)
	if err != nil {
		return fmt.Errorf("TakerTrade[%s],invalid/sold out/canceled ", orderSn)
	}
	if trade.Status != 1 {
		return fmt.Errorf("TakerTrade[%s],trade status[%d] is not active  ", trade.Status)
	}
	takerPayAsset, takerPayAmount, err := getPayToContract(stub)
	if err != nil {
		return err
	}
	//检查assert是否是否相同
	if !takerPayAsset.Equal(trade.WantAsset) {
		return fmt.Errorf("TakerTrade[%s], current asset not match takerFix order want asset", orderSn)
	}
	var takerTradeCount uint64
	takerTradeCount = 0
	//检查金额是否满足
	if takerPayAmount < trade.UnAmount { //未完全成交
		takerTradeCount = takerPayAmount
	} else { //完全成交
		takerTradeCount = trade.UnAmount
	}
	sc, err := decimal.NewFromString(trade.Scale)
	if err != nil {
		return fmt.Errorf("TakerTrade[%s], decimal err: %s", orderSn, err.Error())
	}

	mscale := sc.Mul(amt2AssetAmt( trade.InAsset, 1)).Div(amt2AssetAmt(takerPayAsset, 1))
	makerTradeCount := uint64(mscale.Mul(decimal.New(int64(takerTradeCount), 0)).IntPart()) //todo
	//检查maker 返还金额是否足够
	if makerTradeCount > trade.InUnCount {
		return fmt.Errorf("TakerTrade[%s], makerTradeCount[%d] > trade.InUnCount[%d]", orderSn, makerTradeCount, trade.InUnCount)
	}
	//计算奖励和销毁的费用
	rewardAmount, destructionAmount := calculateFeeRate(stub, trade)
	now, err := stub.GetTxTimestamp(10)
	if err != nil {
		return fmt.Errorf("TakerTrade[%s], GetTxTimestamp err:%s", orderSn, err.Error())
	}
	desAddr, _ := common.StringToAddress(DestructionAddress)
	feeUse := TradeFeeUse{
		Asset:              trade.WantAsset,
		RewardAddress:      trade.RewardAddress,
		RewardAmount:       rewardAmount, //trade.WantAsset.DisplayAmount(10), //trade.WantAsset.DisplayAmount(trade.SaleAmount)
		DestructionAddress: desAddr,
		DestructionAmount:  destructionAmount, //trade.WantAsset.DisplayAmount(5),
	}
	//更新状态数据
	matchRecord := &MatchRecord{
		TradeType:        trade.TradeType,
		TradeOrderSn:     trade.TradeSn,
		TakerReqId:       stub.GetTxID(),
		MakerAddress:     trade.Address,
		MakerAsset:       trade.InAsset,
		MakerAssetAmount: trade.InAmount,
		MakerTradeAmount: makerTradeCount,
		Scale:            trade.Scale,
		TakerAddress:     takerAddress,
		TakerAsset:       takerPayAsset,
		TakerAssetAmount: takerPayAmount,
		TakerTradeAmount: takerTradeCount,
		FeeUse:           feeUse,
		RecordTime:       time.Unix(now.Seconds, 0).Format(TimeFormt),
	}
	log.Debugf("TakerTrade:\n"+
		"TradeType:%d\n"+
		"TradeOrderSn:%s\n"+
		"TakerReqId:%s\n"+
		"MakerAddress:%s\n"+
		"MakerAsset:%s\n"+
		"MakerAssetAmount:%d\n"+
		"MakerTradeAmount:%d\n"+
		"Scale:%s\n"+
		"TakerAddress:%s\n"+
		"TakerAsset:%s\n"+
		"TakerAssetAmount:%d\n"+
		"TakerTradeAmount:%d\n"+
		"FeeUse.asset:%s\n"+
		"FeeUse.RewardAddress:%s\n"+
		"FeeUse.RewardAmount:%d\n"+
		"FeeUse.DestructionAddress:%s\n"+
		"FeeUse.DestructionAmount:%d\n"+
		"RecordTime:%s\n",
		matchRecord.TradeType, matchRecord.TradeOrderSn, matchRecord.TakerReqId, matchRecord.MakerAddress.String(), matchRecord.MakerAsset.String(), matchRecord.MakerAssetAmount, matchRecord.MakerTradeAmount,
		matchRecord.Scale, matchRecord.TakerAddress.String(), matchRecord.TakerAsset.String(), matchRecord.TakerAssetAmount, matchRecord.TakerTradeAmount,
		matchRecord.FeeUse.Asset.String(), matchRecord.FeeUse.RewardAddress.String(), matchRecord.FeeUse.RewardAmount, matchRecord.FeeUse.DestructionAddress.String(), matchRecord.FeeUse.DestructionAmount,
		matchRecord.RecordTime)

	err = saveMatchRecord(stub, matchRecord)
	if err != nil {
		return err
	}

	//计算本次交易后Maker还有多少数量没有交易
	remanCount := trade.UnAmount - takerTradeCount
	if remanCount <= 0 {
		trade.Status = 2
	} else {
		trade.Status = 1
	}
	trade.InUnCount = trade.InUnCount - makerTradeCount
	trade.UnAmount = remanCount

	log.Debugf("TakerTrade, update trade[%v]", trade)
	err = UpdateTradeOrder(stub, trade)
	if err != nil {
		return err
	}
	return transferTokensProcess(stub, trade, matchRecord)
}

func (p *TokenTrade) Cancel(stub shim.ChaincodeStubInterface, orderSn string) error {
	trade, err := getTradeRecordBySn(stub, orderSn)
	if err != nil {
		return errors.New("Cancel, invalid/sold out/canceled trade SN:" + orderSn)
	}
	//检查挂单状态是否有效
	if trade.Status != 1 {
		return errors.New("Cancel,  trade.Status is invalid/ trade SN:" + orderSn)
	}
	//检查是否是Maker或者基金会或者管理地址
	addr, err := stub.GetInvokeAddress()
	if !addr.Equal(trade.Address) && !isFoundationInvoke(stub) && !isTradeContractMgrAddress(stub) {
		return fmt.Errorf("Cancel, you are not the owner or not foundation, invoke addr[%s]-trade.Address[%s]", addr.String(), trade.Address.String())
	}
	err = cancelTradeOrder(stub, trade)
	if err != nil {
		return err
	}

	log.Debugf("Cancel, cancel count :%d", trade.InUnCount)
	if trade.InUnCount <= 0 { //没有返还金额
		log.Debugf("Cancel, returnCount is 0")
		return nil
	}
	_, contractAddr := stub.GetContractID()
	//检查token数量是否满足
	getTokens, err := stub.GetTokenBalance(contractAddr, trade.InAsset)
	if err != nil {
		return err
	}
	for _, tk := range getTokens {
		if tk.Asset.Equal(trade.InAsset) && tk.Amount >= trade.InUnCount {
			//未成交的金额退回
			err = stub.PayOutToken(trade.Address.String(), &modules.AmountAsset{
				Amount: trade.InUnCount,
				Asset:  trade.InAsset,
			}, 0)

			if err != nil {
				return err
			}
		}
	}

	return nil
}
func (p *TokenTrade) Payout(stub shim.ChaincodeStubInterface, addr common.Address, amount decimal.Decimal, asset *modules.Asset) error {
	uint64Amt := ptnjson.JsonAmt2AssetAmt(asset, amount)
	return stub.PayOutToken(addr.String(), &modules.AmountAsset{
		Amount: uint64Amt,
		Asset:  asset,
	}, 0)
}

func (p *TokenTrade) AddTradeOrder(stub shim.ChaincodeStubInterface, order *TradeOrder) error {
	addr := order.Address
	if !KycUser(addr) {
		return errors.New("AddTradeOrder,Please verify your ID")
	}
	err := SaveTradeOrder(stub, order)
	if err != nil {
		return errors.New("AddTradeOrder, saveRecord error:" + err.Error())
	}
	return nil
}

//当前有效挂单
func (p *TokenTrade) GetActiveOrderCount(stub shim.ChaincodeStubInterface) (int, error) {
	return getActiveOrderCount(stub)
}

func (p *TokenTrade) GetActiveOrderList(stub shim.ChaincodeStubInterface, start, end int) ([]*TradeOrderJson, error) {
	return getAllTradeOrder(stub, start, end)
}
func (p *TokenTrade) GetActiveOrdersByMaker(stub shim.ChaincodeStubInterface, addr common.Address) ([]*TradeOrderJson, error) {
	return getTradeOrderByAddress(stub, addr)
}
func (p *TokenTrade) GetActiveOrdersByID(stub shim.ChaincodeStubInterface, orderId string) (*TradeOrderJson, error) {
	order, err := getTradeRecordBySn(stub, orderId)
	if err != nil {
		return nil, err
	}
	log.Debugf("GetActiveOrdersByID, order:%v", order)
	jsSheet := convertSheet(*order)
	log.Debugf("GetActiveOrdersByID, jsSheet:%v", jsSheet)
	return jsSheet, nil
}

//历史挂单
func (p *TokenTrade) GetHistoryOrderCount(stub shim.ChaincodeStubInterface) (int, error) {
	return getHistoryOrderCount(stub)
}
func (p *TokenTrade) GetHistoryOrderList(stub shim.ChaincodeStubInterface, start, end int) ([]*TradeOrderJson, error) {
	return getAllHistoryOrder(stub, start, end)
}
func (p *TokenTrade) GetHistoryOrderBySn(stub shim.ChaincodeStubInterface, tradeSn string) (*TradeOrderJson, error) {
	return getHistoryOrderBySn(stub, tradeSn)
}
func (p *TokenTrade) GetHistoryOrderByMakerAddr(stub shim.ChaincodeStubInterface, addr common.Address) ([]*TradeOrderJson, error) {
	return getHistoryOrderByMakerAddr(stub, addr)
}

//交易订单
func (p *TokenTrade) GetMatchListByOrderSn(stub shim.ChaincodeStubInterface, orderSn string) ([]*MatchRecordJson, error) {
	return getMatchRecordJsonByOrderSn(stub, orderSn)
}
func (p *TokenTrade) GetAllMatchList(stub shim.ChaincodeStubInterface, start, end int) ([]*MatchRecordJson, error) {
	return getAllMatchRecordJson(stub, start, end)
}
func (p *TokenTrade) GetMatchCount(stub shim.ChaincodeStubInterface) (int, error) {
	return getMatchRecordCount(stub)
}

func calculateFeeRate(stub shim.ChaincodeStubInterface, trade *TradeOrder) (rewardAmount, destructionAmount uint64) {
	//检查NFT是否是第一次参与交易
	rewardAmount = 0
	destructionAmount = 0

	//firstTrade := true
	//if isInMatchRecordByAssertId(stub, auction.InAsset) {
	//	firstTrade = false
	//}
	//log.Debugf("calculateFeeRate, assertId[%s] firstTrade[%v]", auction.InAsset.String(), firstTrade)
	//
	//if !auction.RewardAddress.IsZero() {
	//	rewardRate := getAuctionFeeRate(stub, 0)
	//	if firstTrade {
	//		firstLevel := getFirstFeeRateLevel(stub, 0)
	//		rewardRate = rewardRate.Mul(firstLevel)
	//	}
	//	amount := decimal.NewFromFloat(float64(auction.WantAmount)).Mul(rewardRate).IntPart()
	//	rewardAmount = uint64(amount)
	//}
	//destructionRate := getAuctionFeeRate(stub, 1)
	//if firstTrade {
	//	firstLevel := getFirstFeeRateLevel(stub, 1)
	//	destructionRate = destructionRate.Mul(firstLevel)
	//}
	//amount := decimal.NewFromFloat(float64(auction.WantAmount)).Mul(destructionRate).IntPart()
	//destructionAmount = uint64(amount)
	//log.Debugf("calculateFeeRate, sn[%s], rewardAmount[%d], destructionAmount[%d]", auction.TradeSn, rewardAmount, destructionAmount)

	return rewardAmount, destructionAmount
}
