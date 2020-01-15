package exchangecc

import (
	"github.com/shopspring/decimal"
	"testing"
)

func TestCalcDealAmount(t *testing.T) {
	takerDealAmount := uint64(500000000000)
	wantAmount := uint64(300000000000)
	saleAmount := 100000000
	takerDealAmountD := decimal.New(int64(takerDealAmount), 0)
	wantAmountD := decimal.New(int64(wantAmount), 0)
	saleAmountD := decimal.New(int64(saleAmount), 0)
	makerDealAmountD := takerDealAmountD.Mul(saleAmountD).Div(wantAmountD) //*float64(wantAmount)/float64(saleAmount))
	t.Log(makerDealAmountD)
	makerDealAmount := uint64(float64(takerDealAmount) * float64(saleAmount) / float64(wantAmount))
	t.Log(makerDealAmount)
}
