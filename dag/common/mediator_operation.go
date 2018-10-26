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

package common

import (
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/storage"
)

type MediatorCreateOperation struct {
	*core.MediatorInfo
}

//func feePayer(tx *modules.Transaction) (common.Address, error) {
//	return getRequesterAddress(tx)
//}

func (mco *MediatorCreateOperation) Validate() bool {
	return true
}

//func (mco *MediatorCreateOperation) Evaluate() bool {
//	return true
//}

func (mco *MediatorCreateOperation) Apply(statedb storage.IStateDb) {
	statedb.StoreMediatorInfo(mco.MediatorInfo)
	return
}

// Create initial mediators
func GetInitialMediatorMsgs(genesisConf *core.Genesis) []*modules.Message {
	result := make([]*modules.Message, 0)

	for _, mi := range genesisConf.InitialMediatorCandidates {
		mco := &MediatorCreateOperation{
			MediatorInfo: mi,
		}

		msg := &modules.Message{
			App:     modules.OP_MEDIATOR_CREATE,
			Payload: mco,
		}

		result = append(result, msg)
	}

	return result
}

func (unitOp *UnitRepository) ApplyOperation(msg *modules.Message) bool {
	var payload interface{}
	payload = msg.Payload
	mediatorCreateOp, ok := payload.(*MediatorCreateOperation)
	if ok == false {
		log.Error("a invalid Mediator Create Operation!")
		return false
	}

	if !mediatorCreateOp.Validate() {
		log.Error("Mediator Create Operation Validate does not pass!")
		return false
	}

	mediatorCreateOp.Apply(unitOp.statedb)

	return true
}
