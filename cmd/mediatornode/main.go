package main

import (
	"fmt"

	m "github.com/palletone/go-palletone/core/mediatorplugin"
	a "github.com/palletone/go-palletone/core/application"
)

func main() {
	fmt.Println("mediator node starting...")

	// 1. 拆解程序命令行参数和值
	println("拆解程序命令行参数和值...")

	// 2. 根据命令行参数指定的配置文件路径，读取配置文件参数
	println("根据命令行参数指定的配置文件路径，读取配置文件参数...")

	// 3. 注册所需的模块组件
	println("注册所需的模块组件...")
	var mp m.MediatorPlugin

	// 4. 根据命令行参数初始主程序
	println("根据命令行参数初始主程序...")
	var db a.DataBase
	db.Initialize()

	// 5. 根据命令行参数初始程序组件
	println("根据命令行参数初始程序组件...")
	mp.PluginInitialize()

	// 6. 启动主程序
	println("启动主程序...")
	db.Startup()

	// 7. 启动程序组件
	println("启动程序组件...")
}
