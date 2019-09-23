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

package modules

import "github.com/shopspring/decimal"

const (
	ListForApplyBecomeMediator = "ListForApplyBecomeMediator"
	ListForAgreeBecomeMediator = "ListForAgreeBecomeMediator"
	ListForQuit                = "ListForQuit"
	ListForForfeiture          = "ListForForfeiture"

	JuryApplyQuit      = "JuryApplyQuit"
	DeveloperApplyQuit = "DeveloperApplyQuit"

	Developer = "Developer"
	Jury      = "Jury"
	Mediator  = "Mediator"

	Ok = "ok"
	No = "no"

	//获取候选列表
	GetListForMediatorCandidate = "GetListForMediatorCandidate"
	GetListForJuryCandidate     = "GetListForJuryCandidate"
	GetListForDeveloper         = "GetListForDeveloper"
	//查看是否在候选列表中
	IsInMediatorCandidateList = "IsInMediatorCandidateList"
	IsInJuryCandidateList     = "IsInJuryCandidateList"
	IsInDeveloperList         = "IsInDeveloperList"
	//  是否在相应列表中
	IsInBecomeList = "IsInBecomeList"
	//IsInAgressList     = "IsInAgressList"
	IsInQuitList       = "IsInQuitList"
	IsInForfeitureList = "IsInForfeitureList"
	//获取列表
	GetBecomeMediatorApplyList      = "GetBecomeMediatorApplyList"
	GetAgreeForBecomeMediatorList   = "GetAgreeForBecomeMediatorList"
	GetQuitApplyList                = "GetQuitApplyList"
	GetListForForfeitureApplication = "GetListForForfeitureApplication"
	//申请
	ApplyForForfeitureDeposit     = "ApplyForForfeitureDeposit"
	DeveloperPayToDepositContract = "DeveloperPayToDepositContract"
	JuryPayToDepositContract      = "JuryPayToDepositContract"
	//基金会处理
	HandleForForfeitureApplication = "HandleForForfeitureApplication"
	HandleForApplyQuitMediator     = "HandleForApplyQuitMediator"
	HandleForApplyBecomeMediator   = "HandleForApplyBecomeMediator"
	HandleForApplyQuitJury         = "HandleForApplyQuitJury"
	HandleForApplyQuitDev          = "HandleForApplyQuitDev"
	HanldeNodeRemoveFromAgreeList  = "HanldeNodeRemoveFromAgreeList"

	GetDeposit     = "GetNodeBalance"
	GetJuryDeposit = "GetJuryDeposit"

	//  质押相关
	PledgeDeposit           = "PledgeDeposit"
	PledgeWithdraw          = "PledgeWithdraw"
	QueryPledgeStatusByAddr = "QueryPledgeStatusByAddr"
	QueryAllPledgeHistory   = "QueryAllPledgeHistory"
	HandlePledgeReward      = "HandlePledgeReward"
	AllPledgeVotes          = "allPledgeVotes"
	QueryPledgeList         = "QueryPledgeList"
	QueryPledgeListByDate   = "QueryPledgeListByDate"

	//  mediator状态
	Apply    = "Applying"
	Agree    = "Approved"
	Quitting = "Quitting"
	Quited   = "Quited"

	//  时间格式
	//  Layout1 = "2006-01-02 15"
	//  Layout2 = "2006-01-02 15:04"
	//  Layout3 = "2006-01-02 15:04:05"
	//  目前使用 time.Now().UTC().Format(Layout) 返回字符串
	Layout2 = "2006-01-02 15:04:05 MST"

	HandleMediatorInCandidateList = "HandleMediatorInCandidateList"
	HandleJuryInCandidateList     = "HandleJuryInCandidateList"
	HandleDevInList               = "HandleDevInList"
	GetAllMediator                = "GetAllMediator"
	GetAllNode                    = "GetAllNode"
	GetAllJury                    = "GetAllJury"
	UpdateJuryInfo                = "UpdateJuryInfo"
)

//申请退出
type QuitNode struct {
	//Address string `json:"address"` //请求地址
	Role string `json:"role"` //请求角色
	Time string `json:"time"` //请求时间
}

//申请没收保证金
type Forfeiture struct {
	ApplyAddress string `json:"apply_address"` //谁发起的
	//ForfeitureAddress string `json:"forfeiture_address"` //没收节点地址
	ForfeitureRole string `json:"forfeiture_role"` //没收角色
	Extra          string `json:"extra"`           //备注
	ApplyTime      string `json:"apply_time"`      //请求时间
}

//交易的内容
type PayValue struct {
	PayTokens *AmountAsset `json:"pay_tokens"` //数量和资产
	PayTime   string       `json:"pay_time"`   //发生时间
	//PayExtra  string        `json:"pay_extra"`  //额外内容
}

// 保证金信息
type DepositBalance struct {
	Balance   uint64 `json:"balance"`    // 保证金余额
	EnterTime string `json:"enter_time"` // 交保证金的时间
	Role      string `json:"role"`       // 角色，包括mediator、jury和developer
}

type Juror struct {
	DepositBalance
	JurorDepositExtra
}

type JurorDepositExtra struct {
	PublicKey string `json:"public_key"`
	Address   string `json:"address"`
}

// mediator保证金額外信息
type MediatorDepositExtra struct {
	ApplyEnterTime string `json:"apply_enter_time"` // 申请加入时间
	ApplyQuitTime  string `json:"apply_quit_time"`  // 申请退出时间
	Status         string `json:"status"`           // 申请状态  申请、同意、退出
	AgreeTime      string `json:"agree_time"`       // 基金会同意申请时间'
}

// mediator保证金信息
type MediatorDeposit struct {
	MediatorDepositExtra
	DepositBalance
}

func NewMediatorDeposit() *MediatorDeposit {
	md := &MediatorDeposit{}
	md.Status = Quited

	return md
}

type NorNodBal struct {
	AmountAsset  *AmountAsset `json:"amount_asset"`
	MediatorAddr string       `json:"mediator_address"`
}

type Member struct {
	Key   string `json:"key"`
	Value []byte `json:"value"`
}

type DepositBalanceJson struct {
	Balance   decimal.Decimal `json:"balance"`
	EnterTime string          `json:"enter_time"`
	Role      string          `json:"role"`
}

type MediatorDepositJson struct {
	MediatorDepositExtra
	DepositBalanceJson
}

type JuryDepositJson struct {
	DepositBalanceJson
	JurorDepositExtra
}
