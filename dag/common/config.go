package common

import (
	"fmt"

	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/storage"
)

var (
	CONF_PREFIX = "conf"
)

/**
获取配置信息
get config information
*/
func GetConfig(name []byte) []byte {
	key := fmt.Sprintf("%s_%s", CONF_PREFIX, name)
	data := storage.GetPrefix([]byte(key))
	if len(data) != 1 {
		log.Info("Get config ", "error", "not data")
	}
	for _, v := range data {
		return v
	}
	return nil
}

/**
存储配置信息
*/
func SaveConfig(confs []modules.PayloadMapStruct, stateVersion *modules.StateVersion) error {
	for _, conf := range confs {
		key := fmt.Sprintf("%s_%s_%s", CONF_PREFIX, conf.Key, stateVersion.String())
		if storage.Dbconn == nil {
			storage.Dbconn = storage.ReNewDbConn(dagconfig.DbPath)
		}
		if err := storage.Store(storage.Dbconn, key, conf.Value); err != nil {
			log.Error("Save config error.")
			return err
		}
	}
	return nil
}
