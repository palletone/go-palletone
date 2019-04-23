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
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/dag/constants"
	"github.com/palletone/go-palletone/dag/modules"
)

var (
	MediatorSchlDBKey = append(constants.MEDIATOR_SCHEME_PREFIX, []byte("MediatorSchedule")...)
)

func StoreMediatorSchl(db ptndb.Database, ms *modules.MediatorSchedule) error {
	err := StoreBytes(db, MediatorSchlDBKey, ms)
	if err != nil {
		log.Errorf("Store mediator schedule error: %v", err.Error())
	}

	return err
}

func RetrieveMediatorSchl(db ptndb.Database) (*modules.MediatorSchedule, error) {
	ms := new(modules.MediatorSchedule)
	err := retrieve(db, MediatorSchlDBKey, ms)
	if err != nil {
		log.Errorf("Retrieve mediator schedule error: %v", err.Error())
	}

	return ms, err
}
