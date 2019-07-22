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

//
//type TestCertRpc struct {
//	Jsonrpc string `json:"jsonrpc"`
//	Methond string `json:"method"`
//	Params  []interface{} `json:"params"`
//	Id      int    `json:"id"`
//}
//
//
//func TestGenCert(t *testing.T) {
//    addr := "P1PosMZZjMaB1VTR6jayWHUtEkhS87A7N3u"
//
//	certinfo := NewCertInfo("P15dtShJp7FpdgxCqT6RwZGKdfSSdkRk7GQ", "lkk", "HiPalletone", "user", "gptn.mediator1", true)
//	cainfo := CertInfo2Cainfo(*certinfo)
//	//发送请求到CA server 注册用户 生成证书
//	certpem, err := cainfo.Enrolluser()
//	if err != nil {
//		t.Log(err)
//	}
//	err = CertRpcReq("P15dtShJp7FpdgxCqT6RwZGKdfSSdkRk7GQ","P1CdJcmn4J7tUwbJpUNpML4XQ3F2nckJgiX",certpem,"http://localhost:8545")
//	t.Log(err)
//
//	}

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

//func TestCertCRL(t *testing.T) {
//	address := "P1KE4okveuNPmfrHCZQy7cgTczv9VG7npPQ"
//	reason := "no resaon"
//	certinfo := NewCertInfo(address, "lkk", "HiPalletone", "user", "gptn.mediator1", true)
//	cainfo := CertInfo2Cainfo(*certinfo)
//	crlPem,err := cainfo.Revoke(address, reason)
//	t.Log(err)
//
//   t.Log(string(crlPem))
//	params := TestCertRpc{}
//	params.Jsonrpc = "2.0"
//	params.Methond = "contract_ccinvoketx"
//	params.Id = 1
//	form := "P1CdJcmn4J7tUwbJpUNpML4XQ3F2nckJgiX"
//	to := "P1CdJcmn4J7tUwbJpUNpML4XQ3F2nckJgiX"
//	amount := "100"
//	fee := "1"
//	contractid := syscontract.DigitalIdentityContractAddress.String()
//	//method1 := []string{"P15UfoQzo93aSM3R2rDVHemiDJoMKRSLoaD","P15UfoQzo93aSM3R2rDVHemiDJoMKRSLoaD","100","1","PCGTta3M4t3yXu8uRgkKvaWd2d8DRv2vsEk"}
//	method2 := []string{"addCRL",string(crlPem)}
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

//func TestQueryCert(t *testing.T) {
//	//address := "P13pnwkoUmWCxcW6xmLfhpiPHJbWaztSaYg"
//	//certinfo := NewCertInfo(address, "lkk", "HiPalletone", "user", "gptn.mediator1", true)
//	//cainfo := CertInfo2Cainfo(*certinfo)
//
//	params := TestCertRpc{}
//	params.Jsonrpc = "2.0"
//	params.Methond = "contract_ccquery"
//	params.Id = 1
//	contractid := syscontract.DigitalIdentityContractAddress.String()
//	//method1 := []string{"P15UfoQzo93aSM3R2rDVHemiDJoMKRSLoaD","P15UfoQzo93aSM3R2rDVHemiDJoMKRSLoaD","100","1","PCGTta3M4t3yXu8uRgkKvaWd2d8DRv2vsEk"}
//	method2 := []string{"getRootCAHoler"}
//	//method2 := TestAddCertJson{ []string{"addServerCert","P1HrTpdqBmCrNhJMGREu7vtyzmhCiPiztkL","E:\\codes\\go\\src\\github.com\\palletone\\go-palletone\\cmd\\gptn\\data\\certs\\openssl\\powerca\\certs\\powerca.cert.pem"}}
//	params.Params = append(params.Params,contractid)
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
//	var result FatRpc
//	json.Unmarshal(body,&result)
//	t.Log(result.Result)
//	}