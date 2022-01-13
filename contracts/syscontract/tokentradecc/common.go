package tokentradecc

import (
	"github.com/shopspring/decimal"
	"time"
	"github.com/palletone/go-palletone/common/log"
	"encoding/json"
	"github.com/palletone/go-palletone/contracts/shim"
	"github.com/palletone/go-palletone/dag/errors"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/dag/modules"
)

const TradeContractMgrAddressPrefix = "TradeContractMgrAddress-"
const TradeContractFeeRate = "TradeContractFeeRate-"
const FirstFeeRateLevel = "FirstFeeRateLevel-"
const TimeFormt = "2006-01-02 15:04:05"

const DefaultRewardFeeRate = 0.01      //默认奖励费率
const DefaultDestructionFeeRate = 0.01 //默认销毁费率

const DefaultFirstRewardFeeRateLevel = 2.0      //第一次奖励级别，第一次奖励=费率*级别*交易额
const DefaultFirstDestructionFeeRateLevel = 2.0 //第一次销毁级别，第一次销毁=费率*级别*交易额

var DestructionAddress = "P1111111111111111111114oLvT2" //""P1111111111111111111114oLvT2"  "PCGTta3M4t3yXu8uRgkKvaWd2d9Vgsc4zGX"//销毁地址

func getPayToContract(stub shim.ChaincodeStubInterface) (*modules.Asset, uint64, error) {
	payAssets, _ := stub.GetInvokeTokens()
	_, contractAddr := stub.GetContractID()
	for _, invokeAA := range payAssets {
		if invokeAA.Address == contractAddr {
			return invokeAA.Asset, invokeAA.Amount, nil
		}
	}
	return nil, 0, errors.New("no pay to contract token")
}

//  判断是否基金会发起的
func isFoundationInvoke(stub shim.ChaincodeStubInterface) bool {
	//  判断是否基金会发起的
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		log.Error("get invoke address err: ", "error", err)
		return false
	}
	//  获取
	gp, err := stub.GetSystemConfig()
	if err != nil {
		//log.Error("strconv.ParseUint err:", "error", err)
		return false
	}
	foundationAddress := gp.ChainParameters.FoundationAddress
	// 判断当前请求的是否为基金会
	if invokeAddr.String() != foundationAddress {
		log.Debug("invoke address is not foundation address")
		return false
	}
	return true
}

func transferTokensProcess(stub shim.ChaincodeStubInterface, trade *TradeOrder, record *MatchRecord) error {
	if trade == nil || record == nil {
		return errors.New("TransferTokensProcess, ")
	}
	//交换给taker的
	err := stub.PayOutToken(record.TakerAddress.String(), &modules.AmountAsset{
		Amount: record.MakerTradeAmount,
		Asset:  record.MakerAsset,
	}, 0)
	if err != nil {
		return err
	}

	//返还taker多余的
	takerChangeAmount := record.TakerAssetAmount - record.TakerTradeAmount
	if takerChangeAmount > 0 { //剩余的打回
		err = stub.PayOutToken(record.TakerAddress.String(), &modules.AmountAsset{
			Amount: takerChangeAmount,
			Asset:  record.TakerAsset,
		}, 0)
		if err != nil {
			return err
		}
	}

	remainAmount := record.TakerAssetAmount - record.FeeUse.DestructionAmount - record.FeeUse.RewardAmount
	//卖出的
	if remainAmount > 0 {
		err = stub.PayOutToken(record.MakerAddress.String(), &modules.AmountAsset{
			Amount: record.TakerTradeAmount,
			Asset:  record.TakerAsset,
		}, 0)
		if err != nil {
			return err
		}
	}
	//奖励
	if record.FeeUse.RewardAmount > 0 {
		err = stub.PayOutToken(record.FeeUse.RewardAddress.String(), &modules.AmountAsset{
			Amount: record.FeeUse.RewardAmount,
			Asset:  record.FeeUse.Asset,
		}, 0)
		if err != nil {
			return err
		}
	}
	//销毁
	if record.FeeUse.DestructionAmount > 0 {
		err = stub.PayOutToken(DestructionAddress, &modules.AmountAsset{
			Amount: record.FeeUse.DestructionAmount,
			Asset:  record.FeeUse.Asset,
		}, 0)
		if err != nil {
			return err
		}
	}
	return nil
}

func setTradeContractMgrAddress(stub shim.ChaincodeStubInterface, mgrAddress common.Addresses) error {
	if mgrAddress == nil {
		errors.New("setAuctionContractMgrAddress, param address is nil")
	}
	if !isFoundationInvoke(stub) {
		return errors.New("setAuctionContractMgrAddress, the invoke address is err")
	}

	data, _ := json.Marshal(mgrAddress)
	key := TradeContractMgrAddressPrefix
	log.Debugf("setAuctionContractMgrAddress, addrs[%v]", mgrAddress)
	return stub.PutState(key, data)
}

func getTradeContractMgrAddress(stub shim.ChaincodeStubInterface) (mgrAddress common.Addresses, err error) {
	addrs := common.Addresses{}
	key := TradeContractMgrAddressPrefix
	value, err := stub.GetState(key)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(value, &addrs)
	return addrs, nil
}

