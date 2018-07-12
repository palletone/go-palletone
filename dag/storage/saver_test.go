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
	"errors"
	"github.com/palletone/go-palletone/common"
	"log"
	"strconv"
	"testing"
	"time"

	palletdb "github.com/palletone/go-palletone/common/ptndb"
	config "github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/modules"
)

func TestSaveJoint(t *testing.T) {
	if IsGenesisUnit("123") {
		log.Println("faile")
		t.Error("faild")
	} else {
		log.Println("success")
		t.Log("success")
	}
	log.Println(strconv.FormatInt(time.Now().Unix(), 10))
	var ty []modules.IDType16
	var p []common.Hash
	ty = append(ty, modules.PTNCOIN, modules.BTCCOIN)
	log.Println("assets:", ty[0].String(), ty[1].String())
	p = append(p, common.HexToHash("123"), common.HexToHash("456"))
	h := modules.NewHeader(p, ty, uint64(111), uint64(222), []byte("hello"))
	txs := make(modules.Transactions, 0)
	u := modules.NewUnit(h, txs)
	err := SaveJoint(&modules.Joint{Unit: u},
		func() { log.Println("ok") })
	log.Println("error:", err)
}

func TestAddUnitKey(t *testing.T) {
	//keys := GetUnitKeys()
	// if len(keys) <= 0 {
	// 	return errors.New("null keys.")
	// }
	keys := []string{"unit1231526522017", "unit1231526521834"}
	var err error
	if Dbconn == nil {
		Dbconn, err = palletdb.NewLDBDatabase(config.DefaultConfig.DbPath, 0, 0)
		if err != nil {
			log.Println("new db error", err)
			t.Fatal("error1")
		}
	}
	value := []int{123456, 987654}
	for i, v := range keys {
		log.Println("key: ", v, "value: ", value[i])
		if err := Dbconn.Put([]byte(v), ConvertBytes(value[i])); err != nil {
			log.Println("put error", err)
			t.Fatal("error2")
		}
		log.Println("this value:", string(ConvertBytes(value[i])))
	}

	log.Println("success")
}

func TestGetUnitKeys(t *testing.T) {
	t0 := time.Now()

	keys := GetUnitKeys()
	var this []string
	for i, v := range keys {
		var exist bool
		for j := i + 1; j < len(keys); j++ {
			if v == keys[j] {
				log.Println("j:", j)
				exist = true
				log.Println("equal", v)
				break
			}
		}
		if !exist {
			// log.Println("i:", i)
			this = append(this, v)
		}
	}

	err := AddUnitKeys("unit1231526521834")
	if errors.New("key is already exist.").Error() == err.Error() {
		log.Println("success test add unit", keys) // this
	} else {
		log.Println("failed test add  unit ")
	}
	log.Println("times:", (time.Now().UnixNano()-t0.UnixNano())/1e6)
}

func TestDBBatch(t *testing.T) {
	log.Println("db_path:", DBPath)
	table := palletdb.NewTable(Dbconn, "hehe")
	err0 := table.Put([]byte("jay"), []byte("baby"))
	log.Println("err0:", err0)

	b, err := table.Get([]byte("jay"))
	log.Println("b:", string(b), err)

	log.Println("table:", table)
}
