package rwset

import (
	"testing"

	"github.com/palletone/go-palletone/dag/modules"
)

func TestKVRead_String(t *testing.T) {
	kv := &KVRead{
		key:        "key",
		version:    &modules.StateVersion{Height: modules.NewChainIndex(modules.PTNCOIN, 123), TxIndex: 2},
		value:      []byte("value"),
		ContractId: []byte("c1"),
	}
	t.Log(kv.String())
}
func TestKVRWSet_String(t *testing.T) {
	kvRead := &KVRead{
		key:        "key",
		version:    &modules.StateVersion{Height: modules.NewChainIndex(modules.PTNCOIN, 123), TxIndex: 2},
		value:      []byte("value"),
		ContractId: []byte("c1"),
	}
	kvWrite := &KVWrite{key: "wkey", isDelete: false, value: []byte("val"), ContractId: []byte("c2")}
	set := &KVRWSet{
		Reads:  map[string]*KVRead{"a": kvRead},
		Writes: map[string]*KVWrite{"b": kvWrite},
	}
	t.Log(set.String())
}
