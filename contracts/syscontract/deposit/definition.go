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
	"errors"
	"github.com/palletone/go-palletone/dag/modules"
)

const (
	ListForQuit                = "ListForQuit"
	ListForForfeiture          = "ListForForfeiture"
	JuryApplyQuit              = "JuryApplyQuit"
	DeveloperApplyQuit         = "DeveloperApplyQuit"
	ListForApplyBecomeMediator = "ListForApplyBecomeMediator"
	ListForAgreeBecomeMediator = "ListForAgreeBecomeMediator"
	ListForApplyQuitMediator   = "ListForApplyQuitMediator"
	DepositAmountForJury       = "DepositAmountForJury"
	DepositAmountForDeveloper  = "DepositAmountForDeveloper"
	DepositPeriod              = "DepositPeriod"
	Developer                  = "Developer"
	Jury                       = "Jury"
	JuryAndDev                 = "juryAndDev"
	NormalNode                 = "normalNode"
	Mediator                   = "Mediator"
	Ok                         = "ok"
	No                         = "no"
	DTimeDuration              = 1800
	//获取Mediator候选列表
	GetListForMediatorCandidate = "GetListForMediatorCandidate"
	GetQuitApplyList            = "GetQuitApplyList"
	//查看是否在候选列表中
	IsInMediatorCandidateList       = "IsInMediatorCandidateList"
	GetQuitList                     = "GetQuitList"
	GetAgreeForBecomeMediatorList   = "GetAgreeForBecomeMediatorList"
	GetBecomeMediatorApplyList      = "GetBecomeMediatorApplyList"
	GetListForDeveloperCandidate    = "GetListForDeveloperCandidate"
	GetListForJuryCandidate         = "GetListForJuryCandidate"
	GetListForForfeitureApplication = "GetListForForfeitureApplication"
	GetListForCashbackApplication   = "GetListForCashbackApplication"
	HandleForForfeitureApplication  = "HandleForForfeitureApplication"
	ApplyForForfeitureDeposit       = "ApplyForForfeitureDeposit"
	HandleForDeveloperApplyCashback = "HandleForDeveloperApplyCashback"
	DeveloperApplyCashback          = "DeveloperApplyCashback"
	HandleForJuryApplyCashback      = "HandleForJuryApplyCashback"
	JuryApplyCashback               = "JuryApplyCashback"
	HandleForMediatorApplyCashback  = "HandleForMediatorApplyCashback"
	DeveloperPayToDepositContract   = "DeveloperPayToDepositContract"
	JuryPayToDepositContract        = "JuryPayToDepositContract"

	HandleForApplyQuitMediator = "HandleForApplyQuitMediator"
	HandleForApplyQuitJury     = "HandleForApplyQuitJury"
	HandleForApplyQuitDev      = "HandleForApplyQuitDev"

	HandleForApplyBecomeMediator = "HandleForApplyBecomeMediator"
	IsInMediatorQuitList         = "IsInMediatorQuitList"
	IsInCashbackList             = "IsInCashbackList"
	IsInJuryCandidateList        = "IsInJuryCandidateList"
	IsInDeveloperCandidateList   = "IsInDeveloperCandidateList"
	GetDeposit                   = "GetNodeBalance"
	NormalNodePledgeVote         = "normalNodePledgeVote"
	NormalNodeChangeVote         = "normalNodeChangeVote"
	NormalNodeExtractVote        = "normalNodeExtractVote"
	NormalNodeList               = "normalNodeList"
	ExtractPtnList               = "extractPtnList"
	HandleExtractVote            = "handleExtractVote"
	HandleEachDayAward           = "handleEachDayAward"
	AllPledgeVotes               = "allPledgeVotes"
	HandleEachDay                = "handleEachDay"
	MemberList                   = "MemberList"
	MemberListLastDate           = "MemberListLastDate"
	Apply                        = "applying"
	Agree                        = "approved"
	Quitting                     = "quitting"
	Quited                       = "quited"
	//  时间格式
	//  Layout1 = "2006-01-02 15"
	//  Layout2 = "2006-01-02 15:04"
	//  Layout3 = "2006-01-02 15:04:05"
	//  目前使用 time.Now().UTC().Format(Layout) 返回字符串
	Layout1 = "2006-01-02 15"
	Layout2 = "2006-01-02 15:04:05"
)

