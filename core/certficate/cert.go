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
	"crypto/x509"
	"encoding/pem"
	"github.com/palletone/digital-identity/client"
)

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
	// Affiliation associates identity with particular organization.  (gptn.mediator1...)
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

func CertInfo2Cainfo(certinfo CertINfo) client.CaGenInfo {
	return client.CaGenInfo{
		EnrolmentID: certinfo.Address,
		Name:        certinfo.Name,
		Data:        certinfo.Data,
		ECert:       certinfo.ECert,
		Type:        certinfo.Type,
		Affiliation: certinfo.Affiliation,
		Key:         certinfo.Key,
	}

}

func GenCert(certinfo CertINfo) ([]byte, error) {
	cainfo := CertInfo2Cainfo(certinfo)
	//发送请求到CA server 注册用户 生成证书
	certpem, err := cainfo.Enrolluser()
	if err != nil {
		return nil, err
	}
	return certpem, nil
}

func RevokeCert(certinfo CertINfo, reason string) ([]byte, error) {
	cainfo := CertInfo2Cainfo(certinfo)
	crlPem, err := cainfo.Revoke(cainfo.EnrolmentID, reason)
	if err != nil {
		return nil, err
	}
	return crlPem, nil
}
