package rwset

import (
	"errors"
)

// LockBasedTxSimulator is a transaction simulator used in `LockBasedTxMgr`
type baseTxSimulator struct {
	chainid                 string
	txid                    string
	rwsetBuilder            *RWSetBuilder
	writePerformed          bool
	pvtdataQueriesPerformed bool
	doneInvoked             bool
}

type VersionedValue struct {
	Value   []byte
	Version *Version
}

func newLockBasedTxSimulator(chainid string, txid string) (*baseTxSimulator, error) {
	rwsetBuilder := NewRWSetBuilder()
	logger.Debugf("constructing new tx simulator txid = [%s]", txid)
	return &baseTxSimulator{chainid, txid, rwsetBuilder, false, false, false}, nil
}


// GetState implements method in interface `ledger.TxSimulator`
func (s *baseTxSimulator) GetState(ns string, key string) ([]byte, error) {
	if err := s.checkDone(); err != nil {
		return nil, err
	}
	//get value from DB !!!
	var versionedValue *VersionedValue
	//versionedValue, err := db.GetState(ns, key)
	//if err != nil {
	//	return nil, err
	//}

	val, ver := decomposeVersionedValue(versionedValue)
	if s.rwsetBuilder != nil {
		s.rwsetBuilder.AddToReadSet(ns, key, ver)
	}

	return val, nil
}

// SetState implements method in interface `ledger.TxSimulator`
func (s *baseTxSimulator) SetState(ns string, key string, value []byte) error {
	if err := s.checkDone(); err != nil {
		return err
	}
	if s.pvtdataQueriesPerformed {
		return errors.New("pvtdata Queries Performed")
	}
	//ValidateKeyValue

	s.rwsetBuilder.AddToWriteSet(ns, key, value)
	return nil
}

// DeleteState implements method in interface `ledger.TxSimulator`
func (s *baseTxSimulator) DeleteState(ns string, key string) error {
	return s.SetState(ns, key, nil)
}

func (h *baseTxSimulator) checkDone() error {
	if h.doneInvoked {
		return errors.New("This instance should not be used after calling Done()")
	}
	return nil
}

func decomposeVersionedValue(versionedValue *VersionedValue) ([]byte, *Version) {
	var value []byte
	var ver *Version
	if versionedValue != nil {
		value = versionedValue.Value
		ver = versionedValue.Version
	}
	return value, ver
}

func (h *baseTxSimulator) done() {
	if h.doneInvoked {
		return
	}
	//todo

}
