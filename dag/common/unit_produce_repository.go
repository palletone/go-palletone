/*
 *
 *    This file is part of go-palletone.
 *    go-palletone is free software: you can redistribute it and/or modify
 *    it under the terms of the GNU General Public License as published by
 *    the Free Software Foundation, either version 3 of the License, or
 *    (at your option) any later version.
 *    go-palletone is distributed in the hope that it will be useful,
 *    but WITHOUT ANY WARRANTY; without even the implied warranty of
 *    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *    GNU General Public License for more details.
 *    You should have received a copy of the GNU General Public License
 *    along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
 * /
 *
 *  * @author PalletOne core developer <dev@pallet.one>
 *  * @date 2018-2019
 *
 */

package common

import (
	"strconv"
	"time"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/event"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/core/sort"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/modules"
)

type IUnitProduceRepository interface {
	PushUnit(nextUnit *modules.Unit) error
	ApplyUnit(nextUnit *modules.Unit) error
	MediatorVotedResults() map[string]uint64
	Close()
	SubscribeChainMaintenanceEvent(ob AfterChainMaintenanceEventFunc)
	SubscribeActiveMediatorsUpdatedEvent(ch chan<- modules.ActiveMediatorsUpdatedEvent) event.Subscription
}

type UnitProduceRepository struct {
	unitRep  IUnitRepository
	propRep  IPropRepository
	stateRep IStateRepository
	// append by albert·gou 用于活跃mediator更新时的事件订阅
	activeMediatorsUpdatedFeed  event.Feed
	activeMediatorsUpdatedScope event.SubscriptionScope
	observers                   []AfterChainMaintenanceEventFunc

	// append by albert·gou 用于account 各种投票数据统计
	mediatorVoteTally voteTallys
}

type AfterChainMaintenanceEventFunc func(event *modules.ChainMaintenanceEvent)

func NewUnitProduceRepository(unitRep IUnitRepository, propRep IPropRepository, stateRep IStateRepository) *UnitProduceRepository {
	return &UnitProduceRepository{
		unitRep:  unitRep,
		propRep:  propRep,
		stateRep: stateRep,
	}
}

func (rep *UnitProduceRepository) SubscribeChainMaintenanceEvent(ob AfterChainMaintenanceEventFunc) {
	if rep.observers == nil {
		rep.observers = []AfterChainMaintenanceEventFunc{}
	}

	rep.observers = append(rep.observers, ob)
}

// 投票统计辅助结构体
type voteTally struct {
	candidate  common.Address
	votedCount uint64
}

func newVoteTally(candidate common.Address) *voteTally {
	return &voteTally{
		candidate:  candidate,
		votedCount: 0,
	}
}

type voteTallys []*voteTally

func (vts voteTallys) Len() int {
	return len(vts)
}

func (vts voteTallys) Less(i, j int) bool {
	mVoteI := vts[i].votedCount
	mVoteJ := vts[j].votedCount

	if mVoteI != mVoteJ {
		return mVoteI > mVoteJ
	}

	return vts[i].candidate.Less(vts[j].candidate)
}

func (vts voteTallys) Swap(i, j int) {
	vts[i], vts[j] = vts[j], vts[i]
}

func (dag *UnitProduceRepository) SubscribeActiveMediatorsUpdatedEvent(ch chan<- modules.ActiveMediatorsUpdatedEvent) event.Subscription {
	return dag.activeMediatorsUpdatedScope.Track(dag.activeMediatorsUpdatedFeed.Subscribe(ch))
}

func (d *UnitProduceRepository) Close() {
	d.activeMediatorsUpdatedScope.Close()
}

/**
 * Push unit "may fail" in which case every partial change is unwound.  After
 * push unit is successful the block is appended to the chain database on disk.
 *
 * 推块“可能会失败”，在这种情况下，每个部分地更改都会撤销。 推块成功后，该块将附加到磁盘上的链数据库。
 *
 * @return true if we switched forks as a result of this push.
 */
func (rep *UnitProduceRepository) PushUnit(newUnit *modules.Unit) error {
	var err error

	//更新数据库
	err = rep.unitRep.SaveUnit(newUnit, false)
	if err != nil {
		return err
	}

	// 2. 更新状态
	err = rep.ApplyUnit(newUnit)
	if err != nil {
		return err
	}

	return nil
}

