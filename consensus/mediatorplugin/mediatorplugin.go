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

package mediatorplugin

import (
	"fmt"
	"strconv"
	"time"
	"errors"

	"github.com/palletone/go-palletone/common/event"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/core/accounts"
	"github.com/palletone/go-palletone/dag"
	"github.com/palletone/go-palletone/dag/modules"
)

var (
	errTerminated = errors.New("terminated")
)

func newChainBanner(dag *dag.Dag) {
	fmt.Printf("\n" +
		"*   ------- NEW CHAIN -------   *\n" +
		"*   - Welcome to PalletOne! -   *\n" +
		"*   -------------------------   *\n" +
		"\n")

	if modules.GetSlotAtTime(dag.GlobalProp, dag.DynGlobalProp, time.Now()) > 200 {
		fmt.Printf("Your genesis seems to have an old timestamp\n" +
			"Please consider using the --genesis-timestamp option to give your genesis a recent timestamp\n" +
			"\n")
	}
}

func (mp *MediatorPlugin) SubscribeNewProducedUnitEvent(ch chan<- NewProducedUnitEvent) event.Subscription {
	return mp.newProducedUnitScope.Track(mp.newProducedUnitFeed.Subscribe(ch))
}

func (mp *MediatorPlugin) ScheduleProductionLoop() {
	// 1. 计算下一秒的滴答时刻，如果少于50毫秒，则多等一秒开始
	now := time.Now()
	timeToNextSecond := time.Second - time.Duration(now.Nanosecond())
	if timeToNextSecond < 50*time.Millisecond {
		timeToNextSecond += time.Second
	}

	nextWakeup := now.Add(timeToNextSecond)

	// 2. 安排验证单元生产循环
	// production unit until termination is requested
	select {
	case <-mp.quit:
		return
	default:
		go mp.VerifiedUnitProductionLoop(nextWakeup)
	}
}

//验证单元生产状态类型
type ProductionCondition uint8

//验证单元生产状态枚举
const (
	Produced ProductionCondition = iota // 正常生产验证单元
	NotSynced
	NotMyTurn
	NotTimeYet
	NoPrivateKey
	//	LowParticipation
	Lag
	//	Consecutive
	ExceptionProducing
)

func (mp *MediatorPlugin) VerifiedUnitProductionLoop(wakeup time.Time) ProductionCondition {
	// Start to production unit for expiration
	time.Sleep(wakeup.Sub(time.Now()))

	// 1. 尝试生产验证单元
	result, detail := mp.MaybeProduceVerifiedUnit()

	// 2. 打印尝试结果
	switch result {
	case Produced:
		log.Info("Generated VerifiedUnit #" + detail["Num"] + " hash: " + detail["Hash"] +
			" with timestamp " + detail["Timestamp"] + " by mediator: " + detail["Mediator"])
	case NotSynced:
		log.Info("Not producing VerifiedUnit because production is disabled " +
			"until we receive a recent VerifiedUnit (see: --enable-stale-production)")
	case NotTimeYet:
		log.Debug("Not producing VerifiedUnit because next slot time is " + detail["NextTime"] +
			" , but now is " + detail["Now"])
	case NotMyTurn:
		log.Debug("Not producing VerifiedUnit because current scheduled mediator is " +
			detail["ScheduledMediator"])
	case Lag:
		log.Info("Not producing VerifiedUnit because node didn't wake up within 500ms of the slot time." +
			" Scheduled Time is: " + detail["ScheduledTime"] + ", but now is " + detail["Now"])
	case NoPrivateKey:
		log.Info("Not producing VerifiedUnit because I don't have the private key for " +
			detail["ScheduledKey"])
	case ExceptionProducing:
		log.Info("Exception producing unit")
	default:
		log.Info("Unknown condition!")
	}

	// 3. 继续循环生产计划
	mp.ScheduleProductionLoop()

	return result
}

