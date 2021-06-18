package auctioncc

import (
	"github.com/palletone/go-palletone/contracts/shim"
	"github.com/shopspring/decimal"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/dag/errors"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/ptnjson"
	"time"
	"fmt"
)

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
	}
	rate := getAuctionFeeRate(stub, 1)
	amount := decimal.NewFromFloat(float64(auction.WantAmount)).Mul(rate).IntPart()
	destructionAmount := uint64(amount)

	now, err := stub.GetTxTimestamp(10)
	if err != nil {
		return errors.New("takerAuction, GetTxTimestamp err:" + err.Error())
	}
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
		"TargetAmount:%d\n"+
		"StepAmount:%d\n"+
		"StartTime:%s\n"+
		"EndTime:%s\n"+
		"RewardAddress:%s\n"+
		"AuctionSn:%s\n"+
		"CreateTime:%s\n"+
		"Status:%d\n",
		order.AuctionType, order.Address.String(), order.SaleAsset.String(), order.SaleAmount, order.WantAsset.String(), order.WantAmount,
		order.TargetAmount, order.StepAmount, order.StartTime.String(), order.EndTime.String(),
		order.RewardAddress.String(), order.AuctionSn, order.CreateTime.String(), order.Status)

	return p.AddAuctionOrder(stub, order)
}

