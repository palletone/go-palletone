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
 *
 */

package core

// ChainParameters 区块链网络参数结构体的定义
//变量名一定要大些，否则外部无法访问，导致无法进行json编码和解码
type ChainParameters struct {
	// 验证单元之间的间隔时间，以秒为单元。 interval in seconds between verifiedUnits
	MediatorInterval uint8 `json:"mediatorInterval"`

	// 在维护时跳过的verifiedUnitInterval数量。 number of verifiedUnitInterval to skip at maintenance time
	//	MaintenanceSkipSlots uint8
}

func NewChainParams() ChainParameters {
	return ChainParameters{
		MediatorInterval: DefaultMediatorInterval,
	}
}
