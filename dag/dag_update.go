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
 * @author PalletOne core developer Albert·Gou <dev@pallet.one>
 * @date 2018
 *
 */

package dag

//func (dag *Dag) updateLastIrreversibleUnit() {
//	aSize := dag.ActiveMediatorsCount()
//	lastConfirmedUnitNums := make([]int, 0, aSize)
//
//	// 1. 获取所有活跃 mediator 最后确认unit编号
//	meds := dag.GetActiveMediators()
//	for _, add := range meds {
//		med := dag.GetActiveMediator(add)
//		lastConfirmedUnitNums = append(lastConfirmedUnitNums, int(med.LastConfirmedUnitNum))
//	}
//
//	// 2. 排序
//	// todo 应当优化本排序方法，使用第n大元素的方法
//	sort.Ints(lastConfirmedUnitNums)
//
//	// 3. 获取倒数第 > 2/3 个确认unit编号
//	offset := aSize - dag.ChainThreshold()
//	var newLastIrreversibleUnitNum = uint64(lastConfirmedUnitNums[offset])
//
//	// 4. 更新
//	dag.updateLastIrreversibleUnitNum(newLastIrreversibleUnitNum)
//	log.Debugf("new last irreversible unit number is: %v", newLastIrreversibleUnitNum)
//}

//func (dag *Dag) updateLastIrreversibleUnitNum( /*hash common.Hash, */ newLastIrreversibleUnitNum uint64) {
//	dgp := dag.GetDynGlobalProp()
//	token := dagconfig.DagConfig.GetGasToken()
//	_, index, _ := dag.stablePropRep.GetLastStableUnit(token)
//	if newLastIrreversibleUnitNum > index.Index {
//		//dag.stablePropRep.SetLastStableUnit(hash, &modules.ChainIndex{token, true, newLastIrreversibleUnitNum})
//		dgp.LastIrreversibleUnitNum = newLastIrreversibleUnitNum
//		dag.SaveDynGlobalProp(dgp, false)
//	}
//}

//func (dag *Dag) updateGlobalPropDependGroupSign(unitHash common.Hash) {
//	unit, err := dag.GetUnitByHash(unitHash)
//	if err != nil {
//		log.Debug(err.Error())
//
//		return
//	}
//
//	// 1. 根据群签名更新不可逆unit高度
//	//dag.updateLastIrreversibleUnitNum(unitHash, uint64(unit.NumberU64()))
//}
