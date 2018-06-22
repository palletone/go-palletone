/**
@version 0.1
@author albert·gou
@time June 11, 2018
@brief DB和全局变量的初始化功能
*/

package application

import (
	"time"

	d "github.com/palletone/go-palletone/consensus/dpos"
	s "github.com/palletone/go-palletone/consensus/dpos/mediators"
)

type VerifiedUnit struct {
	PreVerifiedUnit *VerifiedUnit // 前一个验证单元的hash
	MediatorSig     string		// 验证单元签名信息
	Timestamp		time.Time	// 时间戳
	VerifiedUnitNum uint32		// 验证单元编号
}

type DataBase struct {
	VerifiedUnits	[]*VerifiedUnit
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
	println("open db!")

	// 2. 初始化区块链数据...
	println("initilize blockchain data start!")

	println("initilize genesis verified uint!")
	gvu := VerifiedUnit{nil, "",
	time.Unix(time.Now().Unix(), 0), 0}	//创世单元
	var vus []*VerifiedUnit
	vus = append(vus, &gvu)

	println("initilize mediators!")
	var ms []*d.Mediator
	ms = append(ms, &Mediator1)
	ms = append(ms, &Mediator2)
	ms = append(ms, &Mediator3)

	db.VerifiedUnits = vus
	db.Mediators = ms

	println("initilize blockchain data end!\n")
}

func (db *DataBase) Startup() {
	// 2. 初始化全局属性...
	println("initilize global property...")
//	db.GlobalProp.ChainParameters.MaintenanceSkipSlots = 3
	db.GlobalProp.ChainParameters.VerifiedUnitInterval = 3

	println("Set active mediators...\n")
	db.GlobalProp.ActiveMediators = append(db.GlobalProp.ActiveMediators, &Mediator1)
	db.GlobalProp.ActiveMediators = append(db.GlobalProp.ActiveMediators, &Mediator2)
	db.GlobalProp.ActiveMediators = append(db.GlobalProp.ActiveMediators, &Mediator3)

	println("initilize dynamic global property...")

	db.DynGlobalProp.LastVerifiedUnitNum = 0
//	db.DynGlobalProp.VerifiedUnitHash = "0x000000"
	db.DynGlobalProp.LastVerifiedUnitTime = time.Unix(time.Now().Unix(), 0)
	db.DynGlobalProp.CurrentMediator = nil
	db.DynGlobalProp.CurrentASlot = 0
//	db.DynGlobalProp.RecentSlotsFilled = 100

	println("Create witness scheduler...\n")
	for _, m := range db.GlobalProp.ActiveMediators {
		db.MediatorSchl.CurrentShuffledMediators =append(db.MediatorSchl.CurrentShuffledMediators, m)
	}
}
