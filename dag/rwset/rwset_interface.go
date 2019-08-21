package rwset

import (
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/dag"
	"github.com/palletone/go-palletone/dag/modules"
)

type TxManager interface {
	Close()
	CloseTxSimulator(chainid, txid string) error
	NewTxSimulator(idag dag.IDag, chainid string, txid string, is_sys bool) (TxSimulator, error)
}

type TxSimulator interface {
	GetState(contractid []byte, ns string, key string) ([]byte, error)
	GetStatesByPrefix(contractid []byte, ns string, prefix string) ([]*modules.KeyValue, error)
	GetTimestamp(ns string, rangeNumber uint32) ([]byte, error)
	SetState(contractid []byte, ns string, key string, value []byte) error
	GetTokenBalance(ns string, addr common.Address, asset *modules.Asset) (map[modules.Asset]uint64, error)
	PayOutToken(ns string, address string, token *modules.Asset, amount uint64, lockTime uint32) error
	DefineToken(ns string, tokenType int32, define []byte, creator string) error
	SupplyToken(ns string, assetId, uniqueId []byte, amt uint64, creator string) error
	DeleteState(contractid []byte, ns string, key string) error
	GetContractStatesById(contractid []byte) (map[string]*modules.ContractStateValue, error)
	GetRwData(ns string) ([]*KVRead, []*KVWrite, error)
	GetPayOutData(ns string) ([]*modules.TokenPayOut, error)
	GetTokenDefineData(ns string) (*modules.TokenDefine, error)
	GetTokenSupplyData(ns string) ([]*modules.TokenSupply, error)
	GetTxSimulationResults() ([]byte, error)
	CheckDone() error
	Done()
	String() string

	//GetChainParameters() ([]byte, error)
	GetGlobalProp() ([]byte, error)
}
