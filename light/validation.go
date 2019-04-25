package light

import (
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/dag/modules"
)

type reqSpvData struct {
	hash common.Hash
	tx   modules.Transaction
}

type respSpvData struct {
	hash   common.Hash
	result bool
}

type ValiPathData struct {
	hash common.Hash
	path *ptndb.MemDatabase
}
