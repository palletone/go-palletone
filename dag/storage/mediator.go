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
	"bytes"
	"fmt"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/common/rlp"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/dag/constants"
)

func mediatorKey(address common.Address) []byte {
	key := append(constants.MEDIATOR_INFO_PREFIX, address.Bytes()...)
	return key
}

//func StoreMediator(db ptndb.Database, med *core.Mediator) error {
//	mi := med.MediatorToInfo()
//
//	err := StoreBytes(db, mediatorKey(med.Address), mi)
//	if err != nil {
//		log.Error(fmt.Sprintf("Store mediator error:%s", err))
//	}
//
//	return err
//}

func StoreMediatorInfo(db ptndb.Database, mi *core.MediatorInfo) error {
	mk := append(constants.MEDIATOR_INFO_PREFIX, []byte(mi.Address)...)

	err := StoreBytes(db, mk, mi)
	if err != nil {
		log.Error(fmt.Sprintf("Store mediator error:%s", err))
	}

	return err
}

func RetrieveMediator(db ptndb.Database, address common.Address) (*core.Mediator, error) {
	mi := new(core.MediatorInfo)

	err := retrieve(db, mediatorKey(address), mi)
	if err != nil {
		log.Error(fmt.Sprintf("Retrieve mediator error: %s", err))
		return nil, nil
	}

	med := mi.InfoToMediator()

	return &med, nil
}

func GetMediatorCount(db ptndb.Database) int {
	mc := getCountByPrefix(db, constants.MEDIATOR_INFO_PREFIX)

	return mc
}

func IsMediator(db ptndb.Database, address common.Address) bool {
	has, err := db.Has(mediatorKey(address))
	if err != nil {
		log.Error(fmt.Sprintf("Error in determining if it is a mediator: %s", err))
	}

	return has
}

func GetMediators(db ptndb.Database) map[common.Address]bool {
	result := make(map[common.Address]bool)

	iter := db.NewIteratorWithPrefix(constants.MEDIATOR_INFO_PREFIX)
	for iter.Next() {
		key := iter.Key()
		address := bytes.TrimPrefix(key, constants.MEDIATOR_INFO_PREFIX)

		result[common.BytesToAddress(address)] = true
	}

	return result
}

func LookupMediator(db ptndb.Database) map[common.Address]core.Mediator {
	result := make(map[common.Address]core.Mediator)

	iter := db.NewIteratorWithPrefix(constants.MEDIATOR_INFO_PREFIX)
	for iter.Next() {
		mi := new(core.MediatorInfo)
		err := rlp.DecodeBytes(iter.Value(), mi)
		if err != nil {
			log.Error(fmt.Sprintf("Error in Decoding Bytes to MediatorInfo: %s", err))
		}

		med := mi.InfoToMediator()
		result[med.Address] = med
	}

	return result
}
