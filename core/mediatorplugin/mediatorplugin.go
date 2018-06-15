package mediatorplugin

import (
	"fmt"

	d "github.com/palletone/go-palletone/consensus/dpos"
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
	Mediaotrs []*d.Mediator
	PrivateKeys []*string
}

func (mp *MediatorPlugin) PluginInitialize() {
	println("\nmediator plugin initialize begin")

	// 1.初始化生产验证单元相关的属性值
	mp.ProductionEnabled = false
	mp.RequiredWitnessParticipation = 0.33

	// 1. 获取当前节点控制的所有mediator
	mp.Mediaotrs = append(mp.Mediaotrs, &Mediator1)
	mp.Mediaotrs = append(mp.Mediaotrs, &Mediator2)
	mp.Mediaotrs = append(mp.Mediaotrs, &Mediator3)

	fmt.Printf("this node controll %d mediators!\n", len(mp.Mediaotrs))

	// 2. 获取当前节点使用的mediator使用的所有签名公私钥
	mp.PrivateKeys = append(mp.PrivateKeys, &Signature1)
	mp.PrivateKeys = append(mp.PrivateKeys, &Signature2)
	mp.PrivateKeys = append(mp.PrivateKeys, &Signature3)

	fmt.Printf("this node controll %d sigkey!\n", len(mp.PrivateKeys))

	println("mediator plugin initialize end\n")
}

func (mp *MediatorPlugin) PluginStartup() {
	println("mediator plugin startup begin")

	// 1. 判断是否满足生产验证单元的条件，主要判断本节点是否控制至少一个mediator账户

	// 2. 开启循环生产计划

	println("mediator plugin startup end")
}

func (mp *MediatorPlugin) ScheduleProductionLoop() {
	// 1. 计算下一秒的滴答时刻，如果少于50毫秒，则从下下一秒开始

	// 2. 安排验证单元生产循环
}

func (mp *MediatorPlugin) VerifiedUnitProductionLoop() {
	// 1. 尝试生产验证单元

	// 2. 打印尝试结果

	// 3. 继续循环生产计划
}

func (mp *MediatorPlugin) MaybeProduceVerifiedUnit() {
	// 1. 判断是否满足生产的各个条件

	// 2. 生产验证单元

	// 3. 向区块链网络广播验证单元

}