func isTradeContractMgrAddress(stub shim.ChaincodeStubInterface) bool {
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		log.Debugf("isAuctionContractMgrAddress, GetInvokeAddress err:%s", err.Error())
		return false
	}
	address, err := getTradeContractMgrAddress(stub)
	if err != nil {
		return false
	}
	for _, ad := range address {
		if ad.Equal(invokeAddr) {
			return true
		}
	}
	return false
}

func setAuctionFeeRate(stub shim.ChaincodeStubInterface, rateType uint8, rate decimal.Decimal) error {
	if !isFoundationInvoke(stub) && !isTradeContractMgrAddress(stub) {
		return errors.New("setAuctionFeeRate, the invoke address is err")
	}
	log.Debugf("setAuctionFeeRate, rateType[%d], rate[%s]", rateType, rate.String())

	// 0 <= rate < 1
	if rate.GreaterThanOrEqual(decimal.NewFromFloat(0)) &&
		rate.LessThan(decimal.NewFromFloat(1)) {
		key := TradeContractFeeRate
		if rateType == 0 {
			key = key + "rewardRate"
		} else {
			key = key + "destructionRate"
		}
		stub.DelState(key)
		brate, _ := rate.GobEncode()
		return stub.PutState(key, brate)
	} else {
		return errors.New("setAuctionFeeRate， value err :feeRate =" + rate.String())
	}
}

func getAuctionFeeRate(stub shim.ChaincodeStubInterface, rateType uint8) (decimal.Decimal) {
	defRate := decimal.Decimal{}
	key := TradeContractFeeRate
	if rateType == 0 {
		key = key + "rewardRate"
		defRate = decimal.NewFromFloat(DefaultRewardFeeRate) //todo
	} else {
		key = key + "destructionRate"
		defRate = decimal.NewFromFloat(DefaultDestructionFeeRate) //todo
	}
	value, err := stub.GetState(key)
	if err != nil { //use default fee rate
		log.Debugf("getAuctionFeeRate, rateType[%d],  use default rate[%s]", rateType, defRate.String())
		return defRate
	}
	if value != nil {
		data := decimal.Decimal{}
		data.GobDecode(value)
		log.Debugf("getAuctionFeeRate, rateType[%d], rate[%s]", rateType, data.String())
		return data
	}
	return defRate
}

func setFirstFeeRateLevel(stub shim.ChaincodeStubInterface, rateType uint8, rate decimal.Decimal) error {
	if !isFoundationInvoke(stub) && !isTradeContractMgrAddress(stub) {
		return errors.New("setFirstFeeRateLevel, the invoke address is err")
	}
	log.Debugf("setFirstFeeRateLevel, rateType[%d], rate[%s]", rateType, rate.String())

	// 0 < level
	if rate.GreaterThan(decimal.NewFromFloat(0)) {
		key := FirstFeeRateLevel
		if rateType == 0 {
			key = key + "rewardRate"
		} else {
			key = key + "destructionRate"
		}
		stub.DelState(key)
		brate, _ := rate.GobEncode()
		return stub.PutState(key, brate)
	} else {
		return errors.New("setFirstFeeRateLevel， value err :feeRate =" + rate.String())
	}
}

func getFirstFeeRateLevel(stub shim.ChaincodeStubInterface, rateType uint8) (decimal.Decimal) {
	defRate := decimal.Decimal{}
	key := FirstFeeRateLevel
	if rateType == 0 {
		key = key + "rewardRate"
		defRate = decimal.NewFromFloat(DefaultFirstRewardFeeRateLevel) //todo
	} else {
		key = key + "destructionRate"
		defRate = decimal.NewFromFloat(DefaultFirstDestructionFeeRateLevel) //todo
	}
	value, err := stub.GetState(key)
	if err != nil { //use default fee rate
		log.Debugf("getFirstFeeRateLevel, rateType[%d],  use default rate[%s]", rateType, defRate.String())
		return defRate
	}
	if value != nil {
		data := decimal.Decimal{}
		data.GobDecode(value)
		log.Debugf("getFirstFeeRateLevel, rateType[%d], rate[%s]", rateType, data.String())
		return data
	}
	return defRate
}

func getTimeFromString(inStr string) (time.Time, error) {
	//l1 := now.Format("2006-01-02 15")
	tm, err := time.Parse(TimeFormt, inStr)
	//tm.Unix()
	if err != nil {
		tm, err = time.Parse(TimeFormt+" +0000 UTC", inStr)
		if err != nil {
			tm, err = time.Parse(TimeFormt+" +0800 CST", inStr)
			return tm, err
		}
	}
	return tm, err

	//	ti := time.Unix(t.Seconds, 0)
	//	return ti.UTC().Format(modules.Layout2)
}

//
func string2decimal(data string) decimal.Decimal {
	value, err := decimal.NewFromString(data)
	if err != nil {
		log.Errorf("string2decimal, err:%s", err)
	}
	return value
}

func checkScaleParam() {

}