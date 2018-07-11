package rwset

type BaseTxMgr struct {
	//db                    DB
	//rwLock            	sync.RWMutex
	baseTxSim               *baseTxSimulator
}

// NewTxSimulator implements method in interface `txmgmt.TxMgr`
func  NewTxSimulator(chainid string, txid string) (*BaseTxMgr, error) {
	logger.Debugf("constructing new tx simulator")
	s, err := newLockBasedTxSimulator(chainid, txid)
	if err != nil {
		return nil, err
	}
	return &BaseTxMgr{s}, nil
}

func (s *BaseTxMgr)GetState(ns string, key string) ([]byte, error) {
	return s.baseTxSim.GetState(ns, key)
}

func (s *BaseTxMgr)SetState(ns string, key string, value []byte) error {
	return s.baseTxSim.SetState(ns, key, value)
}

func (s *BaseTxMgr) DeleteState(ns string, key string) error {
	return s.baseTxSim.DeleteState(ns, key)
}
