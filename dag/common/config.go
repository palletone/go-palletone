package common

import (
	"fmt"

	//"github.com/palletone/go-palletone/common/rlp"
	"bytes"
	"encoding/binary"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/rlp"
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
		log.Error("Get config error.")
	}

	return data
}

func GetIntParam(paramName string) int {
	value := GetConfig([]byte(paramName))
	buf := bytes.NewReader(value)

	var tmp int
	err := binary.Read(buf, binary.BigEndian, tmp)
	if err != nil {
		log.Error("binary.Read failed:", err)
	}

	return tmp
}

/**
存储配置信息
*/
func SaveConfig(confs map[string]interface{}) error {
	for k, v := range confs {
		key := fmt.Sprintf("%s_%s", CONF_PREFIX, k)
		data, err := rlp.EncodeToBytes(v)
		if err != nil {
			log.Error("Save config error.")
			return err
		}
		if err := storage.Store(key, data); err != nil {
			log.Error("Save config error.")
			return err
		}
	}
	return nil
}