// ApplyUnit, 运用下一个 unit 更新整个区块链状态
func (rep *UnitProduceRepository) ApplyUnit(nextUnit *modules.Unit) error {
	defer func(start time.Time) {
		log.Debugf("ApplyUnit cost time: %v", time.Since(start))
	}(time.Now())

	// 计算当前 unit 到上一个 unit 之间的缺失数量，并更新每个mediator的unit的缺失数量
	missed := rep.updateMediatorMissedUnits(nextUnit)

	// 更新全局动态属性值
	rep.updateDynGlobalProp(nextUnit, missed)

	// 更新 mediator 的相关数据
	rep.updateSigningMediator(nextUnit)

	// 更新最新不可逆区块高度
	//rep.updateLastIrreversibleUnit()

	// 判断是否到了链维护周期，并维护
	maintenanceNeeded := !(rep.GetDynGlobalProp().NextMaintenanceTime > uint32(nextUnit.Timestamp()))
	if maintenanceNeeded {
		rep.performChainMaintenance(nextUnit)
	}

	// 更新链维护周期标志
	// n.b., updateMaintenanceFlag() happens this late because GetSlotTime() / GetSlotAtTime() is needed above
	// 由于前面的操作需要调用 GetSlotTime() / GetSlotAtTime() 这两个方法，所以在最后才更新链维护周期标志
	rep.updateMaintenanceFlag(maintenanceNeeded)

	// 洗牌
	rep.updateMediatorSchedule()

	return nil
}

// 根据最新 unit 计算出生产该 unit 的 mediator 缺失的 unit 个数，
// 并更新到 mediator的相应字段中，返回数量
func (rep *UnitProduceRepository) updateMediatorMissedUnits(unit *modules.Unit) uint64 {
	missedUnits := rep.propRep.GetSlotAtTime(time.Unix(unit.Timestamp(), 0))
	if missedUnits == 0 {
		log.Errorf("Trying to push double-produced unit onto current unit?!")
		return 0
	}

	missedUnits--
	log.Debugf("the count of missed units: %v", missedUnits)

	aSize := rep.GetGlobalProp().ActiveMediatorsCount()
	if missedUnits < uint32(aSize) {
		var i uint32
		for i = 0; i < missedUnits; i++ {
			mediatorMissed := rep.propRep.GetScheduledMediator(i + 1)

			med := rep.GetMediator(mediatorMissed)
			med.TotalMissed++
			rep.stateRep.StoreMediator(med)
		}
	}

	return uint64(missedUnits)
}

// UpdateDynGlobalProp, update global dynamic data
func (rep *UnitProduceRepository) updateDynGlobalProp(unit *modules.Unit, missedUnits uint64) {
	log.Debugf("update global dynamic data")
	dgp := rep.GetDynGlobalProp()

	//dgp.HeadUnitNum = unit.NumberU64()
	//dgp.HeadUnitHash = unit.Hash()
	//dgp.HeadUnitTime = unit.Timestamp()
	rep.propRep.SetNewestUnit(unit.Header())

	dgp.LastMediator = unit.Author()
	dgp.IsShuffledSchedule = false
	dgp.RecentSlotsFilled = (dgp.RecentSlotsFilled << (missedUnits + 1)) + 1
	dgp.CurrentASlot += missedUnits + 1

	rep.propRep.StoreDynGlobalProp(dgp)

	return
}

func (rep *UnitProduceRepository) updateMediatorSchedule() {
	gp := rep.GetGlobalProp()
	dgp := rep.GetDynGlobalProp()
	ms := rep.GetMediatorSchl()

	if rep.propRep.UpdateMediatorSchedule(ms, gp, dgp) {
		log.Debugf("shuffle the scheduling order of mediators")
		rep.propRep.StoreMediatorSchl(ms)

		dgp.IsShuffledSchedule = true
		rep.propRep.StoreDynGlobalProp(dgp)
	}

	return
}

func (rep *UnitProduceRepository) updateSigningMediator(newUnit *modules.Unit) {
	// 1. 更新 签名mediator 的LastConfirmedUnitNum
	signingMediator := newUnit.Author()
	med := rep.GetMediator(signingMediator)
	if med == nil {
		log.Errorf("state db have not mediator(%v) info", signingMediator.Str())
		return
	}

	lastConfirmedUnitNum := uint32(newUnit.NumberU64())
	med.LastConfirmedUnitNum = lastConfirmedUnitNum
	rep.stateRep.StoreMediator(med)

	log.Debugf("the LastConfirmedUnitNum of mediator(%v) is: %v", med.Address.Str(), lastConfirmedUnitNum)
}

func (rep *UnitProduceRepository) GetGlobalProp() *modules.GlobalProperty {
	gp, _ := rep.propRep.RetrieveGlobalProp()
	return gp
}

func (rep *UnitProduceRepository) GetDynGlobalProp() *modules.DynamicGlobalProperty {
	dgp, _ := rep.propRep.RetrieveDynGlobalProp()
	return dgp
}

