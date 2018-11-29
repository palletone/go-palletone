package award

import (
	"github.com/palletone/go-palletone/core"
	"math"
	"time"
)

func CalculateAwardsForDepositContractNodes(amount uint64, startTimestamp int64) uint64 {
	coinDayUint64 := getCoinDay(amount, startTimestamp, time.Now().UTC())
	coinDayFloat64 := float64(coinDayUint64)
	//fmt.Println("coinDayFloat64", coinDayFloat64)
	//TODO
	yearRateFloat64 := core.DefaultDepositRate
	//yearRateFloat64 := 0.02
	//fmt.Println("yearRateFloat64", yearRateFloat64)
	awardsFloat64 := coinDayFloat64 * yearRateFloat64 / 365
	//fmt.Println(awardsFloat64)
	awardsUint64 := uint64(awardsFloat64)
	//fmt.Println(awardsUint64)
	return awardsUint64
}

//获取币的币龄
func getCoinDay(amount uint64, startTimestamp int64, endTime time.Time) uint64 {
	startTime := time.Unix(startTimestamp, 0).UTC()
	hourFloat64 := math.Floor(endTime.Sub(startTime).Hours())
	//fmt.Println(endTime.Sub(startTime).Hours(), hour)
	daysFloat64 := hourFloat64 / 24
	//fmt.Println(daysFloat64)
	daysUint64 := uint64(daysFloat64)
	//fmt.Println(daysUint64)
	coinDay := daysUint64 * amount
	return coinDay
}
