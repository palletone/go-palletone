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
package certficate

import (
	"bytes"
	"crypto/x509"
	"encoding/pem"
	"github.com/palletone/digital-identity/client"
	"net/http"

	"encoding/json"
	"github.com/palletone/go-palletone/contracts/syscontract"
	"io/ioutil"
)

const (
	jsonrpc      = "2.0"
	invokemethod = "contract_ccinvoketx"
	querymethod  = "contract_ccquery"
	id           = 1
	amount       = "100"
	fee          = "1"
)

type CertRpc struct {
	Jsonrpc string        `json:"jsonrpc"`
	Methond string        `json:"method"`
	Params  []interface{} `json:"params"`
	Id      int           `json:"id"`
}

type FatRpc struct {
	Jsonrpc string `json:"jsonrpc"`
	Result  string `json:"result"`
	Id      int    `json:"id"`
}

type CertINfo struct {
	//The address as a certificate enrolleid
	Address string `json:"address"`
	//User name
	Name string `json:"name"`
	Data string `json:"data"`
	//ECert : certificate of registration
	ECert bool `json:"ecert"`

	// Type defines type of this identity (user,client, auditor etc...)
	Type string `json:"type"`
	// Affiliation associates identity with particular organisation.  (gptn.mediator1...)
	Affiliation string `json:"affiliation"`

	Key interface{} `json:"key"`
}

type CAGetCertChain struct {
	// RootCertificates is list of pem encoded certificates
	RootCertificates []*x509.Certificate
	// IntermediateCertificates is list of pem encoded intermediate certificates
	IntermediateCertificates []*pem.Block
	CAName                   string
	Version                  string
}

func NewCertInfo(address, name, data, ty, affiliation string, ecert bool) *CertINfo {
	return &CertINfo{
		Address:     address,
		Name:        name,
		Data:        data,
		ECert:       ecert,
		Type:        ty,
		Affiliation: affiliation,
	}
}

func CertInfo2Cainfo(certinfo CertINfo) client.CaGenInfo {
	return client.CaGenInfo{
		EnrolmentId: certinfo.Address,
		Name:        certinfo.Name,
		Data:        certinfo.Data,
		ECert:       certinfo.ECert,
		Type:        certinfo.Type,
		Affiliation: certinfo.Affiliation,
	}

}

func CertChain2Result(cc client.CAGetCertResponse) CAGetCertChain {
	return CAGetCertChain{
		RootCertificates:         cc.RootCertificates,
		IntermediateCertificates: cc.IntermediateCertificates,
		CAName:  cc.CAName,
		Version: cc.Version,
	}
}

func GenCert(certinfo CertINfo, cfg CAConfig) ([]byte, error) {
	cainfo := CertInfo2Cainfo(certinfo)
	//发送请求到CA server 注册用户 生成证书
	certpem, err := cainfo.Enrolluser()
	if err != nil {
		return nil, err
	}
	//immediateca := cfg.ImmediateCa
	//address := cainfo.EnrolmentId
	//url := cfg.CaUrl

	//将证书byte 用户地址 通过rpc调用进行存储
	//if certpem != nil {
	//
	//	err = CertRpcReq(address, immediateca, certpem,url)
	//	if err != nil {
	//		return err
	//	}
	//}

	return certpem, nil
}

func RevokeCert(address string, reason string, cfg CAConfig) error {
	caininfo := client.CaGenInfo{EnrolmentId: address}
	url := cfg.CaUrl
	crlPem, err := caininfo.Revoke(address, reason)
	if err != nil {
		return err
	}
	immediateca := cfg.ImmediateCa
	//吊销证书后将crl byte 通过rpc发送请求 添加到合约中
	if crlPem != nil {
		err = CrlRpcReq(immediateca, crlPem, url)
		if err != nil {
			return err
		}
	}
	return nil
}

func GetIndentity(address string, caname string) (*client.CAGetIdentityResponse, error) {
	cainfo := client.CaGenInfo{}
	idtRep, err := cainfo.GetIndentity(address, caname)
	if err != nil {
		return nil, err
	}

	return idtRep, nil
}

func GetIndentities() (*client.CAListAllIdentitesResponse, error) {
	cainfo := client.CaGenInfo{}
	idtReps, err := cainfo.GetIndentities()
	if err != nil {
		return nil, err
	}

	return idtReps, nil

}

//获取证书链信息
func GetCaCertificateChain(caname string) (*CAGetCertChain, error) {
	cainfo := client.CaGenInfo{}
	certchain, err := cainfo.GetCaCertificateChain(caname)
	if err != nil {
		return nil, err
	}

	cc := CertChain2Result(*certchain)

	return &cc, nil
}

