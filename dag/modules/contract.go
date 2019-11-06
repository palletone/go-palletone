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
 * @author PalletOne core developers <dev@pallet.one>
 * @date 2018
 */

package modules

import (
	"github.com/palletone/go-palletone/common"
)

type Contract struct {
	// 根据用户创建合约实例申请的RequestId截取其后20字节生成
	ContractId   []byte
	TemplateId   []byte
	Name         string
	Status       byte   // 合约状态
	Creator      []byte // address 20bytes
	CreationTime uint64 // creation  date
	DuringTime   uint64 //合约部署持续时间，单位秒
}

func NewContract(templateId []byte, deploy *ContractDeployPayload, creator common.Address, unitTime uint64) *Contract {
	return &Contract{
		ContractId:   deploy.ContractId,
		TemplateId:   templateId,
		Name:         deploy.Name,
		Status:       1,
		Creator:      creator.Bytes(),
		CreationTime: unitTime,
		DuringTime:   deploy.DuringTime,
	}
}
