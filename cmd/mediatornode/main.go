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
 * @brief 实现mediator节点的初始化和启动功能。Implement mediator node initialization and start function
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
