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

package storage

import (
	"fmt"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/dag/constants"
	"github.com/palletone/go-palletone/dag/modules"
)

var (
	MediatorSchlDBKey = append(constants.MEDIATOR_SCHEME_PREFIX, []byte("MediatorSchedule")...)
)

type mediatorSchedule struct {
	CurrentShuffledMediators []string
}

func getMST(ms *modules.MediatorSchedule) mediatorSchedule {
	addStrs := make([]string, 0)

	for _, medAdd := range ms.CurrentShuffledMediators {
		addStr := medAdd.Str()
		addStrs = append(addStrs, addStr)
	}

	mst := mediatorSchedule{
		CurrentShuffledMediators: addStrs,
	}

	return mst
}

func getMS(mst *mediatorSchedule) *modules.MediatorSchedule {
	medAdds := make([]common.Address, 0)

	for _, addStr := range mst.CurrentShuffledMediators {
		medAdd := core.StrToMedAdd(addStr)
		medAdds = append(medAdds, medAdd)
	}

	ms := modules.NewMediatorSchl()
	ms.CurrentShuffledMediators = medAdds

	return ms
}

func StoreMediatorSchl(db ptndb.Database, ms *modules.MediatorSchedule) error {
	mst := getMST(ms)

	err := StoreBytes(db, MediatorSchlDBKey, mst)
	if err != nil {
		log.Error(fmt.Sprintf("Store mediator schedule error: %s", err))
	}

	return err
}

func RetrieveMediatorSchl(db ptndb.Database) (*modules.MediatorSchedule, error) {
	mst := new(mediatorSchedule)

	err := retrieve(db, MediatorSchlDBKey, mst)
	if err != nil {
		log.Error(fmt.Sprintf("Retrieve mediator schedule error: %s", err))
	}

	ms := getMS(mst)

	return ms, err
}
