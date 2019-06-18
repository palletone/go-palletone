package award

import (
	"fmt"
	"time"
)

//计算币龄所得奖励
func CalculateAwardsForDepositContractNodes(coinDays uint64, yearRateFloat64 float64) uint64 {
	coinDayFloat64 := float64(coinDays)
	fmt.Println("coinDayFloat64=", coinDayFloat64)
	//TODO
	//yearRateStr := core.DefaultDepositRate
	//yearRateFloat64, err := strconv.ParseFloat(yearRateStr, 64)
	//if err != nil {
	//	log.Errorf("string to float64 error: %s", err.Error())
	//	return 0
	//}
	//yearRateFloat64 := 0.02
	fmt.Println("yearRateFloat64=", yearRateFloat64)
	awardsFloat64 := coinDayFloat64 * yearRateFloat64 / 365
	fmt.Println("awardsFloat64=", awardsFloat64)
	awardsUint64 := uint64(awardsFloat64)
	fmt.Println("awardsUint64=", awardsUint64)
	return awardsUint64
}

//获取币的币龄
func GetCoinDay(amount uint64, lastModifyTime time.Time, endTime time.Time) uint64 {
	fmt.Println("startTime=", lastModifyTime)
	fmt.Println("endTime=", endTime)
	hourFloat64 := endTime.Sub(lastModifyTime).Hours()
	fmt.Println("hourFloat64=", hourFloat64)
	daysFloat64 := hourFloat64 / 24
	fmt.Println("daysFloat64=", daysFloat64)
	daysUint64 := uint64(daysFloat64)
	fmt.Println("daysUint64=", daysUint64)
	coinDays := daysUint64 * amount
	fmt.Println("coinDays=", coinDays)
	return coinDays
}

//计算币天的利息
func GetCoinDayInterest(receiveTime, spendTime int64, amount uint64, interestRate float64) uint64 {
	holdSecond := spendTime - receiveTime
	if spendTime == 0 || receiveTime == 0 || holdSecond <= 0 {
		return 0
	}
	holdDays := holdSecond / 86400 //24*60*60
	return uint64(float64(uint64(holdDays)*amount) * interestRate)

}

//直接获取持币的奖励
func GetAwardsWithCoins(coinAmount uint64, lastModifyTime int64, yearRateFloat64 float64) uint64 {
	//获取币龄
	startTime := time.Unix(lastModifyTime, 0).UTC()
	coinDays := GetCoinDay(coinAmount, startTime, time.Now().UTC())
	//计算币龄所得奖励
	awards := CalculateAwardsForDepositContractNodes(coinDays, yearRateFloat64)
	return awards
}
