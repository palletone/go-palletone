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
package certficate

import (
	"testing"
	"github.com/palletone/go-palletone/contracts/syscontract"
	"encoding/json"
	"net/http"
	"bytes"
	"io/ioutil"
)

type TestCertRpc struct {
	Jsonrpc string `json:"jsonrpc"`
	Methond string `json:"method"`
	Params  []interface{} `json:"params"`
	Id      int    `json:"id"`
}
//
//
func TestGenCert(t *testing.T) {
	certinfo := NewCertInfo("P1K41R4k868xDtbfF94ky4gLBp5q9XsQstE", "lkkf", "HiPalletone", "user", "gptn.mediator1", true)
	cainfo := CertInfo2Cainfo(*certinfo)
	certpem, err := cainfo.Enrolluser()

 t.Log(string(certpem))
	params := TestCertRpc{}
	params.Jsonrpc = "2.0"
	params.Methond = "ptn_ccinvoketx"
	params.Id = 1
	form := "P19b5Die2W5AV8M9ekMDZcbQ9Ntmuxe3NTT"
	to := "P19b5Die2W5AV8M9ekMDZcbQ9Ntmuxe3NTT"
	amount := "100"
	fee := "1"
	contractid := syscontract.DigitalIdentityContractAddress.String()
	//method1 := []string{"P15UfoQzo93aSM3R2rDVHemiDJoMKRSLoaD","P15UfoQzo93aSM3R2rDVHemiDJoMKRSLoaD","100","1","PCGTta3M4t3yXu8uRgkKvaWd2d8DRv2vsEk"}
	method2 := []string{"addServerCert","P1K41R4k868xDtbfF94ky4gLBp5q9XsQstE",string(certpem)}
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
//  params.Methond = "wallet_getBalance"
//	params.Params = []string{"P15UfoQzo93aSM3R2rDVHemiDJoMKRSLoaD"}
//	params.Id = 1
//
//	reqJson, err := json.Marshal(params)
//	if err != nil {
//		  t.Log(err)
//	}
//  t.Log(string(reqJson))
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

func TestCertCRL(t *testing.T) {
	address := "P12dEMJq7vA1VonEdTyXAYFL7nEjz4QTbk6"
	reason := "no resaon"
	certinfo := NewCertInfo(address, "lkk", "HiPalletone", "user", "gptn.mediator1", true)
	cainfo := CertInfo2Cainfo(*certinfo)
	crlPem,err := cainfo.Revoke(address, reason)
	t.Log(err)

    t.Log(string(crlPem))
	params := TestCertRpc{}
	params.Jsonrpc = "2.0"
	params.Methond = "ptn_ccinvoketx"
	params.Id = 1
	form := "P127LdA9ZkfbD58iHR9QbJ2gtswV1cyokyF"
	to := "P127LdA9ZkfbD58iHR9QbJ2gtswV1cyokyF"
	amount := "100"
	fee := "1"
	contractid := syscontract.DigitalIdentityContractAddress.String()
	//method1 := []string{"P15UfoQzo93aSM3R2rDVHemiDJoMKRSLoaD","P15UfoQzo93aSM3R2rDVHemiDJoMKRSLoaD","100","1","PCGTta3M4t3yXu8uRgkKvaWd2d8DRv2vsEk"}
	method2 := []string{"addCRL",string(crlPem)}
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