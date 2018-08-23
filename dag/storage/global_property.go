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

package storage

import (
	"fmt"

	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/modules"
)

const (
	globalPropDBKey    = "GlobalProperty"
	dynGlobalPropDBKey = "DynamicGlobalProperty"
)

func StoreGlobalProp(gp *modules.GlobalProperty) {
	if Dbconn == nil {
		Dbconn = ReNewDbConn(dagconfig.DbPath)
	}
	err := Store(Dbconn, globalPropDBKey, *gp)
	if err != nil {
		log.Info(fmt.Sprintf("Store global properties error:%s", err))
	}
}

func StoreDynGlobalProp(dgp *modules.DynamicGlobalProperty) {
	if Dbconn == nil {
		Dbconn = ReNewDbConn(dagconfig.DbPath)
	}
	err := Store(Dbconn, dynGlobalPropDBKey, *dgp)
	if err != nil {
		log.Info(fmt.Sprintf("Store dynamic global properties error: %s", err))
	}
}

func RetrieveGlobalProp() *modules.GlobalProperty {
	gp := modules.NewGlobalProp()

	err := Retrieve(globalPropDBKey, gp)
	if err != nil {
		log.Info(fmt.Sprintf("Retrieve global properties error: %s", err))
	}

	return gp
}

func RetrieveDynGlobalProp() *modules.DynamicGlobalProperty {
	dgp := modules.NewDynGlobalProp()

	err := Retrieve(dynGlobalPropDBKey, dgp)
	if err != nil {
		log.Info(fmt.Sprintf("Retrieve dynamic global properties error: %s", err))
	}

	return dgp
}
