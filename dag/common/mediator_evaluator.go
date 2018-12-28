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

type MediatorCreateEvaluator struct {
}

func (mce *MediatorCreateEvaluator) Evaluate() bool {
	return true
}

func (mce *MediatorCreateEvaluator) Apply(statedb storage.IStateDb, mco *modules.MediatorCreateOperation) {
	mi := storage.NewMediatorInfo()
	//mi.AddStr = mco.AddStr
	mi.InitPubKey = mco.InitPubKey
	mi.Node = mco.Node
	mi.Url = mco.Url

	statedb.StoreMediatorInfo(core.StrToMedAdd(mco.AddStr), mi)
	return
}

// Create initial mediators
func GetInitialMediatorMsgs(genesisConf *core.Genesis) []*modules.Message {
	result := make([]*modules.Message, 0)

	for _, mi := range genesisConf.InitialMediatorCandidates {
		mco := &modules.MediatorCreateOperation{
			MediatorInfoBase: mi.MediatorInfoBase,
			Url:              "",
		}

		msg := &modules.Message{
			App:     modules.OP_MEDIATOR_CREATE,
			Payload: mco,
		}

		result = append(result, msg)
	}

	return result
}

func (unitOp *UnitRepository) ApplyOperation(msg *modules.Message, apply bool) bool {
	var payload interface{}
	payload = msg.Payload
	mediatorCreateOp, ok := payload.(*modules.MediatorCreateOperation)
	if ok == false {
		log.Error("a invalid Mediator Create Operation!")
		return false
	}

	if !mediatorCreateOp.Validate() {
		log.Error("Mediator Create Operation Validate does not pass!")
		return false
	}

	var mce MediatorCreateEvaluator
	result := mce.Evaluate()

	if apply {
		mce.Apply(unitOp.statedb, mediatorCreateOp)
	}

	return result
}
