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
	"log"

	"testing"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common/ptndb"
)

func TestDBBatch(t *testing.T) {
	Dbconn, _ := ptndb.NewMemDatabase()
	if Dbconn == nil {
		fmt.Println("Connect mem  db error.")
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
		fmt.Println("Connect mem db error.")
		return
	}
	//l := plog.NewTestLog()
	utxodb := NewUtxoDb(Dbconn)

	//1. construct object
	myplane := NewAirPlane()
	fmt.Println("myplane is :", myplane)
	myplane2 := NewAirPlane()
	fmt.Println("myplane2 is :", myplane2)
	cap := make([]airPlane, 0)
	cap = append(cap, *myplane, *myplane2)
	fmt.Println(" cap :", cap)
	//2. store object
	StoreToRlpBytes(utxodb.db, []byte("testkey"), &cap)
	//3. load object
	something, err := utxodb.db.Get([]byte("testkey"))

	fmt.Println("db get err:", err)
	fmt.Println("byte data:", something)
	p := new([]airPlane)
	err2 := rlp.DecodeBytes(something, p)
	fmt.Println("decoded error:", err2)
	fmt.Printf("decoded data:%v\n", p)

}

//func TestAddToken(t *testing.T) {
//	// dbconn := ReNewDbConn("/Users/jay/code/gocode/src/github.com/palletone/go-palletone/bin/work/gptn/leveldb/")
//	// if dbconn == nil {
//	// 	fmt.Println("Connect to db error.")
//	// 	return
//	// }
//	dbconn, _ := ptndb.NewMemDatabase()
//
//	token := new(modules.TokenInfo)
//	token.TokenHex = modules.PTNCOIN.String()
//	token.Token = modules.PTNCOIN
//	token.Name = "ptn"
//	token.Creator = "jay"
//	token.CreationDate = time.Now().Format(modules.TimeFormatString)
//	infos := new(tokenInfo)
//	infos.Items = make(map[string]*modules.TokenInfo)
//	infos.Items[string(constants.TOKENTYPE)+token.TokenHex] = token
//	// bytes, err := rlp.EncodeToBytes(infos)
//	// if err != nil {
//	// 	t.Errorf("error: %v", err)
//	// 	return
//	// }
//	bytes, _ := json.Marshal(infos)
//	if err := dbconn.Put(constants.TOKENINFOS, bytes); err != nil {
//		t.Error("failed")
//		return
//	}
//
//	if bytes, err := dbconn.Get(constants.TOKENINFOS); err != nil {
//		t.Error("get token infos error:", err)
//		return
//	} else {
//		log.Println("json  bytes:", bytes)
//		token_info := new(tokenInfo)
//		token_info.Items = make(map[string]*modules.TokenInfo)
//		// if err := rlp.DecodeBytes(bytes, &token_info); err != nil {
//		// 	t.Error("decode error:", err)
//		// 	return
//		// }
//		err := json.Unmarshal(bytes, &token_info)
//		log.Println("token_info: ", err, token_info)
//	}
//}
//
//type tokenInfo struct {
//	Items map[string]*modules.TokenInfo //  token_info’json string
//}
