package modules

import (
	"github.com/palletone/go-palletone/dag/util"
)

// type 	Hash 		[]byte
type IDType [256]byte

var (
	PTNCOIN = IDType{'p', 't', 'n', ' ', '0', 'f', '5', ' ', ' '}
	BTCCOIN = IDType{'b', 't', 'c', '0', ' '}
)

func (it *IDType) String() string {
	var b []byte
	length := len(it)
	for _, v := range it {
		b = append(b, v)
	}
	count := 0
	for i := length - 1; i >= 0; i-- {
		if b[i] == ' ' || b[i] == 0 {
			count++
		} else {
			break
		}
	}
	return util.ToString(b[:length-count])
}
