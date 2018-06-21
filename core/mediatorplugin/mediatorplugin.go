/**
@version 0.1
@author albert·gou
@time June 11, 2018
@brief mediator插件的初始化和启动功能
*/

package mediatorplugin

import (
	"fmt"
	"time"

	d "github.com/palletone/go-palletone/consensus/dpos"
	a "github.com/palletone/go-palletone/core/application"
	s "github.com/palletone/go-palletone/consensus/dpos/mediators"
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
	// Percent of witnesses (0-99) that must be participating in order to produce VerifiedUnit.
	// 新开启一个区块链时，必须设为0, 或者100
//	RequiredWitnessParticipation float32
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
	println("\nmediator plugin initialize begin")

	// 1.初始化生产验证单元相关的属性值
	mp.ProductionEnabled = false
//	mp.RequiredWitnessParticipation = 0.33

	// 1. 获取当前节点控制的所有mediator
	mp.MediatorSet = map[d.Mediator]bool{}
	mp.MediatorSet[Mediator1] = true
	mp.MediatorSet[Mediator2] = true
	mp.MediatorSet[Mediator3] = true

	fmt.Printf("this node controll %d mediators!\n", len(mp.MediatorSet))

	// 2. 获取当前节点使用的mediator使用的所有签名公私钥
	mp.SignKeySet = map[string]bool{}
	mp.SignKeySet[Signature1] = true
	mp.SignKeySet[Signature2] = true
	mp.SignKeySet[Signature3] = true

	fmt.Printf("this node controll %d private keys!\n", len(mp.SignKeySet))

	println("mediator plugin initialize end\n")
}

func (mp *MediatorPlugin) PluginStartup(db *a.DataBase, ch chan int) {
	println("\nmediator plugin startup begin")
	mp.DB = db
	mp.ch = ch

	// 1. 判断是否满足生产验证单元的条件，主要判断本节点是否控制至少一个mediator账户
	if len(mp.MediatorSet) == 0 {
		println("No mediaotors configured! Please add mediator and private keys to configuration.")
		mp.CloseChannel(-1)
	} else {
		// 2. 开启循环生产计划
		fmt.Printf("Launching unit verify for %d mediators.\n", len(mp.MediatorSet))
		mp.ProductionEnabled = true		// 此处应由配置文件设置为true

		if mp.ProductionEnabled {
			if mp.DB.DynGlobalProp.LastVerifiedUnitNum == 0 {
				println("*   ------- NEW CHAIN -------   *")
				println("*   - Welcome to PalletOne! -   *")
				println("*   -------------------------   *")
			}
		}

		mp.ScheduleProductionLoop()
	}

	println("mediator plugin startup end\n")
}

func (mp *MediatorPlugin) ScheduleProductionLoop() {
	// 1. 计算下一秒的滴答时刻，如果少于50毫秒，则多等一秒开始
	now := time.Now()
	timeToNextSecond := int64(now.Nanosecond())
	if timeToNextSecond < int64(time.Duration(50)*time.Millisecond) {
		timeToNextSecond += int64(time.Duration(1) * time.Second)
	}

	nextWakeup := now.Add(time.Duration(timeToNextSecond) * time.Nanosecond)

	// 2. 安排验证单元生产循环
	go mp.VerifiedUnitProductionLoop(nextWakeup)
}

//验证单元生产状态类型
type ProductionCondition uint8

//验证单元生产状态枚举
const(
	Produced ProductionCondition = iota
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
	println("尝试生产验证单元")
	result, detail := mp.MaybeProduceVerifiedUnit()

	// 2. 打印尝试结果
	switch result {
	case NoPrivateKey:
		fmt.Printf("Not producing block because I don't have the private key for %v\n", detail["ScheduledKey"])
	case Produced:
//		println()
	default:
		println("Unknow condition!")
	}

	// 3. 继续循环生产计划
	mp.ScheduleProductionLoop()

	return result
}

func (mp *MediatorPlugin) MaybeProduceVerifiedUnit() (ProductionCondition, map[string]string) {
	detail := map[string]string{}

	gp := &mp.DB.GlobalProp
	dgp := &mp.DB.DynGlobalProp
	ms := &mp.DB.MediatorSchl

	nowFine := time.Now()
	now := nowFine.Add( 500 * time.Millisecond)

	// 1. 判断是否满足生产的各个条件
	nextSlotTime := s.GetSlotTime(gp, dgp, 1)
	// If the next VerifiedUnit production opportunity is in the present or future, we're synced.
	if !mp.ProductionEnabled {
		if nextSlotTime.After(now) || nextSlotTime.Equal(now) {
			mp.ProductionEnabled = true
		}else {
			return NotSynced, detail
		}
	}

	slot := s.GetSlotAtTime(gp, dgp, now)
	// is anyone scheduled to produce now or one second in the future?
	if slot == 0 {
		detail["NextTime"] = s.GetSlotTime(gp, dgp, 1).Format("2006-01-02 15:04:05")
		return NotTimeYet,detail
	}

	//
	// this Conditional judgment should fail, because now <= LastVerifiedUnitTime
	// should have resulted in slot == 0.
	//
	// if this panic triggers, there is a serious bug in get_slot_at_time()
	// which would result in allowing a later block to have a timestamp
	// less than or equal to the previous block
	//
	if !now.After(dgp.LastVerifiedUnitTime) {
		panic("\n The later VerifiedUnit have a timestamp less than or equal to the previous!")
	}

	scheduledMediator := ms.GetScheduledMediator(dgp, slot)
	// we must control the Mediator scheduled to produce the next VerifiedUnit.
	if !mp.MediatorSet[*scheduledMediator] {
		detail["ScheduledMediator"] = scheduledMediator.Name
		return NotMyTurn, detail
	}

	scheduledTime := s.GetSlotTime(gp, dgp, slot)
	if scheduledTime.After(now.Add(500 * time.Millisecond)) || scheduledTime.Before(now.Add(-500 * time.Millisecond)) {
		detail["ScheduledTime"] = scheduledTime.Format("2018-06-20 20:50:34")
		detail["Now"] = now.Format("2018-06-20 20:50:34")
		return Lag, detail
	}

	// 此处应该判断scheduledMediator的签名公钥对应的私钥在本节点是否存在
	signKey := scheduledMediator.SignKey
	if !mp.SignKeySet[signKey] {
		detail["ScheduledKey"] = signKey
		return NoPrivateKey, detail
	}

	// 2. 生产验证单元
//	verifiedUnit :=

	// 3. 向区块链网络广播验证单元


	return Produced, detail
}
