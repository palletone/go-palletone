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
	key := append(constants.MEDIATOR_INFO_PREFIX, address.Bytes21()...)
	//key := append(constants.MEDIATOR_INFO_PREFIX, address.Str()...)

	return key
}

// only for serialization(storage)
type MediatorInfo struct {
	//AddStr               string
	InitPubKey           string
	Node                 string
	Url                  string
	TotalMissed          uint64
	LastConfirmedUnitNum uint32
	TotalVotes           uint64
}

func NewMediatorInfo() *MediatorInfo {
	return &MediatorInfo{
		Url:                  "",
		TotalMissed:          0,
		LastConfirmedUnitNum: 0,
		TotalVotes:           0,
	}
}

func mediatorToInfo(md *core.Mediator) *MediatorInfo {
	mi := NewMediatorInfo()
	//mi.AddStr = md.Address.Str()
	mi.InitPubKey = core.PointToStr(md.InitPubKey)
	mi.Node = md.Node.String()
	mi.TotalMissed = md.TotalMissed
	mi.LastConfirmedUnitNum = md.LastConfirmedUnitNum
	mi.TotalVotes = md.TotalVotes

	return mi
}

func (mi *MediatorInfo) infoToMediator() *core.Mediator {
	md := core.NewMediator()
	//md.Address = core.StrToMedAdd(mi.AddStr)
	md.InitPubKey, _ = core.StrToPoint(mi.InitPubKey)
	md.Node = core.StrToMedNode(mi.Node)
	md.TotalMissed = mi.TotalMissed
	md.LastConfirmedUnitNum = mi.LastConfirmedUnitNum
	md.TotalVotes = mi.TotalVotes

	return md
}

func StoreMediator(db ptndb.Database, med *core.Mediator) error {
	mi := mediatorToInfo(med)

	return StoreMediatorInfo(db, med.Address, mi)
}

func StoreMediatorInfo(db ptndb.Database, add common.Address, mi *MediatorInfo) error {
	//log.Debug(fmt.Sprintf("Store Mediator %v:", mi.AddStr))
	//add := core.StrToMedAdd(mi.AddStr)

	err := StoreBytes(db, mediatorKey(add), mi)
	if err != nil {
		log.Error(fmt.Sprintf("Store mediator error:%s", err))
		return err
	}

	return nil
}

func RetrieveMediatorInfo(db ptndb.Database, address common.Address) (*MediatorInfo, error) {
	mi := NewMediatorInfo()

	err := retrieve(db, mediatorKey(address), mi)
	if err != nil {
		log.Error(fmt.Sprintf("Retrieve mediator error: %s", err))
		return nil, err
	}

	return mi, nil
}

func RetrieveMediator(db ptndb.Database, address common.Address) (*core.Mediator, error) {
	mi, err := RetrieveMediatorInfo(db, address)
	if mi == nil || err != nil {
		return nil, err
	}

	med := mi.infoToMediator()
	med.Address = address

	return med, nil
}

func GetMediatorCount(db ptndb.Database) int {
	mc := getCountByPrefix(db, constants.MEDIATOR_INFO_PREFIX)

	return mc
}

func IsMediator(db ptndb.Database, address common.Address) bool {
	has, err := db.Has(mediatorKey(address))
	if err != nil {
		log.Debug(fmt.Sprintf("Error in determining if it is a mediator: %s", err))
	}

	return has
}

func GetMediators(db ptndb.Database) map[common.Address]bool {
	result := make(map[common.Address]bool)

	iter := db.NewIteratorWithPrefix(constants.MEDIATOR_INFO_PREFIX)
	for iter.Next() {
		key := iter.Key()
		if key == nil {
			continue
		}

		//log.Debug(fmt.Sprintf("Get Mediator's key : %s", key))
		addB := bytes.TrimPrefix(key, constants.MEDIATOR_INFO_PREFIX)

		result[common.BytesToAddress(addB)] = true
		//result[core.StrToMedAdd(string(addStr))] = true
	}

	return result
}

func LookupMediator(db ptndb.Database) map[common.Address]*core.Mediator {
	result := make(map[common.Address]*core.Mediator)

	iter := db.NewIteratorWithPrefix(constants.MEDIATOR_INFO_PREFIX)
	for iter.Next() {
		key := iter.Key()
		if key == nil {
			continue
		}

		value := iter.Value()
		if value == nil {
			continue
		}

		mi := NewMediatorInfo()
		err := rlp.DecodeBytes(value, mi)
		if err != nil {
			log.Debug(fmt.Sprintf("Error in Decoding Bytes to MediatorInfo: %s", err))
		}

		addB := bytes.TrimPrefix(key, constants.MEDIATOR_INFO_PREFIX)
		add := common.BytesToAddress(addB)
		med := mi.infoToMediator()
		med.Address = add

		result[add] = med
	}

	return result
}