func (p *AuctionMgr) TakerAuction(stub shim.ChaincodeStubInterface, orderSn string) error {
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
	log.Debugf("TakerAuction, startTime[%d], nowTime[%d], endTime[%d]", auction.StartTime.Unix(), now.Seconds, auction.EndTime.Unix())

	if auction.StartTime.Unix() > 0 && now.Seconds < auction.StartTime.Unix() {
		return errors.New("takerAuction, now.Seconds < auction.StartTime")
	} else if auction.EndTime.Unix() > 0 && now.Seconds > auction.EndTime.Unix() {
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
	needAmount := auction.WantAmount
	if auction.StepAmount != 0 { //设置有step level的情况下
		lastAmount, err := getAuctionLastAmountRecord(stub, auction.AuctionSn)
		if err != nil {
			return nil
		}
		needAmount = lastAmount.TakerAmount + auction.StepAmount
	}
	//检查提交金额数量是否满足
	if takerPayAmount < needAmount {
		return fmt.Errorf("takerAuction, auction[%s] TakerReqId[%s], takerPayAmount[%d] < needAmount [%d]", auction.AuctionSn, stub.GetTxID(), takerPayAmount, needAmount)
	}

	//计算奖励和销毁的费用
	var rewardAmount uint64 = 0
	if !auction.RewardAddress.IsZero() {
		rate := getAuctionFeeRate(stub, 0)
		amount := decimal.NewFromFloat(float64(auction.WantAmount)).Mul(rate).IntPart()
		rewardAmount = uint64(amount)
	}
	rate := getAuctionFeeRate(stub, 1)
	amount := decimal.NewFromFloat(float64(auction.WantAmount)).Mul(rate).IntPart()
	destructionAmount := uint64(amount)
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
	if auction.StepAmount != 0 { //设置有step level的情况下
		lastAmount := AuctionLastAmount{
			AuctionOrderSn: submitRecord.AuctionOrderSn,
			TakerReqId:     submitRecord.TakerReqId,
			TakerAddress:   submitRecord.TakerAddress,
			TakerAsset:     submitRecord.TakerAsset,
			TakerAmount:    submitRecord.TakerAssetAmount,
		}
		err = saveAuctionLastAmountRecord(stub, &lastAmount)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *AuctionMgr) UpdateTakerAuction(stub shim.ChaincodeStubInterface, orderSn string) error {
	takerAddress, _ := stub.GetInvokeAddress()
	auction, err := getAuctionRecordBySn(stub, orderSn)
	if err != nil {
		return errors.New("UpdateTakerAuction, invalid/sold out/canceled takerAuction SN:" + orderSn)
	}
	//检查挂单是否有效
	if auction.Status != 1 {
		return errors.New("UpdateTakerAuction, takerAuction order is invalid")
	}
	//检查时间是否有效
	now, err := stub.GetTxTimestamp(10)
	if err != nil {
		return errors.New("UpdateTakerAuction, GetTxTimestamp err:" + err.Error())
	}
	log.Debugf("UpdateTakerAuction, startTime[%d], nowTime[%d], endTime[%d]", auction.StartTime.Unix(), now.Seconds, auction.EndTime.Unix())

	if now.Seconds < auction.StartTime.Unix() {
		return errors.New("UpdateTakerAuction, now.Seconds < auction.StartTime")
	} else if now.Seconds > auction.EndTime.Unix() {
		return errors.New("UpdateTakerAuction, now.Seconds > auction.EndTime")
	}
	takerPayAsset, takerPayAddAmount, err := getPayToContract(stub)
	if err != nil {
		return err
	}
	//检查assert是否相同
	if !takerPayAsset.Equal(auction.WantAsset) {
		return errors.New("UpdateTakerAuction, current asset not match auction order want asset")
	}

	//根据地址获取之前的订单信息
	record, err := getMatchRecordByAddress(stub, orderSn, takerAddress)
	if err != nil {
		return err
	}
	allPayAmount := record.TakerAssetAmount + takerPayAddAmount
	needAmount := auction.WantAmount
	if auction.StepAmount != 0 { //设置有step level的情况下
		lastAmount, err := getAuctionLastAmountRecord(stub, auction.AuctionSn)
		if err != nil {
			return nil
		}
		needAmount = lastAmount.TakerAmount + auction.StepAmount
	}
	//检查提交金额数量是否满足
	if allPayAmount < needAmount {
		return fmt.Errorf("UpdateTakerAuction, auction[%s] TakerReqId[%s], allPayAmount[%d] < needAmount [%d]", auction.AuctionSn, stub.GetTxID(), allPayAmount, needAmount)
	}
	//重新计算奖励和销毁的费用
	var rewardAmount uint64 = 0
	if !auction.RewardAddress.IsZero() {
		rate := getAuctionFeeRate(stub, 0)
		rewardAmount = uint64(decimal.NewFromFloat(float64(allPayAmount)).Mul(rate).IntPart())
	}
	rate := getAuctionFeeRate(stub, 1)
	destructionAmount := uint64(decimal.NewFromFloat(float64(allPayAmount)).Mul(rate).IntPart())
	desAddr, _ := common.StringToAddress(DestructionAddress)
	feeUse := AuctionFeeUse{
		Asset:              auction.WantAsset,
		RewardAddress:      auction.RewardAddress,
		RewardAmount:       rewardAmount, //auction.WantAsset.DisplayAmount(10), //auction.WantAsset.DisplayAmount(auction.SaleAmount)
		DestructionAddress: desAddr,
		DestructionAmount:  destructionAmount, //auction.WantAsset.DisplayAmount(5),
	}

	//更新状态数据
	record.TakerAssetAmount = allPayAmount
	record.FeeUse = feeUse
	record.recordTime = now.Seconds
	err = saveMatchRecord(stub, record)
	if err != nil {
		return err
	}
	if auction.StepAmount != 0 { //设置有step level的情况下
		lastAmount := AuctionLastAmount{
			AuctionOrderSn: record.AuctionOrderSn,
			TakerReqId:     record.TakerReqId,
			TakerAddress:   record.TakerAddress,
			TakerAsset:     record.TakerAsset,
			TakerAmount:    record.TakerAssetAmount,
		}
		err = saveAuctionLastAmountRecord(stub, &lastAmount)
		if err != nil {
			return err
		}
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
	if !invokeAddress.Equal(auction.Address) || !isFoundationInvoke(stub) || !isAuctionContractMgrAddress(stub) {
		return errors.New("StopAuction addr are not the owner or not foundation")
	}
	//按金额、时间获取成交记录
	matchRecords, err := getMatchRecordByOrderSn(stub, orderSn)
	if err != nil {
		return errors.New("StopAuction getMatchRecordByOrderSn err:" + err.Error())
	}
	maxRecord := getMaxAmountRecord(matchRecords)
	if maxRecord == nil {
		return errors.New("StopAuction, not find max amount record")
	}
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

func (p *AuctionMgr) Cancel(stub shim.ChaincodeStubInterface, orderSn string) error {
	auction, err := getAuctionRecordBySn(stub, orderSn)
	if err != nil {
		return errors.New("invalid/sold out/canceled auction SN:" + orderSn)
	}
	//检查挂单状态是否有效
	if auction.Status != 1 {
		return errors.New(" auction.Status is invalid/ auction SN:" + orderSn)
	}
	//检查拍卖情况下是否有提交订单记录：如果已经有参与竞拍，则不能撤销订单
	if auction.AuctionType != 1 && isInMatchRecord(stub, orderSn) {
		return errors.New("Cancel, there are already bidders ")
	}

	//检查是否是Maker或者基金会或者管理地址
	addr, err := stub.GetInvokeAddress()
	if addr != auction.Address || !isFoundationInvoke(stub) || !isAuctionContractMgrAddress(stub) {
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
