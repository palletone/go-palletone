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
	"github.com/palletone/go-palletone/dag/constants"
	"github.com/palletone/go-palletone/dag/modules"
)

var (
	GlobalPropDBKey    = append(constants.GLOBALPROPERTY_PREFIX, []byte("GlobalProperty")...)
	DynGlobalPropDBKey = append(constants.DYNAMIC_GLOBALPROPERTY_PREFIX, []byte("DynamicGlobalProperty")...)
)

// only for serialization(storage)
type globalProperty struct {
	*modules.GlobalPropBase

	ActiveJuries       []common.Address
	ActiveMediators    []common.Address
	PrecedingMediators []common.Address
}

func getGPT(gp *modules.GlobalProperty) *globalProperty {
	ajs := make([]common.Address, 0)
	ams := make([]common.Address, 0)
	pms := make([]common.Address, 0)

	for juryAdd, _ := range gp.ActiveJuries {
		ajs = append(ajs, juryAdd)
	}

	for medAdd, _ := range gp.ActiveMediators {
		ams = append(ams, medAdd)
	}

	for medAdd, _ := range gp.PrecedingMediators {
		pms = append(pms, medAdd)
	}

	gpt := &globalProperty{
		GlobalPropBase:     gp.GlobalPropBase,
		ActiveJuries:       ajs,
		ActiveMediators:    ams,
		PrecedingMediators: pms,
	}

	return gpt
}

func (gpt *globalProperty) getGP() *modules.GlobalProperty {
	ajs := make(map[common.Address]bool, 0)
	ams := make(map[common.Address]bool, 0)
	pms := make(map[common.Address]bool, 0)

	for _, addStr := range gpt.ActiveJuries {
		ajs[addStr] = true
	}

	for _, addStr := range gpt.ActiveMediators {
		ams[addStr] = true
	}

	for _, addStr := range gpt.PrecedingMediators {
		pms[addStr] = true
	}

	gp := modules.NewGlobalProp()
	gp.GlobalPropBase = gpt.GlobalPropBase
	gp.ActiveJuries = ajs
	gp.ActiveMediators = ams
	gp.PrecedingMediators = pms

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

	gp := gpt.getGP()

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
