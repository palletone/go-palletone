///*
//	This file is part of go-palletone.
//	go-palletone is free software: you can redistribute it and/or modify
//	it under the terms of the GNU General Public License as published by
//	the Free Software Foundation, either version 3 of the License, or
//	(at your option) any later version.
//	go-palletone is distributed in the hope that it will be useful,
//	but WITHOUT ANY WARRANTY; without even the implied warranty of
//	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//	GNU General Public License for more details.
//	You should have received a copy of the GNU General Public License
//	along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
//*/
//
package deposit

import (
	"fmt"
	"testing"
	"time"
)

func TestTime(t *testing.T) {
	// 通过当前时间格式化
	now := time.Now().UTC()
	fmt.Println(now)
	l1 := now.Format("2006-01-02 15")
	fmt.Println(l1)
	l2 := now.Format("2006-01-02 15:04")
	fmt.Println(l2)
	l3 := now.Format("2006-01-02 15:04:05")
	fmt.Println(l3)
	//
	t1, _ := time.Parse("2006-01-02 15", l1)
	fmt.Println(t1)
	t2, _ := time.Parse("2006-01-02 15:04", l2)
	fmt.Println(t2)
	t3, _ := time.Parse("2006-01-02 15:04:05", l3)
	fmt.Println(t3)
	fmt.Println(t1.String())
}

func TestStrToTime(t *testing.T) {
	startTime := "2019-06-23 01:38:54 +0000 UTC"
	fmt.Println("startTime: ", startTime)
	fmt.Println("nowTime:   ", time.Now().UTC())
	st1 := StrToTime(startTime)
	fmt.Println(time.Since(st1))
	fmt.Println(int(time.Since(st1).Hours()))

	startTime2 := "2019-06-23 01:00:00 +0000 UTC"
	fmt.Println("startTime2: ", startTime2)
	fmt.Println("nowTime2:   ", time.Now().UTC())
	st2 := StrToTime(startTime2)
	fmt.Println(time.Since(st2))
	fmt.Println(int(time.Since(st2).Hours()))

	startTime3 := "2019-06-22 01:00:00 +0000 UTC"
	fmt.Println("startTime3: ", startTime3)
	fmt.Println("nowTime3:   ", time.Now().UTC())
	st3 := StrToTime(startTime3)
	fmt.Println(time.Since(st3))
	fmt.Println(int(time.Since(st3).Hours()))

	startTime4 := "2019-06-13 01:00:00 +0000 UTC"
	fmt.Println("startTime4: ", startTime4)
	fmt.Println("nowTime4:   ", time.Now().UTC())
	st4 := StrToTime(startTime4)
	fmt.Println(time.Since(st4))
	fmt.Println(int(time.Since(st4).Hours()))
	fmt.Println(time.Now().Unix())
	fmt.Println(time.Now().UTC().Unix())
	fmt.Println(StrToTime(time.Unix(time.Now().Unix(), 0).UTC().Format("2006-01-02 15:04:05")))
	t1, _ := time.Parse("2006-01-02 15:04:05", time.Unix(time.Now().Unix(), 0).UTC().Format("2006-01-02 15:04:05"))
	fmt.Println(t1)
	fmt.Println(t1.Unix())
	//ts := TimeStr()
	//fmt.Println(ts)
	//st := StrToTime(ts)
	//fmt.Println(st)
}

//
//import (
//	"fmt"
//	"github.com/palletone/go-palletone/common"
//	"testing"
//)
//
//func TestIsInCashbacklist(t *testing.T) {
//	//申请提保证金
//	//type Cashback struct {
//	//	CashbackAddress string        `json:"cashback_address"` //请求地址
//	//	CashbackTokens  *modules.InvokeTokens `json:"cashback_tokens"`  //请求数量
//	//	Role            string        `json:"role"`             //请求角色
//	//	CashbackTime    int64         `json:"cashback_time"`    //请求时间
//	//}
//	cashback := &Cashback{CashbackAddress: "P1GGtw1q4XUm5w5TXcZ6dEStidFLbEkipY8"}
//	cashList := []*Cashback{cashback}
//	isIn := isInCashbacklist("P1GGtw1q4XUm5w5TXcZ6dEStidFLbEkipY8", cashList)
//	if isIn {
//		t.Log("expected in")
//	} else {
//		t.Error("unexpected")
//	}
//
//	isNot := isInCashbacklist("P1GGtw1q4XUm5w5TXcZ6dEStidFLbEkipY9", cashList)
//	if isNot {
//		t.Error("unexpected")
//	} else {
//		t.Log("expected not in")
//	}
//}
//
//func TestMoveInApplyForForfeitureList(t *testing.T) {
//	fmt.Println("============", common.HexToAddress("0x853c98cb67ad40ddc3edc13f9ec52dea67a3a82200").String())
//	//申请没收保证金
//	//type Forfeiture struct {
//	//	ApplyAddress      string        `json:"apply_address"`      //谁发起的
//	//	ForfeitureAddress string        `json:"forfeiture_address"` //没收节点地址
//	//	ApplyTokens       *modules.InvokeTokens `json:"apply_tokens"`       //没收数量
//	//	ForfeitureRole    string        `json:"forfeiture_role"`    //没收角色
//	//	//Extra             string        `json:"extra"`              //备注
//	//	ApplyTime int64 `json:"apply_time"` //请求时间
//	//}
//	forfeiture := &Forfeiture{ForfeitureAddress: "P1GGtw1q4XUm5w5TXcZ6dEStidFLbEkipY8", ApplyTime: 123456}
//	forfeitureList := []*Forfeiture{forfeiture}
//	result, isExist := moveInApplyForForfeitureList(forfeitureList, "P1GGtw1q4XUm5w5TXcZ6dEStidFLbEkipY8", 123456)
//	if isExist && len(result) == 0 {
//		t.Log("expected move successfully")
//	} else {
//		t.Error("unexpected")
//	}
//
//	result, isExist = moveInApplyForForfeitureList(forfeitureList, "P1GGtw1q4XUm5w5TXcZ6dEStidFLbEkipY8", 123457)
//	if isExist && len(result) == 0 {
//		t.Error("unexpected")
//	} else {
//		t.Log("expected move failly")
//	}
//}
