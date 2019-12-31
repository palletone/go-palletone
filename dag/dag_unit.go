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

package dag

import (
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/rlp"
	dagcommon "github.com/palletone/go-palletone/dag/common"
	"time"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/core/accounts/keystore"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/txspool"
)

// GenerateUnit, generate unit
// @author Albert·Gou
func (dag *Dag) GenerateUnit(when time.Time, producer common.Address, groupPubKey []byte, ks *keystore.KeyStore,
	txpool txspool.ITxPool) (*modules.Unit, error) {
	t0 := time.Now()
	defer func(start time.Time) {
		log.Debugf("GenerateUnit cost time: %v", time.Since(start))
	}(t0)

	// 1. 判断是否满足生产的若干条件
	//log.Debugf("generate unit ...")

	// 2. 生产unit，添加交易集、时间戳、签名
	unsign_unit, err := dag.createUnit(producer, txpool, when)
	if err != nil {
		errStr := fmt.Sprintf("GenerateUnit error: %v", err.Error())
		log.Debug(errStr)
		return nil, fmt.Errorf(errStr)
	}
	if unsign_unit == nil || unsign_unit.IsEmpty() {
		errStr := fmt.Sprintf("No unit need to be packaged for now.")
		log.Debug(errStr)
		return nil, fmt.Errorf(errStr)
	}
	sign_unit, err := dagcommon.GetUnitWithSig(unsign_unit, ks, producer)
	sign_unit.UnitHeader.SetGroupPubkey(groupPubKey)

	sign_unit.Size()
	log.Debugf("Generate new unit index:[%d],hash:[%s],size:%s, parent unit[%s],txs[%d], spent time: %s ",
		sign_unit.NumberU64(), sign_unit.Hash().String(), sign_unit.Size().String(),
		sign_unit.UnitHeader.ParentHash()[0].String(), sign_unit.Txs.Len(), time.Since(t0).String())

	//3.将新单元添加到MemDag中
	a, b, c, d, e, err := dag.Memdag.AddUnit(sign_unit, txpool, true)
	if a != nil && err == nil {
		if dag.unstableUnitProduceRep != e {
			log.Debugf("send UnstableRepositoryUpdatedEvent")
			go dag.unstableRepositoryUpdatedFeed.Send(modules.UnstableRepositoryUpdatedEvent{})
		}

		dag.unstableUnitRep = a
		dag.unstableUtxoRep = b
		dag.unstableStateRep = c
		dag.unstablePropRep = d
		dag.unstableUnitProduceRep = e
	} else if err != nil {
		errStr := fmt.Sprintf("Memdag AddUnit[%s] error: %v", sign_unit.Hash().String(), err.Error())
		udata, _ := json.Marshal(sign_unit)
		rdata, _ := rlp.EncodeToBytes(sign_unit)
		log.Errorf("%s, Unit data:%s,Rlp:%x", errStr, string(udata), rdata)
		return nil, fmt.Errorf(errStr)
	}
	sign_unit.ReceivedAt = time.Now()

	//4.PostChainEvents
	//TODO add PostChainEvents
	go func() {
		var (
			events = make([]interface{}, 0, 2)
		)
		events = append(events, modules.ChainHeadEvent{Unit: sign_unit})
		events = append(events, modules.ChainEvent{Unit: sign_unit, Hash: sign_unit.Hash()})
		dag.PostChainEvents(events)
	}()

	return sign_unit, nil
}

// createUnit, create a unit when mediator being produced
//创建未签名的Unit
func (d *Dag) createUnit(mAddr common.Address, txpool txspool.ITxPool,
	when time.Time) (*modules.Unit, error) {

	med, err := d.unstableStateRep.RetrieveMediator(mAddr)
	if err != nil {
		return nil, err
	}

	//return d.unstableUnitRep.CreateUnit(med.GetRewardAdd(), txpool, rep, state.GetJurorReward)
	return d.unstableUnitRep.CreateUnit(med.GetRewardAdd(), txpool, when,
		d.unstablePropRep, d.unstableStateRep.GetJurorReward)
}
