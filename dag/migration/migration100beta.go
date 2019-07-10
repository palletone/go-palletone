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
 *  * @author PalletOne core developer <dev@pallet.one>
 *  * @date 2018-2019
 *
 */
package migration

import (
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/dag/storage"
)

type Migration100_101 struct {
	dagdb   ptndb.Database
	idxdb   ptndb.Database
	utxodb  ptndb.Database
	statedb ptndb.Database
	propdb  ptndb.Database
}

func (m *Migration100_101) FromVersion() string {
	return "1.0.0-beta"
}
func (m *Migration100_101) ToVersion() string {
	return "1.0.1-beta"
}
func (m *Migration100_101) ExecuteUpgrade() error {
	//删除已经花费的UTXO
	dbop := storage.NewUtxoDb(m.utxodb)
	utxos, err := dbop.GetAllUtxos()
	if err != nil {
		return err
	}
	for outpoint, utxo := range utxos {
		if utxo.IsSpent() {
			err = dbop.DeleteUtxo(&outpoint, common.Hash{}, 0)
			if err != nil {
				log.Errorf("Migrate utxo db,delete spent utxo error:%s", err.Error())
				return err
			}
			//log.Debugf("Deleted spent UTXO by key:%s", outpoint.String())
		}
	}
	dagdb := storage.NewDagDb(m.dagdb)
	txs, err := dagdb.GetAllTxs()
	if err != nil {
		log.Error(err.Error())
	}
	log.Debugf("Tx count:%d", len(txs))
	for i, tx := range txs {
		if tx == nil {
			log.Errorf("tx[%d] is nil", i)
		}
		spents := tx.GetSpendOutpoints()
		for _, spent := range spents {
			stxo, err := dbop.GetStxoEntry(spent)
			if err == nil && stxo != nil {
				stxo.SpentByTxId = tx.Hash()
				lookup, _ := dagdb.GetTxLookupEntry(tx.Hash())
				stxo.SpentTime = lookup.Timestamp
				log.Debugf("Update stxo spentTxId:%s,spentTime:%d", stxo.SpentByTxId.String(), stxo.SpentTime)
				dbop.SaveStxoEntry(spent, stxo)
			}
		}
	}
	//// gp 只做了添加和删除的修改
	//fmt.Printf("exec migration , version: %s", m.FromVersion())
	//newGp := modules.NewGlobalProp()
	//newData, err := rlp.EncodeToBytes(newGp)
	//if err != nil {
	//	fmt.Println("ExecuteUpgrade error:" + err.Error())
	//	return err
	//}
	//m.statedb.Put([]byte("gpGlobalProperty"), newData)
	return nil
}
