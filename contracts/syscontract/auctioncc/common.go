package auctioncc

import (
	"github.com/palletone/go-palletone/contracts/shim"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/errors"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/shopspring/decimal"
	"time"
)

const AuctionContractMgrAddressPrefix = "AuctionContractMgrAddress-"
const AuctionContractFeeRate = "AuctionContractFeeRate-"
const TimeFormt = "2006-01-02 15:04:05 UTC"

const DefaultRewardFeeRate = 0.025      //默认奖励费率
const DefaultDestructionFeeRate = 0.025 //默认销毁费率

const FirstRewardFeeRateLevel = 2.0      //第一次奖励级别，第一次奖励=费率*级别*交易额
const FirstDestructionFeeRateLevel = 2.0 //第一次销毁级别，第一次销毁=费率*级别*交易额

//todo tmp
var DestructionAddress = "PCGTta3M4t3yXu8uRgkKvaWd2d9Vgsc4zGX" //""PCLOST00000000000000000000000000000"  //销毁地址
var AuctionTransactionGas = uint64(1)                          //ptn  临时的

func getPayToContract(stub shim.ChaincodeStubInterface) (*modules.Asset, uint64, error) {
	payAssets, _ := stub.GetInvokeTokens()
	for _, invokeAA := range payAssets {
		if invokeAA.Address == myContractAddr {
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

func transferTokensProcess(stub shim.ChaincodeStubInterface, auction *AuctionOrder, record *MatchRecord) error {
	if auction == nil || record == nil {
		return errors.New("TransferTokensProcess, ")
	}
	//交换给taker的
	err := stub.PayOutToken(record.TakerAddress.String(), &modules.AmountAsset{
		Amount: record.MakerAssetAmount,
		Asset:  record.MakerAsset,
	}, 0)
	if err != nil {
		return err
	}

	//返还taker多余的
	if record.AuctionType == 1 {
		takerChangeAmount := record.TakerAssetAmount - auction.WantAmount
		if takerChangeAmount > 0 { //剩余的打回
			err = stub.PayOutToken(record.TakerAddress.String(), &modules.AmountAsset{
				Amount: takerChangeAmount,
				Asset:  record.TakerAsset,
			}, 0)
			if err != nil {
				return err
			}
		}
	}

	remainAmount := record.TakerAssetAmount - record.FeeUse.DestructionAmount - record.FeeUse.RewardAmount
	//卖出的
	if remainAmount > 0 {
		err = stub.PayOutToken(record.MakerAddress.String(), &modules.AmountAsset{
			Amount: remainAmount,
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

func setAuctionContractMgrAddress(stub shim.ChaincodeStubInterface, mgrAddress common.Addresses) error {
	if mgrAddress == nil {
		errors.New("setAuctionContractMgrAddress, param address is nil")
	}
	if !isFoundationInvoke(stub) {
		return errors.New("setAuctionContractMgrAddress, the invoke address is err")
	}

	data, _ := rlp.EncodeToBytes(mgrAddress)
	key := AuctionContractMgrAddressPrefix

	return stub.PutState(key, data)
}

func getAuctionContractMgrAddress(stub shim.ChaincodeStubInterface) (mgrAddress common.Addresses, err error) {
	if !isFoundationInvoke(stub) {
		return nil, errors.New("getAuctionContractMgrAddress, the invoke address is err")
	}
	key := AuctionContractMgrAddressPrefix
	rows, err := stub.GetStateByPrefix(key)
	if err != nil {
		return nil, err
	}

	result := common.Addresses{}
	for _, row := range rows {
		ad := common.Address{}
		rlp.DecodeBytes(row.Value, ad)
		result = append(result, ad)
	}

	return result, nil
}

func isAuctionContractMgrAddress(stub shim.ChaincodeStubInterface) bool {
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		log.Debugf("isAuctionContractMgrAddress, GetInvokeAddress err:%s", err.Error())
		return false
	}
	address, err := getAuctionContractMgrAddress(stub)
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
	if !isFoundationInvoke(stub) && !isAuctionContractMgrAddress(stub) {
		return errors.New("setAuctionFeeRate, the invoke address is err")
	}
	log.Debugf("setAuctionFeeRate, rateType[%d], rate[%s]", rateType, rate.String())

	// 0 <= rate < 1
	if rate.GreaterThanOrEqual(decimal.NewFromFloat(0)) && rate.GreaterThanOrEqual(decimal.NewFromFloat(0)) &&
		rate.LessThan(decimal.NewFromFloat(1)) && rate.LessThan(decimal.NewFromFloat(1)) {
		key := AuctionContractFeeRate
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
	key := AuctionContractFeeRate
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

func getTimeFromString(inStr string) (time.Time, error) {
	tm, err := time.Parse("2006-01-02 15:04:05", inStr)
	//tm.Unix()
	return tm, err

	//	ti := time.Unix(t.Seconds, 0)
	//	return ti.UTC().Format(modules.Layout2)
}

//func getTimeFromSeconds(seconds  int64) (time.Time, error) {
//	ti := time.Unix(seconds, 0)
//	str := ti.UTC().Format("2006-01-02 15:04:05")
//}