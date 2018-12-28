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
/*
 * @author PalletOne core developer Albert·Gou <dev@pallet.one>
 * @date 2018
 */

package modules

import (
	"github.com/palletone/go-palletone/common"
)

type MediatorCreateOperation struct {
	AddStr     string `json:"account"`
	InitPubKey string `json:"initPubKey"`
	Node       string `json:"node"`
	Url        string `json:"url"`
}

func (mco *MediatorCreateOperation) FeePayer() common.Address {
	addr, _ := common.StringToAddress(mco.AddStr)

	return addr
}

func (mco *MediatorCreateOperation) Validate() bool {
	// todo 判断是否已经申请缴纳保证金
	return true
}
