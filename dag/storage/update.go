package storage

import (
	"github.com/palletone/go-palletone/dag/dagconfig"
)

// update
func Update(key, value []byte) error {
	if _, err := Delete(key); err != nil {
		return err
	}
	return Dbconn.Put(key, value)
}

// delete
func Delete(key []byte) error {
	if Dbconn == nil {
		Dbconn = ReNewDbConn(dagconfig.DefaultConfig.DbPath)
	}
	return Dbconn.Delete(key)
}
