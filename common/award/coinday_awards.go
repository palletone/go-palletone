package award

import (
	"fmt"
	"github.com/palletone/go-palletone/core"
	"time"
)

//计算币龄所得奖励
func CalculateAwardsForDepositContractNodes(coinDays uint64) uint64 {
	coinDayFloat64 := float64(coinDays)
	fmt.Println("coinDayFloat64=", coinDayFloat64)
	//TODO
	yearRateFloat64 := core.DefaultDepositRate
	//yearRateFloat64 := 0.02
	fmt.Println("yearRateFloat64", yearRateFloat64)
	awardsFloat64 := coinDayFloat64 * yearRateFloat64 / 365
	fmt.Println("awardsFloat64", awardsFloat64)
	awardsUint64 := uint64(awardsFloat64)
	fmt.Println("awardsUint64", awardsUint64)
	return awardsUint64
}

//获取币的币龄
func GetCoinDay(amount uint64, lastModifyTime int64, endTime time.Time) uint64 {
	startTime := time.Unix(lastModifyTime, 0).UTC()
	fmt.Println("startTime=", startTime)
	fmt.Println("endTime=", endTime)
	hourFloat64 := endTime.Sub(startTime).Hours()
	fmt.Println("hourFloat64=", hourFloat64)
	daysFloat64 := hourFloat64 / 24
	fmt.Println("daysFloat64=", daysFloat64)
	daysUint64 := uint64(daysFloat64)
	fmt.Println("daysUint64=", daysUint64)
	coinDays := daysUint64 * amount
	fmt.Println("coinDays", coinDays)
	return coinDays
}
