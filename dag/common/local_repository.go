/*
 *
 * 	This file is part of go-palletone.
 * 	go-palletone is free software: you can redistribute it and/or modify
 * 	it under the terms of the GNU General Public License as published by
 * 	the Free Software Foundation, either version 3 of the License, or
 * 	(at your option) any later version.
 * 	go-palletone is distributed in the hope that it will be useful,
 * 	but WITHOUT ANY WARRANTY; without even the implied warranty of
 * 	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * 	GNU General Public License for more details.
 * 	You should have received a copy of the GNU General Public License
 * 	along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
 * /
 *
 *
 *  * @author PalletOne core developer  <dev@pallet.one>
 *  * @date 2018-2020
 *
 */

package common

import (
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/storage"
)

type LocalRepository struct {
	db storage.ILocalDb
}

type ILocalRepository interface {
	SaveLocalTx(tx *modules.Transaction) error
	GetLocalTx(txId common.Hash) (*modules.Transaction, modules.TxStatus, error)
	SaveLocalTxStatus(txId common.Hash, status modules.TxStatus) error
}

func NewLocalRepository(localdb storage.ILocalDb) *LocalRepository {
	return &LocalRepository{db: localdb}
}

//通过本地RPC创建或广播的交易
func (rep *LocalRepository) SaveLocalTx(tx *modules.Transaction) error {
	return rep.db.SaveLocalTx(tx)
}

//查询某交易的内容和状态
func (rep *LocalRepository) GetLocalTx(txId common.Hash) (*modules.Transaction, modules.TxStatus, error) {
	tx, err := rep.db.GetLocalTx(txId)
	if err != nil {
		return nil, 0, err
	}
	status, err := rep.db.GetLocalTxStatus(txId)
	if err != nil {
		log.Warnf("Tx[%s] doesn't have Local tx status", txId.String())
		return tx, 0, nil
	}
	return tx, status, nil
}

//保存某交易的状态
func (rep *LocalRepository) SaveLocalTxStatus(txId common.Hash, status modules.TxStatus) error {
	return rep.db.SaveLocalTxStatus(txId, status)
}
