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
package core

import (
	"testing"
	"encoding/json"
	"net/http"
	"bytes"
	"io/ioutil"
	"github.com/palletone/go-palletone/contracts/syscontract"
)


type TestCertRpc struct {
	Jsonrpc string `json:"jsonrpc"`
	Methond string `json:"method"`
	Params  []interface{} `json:"params"`
	Id      int    `json:"id"`
}


func TestGenCert(t *testing.T) {
	//certinfo := NewCertInfo("123456", "lkk", "HiPalletone", "user", "gpt2·2n.mediator1", true)
	//cainfo := CertInfo2Cainfo(*certinfo)
	params := TestCertRpc{}
	params.Jsonrpc = "2.0"
	params.Methond = "ptn_ccinvoketx"
	params.Id = 1
	form := "P135UmGibaAahtiBet3hvZm8pDsu5V1yRhK"
	to := "P135UmGibaAahtiBet3hvZm8pDsu5V1yRhK"
	amount := "100"
	fee := "1"
	contractid := syscontract.DigitalIdentityContractAddress.String()
	//method1 := []string{"P15UfoQzo93aSM3R2rDVHemiDJoMKRSLoaD","P15UfoQzo93aSM3R2rDVHemiDJoMKRSLoaD","100","1","PCGTta3M4t3yXu8uRgkKvaWd2d8DRv2vsEk"}
	method2 := []string{"addServerCert","P135UmGibaAahtiBet3hvZm8pDsu5V1yRhK","This is test cert byte!"}
	//method2 := TestAddCertJson{ []string{"addServerCert","P1HrTpdqBmCrNhJMGREu7vtyzmhCiPiztkL","E:\\codes\\go\\src\\github.com\\palletone\\go-palletone\\cmd\\gptn\\data\\certs\\openssl\\powerca\\certs\\powerca.cert.pem"}}
	params.Params = append(params.Params,form,to,amount,fee,contractid)
	//params.Params = append(params.Params,str2)
	params.Params = append(params.Params,method2)
	reqJson, err := json.Marshal(params)
	if err != nil {
		t.Log(err)
	}
	t.Log(string(reqJson))
	httpReq, err := http.NewRequest("POST", "http://localhost:8545", bytes.NewBuffer(reqJson))
	if err != nil {
		t.Log(err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpClient := &http.Client{}
	resp, err := httpClient.Do(httpReq)
	if err != nil {
		t.Log(err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Log(err)
	}
	t.Log(string(body))

}

//
//func TestRPC(t *testing.T) {
//	params := TestCertRpc{}
//	params.Jsonrpc = "2.0"
//   params.Methond = "wallet_getBalance"
//	params.Params = []string{"P15UfoQzo93aSM3R2rDVHemiDJoMKRSLoaD"}
//	params.Id = 1
//
//	reqJson, err := json.Marshal(params)
//	if err != nil {
//		  t.Log(err)
//	}
//   t.Log(string(reqJson))
//	httpReq, err := http.NewRequest("POST", "http://123.126.106.82:38545", bytes.NewBuffer(reqJson))
//	if err != nil {
//		t.Log(err)
//	}
//	httpReq.Header.Set("Content-Type", "application/json")
//	httpClient := &http.Client{}
//	resp, err := httpClient.Do(httpReq)
//	if err != nil {
//		t.Log(err)
//	}
//
//	defer resp.Body.Close()
//
//	body, err := ioutil.ReadAll(resp.Body)
//	if err != nil {
//		t.Log(err)
//	}
//	t.Log(string(body))
//}