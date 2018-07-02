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
 * @author PalletOne core developers <dev@pallet.one>
 * @date 2018
 * @brief mediator插件的初始化和启动功能。Implement the function of mediator plugin initialization and startup.
 */

package mediatorplugin

import (
	"fmt"
	"time"
	"strconv"

	"github.com/palletone/go-palletone/common/log"

	d "github.com/palletone/go-palletone/consensus/dpos"
	a "github.com/palletone/go-palletone/core/application"
	s "github.com/palletone/go-palletone/consensus/dpos/mediators"
	v "github.com/palletone/go-palletone/dag/verifyunit"
)

var (
	Signature1 = "mediator1`sig"
	Signature2 = "mediator2`sig"
	Signature3 = "mediator3`sig"
)

var (
	Mediator1 = d.Mediator{"mediator1", Signature1}
	Mediator2 = d.Mediator{"mediator2", Signature2}
	Mediator3 = d.Mediator{"mediator3", Signature3}
)

type MediatorPlugin struct {
	// Enable VerifiedUnit production, even if the chain is stale. 新开启一个区块链时，必须设为true
	ProductionEnabled            bool
	// Percent of mediators (0-99) that must be participating in order to produce VerifiedUnit.
	// 新开启一个区块链时，必须设为0, 或者100
//	RequiredMediatorsParticipation float32
	MediatorSet map[d.Mediator]bool
	SignKeySet  map[string]bool
	DB          *a.DataBase
	ch          chan int
}

func (mp *MediatorPlugin) CloseChannel(signal int) {
	mp.ch <- signal
	close(mp.ch)
}

func (mp *MediatorPlugin) PluginInitialize() {
//	println("mediator plugin initialize begin")
	log.Info("mediator plugin initialize begin")

	// 1.初始化生产验证单元相关的属性值
	mp.ProductionEnabled = false
//	mp.RequiredMediatorsParticipation = 0.33

	// 1. 获取当前节点控制的所有mediator
	mp.MediatorSet = map[d.Mediator]bool{}
	mp.MediatorSet[Mediator1] = true
	mp.MediatorSet[Mediator2] = true
	mp.MediatorSet[Mediator3] = true

//	fmt.Printf("this node controll %d mediators!\n", len(mp.MediatorSet))
	log.Info(fmt.Sprintf("this node controll %d mediators!", len(mp.MediatorSet)))

	// 2. 获取当前节点使用的mediator使用的所有签名公私钥
	mp.SignKeySet = map[string]bool{}
	mp.SignKeySet[Signature1] = true
	mp.SignKeySet[Signature2] = true
	mp.SignKeySet[Signature3] = true

//	fmt.Printf("this node controll %d private keys!\n", len(mp.SignKeySet))
	log.Info(fmt.Sprintf("this node controll %d private keys!", len(mp.SignKeySet)))

//	println("mediator plugin initialize end\n")
	log.Info("mediator plugin initialize end")
}

func (mp *MediatorPlugin) PluginStartup(db *a.DataBase, ch chan int) {
//	println("\nmediator plugin startup begin")
	log.Info("mediator plugin startup begin")
	mp.DB = db
	mp.ch = ch

	// 1. 判断是否满足生产验证单元的条件，主要判断本节点是否控制至少一个mediator账户
	if len(mp.MediatorSet) == 0 {
		println("No mediators configured! Please add mediator and private keys to configuration.")
		mp.CloseChannel(-1)
	} else {
		// 2. 开启循环生产计划
//		fmt.Printf("Launching unit verify for %d mediators.\n", len(mp.MediatorSet))
		log.Info(fmt.Sprintf("Launching unit verify for %d mediators.", len(mp.MediatorSet)))
		mp.ProductionEnabled = true		// 此处应由配置文件中的 enable-stale-production 字段设置为true

		if mp.ProductionEnabled {
			if mp.DB.DynGlobalProp.LastVerifiedUnitNum == 0 {
				println()
				println("*   ------- NEW CHAIN -------   *")
				println("*   - Welcome to PalletOne! -   *")
				println("*   -------------------------   *")
				println()
			}
		}

		mp.ScheduleProductionLoop()
	}

//	println("mediator plugin startup end!")
	log.Info("mediator plugin startup end!")
}

func (mp *MediatorPlugin) ScheduleProductionLoop() {
	// 1. 计算下一秒的滴答时刻，如果少于50毫秒，则多等一秒开始
	now := time.Now()
	timeToNextSecond := time.Second - time.Duration(now.Nanosecond())
	if timeToNextSecond < 50 * time.Millisecond {
		timeToNextSecond += time.Second
	}

	nextWakeup := now.Add(timeToNextSecond)

	// 2. 安排验证单元生产循环
	go mp.VerifiedUnitProductionLoop(nextWakeup)
}

//验证单元生产状态类型
type ProductionCondition uint8

//验证单元生产状态枚举
const(
	Produced ProductionCondition = iota	// 正常生产验证单元
	NotSynced
	NotMyTurn
	NotTimeYet
	NoPrivateKey
//	LowParticipation
	Lag
//	Consecutive
//	ExceptionProducing
)

