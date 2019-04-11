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
package main

import (
	"github.com/palletone/digital-identity/client"
	"crypto/x509"
	"encoding/pem"
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
	// Affiliation associates identity with particular organisation.  (gptn.mediator1...)
	Affiliation string `json:"affiliation"`
}

type CAGetCertChain struct {
	// RootCertificates is list of pem encoded certificates
	RootCertificates []*x509.Certificate
	// IntermediateCertificates is list of pem encoded intermediate certificates
	IntermediateCertificates []*pem.Block
	CAName string
	Version string
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
		RootCertificates:cc.RootCertificates,
		IntermediateCertificates:cc.IntermediateCertificates,
		CAName:cc.CAName,
		Version:cc.Version,
	}
}
func GenCert(certinfo CertINfo) error {
	cainfo := CertInfo2Cainfo(certinfo)

	err := cainfo.Enrolluser()
	if err != nil {
		return err
	}

	return nil
}

func RevokeCert(address string, reason string) error {
	caininfo := client.CaGenInfo{}
	err := caininfo.Revoke(address, reason)
	if err != nil {
		return err
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
func GetCaCertificateChain(caname string) (*CAGetCertChain,error) {
	cainfo := client.CaGenInfo{}
	certchain, err := cainfo.GetCaCertificateChain(caname)
	if err != nil {
		return nil,err
	}

	cc := CertChain2Result(*certchain)

	return &cc,nil
}