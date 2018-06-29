package common

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/crypto/sha3"
	"github.com/palletone/go-palletone/common/rlp"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/storage"
)

func RlpHash(x interface{}) (h Hash) {
	hw := sha3.NewKeccak256()
	rlp.Encode(hw, x)
	hw.Sum(h[:0])
	return h
}

func RHashStr(x interface{}) string {
	x_byte, err := json.Marshal(x)
	if err != nil {
		return ""
	}
	s256 := sha256.New()
	s256.Write(x_byte)
	return fmt.Sprintf("%x", s256.Sum(nil))

}

//  last unit
func CurrentUnit() *modules.Unit {
	return &modules.Unit{creationdate: time.Now()}
}

// get unit
func GetUnit(hash *common.Hash, index modules.ChainIndex) *modules.Unit {
	unit_bytes, err := storage.Get(append(storage.UNIT_PREFIX + hash...))
	if err != nil {
		return nil
	}
	unit := new(modules.Unit)
	if err := json.Unmarshal(unit_bytes, &unit); err == nil {
		return unit
	}
	return nil
}
