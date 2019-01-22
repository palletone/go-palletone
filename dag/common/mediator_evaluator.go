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
)

func MediatorCreateEvaluate(op *modules.MediatorCreateOperation) bool {
	// todo 判断是否已经申请缴纳保证金
	return true
}

// Create initial mediators
func GetInitialMediatorMsgs(genesisConf *core.Genesis) []*modules.Message {
	result := make([]*modules.Message, 0)

	for _, mi := range genesisConf.InitialMediatorCandidates {
		mco := &modules.MediatorCreateOperation{
			MediatorInfoBase: mi.MediatorInfoBase,
			Url:              "",
		}

		err := mco.Validate()
		if err != nil {
			panic(err.Error())
		}

		msg := &modules.Message{
			App:     modules.OP_MEDIATOR_CREATE,
			Payload: mco,
		}

		result = append(result, msg)
	}

	return result
}

func (unitOp *UnitRepository) MediatorCreateApply(msg *modules.Message) bool {
	var payload interface{}
	payload = msg.Payload
	mco, ok := payload.(*modules.MediatorCreateOperation)
	if !ok {
		log.Debug("a invalid Mediator Create Operation!")
		return false
	}

	mi := modules.NewMediatorInfo()
	mi.MediatorInfoBase = mco.MediatorInfoBase
	mi.Url = mco.Url

	addr, _ := core.StrToMedAdd(mco.AddStr)
	unitOp.statedb.StoreMediatorInfo(addr, mi)

	return true
}
