/**
@version 0.1
@author albert·gou
@time June 11, 2018
@brief mediator节点的初始化和启动功能
*/

package main

import (
	"fmt"

	a "github.com/palletone/go-palletone/core/application"
	m "github.com/palletone/go-palletone/core/mediatorplugin"
)

func main() {
	fmt.Println("mediator node starting...\n")

	// 1. 拆解程序命令行参数和值
	println("拆解程序命令行参数和值...\n")

	// 2. 根据命令行参数指定的配置文件路径，读取配置文件参数
	println("根据命令行参数指定的配置文件路径，读取配置文件参数...\n")

	// 3. 注册所需的模块组件
	println("注册所需的模块组件...\n")
	var mp m.MediatorPlugin

	// 4. 根据命令行参数初始主程序
	println("根据命令行参数初始主程序...\n")
	var db a.DataBase
	db.Initialize()

	// 5. 根据命令行参数初始程序组件
	println("根据命令行参数初始程序组件...\n")
	mp.PluginInitialize()

	// 6. 启动主程序
	println("启动主程序...\n")
	db.Startup()

	// 7. 启动程序组件
	println("启动程序组件...\n")
	ch := make(chan int)
	go mp.PluginStartup(&db, ch)

	fmt.Printf("Started mediator node on a chain with %d verified uints.\n",
		db.DynGlobalProp.LastVerifiedUnitNum)

	// 8. 本协程睡眠直到其他协程唤醒
	signal := <-ch
	println("Exiting from signal %d", signal)
}