// 获得某地址下关联的证书
func GetHolderCertIDs(address string, cfg CAConfig) (string, error) {
	method := "getHolderCertIDs"
	url := cfg.CaUrl
	body, err := QueryRpcReq(method, address, url)
	if err != nil {
		return "", err
	}
	var result FatRpc

	err = json.Unmarshal(body, &result)
	if err != nil {
		return "", err
	}

	return result.Result, nil
}

// 获得某地址用户颁发的所有证书ID信息
func GetIssuerCertsInfo(address string, cfg CAConfig) (string, error) {
	method := "getIssuerCertsInfo"
	url := cfg.CaUrl
	body, err := QueryRpcReq(method, address, url)
	if err != nil {
		return "", err
	}
	var result FatRpc

	err = json.Unmarshal(body, &result)
	if err != nil {
		return "", err
	}

	return result.Result, nil
}

//获得证书字节
func GetCertBytes(certid string, cfg CAConfig) (string, error) {
	method := "getCertBytes"
	url := cfg.CaUrl
	body, err := QueryRpcReq(method, certid, url)
	if err != nil {
		return "", err
	}
	var result FatRpc

	err = json.Unmarshal(body, &result)
	if err != nil {
		return "", err
	}

	return result.Result, nil
}

// 获取证书的持有者地址
func GetCertHolder(certid string, cfg CAConfig) (string, error) {
	method := "getCertHolder"
	url := cfg.CaUrl
	body, err := QueryRpcReq(method, certid, url)
	if err != nil {
		return "", err
	}
	var result FatRpc

	err = json.Unmarshal(body, &result)
	if err != nil {
		return "", err
	}

	return result.Result, nil
}

// 获取CA证书的持有者
func GetRootCAHoler(cfg CAConfig) (string, error) {
	method := "getRootCAHoler"
	url := cfg.CaUrl
	body, err := QueryRpcReq2(method, url)
	if err != nil {
		return "", err
	}
	var result FatRpc

	err = json.Unmarshal(body, &result)
	if err != nil {
		return "", err
	}

	return result.Result, nil
}

func RpcReq(params CertRpc, url string) ([]byte, error) {
	reqJson, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(reqJson))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpClient := &http.Client{}
	resp, err := httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func CertRpcReq(address string, immediateca string, certbyte []byte, url string) error {
	params := CertRpc{}
	params.Jsonrpc = jsonrpc
	params.Methond = invokemethod
	params.Id = id
	from := immediateca
	to := immediateca
	contractid := syscontract.DigitalIdentityContractAddress.String()

	method2 := []string{"addServerCert", address, string(certbyte)}
	params.Params = append(params.Params, from, to, amount, fee, contractid)

	params.Params = append(params.Params, method2)

	_, err := RpcReq(params, url)
	if err != nil {
		return err
	}
	return nil
}

func CrlRpcReq(immediateca string, crlbyte []byte, url string) error {
	params := CertRpc{}
	params.Jsonrpc = jsonrpc
	params.Methond = invokemethod
	params.Id = id
	from := immediateca
	to := immediateca
	contractid := syscontract.DigitalIdentityContractAddress.String()

	method2 := []string{"addCRL", string(crlbyte)}
	params.Params = append(params.Params, from, to, amount, fee, contractid)

	params.Params = append(params.Params, method2)

	_, err := RpcReq(params, url)
	if err != nil {
		return err
	}
	return nil
}

//GetIssuerCertsInfo rpc req
func QueryRpcReq(data string, method string, url string) ([]byte, error) {
	params := CertRpc{}
	params.Jsonrpc = jsonrpc
	params.Methond = querymethod
	params.Id = id
	contractid := syscontract.DigitalIdentityContractAddress.String()

	method2 := []string{data, method}
	params.Params = append(params.Params, contractid)

	params.Params = append(params.Params, method2)

	body, err := RpcReq(params, url)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func QueryRpcReq2(method string, url string) ([]byte, error) {
	params := CertRpc{}
	params.Jsonrpc = jsonrpc
	params.Methond = querymethod
	params.Id = id
	contractid := syscontract.DigitalIdentityContractAddress.String()

	method2 := []string{method}
	params.Params = append(params.Params, contractid)

	params.Params = append(params.Params, method2)

	body, err := RpcReq(params, url)
	if err != nil {
		return nil, err
	}
	return body, nil
}
