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
package client

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"io/ioutil"
	"os"
	"path"
	"github.com/palletone/digital-identity/config"
	"fmt"
)

type Identity struct {
	Certificate *x509.Certificate
	PrivateKey  interface{}
	MspID       string
}

func (i *Identity) SaveCert(ca *PalletCAClient, enreq *CaEnrollmentRequest, cainfo *CAGetCertResponse) error {
	var mspDir string
	var err error

	is, err := config.IsPathExists(ca.FilePath)
	if err != nil || !is {
		return err
	}
	//保存tls证书
	//	if enreq.Profile == "tls" {
	//		err = saveTLScert(ca, i, cainfo)
	//		if err != nil {
	//			return err
	//		}
	//		return nil
	//	}

	if enreq == nil {
		mspDir = path.Join(ca.FilePath, "/msp")
	} else {
		mspfile := enreq.EnrollmentID + "msp"
		mspDir = path.Join(ca.FilePath, mspfile)
	}
	//保存根证书
	caPath := path.Join(mspDir, "/cacerts")
	err = os.MkdirAll(caPath, os.ModePerm)
	if err != nil {
		return err
	}
	caFile := path.Join(caPath, "ca-cert.pem")
	caPem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: cainfo.RootCertificates[0].Raw,
		},
	)
	err = ioutil.WriteFile(caFile, caPem, 0644)
	if err != nil {
		return err
	}
	//保存中间证书
	if len(cainfo.IntermediateCertificates) > 0 {
		intercaPath := path.Join(mspDir, "/intermediatecerts")
		err = os.MkdirAll(intercaPath, os.ModePerm)
		if err != nil {
			return err
		}
		caFile = path.Join(intercaPath, "intermediate-certs.pem")
		for _, interca := range cainfo.IntermediateCertificates {
			intercaPem := pem.EncodeToMemory(interca)
			fd, openErr := os.OpenFile(caFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, os.ModePerm)
			if openErr != nil {
				return openErr
			}
			_, wirteErr := fd.Write(intercaPem)
			if wirteErr != nil {
				return wirteErr
			}
			_, wirteErr = fd.Write([]byte("\n"))
			if wirteErr != nil {
				return wirteErr
			}
			fd.Close()
		}
	}
	//保存证书
	certPath := path.Join(mspDir + "/signcerts")
	err = os.MkdirAll(certPath, os.ModePerm)
	if err != nil {
		return err
	}
	certFile := path.Join(certPath, "cert.pem")
	certPem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: i.Certificate.Raw,
		},
	)
	err = ioutil.WriteFile(certFile, certPem, 0644)
	if err != nil {
		return err
	}
	//保存私钥
	keyPath := path.Join(mspDir, "/keystore")
	err = os.MkdirAll(keyPath, os.ModePerm)
	if err != nil {
		return err
	}
	keyFile := path.Join(keyPath, "key.pem")
	keyByte, err := x509.MarshalECPrivateKey(i.PrivateKey.(*ecdsa.PrivateKey))
	if err != nil {
		return err
	}
	keyPem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "PRIVATE KEY",
			Bytes: keyByte,
		},
	)
	err = ioutil.WriteFile(keyFile, keyPem, 0644)
	if err != nil {
		return nil
	}
	return nil
}

//Save crl
func SaveCrl(ca *PalletCAClient, request *CARevocationRequest, result *CARevokeResult) ([]byte, error) {
	var err error
	mspfile := request.EnrollmentID + "msp"
	mspDir := path.Join(ca.FilePath, mspfile)
	crlPath := path.Join(mspDir, "/crls")
	err = os.MkdirAll(crlPath, os.ModePerm)
	if err != nil {
		return nil, err
	}
	crlFile := path.Join(crlPath, "crl.pem")

	crl, err := base64.StdEncoding.DecodeString(result.CRL)
	if err != nil {
		return nil, err
	}
	crlPem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "X509 CRL",
			Bytes: crl,
		},
	)
	err = ioutil.WriteFile(crlFile, crlPem, 0644)
	if err != nil {
		return nil, err
	}
	return crlPem, nil
}

func (i *Identity) SaveTLScert(ca *PalletCAClient, cainfo *CAGetCertResponse) error {
	var err error
	mspDir := path.Join(ca.FilePath, "/tlsmsp")

	//保存根证书
	caPath := path.Join(mspDir, "/tlscacerts")
	err = os.MkdirAll(caPath, os.ModePerm)
	if err != nil {
		return err
	}
	caFile := path.Join(caPath, "ca-cert.pem")
	caPem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: cainfo.RootCertificates[0].Raw,
		},
	)
	err = ioutil.WriteFile(caFile, caPem, 0644)
	if err != nil {
		return err
	}
	//保存中间证书
	//if len(cainfo.IntermediateCertificates) > 0 {
	//	intercaPath := path.Join(mspDir, "/tlsintermediatecerts")
	//	err = os.MkdirAll(intercaPath, os.ModePerm)
	//	if err != nil {
	//		return err
	//	}
	//	caFile = path.Join(intercaPath, "intermediate-certs.pem")
	//	for _, interca := range cainfo.IntermediateCertificates {
	//		intercaPem := pem.EncodeToMemory(interca)
	//		fd, openErr := os.OpenFile(caFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, os.ModePerm)
	//		if openErr != nil {
	//			return openErr
	//		}
	//		_, writErr := fd.Write(intercaPem)
	//		if writErr != nil {
	//			return writErr
	//		}
	//		_, writErr = fd.Write([]byte("\n"))
	//		if writErr != nil {
	//			return writErr
	//		}
	//		fd.Close()
	//	}
	//}
	//保存证书
	certPath := path.Join(mspDir + "/signcerts")
	err = os.MkdirAll(certPath, os.ModePerm)
	if err != nil {
		return err
	}
	certFile := path.Join(certPath, "cert.pem")
	certPem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: i.Certificate.Raw,
		},
	)
	err = ioutil.WriteFile(certFile, certPem, 0644)
	if err != nil {
		return err
	}
	//保存私钥
	keyPath := path.Join(mspDir, "/keystore")
	err = os.MkdirAll(keyPath, os.ModePerm)
	if err != nil {
		return err
	}
	keyFile := path.Join(keyPath, "key.pem")
	keyByte, err := x509.MarshalECPrivateKey(i.PrivateKey.(*ecdsa.PrivateKey))
	if err != nil {
		return err
	}
	keyPem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "PRIVATE KEY",
			Bytes: keyByte,
		},
	)
	err = ioutil.WriteFile(keyFile, keyPem, 0644)
	if err != nil {
		return nil
	}
	return nil
}

func (i *Identity) GetCertByte() []byte {
	fmt.Println(i.MspID)

	certPem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: i.Certificate.Raw,
		},
	)
	return certPem
}
