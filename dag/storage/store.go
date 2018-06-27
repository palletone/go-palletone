package storage

import (
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/util"
)

// store unit
func StoreUnit(unit *modules.Unit) {

}
func Store(key string, value interface{}) error {
	if Dbconn == nil {
		Dbconn = ReNewDbConn(dagconfig.DefaultConfig.DbPath)
	}
	val, _ := util.Bytes(value)
	return Dbconn.Put(util.ToByte(key), val)
}
func StoreString(key, value string) error {
	if Dbconn == nil {
		Dbconn = ReNewDbConn(dagconfig.DefaultConfig.DbPath)
	}
	return Dbconn.Put(util.ToByte(key), util.ToByte(value))
}
