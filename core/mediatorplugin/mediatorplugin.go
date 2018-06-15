package mediatorplugin

import (
	"fmt"
	"time"

	d "github.com/palletone/go-palletone/consensus/dpos"
	a "github.com/palletone/go-palletone/core/application"
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
	ProductionEnabled            bool
	RequiredWitnessParticipation float32
	CMediaotors                  []*d.Mediator
	PrivateKeys                  []*string
	DB                           *a.DataBase
	ch                           chan int
}

func (mp *MediatorPlugin) CloseChannel(signal int) {
	mp.ch <- signal
	close(mp.ch)
}

func (mp *MediatorPlugin) PluginInitialize() {
	println("\nmediator plugin initialize begin")

	// 1.初始化生产验证单元相关的属性值
	mp.ProductionEnabled = false
	mp.RequiredWitnessParticipation = 0.33

	// 1. 获取当前节点控制的所有mediator
	mp.CMediaotors = append(mp.CMediaotors, &Mediator1)
	mp.CMediaotors = append(mp.CMediaotors, &Mediator2)
	mp.CMediaotors = append(mp.CMediaotors, &Mediator3)

	fmt.Printf("this node controll %d mediators!\n", len(mp.CMediaotors))

	// 2. 获取当前节点使用的mediator使用的所有签名公私钥
	mp.PrivateKeys = append(mp.PrivateKeys, &Signature1)
	mp.PrivateKeys = append(mp.PrivateKeys, &Signature2)
	mp.PrivateKeys = append(mp.PrivateKeys, &Signature3)

	fmt.Printf("this node controll %d private keys!\n", len(mp.PrivateKeys))

	println("mediator plugin initialize end\n")
}

func (mp *MediatorPlugin) PluginStartup(db *a.DataBase, ch chan int) {
	println("mediator plugin startup begin")
	mp.DB = db
	mp.ch = ch

	// 1. 判断是否满足生产验证单元的条件，主要判断本节点是否控制至少一个mediator账户
	if len(mp.CMediaotors) == 0 {
		println("No mediaotors configured! Please add mediator and private keys to configuration.")
		mp.CloseChannel(-1)
	} else {
		// 2. 开启循环生产计划
		fmt.Printf("Launching unit verify for %d mediators.\n", len(mp.CMediaotors))
		mp.ProductionEnabled = true

		if mp.ProductionEnabled {
			mp.ScheduleProductionLoop()
		}
	}

	println("mediator plugin startup end")
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
	time.Sleep(nextWakeup.Sub(time.Now()))
	mp.VerifiedUnitProductionLoop()
}

func (mp *MediatorPlugin) VerifiedUnitProductionLoop() {
	// 1. 尝试生产验证单元
	println("尝试生产验证单元")

	// 2. 打印尝试结果

	// 3. 继续循环生产计划
	mp.ScheduleProductionLoop()
}

func (mp *MediatorPlugin) MaybeProduceVerifiedUnit() {
	// 1. 判断是否满足生产的各个条件

	// 2. 生产验证单元

	// 3. 向区块链网络广播验证单元

}
