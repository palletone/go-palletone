/*
	This file is part of go-palletone.
	go-palletone is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.
	go-palletone is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.
	You should have received a copy of the GNU General Public License
	along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
*/

package deposit

import (
	"testing"
)

func TestIsInCashbacklist(t *testing.T){
	//申请提保证金
	//type Cashback struct {
	//	CashbackAddress string        `json:"cashback_address"` //请求地址
	//	CashbackTokens  *modules.InvokeTokens `json:"cashback_tokens"`  //请求数量
	//	Role            string        `json:"role"`             //请求角色
	//	CashbackTime    int64         `json:"cashback_time"`    //请求时间
	//}
	cashback := &Cashback{CashbackAddress:"P1GGtw1q4XUm5w5TXcZ6dEStidFLbEkipY8"}
	cashList := []*Cashback{cashback}
	isIn := isInCashbacklist("P1GGtw1q4XUm5w5TXcZ6dEStidFLbEkipY8",cashList)
	if isIn{
		t.Log("expected in")
	}else {
		t.Error("unexpected")
	}

	isNot := isInCashbacklist("P1GGtw1q4XUm5w5TXcZ6dEStidFLbEkipY9",cashList)
	if isNot {
		t.Error("unexpected")
	}else {
		t.Log("expected not in")
	}
}

func TestMoveInApplyForForfeitureList(t *testing.T) {
	//申请没收保证金
	//type Forfeiture struct {
	//	ApplyAddress      string        `json:"apply_address"`      //谁发起的
	//	ForfeitureAddress string        `json:"forfeiture_address"` //没收节点地址
	//	ApplyTokens       *modules.InvokeTokens `json:"apply_tokens"`       //没收数量
	//	ForfeitureRole    string        `json:"forfeiture_role"`    //没收角色
	//	//Extra             string        `json:"extra"`              //备注
	//	ApplyTime int64 `json:"apply_time"` //请求时间
	//}
	forfeiture := &Forfeiture{ForfeitureAddress:"P1GGtw1q4XUm5w5TXcZ6dEStidFLbEkipY8",ApplyTime:123456}
	forfeitureList := []*Forfeiture{forfeiture}
	result,isExist := moveInApplyForForfeitureList(forfeitureList,"P1GGtw1q4XUm5w5TXcZ6dEStidFLbEkipY8",123456)
	if isExist && len(result) == 0 {
		t.Log("expected move successfully")
	} else {
		t.Error("unexpected")
	}

	result, isExist = moveInApplyForForfeitureList(forfeitureList,"P1GGtw1q4XUm5w5TXcZ6dEStidFLbEkipY8",123457)
	if isExist && len(result) == 0 {
		t.Error("unexpected")
	} else {
		t.Log("expected move failly")
	}
}
