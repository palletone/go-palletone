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
