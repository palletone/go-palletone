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
	"fmt"

	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/rlp"
	"github.com/palletone/go-palletone/dag/modules"

)

var (
	CONF_PREFIX = "conf"
)

/**
获取配置信息
get config information
*/
func(statedb *StateDatabase) GetConfig( name []byte) []byte {
	key := fmt.Sprintf("%s_%s", CONF_PREFIX, name)
	data := statedb.GetPrefix( []byte(key))
	if len(data) != 1 {
		log.Info("Get config ", "error", "not data")
	}
	for _, v := range data {
		var b []byte
		if err := rlp.DecodeBytes(v, &b); err != nil {
			return nil
		}
		return b
	}
	return nil
}

/**
存储配置信息
*/
func (statedb *StateDatabase) SaveConfig( confs []modules.PayloadMapStruct, stateVersion *modules.StateVersion) error {
	for _, conf := range confs {
		key := fmt.Sprintf("%s_%s_%s", CONF_PREFIX, conf.Key, stateVersion.String())
		if err := statedb.db.Put([]byte( key), conf.Value); err != nil {
			log.Error("Save config error.")
			return err
		}
	}
	return nil
}
