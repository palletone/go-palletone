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
 *  * @author PalletOne core developers <dev@pallet.one>
 *  * @date 2018-2019
 *
 */
package migration

import (
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/dag/constants"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/storage"
)

type Migration102beta_102gamma struct {
	dagdb   ptndb.Database
	idxdb   ptndb.Database
	utxodb  ptndb.Database
	statedb ptndb.Database
	propdb  ptndb.Database
}

func (m *Migration102beta_102gamma) FromVersion() string {
	return "1.0.2-beta"
}

func (m *Migration102beta_102gamma) ToVersion() string {
	return "1.0.2-gamma"
}

func (m *Migration102beta_102gamma) ExecuteUpgrade() error {
	//转换GLOBALPROPERTY结构体
	//if err := m.upgradeGP(); err != nil {
	//	return err
	//}

	// 转换mediator结构体
	if err := m.upgradeMediatorInfo(); err != nil {
		return err
	}

	return nil
}

func (m *Migration102beta_102gamma) upgradeMediatorInfo() error {
	oldMediatorsIterator := m.statedb.NewIteratorWithPrefix(constants.MEDIATOR_INFO_PREFIX)
	for oldMediatorsIterator.Next() {
		oldMediator := &MediatorInfo101{}
		err := rlp.DecodeBytes(oldMediatorsIterator.Value(), oldMediator)
		if err != nil {
			log.Debugf(err.Error())
			return err
		}

		mib := &core.MediatorInfoBase{
			AddStr:     oldMediator.AddStr,
			RewardAdd:  oldMediator.AddStr,
			InitPubKey: oldMediator.InitPubKey,
			Node:       oldMediator.Node,
		}

		newMediator := &modules.MediatorInfo{
			MediatorInfoBase:   mib,
			MediatorApplyInfo:  oldMediator.MediatorApplyInfo,
			MediatorInfoExpand: oldMediator.MediatorInfoExpand,
		}

		err = storage.StoreToRlpBytes(m.statedb, oldMediatorsIterator.Key(), newMediator)
		if err != nil {
			log.Debugf(err.Error())
			return err
		}
	}

	return nil
}

//func (m *Migration102beta_102gamma) upgradeGP() error {
//	oldGp := &GlobalProperty101{}
//	err := storage.RetrieveFromRlpBytes(m.propdb, constants.GLOBALPROPERTY_KEY, oldGp)
//	if err != nil {
//		log.Errorf(err.Error())
//		return err
//	}
//
//	newData := &modules.GlobalPropertyTemp{}
//	newData.ActiveJuries = oldGp.ActiveJuries
//	newData.ActiveMediators = oldGp.ActiveMediators
//	newData.PrecedingMediators = oldGp.PrecedingMediators
//	newData.ChainParameters = oldGp.ChainParameters
//
//	newData.ImmutableParameters.MinMaintSkipSlots = 2
//	newData.ImmutableParameters.MinimumMediatorCount = oldGp.ImmutableParameters.MinimumMediatorCount
//	newData.ImmutableParameters.MinMediatorInterval = oldGp.ImmutableParameters.MinMediatorInterval
//	newData.ImmutableParameters.UccPrivileged = oldGp.ImmutableParameters.UccPrivileged
//	newData.ImmutableParameters.UccCapDrop = oldGp.ImmutableParameters.UccCapDrop
//	newData.ImmutableParameters.UccNetworkMode = oldGp.ImmutableParameters.UccNetworkMode
//	newData.ImmutableParameters.UccOOMKillDisable = oldGp.ImmutableParameters.UccOOMKillDisable
//
//	err = storage.StoreToRlpBytes(m.propdb, constants.GLOBALPROPERTY_KEY, newData)
//	if err != nil {
//		log.Errorf(err.Error())
//		return err
//	}
//
//	return nil
//}

//type GlobalProperty101 struct {
//	GlobalPropBase101
//
//	ActiveJuries       []common.Address
//	ActiveMediators    []common.Address
//	PrecedingMediators []common.Address
//}
//
//type GlobalPropBase101 struct {
//	ImmutableParameters ImmutableChainParameters101 // 不可改变的区块链网络参数
//	ChainParameters     core.ChainParameters        // 区块链网络参数
//}
//
//type ImmutableChainParameters101 struct {
//	MinimumMediatorCount uint8    `json:"min_mediator_count"`    // 最小活跃mediator数量
//	MinMediatorInterval  uint8    `json:"min_mediator_interval"` // 最小的生产槽间隔时间
//	UccPrivileged        bool     `json:"ucc_privileged"`        // 防止容器以root权限运行
//	UccCapDrop           []string `json:"ucc_cap_drop"`          // 确保容器以最小权限运行
//	UccNetworkMode       string   `json:"ucc_network_mode"`      // 容器运行网络模式
//	UccOOMKillDisable    bool     `json:"ucc_oom_kill_disable"`  // 是否内存使用量超过上限时系统杀死进程
//}

type MediatorInfoBase101 struct {
	AddStr     string `json:"account"`    // mediator账户地址
	InitPubKey string `json:"initPubKey"` // mediator的群签名初始公钥
	Node       string `json:"node"`       // mediator节点网络信息，包括ip和端口等
}

type MediatorInfo101 struct {
	*MediatorInfoBase101
	*core.MediatorApplyInfo
	*core.MediatorInfoExpand
}
