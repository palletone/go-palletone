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

package common

import (
	"time"

	"github.com/palletone/go-palletone/dag/modules"
)

// UpdateGlobalDynProp, update global dynamic data
// @author Albert·Gou
func UpdateGlobalDynProp(gp *modules.GlobalProperty, dgp *modules.DynamicGlobalProperty, unit *modules.Unit) {
	when := time.Unix(unit.UnitHeader.Creationdate, 0)
	dgp.LastVerifiedUnitNum = unit.UnitHeader.Number.Index
	dgp.LastVerifiedUnitTime = when

	missedUnits := uint64(modules.GetSlotAtTime(gp, dgp, when))
	//	println(missedUnits)
	dgp.CurrentASlot += missedUnits + 1
}
