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
 * @date 2018
 */

package storage

import (
	"fmt"
	"github.com/palletone/go-palletone/common/rlp"
	"testing"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/stretchr/testify/assert"
)

func TestGetUnit(t *testing.T) {
	//log.Println("dbconn is nil , renew db  start ...")

	db, _ := ptndb.NewMemDatabase()
	l := log.NewTestLog()
	dagdb := NewDagDb(db, l)
	u, err := dagdb.GetUnit(common.HexToHash("0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347"))
	assert.Nil(t, u, "empty db, must return nil Unit")
	assert.NotNil(t, err)
}

func TestUnitNumberIndex(t *testing.T) {
	key1 := fmt.Sprintf("%s_%s_%d", UNIT_NUMBER_PREFIX, modules.BTCCOIN.String(), 10000)
	key2 := fmt.Sprintf("%s_%s_%d", UNIT_NUMBER_PREFIX, modules.PTNCOIN.String(), 678934)

	if key1 != "nh_btcoin_10000" {
		log.Debug("not equal.", key1)
	}
	if key2 != "nh_ptncoin_678934" {
		log.Debug("not equal.", key2)
	}
}

func TestGetUtxos(t *testing.T) {

	db, _ := ptndb.NewMemDatabase()
	l := log.NewTestLog()
	utxodb := NewUtxoDb(db, l)
	key := new(modules.OutPoint)
	key.MessageIndex = 1
	key.OutIndex = 0
	var hash common.Hash
	hash.SetString("0xwoaibeijingtiananmen")
	key.TxHash = hash

	utxo := new(modules.Utxo)
	utxo.Amount = 10000000000000000

	utxo.Asset = &modules.Asset{AssetId: modules.PTNCOIN, ChainId: 1}
	utxo.LockTime = 123

	utxodb.SaveUtxoEntity(key, utxo)

	utxos, err := utxodb.GetAllUtxos()
	for key, u := range utxos {
		log.Debug("get all utxo error", err)
		log.Debug("key", key.ToKey())
		log.Debug("utxo value", u)
	}
	result := utxodb.GetPrefix(modules.UTXO_PREFIX)
	for key, b := range result {
		log.Debug("result:", key)
		utxo := new(modules.Utxo)
		err := rlp.DecodeBytes(b, utxo)
		log.Debug("utxo ", err, utxo)
	}

	result1 := utxodb.GetPrefix(AddrOutPoint_Prefix)
	for key, b := range result1 {
		log.Debug("result:", key)
		out := new(modules.OutPoint)
		rlp.DecodeBytes(b, out)
		log.Debug("outpoint ", err, out)
		if utxo_byte, err := db.Get(out.ToKey()); err != nil {
			log.Debug("get utxo from outpoint error", err)
		} else {
			utxo := new(modules.Utxo)
			err := rlp.DecodeBytes(utxo_byte, utxo)
			log.Debug("get utxo by outpoint : ", err, utxo)
		}
	}
}
