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
 *  * @author PalletOne core developer <dev@pallet.one>
 *  * @date 2018
 *
 */

package modules

import "github.com/palletone/go-palletone/core"

// only for serialization(storage)
type MediatorInfo struct {
	*core.MediatorInfoBase
	*core.MediatorInfoExpand
}

func NewMediatorInfo() *MediatorInfo {
	return &MediatorInfo{
		MediatorInfoBase:   core.NewMediatorInfoBase(),
		MediatorInfoExpand: core.NewMediatorBase(),
	}
}

func MediatorToInfo(md *core.Mediator) *MediatorInfo {
	mi := NewMediatorInfo()
	mi.AddStr = md.Address.Str()
	mi.InitPubKey = core.PointToStr(md.InitPubKey)
	mi.Node = md.Node.String()
	mi.MediatorInfoExpand = md.MediatorInfoExpand

	return mi
}

func (mi *MediatorInfo) InfoToMediator() *core.Mediator {
	md := core.NewMediator()
	md.Address = core.StrToMedAdd(mi.AddStr)
	md.InitPubKey, _ = core.StrToPoint(mi.InitPubKey)
	md.Node = core.StrToMedNode(mi.Node)
	md.MediatorInfoExpand = mi.MediatorInfoExpand

	return md
}