func (rep *UnitProduceRepository) GetMediatorSchl() *modules.MediatorSchedule {
	ms, _ := rep.propRep.RetrieveMediatorSchl()
	return ms
}

func (rep *UnitProduceRepository) GetMediator(add common.Address) *core.Mediator {
	med, err := rep.stateRep.RetrieveMediator(add)
	if err != nil {
		log.Error("dag", "GetMediator RetrieveMediator err", err, "address", add)
		return nil
	}
	return med
}

func (dag *UnitProduceRepository) performChainMaintenance(nextUnit *modules.Unit) {
	log.Debugf("We are at the maintenance interval")

	// 对每个账户的各种投票信息进行初步统计
	dag.performAccountMaintenance()

	// 统计投票并更新活跃 mediator 列表
	isChanged := dag.updateActiveMediators()

	// 发送更新活跃 mediator 事件，以方便其他模块做相应处理
	go dag.activeMediatorsUpdatedFeed.Send(modules.ActiveMediatorsUpdatedEvent{IsChanged: isChanged})

	// 更新要修改的区块链参数
	dag.updateChainParameters(nextUnit)

	// 计算并更新下一次维护时间
	dag.updateNextMaintenanceTime(nextUnit)

	// 清理中间处理缓存数据
	dag.mediatorVoteTally = nil
	//触发ChainMaintenanceEvent事件
	eventArg := &modules.ChainMaintenanceEvent{}
	for _, eventFunc := range dag.observers {
		go eventFunc(eventArg)
	}
}

func (dag *UnitProduceRepository) updateChainParameters(nextUnit *modules.Unit) {
	log.Debugf("update chain parameters")

	version := &modules.StateVersion{
		Height:  nextUnit.Number(),
		TxIndex: ^uint32(0),
	}
	dag.stateRep.UpdateSysParams(version)
	dag.stateRep.RefreshSysParameters()

	return
}

// 获取账户相关投票数据的直方图
func (dag *UnitProduceRepository) performAccountMaintenance() {
	log.Debugf("Tally account voting mediators")
	// 初始化数据
	mediators := dag.stateRep.GetMediators()
	dag.mediatorVoteTally = make([]*voteTally, 0, len(mediators))

	// 遍历所有账户
	mediatorVoteCount := dag.MediatorVotedResults()

	// 初始化 mediator 的投票数据
	for mediator, _ := range mediators {

		voteTally := newVoteTally(mediator)
		voteTally.votedCount = mediatorVoteCount[mediator.Str()]
		dag.mediatorVoteTally = append(dag.mediatorVoteTally, voteTally)
	}
}

func (dag *UnitProduceRepository) MediatorVotedResults() map[string]uint64 {
	mediatorVoteCount := make(map[string]uint64)

	allAccount := dag.stateRep.LookupAccount()
	for _, info := range allAccount {
		// 遍历该账户投票的mediator
		for med, _ := range info.VotedMediators {
			// 累加投票数量
			mediatorVoteCount[med] += info.Balance
		}
	}

	return mediatorVoteCount
}

func (dag *UnitProduceRepository) updateActiveMediators() bool {
	// 1. 统计出活跃mediator数量n
	maxFn := func(x, y int) int {
		if x > y {
			return x
		}
		return y
	}

	gp := dag.GetGlobalProp()

	// 保证活跃mediator的总数必须大于MinimumMediatorCount
	minMediatorCount := gp.ImmutableParameters.MinimumMediatorCount
	countInSystem := dag.getDesiredActiveMediatorCount()
	mediatorCount := maxFn((countInSystem-1)/2*2+1, int(minMediatorCount))

	mediatorLen := dag.mediatorVoteTally.Len()
	if mediatorLen < mediatorCount {
		log.Debugf("the desired mediator count is %v, the actual mediator count is %v,"+
			" the minimum mediator count is %v", countInSystem, mediatorLen, minMediatorCount)
		// 保证活跃mediator的总数为奇数
		mediatorCount = (mediatorLen-1)/2*2 + 1
	}
	log.Debugf("In this round, The active mediator's count is %v", mediatorCount)

	// 2. 根据每个mediator的得票数，排序出前n个 active mediator
	sort.PartialSort(dag.mediatorVoteTally, mediatorCount)

	// 3. 更新每个mediator的得票数
	for _, voteTally := range dag.mediatorVoteTally {
		med := dag.GetMediator(voteTally.candidate)
		if med == nil {
			log.Errorf("Cannot get mediator by:%s", voteTally.candidate.String())
		}
		med.TotalVotes = voteTally.votedCount
		dag.stateRep.StoreMediator(med)
	}

	// 4. 更新 global property 中的 active mediator 和 Preceding Mediators
	gp.PrecedingMediators = gp.ActiveMediators
	gp.ActiveMediators = make(map[common.Address]bool, mediatorCount)
	for index := 0; index < mediatorCount; index++ {
		voteTally := dag.mediatorVoteTally[index]
		gp.ActiveMediators[voteTally.candidate] = true
	}
	dag.propRep.StoreGlobalProp(gp)

	return isActiveMediatorsChanged(gp)
}

