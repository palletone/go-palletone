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
 * @brief 实现DB和全局变量的初始化功能。implement the initialization function of database and global variables.
 */

package application

import (
	"time"

	"github.com/palletone/go-palletone/common/log"

	d "github.com/palletone/go-palletone/consensus/dpos"
	s "github.com/palletone/go-palletone/consensus/dpos/mediators"
	v "github.com/palletone/go-palletone/dag/verifiedunit"
)

type DataBase struct {
	VerifiedUnits	[]*v.VerifiedUnit
	Mediators     	[]*d.Mediator

	GlobalProp		d.GlobalProperty
	DynGlobalProp	d.DynamicGlobalProperty
	MediatorSchl	s.MediatorSchedule
}

var (
	Mediator1 = d.Mediator{"mediator1", "mediator1`sig"}
	Mediator2 = d.Mediator{"mediator2", "mediator2`sig"}
	Mediator3 = d.Mediator{"mediator3", "mediator3`sig"}
)

func (db *DataBase) Initialize() {
	// 1. 打开区块链数据库...
//	println("open database!")
	log.Info("open database!")

	// 2. 初始化区块链数据...
//	println("initilize blockchain data start!")
	log.Info("initilize blockchain data start!")

//	println("initilize genesis verified uint!")
	log.Info("initilize genesis verified uint!")
	gvu := v.VerifiedUnit{nil, "",
	time.Unix(time.Now().Unix(), 0), 0}	//创世单元
	var vus []*v.VerifiedUnit
	vus = append(vus, &gvu)

//	println("initilize mediators!")
	log.Info("initilize mediators!")
	var ms []*d.Mediator
	ms = append(ms, &Mediator1)
	ms = append(ms, &Mediator2)
	ms = append(ms, &Mediator3)

	db.VerifiedUnits = vus
	db.Mediators = ms

//	println("initilize blockchain data end!\n")
	log.Info("initilize blockchain data end!")
}

func (db *DataBase) Startup() {
	// 2. 初始化全局属性...
//	println("initilize global property...")
	log.Info("initilize global property...")

	gp := &db.GlobalProp
//	gp.ChainParameters.MaintenanceSkipSlots = 3
	gp.ChainParameters.VerifiedUnitInterval = 3

//	println("Set active mediators...")
	log.Info("Set active mediators...")
	gp.ActiveMediators = append(gp.ActiveMediators, &Mediator1)
	gp.ActiveMediators = append(gp.ActiveMediators, &Mediator2)
	gp.ActiveMediators = append(gp.ActiveMediators, &Mediator3)

//	println("initilize dynamic global property...")
	log.Info("initilize dynamic global property...")

	dgp := &db.DynGlobalProp
	vus := &db.VerifiedUnits
	lastVU := (*vus)[len(*vus)-1]
	dgp.LastVerifiedUnitNum = lastVU.VerifiedUnitNum
//	dgp.VerifiedUnitHash = "0x000000"
	dgp.LastVerifiedUnit = lastVU
	dgp.LastVerifiedUnitTime = lastVU.Timestamp
//	dgp.CurrentMediator = nil
	dgp.CurrentASlot = 0
//	dgp.RecentSlotsFilled = 100

	ms := &db.MediatorSchl
//	println("Create mediators scheduler...\n")
	log.Info("Create mediators scheduler...")
	for _, m := range db.GlobalProp.ActiveMediators {
		ms.CurrentShuffledMediators =append(ms.CurrentShuffledMediators, m)
	}

//	ms.UpdateMediatorSchedule(gp, dgp)
}
