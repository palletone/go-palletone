package main

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
	VerifiedUnits []*VerifiedUnit
	Mediators     []*d.Mediator
}

var (
	Mediator1 = d.Mediator{"mediator1", "mediator1`sig"}
	Mediator2 = d.Mediator{"mediator2", "mediator2`sig"}
	Mediator3 = d.Mediator{"mediator3", "mediator3`sig"}
)

func Initialize() (*DataBase) {
	// 1. 打开区块链数据库...
	println("\n open db!")
	var db DataBase

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

	println("initilize blockchain data end!")

	return  &db
}

func Startup(gp *d.GlobalProperty, dgp *d.DynamicGlobalProperty, ms *s.MediatorSchedule) {
	// 2. 初始化全局属性...
	println("initilize global property...")
	gp.ChainParameters.MaintenanceSkipSlots = 3
	gp.ChainParameters.VerifiedUnitInterval = 3

	gp.ActiveMediators = append(gp.ActiveMediators, &Mediator1)
	gp.ActiveMediators = append(gp.ActiveMediators, &Mediator2)
	gp.ActiveMediators = append(gp.ActiveMediators, &Mediator3)

	println("initilize dynamic global property...")

	dgp.VerifiedUnitNum = 0
	dgp.VerifiedUnitHash = "0x000000"
	dgp.VerifiedUnitTime = time.Unix(0, 0)
	dgp.CurrentMediator = nil
	dgp.CurrentASlot = 0
	dgp.RecentSlotsFilled = 100

	println("Set active mediators...")
	for _, m := range gp.ActiveMediators {
		ms.CurrentShuffledMediators =append(ms.CurrentShuffledMediators, m)
	}
}
