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
)

const (
	globalPropDBKey    = "GlobalProperty"
	dynGlobalPropDBKey = "DynamicGlobalProperty"
)

type PropertyDatabase struct {
	db            ptndb.Database
	GlobalProp    *modules.GlobalProperty
	DynGlobalProp *modules.DynamicGlobalProperty
	MediatorSchl  *modules.MediatorSchedule
}
type PropertyDb interface {
	StoreGlobalProp(gp *modules.GlobalProperty) error
	StoreDynGlobalProp(dgp *modules.DynamicGlobalProperty) error
	RetrieveGlobalProp() (*modules.GlobalProperty, error)
	RetrieveDynGlobalProp() (*modules.DynamicGlobalProperty, error)
	StoreMediatorSchl(ms *modules.MediatorSchedule) error
	RetrieveMediatorSchl() (*modules.MediatorSchedule, error)
	GetGlobalProp() *modules.GlobalProperty
	GetDynGlobalProp() *modules.DynamicGlobalProperty
	GetMediatorSchl() *modules.MediatorSchedule

}

func NewPropertyDb(db ptndb.Database) *PropertyDatabase {
	pdb := &PropertyDatabase{db: db}
	gp, err := pdb.RetrieveGlobalProp()
	if err != nil {
		//log.Error(err.Error())
		//return nil, err
	}

	dgp, err := pdb.RetrieveDynGlobalProp()
	if err != nil {
		//log.Error(err.Error())
		//return nil, err
	}

	ms, err := pdb.RetrieveMediatorSchl()
	if err != nil {
		//log.Error(err.Error())
		//return nil, err
	}
	pdb.GlobalProp = gp
	pdb.DynGlobalProp = dgp
	pdb.MediatorSchl = ms
	return pdb
}
func (propdb *PropertyDatabase) GetGlobalProp() *modules.GlobalProperty {
	return propdb.GlobalProp
}
func (propdb *PropertyDatabase) GetDynGlobalProp() *modules.DynamicGlobalProperty {
	return propdb.DynGlobalProp
}
func (propdb *PropertyDatabase) GetMediatorSchl() *modules.MediatorSchedule {
	return propdb.MediatorSchl
}

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

func (propdb *PropertyDatabase) StoreGlobalProp(gp *modules.GlobalProperty) error {

	gpt := getGPT(gp)

	err := Store(propdb.db, globalPropDBKey, gpt)

	if err != nil {
		log.Error(fmt.Sprintf("Store global properties error:%s", err))
	}

	return err
}

func (propdb *PropertyDatabase) StoreDynGlobalProp(dgp *modules.DynamicGlobalProperty) error {

	err := Store(propdb.db, dynGlobalPropDBKey, *dgp)
	if err != nil {
		//log.Error(fmt.Sprintf("Store dynamic global properties error: %s", err))
	}

	return err
}

func (propdb *PropertyDatabase) RetrieveGlobalProp() (*modules.GlobalProperty, error) {
	gpt := new(globalProperty)

	err := Retrieve(propdb.db, globalPropDBKey, gpt)
	if err != nil {
		//log.Error(fmt.Sprintf("Retrieve global properties error: %s", err))
	}

	gp := getGP(gpt)

	return gp, err
}

func (propdb *PropertyDatabase) RetrieveDynGlobalProp() (*modules.DynamicGlobalProperty, error) {
	dgp := modules.NewDynGlobalProp()

	err := Retrieve(propdb.db, dynGlobalPropDBKey, dgp)
	if err != nil {
		//log.Error(fmt.Sprintf("Retrieve dynamic global properties error: %s", err))
	}

	return dgp, err
}