func (d *UnitProduceRepository) getDesiredActiveMediatorCount() int {
	// 获取之前的设置
	activeMediatorStr, _, _ := d.stateRep.GetConfig(modules.DesiredActiveMediatorCount)
	activeMediator, _ := strconv.ParseUint(string(activeMediatorStr), 10, 16)

	// 获取基金会直接修改的设置
	desiredSysParams, err := d.stateRep.GetSysParamWithoutVote()
	if err != nil {
		return int(activeMediator)
	}

	var desiredActiveMediatorStr string
	desiredActiveMediatorStr, ok := desiredSysParams[modules.DesiredActiveMediatorCount]
	if !ok {
		return int(activeMediator)
	}

	desiredActiveMediator, err := strconv.ParseUint(string(desiredActiveMediatorStr), 10, 16)
	if err != nil {
		return int(activeMediator)
	}
	activeMediator = desiredActiveMediator

	// todo 获取通过投票修改的设置

	return int(activeMediator)
}

func (dag *UnitProduceRepository) updateNextMaintenanceTime(nextUnit *modules.Unit) {
	dgp := dag.GetDynGlobalProp()
	gp := dag.GetGlobalProp()

	nextMaintenanceTime := dgp.NextMaintenanceTime
	maintenanceInterval := int64(gp.ChainParameters.MaintenanceInterval)

	if nextUnit.NumberU64() == 1 {
		nextMaintenanceTime = uint32((nextUnit.Timestamp()/maintenanceInterval + 1) * maintenanceInterval)
	} else {
		// We want to find the smallest k such that nextMaintenanceTime + k * maintenanceInterval > HeadUnitTime()
		//  This implies k > ( HeadUnitTime() - nextMaintenanceTime ) / maintenanceInterval
		//
		// Let y be the right-hand side of this inequality, i.e.
		// y = ( HeadUnitTime() - nextMaintenanceTime ) / maintenanceInterval
		//
		// and let the fractional part f be y-floor(y).  Clearly 0 <= f < 1.
		// We can rewrite f = y-floor(y) as floor(y) = y-f.
		//
		// Clearly k = floor(y)+1 has k > y as desired.  Now we must
		// show that this is the least such k, i.e. k-1 <= y.
		//
		// But k-1 = floor(y)+1-1 = floor(y) = y-f <= y.
		// So this k suffices.
		//

		y := (dag.HeadUnitTime() - int64(nextMaintenanceTime)) / maintenanceInterval
		nextMaintenanceTime += uint32((y + 1) * maintenanceInterval)
	}

	dgp.LastMaintenanceTime = dgp.NextMaintenanceTime
	dgp.NextMaintenanceTime = nextMaintenanceTime
	dag.propRep.StoreDynGlobalProp(dgp)

	time := time.Unix(int64(nextMaintenanceTime), 0)
	log.Debugf("nextMaintenanceTime: %v", time.Format("2006-01-02 15:04:05"))

	return
}

func (dag *UnitProduceRepository) updateMaintenanceFlag(newMaintenanceFlag bool) {
	log.Debugf("update maintenance flag: %v", newMaintenanceFlag)

	dgp := dag.GetDynGlobalProp()
	dgp.MaintenanceFlag = newMaintenanceFlag
	dag.propRep.StoreDynGlobalProp(dgp)

	return
}

func (dag *UnitProduceRepository) HeadUnitTime() int64 {
	gasToken := dagconfig.DagConfig.GetGasToken()
	t, _ := dag.propRep.GetNewestUnitTimestamp(gasToken)
	return t
}

// 判断新一届mediator和上一届mediator是否有变化
func isActiveMediatorsChanged(gp *modules.GlobalProperty) bool {
	precedingMediators := gp.PrecedingMediators
	activeMediators := gp.ActiveMediators

	// 首先考虑活跃mediator个数是否改变
	if len(precedingMediators) != len(activeMediators) {
		return true
	}

	for am, _ := range activeMediators {
		if !precedingMediators[am] {
			return true
		}
	}

	return false
}
