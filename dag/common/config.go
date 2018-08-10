package common

import (
	"fmt"

	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/storage"
)

var (
	CONF_PREFIX = "conf_"
)

/**
获取配置信息
get config information
*/
func GetConfig(name []byte) []byte {
	key := fmt.Sprintf("%s_%s", CONF_PREFIX, name)
	data, err := storage.Get([]byte(key))
	if err != nil {
		log.Error("Get config ", "error:", err.Error())
	}

	return data
}

/**
存储配置信息
*/
func SaveConfig(confs map[string]interface{}) error {
	for k, v := range confs {
		key := fmt.Sprintf("%s_%s", CONF_PREFIX, k)
		//data, err := rlp.EncodeToBytes(v)
		//if err != nil {
		//	log.Error("Save config error.")
		//	return err
		//}
		if storage.Dbconn == nil {
			storage.Dbconn = storage.ReNewDbConn(dagconfig.DefaultConfig.DbPath)
		}
		if err := storage.Store(storage.Dbconn, key, v); err != nil {
			log.Error("Save config error.")
			return err
		}
	}
	return nil
}
