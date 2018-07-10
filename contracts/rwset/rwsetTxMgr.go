package rwset


//import "sync"

type BaseTxMgr struct {
	//db                      DB
	//rwLock            		sync.RWMutex
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

func (s *BaseTxMgr)GetState(){

}