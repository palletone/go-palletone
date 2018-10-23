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

package common

import (
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/dag/modules"
)

type MediatorCreateOperation struct {
	core.MediatorInfo
}

func (mco *MediatorCreateOperation) Validate() bool {
	return true
}

func feePayer(tx *modules.Transaction) (common.Address, error) {
	return getRequesterAddress(tx)
}

func (mco *MediatorCreateOperation) Evaluate() bool {
	return true
}

func (mco *MediatorCreateOperation) Apply() {
	return
}
