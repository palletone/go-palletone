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
 *  * @author PalletOne core developer Albert·Gou <dev@pallet.one>
 *  * @date 2018
 *
 */

package modules

import (
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/core"
)

const (
	ApplyMediator      = "ApplyBecomeMediator"
	IsApproved         = "IsInAgreeList"
	MediatorPayDeposit = "MediatorPayToDepositContract"
	MediatorList       = "MediatorList"
	GetMediatorDeposit = "GetMediatorDeposit"
	MediatorApplyQuit  = "MediatorApplyQuit"
	UpdateMediatorInfo = "UpdateMediatorInfo"
)

// mediator 信息
type MediatorInfo struct {
	*core.MediatorInfoBase
	*core.MediatorApplyInfo
	*core.MediatorInfoExpand
}

func NewMediatorInfo() *MediatorInfo {
	return &MediatorInfo{
		MediatorInfoBase:   core.NewMediatorInfoBase(),
		MediatorApplyInfo:  core.NewMediatorApplyInfo(),
		MediatorInfoExpand: core.NewMediatorInfoExpand(),
	}
}

func MediatorToInfo(md *core.Mediator) *MediatorInfo {
	mi := NewMediatorInfo()
	mi.AddStr = md.Address.Str()
	mi.RewardAdd = md.RewardAdd.Str()
	mi.InitPubKey = core.PointToStr(md.InitPubKey)
	mi.Node = md.Node.String()
	*mi.MediatorApplyInfo = *md.MediatorApplyInfo
	*mi.MediatorInfoExpand = *md.MediatorInfoExpand

	return mi
}

func (mi *MediatorInfo) InfoToMediator() *core.Mediator {
	md := core.NewMediator()
	md.Address, _ = core.StrToMedAdd(mi.AddStr)
	md.RewardAdd, _ = core.StrToMedAdd(mi.RewardAdd)
	md.InitPubKey, _ = core.StrToPoint(mi.InitPubKey)
	md.Node, _ = core.StrToMedNode(mi.Node)
	*md.MediatorApplyInfo = *mi.MediatorApplyInfo
	*md.MediatorInfoExpand = *mi.MediatorInfoExpand

	return md
}

// 创建 mediator 所需的参数
type MediatorCreateArgs struct {
	*core.MediatorInfoBase
	*core.MediatorApplyInfo
}

func NewMediatorCreateArgs() *MediatorCreateArgs {
	return &MediatorCreateArgs{
		MediatorInfoBase:  core.NewMediatorInfoBase(),
		MediatorApplyInfo: core.NewMediatorApplyInfo(),
	}
}

// 更新 mediator 信息所需参数
type MediatorUpdateArgs struct {
	AddStr      string  `json:"account"`              // 要更新的mediator地址
	RewardAdd   *string `json:"rewardAdd" rlp:"nil"`  // mediator奖励地址，主要用于接收产块奖励
	InitPubKey  *string `json:"initPubKey" rlp:"nil"` // mediator的群签名初始公钥
	Node        *string `json:"node" rlp:"nil"`       // 节点网络信息，包括ip和端口等
	Logo        *string `json:"logo" rlp:"nil"`       // 节点图标url
	Name        *string `json:"name" rlp:"nil"`       // 节点名称
	Location    *string `json:"loc" rlp:"nil"`        // 节点所在地区
	Url         *string `json:"url" rlp:"nil"`        // 节点宣传网站
	Description *string `json:"applyInfo" rlp:"nil"`  // 节点详细信息描述
}

func (mua *MediatorUpdateArgs) Validate() (common.Address, error) {
	addr, err := core.StrToMedAdd(mua.AddStr)
	if err != nil {
		return addr, err
	}

	if mua.RewardAdd != nil {
		_, err := core.StrToMedAdd(*mua.RewardAdd)
		if err != nil {
			return addr, err
		}
	}

	if mua.InitPubKey != nil {
		_, err = core.StrToPoint(*mua.InitPubKey)
		if err != nil {
			return addr, err
		}
	}

	if mua.Node != nil {
		node, err := core.StrToMedNode(*mua.Node)
		if err != nil {
			return addr, err
		}

		err = node.ValidateComplete()
		if err != nil {
			return addr, err
		}
	}

	return addr, nil
}
