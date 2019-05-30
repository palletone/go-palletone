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
	ApplyMediator           = "ApplyBecomeMediator"
	IsApproved              = "IsSelected"
	MediatorPayDeposit      = "MediatorPayToDepositContract"
	MediatorList            = "MediatorList"
	GetMediatorDeposit      = "GetMediatorDeposit"
	MediatorWithdrawDeposit = "MediatorApplyCashback"
	MediatorApplyQuitList   = "MediatorApplyQuitMediator"
)

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
	mi.InitPubKey = core.PointToStr(md.InitPubKey)
	mi.Node = md.Node.String()
	*mi.MediatorApplyInfo = *md.MediatorApplyInfo
	*mi.MediatorInfoExpand = *md.MediatorInfoExpand

	return mi
}

func (mi *MediatorInfo) InfoToMediator() *core.Mediator {
	md := core.NewMediator()
	md.Address, _ = core.StrToMedAdd(mi.AddStr)
	md.InitPubKey, _ = core.StrToPoint(mi.InitPubKey)
	md.Node, _ = core.StrToMedNode(mi.Node)
	*md.MediatorApplyInfo = *mi.MediatorApplyInfo
	*md.MediatorInfoExpand = *mi.MediatorInfoExpand

	return md
}

type MediatorCreateOperation struct {
	*core.MediatorInfoBase
	*core.MediatorApplyInfo
}

func (mco *MediatorCreateOperation) FeePayer() common.Address {
	addr, _ := common.StringToAddress(mco.AddStr)

	return addr
}
