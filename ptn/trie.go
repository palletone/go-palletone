package ptn

import (
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/trie"
)

func (s *PalletOne) TxsTrie() {
	trdb, err := trie.NewSecure(common.Hash{}, trie.NewDatabase(s.unitDb), 0)
	if err != nil {
		log.Debug("NewProtocolManager trie.NewSecure", "err", err)
		return
	}

	if err := trdb.TryUpdate([]byte("k_ab1"), []byte("v_ab1")); err != nil {
		log.Debug("===========1==========")
		return
	}
	if err := trdb.TryUpdate([]byte("k_ab2"), []byte("v_ab2")); err != nil {
		log.Debug("===========2==========")
		return
	}
	if err := trdb.TryUpdate([]byte("k_ab3"), []byte("v_ab3")); err != nil {
		log.Debug("===========3==========")
		return
	}
	if err := trdb.TryUpdate([]byte("k_ab4"), []byte("v_ab4")); err != nil {
		log.Debug("===========4==========")
		return
	}
	if err := trdb.TryUpdate([]byte("k_ab5"), []byte("v_ab5")); err != nil {
		log.Debug("===========5==========")
		return
	}
	trdb.Commit(nil)
}