func (mp *MediatorPlugin) MaybeProduceVerifiedUnit() (ProductionCondition, map[string]string) {
	//	println("\n尝试生产验证单元...")
	detail := map[string]string{}

	dag := mp.ptn.Dag()
	gp := dag.GlobalProp
	dgp := dag.DynGlobalProp
	ms := dag.MediatorSchl

	// 整秒调整，四舍五入
	nowFine := time.Now()
	now := time.Unix(nowFine.Add(500*time.Millisecond).Unix(), 0)

	// 1. 判断是否满足生产的各个条件
	nextSlotTime := modules.GetSlotTime(gp, dgp, 1)
	// If the next VerifiedUnit production opportunity is in the present or future, we're synced.
	if !mp.productionEnabled {
		if nextSlotTime.After(now) || nextSlotTime.Equal(now) {
			mp.productionEnabled = true
		} else {
			return NotSynced, detail
		}
	}

	slot := modules.GetSlotAtTime(gp, dgp, now)
	// is anyone scheduled to produce now or one second in the future?
	if slot == 0 {
		detail["NextTime"] = nextSlotTime.Format("2006-01-02 15:04:05")
		detail["Now"] = now.Format("2006-01-02 15:04:05")
		return NotTimeYet, detail
	}

	//
	// this Conditional judgment should fail, because now <= LastVerifiedUnitTime
	// should have resulted in slot == 0.
	//
	// if this assert triggers, there is a serious bug in GetSlotAtTime()
	// which would result in allowing a later block to have a timestamp
	// less than or equal to the previous VerifiedUnit
	//
	//if !now.After(dgp.LastVerifiedUnitTime) {
	//	panic("\n The later VerifiedUnit have a timestamp less than or equal to the previous!")
	//}

	scheduledMediator := ms.GetScheduledMediator(dgp, slot)
	// we must control the Mediator scheduled to produce the next VerifiedUnit.
	ma := scheduledMediator.Address
	ps, ok := mp.mediators[ma]
	if !ok {
		detail["ScheduledMediator"] = ma.Str()
		return NotMyTurn, detail
	}

	scheduledTime := modules.GetSlotTime(gp, dgp, slot)
	diff := scheduledTime.Sub(now)
	if diff > 500*time.Millisecond || diff < -500*time.Millisecond {
		detail["ScheduledTime"] = scheduledTime.Format("2006-01-02 15:04:05")
		detail["Now"] = now.Format("2006-01-02 15:04:05")
		return Lag, detail
	}

	// 此处应该判断scheduledMediator的签名公钥对应的私钥在本节点是否存在
	ks := mp.ptn.GetKeyStore()
	err := ks.Unlock(accounts.Account{Address: ma}, ps)
	if err != nil {
		detail["ScheduledKey"] = ma.Str()
		return NoPrivateKey, detail
	}

	// 2. 生产验证单元
	unit := GenerateUnit(mp.ptn.Dag(), scheduledTime, *scheduledMediator, ks, mp.ptn.TxPool())
	if unit.IsEmpty() {
		return ExceptionProducing, detail
	}

	num := unit.UnitHeader.Number.Index
	detail["Num"] = strconv.FormatUint(num, 10)
	time := time.Unix(unit.UnitHeader.Creationdate, 0)
	detail["Timestamp"] = time.Format("2006-01-02 15:04:05")
	detail["Mediator"] = unit.UnitHeader.Authors.Address
	detail["Hash"] = unit.UnitHash.Hex()

	// 3. 异步向区块链网络广播验证单元
	log.Debug("Asynchronously broadcast the new signed verified unit to p2p networks...")
	go mp.newProducedUnitFeed.Send(NewProducedUnitEvent{Unit: unit})

	return Produced, detail
}

func (mp *MediatorPlugin) AddPendingBLSSigned(peer string, unit *modules.Unit) error {
	op := &toBLSSigned{
		origin: peer,
		unit:  unit,
	}

	select {
	case <-mp.quit:
		return errTerminated
	case mp.toBLSSigned <- op:
		return nil
	}
}

func (mp *MediatorPlugin) recPendingBLSSignedLoop() {
	for {
		select {
		// Mediator Plugin terminating, abort operation
		case <-mp.quit:
			return
		case op := <- mp.toBLSSigned:
//			mp.pendingBLS[op.unit.UnitHash] = op
			PushUnit(mp.ptn.Dag(), op.unit)
			go mp.unitBLSSign(op)
		}
	}
}

func (mp *MediatorPlugin) unitBLSSign(toBLSSigned *toBLSSigned)  {
	//todo
	
}
