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
 *  * @date 2018
 *
 */

package storage

import (
	"testing"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/ptndb"

	"github.com/palletone/go-palletone/dag/modules"
	"github.com/stretchr/testify/assert"
)

func TestGetUtxos(t *testing.T) {

	db, _ := ptndb.NewMemDatabase()
	//l := log.NewTestLog()
	utxodb := NewUtxoDb(db)
	key := new(modules.OutPoint)
	key.MessageIndex = 1
	key.OutIndex = 0
	var hash common.Hash
	hash.SetString("0xwoaibeijingtiananmen")
	key.TxHash = hash

	utxo := new(modules.Utxo)
	utxo.Amount = 10000000000000000

	utxo.Asset = &modules.Asset{AssetId: modules.PTNCOIN}
	utxo.LockTime = 123

	utxodb.SaveUtxoEntity(key, utxo)

	utxos, err := utxodb.GetAllUtxos()
	if err != nil {
		log.Debugf("get all utxo error:%s", err)
	}
	for key, u := range utxos {

		log.Debugf("key:%s", key.ToKey())
		log.Debugf("utxo value:%#v", u)
	}
	queryUtxo, err := utxodb.GetUtxoEntry(key)
	assert.Nil(t, err)
	assert.Equal(t, utxo.Bytes(), queryUtxo.Bytes())
	//result := utxodb.GetPrefix(constants.UTXO_PREFIX)
	//for key, b := range result {
	//	log.Debugf("result::%s", key)
	//	utxo := new(modules.Utxo)
	//	err := rlp.DecodeBytes(b, utxo)
	//	if err != nil {
	//		log.Errorf("utxo error:%s ", err)
	//	}
	//}
	//
	//result1 := utxodb.GetPrefix(constants.ADDR_OUTPOINT_PREFIX)
	//for key, b := range result1 {
	//	log.Debugf("result:", key)
	//	out := new(modules.OutPoint)
	//	rlp.DecodeBytes(b, out)
	//	log.Debugf("outpoint ", err, out)
	//	if utxo_byte, err := db.Get(out.ToKey()); err != nil {
	//		log.Errorf("get utxo from outpoint error:%s", err)
	//	} else {
	//		utxo := new(modules.Utxo)
	//		err := rlp.DecodeBytes(utxo_byte, utxo)
	//		log.Errorf("get utxo by outpoint :%s,%s", err, utxo)
	//	}
	//}
}
