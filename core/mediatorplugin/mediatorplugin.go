package mediatorplugin

func PluginInitialize() {
	println("mediator plugin initialize begin")

	// 1. 获取当前节点控制的所有mediator ID

	// 2. 获取当前节点使用的mediator使用的所有签名公私钥

	println("mediator plugin initialize end")
}

func PluginStartup() {
	println("mediator plugin startup begin")

	// 1. 判断是否满足生产验证单元的条件，主要判断本节点是否控制至少一个mediator账户

	// 2. 开启循环生产计划

	println("mediator plugin startup end")
}

func ScheduleProductionLoop() {
	// 1. 计算下一秒的滴答时刻，如果少于50毫秒，则从下下一秒开始

	// 2. 安排验证单元生产循环
}

func VerifiedUnitProductionLoop() {
	// 1. 尝试生产验证单元

	// 2. 打印尝试结果

	// 3. 继续循环生产计划
}

func MaybeProduceVerifiedUnit() {
	// 1. 判断是否满足生产的各个条件

	// 2. 生产验证单元

	// 3. 向区块链网络广播验证单元

}
