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
 * @brief 主要实现mediator调度相关的功能。implements mediator scheduling related functions.
 */
package modules

import "github.com/palletone/go-palletone/common"

type MemdagInfos struct {
	MemStatus map[string]*MemdagStatus `json:"memdag_status"`
}

type MemdagStatus struct {
	Token         AssetId                  `json:"token"`
	StableHeader  *Header                  `json:"stable_header"`  // 最新稳定单元header
	FastHeader    *Header                  `json:"fast_header"`    // 最新不稳定单元header
	Forks         map[uint64][]common.Hash `json:"forks"`          // 分叉单元
	UnstableUnits []common.Hash            `json:"unstable_units"` // 不稳定单元
	OrphanUnits   []common.Hash            `json:"orphan_units"`   // 孤儿单元
}
