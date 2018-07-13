package rwset

type TxSimulator interface {
	GetState(ns string, key string) ([]byte, error)
	SetState(ns string, key string, value []byte) error
	DeleteState(ns string, key string) error

	GetTxSimulationResults() ([]byte, error)
	CheckDone() error
	Done()
}
