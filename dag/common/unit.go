package common

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/palletone/go-palletone/common/crypto/sha3"
	"github.com/palletone/go-palletone/common/rlp"
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
