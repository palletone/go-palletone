/*
	This file is part of go-palletone.
	go-palletone is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.
	go-palletone is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.
	You should have received a copy of the GNU General Public License
	along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
*/

/*
 * @author PalletOne core developers <dev@pallet.one>
 * @date 2018-2020
 */
package rwset

import (
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/dag/modules"
)

type TxManager interface {
	Close()
	CloseTxSimulator(txId string) error
	NewTxSimulator(idag IDataQuery, txId string) (TxSimulator, error)
	GetTxSimulator(txId string) (TxSimulator, error)
}

type TxSimulator interface {
	IStateQuery
	GetState(contractId []byte, ns string, key string) ([]byte, error)
	GetAllStates(contractId []byte, ns string) (map[string]*modules.ContractStateValue, error)
	GetStatesByPrefix(contractId []byte, ns string, prefix string) ([]*modules.KeyValue, error)
	GetTimestamp(ns string, rangeNumber uint32) ([]byte, error)
	SetState(contractid []byte, ns string, key string, value []byte) error
	GetTokenBalance(ns string, addr common.Address, asset *modules.Asset) (map[modules.Asset]uint64, error)
	GetStableTransactionByHash(ns string, hash common.Hash) (*modules.Transaction, error)
	GetStableUnit(ns string, hash common.Hash, unitNumber uint64) (*modules.Unit, error)
	PayOutToken(ns string, address string, token *modules.Asset, amount uint64, lockTime uint32) error
	DefineToken(ns string, tokenType int32, define []byte, creator string) error
	SupplyToken(ns string, assetId, uniqueId []byte, amt uint64, creator string) error
	DeleteState(contractId []byte, ns string, key string) error

	GetRwData(ns string) ([]*KVRead, []*KVWrite, error)
	GetPayOutData(ns string) ([]*modules.TokenPayOut, error)
	GetTokenDefineData(ns string) (*modules.TokenDefine, error)
	GetTokenSupplyData(ns string) ([]*modules.TokenSupply, error)
	//GetTxSimulationResults() ([]byte, error)
	CheckDone() error
	Done()
	Close()
	Rollback() error
	String() string
	//GetChainParameters() ([]byte, error)
	GetGlobalProp() ([]byte, error)
}
