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


type TestCertRpc struct {
	Jsonrpc string `json:"jsonrpc"`
	Methond string `json:"method"`
	Params  []interface{} `json:"params"`
	Id      int    `json:"id"`
}


//func TestGenCert(t *testing.T) {
//	certinfo := NewCertInfo("P1HTX4D6fPUBee2uSFX3rwmVzUXAA8oa", "lkk", "HiPalletone", "user", "gptn.mediator1", true)
//	cainfo := CertInfo2Cainfo(*certinfo)
//	certpem, err := cainfo.Enrolluser()
//
//    t.Log(string(certpem))
//	params := TestCertRpc{}
//	params.Jsonrpc = "2.0"
//	params.Methond = "ptn_ccinvoketx"
//	params.Id = 1
//	form := "P1E6sjjWF6hkGFiZy9uWGBb1fTjKdt82yet"
//	to := "P1E6sjjWF6hkGFiZy9uWGBb1fTjKdt82yet"
//	amount := "100"
//	fee := "1"
//	contractid := syscontract.DigitalIdentityContractAddress.String()
//	//method1 := []string{"P15UfoQzo93aSM3R2rDVHemiDJoMKRSLoaD","P15UfoQzo93aSM3R2rDVHemiDJoMKRSLoaD","100","1","PCGTta3M4t3yXu8uRgkKvaWd2d8DRv2vsEk"}
//	method2 := []string{"addServerCert","P1HTX4D6fPUBee2uSFX3rwmVzUXAA8oa",string(certpem)}
//	//method2 := TestAddCertJson{ []string{"addServerCert","P1HrTpdqBmCrNhJMGREu7vtyzmhCiPiztkL","E:\\codes\\go\\src\\github.com\\palletone\\go-palletone\\cmd\\gptn\\data\\certs\\openssl\\powerca\\certs\\powerca.cert.pem"}}
//	params.Params = append(params.Params,form,to,amount,fee,contractid)
//	//params.Params = append(params.Params,str2)
//	params.Params = append(params.Params,method2)
//	reqJson, err := json.Marshal(params)
//	if err != nil {
//		t.Log(err)
//	}
//	t.Log(string(reqJson))
//	httpReq, err := http.NewRequest("POST", "http://localhost:8545", bytes.NewBuffer(reqJson))
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
//	}

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