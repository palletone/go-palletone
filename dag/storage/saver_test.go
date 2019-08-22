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
	"github.com/palletone/go-palletone/tokenengine"
	"log"
	"testing"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common/ptndb"
)

func TestDBBatch(t *testing.T) {
	Dbconn, _ := ptndb.NewMemDatabase()
	if Dbconn == nil {
		log.Println("Connect mem  db error.")
		return
	}
	table := ptndb.NewTable(Dbconn, "hehe")
	err0 := table.Put([]byte("jay"), []byte("baby"))
	log.Println("err0:", err0)

	b, err := table.Get([]byte("jay"))
	log.Println("b:", string(b), err)

	log.Println("table:", table)
}

type airPlane struct {
	Seats []string
}

func NewAirPlane() *airPlane {
	return &airPlane{
		Seats: append(make([]string, 0), "person", "person", "person", "person", "person", "person", "person", "person", "person", "person", "person"),
	}
}

func TestSaveUtxos(t *testing.T) {
	// 0. initiate db
	Dbconn, err := ptndb.NewMemDatabase()
	if Dbconn == nil {
		log.Println("Connect mem db error.")
		return
	}
	utxodb := NewUtxoDb(Dbconn, tokenengine.Instance)

	//1. construct object
	myplane := NewAirPlane()
	log.Println("myplane is :", myplane)
	myplane2 := NewAirPlane()
	log.Println("myplane2 is :", myplane2)
	cap := make([]airPlane, 0)
	cap = append(cap, *myplane, *myplane2)
	log.Println(" cap :", cap)
	//2. store object
	StoreToRlpBytes(utxodb.db, []byte("testkey"), &cap)
	//3. load object
	something, err := utxodb.db.Get([]byte("testkey"))

	log.Println("db get err:", err)
	log.Println("byte data:", something)
	p := new([]airPlane)
	err2 := rlp.DecodeBytes(something, p)
	log.Println("decoded error:", err2)
	log.Printf("decoded data:%v\n", p)

}
