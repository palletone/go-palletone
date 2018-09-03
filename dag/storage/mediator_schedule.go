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

	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/modules"
)

const (
	mediatorSchlDBKey = "MediatorSchedule"
)

type mediatorSchedule struct {
	CurrentShuffledMediators []core.MediatorInfo
}

func getMST(ms *modules.MediatorSchedule) mediatorSchedule {
	csm := make([]core.MediatorInfo, 0)

	for _, med := range ms.CurrentShuffledMediators {
		medInfo := core.MediatorToInfo(&med)
		csm = append(csm, medInfo)
	}

	mst := mediatorSchedule{
		CurrentShuffledMediators: csm,
	}

	return mst
}

func getMS(mst *mediatorSchedule) *modules.MediatorSchedule {
	csm := make([]core.Mediator, 0)

	for _, medInfo := range mst.CurrentShuffledMediators {
		med := core.InfoToMediator(&medInfo)
		csm = append(csm, med)
	}

	ms := modules.NewMediatorSchl()
	ms.CurrentShuffledMediators = csm

	return ms
}

func StoreMediatorSchl(ms *modules.MediatorSchedule) {
	if Dbconn == nil {
		Dbconn = ReNewDbConn(dagconfig.DbPath)
	}

	mst := getMST(ms)

	err := Store(Dbconn, mediatorSchlDBKey, mst)
	if err != nil {
		log.Error(fmt.Sprintf("Store mediator schedule error: %s", err))
	}
}

func RetrieveMediatorSchl() *modules.MediatorSchedule {
	mst := new(mediatorSchedule)

	err := Retrieve(mediatorSchlDBKey, mst)
	if err != nil {
		log.Error(fmt.Sprintf("Retrieve mediator schedule error: %s", err))
	}

	ms := getMS(mst)

	return ms
}
