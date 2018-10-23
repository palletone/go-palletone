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
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/constants"
)

var (
	GlobalPropDBKey    = append(constants.GLOBALPROPERTY_PREFIX, []byte("GlobalProperty")...)
	DynGlobalPropDBKey = append(constants.DYNAMIC_GLOBALPROPERTY_PREFIX, []byte("DynamicGlobalProperty")...)
)

type globalProperty struct {
	ChainParameters core.ChainParameters

	ActiveMediators []core.MediatorInfo
}

func getGPT(gp *modules.GlobalProperty) globalProperty {
	ams := make([]core.MediatorInfo, 0)

	for _, med := range gp.ActiveMediators {
		medInfo := med.MediatorToInfo()
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
		med := medInfo.InfoToMediator()
		ams[med.Address] = med
	}

	gp := modules.NewGlobalProp()
	gp.ChainParameters = gpt.ChainParameters
	gp.ActiveMediators = ams

	return gp
}

func StoreGlobalProp(db ptndb.Database, gp *modules.GlobalProperty) error {
	gpt := getGPT(gp)

	err := StoreBytes(db, GlobalPropDBKey, gpt)
	if err != nil {
		log.Error(fmt.Sprintf("Store global properties error:%s", err))
	}

	return err
}

func StoreDynGlobalProp(db ptndb.Database, dgp *modules.DynamicGlobalProperty) error {
	err := StoreBytes(db, DynGlobalPropDBKey, *dgp)
	if err != nil {
		log.Error(fmt.Sprintf("Store dynamic global properties error: %s", err))
	}

	return err
}

func RetrieveGlobalProp(db ptndb.Database) (*modules.GlobalProperty, error) {
	gpt := new(globalProperty)

	err := retrieve(db, GlobalPropDBKey, gpt)
	if err != nil {
		log.Error(fmt.Sprintf("Retrieve global properties error: %s", err))
	}

	gp := getGP(gpt)

	return gp, err
}

func RetrieveDynGlobalProp(db ptndb.Database) (*modules.DynamicGlobalProperty, error) {
	dgp := modules.NewDynGlobalProp()

	err := retrieve(db, DynGlobalPropDBKey, dgp)
	if err != nil {
		log.Error(fmt.Sprintf("Retrieve dynamic global properties error: %s", err))
	}

	return dgp, err
}
