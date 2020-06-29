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

package palletcache

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/palletone/go-palletone/dag/palletcache/cache"
	"github.com/palletone/go-palletone/dag/palletcache/redis"
)

var configs string = "redis"

func Init() {
	switch configs {
	case "redis":
		redis.Init()
		Isredispool = true
		log.Println("init redis ,", Isredispool)
	// case "cache":

	default:
		cache.Init()
	}
}
func TestCacheStore(t *testing.T) {

	err := Store("unit", "unit2", 87654321, 0)
	log.Println("re", err)

	re, has := cache.Get([]byte("unit2"))
	log.Println("result:", string(re), has) //  result: 87654321 true

}

func TestRedisStore(t *testing.T) {
	Init()
	err := Store("unit", "unit2", "hello2", 0)
	log.Println("re", err)

	re1, ok1 := Get("unit", "unit2")
	log.Println("re1:", ok1, re1) // true  hello2

	Del("unit", "unit2")
	re2, ok2 := Get("unit", "unit2")
	log.Println("re2:", ok2, re2) // false <nil>
}

func TestGetFileHash(t *testing.T) {
	t0 := time.Now()
	err_count := 0
	count := 10
	createProofTx()
	time.Sleep(5 * time.Second)
	for i := 0; i < count; i++ {
		// test case: localhost
		err := httpDo("POST", "http://39.106.209.219:8545",
			`{"jsonrpc":"2.0","method":"wallet_getFileInfoByFileHash","params":["abcdef"],"id":1}`)
		if err != nil {
			err_count += 1
		}
	}
	fmt.Printf("getFileInfoByFileHash %d spent time:%s", count, time.Since(t0).String())
	//assert.Equal(t, 0, err_count)
}

//P1K3FJLkTf821wHXquD3QYdBqvZc2ooChjs
func createProofTx() {
	// test case: localhost
	err := httpDo("POST", "http://39.106.209.219:8545",
		`{"jsonrpc":"2.0","method":"wallet_createProofOfExistenceTx","params":["P1K3FJLkTf821wHXquD3QYdBqvZc2ooChjs","abcdef","附加信息","123456","1"],"id":1}`)
	if err != nil {
		log.Println(err)
	}
}
func newAccount() {
	// test case: localhost
	err := httpDo("POST", "http://39.106.209.219:8545",
		`{"jsonrpc":"2.0","method":"personal_newAccount","params":["1"],"id":1}`)
	if err != nil {
		log.Println(err)
	}
}
func httpDo(method string, url string, msg string) error {

	client := &http.Client{}
	body := bytes.NewBuffer([]byte(msg))
	req, err := http.NewRequest(method,
		url,
		body)
	if err != nil {
		log.Println("http new request error:", err)
	}
	req.Header.Set("Content-Type", "application/json;charset=utf-8")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return err
	}

	defer resp.Body.Close()
	result_body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return err
	}

	fmt.Println(string(result_body))
	return nil
}
