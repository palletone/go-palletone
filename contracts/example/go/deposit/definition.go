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

import "github.com/palletone/go-palletone/dag/modules"

//申请成为Mediator
type MediatorRegisterInfo struct {
	Address string `json:"address"`
	Content string `json:"content"`
	Time    int64  `json:"time"`
}

//申请提保证金
type Cashback struct {
	CashbackAddress string                `json:"cashback_address"` //请求地址
	CashbackTokens  *modules.InvokeTokens `json:"cashback_tokens"`  //请求数量
	Role            string                `json:"role"`             //请求角色
	CashbackTime    int64                 `json:"cashback_time"`    //请求时间
}

//申请没收保证金
type Forfeiture struct {
	ApplyAddress      string                `json:"apply_address"`      //谁发起的
	ForfeitureAddress string                `json:"forfeiture_address"` //没收节点地址
	ApplyTokens       *modules.InvokeTokens `json:"apply_tokens"`       //没收数量
	ForfeitureRole    string                `json:"forfeiture_role"`    //没收角色
	//Extra             string        `json:"extra"`              //备注
	ApplyTime int64 `json:"apply_time"` //请求时间
}

//交易的内容
type PayValue struct {
	PayTokens *modules.InvokeTokens `json:"pay_tokens"` //数量和资产
	PayTime   int64                 `json:"pay_time"`   //发生时间
	//PayExtra  string        `json:"pay_extra"`  //额外内容
}

//节点状态数据库保存值
type DepositBalance struct {
	TotalAmount      uint64        `json:"total_amount"`      //保证金总量
	LastModifyTime   int64         `json:"last_modify_time"`  //最后一次改变，主要来计算币龄收益
	EnterTime        int64         `json:"enter_time"`        //这是加入列表时的时间
	PayValues        []*PayValue   `json:"pay_values"`        //交付的历史记录
	CashbackValues   []*Cashback   `json:"cashback_values"`   //退款的历史记录
	ForfeitureValues []*Forfeiture `json:"forfeiture_values"` //被没收的历史记录
}