//申请退出
type Cashback struct {
	CashbackAddress string `json:"cashback_address"` //请求地址
	Role            string `json:"role"`             //请求角色
	CashbackTime    string `json:"cashback_time"`    //请求时间
}

//申请没收保证金
type Forfeiture struct {
	ApplyAddress      string               `json:"apply_address"`      //谁发起的
	ForfeitureAddress string               `json:"forfeiture_address"` //没收节点地址
	ApplyTokens       *modules.AmountAsset `json:"apply_tokens"`       //没收数量
	ForfeitureRole    string               `json:"forfeiture_role"`    //没收角色
	Extra             string               `json:"extra"`              //备注
	ApplyTime         string               `json:"apply_time"`         //请求时间
}

//交易的内容
type PayValue struct {
	PayTokens *modules.AmountAsset `json:"pay_tokens"` //数量和资产
	PayTime   string               `json:"pay_time"`   //发生时间
	//PayExtra  string        `json:"pay_extra"`  //额外内容
}

//节点状态数据库保存值
//type DepositBalance struct {
//	TotalAmount      uint64        `json:"total_amount"`      //保证金总量
//	LastModifyTime   int64         `json:"last_modify_time"`  //最后一次改变，主要来计算币龄收益
//	EnterTime        string        `json:"enter_time"`        //这是加入列表时的时间
//	PayValues        []*PayValue   `json:"pay_values"`        //交付的历史记录
//	CashbackValues   []*Cashback   `json:"cashback_values"`   //退款的历史记录
//	ForfeitureValues []*Forfeiture `json:"forfeiture_values"` //被没收的历史记录
//}

type DepositBalance struct {
	Balance        uint64 `json:"balance"`          //  保证金余额
	EnterTime      string `json:"enter_time"`       //  交保证金的时间
	LastModifyTime string `json:"last_modify_time"` //  计算币龄时间
}

type MediatorDeposit struct {
	ApplyEnterTime string `json:"apply_enter_time"` //  申请加入时间
	ApplyQuitTime  string `json:"apply_quit_time"`  //  申请退出时间
	Status         string `json:"status"`           //  申请状态  申请、同意、退出
	AgreeTime      string `json:"agree_time"`       //  基金会同意申请时间'
	DepositBalance
}

func NewMediatorDeposit() *MediatorDeposit {
	return &MediatorDeposit{
		Status: Quited,
	}
}

type NorNodBal struct {
	AmountAsset  *modules.AmountAsset `json:"amount_asset"`
	MediatorAddr string               `json:"mediator_address"`
}

type extractPtn struct {
	Time   string `json:"time"`   //提取质押时间
	Amount uint64 `json:"amount"` //提取质押数量
}

//质押列表
type PledgeList struct {
	TotalAmount uint64           `json:"total_amount"`
	Date        string           `json:"date"` //质押列表所在的日期yyyyMMdd
	Members     []*AddressAmount `json:"members"`
}

//账户质押情况
type AddressAmount struct {
	Address string `json:"address"`
	Amount  uint64 `json:"amount"`
}

func (pl *PledgeList) Add(addr string, amount uint64) {
	pl.TotalAmount += amount
	for _, p := range pl.Members {
		if p.Address == addr {
			p.Amount += amount
			return
		}
	}
	pl.Members = append(pl.Members, &AddressAmount{Address: addr, Amount: amount})
}

//从质押列表中提币，Amount 0表示全部提取
func (pl *PledgeList) Reduce(addr string, amount uint64) (uint64, error) {
	for i, p := range pl.Members {
		if p.Address == addr {
			if amount == 0 {
				amount = p.Amount //如果是0表示全部提取
			}
			if p.Amount < amount {
				return 0, errors.New("Not enough amount")
			}
			pl.TotalAmount -= amount
			if p.Amount == amount {
				pl.Members = append(pl.Members[:i], pl.Members[i+1:]...)
				return amount, nil
			}
			p.Amount -= amount
			return amount, nil
		}
	}
	return 0, errors.New("Address not found")
}

type Member struct {
	Key   string `json:"key"`
	Value []byte `json;"value"`
}
