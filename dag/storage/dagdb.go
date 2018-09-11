/*
 *
 *    This file is part of go-palletone.
 *    go-palletone is free software: you can redistribute it and/or modify
 *    it under the terms of the GNU General Public License as published by
 *    the Free Software Foundation, either version 3 of the License, or
 *    (at your option) any later version.
 *    go-palletone is distributed in the hope that it will be useful,
 *    but WITHOUT ANY WARRANTY; without even the implied warranty of
 *    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *    GNU General Public License for more details.
 *    You should have received a copy of the GNU General Public License
 *    along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
 * /
 *
 *  * @author PalletOne core developer  <dev@pallet.one>
 *  * @date 2018
 *
 */

package storage

import (
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/common"

)
//对DAG对象的操作，包括：Unit，Tx等
type DagDatabase struct {
	db ptndb.Database

}
func NewDagDatabase(db ptndb.Database) *DagDatabase{
	return &DagDatabase{db: db,
	}
}
type DagDb interface {
	GetGenesisUnit() (*modules.Unit, error)
	SaveUnit(unit *modules.Unit, isGenesis bool) error
	SaveHeader(uHash common.Hash, h *modules.Header) error
	SaveTransaction( tx *modules.Transaction) error
	SaveBody(unitHash common.Hash, txsHash []common.Hash) error
	GetBody( unitHash common.Hash) ([]common.Hash, error)
	SaveTransactions( txs *modules.Transactions) error
	SaveNumberByHash( uHash common.Hash, number modules.ChainIndex) error
	SaveHashByNumber( uHash common.Hash, number modules.ChainIndex) error
	SaveTxLookupEntry(unit *modules.Unit) error

	PutCanonicalHash(hash common.Hash, number uint64) error
	PutHeadHeaderHash( hash common.Hash) error
	PutHeadUnitHash(hash common.Hash) error
	PutHeadFastUnitHash( hash common.Hash) error
	PutTrieSyncProgress(count uint64) error

	GetUnit(hash common.Hash) *modules.Unit
	GetUnitTransactions(hash common.Hash) (modules.Transactions, error)
	GetTransaction(hash common.Hash) (*modules.Transaction, common.Hash, uint64, uint64)
}
func(dagdb *DagDatabase) GetGenesisUnit() (*modules.Unit, error){

}
func(dagdb *DagDatabase) SaveUnit(unit *modules.Unit, isGenesis bool) error{

}