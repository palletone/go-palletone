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
 * @date 2018/11/05
 */

package mediatorplugin

import (
	"github.com/palletone/go-palletone/core"
)

type PrivateMediatorAPI struct {
	*MediatorPlugin
}

func NewPrivateMediatorAPI(mp *MediatorPlugin) *PrivateMediatorAPI {
	return &PrivateMediatorAPI{mp}
}

// 创建 mediator 所需的参数, 至少包含普通账户地址
type MediatorCreateArgs struct {
	core.MediatorInfo
}

// 创建 mediator 的执行结果，包含交易哈希，初始dks
type MediatorCreateResult struct {
}

func (a *PrivateMediatorAPI) Create(args MediatorCreateArgs) *MediatorCreateResult {
	// 1. 组装 message

	// 2. 组装 tx

	// 3. 将 tx 放入 pool

	// 4. 返回执行结果

	return nil
}
