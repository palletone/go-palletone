/**
@version 0.1
@author albert·gou
@time June 6, 2018
@brief 主要实现全局属性的保存、更新和获取
*/

package dpos

import (
	"time"

	v "github.com/palletone/go-palletone/dag/verifiedunit"
)

// 区块链属性结构体的定义
type ChainParameters struct {
	// 验证单元之间的间隔时间，以秒为单元。 interval in seconds between verifiedUnits
	VerifiedUnitInterval uint8

	// 在维护时跳过的verifiedUnitInterval数量。 number of verifiedUnitInterval to skip at maintenance time
//	MaintenanceSkipSlots uint8
}

// 全局属性的结构体定义
type GlobalProperty struct {
	ChainParameters ChainParameters // 区块链网络参数

	ActiveMediators []*Mediator // 当前活跃mediator集合；每个维护间隔更新一次
}

// 动态全局属性的结构体定义
type DynamicGlobalProperty struct {
	LastVerifiedUnitNum uint32 // 最近的验证单元编号(数量)

//	VerifiedUnitHash string // 最近的验证单元hash

	LastVerifiedUnit *v.VerifiedUnit	// 最近生产的验证单元

	LastVerifiedUnitTime time.Time // 最近的验证单元时间

//	CurrentMediator *Mediator // 当前生产验证单元的mediator, 用于判断是否连续同一个mediator生产验证单元

//	NextMaintenanceTime time.Time // 下一次系统维护时间

	// 当前的绝对时间槽数量，== 从创世开始所有的时间槽数量 == verifiedUnitNum + 丢失的槽数量
	CurrentASlot uint64

	/**
	在过去的128个见证单元生产slots中miss的数量。
	The count of verifiedUnit production slots that were missed in the past 128 verifiedUnits
	用于计算mediator的参与率。used to compute mediator participation.
	*/
//	RecentSlotsFilled float32
}
