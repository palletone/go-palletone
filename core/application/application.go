package application

import (
	"time"

	d "github.com/palletone/go-palletone/consensus/dpos"
	s "github.com/palletone/go-palletone/consensus/dpos/mediators"
)

type VerifiedUnit struct {
	ParentVerifiedUnit *VerifiedUnit
	MediatorSig string
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
	println("\n open db!")

	// 2. 初始化区块链数据...
	println("initilize blockchain data start!")

	println("initilize genesis verified uint!")
	gvu := VerifiedUnit{nil, ""}	//创世单元
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
	println("\ninitilize global property...")
	db.GlobalProp.ChainParameters.MaintenanceSkipSlots = 3
	db.GlobalProp.ChainParameters.VerifiedUnitInterval = 3

	db.GlobalProp.ActiveMediators = append(db.GlobalProp.ActiveMediators, &Mediator1)
	db.GlobalProp.ActiveMediators = append(db.GlobalProp.ActiveMediators, &Mediator2)
	db.GlobalProp.ActiveMediators = append(db.GlobalProp.ActiveMediators, &Mediator3)

	println("initilize dynamic global property...")

	db.DynGlobalProp.VerifiedUnitNum = 0
	db.DynGlobalProp.VerifiedUnitHash = "0x000000"
	db.DynGlobalProp.VerifiedUnitTime = time.Unix(0, 0)
	db.DynGlobalProp.CurrentMediator = nil
	db.DynGlobalProp.CurrentASlot = 0
	db.DynGlobalProp.RecentSlotsFilled = 100

	println("Set active mediators...\n")
	for _, m := range db.GlobalProp.ActiveMediators {
		db.MediatorSchl.CurrentShuffledMediators =append(db.MediatorSchl.CurrentShuffledMediators, m)
	}
}
