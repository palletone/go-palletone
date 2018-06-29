/**
@version 0.1
@author albert·gou
@time June 11, 2018
@brief mediator节点的初始化和启动功能
*/

package main

import (
	"fmt"

	"github.com/palletone/go-palletone/common/log"

	a "github.com/palletone/go-palletone/core/application"
	m "github.com/palletone/go-palletone/core/mediatorplugin"
)

func main() {
//	fmt.Println("mediator node starting...\n")
	log.Info("mediator node starting...")

	// 1. 拆解程序命令行参数和值
//	println("拆解程序命令行参数和值...\n")
//	println("parse command line...\n")
	log.Info("parse command line...")

	// 2. 根据命令行参数指定的配置文件路径，读取配置文件参数
//	println("根据命令行参数指定的配置文件路径，读取配置文件参数...\n")
//	println("load config file...\n")
	log.Info("load config file...")

	// 3. 注册所需的模块组件
//	println("注册所需的模块组件...\n")
//	println("register plugins...\n")
	log.Info("register plugins...")

	var mp m.MediatorPlugin

	// 4. 根据命令行参数初始主程序
//	println("根据命令行参数初始主程序...\n")
//	println("initialize application...\n")
	log.Info("initialize application...")
	var db a.DataBase
	db.Initialize()

	// 5. 根据命令行参数初始程序组件
//	println("根据命令行参数初始程序组件...\n")
//	println("initialize plugins...\n")
	log.Info("initialize plugins...")
	mp.PluginInitialize()

	// 6. 启动主程序
//	println("启动主程序...\n")
//	println("startup application...\n")
	log.Info("startup application...")
	db.Startup()

	// 7. 启动程序组件
//	println("启动程序组件...\n")
//	println("startup plugins...\n")
	log.Info("startup plugins...")
	ch := make(chan int)
	go mp.PluginStartup(&db, ch)

	//fmt.Printf("Started mediator node on a chain with %d verified uints.\n",
	//	db.DynGlobalProp.LastVerifiedUnitNum)
	log.Info(fmt.Sprintf("Started mediator node on a chain with %d verified uints.",
		db.DynGlobalProp.LastVerifiedUnitNum))

	// 8. 本协程睡眠直到其他协程唤醒
	signal := <-ch
//	println("Exiting from signal %d", signal)
	log.Info(fmt.Sprintf("Exiting from signal %d", signal))
}
