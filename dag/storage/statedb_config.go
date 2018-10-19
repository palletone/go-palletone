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
 *  * @date 2018
 *
 */

package storage

import (
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/dag/constants"
	"github.com/palletone/go-palletone/dag/modules"
)

/**
获取配置信息
get config information
*/
func (statedb *StateDb) GetConfig(name []byte) ([]byte, *modules.StateVersion, error) {
	key := append(constants.CONF_PREFIX, name...)
	return retrieveWithVersion(statedb.db, key)

}

/**
存储配置信息
*/
func (statedb *StateDb) SaveConfig(confs []modules.PayloadMapStruct, stateVersion *modules.StateVersion) error {
	for _, conf := range confs {

		statedb.logger.Debugf("Try to save config key:{%s},Value:{%#x}", conf.Key, conf.Value)

		//if conf.Key == "Mediator" {
		//	mediators := []*core.MediatorInfo{}
		//	rlp.DecodeBytes(conf.Value, &mediators)
		//	statedb.saveMediators(mediators, stateVersion)
		//	continue
		//}

		key := append(constants.CONF_PREFIX, conf.Key...)
		//key := fmt.Sprintf("%s_%s_%s", CONF_PREFIX, conf.Key, stateVersion.String())
		err := StoreBytesWithVersion(statedb.db, key, stateVersion, conf.Value)
		if err != nil {
			log.Error("Save config error.")
			return err
		}
	}
	return nil
}

// todo albert·gou
//func (statedb *StateDb) saveMediators(mediators []*core.MediatorInfo, v *modules.StateVersion) {
//	addressList := []common.Address{}
//	for _, mediator := range mediators {
//		addr, _ := common.StringToAddress(mediator.Address)
//		addressList = append(addressList, addr)
//		statedb.SaveAccountMediatorInfo(addr, mediator, v)
//	}
//	statedb.SaveCandidateMediatorAddrList(addressList, v)
//}
