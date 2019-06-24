package award

import (
	"fmt"
	"time"
)

//计算币龄所得奖励
func CalculateAwardsForDepositContractNodes(coinDays uint64, yearRateFloat64 float64) uint64 {
	fmt.Println("获取币龄利息 ==========================")
	coinDayFloat64 := float64(coinDays)
	fmt.Println("当前年利率 = ", yearRateFloat64)
	awardsFloat64 := coinDayFloat64 * yearRateFloat64 / 365
	fmt.Println("计算利息公式 = ", coinDayFloat64, " * ", yearRateFloat64, " / ", 365)
	fmt.Println("所得币龄利息 = ", uint64(awardsFloat64))
	fmt.Println("获取币龄利息 ==========================")
	return uint64(awardsFloat64)
}

//获取币的币龄
func GetCoinDay(amount uint64, lastModifyTime time.Time) uint64 {
	fmt.Println("获取币龄 ==========================")
	fmt.Println("币的总量 = ", amount)
	fmt.Println("开始时间 = ", lastModifyTime)
	fmt.Println("当前时间 = ", time.Now().UTC())
	dur := time.Since(lastModifyTime)
	fmt.Println("时间间隔 = ", dur)
	fmt.Println("共 ", dur.Hours(), " 个小时，即 ", uint64(dur.Hours()), " 个小时")
	fmt.Println("共 ", int(dur.Hours())/24, " 天")
	coinDays := amount * (uint64(dur.Hours()) / 24)
	fmt.Println("共 ", coinDays, " 币龄")
	fmt.Println("获取币龄 ==========================")
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
func GetAwardsWithCoins(coinAmount uint64, lastModifyTime time.Time, yearRateFloat64 float64) uint64 {
	//获取币龄
	//startTime := time.Unix(lastModifyTime, 0).UTC()
	coinDays := GetCoinDay(coinAmount, lastModifyTime)
	//计算币龄所得奖励
	awards := CalculateAwardsForDepositContractNodes(coinDays, yearRateFloat64)
	return awards
}
