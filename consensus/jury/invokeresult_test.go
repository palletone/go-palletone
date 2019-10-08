/*
 *
 *    This file is part of go-palletone.
 *    go-palletone is free software: you can redistribute it and/or modify
 *    it under the terms of the GNU General Public License as published by
 *    the Free Software Foundation, either version 3 of the License, or
 *    (at your option) any later version.
 *    go-palletone is distributed in the hope that it will be useful,
 *    but WITHOUT ANY WARRANTY; without even the implied warranty of
 *    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *    GNU General Public License for more details.
 *    You should have received a copy of the GNU General Public License
 *    along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
 * /
 *
 *  * @author PalletOne core developers <dev@pallet.one>
 *  * @date 2018-2019
 *
 *
 */

package jury

import (
	"encoding/json"
	"github.com/golang/mock/gomock"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/dag"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	ptn      = modules.NewPTNAsset()
	btc, _   = modules.StringToAsset("BTC")
	addr1, _ = common.StringToAddress("P1JxYp1dRpq2ZeYk58XEkSZJrptYEeuvZyq")
	addr2, _ = common.StringToAddress("P1KP5TZwTY8UowE7X3zSZ3gZDHqwCqcCThR")
)

func TestTokenPayOutGroupByAsset(t *testing.T) {

	//addr3,_:=common.StringToAddress("P1PuhsNTmpsSV36wyoEF49b5dhRdaTQYC2C")
	//no payout
	pay0 := []*modules.TokenPayOut{}
	g0 := tokenPayOutGroupByAsset(pay0)
	assert.Equal(t, 0, len(g0))
	pay1 := []*modules.TokenPayOut{{PayTo: addr1, Amount: 123, Asset: ptn}}
	//only 1 payout
	g1 := tokenPayOutGroupByAsset(pay1)
	assert.Equal(t, 1, len(g1))
	assert.EqualValues(t, 123, g1[*ptn][0].Amount)
	//2 payouts,but same address, same asset
	pay2 := append(pay1, &modules.TokenPayOut{PayTo: addr1, Amount: 1, Asset: ptn})
	g2 := tokenPayOutGroupByAsset(pay2)
	assert.Equal(t, 1, len(g2))
	assert.EqualValues(t, 124, g2[*ptn][0].Amount)
	//3payments,addr1:2 pay  addr2: 1 pay
	pay3 := append(pay2, &modules.TokenPayOut{PayTo: addr2, Amount: 1, Asset: ptn})
	g3 := tokenPayOutGroupByAsset(pay3)
	assert.Equal(t, 1, len(g3))
	assert.Equal(t, 2, len(g3[*ptn]))
	for _, aa := range g3[*ptn] {
		if aa.Address == addr1 {
			assert.EqualValues(t, 124, aa.Amount)
		}
		if aa.Address == addr2 {
			assert.EqualValues(t, 1, aa.Amount)
		}
	}
	pay4 := append(pay3, &modules.TokenPayOut{PayTo: addr1, Amount: 1, Asset: btc})
	g4 := tokenPayOutGroupByAsset(pay4)
	assert.Equal(t, 2, len(g4))
}

func TestResultToContractPayments(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mdag := dag.NewMockIDag(mockCtrl)
	mockResult := make(map[modules.OutPoint]*modules.Utxo)
	mockResult[modules.OutPoint{MessageIndex: 0, OutIndex: 1}] = &modules.Utxo{Amount: 100, Asset: ptn}
	mockResult[modules.OutPoint{MessageIndex: 0, OutIndex: 2}] = &modules.Utxo{Amount: 200, Asset: ptn}
	mockResult[modules.OutPoint{MessageIndex: 0, OutIndex: 3}] = &modules.Utxo{Amount: 150, Asset: ptn}
	mdag.EXPECT().GetAddr1TokenUtxos(gomock.Any(), gomock.Any()).
		Return(mockResult, nil).AnyTimes()

	payouts := []*modules.TokenPayOut{
		{PayTo: addr1, Amount: 123, Asset: ptn},
		{PayTo: addr2, Amount: 321, Asset: ptn},
	}
	result := &modules.ContractInvokeResult{TokenPayOut: payouts}

	payment, err := resultToContractPayments(mdag, result)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(payment))
	d, _ := json.Marshal(payment)
	t.Log(string(d))
}
