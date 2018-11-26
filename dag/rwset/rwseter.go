package rwset

import "github.com/palletone/go-palletone/dag/modules"

type TxSimulator interface {
	GetState(contractid []byte, ns string, key string) ([]byte, error)
	SetState(ns string, key string, value []byte) error
	GetTokenBalance(ns, addr string, asset *modules.Asset) (map[modules.Asset]uint64, error)
	PayOutToken(ns string, address string, token *modules.Asset, amount uint64, lockTime uint32) error
	DefineToken(ns string, tokenType int32, define []byte) error
	SupplyToken(ns string, assetId, uniqueId []byte, amt uint64) error
	DeleteState(ns string, key string) error
	GetContractStatesById(contractid []byte) (map[string]*modules.ContractStateValue, error)
	GetRwData(ns string) (map[string]*KVRead, map[string]*KVWrite, error)
	GetPayOutData(ns string) ([]*modules.TokenPayOut, error)
	GetTxSimulationResults() ([]byte, error)
	CheckDone() error
	Done()
}
