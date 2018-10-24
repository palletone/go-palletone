package rwset

import "github.com/palletone/go-palletone/dag/modules"

type TxSimulator interface {
	GetState(contractid []byte, ns string, key string) ([]byte, error)
	SetState(ns string, key string, value []byte) error
	DeleteState(ns string, key string) error
	GetContractStatesById(contractid []byte) (map[modules.StateVersion][]byte, error)
	GetRwData(ns string) (map[string]*KVRead, map[string]*KVWrite, error)

	GetTxSimulationResults() ([]byte, error)
	CheckDone() error
	Done()
}
