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
	"log"
	"reflect"
	"testing"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/modules"
)

func TestGetUnit(t *testing.T) {
	log.Println("dbconn is nil , renew db  start ...")

	db, _ := ptndb.NewMemDatabase()
	dagdb := NewDagDatabase(db)
	dagdb.GetUnit(common.HexToHash("0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347"))
}

func TestGetContract(t *testing.T) {
	var keys []string
	var results []interface{}
	var origin modules.Contract

	origin.Id = common.HexToHash("123456")

	origin.Name = "test"
	origin.Code = []byte(`log.PrintLn("hello world")`)
	origin.Input = []byte("input")

	Dbconn := ReNewDbConn(dagconfig.DbPath)
	if Dbconn == nil {
		fmt.Println("Connect to db error.")
		return
	}

	log.Println("store error: ", StoreBytes(Dbconn, append(CONTRACT_PTEFIX, origin.Id[:]...), origin))
	keys = append(keys, "Id", "id", "Name", "Code", "code", "codes", "inputs")
	results = append(results, common.HexToHash("123456"), nil, "test", []byte(`log.PrintLn("hello world")`), nil, nil, nil)
	log.Println("test data: ", keys)

	for i, k := range keys {
		data, err := GetContractKeyValue(Dbconn, origin.Id, k)
		if !reflect.DeepEqual(data, results[i]) {
			t.Error("test error:", err, "the expect key is:", k, " value is :", results[i], ",but the return value is: ", data)
		}
	}
}

func TestUnitNumberIndex(t *testing.T) {
	key1 := fmt.Sprintf("%s_%s_%d", UNIT_NUMBER_PREFIX, modules.BTCCOIN.String(), 10000)
	key2 := fmt.Sprintf("%s_%s_%d", UNIT_NUMBER_PREFIX, modules.PTNCOIN.String(), 678934)

	if key1 != "nh_btcoin_10000" {
		log.Println("not equal.", key1)
	}
	if key2 != "nh_ptncoin_678934" {
		log.Println("not equal.", key2)
	}
}

func TestGetContractState(t *testing.T) {
	db, _ := ptndb.NewMemDatabase()
	statedb := NewStateDatabase(db)
	version, value := statedb.GetContractState("contract0000", "name")
	log.Println(version)
	log.Println(value)
	data := statedb.GetContractAllState([]byte("contract0000"))
	for k, v := range data {
		log.Println(k, v)
	}
}

func TestGetUtxos(t *testing.T) {
	//db_path := dagconfig.DefaultDataDir()
	// db_path := "/Users/jay/code/gocode/src/github.com/palletone/go-palletone/bin/work/gptn/leveldb"

	//db, err := ptndb.NewLDBDatabase(db_path, 0, 0)
	db, _ := ptndb.NewMemDatabase()

	utxodb := NewUtxoDatabase(db)
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

	utxodb.SaveUtxoEntity(key.ToKey(), utxo)

	utxos, err := utxodb.GetAllUtxos()
	for key, u := range utxos {
		log.Println("get all utxo error", err)
		log.Println("key", key.ToKey())
		log.Println("utxo value", u)
	}
	result := utxodb.GetPrefix(modules.UTXO_PREFIX)
	for key, b := range result {
		log.Println("result:", key)
		utxo := new(modules.Utxo)
		err := rlp.DecodeBytes(b, utxo)
		log.Println("utxo ", err, utxo)
	}

	result1 := utxodb.GetPrefix(AddrOutPoint_Prefix)
	for key, b := range result1 {
		log.Println("result:", key)
		out := new(modules.OutPoint)
		rlp.DecodeBytes(b, out)
		log.Println("outpoint ", err, out)
		if utxo_byte, err := db.Get(out.ToKey()); err != nil {
			log.Println("get utxo from outpoint error", err)
		} else {
			utxo := new(modules.Utxo)
			err := rlp.DecodeBytes(utxo_byte, utxo)
			log.Println("get utxo by outpoint : ", err, utxo)
		}
	}
}
