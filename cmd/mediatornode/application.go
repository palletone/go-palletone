package main

import (
	d "github.com/palletone/go-palletone/consensus/dpos"
	"time"
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
)

func Initialize() (*DataBase) {
	// 1. 打开区块链数据库...
	var db DataBase

	// 2. 初始化区块链数据...
	gvu := VerifiedUnit{nil, ""}	//创世单元
	var vus []*VerifiedUnit
	vus = append(vus, &gvu)

	var ms []*d.Mediator
	ms = append(ms, &Mediator1)
	ms = append(ms, &Mediator2)

	db.VerifiedUnits = vus
	db.Mediators = ms

	return  &db
}

func Startup(db *DataBase, gp *d.GlobalProperty, dgp *d.DynamicGlobalProperty) {
	// 2. 初始化全局属性...
	gp.ChainParameters.MaintenanceSkipSlots = 3
	gp.ChainParameters.VerifiedUnitInterval = 3

	gp.ActiveMediators = append(gp.ActiveMediators, &Mediator1)
	gp.ActiveMediators = append(gp.ActiveMediators, &Mediator2)

	dgp.VerifiedUnitNum = 0
	dgp.VerifiedUnitHash = "0x000000"
	dgp.VerifiedUnitTime = time.Unix(0, 0)
	dgp.CurrentMediator = nil
	dgp.CurrentASlot = 0
	dgp.RecentSlotsFilled = 100
}