func (mp *MediatorPlugin) VerifiedUnitProductionLoop(wakeup time.Time) ProductionCondition {
	time.Sleep(wakeup.Sub(time.Now()))

	// 1. 尝试生产验证单元
	result, detail := mp.MaybeProduceVerifiedUnit()

	// 2. 打印尝试结果
	switch result {
	case Produced:
		//println("Generated VerifiedUnit #" + detail["Num"] +" with timestamp " +
		//	detail["Timestamp"] + /*" at time " + detail["Now"]*/ " with signature " + detail["MediatorSig"])
		log.Info("Generated VerifiedUnit #" + detail["Num"] +" with timestamp " +
			detail["Timestamp"] + /*" at time " + detail["Now"]*/ " with signature " + detail["MediatorSig"])
	case NotSynced:
		//println("Not producing VerifiedUnit because production is disabled " +
		//	"until we receive a recent VerifiedUnit (see: --enable-stale-production)")
		log.Info("Not producing VerifiedUnit because production is disabled " +
			"until we receive a recent VerifiedUnit (see: --enable-stale-production)")
	case NotTimeYet:
		//fmt.Printf("Not producing VerifiedUnit because next slot time is %v, but now is %v\n",
		//	detail["NextTime"], detail["Now"])
	case NotMyTurn:
		//fmt.Printf("Not producing VerifiedUnit because current scheduled mediator is %v\n",
		//	detail["ScheduledMediator"])
	case Lag:
		//fmt.Printf("Not producing VerifiedUnit because node didn't wake up within 500ms of the slot time." +
		//	" Scheduled Time is: %v, but now is %v\n", detail["ScheduledTime"], detail["Now"])
		log.Info("Not producing VerifiedUnit because node didn't wake up within 500ms of the slot time." +
			" Scheduled Time is: %v, but now is %v\n", detail["ScheduledTime"], detail["Now"])
	case NoPrivateKey:
		//fmt.Printf("Not producing VerifiedUnit because I don't have the private key for %v\n",
		//	detail["ScheduledKey"])
		log.Info("Not producing VerifiedUnit because I don't have the private key for %v\n",
			detail["ScheduledKey"])
	default:
		//println("Unknown condition!")
		log.Info("Unknown condition!")
	}

	// 3. 继续循环生产计划
	mp.ScheduleProductionLoop()

	return result
}

func (mp *MediatorPlugin) MaybeProduceVerifiedUnit() (ProductionCondition, map[string]string) {
//	println("\n尝试生产验证单元...")
	detail := map[string]string{}

	gp := &mp.DB.GlobalProp
	dgp := &mp.DB.DynGlobalProp
	ms := &mp.DB.MediatorSchl

	// 整秒调整，四舍五入
	nowFine := time.Now()
	now := time.Unix(nowFine.Add( 500 * time.Millisecond).Unix(), 0)

	// 1. 判断是否满足生产的各个条件
	nextSlotTime := s.GetSlotTime(gp, dgp, 1)
//	println(nextSlotTime.String())
	// If the next VerifiedUnit production opportunity is in the present or future, we're synced.
	if !mp.ProductionEnabled {
		if nextSlotTime.After(now) || nextSlotTime.Equal(now) {
			mp.ProductionEnabled = true
		}else {
			return NotSynced, detail
		}
	}

	slot := s.GetSlotAtTime(gp, dgp, now)
//	println(slot)
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
	if !mp.MediatorSet[*scheduledMediator] {
		detail["ScheduledMediator"] = scheduledMediator.Name
		return NotMyTurn, detail
	}

	scheduledTime := s.GetSlotTime(gp, dgp, slot)
	diff := scheduledTime.Sub(now)
	if diff > 500 * time.Millisecond || diff < -500 * time.Millisecond {
		detail["ScheduledTime"] = scheduledTime.Format("2006-01-02 15:04:05")
		detail["Now"] = now.Format("2006-01-02 15:04:05")
		return Lag, detail
	}

	// 此处应该判断scheduledMediator的签名公钥对应的私钥在本节点是否存在
	signKey := scheduledMediator.SignKey
	if !mp.SignKeySet[signKey] {
		detail["ScheduledKey"] = signKey
		return NoPrivateKey, detail
	}

	// 2. 生产验证单元
	verifiedUnit := v.GenerateVerifiedUnit(
		scheduledTime,
//		scheduledMediator,
		signKey,
		mp.DB)

	// 3. 异步向区块链网络广播验证单元
//	go println("异步向网络广播新生产的验证单元...")
//	go println("Asynchronously broadcast the new signed verified unit to p2p networks...")
	go log.Info("Asynchronously broadcast the new signed verified unit to p2p networks...")

	detail["Num"] = strconv.FormatUint(uint64(verifiedUnit.VerifiedUnitNum), 10)
	detail["Timestamp"] = verifiedUnit.Timestamp.Format("2006-01-02 15:04:05")
//	detail["Now"] = now.Format("2006-01-02 15:04:05")
	detail["MediatorSig"] = verifiedUnit.MediatorSig
	return Produced, detail
}
