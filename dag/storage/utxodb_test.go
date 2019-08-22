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
	"github.com/palletone/go-palletone/tokenengine"
	"log"
	"testing"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/stretchr/testify/assert"
)

func TestGetUtxos(t *testing.T) {

	db, _ := ptndb.NewMemDatabase()
	utxodb := NewUtxoDb(db, tokenengine.Instance)
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
		log.Printf("get all utxo error:%s", err)
	}
	for key, u := range utxos {

		log.Printf("key:%s", key.ToKey())
		log.Printf("utxo value:%#v", u)
	}
	queryUtxo, err := utxodb.GetUtxoEntry(key)
	assert.Nil(t, err)
	assert.Equal(t, utxo.Bytes(), queryUtxo.Bytes())
}
