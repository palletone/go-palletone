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

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/modules"
)

const (
	globalPropDBKey    = "GlobalProperty"
	dynGlobalPropDBKey = "DynamicGlobalProperty"
)

type globalProperty struct {
	ChainParameters core.ChainParameters

	ActiveMediators []core.MediatorInfo
}

func getGPT(gp *modules.GlobalProperty) globalProperty {
	ams := make([]core.MediatorInfo, 0)

	for _, med := range gp.ActiveMediators {
		medInfo := core.MediatorToInfo(&med)
		ams = append(ams, medInfo)
	}

	gpt := globalProperty{
		ChainParameters: gp.ChainParameters,
		ActiveMediators: ams,
	}

	return gpt
}

func getGP(gpt *globalProperty) *modules.GlobalProperty {
	ams := make(map[common.Address]core.Mediator, 0)
	for _, medInfo := range gpt.ActiveMediators {
		med := core.InfoToMediator(&medInfo)
		ams[med.Address] = med
	}

	gp := modules.NewGlobalProp()
	gp.ChainParameters = gpt.ChainParameters
	gp.ActiveMediators = ams

	return gp
}

func StoreGlobalProp(gp *modules.GlobalProperty) {
	if Dbconn == nil {
		Dbconn = ReNewDbConn(dagconfig.DbPath)
	}

	gpt := getGPT(gp)

	err := Store(Dbconn, globalPropDBKey, gpt)
	if err != nil {
		log.Error(fmt.Sprintf("Store global properties error:%s", err))
	}
}

func StoreDynGlobalProp(dgp *modules.DynamicGlobalProperty) {
	if Dbconn == nil {
		Dbconn = ReNewDbConn(dagconfig.DbPath)
	}

	err := Store(Dbconn, dynGlobalPropDBKey, *dgp)
	if err != nil {
		log.Error(fmt.Sprintf("Store dynamic global properties error: %s", err))
	}
}

func RetrieveGlobalProp() *modules.GlobalProperty {
	gpt := new(globalProperty)

	err := Retrieve(globalPropDBKey, gpt)
	if err != nil {
		log.Error(fmt.Sprintf("Retrieve global properties error: %s", err))
	}

	gp := getGP(gpt)

	return gp
}

func RetrieveDynGlobalProp() *modules.DynamicGlobalProperty {
	dgp := modules.NewDynGlobalProp()

	err := Retrieve(dynGlobalPropDBKey, dgp)
	if err != nil {
		log.Error(fmt.Sprintf("Retrieve dynamic global properties error: %s", err))
	}

	return dgp
}
